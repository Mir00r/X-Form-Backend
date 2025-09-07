package secrets

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// ConfigLoader handles loading configuration from various sources
type ConfigLoader struct {
	configPaths []string
	envPrefix   string
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader(configPaths []string, envPrefix string) *ConfigLoader {
	return &ConfigLoader{
		configPaths: configPaths,
		envPrefix:   envPrefix,
	}
}

// LoadConfig loads configuration from files and environment variables
func (cl *ConfigLoader) LoadConfig() (*Config, error) {
	v := viper.New()

	// Set environment variable prefix
	if cl.envPrefix != "" {
		v.SetEnvPrefix(cl.envPrefix)
	}
	v.AutomaticEnv()

	// Set default values
	cl.setDefaults(v)

	// Try to read configuration files
	for _, configPath := range cl.configPaths {
		if _, err := os.Stat(configPath); err == nil {
			v.SetConfigFile(configPath)
			if err := v.ReadInConfig(); err != nil {
				return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
			}
			break
		}
	}

	// Unmarshal configuration
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cl.validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func (cl *ConfigLoader) setDefaults(v *viper.Viper) {
	// Default provider
	v.SetDefault("provider", "environment")

	// Cache defaults
	v.SetDefault("cache.enabled", true)
	v.SetDefault("cache.ttl", "5m")
	v.SetDefault("cache.max_entries", 1000)

	// Vault defaults
	v.SetDefault("vault.address", "http://localhost:8200")
	v.SetDefault("vault.mount_path", "secret")
	v.SetDefault("vault.tls.enabled", false)
	v.SetDefault("vault.tls.insecure_skip_verify", false)

	// AWS defaults
	v.SetDefault("aws.region", "us-east-1")
	v.SetDefault("aws.ssm_path", "/app/secrets")

	// Kubernetes defaults
	v.SetDefault("kubernetes.namespace", "default")
	v.SetDefault("kubernetes.secret_name", "app-secrets")
	v.SetDefault("kubernetes.in_cluster", false)

	// Environment defaults
	v.SetDefault("environment.prefix", "")
	v.SetDefault("environment.case_sensitive", false)

	// File defaults
	v.SetDefault("file.format", "json")
	v.SetDefault("file.encrypted", false)

	// Security defaults
	v.SetDefault("security.encryption_enabled", false)
	v.SetDefault("security.audit_enabled", false)
}

// validateConfig validates the configuration
func (cl *ConfigLoader) validateConfig(config *Config) error {
	// Validate provider type
	validProviders := map[ProviderType]bool{
		ProviderTypeVault:       true,
		ProviderTypeAWSSecrets:  true,
		ProviderTypeAWSSSM:      true,
		ProviderTypeKubernetes:  true,
		ProviderTypeEnvironment: true,
		ProviderTypeFile:        true,
	}

	if !validProviders[config.Provider] {
		return fmt.Errorf("invalid provider type: %s", config.Provider)
	}

	// Validate fallback providers
	for _, fallback := range config.Fallbacks {
		if !validProviders[fallback] {
			return fmt.Errorf("invalid fallback provider type: %s", fallback)
		}
	}

	// Validate provider-specific configuration
	switch config.Provider {
	case ProviderTypeVault:
		if err := cl.validateVaultConfig(config.Vault); err != nil {
			return fmt.Errorf("invalid vault config: %w", err)
		}
	case ProviderTypeAWSSecrets, ProviderTypeAWSSSM:
		if err := cl.validateAWSConfig(config.AWS); err != nil {
			return fmt.Errorf("invalid aws config: %w", err)
		}
	case ProviderTypeFile:
		if err := cl.validateFileConfig(config.File); err != nil {
			return fmt.Errorf("invalid file config: %w", err)
		}
	}

	return nil
}

// validateVaultConfig validates Vault configuration
func (cl *ConfigLoader) validateVaultConfig(config VaultConfig) error {
	if config.Address == "" {
		return fmt.Errorf("vault address is required")
	}

	// Validate authentication method if specified
	if config.Auth.Method != "" {
		validMethods := map[string]bool{
			"kubernetes": true,
			"aws":        true,
			"userpass":   true,
			"ldap":       true,
			"github":     true,
			"approle":    true,
		}

		if !validMethods[config.Auth.Method] {
			return fmt.Errorf("invalid vault auth method: %s", config.Auth.Method)
		}
	}

	return nil
}

// validateAWSConfig validates AWS configuration
func (cl *ConfigLoader) validateAWSConfig(config AWSConfig) error {
	if config.Region == "" {
		return fmt.Errorf("aws region is required")
	}
	return nil
}

// validateFileConfig validates file configuration
func (cl *ConfigLoader) validateFileConfig(config FileConfig) error {
	if config.Path == "" {
		return fmt.Errorf("file path is required")
	}

	validFormats := map[string]bool{
		"json":       true,
		"yaml":       true,
		"yml":        true,
		"properties": true,
		"env":        true,
	}

	if !validFormats[config.Format] {
		return fmt.Errorf("invalid file format: %s", config.Format)
	}

	return nil
}

// LoadConfigFromFile loads configuration from a specific file
func LoadConfigFromFile(filePath string) (*Config, error) {
	loader := NewConfigLoader([]string{filePath}, "SECRETS")
	return loader.LoadConfig()
}

// LoadConfigFromEnv loads configuration from environment variables only
func LoadConfigFromEnv(envPrefix string) (*Config, error) {
	loader := NewConfigLoader([]string{}, envPrefix)
	return loader.LoadConfig()
}

// SaveConfigToFile saves configuration to a file
func SaveConfigToFile(config *Config, filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *Config {
	return &Config{
		Provider:  ProviderTypeEnvironment,
		Fallbacks: []ProviderType{ProviderTypeFile},
		Cache: CacheConfig{
			Enabled:    true,
			TTL:        5 * time.Minute,
			MaxEntries: 1000,
		},
		Vault: VaultConfig{
			Address:   "http://localhost:8200",
			MountPath: "secret",
			TLS: VaultTLSConfig{
				Enabled:            false,
				InsecureSkipVerify: false,
			},
		},
		AWS: AWSConfig{
			Region:  "us-east-1",
			SSMPath: "/app/secrets",
		},
		Kubernetes: KubernetesConfig{
			Namespace:  "default",
			SecretName: "app-secrets",
			InCluster:  false,
		},
		Environment: EnvironmentConfig{
			Prefix:        "",
			CaseSensitive: false,
		},
		File: FileConfig{
			Path:      "./secrets.json",
			Format:    "json",
			Encrypted: false,
		},
		Security: SecurityConfig{
			EncryptionEnabled: false,
			AuditEnabled:      false,
		},
	}
}

// GetDevConfig returns a configuration suitable for development
func GetDevConfig() *Config {
	config := GetDefaultConfig()
	config.Provider = ProviderTypeEnvironment
	config.Fallbacks = []ProviderType{ProviderTypeFile}
	config.Cache.TTL = 1 * time.Minute
	config.Security.AuditEnabled = true
	return config
}

// GetProdConfig returns a configuration suitable for production
func GetProdConfig() *Config {
	config := GetDefaultConfig()
	config.Provider = ProviderTypeVault
	config.Fallbacks = []ProviderType{ProviderTypeAWSSecrets, ProviderTypeKubernetes}
	config.Cache.TTL = 15 * time.Minute
	config.Security.EncryptionEnabled = true
	config.Security.AuditEnabled = true
	config.Vault.TLS.Enabled = true
	return config
}

// GetKubernetesConfig returns a configuration suitable for Kubernetes deployment
func GetKubernetesConfig() *Config {
	config := GetDefaultConfig()
	config.Provider = ProviderTypeKubernetes
	config.Fallbacks = []ProviderType{ProviderTypeVault, ProviderTypeEnvironment}
	config.Kubernetes.InCluster = true
	config.Vault.Auth.Method = "kubernetes"
	config.Security.AuditEnabled = true
	return config
}
