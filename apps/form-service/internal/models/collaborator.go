package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CollaboratorRole represents the role of a collaborator
type CollaboratorRole string

const (
	CollaboratorRoleViewer CollaboratorRole = "viewer"
	CollaboratorRoleEditor CollaboratorRole = "editor"
	CollaboratorRoleOwner  CollaboratorRole = "owner"
	CollaboratorRoleAdmin  CollaboratorRole = "admin"
)

// IsValid validates if the collaborator role is valid
func (cr CollaboratorRole) IsValid() bool {
	switch cr {
	case CollaboratorRoleViewer, CollaboratorRoleEditor, CollaboratorRoleOwner, CollaboratorRoleAdmin:
		return true
	default:
		return false
	}
}

// CollaboratorStatus represents the status of a collaborator
type CollaboratorStatus string

const (
	CollaboratorStatusPending  CollaboratorStatus = "pending"
	CollaboratorStatusActive   CollaboratorStatus = "active"
	CollaboratorStatusInactive CollaboratorStatus = "inactive"
)

// IsValid validates if the collaborator status is valid
func (cs CollaboratorStatus) IsValid() bool {
	switch cs {
	case CollaboratorStatusPending, CollaboratorStatusActive, CollaboratorStatusInactive:
		return true
	default:
		return false
	}
}

// Collaborator represents a form collaborator entity
type Collaborator struct {
	ID        uuid.UUID          `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FormID    uuid.UUID          `json:"form_id" gorm:"type:uuid;not null;index"`
	UserID    *uuid.UUID         `json:"user_id,omitempty" gorm:"type:uuid;index"`
	Email     string             `json:"email" gorm:"not null;index"`
	Role      CollaboratorRole   `json:"role" gorm:"not null;default:'viewer'"`
	Status    CollaboratorStatus `json:"status" gorm:"not null;default:'pending'"`
	InvitedBy uuid.UUID          `json:"invited_by" gorm:"type:uuid;not null"`
	InvitedAt time.Time          `json:"invited_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	JoinedAt  *time.Time         `json:"joined_at,omitempty"`
	CreatedAt time.Time          `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time          `json:"updated_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt     `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Form *Form `json:"form,omitempty" gorm:"foreignKey:FormID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// TableName returns the table name for the Collaborator model
func (Collaborator) TableName() string {
	return "collaborators"
}

// BeforeCreate hook is called before creating a collaborator
func (c *Collaborator) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// Validate validates the collaborator data
func (c *Collaborator) Validate() error {
	if c.FormID == uuid.Nil {
		return fmt.Errorf("form_id is required")
	}
	if c.Email == "" {
		return fmt.Errorf("email is required")
	}
	if !c.Role.IsValid() {
		return fmt.Errorf("invalid role: %s", c.Role)
	}
	if !c.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", c.Status)
	}
	if c.InvitedBy == uuid.Nil {
		return fmt.Errorf("invited_by is required")
	}
	return nil
}

// CanRead checks if the collaborator can read the form
func (c *Collaborator) CanRead() bool {
	return c.Status == CollaboratorStatusActive &&
		(c.Role == CollaboratorRoleViewer || c.Role == CollaboratorRoleEditor ||
			c.Role == CollaboratorRoleOwner || c.Role == CollaboratorRoleAdmin)
}

// CanEdit checks if the collaborator can edit the form
func (c *Collaborator) CanEdit() bool {
	return c.Status == CollaboratorStatusActive &&
		(c.Role == CollaboratorRoleEditor || c.Role == CollaboratorRoleOwner || c.Role == CollaboratorRoleAdmin)
}

// CanManage checks if the collaborator can manage other collaborators
func (c *Collaborator) CanManage() bool {
	return c.Status == CollaboratorStatusActive &&
		(c.Role == CollaboratorRoleOwner || c.Role == CollaboratorRoleAdmin)
}

// IsOwner checks if the collaborator is the owner
func (c *Collaborator) IsOwner() bool {
	return c.Status == CollaboratorStatusActive && c.Role == CollaboratorRoleOwner
}
