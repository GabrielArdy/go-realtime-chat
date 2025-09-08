package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"realtime-api/internal/events"
	"realtime-api/internal/jwt"
	"realtime-api/internal/logger"
	"realtime-api/internal/model"
	"realtime-api/internal/redis"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type Hub struct {
	clients        map[*Client]bool
	rooms          map[uuid.UUID]map[*Client]bool
	userRooms      map[uuid.UUID][]uuid.UUID // user_id -> room_ids
	register       chan *Client
	unregister     chan *Client
	broadcast      chan []byte
	mutex          sync.RWMutex
	eventPublisher *events.EventPublisher
	redis          *redis.Redis
}

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   uuid.UUID
	username string
	deviceID string
	rooms    map[uuid.UUID]bool
	mutex    sync.RWMutex
}

type Message struct {
	Type      model.WSMessageType `json:"type"`
	Data      interface{}         `json:"data,omitempty"`
	Timestamp time.Time           `json:"timestamp"`
	ID        string              `json:"id,omitempty"`
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// In production, implement proper origin checking
			origin := r.Header.Get("Origin")

			// Allow all origins in development
			// In production, check against allowed origins list
			allowedOrigins := []string{
				"http://localhost:3000", // React dev server
				"http://localhost:8080", // Backend server
				"https://yourapp.com",   // Production frontend
			}

			// For development, allow any origin
			if origin == "" {
				return true // Allow connections without Origin header (e.g., native apps)
			}

			for _, allowed := range allowedOrigins {
				if origin == allowed {
					return true
				}
			}

			logger.Warn("WebSocket connection rejected", logger.WithField("origin", origin))
			return false // Reject unknown origins
		},
	}

	GlobalHub *Hub
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

func NewHub(redis *redis.Redis) *Hub {
	return &Hub{
		clients:        make(map[*Client]bool),
		rooms:          make(map[uuid.UUID]map[*Client]bool),
		userRooms:      make(map[uuid.UUID][]uuid.UUID),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		broadcast:      make(chan []byte, 256),
		eventPublisher: events.NewEventPublisher(redis),
		redis:          redis,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()

			logger.Info("Client connected", logger.WithFields(map[string]interface{}{
				"user_id":   client.userID.String(),
				"username":  client.username,
				"device_id": client.deviceID,
			}))

			// Send confirmation message
			client.send <- h.createMessage(model.WSTypeAuth, map[string]interface{}{
				"status":  "connected",
				"user_id": client.userID,
			})

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				h.removeClientFromAllRooms(client)
				delete(h.clients, client)
				close(client.send)
			}
			h.mutex.Unlock()

			logger.Info("Client disconnected", logger.WithFields(map[string]interface{}{
				"user_id":   client.userID.String(),
				"username":  client.username,
				"device_id": client.deviceID,
			}))

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					h.removeClientFromAllRooms(client)
					delete(h.clients, client)
					close(client.send)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

func (h *Hub) removeClientFromAllRooms(client *Client) {
	client.mutex.RLock()
	for roomID := range client.rooms {
		if room, exists := h.rooms[roomID]; exists {
			delete(room, client)
			if len(room) == 0 {
				delete(h.rooms, roomID)
			}
		}
	}
	client.mutex.RUnlock()

	// Clear user rooms mapping
	if rooms, exists := h.userRooms[client.userID]; exists {
		for _, roomID := range rooms {
			h.broadcastToRoom(roomID, model.WSTypeUserLeave, map[string]interface{}{
				"user_id":  client.userID,
				"username": client.username,
			})
		}
		delete(h.userRooms, client.userID)
	}
}

func (h *Hub) createMessage(msgType model.WSMessageType, data interface{}) []byte {
	msg := Message{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now(),
		ID:        uuid.New().String(),
	}

	msgBytes, _ := json.Marshal(msg)
	return msgBytes
}

func (h *Hub) JoinRoom(userID, roomID uuid.UUID) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, exists := h.rooms[roomID]; !exists {
		h.rooms[roomID] = make(map[*Client]bool)
	}

	// Add user to room for all their clients
	for client := range h.clients {
		if client.userID == userID {
			h.rooms[roomID][client] = true
			client.mutex.Lock()
			client.rooms[roomID] = true
			client.mutex.Unlock()
		}
	}

	// Update user rooms mapping
	if rooms, exists := h.userRooms[userID]; exists {
		// Check if room already exists in user's rooms
		found := false
		for _, id := range rooms {
			if id == roomID {
				found = true
				break
			}
		}
		if !found {
			h.userRooms[userID] = append(rooms, roomID)
		}
	} else {
		h.userRooms[userID] = []uuid.UUID{roomID}
	}
}

func (h *Hub) LeaveRoom(userID, roomID uuid.UUID) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if room, exists := h.rooms[roomID]; exists {
		// Remove user from room for all their clients
		for client := range h.clients {
			if client.userID == userID {
				delete(room, client)
				client.mutex.Lock()
				delete(client.rooms, roomID)
				client.mutex.Unlock()
			}
		}

		if len(room) == 0 {
			delete(h.rooms, roomID)
		}
	}

	// Update user rooms mapping
	if rooms, exists := h.userRooms[userID]; exists {
		newRooms := make([]uuid.UUID, 0)
		for _, id := range rooms {
			if id != roomID {
				newRooms = append(newRooms, id)
			}
		}
		if len(newRooms) == 0 {
			delete(h.userRooms, userID)
		} else {
			h.userRooms[userID] = newRooms
		}
	}
}

func (h *Hub) broadcastToRoom(roomID uuid.UUID, msgType model.WSMessageType, data interface{}) {
	message := h.createMessage(msgType, data)

	h.mutex.RLock()
	if room, exists := h.rooms[roomID]; exists {
		for client := range room {
			select {
			case client.send <- message:
			default:
				delete(room, client)
				close(client.send)
			}
		}
	}
	h.mutex.RUnlock()
}

// BroadcastToRoom is the public method for broadcasting to a room
func (h *Hub) BroadcastToRoom(roomID uuid.UUID, msgType model.WSMessageType, data interface{}) {
	h.broadcastToRoom(roomID, msgType, data)
}

func (h *Hub) BroadcastToUser(userID uuid.UUID, msgType model.WSMessageType, data interface{}) {
	message := h.createMessage(msgType, data)

	h.mutex.RLock()
	for client := range h.clients {
		if client.userID == userID {
			select {
			case client.send <- message:
			default:
				h.removeClientFromAllRooms(client)
				delete(h.clients, client)
				close(client.send)
			}
		}
	}
	h.mutex.RUnlock()
}

func HandleWebSocket(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		logger.Error("WebSocket upgrade failed", logger.WithField("error", err.Error()))
		return err
	}

	// Extract token from query parameter or header
	token := c.QueryParam("token")
	if token == "" {
		token = c.Request().Header.Get("Authorization")
		if token != "" && len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}

	if token == "" {
		conn.Close()
		return echo.NewHTTPError(http.StatusUnauthorized, "missing authentication token")
	}

	// Validate JWT token
	claims, err := jwt.GetService().ValidateToken(token)
	if err != nil {
		conn.Close()
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	client := &Client{
		hub:      GlobalHub,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   claims.UserID,
		username: claims.Username,
		deviceID: claims.DeviceID,
		rooms:    make(map[uuid.UUID]bool),
	}

	client.hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()

	return nil
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket error", logger.WithField("error", err.Error()))
			}
			break
		}

		var wsMsg model.WSMessage
		if err := json.Unmarshal(messageBytes, &wsMsg); err != nil {
			logger.Error("Failed to unmarshal WebSocket message", logger.WithField("error", err.Error()))
			continue
		}

		c.handleMessage(&wsMsg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(wsMsg *model.WSMessage) {
	switch wsMsg.Type {
	case model.WSTypePing:
		c.send <- c.hub.createMessage(model.WSTypePong, nil)

	case model.WSTypeTypingStart:
		c.handleTypingStart(wsMsg.Data)

	case model.WSTypeTypingStop:
		c.handleTypingStop(wsMsg.Data)

	case model.WSTypeUserStatusChange:
		c.handleUserStatusChange(wsMsg.Data)

	default:
		logger.Warn("Unknown WebSocket message type", logger.WithField("type", wsMsg.Type))
	}
}

func (c *Client) handleTypingStart(data interface{}) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	roomIDStr, ok := dataMap["room_id"].(string)
	if !ok {
		return
	}

	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		return
	}

	// Publish typing event using event system
	if c.hub.eventPublisher != nil {
		ctx := context.Background()
		c.hub.eventPublisher.PublishTypingEvent(ctx, roomID, c.userID, true)
	}

	// Broadcast to room members
	c.hub.broadcastToRoom(roomID, model.WSTypeTypingStart, map[string]interface{}{
		"room_id":  roomID,
		"user_id":  c.userID,
		"username": c.username,
	})
}

func (c *Client) handleTypingStop(data interface{}) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	roomIDStr, ok := dataMap["room_id"].(string)
	if !ok {
		return
	}

	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		return
	}

	// Publish typing event using event system
	if c.hub.eventPublisher != nil {
		ctx := context.Background()
		c.hub.eventPublisher.PublishTypingEvent(ctx, roomID, c.userID, false)
	}

	// Broadcast to room members
	c.hub.broadcastToRoom(roomID, model.WSTypeTypingStop, map[string]interface{}{
		"room_id":  roomID,
		"user_id":  c.userID,
		"username": c.username,
	})
}

func (c *Client) handleUserStatusChange(data interface{}) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	status, ok := dataMap["status"].(string)
	if !ok {
		return
	}

	// Publish user status change using event system
	if c.hub.eventPublisher != nil {
		ctx := context.Background()
		c.hub.eventPublisher.PublishUserEvent(ctx, events.UserStatusChange, c.userID, map[string]interface{}{
			"status": status,
		})
	}

	// Broadcast status change to user's rooms
	c.mutex.RLock()
	for roomID := range c.rooms {
		c.hub.broadcastToRoom(roomID, model.WSTypeUserStatusChange, map[string]interface{}{
			"user_id":  c.userID,
			"username": c.username,
			"status":   status,
		})
	}
	c.mutex.RUnlock()
}

func Init(redis *redis.Redis) {
	GlobalHub = NewHub(redis)
	go GlobalHub.Run()

	logger.Info("WebSocket hub initialized")
}

func GetHub() *Hub {
	return GlobalHub
}
