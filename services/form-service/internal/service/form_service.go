package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/models"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/repository"
)

// FormService defines the interface for form business logic
type FormService interface {
	// Form operations
	CreateForm(ctx context.Context, userID uuid.UUID, req CreateFormRequest) (*models.Form, error)
	GetForm(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Form, error)
	GetUserForms(ctx context.Context, userID uuid.UUID, page, limit int) (*PaginatedFormsResponse, error)
	UpdateForm(ctx context.Context, id uuid.UUID, userID uuid.UUID, req UpdateFormRequest) (*models.Form, error)
	DeleteForm(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	PublishForm(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Form, error)

	// Question operations
	AddQuestion(ctx context.Context, formID uuid.UUID, userID uuid.UUID, req AddQuestionRequest) (*models.Question, error)
	UpdateQuestion(ctx context.Context, questionID uuid.UUID, userID uuid.UUID, req UpdateQuestionRequest) (*models.Question, error)
	DeleteQuestion(ctx context.Context, questionID uuid.UUID, userID uuid.UUID) error
	ReorderQuestions(ctx context.Context, formID uuid.UUID, userID uuid.UUID, req ReorderQuestionsRequest) error
}

// CreateFormRequest represents a request to create a form
type CreateFormRequest struct {
	Title       string              `json:"title" binding:"required,max=200"`
	Description string              `json:"description" binding:"max=2000"`
	Settings    models.FormSettings `json:"settings"`
}

// UpdateFormRequest represents a request to update a form
type UpdateFormRequest struct {
	Title       *string              `json:"title,omitempty" binding:"omitempty,max=200"`
	Description *string              `json:"description,omitempty" binding:"omitempty,max=2000"`
	Settings    *models.FormSettings `json:"settings,omitempty"`
}

// AddQuestionRequest represents a request to add a question
type AddQuestionRequest struct {
	Type        models.QuestionType `json:"type" binding:"required"`
	Title       string              `json:"title" binding:"required,max=500"`
	Description string              `json:"description" binding:"max=1000"`
	Order       int                 `json:"order"`
	Options     interface{}         `json:"options,omitempty"`
	Validation  interface{}         `json:"validation,omitempty"`
}

// UpdateQuestionRequest represents a request to update a question
type UpdateQuestionRequest struct {
	Type        *models.QuestionType `json:"type,omitempty"`
	Title       *string              `json:"title,omitempty" binding:"omitempty,max=500"`
	Description *string              `json:"description,omitempty" binding:"omitempty,max=1000"`
	Order       *int                 `json:"order,omitempty"`
	Options     interface{}          `json:"options,omitempty"`
	Validation  interface{}          `json:"validation,omitempty"`
}

// ReorderQuestionsRequest represents a request to reorder questions
type ReorderQuestionsRequest struct {
	QuestionOrders []QuestionOrder `json:"question_orders" binding:"required"`
}

// QuestionOrder represents a question ordering
type QuestionOrder struct {
	ID    uuid.UUID `json:"id" binding:"required"`
	Order int       `json:"order" binding:"min=0"`
}

// PaginatedFormsResponse represents a paginated list of forms
type PaginatedFormsResponse struct {
	Forms      []*models.Form `json:"forms"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"total_pages"`
}

// formService implements FormService interface
type formService struct {
	formRepo     repository.FormRepository
	questionRepo repository.QuestionRepository
}

// NewFormService creates a new form service instance
func NewFormService(formRepo repository.FormRepository, questionRepo repository.QuestionRepository) FormService {
	return &formService{
		formRepo:     formRepo,
		questionRepo: questionRepo,
	}
}

// CreateForm creates a new form
func (s *formService) CreateForm(ctx context.Context, userID uuid.UUID, req CreateFormRequest) (*models.Form, error) {
	form := &models.Form{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Status:      models.FormStatusDraft,
		Settings:    req.Settings,
	}

	if err := s.formRepo.Create(ctx, form); err != nil {
		return nil, fmt.Errorf("failed to create form: %w", err)
	}

	return form, nil
}

// GetForm retrieves a form by ID with access control
func (s *formService) GetForm(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Form, error) {
	form, err := s.formRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get form: %w", err)
	}

	// Check if user has access to the form
	if form.UserID != userID {
		return nil, fmt.Errorf("access denied: user does not own this form")
	}

	return form, nil
}

// GetUserForms retrieves forms for a user with pagination
func (s *formService) GetUserForms(ctx context.Context, userID uuid.UUID, page, limit int) (*PaginatedFormsResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	forms, err := s.formRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user forms: %w", err)
	}

	total, err := s.formRepo.Count(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count user forms: %w", err)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &PaginatedFormsResponse{
		Forms:      forms,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// UpdateForm updates an existing form
func (s *formService) UpdateForm(ctx context.Context, id uuid.UUID, userID uuid.UUID, req UpdateFormRequest) (*models.Form, error) {
	form, err := s.GetForm(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Title != nil {
		form.Title = *req.Title
	}
	if req.Description != nil {
		form.Description = *req.Description
	}
	if req.Settings != nil {
		form.Settings = *req.Settings
	}

	if err := s.formRepo.Update(ctx, form); err != nil {
		return nil, fmt.Errorf("failed to update form: %w", err)
	}

	return form, nil
}

// DeleteForm deletes a form
func (s *formService) DeleteForm(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	form, err := s.GetForm(ctx, id, userID)
	if err != nil {
		return err
	}

	if err := s.formRepo.Delete(ctx, form.ID); err != nil {
		return fmt.Errorf("failed to delete form: %w", err)
	}

	return nil
}

// PublishForm publishes a form
func (s *formService) PublishForm(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Form, error) {
	form, err := s.GetForm(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if form.Status == models.FormStatusPublished {
		return form, nil // Already published
	}

	form.Status = models.FormStatusPublished

	if err := s.formRepo.Update(ctx, form); err != nil {
		return nil, fmt.Errorf("failed to publish form: %w", err)
	}

	return form, nil
}

// AddQuestion adds a new question to a form
func (s *formService) AddQuestion(ctx context.Context, formID uuid.UUID, userID uuid.UUID, req AddQuestionRequest) (*models.Question, error) {
	// Verify user owns the form
	_, err := s.GetForm(ctx, formID, userID)
	if err != nil {
		return nil, err
	}

	question := &models.Question{
		FormID:      formID,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Order:       req.Order,
		// TODO: Convert options and validation to proper JSONB types
	}

	if err := s.questionRepo.Create(ctx, question); err != nil {
		return nil, fmt.Errorf("failed to create question: %w", err)
	}

	return question, nil
}

// UpdateQuestion updates an existing question
func (s *formService) UpdateQuestion(ctx context.Context, questionID uuid.UUID, userID uuid.UUID, req UpdateQuestionRequest) (*models.Question, error) {
	question, err := s.questionRepo.GetByID(ctx, questionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get question: %w", err)
	}

	// Verify user owns the form
	_, err = s.GetForm(ctx, question.FormID, userID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Type != nil {
		question.Type = *req.Type
	}
	if req.Title != nil {
		question.Title = *req.Title
	}
	if req.Description != nil {
		question.Description = *req.Description
	}
	if req.Order != nil {
		question.Order = *req.Order
	}

	if err := s.questionRepo.Update(ctx, question); err != nil {
		return nil, fmt.Errorf("failed to update question: %w", err)
	}

	return question, nil
}

// DeleteQuestion deletes a question
func (s *formService) DeleteQuestion(ctx context.Context, questionID uuid.UUID, userID uuid.UUID) error {
	question, err := s.questionRepo.GetByID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("failed to get question: %w", err)
	}

	// Verify user owns the form
	_, err = s.GetForm(ctx, question.FormID, userID)
	if err != nil {
		return err
	}

	if err := s.questionRepo.Delete(ctx, questionID); err != nil {
		return fmt.Errorf("failed to delete question: %w", err)
	}

	return nil
}

// ReorderQuestions reorders questions in a form
func (s *formService) ReorderQuestions(ctx context.Context, formID uuid.UUID, userID uuid.UUID, req ReorderQuestionsRequest) error {
	// Verify user owns the form
	_, err := s.GetForm(ctx, formID, userID)
	if err != nil {
		return err
	}

	// Update each question's order
	for _, qo := range req.QuestionOrders {
		question, err := s.questionRepo.GetByID(ctx, qo.ID)
		if err != nil {
			return fmt.Errorf("failed to get question %s: %w", qo.ID, err)
		}

		if question.FormID != formID {
			return fmt.Errorf("question %s does not belong to form %s", qo.ID, formID)
		}

		question.Order = qo.Order
		if err := s.questionRepo.Update(ctx, question); err != nil {
			return fmt.Errorf("failed to update question order: %w", err)
		}
	}

	return nil
}
