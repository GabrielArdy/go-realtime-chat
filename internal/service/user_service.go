package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"math"
	"strings"

	"realtime-api/internal/logger"
	"realtime-api/internal/model"
	"realtime-api/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

type UserService interface {
	CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, page, limit int) ([]*model.User, *model.PaginationMeta, error)
	AuthenticateUser(ctx context.Context, req *model.LoginRequest) (*model.User, error)
	UpdateUserStatus(ctx context.Context, userID uuid.UUID, status model.UserStatus) error
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*model.UserProfile, error)
	UpdateUserProfile(ctx context.Context, profile *model.UserProfile) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Check username
	existingUser, err = s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing username: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("username %s already taken", req.Username)
	}

	// Hash password
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &model.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsActive:  true,
		Status:    string(model.UserStatusOffline),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create user profile
	profile := &model.UserProfile{
		UserID:      user.ID,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DisplayName: fmt.Sprintf("%s %s", req.FirstName, req.LastName),
	}

	if err := s.userRepo.CreateOrUpdateProfile(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to create user profile: %w", err)
	}

	logger.Info("User created successfully", logger.WithFields(map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	}))

	return user, nil
}

func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, user *model.User) error {
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	logger.Info("User updated successfully", logger.WithField("user_id", user.ID))
	return nil
}

func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	logger.Info("User deleted successfully", logger.WithField("user_id", id))
	return nil
}

func (s *userService) ListUsers(ctx context.Context, page, limit int) ([]*model.User, *model.PaginationMeta, error) {
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

	users, total, err := s.userRepo.List(ctx, offset, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list users: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	meta := &model.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	return users, meta, nil
}

func (s *userService) AuthenticateUser(ctx context.Context, req *model.LoginRequest) (*model.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("user account is inactive")
	}

	if !verifyPassword(req.Password, user.Password) {
		logger.Warn("Failed login attempt", logger.WithFields(map[string]interface{}{
			"email": req.Email,
			"ip":    ctx.Value("ip"),
		}))
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last seen
	if err := s.userRepo.UpdateLastSeen(ctx, user.ID); err != nil {
		logger.Warn("Failed to update last seen", logger.WithField("user_id", user.ID))
	}

	logger.Info("User authenticated successfully", logger.WithFields(map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
	}))

	return user, nil
}

func (s *userService) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status model.UserStatus) error {
	if err := s.userRepo.UpdateStatus(ctx, userID, status); err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	logger.Info("User status updated", logger.WithFields(map[string]interface{}{
		"user_id": userID,
		"status":  status,
	}))

	return nil
}

func (s *userService) GetUserProfile(ctx context.Context, userID uuid.UUID) (*model.UserProfile, error) {
	profile, err := s.userRepo.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	if profile == nil {
		return nil, fmt.Errorf("user profile not found")
	}
	return profile, nil
}

func (s *userService) UpdateUserProfile(ctx context.Context, profile *model.UserProfile) error {
	if err := s.userRepo.CreateOrUpdateProfile(ctx, profile); err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	logger.Info("User profile updated", logger.WithField("user_id", profile.UserID))
	return nil
}

// Password hashing using Argon2
func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	// Encode salt and hash as base64
	saltB64 := base64.StdEncoding.EncodeToString(salt)
	hashB64 := base64.StdEncoding.EncodeToString(hash)

	return fmt.Sprintf("%s:%s", saltB64, hashB64), nil
}

func verifyPassword(password, hashedPassword string) bool {
	parts := strings.Split(hashedPassword, ":")
	if len(parts) != 2 {
		return false
	}

	salt, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}

	hash, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	testHash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	return subtle.ConstantTimeCompare(hash, testHash) == 1
}
