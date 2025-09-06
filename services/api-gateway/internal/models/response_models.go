package models

import (
	"time"
)

// ResponseSubmissionRequest represents a comprehensive response submission
type ResponseSubmissionRequest struct {
	FormID     string                 `json:"form_id" binding:"required" example:"form_123456789" description:"Form ID to submit response to"`
	Responses  map[string]interface{} `json:"responses" binding:"required" description:"Field responses as key-value pairs"`
	IsComplete bool                   `json:"is_complete" example:"true" description:"Whether submission is complete"`
	IsDraft    bool                   `json:"is_draft" example:"false" description:"Whether this is a draft submission"`
	DeviceInfo *DeviceInfo            `json:"device_info,omitempty" description:"Device information"`
	Location   *LocationInfo          `json:"location,omitempty" description:"Geographic location"`
	UTMParams  *UTMParameters         `json:"utm_params,omitempty" description:"UTM tracking parameters"`
	Referrer   string                 `json:"referrer,omitempty" example:"https://google.com" description:"Referrer URL"`
	UserAgent  string                 `json:"user_agent,omitempty" description:"User agent string"`
	IPAddress  string                 `json:"ip_address,omitempty" description:"Client IP address"`
	StartedAt  *time.Time             `json:"started_at,omitempty" example:"2025-09-06T12:00:00Z" description:"When user started filling form"`
	Metadata   map[string]interface{} `json:"metadata,omitempty" description:"Additional metadata"`
}

// LocationInfo represents geographic location information
type LocationInfo struct {
	Country     string  `json:"country,omitempty" example:"United States" description:"Country name"`
	CountryCode string  `json:"country_code,omitempty" example:"US" description:"Country code"`
	Region      string  `json:"region,omitempty" example:"California" description:"Region/state"`
	City        string  `json:"city,omitempty" example:"San Francisco" description:"City"`
	Latitude    float64 `json:"latitude,omitempty" example:"37.7749" description:"Latitude"`
	Longitude   float64 `json:"longitude,omitempty" example:"-122.4194" description:"Longitude"`
	Timezone    string  `json:"timezone,omitempty" example:"America/Los_Angeles" description:"Timezone"`
}

// UTMParameters represents UTM tracking parameters
type UTMParameters struct {
	Source   string `json:"utm_source,omitempty" example:"google" description:"Traffic source"`
	Medium   string `json:"utm_medium,omitempty" example:"cpc" description:"Marketing medium"`
	Campaign string `json:"utm_campaign,omitempty" example:"summer_sale" description:"Campaign name"`
	Term     string `json:"utm_term,omitempty" example:"form_builder" description:"Search term"`
	Content  string `json:"utm_content,omitempty" example:"banner_ad" description:"Ad content"`
}

// ResponseDetails represents a comprehensive form response
type ResponseDetails struct {
	ID                 string                       `json:"id" example:"response_123456789" description:"Unique response identifier"`
	FormID             string                       `json:"form_id" example:"form_123456789" description:"Associated form ID"`
	UserID             string                       `json:"user_id,omitempty" example:"user_123" description:"User ID (if authenticated)"`
	SessionID          string                       `json:"session_id" example:"session_123456" description:"Session identifier"`
	Status             string                       `json:"status" example:"completed" enums:"draft,completed,abandoned" description:"Response status"`
	IsComplete         bool                         `json:"is_complete" example:"true" description:"Whether response is complete"`
	IsDraft            bool                         `json:"is_draft" example:"false" description:"Whether response is a draft"`
	IsAnonymous        bool                         `json:"is_anonymous" example:"true" description:"Whether response is anonymous"`
	Responses          map[string]ResponseValue     `json:"responses" description:"Field responses"`
	Score              *ResponseScore               `json:"score,omitempty" description:"Response scoring (if applicable)"`
	Validation         ResponseValidation           `json:"validation" description:"Response validation results"`
	Timing             ResponseTiming               `json:"timing" description:"Response timing information"`
	DeviceInfo         *DeviceInfo                  `json:"device_info,omitempty" description:"Device information"`
	Location           *LocationInfo                `json:"location,omitempty" description:"Geographic location"`
	UTMParams          *UTMParameters               `json:"utm_params,omitempty" description:"UTM parameters"`
	Referrer           string                       `json:"referrer,omitempty" example:"https://google.com" description:"Referrer URL"`
	UserAgent          string                       `json:"user_agent,omitempty" description:"User agent string"`
	IPAddress          string                       `json:"ip_address,omitempty" description:"Client IP address (anonymized)"`
	Language           string                       `json:"language,omitempty" example:"en-US" description:"User language"`
	CreatedAt          time.Time                    `json:"created_at" example:"2025-09-06T12:00:00Z" description:"Response creation time"`
	UpdatedAt          time.Time                    `json:"updated_at" example:"2025-09-06T12:05:00Z" description:"Last update time"`
	SubmittedAt        *time.Time                   `json:"submitted_at,omitempty" example:"2025-09-06T12:05:00Z" description:"Submission time"`
	StartedAt          *time.Time                   `json:"started_at,omitempty" example:"2025-09-06T12:00:00Z" description:"Start time"`
	CompletedAt        *time.Time                   `json:"completed_at,omitempty" example:"2025-09-06T12:05:00Z" description:"Completion time"`
	Form               *FormSummary                 `json:"form,omitempty" description:"Associated form summary"`
	User               *BasicUser                   `json:"user,omitempty" description:"User information (if authenticated)"`
	Files              []FileUpload                 `json:"files,omitempty" description:"Uploaded files"`
	Metadata           map[string]interface{}       `json:"metadata,omitempty" description:"Additional metadata"`
	ProcessingStatus   string                       `json:"processing_status" example:"processed" enums:"pending,processing,processed,failed" description:"Backend processing status"`
	IntegrationResults map[string]IntegrationResult `json:"integration_results,omitempty" description:"Third-party integration results"`
}

// ResponseValue represents a field response value with metadata
type ResponseValue struct {
	Value        interface{}            `json:"value" description:"Field value"`
	DisplayValue string                 `json:"display_value,omitempty" example:"Option 1" description:"Human-readable value"`
	Type         string                 `json:"type" example:"text" description:"Value type"`
	FieldID      string                 `json:"field_id" example:"field_1" description:"Field identifier"`
	FieldLabel   string                 `json:"field_label" example:"Full Name" description:"Field label"`
	Files        []FileUpload           `json:"files,omitempty" description:"Uploaded files for this field"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" description:"Field-specific metadata"`
	Validation   *FieldValidationResult `json:"validation,omitempty" description:"Field validation result"`
	TimingInfo   *FieldTiming           `json:"timing_info,omitempty" description:"Field timing information"`
}

// FileUpload represents an uploaded file
type FileUpload struct {
	ID           string                 `json:"id" example:"file_123456789" description:"File identifier"`
	OriginalName string                 `json:"original_name" example:"document.pdf" description:"Original filename"`
	FileName     string                 `json:"file_name" example:"abc123_document.pdf" description:"Stored filename"`
	ContentType  string                 `json:"content_type" example:"application/pdf" description:"File MIME type"`
	Size         int64                  `json:"size" example:"1048576" description:"File size in bytes"`
	URL          string                 `json:"url" example:"https://storage.example.com/files/abc123" description:"File access URL"`
	ThumbnailURL string                 `json:"thumbnail_url,omitempty" description:"Thumbnail URL (for images)"`
	Status       string                 `json:"status" example:"uploaded" enums:"uploading,uploaded,processing,virus_scan,failed" description:"File status"`
	UploadedAt   time.Time              `json:"uploaded_at" example:"2025-09-06T12:00:00Z" description:"Upload timestamp"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty" example:"2025-09-13T12:00:00Z" description:"File expiration time"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" description:"File metadata"`
	VirusScan    *VirusScanResult       `json:"virus_scan,omitempty" description:"Virus scan result"`
}

// VirusScanResult represents virus scan results
type VirusScanResult struct {
	Scanned   bool      `json:"scanned" example:"true" description:"Whether file was scanned"`
	Clean     bool      `json:"clean" example:"true" description:"Whether file is clean"`
	Threats   []string  `json:"threats,omitempty" description:"Detected threats"`
	Scanner   string    `json:"scanner" example:"clamav" description:"Scanner used"`
	ScannedAt time.Time `json:"scanned_at" example:"2025-09-06T12:01:00Z" description:"Scan timestamp"`
}

// ResponseScore represents response scoring information
type ResponseScore struct {
	TotalScore    float64            `json:"total_score" example:"85.5" description:"Total response score"`
	MaxScore      float64            `json:"max_score" example:"100.0" description:"Maximum possible score"`
	Percentage    float64            `json:"percentage" example:"85.5" description:"Score percentage"`
	Grade         string             `json:"grade,omitempty" example:"B+" description:"Letter grade"`
	FieldScores   map[string]float64 `json:"field_scores,omitempty" description:"Individual field scores"`
	ScoringMethod string             `json:"scoring_method" example:"weighted" description:"Scoring method used"`
	PassingScore  float64            `json:"passing_score,omitempty" example:"70.0" description:"Minimum passing score"`
	Passed        bool               `json:"passed" example:"true" description:"Whether response passed"`
}

// ResponseValidation represents response validation results
type ResponseValidation struct {
	IsValid      bool                             `json:"is_valid" example:"true" description:"Whether response is valid"`
	ErrorCount   int                              `json:"error_count" example:"0" description:"Number of validation errors"`
	WarningCount int                              `json:"warning_count" example:"1" description:"Number of validation warnings"`
	FieldResults map[string]FieldValidationResult `json:"field_results" description:"Per-field validation results"`
	GlobalErrors []ValidationError                `json:"global_errors,omitempty" description:"Global validation errors"`
	ValidatedAt  time.Time                        `json:"validated_at" example:"2025-09-06T12:05:00Z" description:"Validation timestamp"`
}

// FieldValidationResult represents field-specific validation results
type FieldValidationResult struct {
	IsValid   bool              `json:"is_valid" example:"true" description:"Whether field is valid"`
	Errors    []ValidationError `json:"errors,omitempty" description:"Field validation errors"`
	Warnings  []ValidationError `json:"warnings,omitempty" description:"Field validation warnings"`
	Processed bool              `json:"processed" example:"true" description:"Whether field was processed"`
}

// ResponseTiming represents response timing information
type ResponseTiming struct {
	TotalTime        int                    `json:"total_time" example:"300" description:"Total time spent (seconds)"`
	ActiveTime       int                    `json:"active_time" example:"180" description:"Active time spent (seconds)"`
	IdleTime         int                    `json:"idle_time" example:"120" description:"Idle time (seconds)"`
	FieldTiming      map[string]FieldTiming `json:"field_timing" description:"Per-field timing"`
	PageTiming       map[string]PageTiming  `json:"page_timing,omitempty" description:"Per-page timing (for multi-page forms)"`
	StartedAt        time.Time              `json:"started_at" example:"2025-09-06T12:00:00Z" description:"Start timestamp"`
	FirstInteraction *time.Time             `json:"first_interaction,omitempty" example:"2025-09-06T12:00:10Z" description:"First interaction timestamp"`
	LastInteraction  *time.Time             `json:"last_interaction,omitempty" example:"2025-09-06T12:04:50Z" description:"Last interaction timestamp"`
	SubmittedAt      *time.Time             `json:"submitted_at,omitempty" example:"2025-09-06T12:05:00Z" description:"Submission timestamp"`
}

// FieldTiming represents timing for a specific field
type FieldTiming struct {
	TimeSpent       int        `json:"time_spent" example:"30" description:"Time spent on field (seconds)"`
	FirstFocus      *time.Time `json:"first_focus,omitempty" example:"2025-09-06T12:01:00Z" description:"First focus timestamp"`
	LastInteraction *time.Time `json:"last_interaction,omitempty" example:"2025-09-06T12:01:30Z" description:"Last interaction timestamp"`
	FocusCount      int        `json:"focus_count" example:"2" description:"Number of times field was focused"`
	ChangeCount     int        `json:"change_count" example:"5" description:"Number of value changes"`
	IdleTime        int        `json:"idle_time" example:"10" description:"Idle time on field (seconds)"`
}

// PageTiming represents timing for a form page
type PageTiming struct {
	TimeSpent int        `json:"time_spent" example:"90" description:"Time spent on page (seconds)"`
	FirstView time.Time  `json:"first_view" example:"2025-09-06T12:01:00Z" description:"First view timestamp"`
	LastView  *time.Time `json:"last_view,omitempty" example:"2025-09-06T12:02:30Z" description:"Last view timestamp"`
	ViewCount int        `json:"view_count" example:"1" description:"Number of page views"`
	ExitCount int        `json:"exit_count" example:"0" description:"Number of exits from page"`
}

// IntegrationResult represents third-party integration result
type IntegrationResult struct {
	Provider    string                 `json:"provider" example:"salesforce" description:"Integration provider"`
	Status      string                 `json:"status" example:"success" enums:"success,failed,pending" description:"Integration status"`
	Message     string                 `json:"message,omitempty" example:"Record created successfully" description:"Result message"`
	ExternalID  string                 `json:"external_id,omitempty" example:"sf_123456" description:"External system ID"`
	ProcessedAt time.Time              `json:"processed_at" example:"2025-09-06T12:06:00Z" description:"Processing timestamp"`
	Data        map[string]interface{} `json:"data,omitempty" description:"Integration-specific data"`
	Error       string                 `json:"error,omitempty" example:"API rate limit exceeded" description:"Error message if failed"`
}

// ResponseListResponse represents a list of responses with pagination
type ResponseListResponse struct {
	Responses  []ResponseSummary     `json:"responses" description:"List of responses"`
	Pagination PaginationResponse    `json:"pagination" description:"Pagination information"`
	Filters    ResponseListFilters   `json:"filters" description:"Applied filters"`
	Sorting    ResponseListSorting   `json:"sorting" description:"Applied sorting"`
	Analytics  ResponseListAnalytics `json:"analytics" description:"Summary analytics"`
}

// ResponseSummary represents a summary of a response for list views
type ResponseSummary struct {
	ID           string                 `json:"id" example:"response_123456789" description:"Response ID"`
	FormID       string                 `json:"form_id" example:"form_123456789" description:"Form ID"`
	FormTitle    string                 `json:"form_title" example:"Customer Survey" description:"Form title"`
	Status       string                 `json:"status" example:"completed" description:"Response status"`
	IsComplete   bool                   `json:"is_complete" example:"true" description:"Completion status"`
	IsAnonymous  bool                   `json:"is_anonymous" example:"true" description:"Anonymous status"`
	Score        *float64               `json:"score,omitempty" example:"85.5" description:"Response score"`
	TimeSpent    int                    `json:"time_spent" example:"300" description:"Time spent (seconds)"`
	CreatedAt    time.Time              `json:"created_at" example:"2025-09-06T12:00:00Z" description:"Creation time"`
	SubmittedAt  *time.Time             `json:"submitted_at,omitempty" example:"2025-09-06T12:05:00Z" description:"Submission time"`
	User         *BasicUser             `json:"user,omitempty" description:"User info (if not anonymous)"`
	KeyResponses map[string]interface{} `json:"key_responses,omitempty" description:"Key response values"`
	Location     *LocationInfo          `json:"location,omitempty" description:"Geographic location"`
	DeviceType   string                 `json:"device_type,omitempty" example:"desktop" description:"Device type"`
}

// ResponseListFilters represents filters for response listing
type ResponseListFilters struct {
	FormIDs     []string    `json:"form_ids,omitempty" description:"Filter by form IDs"`
	Status      []string    `json:"status,omitempty" example:"completed,draft" description:"Filter by status"`
	IsComplete  *bool       `json:"is_complete,omitempty" example:"true" description:"Filter by completion"`
	IsAnonymous *bool       `json:"is_anonymous,omitempty" example:"false" description:"Filter by anonymous status"`
	UserIDs     []string    `json:"user_ids,omitempty" description:"Filter by user IDs"`
	DateRange   DateRange   `json:"date_range,omitempty" description:"Filter by date range"`
	ScoreRange  *ScoreRange `json:"score_range,omitempty" description:"Filter by score range"`
	TimeRange   *TimeRange  `json:"time_range,omitempty" description:"Filter by time spent range"`
	Countries   []string    `json:"countries,omitempty" description:"Filter by countries"`
	DeviceTypes []string    `json:"device_types,omitempty" description:"Filter by device types"`
	Languages   []string    `json:"languages,omitempty" description:"Filter by languages"`
	HasFiles    *bool       `json:"has_files,omitempty" example:"true" description:"Filter by file uploads"`
}

// ScoreRange represents a score range filter
type ScoreRange struct {
	Min *float64 `json:"min,omitempty" example:"70.0" description:"Minimum score"`
	Max *float64 `json:"max,omitempty" example:"100.0" description:"Maximum score"`
}

// TimeRange represents a time range filter
type TimeRange struct {
	Min *int `json:"min,omitempty" example:"60" description:"Minimum time in seconds"`
	Max *int `json:"max,omitempty" example:"600" description:"Maximum time in seconds"`
}

// ResponseListSorting represents sorting options for response listing
type ResponseListSorting struct {
	Field string `json:"field" example:"created_at" enums:"created_at,submitted_at,time_spent,score" description:"Sort field"`
	Order string `json:"order" example:"desc" enums:"asc,desc" description:"Sort order"`
}

// ResponseListAnalytics represents summary analytics for response list
type ResponseListAnalytics struct {
	TotalResponses     int             `json:"total_responses" example:"150" description:"Total responses"`
	CompletedResponses int             `json:"completed_responses" example:"120" description:"Completed responses"`
	DraftResponses     int             `json:"draft_responses" example:"25" description:"Draft responses"`
	AbandonedResponses int             `json:"abandoned_responses" example:"5" description:"Abandoned responses"`
	CompletionRate     float64         `json:"completion_rate" example:"0.80" description:"Overall completion rate"`
	AverageScore       *float64        `json:"average_score,omitempty" example:"82.3" description:"Average score"`
	AverageTime        int             `json:"average_time" example:"240" description:"Average completion time"`
	DeviceBreakdown    DeviceBreakdown `json:"device_breakdown" description:"Device usage breakdown"`
	TopCountries       []LocationData  `json:"top_countries" description:"Top countries"`
}

// ExportRequest represents a data export request
type ExportRequest struct {
	FormIDs      []string            `json:"form_ids" binding:"required" description:"Form IDs to export"`
	Format       string              `json:"format" binding:"required" example:"csv" enums:"csv,xlsx,json,pdf" description:"Export format"`
	Filters      ResponseListFilters `json:"filters,omitempty" description:"Export filters"`
	Fields       []string            `json:"fields,omitempty" description:"Specific fields to include"`
	IncludeFiles bool                `json:"include_files" example:"false" description:"Include file attachments"`
	Options      ExportOptions       `json:"options" description:"Export options"`
}

// ExportOptions represents export configuration options
type ExportOptions struct {
	FileName       string            `json:"file_name,omitempty" example:"responses_export" description:"Custom filename"`
	IncludeHeaders bool              `json:"include_headers" example:"true" description:"Include column headers"`
	DateFormat     string            `json:"date_format" example:"2006-01-02 15:04:05" description:"Date format"`
	Encoding       string            `json:"encoding" example:"UTF-8" description:"File encoding"`
	Delimiter      string            `json:"delimiter" example:"," description:"CSV delimiter"`
	SheetName      string            `json:"sheet_name,omitempty" example:"Responses" description:"Excel sheet name"`
	Password       string            `json:"password,omitempty" description:"Password protect file"`
	Compression    bool              `json:"compression" example:"false" description:"Enable compression"`
	CustomFields   map[string]string `json:"custom_fields,omitempty" description:"Custom field mappings"`
}

// ExportJobResponse represents an export job response
type ExportJobResponse struct {
	JobID          string     `json:"job_id" example:"export_123456789" description:"Export job ID"`
	Status         string     `json:"status" example:"pending" enums:"pending,processing,completed,failed" description:"Job status"`
	Format         string     `json:"format" example:"csv" description:"Export format"`
	Progress       float64    `json:"progress" example:"0.0" description:"Progress percentage"`
	RecordCount    int        `json:"record_count,omitempty" example:"150" description:"Number of records to export"`
	ProcessedCount int        `json:"processed_count" example:"0" description:"Number of records processed"`
	FileSize       int64      `json:"file_size,omitempty" example:"2048" description:"File size in bytes"`
	DownloadURL    string     `json:"download_url,omitempty" description:"Download URL (when completed)"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty" example:"2025-09-13T12:00:00Z" description:"Download expiration"`
	CreatedAt      time.Time  `json:"created_at" example:"2025-09-06T12:00:00Z" description:"Job creation time"`
	StartedAt      *time.Time `json:"started_at,omitempty" example:"2025-09-06T12:00:05Z" description:"Processing start time"`
	CompletedAt    *time.Time `json:"completed_at,omitempty" example:"2025-09-06T12:01:00Z" description:"Completion time"`
	Error          string     `json:"error,omitempty" description:"Error message if failed"`
}
