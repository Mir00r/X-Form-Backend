// Package validator provides parameter validation functionality
// Implements comprehensive request validation following security best practices
package validator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

// ValidationRule defines a validation rule for request parameters
type ValidationRule struct {
	// Field name to validate
	Name string `json:"name"`

	// Required indicates if the field is mandatory
	Required bool `json:"required"`

	// Type specifies the expected data type
	Type string `json:"type"` // string, int, float, bool, email, url, uuid

	// MinLength for string validation
	MinLength int `json:"min_length,omitempty"`

	// MaxLength for string validation
	MaxLength int `json:"max_length,omitempty"`

	// Min value for numeric validation
	Min *float64 `json:"min,omitempty"`

	// Max value for numeric validation
	Max *float64 `json:"max,omitempty"`

	// Pattern for regex validation
	Pattern string `json:"pattern,omitempty"`

	// AllowedValues for enum validation
	AllowedValues []string `json:"allowed_values,omitempty"`

	// CustomMessage for validation error
	CustomMessage string `json:"custom_message,omitempty"`
}

// ValidationConfig holds validation configuration for different endpoints
type ValidationConfig struct {
	// Rules maps endpoint patterns to validation rules
	Rules map[string][]ValidationRule `json:"rules"`

	// GlobalRules applied to all endpoints
	GlobalRules []ValidationRule `json:"global_rules"`

	// SkipValidation for certain endpoints
	SkipValidation []string `json:"skip_validation"`

	// MaxRequestSize in bytes
	MaxRequestSize int64 `json:"max_request_size"`

	// ValidateHeaders enables header validation
	ValidateHeaders bool `json:"validate_headers"`

	// ValidateQuery enables query parameter validation
	ValidateQuery bool `json:"validate_query"`

	// ValidateBody enables request body validation
	ValidateBody bool `json:"validate_body"`
}

// ParameterValidator handles request parameter validation
type ParameterValidator struct {
	config           ValidationConfig
	compiledPatterns map[string]*regexp.Regexp
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "validation failed"
	}

	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// NewParameterValidator creates a new parameter validator
func NewParameterValidator(config ValidationConfig) (*ParameterValidator, error) {
	validator := &ParameterValidator{
		config:           config,
		compiledPatterns: make(map[string]*regexp.Regexp),
	}

	// Compile regex patterns
	allRules := config.GlobalRules
	for _, rules := range config.Rules {
		allRules = append(allRules, rules...)
	}

	for _, rule := range allRules {
		if rule.Pattern != "" {
			compiled, err := regexp.Compile(rule.Pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid regex pattern for rule %s: %w", rule.Name, err)
			}
			validator.compiledPatterns[rule.Pattern] = compiled
		}
	}

	return validator, nil
}

// ValidateRequest validates the entire request according to configured rules
func (v *ParameterValidator) ValidateRequest(c *gin.Context) error {
	path := c.Request.URL.Path
	method := c.Request.Method

	// Check if validation should be skipped
	if v.shouldSkipValidation(path) {
		return nil
	}

	// Get validation rules for this endpoint
	rules := v.getRulesForEndpoint(method, path)
	if len(rules) == 0 {
		return nil // No rules defined
	}

	var errors ValidationErrors

	// Validate headers
	if v.config.ValidateHeaders {
		if headerErrors := v.validateHeaders(c, rules); len(headerErrors) > 0 {
			errors = append(errors, headerErrors...)
		}
	}

	// Validate query parameters
	if v.config.ValidateQuery {
		if queryErrors := v.validateQuery(c, rules); len(queryErrors) > 0 {
			errors = append(errors, queryErrors...)
		}
	}

	// Validate request body
	if v.config.ValidateBody && hasBody(c.Request) {
		if bodyErrors := v.validateBody(c, rules); len(bodyErrors) > 0 {
			errors = append(errors, bodyErrors...)
		}
	}

	// Validate request size
	if v.config.MaxRequestSize > 0 && c.Request.ContentLength > v.config.MaxRequestSize {
		errors = append(errors, ValidationError{
			Field:   "request_size",
			Message: fmt.Sprintf("request size exceeds maximum allowed (%d bytes)", v.config.MaxRequestSize),
			Value:   c.Request.ContentLength,
		})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// validateHeaders validates HTTP headers
func (v *ParameterValidator) validateHeaders(c *gin.Context, rules []ValidationRule) ValidationErrors {
	var errors ValidationErrors

	for _, rule := range rules {
		if !strings.HasPrefix(rule.Name, "header:") {
			continue
		}

		headerName := strings.TrimPrefix(rule.Name, "header:")
		headerValue := c.GetHeader(headerName)

		if err := v.validateValue(rule, headerValue); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// validateQuery validates query parameters
func (v *ParameterValidator) validateQuery(c *gin.Context, rules []ValidationRule) ValidationErrors {
	var errors ValidationErrors

	for _, rule := range rules {
		if !strings.HasPrefix(rule.Name, "query:") {
			continue
		}

		paramName := strings.TrimPrefix(rule.Name, "query:")
		paramValue := c.Query(paramName)

		if err := v.validateValue(rule, paramValue); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// validateBody validates request body
func (v *ParameterValidator) validateBody(c *gin.Context, rules []ValidationRule) ValidationErrors {
	var errors ValidationErrors

	// Parse JSON body
	var bodyData map[string]interface{}
	if err := c.ShouldBindJSON(&bodyData); err != nil {
		errors = append(errors, ValidationError{
			Field:   "body",
			Message: "invalid JSON body",
			Value:   err.Error(),
		})
		return errors
	}

	for _, rule := range rules {
		if !strings.HasPrefix(rule.Name, "body:") {
			continue
		}

		fieldName := strings.TrimPrefix(rule.Name, "body:")
		fieldValue, exists := bodyData[fieldName]

		var stringValue string
		if exists {
			stringValue = fmt.Sprintf("%v", fieldValue)
		}

		if err := v.validateValue(rule, stringValue); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// validateValue validates a single value against a rule
func (v *ParameterValidator) validateValue(rule ValidationRule, value string) *ValidationError {
	// Check required field
	if rule.Required && value == "" {
		return &ValidationError{
			Field:   rule.Name,
			Message: getValidationMessage(rule, "field is required"),
			Value:   value,
		}
	}

	// Skip validation for empty optional fields
	if !rule.Required && value == "" {
		return nil
	}

	// Type validation
	if err := v.validateType(rule, value); err != nil {
		return err
	}

	// Length validation for strings
	if rule.Type == "string" || rule.Type == "" {
		if err := v.validateLength(rule, value); err != nil {
			return err
		}
	}

	// Numeric range validation
	if rule.Type == "int" || rule.Type == "float" {
		if err := v.validateNumericRange(rule, value); err != nil {
			return err
		}
	}

	// Pattern validation
	if rule.Pattern != "" {
		if err := v.validatePattern(rule, value); err != nil {
			return err
		}
	}

	// Enum validation
	if len(rule.AllowedValues) > 0 {
		if err := v.validateEnum(rule, value); err != nil {
			return err
		}
	}

	return nil
}

// validateType validates the data type
func (v *ParameterValidator) validateType(rule ValidationRule, value string) *ValidationError {
	switch rule.Type {
	case "int":
		if _, err := strconv.Atoi(value); err != nil {
			return &ValidationError{
				Field:   rule.Name,
				Message: getValidationMessage(rule, "must be a valid integer"),
				Value:   value,
			}
		}
	case "float":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return &ValidationError{
				Field:   rule.Name,
				Message: getValidationMessage(rule, "must be a valid number"),
				Value:   value,
			}
		}
	case "bool":
		if _, err := strconv.ParseBool(value); err != nil {
			return &ValidationError{
				Field:   rule.Name,
				Message: getValidationMessage(rule, "must be a valid boolean"),
				Value:   value,
			}
		}
	case "email":
		if !isValidEmail(value) {
			return &ValidationError{
				Field:   rule.Name,
				Message: getValidationMessage(rule, "must be a valid email address"),
				Value:   value,
			}
		}
	case "url":
		if !isValidURL(value) {
			return &ValidationError{
				Field:   rule.Name,
				Message: getValidationMessage(rule, "must be a valid URL"),
				Value:   value,
			}
		}
	case "uuid":
		if !isValidUUID(value) {
			return &ValidationError{
				Field:   rule.Name,
				Message: getValidationMessage(rule, "must be a valid UUID"),
				Value:   value,
			}
		}
	}

	return nil
}

// validateLength validates string length
func (v *ParameterValidator) validateLength(rule ValidationRule, value string) *ValidationError {
	length := len(value)

	if rule.MinLength > 0 && length < rule.MinLength {
		return &ValidationError{
			Field:   rule.Name,
			Message: getValidationMessage(rule, fmt.Sprintf("must be at least %d characters long", rule.MinLength)),
			Value:   value,
		}
	}

	if rule.MaxLength > 0 && length > rule.MaxLength {
		return &ValidationError{
			Field:   rule.Name,
			Message: getValidationMessage(rule, fmt.Sprintf("must be at most %d characters long", rule.MaxLength)),
			Value:   value,
		}
	}

	return nil
}

// validateNumericRange validates numeric ranges
func (v *ParameterValidator) validateNumericRange(rule ValidationRule, value string) *ValidationError {
	var numValue float64
	var err error

	if rule.Type == "int" {
		intValue, parseErr := strconv.Atoi(value)
		if parseErr != nil {
			return nil // Type validation should have caught this
		}
		numValue = float64(intValue)
	} else {
		numValue, err = strconv.ParseFloat(value, 64)
		if err != nil {
			return nil // Type validation should have caught this
		}
	}

	if rule.Min != nil && numValue < *rule.Min {
		return &ValidationError{
			Field:   rule.Name,
			Message: getValidationMessage(rule, fmt.Sprintf("must be at least %g", *rule.Min)),
			Value:   value,
		}
	}

	if rule.Max != nil && numValue > *rule.Max {
		return &ValidationError{
			Field:   rule.Name,
			Message: getValidationMessage(rule, fmt.Sprintf("must be at most %g", *rule.Max)),
			Value:   value,
		}
	}

	return nil
}

// validatePattern validates regex patterns
func (v *ParameterValidator) validatePattern(rule ValidationRule, value string) *ValidationError {
	pattern, exists := v.compiledPatterns[rule.Pattern]
	if !exists {
		// Compile on demand (shouldn't happen normally)
		compiled, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return &ValidationError{
				Field:   rule.Name,
				Message: "invalid validation pattern",
				Value:   value,
			}
		}
		pattern = compiled
		v.compiledPatterns[rule.Pattern] = pattern
	}

	if !pattern.MatchString(value) {
		return &ValidationError{
			Field:   rule.Name,
			Message: getValidationMessage(rule, "does not match required pattern"),
			Value:   value,
		}
	}

	return nil
}

// validateEnum validates allowed values
func (v *ParameterValidator) validateEnum(rule ValidationRule, value string) *ValidationError {
	for _, allowed := range rule.AllowedValues {
		if value == allowed {
			return nil
		}
	}

	return &ValidationError{
		Field:   rule.Name,
		Message: getValidationMessage(rule, fmt.Sprintf("must be one of: %s", strings.Join(rule.AllowedValues, ", "))),
		Value:   value,
	}
}

// Helper functions

// shouldSkipValidation checks if validation should be skipped for the given path
func (v *ParameterValidator) shouldSkipValidation(path string) bool {
	for _, skipPath := range v.config.SkipValidation {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// getRulesForEndpoint gets validation rules for a specific endpoint
func (v *ParameterValidator) getRulesForEndpoint(method, path string) []ValidationRule {
	var rules []ValidationRule

	// Add global rules
	rules = append(rules, v.config.GlobalRules...)

	// Add endpoint-specific rules
	for pattern, endpointRules := range v.config.Rules {
		if matchEndpoint(method, path, pattern) {
			rules = append(rules, endpointRules...)
		}
	}

	return rules
}

// matchEndpoint checks if method and path match the pattern
func matchEndpoint(method, path, pattern string) bool {
	parts := strings.SplitN(pattern, " ", 2)
	if len(parts) != 2 {
		return false
	}

	patternMethod := parts[0]
	patternPath := parts[1]

	// Check method
	if patternMethod != "*" && patternMethod != method {
		return false
	}

	// Check path (support wildcards)
	if patternPath == "*" {
		return true
	}

	if strings.HasSuffix(patternPath, "*") {
		prefix := strings.TrimSuffix(patternPath, "*")
		return strings.HasPrefix(path, prefix)
	}

	return path == patternPath
}

// hasBody checks if the request has a body
func hasBody(req *http.Request) bool {
	return req.ContentLength > 0 || req.Header.Get("Transfer-Encoding") == "chunked"
}

// getValidationMessage returns custom message or default
func getValidationMessage(rule ValidationRule, defaultMessage string) string {
	if rule.CustomMessage != "" {
		return rule.CustomMessage
	}
	return defaultMessage
}

// Validation helper functions

// isValidEmail validates email format
func isValidEmail(email string) bool {
	// Basic email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// isValidURL validates URL format
func isValidURL(url string) bool {
	// Basic URL validation
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	return urlRegex.MatchString(url)
}

// isValidUUID validates UUID format
func isValidUUID(uuid string) bool {
	// UUID v4 format validation
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(strings.ToLower(uuid))
}

// isValidJSON validates JSON format
func isValidJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// isAlphanumeric checks if string contains only alphanumeric characters
func isAlphanumeric(str string) bool {
	for _, char := range str {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

// DefaultValidationConfig returns a default validation configuration
func DefaultValidationConfig() ValidationConfig {
	return ValidationConfig{
		Rules:           make(map[string][]ValidationRule),
		GlobalRules:     []ValidationRule{},
		SkipValidation:  []string{"/health", "/metrics", "/swagger"},
		MaxRequestSize:  10 * 1024 * 1024, // 10MB
		ValidateHeaders: true,
		ValidateQuery:   true,
		ValidateBody:    true,
	}
}
