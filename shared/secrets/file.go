package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// FileProvider implements SecretProvider for file-based secrets
type FileProvider struct {
	config FileConfig
	logger *logrus.Logger
	data   map[string]string
}

// NewFileProvider creates a new file-based provider
func NewFileProvider(config FileConfig) (*FileProvider, error) {
	logger := logrus.New()

	provider := &FileProvider{
		config: config,
		logger: logger,
		data:   make(map[string]string),
	}

	// Load secrets from file
	if err := provider.loadSecrets(); err != nil {
		return nil, fmt.Errorf("failed to load secrets from file: %w", err)
	}

	return provider, nil
}

// loadSecrets loads secrets from the configured file
func (f *FileProvider) loadSecrets() error {
	if f.config.Path == "" {
		return fmt.Errorf("file path not configured")
	}

	f.logger.Debugf("Loading secrets from file: %s", f.config.Path)

	// Check if file exists
	if _, err := os.Stat(f.config.Path); os.IsNotExist(err) {
		f.logger.Warnf("Secrets file does not exist: %s", f.config.Path)
		return nil
	}

	// Read file content
	content, err := ioutil.ReadFile(f.config.Path)
	if err != nil {
		return fmt.Errorf("failed to read secrets file: %w", err)
	}

	// Decrypt if encrypted
	if f.config.Encrypted {
		content, err = f.decrypt(content)
		if err != nil {
			return fmt.Errorf("failed to decrypt secrets file: %w", err)
		}
	}

	// Parse based on format
	switch strings.ToLower(f.config.Format) {
	case "json":
		return f.parseJSON(content)
	case "yaml", "yml":
		return f.parseYAML(content)
	case "properties", "env":
		return f.parseProperties(content)
	default:
		return fmt.Errorf("unsupported file format: %s", f.config.Format)
	}
}

// parseJSON parses JSON format secrets
func (f *FileProvider) parseJSON(content []byte) error {
	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	f.data = make(map[string]string)
	for key, value := range data {
		if strValue, ok := value.(string); ok {
			f.data[key] = strValue
		} else {
			// Convert non-string values to JSON
			jsonValue, _ := json.Marshal(value)
			f.data[key] = string(jsonValue)
		}
	}

	return nil
}

// parseYAML parses YAML format secrets using Viper
func (f *FileProvider) parseYAML(content []byte) error {
	v := viper.New()
	v.SetConfigType("yaml")

	if err := v.ReadConfig(strings.NewReader(string(content))); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	f.data = make(map[string]string)
	for key, value := range v.AllSettings() {
		if strValue, ok := value.(string); ok {
			f.data[key] = strValue
		} else {
			// Convert non-string values to JSON
			jsonValue, _ := json.Marshal(value)
			f.data[key] = string(jsonValue)
		}
	}

	return nil
}

// parseProperties parses properties/env format secrets
func (f *FileProvider) parseProperties(content []byte) error {
	lines := strings.Split(string(content), "\n")
	f.data = make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		// Split key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		f.data[key] = value
	}

	return nil
}

// decrypt decrypts the file content (placeholder implementation)
func (f *FileProvider) decrypt(content []byte) ([]byte, error) {
	// This is a placeholder implementation
	// In a real implementation, you would use proper encryption/decryption
	// based on the key provided in f.config.KeyPath
	if f.config.KeyPath == "" {
		return nil, fmt.Errorf("encryption key path not configured")
	}

	// For now, just return the content as-is
	// TODO: Implement actual decryption
	return content, nil
}

// GetSecret retrieves a secret from file data
func (f *FileProvider) GetSecret(ctx context.Context, key string) (string, error) {
	f.logger.Debugf("Getting secret from file: %s", key)

	if value, exists := f.data[key]; exists {
		return value, nil
	}

	return "", fmt.Errorf("secret not found in file: %s", key)
}

// GetSecrets retrieves multiple secrets from file data
func (f *FileProvider) GetSecrets(ctx context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string)

	for _, key := range keys {
		if value, exists := f.data[key]; exists {
			result[key] = value
		}
	}

	return result, nil
}

// SetSecret stores a secret in file data and saves to file
func (f *FileProvider) SetSecret(ctx context.Context, key, value string, metadata map[string]string) error {
	f.logger.Debugf("Setting secret in file: %s", key)

	f.data[key] = value

	// Save to file
	return f.saveSecrets()
}

// DeleteSecret removes a secret from file data and saves to file
func (f *FileProvider) DeleteSecret(ctx context.Context, key string) error {
	f.logger.Debugf("Deleting secret from file: %s", key)

	delete(f.data, key)

	// Save to file
	return f.saveSecrets()
}

// ListSecrets lists all secrets in file with optional prefix
func (f *FileProvider) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	f.logger.Debugf("Listing secrets from file with prefix: %s", prefix)

	var keys []string
	for key := range f.data {
		if prefix == "" || strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// RotateSecret rotates a secret in file data
func (f *FileProvider) RotateSecret(ctx context.Context, key string) error {
	// Generate a new secret value
	newValue, err := GenerateSecretKey(32)
	if err != nil {
		return fmt.Errorf("failed to generate new secret value: %w", err)
	}

	// Set the new value
	metadata := map[string]string{
		"rotated_at": time.Now().UTC().Format(time.RFC3339),
	}

	return f.SetSecret(ctx, key, newValue, metadata)
}

// HealthCheck verifies file provider is working
func (f *FileProvider) HealthCheck(ctx context.Context) error {
	// Check if file is readable
	if f.config.Path == "" {
		return fmt.Errorf("file path not configured")
	}

	if _, err := os.Stat(f.config.Path); err != nil {
		return fmt.Errorf("secrets file not accessible: %w", err)
	}

	return nil
}

// Close closes the file provider
func (f *FileProvider) Close() error {
	// Nothing to close for file provider
	return nil
}

// saveSecrets saves the current data to file
func (f *FileProvider) saveSecrets() error {
	if f.config.Path == "" {
		return fmt.Errorf("file path not configured")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(f.config.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var content []byte
	var err error

	// Generate content based on format
	switch strings.ToLower(f.config.Format) {
	case "json":
		content, err = f.generateJSON()
	case "yaml", "yml":
		content, err = f.generateYAML()
	case "properties", "env":
		content, err = f.generateProperties()
	default:
		return fmt.Errorf("unsupported file format: %s", f.config.Format)
	}

	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	// Encrypt if required
	if f.config.Encrypted {
		content, err = f.encrypt(content)
		if err != nil {
			return fmt.Errorf("failed to encrypt content: %w", err)
		}
	}

	// Write to file
	if err := ioutil.WriteFile(f.config.Path, content, 0600); err != nil {
		return fmt.Errorf("failed to write secrets file: %w", err)
	}

	f.logger.Debugf("Saved secrets to file: %s", f.config.Path)
	return nil
}

// generateJSON generates JSON format content
func (f *FileProvider) generateJSON() ([]byte, error) {
	return json.MarshalIndent(f.data, "", "  ")
}

// generateYAML generates YAML format content
func (f *FileProvider) generateYAML() ([]byte, error) {
	// This is a simple YAML generation
	// For complex YAML, you'd want to use a proper YAML library
	var lines []string
	for key, value := range f.data {
		lines = append(lines, fmt.Sprintf("%s: %s", key, value))
	}
	return []byte(strings.Join(lines, "\n")), nil
}

// generateProperties generates properties format content
func (f *FileProvider) generateProperties() ([]byte, error) {
	var lines []string
	lines = append(lines, "# Secrets file generated at "+time.Now().Format(time.RFC3339))

	for key, value := range f.data {
		// Escape special characters in value
		value = strings.ReplaceAll(value, "\\", "\\\\")
		value = strings.ReplaceAll(value, "=", "\\=")
		value = strings.ReplaceAll(value, "\n", "\\n")

		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}

	return []byte(strings.Join(lines, "\n")), nil
}

// encrypt encrypts the content (placeholder implementation)
func (f *FileProvider) encrypt(content []byte) ([]byte, error) {
	// This is a placeholder implementation
	// In a real implementation, you would use proper encryption
	// based on the key provided in f.config.KeyPath
	if f.config.KeyPath == "" {
		return nil, fmt.Errorf("encryption key path not configured")
	}

	// For now, just return the content as-is
	// TODO: Implement actual encryption
	return content, nil
}
