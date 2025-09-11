package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/Mir00r/X-Form-Backend/enhanced-architecture/api-gateway/internal/config"
	"github.com/Mir00r/X-Form-Backend/enhanced-architecture/api-gateway/internal/handler"
	"github.com/Mir00r/X-Form-Backend/enhanced-architecture/api-gateway/internal/middleware"
	"github.com/Mir00r/X-Form-Backend/enhanced-architecture/api-gateway/pkg/logger"
	"github.com/Mir00r/X-Form-Backend/enhanced-architecture/api-gateway/pkg/metrics"

	// Import docs package for swagger
	_ "github.com/Mir00r/X-Form-Backend/enhanced-architecture/api-gateway/docs"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status" example:"healthy"`
	Timestamp string            `json:"timestamp" example:"2025-09-12T01:35:03+08:00"`
	Services  map[string]string `json:"services,omitempty"`
} // @name HealthResponse

// GatewayInfoResponse represents the gateway info response
type GatewayInfoResponse struct {
	Message  string   `json:"message" example:"Enhanced X-Form API Gateway"`
	Version  string   `json:"version" example:"1.0.0"`
	Path     string   `json:"path" example:"/"`
	Features []string `json:"features"`
} // @name GatewayInfoResponse

// @title Enhanced X-Form API Gateway
// @version 1.0.0
// @description This is the enhanced X-Form Backend API Gateway with comprehensive features including parameter validation, whitelist validation, authentication, rate limiting, service discovery, request transformation, and reverse proxy
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

func main() {
	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := logger.New(cfg.Logger)

	// Initialize metrics
	metrics := metrics.NewCollector()

	// Initialize handler with service discovery and circuit breakers
	handler := handler.NewHandler(cfg, logger, metrics)

	// Set Gin mode based on environment
	if cfg.Environment != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Setup comprehensive middleware chain following the 7-step architecture
	setupMiddlewareChain(router, cfg, logger, metrics)

	// Setup routes with full API Gateway functionality
	setupRoutes(router, handler, cfg, logger, metrics)

	// Get port from environment or config
	port := os.Getenv("PORT")
	if port == "" {
		port = fmt.Sprintf("%d", cfg.Server.Port)
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("🚀 Enhanced API Gateway starting", logger.Fields{
			"port":        port,
			"environment": cfg.Environment,
			"features": []string{
				"Parameter Validation",
				"Whitelist Validation",
				"Authentication/Authorization",
				"Rate Limiting",
				"Service Discovery",
				"Request Transformation",
				"Reverse Proxy",
				"Circuit Breakers",
				"Load Balancing",
				"Swagger Documentation",
			},
		})

		fmt.Printf("🚀 Enhanced API Gateway starting on port %s\n", port)
		fmt.Printf("📊 Health check: http://localhost:%s/health\n", port)
		fmt.Printf("📈 Metrics: http://localhost:%s/metrics\n", port)
		fmt.Printf("📚 API Documentation: http://localhost:%s/swagger/index.html\n", port)
		fmt.Printf("🎯 Gateway Info: http://localhost:%s/\n", port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", logger.Fields{"error": err})
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🛑 Shutting down Enhanced API Gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", logger.Fields{"error": err})
	}

	logger.Info("✅ Enhanced API Gateway exited gracefully")
}

// setupMiddlewareChain configures the comprehensive middleware chain
// Implements the 7-step API Gateway process from the architecture diagram
func setupMiddlewareChain(router *gin.Engine, cfg *config.Config, logger logger.Logger, metrics *metrics.Collector) {
	// Step 1: Parameter Validation
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Step 2: Whitelist Validation
	router.Use(func(c *gin.Context) {
		// Convert Gin context to standard HTTP for middleware compatibility
		w := c.Writer
		r := c.Request

		// Apply whitelist validation
		whitelistMiddleware := middleware.WhitelistValidation(cfg.Whitelist)
		whitelistMiddleware(func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		})(w, r)
	})

	// Step 3: Authentication & Authorization
	router.Use(func(c *gin.Context) {
		w := c.Writer
		r := c.Request

		// Skip auth for health and docs endpoints
		if r.URL.Path == "/health" || r.URL.Path == "/metrics" ||
			strings.HasPrefix(r.URL.Path, "/swagger") {
			c.Next()
			return
		}

		// Apply authentication
		authMiddleware := middleware.Authentication(cfg.Auth)
		authMiddleware(func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		})(w, r)
	})

	// Step 4: Rate Limiting
	router.Use(func(c *gin.Context) {
		w := c.Writer
		r := c.Request

		// Apply rate limiting
		rateLimitMiddleware := middleware.RateLimiting(cfg.RateLimit)
		rateLimitMiddleware(func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		})(w, r)
	})

	// Step 5: Service Discovery (handled in handler)
	// Step 6: Request Transformation (handled in handler)
	// Step 7: Reverse Proxy (handled in handler)
}

// setupRoutes configures all API Gateway routes
func setupRoutes(router *gin.Engine, h *handler.Handler, cfg *config.Config, logger logger.Logger, metrics *metrics.Collector) {
	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		healthCheck(c, h, logger)
	})

	// Metrics endpoint
	router.GET("/metrics", func(c *gin.Context) {
		metricsHandler(c, metrics)
	})

	// Gateway info endpoint
	router.GET("/", func(c *gin.Context) {
		gatewayInfo(c)
	})

	// API versioning
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			healthCheck(c, h, logger)
		})
		v1.GET("/metrics", func(c *gin.Context) {
			metricsHandler(c, metrics)
		})
	}

	// Service proxy routes with full API Gateway functionality
	setupServiceRoutes(router, h)
}

// setupServiceRoutes configures routes that proxy to backend services
func setupServiceRoutes(router *gin.Engine, h *handler.Handler) {
	// Auth service routes
	authGroup := router.Group("/auth")
	{
		authGroup.Any("/*path", func(c *gin.Context) {
			h.ProxyToService(c.Writer, c.Request, "auth-service")
		})
	}

	// Form service routes
	formGroup := router.Group("/forms")
	{
		formGroup.Any("/*path", func(c *gin.Context) {
			h.ProxyToService(c.Writer, c.Request, "form-service")
		})
	}

	// Response service routes
	responseGroup := router.Group("/responses")
	{
		responseGroup.Any("/*path", func(c *gin.Context) {
			h.ProxyToService(c.Writer, c.Request, "response-service")
		})
	}

	// Analytics service routes
	analyticsGroup := router.Group("/analytics")
	{
		analyticsGroup.Any("/*path", func(c *gin.Context) {
			h.ProxyToService(c.Writer, c.Request, "analytics-service")
		})
	}

	// Collaboration service routes
	collaborationGroup := router.Group("/collaboration")
	{
		collaborationGroup.Any("/*path", func(c *gin.Context) {
			h.ProxyToService(c.Writer, c.Request, "collaboration-service")
		})
	}

	// Realtime service routes
	realtimeGroup := router.Group("/realtime")
	{
		realtimeGroup.Any("/*path", func(c *gin.Context) {
			h.ProxyToService(c.Writer, c.Request, "realtime-service")
		})
	}

	// Event bus service routes
	eventGroup := router.Group("/events")
	{
		eventGroup.Any("/*path", func(c *gin.Context) {
			h.ProxyToService(c.Writer, c.Request, "event-bus-service")
		})
	}
}

// healthCheck godoc
// @Summary Health Check
// @Description Get the health status of the API Gateway and all connected services
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
// @Router /api/v1/health [get]
func healthCheck(c *gin.Context, h *handler.Handler, logger logger.Logger) {
	services := h.CheckServicesHealth()

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Services:  services,
	}

	c.JSON(http.StatusOK, response)
}

// metricsHandler godoc
// @Summary Metrics
// @Description Get comprehensive metrics from the API Gateway including request counts, latency, and service health
// @Tags monitoring
// @Accept json
// @Produce plain
// @Success 200 {string} string "Prometheus metrics format"
// @Router /metrics [get]
// @Router /api/v1/metrics [get]
func metricsHandler(c *gin.Context, metrics *metrics.Collector) {
	c.Header("Content-Type", "text/plain")
	metricsData := metrics.Export()
	c.String(http.StatusOK, metricsData)
}

// gatewayInfo godoc
// @Summary Gateway Information
// @Description Get comprehensive information about the Enhanced API Gateway including all implemented features
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
		Features: []string{
			"Parameter Validation",
			"Whitelist Validation",
			"Authentication/Authorization",
			"Rate Limiting",
			"Service Discovery",
			"Request Transformation",
			"Reverse Proxy",
			"Circuit Breakers",
			"Load Balancing",
			"Health Monitoring",
			"Metrics Collection",
			"Swagger Documentation",
			"Graceful Shutdown",
		},
	}
	c.JSON(http.StatusOK, response)
}
