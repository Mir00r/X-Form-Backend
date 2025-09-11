// Package main Enhanced X-Form API Gateway
//
// This is the main API Gateway for X-Form Backend system.
// It provides routing, authentication, and proxy functionality for all microservices.
//
//	@title			Enhanced X-Form API Gateway
//	@version		1.0.0
//	@description	API Gateway for X-Form Backend microservices architecture
//	@termsOfService	http://swagger.io/terms/
//
//	@contact.name	X-Form API Support
//	@contact.url	http://www.x-form.io/support
//	@contact.email	support@x-form.io
//
//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT
//
//	@host		localhost:8080
//	@BasePath	/
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.
//
//	@tag.name			Health
//	@tag.description	Health check endpoints
//
//	@tag.name			Metrics
//	@tag.description	Monitoring and metrics endpoints
//
//	@tag.name			Gateway
//	@tag.description	API Gateway information
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/Mir00r/X-Form-Backend/enhanced-architecture/api-gateway/docs"
)

// setupRoutes configures the HTTP routes for the server
func setupRoutes(mux *http.ServeMux) {
	// Health check endpoint
	mux.HandleFunc("/health", healthHandler)

	// Metrics endpoint
	mux.HandleFunc("/metrics", metricsHandler)

	// Swagger documentation
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// Default handler
	mux.HandleFunc("/", rootHandler)
}

// healthHandler handles health check requests
//
//	@Summary		Health Check
//	@Description	Returns the health status of the API Gateway
//	@Tags			Health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Health status"
//	@Router			/health [get]
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}

// metricsHandler handles metrics requests
//
//	@Summary		Metrics
//	@Description	Returns basic metrics for the API Gateway
//	@Tags			Metrics
//	@Accept			json
//	@Produce		text/plain
//	@Success		200	{string}	string	"Metrics data"
//	@Router			/metrics [get]
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "# Simple metrics\napi_gateway_requests_total 0\n")
}

// rootHandler handles root requests
//
//	@Summary		Gateway Information
//	@Description	Returns information about the API Gateway
//	@Tags			Gateway
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Gateway information"
//	@Router			/ [get]
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `{"message":"Enhanced X-Form API Gateway","version":"1.0.0","path":"%s"}`, r.URL.Path)
}

// Simple main.go that works without complex dependencies
func main() {
	// Create a simple HTTP server
	mux := http.NewServeMux()

	// Setup routes
	setupRoutes(mux)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("🚀 Enhanced API Gateway starting on port %s\n", port)
	fmt.Printf("📊 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("📈 Metrics: http://localhost:%s/metrics\n", port)
	fmt.Printf("📚 API Documentation: http://localhost:%s/swagger/index.html\n", port)

	// Start server
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
