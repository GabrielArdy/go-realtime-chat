package repository

import (
	"context"
	"fmt"
	"time"

	"realtime-api/internal/database"
	"realtime-api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoomRepository interface {
	Create(ctx context.Context, room *model.Room) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Room, error)
	Update(ctx context.Context, room *model.Room) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetUserRooms(ctx context.Context, userID uuid.UUID) ([]model.Room, error)
	GetPublicRooms(ctx context.Context, offset, limit int) ([]model.Room, int64, error)
	SearchRooms(ctx context.Context, query string, offset, limit int) ([]model.Room, int64, error)

	// Room Member management
	AddMember(ctx context.Context, member *model.RoomMember) error
	RemoveMember(ctx context.Context, roomID, userID uuid.UUID) error
	GetRoomMembers(ctx context.Context, roomID uuid.UUID) ([]model.RoomMember, error)
	UpdateMemberRole(ctx context.Context, roomID, userID uuid.UUID, role string) error
	IsUserInRoom(ctx context.Context, roomID, userID uuid.UUID) (bool, error)

	// Room Invites
	CreateInvite(ctx context.Context, invite *model.RoomInvite) error
	GetInviteByCode(ctx context.Context, code string) (*model.RoomInvite, error)
	AcceptInvite(ctx context.Context, inviteID uuid.UUID) error
	RejectInvite(ctx context.Context, inviteID uuid.UUID) error
}

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository() RoomRepository {
	return &roomRepository{
		db: database.GetDB(),
	}
}

func (r *roomRepository) Create(ctx context.Context, room *model.Room) error {
	if err := r.db.WithContext(ctx).Create(room).Error; err != nil {
		return fmt.Errorf("failed to create room: %w", err)
	}
	return nil
}

func (r *roomRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Room, error) {
	var room model.Room
	if err := r.db.WithContext(ctx).
		Preload("CreatedByUser").
		Preload("Members").
		Preload("Members.User").
		First(&room, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get room by ID: %w", err)
	}
	return &room, nil
}

func (r *roomRepository) Update(ctx context.Context, room *model.Room) error {
	if err := r.db.WithContext(ctx).Save(room).Error; err != nil {
		return fmt.Errorf("failed to update room: %w", err)
	}
	return nil
}

func (r *roomRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.Room{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete room: %w", err)
	}
	return nil
}

func (r *roomRepository) GetUserRooms(ctx context.Context, userID uuid.UUID) ([]model.Room, error) {
	var rooms []model.Room
	if err := r.db.WithContext(ctx).
		Joins("JOIN room_members ON rooms.id = room_members.room_id").
		Where("room_members.user_id = ? AND room_members.deleted_at IS NULL", userID).
		Preload("CreatedByUser").
		Find(&rooms).Error; err != nil {
		return nil, fmt.Errorf("failed to get user rooms: %w", err)
	}
	return rooms, nil
}

func (r *roomRepository) GetPublicRooms(ctx context.Context, offset, limit int) ([]model.Room, int64, error) {
	var rooms []model.Room
	var total int64

	query := r.db.WithContext(ctx).Where("is_public = ?", true)

	// Count total records
	if err := query.Model(&model.Room{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count public rooms: %w", err)
	}

	// Get paginated results
	if err := query.Preload("CreatedByUser").Offset(offset).Limit(limit).Find(&rooms).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list public rooms: %w", err)
	}

	return rooms, total, nil
}

func (r *roomRepository) SearchRooms(ctx context.Context, query string, offset, limit int) ([]model.Room, int64, error) {
	var rooms []model.Room
	var total int64

	searchQuery := r.db.WithContext(ctx).Where("is_public = ? AND (name ILIKE ? OR description ILIKE ?)",
		true, "%"+query+"%", "%"+query+"%")

	// Count total records
	if err := searchQuery.Model(&model.Room{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count search rooms: %w", err)
	}

	// Get paginated results
	if err := searchQuery.Preload("CreatedByUser").Offset(offset).Limit(limit).Find(&rooms).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search rooms: %w", err)
	}

	return rooms, total, nil
}

func (r *roomRepository) AddMember(ctx context.Context, member *model.RoomMember) error {
	if err := r.db.WithContext(ctx).Create(member).Error; err != nil {
		return fmt.Errorf("failed to add room member: %w", err)
	}
	return nil
}

func (r *roomRepository) RemoveMember(ctx context.Context, roomID, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Delete(&model.RoomMember{}, "room_id = ? AND user_id = ?", roomID, userID).Error; err != nil {
		return fmt.Errorf("failed to remove room member: %w", err)
	}
	return nil
}

func (r *roomRepository) GetRoomMembers(ctx context.Context, roomID uuid.UUID) ([]model.RoomMember, error) {
	var members []model.RoomMember
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("room_id = ?", roomID).
		Find(&members).Error; err != nil {
		return nil, fmt.Errorf("failed to get room members: %w", err)
	}
	return members, nil
}

func (r *roomRepository) UpdateMemberRole(ctx context.Context, roomID, userID uuid.UUID, role string) error {
	if err := r.db.WithContext(ctx).Model(&model.RoomMember{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Update("role", role).Error; err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}
	return nil
}

func (r *roomRepository) IsUserInRoom(ctx context.Context, roomID, userID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.RoomMember{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check room membership: %w", err)
	}
	return count > 0, nil
}

func (r *roomRepository) CreateInvite(ctx context.Context, invite *model.RoomInvite) error {
	if err := r.db.WithContext(ctx).Create(invite).Error; err != nil {
		return fmt.Errorf("failed to create room invite: %w", err)
	}
	return nil
}

func (r *roomRepository) GetInviteByCode(ctx context.Context, code string) (*model.RoomInvite, error) {
	var invite model.RoomInvite
	if err := r.db.WithContext(ctx).
		Preload("Room").
		Preload("Inviter").
		Where("invite_code = ? AND expires_at > ?", code, time.Now()).
		First(&invite).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get invite by code: %w", err)
	}
	return &invite, nil
}

func (r *roomRepository) AcceptInvite(ctx context.Context, inviteID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&model.RoomInvite{}).
		Where("id = ?", inviteID).
		Update("status", "accepted").Error; err != nil {
		return fmt.Errorf("failed to accept invite: %w", err)
	}
	return nil
}

func (r *roomRepository) RejectInvite(ctx context.Context, inviteID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&model.RoomInvite{}).
		Where("id = ?", inviteID).
		Update("status", "rejected").Error; err != nil {
		return fmt.Errorf("failed to reject invite: %w", err)
	}
	return nil
}
