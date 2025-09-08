package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"realtime-api/internal/config"
	"realtime-api/internal/database"
	"realtime-api/internal/handler"
	"realtime-api/internal/health"
	"realtime-api/internal/jwt"
	"realtime-api/internal/logger"
	"realtime-api/internal/middleware"
	"realtime-api/internal/model"
	"realtime-api/internal/rabbitmq"
	"realtime-api/internal/redis"
	"realtime-api/internal/repository"
	"realtime-api/internal/service"
	"realtime-api/internal/websocket"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.Logger.Level, cfg.Logger.Format, cfg.Logger.Output, cfg.Logger.TimeFormat)
	logger.SetupStandardLogger()

	logger.Info("Starting Realtime API Server", logger.WithFields(map[string]interface{}{
		"version":     "1.0.0",
		"environment": cfg.Server.Environment,
		"port":        cfg.Server.Port,
	}))

	// Initialize database
	db, err := database.Init(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to initialize database", logger.WithField("error", err.Error()))
	}
	defer db.Close()

	// Run database migrations
	if err := db.Migrate(
		&model.User{},
		&model.UserProfile{},
		&model.UserContact{},
		&model.UserSession{},
		&model.Room{},
		&model.RoomMember{},
		&model.RoomInvite{},
		&model.Message{},
		&model.MessageAttachment{},
		&model.MessageReaction{},
		&model.MessageRead{},
		&model.MessageDraft{},
		&model.Notification{},
		&model.FileUpload{},
	); err != nil {
		logger.Fatal("Failed to run database migrations", logger.WithField("error", err.Error()))
	}

	// Initialize Redis
	redisClient, err := redis.Init(&cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to initialize Redis", logger.WithField("error", err.Error()))
	}
	defer redisClient.Close()

	// Initialize RabbitMQ
	rabbitClient, err := rabbitmq.Init(&cfg.RabbitMQ)
	if err != nil {
		logger.Fatal("Failed to initialize RabbitMQ", logger.WithField("error", err.Error()))
	}
	defer rabbitClient.Close()

	// Initialize JWT service
	jwt.Init(&cfg.JWT)

	// Initialize WebSocket hub
	websocket.Init()

	// Initialize health checker
	health.Init()

	// Initialize repositories
	userRepo := repository.NewUserRepository()
	roomRepo := repository.NewRoomRepository()
	messageRepo := repository.NewMessageRepository()

	// Initialize services
	userService := service.NewUserService(userRepo)
	// TODO: Implement RoomService and MessageService
	_ = roomRepo    // Placeholder to avoid unused variable error
	_ = messageRepo // Placeholder to avoid unused variable error

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService)
	// TODO: Implement RoomHandler and MessageHandler

	// Initialize Echo server
	e := echo.New()

	// Hide banner
	e.HideBanner = true

	// Configure timeouts
	e.Server.ReadTimeout = time.Duration(cfg.Server.ReadTimeout) * time.Second
	e.Server.WriteTimeout = time.Duration(cfg.Server.WriteTimeout) * time.Second

	// Global middleware
	e.Use(middleware.RecoveryMiddleware())
	e.Use(middleware.LoggerMiddleware())
	e.Use(middleware.CORSMiddleware())
	e.Use(middleware.RequestIDMiddleware())
	e.Use(echoMiddleware.Secure())
	e.Use(echoMiddleware.Gzip())

	// Rate limiting (100 requests per minute)
	e.Use(middleware.RateLimitMiddleware(100))

	// Health check routes
	e.GET("/health", echo.WrapHandler(http.HandlerFunc(health.HealthHandler)))
	e.GET("/health/ready", echo.WrapHandler(http.HandlerFunc(health.ReadinessHandler)))
	e.GET("/health/live", echo.WrapHandler(http.HandlerFunc(health.LivenessHandler)))

	// API routes
	api := e.Group("/api/v1")

	// User routes
	users := api.Group("/users")
	users.POST("", userHandler.CreateUser)
	users.GET("", userHandler.ListUsers)
	users.GET("/:id", userHandler.GetUser)
	users.PUT("/:id", userHandler.UpdateUser)
	users.DELETE("/:id", userHandler.DeleteUser)

	// Auth routes
	auth := api.Group("/auth")
	auth.POST("/login", userHandler.LoginUser)

	// WebSocket route
	e.GET("/ws", websocket.HandleWebSocket)

	// Root route
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":     "Realtime API Server",
			"version":     "1.0.0",
			"environment": cfg.Server.Environment,
			"timestamp":   time.Now(),
		})
	})

	// Start server in a goroutine
	go func() {
		address := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
		logger.Info("Server starting", logger.WithField("address", address))

		if err := e.Start(address); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", logger.WithField("error", err.Error()))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Server shutting down...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", logger.WithField("error", err.Error()))
	}

	logger.Info("Server shutdown complete")
}
