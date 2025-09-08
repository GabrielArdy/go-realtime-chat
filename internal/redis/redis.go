package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"realtime-api/internal/config"
	"realtime-api/internal/logger"

	"github.com/redis/rueidis"
)

type Redis struct {
	client rueidis.Client
}

type PubSubMessage struct {
	Channel string      `json:"channel"`
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
}

var Client *Redis

func Init(cfg *config.RedisConfig) (*Redis, error) {
	options := rueidis.ClientOption{
		InitAddress: []string{fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)},
		SelectDB:    cfg.Database,
	}

	if cfg.Password != "" {
		options.Password = cfg.Password
	}

	client, err := rueidis.NewClient(options)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	redisClient := &Redis{client: client}
	Client = redisClient

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pong := client.B().Ping().Build()
	resp := client.Do(ctx, pong)
	if err := resp.Error(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	logger.Info("Redis connected successfully", logger.WithFields(map[string]interface{}{
		"host":     cfg.Host,
		"port":     cfg.Port,
		"database": cfg.Database,
	}))

	return redisClient, nil
}

func GetClient() *Redis {
	if Client == nil {
		logger.Fatal("Redis client not initialized")
	}
	return Client
}

func GetRawClient() rueidis.Client {
	if Client == nil {
		logger.Fatal("Redis client not initialized")
	}
	return Client.client
}

// Common Redis operations
func (r *Redis) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	var cmd rueidis.Completed
	if expiration > 0 {
		cmd = r.client.B().Set().Key(key).Value(value).ExSeconds(int64(expiration.Seconds())).Build()
	} else {
		cmd = r.client.B().Set().Key(key).Value(value).Build()
	}

	return r.client.Do(ctx, cmd).Error()
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	cmd := r.client.B().Get().Key(key).Build()
	resp := r.client.Do(ctx, cmd)
	if err := resp.Error(); err != nil {
		return "", err
	}
	return resp.ToString()
}

func (r *Redis) Del(ctx context.Context, keys ...string) (int64, error) {
	cmd := r.client.B().Del().Key(keys...).Build()
	resp := r.client.Do(ctx, cmd)
	if err := resp.Error(); err != nil {
		return 0, err
	}
	return resp.ToInt64()
}

func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	cmd := r.client.B().Exists().Key(key).Build()
	resp := r.client.Do(ctx, cmd)
	if err := resp.Error(); err != nil {
		return false, err
	}
	count, err := resp.ToInt64()
	return count > 0, err
}

func (r *Redis) Incr(ctx context.Context, key string) (int64, error) {
	cmd := r.client.B().Incr().Key(key).Build()
	resp := r.client.Do(ctx, cmd)
	if err := resp.Error(); err != nil {
		return 0, err
	}
	return resp.ToInt64()
}

func (r *Redis) Expire(ctx context.Context, key string, expiration time.Duration) error {
	cmd := r.client.B().Expire().Key(key).Seconds(int64(expiration.Seconds())).Build()
	return r.client.Do(ctx, cmd).Error()
}

func (r *Redis) HSet(ctx context.Context, key string, values map[string]interface{}) error {
	// For now, we'll set each field individually
	for field, value := range values {
		cmd := r.client.B().Hset().Key(key).FieldValue().FieldValue(field, fmt.Sprintf("%v", value)).Build()
		if err := r.client.Do(ctx, cmd).Error(); err != nil {
			return err
		}
	}
	return nil
}

func (r *Redis) HGet(ctx context.Context, key, field string) (string, error) {
	cmd := r.client.B().Hget().Key(key).Field(field).Build()
	resp := r.client.Do(ctx, cmd)
	if err := resp.Error(); err != nil {
		return "", err
	}
	return resp.ToString()
}

func (r *Redis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	cmd := r.client.B().Hgetall().Key(key).Build()
	resp := r.client.Do(ctx, cmd)
	if err := resp.Error(); err != nil {
		return nil, err
	}
	return resp.AsStrMap()
}

func (r *Redis) LPush(ctx context.Context, key string, values ...string) error {
	cmd := r.client.B().Lpush().Key(key).Element(values...).Build()
	return r.client.Do(ctx, cmd).Error()
}

func (r *Redis) RPop(ctx context.Context, key string) (string, error) {
	cmd := r.client.B().Rpop().Key(key).Build()
	resp := r.client.Do(ctx, cmd)
	if err := resp.Error(); err != nil {
		return "", err
	}
	return resp.ToString()
}

func (r *Redis) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := r.client.B().Ping().Build()
	return r.client.Do(ctx, cmd).Error()
}

func (r *Redis) Close() {
	r.client.Close()
}

// Pub/Sub operations for real-time chat
func (r *Redis) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	cmd := r.client.B().Publish().Channel(channel).Message(string(data)).Build()
	result := r.client.Do(ctx, cmd)
	if err := result.Error(); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	logger.Debug("Message published to Redis", logger.WithFields(map[string]interface{}{
		"channel": channel,
		"message": string(data),
	}))

	return nil
}

func (r *Redis) Subscribe(ctx context.Context, channels ...string) (rueidis.DedicatedClient, error) {
	client, cancel := r.client.Dedicate()
	go func() {
		<-ctx.Done()
		cancel()
	}()

	if err := client.Do(ctx, client.B().Subscribe().Channel(channels...).Build()).Error(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to subscribe to channels: %w", err)
	}

	logger.Info("Subscribed to Redis channels", logger.WithField("channels", channels))
	return client, nil
}

// Chat-specific Pub/Sub methods
func (r *Redis) PublishRoomMessage(ctx context.Context, roomID string, message interface{}) error {
	channel := fmt.Sprintf("room:%s", roomID)
	return r.Publish(ctx, channel, PubSubMessage{
		Channel: channel,
		Type:    "message",
		Data:    message,
	})
}

func (r *Redis) PublishUserEvent(ctx context.Context, userID string, eventType string, data interface{}) error {
	channel := fmt.Sprintf("user:%s", userID)
	return r.Publish(ctx, channel, PubSubMessage{
		Channel: channel,
		Type:    eventType,
		Data:    data,
	})
}

func (r *Redis) PublishTypingEvent(ctx context.Context, roomID string, userID string, isTyping bool) error {
	channel := fmt.Sprintf("room:%s:typing", roomID)
	return r.Publish(ctx, channel, PubSubMessage{
		Channel: channel,
		Type:    "typing",
		Data: map[string]interface{}{
			"user_id":   userID,
			"is_typing": isTyping,
		},
	})
}

func (r *Redis) PublishRoomEvent(ctx context.Context, roomID string, eventType string, data interface{}) error {
	channel := fmt.Sprintf("room:%s:events", roomID)
	return r.Publish(ctx, channel, PubSubMessage{
		Channel: channel,
		Type:    eventType,
		Data:    data,
	})
}

// User presence management
func (r *Redis) SetUserOnline(ctx context.Context, userID string) error {
	key := fmt.Sprintf("presence:%s", userID)
	return r.Set(ctx, key, "online", 5*time.Minute) // Auto-expire after 5 minutes
}

func (r *Redis) SetUserOffline(ctx context.Context, userID string) error {
	key := fmt.Sprintf("presence:%s", userID)
	_, err := r.Del(ctx, key)
	return err
}

func (r *Redis) IsUserOnline(ctx context.Context, userID string) (bool, error) {
	key := fmt.Sprintf("presence:%s", userID)
	return r.Exists(ctx, key)
}

// Room membership cache
func (r *Redis) AddUserToRoom(ctx context.Context, roomID, userID string) error {
	key := fmt.Sprintf("room_members:%s", roomID)
	cmd := r.client.B().Sadd().Key(key).Member(userID).Build()
	return r.client.Do(ctx, cmd).Error()
}

func (r *Redis) RemoveUserFromRoom(ctx context.Context, roomID, userID string) error {
	key := fmt.Sprintf("room_members:%s", roomID)
	cmd := r.client.B().Srem().Key(key).Member(userID).Build()
	return r.client.Do(ctx, cmd).Error()
}

func (r *Redis) GetRoomMembers(ctx context.Context, roomID string) ([]string, error) {
	key := fmt.Sprintf("room_members:%s", roomID)
	cmd := r.client.B().Smembers().Key(key).Build()
	result := r.client.Do(ctx, cmd)
	if err := result.Error(); err != nil {
		return nil, err
	}
	return result.AsStrSlice()
}

func (r *Redis) IsUserInRoom(ctx context.Context, roomID, userID string) (bool, error) {
	key := fmt.Sprintf("room_members:%s", roomID)
	cmd := r.client.B().Sismember().Key(key).Member(userID).Build()
	result := r.client.Do(ctx, cmd)
	if err := result.Error(); err != nil {
		return false, err
	}
	return result.AsBool()
}
