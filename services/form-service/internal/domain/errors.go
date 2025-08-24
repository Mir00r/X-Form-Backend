// Package domain defines custom error types for the form service
package domain

import "fmt"

// Error types for different categories of domain errors

// ValidationError represents validation failures in business logic
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string, args ...interface{}) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: fmt.Sprintf(message, args...),
		Code:    "VALIDATION_ERROR",
	}
}

// BusinessRuleError represents violations of business rules
type BusinessRuleError struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e *BusinessRuleError) Error() string {
	return fmt.Sprintf("business rule violation: %s", e.Message)
}

// NewBusinessRuleError creates a new business rule error
func NewBusinessRuleError(message string, args ...interface{}) *BusinessRuleError {
	return &BusinessRuleError{
		Message: fmt.Sprintf(message, args...),
		Code:    "BUSINESS_RULE_VIOLATION",
	}
}

// NotFoundError represents resource not found errors
type NotFoundError struct {
	Resource string `json:"resource"`
	ID       string `json:"id"`
	Message  string `json:"message"`
	Code     string `json:"code"`
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found", e.Resource)
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{
		Resource: resource,
		ID:       id,
		Message:  fmt.Sprintf("%s with ID %s not found", resource, id),
		Code:     "RESOURCE_NOT_FOUND",
	}
}

// AccessDeniedError represents authorization failures
type AccessDeniedError struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	UserID   string `json:"user_id"`
	Message  string `json:"message"`
	Code     string `json:"code"`
}

func (e *AccessDeniedError) Error() string {
	return fmt.Sprintf("access denied: %s", e.Message)
}

// NewAccessDeniedError creates a new access denied error
func NewAccessDeniedError(resource, action, userID, message string) *AccessDeniedError {
	return &AccessDeniedError{
		Resource: resource,
		Action:   action,
		UserID:   userID,
		Message:  message,
		Code:     "ACCESS_DENIED",
	}
}

// ConflictError represents resource conflict errors
type ConflictError struct {
	Resource string `json:"resource"`
	Message  string `json:"message"`
	Code     string `json:"code"`
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict: %s", e.Message)
}

// NewConflictError creates a new conflict error
func NewConflictError(resource, message string) *ConflictError {
	return &ConflictError{
		Resource: resource,
		Message:  message,
		Code:     "RESOURCE_CONFLICT",
	}
}

// InternalError represents internal system errors
type InternalError struct {
	Operation string `json:"operation"`
	Message   string `json:"message"`
	Code      string `json:"code"`
	Cause     error  `json:"-"`
}

func (e *InternalError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("internal error during %s: %s (caused by: %v)", e.Operation, e.Message, e.Cause)
	}
	return fmt.Sprintf("internal error during %s: %s", e.Operation, e.Message)
}

// Unwrap returns the underlying error
func (e *InternalError) Unwrap() error {
	return e.Cause
}

// NewInternalError creates a new internal error
func NewInternalError(operation, message string, cause error) *InternalError {
	return &InternalError{
		Operation: operation,
		Message:   message,
		Code:      "INTERNAL_ERROR",
		Cause:     cause,
	}
}

// Error checking utilities

// IsValidationError checks if error is a validation error
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// IsBusinessRuleError checks if error is a business rule error
func IsBusinessRuleError(err error) bool {
	_, ok := err.(*BusinessRuleError)
	return ok
}

// IsNotFoundError checks if error is a not found error
func IsNotFoundError(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// IsAccessDeniedError checks if error is an access denied error
func IsAccessDeniedError(err error) bool {
	_, ok := err.(*AccessDeniedError)
	return ok
}

// IsConflictError checks if error is a conflict error
func IsConflictError(err error) bool {
	_, ok := err.(*ConflictError)
	return ok
}

// IsInternalError checks if error is an internal error
func IsInternalError(err error) bool {
	_, ok := err.(*InternalError)
	return ok
}
