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

	"github.com/gorilla/mux"
	"github.com/kamkaiz/x-form-backend/collaboration-service/internal/auth"
	"github.com/kamkaiz/x-form-backend/collaboration-service/internal/config"
	redisService "github.com/kamkaiz/x-form-backend/collaboration-service/internal/redis"
	"github.com/kamkaiz/x-form-backend/collaboration-service/internal/websocket"
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

	logger.Info("Starting Real-Time Collaboration Service",
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

	// Setup HTTP router
	router := setupRoutes(hub, logger)

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
		logger.Info("Starting HTTP server", zap.String("addr", server.Addr))
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

// setupRoutes configures HTTP routes
func setupRoutes(hub *websocket.Hub, logger *zap.Logger) *mux.Router {
	router := mux.NewRouter()

	// WebSocket endpoint
	router.HandleFunc("/ws", hub.ServeWS).Methods("GET")

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"collaboration-service"}`))
	}).Methods("GET")

	// Metrics endpoint
	router.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := hub.GetMetrics()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := fmt.Sprintf(`{
			"totalConnections": %d,
			"activeConnections": %d,
			"totalRooms": %d,
			"activeRooms": %d,
			"messagesPerSecond": %d,
			"errorsPerSecond": %d
		}`,
			metrics.TotalConnections,
			metrics.ActiveConnections,
			metrics.TotalRooms,
			metrics.ActiveRooms,
			metrics.MessagesPerSecond,
			metrics.ErrorsPerSecond,
		)

		w.Write([]byte(response))
	}).Methods("GET")

	// CORS middleware
	router.Use(corsMiddleware)

	// Logging middleware
	router.Use(loggingMiddleware(logger))

	return router
}

// corsMiddleware handles CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			logger.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("query", r.URL.RawQuery),
				zap.Int("status", rw.statusCode),
				zap.Duration("duration", duration),
				zap.String("userAgent", r.Header.Get("User-Agent")),
				zap.String("remoteAddr", r.RemoteAddr),
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
