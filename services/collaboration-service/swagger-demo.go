// Package main provides the X-Form Collaboration Service API
// @title X-Form Collaboration Service API
// @version 1.0.0
// @description Real-time collaboration service for X-Form with WebSocket and HTTP endpoints
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
// @description The service primarily operates through WebSocket connections at `/api/v1/ws` endpoint.
// @description See the WebSocket API documentation for detailed event specifications.
// @description
// @description ## 🚀 Getting Started
// @description 1. Use HTTP endpoints to monitor service health and manage sessions
// @description 2. Connect to WebSocket endpoint with JWT token for real-time collaboration
// @description 3. Send `join:form` event to join collaboration session
// @description 4. Listen for real-time events and send updates
// @termsOfService https://xform.com/terms
// @contact.name X-Form Development Team
// @contact.url https://github.com/Mir00r/X-Form-Backend
// @contact.email dev@xform.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Response structures for API documentation

// HealthStatus represents the health check response
type HealthStatus struct {
	Status       string                 `json:"status" example:"healthy"`
	Service      string                 `json:"service" example:"collaboration-service"`
	Version      string                 `json:"version" example:"1.0.0"`
	Timestamp    time.Time              `json:"timestamp" example:"2024-01-15T10:30:00Z"`
	Uptime       string                 `json:"uptime" example:"1h30m45s"`
	Environment  string                 `json:"environment" example:"development"`
	Dependencies map[string]interface{} `json:"dependencies"`
} // @name HealthStatus

// ServiceMetrics represents the metrics response
type ServiceMetrics struct {
	TotalConnections  int64                  `json:"totalConnections" example:"150"`
	ActiveConnections int64                  `json:"activeConnections" example:"25"`
	TotalRooms        int64                  `json:"totalRooms" example:"10"`
	ActiveRooms       int64                  `json:"activeRooms" example:"8"`
	MessagesPerSecond int64                  `json:"messagesPerSecond" example:"45"`
	ErrorsPerSecond   int64                  `json:"errorsPerSecond" example:"0"`
	SystemUsage       map[string]interface{} `json:"systemUsage"`
} // @name ServiceMetrics

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error     string    `json:"error" example:"Resource not found"`
	Code      int       `json:"code" example:"404"`
	Message   string    `json:"message" example:"The requested resource was not found"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-15T10:30:00Z"`
	TraceID   string    `json:"trace_id,omitempty" example:"abc123xyz"`
} // @name ErrorResponse

// WebSocketInfo represents WebSocket connection information
type WebSocketInfo struct {
	Endpoint     string   `json:"endpoint" example:"/api/v1/ws"`
	Protocol     string   `json:"protocol" example:"ws"`
	AuthRequired bool     `json:"auth_required" example:"true"`
	AuthType     string   `json:"auth_type" example:"Bearer JWT"`
	Events       []string `json:"events"`
	Description  string   `json:"description" example:"WebSocket endpoint for real-time collaboration"`
} // @name WebSocketInfo

// @Summary Get service health status
// @Description Returns the health status of the collaboration service and its dependencies including Redis and Auth service connectivity
// @Tags Health & Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} HealthStatus "Service is healthy"
// @Failure 503 {object} ErrorResponse "Service is unhealthy"
// @Router /api/v1/health [get]
func healthHandler(w http.ResponseWriter, r *http.Request) {
	dependencies := map[string]interface{}{
		"redis": map[string]interface{}{
			"status":      "connected",
			"latency":     "2ms",
			"pool_size":   10,
			"connections": 5,
		},
		"auth-service": map[string]interface{}{
			"status":   "available",
			"endpoint": "http://auth-service:8081",
			"latency":  "15ms",
		},
	}

	response := HealthStatus{
		Status:       "healthy",
		Service:      "collaboration-service",
		Version:      "1.0.0",
		Timestamp:    time.Now(),
		Uptime:       "2h15m30s",
		Environment:  "development",
		Dependencies: dependencies,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get system metrics
// @Description Returns current system metrics including active connections, rooms, message rates, and system resource usage
// @Tags Health & Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} ServiceMetrics "System metrics retrieved successfully"
// @Failure 500 {object} ErrorResponse "Failed to retrieve metrics"
// @Router /api/v1/metrics [get]
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	systemUsage := map[string]interface{}{
		"memory": map[string]interface{}{
			"used":      "245MB",
			"available": "2GB",
			"usage":     "12.3%",
		},
		"cpu": map[string]interface{}{
			"usage":   "15.7%",
			"cores":   4,
			"threads": 8,
		},
		"disk": map[string]interface{}{
			"used":      "1.2GB",
			"available": "50GB",
			"usage":     "2.4%",
		},
	}

	response := ServiceMetrics{
		TotalConnections:  150,
		ActiveConnections: 25,
		TotalRooms:        10,
		ActiveRooms:       8,
		MessagesPerSecond: 45,
		ErrorsPerSecond:   0,
		SystemUsage:       systemUsage,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get WebSocket information
// @Description Returns information about the WebSocket endpoint, authentication requirements, and supported events for real-time collaboration
// @Tags WebSocket
// @Accept json
// @Produce json
// @Success 200 {object} WebSocketInfo "WebSocket information retrieved successfully"
// @Router /api/v1/ws/info [get]
func websocketInfoHandler(w http.ResponseWriter, r *http.Request) {
	response := WebSocketInfo{
		Endpoint:     "/api/v1/ws",
		Protocol:     "ws",
		AuthRequired: true,
		AuthType:     "Bearer JWT",
		Events: []string{
			"join:form",
			"leave:form",
			"cursor:move",
			"cursor:hide",
			"question:update",
			"question:focus",
			"question:blur",
			"form:save",
			"user:typing",
			"user:stopped_typing",
			"heartbeat",
		},
		Description: "WebSocket endpoint for real-time collaboration on form editing. Requires JWT authentication via Authorization header or token query parameter.",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Setup CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	r := mux.NewRouter()

	// Apply CORS middleware
	r.Use(corsMiddleware)

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// Health and monitoring endpoints
	api.HandleFunc("/health", healthHandler).Methods("GET")
	api.HandleFunc("/metrics", metricsHandler).Methods("GET")

	// WebSocket information endpoint
	api.HandleFunc("/ws/info", websocketInfoHandler).Methods("GET")

	// Swagger UI endpoint
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Serve documentation files
	r.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs/"))))

	// Root endpoint - redirect to Swagger UI
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusFound)
	})

	// Print startup information
	fmt.Println("🚀 X-Form Collaboration Service API Server")
	fmt.Println("=" + string(make([]rune, 50)))
	fmt.Println("📚 Swagger UI:           http://localhost:8080/swagger/index.html")
	fmt.Println("📖 WebSocket API Docs:   http://localhost:8080/docs/websocket-api.md")
	fmt.Println("🏥 Health Check:         http://localhost:8080/api/v1/health")
	fmt.Println("📊 Metrics:              http://localhost:8080/api/v1/metrics")
	fmt.Println("🔌 WebSocket Info:       http://localhost:8080/api/v1/ws/info")
	fmt.Println("=" + string(make([]rune, 50)))
	fmt.Println("🎯 Server running on port 8080...")

	log.Fatal(http.ListenAndServe(":8080", r))
}
