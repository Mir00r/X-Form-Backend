// Package routes contains versioned API routes for the Form Service
// Following microservices best practices with proper versioning and documentation
package routes

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/application"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/dto"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/handlers"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/validation"
)

// FormRoutesV1 handles version 1 of the Form API
type FormRoutesV1 struct {
	formService     *application.FormApplicationService
	responseHandler *handlers.ResponseHandler
	validator       *validation.FormValidator
}

// NewFormRoutesV1 creates a new instance of version 1 form routes
func NewFormRoutesV1(
	formService *application.FormApplicationService,
	responseHandler *handlers.ResponseHandler,
	validator *validation.FormValidator,
) *FormRoutesV1 {
	return &FormRoutesV1{
		formService:     formService,
		responseHandler: responseHandler,
		validator:       validator,
	}
}

// RegisterRoutes registers all version 1 form routes
func (r *FormRoutesV1) RegisterRoutes(router *gin.RouterGroup) {
	// Form management routes
	r.registerFormRoutes(router)

	// Health and monitoring routes
	r.registerHealthRoutes(router)
}

// registerFormRoutes registers form-related routes
func (r *FormRoutesV1) registerFormRoutes(router *gin.RouterGroup) {
	forms := router.Group("/forms")
	{
		// Create form
		forms.POST("", r.CreateForm)

		// List forms with filtering and pagination
		forms.GET("", r.ListForms)

		// Get specific form
		forms.GET("/:id", r.GetForm)

		// Update form
		forms.PUT("/:id", r.UpdateForm)

		// Delete form
		forms.DELETE("/:id", r.DeleteForm)

		// Publish form
		forms.POST("/:id/publish", r.PublishForm)

		// Close form
		forms.POST("/:id/close", r.CloseForm)

		// Archive form
		forms.POST("/:id/archive", r.ArchiveForm)

		// Get form statistics
		forms.GET("/:id/statistics", r.GetFormStatistics)

		// Duplicate form
		forms.POST("/:id/duplicate", r.DuplicateForm)
	}

	// Public form routes (no authentication required)
	public := router.Group("/public/forms")
	{
		public.GET("/:id", r.GetPublicForm)
	}
}

// registerHealthRoutes registers health and monitoring routes
func (r *FormRoutesV1) registerHealthRoutes(router *gin.RouterGroup) {
	health := router.Group("/health")
	{
		health.GET("", r.HealthCheck)
		health.GET("/ready", r.ReadinessCheck)
		health.GET("/live", r.LivenessCheck)
	}

	// Metrics endpoint
	router.GET("/metrics", r.GetMetrics)
}

// =============================================================================
// Form Management Handlers
// =============================================================================

// CreateForm creates a new form
// @Summary Create a new form
// @Description Create a new form with questions and settings
// @Tags Forms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateFormRequestDTO true "Form creation request"
// @Success 201 {object} dto.SuccessResponse{data=dto.FormResponseDTO} "Form created successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 422 {object} dto.ValidationErrorDTO "Validation failed"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /forms [post]
func (r *FormRoutesV1) CreateForm(c *gin.Context) {
	// Extract user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		r.responseHandler.Unauthorized(c, "User authentication required")
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		r.responseHandler.Unauthorized(c, "Invalid user context")
		return
	}

	// Validate user ID
	if !r.validator.ValidateUserID(c, userIDStr) {
		return
	}

	// Parse and validate request
	var req dto.CreateFormRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		r.responseHandler.BadRequest(c, "Invalid JSON format", err.Error())
		return
	}

	// Validate request
	if !r.validator.ValidateCreateFormRequest(c, &req) {
		return
	}

	// Create form through application service
	form, err := r.formService.CreateForm(c.Request.Context(), userIDStr, &req)
	if err != nil {
		r.handleApplicationError(c, err)
		return
	}

	r.responseHandler.Created(c, form, "Form created successfully")
}

// ListForms lists forms with filtering and pagination
// @Summary List forms
// @Description Get a paginated list of forms with optional filtering
// @Tags Forms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Param status query string false "Filter by status" Enums(draft,published,closed,archived)
// @Param category query string false "Filter by category"
// @Param search query string false "Search in title and description"
// @Param sortBy query string false "Sort field" Enums(created_at,updated_at,title,response_count) default(created_at)
// @Param sortOrder query string false "Sort order" Enums(asc,desc) default(desc)
// @Param createdBy query string false "Filter by creator user ID"
// @Success 200 {object} dto.PaginatedResponse{data=[]dto.FormSummaryDTO} "Forms retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request parameters"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /forms [get]
func (r *FormRoutesV1) ListForms(c *gin.Context) {
	// Extract user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		r.responseHandler.Unauthorized(c, "User authentication required")
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		r.responseHandler.Unauthorized(c, "Invalid user context")
		return
	}

	// Parse query parameters
	req := r.parseListFormsRequest(c)

	// Validate request
	if !r.validator.ValidateFormListRequest(c, &req) {
		return
	}

	// List forms through application service
	forms, pagination, err := r.formService.ListForms(c.Request.Context(), userIDStr, &req)
	if err != nil {
		r.handleApplicationError(c, err)
		return
	}

	r.responseHandler.Paginated(c, forms, pagination, "Forms retrieved successfully")
}

// GetForm retrieves a specific form by ID
// @Summary Get form by ID
// @Description Retrieve a form with all its details
// @Tags Forms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Form ID" format(uuid)
// @Success 200 {object} dto.SuccessResponse{data=dto.FormResponseDTO} "Form retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid form ID"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Form not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /forms/{id} [get]
func (r *FormRoutesV1) GetForm(c *gin.Context) {
	// Extract and validate form ID
	formID := c.Param("id")
	if !r.validator.ValidateFormID(c, formID) {
		return
	}

	// Extract user ID (optional for public forms)
	userID, _ := c.Get("userID")
	userIDStr := ""
	if userID != nil {
		if id, ok := userID.(string); ok {
			userIDStr = id
		}
	}

	// Get form through application service
	form, err := r.formService.GetForm(c.Request.Context(), formID, userIDStr)
	if err != nil {
		r.handleApplicationError(c, err)
		return
	}

	r.responseHandler.Success(c, form, "Form retrieved successfully")
}

// UpdateForm updates an existing form
// @Summary Update form
// @Description Update an existing form's details
// @Tags Forms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Form ID" format(uuid)
// @Param request body dto.UpdateFormRequestDTO true "Form update request"
// @Success 200 {object} dto.SuccessResponse{data=dto.FormResponseDTO} "Form updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - not form owner"
// @Failure 404 {object} dto.ErrorResponse "Form not found"
// @Failure 422 {object} dto.ValidationErrorDTO "Validation failed"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /forms/{id} [put]
func (r *FormRoutesV1) UpdateForm(c *gin.Context) {
	// Extract and validate form ID
	formID := c.Param("id")
	if !r.validator.ValidateFormID(c, formID) {
		return
	}

	// Extract user ID
	userID, exists := c.Get("userID")
	if !exists {
		r.responseHandler.Unauthorized(c, "User authentication required")
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		r.responseHandler.Unauthorized(c, "Invalid user context")
		return
	}

	// Parse and validate request
	var req dto.UpdateFormRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		r.responseHandler.BadRequest(c, "Invalid JSON format", err.Error())
		return
	}

	// Validate request
	if !r.validator.ValidateUpdateFormRequest(c, &req) {
		return
	}

	// Update form through application service
	form, err := r.formService.UpdateForm(c.Request.Context(), formID, userIDStr, &req)
	if err != nil {
		r.handleApplicationError(c, err)
		return
	}

	r.responseHandler.Updated(c, form, "Form updated successfully")
}

// DeleteForm deletes a form
// @Summary Delete form
// @Description Delete a form and all its responses
// @Tags Forms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Form ID" format(uuid)
// @Success 200 {object} dto.SuccessResponse "Form deleted successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid form ID"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - not form owner"
// @Failure 404 {object} dto.ErrorResponse "Form not found"
// @Failure 409 {object} dto.ErrorResponse "Cannot delete published form"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /forms/{id} [delete]
func (r *FormRoutesV1) DeleteForm(c *gin.Context) {
	// Extract and validate form ID
	formID := c.Param("id")
	if !r.validator.ValidateFormID(c, formID) {
		return
	}

	// Extract user ID
	userID, exists := c.Get("userID")
	if !exists {
		r.responseHandler.Unauthorized(c, "User authentication required")
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		r.responseHandler.Unauthorized(c, "Invalid user context")
		return
	}

	// Delete form through application service
	err := r.formService.DeleteForm(c.Request.Context(), formID, userIDStr)
	if err != nil {
		r.handleApplicationError(c, err)
		return
	}

	r.responseHandler.Deleted(c, "Form deleted successfully")
}

// PublishForm publishes a form
// @Summary Publish form
// @Description Publish a form to make it available for responses
// @Tags Forms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Form ID" format(uuid)
// @Param request body dto.PublishFormRequestDTO false "Publish form request"
// @Success 200 {object} dto.SuccessResponse{data=dto.FormResponseDTO} "Form published successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - not form owner"
// @Failure 404 {object} dto.ErrorResponse "Form not found"
// @Failure 409 {object} dto.ErrorResponse "Form already published"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /forms/{id}/publish [post]
func (r *FormRoutesV1) PublishForm(c *gin.Context) {
	// Extract and validate form ID
	formID := c.Param("id")
	if !r.validator.ValidateFormID(c, formID) {
		return
	}

	// Extract user ID
	userID, exists := c.Get("userID")
	if !exists {
		r.responseHandler.Unauthorized(c, "User authentication required")
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		r.responseHandler.Unauthorized(c, "Invalid user context")
		return
	}

	// Parse request (optional)
	var req dto.PublishFormRequestDTO
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			r.responseHandler.BadRequest(c, "Invalid JSON format", err.Error())
			return
		}

		// Validate request
		if !r.validator.ValidatePublishFormRequest(c, &req) {
			return
		}
	}

	// Publish form through application service
	form, err := r.formService.PublishForm(c.Request.Context(), formID, userIDStr, &req)
	if err != nil {
		r.handleApplicationError(c, err)
		return
	}

	r.responseHandler.Success(c, form, "Form published successfully")
}

// CloseForm closes a form
// @Summary Close form
// @Description Close a form to stop accepting new responses
// @Tags Forms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Form ID" format(uuid)
// @Success 200 {object} dto.SuccessResponse{data=dto.FormResponseDTO} "Form closed successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid form ID"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - not form owner"
// @Failure 404 {object} dto.ErrorResponse "Form not found"
// @Failure 409 {object} dto.ErrorResponse "Form not published"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /forms/{id}/close [post]
func (r *FormRoutesV1) CloseForm(c *gin.Context) {
	formID := c.Param("id")
	if !r.validator.ValidateFormID(c, formID) {
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		r.responseHandler.Unauthorized(c, "User authentication required")
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		r.responseHandler.Unauthorized(c, "Invalid user context")
		return
	}

	form, err := r.formService.CloseForm(c.Request.Context(), formID, userIDStr)
	if err != nil {
		r.handleApplicationError(c, err)
		return
	}

	r.responseHandler.Success(c, form, "Form closed successfully")
}

// ArchiveForm archives a form
// @Summary Archive form
// @Description Archive a form for long-term storage
// @Tags Forms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Form ID" format(uuid)
// @Success 200 {object} dto.SuccessResponse{data=dto.FormResponseDTO} "Form archived successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid form ID"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - not form owner"
// @Failure 404 {object} dto.ErrorResponse "Form not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /forms/{id}/archive [post]
func (r *FormRoutesV1) ArchiveForm(c *gin.Context) {
	formID := c.Param("id")
	if !r.validator.ValidateFormID(c, formID) {
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		r.responseHandler.Unauthorized(c, "User authentication required")
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		r.responseHandler.Unauthorized(c, "Invalid user context")
		return
	}

	form, err := r.formService.ArchiveForm(c.Request.Context(), formID, userIDStr)
	if err != nil {
		r.handleApplicationError(c, err)
		return
	}

	r.responseHandler.Success(c, form, "Form archived successfully")
}

// GetFormStatistics retrieves form statistics
// @Summary Get form statistics
// @Description Get detailed statistics for a form
// @Tags Forms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Form ID" format(uuid)
// @Success 200 {object} dto.SuccessResponse{data=dto.FormStatisticsDTO} "Statistics retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid form ID"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - not form owner"
// @Failure 404 {object} dto.ErrorResponse "Form not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /forms/{id}/statistics [get]
func (r *FormRoutesV1) GetFormStatistics(c *gin.Context) {
	formID := c.Param("id")
	if !r.validator.ValidateFormID(c, formID) {
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		r.responseHandler.Unauthorized(c, "User authentication required")
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		r.responseHandler.Unauthorized(c, "Invalid user context")
		return
	}

	statistics, err := r.formService.GetFormStatistics(c.Request.Context(), formID, userIDStr)
	if err != nil {
		r.handleApplicationError(c, err)
		return
	}

	r.responseHandler.Success(c, statistics, "Statistics retrieved successfully")
}

// DuplicateForm creates a copy of an existing form
// @Summary Duplicate form
// @Description Create a copy of an existing form
// @Tags Forms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Form ID" format(uuid)
// @Success 201 {object} dto.SuccessResponse{data=dto.FormResponseDTO} "Form duplicated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid form ID"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Form not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /forms/{id}/duplicate [post]
func (r *FormRoutesV1) DuplicateForm(c *gin.Context) {
	formID := c.Param("id")
	if !r.validator.ValidateFormID(c, formID) {
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		r.responseHandler.Unauthorized(c, "User authentication required")
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		r.responseHandler.Unauthorized(c, "Invalid user context")
		return
	}

	newForm, err := r.formService.DuplicateForm(c.Request.Context(), formID, userIDStr)
	if err != nil {
		r.handleApplicationError(c, err)
		return
	}

	r.responseHandler.Created(c, newForm, "Form duplicated successfully")
}

// GetPublicForm retrieves a public form (no authentication required)
// @Summary Get public form
// @Description Retrieve a public form that can be filled by anyone
// @Tags Public
// @Accept json
// @Produce json
// @Param id path string true "Form ID" format(uuid)
// @Success 200 {object} dto.SuccessResponse{data=dto.FormResponseDTO} "Public form retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid form ID"
// @Failure 404 {object} dto.ErrorResponse "Form not found or not public"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /public/forms/{id} [get]
func (r *FormRoutesV1) GetPublicForm(c *gin.Context) {
	formID := c.Param("id")
	if !r.validator.ValidateFormID(c, formID) {
		return
	}

	form, err := r.formService.GetPublicForm(c.Request.Context(), formID)
	if err != nil {
		r.handleApplicationError(c, err)
		return
	}

	r.responseHandler.Success(c, form, "Public form retrieved successfully")
}

// =============================================================================
// Health and Monitoring Handlers
// =============================================================================

// HealthCheck performs comprehensive health check
// @Summary Health check
// @Description Comprehensive health check including dependencies
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} dto.HealthCheckResponseDTO "Service is healthy"
// @Failure 503 {object} dto.HealthCheckResponseDTO "Service is unhealthy"
// @Router /health [get]
func (r *FormRoutesV1) HealthCheck(c *gin.Context) {
	health, err := r.formService.GetHealthStatus(c.Request.Context())
	if err != nil {
		r.responseHandler.ServiceUnavailable(c, "Health check failed", err.Error())
		return
	}

	r.responseHandler.HealthCheck(c, *health)
}

// ReadinessCheck checks if service is ready to accept traffic
// @Summary Readiness check
// @Description Check if service is ready to accept requests
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} dto.SuccessResponse "Service is ready"
// @Failure 503 {object} dto.ErrorResponse "Service is not ready"
// @Router /health/ready [get]
func (r *FormRoutesV1) ReadinessCheck(c *gin.Context) {
	ready, err := r.formService.IsReady(c.Request.Context())
	if err != nil || !ready {
		r.responseHandler.ServiceUnavailable(c, "Service is not ready")
		return
	}

	r.responseHandler.Success(c, map[string]bool{"ready": true}, "Service is ready")
}

// LivenessCheck checks if service is alive
// @Summary Liveness check
// @Description Check if service is alive and responsive
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} dto.SuccessResponse "Service is alive"
// @Failure 503 {object} dto.ErrorResponse "Service is not responsive"
// @Router /health/live [get]
func (r *FormRoutesV1) LivenessCheck(c *gin.Context) {
	r.responseHandler.Success(c, map[string]bool{"alive": true}, "Service is alive")
}

// GetMetrics retrieves service metrics
// @Summary Get service metrics
// @Description Get performance and usage metrics
// @Tags Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} dto.SuccessResponse{data=dto.HealthMetricsDTO} "Metrics retrieved successfully"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /metrics [get]
func (r *FormRoutesV1) GetMetrics(c *gin.Context) {
	metrics, err := r.formService.GetMetrics(c.Request.Context())
	if err != nil {
		r.responseHandler.InternalServerError(c, "Failed to retrieve metrics", err.Error())
		return
	}

	r.responseHandler.Success(c, metrics, "Metrics retrieved successfully")
}

// =============================================================================
// Helper Methods
// =============================================================================

// parseListFormsRequest parses query parameters for form listing
func (r *FormRoutesV1) parseListFormsRequest(c *gin.Context) dto.FormListRequestDTO {
	req := dto.FormListRequestDTO{
		Page:      1,
		PageSize:  20,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			req.Page = p
		}
	}

	if pageSize := c.Query("pageSize"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
			req.PageSize = ps
		}
	}

	if status := c.Query("status"); status != "" {
		req.Status = status
	}

	if category := c.Query("category"); category != "" {
		req.Category = category
	}

	if search := c.Query("search"); search != "" {
		req.Search = search
	}

	if sortBy := c.Query("sortBy"); sortBy != "" {
		req.SortBy = sortBy
	}

	if sortOrder := c.Query("sortOrder"); sortOrder != "" {
		req.SortOrder = sortOrder
	}

	if createdBy := c.Query("createdBy"); createdBy != "" {
		req.CreatedBy = createdBy
	}

	return req
}

// handleApplicationError handles errors from the application layer
func (r *FormRoutesV1) handleApplicationError(c *gin.Context, err error) {
	// This would typically use a more sophisticated error mapping
	// based on domain error types
	switch err.Error() {
	case "form not found":
		r.responseHandler.NotFound(c, "Form not found")
	case "unauthorized":
		r.responseHandler.Forbidden(c, "You don't have permission to access this form")
	case "form already published":
		r.responseHandler.Conflict(c, "Form is already published")
	case "invalid form status":
		r.responseHandler.BadRequest(c, "Invalid form status for this operation")
	default:
		r.responseHandler.InternalServerError(c, "An error occurred while processing your request", err.Error())
	}
}
