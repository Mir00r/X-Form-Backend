package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EventType represents different types of WebSocket events
type EventType string

const (
	// Room management events
	EventJoinForm          EventType = "join:form"
	EventLeaveForm         EventType = "leave:form"
	EventUserJoined        EventType = "user:joined"
	EventUserLeft          EventType = "user:left"
	EventJoinFormResponse  EventType = "join:form:response"
	EventLeaveFormResponse EventType = "leave:form:response"

	// Cursor events
	EventCursorUpdate EventType = "cursor:update"

	// Question events
	EventQuestionUpdate EventType = "question:update"
	EventQuestionCreate EventType = "question:create"
	EventQuestionDelete EventType = "question:delete"

	// Form events
	EventFormUpdate EventType = "form:update"
	EventFormDelete EventType = "form:delete"

	// System events
	EventError      EventType = "error"
	EventHeartbeat  EventType = "heartbeat"
	EventDisconnect EventType = "disconnect"
	EventRateLimit  EventType = "rate:limit"
	EventPing       EventType = "ping"
	EventPong       EventType = "pong"
)

// Message represents a WebSocket message
type Message struct {
	Type      EventType   `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
	MessageID string      `json:"messageId"`
	UserID    string      `json:"userId,omitempty"`
	FormID    string      `json:"formId,omitempty"`
}

// NewMessage creates a new message with auto-generated ID and timestamp
func NewMessage(eventType EventType, payload interface{}) *Message {
	return &Message{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now(),
		MessageID: uuid.New().String(),
	}
}

// ToJSON converts message to JSON bytes
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON creates message from JSON bytes
func (m *Message) FromJSON(data []byte) error {
	return json.Unmarshal(data, m)
}

// User represents a connected user
type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	Avatar      string    `json:"avatar,omitempty"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
	ConnectedAt time.Time `json:"connectedAt"`
	LastSeen    time.Time `json:"lastSeen"`
	IsOnline    bool      `json:"isOnline"`
	SessionID   string    `json:"sessionId"`
}

// Room represents a collaboration room (form)
type Room struct {
	FormID    string            `json:"formId"`
	Users     map[string]*User  `json:"users"`
	Cursors   map[string]Cursor `json:"cursors"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
	MaxUsers  int               `json:"maxUsers"`
	IsActive  bool              `json:"isActive"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// NewRoom creates a new collaboration room
func NewRoom(formID string, maxUsers int) *Room {
	return &Room{
		FormID:    formID,
		Users:     make(map[string]*User),
		Cursors:   make(map[string]Cursor),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		MaxUsers:  maxUsers,
		IsActive:  true,
		Metadata:  make(map[string]string),
	}
}

// AddUser adds a user to the room
func (r *Room) AddUser(user *User) bool {
	if len(r.Users) >= r.MaxUsers {
		return false
	}
	r.Users[user.ID] = user
	r.UpdatedAt = time.Now()
	return true
}

// RemoveUser removes a user from the room
func (r *Room) RemoveUser(userID string) {
	delete(r.Users, userID)
	delete(r.Cursors, userID)
	r.UpdatedAt = time.Now()
}

// GetUserCount returns the number of users in the room
func (r *Room) GetUserCount() int {
	return len(r.Users)
}

// HasUser checks if a user is in the room
func (r *Room) HasUser(userID string) bool {
	_, exists := r.Users[userID]
	return exists
}

// GetUserList returns a slice of users in the room
func (r *Room) GetUserList() []*User {
	users := make([]*User, 0, len(r.Users))
	for _, user := range r.Users {
		users = append(users, user)
	}
	return users
}

// Cursor represents a user's cursor position
type Cursor struct {
	UserID    string    `json:"userId"`
	Position  Position  `json:"position"`
	Color     string    `json:"color"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Position represents cursor coordinates
type Position struct {
	X          int    `json:"x"`
	Y          int    `json:"y"`
	QuestionID string `json:"questionId,omitempty"`
	Section    string `json:"section,omitempty"`
}

// JoinFormPayload represents the payload for join:form event
type JoinFormPayload struct {
	FormID string `json:"formId" validate:"required"`
}

// LeaveFormPayload represents the payload for leave:form event
type LeaveFormPayload struct {
	FormID string `json:"formId" validate:"required"`
}

// CursorUpdatePayload represents the payload for cursor:update event
type CursorUpdatePayload struct {
	FormID   string   `json:"formId" validate:"required"`
	Position Position `json:"position" validate:"required"`
	Color    string   `json:"color,omitempty"`
	UserID   string   `json:"userId,omitempty"`
	User     *User    `json:"user,omitempty"`
}

// QuestionUpdatePayload represents the payload for question:update event
type QuestionUpdatePayload struct {
	FormID     string                 `json:"formId" validate:"required"`
	QuestionID string                 `json:"questionId" validate:"required"`
	Update     map[string]interface{} `json:"update" validate:"required"`
	Changes    map[string]interface{} `json:"changes" validate:"required"`
	Version    int                    `json:"version,omitempty"`
}

// QuestionCreatePayload represents the payload for question:create event
type QuestionCreatePayload struct {
	FormID   string       `json:"formId" validate:"required"`
	Question QuestionData `json:"question" validate:"required"`
	Position int          `json:"position,omitempty"`
}

// QuestionData represents question information
type QuestionData struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Title    string                 `json:"title"`
	Required bool                   `json:"required"`
	Options  []string               `json:"options,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// QuestionDeletePayload represents the payload for question:delete event
type QuestionDeletePayload struct {
	FormID     string `json:"formId" validate:"required"`
	QuestionID string `json:"questionId" validate:"required"`
}

// UserJoinedPayload represents the payload for user:joined event
type UserJoinedPayload struct {
	FormID string `json:"formId"`
	User   *User  `json:"user"`
}

// UserLeftPayload represents the payload for user:left event
type UserLeftPayload struct {
	FormID string `json:"formId"`
	UserID string `json:"userId"`
}

// ErrorPayload represents the payload for error events
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// JoinFormResponsePayload represents the response payload for join:form event
type JoinFormResponsePayload struct {
	FormID    string    `json:"formId"`
	UserID    string    `json:"userId"`
	Success   bool      `json:"success"`
	RoomUsers []*User   `json:"roomUsers"`
	Timestamp time.Time `json:"timestamp"`
}

// LeaveFormResponsePayload represents the response payload for leave:form event
type LeaveFormResponsePayload struct {
	FormID    string    `json:"formId"`
	UserID    string    `json:"userId"`
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
}

// PongPayload represents the payload for pong events
type PongPayload struct {
	Timestamp time.Time `json:"timestamp"`
}

// HeartbeatPayload represents the payload for heartbeat events
type HeartbeatPayload struct {
	Timestamp time.Time `json:"timestamp"`
	ServerID  string    `json:"serverId,omitempty"`
}

// RateLimitPayload represents the payload for rate limit events
type RateLimitPayload struct {
	Limit     int           `json:"limit"`
	Remaining int           `json:"remaining"`
	ResetTime time.Time     `json:"resetTime"`
	Window    time.Duration `json:"window"`
}

// Connection represents a WebSocket connection
type Connection struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	FormID    string    `json:"formId,omitempty"`
	Connected time.Time `json:"connected"`
	LastPing  time.Time `json:"lastPing"`
	IsActive  bool      `json:"isActive"`
	UserAgent string    `json:"userAgent,omitempty"`
	IPAddress string    `json:"ipAddress,omitempty"`
}

// SessionData represents session information stored in Redis
type SessionData struct {
	UserID      string            `json:"userId"`
	Connections []string          `json:"connections"`
	Rooms       []string          `json:"rooms"`
	Permissions []string          `json:"permissions"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	ExpiresAt   time.Time         `json:"expiresAt"`
}

// UserSession represents a user's session in a form
type UserSession struct {
	UserID    string    `json:"userId"`
	FormID    string    `json:"formId"`
	ClientID  string    `json:"clientId"`
	JoinedAt  time.Time `json:"joinedAt"`
	IsActive  bool      `json:"isActive"`
	UserAgent string    `json:"userAgent,omitempty"`
	IPAddress string    `json:"ipAddress,omitempty"`
}

// CursorPosition represents a user's cursor position in a form
type CursorPosition struct {
	UserID      string    `json:"userId"`
	FormID      string    `json:"formId"`
	QuestionID  string    `json:"questionId,omitempty"`
	X           int       `json:"x"`
	Y           int       `json:"y"`
	Section     string    `json:"section,omitempty"`
	LastUpdated time.Time `json:"lastUpdated"`
}

// QuestionUpdate represents a question update for conflict resolution
type QuestionUpdate struct {
	QuestionID string      `json:"questionId"`
	FormID     string      `json:"formId"`
	UserID     string      `json:"userId"`
	UpdateType string      `json:"updateType"` // create, update, delete
	Changes    interface{} `json:"changes"`
	Timestamp  time.Time   `json:"timestamp"`
	Version    int         `json:"version,omitempty"`
}

// FormPermission represents a user's permission for a specific form
type FormPermission struct {
	FormID      string    `json:"formId"`
	UserID      string    `json:"userId"`
	Role        string    `json:"role"` // owner, editor, viewer
	Permissions []string  `json:"permissions"`
	GrantedAt   time.Time `json:"grantedAt"`
	GrantedBy   string    `json:"grantedBy"`
	ExpiresAt   time.Time `json:"expiresAt,omitempty"`
}

// HasPermission checks if the permission exists in the list
func (fp *FormPermission) HasPermission(permission string) bool {
	for _, p := range fp.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// CanEdit checks if the user can edit the form
func (fp *FormPermission) CanEdit() bool {
	return fp.Role == "owner" || fp.Role == "editor" || fp.HasPermission("edit")
}

// CanView checks if the user can view the form
func (fp *FormPermission) CanView() bool {
	return fp.Role == "owner" || fp.Role == "editor" || fp.Role == "viewer" || fp.HasPermission("view")
}

// RateLimitInfo represents rate limiting information
type RateLimitInfo struct {
	UserID    string        `json:"userId"`
	Key       string        `json:"key"`
	Limit     int           `json:"limit"`
	Count     int           `json:"count"`
	Window    time.Duration `json:"window"`
	ResetTime time.Time     `json:"resetTime"`
	Blocked   bool          `json:"blocked"`
}

// IsAllowed checks if the rate limit allows the action
func (r *RateLimitInfo) IsAllowed() bool {
	return !r.Blocked && r.Count < r.Limit
}

// Remaining returns the remaining requests in the current window
func (r *RateLimitInfo) Remaining() int {
	if r.Count >= r.Limit {
		return 0
	}
	return r.Limit - r.Count
}

// KafkaEvent represents an event to be published to Kafka
type KafkaEvent struct {
	EventID   string      `json:"eventId"`
	Type      string      `json:"type"`
	FormID    string      `json:"formId"`
	UserID    string      `json:"userId"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	Source    string      `json:"source"`
	Version   string      `json:"version"`
}

// NewKafkaEvent creates a new Kafka event
func NewKafkaEvent(eventType, formID, userID string, data interface{}) *KafkaEvent {
	return &KafkaEvent{
		EventID:   uuid.New().String(),
		Type:      eventType,
		FormID:    formID,
		UserID:    userID,
		Data:      data,
		Timestamp: time.Now(),
		Source:    "collaboration-service",
		Version:   "1.0",
	}
}

// ToJSON converts the Kafka event to JSON
func (ke *KafkaEvent) ToJSON() ([]byte, error) {
	return json.Marshal(ke)
}

// WebSocketMetrics represents metrics for WebSocket connections
type WebSocketMetrics struct {
	TotalConnections   int64     `json:"totalConnections"`
	ActiveConnections  int64     `json:"activeConnections"`
	TotalRooms         int64     `json:"totalRooms"`
	ActiveRooms        int64     `json:"activeRooms"`
	MessagesPerSecond  int64     `json:"messagesPerSecond"`
	ErrorsPerSecond    int64     `json:"errorsPerSecond"`
	AverageLatency     int64     `json:"averageLatency"`
	ConnectionsPerUser int64     `json:"connectionsPerUser"`
	RoomsPerUser       int64     `json:"roomsPerUser"`
	LastUpdated        time.Time `json:"lastUpdated"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// Error implements the error interface
func (ve *ValidationError) Error() string {
	return ve.Message
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface for ValidationErrors
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	return ve[0].Message
}

// HasErrors checks if there are any validation errors
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}
