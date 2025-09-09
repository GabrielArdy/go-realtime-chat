package handler

import (
	"net/http"
	"strconv"

	"realtime-api/internal/jwt"
	"realtime-api/internal/logger"
	"realtime-api/internal/model"
	"realtime-api/internal/service"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) RegisterUser(c echo.Context) error {
	var req model.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	// Validate required fields
	if req.Username == "" {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Username is required",
		})
	}
	if req.Email == "" {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Email is required",
		})
	}
	if req.Password == "" {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Password is required",
		})
	}
	if req.FirstName == "" {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "First name is required",
		})
	}
	if req.LastName == "" {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Last name is required",
		})
	}

	// Additional validation
	if len(req.Username) < 3 {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Username must be at least 3 characters long",
		})
	}
	if len(req.Password) < 6 {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Password must be at least 6 characters long",
		})
	}

	// Create user
	user, err := h.userService.CreateUser(c.Request().Context(), &req)
	if err != nil {
		logger.Error("Failed to register user", logger.WithField("error", err.Error()))
		
		// Check for specific errors
		if err.Error() == "user with email "+req.Email+" already exists" {
			return c.JSON(http.StatusConflict, model.APIResponse{
				Success: false,
				Message: "Email address is already registered",
			})
		}
		if err.Error() == "username "+req.Username+" already taken" {
			return c.JSON(http.StatusConflict, model.APIResponse{
				Success: false,
				Message: "Username is already taken",
			})
		}
		
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Failed to register user",
			Error:   "Registration failed, please try again",
		})
	}

	// Remove sensitive information from response
	user.Password = ""

	// Generate JWT token for immediate login after registration
	sessionID := uuid.New()
	deviceID := c.Request().Header.Get("User-Agent")
	if deviceID == "" {
		deviceID = "unknown-device"
	}

	jwtService := jwt.GetService()
	if jwtService == nil {
		logger.Error("JWT service not initialized")
		// Still return success for registration, but without tokens
		return c.JSON(http.StatusCreated, model.APIResponse{
			Success: true,
			Message: "User registered successfully. Please login to continue.",
			Data:    user,
		})
	}

	accessToken, refreshToken, expiresAt, err := jwtService.GenerateTokens(user, sessionID, deviceID)
	if err != nil {
		logger.Error("Failed to generate JWT tokens after registration", logger.WithField("error", err.Error()))
		// Still return success for registration, but without tokens
		return c.JSON(http.StatusCreated, model.APIResponse{
			Success: true,
			Message: "User registered successfully. Please login to continue.",
			Data:    user,
		})
	}

	logger.Info("User registered successfully", logger.WithFields(map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	}))

	return c.JSON(http.StatusCreated, model.APIResponse{
		Success: true,
		Message: "User registered successfully",
		Data: map[string]interface{}{
			"user":          user,
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"expires_at":    expiresAt,
			"session_id":    sessionID,
		},
	})
}

func (h *UserHandler) CreateUser(c echo.Context) error {
	var req model.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	// TODO: Add validation here
	if req.Username == "" {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Username is required",
		})
	}
	if req.Password == "" {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Password is required",
		})
	}
	if req.Email == "" {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Email is required",
		})
	}

	user, err := h.userService.CreateUser(c.Request().Context(), &req)
	if err != nil {
		logger.Error("Failed to create user", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Failed to create user",
			Error:   err.Error(),
		})
	}

	// Remove password from response
	user.Password = ""

	return c.JSON(http.StatusCreated, model.APIResponse{
		Success: true,
		Message: "User created successfully",
		Data:    user,
	})
}

func (h *UserHandler) GetUser(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid user ID format",
			Error:   err.Error(),
		})
	}

	user, err := h.userService.GetUserByID(c.Request().Context(), id)
	if err != nil {
		logger.Error("Failed to get user", logger.WithFields(map[string]interface{}{
			"user_id": id,
			"error":   err.Error(),
		}))
		return c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Message: "User not found",
		})
	}

	// Remove password from response
	user.Password = ""

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "User retrieved successfully",
		Data:    user,
	})
}

func (h *UserHandler) ListUsers(c echo.Context) error {
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	users, meta, err := h.userService.ListUsers(c.Request().Context(), page, limit)
	if err != nil {
		logger.Error("Failed to list users", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Message: "Failed to retrieve users",
			Error:   err.Error(),
		})
	}

	// Remove passwords from response
	for _, user := range users {
		user.Password = ""
	}

	response := model.PaginatedResponse{
		APIResponse: model.APIResponse{
			Success: true,
			Message: "Users retrieved successfully",
			Data:    users,
		},
		Meta: *meta,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *UserHandler) LoginUser(c echo.Context) error {
	var req model.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	user, err := h.userService.AuthenticateUser(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Message: "Authentication failed",
			Error:   "Invalid credentials",
		})
	}

	// Remove password from response
	user.Password = ""

	// Generate JWT token with session
	sessionID := uuid.New()
	deviceID := c.Request().Header.Get("User-Agent") // Use User-Agent as device identifier
	if deviceID == "" {
		deviceID = "unknown-device"
	}

	jwtService := jwt.GetService()
	if jwtService == nil {
		logger.Error("JWT service not initialized")
		return c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Message: "Authentication service unavailable",
		})
	}

	accessToken, refreshToken, expiresAt, err := jwtService.GenerateTokens(user, sessionID, deviceID)
	if err != nil {
		logger.Error("Failed to generate JWT tokens", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Message: "Failed to generate authentication tokens",
		})
	}

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Login successful",
		Data: map[string]interface{}{
			"user":          user,
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"expires_at":    expiresAt,
			"session_id":    sessionID,
		},
	})
}

func (h *UserHandler) UpdateUser(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid user ID format",
			Error:   err.Error(),
		})
	}

	// Get existing user
	user, err := h.userService.GetUserByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Message: "User not found",
			Error:   err.Error(),
		})
	}

	// Bind updates (partial update)
	var updates map[string]interface{}
	if err := c.Bind(&updates); err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	// Apply updates (simplified - in real app, you'd want proper validation)
	if username, ok := updates["username"].(string); ok {
		user.Username = username
	}
	if email, ok := updates["email"].(string); ok {
		user.Email = email
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		user.IsActive = isActive
	}

	if err := h.userService.UpdateUser(c.Request().Context(), user); err != nil {
		logger.Error("Failed to update user", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Message: "Failed to update user",
			Error:   err.Error(),
		})
	}

	// Remove password from response
	user.Password = ""

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "User updated successfully",
		Data:    user,
	})
}

func (h *UserHandler) DeleteUser(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid user ID format",
			Error:   err.Error(),
		})
	}

	if err := h.userService.DeleteUser(c.Request().Context(), id); err != nil {
		logger.Error("Failed to delete user", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Message: "Failed to delete user",
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}
