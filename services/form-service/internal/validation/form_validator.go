// Package validation contains input validation utilities for the Form Service
// Following microservices best practices for security and data integrity
package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/dto"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/handlers"
)

// FormValidator handles form-related validation
type FormValidator struct {
	validator       *validator.Validate
	responseHandler *handlers.ResponseHandler
}

// NewFormValidator creates a new form validator instance
func NewFormValidator(responseHandler *handlers.ResponseHandler) *FormValidator {
	v := validator.New()

	// Register custom validators
	registerCustomValidators(v)

	return &FormValidator{
		validator:       v,
		responseHandler: responseHandler,
	}
}

// =============================================================================
// Custom Validators
// =============================================================================

// registerCustomValidators registers custom validation rules
func registerCustomValidators(v *validator.Validate) {
	// Question type validator
	v.RegisterValidation("question_type", validateQuestionType)

	// Form status validator
	v.RegisterValidation("form_status", validateFormStatus)

	// Sort field validator
	v.RegisterValidation("sort_field", validateSortField)

	// UUID validator
	v.RegisterValidation("uuid", validateUUID)

	// Safe string validator (prevents XSS)
	v.RegisterValidation("safe_string", validateSafeString)

	// File type validator
	v.RegisterValidation("file_type", validateFileType)
}

// validateQuestionType validates question types
func validateQuestionType(fl validator.FieldLevel) bool {
	validTypes := []string{
		"text", "textarea", "number", "email", "date", "datetime",
		"checkbox", "radio", "select", "multiselect", "file", "url",
		"tel", "password", "hidden", "rating", "range", "color",
	}

	questionType := fl.Field().String()
	for _, validType := range validTypes {
		if questionType == validType {
			return true
		}
	}
	return false
}

// validateFormStatus validates form status values
func validateFormStatus(fl validator.FieldLevel) bool {
	validStatuses := []string{"draft", "published", "closed", "archived"}

	status := fl.Field().String()
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// validateSortField validates sort field values
func validateSortField(fl validator.FieldLevel) bool {
	validFields := []string{
		"created_at", "updated_at", "title", "response_count",
		"published_at", "expires_at", "status",
	}

	field := fl.Field().String()
	for _, validField := range validFields {
		if field == validField {
			return true
		}
	}
	return false
}

// validateUUID validates UUID format
func validateUUID(fl validator.FieldLevel) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(fl.Field().String())
}

// validateSafeString validates string for XSS prevention
func validateSafeString(fl validator.FieldLevel) bool {
	str := fl.Field().String()

	// Check for dangerous patterns
	dangerousPatterns := []string{
		"<script", "</script", "javascript:", "onload=", "onerror=",
		"onclick=", "onmouseover=", "onfocus=", "onblur=", "onchange=",
		"onsubmit=", "onreset=", "onselect=", "onkeydown=", "onkeyup=",
		"data:text/html", "vbscript:", "expression(", "url(javascript",
	}

	lowerStr := strings.ToLower(str)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerStr, pattern) {
			return false
		}
	}

	return true
}

// validateFileType validates file types
func validateFileType(fl validator.FieldLevel) bool {
	validTypes := []string{
		"jpg", "jpeg", "png", "gif", "bmp", "webp", "svg",
		"pdf", "doc", "docx", "txt", "rtf", "odt",
		"xls", "xlsx", "csv", "ods",
		"ppt", "pptx", "odp",
		"zip", "rar", "7z", "tar", "gz",
		"mp3", "wav", "ogg", "m4a",
		"mp4", "avi", "mov", "wmv", "webm",
	}

	fileType := strings.ToLower(fl.Field().String())
	for _, validType := range validTypes {
		if fileType == validType {
			return true
		}
	}
	return false
}

// =============================================================================
// Validation Methods
// =============================================================================

// ValidateCreateFormRequest validates form creation request
func (fv *FormValidator) ValidateCreateFormRequest(c *gin.Context, req *dto.CreateFormRequestDTO) bool {
	if err := fv.validator.Struct(req); err != nil {
		fv.handleValidationError(c, err)
		return false
	}

	// Additional business logic validation
	if !fv.validateBusinessRules(c, req) {
		return false
	}

	return true
}

// ValidateUpdateFormRequest validates form update request
func (fv *FormValidator) ValidateUpdateFormRequest(c *gin.Context, req *dto.UpdateFormRequestDTO) bool {
	if err := fv.validator.Struct(req); err != nil {
		fv.handleValidationError(c, err)
		return false
	}

	return true
}

// ValidateFormListRequest validates form listing request
func (fv *FormValidator) ValidateFormListRequest(c *gin.Context, req *dto.FormListRequestDTO) bool {
	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}

	if err := fv.validator.Struct(req); err != nil {
		fv.handleValidationError(c, err)
		return false
	}

	return true
}

// ValidatePublishFormRequest validates form publishing request
func (fv *FormValidator) ValidatePublishFormRequest(c *gin.Context, req *dto.PublishFormRequestDTO) bool {
	if err := fv.validator.Struct(req); err != nil {
		fv.handleValidationError(c, err)
		return false
	}

	return true
}

// ValidateFormID validates form ID parameter
func (fv *FormValidator) ValidateFormID(c *gin.Context, formID string) bool {
	if formID == "" {
		fv.responseHandler.BadRequest(c, "Form ID is required")
		return false
	}

	// Validate UUID format
	if !fv.isValidUUID(formID) {
		fv.responseHandler.BadRequest(c, "Invalid form ID format")
		return false
	}

	return true
}

// ValidateUserID validates user ID parameter
func (fv *FormValidator) ValidateUserID(c *gin.Context, userID string) bool {
	if userID == "" {
		fv.responseHandler.Unauthorized(c, "User ID is required")
		return false
	}

	// Validate UUID format
	if !fv.isValidUUID(userID) {
		fv.responseHandler.BadRequest(c, "Invalid user ID format")
		return false
	}

	return true
}

// =============================================================================
// Business Logic Validation
// =============================================================================

// validateBusinessRules validates business-specific rules
func (fv *FormValidator) validateBusinessRules(c *gin.Context, req *dto.CreateFormRequestDTO) bool {
	errors := make(map[string][]string)

	// Validate questions
	if len(req.Questions) == 0 {
		errors["questions"] = append(errors["questions"], "At least one question is required")
	}

	// Validate question orders are unique
	orders := make(map[int]bool)
	for i, question := range req.Questions {
		if orders[question.Order] {
			errors[fmt.Sprintf("questions[%d].order", i)] = append(
				errors[fmt.Sprintf("questions[%d].order", i)],
				"Question order must be unique",
			)
		}
		orders[question.Order] = true
	}

	// Validate select/radio questions have options
	for i, question := range req.Questions {
		if (question.Type == "select" || question.Type == "radio" || question.Type == "checkbox") && len(question.Options) == 0 {
			errors[fmt.Sprintf("questions[%d].options", i)] = append(
				errors[fmt.Sprintf("questions[%d].options", i)],
				"Select, radio, and checkbox questions must have options",
			)
		}
	}

	// Validate file questions have allowed types
	for i, question := range req.Questions {
		if question.Type == "file" && question.Validation != nil && len(question.Validation.AllowedTypes) == 0 {
			errors[fmt.Sprintf("questions[%d].validation.allowedTypes", i)] = append(
				errors[fmt.Sprintf("questions[%d].validation.allowedTypes", i)],
				"File questions must specify allowed file types",
			)
		}
	}

	// Validate expiration date
	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		errors["expiresAt"] = append(errors["expiresAt"], "Expiration date cannot be in the past")
	}

	// Validate tags
	if len(req.Tags) > 10 {
		errors["tags"] = append(errors["tags"], "Maximum 10 tags allowed")
	}

	for i, tag := range req.Tags {
		if len(tag) > 50 {
			errors[fmt.Sprintf("tags[%d]", i)] = append(
				errors[fmt.Sprintf("tags[%d]", i)],
				"Tag length cannot exceed 50 characters",
			)
		}
	}

	if len(errors) > 0 {
		fv.responseHandler.ValidationError(c, errors, "Business rule validation failed")
		return false
	}

	return true
}

// =============================================================================
// Security Validation
// =============================================================================

// ValidateRequestSecurity performs security validation on requests
func (fv *FormValidator) ValidateRequestSecurity(c *gin.Context) bool {
	// Check Content-Type for POST/PUT requests
	if c.Request.Method == "POST" || c.Request.Method == "PUT" {
		contentType := c.GetHeader("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			fv.responseHandler.BadRequest(c, "Content-Type must be application/json")
			return false
		}
	}

	// Check Content-Length
	if c.Request.ContentLength > 10*1024*1024 { // 10MB limit
		fv.responseHandler.BadRequest(c, "Request body too large")
		return false
	}

	// Validate User-Agent header
	userAgent := c.GetHeader("User-Agent")
	if userAgent == "" {
		fv.responseHandler.BadRequest(c, "User-Agent header is required")
		return false
	}

	// Check for suspicious patterns in headers
	for key, values := range c.Request.Header {
		for _, value := range values {
			if !fv.isSafeString(value) {
				fv.responseHandler.BadRequest(c, fmt.Sprintf("Invalid characters in header: %s", key))
				return false
			}
		}
	}

	return true
}

// =============================================================================
// Helper Methods
// =============================================================================

// handleValidationError processes validation errors and sends appropriate response
func (fv *FormValidator) handleValidationError(c *gin.Context, err error) {
	errors := make(map[string][]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			field := fv.getFieldName(fieldError)
			message := fv.getErrorMessage(fieldError)
			errors[field] = append(errors[field], message)
		}
	} else {
		errors["general"] = append(errors["general"], err.Error())
	}

	fv.responseHandler.ValidationError(c, errors)
}

// getFieldName extracts field name from validation error
func (fv *FormValidator) getFieldName(fieldError validator.FieldError) string {
	// Convert struct field names to JSON field names
	fieldName := fieldError.Field()

	// Convert camelCase to snake_case for consistency
	var result strings.Builder
	for i, r := range fieldName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}

// getErrorMessage generates user-friendly error messages
func (fv *FormValidator) getErrorMessage(fieldError validator.FieldError) string {
	field := fieldError.Field()

	switch fieldError.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, fieldError.Param())
	case "max":
		return fmt.Sprintf("%s cannot exceed %s characters", field, fieldError.Param())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fieldError.Param())
	case "question_type":
		return fmt.Sprintf("%s must be a valid question type", field)
	case "form_status":
		return fmt.Sprintf("%s must be a valid form status", field)
	case "safe_string":
		return fmt.Sprintf("%s contains invalid characters", field)
	case "file_type":
		return fmt.Sprintf("%s must be a valid file type", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// isValidUUID checks if string is a valid UUID
func (fv *FormValidator) isValidUUID(str string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(str)
}

// isSafeString checks if string is safe (no XSS patterns)
func (fv *FormValidator) isSafeString(str string) bool {
	dangerousPatterns := []string{
		"<script", "</script", "javascript:", "onload=", "onerror=",
		"onclick=", "onmouseover=", "onfocus=", "onblur=", "onchange=",
		"onsubmit=", "onreset=", "onselect=", "onkeydown=", "onkeyup=",
		"data:text/html", "vbscript:", "expression(", "url(javascript",
	}

	lowerStr := strings.ToLower(str)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerStr, pattern) {
			return false
		}
	}

	return true
}

// =============================================================================
// Middleware
// =============================================================================

// SecurityValidationMiddleware validates request security
func (fv *FormValidator) SecurityValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !fv.ValidateRequestSecurity(c) {
			c.Abort()
			return
		}
		c.Next()
	}
}
