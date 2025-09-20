// DTO Mapper for Form Service
// Handles conversion between domain objects and DTOs for microservices API layer

package integration

import (
	"time"

	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/domain"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/dto"
)

// FormDTOMapper handles conversion between domain objects and DTOs
type FormDTOMapper struct{}

// NewFormDTOMapper creates a new DTO mapper instance
func NewFormDTOMapper() *FormDTOMapper {
	return &FormDTOMapper{}
}

// =============================================================================
// Domain to DTO Mappings
// =============================================================================

// ToFormResponseDTO converts a domain Form to FormResponseDTO
func (m *FormDTOMapper) ToFormResponseDTO(form *domain.Form) *dto.FormResponseDTO {
	if form == nil {
		return nil
	}

	return &dto.FormResponseDTO{
		ID:            form.ID.String(),
		Title:         form.Title,
		Description:   form.Description,
		Status:        string(form.Status),
		IsAnonymous:   !form.Settings.RequireAuthentication,
		IsPublic:      form.Settings.IsPublic,
		AllowMultiple: !form.Settings.RequireAuthentication, // Approximate mapping
		CreatedBy:     m.ToUserInfoDTO(form.UserID.String()),
		Settings:      m.ToFormSettingsDTO(form.Settings),
		Questions:     m.ToQuestionsResponseDTO(form.Questions),
		Tags:          []string{}, // Default empty if not available
		Category:      "",         // Default empty if not available
		Statistics:    m.ToFormStatisticsDTO(),
		CreatedAt:     form.CreatedAt,
		UpdatedAt:     form.UpdatedAt,
		PublishedAt:   form.PublishedAt,
		ExpiresAt:     form.Settings.ExpiresAt,
	}
}

// ToQuestionsResponseDTO converts domain questions to DTO
func (m *FormDTOMapper) ToQuestionsResponseDTO(questions []domain.Question) []dto.QuestionResponseDTO {
	if questions == nil {
		return nil
	}

	dtoQuestions := make([]dto.QuestionResponseDTO, len(questions))
	for i, question := range questions {
		dtoQuestions[i] = dto.QuestionResponseDTO{
			ID:          question.ID.String(),
			Type:        string(question.Type),
			Label:       question.Title,
			Description: question.Description,
			Required:    question.Required,
			Order:       question.Order,
			Options:     m.ToQuestionOptionsDTO(question.Options),
			Validation:  m.ToQuestionValidationDTO(question.Validation),
			Metadata:    question.Metadata,
			CreatedAt:   time.Now(), // Default timestamp
			UpdatedAt:   time.Now(), // Default timestamp
		}
	}
	return dtoQuestions
}

// ToQuestionOptionsDTO converts domain question options to DTO
func (m *FormDTOMapper) ToQuestionOptionsDTO(options []domain.QuestionOption) []dto.QuestionOptionDTO {
	if options == nil {
		return nil
	}

	dtoOptions := make([]dto.QuestionOptionDTO, len(options))
	for i, option := range options {
		dtoOptions[i] = dto.QuestionOptionDTO{
			ID:    option.ID.String(),
			Label: option.Label,
			Value: option.Value,
			Order: option.Order,
		}
	}
	return dtoOptions
}

// ToQuestionValidationDTO converts domain validation rules to DTO
func (m *FormDTOMapper) ToQuestionValidationDTO(validation domain.ValidationRules) *dto.QuestionValidationDTO {
	return &dto.QuestionValidationDTO{
		MinLength: validation.MinLength,
		MaxLength: validation.MaxLength,
		Pattern:   validation.Pattern,
		MinValue:  validation.MinValue,
		MaxValue:  validation.MaxValue,
		Custom:    validation.CustomRules,
	}
}

// ToFormSettingsDTO converts domain form settings to DTO
func (m *FormDTOMapper) ToFormSettingsDTO(settings domain.FormSettings) dto.FormSettingsDTO {
	return dto.FormSettingsDTO{
		RequireLogin:       settings.RequireAuthentication,
		CollectEmail:       true,  // Default
		ShowProgressBar:    true,  // Default
		AllowDrafts:        false, // Default
		NotifyOnSubmission: settings.EnableNotifications,
		CustomCSS:          "", // Default empty
		RedirectURL:        "", // Default empty
		ThankYouMessage:    "", // Default empty
		Metadata:           make(map[string]interface{}),
	}
}

// ToUserInfoDTO creates a user info DTO from user ID
func (m *FormDTOMapper) ToUserInfoDTO(userID string) dto.UserInfoDTO {
	return dto.UserInfoDTO{
		ID:       userID,
		Username: "user", // Default placeholder
		Email:    "",     // Would be populated from user service
		Name:     "",     // Would be populated from user service
		Avatar:   "",     // Would be populated from user service
	}
}

// ToFormStatisticsDTO creates default form statistics
func (m *FormDTOMapper) ToFormStatisticsDTO() dto.FormStatisticsDTO {
	return dto.FormStatisticsDTO{
		TotalResponses:   0,
		UniqueResponders: 0,
		CompletionRate:   0.0,
		AverageTime:      0,
		LastResponse:     nil,
		ResponseRate:     0.0,
	}
}

// ToFormListResponseDTO converts a slice of domain Forms to list DTO
func (m *FormDTOMapper) ToFormListResponseDTO(forms []domain.Form, pagination dto.Pagination) *dto.FormListResponseDTO {
	formDTOs := make([]dto.FormResponseDTO, len(forms))
	for i, form := range forms {
		formDTOs[i] = *m.ToFormResponseDTO(&form)
	}

	return &dto.FormListResponseDTO{
		Forms:      formDTOs,
		Pagination: pagination,
	}
}

// =============================================================================
// DTO to Domain Mappings
// =============================================================================

// ToFormDomain converts CreateFormRequestDTO to domain Form
func (m *FormDTOMapper) ToFormDomain(createDTO *dto.CreateFormRequestDTO, userID string) *domain.Form {
	if createDTO == nil {
		return nil
	}

	now := time.Now()
	form := &domain.Form{
		Title:       createDTO.Title,
		Description: createDTO.Description,
		Questions:   m.ToQuestionsDomain(createDTO.Questions),
		Settings:    m.ToFormSettingsDomain(createDTO.Settings),
		Status:      domain.FormStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    make(map[string]interface{}),
	}

	// Parse user ID
	if userUUID, err := domain.ParseUUID(userID); err == nil {
		form.UserID = userUUID
	}

	return form
}

// UpdateFormDomain updates an existing domain Form with UpdateFormRequestDTO
func (m *FormDTOMapper) UpdateFormDomain(existing *domain.Form, updateDTO *dto.UpdateFormRequestDTO, userID string) *domain.Form {
	if existing == nil || updateDTO == nil {
		return existing
	}

	// Update fields if provided
	if updateDTO.Title != nil {
		existing.Title = *updateDTO.Title
	}
	if updateDTO.Description != nil {
		existing.Description = *updateDTO.Description
	}
	if updateDTO.Questions != nil {
		existing.Questions = m.ToQuestionsDomain(*updateDTO.Questions)
	}
	if updateDTO.Settings != nil {
		existing.Settings = m.ToFormSettingsDomain(*updateDTO.Settings)
	}

	// Update metadata
	existing.UpdatedAt = time.Now()

	return existing
}

// ToQuestionsDomain converts DTO questions to domain
func (m *FormDTOMapper) ToQuestionsDomain(dtoQuestions []dto.CreateQuestionRequestDTO) []domain.Question {
	if dtoQuestions == nil {
		return nil
	}

	questions := make([]domain.Question, len(dtoQuestions))
	for i, dtoQuestion := range dtoQuestions {
		questions[i] = domain.Question{
			Type:        domain.QuestionType(dtoQuestion.Type),
			Title:       dtoQuestion.Label,
			Description: dtoQuestion.Description,
			Required:    dtoQuestion.Required,
			Options:     m.ToQuestionOptionsDomain(dtoQuestion.Options),
			Validation:  m.ToValidationRulesDomain(dtoQuestion.Validation),
			Order:       dtoQuestion.Order,
			Metadata:    dtoQuestion.Metadata,
		}
	}
	return questions
}

// ToQuestionOptionsDomain converts DTO question options to domain
func (m *FormDTOMapper) ToQuestionOptionsDomain(dtoOptions []dto.QuestionOptionDTO) []domain.QuestionOption {
	if dtoOptions == nil {
		return nil
	}

	options := make([]domain.QuestionOption, len(dtoOptions))
	for i, dtoOption := range dtoOptions {
		options[i] = domain.QuestionOption{
			Label: dtoOption.Label,
			Value: dtoOption.Value,
			Order: dtoOption.Order,
		}
	}
	return options
}

// ToValidationRulesDomain converts DTO validation rules to domain
func (m *FormDTOMapper) ToValidationRulesDomain(dtoValidation dto.QuestionValidationDTO) domain.ValidationRules {
	return domain.ValidationRules{
		MinLength:   dtoValidation.MinLength,
		MaxLength:   dtoValidation.MaxLength,
		Pattern:     dtoValidation.Pattern,
		MinValue:    dtoValidation.MinValue,
		MaxValue:    dtoValidation.MaxValue,
		CustomRules: dtoValidation.Custom,
	}
}

// ToFormSettingsDomain converts DTO form settings to domain
func (m *FormDTOMapper) ToFormSettingsDomain(dtoSettings dto.FormSettingsDTO) domain.FormSettings {
	return domain.FormSettings{
		IsPublic:              true, // Default
		AllowAnonymous:        !dtoSettings.RequireLogin,
		RequireAuthentication: dtoSettings.RequireLogin,
		EnableNotifications:   dtoSettings.NotifyOnSubmission,
		SubmissionLimit:       nil,        // Default
		ExpiresAt:             nil,        // Default
		AllowedDomains:        []string{}, // Default
	}
}

// =============================================================================
// Filter and Query Mappings
// =============================================================================

// ToFormFilter converts FormFilterRequestDTO to domain filter
func (m *FormDTOMapper) ToFormFilter(filterDTO *dto.FormFilterRequestDTO) domain.FormFilter {
	if filterDTO == nil {
		return domain.FormFilter{}
	}

	return domain.FormFilter{
		Title:    filterDTO.Title,
		Category: filterDTO.Category,
		Status:   (*domain.FormStatus)(filterDTO.Status),
		UserID:   filterDTO.CreatedBy,
	}
}

// ToPaginationQuery converts PaginationRequestDTO to domain pagination
func (m *FormDTOMapper) ToPaginationQuery(paginationDTO *dto.PaginationRequestDTO) domain.PaginationQuery {
	if paginationDTO == nil {
		return domain.PaginationQuery{
			Page:  1,
			Limit: 10,
		}
	}

	return domain.PaginationQuery{
		Page:   paginationDTO.Page,
		Limit:  paginationDTO.Limit,
		SortBy: paginationDTO.SortBy,
		Order:  paginationDTO.Order,
	}
}

// ToPaginationResponseDTO converts domain pagination result to DTO
func (m *FormDTOMapper) ToPaginationResponseDTO(page, limit, total int64) dto.Pagination {
	totalPages := (total + limit - 1) / limit

	return dto.Pagination{
		CurrentPage:  int(page),
		TotalPages:   int(totalPages),
		TotalItems:   int(total),
		ItemsPerPage: int(limit),
		HasNext:      page < totalPages,
		HasPrevious:  page > 1,
	}
}

// =============================================================================
// Error Mappings
// =============================================================================

// ToErrorResponseDTO converts domain errors to standardized error DTO
func (m *FormDTOMapper) ToErrorResponseDTO(err error, correlationID string) *dto.ErrorResponse {
	if err == nil {
		return nil
	}

	return &dto.ErrorResponse{
		BaseResponse: dto.BaseResponse{
			Success:       false,
			Message:       "Request failed",
			CorrelationID: correlationID,
			Timestamp:     time.Now(),
			Version:       "v1",
		},
		Error: dto.ErrorDetail{
			Code:      "INTERNAL_ERROR",
			Message:   err.Error(),
			Timestamp: time.Now(),
			RequestID: correlationID,
		},
	}
}

// ToValidationErrorResponseDTO converts validation errors to DTO
func (m *FormDTOMapper) ToValidationErrorResponseDTO(errors map[string]string, correlationID string) *dto.ErrorResponse {
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

// ToFormFieldsResponseDTO converts domain form fields to DTO
func (m *FormDTOMapper) ToFormFieldsResponseDTO(fields []domain.FormField) []dto.FormFieldResponseDTO {
	if fields == nil {
		return nil
	}

	dtoFields := make([]dto.FormFieldResponseDTO, len(fields))
	for i, field := range fields {
		dtoFields[i] = dto.FormFieldResponseDTO{
			ID:          field.ID,
			Label:       field.Label,
			Type:        field.Type,
			Required:    field.Required,
			Placeholder: field.Placeholder,
			Options:     field.Options,
			Validation:  m.ToValidationRulesResponseDTO(field.Validation),
			Order:       field.Order,
			Metadata:    field.Metadata,
		}
	}
	return dtoFields
}

// ToValidationRulesResponseDTO converts domain validation rules to DTO
func (m *FormDTOMapper) ToValidationRulesResponseDTO(validation domain.ValidationRules) dto.ValidationRulesResponseDTO {
	return dto.ValidationRulesResponseDTO{
		MinLength: validation.MinLength,
		MaxLength: validation.MaxLength,
		Pattern:   validation.Pattern,
		Min:       validation.Min,
		Max:       validation.Max,
		Required:  validation.Required,
		Custom:    validation.Custom,
	}
}

// ToFormSettingsResponseDTO converts domain form settings to DTO
func (m *FormDTOMapper) ToFormSettingsResponseDTO(settings domain.FormSettings) dto.FormSettingsResponseDTO {
	return dto.FormSettingsResponseDTO{
		AllowMultipleSubmissions: settings.AllowMultipleSubmissions,
		RequireAuthentication:    settings.RequireAuthentication,
		ShowProgressBar:          settings.ShowProgressBar,
		EmailNotifications:       settings.EmailNotifications,
		SaveProgress:             settings.SaveProgress,
		CustomCSS:                settings.CustomCSS,
		RedirectURL:              settings.RedirectURL,
		SubmissionLimit:          settings.SubmissionLimit,
		ExpirationDate:           settings.ExpirationDate,
	}
}

// ToFormListResponseDTO converts a slice of domain Forms to FormListResponseDTO
func (m *FormDTOMapper) ToFormListResponseDTO(forms []domain.Form, pagination dto.PaginationResponseDTO) *dto.FormListResponseDTO {
	formDTOs := make([]dto.FormResponseDTO, len(forms))
	for i, form := range forms {
		formDTOs[i] = *m.ToFormResponseDTO(&form)
	}

	return &dto.FormListResponseDTO{
		Forms:      formDTOs,
		Pagination: pagination,
	}
}

// =============================================================================
// DTO to Domain Mappings
// =============================================================================

// ToFormDomain converts CreateFormRequestDTO to domain Form
func (m *FormDTOMapper) ToFormDomain(dto *dto.CreateFormRequestDTO, userID string) *domain.Form {
	if dto == nil {
		return nil
	}

	now := time.Now()
	return &domain.Form{
		Title:       dto.Title,
		Description: dto.Description,
		Fields:      m.ToFormFieldsDomain(dto.Fields),
		Settings:    m.ToFormSettingsDomain(dto.Settings),
		IsActive:    true, // Default to active
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   userID,
		UpdatedBy:   userID,
		Version:     1, // Start with version 1
		Tags:        dto.Tags,
		Category:    dto.Category,
		Metadata:    dto.Metadata,
	}
}

// UpdateFormDomain updates an existing domain Form with UpdateFormRequestDTO
func (m *FormDTOMapper) UpdateFormDomain(existing *domain.Form, dto *dto.UpdateFormRequestDTO, userID string) *domain.Form {
	if existing == nil || dto == nil {
		return existing
	}

	// Update fields if provided
	if dto.Title != nil {
		existing.Title = *dto.Title
	}
	if dto.Description != nil {
		existing.Description = *dto.Description
	}
	if dto.Fields != nil {
		existing.Fields = m.ToFormFieldsDomain(*dto.Fields)
	}
	if dto.Settings != nil {
		existing.Settings = m.ToFormSettingsDomain(*dto.Settings)
	}
	if dto.IsActive != nil {
		existing.IsActive = *dto.IsActive
	}
	if dto.Tags != nil {
		existing.Tags = *dto.Tags
	}
	if dto.Category != nil {
		existing.Category = *dto.Category
	}
	if dto.Metadata != nil {
		existing.Metadata = *dto.Metadata
	}

	// Update metadata
	existing.UpdatedAt = time.Now()
	existing.UpdatedBy = userID
	existing.Version++ // Increment version

	return existing
}

// ToFormFieldsDomain converts DTO form fields to domain
func (m *FormDTOMapper) ToFormFieldsDomain(dtoFields []dto.FormFieldRequestDTO) []domain.FormField {
	if dtoFields == nil {
		return nil
	}

	fields := make([]domain.FormField, len(dtoFields))
	for i, dtoField := range dtoFields {
		fields[i] = domain.FormField{
			Label:       dtoField.Label,
			Type:        dtoField.Type,
			Required:    dtoField.Required,
			Placeholder: dtoField.Placeholder,
			Options:     dtoField.Options,
			Validation:  m.ToValidationRulesDomain(dtoField.Validation),
			Order:       dtoField.Order,
			Metadata:    dtoField.Metadata,
		}
	}
	return fields
}

// ToValidationRulesDomain converts DTO validation rules to domain
func (m *FormDTOMapper) ToValidationRulesDomain(dtoValidation dto.ValidationRulesRequestDTO) domain.ValidationRules {
	return domain.ValidationRules{
		MinLength: dtoValidation.MinLength,
		MaxLength: dtoValidation.MaxLength,
		Pattern:   dtoValidation.Pattern,
		Min:       dtoValidation.Min,
		Max:       dtoValidation.Max,
		Required:  dtoValidation.Required,
		Custom:    dtoValidation.Custom,
	}
}

// ToFormSettingsDomain converts DTO form settings to domain
func (m *FormDTOMapper) ToFormSettingsDomain(dtoSettings dto.FormSettingsRequestDTO) domain.FormSettings {
	return domain.FormSettings{
		AllowMultipleSubmissions: dtoSettings.AllowMultipleSubmissions,
		RequireAuthentication:    dtoSettings.RequireAuthentication,
		ShowProgressBar:          dtoSettings.ShowProgressBar,
		EmailNotifications:       dtoSettings.EmailNotifications,
		SaveProgress:             dtoSettings.SaveProgress,
		CustomCSS:                dtoSettings.CustomCSS,
		RedirectURL:              dtoSettings.RedirectURL,
		SubmissionLimit:          dtoSettings.SubmissionLimit,
		ExpirationDate:           dtoSettings.ExpirationDate,
	}
}

// =============================================================================
// Filter and Query Mappings
// =============================================================================

// ToFormFilter converts FormFilterRequestDTO to domain filter
func (m *FormDTOMapper) ToFormFilter(dto *dto.FormFilterRequestDTO) domain.FormFilter {
	if dto == nil {
		return domain.FormFilter{}
	}

	return domain.FormFilter{
		Title:         dto.Title,
		Category:      dto.Category,
		IsActive:      dto.IsActive,
		CreatedBy:     dto.CreatedBy,
		Tags:          dto.Tags,
		CreatedAfter:  dto.CreatedAfter,
		CreatedBefore: dto.CreatedBefore,
	}
}

// ToPaginationQuery converts PaginationRequestDTO to domain pagination
func (m *FormDTOMapper) ToPaginationQuery(dto *dto.PaginationRequestDTO) domain.PaginationQuery {
	if dto == nil {
		return domain.PaginationQuery{
			Page:  1,
			Limit: 10,
		}
	}

	return domain.PaginationQuery{
		Page:   dto.Page,
		Limit:  dto.Limit,
		SortBy: dto.SortBy,
		Order:  dto.Order,
	}
}

// ToPaginationResponseDTO converts domain pagination result to DTO
func (m *FormDTOMapper) ToPaginationResponseDTO(page, limit, total int64) dto.PaginationResponseDTO {
	totalPages := (total + limit - 1) / limit

	return dto.PaginationResponseDTO{
		CurrentPage:  page,
		TotalPages:   totalPages,
		TotalItems:   total,
		ItemsPerPage: limit,
		HasNext:      page < totalPages,
		HasPrevious:  page > 1,
	}
}

// =============================================================================
// Error Mappings
// =============================================================================

// ToErrorResponseDTO converts domain errors to standardized error DTO
func (m *FormDTOMapper) ToErrorResponseDTO(err error, correlationID string) *dto.ErrorResponseDTO {
	if err == nil {
		return nil
	}

	return &dto.ErrorResponseDTO{
		Code:          "INTERNAL_ERROR",
		Message:       err.Error(),
		Timestamp:     time.Now(),
		CorrelationID: correlationID,
		Details:       nil,
	}
}

// ToValidationErrorResponseDTO converts validation errors to DTO
func (m *FormDTOMapper) ToValidationErrorResponseDTO(errors map[string]string, correlationID string) *dto.ErrorResponseDTO {
	details := make(map[string]interface{})
	for field, message := range errors {
		details[field] = message
	}

	return &dto.ErrorResponseDTO{
		Code:          "VALIDATION_ERROR",
		Message:       "Validation failed",
		Timestamp:     time.Now(),
		CorrelationID: correlationID,
		Details:       details,
	}
}

// =============================================================================
// Health Check Mappings
// =============================================================================

// ToHealthResponseDTO converts health status to DTO
func (m *FormDTOMapper) ToHealthResponseDTO(status string, checks map[string]interface{}) *dto.HealthResponseDTO {
	return &dto.HealthResponseDTO{
		Status:    status,
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Checks:    checks,
	}
}
