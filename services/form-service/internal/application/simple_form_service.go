// Package application contains simplified application services
// This is a temporary version that works with current dependencies
package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/domain"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/logger"
)

// SimpleFormApplicationService implements business use cases for forms
// Simplified version that works with current dependencies
type SimpleFormApplicationService struct {
	formRepo domain.FormRepository
	logger   logger.Logger
}

// NewSimpleFormApplicationService creates a new simplified form application service
func NewSimpleFormApplicationService(
	formRepo domain.FormRepository,
) *SimpleFormApplicationService {
	return &SimpleFormApplicationService{
		formRepo: formRepo,
		logger:   logger.NewSimpleLogger(),
	}
}

// CreateForm creates a new form
func (s *SimpleFormApplicationService) CreateForm(ctx context.Context, req *domain.CreateFormRequest) (*domain.FormResponse, error) {
	s.logger.Info("Creating new form:", req.Title)

	// Validate business rules
	if req.Title == "" {
		return nil, domain.NewValidationError("title", "Form title is required")
	}

	// Create form entity
	form := &domain.Form{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		UserID:      req.UserID,
		IsPublished: false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Add questions if provided
	for _, qReq := range req.Questions {
		question := domain.Question{
			ID:       uuid.New().String(),
			FormID:   form.ID,
			Text:     qReq.Text,
			Type:     qReq.Type,
			Required: qReq.Required,
			Options:  qReq.Options,
		}
		form.Questions = append(form.Questions, question)
	}

	// Save to repository
	savedForm, err := s.formRepo.Create(ctx, form)
	if err != nil {
		s.logger.Error("Failed to create form:", err)
		return nil, domain.NewInternalError("Failed to create form", err)
	}

	// Convert to response
	return domain.ToFormResponse(savedForm), nil
}

// GetForm retrieves a form by ID
func (s *SimpleFormApplicationService) GetForm(ctx context.Context, formID string) (*domain.FormResponse, error) {
	s.logger.Info("Getting form:", formID)

	form, err := s.formRepo.GetByID(ctx, formID)
	if err != nil {
		s.logger.Error("Failed to get form:", err)
		return nil, err
	}

	return domain.ToFormResponse(form), nil
}

// UpdateForm updates an existing form
func (s *SimpleFormApplicationService) UpdateForm(ctx context.Context, formID string, req *domain.UpdateFormRequest) (*domain.FormResponse, error) {
	s.logger.Info("Updating form:", formID)

	// Get existing form
	form, err := s.formRepo.GetByID(ctx, formID)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if form.UserID != req.UserID {
		return nil, domain.NewAccessDeniedError("You don't have permission to update this form")
	}

	// Update fields
	if req.Title != nil {
		form.Title = *req.Title
	}
	if req.Description != nil {
		form.Description = *req.Description
	}
	form.UpdatedAt = time.Now()

	// Save changes
	updatedForm, err := s.formRepo.Update(ctx, form)
	if err != nil {
		s.logger.Error("Failed to update form:", err)
		return nil, domain.NewInternalError("Failed to update form", err)
	}

	return domain.ToFormResponse(updatedForm), nil
}

// DeleteForm deletes a form
func (s *SimpleFormApplicationService) DeleteForm(ctx context.Context, formID, userID string) error {
	s.logger.Info("Deleting form:", formID)

	// Get existing form to check ownership
	form, err := s.formRepo.GetByID(ctx, formID)
	if err != nil {
		return err
	}

	// Check ownership
	if form.UserID != userID {
		return domain.NewAccessDeniedError("You don't have permission to delete this form")
	}

	// Delete form
	err = s.formRepo.Delete(ctx, formID)
	if err != nil {
		s.logger.Error("Failed to delete form:", err)
		return domain.NewInternalError("Failed to delete form", err)
	}

	return nil
}

// PublishForm publishes a form making it available for responses
func (s *SimpleFormApplicationService) PublishForm(ctx context.Context, formID, userID string) (*domain.FormResponse, error) {
	s.logger.Info("Publishing form:", formID)

	// Get existing form
	form, err := s.formRepo.GetByID(ctx, formID)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if form.UserID != userID {
		return nil, domain.NewAccessDeniedError("You don't have permission to publish this form")
	}

	// Validate form can be published
	if len(form.Questions) == 0 {
		return nil, domain.NewBusinessRuleError("Cannot publish form without questions")
	}

	// Update publish status
	form.IsPublished = true
	form.UpdatedAt = time.Now()

	// Save changes
	updatedForm, err := s.formRepo.Update(ctx, form)
	if err != nil {
		s.logger.Error("Failed to publish form:", err)
		return nil, domain.NewInternalError("Failed to publish form", err)
	}

	return domain.ToFormResponse(updatedForm), nil
}
