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
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/kamkaiz/x-form-backend/collaboration-service/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error" example:"Internal server error"`
	Code    int    `json:"code" example:"500"`
	Message string `json:"message" example:"Something went wrong"`
} // @name ErrorResponse

// HealthResponse represents a health check response
type HealthResponse struct {
	Status      string            `json:"status" example:"healthy"`
	Version     string            `json:"version" example:"1.0.0"`
	Environment string            `json:"environment" example:"production"`
	Timestamp   time.Time         `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	Services    map[string]string `json:"services" example:"redis:connected,auth-service:available"`
	Uptime      string            `json:"uptime" example:"72h30m45s"`
} // @name HealthResponse

// MetricsResponse represents system metrics
type MetricsResponse struct {
	ActiveConnections int64             `json:"active_connections" example:"150"`
	ActiveRooms       int64             `json:"active_rooms" example:"25"`
	TotalUsers        int64             `json:"total_users" example:"1250"`
	MessageRate       float64           `json:"message_rate" example:"45.6"`
	MemoryUsage       string            `json:"memory_usage" example:"256MB"`
	CPUUsage          string            `json:"cpu_usage" example:"15.2%"`
	RedisConnections  int               `json:"redis_connections" example:"10"`
	RequestCount      map[string]int64  `json:"request_count"`
	ResponseTimes     map[string]string `json:"response_times"`
} // @name MetricsResponse

// RoomInfo represents information about a collaboration room
type RoomInfo struct {
	FormID       string    `json:"form_id" example:"form_123"`
	ActiveUsers  int       `json:"active_users" example:"3"`
	CreatedAt    time.Time `json:"created_at" example:"2023-01-01T10:00:00Z"`
	LastActivity time.Time `json:"last_activity" example:"2023-01-01T10:30:00Z"`
	IsActive     bool      `json:"is_active" example:"true"`
	MaxUsers     int       `json:"max_users" example:"10"`
	CurrentPhase string    `json:"current_phase" example:"editing"`
} // @name RoomInfo

// RoomsResponse represents the list of active rooms
type RoomsResponse struct {
	Rooms []RoomInfo `json:"rooms"`
	Total int        `json:"total" example:"25"`
	Page  int        `json:"page" example:"1"`
	Limit int        `json:"limit" example:"20"`
} // @name RoomsResponse

// UserInfo represents information about a user in a room
type UserInfo struct {
	UserID       string         `json:"user_id" example:"user_456"`
	Username     string         `json:"username" example:"john.doe"`
	Email        string         `json:"email" example:"john@example.com"`
	Role         string         `json:"role" example:"editor"`
	JoinedAt     time.Time      `json:"joined_at" example:"2023-01-01T10:15:00Z"`
	LastActivity time.Time      `json:"last_activity" example:"2023-01-01T10:30:00Z"`
	IsOnline     bool           `json:"is_online" example:"true"`
	Cursor       CursorPosition `json:"cursor"`
} // @name UserInfo

// CursorPosition represents cursor position data
type CursorPosition struct {
	X         float64 `json:"x" example:"120.5"`
	Y         float64 `json:"y" example:"75.2"`
	ElementID string  `json:"element_id" example:"question_1"`
	Timestamp int64   `json:"timestamp" example:"1672531200"`
	IsActive  bool    `json:"is_active" example:"true"`
} // @name CursorPosition

// RoomDetailsResponse represents detailed room information
type RoomDetailsResponse struct {
	Room  RoomInfo   `json:"room"`
	Users []UserInfo `json:"users"`
} // @name RoomDetailsResponse

// UserSession represents a user session
type UserSession struct {
	SessionID     string    `json:"session_id" example:"session_789"`
	UserID        string    `json:"user_id" example:"user_456"`
	FormID        string    `json:"form_id" example:"form_123"`
	ConnectedAt   time.Time `json:"connected_at" example:"2023-01-01T10:15:00Z"`
	LastHeartbeat time.Time `json:"last_heartbeat" example:"2023-01-01T10:30:00Z"`
	IPAddress     string    `json:"ip_address" example:"192.168.1.100"`
	UserAgent     string    `json:"user_agent" example:"Mozilla/5.0..."`
	IsActive      bool      `json:"is_active" example:"true"`
} // @name UserSession

// UserSessionsResponse represents user sessions response
type UserSessionsResponse struct {
	Sessions []UserSession `json:"sessions"`
	Total    int           `json:"total" example:"5"`
} // @name UserSessionsResponse

// WebSocketInfo represents WebSocket connection information
type WebSocketInfo struct {
	Endpoint     string   `json:"endpoint" example:"/api/v1/ws"`
	Protocol     string   `json:"protocol" example:"ws"`
	AuthRequired bool     `json:"auth_required" example:"true"`
	AuthType     string   `json:"auth_type" example:"Bearer JWT"`
	Events       []string `json:"events" example:"join:form,leave:form,cursor:move,question:update"`
	Description  string   `json:"description" example:"WebSocket endpoint for real-time collaboration"`
} // @name WebSocketInfo

// @Summary Get service health status
// @Description Returns the health status of the collaboration service and its dependencies
// @Tags Health & Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse "Service is healthy"
// @Failure 503 {object} ErrorResponse "Service is unhealthy"
// @Router /api/v1/health [get]
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:      "healthy",
		Version:     "1.0.0",
		Environment: "development",
		Timestamp:   time.Now(),
		Services: map[string]string{
			"redis":        "connected",
			"auth-service": "available",
			"database":     "connected",
		},
		Uptime: "72h30m45s",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary Get system metrics
// @Description Returns current system metrics and performance statistics
// @Tags Health & Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} MetricsResponse "System metrics retrieved successfully"
// @Failure 500 {object} ErrorResponse "Failed to retrieve metrics"
// @Router /api/v1/metrics [get]
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	response := MetricsResponse{
		ActiveConnections: 150,
		ActiveRooms:       25,
		TotalUsers:        1250,
		MessageRate:       45.6,
		MemoryUsage:       "256MB",
		CPUUsage:          "15.2%",
		RedisConnections:  10,
		RequestCount: map[string]int64{
			"GET":    1500,
			"POST":   800,
			"PUT":    300,
			"DELETE": 50,
		},
		ResponseTimes: map[string]string{
			"avg": "25ms",
			"p95": "100ms",
			"p99": "250ms",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary Get active collaboration rooms
// @Description Returns a list of all active collaboration rooms with pagination
// @Tags Collaboration
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} RoomsResponse "List of active rooms"
// @Failure 500 {object} ErrorResponse "Failed to retrieve rooms"
// @Router /api/v1/rooms [get]
func roomsHandler(w http.ResponseWriter, r *http.Request) {
	rooms := []RoomInfo{
		{
			FormID:       "form_123",
			ActiveUsers:  3,
			CreatedAt:    time.Now().Add(-2 * time.Hour),
			LastActivity: time.Now().Add(-5 * time.Minute),
			IsActive:     true,
			MaxUsers:     10,
			CurrentPhase: "editing",
		},
		{
			FormID:       "form_456",
			ActiveUsers:  1,
			CreatedAt:    time.Now().Add(-1 * time.Hour),
			LastActivity: time.Now().Add(-2 * time.Minute),
			IsActive:     true,
			MaxUsers:     10,
			CurrentPhase: "review",
		},
	}

	response := RoomsResponse{
		Rooms: rooms,
		Total: len(rooms),
		Page:  1,
		Limit: 20,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary Get room details
// @Description Returns detailed information about a specific collaboration room
// @Tags Collaboration
// @Accept json
// @Produce json
// @Param formId path string true "Form ID"
// @Success 200 {object} RoomDetailsResponse "Room details retrieved successfully"
// @Failure 404 {object} ErrorResponse "Room not found"
// @Failure 500 {object} ErrorResponse "Failed to retrieve room details"
// @Router /api/v1/rooms/{formId} [get]
func roomDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	formID := vars["formId"]

	room := RoomInfo{
		FormID:       formID,
		ActiveUsers:  3,
		CreatedAt:    time.Now().Add(-2 * time.Hour),
		LastActivity: time.Now().Add(-5 * time.Minute),
		IsActive:     true,
		MaxUsers:     10,
		CurrentPhase: "editing",
	}

	users := []UserInfo{
		{
			UserID:       "user_456",
			Username:     "john.doe",
			Email:        "john@example.com",
			Role:         "editor",
			JoinedAt:     time.Now().Add(-30 * time.Minute),
			LastActivity: time.Now().Add(-2 * time.Minute),
			IsOnline:     true,
			Cursor: CursorPosition{
				X:         120.5,
				Y:         75.2,
				ElementID: "question_1",
				Timestamp: time.Now().Unix(),
				IsActive:  true,
			},
		},
		{
			UserID:       "user_789",
			Username:     "jane.smith",
			Email:        "jane@example.com",
			Role:         "reviewer",
			JoinedAt:     time.Now().Add(-15 * time.Minute),
			LastActivity: time.Now().Add(-1 * time.Minute),
			IsOnline:     true,
			Cursor: CursorPosition{
				X:         200.0,
				Y:         150.0,
				ElementID: "question_2",
				Timestamp: time.Now().Unix(),
				IsActive:  false,
			},
		},
	}

	response := RoomDetailsResponse{
		Room:  room,
		Users: users,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary Get users in a room
// @Description Returns a list of all users currently in a collaboration room
// @Tags Collaboration
// @Accept json
// @Produce json
// @Param formId path string true "Form ID"
// @Success 200 {array} UserInfo "List of users in the room"
// @Failure 404 {object} ErrorResponse "Room not found"
// @Failure 500 {object} ErrorResponse "Failed to retrieve users"
// @Router /api/v1/rooms/{formId}/users [get]
func roomUsersHandler(w http.ResponseWriter, r *http.Request) {
	users := []UserInfo{
		{
			UserID:       "user_456",
			Username:     "john.doe",
			Email:        "john@example.com",
			Role:         "editor",
			JoinedAt:     time.Now().Add(-30 * time.Minute),
			LastActivity: time.Now().Add(-2 * time.Minute),
			IsOnline:     true,
			Cursor: CursorPosition{
				X:         120.5,
				Y:         75.2,
				ElementID: "question_1",
				Timestamp: time.Now().Unix(),
				IsActive:  true,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// @Summary Get all user sessions
// @Description Returns a list of all active user sessions across all rooms
// @Tags Session Management
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} UserSessionsResponse "List of user sessions"
// @Failure 500 {object} ErrorResponse "Failed to retrieve sessions"
// @Router /api/v1/sessions [get]
func sessionsHandler(w http.ResponseWriter, r *http.Request) {
	sessions := []UserSession{
		{
			SessionID:     "session_789",
			UserID:        "user_456",
			FormID:        "form_123",
			ConnectedAt:   time.Now().Add(-30 * time.Minute),
			LastHeartbeat: time.Now().Add(-10 * time.Second),
			IPAddress:     "192.168.1.100",
			UserAgent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			IsActive:      true,
		},
	}

	response := UserSessionsResponse{
		Sessions: sessions,
		Total:    len(sessions),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary Get user sessions
// @Description Returns sessions for a specific user
// @Tags Session Management
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} UserSessionsResponse "User sessions retrieved successfully"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Failed to retrieve user sessions"
// @Router /api/v1/sessions/{userId} [get]
func userSessionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	sessions := []UserSession{
		{
			SessionID:     "session_789",
			UserID:        userID,
			FormID:        "form_123",
			ConnectedAt:   time.Now().Add(-30 * time.Minute),
			LastHeartbeat: time.Now().Add(-10 * time.Second),
			IPAddress:     "192.168.1.100",
			UserAgent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			IsActive:      true,
		},
	}

	response := UserSessionsResponse{
		Sessions: sessions,
		Total:    len(sessions),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary Get WebSocket information
// @Description Returns information about WebSocket endpoint and supported events
// @Tags WebSocket
// @Accept json
// @Produce json
// @Success 200 {object} WebSocketInfo "WebSocket information"
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
		Description: "WebSocket endpoint for real-time collaboration on form editing",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// Health and monitoring
	api.HandleFunc("/health", healthHandler).Methods("GET")
	api.HandleFunc("/metrics", metricsHandler).Methods("GET")

	// Collaboration rooms
	api.HandleFunc("/rooms", roomsHandler).Methods("GET")
	api.HandleFunc("/rooms/{formId}", roomDetailsHandler).Methods("GET")
	api.HandleFunc("/rooms/{formId}/users", roomUsersHandler).Methods("GET")

	// Session management
	api.HandleFunc("/sessions", sessionsHandler).Methods("GET")
	api.HandleFunc("/sessions/{userId}", userSessionsHandler).Methods("GET")

	// WebSocket information
	api.HandleFunc("/ws/info", websocketInfoHandler).Methods("GET")

	// Swagger UI
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Serve static files for WebSocket API documentation
	r.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs/"))))

	// Root redirect to swagger
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusFound)
	})

	fmt.Println("🚀 X-Form Collaboration Service API")
	fmt.Println("📚 Swagger UI: http://localhost:8080/swagger/index.html")
	fmt.Println("📖 WebSocket API Docs: http://localhost:8080/docs/websocket-api.md")
	fmt.Println("🏥 Health Check: http://localhost:8080/api/v1/health")
	fmt.Println("📊 Metrics: http://localhost:8080/api/v1/metrics")
	fmt.Println("🔌 WebSocket Info: http://localhost:8080/api/v1/ws/info")

	log.Fatal(http.ListenAndServe(":8080", r))
}
