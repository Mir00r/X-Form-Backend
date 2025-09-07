package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kamkaiz/x-form-backend/shared/secrets"
)

func main() {
	fmt.Println("🚀 X-Form Secrets Management - Direct Test")
	fmt.Println("==========================================")

	// Test 1: Direct environment provider test
	fmt.Println("\n1️⃣ Testing Environment Provider Directly...")
	os.Setenv("XFORM_TEST_SECRET", "test-value-12345")

	envConfig := secrets.EnvironmentConfig{
		Prefix: "XFORM_",
	}

	envProvider, err := secrets.NewEnvironmentProvider(envConfig)
	if err != nil {
		fmt.Printf("   ❌ Failed to create environment provider: %v\n", err)
		return
	}
	fmt.Println("   ✅ Environment provider created successfully")

	// Test 2: Simple secret retrieval
	fmt.Println("\n2️⃣ Testing Secret Retrieval...")
	ctx := context.Background()
	secret, err := envProvider.GetSecret(ctx, "test-secret")
	if err != nil {
		fmt.Printf("   ❌ Failed to get secret: %v\n", err)
		return
	}
	fmt.Printf("   ✅ Secret retrieved: %s\n", secret)

	// Test 3: Test health check
	fmt.Println("\n3️⃣ Testing Health Check...")
	err = envProvider.HealthCheck(ctx)
	if err != nil {
		fmt.Printf("   ❌ Health check failed: %v\n", err)
		return
	}
	fmt.Println("   ✅ Health check passed")

	// Test 4: Test close
	fmt.Println("\n4️⃣ Testing Provider Close...")
	err = envProvider.Close()
	if err != nil {
		fmt.Printf("   ❌ Close failed: %v\n", err)
		return
	}
	fmt.Println("   ✅ Provider closed successfully")

	// Final summary
	fmt.Println("\n📋 Direct Test Summary:")
	fmt.Println("   ✅ Environment Provider Creation: Working")
	fmt.Println("   ✅ Secret Retrieval: Working")
	fmt.Println("   ✅ Health Check: Working")
	fmt.Println("   ✅ Provider Cleanup: Working")
	fmt.Println("\n🎉 Core functionality is working correctly!")
	fmt.Println("   The issue is likely in the SecretManager initialization.")
}
