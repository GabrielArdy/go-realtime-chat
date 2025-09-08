package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"realtime-api/internal/events"
	"realtime-api/internal/logger"
	"realtime-api/internal/model"
	"realtime-api/internal/redis"
	"realtime-api/internal/repository"

	"github.com/google/uuid"
)

type RoomService interface {
	CreateRoom(ctx context.Context, req *model.CreateRoomRequest, creatorID uuid.UUID) (*model.Room, error)
	GetRoomByID(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) (*model.Room, error)
	UpdateRoom(ctx context.Context, roomID uuid.UUID, req *model.UpdateRoomRequest, userID uuid.UUID) (*model.Room, error)
	DeleteRoom(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) error
	GetUserRooms(ctx context.Context, userID uuid.UUID) ([]model.Room, error)
	ListUserChatRooms(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.Room, *model.PaginationMeta, error)
	GetPublicRooms(ctx context.Context, page, limit int) ([]model.Room, *model.PaginationMeta, error)
	SearchRooms(ctx context.Context, query string, page, limit int) ([]model.Room, *model.PaginationMeta, error)

	// Room Member Management
	JoinRoom(ctx context.Context, roomID, userID uuid.UUID) error
	LeaveRoom(ctx context.Context, roomID, userID uuid.UUID) error
	AddMember(ctx context.Context, roomID, userID, inviterID uuid.UUID) error
	RemoveMember(ctx context.Context, roomID, userID, removerID uuid.UUID) error
	GetRoomMembers(ctx context.Context, roomID uuid.UUID) ([]model.RoomMember, error)
	UpdateMemberRole(ctx context.Context, roomID, userID, updaterID uuid.UUID, role string) error

	// Room Invites
	CreateInvite(ctx context.Context, roomID, inviterID uuid.UUID, req *model.CreateInviteRequest) (*model.RoomInvite, error)
	AcceptInvite(ctx context.Context, inviteCode string, userID uuid.UUID) (*model.Room, error)
	RejectInvite(ctx context.Context, inviteCode string, userID uuid.UUID) error

	// Private Message Management
	CreateOrGetDirectRoom(ctx context.Context, userID1, userID2 uuid.UUID) (*model.Room, error)
}

type roomService struct {
	roomRepo       repository.RoomRepository
	userRepo       repository.UserRepository
	redis          *redis.Redis
	eventPublisher *events.EventPublisher
}

func NewRoomService(roomRepo repository.RoomRepository, userRepo repository.UserRepository, redis *redis.Redis) RoomService {
	return &roomService{
		roomRepo:       roomRepo,
		userRepo:       userRepo,
		redis:          redis,
		eventPublisher: events.NewEventPublisher(redis),
	}
}

func (s *roomService) CreateRoom(ctx context.Context, req *model.CreateRoomRequest, creatorID uuid.UUID) (*model.Room, error) {
	// Validate room type
	if req.Type != "direct" && req.Type != "group" && req.Type != "public" && req.Type != "broadcast" {
		return nil, fmt.Errorf("invalid room type")
	}

	// Create room
	room := &model.Room{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Avatar:      req.Avatar,
		IsPublic:    req.IsPublic != nil && *req.IsPublic,
		MaxMembers:  req.MaxMembers,
		CreatedBy:   creatorID,

		// Settings
		AllowFileUpload:      true,
		AllowVoiceMessages:   true,
		AllowVideoMessages:   true,
		MessageRetentionDays: 0,
		RequireApproval:      req.RequireApproval,
		MuteAllMembers:       false,
		OnlyAdminCanPost:     false,
	}

	if err := s.roomRepo.Create(ctx, room); err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	// Add creator as admin member
	member := &model.RoomMember{
		RoomID:   room.ID,
		UserID:   creatorID,
		Role:     "admin",
		JoinedAt: time.Now(),
	}

	if err := s.roomRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add creator as member: %w", err)
	}

	// Cache room membership
	if err := s.redis.AddUserToRoom(ctx, room.ID.String(), creatorID.String()); err != nil {
		logger.Warn("Failed to cache room membership", logger.WithField("error", err.Error()))
	}

	// Publish room creation event
	eventData := events.RoomEventData(room.ID, &creatorID, map[string]interface{}{
		"room_name": room.Name,
		"room_type": room.Type,
	})

	if err := s.eventPublisher.PublishRoomEvent(ctx, events.RoomCreate, room.ID, eventData, &creatorID); err != nil {
		logger.Warn("Failed to publish room creation event", logger.WithField("error", err.Error()))
	}

	logger.Info("Room created successfully", logger.WithFields(map[string]interface{}{
		"room_id":    room.ID,
		"creator_id": creatorID,
		"room_name":  room.Name,
		"room_type":  room.Type,
	}))

	return room, nil
}

func (s *roomService) GetRoomByID(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) (*model.Room, error) {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}
	if room == nil {
		return nil, fmt.Errorf("room not found")
	}

	// Check if user has access to the room
	if !room.IsPublic {
		isMember, err := s.roomRepo.IsUserInRoom(ctx, roomID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to check room membership: %w", err)
		}
		if !isMember {
			return nil, fmt.Errorf("access denied: user is not a member of this room")
		}
	}

	return room, nil
}

func (s *roomService) UpdateRoom(ctx context.Context, roomID uuid.UUID, req *model.UpdateRoomRequest, userID uuid.UUID) (*model.Room, error) {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}
	if room == nil {
		return nil, fmt.Errorf("room not found")
	}

	// Check if user is admin
	members, err := s.roomRepo.GetRoomMembers(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room members: %w", err)
	}

	isAdmin := false
	for _, member := range members {
		if member.UserID == userID && (member.Role == "admin" || member.Role == "owner") {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		return nil, fmt.Errorf("access denied: only admins can update room")
	}

	// Update room fields
	if req.Name != "" {
		room.Name = req.Name
	}
	if req.Description != "" {
		room.Description = req.Description
	}
	if req.Avatar != "" {
		room.Avatar = req.Avatar
	}
	if req.IsPublic != nil {
		room.IsPublic = *req.IsPublic
	}
	if req.MaxMembers > 0 {
		room.MaxMembers = req.MaxMembers
	}

	if err := s.roomRepo.Update(ctx, room); err != nil {
		return nil, fmt.Errorf("failed to update room: %w", err)
	}

	// Publish room update event
	eventData := events.RoomEventData(room.ID, &userID, map[string]interface{}{
		"room_name": room.Name,
	})

	if err := s.eventPublisher.PublishRoomEvent(ctx, events.RoomUpdate, room.ID, eventData, &userID); err != nil {
		logger.Warn("Failed to publish room update event", logger.WithField("error", err.Error()))
	}

	logger.Info("Room updated successfully", logger.WithFields(map[string]interface{}{
		"room_id":    room.ID,
		"updated_by": userID,
	}))

	return room, nil
}

func (s *roomService) DeleteRoom(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) error {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("failed to get room: %w", err)
	}
	if room == nil {
		return fmt.Errorf("room not found")
	}

	// Only room creator can delete
	if room.CreatedBy != userID {
		return fmt.Errorf("access denied: only room creator can delete room")
	}

	if err := s.roomRepo.Delete(ctx, roomID); err != nil {
		return fmt.Errorf("failed to delete room: %w", err)
	}

	// Publish room deletion event
	eventData := events.RoomEventData(room.ID, &userID, map[string]interface{}{
		"room_name": room.Name,
	})

	if err := s.eventPublisher.PublishRoomEvent(ctx, events.RoomDelete, room.ID, eventData, &userID); err != nil {
		logger.Warn("Failed to publish room deletion event", logger.WithField("error", err.Error()))
	}

	logger.Info("Room deleted successfully", logger.WithFields(map[string]interface{}{
		"room_id":    room.ID,
		"deleted_by": userID,
	}))

	return nil
}

func (s *roomService) GetUserRooms(ctx context.Context, userID uuid.UUID) ([]model.Room, error) {
	rooms, err := s.roomRepo.GetUserRooms(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user rooms: %w", err)
	}
	return rooms, nil
}

// ListUserChatRooms returns paginated list of user's chat rooms with additional metadata
func (s *roomService) ListUserChatRooms(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.Room, *model.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Get all user's rooms first
	allRooms, err := s.roomRepo.GetUserRooms(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user chat rooms: %w", err)
	}

	total := len(allRooms)

	// Apply pagination
	offset := (page - 1) * limit
	end := offset + limit
	if end > total {
		end = total
	}

	var rooms []model.Room
	if offset < total {
		rooms = allRooms[offset:end]
	}

	// Enrich rooms with additional metadata for chat list display
	for i := range rooms {
		// For direct rooms (2 members), get the other user's info for display
		if rooms[i].Type == "direct" {
			members, err := s.roomRepo.GetRoomMembers(ctx, rooms[i].ID)
			if err != nil {
				logger.Warn("Failed to get room members for direct room", logger.WithFields(map[string]interface{}{
					"room_id": rooms[i].ID,
					"error":   err.Error(),
				}))
				continue
			}

			// Find the other user in direct room
			for _, member := range members {
				if member.UserID != userID {
					otherUser, err := s.userRepo.GetByID(ctx, member.UserID)
					if err == nil && otherUser != nil {
						// Set room name to other user's name for display
						if rooms[i].Name == "" {
							rooms[i].Name = otherUser.Username
						}
						// Set avatar to other user's avatar if room doesn't have one
						if rooms[i].Avatar == "" && otherUser.Avatar != "" {
							rooms[i].Avatar = otherUser.Avatar
						}
					}
					break
				}
			}
		}

		// Log member count for debugging
		members, err := s.roomRepo.GetRoomMembers(ctx, rooms[i].ID)
		if err == nil {
			logger.Debug("Room member count", logger.WithFields(map[string]interface{}{
				"room_id":      rooms[i].ID,
				"member_count": len(members),
				"room_type":    rooms[i].Type,
			}))
		}
	}

	totalPages := (total + limit - 1) / limit

	meta := &model.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return rooms, meta, nil
}

func (s *roomService) GetPublicRooms(ctx context.Context, page, limit int) ([]model.Room, *model.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	rooms, total, err := s.roomRepo.GetPublicRooms(ctx, offset, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get public rooms: %w", err)
	}

	totalPages := (int(total) + limit - 1) / limit

	meta := &model.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	return rooms, meta, nil
}

func (s *roomService) SearchRooms(ctx context.Context, query string, page, limit int) ([]model.Room, *model.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	rooms, total, err := s.roomRepo.SearchRooms(ctx, query, offset, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to search rooms: %w", err)
	}

	totalPages := (int(total) + limit - 1) / limit

	meta := &model.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	return rooms, meta, nil
}

func (s *roomService) JoinRoom(ctx context.Context, roomID, userID uuid.UUID) error {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("failed to get room: %w", err)
	}
	if room == nil {
		return fmt.Errorf("room not found")
	}

	// Check if room is public or requires approval
	if !room.IsPublic && room.RequireApproval {
		return fmt.Errorf("room requires approval to join")
	}

	// Check if user is already a member
	isMember, err := s.roomRepo.IsUserInRoom(ctx, roomID, userID)
	if err != nil {
		return fmt.Errorf("failed to check room membership: %w", err)
	}
	if isMember {
		return fmt.Errorf("user is already a member of this room")
	}

	// Add user as member
	member := &model.RoomMember{
		RoomID:   roomID,
		UserID:   userID,
		Role:     "member",
		JoinedAt: time.Now(),
	}

	if err := s.roomRepo.AddMember(ctx, member); err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	// Cache room membership
	if err := s.redis.AddUserToRoom(ctx, roomID.String(), userID.String()); err != nil {
		logger.Warn("Failed to cache room membership", logger.WithField("error", err.Error()))
	}

	// Publish user join event
	eventData := events.RoomEventData(roomID, &userID, map[string]interface{}{
		"room_name": room.Name,
	})

	if err := s.eventPublisher.PublishRoomEvent(ctx, events.RoomJoin, roomID, eventData, &userID); err != nil {
		logger.Warn("Failed to publish user join event", logger.WithField("error", err.Error()))
	}

	logger.Info("User joined room successfully", logger.WithFields(map[string]interface{}{
		"room_id": roomID,
		"user_id": userID,
	}))

	return nil
}

func (s *roomService) LeaveRoom(ctx context.Context, roomID, userID uuid.UUID) error {
	// Check if user is a member
	isMember, err := s.roomRepo.IsUserInRoom(ctx, roomID, userID)
	if err != nil {
		return fmt.Errorf("failed to check room membership: %w", err)
	}
	if !isMember {
		return fmt.Errorf("user is not a member of this room")
	}

	if err := s.roomRepo.RemoveMember(ctx, roomID, userID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// Remove from cache
	if err := s.redis.RemoveUserFromRoom(ctx, roomID.String(), userID.String()); err != nil {
		logger.Warn("Failed to remove user from room cache", logger.WithField("error", err.Error()))
	}

	// Publish user leave event
	eventData := events.RoomEventData(roomID, &userID, map[string]interface{}{})

	if err := s.eventPublisher.PublishRoomEvent(ctx, events.RoomLeave, roomID, eventData, &userID); err != nil {
		logger.Warn("Failed to publish user leave event", logger.WithField("error", err.Error()))
	}

	logger.Info("User left room successfully", logger.WithFields(map[string]interface{}{
		"room_id": roomID,
		"user_id": userID,
	}))

	return nil
}

func (s *roomService) AddMember(ctx context.Context, roomID, userID, inviterID uuid.UUID) error {
	// Check if inviter is admin
	members, err := s.roomRepo.GetRoomMembers(ctx, roomID)
	if err != nil {
		return fmt.Errorf("failed to get room members: %w", err)
	}

	isAdmin := false
	for _, member := range members {
		if member.UserID == inviterID && (member.Role == "admin" || member.Role == "owner") {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		return fmt.Errorf("access denied: only admins can add members")
	}

	// Check if user is already a member
	isMember, err := s.roomRepo.IsUserInRoom(ctx, roomID, userID)
	if err != nil {
		return fmt.Errorf("failed to check room membership: %w", err)
	}
	if isMember {
		return fmt.Errorf("user is already a member of this room")
	}

	// Add user as member
	member := &model.RoomMember{
		RoomID:    roomID,
		UserID:    userID,
		Role:      "member",
		JoinedAt:  time.Now(),
		InvitedBy: &inviterID,
	}

	if err := s.roomRepo.AddMember(ctx, member); err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	// Cache room membership
	if err := s.redis.AddUserToRoom(ctx, roomID.String(), userID.String()); err != nil {
		logger.Warn("Failed to cache room membership", logger.WithField("error", err.Error()))
	}

	// Publish member add event
	eventData := events.RoomEventData(roomID, &userID, map[string]interface{}{
		"inviter_id": inviterID,
	})

	if err := s.eventPublisher.PublishRoomEvent(ctx, events.RoomMemberAdd, roomID, eventData, &inviterID); err != nil {
		logger.Warn("Failed to publish member add event", logger.WithField("error", err.Error()))
	}

	return nil
}

func (s *roomService) RemoveMember(ctx context.Context, roomID, userID, removerID uuid.UUID) error {
	// Get room to check type and properties
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("failed to get room: %w", err)
	}
	if room == nil {
		return errors.New("room not found")
	}

	// Check if remover is admin
	members, err := s.roomRepo.GetRoomMembers(ctx, roomID)
	if err != nil {
		return fmt.Errorf("failed to get room members: %w", err)
	}

	// Business rule: Cannot remove members from private rooms (2 members only)
	// Private messages (direct rooms with 2 members) should not allow member removal
	if len(members) == 2 && (room.Type == "direct" || room.Type == "private") {
		return errors.New("cannot remove members from private messages with only 2 participants")
	}

	isAdmin := false
	for _, member := range members {
		if member.UserID == removerID && (member.Role == "admin" || member.Role == "owner") {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		return fmt.Errorf("access denied: only admins can remove members")
	}

	if err := s.roomRepo.RemoveMember(ctx, roomID, userID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// Remove from cache
	if err := s.redis.RemoveUserFromRoom(ctx, roomID.String(), userID.String()); err != nil {
		logger.Warn("Failed to remove user from room cache", logger.WithField("error", err.Error()))
	}

	// Publish member remove event with additional context
	eventData := events.RoomEventData(roomID, &userID, map[string]interface{}{
		"remover_id":   removerID,
		"room_type":    room.Type,
		"member_count": len(members) - 1, // After removal
	})

	if err := s.eventPublisher.PublishRoomEvent(ctx, events.RoomMemberRemove, roomID, eventData, &removerID); err != nil {
		logger.Warn("Failed to publish member remove event", logger.WithField("error", err.Error()))
	}

	return nil
}

func (s *roomService) GetRoomMembers(ctx context.Context, roomID uuid.UUID) ([]model.RoomMember, error) {
	members, err := s.roomRepo.GetRoomMembers(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room members: %w", err)
	}
	return members, nil
}

func (s *roomService) UpdateMemberRole(ctx context.Context, roomID, userID, updaterID uuid.UUID, role string) error {
	// Check if updater is admin
	members, err := s.roomRepo.GetRoomMembers(ctx, roomID)
	if err != nil {
		return fmt.Errorf("failed to get room members: %w", err)
	}

	isAdmin := false
	for _, member := range members {
		if member.UserID == updaterID && (member.Role == "admin" || member.Role == "owner") {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		return fmt.Errorf("access denied: only admins can update member roles")
	}

	if err := s.roomRepo.UpdateMemberRole(ctx, roomID, userID, role); err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	return nil
}

func (s *roomService) CreateInvite(ctx context.Context, roomID, inviterID uuid.UUID, req *model.CreateInviteRequest) (*model.RoomInvite, error) {
	// Check if inviter is member
	isMember, err := s.roomRepo.IsUserInRoom(ctx, roomID, inviterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check room membership: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("access denied: only members can create invites")
	}

	// Generate invite code
	inviteCode := uuid.New().String()[:8] // Short invite code

	// Set expiration
	expiresAt := time.Now().Add(24 * time.Hour) // Default 24 hours
	if req.ExpiresIn > 0 {
		expiresAt = time.Now().Add(time.Duration(req.ExpiresIn) * time.Second)
	}

	invite := &model.RoomInvite{
		RoomID:     roomID,
		InviterID:  inviterID,
		InviteCode: inviteCode,
		ExpiresAt:  &expiresAt,
		Status:     "pending",
		MaxUses:    req.MaxUses,
		UsedCount:  0,
	}

	if err := s.roomRepo.CreateInvite(ctx, invite); err != nil {
		return nil, fmt.Errorf("failed to create invite: %w", err)
	}

	return invite, nil
}

func (s *roomService) AcceptInvite(ctx context.Context, inviteCode string, userID uuid.UUID) (*model.Room, error) {
	invite, err := s.roomRepo.GetInviteByCode(ctx, inviteCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get invite: %w", err)
	}
	if invite == nil {
		return nil, fmt.Errorf("invalid or expired invite")
	}

	// Check if invite is still valid
	if invite.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("invite has expired")
	}

	if invite.MaxUses > 0 && invite.UsedCount >= invite.MaxUses {
		return nil, fmt.Errorf("invite has reached maximum usage")
	}

	// Check if user is already a member
	isMember, err := s.roomRepo.IsUserInRoom(ctx, invite.RoomID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check room membership: %w", err)
	}
	if isMember {
		return nil, fmt.Errorf("user is already a member of this room")
	}

	// Add user as member
	member := &model.RoomMember{
		RoomID:    invite.RoomID,
		UserID:    userID,
		Role:      "member",
		JoinedAt:  time.Now(),
		InvitedBy: &invite.InviterID,
	}

	if err := s.roomRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	// Update invite usage
	if err := s.roomRepo.AcceptInvite(ctx, invite.ID); err != nil {
		logger.Warn("Failed to update invite usage", logger.WithField("error", err.Error()))
	}

	// Cache room membership
	if err := s.redis.AddUserToRoom(ctx, invite.RoomID.String(), userID.String()); err != nil {
		logger.Warn("Failed to cache room membership", logger.WithField("error", err.Error()))
	}

	// Get room details
	room, err := s.roomRepo.GetByID(ctx, invite.RoomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	return room, nil
}

func (s *roomService) RejectInvite(ctx context.Context, inviteCode string, userID uuid.UUID) error {
	invite, err := s.roomRepo.GetInviteByCode(ctx, inviteCode)
	if err != nil {
		return fmt.Errorf("failed to get invite: %w", err)
	}
	if invite == nil {
		return fmt.Errorf("invalid invite")
	}

	if err := s.roomRepo.RejectInvite(ctx, invite.ID); err != nil {
		return fmt.Errorf("failed to reject invite: %w", err)
	}

	return nil
}

// CreateOrGetDirectRoom creates a direct room between two users or returns existing one
func (s *roomService) CreateOrGetDirectRoom(ctx context.Context, user1ID, user2ID uuid.UUID) (*model.Room, error) {
	// Check if direct room already exists between these users
	user1Rooms, err := s.roomRepo.GetUserRooms(ctx, user1ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user rooms: %w", err)
	}

	// Look for existing direct room
	for _, room := range user1Rooms {
		if room.Type == "direct" {
			members, err := s.roomRepo.GetRoomMembers(ctx, room.ID)
			if err != nil {
				continue
			}

			// Check if this room has exactly 2 members and includes both users
			if len(members) == 2 {
				memberUserIDs := make(map[uuid.UUID]bool)
				for _, member := range members {
					memberUserIDs[member.UserID] = true
				}

				if memberUserIDs[user1ID] && memberUserIDs[user2ID] {
					return &room, nil
				}
			}
		}
	}

	// Create new direct room if none exists
	isPublic := false
	createReq := &model.CreateRoomRequest{
		Name:        "", // Direct rooms typically don't have names
		Description: "Direct message",
		Type:        "direct",
		IsPublic:    &isPublic,
	}

	room, err := s.CreateRoom(ctx, createReq, user1ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create direct room: %w", err)
	}

	// Add the second user to the room
	err = s.AddMember(ctx, room.ID, user2ID, user1ID)
	if err != nil {
		// If adding member fails, try to clean up the room
		if deleteErr := s.DeleteRoom(ctx, room.ID, user1ID); deleteErr != nil {
			logger.Error("Failed to cleanup room after member addition failure", logger.WithFields(map[string]interface{}{
				"room_id": room.ID,
				"error":   deleteErr.Error(),
			}))
		}
		return nil, fmt.Errorf("failed to add second user to direct room: %w", err)
	}

	// Publish direct room created event using existing event system
	eventData := events.RoomEventData(room.ID, &user1ID, map[string]interface{}{
		"room_type": "direct",
		"user2_id":  user2ID,
	})

	if err := s.eventPublisher.PublishRoomEvent(ctx, events.RoomCreate, room.ID, eventData, &user1ID); err != nil {
		logger.Error("Failed to publish direct room created event", logger.WithFields(map[string]interface{}{
			"room_id": room.ID,
			"error":   err.Error(),
		}))
	}

	return room, nil
}
