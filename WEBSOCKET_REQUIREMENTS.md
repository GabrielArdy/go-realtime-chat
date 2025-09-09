# 🔌 JAWABAN: Persyaratan Koneksi WebSocket

## ❓ Pertanyaan: "untuk terhubung dengan ws harus membawa apa??"

### ✅ JAWABAN LENGKAP:

Untuk terhubung ke WebSocket (`/ws`), Anda **WAJIB** membawa:

## 1. 🔑 JWT Token yang Valid

**Cara Mendapatkan Token:**
```bash
# 1. Login terlebih dahulu
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@user.com", 
    "password": "password123"
  }'

# 2. Ambil access_token dari response
```

**Response Login:**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {...}
  }
}
```

## 2. 🔗 Format URL WebSocket yang Benar

**Opsi A: Query Parameter (Recommended)**
```
ws://localhost:8080/ws?token=YOUR_JWT_TOKEN
```

**Opsi B: Authorization Header**
```
Authorization: Bearer YOUR_JWT_TOKEN
```

## 3. 🧪 CARA TEST LANGSUNG

### Test Menggunakan HTML Test File:
```bash
# Buka file ini di browser:
open test/websocket-test.html
```

### Test Menggunakan JavaScript:
```javascript
// 1. Login dulu
const loginResponse = await fetch('http://localhost:8080/api/v1/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'test@user.com',
    password: 'password123'
  })
});

const { data } = await loginResponse.json();
const token = data.access_token;

// 2. Connect WebSocket dengan token
const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

ws.onopen = () => console.log('✅ Connected!');
ws.onmessage = (e) => console.log('📨 Received:', JSON.parse(e.data));

// 3. Test ping
ws.send(JSON.stringify({ type: 'ping', data: {} }));
```

## 4. ❌ ERROR yang Mungkin Terjadi

### Error 401 Unauthorized:
```
"missing authentication token"
```
**Solusi:** Pastikan token valid dan tidak kadaluarsa

### Error 429 Rate Limited:
```
"Rate limit exceeded"
```
**Solusi:** Tunggu beberapa detik sebelum mencoba lagi

## 5. 📋 CHECKLIST DEBUG

- [ ] ✅ Sudah login dan dapat `access_token`?
- [ ] ✅ Token belum kadaluarsa (`expires_at`)?
- [ ] ✅ URL format benar: `ws://localhost:8080/ws?token=TOKEN`?
- [ ] ✅ Server sudah running: `docker compose up`?
- [ ] ✅ Health check OK: `curl http://localhost:8080/health/live`?

## 6. 🎯 LANGKAH CEPAT UNTUK TEST

1. **Buka browser, tekan F12 (Developer Console)**
2. **Copy-paste script ini:**

```javascript
// Script lengkap untuk test WebSocket
async function testWebSocket() {
  try {
    // Step 1: Login
    console.log('🔄 Logging in...');
    const loginResponse = await fetch('http://localhost:8080/api/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        email: 'test@user.com',
        password: 'password123'
      })
    });
    
    const loginData = await loginResponse.json();
    if (!loginData.success) {
      throw new Error('Login failed: ' + loginData.message);
    }
    
    const token = loginData.data.access_token;
    console.log('✅ Login successful! Token:', token.substring(0, 20) + '...');
    
    // Step 2: Connect WebSocket
    console.log('🔄 Connecting to WebSocket...');
    const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);
    
    ws.onopen = () => {
      console.log('✅ WebSocket connected successfully!');
      
      // Step 3: Test ping
      ws.send(JSON.stringify({ type: 'ping', data: {} }));
      console.log('📤 Sent ping');
    };
    
    ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      console.log('📨 Received:', message);
    };
    
    ws.onerror = (error) => {
      console.error('❌ WebSocket error:', error);
    };
    
    ws.onclose = (event) => {
      console.log('❌ WebSocket closed:', event.code, event.reason);
    };
    
  } catch (error) {
    console.error('❌ Test failed:', error);
  }
}

// Jalankan test
testWebSocket();
```

3. **Tekan Enter dan lihat hasilnya!**

## 📚 Dokumentasi Lengkap

- **WebSocket Connection Guide:** `docs/WEBSOCKET_CONNECTION_GUIDE.md`
- **Testing Guide:** `docs/WEBSOCKET_TESTING.md` 
- **Frontend Integration:** `docs/FRONTEND_INTEGRATION_GUIDE.md`
- **Test HTML File:** `test/websocket-test.html`

---

### 🎯 RINGKASAN SINGKAT:

**Untuk connect WebSocket:**
1. **Login** → Get JWT token
2. **Connect** → `ws://localhost:8080/ws?token=TOKEN`
3. **Test** → Send ping, receive pong

**Tanpa token = ERROR 401 Unauthorized!** ❌
