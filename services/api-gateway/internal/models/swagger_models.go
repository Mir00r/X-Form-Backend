package models

import (
	"time"
)

// OpenAPIInfo represents comprehensive API information for Swagger
type OpenAPIInfo struct {
	Title          string `json:"title" example:"X-Form API Gateway"`
	Description    string `json:"description" example:"Comprehensive API Gateway for X-Form microservices architecture"`
	Version        string `json:"version" example:"2.0.0"`
	TermsOfService string `json:"termsOfService" example:"https://x-form.com/terms"`
	Contact        struct {
		Name  string `json:"name" example:"X-Form API Team"`
		URL   string `json:"url" example:"https://x-form.com/support"`
		Email string `json:"email" example:"api-support@x-form.com"`
	} `json:"contact"`
	License struct {
		Name string `json:"name" example:"MIT"`
		URL  string `json:"url" example:"https://opensource.org/licenses/MIT"`
	} `json:"license"`
}

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page   int    `form:"page" json:"page" example:"1" minimum:"1" description:"Page number"`
	Limit  int    `form:"limit" json:"limit" example:"10" minimum:"1" maximum:"100" description:"Number of items per page"`
	Sort   string `form:"sort" json:"sort" example:"created_at" description:"Sort field"`
	Order  string `form:"order" json:"order" example:"desc" enums:"asc,desc" description:"Sort order"`
	Search string `form:"search" json:"search" example:"search term" description:"Search term"`
	Filter string `form:"filter" json:"filter" example:"status:active" description:"Filter criteria"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	CurrentPage  int  `json:"current_page" example:"1"`
	PerPage      int  `json:"per_page" example:"10"`
	TotalPages   int  `json:"total_pages" example:"5"`
	TotalItems   int  `json:"total_items" example:"50"`
	HasNext      bool `json:"has_next" example:"true"`
	HasPrevious  bool `json:"has_previous" example:"false"`
	NextPage     *int `json:"next_page,omitempty" example:"2"`
	PreviousPage *int `json:"previous_page,omitempty"`
}

// StandardAPIResponse represents the standard API response format
type StandardAPIResponse struct {
	Success    bool                `json:"success" example:"true" description:"Indicates if the request was successful"`
	Message    string              `json:"message" example:"Operation completed successfully" description:"Human-readable message"`
	Data       interface{}         `json:"data,omitempty" description:"Response data"`
	Error      *DetailedError      `json:"error,omitempty" description:"Error details if success is false"`
	Meta       *ResponseMetadata   `json:"meta,omitempty" description:"Response metadata"`
	Pagination *PaginationResponse `json:"pagination,omitempty" description:"Pagination information"`
	RequestID  string              `json:"request_id" example:"req_123456789" description:"Unique request identifier"`
	Timestamp  time.Time           `json:"timestamp" example:"2025-09-06T12:00:00Z" description:"Response timestamp"`
}

// DetailedError represents comprehensive error information
type DetailedError struct {
	Code       string            `json:"code" example:"VALIDATION_ERROR" description:"Error code"`
	Message    string            `json:"message" example:"Validation failed" description:"Error message"`
	Details    string            `json:"details,omitempty" example:"Field 'email' is required" description:"Detailed error information"`
	Field      string            `json:"field,omitempty" example:"email" description:"Field that caused the error"`
	Timestamp  time.Time         `json:"timestamp" example:"2025-09-06T12:00:00Z" description:"Error timestamp"`
	TraceID    string            `json:"trace_id,omitempty" example:"trace_123456" description:"Trace identifier for debugging"`
	Suggestion string            `json:"suggestion,omitempty" example:"Please provide a valid email address" description:"Suggested fix"`
	MoreInfo   string            `json:"more_info,omitempty" example:"https://docs.x-form.com/errors/validation" description:"Link to documentation"`
	Context    map[string]string `json:"context,omitempty" description:"Additional error context"`
}

// ResponseMetadata represents additional response metadata
type ResponseMetadata struct {
	RequestDuration string            `json:"request_duration,omitempty" example:"150ms" description:"Request processing time"`
	APIVersion      string            `json:"api_version" example:"v1" description:"API version used"`
	ServerInstance  string            `json:"server_instance,omitempty" example:"gateway-01" description:"Server instance that handled the request"`
	RateLimit       *RateLimitInfo    `json:"rate_limit,omitempty" description:"Rate limiting information"`
	Headers         map[string]string `json:"headers,omitempty" description:"Additional response headers"`
	Cache           *CacheInfo        `json:"cache,omitempty" description:"Cache information"`
}

// RateLimitInfo represents rate limiting information
type RateLimitInfo struct {
	Limit     int       `json:"limit" example:"100" description:"Request limit per window"`
	Remaining int       `json:"remaining" example:"95" description:"Remaining requests in current window"`
	Reset     time.Time `json:"reset" example:"2025-09-06T13:00:00Z" description:"When the rate limit resets"`
	Window    string    `json:"window" example:"1h" description:"Rate limit window duration"`
}

// CacheInfo represents caching information
type CacheInfo struct {
	Cached    bool      `json:"cached" example:"true" description:"Whether response was served from cache"`
	TTL       int       `json:"ttl,omitempty" example:"300" description:"Cache TTL in seconds"`
	Key       string    `json:"key,omitempty" example:"cache_key_123" description:"Cache key used"`
	ExpiresAt time.Time `json:"expires_at,omitempty" example:"2025-09-06T12:05:00Z" description:"Cache expiration time"`
}

// HealthCheckResponse represents comprehensive health check response
type HealthCheckResponse struct {
	Status       string                      `json:"status" example:"healthy" enums:"healthy,unhealthy,degraded" description:"Overall health status"`
	Service      string                      `json:"service" example:"api-gateway" description:"Service name"`
	Version      string                      `json:"version" example:"1.0.0" description:"Service version"`
	Environment  string                      `json:"environment" example:"production" description:"Deployment environment"`
	Timestamp    time.Time                   `json:"timestamp" example:"2025-09-06T12:00:00Z" description:"Health check timestamp"`
	Uptime       string                      `json:"uptime" example:"72h30m15s" description:"Service uptime"`
	Checks       map[string]ServiceHealth    `json:"checks" description:"Individual service health checks"`
	System       SystemHealth                `json:"system" description:"System resource information"`
	Dependencies map[string]DependencyHealth `json:"dependencies" description:"External dependencies health"`
}

// ServiceHealth represents individual service health
type ServiceHealth struct {
	Status    string        `json:"status" example:"healthy" enums:"healthy,unhealthy,degraded"`
	LastCheck time.Time     `json:"last_check" example:"2025-09-06T12:00:00Z"`
	Duration  time.Duration `json:"duration" example:"50ms"`
	Error     string        `json:"error,omitempty" example:"Connection timeout"`
	Details   interface{}   `json:"details,omitempty"`
}

// SystemHealth represents system resource health
type SystemHealth struct {
	Memory struct {
		Used    uint64  `json:"used" example:"524288000" description:"Used memory in bytes"`
		Total   uint64  `json:"total" example:"2147483648" description:"Total memory in bytes"`
		Percent float64 `json:"percent" example:"24.4" description:"Memory usage percentage"`
	} `json:"memory"`
	CPU struct {
		Percent float64 `json:"percent" example:"15.7" description:"CPU usage percentage"`
		Cores   int     `json:"cores" example:"4" description:"Number of CPU cores"`
	} `json:"cpu"`
	Disk struct {
		Used    uint64  `json:"used" example:"1073741824" description:"Used disk space in bytes"`
		Total   uint64  `json:"total" example:"10737418240" description:"Total disk space in bytes"`
		Percent float64 `json:"percent" example:"10.0" description:"Disk usage percentage"`
	} `json:"disk"`
	Goroutines int `json:"goroutines" example:"25" description:"Number of active goroutines"`
}

// DependencyHealth represents external dependency health
type DependencyHealth struct {
	Status    string        `json:"status" example:"healthy" enums:"healthy,unhealthy,degraded"`
	URL       string        `json:"url" example:"https://auth-service:3001/health"`
	LastCheck time.Time     `json:"last_check" example:"2025-09-06T12:00:00Z"`
	Duration  time.Duration `json:"duration" example:"100ms"`
	Error     string        `json:"error,omitempty" example:"Service unavailable"`
	Version   string        `json:"version,omitempty" example:"1.2.0"`
}

// AuthenticationRequest represents comprehensive auth request
type AuthenticationRequest struct {
	Email      string      `json:"email" binding:"required,email" example:"user@example.com" description:"User email address"`
	Password   string      `json:"password" binding:"required,min=8,max=128" example:"SecurePassword123!" description:"User password (8-128 characters)"`
	RememberMe bool        `json:"remember_me" example:"false" description:"Extended session duration"`
	DeviceInfo *DeviceInfo `json:"device_info,omitempty" description:"Device information for security"`
	TwoFactor  string      `json:"two_factor,omitempty" example:"123456" description:"Two-factor authentication code"`
}

// DeviceInfo represents device information for security
type DeviceInfo struct {
	UserAgent string `json:"user_agent" example:"Mozilla/5.0..." description:"Browser user agent"`
	IPAddress string `json:"ip_address" example:"192.168.1.1" description:"Client IP address"`
	Platform  string `json:"platform" example:"web" enums:"web,mobile,desktop" description:"Client platform"`
	DeviceID  string `json:"device_id,omitempty" example:"device_123" description:"Unique device identifier"`
}

// RegistrationRequest represents comprehensive user registration
type RegistrationRequest struct {
	FirstName       string      `json:"first_name" binding:"required,min=2,max=50" example:"John" description:"User first name"`
	LastName        string      `json:"last_name" binding:"required,min=2,max=50" example:"Doe" description:"User last name"`
	Email           string      `json:"email" binding:"required,email" example:"john.doe@example.com" description:"User email address"`
	Password        string      `json:"password" binding:"required,min=8,max=128" example:"SecurePassword123!" description:"User password"`
	ConfirmPassword string      `json:"confirm_password" binding:"required" example:"SecurePassword123!" description:"Password confirmation"`
	PhoneNumber     string      `json:"phone_number,omitempty" example:"+1234567890" description:"User phone number"`
	DateOfBirth     *time.Time  `json:"date_of_birth,omitempty" example:"1990-01-01T00:00:00Z" description:"User date of birth"`
	Terms           bool        `json:"terms" binding:"required" example:"true" description:"Acceptance of terms and conditions"`
	Newsletter      bool        `json:"newsletter" example:"false" description:"Newsletter subscription preference"`
	DeviceInfo      *DeviceInfo `json:"device_info,omitempty" description:"Device information"`
}

// AuthenticationResponse represents comprehensive auth response
type AuthenticationResponse struct {
	AccessToken      string       `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." description:"JWT access token"`
	RefreshToken     string       `json:"refresh_token" example:"refresh_token_example" description:"Refresh token for obtaining new access tokens"`
	TokenType        string       `json:"token_type" example:"Bearer" description:"Token type"`
	ExpiresIn        int          `json:"expires_in" example:"3600" description:"Access token expiration time in seconds"`
	ExpiresAt        time.Time    `json:"expires_at" example:"2025-09-06T13:00:00Z" description:"Access token expiration timestamp"`
	Scope            []string     `json:"scope" example:"read,write" description:"Token permissions"`
	User             DetailedUser `json:"user" description:"User information"`
	Session          SessionInfo  `json:"session" description:"Session information"`
	Permissions      []Permission `json:"permissions" description:"User permissions"`
	TwoFactorEnabled bool         `json:"two_factor_enabled" example:"false" description:"Whether 2FA is enabled"`
}

// DetailedUser represents comprehensive user information
type DetailedUser struct {
	ID               string          `json:"id" example:"123e4567-e89b-12d3-a456-426614174000" description:"Unique user identifier"`
	FirstName        string          `json:"first_name" example:"John" description:"User first name"`
	LastName         string          `json:"last_name" example:"Doe" description:"User last name"`
	FullName         string          `json:"full_name" example:"John Doe" description:"User full name"`
	Email            string          `json:"email" example:"john.doe@example.com" description:"User email address"`
	EmailVerified    bool            `json:"email_verified" example:"true" description:"Email verification status"`
	PhoneNumber      string          `json:"phone_number,omitempty" example:"+1234567890" description:"User phone number"`
	PhoneVerified    bool            `json:"phone_verified" example:"false" description:"Phone verification status"`
	Avatar           string          `json:"avatar,omitempty" example:"https://example.com/avatar.jpg" description:"User avatar URL"`
	Role             string          `json:"role" example:"user" enums:"admin,moderator,user" description:"User role"`
	Status           string          `json:"status" example:"active" enums:"active,inactive,suspended" description:"User status"`
	Preferences      UserPreferences `json:"preferences" description:"User preferences"`
	CreatedAt        time.Time       `json:"created_at" example:"2025-01-01T00:00:00Z" description:"Account creation timestamp"`
	UpdatedAt        time.Time       `json:"updated_at" example:"2025-09-06T12:00:00Z" description:"Last update timestamp"`
	LastLoginAt      *time.Time      `json:"last_login_at,omitempty" example:"2025-09-06T11:30:00Z" description:"Last login timestamp"`
	LoginCount       int             `json:"login_count" example:"42" description:"Total login count"`
	TwoFactorEnabled bool            `json:"two_factor_enabled" example:"false" description:"Two-factor authentication status"`
}

// UserPreferences represents user preferences
type UserPreferences struct {
	Language      string               `json:"language" example:"en" description:"Preferred language"`
	Timezone      string               `json:"timezone" example:"UTC" description:"User timezone"`
	Theme         string               `json:"theme" example:"light" enums:"light,dark,auto" description:"UI theme preference"`
	Notifications NotificationSettings `json:"notifications" description:"Notification preferences"`
}

// NotificationSettings represents notification preferences
type NotificationSettings struct {
	Email bool `json:"email" example:"true" description:"Email notifications enabled"`
	SMS   bool `json:"sms" example:"false" description:"SMS notifications enabled"`
	Push  bool `json:"push" example:"true" description:"Push notifications enabled"`
}

// SessionInfo represents session information
type SessionInfo struct {
	ID        string    `json:"id" example:"sess_123456789" description:"Session identifier"`
	CreatedAt time.Time `json:"created_at" example:"2025-09-06T12:00:00Z" description:"Session creation time"`
	ExpiresAt time.Time `json:"expires_at" example:"2025-09-06T20:00:00Z" description:"Session expiration time"`
	IPAddress string    `json:"ip_address" example:"192.168.1.1" description:"Session IP address"`
	UserAgent string    `json:"user_agent" example:"Mozilla/5.0..." description:"Session user agent"`
	IsActive  bool      `json:"is_active" example:"true" description:"Session active status"`
}

// Permission represents user permission
type Permission struct {
	ID          string   `json:"id" example:"perm_123" description:"Permission identifier"`
	Name        string   `json:"name" example:"forms.create" description:"Permission name"`
	Description string   `json:"description" example:"Create new forms" description:"Permission description"`
	Resource    string   `json:"resource" example:"forms" description:"Resource type"`
	Actions     []string `json:"actions" example:"create,read,update" description:"Allowed actions"`
}

// ValidationError represents field validation errors
type ValidationError struct {
	Field   string `json:"field" example:"email" description:"Field name that failed validation"`
	Value   string `json:"value" example:"invalid-email" description:"Invalid value provided"`
	Message string `json:"message" example:"Invalid email format" description:"Validation error message"`
	Code    string `json:"code" example:"INVALID_FORMAT" description:"Validation error code"`
}

// ValidationErrorResponse represents validation error response
type ValidationErrorResponse struct {
	Success   bool              `json:"success" example:"false"`
	Message   string            `json:"message" example:"Validation failed"`
	Errors    []ValidationError `json:"errors" description:"List of validation errors"`
	RequestID string            `json:"request_id" example:"req_123456789"`
	Timestamp time.Time         `json:"timestamp" example:"2025-09-06T12:00:00Z"`
}
