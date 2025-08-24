package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// QuestionType represents the type of a question
type QuestionType string

const (
	QuestionTypeText     QuestionType = "text"
	QuestionTypeTextarea QuestionType = "textarea"
	QuestionTypeNumber   QuestionType = "number"
	QuestionTypeEmail    QuestionType = "email"
	QuestionTypeSelect   QuestionType = "select"
	QuestionTypeRadio    QuestionType = "radio"
	QuestionTypeCheckbox QuestionType = "checkbox"
)

// Question represents a question entity
type Question struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	FormID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"form_id"`
	Type        QuestionType   `gorm:"size:20;not null" json:"type"`
	Title       string         `gorm:"size:500;not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Order       int            `gorm:"not null" json:"order"`
	Options     datatypes.JSON `gorm:"type:jsonb" json:"options,omitempty"`
	Validation  datatypes.JSON `gorm:"type:jsonb" json:"validation"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	Form *Form `gorm:"foreignKey:FormID" json:"form,omitempty"`
}

// BeforeCreate GORM hook called before creating a question
func (q *Question) BeforeCreate(tx *gorm.DB) error {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}

	return q.Validate()
}

// Validate validates the question fields
func (q *Question) Validate() error {
	q.Title = strings.TrimSpace(q.Title)
	q.Description = strings.TrimSpace(q.Description)

	if q.Title == "" {
		return fmt.Errorf("question title is required")
	}
	if len(q.Title) > 500 {
		return fmt.Errorf("question title cannot exceed 500 characters")
	}
	if len(q.Description) > 1000 {
		return fmt.Errorf("question description cannot exceed 1000 characters")
	}
	if q.Order < 0 {
		return fmt.Errorf("question order must be non-negative")
	}

	return nil
}

// TableName returns the table name for GORM
func (Question) TableName() string {
	return "questions"
}
