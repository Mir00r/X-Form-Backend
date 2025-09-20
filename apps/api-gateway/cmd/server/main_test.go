package main

import (
	"testing"
	"os"
)

func TestMain(m *testing.M) {
	// Setup test environment
	os.Setenv("ENV", "test")
	os.Setenv("LOG_LEVEL", "error")
	
	// Run tests
	code := m.Run()
	
	// Cleanup
	os.Exit(code)
}

func TestApplicationCreation(t *testing.T) {
	// Set required environment variables for test
	os.Setenv("ENV", "test")
	os.Setenv("JWT_SECRET", "test-secret")
	
	app, err := NewApplication()
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}
	
	if app == nil {
		t.Fatal("Application should not be nil")
	}
	
	if app.config == nil {
		t.Fatal("Application config should not be nil")
	}
}
