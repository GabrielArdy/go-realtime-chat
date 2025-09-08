# Realtime Chat Server

A real-time chat application server built with Go, featuring WebSocket communication, message queuing with RabbitMQ, and comprehensive chat functionality.

## Features

- 🚀 **Real-time Messaging** - WebSocket-based real-time communication
- 🐰 **Message Queue** - RabbitMQ integration for event distribution
- 🔐 **JWT Authentication** - Secure token-based authentication
- 👥 **User Management** - Complete user registration and profile management
- 🏠 **Room System** - Public/private rooms with member management
- 📎 **File Attachments** - Support for file uploads and attachments
- 😀 **Message Reactions** - Emoji reactions to messages
- 📖 **Read Receipts** - Message read status tracking
- 🔍 **Message Search** - Full-text search across messages
- 💾 **Redis Caching** - High-performance caching layer
- 📊 **Health Monitoring** - Built-in health check endpoints

## Architecture

### Tech Stack
- **Web Framework**: Echo v4
- **ORM**: GORM with PostgreSQL
- **Cache**: Redis with Rueidis client
- **Message Queue**: RabbitMQ
- **WebSocket**: Gorilla WebSocket
- **Authentication**: JWT tokens
- **Configuration**: Viper
- **Logging**: Structured logging
- ✅ User management with authentication
- ✅ RESTful API endpoints
- ✅ Graceful shutdown
- ✅ Environment-based configuration

## Project Structure

```
realtime/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── database/
│   │   └── database.go          # Database connection and setup
│   ├── redis/
│   │   └── redis.go             # Redis client setup
│   ├── logger/
│   │   └── logger.go            # Custom logging system
│   ├── health/
│   │   └── health.go            # Health check system
│   ├── middleware/
│   │   └── middleware.go        # HTTP middleware
│   ├── model/
│   │   └── models.go            # Data models and DTOs
│   ├── repository/
│   │   └── user_repository.go   # Data access layer
│   ├── service/
│   │   └── user_service.go      # Business logic layer
│   └── handler/
│       └── user_handler.go      # HTTP handlers (controllers)
├── pkg/
│   └── utils/
│       └── utils.go             # Utility functions
├── configs/
│   ├── config.yaml              # Development configuration
│   └── config.prod.yaml         # Production configuration
├── go.mod                       # Go module file
└── README.md                    # Project documentation
```

## Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL (default database)
- Redis

### Installation

1. Clone the project and navigate to directory
2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Set up your database and Redis instances

4. Update configuration in `configs/config.yaml` or use environment variables

5. Run the application:
   ```bash
   go run cmd/server/main.go
   ```

The server will start on `http://localhost:8080`

## Configuration

The application uses Viper for configuration management. You can configure the application using:

1. **YAML files** (in `configs/` directory)
2. **Environment variables**

### Environment Variables

All configuration values can be overridden using environment variables with the pattern:
`SECTION_KEY` (e.g., `SERVER_PORT`, `DATABASE_HOST`)

### Configuration Options

```yaml
server:
  host: "0.0.0.0"
  port: "8080"
  read_timeout: 30
  write_timeout: 30
  environment: "development"

database:
  driver: "postgres"  # postgres, mysql, sqlite
  host: "localhost"
  port: "5432"
  username: "postgres"
  password: "password"
  database: "realtime_db"
  ssl_mode: "disable"

redis:
  host: "localhost"
  port: "6379"
  password: ""
  database: 0

logger:
  level: "info"        # debug, info, warn, error, fatal
  format: "json"       # json, text
  output: "stdout"     # stdout, stderr, or file path
```

## API Endpoints

### Health Checks

- `GET /health` - Comprehensive health check
- `GET /health/ready` - Readiness probe (for Kubernetes)
- `GET /health/live` - Liveness probe (for Kubernetes)

### Authentication

- `POST /api/v1/auth/login` - User login

### Users

- `POST /api/v1/users` - Create a new user
- `GET /api/v1/users` - List users (with pagination)
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

## Architecture

This project follows Clean Architecture principles with clear separation of concerns:

1. **Handler Layer** (`internal/handler/`) - HTTP request handling
2. **Service Layer** (`internal/service/`) - Business logic
3. **Repository Layer** (`internal/repository/`) - Data access
4. **Model Layer** (`internal/model/`) - Data structures

## Logging

The application includes a comprehensive logging system with:

- Structured logging (JSON format)
- Multiple log levels (DEBUG, INFO, WARN, ERROR, FATAL)
- Request/response logging
- Error tracking with stack traces
- Configurable output destinations

## Health Monitoring

Built-in health checking system monitors:

- Database connectivity
- Redis connectivity
- System resources (memory, goroutines)
- Application uptime

## Security Features

- Password hashing using Argon2
- Rate limiting middleware
- CORS protection
- Request ID tracking
- Input sanitization

## Development

### Running Tests
```bash
go test ./...
```

### Building for Production
```bash
go build -o realtime-server cmd/server/main.go
```

### Database Migrations

The application automatically runs database migrations on startup using GORM's AutoMigrate feature.

## Deployment

### Docker

Create a `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o realtime-server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/realtime-server .
COPY --from=builder /app/configs ./configs
CMD ["./realtime-server"]
```

### Environment Variables for Production

```bash
export SERVER_PORT=8080
export DATABASE_HOST=your-db-host
export DATABASE_PASSWORD=your-db-password
export REDIS_HOST=your-redis-host
export LOG_LEVEL=info
```

## Contributing

1. Follow Go coding standards
2. Add tests for new features
3. Update documentation
4. Ensure all linters pass

## License

This project is licensed under the MIT License.
