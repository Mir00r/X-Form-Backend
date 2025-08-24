// Package application contains the application services (use cases)
// This layer orchestrates domain entities and repositories to fulfill business requirements
package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/domain"
)

// FormApplicationService implements business use cases for forms
// It follows the Single Responsibility Principle by focusing only on orchestrating business logic
type FormApplicationService struct {
	formRepo domain.FormRepository
	logger   *logrus.Logger
	// TODO: Add event publisher for domain events
	// eventPublisher events.Publisher
}

// NewFormApplicationService creates a new form application service
// Uses Dependency Injection for better testability and loose coupling
func NewFormApplicationService(
	formRepo domain.FormRepository,
	logger *logrus.Logger,
) *FormApplicationService {
	return &FormApplicationService{
		formRepo: formRepo,
		logger:   logger,
	}
}

// CreateForm implements the create form use case
// Applies business validation and orchestrates domain logic
func (s *FormApplicationService) CreateForm(
	ctx context.Context,
	userID uuid.UUID,
	req domain.CreateFormRequest,
) (*domain.Form, error) {
	// Create domain entity
	form := domain.NewForm(userID, req)

	// Apply business validation
	if err := form.Validate(); err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Warn("Form validation failed")
		return nil, err
	}

	// Persist the form
	if err := s.formRepo.Create(ctx, form); err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"form_id": form.ID,
			"error":   err,
		}).Error("Failed to create form")
		return nil, domain.NewInternalError("create form", "failed to persist form", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"form_id": form.ID,
		"title":   form.Title,
	}).Info("Form created successfully")

	// TODO: Publish domain event
	// s.eventPublisher.Publish(events.FormCreated{FormID: form.ID, UserID: userID})

	return form, nil
}

// GetForm implements the get form use case with access control
func (s *FormApplicationService) GetForm(
	ctx context.Context,
	formID uuid.UUID,
	userID *uuid.UUID,
) (*domain.Form, error) {
	form, err := s.formRepo.GetByID(ctx, formID)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"form_id": formID,
			"user_id": userID,
			"error":   err,
		}).Error("Failed to retrieve form")
		return nil, domain.NewInternalError("get form", "failed to retrieve form", err)
	}

	if form == nil {
		return nil, domain.NewNotFoundError("form", formID.String())
	}

	// Apply access control rules
	if userID != nil && form.UserID != *userID {
		// Check if form is public and published
		if !form.Settings.IsPublic || form.Status != domain.FormStatusPublished {
			return nil, domain.NewAccessDeniedError(
				"form", "read", userID.String(),
				"user does not have access to this form",
			)
		}
	}

	return form, nil
}

// UpdateForm implements the update form use case
func (s *FormApplicationService) UpdateForm(
	ctx context.Context,
	formID uuid.UUID,
	userID uuid.UUID,
	req domain.UpdateFormRequest,
) (*domain.Form, error) {
	// Retrieve existing form
	form, err := s.formRepo.GetByID(ctx, formID)
	if err != nil {
		return nil, domain.NewInternalError("update form", "failed to retrieve form", err)
	}

	if form == nil {
		return nil, domain.NewNotFoundError("form", formID.String())
	}

	// Check ownership
	if form.UserID != userID {
		return nil, domain.NewAccessDeniedError(
			"form", "update", userID.String(),
			"user does not own this form",
		)
	}

	// Apply updates
	if err := s.applyFormUpdates(form, req); err != nil {
		return nil, err
	}

	// Validate updated form
	if err := form.Validate(); err != nil {
		return nil, err
	}

	// Update timestamp
	form.UpdatedAt = time.Now()

	// Persist changes
	if err := s.formRepo.Update(ctx, form); err != nil {
		s.logger.WithFields(logrus.Fields{
			"form_id": formID,
			"user_id": userID,
			"error":   err,
		}).Error("Failed to update form")
		return nil, domain.NewInternalError("update form", "failed to persist changes", err)
	}

	s.logger.WithFields(logrus.Fields{
		"form_id": formID,
		"user_id": userID,
	}).Info("Form updated successfully")

	return form, nil
}

// DeleteForm implements the delete form use case
func (s *FormApplicationService) DeleteForm(
	ctx context.Context,
	formID uuid.UUID,
	userID uuid.UUID,
) error {
	// Retrieve form for ownership check
	form, err := s.formRepo.GetByID(ctx, formID)
	if err != nil {
		return domain.NewInternalError("delete form", "failed to retrieve form", err)
	}

	if form == nil {
		return domain.NewNotFoundError("form", formID.String())
	}

	// Check ownership
	if form.UserID != userID {
		return domain.NewAccessDeniedError(
			"form", "delete", userID.String(),
			"user does not own this form",
		)
	}

	// Business rule: cannot delete published forms with responses
	// TODO: Check if form has responses before allowing deletion
	if form.Status == domain.FormStatusPublished {
		return domain.NewBusinessRuleError(
			"cannot delete published form - close it first or contact administrator",
		)
	}

	// Delete form
	if err := s.formRepo.Delete(ctx, formID); err != nil {
		s.logger.WithFields(logrus.Fields{
			"form_id": formID,
			"user_id": userID,
			"error":   err,
		}).Error("Failed to delete form")
		return domain.NewInternalError("delete form", "failed to delete form", err)
	}

	s.logger.WithFields(logrus.Fields{
		"form_id": formID,
		"user_id": userID,
	}).Info("Form deleted successfully")

	return nil
}

// ListUserForms implements the list user forms use case
func (s *FormApplicationService) ListUserForms(
	ctx context.Context,
	userID uuid.UUID,
	filters domain.FormFilters,
) ([]*domain.Form, int64, error) {
	// Ensure user can only see their own forms
	filters.UserID = &userID

	forms, total, err := s.formRepo.GetByUserID(ctx, userID, filters.Offset, filters.Limit)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Error("Failed to list user forms")
		return nil, 0, domain.NewInternalError("list user forms", "failed to retrieve forms", err)
	}

	return forms, total, nil
}

// PublishForm implements the publish form use case
func (s *FormApplicationService) PublishForm(
	ctx context.Context,
	formID uuid.UUID,
	userID uuid.UUID,
) error {
	// Retrieve form
	form, err := s.formRepo.GetByID(ctx, formID)
	if err != nil {
		return domain.NewInternalError("publish form", "failed to retrieve form", err)
	}

	if form == nil {
		return domain.NewNotFoundError("form", formID.String())
	}

	// Check ownership
	if form.UserID != userID {
		return domain.NewAccessDeniedError(
			"form", "publish", userID.String(),
			"user does not own this form",
		)
	}

	// Apply business rules for publishing
	if err := form.Publish(); err != nil {
		return err
	}

	// Update status in repository
	if err := s.formRepo.UpdateStatus(ctx, formID, domain.FormStatusPublished); err != nil {
		s.logger.WithFields(logrus.Fields{
			"form_id": formID,
			"user_id": userID,
			"error":   err,
		}).Error("Failed to publish form")
		return domain.NewInternalError("publish form", "failed to update form status", err)
	}

	s.logger.WithFields(logrus.Fields{
		"form_id": formID,
		"user_id": userID,
		"title":   form.Title,
	}).Info("Form published successfully")

	// TODO: Publish domain event
	// s.eventPublisher.Publish(events.FormPublished{FormID: formID, UserID: userID})

	return nil
}

// CloseForm implements the close form use case
func (s *FormApplicationService) CloseForm(
	ctx context.Context,
	formID uuid.UUID,
	userID uuid.UUID,
) error {
	// Retrieve form
	form, err := s.formRepo.GetByID(ctx, formID)
	if err != nil {
		return domain.NewInternalError("close form", "failed to retrieve form", err)
	}

	if form == nil {
		return domain.NewNotFoundError("form", formID.String())
	}

	// Check ownership
	if form.UserID != userID {
		return domain.NewAccessDeniedError(
			"form", "close", userID.String(),
			"user does not own this form",
		)
	}

	// Apply business rules for closing
	if err := form.Close(); err != nil {
		return err
	}

	// Update status in repository
	if err := s.formRepo.UpdateStatus(ctx, formID, domain.FormStatusClosed); err != nil {
		s.logger.WithFields(logrus.Fields{
			"form_id": formID,
			"user_id": userID,
			"error":   err,
		}).Error("Failed to close form")
		return domain.NewInternalError("close form", "failed to update form status", err)
	}

	s.logger.WithFields(logrus.Fields{
		"form_id": formID,
		"user_id": userID,
	}).Info("Form closed successfully")

	// TODO: Publish domain event
	// s.eventPublisher.Publish(events.FormClosed{FormID: formID, UserID: userID})

	return nil
}

// GetPublicForm implements the get public form use case
func (s *FormApplicationService) GetPublicForm(
	ctx context.Context,
	formID uuid.UUID,
) (*domain.Form, error) {
	form, err := s.formRepo.GetByID(ctx, formID)
	if err != nil {
		return nil, domain.NewInternalError("get public form", "failed to retrieve form", err)
	}

	if form == nil {
		return nil, domain.NewNotFoundError("form", formID.String())
	}

	// Check if form is publicly accessible
	if !form.Settings.IsPublic || form.Status != domain.FormStatusPublished {
		return nil, domain.NewAccessDeniedError(
			"form", "read", "anonymous",
			"form is not publicly accessible",
		)
	}

	// Check if form is expired
	if form.IsExpired() {
		return nil, domain.NewBusinessRuleError("form has expired and is no longer accepting responses")
	}

	return form, nil
}

// Helper methods

// applyFormUpdates applies update request to form entity
func (s *FormApplicationService) applyFormUpdates(form *domain.Form, req domain.UpdateFormRequest) error {
	if req.Title != nil {
		form.Title = *req.Title
	}

	if req.Description != nil {
		form.Description = *req.Description
	}

	if req.Settings != nil {
		form.Settings = *req.Settings
	}

	if req.Questions != nil {
		questions := make([]domain.Question, len(req.Questions))
		for i, qReq := range req.Questions {
			questions[i] = domain.NewQuestion(qReq)
		}
		form.Questions = questions
	}

	if req.Metadata != nil {
		form.Metadata = req.Metadata
	}

	return nil
}
