// Package dto contains Data Transfer Objects for the Form Service API
// Following microservices best practices for API contract stability
package dto

import (
	"time"
)

// =============================================================================
// Base Response DTOs
// =============================================================================

// BaseResponse contains common fields for all API responses
type BaseResponse struct {
	Success       bool      `json:"success"`
	Message       string    `json:"message,omitempty"`
	CorrelationID string    `json:"correlationId,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	Version       string    `json:"version"`
}

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	BaseResponse
	Data interface{} `json:"data,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	BaseResponse
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains detailed error information
type ErrorDetail struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   interface{}            `json:"details,omitempty"`
	Fields    map[string]string      `json:"fields,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	RequestID string                 `json:"requestId,omitempty"`
	Path      string                 `json:"path,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	BaseResponse
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination contains pagination metadata
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
	HasNext    bool  `json:"hasNext"`
	HasPrev    bool  `json:"hasPrev"`
}

// =============================================================================
// Form Request DTOs
// =============================================================================

// CreateFormRequestDTO represents the request to create a new form
type CreateFormRequestDTO struct {
	Title         string                     `json:"title" validate:"required,min=1,max=255" example:"Customer Feedback Form"`
	Description   string                     `json:"description" validate:"max=1000" example:"Please provide your feedback about our service"`
	IsAnonymous   bool                       `json:"isAnonymous" example:"false"`
	IsPublic      bool                       `json:"isPublic" example:"true"`
	AllowMultiple bool                       `json:"allowMultiple" example:"false"`
	ExpiresAt     *time.Time                 `json:"expiresAt,omitempty" example:"2024-12-31T23:59:59Z"`
	Settings      FormSettingsDTO            `json:"settings,omitempty"`
	Questions     []CreateQuestionRequestDTO `json:"questions" validate:"required,min=1,dive"`
	Tags          []string                   `json:"tags,omitempty" validate:"max=10,dive,max=50"`
	Category      string                     `json:"category,omitempty" validate:"max=100" example:"feedback"`
}

// UpdateFormRequestDTO represents the request to update an existing form
type UpdateFormRequestDTO struct {
	Title         *string          `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Description   *string          `json:"description,omitempty" validate:"omitempty,max=1000"`
	IsAnonymous   *bool            `json:"isAnonymous,omitempty"`
	IsPublic      *bool            `json:"isPublic,omitempty"`
	AllowMultiple *bool            `json:"allowMultiple,omitempty"`
	ExpiresAt     *time.Time       `json:"expiresAt,omitempty"`
	Settings      *FormSettingsDTO `json:"settings,omitempty"`
	Tags          []string         `json:"tags,omitempty" validate:"max=10,dive,max=50"`
	Category      *string          `json:"category,omitempty" validate:"omitempty,max=100"`
}

// FormSettingsDTO represents form configuration settings
type FormSettingsDTO struct {
	RequireLogin       bool                   `json:"requireLogin" example:"false"`
	CollectEmail       bool                   `json:"collectEmail" example:"true"`
	ShowProgressBar    bool                   `json:"showProgressBar" example:"true"`
	AllowDrafts        bool                   `json:"allowDrafts" example:"false"`
	NotifyOnSubmission bool                   `json:"notifyOnSubmission" example:"true"`
	CustomCSS          string                 `json:"customCss,omitempty" validate:"max=5000"`
	RedirectURL        string                 `json:"redirectUrl,omitempty" validate:"omitempty,url,max=500"`
	ThankYouMessage    string                 `json:"thankYouMessage,omitempty" validate:"max=1000"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// PublishFormRequestDTO represents the request to publish a form
type PublishFormRequestDTO struct {
	Message           string     `json:"message,omitempty" validate:"max=500" example:"Form is now live and ready for responses"`
	NotifySubscribers bool       `json:"notifySubscribers" example:"false"`
	ScheduleAt        *time.Time `json:"scheduleAt,omitempty" example:"2024-01-15T09:00:00Z"`
}

// =============================================================================
// Question Request DTOs
// =============================================================================

// CreateQuestionRequestDTO represents the request to create a new question
type CreateQuestionRequestDTO struct {
	Type        string                 `json:"type" validate:"required,oneof=text textarea number email date checkbox radio select file" example:"text"`
	Label       string                 `json:"label" validate:"required,min=1,max=500" example:"What is your name?"`
	Description string                 `json:"description,omitempty" validate:"max=1000" example:"Please enter your full name"`
	Required    bool                   `json:"required" example:"true"`
	Order       int                    `json:"order" validate:"min=0" example:"1"`
	Options     []QuestionOptionDTO    `json:"options,omitempty" validate:"dive"`
	Validation  *QuestionValidationDTO `json:"validation,omitempty"`
	Conditional *ConditionalLogicDTO   `json:"conditional,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// QuestionOptionDTO represents an option for select/radio/checkbox questions
type QuestionOptionDTO struct {
	Value string `json:"value" validate:"required,max=255" example:"option1"`
	Label string `json:"label" validate:"required,max=255" example:"Option 1"`
	Order int    `json:"order" validate:"min=0" example:"1"`
}

// QuestionValidationDTO represents validation rules for questions
type QuestionValidationDTO struct {
	MinLength    *int     `json:"minLength,omitempty" example:"2"`
	MaxLength    *int     `json:"maxLength,omitempty" example:"100"`
	Pattern      string   `json:"pattern,omitempty" example:"^[a-zA-Z ]+$"`
	MinValue     *float64 `json:"minValue,omitempty" example:"0"`
	MaxValue     *float64 `json:"maxValue,omitempty" example:"100"`
	AllowedTypes []string `json:"allowedTypes,omitempty" example:"jpg,png,pdf"`
	MaxFileSize  *int64   `json:"maxFileSize,omitempty" example:"5242880"`
}

// ConditionalLogicDTO represents conditional display logic
type ConditionalLogicDTO struct {
	ShowIf []ConditionDTO `json:"showIf,omitempty"`
	HideIf []ConditionDTO `json:"hideIf,omitempty"`
	Logic  string         `json:"logic" validate:"oneof=AND OR" example:"AND"`
}

// ConditionDTO represents a single condition
type ConditionDTO struct {
	QuestionID string `json:"questionId" validate:"required" example:"q1"`
	Operator   string `json:"operator" validate:"required,oneof=equals not_equals contains not_contains greater_than less_than" example:"equals"`
	Value      string `json:"value" example:"yes"`
}

// =============================================================================
// Form Response DTOs
// =============================================================================

// FormResponseDTO represents a complete form with all details
type FormResponseDTO struct {
	ID            string                `json:"id" example:"f123e4567-e89b-12d3-a456-426614174000"`
	Title         string                `json:"title" example:"Customer Feedback Form"`
	Description   string                `json:"description,omitempty" example:"Please provide your feedback"`
	Status        string                `json:"status" example:"published"`
	IsAnonymous   bool                  `json:"isAnonymous" example:"false"`
	IsPublic      bool                  `json:"isPublic" example:"true"`
	AllowMultiple bool                  `json:"allowMultiple" example:"false"`
	CreatedBy     UserInfoDTO           `json:"createdBy"`
	Settings      FormSettingsDTO       `json:"settings"`
	Questions     []QuestionResponseDTO `json:"questions"`
	Tags          []string              `json:"tags,omitempty"`
	Category      string                `json:"category,omitempty" example:"feedback"`
	Statistics    FormStatisticsDTO     `json:"statistics"`
	CreatedAt     time.Time             `json:"createdAt" example:"2024-01-01T12:00:00Z"`
	UpdatedAt     time.Time             `json:"updatedAt" example:"2024-01-01T12:00:00Z"`
	PublishedAt   *time.Time            `json:"publishedAt,omitempty" example:"2024-01-01T12:00:00Z"`
	ExpiresAt     *time.Time            `json:"expiresAt,omitempty" example:"2024-12-31T23:59:59Z"`
}

// FormSummaryDTO represents a brief form summary for listings
type FormSummaryDTO struct {
	ID          string            `json:"id" example:"f123e4567-e89b-12d3-a456-426614174000"`
	Title       string            `json:"title" example:"Customer Feedback Form"`
	Description string            `json:"description,omitempty" example:"Please provide your feedback"`
	Status      string            `json:"status" example:"published"`
	IsPublic    bool              `json:"isPublic" example:"true"`
	CreatedBy   UserInfoDTO       `json:"createdBy"`
	Statistics  FormStatisticsDTO `json:"statistics"`
	CreatedAt   time.Time         `json:"createdAt" example:"2024-01-01T12:00:00Z"`
	UpdatedAt   time.Time         `json:"updatedAt" example:"2024-01-01T12:00:00Z"`
	PublishedAt *time.Time        `json:"publishedAt,omitempty" example:"2024-01-01T12:00:00Z"`
	ExpiresAt   *time.Time        `json:"expiresAt,omitempty" example:"2024-12-31T23:59:59Z"`
}

// QuestionResponseDTO represents a question with all details
type QuestionResponseDTO struct {
	ID          string                 `json:"id" example:"q123e4567-e89b-12d3-a456-426614174000"`
	Type        string                 `json:"type" example:"text"`
	Label       string                 `json:"label" example:"What is your name?"`
	Description string                 `json:"description,omitempty" example:"Please enter your full name"`
	Required    bool                   `json:"required" example:"true"`
	Order       int                    `json:"order" example:"1"`
	Options     []QuestionOptionDTO    `json:"options,omitempty"`
	Validation  *QuestionValidationDTO `json:"validation,omitempty"`
	Conditional *ConditionalLogicDTO   `json:"conditional,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"createdAt" example:"2024-01-01T12:00:00Z"`
	UpdatedAt   time.Time              `json:"updatedAt" example:"2024-01-01T12:00:00Z"`
}

// UserInfoDTO represents basic user information
type UserInfoDTO struct {
	ID       string `json:"id" example:"u123e4567-e89b-12d3-a456-426614174000"`
	Username string `json:"username" example:"john_doe"`
	Email    string `json:"email,omitempty" example:"john@example.com"`
	Name     string `json:"name,omitempty" example:"John Doe"`
	Avatar   string `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
}

// FormStatisticsDTO represents form usage statistics
type FormStatisticsDTO struct {
	TotalResponses   int64      `json:"totalResponses" example:"150"`
	UniqueResponders int64      `json:"uniqueResponders" example:"140"`
	CompletionRate   float64    `json:"completionRate" example:"0.85"`
	AverageTime      int        `json:"averageTimeSeconds" example:"120"`
	LastResponse     *time.Time `json:"lastResponse,omitempty" example:"2024-01-01T12:00:00Z"`
	ResponseRate     float64    `json:"responseRate" example:"0.75"`
}

// =============================================================================
// Listing and Query DTOs
// =============================================================================

// FormListRequestDTO represents parameters for listing forms
type FormListRequestDTO struct {
	Page      int        `json:"page" validate:"min=1" example:"1"`
	PageSize  int        `json:"pageSize" validate:"min=1,max=100" example:"20"`
	Status    string     `json:"status,omitempty" validate:"omitempty,oneof=draft published closed archived" example:"published"`
	Category  string     `json:"category,omitempty" validate:"max=100" example:"feedback"`
	Tags      []string   `json:"tags,omitempty" validate:"max=10,dive,max=50"`
	Search    string     `json:"search,omitempty" validate:"max=255" example:"customer feedback"`
	SortBy    string     `json:"sortBy,omitempty" validate:"omitempty,oneof=created_at updated_at title response_count" example:"created_at"`
	SortOrder string     `json:"sortOrder,omitempty" validate:"omitempty,oneof=asc desc" example:"desc"`
	CreatedBy string     `json:"createdBy,omitempty" example:"u123e4567-e89b-12d3-a456-426614174000"`
	DateFrom  *time.Time `json:"dateFrom,omitempty" example:"2024-01-01T00:00:00Z"`
	DateTo    *time.Time `json:"dateTo,omitempty" example:"2024-12-31T23:59:59Z"`
}

// =============================================================================
// Health and Monitoring DTOs
// =============================================================================

// HealthCheckResponseDTO represents the health check response
type HealthCheckResponseDTO struct {
	Status       string                `json:"status" example:"healthy"`
	Service      string                `json:"service" example:"form-service"`
	Version      string                `json:"version" example:"1.0.0"`
	Environment  string                `json:"environment" example:"production"`
	Timestamp    time.Time             `json:"timestamp" example:"2024-01-01T12:00:00Z"`
	Uptime       string                `json:"uptime" example:"72h30m45s"`
	Dependencies HealthDependenciesDTO `json:"dependencies"`
	Metrics      HealthMetricsDTO      `json:"metrics"`
	Features     []string              `json:"features"`
}

// HealthDependenciesDTO represents the status of external dependencies
type HealthDependenciesDTO struct {
	Database       DependencyStatusDTO `json:"database"`
	AuthService    DependencyStatusDTO `json:"authService,omitempty"`
	EmailService   DependencyStatusDTO `json:"emailService,omitempty"`
	StorageService DependencyStatusDTO `json:"storageService,omitempty"`
}

// DependencyStatusDTO represents the status of a single dependency
type DependencyStatusDTO struct {
	Status       string    `json:"status" example:"healthy"`
	ResponseTime int64     `json:"responseTimeMs" example:"25"`
	LastChecked  time.Time `json:"lastChecked" example:"2024-01-01T12:00:00Z"`
	Error        string    `json:"error,omitempty"`
}

// HealthMetricsDTO represents service metrics
type HealthMetricsDTO struct {
	RequestsPerSecond   float64 `json:"requestsPerSecond" example:"10.5"`
	AverageResponseTime int     `json:"averageResponseTimeMs" example:"150"`
	ErrorRate           float64 `json:"errorRate" example:"0.001"`
	ActiveConnections   int     `json:"activeConnections" example:"25"`
	MemoryUsagePercent  float64 `json:"memoryUsagePercent" example:"65.5"`
	CPUUsagePercent     float64 `json:"cpuUsagePercent" example:"45.2"`
	DiskUsagePercent    float64 `json:"diskUsagePercent" example:"35.8"`
}

// =============================================================================
// Validation Error DTOs
// =============================================================================

// ValidationErrorDTO represents a validation error response
type ValidationErrorDTO struct {
	BaseResponse
	Error ValidationErrorDetail `json:"error"`
}

// ValidationErrorDetail contains detailed validation error information
type ValidationErrorDetail struct {
	Code      string                 `json:"code" example:"VALIDATION_ERROR"`
	Message   string                 `json:"message" example:"Input validation failed"`
	Fields    map[string][]string    `json:"fields"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	RequestID string                 `json:"requestId"`
	Path      string                 `json:"path"`
	Timestamp time.Time              `json:"timestamp"`
}

// =============================================================================
// API Metadata DTOs
// =============================================================================

// APIInfoDTO represents API information for discovery
type APIInfoDTO struct {
	Name        string      `json:"name" example:"Form Service API"`
	Version     string      `json:"version" example:"v1.0.0"`
	Description string      `json:"description" example:"Comprehensive form management service"`
	Contact     ContactDTO  `json:"contact"`
	License     LicenseDTO  `json:"license"`
	Servers     []ServerDTO `json:"servers"`
	Features    []string    `json:"features"`
}

// ContactDTO represents API contact information
type ContactDTO struct {
	Name  string `json:"name" example:"API Team"`
	Email string `json:"email" example:"api@example.com"`
	URL   string `json:"url" example:"https://api.example.com/support"`
}

// LicenseDTO represents API license information
type LicenseDTO struct {
	Name string `json:"name" example:"MIT"`
	URL  string `json:"url" example:"https://opensource.org/licenses/MIT"`
}

// ServerDTO represents API server information
type ServerDTO struct {
	URL         string `json:"url" example:"https://api.example.com/v1"`
	Description string `json:"description" example:"Production server"`
	Environment string `json:"environment" example:"production"`
}

// =============================================================================
// Rate Limiting DTOs
// =============================================================================

// RateLimitInfoDTO represents rate limiting information
type RateLimitInfoDTO struct {
	Limit     int       `json:"limit" example:"100"`
	Remaining int       `json:"remaining" example:"85"`
	Reset     time.Time `json:"reset" example:"2024-01-01T13:00:00Z"`
	Window    string    `json:"window" example:"1h"`
}

// RateLimitExceededDTO represents rate limit exceeded error
type RateLimitExceededDTO struct {
	BaseResponse
	Error      ErrorDetail      `json:"error"`
	RateLimit  RateLimitInfoDTO `json:"rateLimit"`
	RetryAfter int              `json:"retryAfterSeconds" example:"3600"`
}
