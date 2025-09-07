package secrets

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// EnvironmentProvider implements SecretProvider for environment variables
type EnvironmentProvider struct {
	config EnvironmentConfig
	logger *logrus.Logger
}

// NewEnvironmentProvider creates a new environment variable provider
func NewEnvironmentProvider(config EnvironmentConfig) (*EnvironmentProvider, error) {
	logger := logrus.New()
	
	return &EnvironmentProvider{
		config: config,
		logger: logger,
	}, nil
}

// GetSecret retrieves a secret from environment variables
func (e *EnvironmentProvider) GetSecret(ctx context.Context, key string) (string, error) {
	envKey := e.buildEnvKey(key)
	e.logger.Debugf("Getting environment variable: %s", envKey)
	
	value := os.Getenv(envKey)
	if value == "" {
		return "", fmt.Errorf("environment variable not found: %s", envKey)
	}
	
	return value, nil
}

// GetSecrets retrieves multiple secrets from environment variables
func (e *EnvironmentProvider) GetSecrets(ctx context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string)
	
	for _, key := range keys {
		value, err := e.GetSecret(ctx, key)
		if err != nil {
			e.logger.Warnf("Failed to get environment variable %s: %v", key, err)
			continue
		}
		result[key] = value
	}
	
	return result, nil
}

// SetSecret stores a secret as an environment variable (for current process only)
func (e *EnvironmentProvider) SetSecret(ctx context.Context, key, value string, metadata map[string]string) error {
	envKey := e.buildEnvKey(key)
	e.logger.Debugf("Setting environment variable: %s", envKey)
	
	err := os.Setenv(envKey, value)
	if err != nil {
		return fmt.Errorf("failed to set environment variable %s: %w", envKey, err)
	}
	
	return nil
}

// DeleteSecret removes a secret from environment variables (for current process only)
func (e *EnvironmentProvider) DeleteSecret(ctx context.Context, key string) error {
	envKey := e.buildEnvKey(key)
	e.logger.Debugf("Unsetting environment variable: %s", envKey)
	
	err := os.Unsetenv(envKey)
	if err != nil {
		return fmt.Errorf("failed to unset environment variable %s: %w", envKey, err)
	}
	
	return nil
}

// ListSecrets lists all environment variables with optional prefix
func (e *EnvironmentProvider) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	e.logger.Debugf("Listing environment variables with prefix: %s", prefix)
	
	var keys []string
	envPrefix := e.buildEnvKey(prefix)
	
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		envKey := parts[0]
		
		// Check if it matches our prefix pattern
		if e.config.Prefix != "" {
			if !strings.HasPrefix(envKey, e.config.Prefix) {
				continue
			}
			// Remove the prefix to get the original key
			key := envKey[len(e.config.Prefix):]
			if prefix == "" || strings.HasPrefix(key, prefix) {
				keys = append(keys, key)
			}
		} else {
			// No prefix configured, check mapping or direct match
			if e.config.Mapping != nil {
				for originalKey, mappedKey := range e.config.Mapping {
					if mappedKey == envKey && (prefix == "" || strings.HasPrefix(originalKey, prefix)) {
						keys = append(keys, originalKey)
					}
				}
			} else {
				// Direct match
				if prefix == "" || strings.HasPrefix(envKey, envPrefix) {
					keys = append(keys, envKey)
				}
			}
		}
	}
	
	return keys, nil
}

// RotateSecret rotates a secret in environment variables (generates new value)
func (e *EnvironmentProvider) RotateSecret(ctx context.Context, key string) error {
	// Generate a new secret value
	newValue, err := GenerateSecretKey(32)
	if err != nil {
		return fmt.Errorf("failed to generate new secret value: %w", err)
	}
	
	// Set the new value
	return e.SetSecret(ctx, key, newValue, nil)
}

// HealthCheck verifies environment provider is working
func (e *EnvironmentProvider) HealthCheck(ctx context.Context) error {
	// Environment provider is always available
	return nil
}

// Close closes the environment provider
func (e *EnvironmentProvider) Close() error {
	// Nothing to close for environment provider
	return nil
}

// buildEnvKey builds the environment variable key based on configuration
func (e *EnvironmentProvider) buildEnvKey(key string) string {
	// Check if there's a specific mapping for this key
	if e.config.Mapping != nil {
		if mappedKey, exists := e.config.Mapping[key]; exists {
			return mappedKey
		}
	}
	
	// Apply prefix if configured
	if e.config.Prefix != "" {
		return e.config.Prefix + key
	}
	
	// Convert to uppercase if not case sensitive (default behavior)
	if !e.config.CaseSensitive {
		return strings.ToUpper(key)
	}
	
	return key
}
