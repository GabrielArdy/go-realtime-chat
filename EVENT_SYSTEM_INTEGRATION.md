# Realtime API Event System Integration

## Overview

Event system yang telah diimplementasikan menggunakan structured events dengan format `event.[level].[action]` yang terintegrasi dengan WebSocket untuk komunikasi real-time.

## Event System Architecture

### Event Levels
- `user` - Events related to user activities
- `room` - Events related to room management  
- `message` - Events related to messaging
- `system` - Events related to system status

### Core Components

1. **EventPublisher** - Publishes structured events to Redis channels
2. **EventSubscriber** - Subscribes to Redis channels and processes events
3. **EventRouter** - Routes events to appropriate handlers
4. **WebSocket Hub** - Manages real-time connections and broadcasts

## Event Types

### User Events
```
event.user.online
event.user.offline
event.user.typing.start
event.user.typing.stop
event.user.status.change
event.user.profile.update
```

### Room Events
```
event.room.create
event.room.update
event.room.delete
event.room.join
event.room.leave
event.room.member.add
event.room.member.remove
event.room.invite.create
event.room.invite.accept
event.room.invite.reject
```

### Message Events
```
event.message.send
event.message.edit
event.message.delete
event.message.read
event.message.reaction.add
event.message.reaction.remove
```

### System Events
```
event.system.maintenance
event.system.shutdown
event.system.broadcast
```

## API Endpoints

### User Management
```
POST   /api/v1/users                    # Create user
GET    /api/v1/users                    # List users
GET    /api/v1/users/:id                # Get user
PUT    /api/v1/users/:id                # Update user
DELETE /api/v1/users/:id                # Delete user
POST   /api/v1/auth/login               # Login user
```

### Room Management
```
POST   /api/v1/rooms                    # Create room
GET    /api/v1/rooms                    # List rooms
GET    /api/v1/rooms/:id                # Get room
PUT    /api/v1/rooms/:id                # Update room
DELETE /api/v1/rooms/:id                # Delete room
POST   /api/v1/rooms/:id/join           # Join room
POST   /api/v1/rooms/:id/leave          # Leave room
GET    /api/v1/rooms/:id/members        # Get room members
POST   /api/v1/rooms/:id/members        # Add member
DELETE /api/v1/rooms/:id/members/:user_id # Remove member
POST   /api/v1/rooms/:id/invites        # Create invite
POST   /api/v1/rooms/invites/:code/accept # Accept invite
POST   /api/v1/rooms/invites/:code/reject # Reject invite
```

### Message Management
```
POST   /api/v1/messages                 # Send message
GET    /api/v1/messages/:id             # Get message
PUT    /api/v1/messages/:id             # Edit message
DELETE /api/v1/messages/:id             # Delete message
POST   /api/v1/messages/:id/reactions   # Add reaction
DELETE /api/v1/messages/:id/reactions   # Remove reaction
POST   /api/v1/messages/:id/read        # Mark as read
GET    /api/v1/rooms/:room_id/messages  # Get room messages
POST   /api/v1/rooms/:room_id/typing/start # Start typing
POST   /api/v1/rooms/:room_id/typing/stop  # Stop typing
```

### Event System (Monitoring)
```
GET    /api/v1/events/metrics           # Get event metrics
POST   /api/v1/events/system            # Publish system event
GET    /api/v1/events/history           # Get event history
```

## WebSocket Integration

### Connection
```javascript
const ws = new WebSocket('ws://localhost:8080/ws?token=YOUR_JWT_TOKEN');
```

### Message Types
```javascript
// Ping/Pong
{ "type": "ping" }
{ "type": "pong" }

// Authentication
{ "type": "auth", "data": { "status": "connected", "user_id": "..." } }

// Typing indicators
{ "type": "typing_start", "data": { "room_id": "...", "user_id": "...", "username": "..." } }
{ "type": "typing_stop", "data": { "room_id": "...", "user_id": "...", "username": "..." } }

// User status
{ "type": "user_status_change", "data": { "user_id": "...", "username": "...", "status": "..." } }

// Room events
{ "type": "user_join", "data": { "user_id": "...", "username": "..." } }
{ "type": "user_leave", "data": { "user_id": "...", "username": "..." } }
```

## Event Flow Example

### Message Sending Flow
1. User sends POST to `/api/v1/messages`
2. MessageService.SendMessage() is called
3. EventPublisher publishes `event.message.send`
4. Event is broadcast to Redis channel `room:{room_id}`
5. WebSocket Hub receives event and broadcasts to connected clients
6. All room members receive real-time notification

### Room Join Flow
1. User sends POST to `/api/v1/rooms/:id/join`
2. RoomService.JoinRoom() is called
3. EventPublisher publishes `event.room.join`
4. WebSocket Hub.JoinRoom() adds user to room
5. Event is broadcast to all room members
6. Real-time notification sent to existing members

## Configuration

### Redis Channels
- `room:{room_id}` - Room-specific events
- `user:{user_id}` - User-specific events
- `presence` - Presence/status events
- `system` - System-wide events
- `global` - Global broadcast events

### Event Data Structure
```json
{
  "id": "event-uuid",
  "type": "event.message.send",
  "level": "message",
  "action": "send",
  "data": {
    "message_id": "msg-uuid",
    "content": "Hello world",
    "timestamp": "2024-01-01T00:00:00Z"
  },
  "metadata": {
    "client_info": "...",
    "ip_address": "..."
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "user_id": "user-uuid",
  "room_id": "room-uuid"
}
```

## Usage Examples

### Start Server
```bash
cd /home/gli-it/Projects/Sandbox/realtime
./bin/server
```

### WebSocket Client Example
```javascript
const ws = new WebSocket('ws://localhost:8080/ws?token=your-jwt-token');

ws.onopen = () => {
    console.log('Connected to WebSocket');
    
    // Start typing in a room
    ws.send(JSON.stringify({
        type: 'typing_start',
        data: { room_id: 'room-uuid' }
    }));
};

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    console.log('Received:', message);
    
    switch(message.type) {
        case 'typing_start':
            console.log(`${message.data.username} started typing`);
            break;
        case 'auth':
            console.log('Authenticated successfully');
            break;
    }
};
```

### API Request Example
```bash
# Send a message
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "room_id": "room-uuid",
    "content": "Hello everyone!",
    "type": "text"
  }'

# Join a room
curl -X POST http://localhost:8080/api/v1/rooms/room-uuid/join
```

## Health Checks
```bash
curl http://localhost:8080/health        # Basic health
curl http://localhost:8080/health/ready  # Readiness
curl http://localhost:8080/health/live   # Liveness
```

## Event System Benefits

1. **Structured Events** - Consistent event format across the application
2. **Real-time Communication** - WebSocket integration for instant updates
3. **Scalable Architecture** - Redis pub/sub for horizontal scaling
4. **Event Routing** - Flexible handler registration system
5. **Monitoring** - Built-in event metrics and history
6. **Type Safety** - Predefined event constants and structures
