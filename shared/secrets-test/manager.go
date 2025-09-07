package secrets

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// SecretProvider defines the interface for secret management providers
type SecretProvider interface {
	// GetSecret retrieves a secret by key
	GetSecret(ctx context.Context, key string) (string, error)
	
	// GetSecrets retrieves multiple secrets by keys
	GetSecrets(ctx context.Context, keys []string) (map[string]string, error)
	
	// SetSecret stores a secret (if supported)
	SetSecret(ctx context.Context, key, value string, metadata map[string]string) error
	
	// DeleteSecret removes a secret (if supported)
	DeleteSecret(ctx context.Context, key string) error
	
	// ListSecrets lists available secret keys with optional prefix
	ListSecrets(ctx context.Context, prefix string) ([]string, error)
	
	// RotateSecret rotates a secret (if supported)
	RotateSecret(ctx context.Context, key string) error
	
	// HealthCheck verifies the provider is accessible
	HealthCheck(ctx context.Context) error
	
	// Close closes the provider connection
	Close() error
}

// ProviderType represents the type of secret provider
type ProviderType string

const (
	ProviderTypeVault       ProviderType = "vault"
	ProviderTypeAWSSecrets  ProviderType = "aws-secrets"
	ProviderTypeAWSSSM      ProviderType = "aws-ssm"
	ProviderTypeKubernetes  ProviderType = "kubernetes"
	ProviderTypeEnvironment ProviderType = "environment"
	ProviderTypeFile        ProviderType = "file"
)

// Config holds the configuration for the secrets manager
type Config struct {
	// Primary provider configuration
	Provider ProviderType `json:"provider" yaml:"provider" mapstructure:"provider"`
	
	// Fallback providers (in order of preference)
	Fallbacks []ProviderType `json:"fallbacks" yaml:"fallbacks" mapstructure:"fallbacks"`
	
	// Cache configuration
	Cache CacheConfig `json:"cache" yaml:"cache" mapstructure:"cache"`
	
	// Vault-specific configuration
	Vault VaultConfig `json:"vault" yaml:"vault" mapstructure:"vault"`
	
	// AWS-specific configuration
	AWS AWSConfig `json:"aws" yaml:"aws" mapstructure:"aws"`
	
	// Kubernetes-specific configuration
	Kubernetes KubernetesConfig `json:"kubernetes" yaml:"kubernetes" mapstructure:"kubernetes"`
	
	// Environment-specific configuration
	Environment EnvironmentConfig `json:"environment" yaml:"environment" mapstructure:"environment"`
	
	// File-specific configuration
	File FileConfig `json:"file" yaml:"file" mapstructure:"file"`
	
	// Security settings
	Security SecurityConfig `json:"security" yaml:"security" mapstructure:"security"`
}

// CacheConfig holds cache-related configuration
type CacheConfig struct {
	Enabled    bool          `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	TTL        time.Duration `json:"ttl" yaml:"ttl" mapstructure:"ttl"`
	MaxEntries int           `json:"max_entries" yaml:"max_entries" mapstructure:"max_entries"`
}

// VaultConfig holds Vault-specific configuration
type VaultConfig struct {
	Address    string            `json:"address" yaml:"address" mapstructure:"address"`
	Token      string            `json:"token" yaml:"token" mapstructure:"token"`
	TokenPath  string            `json:"token_path" yaml:"token_path" mapstructure:"token_path"`
	Namespace  string            `json:"namespace" yaml:"namespace" mapstructure:"namespace"`
	MountPath  string            `json:"mount_path" yaml:"mount_path" mapstructure:"mount_path"`
	TLS        VaultTLSConfig    `json:"tls" yaml:"tls" mapstructure:"tls"`
	Auth       VaultAuthConfig   `json:"auth" yaml:"auth" mapstructure:"auth"`
	Headers    map[string]string `json:"headers" yaml:"headers" mapstructure:"headers"`
}

// VaultTLSConfig holds Vault TLS configuration
type VaultTLSConfig struct {
	Enabled            bool   `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	CACert             string `json:"ca_cert" yaml:"ca_cert" mapstructure:"ca_cert"`
	ClientCert         string `json:"client_cert" yaml:"client_cert" mapstructure:"client_cert"`
	ClientKey          string `json:"client_key" yaml:"client_key" mapstructure:"client_key"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify" yaml:"insecure_skip_verify" mapstructure:"insecure_skip_verify"`
}

// VaultAuthConfig holds Vault authentication configuration
type VaultAuthConfig struct {
	Method     string            `json:"method" yaml:"method" mapstructure:"method"`
	Parameters map[string]string `json:"parameters" yaml:"parameters" mapstructure:"parameters"`
}

// AWSConfig holds AWS-specific configuration
type AWSConfig struct {
	Region          string `json:"region" yaml:"region" mapstructure:"region"`
	AccessKeyID     string `json:"access_key_id" yaml:"access_key_id" mapstructure:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key" yaml:"secret_access_key" mapstructure:"secret_access_key"`
	SessionToken    string `json:"session_token" yaml:"session_token" mapstructure:"session_token"`
	Profile         string `json:"profile" yaml:"profile" mapstructure:"profile"`
	RoleARN         string `json:"role_arn" yaml:"role_arn" mapstructure:"role_arn"`
	SSMPath         string `json:"ssm_path" yaml:"ssm_path" mapstructure:"ssm_path"`
}

// KubernetesConfig holds Kubernetes-specific configuration
type KubernetesConfig struct {
	Namespace   string `json:"namespace" yaml:"namespace" mapstructure:"namespace"`
	SecretName  string `json:"secret_name" yaml:"secret_name" mapstructure:"secret_name"`
	ConfigPath  string `json:"config_path" yaml:"config_path" mapstructure:"config_path"`
	InCluster   bool   `json:"in_cluster" yaml:"in_cluster" mapstructure:"in_cluster"`
}

// EnvironmentConfig holds environment variable configuration
type EnvironmentConfig struct {
	Prefix          string            `json:"prefix" yaml:"prefix" mapstructure:"prefix"`
	Mapping         map[string]string `json:"mapping" yaml:"mapping" mapstructure:"mapping"`
	CaseSensitive   bool              `json:"case_sensitive" yaml:"case_sensitive" mapstructure:"case_sensitive"`
}

// FileConfig holds file-based configuration
type FileConfig struct {
	Path      string `json:"path" yaml:"path" mapstructure:"path"`
	Format    string `json:"format" yaml:"format" mapstructure:"format"` // json, yaml, properties
	Encrypted bool   `json:"encrypted" yaml:"encrypted" mapstructure:"encrypted"`
	KeyPath   string `json:"key_path" yaml:"key_path" mapstructure:"key_path"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	EncryptionEnabled bool   `json:"encryption_enabled" yaml:"encryption_enabled" mapstructure:"encryption_enabled"`
	EncryptionKey     string `json:"encryption_key" yaml:"encryption_key" mapstructure:"encryption_key"`
	AuditEnabled      bool   `json:"audit_enabled" yaml:"audit_enabled" mapstructure:"audit_enabled"`
	AuditPath         string `json:"audit_path" yaml:"audit_path" mapstructure:"audit_path"`
}

// SecretManager manages secrets from multiple providers with fallback support
type SecretManager struct {
	config    Config
	primary   SecretProvider
	fallbacks []SecretProvider
	cache     *secretCache
	logger    *logrus.Logger
	mu        sync.RWMutex
}

// NewSecretManager creates a new secret manager with the given configuration
func NewSecretManager(config Config) (*SecretManager, error) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	
	sm := &SecretManager{
		config: config,
		logger: logger,
	}
	
	// Initialize cache if enabled
	if config.Cache.Enabled {
		sm.cache = newSecretCache(config.Cache)
	}
	
	// Initialize primary provider
	primary, err := sm.createProvider(config.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create primary provider %s: %w", config.Provider, err)
	}
	sm.primary = primary
	
	// Initialize fallback providers
	for _, providerType := range config.Fallbacks {
		provider, err := sm.createProvider(providerType)
		if err != nil {
			logger.Warnf("Failed to create fallback provider %s: %v", providerType, err)
			continue
		}
		sm.fallbacks = append(sm.fallbacks, provider)
	}
	
	return sm, nil
}

// GetSecret retrieves a secret with fallback support and caching
func (sm *SecretManager) GetSecret(ctx context.Context, key string) (string, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	// Check cache first
	if sm.cache != nil {
		if value, found := sm.cache.Get(key); found {
			sm.logger.Debugf("Secret cache hit for key: %s", key)
			return value, nil
		}
	}
	
	// Try primary provider
	value, err := sm.tryProvider(ctx, sm.primary, "primary", key)
	if err == nil {
		// Cache the value
		if sm.cache != nil {
			sm.cache.Set(key, value)
		}
		return value, nil
	}
	
	sm.logger.Warnf("Primary provider failed for key %s: %v", key, err)
	
	// Try fallback providers
	for i, provider := range sm.fallbacks {
		value, err := sm.tryProvider(ctx, provider, fmt.Sprintf("fallback-%d", i), key)
		if err == nil {
			// Cache the value
			if sm.cache != nil {
				sm.cache.Set(key, value)
			}
			return value, nil
		}
		sm.logger.Warnf("Fallback provider %d failed for key %s: %v", i, key, err)
	}
	
	return "", fmt.Errorf("all providers failed to retrieve secret for key: %s", key)
}

// GetSecrets retrieves multiple secrets efficiently
func (sm *SecretManager) GetSecrets(ctx context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string)
	var missingKeys []string
	
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	// Check cache for all keys first
	if sm.cache != nil {
		for _, key := range keys {
			if value, found := sm.cache.Get(key); found {
				result[key] = value
			} else {
				missingKeys = append(missingKeys, key)
			}
		}
	} else {
		missingKeys = keys
	}
	
	if len(missingKeys) == 0 {
		return result, nil
	}
	
	// Try to get missing keys from primary provider
	if secrets, err := sm.primary.GetSecrets(ctx, missingKeys); err == nil {
		for key, value := range secrets {
			result[key] = value
			if sm.cache != nil {
				sm.cache.Set(key, value)
			}
		}
		return result, nil
	}
	
	// Fallback to individual secret retrieval
	for _, key := range missingKeys {
		if value, err := sm.GetSecret(ctx, key); err == nil {
			result[key] = value
		}
	}
	
	return result, nil
}

// SetSecret stores a secret using the primary provider
func (sm *SecretManager) SetSecret(ctx context.Context, key, value string, metadata map[string]string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if err := sm.primary.SetSecret(ctx, key, value, metadata); err != nil {
		return fmt.Errorf("failed to set secret %s: %w", key, err)
	}
	
	// Update cache
	if sm.cache != nil {
		sm.cache.Set(key, value)
	}
	
	sm.logger.Infof("Successfully set secret: %s", key)
	return nil
}

// DeleteSecret removes a secret using the primary provider
func (sm *SecretManager) DeleteSecret(ctx context.Context, key string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if err := sm.primary.DeleteSecret(ctx, key); err != nil {
		return fmt.Errorf("failed to delete secret %s: %w", key, err)
	}
	
	// Remove from cache
	if sm.cache != nil {
		sm.cache.Delete(key)
	}
	
	sm.logger.Infof("Successfully deleted secret: %s", key)
	return nil
}

// RotateSecret rotates a secret using the primary provider
func (sm *SecretManager) RotateSecret(ctx context.Context, key string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if err := sm.primary.RotateSecret(ctx, key); err != nil {
		return fmt.Errorf("failed to rotate secret %s: %w", key, err)
	}
	
	// Clear from cache to force refresh
	if sm.cache != nil {
		sm.cache.Delete(key)
	}
	
	sm.logger.Infof("Successfully rotated secret: %s", key)
	return nil
}

// HealthCheck checks the health of all providers
func (sm *SecretManager) HealthCheck(ctx context.Context) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	// Check primary provider
	if err := sm.primary.HealthCheck(ctx); err != nil {
		return fmt.Errorf("primary provider health check failed: %w", err)
	}
	
	// Check fallback providers (non-blocking)
	for i, provider := range sm.fallbacks {
		if err := provider.HealthCheck(ctx); err != nil {
			sm.logger.Warnf("Fallback provider %d health check failed: %v", i, err)
		}
	}
	
	return nil
}

// Close closes all provider connections
func (sm *SecretManager) Close() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	var errors []error
	
	if err := sm.primary.Close(); err != nil {
		errors = append(errors, fmt.Errorf("primary provider close failed: %w", err))
	}
	
	for i, provider := range sm.fallbacks {
		if err := provider.Close(); err != nil {
			errors = append(errors, fmt.Errorf("fallback provider %d close failed: %w", i, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("errors closing providers: %v", errors)
	}
	
	return nil
}

// RefreshCache clears the cache to force refresh of secrets
func (sm *SecretManager) RefreshCache() {
	if sm.cache != nil {
		sm.cache.Clear()
		sm.logger.Info("Secret cache cleared")
	}
}

// GetCacheStats returns cache statistics
func (sm *SecretManager) GetCacheStats() map[string]interface{} {
	if sm.cache != nil {
		return sm.cache.Stats()
	}
	return nil
}

// tryProvider attempts to get a secret from a specific provider
func (sm *SecretManager) tryProvider(ctx context.Context, provider SecretProvider, name, key string) (string, error) {
	sm.logger.Debugf("Trying %s provider for key: %s", name, key)
	
	value, err := provider.GetSecret(ctx, key)
	if err != nil {
		return "", err
	}
	
	sm.logger.Debugf("Successfully retrieved secret from %s provider for key: %s", name, key)
	return value, nil
}

// GenerateSecretKey generates a cryptographically secure random key
func GenerateSecretKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
