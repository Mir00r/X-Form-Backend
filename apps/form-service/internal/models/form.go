package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// FormStatus represents the status of a form
type FormStatus string

const (
	FormStatusDraft     FormStatus = "draft"
	FormStatusPublished FormStatus = "published"
	FormStatusClosed    FormStatus = "closed"
)

// IsValid validates if the form status is valid
func (fs FormStatus) IsValid() bool {
	switch fs {
	case FormStatusDraft, FormStatusPublished, FormStatusClosed:
		return true
	default:
		return false
	}
}

// FormSettings represents the settings of a form
type FormSettings struct {
	AcceptingResponses    bool   `json:"accepting_responses"`
	RequireSignIn         bool   `json:"require_sign_in"`
	ConfirmationMessage   string `json:"confirmation_message"`
	AllowMultipleResponse bool   `json:"allow_multiple_response"`
	ShowProgressBar       bool   `json:"show_progress_bar"`
	ShuffleQuestions      bool   `json:"shuffle_questions"`
}

// Validate validates the form settings
func (fs FormSettings) Validate() error {
	if len(fs.ConfirmationMessage) > 1000 {
		return fmt.Errorf("confirmation message cannot exceed 1000 characters")
	}
	return nil
}

// Form represents a form entity
type Form struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Title       string         `gorm:"size:200;not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Status      FormStatus     `gorm:"size:20;not null;default:'draft'" json:"status"`
	Settings    datatypes.JSON `gorm:"type:jsonb" json:"settings"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Computed fields (not stored in database)
	QuestionCount     int `gorm:"-" json:"question_count,omitempty"`
	CollaboratorCount int `gorm:"-" json:"collaborator_count,omitempty"`
	ResponseCount     int `gorm:"-" json:"response_count,omitempty"`
}

// BeforeCreate hook is called before creating a form
func (f *Form) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}

	if err := f.Validate(); err != nil {
		return err
	}

	if f.Status == "" {
		f.Status = FormStatusDraft
	}

	return nil
}

// Validate validates the form fields
func (f *Form) Validate() error {
	f.Title = strings.TrimSpace(f.Title)
	f.Description = strings.TrimSpace(f.Description)

	if f.Title == "" {
		return fmt.Errorf("form title is required")
	}
	if len(f.Title) > 200 {
		return fmt.Errorf("form title cannot exceed 200 characters")
	}
	if len(f.Description) > 2000 {
		return fmt.Errorf("form description cannot exceed 2000 characters")
	}
	if !f.Status.IsValid() {
		return fmt.Errorf("invalid form status: %s", f.Status)
	}

	// Validate settings if they exist
	if len(f.Settings) > 0 {
		var settings FormSettings
		if err := json.Unmarshal(f.Settings, &settings); err != nil {
			return fmt.Errorf("invalid form settings JSON: %w", err)
		}
		if err := settings.Validate(); err != nil {
			return fmt.Errorf("invalid form settings: %w", err)
		}
	}

	return nil
}

// TableName returns the table name for GORM
func (Form) TableName() string {
	return "forms"
}
