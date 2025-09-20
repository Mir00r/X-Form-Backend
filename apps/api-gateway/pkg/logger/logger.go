// Package logger provides structured logging functionality
// Implements logging interface following industry best practices for observability
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogConfig holds logger configuration
type LogConfig struct {
	// Level is the minimum log level to output
	Level string `mapstructure:"level" validate:"required,oneof=debug info warn error fatal"`

	// Format is the log output format
	Format string `mapstructure:"format" validate:"required,oneof=json text pretty"`

	// Output is the log output destination
	Output string `mapstructure:"output" validate:"required,oneof=stdout stderr file"`

	// FilePath is the path to the log file (if Output is "file")
	FilePath string `mapstructure:"file_path"`

	// IncludeCaller controls whether to include caller information
	IncludeCaller bool `mapstructure:"include_caller"`

	// IncludeTimestamp controls whether to include timestamp
	IncludeTimestamp bool `mapstructure:"include_timestamp"`

	// TimeFormat is the format for timestamps
	TimeFormat string `mapstructure:"time_format"`

	// ServiceName is the name of the service
	ServiceName string `mapstructure:"service_name"`

	// ServiceVersion is the version of the service
	ServiceVersion string `mapstructure:"service_version"`
}

// LogLevel represents the severity level of a log entry
type LogLevel int

const (
	// DebugLevel for debug information
	DebugLevel LogLevel = iota
	// InfoLevel for general information
	InfoLevel
	// WarnLevel for warning messages
	WarnLevel
	// ErrorLevel for error messages
	ErrorLevel
	// FatalLevel for fatal errors that cause program termination
	FatalLevel
)

// String returns the string representation of log level
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	default:
		return "unknown"
	}
}

// ParseLogLevel parses string to LogLevel
func ParseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}

// Fields represents structured log fields
type Fields map[string]interface{}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Fields    Fields    `json:"fields,omitempty"`
	Caller    string    `json:"caller,omitempty"`
	Service   string    `json:"service,omitempty"`
	Version   string    `json:"version,omitempty"`
}

// Logger interface defines logging operations
type Logger interface {
	// Basic logging methods
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)

	// Formatted logging methods
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	// Structured logging methods
	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger

	// Error handling
	WithError(err error) Logger

	// Close closes the logger
	Close() error

	// Configuration methods
	SetLevel(level LogLevel)
	SetOutput(output io.Writer)

	// Structured logging - use a single method signature
	Log(level LogLevel, msg string, fields ...Fields)
}

// jsonLogger implements the Logger interface with JSON formatting
type jsonLogger struct {
	level          LogLevel
	output         io.Writer
	fields         Fields
	includeCaller  bool
	serviceName    string
	serviceVersion string
}

// New creates a new logger instance
func New(config LogConfig) Logger {
	// Parse log level
	level := ParseLogLevel(config.Level)

	// Set output writer
	var output io.Writer
	switch strings.ToLower(config.Output) {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	case "file":
		if config.FilePath == "" {
			log.Println("Warning: file output specified but no file path provided, defaulting to stdout")
			output = os.Stdout
		} else {
			file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				log.Printf("Error opening log file: %v, defaulting to stdout", err)
				output = os.Stdout
			} else {
				output = file
			}
		}
	default:
		output = os.Stdout
	}

	// Create logger based on format
	switch strings.ToLower(config.Format) {
	case "json":
		return &jsonLogger{
			level:          level,
			output:         output,
			fields:         make(Fields),
			includeCaller:  config.IncludeCaller,
			serviceName:    config.ServiceName,
			serviceVersion: config.ServiceVersion,
		}
	case "text", "pretty":
		// For simplicity, we'll just use JSON logger for now
		// In a real implementation, you would create separate loggers for different formats
		return &jsonLogger{
			level:          level,
			output:         output,
			fields:         make(Fields),
			includeCaller:  config.IncludeCaller,
			serviceName:    config.ServiceName,
			serviceVersion: config.ServiceVersion,
		}
	default:
		return &jsonLogger{
			level:          level,
			output:         output,
			fields:         make(Fields),
			includeCaller:  config.IncludeCaller,
			serviceName:    config.ServiceName,
			serviceVersion: config.ServiceVersion,
		}
	}
}

// log logs a message at the specified level
func (l *jsonLogger) log(level LogLevel, msg string, fields Fields) {
	if level < l.level {
		return
	}

	// Create log entry
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level.String(),
		Message:   msg,
		Service:   l.serviceName,
		Version:   l.serviceVersion,
	}

	// Add caller information if enabled
	if l.includeCaller {
		_, file, line, ok := runtime.Caller(2)
		if ok {
			entry.Caller = fmt.Sprintf("%s:%d", file, line)
		}
	}

	// Merge fields
	if len(l.fields) > 0 || len(fields) > 0 {
		entry.Fields = make(Fields)
		// Add logger fields
		for k, v := range l.fields {
			entry.Fields[k] = v
		}
		// Add message fields (overrides logger fields)
		for k, v := range fields {
			entry.Fields[k] = v
		}
	}

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling log entry: %v\n", err)
		return
	}

	// Write to output
	data = append(data, '\n')
	_, err = l.output.Write(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing log entry: %v\n", err)
	}

	// If fatal, exit
	if level == FatalLevel {
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *jsonLogger) Debug(msg string) {
	l.log(DebugLevel, msg, nil)
}

// Info logs an info message
func (l *jsonLogger) Info(msg string) {
	l.log(InfoLevel, msg, nil)
}

// Warn logs a warning message
func (l *jsonLogger) Warn(msg string) {
	l.log(WarnLevel, msg, nil)
}

// Error logs an error message
func (l *jsonLogger) Error(msg string) {
	l.log(ErrorLevel, msg, nil)
}

// Fatal logs a fatal message and exits
func (l *jsonLogger) Fatal(msg string) {
	l.log(FatalLevel, msg, nil)
}

// Debugf logs a formatted debug message
func (l *jsonLogger) Debugf(format string, args ...interface{}) {
	l.log(DebugLevel, fmt.Sprintf(format, args...), nil)
}

// Infof logs a formatted info message
func (l *jsonLogger) Infof(format string, args ...interface{}) {
	l.log(InfoLevel, fmt.Sprintf(format, args...), nil)
}

// Warnf logs a formatted warning message
func (l *jsonLogger) Warnf(format string, args ...interface{}) {
	l.log(WarnLevel, fmt.Sprintf(format, args...), nil)
}

// Errorf logs a formatted error message
func (l *jsonLogger) Errorf(format string, args ...interface{}) {
	l.log(ErrorLevel, fmt.Sprintf(format, args...), nil)
}

// Fatalf logs a formatted fatal message and exits
func (l *jsonLogger) Fatalf(format string, args ...interface{}) {
	l.log(FatalLevel, fmt.Sprintf(format, args...), nil)
}

// WithField returns a new logger with the field added
func (l *jsonLogger) WithField(key string, value interface{}) Logger {
	newLogger := &jsonLogger{
		level:          l.level,
		output:         l.output,
		fields:         make(Fields),
		includeCaller:  l.includeCaller,
		serviceName:    l.serviceName,
		serviceVersion: l.serviceVersion,
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new field
	newLogger.fields[key] = value

	return newLogger
}

// WithFields returns a new logger with the fields added
func (l *jsonLogger) WithFields(fields Fields) Logger {
	newLogger := &jsonLogger{
		level:          l.level,
		output:         l.output,
		fields:         make(Fields),
		includeCaller:  l.includeCaller,
		serviceName:    l.serviceName,
		serviceVersion: l.serviceVersion,
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// WithError returns a new logger with the error added
func (l *jsonLogger) WithError(err error) Logger {
	return l.WithField("error", err.Error())
}

// Close closes the logger
func (l *jsonLogger) Close() error {
	// Close file if output is a file
	if closer, ok := l.output.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// Implementation of additional Logger methods

// SetLevel sets the logger level
func (l *jsonLogger) SetLevel(level LogLevel) {
	l.level = level
}

// SetOutput sets the logger output
func (l *jsonLogger) SetOutput(output io.Writer) {
	l.output = output
}

// Log logs a message with the specified level
func (l *jsonLogger) Log(level LogLevel, msg string, fields ...Fields) {
	var fieldsToUse Fields
	if len(fields) > 0 {
		fieldsToUse = fields[0]
	}
	l.log(level, msg, fieldsToUse)
}

// Enhanced logging utility functions for production use

// CreateStructuredLogger creates a logger with production-ready configuration
func CreateStructuredLogger(serviceName, serviceVersion string) Logger {
	config := LogConfig{
		Level:            "info",
		Format:           "json",
		Output:           "stdout",
		IncludeCaller:    true,
		IncludeTimestamp: true,
		TimeFormat:       time.RFC3339Nano,
		ServiceName:      serviceName,
		ServiceVersion:   serviceVersion,
	}
	return New(config)
}

// LogWithTraceID logs a message with trace ID context
func LogWithTraceID(logger Logger, level LogLevel, msg, traceID string) {
	fields := Fields{"trace_id": traceID}
	logger.Log(level, msg, fields)
}

// LogWithRequestContext logs with comprehensive request context
func LogWithRequestContext(logger Logger, level LogLevel, msg string, method, path, clientIP, traceID string, duration time.Duration, statusCode int) {
	fields := Fields{
		"trace_id":    traceID,
		"method":      method,
		"path":        path,
		"client_ip":   clientIP,
		"duration_ms": duration.Milliseconds(),
		"status_code": statusCode,
		"timestamp":   time.Now().Format(time.RFC3339Nano),
	}
	logger.Log(level, msg, fields)
}

// LogAuditEvent logs audit events for security monitoring
func LogAuditEvent(logger Logger, event, userID, resource string, additionalFields Fields) {
	fields := Fields{
		"audit_event":     event,
		"audit_user_id":   userID,
		"audit_resource":  resource,
		"audit_timestamp": time.Now().Format(time.RFC3339Nano),
		"event_type":      "audit",
	}

	// Merge additional fields
	if additionalFields != nil {
		for k, v := range additionalFields {
			fields[k] = v
		}
	}

	logger.Log(InfoLevel, fmt.Sprintf("AUDIT: %s", event), fields)
}

// LogPerformanceMetrics logs performance-related information
func LogPerformanceMetrics(logger Logger, operation string, duration time.Duration, success bool, additionalMetrics Fields) {
	fields := Fields{
		"operation":   operation,
		"duration_ms": duration.Milliseconds(),
		"success":     success,
		"metric_type": "performance",
		"timestamp":   time.Now().Format(time.RFC3339Nano),
	}

	if additionalMetrics != nil {
		for k, v := range additionalMetrics {
			fields[k] = v
		}
	}

	level := InfoLevel
	if duration > 1*time.Second {
		level = WarnLevel
	}
	if !success {
		level = ErrorLevel
	}

	logger.Log(level, fmt.Sprintf("Performance: %s", operation), fields)
}

// LogBusinessEvent logs business-specific events
func LogBusinessEvent(logger Logger, eventType, entity, action string, userID string, additionalData Fields) {
	fields := Fields{
		"business_event": eventType,
		"entity":         entity,
		"action":         action,
		"user_id":        userID,
		"event_type":     "business",
		"timestamp":      time.Now().Format(time.RFC3339Nano),
	}

	if additionalData != nil {
		for k, v := range additionalData {
			fields[k] = v
		}
	}

	logger.Log(InfoLevel, fmt.Sprintf("Business Event: %s %s", action, entity), fields)
}

// These methods are now implemented in the main Debug, Info, Warn, Error methods above

// Config holds logger configuration
type Config struct {
	// Level is the minimum log level to output
	Level LogLevel `json:"level"`

	// Format specifies the log format: "json" or "text"
	Format string `json:"format"`

	// Output specifies where to write logs: "stdout", "stderr", or file path
	Output string `json:"output"`

	// Service name for log entries
	Service string `json:"service"`

	// Version for log entries
	Version string `json:"version"`

	// EnableCaller adds caller information to logs
	EnableCaller bool `json:"enable_caller"`

	// EnableTimestamp adds timestamp to logs
	EnableTimestamp bool `json:"enable_timestamp"`

	// TimeFormat specifies timestamp format
	TimeFormat string `json:"time_format"`

	// PrettyPrint enables pretty printing for JSON format
	PrettyPrint bool `json:"pretty_print"`
}

// StructuredLogger implements the Logger interface
type StructuredLogger struct {
	config Config
	output io.Writer
	fields Fields
}

// NewLogger creates a new structured logger
func NewLogger(config Config) Logger {
	logger := &StructuredLogger{
		config: config,
		fields: make(Fields),
	}

	// Set default values
	if logger.config.TimeFormat == "" {
		logger.config.TimeFormat = time.RFC3339
	}

	// Configure output
	switch config.Output {
	case "stderr":
		logger.output = os.Stderr
	case "stdout", "":
		logger.output = os.Stdout
	default:
		// File output
		if file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
			logger.output = file
		} else {
			// Fallback to stdout
			logger.output = os.Stdout
		}
	}

	return logger
}

// Debug logs debug message
func (l *StructuredLogger) Debug(msg string) {
	l.log(DebugLevel, msg)
}

// Info logs info message
func (l *StructuredLogger) Info(msg string) {
	l.log(InfoLevel, msg)
}

// Warn logs warning message
func (l *StructuredLogger) Warn(msg string) {
	l.log(WarnLevel, msg)
}

// Error logs error message
func (l *StructuredLogger) Error(msg string) {
	l.log(ErrorLevel, msg)
}

// Fatal logs fatal message and exits
func (l *StructuredLogger) Fatal(msg string) {
	l.log(FatalLevel, msg)
	os.Exit(1)
}

// Debugf logs formatted debug message
func (l *StructuredLogger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

// Infof logs formatted info message
func (l *StructuredLogger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

// Warnf logs formatted warning message
func (l *StructuredLogger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

// Errorf logs formatted error message
func (l *StructuredLogger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

// Fatalf logs formatted fatal message and exits
func (l *StructuredLogger) Fatalf(format string, args ...interface{}) {
	l.Fatal(fmt.Sprintf(format, args...))
}

// WithField adds a field to the logger context
func (l *StructuredLogger) WithField(key string, value interface{}) Logger {
	newLogger := l.clone()
	newLogger.fields[key] = value
	return newLogger
}

// WithFields adds multiple fields to the logger context
func (l *StructuredLogger) WithFields(fields Fields) Logger {
	newLogger := l.clone()
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	return newLogger
}

// WithError adds error field to the logger context
func (l *StructuredLogger) WithError(err error) Logger {
	if err == nil {
		return l
	}
	return l.WithField("error", err.Error())
}

// Log logs a message with specified level
func (l *StructuredLogger) Log(level LogLevel, msg string, fields ...Fields) {
	l.log(level, msg)
}

// Close closes the logger (no-op for StructuredLogger)
func (l *StructuredLogger) Close() error {
	return nil
}

// SetLevel sets the minimum log level
func (l *StructuredLogger) SetLevel(level LogLevel) {
	l.config.Level = level
}

// SetOutput sets the log output writer
func (l *StructuredLogger) SetOutput(output io.Writer) {
	l.output = output
}

// Internal logging method
func (l *StructuredLogger) log(level LogLevel, msg string) {
	// Check if level is enabled
	if level < l.config.Level {
		return
	}

	// Create log entry
	entry := LogEntry{
		Level:   level.String(),
		Message: msg,
		Service: l.config.Service,
		Version: l.config.Version,
	}

	// Add timestamp
	if l.config.EnableTimestamp {
		entry.Timestamp = time.Now()
	}

	// Add caller information
	if l.config.EnableCaller {
		if pc, file, line, ok := runtime.Caller(3); ok {
			funcName := runtime.FuncForPC(pc).Name()
			entry.Caller = fmt.Sprintf("%s:%d %s", file, line, funcName)
		}
	}

	// Add fields
	if len(l.fields) > 0 {
		entry.Fields = make(Fields)
		for k, v := range l.fields {
			entry.Fields[k] = v
		}
	}

	// Write log entry
	l.writeEntry(entry)
}

// writeEntry writes log entry to output
func (l *StructuredLogger) writeEntry(entry LogEntry) {
	var output string

	switch l.config.Format {
	case "json":
		output = l.formatJSON(entry)
	default:
		output = l.formatText(entry)
	}

	// Write to output
	fmt.Fprintln(l.output, output)
}

// formatJSON formats entry as JSON
func (l *StructuredLogger) formatJSON(entry LogEntry) string {
	if l.config.PrettyPrint {
		if data, err := json.MarshalIndent(entry, "", "  "); err == nil {
			return string(data)
		}
	}

	if data, err := json.Marshal(entry); err == nil {
		return string(data)
	}

	// Fallback to text format
	return l.formatText(entry)
}

// formatText formats entry as human-readable text
func (l *StructuredLogger) formatText(entry LogEntry) string {
	var parts []string

	// Timestamp
	if !entry.Timestamp.IsZero() {
		parts = append(parts, entry.Timestamp.Format(l.config.TimeFormat))
	}

	// Level
	parts = append(parts, strings.ToUpper(entry.Level))

	// Service
	if entry.Service != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Service))
	}

	// Message
	parts = append(parts, entry.Message)

	// Fields
	if len(entry.Fields) > 0 {
		var fieldParts []string
		for k, v := range entry.Fields {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, v))
		}
		parts = append(parts, fmt.Sprintf("{%s}", strings.Join(fieldParts, " ")))
	}

	// Caller
	if entry.Caller != "" {
		parts = append(parts, fmt.Sprintf("caller=%s", entry.Caller))
	}

	return strings.Join(parts, " ")
}

// clone creates a copy of the logger with the same configuration
func (l *StructuredLogger) clone() *StructuredLogger {
	newLogger := &StructuredLogger{
		config: l.config,
		output: l.output,
		fields: make(Fields),
	}

	// Copy fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// Standard logger instance for global use
var std Logger

// init initializes the standard logger
func init() {
	std = NewLogger(DefaultConfig())
}

// DefaultConfig returns default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:           InfoLevel,
		Format:          "json",
		Output:          "stdout",
		Service:         "api-gateway",
		Version:         "1.0.0",
		EnableCaller:    false,
		EnableTimestamp: true,
		TimeFormat:      time.RFC3339,
		PrettyPrint:     false,
	}
}

// Package-level convenience functions using the standard logger

// Debug logs debug message using standard logger
func Debug(msg string) {
	std.Debug(msg)
}

// Info logs info message using standard logger
func Info(msg string) {
	std.Info(msg)
}

// Warn logs warning message using standard logger
func Warn(msg string) {
	std.Warn(msg)
}

// Error logs error message using standard logger
func Error(msg string) {
	std.Error(msg)
}

// Fatal logs fatal message using standard logger
func Fatal(msg string) {
	std.Fatal(msg)
}

// Debugf logs formatted debug message using standard logger
func Debugf(format string, args ...interface{}) {
	std.Debugf(format, args...)
}

// Infof logs formatted info message using standard logger
func Infof(format string, args ...interface{}) {
	std.Infof(format, args...)
}

// Warnf logs formatted warning message using standard logger
func Warnf(format string, args ...interface{}) {
	std.Warnf(format, args...)
}

// Errorf logs formatted error message using standard logger
func Errorf(format string, args ...interface{}) {
	std.Errorf(format, args...)
}

// Fatalf logs formatted fatal message using standard logger
func Fatalf(format string, args ...interface{}) {
	std.Fatalf(format, args...)
}

// WithField adds a field using standard logger
func WithField(key string, value interface{}) Logger {
	return std.WithField(key, value)
}

// WithFields adds multiple fields using standard logger
func WithFields(fields Fields) Logger {
	return std.WithFields(fields)
}

// WithError adds error field using standard logger
func WithError(err error) Logger {
	return std.WithError(err)
}

// SetLevel sets the level for standard logger
func SetLevel(level LogLevel) {
	std.SetLevel(level)
}

// SetOutput sets the output for standard logger
func SetOutput(output io.Writer) {
	std.SetOutput(output)
}

// GetStandardLogger returns the standard logger instance
func GetStandardLogger() Logger {
	return std
}

// SetStandardLogger sets a new standard logger
func SetStandardLogger(logger Logger) {
	std = logger
}

// SimpleLogger provides a simple text-based logger for basic use cases
type SimpleLogger struct {
	level  LogLevel
	output io.Writer
	prefix string
}

// NewSimpleLogger creates a simple logger
func NewSimpleLogger(level LogLevel, output io.Writer, prefix string) Logger {
	return &SimpleLogger{
		level:  level,
		output: output,
		prefix: prefix,
	}
}

// Debug logs debug message
func (s *SimpleLogger) Debug(msg string) {
	if s.level <= DebugLevel {
		s.write("DEBUG", msg)
	}
}

// Info logs info message
func (s *SimpleLogger) Info(msg string) {
	if s.level <= InfoLevel {
		s.write("INFO", msg)
	}
}

// Warn logs warning message
func (s *SimpleLogger) Warn(msg string) {
	if s.level <= WarnLevel {
		s.write("WARN", msg)
	}
}

// Error logs error message
func (s *SimpleLogger) Error(msg string) {
	if s.level <= ErrorLevel {
		s.write("ERROR", msg)
	}
}

// Fatal logs fatal message
func (s *SimpleLogger) Fatal(msg string) {
	s.write("FATAL", msg)
	os.Exit(1)
}

// Debugf logs formatted debug message
func (s *SimpleLogger) Debugf(format string, args ...interface{}) {
	s.Debug(fmt.Sprintf(format, args...))
}

// Infof logs formatted info message
func (s *SimpleLogger) Infof(format string, args ...interface{}) {
	s.Info(fmt.Sprintf(format, args...))
}

// Warnf logs formatted warning message
func (s *SimpleLogger) Warnf(format string, args ...interface{}) {
	s.Warn(fmt.Sprintf(format, args...))
}

// Errorf logs formatted error message
func (s *SimpleLogger) Errorf(format string, args ...interface{}) {
	s.Error(fmt.Sprintf(format, args...))
}

// Fatalf logs formatted fatal message
func (s *SimpleLogger) Fatalf(format string, args ...interface{}) {
	s.Fatal(fmt.Sprintf(format, args...))
}

// WithField returns the same logger (simple logger doesn't support fields)
func (s *SimpleLogger) WithField(key string, value interface{}) Logger {
	return s
}

// WithFields returns the same logger (simple logger doesn't support fields)
func (s *SimpleLogger) WithFields(fields Fields) Logger {
	return s
}

// WithError returns the same logger (simple logger doesn't support fields)
func (s *SimpleLogger) WithError(err error) Logger {
	return s
}

// Log logs a message with specified level
func (s *SimpleLogger) Log(level LogLevel, msg string, fields ...Fields) {
	if s.level <= level {
		s.write(strings.ToUpper(level.String()), msg)
	}
}

// Close closes the logger (no-op for SimpleLogger)
func (s *SimpleLogger) Close() error {
	return nil
}

// SetLevel sets the minimum log level
func (s *SimpleLogger) SetLevel(level LogLevel) {
	s.level = level
}

// SetOutput sets the log output writer
func (s *SimpleLogger) SetOutput(output io.Writer) {
	s.output = output
}

// write writes a log message
func (s *SimpleLogger) write(level, msg string) {
	timestamp := time.Now().Format(time.RFC3339)
	logMsg := fmt.Sprintf("%s [%s] %s%s\n", timestamp, level, s.prefix, msg)

	if s.output != nil {
		s.output.Write([]byte(logMsg))
	} else {
		log.Print(logMsg)
	}
}
