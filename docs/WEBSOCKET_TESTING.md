# WebSocket Test & Debugging

## 🔧 Cara Test Koneksi WebSocket

### 1. Menggunakan Test HTML (Recommended)

Buka file `test/websocket-test.html` di browser untuk test interaktif:

```bash
# Dari root project
open test/websocket-test.html
# atau
firefox test/websocket-test.html
# atau drag & drop ke browser
```

**Fitur Test HTML:**
- ✅ Login otomatis untuk mendapatkan JWT token
- ✅ Test koneksi WebSocket dengan token
- ✅ Kirim berbagai jenis pesan (ping, typing, status)
- ✅ Monitor log pesan real-time
- ✅ Tampilan status koneksi yang jelas

### 2. Menggunakan Browser DevTools

1. **Buka Developer Console**
2. **Login terlebih dahulu:**
```javascript
// Login untuk mendapatkan token
fetch('http://localhost:8080/api/v1/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'test@user.com',
    password: 'password123'
  })
})
.then(r => r.json())
.then(data => {
  if (data.success) {
    console.log('Token:', data.data.access_token);
    window.authToken = data.data.access_token;
  }
});
```

3. **Buat koneksi WebSocket:**
```javascript
// Gunakan token dari login
const ws = new WebSocket(`ws://localhost:8080/ws?token=${window.authToken}`);

ws.onopen = () => console.log('✅ Connected');
ws.onmessage = (e) => console.log('📨 Received:', JSON.parse(e.data));
ws.onclose = (e) => console.log('❌ Closed:', e.code, e.reason);
ws.onerror = (e) => console.error('❌ Error:', e);

// Test ping
ws.send(JSON.stringify({ type: 'ping', data: {} }));
```

### 3. Menggunakan cURL (HTTP Upgrade)

```bash
# Test dengan token yang valid
curl -i "http://localhost:8080/ws?token=YOUR_JWT_TOKEN" \
  -H "Connection: Upgrade" \
  -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Version: 13" \
  -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ=="
```

**Expected Response:**
```
HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
```

### 4. Menggunakan WebSocket Client Tools

#### wscat (Node.js)
```bash
# Install wscat
npm install -g wscat

# Connect dengan token
wscat -c "ws://localhost:8080/ws?token=YOUR_JWT_TOKEN"

# Send messages
> {"type":"ping","data":{}}
< {"type":"pong","data":null,"timestamp":"2025-09-09T04:30:00Z","id":"uuid"}
```

#### websocat (Rust)
```bash
# Install websocat
cargo install websocat
# atau download binary

# Connect
echo '{"type":"ping","data":{}}' | websocat "ws://localhost:8080/ws?token=YOUR_JWT_TOKEN"
```

## 🐛 Troubleshooting

### Error: 401 Unauthorized
```
WebSocket connection failed: 401 Unauthorized
```

**Penyebab:** Token JWT tidak valid, kadaluarsa, atau tidak ada

**Solusi:**
1. Pastikan sudah login terlebih dahulu
2. Copy `access_token` dengan benar dari response login
3. Periksa token tidak kadaluarsa (lihat `expires_at`)
4. Pastikan format URL benar: `ws://localhost:8080/ws?token=TOKEN`

### Error: 429 Rate Limited
```
HTTP/1.1 429 Too Many Requests
```

**Penyebab:** Terlalu banyak request dalam waktu singkat

**Solusi:**
1. Tunggu beberapa detik sebelum mencoba lagi
2. Implementasikan rate limiting di client
3. Gunakan exponential backoff untuk reconnection

### Error: Connection Refused
```
WebSocket connection failed: Connection refused
```

**Penyebab:** Server tidak berjalan atau URL salah

**Solusi:**
1. Pastikan server sudah running: `docker compose up`
2. Periksa port: `curl http://localhost:8080/health/live`
3. Periksa URL WebSocket: `ws://localhost:8080/ws`

### Error: Origin Not Allowed
```
WebSocket connection rejected
```

**Penyebab:** Origin tidak diizinkan oleh server

**Solusi:**
- Server dikonfigurasi untuk menerima origin:
  - `http://localhost:3000` (React dev)
  - `http://localhost:8080` (Backend)
  - `https://yourapp.com` (Production)
- Untuk testing, origin kosong juga diizinkan

### Error: Invalid Message Format
```
Failed to parse WebSocket message
```

**Penyebab:** Format JSON pesan tidak valid

**Solusi:**
Pastikan pesan mengikuti format yang benar:
```json
{
  "type": "message_type",
  "data": {},
  "timestamp": "2025-09-09T04:30:00Z"
}
```

## 📊 Debug Server Logs

Monitor server logs untuk debug:

```bash
# Monitor real-time logs
docker compose logs -f realtime-api

# Cari WebSocket events
docker compose logs realtime-api | grep -i websocket

# Cari authentication errors
docker compose logs realtime-api | grep -i "401\|unauthorized\|token"
```

**Log Patterns:**
- ✅ Successful connection: `"WebSocket connected"`
- ❌ Auth failure: `"missing authentication token"`
- ❌ Rate limit: `"Rate limit exceeded"`
- 🔄 Ping/Pong: `"type":"ping"` / `"type":"pong"`

## 🔍 Network Debugging

### Browser Network Tab
1. Buka Developer Tools → Network
2. Filter by "WS" (WebSocket)
3. Monitor WebSocket frames:
   - ⬆️ Outgoing: Pesan yang dikirim
   - ⬇️ Incoming: Pesan yang diterima
   - 🔌 Connection: Status koneksi

### Check Connection Details
```javascript
// Dalam browser console
console.log('WebSocket state:', ws.readyState);
// 0 = CONNECTING, 1 = OPEN, 2 = CLOSING, 3 = CLOSED

console.log('WebSocket URL:', ws.url);
console.log('WebSocket protocol:', ws.protocol);
```

## 📝 Test Scenarios

### Basic Connection Test
1. Login → Get token
2. Connect WebSocket dengan token
3. Send ping → Expect pong
4. Disconnect

### Typing Indicator Test
1. Connect WebSocket
2. Send `typing_start` dengan `room_id`
3. Send `typing_stop` dengan `room_id`
4. Verify events diterima

### Status Update Test
1. Connect WebSocket
2. Send `user_status_change` dengan status baru
3. Verify status berubah

### Reconnection Test
1. Connect WebSocket
2. Force disconnect (close tab/kill connection)
3. Verify auto-reconnection works
4. Check exponential backoff

## 🚀 Production Checklist

- [ ] **HTTPS/WSS**: Gunakan `wss://` untuk production
- [ ] **Token Refresh**: Implementasikan auto-refresh token
- [ ] **Error Handling**: Handle semua error scenarios
- [ ] **Reconnection**: Implementasikan robust reconnection
- [ ] **Rate Limiting**: Respect server rate limits
- [ ] **Origin Validation**: Pastikan origin terdaftar
- [ ] **Connection Pooling**: Batasi jumlah koneksi concurrent
- [ ] **Monitoring**: Log dan monitor WebSocket health
