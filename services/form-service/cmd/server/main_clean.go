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

	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/config"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/database"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/middleware"

	// Original handlers for working version
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/handlers"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/repository"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/service"
	"github.com/gin-gonic/gin"
)

// ApplicationContainer demonstrates Clean Architecture principles
// while maintaining compatibility with existing code
type ApplicationContainer struct {
	Config      *config.Config
	FormHandler *handlers.FormHandler
}

// NewApplicationContainer creates a new application container with Clean Architecture principles
func NewApplicationContainer() (*ApplicationContainer, error) {
	// Load configuration (Infrastructure concern)
	cfg := config.Load()

	// Initialize database (Infrastructure layer)
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate database schema
	if err := database.Migrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Initialize Redis client (Infrastructure layer)
	redisClient := database.ConnectRedis(cfg.RedisURL)

	// Initialize repositories (Data Access Layer)
	// Following Repository Pattern - abstracts data access
	formRepo := repository.NewFormRepository(db)
	questionRepo := repository.NewQuestionRepository(db)

	// Initialize services (Business Logic Layer)
	// Following Service Layer Pattern - encapsulates business logic
	formService := service.NewFormService(formRepo, questionRepo, redisClient)

	// Initialize handlers (Presentation Layer)
	// Following MVC pattern - handles HTTP concerns
	formHandler := handlers.NewFormHandler(formService)

	return &ApplicationContainer{
		Config:      cfg,
		FormHandler: formHandler,
	}, nil
}

func main() {
	// Initialize application container with dependency injection
	container, err := NewApplicationContainer()
	if err != nil {
		log.Fatalf("Failed to initialize application container: %v", err)
	}

	// Configure Gin based on environment
	if container.Config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup and start HTTP server
	server := setupHTTPServer(container)
	startServerGracefully(server, container.Config.Port)
}

// setupHTTPServer configures the HTTP server with middleware and routes
func setupHTTPServer(container *ApplicationContainer) *http.Server {
	router := setupRouter(container)

	return &http.Server{
		Addr:         fmt.Sprintf(":%s", container.Config.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// setupRouter configures routes and middleware following RESTful principles
func setupRouter(container *ApplicationContainer) *gin.Engine {
	cfg := container.Config
	formHandler := container.FormHandler

	router := gin.New()

	// Apply middleware (Cross-cutting concerns)
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.Security())

	// Health check endpoint for monitoring
	router.GET("/health", healthCheckHandler)

	// API versioning for backward compatibility
	api := router.Group("/api/v1")
	{
		// Form resource routes following REST conventions
		forms := api.Group("/forms")
		{
			// CRUD operations with proper HTTP methods
			forms.POST("", middleware.AuthRequired(cfg.JWTSecret), formHandler.CreateForm)
			forms.GET("/:id", middleware.OptionalAuth(cfg.JWTSecret), formHandler.GetForm)
			forms.PUT("/:id", middleware.AuthRequired(cfg.JWTSecret), formHandler.UpdateForm)
			forms.DELETE("/:id", middleware.AuthRequired(cfg.JWTSecret), formHandler.DeleteForm)

			// Additional business operations
			forms.POST("/:id/publish", middleware.AuthRequired(cfg.JWTSecret), formHandler.PublishForm)
		}
	}

	return router
}

// healthCheckHandler provides service health information
func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":       "healthy",
		"service":      "form-service",
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"version":      "1.0.0",
		"architecture": "Clean Architecture with SOLID Principles",
	})
}

// startServerGracefully starts the server with graceful shutdown support
func startServerGracefully(server *http.Server, port string) {
	// Start server in a goroutine for non-blocking execution
	go func() {
		log.Printf("🚀 Form service starting on port %s", port)
		log.Printf("📊 Environment: %s", os.Getenv("ENVIRONMENT"))
		log.Printf("🏗️  Architecture: Clean Architecture with SOLID Principles")
		log.Printf("🔧 Dependency Injection: Container Pattern")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Give outstanding requests time to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}
