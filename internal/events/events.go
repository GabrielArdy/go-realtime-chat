package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"realtime-api/internal/redis"

	"github.com/google/uuid"
)

// Event levels
const (
	LevelUser    = "user"
	LevelRoom    = "room"
	LevelMessage = "message"
	LevelSystem  = "system"
)

// User events
const (
	UserOnline        = "event.user.online"
	UserOffline       = "event.user.offline"
	UserTypingStart   = "event.user.typing.start"
	UserTypingStop    = "event.user.typing.stop"
	UserStatusChange  = "event.user.status.change"
	UserProfileUpdate = "event.user.profile.update"
)

// Room events
const (
	RoomCreate           = "event.room.create"
	RoomUpdate           = "event.room.update"
	RoomDelete           = "event.room.delete"
	RoomJoin             = "event.room.join"
	RoomLeave            = "event.room.leave"
	RoomMemberAdd        = "event.room.member.add"
	RoomMemberRemove     = "event.room.member.remove"
	RoomMemberRoleUpdate = "event.room.member.role.update"
	RoomInviteCreate     = "event.room.invite.create"
	RoomInviteAccept     = "event.room.invite.accept"
	RoomInviteReject     = "event.room.invite.reject"
)

// Message events
const (
	MessageSend           = "event.message.send"
	MessageEdit           = "event.message.edit"
	MessageDelete         = "event.message.delete"
	MessageRead           = "event.message.read"
	MessageReactionAdd    = "event.message.reaction.add"
	MessageReactionRemove = "event.message.reaction.remove"
)

// System events
const (
	SystemMaintenance = "event.system.maintenance"
	SystemShutdown    = "event.system.shutdown"
	SystemBroadcast   = "event.system.broadcast"
)

// Event represents a structured event with metadata
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Level     string                 `json:"level"`
	Action    string                 `json:"action"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    *uuid.UUID             `json:"user_id,omitempty"`
	RoomID    *uuid.UUID             `json:"room_id,omitempty"`
}

// EventPublisher handles publishing events to Redis
type EventPublisher struct {
	redis *redis.Redis
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(redis *redis.Redis) *EventPublisher {
	return &EventPublisher{
		redis: redis,
	}
}

// PublishUserEvent publishes user-related events
func (ep *EventPublisher) PublishUserEvent(ctx context.Context, eventType string, userID uuid.UUID, data map[string]interface{}) error {
	event := &Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Level:     LevelUser,
		Action:    extractAction(eventType),
		Data:      data,
		Timestamp: time.Now(),
		UserID:    &userID,
	}

	return ep.publishEvent(ctx, fmt.Sprintf("user:%s", userID.String()), event)
}

// PublishRoomEvent publishes room-related events
func (ep *EventPublisher) PublishRoomEvent(ctx context.Context, eventType string, roomID uuid.UUID, data map[string]interface{}, userID *uuid.UUID) error {
	event := &Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Level:     LevelRoom,
		Action:    extractAction(eventType),
		Data:      data,
		Timestamp: time.Now(),
		UserID:    userID,
		RoomID:    &roomID,
	}

	return ep.publishEvent(ctx, fmt.Sprintf("room:%s", roomID.String()), event)
}

// PublishMessageEvent publishes message-related events
func (ep *EventPublisher) PublishMessageEvent(ctx context.Context, eventType string, roomID uuid.UUID, messageID uuid.UUID, data map[string]interface{}, userID *uuid.UUID) error {
	event := &Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Level:     LevelMessage,
		Action:    extractAction(eventType),
		Data:      data,
		Timestamp: time.Now(),
		UserID:    userID,
		RoomID:    &roomID,
	}

	// Add message ID to data
	if event.Data == nil {
		event.Data = make(map[string]interface{})
	}
	event.Data["message_id"] = messageID

	return ep.publishEvent(ctx, fmt.Sprintf("room:%s", roomID.String()), event)
}

// PublishTypingEvent publishes typing indicator events
func (ep *EventPublisher) PublishTypingEvent(ctx context.Context, roomID uuid.UUID, userID uuid.UUID, isTyping bool) error {
	eventType := UserTypingStart
	if !isTyping {
		eventType = UserTypingStop
	}

	data := map[string]interface{}{
		"is_typing": isTyping,
	}

	event := &Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Level:     LevelUser,
		Action:    extractAction(eventType),
		Data:      data,
		Timestamp: time.Now(),
		UserID:    &userID,
		RoomID:    &roomID,
	}

	// Publish to both user and room channels
	if err := ep.publishEvent(ctx, fmt.Sprintf("user:%s", userID.String()), event); err != nil {
		return err
	}

	return ep.publishEvent(ctx, fmt.Sprintf("room:%s", roomID.String()), event)
}

// PublishSystemEvent publishes system-wide events
func (ep *EventPublisher) PublishSystemEvent(ctx context.Context, eventType string, data map[string]interface{}) error {
	event := &Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Level:     LevelSystem,
		Action:    extractAction(eventType),
		Data:      data,
		Timestamp: time.Now(),
	}

	return ep.publishEvent(ctx, "system", event)
}

// PublishPresenceEvent publishes user presence events
func (ep *EventPublisher) PublishPresenceEvent(ctx context.Context, userID uuid.UUID, status string, metadata map[string]interface{}) error {
	eventType := UserOnline
	if status == "offline" {
		eventType = UserOffline
	}

	data := map[string]interface{}{
		"status": status,
	}

	// Merge additional metadata
	for k, v := range metadata {
		data[k] = v
	}

	event := &Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Level:     LevelUser,
		Action:    extractAction(eventType),
		Data:      data,
		Metadata:  metadata,
		Timestamp: time.Now(),
		UserID:    &userID,
	}

	// Publish to user-specific channel
	if err := ep.publishEvent(ctx, fmt.Sprintf("user:%s", userID.String()), event); err != nil {
		return err
	}

	// Also publish to global presence channel
	return ep.publishEvent(ctx, "presence", event)
}

// PublishToChannel publishes event to a specific channel
func (ep *EventPublisher) PublishToChannel(ctx context.Context, channel string, event *Event) error {
	return ep.publishEvent(ctx, channel, event)
}

// PublishGlobalEvent publishes event to global channel (all connected users)
func (ep *EventPublisher) PublishGlobalEvent(ctx context.Context, eventType string, data map[string]interface{}) error {
	event := &Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Level:     extractLevel(eventType),
		Action:    extractAction(eventType),
		Data:      data,
		Timestamp: time.Now(),
	}

	return ep.publishEvent(ctx, "global", event)
}

// Private methods

func (ep *EventPublisher) publishEvent(ctx context.Context, channel string, event *Event) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return ep.redis.PublishRoomMessage(ctx, channel, string(eventData))
}

// extractLevel extracts level from event type (event.level.action)
func extractLevel(eventType string) string {
	parts := splitEventType(eventType)
	if len(parts) >= 2 {
		return parts[1]
	}
	return "unknown"
}

// extractAction extracts action from event type (event.level.action)
func extractAction(eventType string) string {
	parts := splitEventType(eventType)
	if len(parts) >= 3 {
		// Join remaining parts for multi-word actions
		action := ""
		for i := 2; i < len(parts); i++ {
			if i > 2 {
				action += "."
			}
			action += parts[i]
		}
		return action
	}
	return "unknown"
}

// splitEventType splits event type by dots
func splitEventType(eventType string) []string {
	parts := []string{}
	current := ""

	for _, char := range eventType {
		if char == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// Event data structure helpers

// UserEventData creates standardized user event data
func UserEventData(userID uuid.UUID, additionalData map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{
		"user_id": userID,
	}

	for k, v := range additionalData {
		data[k] = v
	}

	return data
}

// RoomEventData creates standardized room event data
func RoomEventData(roomID uuid.UUID, userID *uuid.UUID, additionalData map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{
		"room_id": roomID,
	}

	if userID != nil {
		data["user_id"] = *userID
	}

	for k, v := range additionalData {
		data[k] = v
	}

	return data
}

// MessageEventData creates standardized message event data
func MessageEventData(messageID, roomID uuid.UUID, userID *uuid.UUID, additionalData map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{
		"message_id": messageID,
		"room_id":    roomID,
	}

	if userID != nil {
		data["user_id"] = *userID
	}

	for k, v := range additionalData {
		data[k] = v
	}

	return data
}

// TypingEventData creates standardized typing event data
func TypingEventData(roomID, userID uuid.UUID, isTyping bool) map[string]interface{} {
	return map[string]interface{}{
		"room_id":   roomID,
		"user_id":   userID,
		"is_typing": isTyping,
	}
}

// PresenceEventData creates standardized presence event data
func PresenceEventData(userID uuid.UUID, status string, lastSeen *time.Time, additionalData map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{
		"user_id": userID,
		"status":  status,
	}

	if lastSeen != nil {
		data["last_seen"] = *lastSeen
	}

	for k, v := range additionalData {
		data[k] = v
	}

	return data
}
