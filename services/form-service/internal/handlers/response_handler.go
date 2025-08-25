// Package handlers contains API response utilities for the Form Service
// Following microservices best practices for consistent API responses
package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/dto"
)

// ResponseHandler handles standardized API responses with correlation IDs
type ResponseHandler struct {
	version string
}

// NewResponseHandler creates a new response handler instance
func NewResponseHandler(version string) *ResponseHandler {
	return &ResponseHandler{
		version: version,
	}
}

// =============================================================================
// Success Responses
// =============================================================================

// Success sends a successful response with optional data
func (h *ResponseHandler) Success(c *gin.Context, data interface{}, message ...string) {
	msg := "Success"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	response := dto.SuccessResponse{
		BaseResponse: h.createBaseResponse(c, true, msg),
		Data:         data,
	}

	c.JSON(http.StatusOK, response)
}

// Created sends a 201 Created response
func (h *ResponseHandler) Created(c *gin.Context, data interface{}, message ...string) {
	msg := "Resource created successfully"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	response := dto.SuccessResponse{
		BaseResponse: h.createBaseResponse(c, true, msg),
		Data:         data,
	}

	c.JSON(http.StatusCreated, response)
}

// Updated sends a 200 OK response for updates
func (h *ResponseHandler) Updated(c *gin.Context, data interface{}, message ...string) {
	msg := "Resource updated successfully"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	response := dto.SuccessResponse{
		BaseResponse: h.createBaseResponse(c, true, msg),
		Data:         data,
	}

	c.JSON(http.StatusOK, response)
}

// Deleted sends a 200 OK response for deletions
func (h *ResponseHandler) Deleted(c *gin.Context, message ...string) {
	msg := "Resource deleted successfully"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	response := dto.SuccessResponse{
		BaseResponse: h.createBaseResponse(c, true, msg),
		Data:         nil,
	}

	c.JSON(http.StatusOK, response)
}

// Paginated sends a paginated response
func (h *ResponseHandler) Paginated(c *gin.Context, data interface{}, pagination dto.Pagination, message ...string) {
	msg := "Data retrieved successfully"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	response := dto.PaginatedResponse{
		BaseResponse: h.createBaseResponse(c, true, msg),
		Data:         data,
		Pagination:   pagination,
	}

	c.JSON(http.StatusOK, response)
}

// =============================================================================
// Error Responses
// =============================================================================

// BadRequest sends a 400 Bad Request response
func (h *ResponseHandler) BadRequest(c *gin.Context, message string, details ...interface{}) {
	h.sendError(c, http.StatusBadRequest, "BAD_REQUEST", message, details...)
}

// Unauthorized sends a 401 Unauthorized response
func (h *ResponseHandler) Unauthorized(c *gin.Context, message string, details ...interface{}) {
	h.sendError(c, http.StatusUnauthorized, "UNAUTHORIZED", message, details...)
}

// Forbidden sends a 403 Forbidden response
func (h *ResponseHandler) Forbidden(c *gin.Context, message string, details ...interface{}) {
	h.sendError(c, http.StatusForbidden, "FORBIDDEN", message, details...)
}

// NotFound sends a 404 Not Found response
func (h *ResponseHandler) NotFound(c *gin.Context, message string, details ...interface{}) {
	h.sendError(c, http.StatusNotFound, "NOT_FOUND", message, details...)
}

// Conflict sends a 409 Conflict response
func (h *ResponseHandler) Conflict(c *gin.Context, message string, details ...interface{}) {
	h.sendError(c, http.StatusConflict, "CONFLICT", message, details...)
}

// UnprocessableEntity sends a 422 Unprocessable Entity response
func (h *ResponseHandler) UnprocessableEntity(c *gin.Context, message string, details ...interface{}) {
	h.sendError(c, http.StatusUnprocessableEntity, "UNPROCESSABLE_ENTITY", message, details...)
}

// TooManyRequests sends a 429 Too Many Requests response
func (h *ResponseHandler) TooManyRequests(c *gin.Context, message string, retryAfter int, rateLimit dto.RateLimitInfoDTO) {
	baseResponse := h.createBaseResponse(c, false, "")

	response := dto.RateLimitExceededDTO{
		BaseResponse: baseResponse,
		Error: dto.ErrorDetail{
			Code:      "RATE_LIMIT_EXCEEDED",
			Message:   message,
			RequestID: h.getCorrelationID(c),
			Path:      c.Request.URL.Path,
			Timestamp: time.Now().UTC(),
		},
		RateLimit:  rateLimit,
		RetryAfter: retryAfter,
	}

	c.Header("Retry-After", string(retryAfter))
	c.Header("X-RateLimit-Limit", string(rateLimit.Limit))
	c.Header("X-RateLimit-Remaining", string(rateLimit.Remaining))
	c.Header("X-RateLimit-Reset", rateLimit.Reset.Format(time.RFC3339))

	c.JSON(http.StatusTooManyRequests, response)
}

// InternalServerError sends a 500 Internal Server Error response
func (h *ResponseHandler) InternalServerError(c *gin.Context, message string, details ...interface{}) {
	h.sendError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message, details...)
}

// ServiceUnavailable sends a 503 Service Unavailable response
func (h *ResponseHandler) ServiceUnavailable(c *gin.Context, message string, details ...interface{}) {
	h.sendError(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message, details...)
}

// =============================================================================
// Validation Error Responses
// =============================================================================

// ValidationError sends a 400 Bad Request response with validation details
func (h *ResponseHandler) ValidationError(c *gin.Context, fieldErrors map[string][]string, message ...string) {
	msg := "Input validation failed"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	baseResponse := h.createBaseResponse(c, false, "")

	response := dto.ValidationErrorDTO{
		BaseResponse: baseResponse,
		Error: dto.ValidationErrorDetail{
			Code:      "VALIDATION_ERROR",
			Message:   msg,
			Fields:    fieldErrors,
			RequestID: h.getCorrelationID(c),
			Path:      c.Request.URL.Path,
			Timestamp: time.Now().UTC(),
		},
	}

	c.JSON(http.StatusBadRequest, response)
}

// =============================================================================
// Health Check Responses
// =============================================================================

// HealthCheck sends a health check response
func (h *ResponseHandler) HealthCheck(c *gin.Context, healthData dto.HealthCheckResponseDTO) {
	statusCode := http.StatusOK
	if healthData.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, healthData)
}

// =============================================================================
// Helper Methods
// =============================================================================

// sendError is a helper method to send standardized error responses
func (h *ResponseHandler) sendError(c *gin.Context, statusCode int, code, message string, details ...interface{}) {
	var detail interface{}
	if len(details) > 0 {
		detail = details[0]
	}

	baseResponse := h.createBaseResponse(c, false, "")

	response := dto.ErrorResponse{
		BaseResponse: baseResponse,
		Error: dto.ErrorDetail{
			Code:      code,
			Message:   message,
			Details:   detail,
			RequestID: h.getCorrelationID(c),
			Path:      c.Request.URL.Path,
			Timestamp: time.Now().UTC(),
		},
	}

	c.JSON(statusCode, response)
}

// createBaseResponse creates a base response with common fields
func (h *ResponseHandler) createBaseResponse(c *gin.Context, success bool, message string) dto.BaseResponse {
	return dto.BaseResponse{
		Success:       success,
		Message:       message,
		CorrelationID: h.getCorrelationID(c),
		Timestamp:     time.Now().UTC(),
		Version:       h.version,
	}
}

// getCorrelationID extracts or generates a correlation ID
func (h *ResponseHandler) getCorrelationID(c *gin.Context) string {
	// Try to get from headers
	if correlationID := c.GetHeader("X-Correlation-ID"); correlationID != "" {
		return correlationID
	}

	// Try to get from context
	if correlationID, exists := c.Get("correlationID"); exists {
		if id, ok := correlationID.(string); ok {
			return id
		}
	}

	// Generate new correlation ID
	return uuid.New().String()
}

// =============================================================================
// Middleware for Correlation ID
// =============================================================================

// CorrelationIDMiddleware adds correlation ID to request context
func CorrelationIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Set in context for use in handlers
		c.Set("correlationID", correlationID)

		// Set in response headers
		c.Header("X-Correlation-ID", correlationID)

		// Add to request context for downstream services
		ctx := context.WithValue(c.Request.Context(), "correlationID", correlationID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// =============================================================================
// Request Metrics Middleware
// =============================================================================

// RequestMetricsMiddleware collects request metrics
func RequestMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get correlation ID
		correlationID := getCorrelationIDFromContext(c)

		// Set metrics headers
		c.Header("X-Response-Time", duration.String())
		c.Header("X-Request-ID", correlationID)

		// Log request metrics (can be sent to monitoring system)
		// This would typically integrate with Prometheus or similar
		logRequestMetrics(c, duration, correlationID)
	}
}

// getCorrelationIDFromContext extracts correlation ID from context
func getCorrelationIDFromContext(c *gin.Context) string {
	// Try to get from headers
	if correlationID := c.GetHeader("X-Correlation-ID"); correlationID != "" {
		return correlationID
	}

	// Try to get from context
	if correlationID, exists := c.Get("correlationID"); exists {
		if id, ok := correlationID.(string); ok {
			return id
		}
	}

	// Generate new correlation ID
	return uuid.New().String()
}

// logRequestMetrics logs request metrics for monitoring
func logRequestMetrics(c *gin.Context, duration time.Duration, correlationID string) {
	// This is a placeholder for actual metrics collection
	// In production, this would send metrics to Prometheus, Grafana, etc.

	metrics := map[string]interface{}{
		"method":         c.Request.Method,
		"path":           c.Request.URL.Path,
		"status_code":    c.Writer.Status(),
		"duration_ms":    duration.Milliseconds(),
		"correlation_id": correlationID,
		"timestamp":      time.Now().UTC(),
		"user_agent":     c.GetHeader("User-Agent"),
		"remote_addr":    c.ClientIP(),
	}

	// In a real implementation, you would send these metrics to your monitoring system
	// Example: prometheus.Histogram.WithLabelValues(...).Observe(duration.Seconds())
	_ = metrics // Suppress unused variable warning
}

// =============================================================================
// Global Error Handler
// =============================================================================

// ErrorHandler handles global errors and panics
func ErrorHandler(handler *ResponseHandler) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			handler.InternalServerError(c, "Internal server error occurred", map[string]interface{}{
				"error": err,
			})
		} else {
			handler.InternalServerError(c, "An unexpected error occurred")
		}
		c.Abort()
	})
}

// =============================================================================
// Security Headers Middleware
// =============================================================================

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")

		// API versioning headers
		c.Header("API-Version", "v1")
		c.Header("API-Supported-Versions", "v1")

		c.Next()
	}
}
