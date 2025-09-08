package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"realtime-api/internal/redis"

	"github.com/redis/rueidis"
)

// EventSubscriber handles subscribing to events from Redis
type EventSubscriber struct {
	redis *redis.Redis
}

// NewEventSubscriber creates a new event subscriber
func NewEventSubscriber(redis *redis.Redis) *EventSubscriber {
	return &EventSubscriber{
		redis: redis,
	}
}

// EventHandler is a function type for handling events
type EventHandler func(event *Event) error

// EventRouter routes events to appropriate handlers
type EventRouter struct {
	handlers map[string]EventHandler
}

// NewEventRouter creates a new event router
func NewEventRouter() *EventRouter {
	return &EventRouter{
		handlers: make(map[string]EventHandler),
	}
}

// Register registers an event handler for a specific event type
func (er *EventRouter) Register(eventType string, handler EventHandler) {
	er.handlers[eventType] = handler
}

// Route routes an event to the appropriate handler
func (er *EventRouter) Route(event *Event) error {
	if handler, exists := er.handlers[event.Type]; exists {
		return handler(event)
	}

	// Log unhandled events for debugging
	log.Printf("Unhandled event type: %s", event.Type)
	return nil
}

// SubscribeToChannel subscribes to a specific Redis channel
func (es *EventSubscriber) SubscribeToChannel(ctx context.Context, channel string, router *EventRouter) error {
	// Subscribe to channel using Redis Subscribe method
	client, err := es.redis.Subscribe(ctx, channel)
	if err != nil {
		return fmt.Errorf("failed to subscribe to channel %s: %w", channel, err)
	}
	defer client.Close()

	log.Printf("Subscribed to channel: %s", channel)

	// Start message processing loop
	for {
		select {
		case <-ctx.Done():
			log.Printf("Context cancelled, unsubscribing from channel: %s", channel)
			return ctx.Err()
		default:
			// Receive message
			err := client.Receive(ctx,
				client.B().Subscribe().Channel(channel).Build(),
				func(msg rueidis.PubSubMessage) {
					// Try to parse as JSON string first
					var eventData string
					if err := json.Unmarshal([]byte(msg.Message), &eventData); err == nil {
						// It was a JSON string, now parse the actual event
						var event Event
						if err := json.Unmarshal([]byte(eventData), &event); err != nil {
							log.Printf("Failed to unmarshal event from JSON string in channel %s: %v", channel, err)
							return
						}

						// Route event to handler
						if err := router.Route(&event); err != nil {
							log.Printf("Error handling event %s from channel %s: %v", event.Type, channel, err)
						}
						return
					}

					// Try to parse directly as Event object
					var event Event
					if err := json.Unmarshal([]byte(msg.Message), &event); err != nil {
						log.Printf("Failed to unmarshal event from channel %s: %v", channel, err)
						return
					}

					// Route event to handler
					if err := router.Route(&event); err != nil {
						log.Printf("Error handling event %s from channel %s: %v", event.Type, channel, err)
					}
				})

			if err != nil {
				log.Printf("Error receiving from channel %s: %v", channel, err)
				// Continue listening despite errors
				continue
			}
		}
	}
}

// SubscribeToRoom subscribes to room events
func (es *EventSubscriber) SubscribeToRoom(ctx context.Context, roomID string, router *EventRouter) error {
	return es.SubscribeToChannel(ctx, fmt.Sprintf("room:%s", roomID), router)
}

// SubscribeToUser subscribes to user events
func (es *EventSubscriber) SubscribeToUser(ctx context.Context, userID string, router *EventRouter) error {
	return es.SubscribeToChannel(ctx, fmt.Sprintf("user:%s", userID), router)
}

// SubscribeToPresence subscribes to presence events
func (es *EventSubscriber) SubscribeToPresence(ctx context.Context, router *EventRouter) error {
	return es.SubscribeToChannel(ctx, "presence", router)
}

// SubscribeToSystem subscribes to system events
func (es *EventSubscriber) SubscribeToSystem(ctx context.Context, router *EventRouter) error {
	return es.SubscribeToChannel(ctx, "system", router)
}

// SubscribeToGlobal subscribes to global events
func (es *EventSubscriber) SubscribeToGlobal(ctx context.Context, router *EventRouter) error {
	return es.SubscribeToChannel(ctx, "global", router)
}

// Example event handlers

// UserEventHandler handles user-related events
func UserEventHandler(event *Event) error {
	log.Printf("User Event: %s | User: %v | Data: %+v",
		event.Type, event.UserID, event.Data)

	switch event.Type {
	case UserOnline:
		// Handle user online
		log.Printf("User %v is now online", event.UserID)

	case UserOffline:
		// Handle user offline
		log.Printf("User %v is now offline", event.UserID)

	case UserTypingStart:
		// Handle typing start
		if roomID, ok := event.Data["room_id"]; ok {
			log.Printf("User %v started typing in room %v", event.UserID, roomID)
		}

	case UserTypingStop:
		// Handle typing stop
		if roomID, ok := event.Data["room_id"]; ok {
			log.Printf("User %v stopped typing in room %v", event.UserID, roomID)
		}
	}

	return nil
}

// RoomEventHandler handles room-related events
func RoomEventHandler(event *Event) error {
	log.Printf("Room Event: %s | Room: %v | User: %v | Data: %+v",
		event.Type, event.RoomID, event.UserID, event.Data)

	switch event.Type {
	case RoomCreate:
		// Handle room creation
		if roomName, ok := event.Data["room_name"]; ok {
			log.Printf("Room %v (%v) created by user %v", event.RoomID, roomName, event.UserID)
		}

	case RoomJoin:
		// Handle user joining room
		log.Printf("User %v joined room %v", event.UserID, event.RoomID)

	case RoomLeave:
		// Handle user leaving room
		log.Printf("User %v left room %v", event.UserID, event.RoomID)

	case RoomMemberAdd:
		// Handle member addition
		if inviterID, ok := event.Data["inviter_id"]; ok {
			log.Printf("User %v added to room %v by %v", event.UserID, event.RoomID, inviterID)
		}

	case RoomMemberRemove:
		// Handle member removal
		if removerID, ok := event.Data["remover_id"]; ok {
			log.Printf("User %v removed from room %v by %v", event.UserID, event.RoomID, removerID)
		}
	}

	return nil
}

// MessageEventHandler handles message-related events
func MessageEventHandler(event *Event) error {
	log.Printf("Message Event: %s | Room: %v | User: %v | Data: %+v",
		event.Type, event.RoomID, event.UserID, event.Data)

	switch event.Type {
	case MessageSend:
		// Handle new message
		if content, ok := event.Data["content"]; ok {
			log.Printf("New message in room %v from user %v: %v",
				event.RoomID, event.UserID, content)
		}

	case MessageEdit:
		// Handle message edit
		if messageID, ok := event.Data["message_id"]; ok {
			log.Printf("Message %v edited by user %v in room %v",
				messageID, event.UserID, event.RoomID)
		}

	case MessageDelete:
		// Handle message deletion
		if messageID, ok := event.Data["message_id"]; ok {
			log.Printf("Message %v deleted by user %v in room %v",
				messageID, event.UserID, event.RoomID)
		}

	case MessageRead:
		// Handle message read
		if messageID, ok := event.Data["message_id"]; ok {
			log.Printf("Message %v read by user %v in room %v",
				messageID, event.UserID, event.RoomID)
		}

	case MessageReactionAdd:
		// Handle reaction addition
		if emoji, ok := event.Data["emoji"]; ok {
			if messageID, ok := event.Data["message_id"]; ok {
				log.Printf("User %v reacted with %v to message %v in room %v",
					event.UserID, emoji, messageID, event.RoomID)
			}
		}
	}

	return nil
}

// SystemEventHandler handles system-related events
func SystemEventHandler(event *Event) error {
	log.Printf("System Event: %s | Data: %+v", event.Type, event.Data)

	switch event.Type {
	case SystemMaintenance:
		// Handle maintenance event
		log.Printf("System maintenance scheduled")

	case SystemShutdown:
		// Handle shutdown event
		log.Printf("System shutdown initiated")

	case SystemBroadcast:
		// Handle broadcast event
		if message, ok := event.Data["message"]; ok {
			log.Printf("System broadcast: %v", message)
		}
	}

	return nil
}
