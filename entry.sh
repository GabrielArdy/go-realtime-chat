#!/bin/sh

# Entry script for Realtime Chat API Server
# This script handles startup, environment setup, and graceful shutdown

set -e

echo "🚀 Starting Realtime Chat API Server..."
echo "   Version: 1.0.0"
echo "   Environment: ${APP_ENV:-development}"
echo "   Port: ${SERVER_PORT:-8080}"

# Function to handle shutdown signals
cleanup() {
    echo ""
    echo "🛑 Received shutdown signal. Gracefully shutting down..."
    
    if [ ! -z "$SERVER_PID" ]; then
        echo "   Stopping server process (PID: $SERVER_PID)..."
        kill -TERM "$SERVER_PID" 2>/dev/null || true
        wait "$SERVER_PID" 2>/dev/null || true
    fi
    
    echo "✅ Server shutdown complete"
    exit 0
}

# Function to check if required files exist
check_requirements() {
    if [ ! -f "./realtime-server" ]; then
        echo "❌ Error: realtime-server binary not found"
        exit 1
    fi
    
    if [ ! -d "./configs" ]; then
        echo "❌ Error: configs directory not found"
        exit 1
    fi
    
    echo "✅ All required files found"
}

# Function to wait for dependencies (Redis, Database, etc.)
wait_for_dependencies() {
    echo "⏳ Checking dependencies..."
    
    # Wait for Redis if REDIS_HOST is set
    if [ ! -z "$REDIS_HOST" ]; then
        echo "   Waiting for Redis at $REDIS_HOST:${REDIS_PORT:-6379}..."
        while ! nc -z "$REDIS_HOST" "${REDIS_PORT:-6379}" 2>/dev/null; do
            echo "   Redis not ready, waiting 2 seconds..."
            sleep 2
        done
        echo "   ✅ Redis is ready"
    fi
    
    # Wait for Database if DB_HOST is set
    if [ ! -z "$DB_HOST" ]; then
        echo "   Waiting for Database at $DB_HOST:${DB_PORT:-5432}..."
        while ! nc -z "$DB_HOST" "${DB_PORT:-5432}" 2>/dev/null; do
            echo "   Database not ready, waiting 2 seconds..."
            sleep 2
        done
        echo "   ✅ Database is ready"
    fi
    
    # Wait for RabbitMQ if RABBITMQ_HOST is set
    if [ ! -z "$RABBITMQ_HOST" ]; then
        echo "   Waiting for RabbitMQ at $RABBITMQ_HOST:${RABBITMQ_PORT:-5672}..."
        while ! nc -z "$RABBITMQ_HOST" "${RABBITMQ_PORT:-5672}" 2>/dev/null; do
            echo "   RabbitMQ not ready, waiting 2 seconds..."
            sleep 2
        done
        echo "   ✅ RabbitMQ is ready"
    fi
    
    echo "✅ All dependencies are ready"
}

# Function to validate configuration
validate_config() {
    echo "🔧 Validating configuration..."
    
    # Check if config file exists
    CONFIG_FILE="./configs/config.yml"
    if [ ! -f "$CONFIG_FILE" ]; then
        echo "   ⚠️  Warning: config.yml not found, using environment variables"
    else
        echo "   ✅ Configuration file found"
    fi
    
    # Validate required environment variables
    if [ -z "$JWT_SECRET" ] && [ -z "$JWT_SECRET_KEY" ]; then
        echo "   ⚠️  Warning: No JWT secret configured. Using default (not recommended for production)"
    fi
    
    echo "✅ Configuration validation complete"
}

# Function to setup logging
setup_logging() {
    # Create logs directory if it doesn't exist
    mkdir -p /tmp/logs
    
    # Set log level based on environment
    if [ -z "$LOG_LEVEL" ]; then
        if [ "$APP_ENV" = "production" ]; then
            export LOG_LEVEL="info"
        else
            export LOG_LEVEL="debug"
        fi
    fi
    
    echo "📝 Logging configured (level: $LOG_LEVEL)"
}

# Function to start the server
start_server() {
    echo "🎯 Starting Realtime Chat API Server..."
    echo "   Binary: ./realtime-server"
    echo "   Config: ./configs/"
    echo "   PID: $$"
    echo ""
    
    # Start the server in background and capture PID
    ./realtime-server &
    SERVER_PID=$!
    
    echo "🚀 Server started successfully (PID: $SERVER_PID)"
    echo "   Health check: http://localhost:${SERVER_PORT:-8080}/health"
    echo "   WebSocket: ws://localhost:${SERVER_PORT:-8080}/ws"
    echo ""
    echo "📊 Server is ready to accept connections!"
    
    # Wait for the server process
    wait "$SERVER_PID"
}

# Main execution flow
main() {
    echo "=================================================="
    echo "  Realtime Chat API Server - Container Startup"
    echo "=================================================="
    
    # Setup signal handlers for graceful shutdown
    trap cleanup SIGTERM SIGINT SIGQUIT
    
    # Execute startup checks and setup
    check_requirements
    validate_config
    setup_logging
    wait_for_dependencies
    
    echo ""
    echo "🎉 All startup checks passed!"
    echo ""
    
    # Start the server (this will block until shutdown)
    start_server
}

# Execute main function
main "$@"
