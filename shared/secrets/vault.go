package secrets

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
)

// VaultProvider implements SecretProvider for HashiCorp Vault
type VaultProvider struct {
	client    *api.Client
	config    VaultConfig
	logger    *logrus.Logger
	mountPath string
}

// NewVaultProvider creates a new Vault provider
func NewVaultProvider(config VaultConfig) (SecretProvider, error) {
	vaultConfig := api.DefaultConfig()
	vaultConfig.Address = config.Address

	if config.TLS.CACert != "" {
		tlsConfig := &api.TLSConfig{
			CACert: config.TLS.CACert,
		}
		vaultConfig.ConfigureTLS(tlsConfig)
	}

	client, err := api.NewClient(vaultConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	provider := &VaultProvider{
		client: client,
		config: config,
	}

	// Authenticate with Vault
	if err := provider.authenticate(); err != nil {
		return nil, fmt.Errorf("failed to authenticate with vault: %w", err)
	}

	return provider, nil
}

// authenticate handles Vault authentication
func (v *VaultProvider) authenticate() error {
	// If token is provided directly, use it
	if v.config.Token != "" {
		v.client.SetToken(v.config.Token)
		return nil
	}

	// If token path is provided, read token from file
	if v.config.TokenPath != "" {
		tokenBytes, err := ioutil.ReadFile(v.config.TokenPath)
		if err != nil {
			return fmt.Errorf("failed to read token from file %s: %w", v.config.TokenPath, err)
		}
		token := strings.TrimSpace(string(tokenBytes))
		v.client.SetToken(token)
		return nil
	}

	// Handle different authentication methods
	switch v.config.Auth.Method {
	case "kubernetes":
		return v.authenticateKubernetes()
	case "aws":
		return v.authenticateAWS()
	case "userpass":
		return v.authenticateUserpass()
	case "ldap":
		return v.authenticateLDAP()
	case "github":
		return v.authenticateGitHub()
	case "approle":
		return v.authenticateAppRole()
	case "":
		// Try to get token from environment
		if token := os.Getenv("VAULT_TOKEN"); token != "" {
			v.client.SetToken(token)
			return nil
		}
		return fmt.Errorf("no authentication method specified and VAULT_TOKEN not set")
	default:
		return fmt.Errorf("unsupported authentication method: %s", v.config.Auth.Method)
	}
}

// authenticateKubernetes handles Kubernetes service account authentication
func (v *VaultProvider) authenticateKubernetes() error {
	role := v.config.Auth.Parameters["role"]
	if role == "" {
		return fmt.Errorf("kubernetes auth requires 'role' parameter")
	}

	// Read service account token
	tokenPath := v.config.Auth.Parameters["token_path"]
	if tokenPath == "" {
		tokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	}

	tokenBytes, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		return fmt.Errorf("failed to read service account token: %w", err)
	}

	// Authenticate with Vault
	options := map[string]interface{}{
		"role": role,
		"jwt":  string(tokenBytes),
	}

	mountPath := v.config.Auth.Parameters["mount_path"]
	if mountPath == "" {
		mountPath = "kubernetes"
	}

	secret, err := v.client.Logical().Write(fmt.Sprintf("auth/%s/login", mountPath), options)
	if err != nil {
		return fmt.Errorf("kubernetes authentication failed: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("no auth information returned from kubernetes authentication")
	}

	v.client.SetToken(secret.Auth.ClientToken)
	return nil
}

// authenticateAWS handles AWS IAM authentication
func (v *VaultProvider) authenticateAWS() error {
	role := v.config.Auth.Parameters["role"]
	if role == "" {
		return fmt.Errorf("aws auth requires 'role' parameter")
	}

	// This is a simplified implementation
	// In a real implementation, you would generate the required AWS signature
	options := map[string]interface{}{
		"role": role,
	}

	mountPath := v.config.Auth.Parameters["mount_path"]
	if mountPath == "" {
		mountPath = "aws"
	}

	secret, err := v.client.Logical().Write(fmt.Sprintf("auth/%s/login", mountPath), options)
	if err != nil {
		return fmt.Errorf("aws authentication failed: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("no auth information returned from aws authentication")
	}

	v.client.SetToken(secret.Auth.ClientToken)
	return nil
}

// authenticateUserpass handles username/password authentication
func (v *VaultProvider) authenticateUserpass() error {
	username := v.config.Auth.Parameters["username"]
	password := v.config.Auth.Parameters["password"]

	if username == "" || password == "" {
		return fmt.Errorf("userpass auth requires 'username' and 'password' parameters")
	}

	options := map[string]interface{}{
		"password": password,
	}

	mountPath := v.config.Auth.Parameters["mount_path"]
	if mountPath == "" {
		mountPath = "userpass"
	}

	secret, err := v.client.Logical().Write(fmt.Sprintf("auth/%s/login/%s", mountPath, username), options)
	if err != nil {
		return fmt.Errorf("userpass authentication failed: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("no auth information returned from userpass authentication")
	}

	v.client.SetToken(secret.Auth.ClientToken)
	return nil
}

// authenticateLDAP handles LDAP authentication
func (v *VaultProvider) authenticateLDAP() error {
	username := v.config.Auth.Parameters["username"]
	password := v.config.Auth.Parameters["password"]

	if username == "" || password == "" {
		return fmt.Errorf("ldap auth requires 'username' and 'password' parameters")
	}

	options := map[string]interface{}{
		"password": password,
	}

	mountPath := v.config.Auth.Parameters["mount_path"]
	if mountPath == "" {
		mountPath = "ldap"
	}

	secret, err := v.client.Logical().Write(fmt.Sprintf("auth/%s/login/%s", mountPath, username), options)
	if err != nil {
		return fmt.Errorf("ldap authentication failed: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("no auth information returned from ldap authentication")
	}

	v.client.SetToken(secret.Auth.ClientToken)
	return nil
}

// authenticateGitHub handles GitHub authentication
func (v *VaultProvider) authenticateGitHub() error {
	token := v.config.Auth.Parameters["token"]
	if token == "" {
		return fmt.Errorf("github auth requires 'token' parameter")
	}

	options := map[string]interface{}{
		"token": token,
	}

	mountPath := v.config.Auth.Parameters["mount_path"]
	if mountPath == "" {
		mountPath = "github"
	}

	secret, err := v.client.Logical().Write(fmt.Sprintf("auth/%s/login", mountPath), options)
	if err != nil {
		return fmt.Errorf("github authentication failed: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("no auth information returned from github authentication")
	}

	v.client.SetToken(secret.Auth.ClientToken)
	return nil
}

// authenticateAppRole handles AppRole authentication
func (v *VaultProvider) authenticateAppRole() error {
	roleID := v.config.Auth.Parameters["role_id"]
	secretID := v.config.Auth.Parameters["secret_id"]

	if roleID == "" || secretID == "" {
		return fmt.Errorf("approle auth requires 'role_id' and 'secret_id' parameters")
	}

	options := map[string]interface{}{
		"role_id":   roleID,
		"secret_id": secretID,
	}

	mountPath := v.config.Auth.Parameters["mount_path"]
	if mountPath == "" {
		mountPath = "approle"
	}

	secret, err := v.client.Logical().Write(fmt.Sprintf("auth/%s/login", mountPath), options)
	if err != nil {
		return fmt.Errorf("approle authentication failed: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("no auth information returned from approle authentication")
	}

	v.client.SetToken(secret.Auth.ClientToken)
	return nil
}

// GetSecret retrieves a secret from Vault
func (v *VaultProvider) GetSecret(ctx context.Context, key string) (string, error) {
	// Handle both KV v1 and v2 paths
	path := v.buildSecretPath(key)

	v.logger.Debugf("Reading secret from Vault path: %s", path)

	secret, err := v.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return "", fmt.Errorf("failed to read secret from Vault: %w", err)
	}

	if secret == nil {
		return "", fmt.Errorf("secret not found: %s", key)
	}

	// Handle KV v2 format
	if secret.Data != nil {
		if data, ok := secret.Data["data"].(map[string]interface{}); ok {
			if value, exists := data["value"]; exists {
				if strValue, ok := value.(string); ok {
					return strValue, nil
				}
			}
			// Try the key itself
			if value, exists := data[key]; exists {
				if strValue, ok := value.(string); ok {
					return strValue, nil
				}
			}
		}

		// Handle KV v1 format
		if value, exists := secret.Data["value"]; exists {
			if strValue, ok := value.(string); ok {
				return strValue, nil
			}
		}

		// Try the key itself in v1 format
		if value, exists := secret.Data[key]; exists {
			if strValue, ok := value.(string); ok {
				return strValue, nil
			}
		}
	}

	return "", fmt.Errorf("secret value not found or invalid format for key: %s", key)
}

// GetSecrets retrieves multiple secrets from Vault
func (v *VaultProvider) GetSecrets(ctx context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string)

	for _, key := range keys {
		value, err := v.GetSecret(ctx, key)
		if err != nil {
			v.logger.Warnf("Failed to get secret %s: %v", key, err)
			continue
		}
		result[key] = value
	}

	return result, nil
}

// SetSecret stores a secret in Vault
func (v *VaultProvider) SetSecret(ctx context.Context, key, value string, metadata map[string]string) error {
	path := v.buildSecretPath(key)

	data := map[string]interface{}{
		"value": value,
	}

	// Add metadata
	for k, v := range metadata {
		data[k] = v
	}

	// Handle KV v2 format
	if v.isKVv2() {
		payload := map[string]interface{}{
			"data": data,
		}
		if metadata != nil {
			payload["metadata"] = metadata
		}
		data = payload
	}

	v.logger.Debugf("Writing secret to Vault path: %s", path)

	_, err := v.client.Logical().WriteWithContext(ctx, path, data)
	if err != nil {
		return fmt.Errorf("failed to write secret to Vault: %w", err)
	}

	return nil
}

// DeleteSecret removes a secret from Vault
func (v *VaultProvider) DeleteSecret(ctx context.Context, key string) error {
	path := v.buildSecretPath(key)

	v.logger.Debugf("Deleting secret from Vault path: %s", path)

	_, err := v.client.Logical().DeleteWithContext(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete secret from Vault: %w", err)
	}

	return nil
}

// ListSecrets lists secrets in Vault with optional prefix
func (v *VaultProvider) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	path := v.mountPath
	if v.isKVv2() {
		path = path + "/metadata"
	}
	if prefix != "" {
		path = path + "/" + prefix
	}

	v.logger.Debugf("Listing secrets from Vault path: %s", path)

	secret, err := v.client.Logical().ListWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets from Vault: %w", err)
	}

	if secret == nil || secret.Data == nil {
		return []string{}, nil
	}

	if keys, ok := secret.Data["keys"].([]interface{}); ok {
		result := make([]string, 0, len(keys))
		for _, key := range keys {
			if strKey, ok := key.(string); ok {
				result = append(result, strKey)
			}
		}
		return result, nil
	}

	return []string{}, nil
}

// RotateSecret rotates a secret in Vault (generates new value)
func (v *VaultProvider) RotateSecret(ctx context.Context, key string) error {
	// Generate a new secret value
	newValue, err := GenerateSecretKey(32)
	if err != nil {
		return fmt.Errorf("failed to generate new secret value: %w", err)
	}

	// Set the new value
	metadata := map[string]string{
		"rotated_at": time.Now().UTC().Format(time.RFC3339),
		"rotated_by": "secret-manager",
	}

	return v.SetSecret(ctx, key, newValue, metadata)
}

// HealthCheck verifies Vault connectivity
func (v *VaultProvider) HealthCheck(ctx context.Context) error {
	health, err := v.client.Sys().HealthWithContext(ctx)
	if err != nil {
		return fmt.Errorf("vault health check failed: %w", err)
	}

	if health == nil {
		return fmt.Errorf("vault health check returned nil response")
	}

	if !health.Initialized {
		return fmt.Errorf("vault is not initialized")
	}

	if health.Sealed {
		return fmt.Errorf("vault is sealed")
	}

	return nil
}

// Close closes the Vault provider connection
func (v *VaultProvider) Close() error {
	// Vault client doesn't require explicit closing
	return nil
}

// buildSecretPath builds the full path for a secret
func (v *VaultProvider) buildSecretPath(key string) string {
	if v.isKVv2() {
		return fmt.Sprintf("%s/data/%s", v.mountPath, key)
	}
	return fmt.Sprintf("%s/%s", v.mountPath, key)
}

// isKVv2 checks if the mount is KV version 2
func (v *VaultProvider) isKVv2() bool {
	// This is a simplified check. In a real implementation,
	// you would query the mount information from Vault
	return strings.Contains(v.mountPath, "kv") || v.mountPath == "secret"
}
