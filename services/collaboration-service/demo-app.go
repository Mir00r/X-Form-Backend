// Package main provides the X-Form Collaboration Service API Demo
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
// @description See the [WebSocket API documentation](/docs/websocket-api.md) for detailed event specifications.
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
	"strconv"
	"time"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Response structures for API documentation

// APIErrorResponse represents an error response
type APIErrorResponse struct {
	Error     string `json:"error" example:"Resource not found"`
	Code      int    `json:"code" example:"404"`
	Message   string `json:"message" example:"The requested resource was not found"`
	Timestamp string `json:"timestamp" example:"2023-01-01T00:00:00Z"`
} // @name APIErrorResponse

// ServiceHealthResponse represents a health check response
type ServiceHealthResponse struct {
	Status      string            `json:"status" example:"healthy"`
	Version     string            `json:"version" example:"1.0.0"`
	Environment string            `json:"environment" example:"production"`
	Timestamp   time.Time         `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	Services    map[string]string `json:"services"`
	Uptime      string            `json:"uptime" example:"72h30m45s"`
} // @name ServiceHealthResponse

// SystemMetricsResponse represents system metrics
type SystemMetricsResponse struct {
	ActiveConnections int64             `json:"active_connections" example:"150"`
	ActiveRooms       int64             `json:"active_rooms" example:"25"`
	TotalUsers        int64             `json:"total_users" example:"1250"`
	MessageRate       float64           `json:"message_rate" example:"45.6"`
	MemoryUsage       string            `json:"memory_usage" example:"256MB"`
	CPUUsage          float64           `json:"cpu_usage" example:"15.2"`
	RedisConnections  int               `json:"redis_connections" example:"10"`
	RequestCount      map[string]int64  `json:"request_count"`
	ResponseTimes     map[string]string `json:"response_times"`
} // @name SystemMetricsResponse

// CursorPositionData represents cursor position data
type CursorPositionData struct {
	X         float64 `json:"x" example:"120.5"`
	Y         float64 `json:"y" example:"75.2"`
	ElementID string  `json:"element_id" example:"question_1"`
	Timestamp int64   `json:"timestamp" example:"1672531200"`
	IsActive  bool    `json:"is_active" example:"true"`
} // @name CursorPositionData

// CollaboratorInfo represents information about a user in a room
type CollaboratorInfo struct {
	UserID       string             `json:"user_id" example:"user_456"`
	Username     string             `json:"username" example:"john.doe"`
	Email        string             `json:"email" example:"john@example.com"`
	Role         string             `json:"role" example:"editor"`
	JoinedAt     time.Time          `json:"joined_at" example:"2023-01-01T10:15:00Z"`
	LastActivity time.Time          `json:"last_activity" example:"2023-01-01T10:30:00Z"`
	IsOnline     bool               `json:"is_online" example:"true"`
	Cursor       CursorPositionData `json:"cursor"`
} // @name CollaboratorInfo

// CollaborationRoomInfo represents information about a collaboration room
type CollaborationRoomInfo struct {
	FormID       string    `json:"form_id" example:"form_123"`
	ActiveUsers  int       `json:"active_users" example:"3"`
	CreatedAt    time.Time `json:"created_at" example:"2023-01-01T10:00:00Z"`
	LastActivity time.Time `json:"last_activity" example:"2023-01-01T10:30:00Z"`
	IsActive     bool      `json:"is_active" example:"true"`
	MaxUsers     int       `json:"max_users" example:"10"`
	CurrentPhase string    `json:"current_phase" example:"editing"`
} // @name CollaborationRoomInfo

// RoomsListResponse represents the list of active rooms
type RoomsListResponse struct {
	Rooms []CollaborationRoomInfo `json:"rooms"`
	Total int                     `json:"total" example:"25"`
	Page  int                     `json:"page" example:"1"`
	Limit int                     `json:"limit" example:"20"`
} // @name RoomsListResponse

// RoomDetailsInfo represents detailed room information
type RoomDetailsInfo struct {
	Room  CollaborationRoomInfo `json:"room"`
	Users []CollaboratorInfo    `json:"users"`
} // @name RoomDetailsInfo

// ActiveUserSession represents a user session
type ActiveUserSession struct {
	SessionID     string    `json:"session_id" example:"session_789"`
	UserID        string    `json:"user_id" example:"user_456"`
	FormID        string    `json:"form_id" example:"form_123"`
	ConnectedAt   time.Time `json:"connected_at" example:"2023-01-01T10:15:00Z"`
	LastHeartbeat time.Time `json:"last_heartbeat" example:"2023-01-01T10:30:00Z"`
	IPAddress     string    `json:"ip_address" example:"192.168.1.100"`
	UserAgent     string    `json:"user_agent" example:"Mozilla/5.0..."`
	IsActive      bool      `json:"is_active" example:"true"`
} // @name ActiveUserSession

// SessionsListResponse represents user sessions response
type SessionsListResponse struct {
	Sessions []ActiveUserSession `json:"sessions"`
	Total    int                 `json:"total" example:"5"`
} // @name SessionsListResponse

// WebSocketConnectionInfo represents WebSocket connection information
type WebSocketConnectionInfo struct {
	Endpoint     string   `json:"endpoint" example:"/api/v1/ws"`
	Protocol     string   `json:"protocol" example:"ws"`
	AuthRequired bool     `json:"auth_required" example:"true"`
	AuthType     string   `json:"auth_type" example:"Bearer JWT"`
	Events       []string `json:"events"`
	Description  string   `json:"description" example:"WebSocket endpoint for real-time collaboration"`
} // @name WebSocketConnectionInfo

// WebSocketEventInfo represents a WebSocket event structure
type WebSocketEventInfo struct {
	Type        string `json:"type" example:"join:form"`
	Description string `json:"description" example:"User joins a form collaboration session"`
	Example     string `json:"example,omitempty"`
} // @name WebSocketEventInfo

// WebSocketEventsListResponse represents available WebSocket events
type WebSocketEventsListResponse struct {
	Events []WebSocketEventInfo `json:"events"`
	Total  int                  `json:"total" example:"11"`
} // @name WebSocketEventsListResponse

// HTTP Handlers

// @Summary Get service health status
// @Description Returns the health status of the collaboration service and its dependencies
// @Tags Health & Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} ServiceHealthResponse "Service is healthy"
// @Failure 503 {object} APIErrorResponse "Service is unhealthy"
// @Router /api/v1/health [get]
func getHealthStatus(w http.ResponseWriter, r *http.Request) {
	response := ServiceHealthResponse{
		Status:      "healthy",
		Version:     "1.0.0",
		Environment: "development",
		Timestamp:   time.Now(),
		Services: map[string]string{
			"redis":        "connected",
			"auth-service": "available",
			"database":     "connected",
			"websocket":    "active",
		},
		Uptime: "72h30m45s",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get system metrics
// @Description Returns current system metrics and performance statistics
// @Tags Health & Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} SystemMetricsResponse "System metrics retrieved successfully"
// @Failure 500 {object} APIErrorResponse "Failed to retrieve metrics"
// @Router /api/v1/metrics [get]
func getSystemMetrics(w http.ResponseWriter, r *http.Request) {
	response := SystemMetricsResponse{
		ActiveConnections: 150,
		ActiveRooms:       25,
		TotalUsers:        1250,
		MessageRate:       45.6,
		MemoryUsage:       "256MB",
		CPUUsage:          15.2,
		RedisConnections:  10,
		RequestCount: map[string]int64{
			"GET":       1500,
			"POST":      800,
			"WebSocket": 2500,
		},
		ResponseTimes: map[string]string{
			"avg": "25ms",
			"p95": "100ms",
			"p99": "250ms",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get active collaboration rooms
// @Description Returns a list of all active collaboration rooms with pagination
// @Tags Collaboration
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} RoomsListResponse "List of active rooms"
// @Failure 500 {object} APIErrorResponse "Failed to retrieve rooms"
// @Router /api/v1/rooms [get]
func getActiveRooms(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page := 1
	limit := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Mock data
	rooms := []CollaborationRoomInfo{
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
		{
			FormID:       "form_789",
			ActiveUsers:  5,
			CreatedAt:    time.Now().Add(-3 * time.Hour),
			LastActivity: time.Now().Add(-1 * time.Minute),
			IsActive:     true,
			MaxUsers:     10,
			CurrentPhase: "collaboration",
		},
	}

	response := RoomsListResponse{
		Rooms: rooms,
		Total: len(rooms),
		Page:  page,
		Limit: limit,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get room details
// @Description Returns detailed information about a specific collaboration room
// @Tags Collaboration
// @Accept json
// @Produce json
// @Param formId path string true "Form ID" example("form_123")
// @Success 200 {object} RoomDetailsInfo "Room details retrieved successfully"
// @Failure 404 {object} APIErrorResponse "Room not found"
// @Failure 500 {object} APIErrorResponse "Failed to retrieve room details"
// @Router /api/v1/rooms/{formId} [get]
func getRoomDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	formID := vars["formId"]

	if formID == "" {
		errorResponse := APIErrorResponse{
			Error:     "Invalid form ID",
			Code:      400,
			Message:   "Form ID cannot be empty",
			Timestamp: time.Now().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	room := CollaborationRoomInfo{
		FormID:       formID,
		ActiveUsers:  3,
		CreatedAt:    time.Now().Add(-2 * time.Hour),
		LastActivity: time.Now().Add(-5 * time.Minute),
		IsActive:     true,
		MaxUsers:     10,
		CurrentPhase: "editing",
	}

	users := []CollaboratorInfo{
		{
			UserID:       "user_456",
			Username:     "john.doe",
			Email:        "john@example.com",
			Role:         "editor",
			JoinedAt:     time.Now().Add(-30 * time.Minute),
			LastActivity: time.Now().Add(-2 * time.Minute),
			IsOnline:     true,
			Cursor: CursorPositionData{
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
			Cursor: CursorPositionData{
				X:         200.0,
				Y:         150.0,
				ElementID: "question_2",
				Timestamp: time.Now().Unix(),
				IsActive:  false,
			},
		},
	}

	response := RoomDetailsInfo{
		Room:  room,
		Users: users,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get users in a room
// @Description Returns a list of all users currently in a collaboration room
// @Tags Collaboration
// @Accept json
// @Produce json
// @Param formId path string true "Form ID" example("form_123")
// @Success 200 {array} CollaboratorInfo "List of users in the room"
// @Failure 404 {object} APIErrorResponse "Room not found"
// @Failure 500 {object} APIErrorResponse "Failed to retrieve users"
// @Router /api/v1/rooms/{formId}/users [get]
func getRoomUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	formID := vars["formId"]

	if formID == "" {
		errorResponse := APIErrorResponse{
			Error:     "Invalid form ID",
			Code:      400,
			Message:   "Form ID cannot be empty",
			Timestamp: time.Now().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	users := []CollaboratorInfo{
		{
			UserID:       "user_456",
			Username:     "john.doe",
			Email:        "john@example.com",
			Role:         "editor",
			JoinedAt:     time.Now().Add(-30 * time.Minute),
			LastActivity: time.Now().Add(-2 * time.Minute),
			IsOnline:     true,
			Cursor: CursorPositionData{
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
			Cursor: CursorPositionData{
				X:         200.0,
				Y:         150.0,
				ElementID: "question_2",
				Timestamp: time.Now().Unix(),
				IsActive:  false,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

// @Summary Get all user sessions
// @Description Returns a list of all active user sessions across all rooms
// @Tags Session Management
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} SessionsListResponse "List of user sessions"
// @Failure 500 {object} APIErrorResponse "Failed to retrieve sessions"
// @Router /api/v1/sessions [get]
func getAllSessions(w http.ResponseWriter, r *http.Request) {
	page := 1
	limit := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	sessions := []ActiveUserSession{
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
		{
			SessionID:     "session_890",
			UserID:        "user_789",
			FormID:        "form_123",
			ConnectedAt:   time.Now().Add(-15 * time.Minute),
			LastHeartbeat: time.Now().Add(-5 * time.Second),
			IPAddress:     "192.168.1.101",
			UserAgent:     "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			IsActive:      true,
		},
	}

	response := SessionsListResponse{
		Sessions: sessions,
		Total:    len(sessions),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get user sessions
// @Description Returns sessions for a specific user
// @Tags Session Management
// @Accept json
// @Produce json
// @Param userId path string true "User ID" example("user_456")
// @Success 200 {object} SessionsListResponse "User sessions retrieved successfully"
// @Failure 404 {object} APIErrorResponse "User not found"
// @Failure 500 {object} APIErrorResponse "Failed to retrieve user sessions"
// @Router /api/v1/sessions/{userId} [get]
func getUserSessions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	if userID == "" {
		errorResponse := APIErrorResponse{
			Error:     "Invalid user ID",
			Code:      400,
			Message:   "User ID cannot be empty",
			Timestamp: time.Now().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	sessions := []ActiveUserSession{
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

	response := SessionsListResponse{
		Sessions: sessions,
		Total:    len(sessions),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get WebSocket information
// @Description Returns information about WebSocket endpoint and supported events
// @Tags WebSocket
// @Accept json
// @Produce json
// @Success 200 {object} WebSocketConnectionInfo "WebSocket information"
// @Router /api/v1/ws/info [get]
func getWebSocketInfo(w http.ResponseWriter, r *http.Request) {
	response := WebSocketConnectionInfo{
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
		Description: "WebSocket endpoint for real-time collaboration on form editing. Requires JWT authentication.",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get WebSocket events documentation
// @Description Returns detailed documentation of all available WebSocket events
// @Tags WebSocket
// @Accept json
// @Produce json
// @Success 200 {object} WebSocketEventsListResponse "WebSocket events documentation"
// @Router /api/v1/ws/events [get]
func getWebSocketEvents(w http.ResponseWriter, r *http.Request) {
	events := []WebSocketEventInfo{
		{
			Type:        "join:form",
			Description: "User joins a form collaboration session",
			Example:     `{"type":"join:form","payload":{"form_id":"form_123","user_id":"user_456"}}`,
		},
		{
			Type:        "leave:form",
			Description: "User leaves a form collaboration session",
			Example:     `{"type":"leave:form","payload":{"form_id":"form_123","user_id":"user_456"}}`,
		},
		{
			Type:        "cursor:move",
			Description: "User moves cursor position",
			Example:     `{"type":"cursor:move","payload":{"x":120.5,"y":75.2,"element_id":"question_1"}}`,
		},
		{
			Type:        "cursor:hide",
			Description: "User hides cursor",
			Example:     `{"type":"cursor:hide","payload":{"user_id":"user_456"}}`,
		},
		{
			Type:        "question:update",
			Description: "Question content is updated",
			Example:     `{"type":"question:update","payload":{"question_id":"q1","content":"Updated question text"}}`,
		},
		{
			Type:        "question:focus",
			Description: "User focuses on a question",
			Example:     `{"type":"question:focus","payload":{"question_id":"q1","user_id":"user_456"}}`,
		},
		{
			Type:        "question:blur",
			Description: "User unfocuses from a question",
			Example:     `{"type":"question:blur","payload":{"question_id":"q1","user_id":"user_456"}}`,
		},
		{
			Type:        "form:save",
			Description: "Form is saved",
			Example:     `{"type":"form:save","payload":{"form_id":"form_123","user_id":"user_456"}}`,
		},
		{
			Type:        "user:typing",
			Description: "User starts typing",
			Example:     `{"type":"user:typing","payload":{"user_id":"user_456","element_id":"question_1"}}`,
		},
		{
			Type:        "user:stopped_typing",
			Description: "User stops typing",
			Example:     `{"type":"user:stopped_typing","payload":{"user_id":"user_456","element_id":"question_1"}}`,
		},
		{
			Type:        "heartbeat",
			Description: "Connection heartbeat to maintain session",
			Example:     `{"type":"heartbeat","payload":{"timestamp":1672531200}}`,
		},
	}

	response := WebSocketEventsListResponse{
		Events: events,
		Total:  len(events),
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
	api.HandleFunc("/health", getHealthStatus).Methods("GET")
	api.HandleFunc("/metrics", getSystemMetrics).Methods("GET")

	// Collaboration room endpoints
	api.HandleFunc("/rooms", getActiveRooms).Methods("GET")
	api.HandleFunc("/rooms/{formId}", getRoomDetails).Methods("GET")
	api.HandleFunc("/rooms/{formId}/users", getRoomUsers).Methods("GET")

	// Session management endpoints
	api.HandleFunc("/sessions", getAllSessions).Methods("GET")
	api.HandleFunc("/sessions/{userId}", getUserSessions).Methods("GET")

	// WebSocket information endpoints
	api.HandleFunc("/ws/info", getWebSocketInfo).Methods("GET")
	api.HandleFunc("/ws/events", getWebSocketEvents).Methods("GET")

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
	fmt.Println("🏠 Active Rooms:         http://localhost:8080/api/v1/rooms")
	fmt.Println("👥 User Sessions:        http://localhost:8080/api/v1/sessions")
	fmt.Println("🔌 WebSocket Info:       http://localhost:8080/api/v1/ws/info")
	fmt.Println("📋 WebSocket Events:     http://localhost:8080/api/v1/ws/events")
	fmt.Println("=" + string(make([]rune, 50)))
	fmt.Println("🎯 Server running on port 8080...")

	log.Fatal(http.ListenAndServe(":8080", r))
}
