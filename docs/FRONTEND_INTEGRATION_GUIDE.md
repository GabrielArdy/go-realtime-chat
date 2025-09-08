# Frontend Integration Guide
## Real-time Chat API Integration

### 📋 Table of Contents
1. [Overview](#overview)
2. [API Schemas](#api-schemas)
3. [WebSocket Payload Schemas](#websocket-payload-schemas)
4. [Authentication Flow](#authentication-flow)
5. [WebSocket Connection](#websocket-connection)
6. [API Endpoints](#api-endpoints)
7. [Real-time Events](#real-time-events)
8. [Chat Implementation](#chat-implementation)
9. [Code Examples](#code-examples)
10. [Error Handling](#error-handling)
11. [Best Practices](#best-practices)

---

## 🎯 Overview

This guide provides complete integration instructions for frontend applications to connect with our real-time chat API. The system supports:

- **Real-time messaging** via WebSocket
- **Event-driven architecture** for instant updates
- **RESTful API** for CRUD operations
- **JWT authentication** for security
- **Private & Group chats** with member management

### Architecture Flow
```
Frontend App ↔ REST API (CRUD) ↔ Event System ↔ WebSocket ↔ Real-time Updates
```

---

## 📊 API Schemas

### Standard Response Format

#### Success Response
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {
    // Response data here
  },
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

#### Error Response
```json
{
  "success": false,
  "message": "Error description",
  "error": {
    "code": "VALIDATION_ERROR",
    "details": [
      {
        "field": "username",
        "message": "Username is required"
      }
    ]
  }
}
```

### Authentication Schemas

#### Login Request
```json
{
  "username": "john_doe",
  "password": "securePassword123"
}
```

#### Login Response
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "uuid-string",
      "username": "john_doe",
      "email": "john@example.com",
      "display_name": "John Doe",
      "avatar": "https://example.com/avatar.jpg",
      "status": "online",
      "is_verified": true,
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z"
    }
  }
}
```

#### Register Request
```json
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "securePassword123",
  "display_name": "John Doe"
}
```

### User Schemas

#### User Object
```json
{
  "id": "uuid-string",
  "username": "john_doe",
  "email": "john@example.com",
  "display_name": "John Doe",
  "avatar": "https://example.com/avatar.jpg",
  "status": "online",
  "last_seen": "2023-01-01T00:00:00Z",
  "is_verified": true,
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z"
}
```

### Room Schemas

#### Room Object
```json
{
  "id": "uuid-string",
  "name": "General Chat",
  "description": "Main discussion room",
  "type": "group",
  "avatar": "https://example.com/room-avatar.jpg",
  "is_public": true,
  "max_members": 100,
  "member_count": 25,
  "created_by": "uuid-string",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z",
  "last_message": {
    "id": "uuid-string",
    "content": "Hello everyone!",
    "type": "text",
    "user": {
      "id": "uuid-string",
      "username": "jane_doe",
      "display_name": "Jane Doe"
    },
    "created_at": "2023-01-01T12:00:00Z"
  },
  "unread_count": 3
}
```

#### Create Room Request
```json
{
  "name": "Project Discussion",
  "description": "Discussion for project XYZ",
  "type": "group",
  "avatar": "https://example.com/avatar.jpg",
  "is_public": false,
  "max_members": 50
}
```

#### Room List Response
```json
{
  "success": true,
  "data": {
    "rooms": [
      {
        // Room object here
      }
    ]
  },
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 5,
    "total_pages": 1
  }
}
```

### Message Schemas

#### Message Object
```json
{
  "id": "uuid-string",
  "room_id": "uuid-string",
  "user_id": "uuid-string",
  "content": "Hello, how are you?",
  "type": "text",
  "reply_to": "uuid-string",
  "attachments": [
    {
      "id": "uuid-string",
      "filename": "document.pdf",
      "url": "https://example.com/files/document.pdf",
      "type": "file",
      "size": 1024000
    }
  ],
  "reactions": [
    {
      "emoji": "👍",
      "count": 3,
      "users": ["uuid-1", "uuid-2", "uuid-3"],
      "user_reacted": true
    }
  ],
  "user": {
    "id": "uuid-string",
    "username": "john_doe",
    "display_name": "John Doe",
    "avatar": "https://example.com/avatar.jpg"
  },
  "edited_at": null,
  "created_at": "2023-01-01T12:00:00Z",
  "updated_at": "2023-01-01T12:00:00Z"
}
```

#### Send Message Request
```json
{
  "room_id": "uuid-string",
  "content": "Hello everyone!",
  "type": "text",
  "reply_to": "uuid-string",
  "attachments": [
    {
      "filename": "image.jpg",
      "url": "https://example.com/uploads/image.jpg",
      "type": "image",
      "size": 524288
    }
  ]
}
```

#### Edit Message Request
```json
{
  "content": "Updated message content"
}
```

#### Add Reaction Request
```json
{
  "emoji": "👍"
}
```

#### Messages List Response
```json
{
  "success": true,
  "data": {
    "messages": [
      {
        // Message object here
      }
    ]
  },
  "pagination": {
    "page": 1,
    "limit": 50,
    "total": 150,
    "total_pages": 3
  }
}
```

### Room Member Schemas

#### Room Member Object
```json
{
  "user_id": "uuid-string",
  "room_id": "uuid-string",
  "role": "member",
  "joined_at": "2023-01-01T00:00:00Z",
  "user": {
    "id": "uuid-string",
    "username": "john_doe",
    "display_name": "John Doe",
    "avatar": "https://example.com/avatar.jpg",
    "status": "online"
  }
}
```

#### Add Member Request
```json
{
  "user_id": "uuid-string",
  "role": "member"
}
```

---

## 🔌 WebSocket Payload Schemas

### Connection & Authentication

#### Authentication Message (Client → Server)
```json
{
  "type": "auth",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Authentication Response (Server → Client)
```json
{
  "type": "auth_response",
  "success": true,
  "message": "Authentication successful",
  "user_id": "uuid-string"
}
```

#### Join Room Message (Client → Server)
```json
{
  "type": "join_room",
  "room_id": "uuid-string"
}
```

#### Leave Room Message (Client → Server)
```json
{
  "type": "leave_room",
  "room_id": "uuid-string"
}
```

### Message Events

#### New Message (Server → Client)
```json
{
  "type": "message",
  "event_id": "uuid-string",
  "room_id": "uuid-string",
  "message": {
    "id": "uuid-string",
    "room_id": "uuid-string",
    "user_id": "uuid-string",
    "content": "Hello everyone!",
    "type": "text",
    "reply_to": null,
    "attachments": [],
    "reactions": [],
    "user": {
      "id": "uuid-string",
      "username": "john_doe",
      "display_name": "John Doe",
      "avatar": "https://example.com/avatar.jpg"
    },
    "created_at": "2023-01-01T12:00:00Z",
    "updated_at": "2023-01-01T12:00:00Z"
  },
  "timestamp": "2023-01-01T12:00:00Z"
}
```

#### Message Edited (Server → Client)
```json
{
  "type": "message_edit",
  "event_id": "uuid-string",
  "room_id": "uuid-string",
  "message_id": "uuid-string",
  "content": "Updated message content",
  "edited_at": "2023-01-01T12:05:00Z",
  "user_id": "uuid-string",
  "timestamp": "2023-01-01T12:05:00Z"
}
```

#### Message Deleted (Server → Client)
```json
{
  "type": "message_delete",
  "event_id": "uuid-string",
  "room_id": "uuid-string",
  "message_id": "uuid-string",
  "user_id": "uuid-string",
  "timestamp": "2023-01-01T12:10:00Z"
}
```

#### Message Reaction (Server → Client)
```json
{
  "type": "message_reaction",
  "event_id": "uuid-string",
  "room_id": "uuid-string",
  "message_id": "uuid-string",
  "emoji": "👍",
  "action": "add",
  "user_id": "uuid-string",
  "reaction_count": 5,
  "timestamp": "2023-01-01T12:15:00Z"
}
```

### Typing Events

#### Start Typing (Client → Server)
```json
{
  "type": "typing_start",
  "room_id": "uuid-string"
}
```

#### Stop Typing (Client → Server)
```json
{
  "type": "typing_stop",
  "room_id": "uuid-string"
}
```

#### Typing Started (Server → Client)
```json
{
  "type": "typing_start",
  "event_id": "uuid-string",
  "room_id": "uuid-string",
  "user_id": "uuid-string",
  "user": {
    "id": "uuid-string",
    "username": "jane_doe",
    "display_name": "Jane Doe"
  },
  "timestamp": "2023-01-01T12:20:00Z"
}
```

#### Typing Stopped (Server → Client)
```json
{
  "type": "typing_stop",
  "event_id": "uuid-string",
  "room_id": "uuid-string",
  "user_id": "uuid-string",
  "user": {
    "id": "uuid-string",
    "username": "jane_doe",
    "display_name": "Jane Doe"
  },
  "timestamp": "2023-01-01T12:22:00Z"
}
```

### Room Events

#### User Joined Room (Server → Client)
```json
{
  "type": "user_join",
  "event_id": "uuid-string",
  "room_id": "uuid-string",
  "user_id": "uuid-string",
  "user": {
    "id": "uuid-string",
    "username": "new_user",
    "display_name": "New User",
    "avatar": "https://example.com/avatar.jpg"
  },
  "role": "member",
  "timestamp": "2023-01-01T12:25:00Z"
}
```

#### User Left Room (Server → Client)
```json
{
  "type": "user_leave",
  "event_id": "uuid-string",
  "room_id": "uuid-string",
  "user_id": "uuid-string",
  "user": {
    "id": "uuid-string",
    "username": "leaving_user",
    "display_name": "Leaving User"
  },
  "timestamp": "2023-01-01T12:30:00Z"
}
```

#### Room Updated (Server → Client)
```json
{
  "type": "room_update",
  "event_id": "uuid-string",
  "room_id": "uuid-string",
  "changes": {
    "name": "Updated Room Name",
    "description": "New description"
  },
  "updated_by": "uuid-string",
  "timestamp": "2023-01-01T12:35:00Z"
}
```

#### Room Deleted (Server → Client)
```json
{
  "type": "room_delete",
  "event_id": "uuid-string",
  "room_id": "uuid-string",
  "deleted_by": "uuid-string",
  "timestamp": "2023-01-01T12:40:00Z"
}
```

### User Status Events

#### User Status Change (Server → Client)
```json
{
  "type": "user_status_change",
  "event_id": "uuid-string",
  "user_id": "uuid-string",
  "status": "online",
  "last_seen": "2023-01-01T12:45:00Z",
  "timestamp": "2023-01-01T12:45:00Z"
}
```

#### User Profile Update (Server → Client)
```json
{
  "type": "user_profile_update",
  "event_id": "uuid-string",
  "user_id": "uuid-string",
  "changes": {
    "display_name": "Updated Name",
    "avatar": "https://example.com/new-avatar.jpg"
  },
  "timestamp": "2023-01-01T12:50:00Z"
}
```

### System Events

#### Notification (Server → Client)
```json
{
  "type": "notification",
  "event_id": "uuid-string",
  "notification_type": "mention",
  "title": "You were mentioned",
  "message": "John Doe mentioned you in General Chat",
  "data": {
    "room_id": "uuid-string",
    "message_id": "uuid-string",
    "mentioned_by": "uuid-string"
  },
  "priority": "high",
  "timestamp": "2023-01-01T12:55:00Z"
}
```

#### Error Event (Server → Client)
```json
{
  "type": "error",
  "event_id": "uuid-string",
  "error_code": "UNAUTHORIZED",
  "message": "Authentication required",
  "details": {
    "action": "join_room",
    "room_id": "uuid-string"
  },
  "timestamp": "2023-01-01T13:00:00Z"
}
```

#### Heartbeat/Ping (Bidirectional)
```json
{
  "type": "ping",
  "timestamp": "2023-01-01T13:05:00Z"
}
```

```json
{
  "type": "pong",
  "timestamp": "2023-01-01T13:05:00Z"
}
```

### File Upload Events

#### File Upload Progress (Server → Client)
```json
{
  "type": "upload_progress",
  "event_id": "uuid-string",
  "upload_id": "uuid-string",
  "filename": "document.pdf",
  "progress": 65,
  "total_size": 1024000,
  "uploaded_size": 665600,
  "timestamp": "2023-01-01T13:10:00Z"
}
```

#### File Upload Complete (Server → Client)
```json
{
  "type": "upload_complete",
  "event_id": "uuid-string",
  "upload_id": "uuid-string",
  "file": {
    "id": "uuid-string",
    "filename": "document.pdf",
    "url": "https://example.com/files/document.pdf",
    "type": "file",
    "size": 1024000,
    "mime_type": "application/pdf"
  },
  "timestamp": "2023-01-01T13:12:00Z"
}
```

### WebSocket Message Types Summary

| Event Type | Direction | Description |
|------------|-----------|-------------|
| `auth` | Client → Server | Authenticate connection |
| `auth_response` | Server → Client | Authentication result |
| `join_room` | Client → Server | Join a room |
| `leave_room` | Client → Server | Leave a room |
| `message` | Server → Client | New message received |
| `message_edit` | Server → Client | Message was edited |
| `message_delete` | Server → Client | Message was deleted |
| `message_reaction` | Server → Client | Reaction added/removed |
| `typing_start` | Bidirectional | User started typing |
| `typing_stop` | Bidirectional | User stopped typing |
| `user_join` | Server → Client | User joined room |
| `user_leave` | Server → Client | User left room |
| `user_status_change` | Server → Client | User online status changed |
| `user_profile_update` | Server → Client | User profile updated |
| `room_update` | Server → Client | Room details updated |
| `room_delete` | Server → Client | Room was deleted |
| `notification` | Server → Client | System notification |
| `error` | Server → Client | Error occurred |
| `ping`/`pong` | Bidirectional | Connection heartbeat |
| `upload_progress` | Server → Client | File upload progress |
| `upload_complete` | Server → Client | File upload completed |

### Backend Event Type Constants

The backend defines the following event type constants that will be sent through WebSocket messages:

#### User Events
```javascript
const USER_EVENTS = {
  USER_ONLINE: "event.user.online",
  USER_OFFLINE: "event.user.offline",
  USER_TYPING_START: "event.user.typing.start",
  USER_TYPING_STOP: "event.user.typing.stop",
  USER_STATUS_CHANGE: "event.user.status.change",
  USER_PROFILE_UPDATE: "event.user.profile.update"
};
```

#### Room Events
```javascript
const ROOM_EVENTS = {
  ROOM_CREATE: "event.room.create",
  ROOM_UPDATE: "event.room.update",
  ROOM_DELETE: "event.room.delete",
  ROOM_JOIN: "event.room.join",
  ROOM_LEAVE: "event.room.leave",
  ROOM_MEMBER_ADD: "event.room.member.add",
  ROOM_MEMBER_REMOVE: "event.room.member.remove",
  ROOM_MEMBER_ROLE_UPDATE: "event.room.member.role.update",
  ROOM_INVITE_CREATE: "event.room.invite.create",
  ROOM_INVITE_ACCEPT: "event.room.invite.accept",
  ROOM_INVITE_REJECT: "event.room.invite.reject"
};
```

#### Message Events
```javascript
const MESSAGE_EVENTS = {
  MESSAGE_SEND: "event.message.send",
  MESSAGE_EDIT: "event.message.edit",
  MESSAGE_DELETE: "event.message.delete",
  MESSAGE_READ: "event.message.read",
  MESSAGE_REACTION_ADD: "event.message.reaction.add",
  MESSAGE_REACTION_REMOVE: "event.message.reaction.remove"
};
```

#### System Events
```javascript
const SYSTEM_EVENTS = {
  SYSTEM_MAINTENANCE: "event.system.maintenance",
  SYSTEM_SHUTDOWN: "event.system.shutdown",
  SYSTEM_BROADCAST: "event.system.broadcast"
};
```

#### All Event Types Combined
```javascript
const EVENT_TYPES = {
  ...USER_EVENTS,
  ...ROOM_EVENTS,
  ...MESSAGE_EVENTS,
  ...SYSTEM_EVENTS
};
```

#### Event Levels
```javascript
const EVENT_LEVELS = {
  USER: "user",
  ROOM: "room",
  MESSAGE: "message",
  SYSTEM: "system"
};
```

#### Example Usage in Frontend
```javascript
// Register specific event handlers using constants
chatWS.on(EVENT_TYPES.MESSAGE_SEND, handleNewMessage);
chatWS.on(EVENT_TYPES.MESSAGE_EDIT, handleMessageEdit);
chatWS.on(EVENT_TYPES.MESSAGE_DELETE, handleMessageDelete);
chatWS.on(EVENT_TYPES.USER_TYPING_START, handleTypingStart);
chatWS.on(EVENT_TYPES.USER_TYPING_STOP, handleTypingStop);
chatWS.on(EVENT_TYPES.ROOM_JOIN, handleUserJoin);
chatWS.on(EVENT_TYPES.ROOM_LEAVE, handleUserLeave);
chatWS.on(EVENT_TYPES.MESSAGE_REACTION_ADD, handleMessageReaction);
chatWS.on(EVENT_TYPES.USER_STATUS_CHANGE, handleUserStatusChange);

// Check event types
const isMessageEvent = (eventType) => {
  return Object.values(MESSAGE_EVENTS).includes(eventType);
};

const isUserEvent = (eventType) => {
  return Object.values(USER_EVENTS).includes(eventType);
};

// Event handler with type checking
chatWS.handleMessage = (message) => {
  const { type } = message;
  
  switch (type) {
    case EVENT_TYPES.MESSAGE_SEND:
      handleNewMessage(message);
      break;
    case EVENT_TYPES.MESSAGE_EDIT:
      handleMessageEdit(message);
      break;
    case EVENT_TYPES.MESSAGE_DELETE:
      handleMessageDelete(message);
      break;
    case EVENT_TYPES.USER_TYPING_START:
      handleTypingStart(message);
      break;
    case EVENT_TYPES.USER_TYPING_STOP:
      handleTypingStop(message);
      break;
    case EVENT_TYPES.ROOM_JOIN:
      handleUserJoin(message);
      break;
    case EVENT_TYPES.ROOM_LEAVE:
      handleUserLeave(message);
      break;
    case EVENT_TYPES.MESSAGE_REACTION_ADD:
    case EVENT_TYPES.MESSAGE_REACTION_REMOVE:
      handleMessageReaction(message);
      break;
    case EVENT_TYPES.USER_STATUS_CHANGE:
      handleUserStatusChange(message);
      break;
    case EVENT_TYPES.USER_PROFILE_UPDATE:
      handleUserProfileUpdate(message);
      break;
    case EVENT_TYPES.ROOM_UPDATE:
      handleRoomUpdate(message);
      break;
    case EVENT_TYPES.ROOM_DELETE:
      handleRoomDelete(message);
      break;
    case EVENT_TYPES.SYSTEM_BROADCAST:
      handleSystemBroadcast(message);
      break;
    default:
      console.warn(`Unhandled event type: ${type}`);
  }
};
```

---

## 🔐 Authentication Flow

### 1. User Login
```javascript
// POST /api/v1/auth/login
const loginUser = async (username, password) => {
  try {
    const response = await fetch('/api/v1/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        username,
        password
      })
    });
    
    const data = await response.json();
    
    if (data.success) {
      // Store JWT token
      localStorage.setItem('token', data.data.token);
      localStorage.setItem('user', JSON.stringify(data.data.user));
      return data.data;
    }
  } catch (error) {
    console.error('Login failed:', error);
    throw error;
  }
};
```

### 2. Token Management
```javascript
// Get stored token for API calls
const getAuthToken = () => {
  return localStorage.getItem('token');
};

// Create authenticated fetch wrapper
const authenticatedFetch = async (url, options = {}) => {
  const token = getAuthToken();
  
  return fetch(url, {
    ...options,
    headers: {
      ...options.headers,
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    }
  });
};
```

---

## 🔌 WebSocket Connection

### 1. Establish Connection
```javascript
class ChatWebSocket {
  constructor(token) {
    this.token = token;
    this.ws = null;
    this.reconnectInterval = 5000;
    this.maxReconnectAttempts = 5;
    this.reconnectAttempts = 0;
    this.eventHandlers = new Map();
  }

  connect() {
    try {
      // WebSocket connection with auth token
      this.ws = new WebSocket(`ws://localhost:8080/ws?token=${this.token}`);
      
      this.ws.onopen = this.onOpen.bind(this);
      this.ws.onmessage = this.onMessage.bind(this);
      this.ws.onclose = this.onClose.bind(this);
      this.ws.onerror = this.onError.bind(this);
      
    } catch (error) {
      console.error('WebSocket connection failed:', error);
      this.scheduleReconnect();
    }
  }

  onOpen(event) {
    console.log('WebSocket connected');
    this.reconnectAttempts = 0;
    
    // Send authentication message
    this.send({
      type: 'auth',
      token: this.token
    });
  }

  onMessage(event) {
    try {
      const message = JSON.parse(event.data);
      this.handleMessage(message);
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error);
    }
  }

  onClose(event) {
    console.log('WebSocket disconnected');
    this.scheduleReconnect();
  }

  onError(error) {
    console.error('WebSocket error:', error);
  }

  send(data) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }

  // Event handler registration
  on(eventType, handler) {
    if (!this.eventHandlers.has(eventType)) {
      this.eventHandlers.set(eventType, []);
    }
    this.eventHandlers.get(eventType).push(handler);
  }

  handleMessage(message) {
    const handlers = this.eventHandlers.get(message.type);
    if (handlers) {
      handlers.forEach(handler => handler(message));
    }
  }

  scheduleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      setTimeout(() => {
        this.reconnectAttempts++;
        console.log(`Reconnecting... Attempt ${this.reconnectAttempts}`);
        this.connect();
      }, this.reconnectInterval);
    }
  }
}
```

### 2. Initialize WebSocket
```javascript
// Initialize WebSocket after login
const initializeWebSocket = (token) => {
  const chatWS = new ChatWebSocket(token);
  
  // Register event handlers
  chatWS.on('message', handleNewMessage);
  chatWS.on('message_edit', handleMessageEdit);
  chatWS.on('message_delete', handleMessageDelete);
  chatWS.on('typing_start', handleTypingStart);
  chatWS.on('typing_stop', handleTypingStop);
  chatWS.on('user_join', handleUserJoin);
  chatWS.on('user_leave', handleUserLeave);
  chatWS.on('message_reaction', handleMessageReaction);
  chatWS.on('notification', handleNotification);
  
  chatWS.connect();
  return chatWS;
};
```

---

## 🛠 API Endpoints

### 1. Room Management

#### Get User's Chat Rooms
```javascript
// GET /api/v1/rooms/my-chats?page=1&limit=20
const getUserChatRooms = async (page = 1, limit = 20) => {
  try {
    const response = await authenticatedFetch(
      `/api/v1/rooms/my-chats?page=${page}&limit=${limit}`
    );
    const data = await response.json();
    return data.data;
  } catch (error) {
    console.error('Failed to get chat rooms:', error);
    throw error;
  }
};
```

#### Create Room
```javascript
// POST /api/v1/rooms
const createRoom = async (roomData) => {
  try {
    const response = await authenticatedFetch('/api/v1/rooms', {
      method: 'POST',
      body: JSON.stringify({
        name: roomData.name,
        description: roomData.description,
        type: roomData.type, // 'direct', 'group', 'public', 'broadcast'
        avatar: roomData.avatar,
        is_public: roomData.isPublic,
        max_members: roomData.maxMembers
      })
    });
    
    const data = await response.json();
    return data.data;
  } catch (error) {
    console.error('Failed to create room:', error);
    throw error;
  }
};
```

#### Create/Get Direct Room
```javascript
// POST /api/v1/rooms/direct/{user_id}
const createOrGetDirectRoom = async (otherUserId) => {
  try {
    const response = await authenticatedFetch(`/api/v1/rooms/direct/${otherUserId}`, {
      method: 'POST'
    });
    
    const data = await response.json();
    return data.data;
  } catch (error) {
    console.error('Failed to create/get direct room:', error);
    throw error;
  }
};
```

#### Join Room
```javascript
// POST /api/v1/rooms/{id}/join
const joinRoom = async (roomId) => {
  try {
    const response = await authenticatedFetch(`/api/v1/rooms/${roomId}/join`, {
      method: 'POST'
    });
    
    return await response.json();
  } catch (error) {
    console.error('Failed to join room:', error);
    throw error;
  }
};
```

### 2. Message Management

#### Get Room Messages
```javascript
// GET /api/v1/rooms/{room_id}/messages?page=1&limit=50
const getRoomMessages = async (roomId, page = 1, limit = 50) => {
  try {
    const response = await authenticatedFetch(
      `/api/v1/rooms/${roomId}/messages?page=${page}&limit=${limit}`
    );
    const data = await response.json();
    return data.data;
  } catch (error) {
    console.error('Failed to get room messages:', error);
    throw error;
  }
};
```

#### Send Message
```javascript
// POST /api/v1/messages
const sendMessage = async (messageData) => {
  try {
    const response = await authenticatedFetch('/api/v1/messages', {
      method: 'POST',
      body: JSON.stringify({
        room_id: messageData.roomId,
        content: messageData.content,
        type: messageData.type || 'text', // 'text', 'image', 'file', etc.
        reply_to: messageData.replyTo, // Optional: message ID to reply to
        attachments: messageData.attachments // Optional: file attachments
      })
    });
    
    const data = await response.json();
    return data.data;
  } catch (error) {
    console.error('Failed to send message:', error);
    throw error;
  }
};
```

#### Edit Message
```javascript
// PUT /api/v1/messages/{id}
const editMessage = async (messageId, newContent) => {
  try {
    const response = await authenticatedFetch(`/api/v1/messages/${messageId}`, {
      method: 'PUT',
      body: JSON.stringify({
        content: newContent
      })
    });
    
    return await response.json();
  } catch (error) {
    console.error('Failed to edit message:', error);
    throw error;
  }
};
```

#### Add Reaction
```javascript
// POST /api/v1/messages/{id}/reactions
const addReaction = async (messageId, emoji) => {
  try {
    const response = await authenticatedFetch(`/api/v1/messages/${messageId}/reactions`, {
      method: 'POST',
      body: JSON.stringify({
        emoji: emoji
      })
    });
    
    return await response.json();
  } catch (error) {
    console.error('Failed to add reaction:', error);
    throw error;
  }
};
```

### 3. Typing Indicators

#### Start Typing
```javascript
// POST /api/v1/rooms/{room_id}/typing/start
const startTyping = async (roomId) => {
  try {
    await authenticatedFetch(`/api/v1/rooms/${roomId}/typing/start`, {
      method: 'POST'
    });
  } catch (error) {
    console.error('Failed to start typing:', error);
  }
};
```

#### Stop Typing
```javascript
// POST /api/v1/rooms/{room_id}/typing/stop
const stopTyping = async (roomId) => {
  try {
    await authenticatedFetch(`/api/v1/rooms/${roomId}/typing/stop`, {
      method: 'POST'
    });
  } catch (error) {
    console.error('Failed to stop typing:', error);
  }
};
```

---

## ⚡ Real-time Events

### WebSocket Message Types

#### 1. Message Events
```javascript
// New message received
chatWS.on('message', (data) => {
  console.log('New message:', data);
  // Update UI with new message
  addMessageToUI(data);
});

// Message edited
chatWS.on('message_edit', (data) => {
  console.log('Message edited:', data);
  updateMessageInUI(data.message_id, data);
});

// Message deleted
chatWS.on('message_delete', (data) => {
  console.log('Message deleted:', data);
  removeMessageFromUI(data.message_id);
});

// Message reaction added/removed
chatWS.on('message_reaction', (data) => {
  console.log('Message reaction:', data);
  updateMessageReactions(data.message_id, data);
});
```

#### 2. Typing Events
```javascript
// User started typing
chatWS.on('typing_start', (data) => {
  console.log('User started typing:', data);
  showTypingIndicator(data.user_id, data.room_id);
});

// User stopped typing
chatWS.on('typing_stop', (data) => {
  console.log('User stopped typing:', data);
  hideTypingIndicator(data.user_id, data.room_id);
});
```

#### 3. Room Events
```javascript
// User joined room
chatWS.on('user_join', (data) => {
  console.log('User joined:', data);
  addUserToRoomUI(data.user_id, data.room_id);
});

// User left room
chatWS.on('user_leave', (data) => {
  console.log('User left:', data);
  removeUserFromRoomUI(data.user_id, data.room_id);
});
```

#### 4. User Status Events
```javascript
// User online/offline status
chatWS.on('user_status_change', (data) => {
  console.log('User status changed:', data);
  updateUserStatus(data.user_id, data.status);
});
```

#### 5. System Events
```javascript
// Notifications
chatWS.on('notification', (data) => {
  console.log('Notification:', data);
  showNotification(data);
});

// Errors
chatWS.on('error', (data) => {
  console.error('WebSocket error:', data);
  handleWebSocketError(data);
});
```

---

## 💬 Chat Implementation

### 1. Chat Room Component (React Example)

```javascript
import React, { useState, useEffect, useRef } from 'react';

const ChatRoom = ({ roomId, chatWS }) => {
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState('');
  const [typingUsers, setTypingUsers] = useState(new Set());
  const [isLoading, setIsLoading] = useState(true);
  const messagesEndRef = useRef(null);
  const typingTimeoutRef = useRef(null);

  useEffect(() => {
    loadMessages();
    setupWebSocketHandlers();
    
    return () => {
      // Cleanup
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
      }
    };
  }, [roomId]);

  const loadMessages = async () => {
    try {
      setIsLoading(true);
      const messagesData = await getRoomMessages(roomId);
      setMessages(messagesData.messages);
    } catch (error) {
      console.error('Failed to load messages:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const setupWebSocketHandlers = () => {
    // Handle new messages
    chatWS.on('message', (data) => {
      if (data.room_id === roomId) {
        setMessages(prev => [...prev, data]);
        scrollToBottom();
      }
    });

    // Handle typing indicators
    chatWS.on('typing_start', (data) => {
      if (data.room_id === roomId) {
        setTypingUsers(prev => new Set([...prev, data.user_id]));
      }
    });

    chatWS.on('typing_stop', (data) => {
      if (data.room_id === roomId) {
        setTypingUsers(prev => {
          const newSet = new Set(prev);
          newSet.delete(data.user_id);
          return newSet;
        });
      }
    });

    // Handle message edits
    chatWS.on('message_edit', (data) => {
      if (data.room_id === roomId) {
        setMessages(prev => 
          prev.map(msg => 
            msg.id === data.message_id 
              ? { ...msg, content: data.content, edited_at: data.edited_at }
              : msg
          )
        );
      }
    });

    // Handle message deletions
    chatWS.on('message_delete', (data) => {
      if (data.room_id === roomId) {
        setMessages(prev => 
          prev.filter(msg => msg.id !== data.message_id)
        );
      }
    });
  };

  const handleSendMessage = async (e) => {
    e.preventDefault();
    
    if (!newMessage.trim()) return;

    try {
      await sendMessage({
        roomId,
        content: newMessage.trim(),
        type: 'text'
      });
      
      setNewMessage('');
      await stopTyping(roomId);
    } catch (error) {
      console.error('Failed to send message:', error);
    }
  };

  const handleTyping = async (e) => {
    setNewMessage(e.target.value);

    // Start typing indicator
    if (e.target.value.length === 1) {
      await startTyping(roomId);
    }

    // Reset typing timeout
    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current);
    }

    // Stop typing after 3 seconds of inactivity
    typingTimeoutRef.current = setTimeout(async () => {
      await stopTyping(roomId);
    }, 3000);
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  return (
    <div className="chat-room">
      <div className="messages-container">
        {isLoading ? (
          <div>Loading messages...</div>
        ) : (
          messages.map(message => (
            <MessageComponent 
              key={message.id} 
              message={message}
              onReaction={(emoji) => addReaction(message.id, emoji)}
            />
          ))
        )}
        
        {typingUsers.size > 0 && (
          <TypingIndicator users={Array.from(typingUsers)} />
        )}
        
        <div ref={messagesEndRef} />
      </div>

      <form onSubmit={handleSendMessage} className="message-input-form">
        <input
          type="text"
          value={newMessage}
          onChange={handleTyping}
          placeholder="Type a message..."
          className="message-input"
        />
        <button type="submit" disabled={!newMessage.trim()}>
          Send
        </button>
      </form>
    </div>
  );
};
```

### 2. Chat List Component

```javascript
const ChatList = ({ onRoomSelect }) => {
  const [rooms, setRooms] = useState([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    loadChatRooms();
  }, []);

  const loadChatRooms = async () => {
    try {
      setIsLoading(true);
      const roomsData = await getUserChatRooms();
      setRooms(roomsData.rooms);
    } catch (error) {
      console.error('Failed to load chat rooms:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const startDirectChat = async (userId) => {
    try {
      const room = await createOrGetDirectRoom(userId);
      onRoomSelect(room.id);
    } catch (error) {
      console.error('Failed to start direct chat:', error);
    }
  };

  return (
    <div className="chat-list">
      <h3>Chats</h3>
      
      {isLoading ? (
        <div>Loading chats...</div>
      ) : (
        <div className="rooms-list">
          {rooms.map(room => (
            <div 
              key={room.id} 
              className="room-item"
              onClick={() => onRoomSelect(room.id)}
            >
              <div className="room-avatar">
                {room.avatar ? (
                  <img src={room.avatar} alt={room.name} />
                ) : (
                  <div className="avatar-placeholder">
                    {room.name?.charAt(0) || '?'}
                  </div>
                )}
              </div>
              
              <div className="room-info">
                <div className="room-name">{room.name || 'Unknown'}</div>
                <div className="room-last-message">
                  {room.last_message?.content || 'No messages yet'}
                </div>
              </div>
              
              <div className="room-meta">
                <div className="room-time">
                  {formatTime(room.updated_at)}
                </div>
                {room.unread_count > 0 && (
                  <div className="unread-badge">
                    {room.unread_count}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
```

---

## 🚨 Error Handling

### 1. API Error Handling
```javascript
const handleAPIError = (error, context) => {
  console.error(`${context} error:`, error);
  
  if (error.status === 401) {
    // Token expired, redirect to login
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    window.location.href = '/login';
  } else if (error.status === 403) {
    // Forbidden
    showErrorNotification('You don\'t have permission to perform this action');
  } else if (error.status === 404) {
    // Not found
    showErrorNotification('Resource not found');
  } else if (error.status >= 500) {
    // Server error
    showErrorNotification('Server error. Please try again later.');
  } else {
    // Other errors
    showErrorNotification(error.message || 'An unexpected error occurred');
  }
};
```

### 2. WebSocket Error Handling
```javascript
const handleWebSocketError = (error) => {
  console.error('WebSocket error:', error);
  
  // Show connection status to user
  showConnectionStatus('disconnected');
  
  // Try to reconnect
  setTimeout(() => {
    if (chatWS) {
      chatWS.connect();
    }
  }, 5000);
};
```

---

## ✨ Best Practices

### 1. Performance Optimization

#### Message Virtualization
```javascript
// For large message lists, use virtualization
import { VariableSizeList as List } from 'react-window';

const VirtualizedMessageList = ({ messages }) => {
  const getItemSize = (index) => {
    // Calculate message height based on content
    return calculateMessageHeight(messages[index]);
  };

  return (
    <List
      height={400}
      itemCount={messages.length}
      itemSize={getItemSize}
      itemData={messages}
    >
      {MessageItem}
    </List>
  );
};
```

#### Message Pagination
```javascript
const useInfiniteMessages = (roomId) => {
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [page, setPage] = useState(1);

  const loadMoreMessages = async () => {
    if (loading || !hasMore) return;

    try {
      setLoading(true);
      const data = await getRoomMessages(roomId, page, 50);
      
      if (data.messages.length === 0) {
        setHasMore(false);
      } else {
        setMessages(prev => [...data.messages, ...prev]);
        setPage(prev => prev + 1);
      }
    } catch (error) {
      console.error('Failed to load more messages:', error);
    } finally {
      setLoading(false);
    }
  };

  return { messages, loadMoreMessages, loading, hasMore };
};
```

### 2. State Management

#### Using Redux/Zustand for Global State
```javascript
// Zustand store example
import { create } from 'zustand';

const useChatStore = create((set, get) => ({
  // State
  rooms: [],
  currentRoomId: null,
  messages: {},
  typingUsers: {},
  onlineUsers: new Set(),
  
  // Actions
  setRooms: (rooms) => set({ rooms }),
  
  setCurrentRoom: (roomId) => set({ currentRoomId: roomId }),
  
  addMessage: (message) => set((state) => ({
    messages: {
      ...state.messages,
      [message.room_id]: [
        ...(state.messages[message.room_id] || []),
        message
      ]
    }
  })),
  
  updateMessage: (roomId, messageId, updates) => set((state) => ({
    messages: {
      ...state.messages,
      [roomId]: state.messages[roomId]?.map(msg =>
        msg.id === messageId ? { ...msg, ...updates } : msg
      ) || []
    }
  })),
  
  setTypingUsers: (roomId, users) => set((state) => ({
    typingUsers: {
      ...state.typingUsers,
      [roomId]: users
    }
  })),
  
  setUserOnline: (userId) => set((state) => ({
    onlineUsers: new Set([...state.onlineUsers, userId])
  })),
  
  setUserOffline: (userId) => set((state) => {
    const newOnlineUsers = new Set(state.onlineUsers);
    newOnlineUsers.delete(userId);
    return { onlineUsers: newOnlineUsers };
  })
}));
```

### 3. Security Best Practices

#### Token Refresh
```javascript
const refreshToken = async () => {
  try {
    const refreshToken = localStorage.getItem('refreshToken');
    const response = await fetch('/api/v1/auth/refresh', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${refreshToken}`
      }
    });
    
    const data = await response.json();
    
    if (data.success) {
      localStorage.setItem('token', data.data.token);
      return data.data.token;
    }
  } catch (error) {
    console.error('Token refresh failed:', error);
    // Redirect to login
    window.location.href = '/login';
  }
};

// Auto-refresh token before expiry
const setupTokenRefresh = () => {
  const token = localStorage.getItem('token');
  if (token) {
    const payload = JSON.parse(atob(token.split('.')[1]));
    const expiryTime = payload.exp * 1000;
    const currentTime = Date.now();
    const timeUntilExpiry = expiryTime - currentTime;
    
    // Refresh 5 minutes before expiry
    const refreshTime = timeUntilExpiry - (5 * 60 * 1000);
    
    if (refreshTime > 0) {
      setTimeout(refreshToken, refreshTime);
    }
  }
};
```

#### Input Sanitization
```javascript
const sanitizeMessage = (content) => {
  // Remove potentially harmful content
  return content
    .replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '')
    .replace(/javascript:/gi, '')
    .trim();
};
```

### 4. Accessibility

#### Keyboard Navigation
```javascript
const MessageInput = ({ onSend }) => {
  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      onSend();
    }
  };

  return (
    <textarea
      placeholder="Type a message... (Enter to send, Shift+Enter for new line)"
      onKeyDown={handleKeyDown}
      aria-label="Message input"
      role="textbox"
      aria-multiline="true"
    />
  );
};
```

#### Screen Reader Support
```javascript
const Message = ({ message, isOwn }) => {
  return (
    <div
      className={`message ${isOwn ? 'own' : 'other'}`}
      role="article"
      aria-label={`Message from ${message.user.username} at ${formatTime(message.created_at)}`}
    >
      <div className="message-author" aria-hidden="true">
        {message.user.username}
      </div>
      <div className="message-content">
        {message.content}
      </div>
      <div className="message-time" aria-hidden="true">
        {formatTime(message.created_at)}
      </div>
    </div>
  );
};
```

---

## 🔧 Environment Configuration

### Development Environment
```javascript
// config/development.js
export const config = {
  API_BASE_URL: 'http://localhost:8080/api/v1',
  WS_URL: 'ws://localhost:8080/ws',
  DEBUG: true,
  RECONNECT_INTERVAL: 3000,
  MESSAGE_PAGINATION_SIZE: 50,
  TYPING_TIMEOUT: 3000
};
```

### Production Environment
```javascript
// config/production.js
export const config = {
  API_BASE_URL: 'https://api.yourapp.com/api/v1',
  WS_URL: 'wss://api.yourapp.com/ws',
  DEBUG: false,
  RECONNECT_INTERVAL: 5000,
  MESSAGE_PAGINATION_SIZE: 50,
  TYPING_TIMEOUT: 3000
};
```

---

## 📱 Mobile Considerations

### React Native WebSocket
```javascript
// Install: npm install react-native-websocket
import WebSocket from 'react-native-websocket';

const ChatWebSocketRN = () => {
  return (
    <WebSocket
      url={`${WS_URL}?token=${token}`}
      onOpen={() => console.log('WebSocket connected')}
      onMessage={(message) => handleMessage(JSON.parse(message.data))}
      onError={(error) => console.error('WebSocket error:', error)}
      onClose={() => console.log('WebSocket disconnected')}
      reconnect={true}
      reconnectIntervalInMilliSeconds={5000}
    />
  );
};
```

### Handling Background/Foreground
```javascript
import { AppState } from 'react-native';

const useAppStateWebSocket = (chatWS) => {
  useEffect(() => {
    const handleAppStateChange = (nextAppState) => {
      if (nextAppState === 'background') {
        // App going to background
        chatWS?.disconnect();
      } else if (nextAppState === 'active') {
        // App coming to foreground
        chatWS?.connect();
      }
    };

    const subscription = AppState.addEventListener(
      'change',
      handleAppStateChange
    );

    return () => subscription?.remove();
  }, [chatWS]);
};
```

---

## 🧪 Testing

### Unit Testing WebSocket
```javascript
// __tests__/websocket.test.js
import { ChatWebSocket } from '../src/websocket';

// Mock WebSocket
global.WebSocket = jest.fn(() => ({
  send: jest.fn(),
  close: jest.fn(),
  addEventListener: jest.fn(),
  removeEventListener: jest.fn(),
  readyState: 1
}));

describe('ChatWebSocket', () => {
  let chatWS;

  beforeEach(() => {
    chatWS = new ChatWebSocket('test-token');
  });

  test('should connect with token', () => {
    chatWS.connect();
    expect(global.WebSocket).toHaveBeenCalledWith(
      'ws://localhost:8080/ws?token=test-token'
    );
  });

  test('should handle message events', () => {
    const handler = jest.fn();
    chatWS.on('message', handler);
    
    chatWS.handleMessage({
      type: 'message',
      data: { content: 'test message' }
    });
    
    expect(handler).toHaveBeenCalledWith({
      type: 'message',
      data: { content: 'test message' }
    });
  });
});
```

---

## 📋 Troubleshooting

### Common Issues

#### 1. WebSocket Connection Failed
```javascript
// Check if server is running
// Verify token is valid
// Check CORS settings
// Ensure proper protocol (ws:// for HTTP, wss:// for HTTPS)
```

#### 2. Messages Not Updating
```javascript
// Verify event handlers are registered
// Check if user is authenticated
// Ensure proper room membership
// Check browser console for errors
```

#### 3. Authentication Issues
```javascript
// Verify token format and expiry
// Check if token is included in requests
// Ensure user has proper permissions
// Check server logs for auth errors
```

---

## 📚 Complete Integration Example

```javascript
// App.js - Complete integration example
import React, { useState, useEffect } from 'react';
import { ChatWebSocket } from './websocket';
import ChatList from './components/ChatList';
import ChatRoom from './components/ChatRoom';

const App = () => {
  const [user, setUser] = useState(null);
  const [token, setToken] = useState(null);
  const [chatWS, setChatWS] = useState(null);
  const [currentRoomId, setCurrentRoomId] = useState(null);

  useEffect(() => {
    // Check for stored auth
    const storedToken = localStorage.getItem('token');
    const storedUser = localStorage.getItem('user');
    
    if (storedToken && storedUser) {
      setToken(storedToken);
      setUser(JSON.parse(storedUser));
      initializeWebSocket(storedToken);
    }
  }, []);

  const initializeWebSocket = (authToken) => {
    const ws = new ChatWebSocket(authToken);
    
    // Setup global event handlers
    ws.on('notification', (data) => {
      showNotification(data.message);
    });
    
    ws.on('error', (error) => {
      console.error('WebSocket error:', error);
    });
    
    ws.connect();
    setChatWS(ws);
  };

  const handleLogin = async (credentials) => {
    try {
      const result = await loginUser(credentials.username, credentials.password);
      setUser(result.user);
      setToken(result.token);
      initializeWebSocket(result.token);
    } catch (error) {
      console.error('Login failed:', error);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    chatWS?.disconnect();
    setUser(null);
    setToken(null);
    setChatWS(null);
    setCurrentRoomId(null);
  };

  if (!user || !token) {
    return <LoginForm onLogin={handleLogin} />;
  }

  return (
    <div className="app">
      <div className="app-header">
        <h1>Chat App</h1>
        <div className="user-info">
          Welcome, {user.username}
          <button onClick={handleLogout}>Logout</button>
        </div>
      </div>
      
      <div className="app-content">
        <div className="sidebar">
          <ChatList onRoomSelect={setCurrentRoomId} />
        </div>
        
        <div className="main-content">
          {currentRoomId ? (
            <ChatRoom 
              roomId={currentRoomId} 
              chatWS={chatWS}
              currentUser={user}
            />
          ) : (
            <div className="no-chat-selected">
              Select a chat to start messaging
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default App;
```

---

## 🎉 Conclusion

This guide provides complete integration instructions for building a real-time chat application with our API. Key takeaways:

1. **Authentication** using JWT tokens
2. **WebSocket** for real-time communication
3. **RESTful API** for CRUD operations
4. **Event-driven** updates for instant UI changes
5. **Error handling** and reconnection logic
6. **Performance optimization** for large chat lists
7. **Accessibility** and mobile support

For additional support or questions, please refer to the API documentation or contact our development team.

---

**Happy coding! 🚀**
