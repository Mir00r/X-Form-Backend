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

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

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

// Mock WebSocket Hub for demo
type MockHub struct {
	TotalConnections  int64
	ActiveConnections int64
	TotalRooms        int64
	ActiveRooms       int64
	MessagesPerSecond int64
	ErrorsPerSecond   int64
}

func (h *MockHub) GetMetrics() *MockHub {
	return h
}

// Response types for Swagger documentation
type HealthResponse struct {
	Status       string                 `json:"status" example:"healthy"`
	Service      string                 `json:"service" example:"collaboration-service"`
	Version      string                 `json:"version" example:"1.0.0"`
	Timestamp    time.Time              `json:"timestamp" example:"2024-01-15T10:30:00Z"`
	Uptime       string                 `json:"uptime" example:"1h30m45s"`
	Environment  string                 `json:"environment" example:"development"`
	Dependencies map[string]interface{} `json:"dependencies"`
}

type MetricsResponse struct {
	TotalConnections  int64   `json:"totalConnections" example:"150"`
	ActiveConnections int64   `json:"activeConnections" example:"25"`
	TotalRooms        int64   `json:"totalRooms" example:"10"`
	ActiveRooms       int64   `json:"activeRooms" example:"8"`
	MessagesPerSecond int64   `json:"messagesPerSecond" example:"45"`
	ErrorsPerSecond   int64   `json:"errorsPerSecond" example:"0"`
	AverageLatency    float64 `json:"averageLatency" example:"12.5"`
	MemoryUsage       string  `json:"memoryUsage" example:"256MB"`
	CPUUsage          float64 `json:"cpuUsage" example:"15.2"`
}

type RoomInfo struct {
	FormID       string    `json:"formId" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserCount    int       `json:"userCount" example:"3"`
	MaxUsers     int       `json:"maxUsers" example:"10"`
	CreatedAt    time.Time `json:"createdAt" example:"2024-01-15T10:00:00Z"`
	UpdatedAt    time.Time `json:"updatedAt" example:"2024-01-15T10:30:00Z"`
	IsActive     bool      `json:"isActive" example:"true"`
	LastActivity time.Time `json:"lastActivity" example:"2024-01-15T10:29:45Z"`
}

type RoomsResponse struct {
	Rooms       []RoomInfo `json:"rooms"`
	TotalCount  int        `json:"totalCount" example:"10"`
	ActiveCount int        `json:"activeCount" example:"8"`
	Timestamp   time.Time  `json:"timestamp" example:"2024-01-15T10:30:00Z"`
}

type UserInfo struct {
	ID             string          `json:"id" example:"user123"`
	Email          string          `json:"email" example:"john.doe@example.com"`
	Name           string          `json:"name" example:"John Doe"`
	Avatar         string          `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
	Role           string          `json:"role" example:"editor"`
	ConnectedAt    time.Time       `json:"connectedAt" example:"2024-01-15T10:15:00Z"`
	LastSeen       time.Time       `json:"lastSeen" example:"2024-01-15T10:29:30Z"`
	IsOnline       bool            `json:"isOnline" example:"true"`
	CursorPosition *CursorPosition `json:"cursorPosition,omitempty"`
}

type CursorPosition struct {
	X          int    `json:"x" example:"100"`
	Y          int    `json:"y" example:"200"`
	QuestionID string `json:"questionId,omitempty" example:"q1"`
	Section    string `json:"section,omitempty" example:"title"`
}

type RoomDetailsResponse struct {
	FormID       string            `json:"formId" example:"550e8400-e29b-41d4-a716-446655440000"`
	Users        []UserInfo        `json:"users"`
	UserCount    int               `json:"userCount" example:"3"`
	MaxUsers     int               `json:"maxUsers" example:"10"`
	CreatedAt    time.Time         `json:"createdAt" example:"2024-01-15T10:00:00Z"`
	UpdatedAt    time.Time         `json:"updatedAt" example:"2024-01-15T10:30:00Z"`
	IsActive     bool              `json:"isActive" example:"true"`
	LastActivity time.Time         `json:"lastActivity" example:"2024-01-15T10:29:45Z"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type UserSession struct {
	UserID       string    `json:"userId" example:"user123"`
	Connections  []string  `json:"connections"`
	Rooms        []string  `json:"rooms"`
	ConnectedAt  time.Time `json:"connectedAt" example:"2024-01-15T10:15:00Z"`
	LastActivity time.Time `json:"lastActivity" example:"2024-01-15T10:29:30Z"`
	IsActive     bool      `json:"isActive" example:"true"`
}

type UserSessionsResponse struct {
	Sessions    []UserSession `json:"sessions"`
	TotalCount  int           `json:"totalCount" example:"25"`
	ActiveCount int           `json:"activeCount" example:"20"`
	Timestamp   time.Time     `json:"timestamp" example:"2024-01-15T10:30:00Z"`
}

type ErrorResponse struct {
	Error     string    `json:"error" example:"Resource not found"`
	Code      string    `json:"code" example:"NOT_FOUND"`
	Message   string    `json:"message" example:"The requested resource was not found"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-15T10:30:00Z"`
	Path      string    `json:"path" example:"/api/v1/rooms/invalid-id"`
}

type WebSocketInfo struct {
	Endpoint     string            `json:"endpoint" example:"ws://localhost:8080/api/v1/ws"`
	AuthMethod   string            `json:"authMethod" example:"JWT Bearer Token"`
	Events       map[string]string `json:"events"`
	ExampleUsage string            `json:"exampleUsage"`
}

// Global mock hub
var mockHub = &MockHub{
	TotalConnections:  150,
	ActiveConnections: 25,
	TotalRooms:        10,
	ActiveRooms:       8,
	MessagesPerSecond: 45,
	ErrorsPerSecond:   0,
}

var startTime = time.Now()

func main() {
	// Setup HTTP router with Swagger
	router := setupSwaggerRoutes()

	// Setup HTTP server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		fmt.Println("🚀 X-Form Collaboration Service Demo Started Successfully!")
		fmt.Println("========================================================")
		fmt.Println("🌐 Server running on: http://localhost:8080")
		fmt.Println("📖 API Documentation: http://localhost:8080/swagger/")
		fmt.Println("🔌 WebSocket Endpoint: ws://localhost:8080/api/v1/ws")
		fmt.Println("🏥 Health Check: http://localhost:8080/api/v1/health")
		fmt.Println("📊 Metrics: http://localhost:8080/api/v1/metrics")
		fmt.Println("========================================================")
		fmt.Println("🔗 Real-time Collaboration Features:")
		fmt.Println("1. WebSocket-based real-time communication")
		fmt.Println("2. Form collaboration with multiple users")
		fmt.Println("3. Cursor tracking and live updates")
		fmt.Println("4. Question management with real-time sync")
		fmt.Println("5. Session management and monitoring")
		fmt.Println("========================================================")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server stopped")
}

func setupSwaggerRoutes() *mux.Router {
	router := mux.NewRouter()

	// Swagger documentation endpoint
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// API documentation redirect
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})

	// API v1 routes
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	// WebSocket endpoint (documentation only - mock endpoint)
	apiV1.HandleFunc("/ws", getWebSocketInfo).Methods("GET")

	// Health check endpoint
	apiV1.HandleFunc("/health", getHealthCheck).Methods("GET")

	// Metrics endpoint
	apiV1.HandleFunc("/metrics", getMetrics).Methods("GET")

	// Room management endpoints
	apiV1.HandleFunc("/rooms", getRooms).Methods("GET")
	apiV1.HandleFunc("/rooms/{formId}", getRoomDetails).Methods("GET")
	apiV1.HandleFunc("/rooms/{formId}/users", getRoomUsers).Methods("GET")

	// User session endpoints
	apiV1.HandleFunc("/sessions", getUserSessions).Methods("GET")
	apiV1.HandleFunc("/sessions/{userId}", getUserSessionDetails).Methods("GET")

	// CORS middleware
	router.Use(corsMiddleware)

	// Logging middleware
	router.Use(loggingMiddleware)

	return router
}

// @Summary Get WebSocket connection information
// @Description Returns information about the WebSocket endpoint for real-time collaboration
// @Tags WebSocket
// @Accept json
// @Produce json
// @Success 200 {object} WebSocketInfo "WebSocket information retrieved successfully"
// @Router /api/v1/ws [get]
func getWebSocketInfo(w http.ResponseWriter, r *http.Request) {
	info := WebSocketInfo{
		Endpoint:   "ws://localhost:8080/api/v1/ws",
		AuthMethod: "JWT Bearer Token",
		Events: map[string]string{
			"join:form":       "Join a form collaboration session",
			"leave:form":      "Leave a form collaboration session",
			"cursor:update":   "Update cursor position",
			"question:create": "Create new question",
			"question:update": "Update existing question",
			"question:delete": "Delete question",
			"ping":            "Keep-alive ping",
		},
		ExampleUsage: `
// JavaScript WebSocket connection example
const ws = new WebSocket('ws://localhost:8080/api/v1/ws?token=YOUR_JWT_TOKEN');

ws.onopen = () => {
  // Join a form collaboration session
  ws.send(JSON.stringify({
    type: 'join:form',
    payload: { formId: 'your-form-id' },
    timestamp: new Date().toISOString(),
    messageId: 'msg123'
  }));
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};
		`,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(info)
}

// @Summary Get service health status
// @Description Returns the health status of the collaboration service and its dependencies
// @Tags Health & Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse "Service is healthy"
// @Failure 503 {object} ErrorResponse "Service is unhealthy"
// @Router /api/v1/health [get]
func getHealthCheck(w http.ResponseWriter, r *http.Request) {
	health := HealthResponse{
		Status:      "healthy",
		Service:     "collaboration-service",
		Version:     "1.0.0",
		Timestamp:   time.Now(),
		Uptime:      time.Since(startTime).String(),
		Environment: "development",
		Dependencies: map[string]interface{}{
			"redis": map[string]interface{}{
				"status":       "healthy",
				"responseTime": "2ms",
				"lastChecked":  time.Now(),
			},
			"websocket": map[string]interface{}{
				"status":            "healthy",
				"activeConnections": mockHub.ActiveConnections,
				"lastChecked":       time.Now(),
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}

// @Summary Get service metrics
// @Description Returns real-time metrics and statistics for the collaboration service
// @Tags Health & Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} MetricsResponse "Service metrics retrieved successfully"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/metrics [get]
func getMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := mockHub.GetMetrics()

	response := MetricsResponse{
		TotalConnections:  metrics.TotalConnections,
		ActiveConnections: metrics.ActiveConnections,
		TotalRooms:        metrics.TotalRooms,
		ActiveRooms:       metrics.ActiveRooms,
		MessagesPerSecond: metrics.MessagesPerSecond,
		ErrorsPerSecond:   metrics.ErrorsPerSecond,
		AverageLatency:    12.5,
		MemoryUsage:       "256MB",
		CPUUsage:          15.2,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get all collaboration rooms
// @Description Returns a list of all active collaboration rooms (forms) with basic information
// @Tags Room Management
// @Accept json
// @Produce json
// @Param active query bool false "Filter by active rooms only" default(false)
// @Param limit query int false "Maximum number of rooms to return" default(50)
// @Param offset query int false "Number of rooms to skip" default(0)
// @Success 200 {object} RoomsResponse "Rooms retrieved successfully"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/rooms [get]
func getRooms(w http.ResponseWriter, r *http.Request) {
	rooms := []RoomInfo{
		{
			FormID:       "550e8400-e29b-41d4-a716-446655440000",
			UserCount:    3,
			MaxUsers:     10,
			CreatedAt:    time.Now().Add(-2 * time.Hour),
			UpdatedAt:    time.Now().Add(-10 * time.Minute),
			IsActive:     true,
			LastActivity: time.Now().Add(-2 * time.Minute),
		},
		{
			FormID:       "550e8400-e29b-41d4-a716-446655440001",
			UserCount:    1,
			MaxUsers:     10,
			CreatedAt:    time.Now().Add(-1 * time.Hour),
			UpdatedAt:    time.Now().Add(-5 * time.Minute),
			IsActive:     true,
			LastActivity: time.Now().Add(-1 * time.Minute),
		},
	}

	response := RoomsResponse{
		Rooms:       rooms,
		TotalCount:  len(rooms),
		ActiveCount: len(rooms),
		Timestamp:   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get room details
// @Description Returns detailed information about a specific collaboration room including users and their cursor positions
// @Tags Room Management
// @Accept json
// @Produce json
// @Param formId path string true "Form ID" example("550e8400-e29b-41d4-a716-446655440000")
// @Success 200 {object} RoomDetailsResponse "Room details retrieved successfully"
// @Failure 404 {object} ErrorResponse "Room not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/rooms/{formId} [get]
func getRoomDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	formID := vars["formId"]

	if formID == "550e8400-e29b-41d4-a716-446655440000" {
		users := []UserInfo{
			{
				ID:          "user123",
				Email:       "john.doe@example.com",
				Name:        "John Doe",
				Role:        "editor",
				ConnectedAt: time.Now().Add(-30 * time.Minute),
				LastSeen:    time.Now().Add(-1 * time.Minute),
				IsOnline:    true,
				CursorPosition: &CursorPosition{
					X:          100,
					Y:          200,
					QuestionID: "q1",
					Section:    "title",
				},
			},
		}

		response := RoomDetailsResponse{
			FormID:       formID,
			Users:        users,
			UserCount:    len(users),
			MaxUsers:     10,
			CreatedAt:    time.Now().Add(-2 * time.Hour),
			UpdatedAt:    time.Now().Add(-10 * time.Minute),
			IsActive:     true,
			LastActivity: time.Now().Add(-1 * time.Minute),
			Metadata:     map[string]string{"type": "survey", "title": "Customer Feedback"},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	} else {
		errorResponse := ErrorResponse{
			Error:     "Room not found",
			Code:      "NOT_FOUND",
			Message:   "The requested room was not found",
			Timestamp: time.Now(),
			Path:      r.URL.Path,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse)
	}
}

// @Summary Get room users
// @Description Returns a list of users currently in a specific collaboration room
// @Tags Room Management
// @Accept json
// @Produce json
// @Param formId path string true "Form ID" example("550e8400-e29b-41d4-a716-446655440000")
// @Param online query bool false "Filter by online users only" default(false)
// @Success 200 {array} UserInfo "Room users retrieved successfully"
// @Failure 404 {object} ErrorResponse "Room not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/rooms/{formId}/users [get]
func getRoomUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	formID := vars["formId"]

	if formID == "550e8400-e29b-41d4-a716-446655440000" {
		users := []UserInfo{
			{
				ID:          "user123",
				Email:       "john.doe@example.com",
				Name:        "John Doe",
				Role:        "editor",
				ConnectedAt: time.Now().Add(-30 * time.Minute),
				LastSeen:    time.Now().Add(-1 * time.Minute),
				IsOnline:    true,
			},
			{
				ID:          "user456",
				Email:       "jane.smith@example.com",
				Name:        "Jane Smith",
				Role:        "viewer",
				ConnectedAt: time.Now().Add(-15 * time.Minute),
				LastSeen:    time.Now().Add(-30 * time.Second),
				IsOnline:    true,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(users)
	} else {
		errorResponse := ErrorResponse{
			Error:     "Room not found",
			Code:      "NOT_FOUND",
			Message:   "The requested room was not found",
			Timestamp: time.Now(),
			Path:      r.URL.Path,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse)
	}
}

// @Summary Get all user sessions
// @Description Returns a list of all active user sessions across the service
// @Tags Session Management
// @Accept json
// @Produce json
// @Param active query bool false "Filter by active sessions only" default(false)
// @Param limit query int false "Maximum number of sessions to return" default(50)
// @Param offset query int false "Number of sessions to skip" default(0)
// @Success 200 {object} UserSessionsResponse "User sessions retrieved successfully"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sessions [get]
func getUserSessions(w http.ResponseWriter, r *http.Request) {
	sessions := []UserSession{
		{
			UserID:       "user123",
			Connections:  []string{"conn1", "conn2"},
			Rooms:        []string{"550e8400-e29b-41d4-a716-446655440000"},
			ConnectedAt:  time.Now().Add(-30 * time.Minute),
			LastActivity: time.Now().Add(-1 * time.Minute),
			IsActive:     true,
		},
		{
			UserID:       "user456",
			Connections:  []string{"conn3"},
			Rooms:        []string{"550e8400-e29b-41d4-a716-446655440001"},
			ConnectedAt:  time.Now().Add(-15 * time.Minute),
			LastActivity: time.Now().Add(-30 * time.Second),
			IsActive:     true,
		},
	}

	response := UserSessionsResponse{
		Sessions:    sessions,
		TotalCount:  len(sessions),
		ActiveCount: len(sessions),
		Timestamp:   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get user session details
// @Description Returns detailed information about a specific user's session including connections and room participation
// @Tags Session Management
// @Accept json
// @Produce json
// @Param userId path string true "User ID" example("user123")
// @Success 200 {object} UserSession "User session details retrieved successfully"
// @Failure 404 {object} ErrorResponse "User session not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/sessions/{userId} [get]
func getUserSessionDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	if userID == "user123" {
		session := UserSession{
			UserID:       userID,
			Connections:  []string{"conn1", "conn2"},
			Rooms:        []string{"550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440001"},
			ConnectedAt:  time.Now().Add(-30 * time.Minute),
			LastActivity: time.Now().Add(-1 * time.Minute),
			IsActive:     true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(session)
	} else {
		errorResponse := ErrorResponse{
			Error:     "User session not found",
			Code:      "NOT_FOUND",
			Message:   "The requested user session was not found",
			Timestamp: time.Now(),
			Path:      r.URL.Path,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse)
	}
}

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

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)
		duration := time.Since(start)
		fmt.Printf("HTTP %s %s %d %v\n", r.Method, r.URL.Path, rw.statusCode, duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
