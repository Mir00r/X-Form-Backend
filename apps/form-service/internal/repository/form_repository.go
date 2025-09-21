// Package repository provides data access layer for form management
// Following Repository Pattern with interface segregation
package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/models"
)

// FormRepository defines the interface for form data operations
type FormRepository interface {
	// Form CRUD operations
	Create(ctx context.Context, form *models.Form) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Form, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Form, error)
	Update(ctx context.Context, form *models.Form) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context, userID uuid.UUID) (int64, error)

	// Form access control
	CanUserAccess(ctx context.Context, formID, userID uuid.UUID) (bool, error)
	CanUserEdit(ctx context.Context, formID, userID uuid.UUID) (bool, error)
}

// QuestionRepository defines the interface for question data operations
type QuestionRepository interface {
	// Question CRUD operations
	Create(ctx context.Context, question *models.Question) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Question, error)
	GetByFormID(ctx context.Context, formID uuid.UUID) ([]*models.Question, error)
	Update(ctx context.Context, question *models.Question) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Question ordering
	UpdateOrder(ctx context.Context, formID uuid.UUID, questionOrders []QuestionOrder) error
	GetMaxOrder(ctx context.Context, formID uuid.UUID) (int, error)
}

// CollaboratorRepository defines the interface for collaborator data operations
type CollaboratorRepository interface {
	// Collaborator CRUD operations
	Create(ctx context.Context, collaborator *models.Collaborator) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Collaborator, error)
	GetByFormID(ctx context.Context, formID uuid.UUID) ([]*models.Collaborator, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Collaborator, error)
	Update(ctx context.Context, collaborator *models.Collaborator) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Collaborator specific operations
	FindByFormAndEmail(ctx context.Context, formID uuid.UUID, email string) (*models.Collaborator, error)
	FindByFormAndUser(ctx context.Context, formID, userID uuid.UUID) (*models.Collaborator, error)
}

// QuestionOrder represents a question ordering request
type QuestionOrder struct {
	ID    uuid.UUID `json:"id"`
	Order int       `json:"order"`
}

// formRepository implements FormRepository interface
type formRepository struct {
	db *gorm.DB
}

// NewFormRepository creates a new form repository instance
func NewFormRepository(db *gorm.DB) FormRepository {
	return &formRepository{db: db}
}

// Create creates a new form in the database
func (r *formRepository) Create(ctx context.Context, form *models.Form) error {
	// Settings are handled in the BeforeCreate hook of the model
	return r.db.WithContext(ctx).Create(form).Error
}

// GetByID retrieves a form by its ID with all related data
func (r *formRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Form, error) {
	var form models.Form

	err := r.db.WithContext(ctx).
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"order\" ASC")
		}).
		Preload("Collaborators").
		First(&form, "id = ?", id).Error

	if err != nil {
		return nil, err
	}

	// Load computed fields
	r.loadComputedFields(ctx, &form)

	return &form, nil
}

// GetByUserID retrieves forms for a specific user with pagination
func (r *formRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Form, error) {
	var forms []*models.Form

	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&forms).Error
	if err != nil {
		return nil, err
	}

	// Load computed fields for each form
	for _, form := range forms {
		r.loadComputedFields(ctx, form)
	}

	return forms, nil
}

// Update updates an existing form
func (r *formRepository) Update(ctx context.Context, form *models.Form) error {
	return r.db.WithContext(ctx).Save(form).Error
}

// Delete soft deletes a form
func (r *formRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Form{}, "id = ?", id).Error
}

// Count returns the total number of forms for a user
func (r *formRepository) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Form{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	return count, err
}

// CanUserAccess checks if a user can access a form (view permission)
func (r *formRepository) CanUserAccess(ctx context.Context, formID, userID uuid.UUID) (bool, error) {
	var count int64

	// Check if user is the owner
	err := r.db.WithContext(ctx).
		Model(&models.Form{}).
		Where("id = ? AND user_id = ?", formID, userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	// Check if user is a collaborator
	err = r.db.WithContext(ctx).
		Model(&models.Collaborator{}).
		Where("form_id = ? AND user_id = ?", formID, userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CanUserEdit checks if a user can edit a form
func (r *formRepository) CanUserEdit(ctx context.Context, formID, userID uuid.UUID) (bool, error) {
	var count int64

	// Check if user is the owner
	err := r.db.WithContext(ctx).
		Model(&models.Form{}).
		Where("id = ? AND user_id = ?", formID, userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	// Check if user is a collaborator with edit permissions
	err = r.db.WithContext(ctx).
		Model(&models.Collaborator{}).
		Where("form_id = ? AND user_id = ? AND role IN (?)",
			formID, userID, []string{"owner", "editor"}).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// loadComputedFields loads computed fields for a form
func (r *formRepository) loadComputedFields(ctx context.Context, form *models.Form) {
	// Load question count
	var questionCount int64
	r.db.WithContext(ctx).
		Model(&models.Question{}).
		Where("form_id = ?", form.ID).
		Count(&questionCount)
	form.QuestionCount = int(questionCount)

	// Load collaborator count
	var collaboratorCount int64
	r.db.WithContext(ctx).
		Model(&models.Collaborator{}).
		Where("form_id = ?", form.ID).
		Count(&collaboratorCount)
	form.CollaboratorCount = int(collaboratorCount)

	// TODO: Load response count when response service is implemented
	form.ResponseCount = 0
}

// questionRepository implements QuestionRepository interface
type questionRepository struct {
	db *gorm.DB
}

// NewQuestionRepository creates a new question repository instance
func NewQuestionRepository(db *gorm.DB) QuestionRepository {
	return &questionRepository{db: db}
}

// Create creates a new question in the database
func (r *questionRepository) Create(ctx context.Context, question *models.Question) error {
	// Auto-assign order if not provided
	if question.Order == 0 {
		maxOrder, err := r.GetMaxOrder(ctx, question.FormID)
		if err != nil {
			return fmt.Errorf("failed to get max order: %w", err)
		}
		question.Order = maxOrder + 1
	}

	return r.db.WithContext(ctx).Create(question).Error
}

// GetByID retrieves a question by its ID
func (r *questionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Question, error) {
	var question models.Question

	err := r.db.WithContext(ctx).
		Preload("Form").
		First(&question, "id = ?", id).Error

	if err != nil {
		return nil, err
	}

	return &question, nil
}

// GetByFormID retrieves all questions for a form, ordered by their order field
func (r *questionRepository) GetByFormID(ctx context.Context, formID uuid.UUID) ([]*models.Question, error) {
	var questions []*models.Question

	err := r.db.WithContext(ctx).
		Where("form_id = ?", formID).
		Order("\"order\" ASC").
		Find(&questions).Error

	if err != nil {
		return nil, err
	}

	return questions, nil
}

// Update updates an existing question
func (r *questionRepository) Update(ctx context.Context, question *models.Question) error {
	return r.db.WithContext(ctx).Save(question).Error
}

// Delete soft deletes a question
func (r *questionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Question{}, "id = ?", id).Error
}

// UpdateOrder updates the order of multiple questions in a transaction
func (r *questionRepository) UpdateOrder(ctx context.Context, formID uuid.UUID, questionOrders []QuestionOrder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, qo := range questionOrders {
			err := tx.Model(&models.Question{}).
				Where("id = ? AND form_id = ?", qo.ID, formID).
				Update("order", qo.Order).Error

			if err != nil {
				return err
			}
		}
		return nil
	})
}

// GetMaxOrder returns the maximum order value for questions in a form
func (r *questionRepository) GetMaxOrder(ctx context.Context, formID uuid.UUID) (int, error) {
	var maxOrder int

	err := r.db.WithContext(ctx).
		Model(&models.Question{}).
		Where("form_id = ?", formID).
		Select("COALESCE(MAX(\"order\"), 0)").
		Scan(&maxOrder).Error

	return maxOrder, err
}

// collaboratorRepository implements CollaboratorRepository interface
type collaboratorRepository struct {
	db *gorm.DB
}

// NewCollaboratorRepository creates a new collaborator repository instance
func NewCollaboratorRepository(db *gorm.DB) CollaboratorRepository {
	return &collaboratorRepository{db: db}
}

// Create creates a new collaborator in the database
func (r *collaboratorRepository) Create(ctx context.Context, collaborator *models.Collaborator) error {
	return r.db.WithContext(ctx).Create(collaborator).Error
}

// GetByID retrieves a collaborator by ID
func (r *collaboratorRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Collaborator, error) {
	var collaborator models.Collaborator

	err := r.db.WithContext(ctx).
		Preload("Form").
		First(&collaborator, "id = ?", id).Error

	if err != nil {
		return nil, err
	}

	return &collaborator, nil
}

// GetByFormID retrieves all collaborators for a form
func (r *collaboratorRepository) GetByFormID(ctx context.Context, formID uuid.UUID) ([]*models.Collaborator, error) {
	var collaborators []*models.Collaborator

	err := r.db.WithContext(ctx).
		Where("form_id = ?", formID).
		Order("created_at ASC").
		Find(&collaborators).Error

	if err != nil {
		return nil, err
	}

	return collaborators, nil
}

// GetByUserID retrieves all collaborations for a user
func (r *collaboratorRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Collaborator, error) {
	var collaborators []*models.Collaborator

	err := r.db.WithContext(ctx).
		Preload("Form").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&collaborators).Error

	if err != nil {
		return nil, err
	}

	return collaborators, nil
}

// Update updates an existing collaborator
func (r *collaboratorRepository) Update(ctx context.Context, collaborator *models.Collaborator) error {
	return r.db.WithContext(ctx).Save(collaborator).Error
}

// Delete removes a collaborator
func (r *collaboratorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Collaborator{}, "id = ?", id).Error
}

// FindByFormAndEmail finds a collaborator by form ID and email
func (r *collaboratorRepository) FindByFormAndEmail(ctx context.Context, formID uuid.UUID, email string) (*models.Collaborator, error) {
	var collaborator models.Collaborator

	err := r.db.WithContext(ctx).
		Where("form_id = ? AND email = ?", formID, email).
		First(&collaborator).Error

	if err != nil {
		return nil, err
	}

	return &collaborator, nil
}

// FindByFormAndUser finds a collaborator by form ID and user ID
func (r *collaboratorRepository) FindByFormAndUser(ctx context.Context, formID, userID uuid.UUID) (*models.Collaborator, error) {
	var collaborator models.Collaborator

	err := r.db.WithContext(ctx).
		Where("form_id = ? AND user_id = ?", formID, userID).
		First(&collaborator).Error

	if err != nil {
		return nil, err
	}

	return &collaborator, nil
}
