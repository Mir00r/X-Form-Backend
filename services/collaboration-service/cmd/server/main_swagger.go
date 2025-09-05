package main

// @title X-Form Collaboration Service API
// @version 1.0.0
// @description Real-time collaboration service for X-Form with WebSocket and HTTP endpoints. Enables multiple users to collaborate on forms in real-time with cursor tracking, live updates, and session management.
// @description
// @description ## 🏗️ Architecture Features
// @description - **Real-time WebSocket Communication** for instant collaboration
// @description - **Form Session Management** with Redis persistence
// @description - **Cursor Tracking** and live position sharing
// @description - **Question Management** with real-time updates
// @description - **JWT Authentication** with service-to-service communication
// @description - **Rate Limiting** and connection management
// @description
// @description ## 🔐 Security Features
// @description - **JWT Bearer Authentication** for WebSocket connections
// @description - **Rate limiting** per user and connection
// @description - **CORS** configuration for cross-origin requests
// @description - **Connection validation** and session management
// @description
// @description ## 🌐 WebSocket Events
// @description The service primarily operates through WebSocket connections at `/ws` endpoint.
// @description
// @description ### Client → Server Events:
// @description - `join:form` - Join a form collaboration session
// @description - `leave:form` - Leave a form collaboration session
// @description - `cursor:update` - Update cursor position
// @description - `question:update` - Update existing question
// @description - `question:create` - Create new question
// @description - `question:delete` - Delete question
// @description - `ping` - Keep-alive ping
// @description
// @description ### Server → Client Events:
// @description - `join:form:response` - Response to join request
// @description - `user:joined` - Notify when user joins
// @description - `user:left` - Notify when user leaves
// @description - `cursor:update` - Broadcast cursor updates
// @description - `question:update` - Broadcast question updates
// @description - `error` - Error notifications
// @description - `pong` - Response to ping
// @description
// @description ## 🚀 Getting Started
// @description 1. Connect to WebSocket endpoint with JWT token
// @description 2. Send `join:form` event to join collaboration session
// @description 3. Listen for real-time events and updates
// @description 4. Send updates for cursor movement and form changes
// @description
// @termsOfService https://xform.com/terms
// @contact.name X-Form Development Team
// @contact.email dev@xform.com
// @contact.url https://github.com/Mir00r/X-Form-Backend
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token for authentication. Format: "Bearer {token}"

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/kamkaiz/x-form-backend/collaboration-service/internal/auth"
	"github.com/kamkaiz/x-form-backend/collaboration-service/internal/config"
	redisService "github.com/kamkaiz/x-form-backend/collaboration-service/internal/redis"
	"github.com/kamkaiz/x-form-backend/collaboration-service/internal/websocket"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	logger.Info("Starting Real-Time Collaboration Service with Swagger Documentation",
		zap.String("version", "1.0.0"),
		zap.String("port", cfg.Server.Port),
		zap.String("environment", cfg.Server.Environment))

	// Initialize Redis service
	redis, err := redisService.NewService(&cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to initialize Redis service", zap.Error(err))
	}
	defer redis.Close()

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redis.Ping(ctx); err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	logger.Info("Redis connection established")

	// Initialize auth service
	authService := auth.NewService(
		cfg.Auth.JWTSecret,
		cfg.Auth.ServiceSecret,
		cfg.Auth.JWTExpiration,
	)

	// Initialize WebSocket hub
	hub := websocket.NewHub(redis, authService, &cfg.WebSocket, logger)

	// Start WebSocket hub
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	// Setup HTTP router with Swagger
	router := setupRoutesWithSwagger(hub, logger)

	// Setup HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting HTTP server with Swagger UI",
			zap.String("addr", server.Addr),
			zap.String("swagger_url", fmt.Sprintf("http://localhost:%s/swagger/", cfg.Server.Port)))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped")
}

// setupRoutesWithSwagger configures HTTP routes with Swagger documentation
func setupRoutesWithSwagger(hub *websocket.Hub, logger *zap.Logger) *mux.Router {
	router := mux.NewRouter()

	// Swagger documentation endpoint
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// API documentation redirect
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})

	// API v1 routes
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	// WebSocket endpoint
	apiV1.HandleFunc("/ws", hub.ServeWS).Methods("GET")

	// Health check endpoint
	apiV1.HandleFunc("/health", getHealthCheck).Methods("GET")

	// Metrics endpoint
	apiV1.HandleFunc("/metrics", getMetrics(hub)).Methods("GET")

	// Room management endpoints
	apiV1.HandleFunc("/rooms", getRooms(hub)).Methods("GET")
	apiV1.HandleFunc("/rooms/{formId}", getRoomDetails(hub)).Methods("GET")
	apiV1.HandleFunc("/rooms/{formId}/users", getRoomUsers(hub)).Methods("GET")

	// User session endpoints
	apiV1.HandleFunc("/sessions", getUserSessions(hub)).Methods("GET")
	apiV1.HandleFunc("/sessions/{userId}", getUserSessionDetails(hub)).Methods("GET")

	// CORS middleware
	router.Use(corsMiddleware)

	// Logging middleware
	router.Use(loggingMiddleware(logger))

	return router
}
