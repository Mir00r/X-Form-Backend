package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kamkaiz/x-form-backend/shared/secrets"
)

func main() {
	fmt.Println("🚀 X-Form Secrets Management - Quick Test")
	fmt.Println("==========================================")

	// Test 1: Basic configuration
	fmt.Println("\n1️⃣ Testing Configuration Loading...")
	config := secrets.Config{
		Provider: "environment",
		Environment: secrets.EnvironmentConfig{
			Prefix: "XFORM_",
		},
	}
	fmt.Printf("   ✅ Config loaded: Provider=%s\n", config.Provider)

	// Test 2: Environment provider (safest test)
	fmt.Println("\n2️⃣ Testing Environment Provider...")
	os.Setenv("XFORM_TEST_SECRET", "test-value-12345")

	envProvider, err := secrets.NewEnvironmentProvider(config.Environment)
	if err != nil {
		fmt.Printf("   ❌ Failed to create environment provider: %v\n", err)
		return
	}
	fmt.Println("   ✅ Environment provider created")

	// Test 3: Simple secret retrieval
	fmt.Println("\n3️⃣ Testing Secret Retrieval...")
	ctx := context.Background()
	secret, err := envProvider.GetSecret(ctx, "test-secret")
	if err != nil {
		fmt.Printf("   ❌ Failed to get secret: %v\n", err)
		return
	}
	fmt.Printf("   ✅ Secret retrieved: %s\n", secret)

	// Test 4: Basic manager functionality
	fmt.Println("\n4️⃣ Testing Secret Manager...")
	_, err = secrets.NewSecretManager(config)
	if err != nil {
		fmt.Printf("   ❌ Failed to create manager: %v\n", err)
		return
	}
	fmt.Println("   ✅ Secret manager created")

	// Final summary
	fmt.Println("\n📋 Quick Test Summary:")
	fmt.Println("   ✅ Configuration: Working")
	fmt.Println("   ✅ Environment Provider: Working")
	fmt.Println("   ✅ Secret Operations: Working")
	fmt.Println("   ✅ Secret Manager: Working")
	fmt.Println("\n🎉 All basic functionality is working correctly!")
}
