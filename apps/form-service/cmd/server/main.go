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

	// Repository and Service layers (following Clean Architecture)
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/handlers"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/repository"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/service"
	"github.com/gin-gonic/gin"
)

// ApplicationContainer demonstrates Clean Architecture principles
// Implements Dependency Injection Container pattern
type ApplicationContainer struct {
	Config      *config.Config
	FormHandler *handlers.FormHandler
}

// NewApplicationContainer creates application dependencies following SOLID principles
// Single Responsibility: Each component has one reason to change
// Dependency Inversion: High-level modules don't depend on low-level modules
func NewApplicationContainer() (*ApplicationContainer, error) {
	// Load configuration (Infrastructure concern)
	cfg := config.Load()

	// Initialize database connection (Infrastructure layer)
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate database schema
	if err := database.Migrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Initialize repositories (Data Access Layer)
	// Repository Pattern: Abstracts data persistence concerns
	formRepo := repository.NewFormRepository(db)
	questionRepo := repository.NewQuestionRepository(db)

	// Initialize services (Business Logic Layer)
	// Service Layer Pattern: Encapsulates business rules and use cases
	formService := service.NewFormService(formRepo, questionRepo)

	// Initialize handlers (Presentation Layer)
	// Controller Pattern: Handles HTTP requests and responses
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

	// Configure environment-specific settings
	if container.Config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup and start HTTP server with graceful shutdown
	server := setupHTTPServer(container)
	startServerGracefully(server, container.Config.Port)
}

// setupHTTPServer configures the HTTP server with timeouts
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
	// Each middleware follows Single Responsibility Principle
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.Security())

	// Health check endpoint for monitoring and observability
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":       "healthy",
			"service":      "form-service",
			"timestamp":    time.Now().UTC().Format(time.RFC3339),
			"version":      "1.0.0",
			"architecture": "Clean Architecture with SOLID Principles",
		})
	})

	// API versioning for backward compatibility
	api := router.Group("/api/v1")
	{
		// Form resource routes following REST conventions
		forms := api.Group("/forms")
		{
			// CRUD operations with proper HTTP methods
			// Each route follows Interface Segregation Principle
			forms.POST("", middleware.AuthRequired(cfg.JWTSecret), formHandler.CreateForm)
			forms.GET("/:id", middleware.OptionalAuth(cfg.JWTSecret), formHandler.GetForm)
			forms.PUT("/:id", middleware.AuthRequired(cfg.JWTSecret), formHandler.UpdateForm)
			forms.DELETE("/:id", middleware.AuthRequired(cfg.JWTSecret), formHandler.DeleteForm)
			forms.POST("/:id/publish", middleware.AuthRequired(cfg.JWTSecret), formHandler.PublishForm)
		}
	}

	return router
}

// startServerGracefully starts the server with graceful shutdown support
// Follows Open/Closed Principle: Open for extension, closed for modification
func startServerGracefully(server *http.Server, port string) {
	// Start server in a goroutine for non-blocking execution
	go func() {
		log.Printf("🚀 Form service starting on port %s", port)
		log.Printf("📊 Environment: %s", os.Getenv("ENVIRONMENT"))
		log.Printf("🏗️  Architecture: Clean Architecture with SOLID Principles")
		log.Printf("🔧 Dependency Injection: Container Pattern")
		log.Printf("📋 Following SOLID Principles:")
		log.Printf("   • Single Responsibility: Each layer has one responsibility")
		log.Printf("   • Open/Closed: Open for extension, closed for modification")
		log.Printf("   • Liskov Substitution: Interfaces enable substitutability")
		log.Printf("   • Interface Segregation: Small, focused interfaces")
		log.Printf("   • Dependency Inversion: Depend on abstractions, not concretions")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Give outstanding requests time to complete (graceful shutdown)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}
