package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"realtime-api/internal/database"
	"realtime-api/internal/logger"
	"realtime-api/internal/redis"
)

type HealthChecker struct {
	checks map[string]CheckFunc
}

type CheckFunc func(ctx context.Context) CheckResult

type CheckResult struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	System    SystemInfo             `json:"system"`
	Checks    map[string]CheckResult `json:"checks"`
}

type SystemInfo struct {
	GoVersion    string      `json:"go_version"`
	NumGoroutine int         `json:"num_goroutine"`
	NumCPU       int         `json:"num_cpu"`
	MemoryStats  MemoryStats `json:"memory_stats"`
}

type MemoryStats struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	NumGC      uint32 `json:"num_gc"`
}

var (
	DefaultHealthChecker *HealthChecker
	startTime            = time.Now()
	version              = "1.0.0" // This should be set during build
)

func Init() *HealthChecker {
	hc := &HealthChecker{
		checks: make(map[string]CheckFunc),
	}

	// Register default checks
	hc.RegisterCheck("database", DatabaseCheck)
	hc.RegisterCheck("redis", RedisCheck)

	DefaultHealthChecker = hc
	return hc
}

func (hc *HealthChecker) RegisterCheck(name string, check CheckFunc) {
	hc.checks[name] = check
}

func (hc *HealthChecker) Check(ctx context.Context) HealthStatus {
	status := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   version,
		Uptime:    time.Since(startTime).String(),
		System:    getSystemInfo(),
		Checks:    make(map[string]CheckResult),
	}

	// Run all health checks
	for name, check := range hc.checks {
		checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		result := check(checkCtx)
		cancel()

		status.Checks[name] = result

		if result.Status != "healthy" {
			status.Status = "unhealthy"
		}
	}

	return status
}

func getSystemInfo() SystemInfo {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return SystemInfo{
		GoVersion:    runtime.Version(),
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
		MemoryStats: MemoryStats{
			Alloc:      memStats.Alloc,
			TotalAlloc: memStats.TotalAlloc,
			Sys:        memStats.Sys,
			NumGC:      memStats.NumGC,
		},
	}
}

// Default health checks
func DatabaseCheck(ctx context.Context) CheckResult {
	if database.DB == nil {
		return CheckResult{
			Status: "unhealthy",
			Error:  "Database not initialized",
		}
	}

	if err := database.DB.Health(); err != nil {
		return CheckResult{
			Status: "unhealthy",
			Error:  fmt.Sprintf("Database connection failed: %v", err),
		}
	}

	return CheckResult{
		Status:  "healthy",
		Message: "Database connection is healthy",
	}
}

func RedisCheck(ctx context.Context) CheckResult {
	if redis.Client == nil {
		return CheckResult{
			Status: "unhealthy",
			Error:  "Redis client not initialized",
		}
	}

	if err := redis.Client.Health(); err != nil {
		return CheckResult{
			Status: "unhealthy",
			Error:  fmt.Sprintf("Redis connection failed: %v", err),
		}
	}

	return CheckResult{
		Status:  "healthy",
		Message: "Redis connection is healthy",
	}
}

// HTTP Handler for health endpoint
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status := DefaultHealthChecker.Check(ctx)

	w.Header().Set("Content-Type", "application/json")

	if status.Status == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if err := json.NewEncoder(w).Encode(status); err != nil {
		logger.Error("Failed to encode health status", logger.WithField("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Readiness check - simpler version for k8s readiness probe
func ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}

// Liveness check - very basic check for k8s liveness probe
func LivenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}
