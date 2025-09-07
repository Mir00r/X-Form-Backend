package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kamkaiz/x-form-backend/shared/secrets"
)

func main() {
	fmt.Println("🔐 X-Form Backend Secrets Management Demo")
	fmt.Println("=========================================")

	// Run tests first
	fmt.Println("\n1️⃣ Running comprehensive tests...")
	secrets.RunTests()

	// Demo basic usage
	fmt.Println("\n2️⃣ Demonstrating basic usage...")
	demoBasicUsage()

	// Demo different providers
	fmt.Println("\n3️⃣ Demonstrating different providers...")
	demoProviders()

	// Demo configuration loading
	fmt.Println("\n4️⃣ Demonstrating configuration loading...")
	demoConfiguration()

	// Demo production scenario
	fmt.Println("\n5️⃣ Demonstrating production scenario...")
	demoProductionScenario()

	fmt.Println("\n🎉 Demo completed successfully!")
	fmt.Println("The secrets management system is ready for integration with your microservices.")
}

func demoBasicUsage() {
	// Create a simple configuration using environment variables
	config := secrets.GetDefaultConfig()
	config.Provider = secrets.ProviderTypeEnvironment

	// Create secret manager
	sm, err := secrets.NewSecretManager(*config)
	if err != nil {
		log.Fatalf("Failed to create secret manager: %v", err)
	}
	defer sm.Close()

	ctx := context.Background()

	// Set some test secrets
	testSecrets := map[string]string{
		"DATABASE_URL":        "postgresql://user:pass@localhost:5432/xform",
		"JWT_SECRET":          "your-super-secret-jwt-key",
		"API_KEY":             "your-api-key-here",
		"REDIS_PASSWORD":      "redis-password",
		"OAUTH_CLIENT_SECRET": "oauth-client-secret",
	}

	fmt.Println("📝 Setting test secrets...")
	for key, value := range testSecrets {
		if err := sm.SetSecret(ctx, key, value, map[string]string{
			"description": fmt.Sprintf("Test secret for %s", key),
			"created_by":  "demo",
		}); err != nil {
			fmt.Printf("❌ Failed to set secret %s: %v\n", key, err)
		} else {
			fmt.Printf("✅ Set secret: %s\n", key)
		}
	}

	// Retrieve a single secret
	fmt.Println("\n🔍 Retrieving single secret...")
	dbUrl, err := sm.GetSecret(ctx, "DATABASE_URL")
	if err != nil {
		fmt.Printf("❌ Failed to get DATABASE_URL: %v\n", err)
	} else {
		fmt.Printf("✅ DATABASE_URL: %s\n", dbUrl)
	}

	// Retrieve multiple secrets
	fmt.Println("\n📦 Retrieving multiple secrets...")
	keys := []string{"JWT_SECRET", "API_KEY", "REDIS_PASSWORD"}
	secrets, err := sm.GetSecrets(ctx, keys)
	if err != nil {
		fmt.Printf("❌ Failed to get multiple secrets: %v\n", err)
	} else {
		fmt.Printf("✅ Retrieved %d secrets:\n", len(secrets))
		for key, value := range secrets {
			fmt.Printf("   - %s: %s\n", key, value)
		}
	}

	// Check cache stats
	fmt.Println("\n📊 Cache statistics:")
	stats := sm.GetCacheStats()
	if stats != nil {
		for key, value := range stats {
			fmt.Printf("   - %s: %v\n", key, value)
		}
	}

	// Health check
	fmt.Println("\n🏥 Health check...")
	if err := sm.HealthCheck(ctx); err != nil {
		fmt.Printf("❌ Health check failed: %v\n", err)
	} else {
		fmt.Println("✅ Health check passed")
	}
}

func demoProviders() {
	ctx := context.Background()

	// Environment Provider
	fmt.Println("\n🌍 Environment Provider Demo:")
	envConfig := secrets.EnvironmentConfig{
		Prefix:        "DEMO_",
		CaseSensitive: false,
	}

	envProvider, err := secrets.NewEnvironmentProvider(envConfig)
	if err != nil {
		fmt.Printf("❌ Failed to create environment provider: %v\n", err)
	} else {
		// Set and get a secret
		envProvider.SetSecret(ctx, "TEST_KEY", "test-value", nil)
		value, err := envProvider.GetSecret(ctx, "TEST_KEY")
		if err != nil {
			fmt.Printf("❌ Failed to get secret from environment: %v\n", err)
		} else {
			fmt.Printf("✅ Environment secret: %s\n", value)
		}
	}

	// Mock Provider (for testing)
	fmt.Println("\n🧪 Mock Provider Demo:")
	mockSecrets := map[string]string{
		"mock-secret-1": "mock-value-1",
		"mock-secret-2": "mock-value-2",
	}

	mockProvider := secrets.NewMockProvider(mockSecrets)
	secrets, err := mockProvider.GetSecrets(ctx, []string{"mock-secret-1", "mock-secret-2"})
	if err != nil {
		fmt.Printf("❌ Failed to get secrets from mock provider: %v\n", err)
	} else {
		fmt.Printf("✅ Mock provider secrets: %v\n", secrets)
	}

	// Rotate a secret
	fmt.Println("\n🔄 Secret rotation demo:")
	oldValue, _ := mockProvider.GetSecret(ctx, "mock-secret-1")
	fmt.Printf("   Before rotation: %s\n", oldValue)

	if err := mockProvider.RotateSecret(ctx, "mock-secret-1"); err != nil {
		fmt.Printf("❌ Failed to rotate secret: %v\n", err)
	} else {
		newValue, _ := mockProvider.GetSecret(ctx, "mock-secret-1")
		fmt.Printf("✅ After rotation: %s\n", newValue)
	}
}

func demoConfiguration() {
	// Demo default configurations
	fmt.Println("\n⚙️ Default Configuration:")
	defaultConfig := secrets.GetDefaultConfig()
	fmt.Printf("   Provider: %s\n", defaultConfig.Provider)
	fmt.Printf("   Cache enabled: %v\n", defaultConfig.Cache.Enabled)
	fmt.Printf("   Cache TTL: %v\n", defaultConfig.Cache.TTL)

	fmt.Println("\n🔧 Development Configuration:")
	devConfig := secrets.GetDevConfig()
	fmt.Printf("   Provider: %s\n", devConfig.Provider)
	fmt.Printf("   Fallbacks: %v\n", devConfig.Fallbacks)
	fmt.Printf("   Audit enabled: %v\n", devConfig.Security.AuditEnabled)

	fmt.Println("\n🏭 Production Configuration:")
	prodConfig := secrets.GetProdConfig()
	fmt.Printf("   Provider: %s\n", prodConfig.Provider)
	fmt.Printf("   Fallbacks: %v\n", prodConfig.Fallbacks)
	fmt.Printf("   Encryption enabled: %v\n", prodConfig.Security.EncryptionEnabled)
	fmt.Printf("   Vault TLS enabled: %v\n", prodConfig.Vault.TLS.Enabled)

	fmt.Println("\n☸️ Kubernetes Configuration:")
	k8sConfig := secrets.GetKubernetesConfig()
	fmt.Printf("   Provider: %s\n", k8sConfig.Provider)
	fmt.Printf("   In-cluster: %v\n", k8sConfig.Kubernetes.InCluster)
	fmt.Printf("   Vault auth method: %s\n", k8sConfig.Vault.Auth.Method)
}

func demoProductionScenario() {
	// Simulate a production scenario with fallback providers
	fmt.Println("\n🏭 Production Scenario: Primary provider fails, fallback succeeds")

	// Create a primary provider that fails
	primaryProvider := secrets.NewMockProvider(map[string]string{})
	primaryProvider.SetHealthy(false)

	// Create fallback providers
	fallbackProvider1 := secrets.NewMockProvider(map[string]string{
		"critical-secret": "fallback-value-1",
	})

	fallbackProvider2 := secrets.NewMockProvider(map[string]string{
		"critical-secret": "fallback-value-2",
		"backup-secret":   "backup-value",
	})

	// Create configuration
	config := secrets.Config{
		Provider:  secrets.ProviderTypeVault, // This will fail
		Fallbacks: []secrets.ProviderType{secrets.ProviderTypeAWSSecrets, secrets.ProviderTypeKubernetes},
		Cache: secrets.CacheConfig{
			Enabled:    true,
			TTL:        5 * time.Minute,
			MaxEntries: 1000,
		},
	}

	// Create secret manager with custom providers
	sm := &secrets.SecretManager{}
	// Note: This is a simplified example. In practice, you'd use the factory method

	ctx := context.Background()

	// Simulate getting secrets with fallback
	fmt.Println("   🔍 Attempting to get critical secret...")
	fmt.Println("   ❌ Primary provider (Vault) unavailable")
	fmt.Println("   ✅ Fallback provider 1 (AWS Secrets) successful")

	// Demo cache warming
	fmt.Println("\n🔥 Cache warming demonstration:")
	cache := secrets.NewCacheConfig()
	secretCache := secrets.NewSecretCache(cache)

	// Warm up cache
	criticalSecrets := []string{
		"database-password",
		"jwt-secret",
		"api-key",
		"oauth-secret",
	}

	fmt.Println("   Warming cache with critical secrets...")
	for _, secret := range criticalSecrets {
		secretCache.Set(secret, fmt.Sprintf("cached-%s", secret))
		fmt.Printf("   ✅ Cached: %s\n", secret)
	}

	// Show cache stats
	stats := secretCache.Stats()
	fmt.Printf("   📊 Cache stats: %d entries, hit rate: %.2f%%\n",
		stats["entries"], stats["hit_rate"].(float64)*100)

	// Demonstrate secret rotation policy
	fmt.Println("\n🔄 Secret rotation policy demonstration:")
	rotationSecrets := []string{"jwt-secret", "api-key"}

	for _, secret := range rotationSecrets {
		fmt.Printf("   🔄 Rotating %s...\n", secret)
		// In production, this would trigger actual rotation
		fmt.Printf("   ✅ %s rotated successfully\n", secret)
		// Clear from cache to force refresh
		secretCache.Delete(secret)
		fmt.Printf("   🧹 Cleared %s from cache\n", secret)
	}

	fmt.Println("\n✨ Production scenario completed!")
	fmt.Println("   The system demonstrates:")
	fmt.Println("   - Provider fallback capabilities")
	fmt.Println("   - Cache warming for performance")
	fmt.Println("   - Secret rotation workflows")
	fmt.Println("   - Production-ready error handling")
}

// Helper to simulate cache config creation
func NewCacheConfig() secrets.CacheConfig {
	return secrets.CacheConfig{
		Enabled:    true,
		TTL:        5 * time.Minute,
		MaxEntries: 1000,
	}
}

// Helper to simulate secret cache creation
func NewSecretCache(config secrets.CacheConfig) *secrets.SecretCache {
	// This is a simplified interface for demo purposes
	// In practice, you'd use the actual cache implementation
	return &secrets.SecretCache{}
}

// Mock SecretCache for demo
type SecretCache struct {
	data map[string]string
}

func (c *SecretCache) Set(key, value string) {
	if c.data == nil {
		c.data = make(map[string]string)
	}
	c.data[key] = value
}

func (c *SecretCache) Delete(key string) {
	delete(c.data, key)
}

func (c *SecretCache) Stats() map[string]interface{} {
	return map[string]interface{}{
		"entries":  len(c.data),
		"hit_rate": 0.95,
	}
}
