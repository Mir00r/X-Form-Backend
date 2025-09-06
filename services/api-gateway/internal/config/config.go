// Package config provides centralized configuration management for the API Gateway
// following the "Schema First" principle with environment-based configuration
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config represents the complete configuration structure for the API Gateway
// This follows the "Single Responsibility" principle by centralizing all configuration concerns
type Config struct {
	// Server Configuration
	Server ServerConfig `json:"server" yaml:"server"`

	// Security Configuration - implements "Gateway-level authentication & policy enforcement"
	Security SecurityConfig `json:"security" yaml:"security"`

	// Traefik Integration Configuration
	Traefik TraefikConfig `json:"traefik" yaml:"traefik"`

	// Tyk API Management Configuration
	Tyk TykConfig `json:"tyk" yaml:"tyk"`

	// Microservices Configuration - supports "API as a contract"
	Services ServicesConfig `json:"services" yaml:"services"`

	// Observability Configuration
	Observability ObservabilityConfig `json:"observability" yaml:"observability"`

	// Event-driven Configuration - supports "Event-driven for cross-cutting"
	Events EventsConfig `json:"events" yaml:"events"`

	// Legacy fields for backward compatibility
	Port                string
	MetricsPort         string
	Environment         string
	JWTSecret           string
	Version             string
	AuthServiceURL      string
	FormServiceURL      string
	ResponseServiceURL  string
	AnalyticsServiceURL string
	FileServiceURL      string
	RealtimeServiceURL  string
	RedisURL            string
	KongAdminURL        string
	JaegerURL           string
}

// ServerConfig defines HTTP server configuration
type ServerConfig struct {
	Port         string        `json:"port" yaml:"port"`
	Host         string        `json:"host" yaml:"host"`
	Environment  string        `json:"environment" yaml:"environment"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	TLS          TLSConfig     `json:"tls" yaml:"tls"`
}

// TLSConfig defines TLS configuration for secure communication
type TLSConfig struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`
	CertFile string `json:"cert_file" yaml:"cert_file"`
	KeyFile  string `json:"key_file" yaml:"key_file"`
	MTLSMode string `json:"mtls_mode" yaml:"mtls_mode"` // none, request, require
}

// SecurityConfig implements comprehensive security configuration
// Following "JWKS & RSA keys" and "Gateway-level authentication" principles
type SecurityConfig struct {
	// JWT Configuration
	JWT JWTConfig `json:"jwt" yaml:"jwt"`

	// JWKS Configuration for key rotation
	JWKS JWKSConfig `json:"jwks" yaml:"jwks"`

	// mTLS Configuration for service-to-service communication
	MTLS MTLSConfig `json:"mtls" yaml:"mtls"`

	// Rate Limiting Configuration
	RateLimit RateLimitConfig `json:"rate_limit" yaml:"rate_limit"`

	// CORS Configuration
	CORS CORSConfig `json:"cors" yaml:"cors"`

	// Input Validation Configuration
	Validation ValidationConfig `json:"validation" yaml:"validation"`
}

// JWTConfig defines JWT authentication configuration
type JWTConfig struct {
	Secret         string        `json:"-" yaml:"-"` // Hidden from JSON/YAML output
	Algorithm      string        `json:"algorithm" yaml:"algorithm"`
	Issuer         string        `json:"issuer" yaml:"issuer"`
	Audience       string        `json:"audience" yaml:"audience"`
	ExpirationTime time.Duration `json:"expiration_time" yaml:"expiration_time"`
	RefreshTime    time.Duration `json:"refresh_time" yaml:"refresh_time"`
	KeyID          string        `json:"key_id" yaml:"key_id"`
}

// JWKSConfig defines JSON Web Key Set configuration for key rotation
type JWKSConfig struct {
	Endpoint        string        `json:"endpoint" yaml:"endpoint"`
	CacheTimeout    time.Duration `json:"cache_timeout" yaml:"cache_timeout"`
	RefreshInterval time.Duration `json:"refresh_interval" yaml:"refresh_interval"`
	KeyIDHeader     string        `json:"key_id_header" yaml:"key_id_header"`
}

// MTLSConfig defines mutual TLS configuration for service-to-service communication
type MTLSConfig struct {
	Enabled    bool   `json:"enabled" yaml:"enabled"`
	CACertFile string `json:"ca_cert_file" yaml:"ca_cert_file"`
	CertFile   string `json:"cert_file" yaml:"cert_file"`
	KeyFile    string `json:"key_file" yaml:"key_file"`
	VerifyMode string `json:"verify_mode" yaml:"verify_mode"` // none, optional, strict
}

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	Enabled        bool           `json:"enabled" yaml:"enabled"`
	GlobalLimit    int            `json:"global_limit" yaml:"global_limit"`
	PerUserLimit   int            `json:"per_user_limit" yaml:"per_user_limit"`
	WindowDuration time.Duration  `json:"window_duration" yaml:"window_duration"`
	EndpointLimits map[string]int `json:"endpoint_limits" yaml:"endpoint_limits"`
	UserTierLimits map[string]int `json:"user_tier_limits" yaml:"user_tier_limits"`
	RedisURL       string         `json:"redis_url" yaml:"redis_url"`
}

// CORSConfig defines Cross-Origin Resource Sharing configuration
type CORSConfig struct {
	Enabled          bool     `json:"enabled" yaml:"enabled"`
	AllowedOrigins   []string `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods" yaml:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers" yaml:"allowed_headers"`
	ExposedHeaders   []string `json:"exposed_headers" yaml:"exposed_headers"`
	AllowCredentials bool     `json:"allow_credentials" yaml:"allow_credentials"`
	MaxAge           int      `json:"max_age" yaml:"max_age"`
}

// ValidationConfig defines input validation configuration
type ValidationConfig struct {
	Enabled           bool  `json:"enabled" yaml:"enabled"`
	JSONSchemaEnabled bool  `json:"json_schema_enabled" yaml:"json_schema_enabled"`
	StrictMode        bool  `json:"strict_mode" yaml:"strict_mode"`
	MaxRequestSize    int64 `json:"max_request_size" yaml:"max_request_size"`
}

// TraefikConfig defines Traefik ingress configuration
type TraefikConfig struct {
	Enabled    bool              `json:"enabled" yaml:"enabled"`
	EntryPoint string            `json:"entry_point" yaml:"entry_point"`
	Router     TraefikRouter     `json:"router" yaml:"router"`
	Middleware TraefikMiddleware `json:"middleware" yaml:"middleware"`
	TLS        TraefikTLS        `json:"tls" yaml:"tls"`
}

// TraefikRouter defines Traefik routing configuration
type TraefikRouter struct {
	Rule        string   `json:"rule" yaml:"rule"`
	Priority    int      `json:"priority" yaml:"priority"`
	Service     string   `json:"service" yaml:"service"`
	Middlewares []string `json:"middlewares" yaml:"middlewares"`
}

// TraefikMiddleware defines Traefik middleware configuration
type TraefikMiddleware struct {
	RateLimit      TraefikRateLimit      `json:"rate_limit" yaml:"rate_limit"`
	Retry          TraefikRetry          `json:"retry" yaml:"retry"`
	CircuitBreaker TraefikCircuitBreaker `json:"circuit_breaker" yaml:"circuit_breaker"`
	Headers        TraefikHeaders        `json:"headers" yaml:"headers"`
}

// TraefikRateLimit defines Traefik rate limiting
type TraefikRateLimit struct {
	Average int64 `json:"average" yaml:"average"`
	Burst   int64 `json:"burst" yaml:"burst"`
}

// TraefikRetry defines Traefik retry configuration
type TraefikRetry struct {
	Attempts        int           `json:"attempts" yaml:"attempts"`
	InitialInterval time.Duration `json:"initial_interval" yaml:"initial_interval"`
}

// TraefikCircuitBreaker defines Traefik circuit breaker configuration
type TraefikCircuitBreaker struct {
	Expression        string        `json:"expression" yaml:"expression"`
	CheckPeriod       time.Duration `json:"check_period" yaml:"check_period"`
	FallbackDuration  time.Duration `json:"fallback_duration" yaml:"fallback_duration"`
	RecoveryDuration  time.Duration `json:"recovery_duration" yaml:"recovery_duration"`
	ResponseCodeRatio float64       `json:"response_code_ratio" yaml:"response_code_ratio"`
}

// TraefikHeaders defines Traefik header manipulation
type TraefikHeaders struct {
	CustomRequestHeaders  map[string]string `json:"custom_request_headers" yaml:"custom_request_headers"`
	CustomResponseHeaders map[string]string `json:"custom_response_headers" yaml:"custom_response_headers"`
	SecureHeaders         bool              `json:"secure_headers" yaml:"secure_headers"`
}

// TraefikTLS defines Traefik TLS configuration
type TraefikTLS struct {
	CertResolver string   `json:"cert_resolver" yaml:"cert_resolver"`
	Domains      []string `json:"domains" yaml:"domains"`
	Options      string   `json:"options" yaml:"options"`
}

// TykConfig defines Tyk API Management configuration
type TykConfig struct {
	Enabled      bool         `json:"enabled" yaml:"enabled"`
	GatewayURL   string       `json:"gateway_url" yaml:"gateway_url"`
	DashboardURL string       `json:"dashboard_url" yaml:"dashboard_url"`
	APIKey       string       `json:"-" yaml:"-"` // Hidden from output
	OrgID        string       `json:"org_id" yaml:"org_id"`
	Policies     TykPolicies  `json:"policies" yaml:"policies"`
	Analytics    TykAnalytics `json:"analytics" yaml:"analytics"`
	Portal       TykPortal    `json:"portal" yaml:"portal"`
}

// TykPolicies defines Tyk policy configuration
type TykPolicies struct {
	DefaultPolicy string            `json:"default_policy" yaml:"default_policy"`
	RateLimits    map[string]int    `json:"rate_limits" yaml:"rate_limits"`
	Quotas        map[string]int    `json:"quotas" yaml:"quotas"`
	AccessRights  map[string]string `json:"access_rights" yaml:"access_rights"`
}

// TykAnalytics defines Tyk analytics configuration
type TykAnalytics struct {
	Enabled           bool `json:"enabled" yaml:"enabled"`
	DetailedRecording bool `json:"detailed_recording" yaml:"detailed_recording"`
	GeoIP             bool `json:"geo_ip" yaml:"geo_ip"`
	Retention         int  `json:"retention" yaml:"retention"` // days
}

// TykPortal defines Tyk developer portal configuration
type TykPortal struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	URL     string `json:"url" yaml:"url"`
	Theme   string `json:"theme" yaml:"theme"`
}

// ServicesConfig defines microservice configuration following "API as a contract"
type ServicesConfig struct {
	AuthService          ServiceConfig `json:"auth_service" yaml:"auth_service"`
	FormService          ServiceConfig `json:"form_service" yaml:"form_service"`
	ResponseService      ServiceConfig `json:"response_service" yaml:"response_service"`
	AnalyticsService     ServiceConfig `json:"analytics_service" yaml:"analytics_service"`
	CollaborationService ServiceConfig `json:"collaboration_service" yaml:"collaboration_service"`
	RealtimeService      ServiceConfig `json:"realtime_service" yaml:"realtime_service"`
}

// ServiceConfig defines individual service configuration
type ServiceConfig struct {
	URL            string               `json:"url" yaml:"url"`
	HealthEndpoint string               `json:"health_endpoint" yaml:"health_endpoint"`
	Timeout        time.Duration        `json:"timeout" yaml:"timeout"`
	RetryAttempts  int                  `json:"retry_attempts" yaml:"retry_attempts"`
	CircuitBreaker CircuitBreakerConfig `json:"circuit_breaker" yaml:"circuit_breaker"`
	LoadBalancer   LoadBalancerConfig   `json:"load_balancer" yaml:"load_balancer"`
	TLS            ServiceTLSConfig     `json:"tls" yaml:"tls"`
}

// CircuitBreakerConfig defines circuit breaker configuration for services
type CircuitBreakerConfig struct {
	Enabled       bool          `json:"enabled" yaml:"enabled"`
	Threshold     int           `json:"threshold" yaml:"threshold"`
	Timeout       time.Duration `json:"timeout" yaml:"timeout"`
	MaxRequests   uint32        `json:"max_requests" yaml:"max_requests"`
	Interval      time.Duration `json:"interval" yaml:"interval"`
	OnStateChange string        `json:"on_state_change" yaml:"on_state_change"`
}

// LoadBalancerConfig defines load balancing configuration
type LoadBalancerConfig struct {
	Strategy string   `json:"strategy" yaml:"strategy"` // round_robin, weighted, least_conn
	Backends []string `json:"backends" yaml:"backends"`
	Weights  []int    `json:"weights" yaml:"weights"`
}

// ServiceTLSConfig defines TLS configuration for service communication
type ServiceTLSConfig struct {
	Enabled            bool   `json:"enabled" yaml:"enabled"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`
	CACertFile         string `json:"ca_cert_file" yaml:"ca_cert_file"`
	CertFile           string `json:"cert_file" yaml:"cert_file"`
	KeyFile            string `json:"key_file" yaml:"key_file"`
}

// ObservabilityConfig defines monitoring and observability configuration
type ObservabilityConfig struct {
	Metrics     MetricsConfig     `json:"metrics" yaml:"metrics"`
	Tracing     TracingConfig     `json:"tracing" yaml:"tracing"`
	Logging     LoggingConfig     `json:"logging" yaml:"logging"`
	HealthCheck HealthCheckConfig `json:"health_check" yaml:"health_check"`
}

// MetricsConfig defines metrics collection configuration
type MetricsConfig struct {
	Enabled    bool             `json:"enabled" yaml:"enabled"`
	Port       string           `json:"port" yaml:"port"`
	Path       string           `json:"path" yaml:"path"`
	Namespace  string           `json:"namespace" yaml:"namespace"`
	Prometheus PrometheusConfig `json:"prometheus" yaml:"prometheus"`
}

// PrometheusConfig defines Prometheus-specific configuration
type PrometheusConfig struct {
	Enabled     bool   `json:"enabled" yaml:"enabled"`
	Registry    string `json:"registry" yaml:"registry"`
	PushGateway string `json:"push_gateway" yaml:"push_gateway"`
}

// TracingConfig defines distributed tracing configuration
type TracingConfig struct {
	Enabled     bool     `json:"enabled" yaml:"enabled"`
	Provider    string   `json:"provider" yaml:"provider"` // jaeger, zipkin, otel
	Endpoint    string   `json:"endpoint" yaml:"endpoint"`
	ServiceName string   `json:"service_name" yaml:"service_name"`
	SampleRate  float64  `json:"sample_rate" yaml:"sample_rate"`
	Headers     []string `json:"headers" yaml:"headers"`
}

// LoggingConfig defines logging configuration
type LoggingConfig struct {
	Level       string        `json:"level" yaml:"level"`
	Format      string        `json:"format" yaml:"format"` // json, text
	Output      []string      `json:"output" yaml:"output"` // stdout, file, syslog
	File        LogFileConfig `json:"file" yaml:"file"`
	Structured  bool          `json:"structured" yaml:"structured"`
	Correlation bool          `json:"correlation" yaml:"correlation"`
}

// LogFileConfig defines file logging configuration
type LogFileConfig struct {
	Path       string `json:"path" yaml:"path"`
	MaxSize    int    `json:"max_size" yaml:"max_size"` // MB
	MaxBackups int    `json:"max_backups" yaml:"max_backups"`
	MaxAge     int    `json:"max_age" yaml:"max_age"` // days
	Compress   bool   `json:"compress" yaml:"compress"`
}

// HealthCheckConfig defines health check configuration
type HealthCheckConfig struct {
	Enabled     bool          `json:"enabled" yaml:"enabled"`
	Path        string        `json:"path" yaml:"path"`
	Interval    time.Duration `json:"interval" yaml:"interval"`
	Timeout     time.Duration `json:"timeout" yaml:"timeout"`
	Retries     int           `json:"retries" yaml:"retries"`
	StartPeriod time.Duration `json:"start_period" yaml:"start_period"`
}

// EventsConfig defines event-driven configuration for cross-cutting concerns
type EventsConfig struct {
	Enabled  bool           `json:"enabled" yaml:"enabled"`
	Provider string         `json:"provider" yaml:"provider"` // kafka, nats, redis, rabbitmq
	Kafka    KafkaConfig    `json:"kafka" yaml:"kafka"`
	Redis    RedisConfig    `json:"redis" yaml:"redis"`
	NATS     NATSConfig     `json:"nats" yaml:"nats"`
	RabbitMQ RabbitMQConfig `json:"rabbitmq" yaml:"rabbitmq"`
}

// KafkaConfig defines Kafka event streaming configuration
type KafkaConfig struct {
	Brokers     []string          `json:"brokers" yaml:"brokers"`
	ClientID    string            `json:"client_id" yaml:"client_id"`
	GroupID     string            `json:"group_id" yaml:"group_id"`
	Topics      map[string]string `json:"topics" yaml:"topics"`
	Compression string            `json:"compression" yaml:"compression"`
	TLS         EventsTLSConfig   `json:"tls" yaml:"tls"`
}

// RedisConfig defines Redis configuration for events
type RedisConfig struct {
	URL      string          `json:"url" yaml:"url"`
	Password string          `json:"-" yaml:"-"` // Hidden
	DB       int             `json:"db" yaml:"db"`
	Prefix   string          `json:"prefix" yaml:"prefix"`
	TLS      EventsTLSConfig `json:"tls" yaml:"tls"`
}

// NATSConfig defines NATS configuration for events
type NATSConfig struct {
	URL     string          `json:"url" yaml:"url"`
	Subject string          `json:"subject" yaml:"subject"`
	Queue   string          `json:"queue" yaml:"queue"`
	Token   string          `json:"-" yaml:"-"` // Hidden
	TLS     EventsTLSConfig `json:"tls" yaml:"tls"`
}

// RabbitMQConfig defines RabbitMQ configuration for events
type RabbitMQConfig struct {
	URL          string          `json:"url" yaml:"url"`
	Exchange     string          `json:"exchange" yaml:"exchange"`
	ExchangeType string          `json:"exchange_type" yaml:"exchange_type"`
	Queue        string          `json:"queue" yaml:"queue"`
	RoutingKey   string          `json:"routing_key" yaml:"routing_key"`
	TLS          EventsTLSConfig `json:"tls" yaml:"tls"`
}

// EventsTLSConfig defines TLS configuration for event systems
type EventsTLSConfig struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`
	CertFile string `json:"cert_file" yaml:"cert_file"`
	KeyFile  string `json:"key_file" yaml:"key_file"`
	CAFile   string `json:"ca_file" yaml:"ca_file"`
}

// Load configuration with enhanced structure and backward compatibility
func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		// Legacy fields for backward compatibility
		Port:        getEnv("PORT", "8080"),
		MetricsPort: getEnv("METRICS_PORT", "9090"),
		Environment: getEnv("ENVIRONMENT", "development"),
		JWTSecret:   getEnv("JWT_SECRET", "your-jwt-secret-key"),
		Version:     getEnv("VERSION", "1.0.0"),

		// Service URLs
		AuthServiceURL:      getEnv("AUTH_SERVICE_URL", "http://auth-service:3001"),
		FormServiceURL:      getEnv("FORM_SERVICE_URL", "http://form-service:8001"),
		ResponseServiceURL:  getEnv("RESPONSE_SERVICE_URL", "http://response-service:3002"),
		AnalyticsServiceURL: getEnv("ANALYTICS_SERVICE_URL", "http://analytics-service:5001"),
		FileServiceURL:      getEnv("FILE_SERVICE_URL", "http://file-service:3003"),
		RealtimeServiceURL:  getEnv("REALTIME_SERVICE_URL", "http://realtime-service:8002"),

		RedisURL:     getEnv("REDIS_URL", "redis://redis:6379"),
		KongAdminURL: getEnv("KONG_ADMIN_URL", "http://kong:8001"),
		JaegerURL:    getEnv("JAEGER_URL", "http://jaeger:14268/api/traces"),

		// Enhanced configuration structure
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			Host:         getEnv("HOST", "0.0.0.0"),
			Environment:  getEnv("ENVIRONMENT", "development"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
			TLS: TLSConfig{
				Enabled:  getBoolEnv("SERVER_TLS_ENABLED", false),
				CertFile: getEnv("SERVER_TLS_CERT_FILE", ""),
				KeyFile:  getEnv("SERVER_TLS_KEY_FILE", ""),
				MTLSMode: getEnv("SERVER_MTLS_MODE", "none"),
			},
		},
		Security: SecurityConfig{
			JWT: JWTConfig{
				Secret:         getEnv("JWT_SECRET", "your-jwt-secret-key"),
				Algorithm:      getEnv("JWT_ALGORITHM", "HS256"),
				Issuer:         getEnv("JWT_ISSUER", "x-form-api-gateway"),
				Audience:       getEnv("JWT_AUDIENCE", "x-form-services"),
				ExpirationTime: getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
				RefreshTime:    getDurationEnv("JWT_REFRESH", 7*24*time.Hour),
				KeyID:          getEnv("JWT_KEY_ID", "default"),
			},
			JWKS: JWKSConfig{
				Endpoint:        getEnv("JWKS_ENDPOINT", ""),
				CacheTimeout:    getDurationEnv("JWKS_CACHE_TIMEOUT", 1*time.Hour),
				RefreshInterval: getDurationEnv("JWKS_REFRESH_INTERVAL", 5*time.Minute),
				KeyIDHeader:     getEnv("JWKS_KEY_ID_HEADER", "kid"),
			},
			MTLS: MTLSConfig{
				Enabled:    getBoolEnv("MTLS_ENABLED", false),
				CACertFile: getEnv("MTLS_CA_CERT_FILE", ""),
				CertFile:   getEnv("MTLS_CERT_FILE", ""),
				KeyFile:    getEnv("MTLS_KEY_FILE", ""),
				VerifyMode: getEnv("MTLS_VERIFY_MODE", "none"),
			},
			RateLimit: RateLimitConfig{
				Enabled:        getBoolEnv("RATE_LIMIT_ENABLED", true),
				GlobalLimit:    getIntEnv("RATE_LIMIT_GLOBAL", 1000),
				PerUserLimit:   getIntEnv("RATE_LIMIT_PER_USER", 100),
				WindowDuration: getDurationEnv("RATE_LIMIT_WINDOW", 1*time.Hour),
				RedisURL:       getEnv("REDIS_URL", "redis://redis:6379"),
			},
			CORS: CORSConfig{
				Enabled:          getBoolEnv("CORS_ENABLED", true),
				AllowedOrigins:   getSliceEnv("CORS_ALLOWED_ORIGINS", []string{"*"}),
				AllowedMethods:   getSliceEnv("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
				AllowedHeaders:   getSliceEnv("CORS_ALLOWED_HEADERS", []string{"*"}),
				ExposedHeaders:   getSliceEnv("CORS_EXPOSED_HEADERS", []string{"Content-Length"}),
				AllowCredentials: getBoolEnv("CORS_ALLOW_CREDENTIALS", true),
				MaxAge:           getIntEnv("CORS_MAX_AGE", 86400),
			},
			Validation: ValidationConfig{
				Enabled:           getBoolEnv("VALIDATION_ENABLED", true),
				JSONSchemaEnabled: getBoolEnv("JSON_SCHEMA_ENABLED", true),
				StrictMode:        getBoolEnv("VALIDATION_STRICT_MODE", false),
				MaxRequestSize:    getInt64Env("MAX_REQUEST_SIZE", 10<<20), // 10MB
			},
		},
		Traefik: TraefikConfig{
			Enabled:    getBoolEnv("TRAEFIK_ENABLED", false),
			EntryPoint: getEnv("TRAEFIK_ENTRY_POINT", "websecure"),
		},
		Tyk: TykConfig{
			Enabled:      getBoolEnv("TYK_ENABLED", false),
			GatewayURL:   getEnv("TYK_GATEWAY_URL", ""),
			DashboardURL: getEnv("TYK_DASHBOARD_URL", ""),
			APIKey:       getEnv("TYK_API_KEY", ""),
			OrgID:        getEnv("TYK_ORG_ID", ""),
		},
		Services: ServicesConfig{
			AuthService: ServiceConfig{
				URL:            getEnv("AUTH_SERVICE_URL", "http://auth-service:3001"),
				HealthEndpoint: getEnv("AUTH_SERVICE_HEALTH", "/health"),
				Timeout:        getDurationEnv("AUTH_SERVICE_TIMEOUT", 30*time.Second),
				RetryAttempts:  getIntEnv("AUTH_SERVICE_RETRIES", 3),
			},
			FormService: ServiceConfig{
				URL:            getEnv("FORM_SERVICE_URL", "http://form-service:8001"),
				HealthEndpoint: getEnv("FORM_SERVICE_HEALTH", "/health"),
				Timeout:        getDurationEnv("FORM_SERVICE_TIMEOUT", 30*time.Second),
				RetryAttempts:  getIntEnv("FORM_SERVICE_RETRIES", 3),
			},
			ResponseService: ServiceConfig{
				URL:            getEnv("RESPONSE_SERVICE_URL", "http://response-service:3002"),
				HealthEndpoint: getEnv("RESPONSE_SERVICE_HEALTH", "/health"),
				Timeout:        getDurationEnv("RESPONSE_SERVICE_TIMEOUT", 30*time.Second),
				RetryAttempts:  getIntEnv("RESPONSE_SERVICE_RETRIES", 3),
			},
			AnalyticsService: ServiceConfig{
				URL:            getEnv("ANALYTICS_SERVICE_URL", "http://analytics-service:5001"),
				HealthEndpoint: getEnv("ANALYTICS_SERVICE_HEALTH", "/health"),
				Timeout:        getDurationEnv("ANALYTICS_SERVICE_TIMEOUT", 30*time.Second),
				RetryAttempts:  getIntEnv("ANALYTICS_SERVICE_RETRIES", 3),
			},
			CollaborationService: ServiceConfig{
				URL:            getEnv("COLLABORATION_SERVICE_URL", "http://collaboration-service:8004"),
				HealthEndpoint: getEnv("COLLABORATION_SERVICE_HEALTH", "/health"),
				Timeout:        getDurationEnv("COLLABORATION_SERVICE_TIMEOUT", 30*time.Second),
				RetryAttempts:  getIntEnv("COLLABORATION_SERVICE_RETRIES", 3),
			},
			RealtimeService: ServiceConfig{
				URL:            getEnv("REALTIME_SERVICE_URL", "http://realtime-service:8002"),
				HealthEndpoint: getEnv("REALTIME_SERVICE_HEALTH", "/health"),
				Timeout:        getDurationEnv("REALTIME_SERVICE_TIMEOUT", 30*time.Second),
				RetryAttempts:  getIntEnv("REALTIME_SERVICE_RETRIES", 3),
			},
		},
		Observability: ObservabilityConfig{
			Metrics: MetricsConfig{
				Enabled:   getBoolEnv("METRICS_ENABLED", true),
				Port:      getEnv("METRICS_PORT", "9090"),
				Path:      getEnv("METRICS_PATH", "/metrics"),
				Namespace: getEnv("METRICS_NAMESPACE", "x_form_api_gateway"),
			},
			Tracing: TracingConfig{
				Enabled:     getBoolEnv("TRACING_ENABLED", false),
				Provider:    getEnv("TRACING_PROVIDER", "jaeger"),
				Endpoint:    getEnv("TRACING_ENDPOINT", ""),
				ServiceName: getEnv("TRACING_SERVICE_NAME", "x-form-api-gateway"),
				SampleRate:  getFloat64Env("TRACING_SAMPLE_RATE", 0.1),
			},
			Logging: LoggingConfig{
				Level:       getEnv("LOG_LEVEL", "info"),
				Format:      getEnv("LOG_FORMAT", "json"),
				Output:      getSliceEnv("LOG_OUTPUT", []string{"stdout"}),
				Structured:  getBoolEnv("LOG_STRUCTURED", true),
				Correlation: getBoolEnv("LOG_CORRELATION", true),
			},
			HealthCheck: HealthCheckConfig{
				Enabled:     getBoolEnv("HEALTH_CHECK_ENABLED", true),
				Path:        getEnv("HEALTH_CHECK_PATH", "/health"),
				Interval:    getDurationEnv("HEALTH_CHECK_INTERVAL", 30*time.Second),
				Timeout:     getDurationEnv("HEALTH_CHECK_TIMEOUT", 5*time.Second),
				Retries:     getIntEnv("HEALTH_CHECK_RETRIES", 3),
				StartPeriod: getDurationEnv("HEALTH_CHECK_START_PERIOD", 30*time.Second),
			},
		},
		Events: EventsConfig{
			Enabled:  getBoolEnv("EVENTS_ENABLED", false),
			Provider: getEnv("EVENTS_PROVIDER", "redis"),
			Redis: RedisConfig{
				URL:    getEnv("EVENTS_REDIS_URL", "redis://redis:6379"),
				DB:     getIntEnv("EVENTS_REDIS_DB", 0),
				Prefix: getEnv("EVENTS_REDIS_PREFIX", "x-form:events:"),
			},
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Printf("Configuration validation failed: %v", err)
	}

	return config
}

// Validate validates the configuration for completeness and correctness
func (c *Config) Validate() error {
	// Validate JWT configuration
	if c.Security.JWT.Secret == "" && c.Security.JWKS.Endpoint == "" {
		return fmt.Errorf("either JWT_SECRET or JWKS_ENDPOINT must be provided")
	}

	// Validate service URLs
	services := map[string]string{
		"auth-service":          c.Services.AuthService.URL,
		"form-service":          c.Services.FormService.URL,
		"response-service":      c.Services.ResponseService.URL,
		"analytics-service":     c.Services.AnalyticsService.URL,
		"collaboration-service": c.Services.CollaborationService.URL,
		"realtime-service":      c.Services.RealtimeService.URL,
	}

	for name, url := range services {
		if url == "" {
			log.Printf("Warning: %s URL not configured", name)
		}
	}

	return nil
}

// Utility functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getFloat64Env(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
