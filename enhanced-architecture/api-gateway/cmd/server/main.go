package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Import docs package for swagger
	_ "github.com/Mir00r/X-Form-Backend/enhanced-architecture/api-gateway/docs"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status" example:"healthy"`
	Timestamp string `json:"timestamp" example:"2025-09-12T01:35:03+08:00"`
} // @name HealthResponse

// MetricsResponse represents the metrics response
type MetricsResponse struct {
	Metrics string `json:"metrics" example:"api_gateway_requests_total 0"`
} // @name MetricsResponse

// GatewayInfoResponse represents the gateway info response
type GatewayInfoResponse struct {
	Message string `json:"message" example:"Enhanced X-Form API Gateway"`
	Version string `json:"version" example:"1.0.0"`
	Path    string `json:"path" example:"/"`
} // @name GatewayInfoResponse

// @title Enhanced X-Form API Gateway
// @version 1.0.0
// @description This is the enhanced X-Form Backend API Gateway with comprehensive documentation
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// setupGinRoutes configures the Gin routes for the server
func setupGinRoutes(router *gin.Engine) {
	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", healthCheck)
		v1.GET("/metrics", metricsHandler)
	}

	// Root endpoints (backward compatibility)
	router.GET("/health", healthCheck)
	router.GET("/metrics", metricsHandler)
	router.GET("/", gatewayInfo)
}

// healthCheck godoc
// @Summary Health Check
// @Description Get the health status of the API Gateway
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
// @Router /api/v1/health [get]
func healthCheck(c *gin.Context) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
	}
	c.JSON(http.StatusOK, response)
}

// metricsHandler godoc
// @Summary Metrics
// @Description Get basic metrics from the API Gateway
// @Tags monitoring
// @Accept json
// @Produce plain
// @Success 200 {string} string "api_gateway_requests_total 0"
// @Router /metrics [get]
// @Router /api/v1/metrics [get]
func metricsHandler(c *gin.Context) {
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, "# Simple metrics\napi_gateway_requests_total 0\n")
}

// gatewayInfo godoc
// @Summary Gateway Information
// @Description Get information about the API Gateway
// @Tags info
// @Accept json
// @Produce json
// @Success 200 {object} GatewayInfoResponse
// @Router / [get]
func gatewayInfo(c *gin.Context) {
	response := GatewayInfoResponse{
		Message: "Enhanced X-Form API Gateway",
		Version: "1.0.0",
		Path:    c.Request.URL.Path,
	}
	c.JSON(http.StatusOK, response)
}

func main() {
	// Set Gin mode based on environment
	if os.Getenv("ENV") != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.Default()

	// Add basic middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Setup routes
	setupGinRoutes(router)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Enhanced API Gateway starting on port %s\n", port)
	fmt.Printf("📊 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("📈 Metrics: http://localhost:%s/metrics\n", port)
	fmt.Printf("📚 API Documentation: http://localhost:%s/swagger/index.html\n", port)

	// Start server
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
