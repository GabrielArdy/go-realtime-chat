package handler

import (
	"fmt"
	"net/http"
	"strings"

	"realtime-api/internal/jwt"
	"realtime-api/internal/model"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// extractTokenFromHeader extracts JWT token from Authorization header
func extractTokenFromHeader(c echo.Context) (string, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is required")
	}

	// Extract token from "Bearer <token>" format
	if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
		return authHeader[7:], nil
	}

	// Try to get token directly if not in Bearer format
	return authHeader, nil
}

// validateTokenAndGetClaims validates JWT token and returns claims
func validateTokenAndGetClaims(tokenString string) (*jwt.Claims, error) {
	jwtService := jwt.GetService()
	if jwtService == nil {
		return nil, fmt.Errorf("JWT service not initialized")
	}

	claims, err := jwtService.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return claims, nil
}

// GetUserIDFromContext extracts the user ID from the JWT token in Authorization header
func GetUserIDFromContext(c echo.Context) (uuid.UUID, error) {
	token, err := extractTokenFromHeader(c)
	if err != nil {
		return uuid.Nil, err
	}

	claims, err := validateTokenAndGetClaims(token)
	if err != nil {
		return uuid.Nil, err
	}

	return claims.UserID, nil
}

// RequireAuth is a helper that extracts user ID from Authorization header and returns error response if not authenticated
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

// GetUsernameFromContext extracts the username from the JWT token in Authorization header
func GetUsernameFromContext(c echo.Context) (string, error) {
	token, err := extractTokenFromHeader(c)
	if err != nil {
		return "", err
	}

	claims, err := validateTokenAndGetClaims(token)
	if err != nil {
		return "", err
	}

	return claims.Username, nil
}

// GetDeviceIDFromContext extracts the device ID from the JWT token in Authorization header
func GetDeviceIDFromContext(c echo.Context) (string, error) {
	token, err := extractTokenFromHeader(c)
	if err != nil {
		return "", err
	}

	claims, err := validateTokenAndGetClaims(token)
	if err != nil {
		return "", err
	}

	return claims.DeviceID, nil
}

// GetSessionIDFromContext extracts the session ID from the JWT token in Authorization header
func GetSessionIDFromContext(c echo.Context) (uuid.UUID, error) {
	token, err := extractTokenFromHeader(c)
	if err != nil {
		return uuid.Nil, err
	}

	claims, err := validateTokenAndGetClaims(token)
	if err != nil {
		return uuid.Nil, err
	}

	return claims.SessionID, nil
}

// GetEmailFromContext extracts the email from the JWT token in Authorization header
func GetEmailFromContext(c echo.Context) (string, error) {
	token, err := extractTokenFromHeader(c)
	if err != nil {
		return "", err
	}

	claims, err := validateTokenAndGetClaims(token)
	if err != nil {
		return "", err
	}

	return claims.Email, nil
}

// GetAllClaimsFromContext extracts all claims from the JWT token in Authorization header
func GetAllClaimsFromContext(c echo.Context) (*jwt.Claims, error) {
	token, err := extractTokenFromHeader(c)
	if err != nil {
		return nil, err
	}

	claims, err := validateTokenAndGetClaims(token)
	if err != nil {
		return nil, err
	}

	return claims, nil
}
