package secrets

import (
	"context"
	"fmt"
)

// createProvider creates a secret provider based on the provider type
func (sm *SecretManager) createProvider(providerType ProviderType) (SecretProvider, error) {
	switch providerType {
	case ProviderTypeVault:
		return NewVaultProvider(sm.config.Vault)
	case ProviderTypeAWSSecrets:
		return NewAWSSecretsProvider(sm.config.AWS)
	case ProviderTypeAWSSSM:
		return NewAWSSSMProvider(sm.config.AWS)
	case ProviderTypeKubernetes:
		return NewKubernetesProvider(sm.config.Kubernetes)
	case ProviderTypeEnvironment:
		return NewEnvironmentProvider(sm.config.Environment)
	case ProviderTypeFile:
		return NewFileProvider(sm.config.File)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}

// ProviderFactory creates providers for testing and custom implementations
type ProviderFactory interface {
	CreateProvider(providerType ProviderType, config interface{}) (SecretProvider, error)
}

// DefaultProviderFactory is the default implementation of ProviderFactory
type DefaultProviderFactory struct{}

// CreateProvider creates a provider using the default factory
func (f *DefaultProviderFactory) CreateProvider(providerType ProviderType, config interface{}) (SecretProvider, error) {
	switch providerType {
	case ProviderTypeVault:
		if vaultConfig, ok := config.(VaultConfig); ok {
			return NewVaultProvider(vaultConfig)
		}
		return nil, fmt.Errorf("invalid vault config type")
	case ProviderTypeAWSSecrets:
		if awsConfig, ok := config.(AWSConfig); ok {
			return NewAWSSecretsProvider(awsConfig)
		}
		return nil, fmt.Errorf("invalid aws config type")
	case ProviderTypeAWSSSM:
		if awsConfig, ok := config.(AWSConfig); ok {
			return NewAWSSSMProvider(awsConfig)
		}
		return nil, fmt.Errorf("invalid aws config type")
	case ProviderTypeKubernetes:
		if k8sConfig, ok := config.(KubernetesConfig); ok {
			return NewKubernetesProvider(k8sConfig)
		}
		return nil, fmt.Errorf("invalid kubernetes config type")
	case ProviderTypeEnvironment:
		if envConfig, ok := config.(EnvironmentConfig); ok {
			return NewEnvironmentProvider(envConfig)
		}
		return nil, fmt.Errorf("invalid environment config type")
	case ProviderTypeFile:
		if fileConfig, ok := config.(FileConfig); ok {
			return NewFileProvider(fileConfig)
		}
		return nil, fmt.Errorf("invalid file config type")
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}

// MockProvider is a simple mock implementation for testing
type MockProvider struct {
	secrets map[string]string
	healthy bool
}

// NewMockProvider creates a new mock provider
func NewMockProvider(secrets map[string]string) *MockProvider {
	return &MockProvider{
		secrets: secrets,
		healthy: true,
	}
}

// GetSecret implements SecretProvider
func (m *MockProvider) GetSecret(ctx context.Context, key string) (string, error) {
	if !m.healthy {
		return "", fmt.Errorf("mock provider is unhealthy")
	}

	if value, exists := m.secrets[key]; exists {
		return value, nil
	}
	return "", fmt.Errorf("secret not found: %s", key)
}

// GetSecrets implements SecretProvider
func (m *MockProvider) GetSecrets(ctx context.Context, keys []string) (map[string]string, error) {
	if !m.healthy {
		return nil, fmt.Errorf("mock provider is unhealthy")
	}

	result := make(map[string]string)
	for _, key := range keys {
		if value, exists := m.secrets[key]; exists {
			result[key] = value
		}
	}
	return result, nil
}

// SetSecret implements SecretProvider
func (m *MockProvider) SetSecret(ctx context.Context, key, value string, metadata map[string]string) error {
	if !m.healthy {
		return fmt.Errorf("mock provider is unhealthy")
	}

	if m.secrets == nil {
		m.secrets = make(map[string]string)
	}
	m.secrets[key] = value
	return nil
}

// DeleteSecret implements SecretProvider
func (m *MockProvider) DeleteSecret(ctx context.Context, key string) error {
	if !m.healthy {
		return fmt.Errorf("mock provider is unhealthy")
	}

	delete(m.secrets, key)
	return nil
}

// ListSecrets implements SecretProvider
func (m *MockProvider) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	if !m.healthy {
		return nil, fmt.Errorf("mock provider is unhealthy")
	}

	var keys []string
	for key := range m.secrets {
		if prefix == "" || len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

// RotateSecret implements SecretProvider
func (m *MockProvider) RotateSecret(ctx context.Context, key string) error {
	if !m.healthy {
		return fmt.Errorf("mock provider is unhealthy")
	}

	// Generate a new secret value
	newValue, err := GenerateSecretKey(32)
	if err != nil {
		return err
	}

	m.secrets[key] = newValue
	return nil
}

// HealthCheck implements SecretProvider
func (m *MockProvider) HealthCheck(ctx context.Context) error {
	if !m.healthy {
		return fmt.Errorf("mock provider is unhealthy")
	}
	return nil
}

// Close implements SecretProvider
func (m *MockProvider) Close() error {
	return nil
}

// SetHealthy sets the health status of the mock provider
func (m *MockProvider) SetHealthy(healthy bool) {
	m.healthy = healthy
}

// AddSecret adds a secret to the mock provider
func (m *MockProvider) AddSecret(key, value string) {
	if m.secrets == nil {
		m.secrets = make(map[string]string)
	}
	m.secrets[key] = value
}
