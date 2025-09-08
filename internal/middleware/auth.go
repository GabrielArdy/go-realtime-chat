package middleware

import (
	"net/http"
	"strings"

	"realtime-api/internal/jwt"
	"realtime-api/internal/logger"
	"realtime-api/internal/model"

	"github.com/labstack/echo/v4"
)

func JWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, model.APIResponse{
					Success: false,
					Message: "Missing authorization header",
				})
			}

			// Check if header starts with "Bearer "
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, model.APIResponse{
					Success: false,
					Message: "Invalid authorization header format",
				})
			}

			// Extract token
			token := authHeader[7:] // Remove "Bearer " prefix
			if token == "" {
				return c.JSON(http.StatusUnauthorized, model.APIResponse{
					Success: false,
					Message: "Missing token",
				})
			}

			// Validate token
			claims, err := jwt.GetService().ValidateToken(token)
			if err != nil {
				logger.Warn("Invalid JWT token", logger.WithFields(map[string]interface{}{
					"error": err.Error(),
					"ip":    c.RealIP(),
				}))
				return c.JSON(http.StatusUnauthorized, model.APIResponse{
					Success: false,
					Message: "Invalid token",
				})
			}

			// Set user context
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("device_id", claims.DeviceID)
			c.Set("claims", claims)

			logger.Debug("User authenticated", logger.WithFields(map[string]interface{}{
				"user_id":  claims.UserID,
				"username": claims.Username,
			}))

			return next(c)
		}
	}
}

func OptionalJWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				token := authHeader[7:] // Remove "Bearer " prefix
				if token != "" {
					// Validate token
					claims, err := jwt.GetService().ValidateToken(token)
					if err == nil {
						// Set user context if token is valid
						c.Set("user_id", claims.UserID)
						c.Set("username", claims.Username)
						c.Set("device_id", claims.DeviceID)
						c.Set("claims", claims)
					}
				}
			}

			return next(c)
		}
	}
}
