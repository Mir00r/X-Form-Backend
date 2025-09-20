package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
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

// Value implements the driver.Valuer interface for FormSettings
func (fs FormSettings) Value() (interface{}, error) {
	return json.Marshal(fs)
}

// Scan implements the sql.Scanner interface for FormSettings
func (fs *FormSettings) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into FormSettings", value)
	}

	return json.Unmarshal(bytes, fs)
}

// Form represents a form entity
type Form struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Title       string         `gorm:"size:200;not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Status      FormStatus     `gorm:"size:20;not null;default:'draft'" json:"status"`
	Settings    FormSettings   `gorm:"type:jsonb" json:"settings"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate GORM hook called before creating a form
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

	if f.Settings.ConfirmationMessage == "" {
		f.Settings.ConfirmationMessage = "Thank you for your response!"
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

	return f.Settings.Validate()
}

// TableName returns the table name for GORM
func (Form) TableName() string {
	return "forms"
}
