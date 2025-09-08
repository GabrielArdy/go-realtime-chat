# Makefile for Realtime API Server

.PHONY: build run clean test lint docker-build docker-run help

# Variables
BINARY_NAME=realtime-server
MAIN_PATH=cmd/server/main.go
DOCKER_IMAGE=realtime-api
DOCKER_TAG=latest

# Default target
.DEFAULT_GOAL := help

## Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) $(MAIN_PATH)

## Run the application
run:
	@echo "Running the application..."
	@go run $(MAIN_PATH)

## Clean build artifacts
clean:
	@echo "Cleaning..."
	@go clean
	@rm -f $(BINARY_NAME)

## Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

## Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

## Lint the code
lint:
	@echo "Running linter..."
	@golangci-lint run

## Format the code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

## Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

## Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download

## Run with development config
dev:
	@echo "Running in development mode..."
	@go run $(MAIN_PATH)

## Run with production config
prod:
	@echo "Running in production mode..."
	@CONFIG_PATH=./configs/config.prod.yaml go run $(MAIN_PATH)

## Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

## Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)

## Generate API documentation
docs:
	@echo "Generating documentation..."
	@# Add your documentation generation command here

## Database migration (if you add a migration tool)
migrate-up:
	@echo "Running database migrations..."
	@# Add your migration command here

## Database rollback (if you add a migration tool)
migrate-down:
	@echo "Rolling back database migrations..."
	@# Add your rollback command here

## Setup development environment
setup-dev:
	@echo "Setting up development environment..."
	@go mod tidy
	@# Add any other setup commands

## Show available commands
help:
	@echo "Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""
