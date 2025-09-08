package handler

import (
	"fmt"
	"net/http"

	"realtime-api/internal/model"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// GetUserIDFromContext extracts the user ID from the JWT token stored in the Echo context
func GetUserIDFromContext(c echo.Context) (uuid.UUID, error) {
	userID := c.Get("user_id")
	if userID == nil {
		return uuid.Nil, fmt.Errorf("user not authenticated")
	}

	switch v := userID.(type) {
	case uuid.UUID:
		return v, nil
	case string:
		return uuid.Parse(v)
	default:
		return uuid.Nil, fmt.Errorf("invalid user ID format")
	}
}

// RequireAuth is a helper that extracts user ID and returns error response if not authenticated
func RequireAuth(c echo.Context) (uuid.UUID, *echo.HTTPError) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Message: "Authentication required",
			Error:   err.Error(),
		})
	}
	return userID, nil
}

// GetUsernameFromContext extracts the username from the JWT token stored in the Echo context
func GetUsernameFromContext(c echo.Context) (string, error) {
	username := c.Get("username")
	if username == nil {
		return "", fmt.Errorf("username not found in context")
	}

	if str, ok := username.(string); ok {
		return str, nil
	}

	return "", fmt.Errorf("invalid username format")
}

// GetDeviceIDFromContext extracts the device ID from the JWT token stored in the Echo context
func GetDeviceIDFromContext(c echo.Context) (string, error) {
	deviceID := c.Get("device_id")
	if deviceID == nil {
		return "", fmt.Errorf("device ID not found in context")
	}

	if str, ok := deviceID.(string); ok {
		return str, nil
	}

	return "", fmt.Errorf("invalid device ID format")
}
