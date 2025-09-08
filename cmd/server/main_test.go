package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"realtime-api/internal/config"
	"realtime-api/internal/logger"
	"realtime-api/internal/model"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHealthEndpoint(t *testing.T) {
	// Initialize logger for tests
	logger.Init("info", "json", "stdout", "")

	// Initialize config
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:        "localhost",
			Port:        "8080",
			Environment: "test",
		},
	}
	config.AppConfig = cfg

	// Create Echo instance
	e := echo.New()

	// Add health endpoint
	e.GET("/health/live", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":    "alive",
			"timestamp": "2023-01-01T12:00:00Z",
		})
	})

	// Test request
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "alive", response["status"])
}

func TestAPIResponse(t *testing.T) {
	// Test API response structure
	response := model.APIResponse{
		Success: true,
		Message: "Test message",
		Data:    map[string]string{"key": "value"},
	}

	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)

	var unmarshaled model.APIResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, true, unmarshaled.Success)
	assert.Equal(t, "Test message", unmarshaled.Message)
}

func TestUserModel(t *testing.T) {
	// Test user model creation
	user := model.User{
		Username: "testuser",
		Email:    "test@example.com",
		IsActive: true,
	}

	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, true, user.IsActive)
}

func TestCreateUserRequest(t *testing.T) {
	// Test create user request
	createReq := model.CreateUserRequest{
		Username:  "newuser",
		Email:     "new@example.com",
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
	}

	jsonData, err := json.Marshal(createReq)
	assert.NoError(t, err)

	var unmarshaled model.CreateUserRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, "newuser", unmarshaled.Username)
	assert.Equal(t, "new@example.com", unmarshaled.Email)
}

func TestEchoJSONBinding(t *testing.T) {
	// Test Echo JSON binding
	e := echo.New()

	reqBody := model.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	jsonData, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	var boundReq model.LoginRequest
	err := c.Bind(&boundReq)
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", boundReq.Email)
	assert.Equal(t, "password123", boundReq.Password)
}
