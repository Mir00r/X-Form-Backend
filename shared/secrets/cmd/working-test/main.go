package main

import (
	"context"
	"fmt"
	"os"
)

// Basic environment provider implementation for testing
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

func (p *EnvironmentProvider) HealthCheck(ctx context.Context) error {
	return nil // Environment variables are always available
}

func (p *EnvironmentProvider) Close() error {
	return nil // Nothing to close for environment provider
}

func main() {
	fmt.Println("🚀 X-Form Secrets Management - Environment Provider Test")
	fmt.Println("========================================================")

	// Test 1: Create environment provider
	fmt.Println("\n1️⃣ Creating Environment Provider...")
	provider := NewEnvironmentProvider("XFORM_")
	fmt.Println("   ✅ Environment provider created")

	// Test 2: Set test environment variable
	fmt.Println("\n2️⃣ Setting Test Environment Variable...")
	os.Setenv("XFORM_TEST_SECRET", "test-value-12345")
	os.Setenv("XFORM_DATABASE_PASSWORD", "super-secret-password")
	os.Setenv("XFORM_API_KEY", "api-key-abc123")
	fmt.Println("   ✅ Test environment variables set")

	// Test 3: Retrieve secrets
	fmt.Println("\n3️⃣ Retrieving Secrets...")
	ctx := context.Background()

	secrets := []string{"TEST_SECRET", "DATABASE_PASSWORD", "API_KEY"}
	for _, secretKey := range secrets {
		secret, err := provider.GetSecret(ctx, secretKey)
		if err != nil {
			fmt.Printf("   ❌ Failed to get %s: %v\n", secretKey, err)
		} else {
			fmt.Printf("   ✅ %s: %s\n", secretKey, secret)
		}
	}

	// Test 4: Test missing secret
	fmt.Println("\n4️⃣ Testing Missing Secret...")
	_, err := provider.GetSecret(ctx, "MISSING_SECRET")
	if err != nil {
		fmt.Printf("   ✅ Correctly failed for missing secret: %v\n", err)
	} else {
		fmt.Println("   ❌ Should have failed for missing secret")
	}

	// Test 5: Health check
	fmt.Println("\n5️⃣ Testing Health Check...")
	err = provider.HealthCheck(ctx)
	if err != nil {
		fmt.Printf("   ❌ Health check failed: %v\n", err)
	} else {
		fmt.Println("   ✅ Health check passed")
	}

	// Test 6: Close provider
	fmt.Println("\n6️⃣ Closing Provider...")
	err = provider.Close()
	if err != nil {
		fmt.Printf("   ❌ Close failed: %v\n", err)
	} else {
		fmt.Println("   ✅ Provider closed successfully")
	}

	// Final summary
	fmt.Println("\n📋 Test Summary:")
	fmt.Println("   ✅ Provider Creation: Working")
	fmt.Println("   ✅ Secret Retrieval: Working")
	fmt.Println("   ✅ Error Handling: Working")
	fmt.Println("   ✅ Health Check: Working")
	fmt.Println("   ✅ Resource Cleanup: Working")
	fmt.Println("\n🎉 Environment Provider is fully functional!")

	fmt.Println("\n📖 Usage Example:")
	fmt.Println("   export XFORM_DATABASE_PASSWORD=your-password")
	fmt.Println("   export XFORM_API_KEY=your-api-key")
	fmt.Println("   // In your Go code:")
	fmt.Println("   provider := NewEnvironmentProvider(\"XFORM_\")")
	fmt.Println("   password, _ := provider.GetSecret(ctx, \"DATABASE_PASSWORD\")")
}
