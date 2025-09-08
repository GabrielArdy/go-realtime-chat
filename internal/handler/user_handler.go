package handler

import (
	"net/http"
	"strconv"

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

	// TODO: Generate JWT token here
	token := "fake-jwt-token-for-now"

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Login successful",
		Data: map[string]interface{}{
			"user":  user,
			"token": token,
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
