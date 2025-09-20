// Package config provides configuration management for the API Gateway
// Following the Single Responsibility Principle and Configuration as Code pattern
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the complete application configuration
// Using struct tags for validation and environment variable mapping
type Config struct {
	// Application metadata
	Version     string `mapstructure:"version" validate:"required"`
	Environment string `mapstructure:"environment" validate:"required,oneof=development staging production"`

	// Server configuration
	Server ServerConfig `mapstructure:"server" validate:"required"`

	// Database configuration (if needed)
	Database DatabaseConfig `mapstructure:"database"`

	// Security configuration
	Security SecurityConfig `mapstructure:"security" validate:"required"`

	// Authentication configuration
	Auth AuthConfig `mapstructure:"auth" validate:"required"`

	// Services configuration for service discovery
	Services ServicesConfig `mapstructure:"services" validate:"required"`

	// Proxy configuration
	Proxy ProxyConfig `mapstructure:"proxy" validate:"required"`

	// Validation configuration
	Validation ValidationConfig `mapstructure:"validation" validate:"required"`

	// Logging configuration
	Log LogConfig `mapstructure:"log" validate:"required"`

	// Metrics configuration
	Metrics MetricsConfig `mapstructure:"metrics"`

	// CORS configuration
	CORS CORSConfig `mapstructure:"cors" validate:"required"`

	// Rate limiting configuration
	RateLimit RateLimitConfig `mapstructure:"rate_limit" validate:"required"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string        `mapstructure:"host" validate:"required"`
	Port         string        `mapstructure:"port" validate:"required,min=1,max=65535"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" validate:"required"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" validate:"required"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout" validate:"required"`
	Timeout      time.Duration `mapstructure:"timeout" validate:"required"`
	// TLS configuration
	TLS TLSConfig `mapstructure:"tls"`
}

// TLSConfig holds TLS/SSL configuration
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
	// Minimum TLS version
	MinVersion string `mapstructure:"min_version" validate:"oneof=1.2 1.3"`
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver" validate:"oneof=postgres mysql sqlite"`
	DSN             string        `mapstructure:"dsn" validate:"required"`
	MaxOpenConns    int           `mapstructure:"max_open_conns" validate:"min=1"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" validate:"min=1"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	// JWT configuration
	JWT JWTConfig `mapstructure:"jwt" validate:"required"`

	// Whitelist configuration for IP filtering
	Whitelist WhitelistConfig `mapstructure:"whitelist"`

	// Rate limiting configuration
	RateLimit RateLimitConfig `mapstructure:"rate_limit" validate:"required"`

	// Security headers
	Headers SecurityHeadersConfig `mapstructure:"headers"`

	// CORS configuration
	CORS CORSConfig `mapstructure:"cors"`

	// Validation configuration
	Validation ValidationConfig `mapstructure:"validation"`
}

// JWTConfig holds JWT-specific configuration
type JWTConfig struct {
	Secret         string        `mapstructure:"secret" validate:"required,min=32"`
	PublicKey      string        `mapstructure:"public_key"`
	PrivateKey     string        `mapstructure:"private_key"`
	Algorithm      string        `mapstructure:"algorithm" validate:"required,oneof=HS256 HS384 HS512 RS256 RS384 RS512"`
	ExpirationTime time.Duration `mapstructure:"expiration_time" validate:"required"`
	RefreshTime    time.Duration `mapstructure:"refresh_time" validate:"required"`
	Issuer         string        `mapstructure:"issuer" validate:"required"`
	Audience       string        `mapstructure:"audience" validate:"required"`
}

// WhitelistConfig holds IP whitelist configuration
type WhitelistConfig struct {
	Enabled     bool     `mapstructure:"enabled"`
	AllowedIPs  []string `mapstructure:"allowed_ips"`
	BlockedIPs  []string `mapstructure:"blocked_ips"`
	TrustProxy  bool     `mapstructure:"trust_proxy"`
	ProxyHeader string   `mapstructure:"proxy_header"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled bool          `mapstructure:"enabled"`
	RPS     int           `mapstructure:"rps" validate:"min=1"`
	Burst   int           `mapstructure:"burst" validate:"min=1"`
	Window  time.Duration `mapstructure:"window" validate:"required"`
	// Per-endpoint rate limits
	Endpoints map[string]EndpointRateLimit `mapstructure:"endpoints"`
}

// EndpointRateLimit holds endpoint-specific rate limiting
type EndpointRateLimit struct {
	RPS    int           `mapstructure:"rps" validate:"min=1"`
	Burst  int           `mapstructure:"burst" validate:"min=1"`
	Window time.Duration `mapstructure:"window" validate:"required"`
}

// SecurityHeadersConfig holds security headers configuration
type SecurityHeadersConfig struct {
	ContentSecurityPolicy string `mapstructure:"content_security_policy"`
	FrameOptions          string `mapstructure:"frame_options" validate:"oneof=DENY SAMEORIGIN"`
	ContentTypeNoSniff    bool   `mapstructure:"content_type_no_sniff"`
	XSSProtection         string `mapstructure:"xss_protection"`
	ReferrerPolicy        string `mapstructure:"referrer_policy"`
}

// AuthConfig holds authentication service configuration
type AuthConfig struct {
	ServiceURL    string        `mapstructure:"service_url" validate:"required,url"`
	Timeout       time.Duration `mapstructure:"timeout" validate:"required"`
	RetryAttempts int           `mapstructure:"retry_attempts" validate:"min=0,max=5"`
	RetryDelay    time.Duration `mapstructure:"retry_delay"`
	// JWKS endpoint for public key validation
	JWKSEndpoint  string        `mapstructure:"jwks_endpoint" validate:"url"`
	JWKSCacheTime time.Duration `mapstructure:"jwks_cache_time"`
}

// ServicesConfig holds configuration for all backend services
type ServicesConfig struct {
	// Service discovery configuration
	Discovery ServiceDiscoveryConfig `mapstructure:"discovery" validate:"required"`

	// Individual service configurations
	Services map[string]ServiceConfig `mapstructure:"services" validate:"required"`
}

// ServiceDiscoveryConfig holds service discovery configuration
type ServiceDiscoveryConfig struct {
	Type     string        `mapstructure:"type" validate:"required,oneof=consul etcd static"`
	Address  string        `mapstructure:"address" validate:"required"`
	Timeout  time.Duration `mapstructure:"timeout" validate:"required"`
	Interval time.Duration `mapstructure:"interval" validate:"required"`
}

// ServiceConfig holds individual service configuration
type ServiceConfig struct {
	Name        string            `mapstructure:"name" validate:"required"`
	URL         string            `mapstructure:"url" validate:"required,url"`
	HealthCheck string            `mapstructure:"health_check" validate:"required"`
	Timeout     time.Duration     `mapstructure:"timeout" validate:"required"`
	Retries     int               `mapstructure:"retries" validate:"min=0,max=5"`
	LoadBalance LoadBalanceConfig `mapstructure:"load_balance"`
	Metadata    map[string]string `mapstructure:"metadata"`
}

// LoadBalanceConfig holds load balancing configuration
type LoadBalanceConfig struct {
	Algorithm string `mapstructure:"algorithm" validate:"oneof=round_robin least_conn weighted"`
	Weight    int    `mapstructure:"weight" validate:"min=1"`
}

// ProxyConfig holds reverse proxy configuration
type ProxyConfig struct {
	Timeout   time.Duration `mapstructure:"timeout" validate:"required"`
	KeepAlive time.Duration `mapstructure:"keep_alive" validate:"required"`

	// Add any other proxy configuration fields here
}

// Load loads the configuration from file and environment variables
func Load() (*Config, error) {
	v := viper.New()

	// Set default configuration file path
	configPath := "./config"
	configName := "config"
	configType := "yaml"

	// Check if config file path is set via environment variable
	if os.Getenv("CONFIG_PATH") != "" {
		configPath = os.Getenv("CONFIG_PATH")
	}

	// Check if config name is set via environment variable
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	// Check if config type is set via environment variable
	if os.Getenv("CONFIG_TYPE") != "" {
		configType = os.Getenv("CONFIG_TYPE")
	}

	// Set configuration file settings
	v.AddConfigPath(configPath)
	v.SetConfigName(configName)
	v.SetConfigType(configType)

	// Set environment variable prefix
	v.SetEnvPrefix("API_GATEWAY")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set default values
	setDefaults(v)

	// Read configuration file
	if err := v.ReadInConfig(); err != nil {
		// If config file is not found, log a warning but continue with defaults and env vars
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal configuration into struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// setDefaults sets default values for configuration
func setDefaults(v *viper.Viper) {
	// Application defaults
	v.SetDefault("version", "1.0.0")
	v.SetDefault("environment", "development")

	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 30)
	v.SetDefault("server.write_timeout", 30)
	v.SetDefault("server.idle_timeout", 60)
	v.SetDefault("server.timeout", 30)

	// Security defaults
	v.SetDefault("security.jwt.algorithm", "HS256")
	v.SetDefault("security.jwt.expiration_time", 3600)
	v.SetDefault("security.jwt.refresh_time", 86400)
	v.SetDefault("security.jwt.issuer", "api-gateway")
	v.SetDefault("security.jwt.audience", "users")

	// Rate limiting defaults
	v.SetDefault("security.rate_limit.enabled", true)
	v.SetDefault("security.rate_limit.rps", 100)
	v.SetDefault("security.rate_limit.burst", 200)
	v.SetDefault("security.rate_limit.window", 60)

	// Proxy defaults
	v.SetDefault("proxy.timeout", 30)
	v.SetDefault("proxy.keep_alive", 60)
	v.SetDefault("proxy.max_idle_conns", 100)
	v.SetDefault("proxy.max_conns_per_host", 100)
	v.SetDefault("proxy.idle_conn_timeout", 90)

	// Logger defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.output", "stdout")

	// CORS defaults
	v.SetDefault("cors.allowed_origins", []string{"*"})
	v.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allowed_headers", []string{"Origin", "Content-Type", "Accept", "Authorization"})
	v.SetDefault("cors.exposed_headers", []string{"Content-Length", "Content-Type"})
	v.SetDefault("cors.allow_credentials", true)
	v.SetDefault("cors.max_age", 86400)
}

// validateConfig validates the configuration
func validateConfig(cfg *Config) error {
	// Basic validation
	port, err := strconv.Atoi(cfg.Server.Port)
	if err != nil || port <= 0 || port > 65535 {
		return fmt.Errorf("invalid server port: %s", cfg.Server.Port)
	}

	// For simplicity, we'll skip complex validation for now
	// In a production environment, you would use a validation library like go-playground/validator

	return nil
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled          bool          `mapstructure:"enabled"`
	Threshold        int           `mapstructure:"threshold" validate:"min=1"`
	Timeout          time.Duration `mapstructure:"timeout" validate:"required"`
	MaxRequests      int           `mapstructure:"max_requests" validate:"min=1"`
	Interval         time.Duration `mapstructure:"interval" validate:"required"`
	FailureThreshold float64       `mapstructure:"failure_threshold" validate:"min=0,max=1"`
}

// ValidationConfig holds parameter validation configuration
type ValidationConfig struct {
	Enabled bool                      `mapstructure:"enabled"`
	Rules   map[string]ValidationRule `mapstructure:"rules"`
}

// ValidationRule holds validation rules for specific endpoints
type ValidationRule struct {
	Methods  []string          `mapstructure:"methods" validate:"required"`
	Required []string          `mapstructure:"required"`
	Optional []string          `mapstructure:"optional"`
	Patterns map[string]string `mapstructure:"patterns"`
	MaxSizes map[string]int    `mapstructure:"max_sizes"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string `mapstructure:"level" validate:"required,oneof=debug info warn error fatal panic"`
	Format string `mapstructure:"format" validate:"required,oneof=json text"`
	Output string `mapstructure:"output" validate:"required"`
	// Structured logging fields
	Fields map[string]interface{} `mapstructure:"fields"`
}

// MetricsConfig holds metrics collection configuration
type MetricsConfig struct {
	Enabled   bool          `mapstructure:"enabled"`
	Port      string        `mapstructure:"port" validate:"min=1,max=65535"`
	Path      string        `mapstructure:"path" validate:"required"`
	Namespace string        `mapstructure:"namespace" validate:"required"`
	Interval  time.Duration `mapstructure:"interval" validate:"required"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	Enabled          bool     `mapstructure:"enabled"`
	AllowedOrigins   []string `mapstructure:"allowed_origins" validate:"required"`
	AllowedMethods   []string `mapstructure:"allowed_methods" validate:"required"`
	AllowedHeaders   []string `mapstructure:"allowed_headers" validate:"required"`
	ExposedHeaders   []string `mapstructure:"exposed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age" validate:"min=0"`
}

// LoadConfig loads configuration from various sources
// Following the 12-factor app methodology for configuration
func LoadConfig() (*Config, error) {
	// Set default configuration file name
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Add configuration search paths
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/xform/")

	// Enable automatic environment variable reading
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("XFORM")

	// Set default values following best practices
	setDefaults(viper.GetViper())

	// Read configuration file (optional - can run on env vars only)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found; rely on environment variables and defaults
	}

	// Unmarshal configuration into struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// GetServiceConfig returns configuration for a specific service
func (c *Config) GetServiceConfig(serviceName string) (ServiceConfig, bool) {
	service, exists := c.Services.Services[serviceName]
	return service, exists
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}
