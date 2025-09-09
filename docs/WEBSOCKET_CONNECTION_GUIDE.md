# WebSocket Connection Guide

## Persyaratan untuk Terhubung ke WebSocket

Untuk terhubung ke WebSocket server, Anda **WAJIB** membawa:

### 1. JWT Token yang Valid

WebSocket endpoint memerlukan JWT token yang valid untuk autentikasi. Token dapat diberikan melalui:

**Opsi A: Query Parameter (Recommended)**
```
ws://localhost:8080/ws?token=YOUR_JWT_TOKEN
```

**Opsi B: Authorization Header**
```
Authorization: Bearer YOUR_JWT_TOKEN
```

### 2. URL WebSocket

- **Development**: `ws://localhost:8080/ws`
- **Production**: `wss://your-domain.com/ws`

## Cara Mendapatkan JWT Token

### 1. Login terlebih dahulu
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "your@email.com",
    "password": "yourpassword"
  }'
```

### 2. Ambil access_token dari response
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {...},
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "...",
    "expires_at": "2025-09-09T05:26:49Z",
    "session_id": "..."
  }
}
```

## Contoh Koneksi WebSocket

### JavaScript (Browser)
```javascript
// Ambil token dari login response
const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...';

// Buat koneksi WebSocket dengan token
const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

ws.onopen = function(event) {
    console.log('WebSocket connected');
};

ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    console.log('Received:', message);
};

ws.onerror = function(error) {
    console.error('WebSocket error:', error);
};

ws.onclose = function(event) {
    console.log('WebSocket closed:', event.code, event.reason);
};

// Kirim pesan
ws.send(JSON.stringify({
    type: 'ping',
    data: {}
}));
```

### Node.js
```javascript
const WebSocket = require('ws');

const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...';
const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

ws.on('open', function open() {
    console.log('WebSocket connected');
    
    // Send ping
    ws.send(JSON.stringify({
        type: 'ping',
        data: {}
    }));
});

ws.on('message', function message(data) {
    const msg = JSON.parse(data.toString());
    console.log('Received:', msg);
});
```

### Python
```python
import asyncio
import websockets
import json

async def websocket_client():
    token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    uri = f"ws://localhost:8080/ws?token={token}"
    
    async with websockets.connect(uri) as websocket:
        print("WebSocket connected")
        
        # Send ping
        await websocket.send(json.dumps({
            "type": "ping",
            "data": {}
        }))
        
        # Receive messages
        async for message in websocket:
            data = json.loads(message)
            print(f"Received: {data}")

asyncio.run(websocket_client())
```

## Format Pesan WebSocket

### Ping/Pong
```json
// Send ping
{
  "type": "ping",
  "data": {}
}

// Receive pong
{
  "type": "pong",
  "data": null,
  "timestamp": "2025-09-09T04:30:00Z",
  "id": "uuid"
}
```

### Typing Indicators
```json
// Start typing
{
  "type": "typing_start",
  "data": {
    "room_id": "room-uuid"
  }
}

// Stop typing
{
  "type": "typing_stop",
  "data": {
    "room_id": "room-uuid"
  }
}
```

### User Status Change
```json
{
  "type": "user_status_change",
  "data": {
    "status": "online" // online, offline, away, busy, invisible
  }
}
```

## Error Codes

### HTTP Errors
- `401 Unauthorized`: Token tidak valid atau tidak ada
- `429 Too Many Requests`: Rate limit terlampaui
- `404 Not Found`: Endpoint WebSocket tidak ditemukan

### WebSocket Close Codes
- `1000`: Normal closure
- `1001`: Going away
- `1002`: Protocol error
- `1003`: Unsupported data
- `1006`: Abnormal closure
- `1011`: Internal server error

## Troubleshooting

### 1. Koneksi Gagal (401 Unauthorized)
- **Masalah**: Token JWT tidak valid atau kadaluarsa
- **Solusi**: Login ulang untuk mendapatkan token baru

### 2. Koneksi Ditolak (429 Rate Limited)
- **Masalah**: Terlalu banyak request dalam waktu singkat
- **Solusi**: Tunggu beberapa detik sebelum mencoba lagi

### 3. WebSocket Disconnect Otomatis
- **Masalah**: Koneksi tidak stabil atau server restart
- **Solusi**: Implementasikan reconnection logic dengan exponential backoff

### 4. CORS Issues (Browser)
- **Masalah**: Browser memblokir koneksi dari origin berbeda
- **Solusi**: Server sudah dikonfigurasi untuk menerima koneksi dari localhost

## Implementasi Reconnection

```javascript
class WebSocketManager {
    constructor(token) {
        this.token = token;
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000; // 1 second
    }
    
    connect() {
        try {
            this.ws = new WebSocket(`ws://localhost:8080/ws?token=${this.token}`);
            
            this.ws.onopen = () => {
                console.log('WebSocket connected');
                this.reconnectAttempts = 0;
            };
            
            this.ws.onclose = (event) => {
                console.log('WebSocket closed:', event.code);
                if (event.code !== 1000) { // Not normal closure
                    this.reconnect();
                }
            };
            
            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
            };
            
        } catch (error) {
            console.error('Connection failed:', error);
            this.reconnect();
        }
    }
    
    reconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
            
            console.log(`Reconnecting in ${delay}ms... (attempt ${this.reconnectAttempts})`);
            
            setTimeout(() => {
                this.connect();
            }, delay);
        } else {
            console.error('Max reconnection attempts reached');
        }
    }
}

// Usage
const token = 'your-jwt-token';
const wsManager = new WebSocketManager(token);
wsManager.connect();
```

## Best Practices

1. **Selalu validasi token** sebelum membuat koneksi
2. **Implementasikan reconnection logic** untuk koneksi yang stabil
3. **Handle error dengan graceful** untuk UX yang baik
4. **Gunakan ping/pong** untuk menjaga koneksi tetap hidup
5. **Batasi rate limit** pada client side untuk menghindari 429 errors

## Origin yang Diizinkan

Server dikonfigurasi untuk menerima koneksi dari:
- `http://localhost:3000` (React dev server)
- `http://localhost:8080` (Backend server)
- `https://yourapp.com` (Production frontend)
- Origin kosong (untuk native apps)

Untuk origin lain, hubungi administrator untuk menambahkannya ke whitelist.
