package secrets

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestSecretManager tests the basic functionality of SecretManager
func TestSecretManager(t *testing.T) {
	// Create a mock provider
	mockSecrets := map[string]string{
		"test-key":    "test-value",
		"db-password": "super-secret",
	}

	mockProvider := NewMockProvider(mockSecrets)

	// Create config with mock provider
	config := Config{
		Provider: ProviderTypeEnvironment, // Will be overridden by factory
		Cache: CacheConfig{
			Enabled:    true,
			TTL:        1 * time.Minute,
			MaxEntries: 100,
		},
	}

	// Create secret manager with custom factory
	sm := &SecretManager{
		config:  config,
		primary: mockProvider,
		cache:   newSecretCache(config.Cache),
	}

	ctx := context.Background()

	// Test GetSecret
	value, err := sm.GetSecret(ctx, "test-key")
	if err != nil {
		t.Fatalf("Failed to get secret: %v", err)
	}
	if value != "test-value" {
		t.Fatalf("Expected 'test-value', got '%s'", value)
	}

	// Test GetSecrets
	secrets, err := sm.GetSecrets(ctx, []string{"test-key", "db-password"})
	if err != nil {
		t.Fatalf("Failed to get secrets: %v", err)
	}
	if len(secrets) != 2 {
		t.Fatalf("Expected 2 secrets, got %d", len(secrets))
	}

	// Test SetSecret
	err = sm.SetSecret(ctx, "new-key", "new-value", nil)
	if err != nil {
		t.Fatalf("Failed to set secret: %v", err)
	}

	// Verify the secret was set
	value, err = sm.GetSecret(ctx, "new-key")
	if err != nil {
		t.Fatalf("Failed to get newly set secret: %v", err)
	}
	if value != "new-value" {
		t.Fatalf("Expected 'new-value', got '%s'", value)
	}

	// Test cache
	stats := sm.GetCacheStats()
	if stats == nil {
		t.Fatal("Expected cache stats, got nil")
	}

	// Test HealthCheck
	err = sm.HealthCheck(ctx)
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}

	fmt.Println("✅ All SecretManager tests passed!")
}

// TestProviderFallback tests fallback provider functionality
func TestProviderFallback(t *testing.T) {
	// Create primary provider that fails
	primaryProvider := NewMockProvider(map[string]string{})
	primaryProvider.SetHealthy(false)

	// Create fallback provider that works
	fallbackProvider := NewMockProvider(map[string]string{
		"fallback-key": "fallback-value",
	})

	config := Config{
		Cache: CacheConfig{
			Enabled:    true,
			TTL:        1 * time.Minute,
			MaxEntries: 100,
		},
	}

	sm := &SecretManager{
		config:    config,
		primary:   primaryProvider,
		fallbacks: []SecretProvider{fallbackProvider},
		cache:     newSecretCache(config.Cache),
	}

	ctx := context.Background()

	// Test fallback
	value, err := sm.GetSecret(ctx, "fallback-key")
	if err != nil {
		t.Fatalf("Failed to get secret from fallback: %v", err)
	}
	if value != "fallback-value" {
		t.Fatalf("Expected 'fallback-value', got '%s'", value)
	}

	fmt.Println("✅ Provider fallback test passed!")
}

// TestCache tests the cache functionality
func TestCache(t *testing.T) {
	cache := newSecretCache(CacheConfig{
		Enabled:    true,
		TTL:        100 * time.Millisecond,
		MaxEntries: 2,
	})

	// Test Set and Get
	cache.Set("key1", "value1")
	value, found := cache.Get("key1")
	if !found {
		t.Fatal("Expected to find key1 in cache")
	}
	if value != "value1" {
		t.Fatalf("Expected 'value1', got '%s'", value)
	}

	// Test TTL expiration
	time.Sleep(150 * time.Millisecond)
	_, found = cache.Get("key1")
	if found {
		t.Fatal("Expected key1 to be expired")
	}

	// Test max entries eviction
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3") // Should evict oldest

	stats := cache.Stats()
	if stats["entries"].(int) > 2 {
		t.Fatalf("Expected max 2 entries, got %d", stats["entries"].(int))
	}

	fmt.Println("✅ Cache tests passed!")
}

// TestEnvironmentProvider tests the environment variable provider
func TestEnvironmentProvider(t *testing.T) {
	config := EnvironmentConfig{
		Prefix:        "TEST_",
		CaseSensitive: false,
	}

	provider, err := NewEnvironmentProvider(config)
	if err != nil {
		t.Fatalf("Failed to create environment provider: %v", err)
	}

	ctx := context.Background()

	// Set a test environment variable
	err = provider.SetSecret(ctx, "SECRET_KEY", "test-secret-value", nil)
	if err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}

	// Get the secret
	value, err := provider.GetSecret(ctx, "SECRET_KEY")
	if err != nil {
		t.Fatalf("Failed to get environment variable: %v", err)
	}
	if value != "test-secret-value" {
		t.Fatalf("Expected 'test-secret-value', got '%s'", value)
	}

	// Test health check
	err = provider.HealthCheck(ctx)
	if err != nil {
		t.Fatalf("Environment provider health check failed: %v", err)
	}

	fmt.Println("✅ Environment provider tests passed!")
}

// BenchmarkSecretManager benchmarks the secret manager performance
func BenchmarkSecretManager(b *testing.B) {
	mockSecrets := make(map[string]string)
	for i := 0; i < 1000; i++ {
		mockSecrets[fmt.Sprintf("key-%d", i)] = fmt.Sprintf("value-%d", i)
	}

	mockProvider := NewMockProvider(mockSecrets)
	config := Config{
		Cache: CacheConfig{
			Enabled:    true,
			TTL:        5 * time.Minute,
			MaxEntries: 1000,
		},
	}

	sm := &SecretManager{
		config:  config,
		primary: mockProvider,
		cache:   newSecretCache(config.Cache),
	}

	ctx := context.Background()

	b.ResetTimer()

	b.Run("GetSecret", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key-%d", i%1000)
			_, err := sm.GetSecret(ctx, key)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("GetSecrets", func(b *testing.B) {
		keys := []string{"key-1", "key-2", "key-3", "key-4", "key-5"}
		for i := 0; i < b.N; i++ {
			_, err := sm.GetSecrets(ctx, keys)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Example usage
func ExampleSecretManager() {
	// Create default configuration
	config := GetDefaultConfig()
	config.Provider = ProviderTypeEnvironment

	// Create secret manager
	sm, err := NewSecretManager(*config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create secret manager: %v", err))
	}
	defer sm.Close()

	ctx := context.Background()

	// Get a secret
	secret, err := sm.GetSecret(ctx, "DATABASE_PASSWORD")
	if err != nil {
		fmt.Printf("Failed to get secret: %v\n", err)
		return
	}

	fmt.Printf("Secret retrieved: %s\n", secret)

	// Get multiple secrets
	secrets, err := sm.GetSecrets(ctx, []string{"DATABASE_PASSWORD", "API_KEY", "JWT_SECRET"})
	if err != nil {
		fmt.Printf("Failed to get secrets: %v\n", err)
		return
	}

	fmt.Printf("Retrieved %d secrets\n", len(secrets))
}

// RunTests runs all tests
func RunTests() {
	fmt.Println("🚀 Running secrets management tests...")

	// Run basic tests (simplified since we can't use testing.T)
	fmt.Println("📋 Testing basic functionality...")

	// Test cache
	fmt.Println("🔄 Testing cache...")
	cache := newSecretCache(CacheConfig{
		Enabled:    true,
		TTL:        1 * time.Second,
		MaxEntries: 10,
	})

	cache.Set("test", "value")
	if value, found := cache.Get("test"); !found || value != "value" {
		panic("Cache test failed")
	}

	// Test mock provider
	fmt.Println("🔌 Testing mock provider...")
	mockProvider := NewMockProvider(map[string]string{
		"test": "value",
	})

	ctx := context.Background()
	value, err := mockProvider.GetSecret(ctx, "test")
	if err != nil || value != "value" {
		panic("Mock provider test failed")
	}

	// Test environment provider
	fmt.Println("🌍 Testing environment provider...")
	envProvider, err := NewEnvironmentProvider(EnvironmentConfig{})
	if err != nil {
		panic(fmt.Sprintf("Failed to create environment provider: %v", err))
	}

	if err := envProvider.HealthCheck(ctx); err != nil {
		panic(fmt.Sprintf("Environment provider health check failed: %v", err))
	}

	fmt.Println("✅ All tests passed successfully!")
	fmt.Println("🎉 Secrets management module is ready for production use!")
}
