// Package handlers contains HTTP handlers for the form service
// This layer handles HTTP concerns and translates between HTTP and domain models
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/application"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/domain"
)

// FormHandler handles HTTP requests for form operations
// Follows Single Responsibility Principle by only handling HTTP concerns
type FormHandler struct {
	formService *application.FormApplicationService
}

// NewFormHandler creates a new form handler instance
// Uses Dependency Injection for better testability
func NewFormHandler(formService *application.FormApplicationService) *FormHandler {
	return &FormHandler{
		formService: formService,
	}
}

// RegisterRoutes registers all form-related routes
// Organizes routes in a clean, RESTful manner
func (h *FormHandler) RegisterRoutes(router *gin.RouterGroup) {
	forms := router.Group("/forms")
	{
		forms.POST("", h.CreateForm)
		forms.GET("", h.ListUserForms)
		forms.GET("/:id", h.GetForm)
		forms.PUT("/:id", h.UpdateForm)
		forms.DELETE("/:id", h.DeleteForm)
		forms.POST("/:id/publish", h.PublishForm)
		forms.POST("/:id/close", h.CloseForm)
	}

	// Public routes
	public := router.Group("/public/forms")
	{
		public.GET("/:id", h.GetPublicForm)
	}
}

// HTTP Response structures

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
}

// ErrorDetail contains detailed error information
type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination contains pagination metadata
type Pagination struct {
	Offset int   `json:"offset"`
	Limit  int   `json:"limit"`
	Total  int64 `json:"total"`
}

// HTTP Handlers

// CreateForm handles form creation requests
func (h *FormHandler) CreateForm(c *gin.Context) {
	userID, err := h.extractUserID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	var req domain.CreateFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	form, err := h.formService.CreateForm(c.Request.Context(), userID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.sendSuccessResponse(c, http.StatusCreated, "Form created successfully", form)
}

// GetForm handles form retrieval requests
func (h *FormHandler) GetForm(c *gin.Context) {
	formID, err := h.extractFormID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	userID, _ := h.extractOptionalUserID(c)

	form, err := h.formService.GetForm(c.Request.Context(), formID, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.sendSuccessResponse(c, http.StatusOK, "Form retrieved successfully", form)
}

// UpdateForm handles form update requests
func (h *FormHandler) UpdateForm(c *gin.Context) {
	formID, err := h.extractFormID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	userID, err := h.extractUserID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	var req domain.UpdateFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	form, err := h.formService.UpdateForm(c.Request.Context(), formID, userID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.sendSuccessResponse(c, http.StatusOK, "Form updated successfully", form)
}

// DeleteForm handles form deletion requests
func (h *FormHandler) DeleteForm(c *gin.Context) {
	formID, err := h.extractFormID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	userID, err := h.extractUserID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	err = h.formService.DeleteForm(c.Request.Context(), formID, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.sendSuccessResponse(c, http.StatusOK, "Form deleted successfully", nil)
}

// ListUserForms handles user forms listing requests
func (h *FormHandler) ListUserForms(c *gin.Context) {
	userID, err := h.extractUserID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	filters, err := h.extractFormFilters(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	forms, total, err := h.formService.ListUserForms(c.Request.Context(), userID, filters)
	if err != nil {
		h.handleError(c, err)
		return
	}

	pagination := Pagination{
		Offset: filters.Offset,
		Limit:  filters.Limit,
		Total:  total,
	}

	h.sendPaginatedResponse(c, http.StatusOK, "Forms retrieved successfully", forms, pagination)
}

// PublishForm handles form publishing requests
func (h *FormHandler) PublishForm(c *gin.Context) {
	formID, err := h.extractFormID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	userID, err := h.extractUserID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	err = h.formService.PublishForm(c.Request.Context(), formID, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.sendSuccessResponse(c, http.StatusOK, "Form published successfully", nil)
}

// CloseForm handles form closing requests
func (h *FormHandler) CloseForm(c *gin.Context) {
	formID, err := h.extractFormID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	userID, err := h.extractUserID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	err = h.formService.CloseForm(c.Request.Context(), formID, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.sendSuccessResponse(c, http.StatusOK, "Form closed successfully", nil)
}

// GetPublicForm handles public form retrieval requests
func (h *FormHandler) GetPublicForm(c *gin.Context) {
	formID, err := h.extractFormID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	form, err := h.formService.GetPublicForm(c.Request.Context(), formID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.sendSuccessResponse(c, http.StatusOK, "Public form retrieved successfully", form)
}

// Helper methods

// extractUserID extracts user ID from request context
func (h *FormHandler) extractUserID(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, domain.NewAccessDeniedError("user", "authenticate", "", "user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return uuid.Nil, domain.NewValidationError("user_id", "invalid user ID format")
	}

	return userID, nil
}

// extractOptionalUserID extracts user ID from request context if present
func (h *FormHandler) extractOptionalUserID(c *gin.Context) (*uuid.UUID, error) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return nil, nil
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return nil, domain.NewValidationError("user_id", "invalid user ID format")
	}

	return &userID, nil
}

// extractFormID extracts form ID from URL parameters
func (h *FormHandler) extractFormID(c *gin.Context) (uuid.UUID, error) {
	formIDStr := c.Param("id")
	if formIDStr == "" {
		return uuid.Nil, domain.NewValidationError("id", "form ID is required")
	}

	formID, err := uuid.Parse(formIDStr)
	if err != nil {
		return uuid.Nil, domain.NewValidationError("id", "invalid form ID format")
	}

	return formID, nil
}

// extractFormFilters extracts form filters from query parameters
func (h *FormHandler) extractFormFilters(c *gin.Context) (domain.FormFilters, error) {
	filters := domain.FormFilters{
		Offset: 0,
		Limit:  20, // Default limit
	}

	// Extract offset
	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return filters, domain.NewValidationError("offset", "invalid offset value")
		}
		filters.Offset = offset
	}

	// Extract limit
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 100 {
			return filters, domain.NewValidationError("limit", "invalid limit value (max 100)")
		}
		filters.Limit = limit
	}

	// Extract search
	if search := c.Query("search"); search != "" {
		filters.Search = &search
	}

	// Extract status
	if statusStr := c.Query("status"); statusStr != "" {
		status := domain.FormStatus(statusStr)
		filters.Status = &status
	}

	return filters, nil
}

// Response helpers

// sendSuccessResponse sends a successful response
func (h *FormHandler) sendSuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// sendPaginatedResponse sends a paginated response
func (h *FormHandler) sendPaginatedResponse(c *gin.Context, statusCode int, message string, data interface{}, pagination Pagination) {
	c.JSON(statusCode, PaginatedResponse{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	})
}

// Error handling

// handleError handles different types of domain errors
func (h *FormHandler) handleError(c *gin.Context, err error) {
	var statusCode int
	var errorCode string
	var message string

	// Map domain errors to HTTP status codes (follows Open/Closed Principle)
	switch {
	case domain.IsValidationError(err):
		statusCode = http.StatusBadRequest
		errorCode = "VALIDATION_ERROR"
		message = err.Error()

	case domain.IsNotFoundError(err):
		statusCode = http.StatusNotFound
		errorCode = "NOT_FOUND"
		message = err.Error()

	case domain.IsAccessDeniedError(err):
		statusCode = http.StatusForbidden
		errorCode = "ACCESS_DENIED"
		message = err.Error()

	case domain.IsBusinessRuleError(err):
		statusCode = http.StatusConflict
		errorCode = "BUSINESS_RULE_VIOLATION"
		message = err.Error()

	case domain.IsConflictError(err):
		statusCode = http.StatusConflict
		errorCode = "CONFLICT"
		message = err.Error()

	default:
		// Internal server error for unknown errors
		statusCode = http.StatusInternalServerError
		errorCode = "INTERNAL_ERROR"
		message = "An internal error occurred"

		// TODO: Log the actual error for debugging
		// h.logger.WithError(err).Error("Unhandled error in form handler")
	}

	h.sendErrorResponse(c, statusCode, errorCode, message, nil)
}

// handleValidationError handles request validation errors
func (h *FormHandler) handleValidationError(c *gin.Context, err error) {
	h.sendErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Request validation failed", err.Error())
}

// sendErrorResponse sends an error response
func (h *FormHandler) sendErrorResponse(c *gin.Context, statusCode int, code, message string, details interface{}) {
	c.JSON(statusCode, ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck handles health check requests
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "form-service",
		"timestamp": "2024-08-24T14:44:05Z", // TODO: Use actual timestamp
	})
}
