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
	"realtime-api/internal/events"
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

	"github.com/google/uuid"
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

	// ===== Initialize Event System =====
	logger.Info("Initializing event system...")

	// Initialize Event Subscriber
	eventSubscriber := events.NewEventSubscriber(redisClient)

	// Initialize Event Router
	eventRouter := events.NewEventRouter()

	// Initialize WebSocket hub
	websocket.Init(redisClient)
	websocketHub := websocket.GetHub()

	// Setup event handlers for real-time functionality
	setupEventHandlers(eventRouter, websocketHub)

	// Start event processing in background
	eventCtx, eventCancel := context.WithCancel(context.Background())
	defer eventCancel()

	go func() {
		logger.Info("Starting event subscriber for real-time processing...")
		// Subscribe to different event channels
		go func() {
			if err := eventSubscriber.SubscribeToGlobal(eventCtx, eventRouter); err != nil {
				logger.Error("Failed to subscribe to global events", logger.WithField("error", err.Error()))
			}
		}()

		go func() {
			if err := eventSubscriber.SubscribeToSystem(eventCtx, eventRouter); err != nil {
				logger.Error("Failed to subscribe to system events", logger.WithField("error", err.Error()))
			}
		}()

		go func() {
			if err := eventSubscriber.SubscribeToPresence(eventCtx, eventRouter); err != nil {
				logger.Error("Failed to subscribe to presence events", logger.WithField("error", err.Error()))
			}
		}()
	}()

	// Initialize health checker
	health.Init()

	// Initialize repositories
	userRepo := repository.NewUserRepository()
	roomRepo := repository.NewRoomRepository()
	messageRepo := repository.NewMessageRepository()

	// Initialize services
	userService := service.NewUserService(userRepo)
	roomService := service.NewRoomService(roomRepo, userRepo, redisClient)
	messageService := service.NewMessageService(messageRepo, roomRepo, userRepo, redisClient)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService)
	roomHandler := handler.NewRoomHandler(roomService)
	messageHandler := handler.NewMessageHandler(messageService)
	eventHandler := handler.NewEventHandler(redisClient)

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

	// Room routes
	rooms := api.Group("/rooms")
	rooms.POST("", roomHandler.CreateRoom)
	rooms.GET("", roomHandler.ListRooms)
	rooms.GET("/my-chats", roomHandler.ListUserChatRooms) // New endpoint for chat list
	rooms.GET("/:id", roomHandler.GetRoom)
	rooms.PUT("/:id", roomHandler.UpdateRoom)
	rooms.DELETE("/:id", roomHandler.DeleteRoom)
	rooms.POST("/:id/join", roomHandler.JoinRoom)
	rooms.POST("/:id/leave", roomHandler.LeaveRoom)
	rooms.GET("/:id/members", roomHandler.GetRoomMembers)
	rooms.POST("/:id/members", roomHandler.AddMember)
	rooms.DELETE("/:id/members/:user_id", roomHandler.RemoveMember)
	rooms.POST("/:id/invites", roomHandler.CreateInvite)
	rooms.POST("/invites/:invite_code/accept", roomHandler.AcceptInvite)
	rooms.POST("/invites/:invite_code/reject", roomHandler.RejectInvite)

	// Direct room routes
	rooms.POST("/direct/:user_id", roomHandler.CreateOrGetDirectRoom) // New endpoint for direct messages

	// Message routes
	messages := api.Group("/messages")
	messages.POST("", messageHandler.SendMessage)
	messages.GET("/:id", messageHandler.GetMessage)
	messages.PUT("/:id", messageHandler.EditMessage)
	messages.DELETE("/:id", messageHandler.DeleteMessage)
	messages.POST("/:id/reactions", messageHandler.ReactToMessage)
	messages.DELETE("/:id/reactions", messageHandler.RemoveReaction)
	messages.POST("/:id/read", messageHandler.MarkAsRead)

	// Room-specific message routes
	rooms.GET("/:room_id/messages", messageHandler.GetRoomMessages)
	rooms.POST("/:room_id/typing/start", messageHandler.StartTyping)
	rooms.POST("/:room_id/typing/stop", messageHandler.StopTyping)

	// Event system routes (for monitoring/debugging)
	events := api.Group("/events")
	events.GET("/metrics", eventHandler.GetEventMetrics)
	events.POST("/system", eventHandler.PublishSystemEvent)
	events.GET("/history", eventHandler.GetEventHistory)

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
		logger.Info("Server starting", logger.WithFields(map[string]interface{}{
			"address":      address,
			"event_system": "enabled",
		}))

		if err := e.Start(address); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", logger.WithField("error", err.Error()))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Server shutting down...")

	// Cancel event processing
	eventCancel()

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", logger.WithField("error", err.Error()))
	}

	logger.Info("Server shutdown complete")
}

// setupEventHandlers configures event routing to WebSocket for real-time functionality
func setupEventHandlers(router *events.EventRouter, hub *websocket.Hub) {
	logger.Info("Setting up event handlers for real-time functionality...")

	// User events - Online/Offline status
	router.Register("event.user.online", func(event *events.Event) error {
		logger.Debug("User online event", logger.WithFields(map[string]interface{}{
			"user_id": event.UserID,
		}))

		if event.UserID != nil {
			hub.BroadcastToUser(*event.UserID, model.WSTypeUserStatusChange, map[string]interface{}{
				"status":  "online",
				"user_id": *event.UserID,
				"data":    event.Data,
			})
		}
		return nil
	})

	router.Register("event.user.offline", func(event *events.Event) error {
		logger.Debug("User offline event", logger.WithFields(map[string]interface{}{
			"user_id": event.UserID,
		}))

		if event.UserID != nil {
			hub.BroadcastToUser(*event.UserID, model.WSTypeUserStatusChange, map[string]interface{}{
				"status":  "offline",
				"user_id": *event.UserID,
				"data":    event.Data,
			})
		}
		return nil
	})

	// Typing events - Real-time typing indicators
	router.Register("event.user.typing.start", func(event *events.Event) error {
		if roomIDStr, ok := event.Data["room_id"].(string); ok {
			if roomID, err := uuid.Parse(roomIDStr); err == nil {
				hub.BroadcastToRoom(roomID, model.WSTypeTypingStart, map[string]interface{}{
					"user_id": event.UserID,
					"room_id": roomID,
				})
			}
		}
		return nil
	})

	router.Register("event.user.typing.stop", func(event *events.Event) error {
		if roomIDStr, ok := event.Data["room_id"].(string); ok {
			if roomID, err := uuid.Parse(roomIDStr); err == nil {
				hub.BroadcastToRoom(roomID, model.WSTypeTypingStop, map[string]interface{}{
					"user_id": event.UserID,
					"room_id": roomID,
				})
			}
		}
		return nil
	})

	// Room events - Join/Leave/Create real-time notifications
	router.Register("event.room.create", func(event *events.Event) error {
		if event.RoomID != nil {
			hub.BroadcastToRoom(*event.RoomID, model.WSTypeNotification, map[string]interface{}{
				"type":    "room_created",
				"room_id": *event.RoomID,
				"user_id": event.UserID,
				"data":    event.Data,
			})
		}
		return nil
	})

	router.Register("event.room.join", func(event *events.Event) error {
		if event.RoomID != nil {
			hub.BroadcastToRoom(*event.RoomID, model.WSTypeUserJoin, map[string]interface{}{
				"room_id": *event.RoomID,
				"user_id": event.UserID,
				"data":    event.Data,
			})
		}
		return nil
	})

	router.Register("event.room.leave", func(event *events.Event) error {
		if event.RoomID != nil {
			hub.BroadcastToRoom(*event.RoomID, model.WSTypeUserLeave, map[string]interface{}{
				"room_id": *event.RoomID,
				"user_id": event.UserID,
				"data":    event.Data,
			})
		}
		return nil
	})

	router.Register("event.room.member.add", func(event *events.Event) error {
		if event.RoomID != nil {
			hub.BroadcastToRoom(*event.RoomID, model.WSTypeNotification, map[string]interface{}{
				"type":    "member_added",
				"room_id": *event.RoomID,
				"data":    event.Data,
			})
		}
		return nil
	})

	router.Register("event.room.member.remove", func(event *events.Event) error {
		if event.RoomID != nil {
			hub.BroadcastToRoom(*event.RoomID, model.WSTypeNotification, map[string]interface{}{
				"type":    "member_removed",
				"room_id": *event.RoomID,
				"data":    event.Data,
			})
		}
		return nil
	})

	// Message events - Real-time message delivery
	router.Register("event.message.send", func(event *events.Event) error {
		if event.RoomID != nil {
			hub.BroadcastToRoom(*event.RoomID, model.WSTypeMessage, event.Data)
		}
		return nil
	})

	router.Register("event.message.edit", func(event *events.Event) error {
		if event.RoomID != nil {
			hub.BroadcastToRoom(*event.RoomID, model.WSTypeMessageEdit, event.Data)
		}
		return nil
	})

	router.Register("event.message.delete", func(event *events.Event) error {
		if event.RoomID != nil {
			hub.BroadcastToRoom(*event.RoomID, model.WSTypeMessageDelete, event.Data)
		}
		return nil
	})

	router.Register("event.message.read", func(event *events.Event) error {
		if event.RoomID != nil {
			hub.BroadcastToRoom(*event.RoomID, model.WSTypeNotification, map[string]interface{}{
				"type":    "message_read",
				"room_id": *event.RoomID,
				"user_id": event.UserID,
				"data":    event.Data,
			})
		}
		return nil
	})

	router.Register("event.message.reaction.add", func(event *events.Event) error {
		if event.RoomID != nil {
			hub.BroadcastToRoom(*event.RoomID, model.WSTypeMessageReaction, map[string]interface{}{
				"action":  "add",
				"room_id": *event.RoomID,
				"user_id": event.UserID,
				"data":    event.Data,
			})
		}
		return nil
	})

	router.Register("event.message.reaction.remove", func(event *events.Event) error {
		if event.RoomID != nil {
			hub.BroadcastToRoom(*event.RoomID, model.WSTypeMessageReaction, map[string]interface{}{
				"action":  "remove",
				"room_id": *event.RoomID,
				"user_id": event.UserID,
				"data":    event.Data,
			})
		}
		return nil
	})

	// System events - Global notifications (broadcast to all connected users)
	router.Register("event.system.maintenance", func(event *events.Event) error {
		// Since there's no BroadcastGlobal, we'll use notification type
		logger.Info("System maintenance event", logger.WithField("data", event.Data))
		return nil
	})

	router.Register("event.system.shutdown", func(event *events.Event) error {
		logger.Info("System shutdown event", logger.WithField("data", event.Data))
		return nil
	})

	router.Register("event.system.announcement", func(event *events.Event) error {
		logger.Info("System announcement event", logger.WithField("data", event.Data))
		return nil
	})

	logger.Info("Event handlers registered successfully", logger.WithFields(map[string]interface{}{
		"handlers_count": "16",
		"categories":     []string{"user", "typing", "room", "message", "system"},
	}))
}
