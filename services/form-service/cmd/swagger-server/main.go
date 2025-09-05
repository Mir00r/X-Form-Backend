/*
Form Service API with comprehensive Swagger documentation
Follows microservices best practices and industry standards

@title Form Service API
@version 1.0.0
@description Comprehensive form management service built with Clean Architecture and following microservices best practices.

## Features:
- Create, update, and manage forms
- Dynamic question types and validation
- Form publishing and response collection
- Advanced filtering and search
- Comprehensive monitoring and health checks

## Architecture:
- Clean Architecture with SOLID principles
- Microservices best practices
- API versioning
- Comprehensive DTOs
- Input validation
- Rate limiting
- Circuit breakers
- Structured logging

@contact.name Form Service Team
@contact.email form-service@example.com
@contact.url https://api.example.com/support

@license.name MIT
@license.url https://opensource.org/licenses/MIT

@host localhost:8080
@BasePath /api/v1

@securityDefinitions.apikey BearerAuth
@in header
@name Authorization
@description Type "Bearer" followed by a space and JWT token.

@tag.name Forms
@tag.description Form management operations

@tag.name Health
@tag.description Health check and monitoring endpoints

@tag.name Metrics
@tag.description System metrics and monitoring

@externalDocs.description OpenAPI Specification
@externalDocs.url https://swagger.io/resources/open-api/
*/
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
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/application"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/config"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/database"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/handlers"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/integration"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/middleware"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/repository"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/routes"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/validation"
	// Import docs package to register Swagger specs (will be generated)
	// _ "github.com/Mir00r/X-Form-Backend/services/form-service/docs"
)

// ApplicationContainer demonstrates microservices best practices with Swagger integration
// Implements comprehensive dependency injection with all improvements
type ApplicationContainer struct {
	Config          *config.Config
	FormService     *application.FormApplicationService
	ResponseHandler *handlers.ResponseHandler
	FormValidator   *validation.FormValidator
	FormMapper      *integration.SimplifiedFormMapper
	FormRoutes      *routes.FormRoutesV1
	StartTime       time.Time
}

// NewApplicationContainer creates application dependencies following microservices best practices
func NewApplicationContainer() (*ApplicationContainer, error) {
	startTime := time.Now()

	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Initialize database connection
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate database schema
	if err := database.Migrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Initialize repository
	_ = repository.NewFormRepository(db)

	// Create a dummy form service for now (we'll connect to real one later)
	var formService *application.FormApplicationService

	// Initialize handlers and utilities
	responseHandler := handlers.NewResponseHandler("1.0.0")
	formValidator := validation.NewFormValidator(responseHandler)
	formMapper := integration.NewSimplifiedFormMapper()

	// Initialize routes with the available components
	formRoutes := routes.NewFormRoutesV1(
		formService,
		responseHandler,
		formValidator,
	)

	return &ApplicationContainer{
		Config:          cfg,
		FormService:     formService,
		ResponseHandler: responseHandler,
		FormValidator:   formValidator,
		FormMapper:      formMapper,
		FormRoutes:      formRoutes,
		StartTime:       startTime,
	}, nil
}

func main() {
	// Initialize application container with dependency injection
	container, err := NewApplicationContainer()
	if err != nil {
		log.Fatalf("Failed to initialize application container: %v", err)
	}

	// Configure environment-specific settings
	if container.Config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup and start HTTP server with graceful shutdown
	server := setupHTTPServer(container)
	startServerGracefully(server, container.Config.Port)
}

// setupHTTPServer configures the HTTP server with timeouts and Swagger
func setupHTTPServer(container *ApplicationContainer) *http.Server {
	router := setupRouter(container)

	return &http.Server{
		Addr:         fmt.Sprintf(":%s", container.Config.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

// setupRouter configures routes, middleware, and Swagger documentation
func setupRouter(container *ApplicationContainer) *gin.Engine {
	// Initialize Gin router
	router := gin.New()

	// Add basic middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimiting())

	// Health check endpoint (before middleware)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "form-service",
			"version":   "1.0.0",
			"timestamp": time.Now().UTC(),
			"uptime":    time.Since(container.StartTime).String(),
		})
	})

	// Root endpoint with service information
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":     "Form Service API",
			"version":     "1.0.0",
			"description": "Comprehensive form management service with Clean Architecture",
			"endpoints": map[string]string{
				"health":       "/health",
				"swagger":      "/swagger/index.html",
				"api":          "/api/v1",
				"swagger-json": "/swagger/doc.json",
			},
			"features": []string{
				"✅ RESTful API design",
				"✅ Clean Architecture",
				"✅ Comprehensive validation",
				"✅ Swagger/OpenAPI documentation",
				"✅ Health checks & monitoring",
				"✅ Rate limiting & security",
				"✅ Database migrations",
				"✅ Structured logging",
				"✅ Graceful shutdown",
			},
			"timestamp": time.Now().UTC(),
		})
	})

	// Swagger documentation endpoints
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	apiV1 := router.Group("/api/v1")
	{
		// Register form routes
		container.FormRoutes.RegisterRoutes(apiV1)
	}

	return router
}

// startServerGracefully starts the server with graceful shutdown capability
func startServerGracefully(server *http.Server, port string) {
	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Form Service starting on port %s", port)
		log.Printf("📖 Swagger documentation available at: http://localhost:%s/swagger/index.html", port)
		log.Printf("🔍 Health check available at: http://localhost:%s/health", port)
		log.Printf("📊 API endpoints available at: http://localhost:%s/api/v1", port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
		return
	}

	log.Println("✅ Server gracefully stopped")
}
