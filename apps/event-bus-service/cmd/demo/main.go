// Package main demonstrates the comprehensive observability implementation
// for the Event Bus Service with OpenTelemetry, Prometheus, and Error Tracking.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/config"
	"github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/telemetry"
)

// DemoConfig represents basic configuration for the demo
type DemoConfig struct {
	Environment string `json:"environment"`
	Version     string `json:"version"`
	Port        int    `json:"port"`
}

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Create demo configuration
	cfg := &config.Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Version:     getEnv("VERSION", "1.0.0-demo"),
	}

	logger.Info("Starting Event Bus Observability Demo",
		zap.String("environment", cfg.Environment),
		zap.String("version", cfg.Version))

	// Initialize telemetry provider
	telemetryProvider, err := telemetry.New(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize telemetry provider", zap.Error(err))
	}

	// Create HTTP server with observability middleware
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		health := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   cfg.Version,
			"telemetry": telemetryProvider.Health(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health)
	})

	// Demo endpoints to test observability
	mux.HandleFunc("/api/events", handleEvents(telemetryProvider, logger))
	mux.HandleFunc("/api/kafka", handleKafka(telemetryProvider, logger))
	mux.HandleFunc("/api/cdc", handleCDC(telemetryProvider, logger))
	mux.HandleFunc("/api/error", handleError(telemetryProvider, logger))

	// Metrics endpoint (Prometheus)
	mux.Handle("/metrics", telemetryProvider.Metrics().Handler())

	// Create server with middleware
	handler := telemetryProvider.HTTPMiddleware()(mux)
	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting HTTP server", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	// Shutdown telemetry
	if err := telemetryProvider.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown telemetry provider", zap.Error(err))
	}

	logger.Info("Server exited")
}

// handleEvents demonstrates event processing observability
func handleEvents(tp *telemetry.Provider, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		eventType := r.URL.Query().Get("type")
		source := r.URL.Query().Get("source")

		if eventType == "" {
			eventType = "demo_event"
		}
		if source == "" {
			source = "demo_source"
		}

		// Use event processing observability
		newCtx, cleanup := tp.EventProcessingObservability(ctx, eventType, source)
		defer cleanup(nil) // No error for successful demo

		// Simulate event processing
		time.Sleep(100 * time.Millisecond)

		// Log and respond
		logger.Info("Event processed",
			zap.String("event_type", eventType),
			zap.String("source", source))

		response := map[string]interface{}{
			"status":     "processed",
			"event_type": eventType,
			"source":     source,
			"timestamp":  time.Now().UTC(),
			"context":    fmt.Sprintf("%v", newCtx),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// handleKafka demonstrates Kafka operation observability
func handleKafka(tp *telemetry.Provider, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		operation := r.URL.Query().Get("operation")
		topic := r.URL.Query().Get("topic")

		if operation == "" {
			operation = "produce"
		}
		if topic == "" {
			topic = "demo-topic"
		}

		// Use Kafka observability
		newCtx, cleanup := tp.KafkaObservability(ctx, operation, topic, 0)
		defer cleanup(nil) // No error for successful demo

		// Simulate Kafka operation
		time.Sleep(50 * time.Millisecond)

		logger.Info("Kafka operation completed",
			zap.String("operation", operation),
			zap.String("topic", topic))

		response := map[string]interface{}{
			"status":    "completed",
			"operation": operation,
			"topic":     topic,
			"partition": 0,
			"timestamp": time.Now().UTC(),
			"context":   fmt.Sprintf("%v", newCtx),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// handleCDC demonstrates CDC operation observability
func handleCDC(tp *telemetry.Provider, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		connector := r.URL.Query().Get("connector")
		table := r.URL.Query().Get("table")
		operation := r.URL.Query().Get("operation")

		if connector == "" {
			connector = "postgres-connector"
		}
		if table == "" {
			table = "demo_table"
		}
		if operation == "" {
			operation = "insert"
		}

		// Use CDC observability
		newCtx, cleanup := tp.CDCObservability(ctx, connector, table, operation)
		defer cleanup(nil) // No error for successful demo

		// Simulate CDC processing
		time.Sleep(75 * time.Millisecond)

		logger.Info("CDC operation completed",
			zap.String("connector", connector),
			zap.String("table", table),
			zap.String("operation", operation))

		response := map[string]interface{}{
			"status":    "processed",
			"connector": connector,
			"table":     table,
			"operation": operation,
			"timestamp": time.Now().UTC(),
			"context":   fmt.Sprintf("%v", newCtx),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// handleError demonstrates error tracking
func handleError(tp *telemetry.Provider, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		errorType := r.URL.Query().Get("type")
		if errorType == "" {
			errorType = "demo"
		}

		// Create demo error
		err := fmt.Errorf("demo error: %s", errorType)

		// Capture error with telemetry
		eventID := tp.CaptureError(err, "demo", "error_endpoint", map[string]interface{}{
			"error_type": errorType,
			"timestamp":  time.Now().UTC(),
		})

		logger.Error("Demo error captured",
			zap.String("error_type", errorType),
			zap.String("event_id", eventID),
			zap.Error(err))

		// Return error response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		response := map[string]interface{}{
			"error":     err.Error(),
			"event_id":  eventID,
			"type":      errorType,
			"timestamp": time.Now().UTC(),
		}

		json.NewEncoder(w).Encode(response)
	}
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
