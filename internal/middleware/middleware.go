package middleware

import (
	"time"

	"realtime-api/internal/logger"

	"github.com/labstack/echo/v4"
)

// LoggerMiddleware logs HTTP requests
func LoggerMiddleware() echo.MiddlewareFunc {
	return echo.MiddlewareFunc(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			duration := time.Since(start)

			req := c.Request()
			res := c.Response()

			fields := map[string]interface{}{
				"method":     req.Method,
				"path":       req.URL.Path,
				"status":     res.Status,
				"duration":   duration.String(),
				"ip":         c.RealIP(),
				"user_agent": req.UserAgent(),
				"bytes_in":   req.ContentLength,
				"bytes_out":  res.Size,
			}

			// Add query parameters if they exist
			if req.URL.RawQuery != "" {
				fields["query"] = req.URL.RawQuery
			}

			// Add error if exists
			if err != nil {
				fields["error"] = err.Error()
				logger.Error("HTTP request failed", fields)
			} else {
				logger.Info("HTTP request completed", fields)
			}

			return err
		}
	})
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware() echo.MiddlewareFunc {
	return echo.MiddlewareFunc(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Panic recovered", logger.WithFields(map[string]interface{}{
						"panic":  r,
						"method": c.Request().Method,
						"path":   c.Request().URL.Path,
						"ip":     c.RealIP(),
					}))

					c.JSON(500, map[string]interface{}{
						"success": false,
						"message": "Internal server error",
						"error":   "An unexpected error occurred",
					})
				}
			}()

			return next(c)
		}
	})
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() echo.MiddlewareFunc {
	return echo.MiddlewareFunc(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			origin := c.Request().Header.Get("Origin")

			// Set CORS headers
			c.Response().Header().Set("Access-Control-Allow-Origin", "*")
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			c.Response().Header().Set("Access-Control-Expose-Headers", "Content-Length")
			c.Response().Header().Set("Access-Control-Allow-Credentials", "true")

			// Handle preflight requests
			if c.Request().Method == "OPTIONS" {
				return c.NoContent(200)
			}

			logger.Debug("CORS middleware applied", logger.WithField("origin", origin))

			return next(c)
		}
	})
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() echo.MiddlewareFunc {
	return echo.MiddlewareFunc(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := c.Request().Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}

			c.Response().Header().Set("X-Request-ID", requestID)
			c.Set("request_id", requestID)

			return next(c)
		}
	})
}

// RateLimitMiddleware implements basic rate limiting
func RateLimitMiddleware(requestsPerMinute int) echo.MiddlewareFunc {
	// This is a simple in-memory rate limiter
	// For production, consider using Redis-based rate limiting

	type client struct {
		requests []time.Time
	}

	clients := make(map[string]*client)

	return echo.MiddlewareFunc(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			now := time.Now()

			// Clean up old clients periodically
			if len(clients) > 1000 {
				for k, v := range clients {
					if len(v.requests) == 0 || now.Sub(v.requests[len(v.requests)-1]) > time.Hour {
						delete(clients, k)
					}
				}
			}

			// Get or create client
			if clients[ip] == nil {
				clients[ip] = &client{requests: make([]time.Time, 0)}
			}

			client := clients[ip]

			// Remove requests older than 1 minute
			cutoff := now.Add(-time.Minute)
			newRequests := make([]time.Time, 0)
			for _, reqTime := range client.requests {
				if reqTime.After(cutoff) {
					newRequests = append(newRequests, reqTime)
				}
			}
			client.requests = newRequests

			// Check rate limit
			if len(client.requests) >= requestsPerMinute {
				logger.Warn("Rate limit exceeded", logger.WithFields(map[string]interface{}{
					"ip":       ip,
					"requests": len(client.requests),
					"limit":    requestsPerMinute,
				}))

				return c.JSON(429, map[string]interface{}{
					"success": false,
					"message": "Rate limit exceeded",
					"error":   "Too many requests",
				})
			}

			// Add current request
			client.requests = append(client.requests, now)

			return next(c)
		}
	})
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	// Simple implementation - for production consider using uuid
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
