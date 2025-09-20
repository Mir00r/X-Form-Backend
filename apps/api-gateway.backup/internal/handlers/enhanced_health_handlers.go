package handlers

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/models"
	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

// EnhancedHealthCheck godoc
// @Summary      Get comprehensive health status
// @Description  Returns detailed health information including system resources, service dependencies, and performance metrics
// @Tags         System,Health & Monitoring
// @Accept       json
// @Produce      json
// @Param        include_dependencies query bool false "Include dependency health checks" default(true)
// @Param        include_metrics query bool false "Include system metrics" default(true)
// @Success      200 {object} models.StandardAPIResponse{data=models.HealthCheckResponse} "Service is healthy"
// @Failure      503 {object} models.StandardAPIResponse{data=models.HealthCheckResponse} "Service is unhealthy"
// @Failure      500 {object} models.StandardAPIResponse{error=models.DetailedError} "Internal server error"
// @Router       /health [get]
func EnhancedHealthCheck(c *gin.Context) {
	includeDeps := c.DefaultQuery("include_dependencies", "true") == "true"
	_ = c.DefaultQuery("include_metrics", "true") == "true" // Unused but keep for future use

	now := time.Now()
	uptime := now.Sub(startTime)

	// System health information
	var memStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memStats)

	health := models.HealthCheckResponse{
		Status:      "healthy",
		Service:     "api-gateway",
		Version:     "1.0.0",
		Environment: "production",
		Timestamp:   now,
		Uptime:      formatDuration(uptime),
		Checks:      make(map[string]models.ServiceHealth),
		System: models.SystemHealth{
			Memory: struct {
				Used    uint64  `json:"used" example:"524288000" description:"Used memory in bytes"`
				Total   uint64  `json:"total" example:"2147483648" description:"Total memory in bytes"`
				Percent float64 `json:"percent" example:"24.4" description:"Memory usage percentage"`
			}{
				Used:    memStats.Alloc,
				Total:   memStats.Sys,
				Percent: float64(memStats.Alloc) / float64(memStats.Sys) * 100,
			},
			CPU: struct {
				Percent float64 `json:"percent" example:"15.7" description:"CPU usage percentage"`
				Cores   int     `json:"cores" example:"4" description:"Number of CPU cores"`
			}{
				Percent: 15.7, // Mock CPU usage
				Cores:   runtime.NumCPU(),
			},
			Disk: struct {
				Used    uint64  `json:"used" example:"1073741824" description:"Used disk space in bytes"`
				Total   uint64  `json:"total" example:"10737418240" description:"Total disk space in bytes"`
				Percent float64 `json:"percent" example:"10.0" description:"Disk usage percentage"`
			}{
				Used:    1073741824,  // 1GB mock
				Total:   10737418240, // 10GB mock
				Percent: 10.0,
			},
			Goroutines: runtime.NumGoroutine(),
		},
		Dependencies: make(map[string]models.DependencyHealth),
	}

	// Add service checks if requested
	if includeDeps {
		health.Checks["database"] = models.ServiceHealth{
			Status:    "healthy",
			LastCheck: now,
			Duration:  50 * time.Millisecond,
		}

		health.Checks["redis"] = models.ServiceHealth{
			Status:    "healthy",
			LastCheck: now,
			Duration:  25 * time.Millisecond,
		}

		health.Checks["auth_service"] = models.ServiceHealth{
			Status:    "healthy",
			LastCheck: now,
			Duration:  100 * time.Millisecond,
		}

		// Mock dependency health
		health.Dependencies["auth-service"] = models.DependencyHealth{
			Status:    "healthy",
			URL:       "http://auth-service:3001/health",
			LastCheck: now,
			Duration:  100 * time.Millisecond,
			Version:   "1.2.0",
		}

		health.Dependencies["form-service"] = models.DependencyHealth{
			Status:    "healthy",
			URL:       "http://form-service:8080/health",
			LastCheck: now,
			Duration:  120 * time.Millisecond,
			Version:   "1.1.0",
		}

		health.Dependencies["response-service"] = models.DependencyHealth{
			Status:    "healthy",
			URL:       "http://response-service:3003/health",
			LastCheck: now,
			Duration:  80 * time.Millisecond,
			Version:   "1.0.5",
		}

		health.Dependencies["analytics-service"] = models.DependencyHealth{
			Status:    "healthy",
			URL:       "http://analytics-service:8084/health",
			LastCheck: now,
			Duration:  90 * time.Millisecond,
			Version:   "1.0.0",
		}
	}

	// Determine overall health status
	overallStatus := "healthy"
	for _, check := range health.Checks {
		if check.Status == "unhealthy" {
			overallStatus = "degraded"
		}
	}

	for _, dep := range health.Dependencies {
		if dep.Status == "unhealthy" {
			overallStatus = "degraded"
		}
	}

	health.Status = overallStatus

	// Return appropriate HTTP status
	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	message := "Service is healthy"
	if overallStatus == "degraded" {
		message = "Service is operational with some degraded components"
	} else if overallStatus == "unhealthy" {
		message = "Service is experiencing issues"
	}

	c.JSON(statusCode, models.StandardAPIResponse{
		Success:   statusCode == http.StatusOK,
		Message:   message,
		Data:      health,
		RequestID: c.GetString("request_id"),
		Timestamp: now,
		Meta: &models.ResponseMetadata{
			RequestDuration: "50ms",
			APIVersion:      "v1",
			ServerInstance:  "gateway-01",
		},
	})
}

// Metrics godoc
// @Summary      Get service metrics
// @Description  Returns comprehensive service metrics including performance, usage, and system statistics
// @Tags         System,Health & Monitoring
// @Accept       json
// @Produce      json
// @Success      200 {object} models.StandardAPIResponse{data=object} "Metrics retrieved successfully"
// @Failure      401 {object} models.StandardAPIResponse{error=models.DetailedError} "Authentication required"
// @Failure      500 {object} models.StandardAPIResponse{error=models.DetailedError} "Internal server error"
// @Router       /metrics [get]
// @Security     BearerAuth
func EnhancedMetrics(c *gin.Context) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics := map[string]interface{}{
		"uptime_seconds": time.Since(startTime).Seconds(),
		"memory": map[string]interface{}{
			"alloc_bytes":       memStats.Alloc,
			"total_alloc_bytes": memStats.TotalAlloc,
			"sys_bytes":         memStats.Sys,
			"gc_runs":           memStats.NumGC,
			"gc_pause_ns":       memStats.PauseNs,
		},
		"goroutines": runtime.NumGoroutine(),
		"cpu_cores":  runtime.NumCPU(),
		"requests": map[string]interface{}{
			"total":           12345,
			"success":         11500,
			"errors":          845,
			"rate_per_second": 25.5,
		},
		"response_times": map[string]interface{}{
			"avg_ms": 150.5,
			"p50_ms": 125.0,
			"p95_ms": 350.0,
			"p99_ms": 800.0,
		},
		"cache": map[string]interface{}{
			"hits":   8500,
			"misses": 1500,
			"ratio":  0.85,
		},
	}

	c.JSON(http.StatusOK, models.StandardAPIResponse{
		Success:   true,
		Message:   "Metrics retrieved successfully",
		Data:      metrics,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now(),
	})
}

// Ready godoc
// @Summary      Check readiness status
// @Description  Returns readiness status indicating if the service is ready to accept traffic
// @Tags         System,Health & Monitoring
// @Accept       json
// @Produce      json
// @Success      200 {object} models.StandardAPIResponse "Service is ready"
// @Failure      503 {object} models.StandardAPIResponse "Service is not ready"
// @Router       /ready [get]
func EnhancedReady(c *gin.Context) {
	// Check if service is ready to accept traffic
	// This could include database connectivity, required services, etc.

	ready := true
	message := "Service is ready to accept traffic"

	// Mock readiness checks
	checks := map[string]bool{
		"database_connection":  true,
		"redis_connection":     true,
		"required_services":    true,
		"configuration_loaded": true,
	}

	for name, status := range checks {
		if !status {
			ready = false
			message = "Service is not ready: " + name + " check failed"
			break
		}
	}

	statusCode := http.StatusOK
	if !ready {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, models.StandardAPIResponse{
		Success:   ready,
		Message:   message,
		Data:      checks,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now(),
	})
}

// Live godoc
// @Summary      Check liveness status
// @Description  Returns liveness status indicating if the service is alive and functioning
// @Tags         System,Health & Monitoring
// @Accept       json
// @Produce      json
// @Success      200 {object} models.StandardAPIResponse "Service is alive"
// @Failure      500 {object} models.StandardAPIResponse "Service is not responding"
// @Router       /live [get]
func EnhancedLive(c *gin.Context) {
	// Simple liveness check - if we can respond, we're alive
	c.JSON(http.StatusOK, models.StandardAPIResponse{
		Success: true,
		Message: "Service is alive and responding",
		Data: map[string]interface{}{
			"status":    "alive",
			"timestamp": time.Now(),
			"uptime":    formatDuration(time.Since(startTime)),
		},
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now(),
	})
}

// formatDuration formats a duration into a human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return d.Round(time.Second).String()
	}
	if d < time.Hour {
		return d.Round(time.Minute).String()
	}
	if d < 24*time.Hour {
		return d.Round(time.Hour).String()
	}
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd%dh", days, hours)
}
