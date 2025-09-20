// Package domain defines the core business entities and contracts for the form service
// This layer contains the business rules and is independent of external concerns
package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Form represents the core form entity with business logic
type Form struct {
	ID          uuid.UUID              `json:"id"`
	UserID      uuid.UUID              `json:"user_id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Questions   []Question             `json:"questions"`
	Settings    FormSettings           `json:"settings"`
	Status      FormStatus             `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	PublishedAt *time.Time             `json:"published_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Question represents a form question with validation rules
type Question struct {
	ID          uuid.UUID              `json:"id"`
	Type        QuestionType           `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Required    bool                   `json:"required"`
	Options     []QuestionOption       `json:"options,omitempty"`
	Validation  ValidationRules        `json:"validation,omitempty"`
	Order       int                    `json:"order"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// QuestionOption represents an option for choice-based questions
type QuestionOption struct {
	ID    uuid.UUID `json:"id"`
	Label string    `json:"label"`
	Value string    `json:"value"`
	Order int       `json:"order"`
}

// FormSettings contains form configuration
type FormSettings struct {
	IsPublic              bool       `json:"is_public"`
	AllowAnonymous        bool       `json:"allow_anonymous"`
	RequireAuthentication bool       `json:"require_authentication"`
	EnableNotifications   bool       `json:"enable_notifications"`
	SubmissionLimit       *int       `json:"submission_limit,omitempty"`
	ExpiresAt             *time.Time `json:"expires_at,omitempty"`
	AllowedDomains        []string   `json:"allowed_domains,omitempty"`
}

// ValidationRules defines validation constraints for questions
type ValidationRules struct {
	MinLength   *int     `json:"min_length,omitempty"`
	MaxLength   *int     `json:"max_length,omitempty"`
	Pattern     *string  `json:"pattern,omitempty"`
	MinValue    *float64 `json:"min_value,omitempty"`
	MaxValue    *float64 `json:"max_value,omitempty"`
	CustomRules []string `json:"custom_rules,omitempty"`
}

// Enums
type FormStatus string
type QuestionType string

const (
	FormStatusDraft     FormStatus = "draft"
	FormStatusPublished FormStatus = "published"
	FormStatusClosed    FormStatus = "closed"
	FormStatusArchived  FormStatus = "archived"
)

const (
	QuestionTypeText           QuestionType = "text"
	QuestionTypeTextarea       QuestionType = "textarea"
	QuestionTypeNumber         QuestionType = "number"
	QuestionTypeEmail          QuestionType = "email"
	QuestionTypeURL            QuestionType = "url"
	QuestionTypeDate           QuestionType = "date"
	QuestionTypeTime           QuestionType = "time"
	QuestionTypeDateTime       QuestionType = "datetime"
	QuestionTypeSingleChoice   QuestionType = "single_choice"
	QuestionTypeMultipleChoice QuestionType = "multiple_choice"
	QuestionTypeDropdown       QuestionType = "dropdown"
	QuestionTypeBoolean        QuestionType = "boolean"
	QuestionTypeRating         QuestionType = "rating"
	QuestionTypeFile           QuestionType = "file"
)

// Business logic methods

// Validate ensures the form meets business requirements
func (f *Form) Validate() error {
	if f.Title == "" {
		return NewValidationError("title", "form title is required")
	}

	if len(f.Questions) == 0 {
		return NewValidationError("questions", "form must have at least one question")
	}

	// Validate questions
	for i, question := range f.Questions {
		if err := question.Validate(); err != nil {
			return NewValidationError("questions", "question %d: %v", i+1, err)
		}
	}

	return nil
}

// CanBePublished checks if form can be published
func (f *Form) CanBePublished() bool {
	return f.Status == FormStatusDraft && len(f.Questions) > 0 && f.Title != ""
}

// Publish changes form status to published
func (f *Form) Publish() error {
	if !f.CanBePublished() {
		return NewBusinessRuleError("form cannot be published in current state")
	}

	now := time.Now()
	f.Status = FormStatusPublished
	f.PublishedAt = &now
	f.UpdatedAt = now

	return nil
}

// Close marks the form as closed for new submissions
func (f *Form) Close() error {
	if f.Status != FormStatusPublished {
		return NewBusinessRuleError("only published forms can be closed")
	}

	f.Status = FormStatusClosed
	f.UpdatedAt = time.Now()

	return nil
}

// IsExpired checks if the form has expired
func (f *Form) IsExpired() bool {
	return f.Settings.ExpiresAt != nil && time.Now().After(*f.Settings.ExpiresAt)
}

// CanAcceptSubmissions checks if form can accept new submissions
func (f *Form) CanAcceptSubmissions() bool {
	return f.Status == FormStatusPublished && !f.IsExpired()
}

// Validate ensures the question meets business requirements
func (q *Question) Validate() error {
	if q.Title == "" {
		return NewValidationError("title", "question title is required")
	}

	switch q.Type {
	case QuestionTypeSingleChoice, QuestionTypeMultipleChoice, QuestionTypeDropdown:
		if len(q.Options) < 2 {
			return NewValidationError("options", "choice questions must have at least 2 options")
		}
	case QuestionTypeRating:
		if q.Validation.MinValue == nil || q.Validation.MaxValue == nil {
			return NewValidationError("validation", "rating questions must have min and max values")
		}
	}

	return nil
}

// Repository interfaces (ports) - following Dependency Inversion Principle

// FormRepository defines the contract for form data persistence
type FormRepository interface {
	// Create stores a new form
	Create(ctx context.Context, form *Form) error

	// GetByID retrieves a form by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*Form, error)

	// GetByUserID retrieves all forms for a user with pagination
	GetByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*Form, int64, error)

	// Update modifies an existing form
	Update(ctx context.Context, form *Form) error

	// Delete removes a form
	Delete(ctx context.Context, id uuid.UUID) error

	// GetPublishedForms retrieves published forms with filters
	GetPublishedForms(ctx context.Context, filters FormFilters) ([]*Form, error)

	// UpdateStatus changes form status
	UpdateStatus(ctx context.Context, id uuid.UUID, status FormStatus) error
}

// FormFilters represents query filters for forms
type FormFilters struct {
	UserID        *uuid.UUID  `json:"user_id,omitempty"`
	Status        *FormStatus `json:"status,omitempty"`
	Search        *string     `json:"search,omitempty"`
	Tags          []string    `json:"tags,omitempty"`
	CreatedAfter  *time.Time  `json:"created_after,omitempty"`
	CreatedBefore *time.Time  `json:"created_before,omitempty"`
	Offset        int         `json:"offset"`
	Limit         int         `json:"limit"`
}

// Service interfaces

// FormService defines the core form business operations
type FormService interface {
	// CreateForm creates a new form with validation
	CreateForm(ctx context.Context, userID uuid.UUID, req CreateFormRequest) (*Form, error)

	// GetForm retrieves a form with access control
	GetForm(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*Form, error)

	// UpdateForm modifies an existing form
	UpdateForm(ctx context.Context, id uuid.UUID, userID uuid.UUID, req UpdateFormRequest) (*Form, error)

	// DeleteForm removes a form
	DeleteForm(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// ListUserForms retrieves all forms for a user
	ListUserForms(ctx context.Context, userID uuid.UUID, filters FormFilters) ([]*Form, int64, error)

	// PublishForm publishes a draft form
	PublishForm(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// CloseForm closes a published form
	CloseForm(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// GetPublicForm retrieves a published form for public access
	GetPublicForm(ctx context.Context, id uuid.UUID) (*Form, error)
}

// Request/Response DTOs

// CreateFormRequest represents a form creation request
type CreateFormRequest struct {
	Title       string                 `json:"title" binding:"required,min=1,max=200"`
	Description string                 `json:"description" binding:"max=1000"`
	Questions   []QuestionRequest      `json:"questions" binding:"required,min=1"`
	Settings    FormSettings           `json:"settings"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateFormRequest represents a form update request
type UpdateFormRequest struct {
	Title       *string                `json:"title,omitempty" binding:"omitempty,min=1,max=200"`
	Description *string                `json:"description,omitempty" binding:"omitempty,max=1000"`
	Questions   []QuestionRequest      `json:"questions,omitempty"`
	Settings    *FormSettings          `json:"settings,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// QuestionRequest represents a question in a form request
type QuestionRequest struct {
	ID          *uuid.UUID              `json:"id,omitempty"`
	Type        QuestionType            `json:"type" binding:"required"`
	Title       string                  `json:"title" binding:"required,min=1,max=500"`
	Description string                  `json:"description,omitempty" binding:"max=1000"`
	Required    bool                    `json:"required"`
	Options     []QuestionOptionRequest `json:"options,omitempty"`
	Validation  ValidationRules         `json:"validation,omitempty"`
	Order       int                     `json:"order"`
	Metadata    map[string]interface{}  `json:"metadata,omitempty"`
}

// QuestionOptionRequest represents a question option in a request
type QuestionOptionRequest struct {
	ID    *uuid.UUID `json:"id,omitempty"`
	Label string     `json:"label" binding:"required,min=1,max=200"`
	Value string     `json:"value" binding:"required,min=1,max=200"`
	Order int        `json:"order"`
}

// Utility functions for creating domain entities

// NewForm creates a new form instance with defaults
func NewForm(userID uuid.UUID, req CreateFormRequest) *Form {
	now := time.Now()
	formID := uuid.New()

	questions := make([]Question, len(req.Questions))
	for i, qReq := range req.Questions {
		questions[i] = NewQuestion(qReq)
	}

	return &Form{
		ID:          formID,
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Questions:   questions,
		Settings:    req.Settings,
		Status:      FormStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    req.Metadata,
	}
}

// NewQuestion creates a new question instance
func NewQuestion(req QuestionRequest) Question {
	questionID := uuid.New()
	if req.ID != nil {
		questionID = *req.ID
	}

	options := make([]QuestionOption, len(req.Options))
	for i, optReq := range req.Options {
		optionID := uuid.New()
		if optReq.ID != nil {
			optionID = *optReq.ID
		}

		options[i] = QuestionOption{
			ID:    optionID,
			Label: optReq.Label,
			Value: optReq.Value,
			Order: optReq.Order,
		}
	}

	return Question{
		ID:          questionID,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Required:    req.Required,
		Options:     options,
		Validation:  req.Validation,
		Order:       req.Order,
		Metadata:    req.Metadata,
	}
}
