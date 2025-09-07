# Secrets Management Module

This module provides a comprehensive secrets management solution for the X-Form Backend microservices architecture. It supports multiple secret providers with automatic fallback, caching, and production-ready security features.

## Features

### 🔐 Multiple Provider Support
- **HashiCorp Vault**: Production-grade secret management with multiple auth methods
- **AWS Secrets Manager**: Managed secrets service with automatic rotation
- **AWS Systems Manager Parameter Store**: Hierarchical parameter storage with encryption
- **Kubernetes Secrets**: Native Kubernetes secret integration
- **Environment Variables**: Simple environment-based secrets for development
- **File-based**: JSON/YAML/Properties file storage with optional encryption

### 🔄 Provider Fallback
- Automatic failover between providers
- Configurable fallback chains
- Health monitoring and circuit breaking
- Zero-downtime secret access

### ⚡ Performance & Caching
- In-memory caching with configurable TTL
- Cache warming and preloading
- Automatic cache invalidation
- Performance metrics and monitoring

### 🛡️ Security Features
- Encryption at rest and in transit
- Audit logging for all secret operations
- Secret rotation automation
- Access control and authentication

### 🔧 Production Ready
- Comprehensive error handling
- Retry mechanisms with exponential backoff
- Health checks and monitoring
- Graceful degradation

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"
    
    "github.com/kamkaiz/x-form-backend/shared/secrets"
)

func main() {
    // Load configuration
    config := secrets.GetDefaultConfig()
    config.Provider = secrets.ProviderTypeEnvironment
    
    // Create secret manager
    sm, err := secrets.NewSecretManager(*config)
    if err != nil {
        log.Fatal(err)
    }
    defer sm.Close()
    
    ctx := context.Background()
    
    // Get a secret
    dbPassword, err := sm.GetSecret(ctx, "DATABASE_PASSWORD")
    if err != nil {
        log.Fatal(err)
    }
    
    // Get multiple secrets
    secrets, err := sm.GetSecrets(ctx, []string{
        "DATABASE_PASSWORD",
        "JWT_SECRET",
        "API_KEY",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Use secrets in your application
    connectToDatabase(dbPassword)
    configureJWT(secrets["JWT_SECRET"])
}
```

### Configuration Examples

#### Development Environment
```go
config := secrets.GetDevConfig()
// Uses environment variables with file fallback
// Short cache TTL for development
// Audit logging enabled
```

#### Production Environment
```go
config := secrets.GetProdConfig()
// Uses Vault with AWS Secrets Manager fallback
// Long cache TTL for performance
// Encryption and audit logging enabled
```

#### Kubernetes Deployment
```go
config := secrets.GetKubernetesConfig()
// Uses Kubernetes secrets with Vault fallback
// In-cluster authentication
// Service account token authentication
```

## Provider Configuration

### HashiCorp Vault

```yaml
provider: vault
vault:
  address: https://vault.company.com:8200
  mount_path: secret
  namespace: myapp
  auth:
    method: kubernetes
    parameters:
      role: myapp-role
  tls:
    enabled: true
    ca_cert: /etc/ssl/vault-ca.pem
```

### AWS Secrets Manager

```yaml
provider: aws-secrets
aws:
  region: us-east-1
  role_arn: arn:aws:iam::123456789012:role/SecretsRole
```

### Kubernetes Secrets

```yaml
provider: kubernetes
kubernetes:
  namespace: myapp
  secret_name: app-secrets
  in_cluster: true
```

## Integration with Microservices

### Auth Service Integration

```go
// services/auth-service/src/config/secrets.go
package config

import (
    "context"
    "github.com/kamkaiz/x-form-backend/shared/secrets"
)

type AuthConfig struct {
    JWTSecret        string
    DatabaseURL      string
    OAuthClientID    string
    OAuthClientSecret string
}

func LoadAuthConfig() (*AuthConfig, error) {
    // Load secrets configuration
    secretsConfig := secrets.GetKubernetesConfig()
    sm, err := secrets.NewSecretManager(*secretsConfig)
    if err != nil {
        return nil, err
    }
    defer sm.Close()
    
    ctx := context.Background()
    
    // Get all required secrets
    secretKeys := []string{
        "JWT_SECRET",
        "DATABASE_URL", 
        "OAUTH_CLIENT_ID",
        "OAUTH_CLIENT_SECRET",
    }
    
    secretValues, err := sm.GetSecrets(ctx, secretKeys)
    if err != nil {
        return nil, err
    }
    
    return &AuthConfig{
        JWTSecret:         secretValues["JWT_SECRET"],
        DatabaseURL:       secretValues["DATABASE_URL"],
        OAuthClientID:     secretValues["OAUTH_CLIENT_ID"],
        OAuthClientSecret: secretValues["OAUTH_CLIENT_SECRET"],
    }, nil
}
```

### Form Service Integration

```go
// services/form-service/internal/config/config.go
package config

import (
    "github.com/kamkaiz/x-form-backend/shared/secrets"
)

type Config struct {
    Database struct {
        URL      string
        Password string
    }
    Redis struct {
        URL      string
        Password string
    }
    Storage struct {
        AccessKey string
        SecretKey string
    }
}

func Load() (*Config, error) {
    sm, err := secrets.NewSecretManager(*secrets.GetProdConfig())
    if err != nil {
        return nil, err
    }
    defer sm.Close()
    
    ctx := context.Background()
    
    // Get all secrets at once for better performance
    allSecrets, err := sm.GetSecrets(ctx, []string{
        "POSTGRES_PASSWORD",
        "REDIS_PASSWORD", 
        "S3_ACCESS_KEY",
        "S3_SECRET_KEY",
    })
    if err != nil {
        return nil, err
    }
    
    return &Config{
        Database: struct {
            URL      string
            Password string
        }{
            URL:      os.Getenv("DATABASE_URL"),
            Password: allSecrets["POSTGRES_PASSWORD"],
        },
        Redis: struct {
            URL      string
            Password string
        }{
            URL:      os.Getenv("REDIS_URL"),
            Password: allSecrets["REDIS_PASSWORD"],
        },
        Storage: struct {
            AccessKey string
            SecretKey string
        }{
            AccessKey: allSecrets["S3_ACCESS_KEY"],
            SecretKey: allSecrets["S3_SECRET_KEY"],
        },
    }, nil
}
```

## Kubernetes Integration

### Helm Chart Integration

```yaml
# infrastructure/helm/x-form-backend/templates/secrets.yaml
{{- if .Values.secrets.vault.enabled }}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ include "x-form-backend.fullname" . }}-vault
  annotations:
    vault.hashicorp.com/auth-method: kubernetes
    vault.hashicorp.com/role: {{ .Values.secrets.vault.role }}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "x-form-backend.fullname" . }}-secrets-config
data:
  config.yaml: |
    provider: vault
    fallbacks: [kubernetes, environment]
    vault:
      address: {{ .Values.secrets.vault.address }}
      auth:
        method: kubernetes
        parameters:
          role: {{ .Values.secrets.vault.role }}
    cache:
      enabled: true
      ttl: {{ .Values.secrets.cache.ttl }}
{{- end }}
```

### Deployment with Secrets

```yaml
# infrastructure/helm/x-form-backend/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      serviceAccountName: {{ include "x-form-backend.fullname" . }}-vault
      containers:
      - name: auth-service
        image: {{ .Values.authService.image }}
        env:
        - name: SECRETS_CONFIG_PATH
          value: /etc/secrets/config.yaml
        volumeMounts:
        - name: secrets-config
          mountPath: /etc/secrets
          readOnly: true
        # Vault Agent sidecar for automatic token renewal
      - name: vault-agent
        image: vault:1.15.0
        command: ['vault', 'agent', '-config=/etc/vault/config.hcl']
        volumeMounts:
        - name: vault-config
          mountPath: /etc/vault
      volumes:
      - name: secrets-config
        configMap:
          name: {{ include "x-form-backend.fullname" . }}-secrets-config
```

## CI/CD Pipeline Integration

### GitHub Actions Secret Management

```yaml
# .github/workflows/ci-cd.yml
- name: Deploy Secrets to Vault
  run: |
    # Install Vault CLI
    curl -fsSL https://releases.hashicorp.com/vault/1.15.0/vault_1.15.0_linux_amd64.zip -o vault.zip
    unzip vault.zip && chmod +x vault
    
    # Authenticate with Vault
    export VAULT_ADDR=${{ secrets.VAULT_ADDR }}
    ./vault auth -method=github token=${{ secrets.GITHUB_TOKEN }}
    
    # Deploy secrets for environment
    ./vault kv put secret/x-form-backend/prod \
      database_password="${{ secrets.DATABASE_PASSWORD }}" \
      jwt_secret="${{ secrets.JWT_SECRET }}" \
      api_key="${{ secrets.API_KEY }}"

- name: Update Kubernetes Secrets
  run: |
    # Update Kubernetes secrets as fallback
    kubectl create secret generic app-secrets \
      --from-literal=DATABASE_PASSWORD="${{ secrets.DATABASE_PASSWORD }}" \
      --from-literal=JWT_SECRET="${{ secrets.JWT_SECRET }}" \
      --dry-run=client -o yaml | kubectl apply -f -
```

## Monitoring and Observability

### Metrics Collection

```go
// Add to your monitoring setup
package monitoring

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/kamkaiz/x-form-backend/shared/secrets"
)

var (
    secretsRetrieved = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "secrets_retrieved_total",
            Help: "Total number of secrets retrieved",
        },
        []string{"provider", "status"},
    )
    
    secretsCacheHitRate = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "secrets_cache_hit_rate",
            Help: "Cache hit rate for secrets",
        },
    )
)

func MonitorSecrets(sm *secrets.SecretManager) {
    // Collect cache metrics
    if stats := sm.GetCacheStats(); stats != nil {
        if hitRate, ok := stats["hit_rate"].(float64); ok {
            secretsCacheHitRate.Set(hitRate)
        }
    }
}
```

### Health Check Endpoint

```go
// Add to your health check endpoints
func SecretsHealthCheck(sm *secrets.SecretManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := context.Background()
        
        if err := sm.HealthCheck(ctx); err != nil {
            c.JSON(503, gin.H{
                "status": "unhealthy",
                "error":  err.Error(),
            })
            return
        }
        
        // Get cache stats
        stats := sm.GetCacheStats()
        
        c.JSON(200, gin.H{
            "status": "healthy",
            "cache":  stats,
        })
    }
}
```

## Security Best Practices

### 1. Least Privilege Access
- Configure minimal required permissions for each service
- Use service-specific Vault policies
- Implement role-based access control

### 2. Secret Rotation
- Implement automatic secret rotation policies
- Use short-lived tokens where possible
- Monitor for stale or expired secrets

### 3. Audit Logging
- Enable comprehensive audit logging
- Monitor for unusual access patterns
- Set up alerts for security events

### 4. Network Security
- Use TLS for all secret communications
- Implement network policies in Kubernetes
- Restrict access to secret storage systems

## Migration Guide

### Migrating from Environment Variables

1. **Inventory Current Secrets**
   ```bash
   # Find all environment variables used
   grep -r "process.env\|os.Getenv" services/
   ```

2. **Create Secret Mapping**
   ```go
   // Create migration mapping
   secretMapping := map[string]string{
       "DATABASE_URL":      "database/connection-string",
       "JWT_SECRET":        "auth/jwt-signing-key", 
       "OAUTH_CLIENT_SECRET": "auth/oauth-client-secret",
   }
   ```

3. **Deploy Secrets to Vault**
   ```bash
   # Deploy using Terraform or Vault CLI
   vault kv put secret/x-form-backend/prod @secrets.json
   ```

4. **Update Application Code**
   ```go
   // Replace direct environment access
   // Before:
   dbURL := os.Getenv("DATABASE_URL")
   
   // After:
   dbURL, err := sm.GetSecret(ctx, "database/connection-string")
   ```

5. **Test and Validate**
   ```bash
   # Run integration tests
   go test ./...
   
   # Verify secret access in staging
   kubectl exec -it pod/auth-service -- sh -c "curl localhost:8080/health"
   ```

## Troubleshooting

### Common Issues

#### 1. Vault Connection Errors
```
Error: failed to create vault client: Get "https://vault:8200/v1/sys/health": dial tcp: lookup vault on 10.96.0.10:53: no such host
```

**Solution**: Check Vault address and network connectivity
```bash
# Test Vault connectivity
nslookup vault.default.svc.cluster.local
curl -k https://vault:8200/v1/sys/health
```

#### 2. Authentication Failures
```
Error: kubernetes authentication failed: permission denied
```

**Solution**: Verify Kubernetes service account and Vault role configuration
```bash
# Check service account
kubectl get serviceaccount vault-auth -o yaml

# Verify Vault role
vault read auth/kubernetes/role/myapp-role
```

#### 3. Cache Performance Issues
```
Cache hit rate below 50%
```

**Solution**: Tune cache configuration
```go
config.Cache.TTL = 15 * time.Minute  // Increase TTL
config.Cache.MaxEntries = 2000       // Increase capacity
```

## Performance Optimization

### Cache Optimization
- Set appropriate TTL based on secret change frequency
- Pre-warm cache with critical secrets at startup
- Monitor cache hit rates and adjust accordingly

### Provider Selection
- Use fastest provider as primary (usually local cache/file)
- Configure geographically close providers
- Implement circuit breakers for failed providers

### Batch Operations
- Use `GetSecrets()` for multiple secrets
- Minimize individual secret requests
- Implement secret prefetching for predictable access patterns

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Run the test suite: `go test ./...`
5. Submit a pull request

## License

This module is part of the X-Form Backend project and follows the same licensing terms.
