package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/config"
	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/gateway"
	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/middleware"
	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/proxy"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logging
	logrus.SetLevel(logrus.InfoLevel)
	if cfg.Environment == "development" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize proxy manager
	proxyManager := proxy.NewManager(cfg)

	// Initialize gateway
	gatewayInstance := gateway.New(cfg, proxyManager)

	// Setup router
	router := setupRouter(cfg, gatewayInstance)

	// Setup metrics server
	go func() {
		metricsRouter := gin.New()
		metricsRouter.GET("/metrics", gin.WrapH(promhttp.Handler()))
		metricsRouter.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "healthy"})
		})

		log.Printf("Metrics server starting on port %s", cfg.MetricsPort)
		if err := http.ListenAndServe(":"+cfg.MetricsPort, metricsRouter); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	// Setup main server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("API Gateway starting on port %s", cfg.Port)
		log.Printf("Environment: %s", cfg.Environment)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down API Gateway...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("API Gateway exiting")
}

func setupRouter(cfg *config.Config, gateway *gateway.Gateway) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: gin.DefaultLogFormatter,
		SkipPaths: []string{"/health", "/metrics"},
	}))
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.Metrics())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "api-gateway",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   cfg.Version,
		})
	})

	// API versioned routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (no auth required)
		authGroup := v1.Group("/auth")
		{
			authGroup.Any("/*path", gateway.ProxyToAuth)
		}

		// User routes (auth required)
		userGroup := v1.Group("/user")
		userGroup.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			userGroup.Any("/*path", gateway.ProxyToAuth)
		}

		// Form routes (auth required)
		formsGroup := v1.Group("/forms")
		formsGroup.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			formsGroup.Any("/*path", gateway.ProxyToForm)
		}

		// Response routes (mixed auth)
		responsesGroup := v1.Group("/responses")
		{
			// Public submission endpoint
			responsesGroup.POST("/:formId/submit", gateway.ProxyToResponse)

			// Protected endpoints
			protectedResponses := responsesGroup.Group("")
			protectedResponses.Use(middleware.AuthRequired(cfg.JWTSecret))
			{
				protectedResponses.Any("/*path", gateway.ProxyToResponse)
			}
		}

		// Analytics routes (auth required)
		analyticsGroup := v1.Group("/analytics")
		analyticsGroup.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			analyticsGroup.Any("/*path", gateway.ProxyToAnalytics)
		}

		// File routes (auth required)
		filesGroup := v1.Group("/files")
		filesGroup.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			filesGroup.Any("/*path", gateway.ProxyToFile)
		}
	}

	// Public form access (no /api prefix)
	publicForms := router.Group("/forms")
	{
		publicForms.GET("/:formId", gateway.ProxyToForm)
		publicForms.POST("/:formId/submit", gateway.ProxyToResponse)
	}

	// WebSocket proxy (handled by Traefik directly to real-time service)
	// This is just for documentation - actual WS traffic bypasses gateway
	router.GET("/ws/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":   "WebSocket traffic is handled directly by Traefik",
			"endpoint":  "ws://ws.xform.dev/forms/:id/updates",
			"protocols": []string{"ws", "wss"},
		})
	})

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Route not found",
			"message": fmt.Sprintf("Cannot %s %s", c.Request.Method, c.Request.URL.Path),
		})
	})

	return router
}
