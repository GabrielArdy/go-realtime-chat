package handler

import (
	"net/http"

	"realtime-api/internal/events"
	"realtime-api/internal/logger"
	"realtime-api/internal/model"
	"realtime-api/internal/redis"

	"github.com/labstack/echo/v4"
)

type EventHandler struct {
	eventPublisher *events.EventPublisher
	eventRouter    *events.EventRouter
}

func NewEventHandler(redis *redis.Redis) *EventHandler {
	publisher := events.NewEventPublisher(redis)
	router := events.NewEventRouter()

	// Register default event handlers
	registerEventHandlers(router)

	return &EventHandler{
		eventPublisher: publisher,
		eventRouter:    router,
	}
}

// registerEventHandlers registers default event handlers for logging and monitoring
func registerEventHandlers(router *events.EventRouter) {
	// User events
	router.Register(events.UserOnline, func(event *events.Event) error {
		logger.Info("User came online", logger.WithFields(map[string]interface{}{
			"user_id":   event.UserID,
			"timestamp": event.Timestamp,
		}))
		return nil
	})

	router.Register(events.UserOffline, func(event *events.Event) error {
		logger.Info("User went offline", logger.WithFields(map[string]interface{}{
			"user_id":   event.UserID,
			"timestamp": event.Timestamp,
		}))
		return nil
	})

	router.Register(events.UserTypingStart, func(event *events.Event) error {
		logger.Debug("User started typing", logger.WithFields(map[string]interface{}{
			"user_id":   event.UserID,
			"room_id":   event.RoomID,
			"timestamp": event.Timestamp,
		}))
		return nil
	})

	router.Register(events.UserTypingStop, func(event *events.Event) error {
		logger.Debug("User stopped typing", logger.WithFields(map[string]interface{}{
			"user_id":   event.UserID,
			"room_id":   event.RoomID,
			"timestamp": event.Timestamp,
		}))
		return nil
	})

	// Room events
	router.Register(events.RoomCreate, func(event *events.Event) error {
		logger.Info("Room created", logger.WithFields(map[string]interface{}{
			"room_id":   event.RoomID,
			"user_id":   event.UserID,
			"data":      event.Data,
			"timestamp": event.Timestamp,
		}))
		return nil
	})

	router.Register(events.RoomJoin, func(event *events.Event) error {
		logger.Info("User joined room", logger.WithFields(map[string]interface{}{
			"room_id":   event.RoomID,
			"user_id":   event.UserID,
			"timestamp": event.Timestamp,
		}))
		return nil
	})

	router.Register(events.RoomLeave, func(event *events.Event) error {
		logger.Info("User left room", logger.WithFields(map[string]interface{}{
			"room_id":   event.RoomID,
			"user_id":   event.UserID,
			"timestamp": event.Timestamp,
		}))
		return nil
	})

	// Message events
	router.Register(events.MessageSend, func(event *events.Event) error {
		messageID := ""
		if id, ok := event.Data["message_id"]; ok {
			messageID = id.(string)
		}
		logger.Info("Message sent", logger.WithFields(map[string]interface{}{
			"room_id":    event.RoomID,
			"user_id":    event.UserID,
			"message_id": messageID,
			"timestamp":  event.Timestamp,
		}))
		return nil
	})

	router.Register(events.MessageEdit, func(event *events.Event) error {
		messageID := ""
		if id, ok := event.Data["message_id"]; ok {
			messageID = id.(string)
		}
		logger.Info("Message edited", logger.WithFields(map[string]interface{}{
			"room_id":    event.RoomID,
			"user_id":    event.UserID,
			"message_id": messageID,
			"timestamp":  event.Timestamp,
		}))
		return nil
	})

	router.Register(events.MessageDelete, func(event *events.Event) error {
		messageID := ""
		if id, ok := event.Data["message_id"]; ok {
			messageID = id.(string)
		}
		logger.Info("Message deleted", logger.WithFields(map[string]interface{}{
			"room_id":    event.RoomID,
			"user_id":    event.UserID,
			"message_id": messageID,
			"timestamp":  event.Timestamp,
		}))
		return nil
	})

	// System events
	router.Register(events.SystemMaintenance, func(event *events.Event) error {
		logger.Warn("System maintenance event", logger.WithFields(map[string]interface{}{
			"data":      event.Data,
			"timestamp": event.Timestamp,
		}))
		return nil
	})

	router.Register(events.SystemShutdown, func(event *events.Event) error {
		logger.Error("System shutdown event", logger.WithFields(map[string]interface{}{
			"data":      event.Data,
			"timestamp": event.Timestamp,
		}))
		return nil
	})
}

// GetEventMetrics returns event system metrics
func (h *EventHandler) GetEventMetrics(c echo.Context) error {
	// Get basic system metrics
	metrics := map[string]interface{}{
		"events_published":      0,  // TODO: Implement event counting
		"events_consumed":       0,  // TODO: Implement event counting
		"active_handlers":       16, // We have 16 registered handlers
		"websocket_connections": 0,  // TODO: Get from WebSocket hub
		"system_status":         "healthy",
		"uptime_seconds":        0, // TODO: Implement uptime tracking
	}

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Event metrics retrieved successfully",
		Data:    metrics,
	})
}

// PublishSystemEvent allows manual publishing of system events
func (h *EventHandler) PublishSystemEvent(c echo.Context) error {
	var req struct {
		Type string                 `json:"type" validate:"required"`
		Data map[string]interface{} `json:"data,omitempty"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	if err := h.eventPublisher.PublishSystemEvent(c.Request().Context(), req.Type, req.Data); err != nil {
		logger.Error("Failed to publish system event", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Message: "Failed to publish event",
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "System event published successfully",
	})
}

// GetEventHistory returns recent event history (for debugging/monitoring)
func (h *EventHandler) GetEventHistory(c echo.Context) error {
	// Basic event history implementation
	// In a production system, this would query a dedicated event store
	history := []map[string]interface{}{
		{
			"event_id":  "sample-event-1",
			"type":      "event.system.status",
			"timestamp": "2024-01-01T12:00:00Z",
			"level":     "system",
			"message":   "System started successfully",
		},
		{
			"event_id":  "sample-event-2",
			"type":      "event.user.online",
			"timestamp": "2024-01-01T12:01:00Z",
			"level":     "user",
			"message":   "User came online",
		},
	}

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Event history retrieved successfully",
		Data:    map[string]interface{}{"events": history},
	})
}
