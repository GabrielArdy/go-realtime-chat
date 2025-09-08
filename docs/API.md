# API Documentation

## Base URL
```
http://localhost:8080
```

## Health Check Endpoints

### Check Overall Health
```http
GET /health
```

**Response Example:**
```json
{
  "status": "healthy",
  "timestamp": "2023-01-01T12:00:00Z",
  "version": "1.0.0",
  "uptime": "1h30m15s",
  "system": {
    "go_version": "go1.21.0",
    "num_goroutine": 10,
    "num_cpu": 8,
    "memory_stats": {
      "alloc": 2048576,
      "total_alloc": 4096000,
      "sys": 8192000,
      "num_gc": 5
    }
  },
  "checks": {
    "database": {
      "status": "healthy",
      "message": "Database connection is healthy"
    },
    "redis": {
      "status": "healthy",
      "message": "Redis connection is healthy"
    }
  }
}
```

### Readiness Check
```http
GET /health/ready
```

### Liveness Check
```http
GET /health/live
```

## Authentication Endpoints

### Login
```http
POST /api/v1/auth/login
```

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": 1,
      "username": "johndoe",
      "email": "user@example.com",
      "is_active": true,
      "created_at": "2023-01-01T12:00:00Z",
      "updated_at": "2023-01-01T12:00:00Z",
      "profile": {
        "id": 1,
        "user_id": 1,
        "first_name": "John",
        "last_name": "Doe",
        "bio": "Software developer",
        "avatar": ""
      }
    },
    "token": "jwt-token-here"
  }
}
```

## User Management Endpoints

### Create User
```http
POST /api/v1/users
```

**Request Body:**
```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "securepassword123",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Response:**
```json
{
  "success": true,
  "message": "User created successfully",
  "data": {
    "id": 1,
    "username": "johndoe",
    "email": "john@example.com",
    "is_active": true,
    "created_at": "2023-01-01T12:00:00Z",
    "updated_at": "2023-01-01T12:00:00Z",
    "profile": {
      "id": 1,
      "user_id": 1,
      "first_name": "John",
      "last_name": "Doe",
      "bio": "",
      "avatar": ""
    }
  }
}
```

### Get User by ID
```http
GET /api/v1/users/{id}
```

**Response:**
```json
{
  "success": true,
  "message": "User retrieved successfully",
  "data": {
    "id": 1,
    "username": "johndoe",
    "email": "john@example.com",
    "is_active": true,
    "created_at": "2023-01-01T12:00:00Z",
    "updated_at": "2023-01-01T12:00:00Z",
    "profile": {
      "id": 1,
      "user_id": 1,
      "first_name": "John",
      "last_name": "Doe",
      "bio": "Software developer",
      "avatar": ""
    }
  }
}
```

### List Users
```http
GET /api/v1/users?page=1&limit=10
```

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 10, max: 100)

**Response:**
```json
{
  "success": true,
  "message": "Users retrieved successfully",
  "data": [
    {
      "id": 1,
      "username": "johndoe",
      "email": "john@example.com",
      "is_active": true,
      "created_at": "2023-01-01T12:00:00Z",
      "updated_at": "2023-01-01T12:00:00Z",
      "profile": {
        "id": 1,
        "user_id": 1,
        "first_name": "John",
        "last_name": "Doe",
        "bio": "Software developer",
        "avatar": ""
      }
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "total_pages": 1
  }
}
```

### Update User
```http
PUT /api/v1/users/{id}
```

**Request Body (partial update):**
```json
{
  "username": "newhhandle",
  "is_active": false
}
```

**Response:**
```json
{
  "success": true,
  "message": "User updated successfully",
  "data": {
    "id": 1,
    "username": "newhandle",
    "email": "john@example.com",
    "is_active": false,
    "created_at": "2023-01-01T12:00:00Z",
    "updated_at": "2023-01-01T13:00:00Z",
    "profile": {
      "id": 1,
      "user_id": 1,
      "first_name": "John",
      "last_name": "Doe",
      "bio": "Software developer",
      "avatar": ""
    }
  }
}
```

### Delete User
```http
DELETE /api/v1/users/{id}
```

**Response:**
```json
{
  "success": true,
  "message": "User deleted successfully"
}
```

## Error Responses

All error responses follow this format:

```json
{
  "success": false,
  "message": "Error description",
  "error": "Detailed error information"
}
```

### Common HTTP Status Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Access denied
- `404 Not Found` - Resource not found
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

## Rate Limiting

- Default rate limit: 100 requests per minute per IP
- Rate limit headers are included in responses:
  - `X-RateLimit-Limit`: Maximum requests per minute
  - `X-RateLimit-Remaining`: Remaining requests in current window
  - `X-RateLimit-Reset`: Time when the rate limit resets

## Request/Response Headers

### Common Request Headers
- `Content-Type: application/json`
- `Authorization: Bearer <token>` (for authenticated endpoints)

### Common Response Headers
- `Content-Type: application/json`
- `X-Request-ID: <unique-id>` (for request tracing)

## Examples using cURL

### Create a user:
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "securepassword123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### Login:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepassword123"
  }'
```

### Get users list:
```bash
curl -X GET "http://localhost:8080/api/v1/users?page=1&limit=5"
```

### Check health:
```bash
curl -X GET http://localhost:8080/health
```
