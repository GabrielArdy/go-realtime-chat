package service

import (
	"context"
	"fmt"
	"time"

	"realtime-api/internal/events"
	"realtime-api/internal/logger"
	"realtime-api/internal/model"
	"realtime-api/internal/redis"
	"realtime-api/internal/repository"

	"github.com/google/uuid"
)

type MessageService interface {
	SendMessage(ctx context.Context, req *model.SendMessageRequest, senderID uuid.UUID) (*model.Message, error)
	GetMessages(ctx context.Context, roomID uuid.UUID, userID uuid.UUID, page, limit int) ([]model.Message, *model.PaginationMeta, error)
	GetMessageByID(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) (*model.Message, error)
	EditMessage(ctx context.Context, messageID uuid.UUID, req *model.EditMessageRequest, userID uuid.UUID) (*model.Message, error)
	DeleteMessage(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) error

	// Message Reactions
	ReactToMessage(ctx context.Context, messageID uuid.UUID, req *model.ReactToMessageRequest, userID uuid.UUID) error
	RemoveReaction(ctx context.Context, messageID uuid.UUID, emoji string, userID uuid.UUID) error

	// Message Read Status
	MarkAsRead(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) error

	// Typing Indicators
	StartTyping(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) error
	StopTyping(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) error
}

type messageService struct {
	messageRepo    repository.MessageRepository
	roomRepo       repository.RoomRepository
	userRepo       repository.UserRepository
	redis          *redis.Redis
	eventPublisher *events.EventPublisher
}

func NewMessageService(messageRepo repository.MessageRepository, roomRepo repository.RoomRepository, userRepo repository.UserRepository, redis *redis.Redis) MessageService {
	return &messageService{
		messageRepo:    messageRepo,
		roomRepo:       roomRepo,
		userRepo:       userRepo,
		redis:          redis,
		eventPublisher: events.NewEventPublisher(redis),
	}
}

func (s *messageService) SendMessage(ctx context.Context, req *model.SendMessageRequest, senderID uuid.UUID) (*model.Message, error) {
	// Validate sender is member of the room
	isMember, err := s.roomRepo.IsUserInRoom(ctx, req.RoomID, senderID)
	if err != nil {
		return nil, fmt.Errorf("failed to check room membership: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("access denied: user is not a member of this room")
	}

	// Get room to check settings
	room, err := s.roomRepo.GetByID(ctx, req.RoomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}
	if room == nil {
		return nil, fmt.Errorf("room not found")
	}

	// Check if room allows posting from this user
	if room.OnlyAdminCanPost {
		members, err := s.roomRepo.GetRoomMembers(ctx, req.RoomID)
		if err != nil {
			return nil, fmt.Errorf("failed to get room members: %w", err)
		}

		isAdmin := false
		for _, member := range members {
			if member.UserID == senderID && (member.Role == "admin" || member.Role == "owner") {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			return nil, fmt.Errorf("access denied: only admins can post in this room")
		}
	}

	// Validate message type
	if req.Type == "" {
		req.Type = "text"
	}

	// Create message
	message := &model.Message{
		RoomID:    req.RoomID,
		SenderID:  senderID,
		Type:      req.Type,
		Content:   req.Content,
		Metadata:  req.Metadata,
		ReplyToID: req.ReplyToID,
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Load message with relationships
	messageWithDetails, err := s.messageRepo.GetByID(ctx, message.ID)
	if err != nil {
		logger.Warn("Failed to load message with details", logger.WithField("error", err.Error()))
		messageWithDetails = message
	}

	// Publish message to Redis for real-time delivery
	eventData := events.MessageEventData(message.ID, message.RoomID, &message.SenderID, map[string]interface{}{
		"type":        message.Type,
		"content":     message.Content,
		"metadata":    message.Metadata,
		"reply_to_id": message.ReplyToID,
		"created_at":  message.CreatedAt,
	})

	if err := s.eventPublisher.PublishMessageEvent(ctx, events.MessageSend, message.RoomID, message.ID, eventData, &message.SenderID); err != nil {
		logger.Warn("Failed to publish message to Redis", logger.WithField("error", err.Error()))
	}

	// Stop typing indicator for sender
	if err := s.StopTyping(ctx, req.RoomID, senderID); err != nil {
		logger.Warn("Failed to stop typing indicator", logger.WithField("error", err.Error()))
	}

	logger.Info("Message sent successfully", logger.WithFields(map[string]interface{}{
		"message_id": message.ID,
		"room_id":    message.RoomID,
		"sender_id":  message.SenderID,
		"type":       message.Type,
	}))

	return messageWithDetails, nil
}

func (s *messageService) GetMessages(ctx context.Context, roomID uuid.UUID, userID uuid.UUID, page, limit int) ([]model.Message, *model.PaginationMeta, error) {
	// Check if user is member of the room
	isMember, err := s.roomRepo.IsUserInRoom(ctx, roomID, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check room membership: %w", err)
	}
	if !isMember {
		return nil, nil, fmt.Errorf("access denied: user is not a member of this room")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	messages, total, err := s.messageRepo.GetRoomMessages(ctx, roomID, offset, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get messages: %w", err)
	}

	totalPages := (int(total) + limit - 1) / limit

	meta := &model.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	return messages, meta, nil
}

func (s *messageService) GetMessageByID(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) (*model.Message, error) {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	if message == nil {
		return nil, fmt.Errorf("message not found")
	}

	// Check if user is member of the room
	isMember, err := s.roomRepo.IsUserInRoom(ctx, message.RoomID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check room membership: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("access denied: user is not a member of this room")
	}

	return message, nil
}

func (s *messageService) EditMessage(ctx context.Context, messageID uuid.UUID, req *model.EditMessageRequest, userID uuid.UUID) (*model.Message, error) {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	if message == nil {
		return nil, fmt.Errorf("message not found")
	}

	// Check if user is the sender
	if message.SenderID != userID {
		return nil, fmt.Errorf("access denied: only the sender can edit this message")
	}

	// Check if message is too old to edit (optional)
	if time.Since(message.CreatedAt) > 24*time.Hour {
		return nil, fmt.Errorf("message is too old to edit")
	}

	// Update message
	message.Content = req.Content
	message.Metadata = req.Metadata
	message.IsEdited = true
	message.EditedAt = &[]time.Time{time.Now()}[0]

	if err := s.messageRepo.Update(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	// Publish message edit event
	eventData := events.MessageEventData(message.ID, message.RoomID, &message.SenderID, map[string]interface{}{
		"content":   message.Content,
		"metadata":  message.Metadata,
		"is_edited": message.IsEdited,
		"edited_at": message.EditedAt,
	})

	if err := s.eventPublisher.PublishMessageEvent(ctx, events.MessageEdit, message.RoomID, message.ID, eventData, &message.SenderID); err != nil {
		logger.Warn("Failed to publish message edit event", logger.WithField("error", err.Error()))
	}

	logger.Info("Message edited successfully", logger.WithFields(map[string]interface{}{
		"message_id": message.ID,
		"room_id":    message.RoomID,
		"sender_id":  message.SenderID,
	}))

	return message, nil
}

func (s *messageService) DeleteMessage(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) error {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}
	if message == nil {
		return fmt.Errorf("message not found")
	}

	// Check if user is the sender or admin
	canDelete := message.SenderID == userID

	if !canDelete {
		// Check if user is admin in the room
		members, err := s.roomRepo.GetRoomMembers(ctx, message.RoomID)
		if err != nil {
			return fmt.Errorf("failed to get room members: %w", err)
		}

		for _, member := range members {
			if member.UserID == userID && (member.Role == "admin" || member.Role == "owner") {
				canDelete = true
				break
			}
		}
	}

	if !canDelete {
		return fmt.Errorf("access denied: only the sender or room admin can delete this message")
	}

	// Soft delete by marking as deleted
	message.IsDeleted = true
	message.Content = "This message was deleted"
	message.Metadata = ""

	if err := s.messageRepo.Update(ctx, message); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	// Publish message deletion event
	eventData := events.MessageEventData(message.ID, message.RoomID, &userID, map[string]interface{}{
		"is_deleted": message.IsDeleted,
	})

	if err := s.eventPublisher.PublishMessageEvent(ctx, events.MessageDelete, message.RoomID, message.ID, eventData, &userID); err != nil {
		logger.Warn("Failed to publish message deletion event", logger.WithField("error", err.Error()))
	}

	logger.Info("Message deleted successfully", logger.WithFields(map[string]interface{}{
		"message_id": message.ID,
		"room_id":    message.RoomID,
		"deleted_by": userID,
	}))

	return nil
}

func (s *messageService) ReactToMessage(ctx context.Context, messageID uuid.UUID, req *model.ReactToMessageRequest, userID uuid.UUID) error {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}
	if message == nil {
		return fmt.Errorf("message not found")
	}

	// Check if user is member of the room
	isMember, err := s.roomRepo.IsUserInRoom(ctx, message.RoomID, userID)
	if err != nil {
		return fmt.Errorf("failed to check room membership: %w", err)
	}
	if !isMember {
		return fmt.Errorf("access denied: user is not a member of this room")
	}

	// Add or update reaction
	reaction := &model.MessageReaction{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     req.Emoji,
	}

	if err := s.messageRepo.AddReaction(ctx, reaction); err != nil {
		return fmt.Errorf("failed to add reaction: %w", err)
	}

	// Publish reaction event
	eventData := events.MessageEventData(messageID, message.RoomID, &userID, map[string]interface{}{
		"emoji": req.Emoji,
	})

	if err := s.eventPublisher.PublishMessageEvent(ctx, events.MessageReactionAdd, message.RoomID, messageID, eventData, &userID); err != nil {
		logger.Warn("Failed to publish reaction event", logger.WithField("error", err.Error()))
	}

	return nil
}

func (s *messageService) RemoveReaction(ctx context.Context, messageID uuid.UUID, emoji string, userID uuid.UUID) error {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}
	if message == nil {
		return fmt.Errorf("message not found")
	}

	if err := s.messageRepo.RemoveReaction(ctx, messageID, userID, emoji); err != nil {
		return fmt.Errorf("failed to remove reaction: %w", err)
	}

	// Publish reaction removal event
	eventData := events.MessageEventData(messageID, message.RoomID, &userID, map[string]interface{}{
		"emoji": emoji,
	})

	if err := s.eventPublisher.PublishMessageEvent(ctx, events.MessageReactionRemove, message.RoomID, messageID, eventData, &userID); err != nil {
		logger.Warn("Failed to publish reaction removal event", logger.WithField("error", err.Error()))
	}

	return nil
}

func (s *messageService) MarkAsRead(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) error {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}
	if message == nil {
		return fmt.Errorf("message not found")
	}

	// Check if user is member of the room
	isMember, err := s.roomRepo.IsUserInRoom(ctx, message.RoomID, userID)
	if err != nil {
		return fmt.Errorf("failed to check room membership: %w", err)
	}
	if !isMember {
		return fmt.Errorf("access denied: user is not a member of this room")
	}

	// Mark message as read
	if err := s.messageRepo.MarkAsRead(ctx, messageID, userID); err != nil {
		return fmt.Errorf("failed to mark message as read: %w", err)
	}

	// Publish read event
	eventData := events.MessageEventData(messageID, message.RoomID, &userID, map[string]interface{}{
		"read_at": time.Now(),
	})

	if err := s.eventPublisher.PublishMessageEvent(ctx, events.MessageRead, message.RoomID, messageID, eventData, &userID); err != nil {
		logger.Warn("Failed to publish read event", logger.WithField("error", err.Error()))
	}

	return nil
}

func (s *messageService) StartTyping(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) error {
	// Check if user is member of the room
	isMember, err := s.roomRepo.IsUserInRoom(ctx, roomID, userID)
	if err != nil {
		return fmt.Errorf("failed to check room membership: %w", err)
	}
	if !isMember {
		return fmt.Errorf("access denied: user is not a member of this room")
	}

	// Publish typing start event
	if err := s.eventPublisher.PublishTypingEvent(ctx, roomID, userID, true); err != nil {
		return fmt.Errorf("failed to publish typing event: %w", err)
	}

	return nil
}

func (s *messageService) StopTyping(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) error {
	// Publish typing stop event
	if err := s.eventPublisher.PublishTypingEvent(ctx, roomID, userID, false); err != nil {
		return fmt.Errorf("failed to publish typing event: %w", err)
	}

	return nil
}
