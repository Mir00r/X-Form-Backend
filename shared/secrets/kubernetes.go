package secrets

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// KubernetesProvider implements SecretProvider for Kubernetes Secrets
type KubernetesProvider struct {
	client    kubernetes.Interface
	config    KubernetesConfig
	logger    *logrus.Logger
	namespace string
}

// NewKubernetesProvider creates a new Kubernetes provider
func NewKubernetesProvider(config KubernetesConfig) (*KubernetesProvider, error) {
	logger := logrus.New()

	// Set default namespace
	namespace := config.Namespace
	if namespace == "" {
		namespace = "default"
	}

	// Create Kubernetes client
	var kubeConfig *rest.Config
	var err error

	if config.InCluster {
		// Use in-cluster configuration
		kubeConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
		}
	} else {
		// Use kubeconfig file
		configPath := config.ConfigPath
		if configPath == "" {
			configPath = clientcmd.RecommendedHomeFile
		}

		kubeConfig, err = clientcmd.BuildConfigFromFlags("", configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create kubeconfig: %w", err)
		}
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &KubernetesProvider{
		client:    clientset,
		config:    config,
		logger:    logger,
		namespace: namespace,
	}, nil
}

// GetSecret retrieves a secret from Kubernetes
func (k *KubernetesProvider) GetSecret(ctx context.Context, key string) (string, error) {
	secretName := k.config.SecretName
	if secretName == "" {
		secretName = "app-secrets"
	}

	k.logger.Debugf("Getting secret from Kubernetes: %s/%s", secretName, key)

	secret, err := k.client.CoreV1().Secrets(k.namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get secret from Kubernetes: %w", err)
	}

	if secret.Data == nil {
		return "", fmt.Errorf("secret %s has no data", secretName)
	}

	if value, exists := secret.Data[key]; exists {
		return string(value), nil
	}

	return "", fmt.Errorf("key %s not found in secret %s", key, secretName)
}

// GetSecrets retrieves multiple secrets from Kubernetes
func (k *KubernetesProvider) GetSecrets(ctx context.Context, keys []string) (map[string]string, error) {
	secretName := k.config.SecretName
	if secretName == "" {
		secretName = "app-secrets"
	}

	k.logger.Debugf("Getting secrets from Kubernetes: %s", secretName)

	secret, err := k.client.CoreV1().Secrets(k.namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret from Kubernetes: %w", err)
	}

	result := make(map[string]string)
	if secret.Data != nil {
		for _, key := range keys {
			if value, exists := secret.Data[key]; exists {
				result[key] = string(value)
			}
		}
	}

	return result, nil
}

// SetSecret stores a secret in Kubernetes
func (k *KubernetesProvider) SetSecret(ctx context.Context, key, value string, metadata map[string]string) error {
	secretName := k.config.SecretName
	if secretName == "" {
		secretName = "app-secrets"
	}

	k.logger.Debugf("Setting secret in Kubernetes: %s/%s", secretName, key)

	// Get existing secret or create new one
	secret, err := k.client.CoreV1().Secrets(k.namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		// Create new secret
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: k.namespace,
			},
			Type: corev1.SecretTypeOpaque,
			Data: make(map[string][]byte),
		}
	}

	// Update the secret data
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	secret.Data[key] = []byte(value)

	// Add metadata as annotations
	if secret.Annotations == nil {
		secret.Annotations = make(map[string]string)
	}
	for k, v := range metadata {
		secret.Annotations[fmt.Sprintf("secrets.x-form/%s", k)] = v
	}

	// Update or create the secret
	if secret.CreationTimestamp.IsZero() {
		_, err = k.client.CoreV1().Secrets(k.namespace).Create(ctx, secret, metav1.CreateOptions{})
	} else {
		_, err = k.client.CoreV1().Secrets(k.namespace).Update(ctx, secret, metav1.UpdateOptions{})
	}

	if err != nil {
		return fmt.Errorf("failed to update Kubernetes secret: %w", err)
	}

	return nil
}

// DeleteSecret removes a secret key from Kubernetes
func (k *KubernetesProvider) DeleteSecret(ctx context.Context, key string) error {
	secretName := k.config.SecretName
	if secretName == "" {
		secretName = "app-secrets"
	}

	k.logger.Debugf("Deleting secret key from Kubernetes: %s/%s", secretName, key)

	// Get existing secret
	secret, err := k.client.CoreV1().Secrets(k.namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get secret from Kubernetes: %w", err)
	}

	// Remove the key
	if secret.Data != nil {
		delete(secret.Data, key)
	}

	// Update the secret
	_, err = k.client.CoreV1().Secrets(k.namespace).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update Kubernetes secret: %w", err)
	}

	return nil
}

// ListSecrets lists all secrets in Kubernetes with optional prefix
func (k *KubernetesProvider) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	secretName := k.config.SecretName
	if secretName == "" {
		secretName = "app-secrets"
	}

	k.logger.Debugf("Listing secrets from Kubernetes: %s", secretName)

	secret, err := k.client.CoreV1().Secrets(k.namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret from Kubernetes: %w", err)
	}

	var keys []string
	if secret.Data != nil {
		for key := range secret.Data {
			if prefix == "" || len(key) >= len(prefix) && key[:len(prefix)] == prefix {
				keys = append(keys, key)
			}
		}
	}

	return keys, nil
}

// RotateSecret rotates a secret in Kubernetes
func (k *KubernetesProvider) RotateSecret(ctx context.Context, key string) error {
	// Generate a new secret value
	newValue, err := GenerateSecretKey(32)
	if err != nil {
		return fmt.Errorf("failed to generate new secret value: %w", err)
	}

	// Set the new value
	metadata := map[string]string{
		"rotated_at": "now",
		"rotated_by": "secret-manager",
	}

	return k.SetSecret(ctx, key, newValue, metadata)
}

// HealthCheck verifies Kubernetes connectivity
func (k *KubernetesProvider) HealthCheck(ctx context.Context) error {
	// Try to list secrets to verify connectivity
	_, err := k.client.CoreV1().Secrets(k.namespace).List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return fmt.Errorf("kubernetes health check failed: %w", err)
	}

	return nil
}

// Close closes the Kubernetes provider
func (k *KubernetesProvider) Close() error {
	// Kubernetes client doesn't require explicit closing
	return nil
}
