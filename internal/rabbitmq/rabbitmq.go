package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"realtime-api/internal/config"
	"realtime-api/internal/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	config     *config.RabbitMQConfig
}

type MessageHandler func(body []byte) error

var Client *RabbitMQ

func Init(cfg *config.RabbitMQConfig) (*RabbitMQ, error) {
	var url string
	if cfg.URL != "" {
		url = cfg.URL
	} else {
		url = fmt.Sprintf("amqp://%s:%s@%s:%s%s",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.VHost)
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = ch.ExchangeDeclare(
		cfg.Exchange, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	_, err = ch.QueueDeclare(
		cfg.Queue, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = ch.QueueBind(
		cfg.Queue,      // queue name
		cfg.RoutingKey, // routing key
		cfg.Exchange,   // exchange
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	rabbitMQ := &RabbitMQ{
		connection: conn,
		channel:    ch,
		config:     cfg,
	}

	Client = rabbitMQ

	logger.Info("RabbitMQ connected successfully", logger.WithFields(map[string]interface{}{
		"host":     cfg.Host,
		"port":     cfg.Port,
		"exchange": cfg.Exchange,
		"queue":    cfg.Queue,
	}))

	return rabbitMQ, nil
}

func (r *RabbitMQ) PublishMessage(routingKey string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = r.channel.PublishWithContext(
		ctx,
		r.config.Exchange, // exchange
		routingKey,        // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // make message persistent
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	logger.Debug("Message published to RabbitMQ", logger.WithFields(map[string]interface{}{
		"routing_key": routingKey,
		"message":     string(body),
	}))

	return nil
}

func (r *RabbitMQ) ConsumeMessages(queueName string, handler MessageHandler) error {
	msgs, err := r.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for d := range msgs {
			err := handler(d.Body)
			if err != nil {
				logger.Error("Failed to handle message", logger.WithFields(map[string]interface{}{
					"error":   err.Error(),
					"message": string(d.Body),
				}))
				d.Nack(false, true) // negative acknowledge and requeue
			} else {
				d.Ack(false) // acknowledge
			}
		}
	}()

	logger.Info("Started consuming messages", logger.WithField("queue", queueName))
	return nil
}

func (r *RabbitMQ) PublishUserEvent(userID string, eventType string, data interface{}) error {
	event := map[string]interface{}{
		"user_id":    userID,
		"event_type": eventType,
		"data":       data,
		"timestamp":  time.Now(),
	}

	routingKey := fmt.Sprintf("user.%s.%s", userID, eventType)
	return r.PublishMessage(routingKey, event)
}

func (r *RabbitMQ) PublishRoomEvent(roomID string, eventType string, data interface{}) error {
	event := map[string]interface{}{
		"room_id":    roomID,
		"event_type": eventType,
		"data":       data,
		"timestamp":  time.Now(),
	}

	routingKey := fmt.Sprintf("room.%s.%s", roomID, eventType)
	return r.PublishMessage(routingKey, event)
}

func (r *RabbitMQ) PublishMessageEvent(messageData interface{}) error {
	routingKey := "message.new"
	return r.PublishMessage(routingKey, messageData)
}

func (r *RabbitMQ) PublishTypingEvent(roomID, userID string, isTyping bool) error {
	event := map[string]interface{}{
		"room_id":   roomID,
		"user_id":   userID,
		"is_typing": isTyping,
		"timestamp": time.Now(),
	}

	routingKey := fmt.Sprintf("typing.%s", roomID)
	return r.PublishMessage(routingKey, event)
}

func (r *RabbitMQ) PublishNotificationEvent(userID string, notification interface{}) error {
	event := map[string]interface{}{
		"user_id":      userID,
		"notification": notification,
		"timestamp":    time.Now(),
	}

	routingKey := fmt.Sprintf("notification.%s", userID)
	return r.PublishMessage(routingKey, event)
}

func (r *RabbitMQ) Health() error {
	if r.connection == nil || r.connection.IsClosed() {
		return fmt.Errorf("RabbitMQ connection is closed")
	}
	if r.channel == nil || r.channel.IsClosed() {
		return fmt.Errorf("RabbitMQ channel is closed")
	}
	return nil
}

func (r *RabbitMQ) Close() error {
	var err error
	if r.channel != nil {
		err = r.channel.Close()
	}
	if r.connection != nil {
		connErr := r.connection.Close()
		if err == nil {
			err = connErr
		}
	}
	return err
}

func GetClient() *RabbitMQ {
	return Client
}
