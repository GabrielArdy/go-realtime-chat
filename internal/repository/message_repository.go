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

type MessageRepository interface {
	Create(ctx context.Context, message *model.Message) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Message, error)
	Update(ctx context.Context, message *model.Message) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetRoomMessages(ctx context.Context, roomID uuid.UUID, offset, limit int) ([]model.Message, int64, error)
	GetMessagesSince(ctx context.Context, roomID uuid.UUID, since time.Time) ([]model.Message, error)
	SearchMessages(ctx context.Context, roomID uuid.UUID, query string, offset, limit int) ([]model.Message, int64, error)
	MarkAsRead(ctx context.Context, messageID, userID uuid.UUID) error
	GetUnreadCount(ctx context.Context, roomID, userID uuid.UUID) (int64, error)

	// Message Attachments
	AddAttachment(ctx context.Context, attachment *model.MessageAttachment) error
	GetMessageAttachments(ctx context.Context, messageID uuid.UUID) ([]model.MessageAttachment, error)
	DeleteAttachment(ctx context.Context, attachmentID uuid.UUID) error

	// Message Reactions
	AddReaction(ctx context.Context, reaction *model.MessageReaction) error
	RemoveReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error
	GetMessageReactions(ctx context.Context, messageID uuid.UUID) ([]model.MessageReaction, error)

	// Message Threading
	GetThreadMessages(ctx context.Context, parentMessageID uuid.UUID, offset, limit int) ([]model.Message, int64, error)
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository() MessageRepository {
	return &messageRepository{
		db: database.GetDB(),
	}
}

func (r *messageRepository) Create(ctx context.Context, message *model.Message) error {
	if err := r.db.WithContext(ctx).Create(message).Error; err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}
	return nil
}

func (r *messageRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Message, error) {
	var message model.Message
	if err := r.db.WithContext(ctx).
		Preload("Sender").
		Preload("Attachments").
		Preload("Reactions").
		Preload("Reactions.User").
		First(&message, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get message by ID: %w", err)
	}
	return &message, nil
}

func (r *messageRepository) Update(ctx context.Context, message *model.Message) error {
	if err := r.db.WithContext(ctx).Save(message).Error; err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}
	return nil
}

func (r *messageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.Message{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}

func (r *messageRepository) GetRoomMessages(ctx context.Context, roomID uuid.UUID, offset, limit int) ([]model.Message, int64, error) {
	var messages []model.Message
	var total int64

	query := r.db.WithContext(ctx).Where("room_id = ?", roomID)

	// Count total records
	if err := query.Model(&model.Message{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count room messages: %w", err)
	}

	// Get paginated results
	if err := query.
		Preload("Sender").
		Preload("Attachments").
		Preload("Reactions").
		Preload("Reactions.User").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get room messages: %w", err)
	}

	return messages, total, nil
}

func (r *messageRepository) GetMessagesSince(ctx context.Context, roomID uuid.UUID, since time.Time) ([]model.Message, error) {
	var messages []model.Message
	if err := r.db.WithContext(ctx).
		Where("room_id = ? AND created_at > ?", roomID, since).
		Preload("Sender").
		Preload("Attachments").
		Preload("Reactions").
		Preload("Reactions.User").
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to get messages since: %w", err)
	}
	return messages, nil
}

func (r *messageRepository) SearchMessages(ctx context.Context, roomID uuid.UUID, query string, offset, limit int) ([]model.Message, int64, error) {
	var messages []model.Message
	var total int64

	searchQuery := r.db.WithContext(ctx).
		Where("room_id = ? AND content ILIKE ?", roomID, "%"+query+"%")

	// Count total records
	if err := searchQuery.Model(&model.Message{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count search messages: %w", err)
	}

	// Get paginated results
	if err := searchQuery.
		Preload("Sender").
		Preload("Attachments").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search messages: %w", err)
	}

	return messages, total, nil
}

func (r *messageRepository) MarkAsRead(ctx context.Context, messageID, userID uuid.UUID) error {
	// Check if read receipt already exists
	var existing model.MessageRead
	err := r.db.WithContext(ctx).
		Where("message_id = ? AND user_id = ?", messageID, userID).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new read receipt
		receipt := &model.MessageRead{
			MessageID: messageID,
			UserID:    userID,
			ReadAt:    time.Now(),
		}

		if err := r.db.WithContext(ctx).Create(receipt).Error; err != nil {
			return fmt.Errorf("failed to create read receipt: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check existing read receipt: %w", err)
	} else {
		// Update existing read receipt
		existing.ReadAt = time.Now()
		if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
			return fmt.Errorf("failed to update read receipt: %w", err)
		}
	}

	return nil
}

func (r *messageRepository) GetUnreadCount(ctx context.Context, roomID, userID uuid.UUID) (int64, error) {
	var count int64

	// Count messages in room that don't have read receipts for this user
	if err := r.db.WithContext(ctx).
		Model(&model.Message{}).
		Where("room_id = ? AND sender_id != ?", roomID, userID).
		Where("id NOT IN (?)",
			r.db.Select("message_id").
				Table("message_reads").
				Where("user_id = ?", userID),
		).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

func (r *messageRepository) AddAttachment(ctx context.Context, attachment *model.MessageAttachment) error {
	if err := r.db.WithContext(ctx).Create(attachment).Error; err != nil {
		return fmt.Errorf("failed to add attachment: %w", err)
	}
	return nil
}

func (r *messageRepository) GetMessageAttachments(ctx context.Context, messageID uuid.UUID) ([]model.MessageAttachment, error) {
	var attachments []model.MessageAttachment
	if err := r.db.WithContext(ctx).
		Where("message_id = ?", messageID).
		Find(&attachments).Error; err != nil {
		return nil, fmt.Errorf("failed to get message attachments: %w", err)
	}
	return attachments, nil
}

func (r *messageRepository) DeleteAttachment(ctx context.Context, attachmentID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.MessageAttachment{}, "id = ?", attachmentID).Error; err != nil {
		return fmt.Errorf("failed to delete attachment: %w", err)
	}
	return nil
}

func (r *messageRepository) AddReaction(ctx context.Context, reaction *model.MessageReaction) error {
	if err := r.db.WithContext(ctx).Create(reaction).Error; err != nil {
		return fmt.Errorf("failed to add reaction: %w", err)
	}
	return nil
}

func (r *messageRepository) RemoveReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	if err := r.db.WithContext(ctx).
		Delete(&model.MessageReaction{}, "message_id = ? AND user_id = ? AND emoji = ?", messageID, userID, emoji).Error; err != nil {
		return fmt.Errorf("failed to remove reaction: %w", err)
	}
	return nil
}

func (r *messageRepository) GetMessageReactions(ctx context.Context, messageID uuid.UUID) ([]model.MessageReaction, error) {
	var reactions []model.MessageReaction
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("message_id = ?", messageID).
		Find(&reactions).Error; err != nil {
		return nil, fmt.Errorf("failed to get message reactions: %w", err)
	}
	return reactions, nil
}

func (r *messageRepository) GetThreadMessages(ctx context.Context, parentMessageID uuid.UUID, offset, limit int) ([]model.Message, int64, error) {
	var messages []model.Message
	var total int64

	query := r.db.WithContext(ctx).Where("parent_message_id = ?", parentMessageID)

	// Count total records
	if err := query.Model(&model.Message{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count thread messages: %w", err)
	}

	// Get paginated results
	if err := query.
		Preload("Sender").
		Preload("Attachments").
		Preload("Reactions").
		Preload("Reactions.User").
		Order("created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get thread messages: %w", err)
	}

	return messages, total, nil
}
