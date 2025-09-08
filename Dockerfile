# Build stage
FROM golang:1.22-alpine3.19 AS builder

# Install git and build dependencies
RUN apk update && \
    apk add --no-cache git gcc musl-dev && \
    rm -rf /var/cache/apk/*

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags='-w -s -extldflags "-static"' \
    -o realtime-server \
    cmd/server/main.go

# Final stage
FROM alpine:3.19

# Update package index and install dependencies
RUN apk update && \
    apk add --no-cache ca-certificates tzdata curl && \
    rm -rf /var/cache/apk/*

# Create app directory and user
WORKDIR /app
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy binary from builder stage
COPY --from=builder /app/realtime-server .

# Copy configuration files
COPY --from=builder /app/configs ./configs

# Create entry script
COPY entry.sh .

# Make scripts executable and set ownership
RUN chmod +x ./realtime-server && \
    chmod +x ./entry.sh && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
  CMD curl -f http://localhost:8080/health/live || exit 1

# Run the application via entry script
CMD ["./entry.sh"]
