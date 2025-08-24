// Package logger provides a simple logging interface
// This is a temporary solution until we add proper logrus dependency
package logger

import (
	"log"
	"os"
)

// Logger interface for dependency injection
type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Warn(args ...interface{})
}

// SimpleLogger implements Logger using standard log package
type SimpleLogger struct {
	logger *log.Logger
}

// NewSimpleLogger creates a new simple logger
func NewSimpleLogger() *SimpleLogger {
	return &SimpleLogger{
		logger: log.New(os.Stdout, "[FORM-SERVICE] ", log.LstdFlags|log.Lshortfile),
	}
}

// Info logs info messages
func (l *SimpleLogger) Info(args ...interface{}) {
	l.logger.Println(append([]interface{}{"INFO:"}, args...)...)
}

// Error logs error messages
func (l *SimpleLogger) Error(args ...interface{}) {
	l.logger.Println(append([]interface{}{"ERROR:"}, args...)...)
}

// Debug logs debug messages
func (l *SimpleLogger) Debug(args ...interface{}) {
	l.logger.Println(append([]interface{}{"DEBUG:"}, args...)...)
}

// Warn logs warning messages
func (l *SimpleLogger) Warn(args ...interface{}) {
	l.logger.Println(append([]interface{}{"WARN:"}, args...)...)
}
