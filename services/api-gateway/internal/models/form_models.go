package models

import (
	"time"
)

// FormCreationRequest represents a comprehensive form creation request
type FormCreationRequest struct {
	Title               string                `json:"title" binding:"required,min=3,max=200" example:"Customer Feedback Survey" description:"Form title"`
	Description         string                `json:"description,omitempty" example:"Please provide your valuable feedback" description:"Form description"`
	Category            string                `json:"category,omitempty" example:"feedback" description:"Form category"`
	Tags                []string              `json:"tags,omitempty" example:"feedback,customer,survey" description:"Form tags for organization"`
	IsPublic            bool                  `json:"is_public" example:"true" description:"Whether form is publicly accessible"`
	AllowAnonymous      bool                  `json:"allow_anonymous" example:"true" description:"Allow anonymous submissions"`
	RequireAuth         bool                  `json:"require_auth" example:"false" description:"Require authentication to submit"`
	MultipleSubmissions bool                  `json:"multiple_submissions" example:"false" description:"Allow multiple submissions per user"`
	SubmissionLimit     int                   `json:"submission_limit,omitempty" example:"1000" description:"Maximum number of submissions"`
	ExpiresAt           *time.Time            `json:"expires_at,omitempty" example:"2025-12-31T23:59:59Z" description:"Form expiration date"`
	Fields              []FormFieldDefinition `json:"fields" binding:"required,min=1" description:"Form fields"`
	Settings            FormSettings          `json:"settings" description:"Form settings"`
	Notifications       NotificationConfig    `json:"notifications" description:"Notification configuration"`
	Integrations        IntegrationConfig     `json:"integrations" description:"Integration settings"`
	Branding            BrandingConfig        `json:"branding" description:"Branding configuration"`
}

// FormUpdateRequest represents a form update request with comprehensive configuration
type FormUpdateRequest struct {
	Title       *string               `json:"title,omitempty" example:"Updated Contact Form" description:"Form title (optional)"`
	Description *string               `json:"description,omitempty" example:"Please fill out your contact information" description:"Form description (optional)"`
	Fields      []FormFieldDefinition `json:"fields,omitempty" description:"Form field definitions (optional)"`
	Settings    *FormSettings         `json:"settings,omitempty" description:"Form configuration settings (optional)"`
	Status      *string               `json:"status,omitempty" example:"published" enums:"draft,published,archived" description:"Form status (optional)"`
	Tags        []string              `json:"tags,omitempty" example:"contact,support,general" description:"Form tags for categorization (optional)"`
	Category    *string               `json:"category,omitempty" example:"support" description:"Form category (optional)"`
}

// FormFieldDefinition represents a comprehensive form field definition
type FormFieldDefinition struct {
	ID           string                 `json:"id" example:"field_1" description:"Unique field identifier"`
	Type         string                 `json:"type" binding:"required" example:"text" enums:"text,email,number,tel,url,textarea,select,radio,checkbox,date,time,datetime,file,rating,matrix,slider" description:"Field type"`
	Label        string                 `json:"label" binding:"required" example:"Full Name" description:"Field label"`
	Description  string                 `json:"description,omitempty" example:"Enter your full legal name" description:"Field description"`
	Placeholder  string                 `json:"placeholder,omitempty" example:"John Doe" description:"Field placeholder text"`
	Required     bool                   `json:"required" example:"true" description:"Whether field is required"`
	Order        int                    `json:"order" example:"1" description:"Field display order"`
	DefaultValue string                 `json:"default_value,omitempty" example:"" description:"Default field value"`
	Options      []FieldOption          `json:"options,omitempty" description:"Field options for select/radio/checkbox fields"`
	Validation   FieldValidation        `json:"validation" description:"Field validation rules"`
	Conditional  *ConditionalLogic      `json:"conditional,omitempty" description:"Conditional display logic"`
	Styling      FieldStyling           `json:"styling" description:"Field styling options"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" description:"Additional field metadata"`
}

// FieldOption represents an option for select/radio/checkbox fields
type FieldOption struct {
	Value    string `json:"value" example:"option1" description:"Option value"`
	Label    string `json:"label" example:"Option 1" description:"Option display label"`
	Selected bool   `json:"selected" example:"false" description:"Whether option is selected by default"`
	Disabled bool   `json:"disabled" example:"false" description:"Whether option is disabled"`
	Color    string `json:"color,omitempty" example:"#007bff" description:"Option color"`
	Icon     string `json:"icon,omitempty" example:"star" description:"Option icon"`
}

// FieldValidation represents field validation rules
type FieldValidation struct {
	MinLength    *int                   `json:"min_length,omitempty" example:"2" description:"Minimum character length"`
	MaxLength    *int                   `json:"max_length,omitempty" example:"100" description:"Maximum character length"`
	MinValue     *float64               `json:"min_value,omitempty" example:"0" description:"Minimum numeric value"`
	MaxValue     *float64               `json:"max_value,omitempty" example:"100" description:"Maximum numeric value"`
	Pattern      string                 `json:"pattern,omitempty" example:"^[A-Za-z]+$" description:"Regex validation pattern"`
	AllowedTypes []string               `json:"allowed_types,omitempty" example:"image/jpeg,image/png" description:"Allowed file types"`
	MaxFileSize  *int64                 `json:"max_file_size,omitempty" example:"5242880" description:"Maximum file size in bytes"`
	MinFiles     *int                   `json:"min_files,omitempty" example:"1" description:"Minimum number of files"`
	MaxFiles     *int                   `json:"max_files,omitempty" example:"5" description:"Maximum number of files"`
	Custom       map[string]interface{} `json:"custom,omitempty" description:"Custom validation rules"`
}

// ConditionalLogic represents conditional field display logic
type ConditionalLogic struct {
	ShowIf   []ConditionRule `json:"show_if,omitempty" description:"Conditions to show field"`
	HideIf   []ConditionRule `json:"hide_if,omitempty" description:"Conditions to hide field"`
	Operator string          `json:"operator" example:"AND" enums:"AND,OR" description:"Logical operator for multiple conditions"`
}

// ConditionRule represents a single condition rule
type ConditionRule struct {
	FieldID  string      `json:"field_id" example:"field_2" description:"Field ID to check"`
	Operator string      `json:"operator" example:"equals" enums:"equals,not_equals,contains,not_contains,greater_than,less_than,is_empty,is_not_empty" description:"Condition operator"`
	Value    interface{} `json:"value" example:"yes" description:"Value to compare against"`
}

// FieldStyling represents field styling options
type FieldStyling struct {
	Width           string `json:"width,omitempty" example:"100%" description:"Field width"`
	Height          string `json:"height,omitempty" example:"auto" description:"Field height"`
	BackgroundColor string `json:"background_color,omitempty" example:"#ffffff" description:"Background color"`
	BorderColor     string `json:"border_color,omitempty" example:"#cccccc" description:"Border color"`
	TextColor       string `json:"text_color,omitempty" example:"#333333" description:"Text color"`
	FontSize        string `json:"font_size,omitempty" example:"14px" description:"Font size"`
	FontWeight      string `json:"font_weight,omitempty" example:"normal" description:"Font weight"`
	BorderRadius    string `json:"border_radius,omitempty" example:"4px" description:"Border radius"`
	Margin          string `json:"margin,omitempty" example:"10px" description:"Field margin"`
	Padding         string `json:"padding,omitempty" example:"8px" description:"Field padding"`
	CustomCSS       string `json:"custom_css,omitempty" description:"Custom CSS styles"`
}

// FormSettings represents comprehensive form settings
type FormSettings struct {
	Theme              string            `json:"theme" example:"default" description:"Form theme"`
	Layout             string            `json:"layout" example:"vertical" enums:"vertical,horizontal,grid" description:"Form layout"`
	ProgressBar        bool              `json:"progress_bar" example:"true" description:"Show progress bar"`
	PageBreaks         bool              `json:"page_breaks" example:"false" description:"Enable page breaks"`
	AutoSave           bool              `json:"auto_save" example:"true" description:"Auto-save responses"`
	SaveDraftButton    bool              `json:"save_draft_button" example:"true" description:"Show save draft button"`
	RequiredFieldsNote string            `json:"required_fields_note" example:"* Required fields" description:"Required fields note"`
	SubmitButtonText   string            `json:"submit_button_text" example:"Submit" description:"Submit button text"`
	SuccessMessage     string            `json:"success_message" example:"Thank you for your submission!" description:"Success message"`
	SuccessRedirectURL string            `json:"success_redirect_url,omitempty" description:"URL to redirect after successful submission"`
	ErrorMessage       string            `json:"error_message" example:"Please correct the errors below" description:"Error message"`
	Captcha            CaptchaSettings   `json:"captcha" description:"Captcha settings"`
	RateLimiting       RateLimitSettings `json:"rate_limiting" description:"Rate limiting settings"`
	CustomCSS          string            `json:"custom_css,omitempty" description:"Custom CSS for the form"`
	CustomJS           string            `json:"custom_js,omitempty" description:"Custom JavaScript for the form"`
	Analytics          AnalyticsSettings `json:"analytics" description:"Analytics settings"`
}

// CaptchaSettings represents captcha configuration
type CaptchaSettings struct {
	Enabled   bool    `json:"enabled" example:"false" description:"Enable captcha"`
	Type      string  `json:"type" example:"recaptcha" enums:"recaptcha,hcaptcha,custom" description:"Captcha type"`
	SiteKey   string  `json:"site_key,omitempty" description:"Captcha site key"`
	Threshold float64 `json:"threshold" example:"0.5" description:"Captcha score threshold"`
}

// RateLimitSettings represents rate limiting configuration
type RateLimitSettings struct {
	Enabled            bool `json:"enabled" example:"true" description:"Enable rate limiting"`
	MaxSubmissions     int  `json:"max_submissions" example:"5" description:"Maximum submissions per window"`
	WindowMinutes      int  `json:"window_minutes" example:"60" description:"Rate limit window in minutes"`
	BlockDurationHours int  `json:"block_duration_hours" example:"24" description:"Block duration in hours"`
}

// AnalyticsSettings represents analytics configuration
type AnalyticsSettings struct {
	Enabled         bool     `json:"enabled" example:"true" description:"Enable analytics"`
	TrackViews      bool     `json:"track_views" example:"true" description:"Track form views"`
	TrackDropoffs   bool     `json:"track_dropoffs" example:"true" description:"Track form dropoffs"`
	TrackTime       bool     `json:"track_time" example:"true" description:"Track completion time"`
	GoogleAnalytics string   `json:"google_analytics,omitempty" description:"Google Analytics ID"`
	CustomEvents    []string `json:"custom_events,omitempty" description:"Custom analytics events"`
}

// NotificationConfig represents notification settings
type NotificationConfig struct {
	Enabled          bool            `json:"enabled" example:"true" description:"Enable notifications"`
	AdminEmails      []string        `json:"admin_emails,omitempty" description:"Admin email addresses"`
	UserConfirmation bool            `json:"user_confirmation" example:"true" description:"Send confirmation to user"`
	EmailTemplate    string          `json:"email_template,omitempty" description:"Email template ID"`
	SlackWebhook     string          `json:"slack_webhook,omitempty" description:"Slack webhook URL"`
	DiscordWebhook   string          `json:"discord_webhook,omitempty" description:"Discord webhook URL"`
	CustomWebhooks   []WebhookConfig `json:"custom_webhooks,omitempty" description:"Custom webhook configurations"`
}

// WebhookConfig represents webhook configuration
type WebhookConfig struct {
	URL        string            `json:"url" example:"https://api.example.com/webhook" description:"Webhook URL"`
	Method     string            `json:"method" example:"POST" enums:"POST,PUT,PATCH" description:"HTTP method"`
	Headers    map[string]string `json:"headers,omitempty" description:"Custom headers"`
	Secret     string            `json:"secret,omitempty" description:"Webhook secret for verification"`
	Events     []string          `json:"events" example:"submission,update" description:"Events to trigger webhook"`
	Retries    int               `json:"retries" example:"3" description:"Number of retry attempts"`
	RetryDelay int               `json:"retry_delay" example:"300" description:"Retry delay in seconds"`
}

// IntegrationConfig represents third-party integrations
type IntegrationConfig struct {
	Enabled     bool                   `json:"enabled" example:"false" description:"Enable integrations"`
	Email       EmailIntegration       `json:"email" description:"Email integration settings"`
	CRM         CRMIntegration         `json:"crm" description:"CRM integration settings"`
	Spreadsheet SpreadsheetIntegration `json:"spreadsheet" description:"Spreadsheet integration settings"`
	Database    DatabaseIntegration    `json:"database" description:"Database integration settings"`
	Zapier      ZapierIntegration      `json:"zapier" description:"Zapier integration settings"`
	Custom      map[string]interface{} `json:"custom,omitempty" description:"Custom integrations"`
}

// EmailIntegration represents email service integration
type EmailIntegration struct {
	Provider  string            `json:"provider" example:"sendgrid" enums:"sendgrid,mailgun,smtp" description:"Email provider"`
	APIKey    string            `json:"api_key,omitempty" description:"API key"`
	FromEmail string            `json:"from_email" example:"noreply@example.com" description:"From email address"`
	FromName  string            `json:"from_name" example:"Form System" description:"From name"`
	Templates map[string]string `json:"templates,omitempty" description:"Email templates"`
}

// CRMIntegration represents CRM integration
type CRMIntegration struct {
	Provider string            `json:"provider" example:"salesforce" enums:"salesforce,hubspot,pipedrive" description:"CRM provider"`
	APIKey   string            `json:"api_key,omitempty" description:"API key"`
	Endpoint string            `json:"endpoint,omitempty" description:"API endpoint"`
	Mapping  map[string]string `json:"mapping,omitempty" description:"Field mapping"`
}

// SpreadsheetIntegration represents spreadsheet integration
type SpreadsheetIntegration struct {
	Provider      string `json:"provider" example:"google_sheets" enums:"google_sheets,excel_online" description:"Spreadsheet provider"`
	SpreadsheetID string `json:"spreadsheet_id,omitempty" description:"Spreadsheet ID"`
	SheetName     string `json:"sheet_name,omitempty" description:"Sheet name"`
	APIKey        string `json:"api_key,omitempty" description:"API key"`
}

// DatabaseIntegration represents database integration
type DatabaseIntegration struct {
	Type     string            `json:"type" example:"mysql" enums:"mysql,postgresql,mongodb" description:"Database type"`
	Host     string            `json:"host,omitempty" description:"Database host"`
	Database string            `json:"database,omitempty" description:"Database name"`
	Table    string            `json:"table,omitempty" description:"Table name"`
	Mapping  map[string]string `json:"mapping,omitempty" description:"Field mapping"`
}

// ZapierIntegration represents Zapier integration
type ZapierIntegration struct {
	WebhookURL string   `json:"webhook_url,omitempty" description:"Zapier webhook URL"`
	Events     []string `json:"events,omitempty" description:"Events to send to Zapier"`
}

// BrandingConfig represents form branding configuration
type BrandingConfig struct {
	Logo            string `json:"logo,omitempty" description:"Logo URL"`
	FaviconURL      string `json:"favicon_url,omitempty" description:"Favicon URL"`
	PrimaryColor    string `json:"primary_color" example:"#007bff" description:"Primary brand color"`
	SecondaryColor  string `json:"secondary_color" example:"#6c757d" description:"Secondary brand color"`
	AccentColor     string `json:"accent_color" example:"#28a745" description:"Accent color"`
	BackgroundColor string `json:"background_color" example:"#ffffff" description:"Background color"`
	TextColor       string `json:"text_color" example:"#333333" description:"Text color"`
	FontFamily      string `json:"font_family" example:"Arial, sans-serif" description:"Font family"`
	BorderRadius    string `json:"border_radius" example:"4px" description:"Border radius"`
	CustomCSS       string `json:"custom_css,omitempty" description:"Custom CSS"`
	CompanyName     string `json:"company_name,omitempty" example:"Acme Corp" description:"Company name"`
	CompanyURL      string `json:"company_url,omitempty" example:"https://acme.com" description:"Company URL"`
}

// FormResponse represents a comprehensive form
type FormResponse struct {
	ID                  string                `json:"id" example:"form_123456789" description:"Unique form identifier"`
	Title               string                `json:"title" example:"Customer Feedback Survey" description:"Form title"`
	Description         string                `json:"description,omitempty" example:"Please provide your valuable feedback" description:"Form description"`
	Category            string                `json:"category,omitempty" example:"feedback" description:"Form category"`
	Tags                []string              `json:"tags,omitempty" example:"feedback,customer,survey" description:"Form tags"`
	Status              string                `json:"status" example:"published" enums:"draft,published,archived,expired" description:"Form status"`
	IsPublic            bool                  `json:"is_public" example:"true" description:"Whether form is publicly accessible"`
	AllowAnonymous      bool                  `json:"allow_anonymous" example:"true" description:"Allow anonymous submissions"`
	RequireAuth         bool                  `json:"require_auth" example:"false" description:"Require authentication"`
	MultipleSubmissions bool                  `json:"multiple_submissions" example:"false" description:"Allow multiple submissions"`
	SubmissionLimit     int                   `json:"submission_limit,omitempty" example:"1000" description:"Maximum submissions"`
	SubmissionCount     int                   `json:"submission_count" example:"42" description:"Current submission count"`
	ViewCount           int                   `json:"view_count" example:"150" description:"Form view count"`
	CompletionRate      float64               `json:"completion_rate" example:"0.85" description:"Form completion rate"`
	AverageTime         int                   `json:"average_time" example:"180" description:"Average completion time in seconds"`
	CreatedAt           time.Time             `json:"created_at" example:"2025-01-01T00:00:00Z" description:"Creation timestamp"`
	UpdatedAt           time.Time             `json:"updated_at" example:"2025-09-06T12:00:00Z" description:"Last update timestamp"`
	PublishedAt         *time.Time            `json:"published_at,omitempty" example:"2025-01-01T00:00:00Z" description:"Publication timestamp"`
	ExpiresAt           *time.Time            `json:"expires_at,omitempty" example:"2025-12-31T23:59:59Z" description:"Expiration timestamp"`
	LastSubmissionAt    *time.Time            `json:"last_submission_at,omitempty" example:"2025-09-06T11:30:00Z" description:"Last submission timestamp"`
	Fields              []FormFieldDefinition `json:"fields" description:"Form fields"`
	Settings            FormSettings          `json:"settings" description:"Form settings"`
	Notifications       NotificationConfig    `json:"notifications" description:"Notification config"`
	Integrations        IntegrationConfig     `json:"integrations" description:"Integration config"`
	Branding            BrandingConfig        `json:"branding" description:"Branding config"`
	Owner               DetailedUser          `json:"owner" description:"Form owner"`
	Collaborators       []FormCollaborator    `json:"collaborators,omitempty" description:"Form collaborators"`
	Permissions         FormPermissions       `json:"permissions" description:"Form permissions"`
	SEO                 SEOConfig             `json:"seo" description:"SEO configuration"`
	Analytics           FormAnalytics         `json:"analytics" description:"Form analytics"`
	QRCode              string                `json:"qr_code,omitempty" description:"QR code URL"`
	ShortURL            string                `json:"short_url,omitempty" description:"Short URL"`
	EmbedCode           string                `json:"embed_code,omitempty" description:"Embed code"`
}

// FormCollaborator represents a form collaborator
type FormCollaborator struct {
	UserID      string    `json:"user_id" example:"user_123" description:"User ID"`
	Email       string    `json:"email" example:"collaborator@example.com" description:"Collaborator email"`
	Role        string    `json:"role" example:"editor" enums:"viewer,editor,admin" description:"Collaborator role"`
	Permissions []string  `json:"permissions" example:"read,write" description:"Specific permissions"`
	AddedAt     time.Time `json:"added_at" example:"2025-01-01T00:00:00Z" description:"Added timestamp"`
	AddedBy     string    `json:"added_by" example:"user_456" description:"Added by user ID"`
}

// FormPermissions represents form permissions
type FormPermissions struct {
	CanView                bool `json:"can_view" example:"true" description:"Can view form"`
	CanEdit                bool `json:"can_edit" example:"true" description:"Can edit form"`
	CanDelete              bool `json:"can_delete" example:"false" description:"Can delete form"`
	CanShare               bool `json:"can_share" example:"true" description:"Can share form"`
	CanAnalyze             bool `json:"can_analyze" example:"true" description:"Can view analytics"`
	CanExport              bool `json:"can_export" example:"true" description:"Can export data"`
	CanManageCollaborators bool `json:"can_manage_collaborators" example:"false" description:"Can manage collaborators"`
}

// SEOConfig represents SEO configuration
type SEOConfig struct {
	Title       string   `json:"title,omitempty" example:"Customer Feedback Survey" description:"SEO title"`
	Description string   `json:"description,omitempty" example:"Help us improve by providing feedback" description:"SEO description"`
	Keywords    []string `json:"keywords,omitempty" example:"feedback,survey,customer" description:"SEO keywords"`
	OGImage     string   `json:"og_image,omitempty" description:"Open Graph image URL"`
	Canonical   string   `json:"canonical,omitempty" description:"Canonical URL"`
	NoIndex     bool     `json:"no_index" example:"false" description:"Prevent indexing"`
}

// FormAnalytics represents form analytics data
type FormAnalytics struct {
	Views             int                       `json:"views" example:"150" description:"Total views"`
	UniqueViews       int                       `json:"unique_views" example:"120" description:"Unique views"`
	Submissions       int                       `json:"submissions" example:"42" description:"Total submissions"`
	CompletionRate    float64                   `json:"completion_rate" example:"0.85" description:"Completion rate"`
	AverageTime       int                       `json:"average_time" example:"180" description:"Average completion time (seconds)"`
	DropoffPoints     []DropoffPoint            `json:"dropoff_points" description:"Form dropoff points"`
	FieldAnalytics    map[string]FieldAnalytics `json:"field_analytics" description:"Per-field analytics"`
	DeviceBreakdown   DeviceBreakdown           `json:"device_breakdown" description:"Device usage breakdown"`
	LocationBreakdown []LocationData            `json:"location_breakdown" description:"Geographic breakdown"`
	TimeBreakdown     TimeBreakdown             `json:"time_breakdown" description:"Time-based analytics"`
	ReferrerBreakdown []ReferrerData            `json:"referrer_breakdown" description:"Traffic source breakdown"`
}

// DropoffPoint represents a form dropoff point
type DropoffPoint struct {
	FieldID     string  `json:"field_id" example:"field_3" description:"Field where dropoff occurred"`
	FieldLabel  string  `json:"field_label" example:"Phone Number" description:"Field label"`
	DropoffRate float64 `json:"dropoff_rate" example:"0.15" description:"Dropoff rate at this field"`
	Views       int     `json:"views" example:"100" description:"Views at this point"`
	Completions int     `json:"completions" example:"85" description:"Completions from this point"`
}

// FieldAnalytics represents analytics for a specific field
type FieldAnalytics struct {
	Views             int            `json:"views" example:"100" description:"Field views"`
	Completions       int            `json:"completions" example:"95" description:"Field completions"`
	CompletionRate    float64        `json:"completion_rate" example:"0.95" description:"Field completion rate"`
	AverageTime       int            `json:"average_time" example:"30" description:"Average time spent on field"`
	ValidationErrors  int            `json:"validation_errors" example:"5" description:"Validation errors"`
	ValueDistribution map[string]int `json:"value_distribution,omitempty" description:"Value distribution for choice fields"`
}

// DeviceBreakdown represents device usage analytics
type DeviceBreakdown struct {
	Desktop int `json:"desktop" example:"60" description:"Desktop users"`
	Mobile  int `json:"mobile" example:"35" description:"Mobile users"`
	Tablet  int `json:"tablet" example:"5" description:"Tablet users"`
}

// LocationData represents geographic analytics
type LocationData struct {
	Country     string  `json:"country" example:"United States" description:"Country name"`
	CountryCode string  `json:"country_code" example:"US" description:"Country code"`
	Count       int     `json:"count" example:"25" description:"User count"`
	Percentage  float64 `json:"percentage" example:"59.5" description:"Percentage of total"`
}

// TimeBreakdown represents time-based analytics
type TimeBreakdown struct {
	HourlyBreakdown []HourlyData `json:"hourly_breakdown" description:"Hourly usage data"`
	DailyBreakdown  []DailyData  `json:"daily_breakdown" description:"Daily usage data"`
	WeeklyBreakdown []WeeklyData `json:"weekly_breakdown" description:"Weekly usage data"`
}

// HourlyData represents hourly analytics
type HourlyData struct {
	Hour  int `json:"hour" example:"14" description:"Hour (0-23)"`
	Count int `json:"count" example:"8" description:"Submission count"`
}

// DailyData represents daily analytics
type DailyData struct {
	Date  string `json:"date" example:"2025-09-06" description:"Date"`
	Count int    `json:"count" example:"15" description:"Submission count"`
}

// WeeklyData represents weekly analytics
type WeeklyData struct {
	Week  string `json:"week" example:"2025-W36" description:"Week (ISO format)"`
	Count int    `json:"count" example:"50" description:"Submission count"`
}

// ReferrerData represents traffic source analytics
type ReferrerData struct {
	Source     string  `json:"source" example:"google.com" description:"Referrer source"`
	Count      int     `json:"count" example:"20" description:"Visitor count"`
	Percentage float64 `json:"percentage" example:"40.0" description:"Percentage of total"`
}

// FormListResponse represents a list of forms with pagination
type FormListResponse struct {
	Forms      []FormSummary      `json:"forms" description:"List of forms"`
	Pagination PaginationResponse `json:"pagination" description:"Pagination information"`
	Filters    FormListFilters    `json:"filters" description:"Applied filters"`
	Sorting    FormListSorting    `json:"sorting" description:"Applied sorting"`
}

// FormSummary represents a summary of a form for list views
type FormSummary struct {
	ID              string    `json:"id" example:"form_123456789" description:"Form ID"`
	Title           string    `json:"title" example:"Customer Feedback Survey" description:"Form title"`
	Description     string    `json:"description,omitempty" example:"Brief description" description:"Form description"`
	Status          string    `json:"status" example:"published" description:"Form status"`
	IsPublic        bool      `json:"is_public" example:"true" description:"Public accessibility"`
	SubmissionCount int       `json:"submission_count" example:"42" description:"Submission count"`
	ViewCount       int       `json:"view_count" example:"150" description:"View count"`
	CompletionRate  float64   `json:"completion_rate" example:"0.85" description:"Completion rate"`
	CreatedAt       time.Time `json:"created_at" example:"2025-01-01T00:00:00Z" description:"Creation date"`
	UpdatedAt       time.Time `json:"updated_at" example:"2025-09-06T12:00:00Z" description:"Last update"`
	Owner           BasicUser `json:"owner" description:"Form owner"`
	Tags            []string  `json:"tags,omitempty" description:"Form tags"`
	Category        string    `json:"category,omitempty" description:"Form category"`
}

// BasicUser represents basic user information
type BasicUser struct {
	ID     string `json:"id" example:"user_123" description:"User ID"`
	Name   string `json:"name" example:"John Doe" description:"User name"`
	Email  string `json:"email" example:"john@example.com" description:"User email"`
	Avatar string `json:"avatar,omitempty" description:"Avatar URL"`
}

// FormListFilters represents filters for form listing
type FormListFilters struct {
	Status    []string  `json:"status,omitempty" example:"published,draft" description:"Filter by status"`
	Category  []string  `json:"category,omitempty" example:"feedback,survey" description:"Filter by category"`
	Tags      []string  `json:"tags,omitempty" example:"customer,feedback" description:"Filter by tags"`
	Owner     string    `json:"owner,omitempty" example:"user_123" description:"Filter by owner"`
	IsPublic  *bool     `json:"is_public,omitempty" example:"true" description:"Filter by public status"`
	DateRange DateRange `json:"date_range,omitempty" description:"Filter by date range"`
}

// FormListSorting represents sorting options for form listing
type FormListSorting struct {
	Field string `json:"field" example:"created_at" enums:"created_at,updated_at,title,submission_count,view_count" description:"Sort field"`
	Order string `json:"order" example:"desc" enums:"asc,desc" description:"Sort order"`
}

// DateRange represents a date range filter
type DateRange struct {
	From *time.Time `json:"from,omitempty" example:"2025-01-01T00:00:00Z" description:"Start date"`
	To   *time.Time `json:"to,omitempty" example:"2025-12-31T23:59:59Z" description:"End date"`
}
