package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JSON custom type for handling JSON fields
type JSON map[string]interface{}

func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return fmt.Errorf("cannot scan %T into JSON", value)
	}
}

// User model
type User struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Email         string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash  string    `gorm:"column:password_hash" json:"-"`
	FirstName     string    `gorm:"column:first_name" json:"firstName"`
	LastName      string    `gorm:"column:last_name" json:"lastName"`
	AvatarURL     *string   `gorm:"column:avatar_url" json:"avatarUrl,omitempty"`
	Provider      string    `gorm:"default:'email'" json:"provider"`
	ProviderID    *string   `gorm:"column:provider_id" json:"providerId,omitempty"`
	EmailVerified bool      `gorm:"column:email_verified;default:false" json:"emailVerified"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// Form model
type Form struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Title       string     `gorm:"not null" json:"title"`
	Description *string    `json:"description,omitempty"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"userId"`
	User        User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Schema      JSON       `gorm:"type:jsonb;not null" json:"schema"`
	Settings    JSON       `gorm:"type:jsonb;default:'{}'" json:"settings"`
	Status      string     `gorm:"default:'draft'" json:"status"` // draft, published, closed
	PublishedAt *time.Time `gorm:"column:published_at" json:"publishedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	Questions   []Question `gorm:"foreignKey:FormID" json:"questions,omitempty"`
}

// Question model (normalized from form schema)
type Question struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	FormID      uuid.UUID `gorm:"type:uuid;not null;index" json:"formId"`
	Form        Form      `gorm:"foreignKey:FormID" json:"form,omitempty"`
	QuestionID  string    `gorm:"column:question_id;not null" json:"questionId"`
	Type        string    `gorm:"column:question_type;not null" json:"type"`
	Title       string    `gorm:"not null" json:"title"`
	Description *string   `json:"description,omitempty"`
	Required    bool      `gorm:"default:false" json:"required"`
	Options     JSON      `gorm:"type:jsonb" json:"options,omitempty"`
	Validation  JSON      `gorm:"type:jsonb" json:"validation,omitempty"`
	OrderIndex  int       `gorm:"column:order_index;not null" json:"orderIndex"`
	CreatedAt   time.Time `json:"createdAt"`
}

// FormCollaborator model (for team features - post-MVP)
type FormCollaborator struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	FormID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"formId"`
	Form      Form       `gorm:"foreignKey:FormID" json:"form,omitempty"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"userId"`
	User      User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role      string     `gorm:"default:'viewer'" json:"role"` // owner, editor, viewer
	InvitedBy *uuid.UUID `gorm:"type:uuid" json:"invitedBy,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

// FileUpload model
type FileUpload struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"userId"`
	User         User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	FormID       *uuid.UUID `gorm:"type:uuid;index" json:"formId,omitempty"`
	Form         *Form      `gorm:"foreignKey:FormID" json:"form,omitempty"`
	OriginalName string     `gorm:"column:original_name;not null" json:"originalName"`
	FileName     string     `gorm:"column:file_name;not null" json:"fileName"`
	FilePath     string     `gorm:"column:file_path;not null" json:"filePath"`
	FileSize     int64      `gorm:"column:file_size;not null" json:"fileSize"`
	MimeType     string     `gorm:"column:mime_type;not null" json:"mimeType"`
	S3Key        *string    `gorm:"column:s3_key" json:"s3Key,omitempty"`
	S3Bucket     *string    `gorm:"column:s3_bucket" json:"s3Bucket,omitempty"`
	Status       string     `gorm:"default:'processing'" json:"status"` // processing, ready, error
	Metadata     JSON       `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt    time.Time  `json:"createdAt"`
}

// BeforeCreate hook for forms
func (f *Form) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

// BeforeCreate hook for questions
func (q *Question) BeforeCreate(tx *gorm.DB) error {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	return nil
}
