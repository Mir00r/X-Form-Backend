package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/kamkaiz/x-form-backend/collaboration-service/internal/websocket"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status       string                 `json:"status" example:"healthy"`
	Service      string                 `json:"service" example:"collaboration-service"`
	Version      string                 `json:"version" example:"1.0.0"`
	Timestamp    time.Time              `json:"timestamp" example:"2024-01-15T10:30:00Z"`
	Uptime       string                 `json:"uptime" example:"1h30m45s"`
	Environment  string                 `json:"environment" example:"development"`
	Dependencies map[string]interface{} `json:"dependencies"`
}

// MetricsResponse represents the metrics response
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

// RoomInfo represents room information
type RoomInfo struct {
	FormID       string    `json:"formId" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserCount    int       `json:"userCount" example:"3"`
	MaxUsers     int       `json:"maxUsers" example:"10"`
	CreatedAt    time.Time `json:"createdAt" example:"2024-01-15T10:00:00Z"`
	UpdatedAt    time.Time `json:"updatedAt" example:"2024-01-15T10:30:00Z"`
	IsActive     bool      `json:"isActive" example:"true"`
	LastActivity time.Time `json:"lastActivity" example:"2024-01-15T10:29:45Z"`
}

// RoomsResponse represents the rooms list response
type RoomsResponse struct {
	Rooms       []RoomInfo `json:"rooms"`
	TotalCount  int        `json:"totalCount" example:"10"`
	ActiveCount int        `json:"activeCount" example:"8"`
	Timestamp   time.Time  `json:"timestamp" example:"2024-01-15T10:30:00Z"`
}

// UserInfo represents user information in a room
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

// CursorPosition represents cursor position
type CursorPosition struct {
	X          int    `json:"x" example:"100"`
	Y          int    `json:"y" example:"200"`
	QuestionID string `json:"questionId,omitempty" example:"q1"`
	Section    string `json:"section,omitempty" example:"title"`
}

// RoomDetailsResponse represents detailed room information
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

// UserSession represents user session information
type UserSession struct {
	UserID       string    `json:"userId" example:"user123"`
	Connections  []string  `json:"connections"`
	Rooms        []string  `json:"rooms"`
	ConnectedAt  time.Time `json:"connectedAt" example:"2024-01-15T10:15:00Z"`
	LastActivity time.Time `json:"lastActivity" example:"2024-01-15T10:29:30Z"`
	IsActive     bool      `json:"isActive" example:"true"`
}

// UserSessionsResponse represents user sessions response
type UserSessionsResponse struct {
	Sessions    []UserSession `json:"sessions"`
	TotalCount  int           `json:"totalCount" example:"25"`
	ActiveCount int           `json:"activeCount" example:"20"`
	Timestamp   time.Time     `json:"timestamp" example:"2024-01-15T10:30:00Z"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error     string    `json:"error" example:"Resource not found"`
	Code      string    `json:"code" example:"NOT_FOUND"`
	Message   string    `json:"message" example:"The requested resource was not found"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-15T10:30:00Z"`
	Path      string    `json:"path" example:"/api/v1/rooms/invalid-id"`
}

// getHealthCheck handles health check endpoint
// @Summary Get service health status
// @Description Returns the health status of the collaboration service and its dependencies
// @Tags Health & Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse "Service is healthy"
// @Failure 503 {object} ErrorResponse "Service is unhealthy"
// @Router /api/v1/health [get]
func getHealthCheck(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Mock health check - in real implementation, check Redis, etc.
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
				"activeConnections": 25,
				"lastChecked":       time.Now(),
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}

// getMetrics returns metrics handler
// @Summary Get service metrics
// @Description Returns real-time metrics and statistics for the collaboration service
// @Tags Health & Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} MetricsResponse "Service metrics retrieved successfully"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/metrics [get]
func getMetrics(hub *websocket.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := hub.GetMetrics()

		response := MetricsResponse{
			TotalConnections:  metrics.TotalConnections,
			ActiveConnections: metrics.ActiveConnections,
			TotalRooms:        metrics.TotalRooms,
			ActiveRooms:       metrics.ActiveRooms,
			MessagesPerSecond: metrics.MessagesPerSecond,
			ErrorsPerSecond:   metrics.ErrorsPerSecond,
			AverageLatency:    12.5,    // Mock data
			MemoryUsage:       "256MB", // Mock data
			CPUUsage:          15.2,    // Mock data
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// getRooms returns rooms handler
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
func getRooms(hub *websocket.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Mock data - in real implementation, get from hub
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
}

// getRoomDetails returns room details handler
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
func getRoomDetails(hub *websocket.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		formID := vars["formId"]

		// Mock data - in real implementation, get from hub
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
}

// getRoomUsers returns room users handler
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
func getRoomUsers(hub *websocket.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		formID := vars["formId"]

		// Mock data - in real implementation, get from hub
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
}

// getUserSessions returns user sessions handler
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
func getUserSessions(hub *websocket.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Mock data - in real implementation, get from hub/redis
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
}

// getUserSessionDetails returns user session details handler
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
func getUserSessionDetails(hub *websocket.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userID := vars["userId"]

		// Mock data - in real implementation, get from hub/redis
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
func loggingMiddleware(logger interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			// Simple logging without zap dependency for now
			println("HTTP request:", r.Method, r.URL.Path, rw.statusCode, duration.String())
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
