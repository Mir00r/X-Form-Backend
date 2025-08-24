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
}

// QuestionRepository defines the interface for question data operations
type QuestionRepository interface {
	// Question CRUD operations
	Create(ctx context.Context, question *models.Question) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Question, error)
	GetByFormID(ctx context.Context, formID uuid.UUID) ([]*models.Question, error)
	Update(ctx context.Context, question *models.Question) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetMaxOrder(ctx context.Context, formID uuid.UUID) (int, error)
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
	return r.db.WithContext(ctx).Create(form).Error
}

// GetByID retrieves a form by its ID
func (r *formRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Form, error) {
	var form models.Form
	err := r.db.WithContext(ctx).First(&form, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
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
	err := r.db.WithContext(ctx).First(&question, "id = ?", id).Error
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
