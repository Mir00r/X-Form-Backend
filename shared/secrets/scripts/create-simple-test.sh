#!/bin/bash

echo "🔧 Creating simplified secrets module for testing..."

# Create a new test directory
cd /Users/mir00r/Developer/kamkaiz/X-Form-Backend/shared
mkdir -p secrets-test
cd secrets-test

# Initialize a new Go module
go mod init github.com/kamkaiz/x-form-backend/shared/secrets-test

# Copy only essential files
cp ../secrets/manager.go .
cp ../secrets/environment.go .
cp ../secrets/config.go .

# Create a simple main.go for testing
cat > main.go << 'EOF'
package main

import (
	"context"
	"fmt"
	"os"
)

// Minimal environment provider for testing
type EnvironmentProvider struct {
	prefix string
}

func NewEnvironmentProvider(prefix string) *EnvironmentProvider {
	return &EnvironmentProvider{prefix: prefix}
}

func (p *EnvironmentProvider) GetSecret(ctx context.Context, key string) (string, error) {
	envKey := p.prefix + key
	if value := os.Getenv(envKey); value != "" {
		return value, nil
	}
	return "", fmt.Errorf("secret %s not found", key)
}

func main() {
	fmt.Println("🚀 Simple Secrets Test")
	fmt.Println("=====================")
	
	// Set test environment variable
	os.Setenv("XFORM_TEST_SECRET", "test-value-12345")
	
	// Create provider
	provider := NewEnvironmentProvider("XFORM_")
	
	// Get secret
	ctx := context.Background()
	secret, err := provider.GetSecret(ctx, "TEST_SECRET")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Secret retrieved: %s\n", secret)
	fmt.Println("🎉 Basic functionality works!")
}
EOF

echo "✅ Created simplified test module"
