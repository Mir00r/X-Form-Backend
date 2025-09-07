package secrets

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/sirupsen/logrus"
)

// AWSSecretsProvider implements SecretProvider for AWS Secrets Manager
type AWSSecretsProvider struct {
	client *secretsmanager.Client
	config AWSConfig
	logger *logrus.Logger
}

// NewAWSSecretsProvider creates a new AWS Secrets Manager provider
func NewAWSSecretsProvider(config AWSConfig) (*AWSSecretsProvider, error) {
	logger := logrus.New()

	// Load AWS configuration
	cfg, err := loadAWSConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create Secrets Manager client
	client := secretsmanager.NewFromConfig(cfg)

	return &AWSSecretsProvider{
		client: client,
		config: config,
		logger: logger,
	}, nil
}

// GetSecret retrieves a secret from AWS Secrets Manager
func (a *AWSSecretsProvider) GetSecret(ctx context.Context, key string) (string, error) {
	a.logger.Debugf("Getting secret from AWS Secrets Manager: %s", key)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(key),
	}

	result, err := a.client.GetSecretValue(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get secret from AWS Secrets Manager: %w", err)
	}

	if result.SecretString != nil {
		return *result.SecretString, nil
	}

	return "", fmt.Errorf("secret %s has no string value", key)
}

// GetSecrets retrieves multiple secrets from AWS Secrets Manager
func (a *AWSSecretsProvider) GetSecrets(ctx context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string)

	for _, key := range keys {
		value, err := a.GetSecret(ctx, key)
		if err != nil {
			a.logger.Warnf("Failed to get secret %s: %v", key, err)
			continue
		}
		result[key] = value
	}

	return result, nil
}

// SetSecret stores a secret in AWS Secrets Manager
func (a *AWSSecretsProvider) SetSecret(ctx context.Context, key, value string, metadata map[string]string) error {
	a.logger.Debugf("Setting secret in AWS Secrets Manager: %s", key)

	// Try to update existing secret first
	updateInput := &secretsmanager.UpdateSecretInput{
		SecretId:     aws.String(key),
		SecretString: aws.String(value),
	}

	if description, exists := metadata["description"]; exists {
		updateInput.Description = aws.String(description)
	}

	_, err := a.client.UpdateSecret(ctx, updateInput)
	if err != nil {
		// If secret doesn't exist, create it
		createInput := &secretsmanager.CreateSecretInput{
			Name:         aws.String(key),
			SecretString: aws.String(value),
		}

		if description, exists := metadata["description"]; exists {
			createInput.Description = aws.String(description)
		}

		_, createErr := a.client.CreateSecret(ctx, createInput)
		if createErr != nil {
			return fmt.Errorf("failed to create secret in AWS Secrets Manager: %w", createErr)
		}
	}

	return nil
}

// DeleteSecret removes a secret from AWS Secrets Manager
func (a *AWSSecretsProvider) DeleteSecret(ctx context.Context, key string) error {
	a.logger.Debugf("Deleting secret from AWS Secrets Manager: %s", key)

	input := &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(key),
		ForceDeleteWithoutRecovery: aws.Bool(true),
	}

	_, err := a.client.DeleteSecret(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete secret from AWS Secrets Manager: %w", err)
	}

	return nil
}

// ListSecrets lists secrets in AWS Secrets Manager with optional prefix
func (a *AWSSecretsProvider) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	a.logger.Debugf("Listing secrets from AWS Secrets Manager with prefix: %s", prefix)

	input := &secretsmanager.ListSecretsInput{
		MaxResults: aws.Int32(100),
	}

	var secrets []string
	paginator := secretsmanager.NewListSecretsPaginator(a.client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list secrets from AWS Secrets Manager: %w", err)
		}

		for _, secret := range page.SecretList {
			if secret.Name != nil {
				name := *secret.Name
				if prefix == "" || len(name) >= len(prefix) && name[:len(prefix)] == prefix {
					secrets = append(secrets, name)
				}
			}
		}
	}

	return secrets, nil
}

// RotateSecret rotates a secret in AWS Secrets Manager
func (a *AWSSecretsProvider) RotateSecret(ctx context.Context, key string) error {
	a.logger.Debugf("Rotating secret in AWS Secrets Manager: %s", key)

	input := &secretsmanager.RotateSecretInput{
		SecretId: aws.String(key),
	}

	_, err := a.client.RotateSecret(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to rotate secret in AWS Secrets Manager: %w", err)
	}

	return nil
}

// HealthCheck verifies AWS Secrets Manager connectivity
func (a *AWSSecretsProvider) HealthCheck(ctx context.Context) error {
	// Try to list secrets to verify connectivity
	input := &secretsmanager.ListSecretsInput{
		MaxResults: aws.Int32(1),
	}

	_, err := a.client.ListSecrets(ctx, input)
	if err != nil {
		return fmt.Errorf("aws secrets manager health check failed: %w", err)
	}

	return nil
}

// Close closes the AWS Secrets Manager provider
func (a *AWSSecretsProvider) Close() error {
	// AWS SDK clients don't require explicit closing
	return nil
}

// AWSSSMProvider implements SecretProvider for AWS Systems Manager Parameter Store
type AWSSSMProvider struct {
	client   *ssm.Client
	config   AWSConfig
	logger   *logrus.Logger
	basePath string
}

// NewAWSSSMProvider creates a new AWS SSM Parameter Store provider
func NewAWSSSMProvider(config AWSConfig) (*AWSSSMProvider, error) {
	logger := logrus.New()

	// Load AWS configuration
	cfg, err := loadAWSConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create SSM client
	client := ssm.NewFromConfig(cfg)

	// Set base path for parameters
	basePath := config.SSMPath
	if basePath == "" {
		basePath = "/app/secrets"
	}

	return &AWSSSMProvider{
		client:   client,
		config:   config,
		logger:   logger,
		basePath: basePath,
	}, nil
}

// GetSecret retrieves a parameter from AWS SSM Parameter Store
func (a *AWSSSMProvider) GetSecret(ctx context.Context, key string) (string, error) {
	path := a.buildParameterPath(key)
	a.logger.Debugf("Getting parameter from AWS SSM: %s", path)

	input := &ssm.GetParameterInput{
		Name:           aws.String(path),
		WithDecryption: aws.Bool(true),
	}

	result, err := a.client.GetParameter(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get parameter from AWS SSM: %w", err)
	}

	if result.Parameter == nil || result.Parameter.Value == nil {
		return "", fmt.Errorf("parameter %s not found", path)
	}

	return *result.Parameter.Value, nil
}

// GetSecrets retrieves multiple parameters from AWS SSM Parameter Store
func (a *AWSSSMProvider) GetSecrets(ctx context.Context, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return map[string]string{}, nil
	}

	// Build parameter names
	var names []string
	for _, key := range keys {
		names = append(names, a.buildParameterPath(key))
	}

	a.logger.Debugf("Getting parameters from AWS SSM: %v", names)

	input := &ssm.GetParametersInput{
		Names:          names,
		WithDecryption: aws.Bool(true),
	}

	result, err := a.client.GetParameters(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get parameters from AWS SSM: %w", err)
	}

	// Build result map
	resultMap := make(map[string]string)
	for _, param := range result.Parameters {
		if param.Name != nil && param.Value != nil {
			// Extract key from full path
			key := a.extractKeyFromPath(*param.Name)
			resultMap[key] = *param.Value
		}
	}

	return resultMap, nil
}

// SetSecret stores a parameter in AWS SSM Parameter Store
func (a *AWSSSMProvider) SetSecret(ctx context.Context, key, value string, metadata map[string]string) error {
	path := a.buildParameterPath(key)
	a.logger.Debugf("Setting parameter in AWS SSM: %s", path)

	input := &ssm.PutParameterInput{
		Name:      aws.String(path),
		Value:     aws.String(value),
		Type:      "SecureString",
		Overwrite: aws.Bool(true),
	}

	if description, exists := metadata["description"]; exists {
		input.Description = aws.String(description)
	}

	_, err := a.client.PutParameter(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put parameter in AWS SSM: %w", err)
	}

	return nil
}

// DeleteSecret removes a parameter from AWS SSM Parameter Store
func (a *AWSSSMProvider) DeleteSecret(ctx context.Context, key string) error {
	path := a.buildParameterPath(key)
	a.logger.Debugf("Deleting parameter from AWS SSM: %s", path)

	input := &ssm.DeleteParameterInput{
		Name: aws.String(path),
	}

	_, err := a.client.DeleteParameter(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete parameter from AWS SSM: %w", err)
	}

	return nil
}

// ListSecrets lists parameters in AWS SSM Parameter Store with optional prefix
func (a *AWSSSMProvider) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	path := a.basePath
	if prefix != "" {
		path = a.buildParameterPath(prefix)
	}

	a.logger.Debugf("Listing parameters from AWS SSM with path: %s", path)

	input := &ssm.GetParametersByPathInput{
		Path:       aws.String(path),
		Recursive:  aws.Bool(true),
		MaxResults: aws.Int32(10),
	}

	var parameters []string
	paginator := ssm.NewGetParametersByPathPaginator(a.client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list parameters from AWS SSM: %w", err)
		}

		for _, param := range page.Parameters {
			if param.Name != nil {
				key := a.extractKeyFromPath(*param.Name)
				parameters = append(parameters, key)
			}
		}
	}

	return parameters, nil
}

// RotateSecret rotates a parameter in AWS SSM Parameter Store
func (a *AWSSSMProvider) RotateSecret(ctx context.Context, key string) error {
	// Generate a new secret value
	newValue, err := GenerateSecretKey(32)
	if err != nil {
		return fmt.Errorf("failed to generate new secret value: %w", err)
	}

	// Set the new value
	metadata := map[string]string{
		"description": fmt.Sprintf("Rotated at %s", "now"),
	}

	return a.SetSecret(ctx, key, newValue, metadata)
}

// HealthCheck verifies AWS SSM Parameter Store connectivity
func (a *AWSSSMProvider) HealthCheck(ctx context.Context) error {
	// Try to describe parameters to verify connectivity
	input := &ssm.DescribeParametersInput{
		MaxResults: aws.Int32(1),
	}

	_, err := a.client.DescribeParameters(ctx, input)
	if err != nil {
		return fmt.Errorf("aws ssm health check failed: %w", err)
	}

	return nil
}

// Close closes the AWS SSM provider
func (a *AWSSSMProvider) Close() error {
	// AWS SDK clients don't require explicit closing
	return nil
}

// buildParameterPath builds the full parameter path
func (a *AWSSSMProvider) buildParameterPath(key string) string {
	if a.basePath == "" {
		return "/" + key
	}
	return a.basePath + "/" + key
}

// extractKeyFromPath extracts the key from a full parameter path
func (a *AWSSSMProvider) extractKeyFromPath(path string) string {
	if a.basePath == "" {
		if len(path) > 1 && path[0] == '/' {
			return path[1:]
		}
		return path
	}

	prefix := a.basePath + "/"
	if len(path) > len(prefix) && path[:len(prefix)] == prefix {
		return path[len(prefix):]
	}
	return path
}

// loadAWSConfig loads AWS configuration with various authentication methods
func loadAWSConfig(awsConfig AWSConfig) (aws.Config, error) {
	var options []func(*config.LoadOptions) error

	// Set region if provided
	if awsConfig.Region != "" {
		options = append(options, config.WithRegion(awsConfig.Region))
	}

	// Set credentials if provided
	if awsConfig.AccessKeyID != "" && awsConfig.SecretAccessKey != "" {
		options = append(options, config.WithCredentialsProvider(
			aws.NewCredentialsCache(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     awsConfig.AccessKeyID,
					SecretAccessKey: awsConfig.SecretAccessKey,
					SessionToken:    awsConfig.SessionToken,
				}, nil
			})),
		))
	}

	// Set profile if provided
	if awsConfig.Profile != "" {
		options = append(options, config.WithSharedConfigProfile(awsConfig.Profile))
	}

	return config.LoadDefaultConfig(context.TODO(), options...)
}
