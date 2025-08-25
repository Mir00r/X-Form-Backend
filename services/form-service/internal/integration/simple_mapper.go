// Simplified DTO Mapper for Form Service Integration
// Provides basic conversion between domain and DTO objects with realistic mappings

package integration

import (
	"time"

	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/domain"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/dto"
	"github.com/google/uuid"
)

// SimplifiedFormMapper handles basic conversions between domain and DTO
type SimplifiedFormMapper struct{}

// NewSimplifiedFormMapper creates a new simplified mapper
func NewSimplifiedFormMapper() *SimplifiedFormMapper {
	return &SimplifiedFormMapper{}
}

// =============================================================================
// Domain to DTO Mappings
// =============================================================================

// ToFormResponseDTO converts domain Form to response DTO
func (m *SimplifiedFormMapper) ToFormResponseDTO(form *domain.Form) *dto.FormResponseDTO {
	if form == nil {
		return nil
	}

	return &dto.FormResponseDTO{
		ID:            form.ID.String(),
		Title:         form.Title,
		Description:   form.Description,
		Status:        string(form.Status),
		IsAnonymous:   form.Settings.AllowAnonymous,
		IsPublic:      form.Settings.IsPublic,
		AllowMultiple: !form.Settings.RequireAuthentication,
		CreatedBy:     m.toUserInfoDTO(form.UserID),
		Settings:      m.toFormSettingsDTO(form.Settings),
		Questions:     m.toQuestionResponseDTOs(form.Questions),
		Tags:          []string{}, // Default empty
		Category:      "",         // Default empty
		Statistics:    m.defaultFormStatistics(),
		CreatedAt:     form.CreatedAt,
		UpdatedAt:     form.UpdatedAt,
		PublishedAt:   form.PublishedAt,
		ExpiresAt:     form.Settings.ExpiresAt,
	}
}

// toQuestionResponseDTOs converts domain questions to response DTOs
func (m *SimplifiedFormMapper) toQuestionResponseDTOs(questions []domain.Question) []dto.QuestionResponseDTO {
	if questions == nil {
		return []dto.QuestionResponseDTO{}
	}

	result := make([]dto.QuestionResponseDTO, len(questions))
	for i, q := range questions {
		result[i] = dto.QuestionResponseDTO{
			ID:          q.ID.String(),
			Type:        string(q.Type),
			Label:       q.Title,
			Description: q.Description,
			Required:    q.Required,
			Order:       q.Order,
			Options:     m.toQuestionOptionDTOs(q.Options),
			Validation:  m.toQuestionValidationDTO(q.Validation),
			Metadata:    q.Metadata,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}
	return result
}

// toQuestionOptionDTOs converts domain options to DTOs
func (m *SimplifiedFormMapper) toQuestionOptionDTOs(options []domain.QuestionOption) []dto.QuestionOptionDTO {
	if options == nil {
		return []dto.QuestionOptionDTO{}
	}

	result := make([]dto.QuestionOptionDTO, len(options))
	for i, opt := range options {
		result[i] = dto.QuestionOptionDTO{
			Value: opt.Value,
			Label: opt.Label,
			Order: opt.Order,
		}
	}
	return result
}

// toQuestionValidationDTO converts domain validation to DTO
func (m *SimplifiedFormMapper) toQuestionValidationDTO(validation domain.ValidationRules) *dto.QuestionValidationDTO {
	var pattern string
	if validation.Pattern != nil {
		pattern = *validation.Pattern
	}

	return &dto.QuestionValidationDTO{
		MinLength: validation.MinLength,
		MaxLength: validation.MaxLength,
		Pattern:   pattern,
		MinValue:  validation.MinValue,
		MaxValue:  validation.MaxValue,
	}
}

// toFormSettingsDTO converts domain settings to DTO
func (m *SimplifiedFormMapper) toFormSettingsDTO(settings domain.FormSettings) dto.FormSettingsDTO {
	return dto.FormSettingsDTO{
		RequireLogin:       settings.RequireAuthentication,
		CollectEmail:       true,
		ShowProgressBar:    true,
		AllowDrafts:        false,
		NotifyOnSubmission: settings.EnableNotifications,
		CustomCSS:          "",
		RedirectURL:        "",
		ThankYouMessage:    "",
		Metadata:           make(map[string]interface{}),
	}
}

// toUserInfoDTO creates user info from user ID
func (m *SimplifiedFormMapper) toUserInfoDTO(userID uuid.UUID) dto.UserInfoDTO {
	return dto.UserInfoDTO{
		ID:       userID.String(),
		Username: "user",
		Email:    "",
		Name:     "",
		Avatar:   "",
	}
}

// defaultFormStatistics returns default statistics
func (m *SimplifiedFormMapper) defaultFormStatistics() dto.FormStatisticsDTO {
	return dto.FormStatisticsDTO{
		TotalResponses:   0,
		UniqueResponders: 0,
		CompletionRate:   0.0,
		AverageTime:      0,
		LastResponse:     nil,
		ResponseRate:     0.0,
	}
}

// =============================================================================
// DTO to Domain Mappings
// =============================================================================

// ToFormDomain converts CreateFormRequestDTO to domain Form
func (m *SimplifiedFormMapper) ToFormDomain(createDTO *dto.CreateFormRequestDTO, userID string) (*domain.Form, error) {
	if createDTO == nil {
		return nil, domain.NewValidationError("dto", "create form DTO is required")
	}

	// Parse user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.NewValidationError("userID", "invalid user ID format")
	}

	// Convert questions
	questions, err := m.toQuestionsDomain(createDTO.Questions)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	form := &domain.Form{
		UserID:      userUUID,
		Title:       createDTO.Title,
		Description: createDTO.Description,
		Questions:   questions,
		Settings:    m.toFormSettingsDomain(createDTO.Settings),
		Status:      domain.FormStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    make(map[string]interface{}),
	}

	return form, nil
}

// UpdateFormDomain updates existing form with UpdateFormRequestDTO
func (m *SimplifiedFormMapper) UpdateFormDomain(existing *domain.Form, updateDTO *dto.UpdateFormRequestDTO) error {
	if existing == nil {
		return domain.NewValidationError("form", "existing form is required")
	}
	if updateDTO == nil {
		return domain.NewValidationError("dto", "update form DTO is required")
	}

	// Update fields if provided
	if updateDTO.Title != nil {
		existing.Title = *updateDTO.Title
	}
	if updateDTO.Description != nil {
		existing.Description = *updateDTO.Description
	}
	if updateDTO.Settings != nil {
		existing.Settings = m.toFormSettingsDomain(*updateDTO.Settings)
	}

	existing.UpdatedAt = time.Now()
	return nil
}

// toQuestionsDomain converts DTO questions to domain
func (m *SimplifiedFormMapper) toQuestionsDomain(dtoQuestions []dto.CreateQuestionRequestDTO) ([]domain.Question, error) {
	if len(dtoQuestions) == 0 {
		return nil, domain.NewValidationError("questions", "at least one question is required")
	}

	questions := make([]domain.Question, len(dtoQuestions))
	for i, dtoQ := range dtoQuestions {
		questions[i] = domain.Question{
			ID:          uuid.New(),
			Type:        domain.QuestionType(dtoQ.Type),
			Title:       dtoQ.Label,
			Description: dtoQ.Description,
			Required:    dtoQ.Required,
			Options:     m.toQuestionOptionsDomain(dtoQ.Options),
			Validation:  m.toValidationRulesDomain(dtoQ.Validation),
			Order:       dtoQ.Order,
			Metadata:    dtoQ.Metadata,
		}
	}
	return questions, nil
}

// toQuestionOptionsDomain converts DTO options to domain
func (m *SimplifiedFormMapper) toQuestionOptionsDomain(dtoOptions []dto.QuestionOptionDTO) []domain.QuestionOption {
	if dtoOptions == nil {
		return []domain.QuestionOption{}
	}

	options := make([]domain.QuestionOption, len(dtoOptions))
	for i, dtoOpt := range dtoOptions {
		options[i] = domain.QuestionOption{
			ID:    uuid.New(),
			Label: dtoOpt.Label,
			Value: dtoOpt.Value,
			Order: dtoOpt.Order,
		}
	}
	return options
}

// toValidationRulesDomain converts DTO validation to domain
func (m *SimplifiedFormMapper) toValidationRulesDomain(dtoValidation *dto.QuestionValidationDTO) domain.ValidationRules {
	if dtoValidation == nil {
		return domain.ValidationRules{}
	}

	var pattern *string
	if dtoValidation.Pattern != "" {
		pattern = &dtoValidation.Pattern
	}

	return domain.ValidationRules{
		MinLength:   dtoValidation.MinLength,
		MaxLength:   dtoValidation.MaxLength,
		Pattern:     pattern,
		MinValue:    dtoValidation.MinValue,
		MaxValue:    dtoValidation.MaxValue,
		CustomRules: []string{}, // Default empty
	}
}

// toFormSettingsDomain converts DTO settings to domain
func (m *SimplifiedFormMapper) toFormSettingsDomain(dtoSettings dto.FormSettingsDTO) domain.FormSettings {
	return domain.FormSettings{
		IsPublic:              !dtoSettings.RequireLogin,
		AllowAnonymous:        !dtoSettings.RequireLogin,
		RequireAuthentication: dtoSettings.RequireLogin,
		EnableNotifications:   dtoSettings.NotifyOnSubmission,
		SubmissionLimit:       nil,
		ExpiresAt:             nil,
		AllowedDomains:        []string{},
	}
}

// =============================================================================
// Response Helpers
// =============================================================================

// ToSuccessResponse creates a standardized success response
func (m *SimplifiedFormMapper) ToSuccessResponse(data interface{}, message string) *dto.SuccessResponse {
	return &dto.SuccessResponse{
		BaseResponse: dto.BaseResponse{
			Success:   true,
			Message:   message,
			Timestamp: time.Now(),
			Version:   "v1",
		},
		Data: data,
	}
}

// ToErrorResponse creates a standardized error response
func (m *SimplifiedFormMapper) ToErrorResponse(code, message, correlationID string) *dto.ErrorResponse {
	return &dto.ErrorResponse{
		BaseResponse: dto.BaseResponse{
			Success:       false,
			Message:       "Request failed",
			CorrelationID: correlationID,
			Timestamp:     time.Now(),
			Version:       "v1",
		},
		Error: dto.ErrorDetail{
			Code:      code,
			Message:   message,
			Timestamp: time.Now(),
			RequestID: correlationID,
		},
	}
}

// ToValidationErrorResponse creates a validation error response
func (m *SimplifiedFormMapper) ToValidationErrorResponse(errors map[string]string, correlationID string) *dto.ErrorResponse {
	return &dto.ErrorResponse{
		BaseResponse: dto.BaseResponse{
			Success:       false,
			Message:       "Validation failed",
			CorrelationID: correlationID,
			Timestamp:     time.Now(),
			Version:       "v1",
		},
		Error: dto.ErrorDetail{
			Code:      "VALIDATION_ERROR",
			Message:   "Input validation failed",
			Fields:    errors,
			Timestamp: time.Now(),
			RequestID: correlationID,
		},
	}
}

// ToPaginationResponse creates pagination metadata
func (m *SimplifiedFormMapper) ToPaginationResponse(page, pageSize int, total int64) dto.Pagination {
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return dto.Pagination{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}
