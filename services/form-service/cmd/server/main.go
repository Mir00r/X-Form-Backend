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
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/handlers"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/middleware"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/repository"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate database schema
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize Redis client
	redisClient := database.ConnectRedis(cfg.RedisURL)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	formRepo := repository.NewFormRepository(db)
	questionRepo := repository.NewQuestionRepository(db)

	// Initialize services
	formService := service.NewFormService(formRepo, questionRepo, redisClient)

	// Initialize handlers
	formHandler := handlers.NewFormHandler(formService)

	// Setup router
	router := setupRouter(cfg, formHandler)

	// Setup server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Form service starting on port %s", cfg.Port)
		log.Printf("Environment: %s", cfg.Environment)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func setupRouter(cfg *config.Config, formHandler *handlers.FormHandler) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.Security())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "form-service",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   "1.0.0",
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Form routes
		forms := api.Group("/forms")
		{
			forms.POST("", middleware.AuthRequired(cfg.JWTSecret), formHandler.CreateForm)
			forms.GET("", middleware.AuthRequired(cfg.JWTSecret), formHandler.GetUserForms)
			forms.GET("/:id", middleware.OptionalAuth(cfg.JWTSecret), formHandler.GetForm)
			forms.PUT("/:id", middleware.AuthRequired(cfg.JWTSecret), formHandler.UpdateForm)
			forms.DELETE("/:id", middleware.AuthRequired(cfg.JWTSecret), formHandler.DeleteForm)
			forms.POST("/:id/publish", middleware.AuthRequired(cfg.JWTSecret), formHandler.PublishForm)
			forms.POST("/:id/unpublish", middleware.AuthRequired(cfg.JWTSecret), formHandler.UnpublishForm)
		}
	}

	return router
}
