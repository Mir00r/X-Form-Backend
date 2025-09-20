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

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	_ "github.com/Mir00r/X-Form-Backend/services/api-gateway/docs" // Import for swagger
	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/config"
	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/discovery"
	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/handlers"
	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/jwt"
	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/middleware"
	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/traefik"
	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/tyk"
	"github.com/Mir00r/X-Form-Backend/shared/observability"
)

// @title           X-Form API Gateway
// @version         2.0.0
// @description     Comprehensive API Gateway for X-Form microservices architecture with advanced features including authentication, form management, response collection, real-time analytics, and comprehensive monitoring.
// @description
// @description     ## Features
// @description     - **Authentication & Authorization**: JWT-based authentication with role-based access control
// @description     - **Form Management**: Create, update, delete, and publish forms with advanced field types
// @description     - **Response Collection**: Collect and validate form responses with file uploads
// @description     - **Real-time Analytics**: Comprehensive analytics with insights and dashboard capabilities
// @description     - **Data Export**: Export responses in multiple formats (CSV, Excel, JSON, PDF)
// @description     - **Rate Limiting**: Advanced rate limiting with per-user and per-endpoint controls
// @description     - **Monitoring**: Health checks, metrics, and performance monitoring
// @description     - **Security**: CORS, security headers, request validation, and data protection
// @description
// @description     ## Authentication
// @description     This API uses JWT (JSON Web Tokens) for authentication. Include the token in the Authorization header:
// @description     ```
// @description     Authorization: Bearer <your-jwt-token>
// @description     ```
// @description
// @description     ## Rate Limiting
// @description     API requests are rate limited to ensure fair usage:
// @description     - Anonymous users: 100 requests per hour
// @description     - Authenticated users: 1000 requests per hour
// @description     - Premium users: 10000 requests per hour
// @description
// @description     ## Error Handling
// @description     The API returns consistent error responses with detailed information:
// @description     - 4xx errors: Client-side issues (validation, authentication, etc.)
// @description     - 5xx errors: Server-side issues
// @description
// @description     ## Pagination
// @description     List endpoints support pagination with the following parameters:
// @description     - `page`: Page number (default: 1)
// @description     - `limit`: Items per page (default: 10, max: 100)
// @description     - `sort`: Sort field
// @description     - `order`: Sort order (asc/desc)

// @termsOfService  https://x-form.com/terms
// @contact.name    X-Form API Support Team
// @contact.url     https://x-form.com/support
// @contact.email   api-support@x-form.com
// @license.name    MIT License
// @license.url     https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @schemes   http https
// @produce   application/json
// @consumes  application/json

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and your JWT token. Example: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API key for service-to-service communication

func main() {
	// Initialize structured logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg := config.Load()

	// Initialize observability
	obsConfig := observability.DefaultConfig("api-gateway")
	obsProvider, err := observability.New(obsConfig, logger)
	if err != nil {
		logger.Fatal("Failed to initialize observability", zap.Error(err))
	}
	defer obsProvider.Shutdown(context.Background())

	logger.Info("X-Form API Gateway starting",
		zap.String("version", cfg.Version),
		zap.String("environment", cfg.Environment),
	)

	// Initialize services
	jwtService := jwt.NewJWTService(cfg)
	serviceDiscovery := discovery.NewServiceDiscovery(cfg)
	traefikService := traefik.NewTraefikService(cfg)
	tykService := tyk.NewTykService(cfg)

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	r := gin.Default()

	// Add observability middleware first
	r.Use(obsProvider.GinMiddleware())

	// Add core middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.Recovery())

	// Add integration middleware
	if cfg.Traefik.Enabled {
		r.Use(traefikService.TraefikMiddleware())
		log.Printf("✅ Traefik integration enabled")
	}

	if cfg.Tyk.Enabled {
		r.Use(tykService.TykMiddleware())
		log.Printf("✅ Tyk API management enabled")
	}

	// Health check endpoints
	r.GET("/health", traefikService.HealthCheck())
	r.GET("/ready", handlers.EnhancedReady)
	r.GET("/live", handlers.EnhancedLive)

	// Metrics endpoint - using observability provider
	r.GET("/metrics", gin.WrapH(obsProvider.Metrics().Handler()))
	logger.Info("Metrics endpoint enabled", zap.String("path", "/metrics"))

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})

	// Service discovery endpoints
	serviceGroup := r.Group("/api/gateway")
	{
		serviceGroup.GET("/services", serviceDiscovery.GetServicesEndpoint())
		serviceGroup.GET("/services/:service/health", serviceDiscovery.GetServiceHealthEndpoint())
		serviceGroup.GET("/services/metrics", serviceDiscovery.GetServiceMetricsEndpoint())

		// JWT validation endpoints
		serviceGroup.POST("/jwt/validate", jwtService.ValidateJWTEndpoint())
		serviceGroup.GET("/jwt/jwks", jwtService.GetJWKSEndpoint())

		// Traefik configuration endpoint
		if cfg.Traefik.Enabled {
			serviceGroup.GET("/traefik/config", func(c *gin.Context) {
				c.JSON(http.StatusOK, traefikService.GetTraefikConfig())
			})
		}
	}

	// API v1 routes with JWT authentication
	v1 := r.Group("/api/v1")
	v1.Use(jwtService.JWTMiddleware())
	{
		// Authentication service routes (proxied)
		auth := v1.Group("/auth")
		{
			auth.Any("/*path", serviceDiscovery.ProxyRequest("auth-service"))
		}

		// Form service routes (proxied)
		forms := v1.Group("/forms")
		{
			forms.Any("/*path", serviceDiscovery.ProxyRequest("form-service"))
		}

		// Response service routes (proxied)
		responses := v1.Group("/responses")
		{
			responses.Any("/*path", serviceDiscovery.ProxyRequest("response-service"))
		}

		// Analytics service routes (proxied)
		analytics := v1.Group("/analytics")
		{
			analytics.Any("/*path", serviceDiscovery.ProxyRequest("analytics-service"))
		}

		// Collaboration service routes (proxied)
		collaboration := v1.Group("/collaboration")
		{
			collaboration.Any("/*path", serviceDiscovery.ProxyRequest("collaboration-service"))
		}

		// Realtime service routes (proxied via WebSocket)
		realtime := v1.Group("/realtime")
		{
			realtime.Any("/*path", serviceDiscovery.ProxyRequest("realtime-service"))
		}
	}

	// Initialize Tyk API definitions if enabled
	if cfg.Tyk.Enabled {
		go initializeTykAPIs(tykService, cfg)
	}

	// Create HTTP server with configured timeouts
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🌐 Server starting on %s:%s", cfg.Server.Host, cfg.Server.Port)
		log.Printf("📚 Swagger documentation: http://localhost:%s/swagger/index.html", cfg.Server.Port)
		log.Printf("💚 Health check: http://localhost:%s/health", cfg.Server.Port)

		if cfg.Observability.Metrics.Enabled {
			log.Printf("📊 Metrics: http://localhost:%s%s", cfg.Server.Port, cfg.Observability.Metrics.Path)
		}

		log.Printf("🔐 JWT validation: %s", getJWTMethod(cfg))
		log.Printf("⚡ Features: Traefik(%v), Tyk(%v), JWKS(%v), mTLS(%v)",
			cfg.Traefik.Enabled, cfg.Tyk.Enabled,
			cfg.Security.JWKS.Endpoint != "", cfg.Security.MTLS.Enabled)

		var err error
		if cfg.Server.TLS.Enabled {
			log.Printf("🔒 Starting HTTPS server with TLS")
			err = server.ListenAndServeTLS(cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile)
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop service discovery
	serviceDiscovery.Stop()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
	} else {
		log.Println("✅ Server exited gracefully")
	}
}

// initializeTykAPIs initializes API definitions in Tyk
func initializeTykAPIs(tykService *tyk.TykService, cfg *config.Config) {
	ctx := context.Background()

	// Define service configurations
	services := map[string]string{
		"auth-service":          "/api/v1/auth",
		"form-service":          "/api/v1/forms",
		"response-service":      "/api/v1/responses",
		"analytics-service":     "/api/v1/analytics",
		"collaboration-service": "/api/v1/collaboration",
		"realtime-service":      "/api/v1/realtime",
	}

	// Create API definitions for each service
	for serviceName, listenPath := range services {
		serviceConfig, exists := getServiceConfig(cfg, serviceName)
		if !exists {
			continue
		}

		apiDef := tykService.GetAPIDefinitionForService(serviceName, serviceConfig.URL, listenPath)

		if err := tykService.CreateAPIDefinition(ctx, apiDef); err != nil {
			log.Printf("❌ Failed to create Tyk API definition for %s: %v", serviceName, err)
		} else {
			log.Printf("✅ Created Tyk API definition for %s", serviceName)
		}
	}

	// Reload API definitions
	if err := tykService.ReloadAPIDefinitions(ctx); err != nil {
		log.Printf("❌ Failed to reload Tyk API definitions: %v", err)
	} else {
		log.Printf("✅ Reloaded Tyk API definitions")
	}
}

// getServiceConfig returns service configuration by name
func getServiceConfig(cfg *config.Config, serviceName string) (config.ServiceConfig, bool) {
	switch serviceName {
	case "auth-service":
		return cfg.Services.AuthService, true
	case "form-service":
		return cfg.Services.FormService, true
	case "response-service":
		return cfg.Services.ResponseService, true
	case "analytics-service":
		return cfg.Services.AnalyticsService, true
	case "collaboration-service":
		return cfg.Services.CollaborationService, true
	case "realtime-service":
		return cfg.Services.RealtimeService, true
	default:
		return config.ServiceConfig{}, false
	}
}

// getJWTMethod returns a description of the JWT validation method
func getJWTMethod(cfg *config.Config) string {
	if cfg.Security.JWKS.Endpoint != "" {
		return fmt.Sprintf("JWKS (%s)", cfg.Security.JWKS.Endpoint)
	}
	return "Shared Secret"
}
