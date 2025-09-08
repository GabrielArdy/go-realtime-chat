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

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*model.User, int64, error)
	UpdateLastSeen(ctx context.Context, userID uuid.UUID) error
	UpdateStatus(ctx context.Context, userID uuid.UUID, status model.UserStatus) error
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*model.UserProfile, error)
	CreateOrUpdateProfile(ctx context.Context, profile *model.UserProfile) error
	GetUserContacts(ctx context.Context, userID uuid.UUID) ([]model.UserContact, error)
	AddContact(ctx context.Context, contact *model.UserContact) error
	RemoveContact(ctx context.Context, userID, contactID uuid.UUID) error
	UpdateContactStatus(ctx context.Context, userID, contactID uuid.UUID, status model.ContactStatus) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository() UserRepository {
	return &userRepository{
		db: database.GetDB(),
	}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Preload("Profile").First(&user, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Preload("Profile").Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Preload("Profile").Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.User{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).Preload("Profile").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

func (r *userRepository) UpdateLastSeen(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("last_seen", time.Now()).Error; err != nil {
		return fmt.Errorf("failed to update last seen: %w", err)
	}
	return nil
}

func (r *userRepository) UpdateStatus(ctx context.Context, userID uuid.UUID, status model.UserStatus) error {
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	return nil
}

func (r *userRepository) GetUserProfile(ctx context.Context, userID uuid.UUID) (*model.UserProfile, error) {
	var profile model.UserProfile
	if err := r.db.WithContext(ctx).First(&profile, "user_id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	return &profile, nil
}

func (r *userRepository) CreateOrUpdateProfile(ctx context.Context, profile *model.UserProfile) error {
	if err := r.db.WithContext(ctx).Save(profile).Error; err != nil {
		return fmt.Errorf("failed to save user profile: %w", err)
	}
	return nil
}

func (r *userRepository) GetUserContacts(ctx context.Context, userID uuid.UUID) ([]model.UserContact, error) {
	var contacts []model.UserContact
	if err := r.db.WithContext(ctx).Preload("Contact").Where("user_id = ?", userID).Find(&contacts).Error; err != nil {
		return nil, fmt.Errorf("failed to get user contacts: %w", err)
	}
	return contacts, nil
}

func (r *userRepository) AddContact(ctx context.Context, contact *model.UserContact) error {
	if err := r.db.WithContext(ctx).Create(contact).Error; err != nil {
		return fmt.Errorf("failed to add contact: %w", err)
	}
	return nil
}

func (r *userRepository) RemoveContact(ctx context.Context, userID, contactID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.UserContact{}, "user_id = ? AND contact_id = ?", userID, contactID).Error; err != nil {
		return fmt.Errorf("failed to remove contact: %w", err)
	}
	return nil
}

func (r *userRepository) UpdateContactStatus(ctx context.Context, userID, contactID uuid.UUID, status model.ContactStatus) error {
	if err := r.db.WithContext(ctx).Model(&model.UserContact{}).
		Where("user_id = ? AND contact_id = ?", userID, contactID).
		Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update contact status: %w", err)
	}
	return nil
}
