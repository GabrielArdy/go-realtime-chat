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

#### Register Request
```json
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "securePassword123",
  "first_name": "John",
  "last_name": "Doe",
  "phone_number": "+1234567890",
  "bio": "Software Developer"
}
```

#### Register Response (Success)
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "user": {
      "id": "uuid-string",
      "username": "john_doe",
      "email": "john@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "avatar": "",
      "phone_number": "+1234567890",
      "bio": "Software Developer",
      "status": "offline",
      "last_seen": null,
      "is_active": true,
      "is_verified": false,
      "language": "en",
      "timezone": "UTC",
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2023-01-01T01:00:00Z",
    "session_id": "uuid-string"
  }
}
```

#### Register Response (Error - Email Exists)
```json
{
  "success": false,
  "message": "Email address is already registered"
}
```

#### Register Response (Error - Username Taken)
```json
{
  "success": false,
  "message": "Username is already taken"
}
```

#### Register Response (Error - Validation)
```json
{
  "success": false,
  "message": "First name is required"
}
```

#### Login Request
```json
{
  "email": "john@example.com",
  "password": "securePassword123",
  "device_id": "browser-uuid-or-identifier",
  "device_type": "web"
}
```

#### Login Response (Success)
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": "uuid-string",
      "username": "john_doe",
      "email": "john@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "avatar": "https://example.com/avatar.jpg",
      "phone_number": "+1234567890",
      "bio": "Software Developer",
      "status": "online",
      "last_seen": "2023-01-01T00:00:00Z",
      "is_active": true,
      "is_verified": true,
      "language": "en",
      "timezone": "UTC",
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2023-01-01T01:00:00Z",
    "session_id": "uuid-string"
  }
}
```

#### Login Response (Error - Invalid Credentials)
```json
{
  "success": false,
  "message": "Authentication failed",
  "error": "Invalid credentials"
}
```

#### Login Response (Error - Account Inactive)
```json
{
  "success": false,
  "message": "Authentication failed",
  "error": "User account is inactive"
}
```

#### Refresh Token Request
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Refresh Token Response
```json
{
  "success": true,
  "message": "Token refreshed successfully",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2023-01-01T02:00:00Z"
  }
}
```

#### Logout Request
```json
{
  "session_id": "uuid-string"
}
```

#### Logout Response
```json
{
  "success": true,
  "message": "Logout successful"
}
```

### User Schemas

#### User Object
```json
{
  "id": "uuid-string",
  "username": "john_doe",
  "email": "john@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "avatar": "https://example.com/avatar.jpg",
  "phone_number": "+1234567890",
  "bio": "Software Developer passionate about real-time applications",
  "status": "online",
  "last_seen": "2023-01-01T00:00:00Z",
  "is_active": true,
  "is_verified": true,
  "language": "en",
  "timezone": "UTC",
  "notification_sound": true,
  "email_notifications": true,
  "push_notifications": true,
  "show_online_status": true,
  "show_read_receipts": true,
  "allow_direct_messages": true,
  "auto_join_public_rooms": false,
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z"
}
```

#### Update User Profile Request
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "bio": "Updated bio text",
  "avatar": "https://example.com/new-avatar.jpg",
  "phone_number": "+1234567890",
  "status": "online"
}
```

#### Update User Settings Request
```json
{
  "language": "en",
  "timezone": "America/New_York",
  "notification_sound": true,
  "email_notifications": false,
  "push_notifications": true,
  "show_online_status": true,
  "show_read_receipts": false,
  "allow_direct_messages": true,
  "auto_join_public_rooms": false
}
```
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

### 1. User Registration
```javascript
// POST /api/v1/auth/register
const registerUser = async (userData) => {
  try {
    const response = await fetch('/api/v1/auth/register', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        username: userData.username,
        email: userData.email,
        password: userData.password,
        first_name: userData.firstName,
        last_name: userData.lastName,
        phone_number: userData.phoneNumber, // optional
        bio: userData.bio // optional
      })
    });
    
    const data = await response.json();
    
    if (data.success) {
      // Store JWT tokens
      localStorage.setItem('access_token', data.data.access_token);
      localStorage.setItem('refresh_token', data.data.refresh_token);
      localStorage.setItem('user', JSON.stringify(data.data.user));
      localStorage.setItem('session_id', data.data.session_id);
      return data.data;
    } else {
      throw new Error(data.message);
    }
  } catch (error) {
    console.error('Registration failed:', error);
    throw error;
  }
};

// Example usage
const handleRegistration = async (formData) => {
  try {
    const result = await registerUser({
      username: formData.username,
      email: formData.email,
      password: formData.password,
      firstName: formData.firstName,
      lastName: formData.lastName,
      phoneNumber: formData.phoneNumber,
      bio: formData.bio
    });
    
    // User is automatically logged in after registration
    console.log('Registration successful:', result);
    // Redirect to dashboard or initialize app
    initializeApp(result);
  } catch (error) {
    // Handle specific errors
    if (error.message === 'Email address is already registered') {
      showError('This email is already registered. Please try logging in instead.');
    } else if (error.message === 'Username is already taken') {
      showError('This username is already taken. Please choose another one.');
    } else {
      showError('Registration failed. Please try again.');
    }
  }
};
```

### 2. User Login
```javascript
// POST /api/v1/auth/login
const loginUser = async (email, password, deviceId = null) => {
  try {
    // Generate device ID if not provided
    const actualDeviceId = deviceId || generateDeviceId();
    
    const response = await fetch('/api/v1/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        email,
        password,
        device_id: actualDeviceId,
        device_type: getDeviceType() // 'web', 'mobile', 'desktop'
      })
    });
    
    const data = await response.json();
    
    if (data.success) {
      // Store JWT tokens and user data
      localStorage.setItem('access_token', data.data.access_token);
      localStorage.setItem('refresh_token', data.data.refresh_token);
      localStorage.setItem('user', JSON.stringify(data.data.user));
      localStorage.setItem('session_id', data.data.session_id);
      localStorage.setItem('device_id', actualDeviceId);
      return data.data;
    } else {
      throw new Error(data.message || data.error);
    }
  } catch (error) {
    console.error('Login failed:', error);
    throw error;
  }
};

// Helper functions
const generateDeviceId = () => {
  return 'web_' + Math.random().toString(36).substr(2, 9) + '_' + Date.now();
};

const getDeviceType = () => {
  if (typeof window !== 'undefined') {
    if (window.navigator.userAgent.includes('Mobile')) {
      return 'mobile';
    }
    return 'web';
  }
  return 'unknown';
};
```

### 3. Token Management
```javascript
// Get stored tokens for API calls
const getAuthTokens = () => {
  return {
    accessToken: localStorage.getItem('access_token'),
    refreshToken: localStorage.getItem('refresh_token'),
    sessionId: localStorage.getItem('session_id')
  };
};

// Check if user is authenticated
const isAuthenticated = () => {
  const { accessToken } = getAuthTokens();
  if (!accessToken) return false;
  
  try {
    const payload = JSON.parse(atob(accessToken.split('.')[1]));
    const expiryTime = payload.exp * 1000;
    return Date.now() < expiryTime;
  } catch (error) {
    return false;
  }
};

// Refresh access token
const refreshAccessToken = async () => {
  try {
    const { refreshToken } = getAuthTokens();
    if (!refreshToken) {
      throw new Error('No refresh token available');
    }
    
    const response = await fetch('/api/v1/auth/refresh', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${refreshToken}`
      }
    });
    
    const data = await response.json();
    
    if (data.success) {
      localStorage.setItem('access_token', data.data.access_token);
      localStorage.setItem('refresh_token', data.data.refresh_token);
      return data.data.access_token;
    } else {
      throw new Error('Token refresh failed');
    }
  } catch (error) {
    console.error('Token refresh failed:', error);
    // Clear tokens and redirect to login
    clearAuthData();
    window.location.href = '/login';
    throw error;
  }
};

// Create authenticated fetch wrapper
const authenticatedFetch = async (url, options = {}) => {
  let { accessToken } = getAuthTokens();
  
  // Check if token needs refresh
  if (!isAuthenticated()) {
    try {
      accessToken = await refreshAccessToken();
    } catch (error) {
      throw new Error('Authentication required');
    }
  }
  
  return fetch(url, {
    ...options,
    headers: {
      ...options.headers,
      'Authorization': `Bearer ${accessToken}`,
      'Content-Type': 'application/json',
    }
  });
};
```

### 4. Logout
```javascript
// POST /api/v1/auth/logout
const logoutUser = async () => {
  try {
    const { sessionId } = getAuthTokens();
    
    if (sessionId) {
      await authenticatedFetch('/api/v1/auth/logout', {
        method: 'POST',
        body: JSON.stringify({
          session_id: sessionId
        })
      });
    }
  } catch (error) {
    console.error('Logout request failed:', error);
    // Continue with local cleanup even if server request fails
  } finally {
    // Clear local storage
    clearAuthData();
    // Redirect to login
    window.location.href = '/login';
  }
};

const clearAuthData = () => {
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
  localStorage.removeItem('user');
  localStorage.removeItem('session_id');
  localStorage.removeItem('device_id');
};
```

### 5. Auto-refresh Token Setup
```javascript
// Setup automatic token refresh
const setupTokenRefresh = () => {
  const { accessToken } = getAuthTokens();
  if (!accessToken) return;
  
  try {
    const payload = JSON.parse(atob(accessToken.split('.')[1]));
    const expiryTime = payload.exp * 1000;
    const currentTime = Date.now();
    const timeUntilExpiry = expiryTime - currentTime;
    
    // Refresh 5 minutes before expiry
    const refreshTime = timeUntilExpiry - (5 * 60 * 1000);
    
    if (refreshTime > 0) {
      setTimeout(async () => {
        try {
          await refreshAccessToken();
          setupTokenRefresh(); // Setup next refresh
        } catch (error) {
          console.error('Auto token refresh failed:', error);
        }
      }, refreshTime);
    }
  } catch (error) {
    console.error('Failed to setup token refresh:', error);
  }
};

// Call this after login or app initialization
setupTokenRefresh();
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

## � Authentication Examples

### 1. Registration Form Component (React)

```javascript
import React, { useState } from 'react';

const RegistrationForm = ({ onRegistrationSuccess }) => {
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
    firstName: '',
    lastName: '',
    phoneNumber: '',
    bio: ''
  });
  
  const [errors, setErrors] = useState({});
  const [isLoading, setIsLoading] = useState(false);

  const validateForm = () => {
    const newErrors = {};

    // Required fields
    if (!formData.username.trim()) {
      newErrors.username = 'Username is required';
    } else if (formData.username.length < 3) {
      newErrors.username = 'Username must be at least 3 characters';
    }

    if (!formData.email.trim()) {
      newErrors.email = 'Email is required';
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = 'Email is invalid';
    }

    if (!formData.password) {
      newErrors.password = 'Password is required';
    } else if (formData.password.length < 6) {
      newErrors.password = 'Password must be at least 6 characters';
    }

    if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match';
    }

    if (!formData.firstName.trim()) {
      newErrors.firstName = 'First name is required';
    }

    if (!formData.lastName.trim()) {
      newErrors.lastName = 'Last name is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!validateForm()) return;

    setIsLoading(true);
    setErrors({});

    try {
      const result = await registerUser({
        username: formData.username,
        email: formData.email,
        password: formData.password,
        firstName: formData.firstName,
        lastName: formData.lastName,
        phoneNumber: formData.phoneNumber || undefined,
        bio: formData.bio || undefined
      });

      onRegistrationSuccess(result);
    } catch (error) {
      if (error.message === 'Email address is already registered') {
        setErrors({ email: 'This email is already registered' });
      } else if (error.message === 'Username is already taken') {
        setErrors({ username: 'This username is already taken' });
      } else {
        setErrors({ general: 'Registration failed. Please try again.' });
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
    
    // Clear error when user starts typing
    if (errors[name]) {
      setErrors(prev => ({
        ...prev,
        [name]: ''
      }));
    }
  };

  return (
    <form onSubmit={handleSubmit} className="registration-form">
      <h2>Create Account</h2>
      
      {errors.general && (
        <div className="error-banner">{errors.general}</div>
      )}

      <div className="form-row">
        <div className="form-group">
          <label htmlFor="firstName">First Name *</label>
          <input
            type="text"
            id="firstName"
            name="firstName"
            value={formData.firstName}
            onChange={handleChange}
            className={errors.firstName ? 'error' : ''}
            disabled={isLoading}
          />
          {errors.firstName && <span className="error">{errors.firstName}</span>}
        </div>

        <div className="form-group">
          <label htmlFor="lastName">Last Name *</label>
          <input
            type="text"
            id="lastName"
            name="lastName"
            value={formData.lastName}
            onChange={handleChange}
            className={errors.lastName ? 'error' : ''}
            disabled={isLoading}
          />
          {errors.lastName && <span className="error">{errors.lastName}</span>}
        </div>
      </div>

      <div className="form-group">
        <label htmlFor="username">Username *</label>
        <input
          type="text"
          id="username"
          name="username"
          value={formData.username}
          onChange={handleChange}
          className={errors.username ? 'error' : ''}
          disabled={isLoading}
          placeholder="Choose a unique username"
        />
        {errors.username && <span className="error">{errors.username}</span>}
      </div>

      <div className="form-group">
        <label htmlFor="email">Email *</label>
        <input
          type="email"
          id="email"
          name="email"
          value={formData.email}
          onChange={handleChange}
          className={errors.email ? 'error' : ''}
          disabled={isLoading}
          placeholder="your@email.com"
        />
        {errors.email && <span className="error">{errors.email}</span>}
      </div>

      <div className="form-row">
        <div className="form-group">
          <label htmlFor="password">Password *</label>
          <input
            type="password"
            id="password"
            name="password"
            value={formData.password}
            onChange={handleChange}
            className={errors.password ? 'error' : ''}
            disabled={isLoading}
            placeholder="At least 6 characters"
          />
          {errors.password && <span className="error">{errors.password}</span>}
        </div>

        <div className="form-group">
          <label htmlFor="confirmPassword">Confirm Password *</label>
          <input
            type="password"
            id="confirmPassword"
            name="confirmPassword"
            value={formData.confirmPassword}
            onChange={handleChange}
            className={errors.confirmPassword ? 'error' : ''}
            disabled={isLoading}
            placeholder="Repeat your password"
          />
          {errors.confirmPassword && <span className="error">{errors.confirmPassword}</span>}
        </div>
      </div>

      <div className="form-group">
        <label htmlFor="phoneNumber">Phone Number</label>
        <input
          type="tel"
          id="phoneNumber"
          name="phoneNumber"
          value={formData.phoneNumber}
          onChange={handleChange}
          disabled={isLoading}
          placeholder="+1234567890 (optional)"
        />
      </div>

      <div className="form-group">
        <label htmlFor="bio">Bio</label>
        <textarea
          id="bio"
          name="bio"
          value={formData.bio}
          onChange={handleChange}
          disabled={isLoading}
          placeholder="Tell us about yourself (optional)"
          rows="3"
        />
      </div>

      <button 
        type="submit" 
        disabled={isLoading}
        className="submit-button"
      >
        {isLoading ? 'Creating Account...' : 'Create Account'}
      </button>

      <p className="login-link">
        Already have an account? <a href="/login">Sign in here</a>
      </p>
    </form>
  );
};

export default RegistrationForm;
```

### 2. Login Form Component (React)

```javascript
import React, { useState } from 'react';

const LoginForm = ({ onLoginSuccess }) => {
  const [formData, setFormData] = useState({
    email: '',
    password: '',
    rememberMe: false
  });
  
  const [errors, setErrors] = useState({});
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!formData.email || !formData.password) {
      setErrors({ general: 'Please fill in all fields' });
      return;
    }

    setIsLoading(true);
    setErrors({});

    try {
      const result = await loginUser(formData.email, formData.password);
      onLoginSuccess(result);
    } catch (error) {
      if (error.message === 'Authentication failed' || error.message === 'Invalid credentials') {
        setErrors({ general: 'Invalid email or password' });
      } else if (error.message === 'User account is inactive') {
        setErrors({ general: 'Your account has been deactivated. Please contact support.' });
      } else {
        setErrors({ general: 'Login failed. Please try again.' });
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }));
    
    // Clear errors when user starts typing
    if (errors.general) {
      setErrors({});
    }
  };

  return (
    <form onSubmit={handleSubmit} className="login-form">
      <h2>Sign In</h2>
      
      {errors.general && (
        <div className="error-banner">{errors.general}</div>
      )}

      <div className="form-group">
        <label htmlFor="email">Email</label>
        <input
          type="email"
          id="email"
          name="email"
          value={formData.email}
          onChange={handleChange}
          disabled={isLoading}
          placeholder="your@email.com"
          required
        />
      </div>

      <div className="form-group">
        <label htmlFor="password">Password</label>
        <input
          type="password"
          id="password"
          name="password"
          value={formData.password}
          onChange={handleChange}
          disabled={isLoading}
          placeholder="Your password"
          required
        />
      </div>

      <div className="form-group checkbox-group">
        <label className="checkbox-label">
          <input
            type="checkbox"
            name="rememberMe"
            checked={formData.rememberMe}
            onChange={handleChange}
            disabled={isLoading}
          />
          Remember me
        </label>
      </div>

      <button 
        type="submit" 
        disabled={isLoading || !formData.email || !formData.password}
        className="submit-button"
      >
        {isLoading ? 'Signing In...' : 'Sign In'}
      </button>

      <div className="form-links">
        <a href="/forgot-password">Forgot password?</a>
        <span>•</span>
        <a href="/register">Create account</a>
      </div>
    </form>
  );
};

export default LoginForm;
```

### 3. Authentication Context Provider (React)

```javascript
import React, { createContext, useContext, useState, useEffect } from 'react';

const AuthContext = createContext();

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  useEffect(() => {
    checkAuthStatus();
  }, []);

  const checkAuthStatus = async () => {
    try {
      const storedUser = localStorage.getItem('user');
      const { accessToken } = getAuthTokens();

      if (storedUser && accessToken && isTokenValid(accessToken)) {
        setUser(JSON.parse(storedUser));
        setIsAuthenticated(true);
        setupTokenRefresh();
      } else {
        clearAuthData();
      }
    } catch (error) {
      console.error('Auth check failed:', error);
      clearAuthData();
    } finally {
      setLoading(false);
    }
  };

  const isTokenValid = (token) => {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      return Date.now() < payload.exp * 1000;
    } catch (error) {
      return false;
    }
  };

  const login = async (email, password) => {
    setLoading(true);
    try {
      const result = await loginUser(email, password);
      setUser(result.user);
      setIsAuthenticated(true);
      setupTokenRefresh();
      return result;
    } catch (error) {
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const register = async (userData) => {
    setLoading(true);
    try {
      const result = await registerUser(userData);
      setUser(result.user);
      setIsAuthenticated(true);
      setupTokenRefresh();
      return result;
    } catch (error) {
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const logout = async () => {
    setLoading(true);
    try {
      await logoutUser();
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      setUser(null);
      setIsAuthenticated(false);
      setLoading(false);
    }
  };

  const updateUser = (updatedUser) => {
    setUser(updatedUser);
    localStorage.setItem('user', JSON.stringify(updatedUser));
  };

  const value = {
    user,
    loading,
    isAuthenticated,
    login,
    register,
    logout,
    updateUser,
    checkAuthStatus
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};
```

### 4. Protected Route Component

```javascript
import React from 'react';
import { useAuth } from './AuthContext';

const ProtectedRoute = ({ children, fallback = null }) => {
  const { isAuthenticated, loading } = useAuth();

  if (loading) {
    return (
      <div className="loading-spinner">
        <div>Loading...</div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return fallback || <LoginForm />;
  }

  return children;
};

export default ProtectedRoute;
```

### 5. Authentication Form Styling (CSS)

```css
/* Authentication Forms Styling */
.auth-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.registration-form,
.login-form {
  background: white;
  padding: 40px;
  border-radius: 10px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
  width: 100%;
  max-width: 500px;
}

.registration-form h2,
.login-form h2 {
  text-align: center;
  margin-bottom: 30px;
  color: #333;
  font-size: 28px;
  font-weight: 600;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 15px;
  margin-bottom: 15px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  color: #555;
  font-weight: 500;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 12px;
  border: 2px solid #e1e5e9;
  border-radius: 6px;
  font-size: 16px;
  transition: border-color 0.3s ease;
  box-sizing: border-box;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #667eea;
}

.form-group input.error,
.form-group textarea.error {
  border-color: #e74c3c;
}

.form-group .error {
  color: #e74c3c;
  font-size: 14px;
  margin-top: 5px;
  display: block;
}

.error-banner {
  background-color: #ffe6e6;
  color: #d8000c;
  padding: 12px;
  border-radius: 6px;
  margin-bottom: 20px;
  border: 1px solid #ffb3b3;
  text-align: center;
}

.checkbox-group {
  margin-bottom: 25px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  cursor: pointer;
  color: #555;
}

.checkbox-label input[type="checkbox"] {
  width: auto;
  margin-right: 8px;
}

.submit-button {
  width: 100%;
  padding: 14px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.3s ease;
  margin-bottom: 20px;
}

.submit-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.submit-button:hover:not(:disabled) {
  opacity: 0.9;
}

.form-links {
  text-align: center;
  color: #666;
}

.form-links a {
  color: #667eea;
  text-decoration: none;
  font-weight: 500;
}

.form-links a:hover {
  text-decoration: underline;
}

.form-links span {
  margin: 0 10px;
  color: #ccc;
}

.login-link {
  text-align: center;
  color: #666;
  margin: 0;
}

.login-link a {
  color: #667eea;
  text-decoration: none;
  font-weight: 500;
}

.login-link a:hover {
  text-decoration: underline;
}

.loading-spinner {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  font-size: 18px;
  color: #666;
}

/* Responsive Design */
@media (max-width: 768px) {
  .auth-container {
    padding: 10px;
  }
  
  .registration-form,
  .login-form {
    padding: 30px 20px;
  }
  
  .form-row {
    grid-template-columns: 1fr;
    gap: 0;
  }
  
  .registration-form h2,
  .login-form h2 {
    font-size: 24px;
  }
}

/* Dark theme support */
@media (prefers-color-scheme: dark) {
  .registration-form,
  .login-form {
    background: #2c2c2c;
    color: #fff;
  }
  
  .registration-form h2,
  .login-form h2 {
    color: #fff;
  }
  
  .form-group label {
    color: #ccc;
  }
  
  .form-group input,
  .form-group textarea {
    background: #3a3a3a;
    border-color: #555;
    color: #fff;
  }
  
  .form-group input:focus,
  .form-group textarea:focus {
    border-color: #667eea;
  }
  
  .checkbox-label {
    color: #ccc;
  }
  
  .form-links {
    color: #aaa;
  }
  
  .login-link {
    color: #aaa;
  }
}
```

---

## �💬 Chat Implementation

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
