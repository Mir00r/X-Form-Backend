package models

import (
	"time"
)

// APIResponse represents a standard API response
type APIResponse struct {
	Success   bool        `json:"success" example:"true"`
	Message   string      `json:"message" example:"Operation completed successfully"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Meta      *MetaInfo   `json:"meta,omitempty"`
	RequestID string      `json:"request_id,omitempty" example:"uuid-request-id"`
	Timestamp time.Time   `json:"timestamp" example:"2024-01-01T00:00:00Z"`
}

// ErrorInfo represents detailed error information
type ErrorInfo struct {
	Code    string `json:"code" example:"VALIDATION_ERROR"`
	Message string `json:"message" example:"Invalid input parameters"`
	Details string `json:"details,omitempty" example:"Field 'email' is required"`
	Field   string `json:"field,omitempty" example:"email"`
}

// MetaInfo represents metadata for responses
type MetaInfo struct {
	Page       int `json:"page,omitempty" example:"1"`
	Limit      int `json:"limit,omitempty" example:"10"`
	Total      int `json:"total,omitempty" example:"100"`
	TotalPages int `json:"total_pages,omitempty" example:"10"`
}

// ServiceConfig represents the configuration for a microservice
type ServiceConfig struct {
	Name           string            `json:"name" yaml:"name" example:"auth-service"`
	BaseURL        string            `json:"base_url" yaml:"base_url" example:"http://auth-service:3001"`
	Prefix         string            `json:"prefix" yaml:"prefix" example:"/api/auth"`
	Timeout        int               `json:"timeout" yaml:"timeout" example:"30"`
	HealthPath     string            `json:"health_path" yaml:"health_path" example:"/health"`
	Retries        int               `json:"retries" yaml:"retries" example:"3"`
	CircuitBreaker bool              `json:"circuit_breaker" yaml:"circuit_breaker" example:"true"`
	RateLimit      *RateLimitConfig  `json:"rate_limit,omitempty" yaml:"rate_limit,omitempty"`
	Headers        map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond int           `json:"requests_per_second" yaml:"requests_per_second" example:"100"`
	BurstSize         int           `json:"burst_size" yaml:"burst_size" example:"10"`
	Window            time.Duration `json:"window" yaml:"window" example:"60s"`
}

// GatewayConfig represents the overall gateway configuration
type GatewayConfig struct {
	Host        string                    `json:"host" yaml:"host" example:"0.0.0.0"`
	Port        int                       `json:"port" yaml:"port" example:"8080"`
	Environment string                    `json:"environment" yaml:"environment" example:"development"`
	LogLevel    string                    `json:"log_level" yaml:"log_level" example:"info"`
	JWTSecret   string                    `json:"jwt_secret" yaml:"jwt_secret"`
	Services    map[string]*ServiceConfig `json:"services" yaml:"services"`
	CORS        *CORSConfig               `json:"cors,omitempty" yaml:"cors,omitempty"`
	RateLimit   *RateLimitConfig          `json:"rate_limit,omitempty" yaml:"rate_limit,omitempty"`
	Monitoring  *MonitoringConfig         `json:"monitoring,omitempty" yaml:"monitoring,omitempty"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods" yaml:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers" yaml:"allowed_headers"`

	MaxAge int `json:"max_age" yaml:"max_age" example:"86400"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	Enabled     bool   `json:"enabled" yaml:"enabled" example:"true"`
	MetricsPath string `json:"metrics_path" yaml:"metrics_path" example:"/metrics"`
	HealthPath  string `json:"health_path" yaml:"health_path" example:"/health"`
}

// User represents a user in the system
type User struct {
	ID        string    `json:"id" example:"user-123"`
	Email     string    `json:"email" example:"user@example.com"`
	Username  string    `json:"username" example:"johndoe"`
	Roles     []string  `json:"roles" example:"admin,user"`
	Active    bool      `json:"active" example:"true"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID    string   `json:"user_id" example:"user-123"`
	Email     string   `json:"email" example:"user@example.com"`
	Username  string   `json:"username" example:"johndoe"`
	Roles     []string `json:"roles" example:"admin,user"`
	Issuer    string   `json:"iss" example:"x-form-backend"`
	Subject   string   `json:"sub" example:"user-123"`
	Audience  string   `json:"aud" example:"x-form-frontend"`
	IssuedAt  int64    `json:"iat" example:"1640995200"`
	ExpiresAt int64    `json:"exp" example:"1641081600"`
}

// Form represents a form in the system
type Form struct {
	ID          string       `json:"id" example:"form-123"`
	Title       string       `json:"title" example:"Customer Feedback Form"`
	Description string       `json:"description" example:"Please provide your feedback"`
	CreatorID   string       `json:"creator_id" example:"user-123"`
	Status      string       `json:"status" example:"active"`
	Fields      []FormField  `json:"fields"`
	Settings    FormSettings `json:"settings"`
	CreatedAt   time.Time    `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time    `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// FormField represents a field in a form
type FormField struct {
	ID         string                 `json:"id" example:"field-123"`
	Type       string                 `json:"type" example:"text"`
	Label      string                 `json:"label" example:"Full Name"`
	Required   bool                   `json:"required" example:"true"`
	Options    []string               `json:"options,omitempty"`
	Validation map[string]interface{} `json:"validation,omitempty"`
	Order      int                    `json:"order" example:"1"`
}

// FormSettings represents form settings
type FormSettings struct {
	IsPublic       bool       `json:"is_public" example:"true"`
	AllowAnonymous bool       `json:"allow_anonymous" example:"false"`
	MaxResponses   int        `json:"max_responses" example:"1000"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	RequireAuth    bool       `json:"require_auth" example:"false"`
	NotifyOnSubmit bool       `json:"notify_on_submit" example:"true"`
}

// FormResponse represents a response to a form
type FormResponse struct {
	ID        string                 `json:"id" example:"response-123"`
	FormID    string                 `json:"form_id" example:"form-123"`
	UserID    string                 `json:"user_id,omitempty" example:"user-123"`
	Data      map[string]interface{} `json:"data"`
	IP        string                 `json:"ip" example:"192.168.1.1"`
	UserAgent string                 `json:"user_agent" example:"Mozilla/5.0..."`
	CreatedAt time.Time              `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// Analytics represents analytics data
type Analytics struct {
	FormID           string    `json:"form_id" example:"form-123"`
	TotalResponses   int       `json:"total_responses" example:"150"`
	UniqueResponders int       `json:"unique_responders" example:"145"`
	AverageTime      float64   `json:"average_completion_time" example:"120.5"`
	CompletionRate   float64   `json:"completion_rate" example:"0.85"`
	Period           string    `json:"period" example:"30d"`
	GeneratedAt      time.Time `json:"generated_at" example:"2024-01-01T00:00:00Z"`
}

// Collaboration represents collaboration data
type Collaboration struct {
	FormID     string     `json:"form_id" example:"form-123"`
	UserID     string     `json:"user_id" example:"user-123"`
	Permission string     `json:"permission" example:"edit"`
	InvitedBy  string     `json:"invited_by" example:"user-456"`
	Status     string     `json:"status" example:"accepted"`
	InvitedAt  time.Time  `json:"invited_at" example:"2024-01-01T00:00:00Z"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty" example:"2024-01-01T00:00:00Z"`
}

// Notification represents a real-time notification
type Notification struct {
	ID        string                 `json:"id" example:"notif-123"`
	UserID    string                 `json:"user_id" example:"user-123"`
	Type      string                 `json:"type" example:"form_response"`
	Title     string                 `json:"title" example:"New Form Response"`
	Message   string                 `json:"message" example:"You have a new response"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Read      bool                   `json:"read" example:"false"`
	CreatedAt time.Time              `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// FileUpload represents a file upload
type FileUpload struct {
	ID        string    `json:"id" example:"file-123"`
	UserID    string    `json:"user_id" example:"user-123"`
	FormID    string    `json:"form_id,omitempty" example:"form-123"`
	FieldID   string    `json:"field_id,omitempty" example:"field-123"`
	Filename  string    `json:"filename" example:"document.pdf"`
	Size      int64     `json:"size" example:"1024000"`
	MimeType  string    `json:"mime_type" example:"application/pdf"`
	URL       string    `json:"url" example:"https://s3.amazonaws.com/bucket/file.pdf"`
	Status    string    `json:"status" example:"uploaded"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// Webhook represents a webhook configuration
type Webhook struct {
	ID        string            `json:"id" example:"webhook-123"`
	FormID    string            `json:"form_id" example:"form-123"`
	URL       string            `json:"url" example:"https://example.com/webhook"`
	Events    []string          `json:"events" example:"form.response.created"`
	Headers   map[string]string `json:"headers,omitempty"`
	Secret    string            `json:"secret,omitempty"`
	Active    bool              `json:"active" example:"true"`
	CreatedAt time.Time         `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID         string                 `json:"id" example:"audit-123"`
	UserID     string                 `json:"user_id" example:"user-123"`
	Action     string                 `json:"action" example:"form.created"`
	Resource   string                 `json:"resource" example:"form"`
	ResourceID string                 `json:"resource_id" example:"form-123"`
	Details    map[string]interface{} `json:"details,omitempty"`
	IP         string                 `json:"ip" example:"192.168.1.1"`
	UserAgent  string                 `json:"user_agent" example:"Mozilla/5.0..."`
	CreatedAt  time.Time              `json:"created_at" example:"2024-01-01T00:00:00Z"`
}
