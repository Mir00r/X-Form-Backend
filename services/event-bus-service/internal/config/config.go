// Package config provides configuration management for the Event Bus Service
// This package implements a comprehensive configuration system that supports
// environment variables, configuration files, and default values following
// the 12-factor app methodology and enterprise best practices.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Config represents the complete configuration for the Event Bus Service
// It includes all subsystem configurations required for enterprise-grade
// event streaming and change data capture operations.
type Config struct {
	// Server configuration for HTTP API and health endpoints
	Server ServerConfig `mapstructure:"server" yaml:"server" json:"server"`

	// Environment and deployment configuration
	Environment string `mapstructure:"environment" yaml:"environment" json:"environment"`
	Version     string `mapstructure:"version" yaml:"version" json:"version"`

	// Kafka configuration for event streaming
	Kafka KafkaConfig `mapstructure:"kafka" yaml:"kafka" json:"kafka"`

	// Debezium configuration for Change Data Capture
	Debezium DebeziumConfig `mapstructure:"debezium" yaml:"debezium" json:"debezium"`

	// Database configuration for multiple database connections
	Databases DatabasesConfig `mapstructure:"databases" yaml:"databases" json:"databases"`

	// Redis configuration for caching and state management
	Redis RedisConfig `mapstructure:"redis" yaml:"redis" json:"redis"`

	// Security configuration for authentication and authorization
	Security SecurityConfig `mapstructure:"security" yaml:"security" json:"security"`

	// Observability configuration for metrics, logging, and tracing
	Observability ObservabilityConfig `mapstructure:"observability" yaml:"observability" json:"observability"`

	// Event processing configuration
	EventProcessing EventProcessingConfig `mapstructure:"event_processing" yaml:"event_processing" json:"event_processing"`

	// Service discovery and integration configuration
	Services ServicesConfig `mapstructure:"services" yaml:"services" json:"services"`

	// Rate limiting and circuit breaker configuration
	RateLimiting RateLimitingConfig `mapstructure:"rate_limiting" yaml:"rate_limiting" json:"rate_limiting"`
}

// ServerConfig defines HTTP server configuration
type ServerConfig struct {
	Host string `mapstructure:"host" yaml:"host" json:"host"`
	Port string `mapstructure:"port" yaml:"port" json:"port"`

	// Timeout configurations for robust server operation
	ReadTimeout  time.Duration `mapstructure:"read_timeout" yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout" yaml:"idle_timeout" json:"idle_timeout"`

	// TLS configuration for secure communication
	TLS TLSConfig `mapstructure:"tls" yaml:"tls" json:"tls"`

	// CORS configuration for web client support
	CORS CORSConfig `mapstructure:"cors" yaml:"cors" json:"cors"`
}

// TLSConfig defines TLS/SSL configuration
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	CertFile string `mapstructure:"cert_file" yaml:"cert_file" json:"cert_file"`
	KeyFile  string `mapstructure:"key_file" yaml:"key_file" json:"key_file"`
}

// CORSConfig defines Cross-Origin Resource Sharing configuration
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins" yaml:"allowed_origins" json:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods" yaml:"allowed_methods" json:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers" yaml:"allowed_headers" json:"allowed_headers"`
}

// KafkaConfig defines Apache Kafka configuration for event streaming
type KafkaConfig struct {
	// Broker configuration
	Brokers   []string `mapstructure:"brokers" yaml:"brokers" json:"brokers"`
	ClientID  string   `mapstructure:"client_id" yaml:"client_id" json:"client_id"`
	Version   string   `mapstructure:"version" yaml:"version" json:"version"`
	Mechanism string   `mapstructure:"mechanism" yaml:"mechanism" json:"mechanism"`

	// Security configuration for Kafka
	Security KafkaSecurityConfig `mapstructure:"security" yaml:"security" json:"security"`

	// Producer configuration for publishing events
	Producer KafkaProducerConfig `mapstructure:"producer" yaml:"producer" json:"producer"`

	// Consumer configuration for consuming events
	Consumer KafkaConsumerConfig `mapstructure:"consumer" yaml:"consumer" json:"consumer"`

	// Admin configuration for topic management
	Admin KafkaAdminConfig `mapstructure:"admin" yaml:"admin" json:"admin"`

	// Schema Registry configuration for Avro/JSON Schema support
	SchemaRegistry SchemaRegistryConfig `mapstructure:"schema_registry" yaml:"schema_registry" json:"schema_registry"`
}

// KafkaSecurityConfig defines Kafka security settings
type KafkaSecurityConfig struct {
	Protocol string `mapstructure:"protocol" yaml:"protocol" json:"protocol"` // PLAINTEXT, SASL_PLAINTEXT, SASL_SSL, SSL
	Username string `mapstructure:"username" yaml:"username" json:"username"`
	Password string `mapstructure:"password" yaml:"password" json:"password"`

	// TLS configuration for Kafka
	TLS KafkaTLSConfig `mapstructure:"tls" yaml:"tls" json:"tls"`

	// SASL configuration
	SASL KafkaSASLConfig `mapstructure:"sasl" yaml:"sasl" json:"sasl"`
}

// KafkaTLSConfig defines Kafka TLS configuration
type KafkaTLSConfig struct {
	Enabled            bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	CertFile           string `mapstructure:"cert_file" yaml:"cert_file" json:"cert_file"`
	KeyFile            string `mapstructure:"key_file" yaml:"key_file" json:"key_file"`
	CAFile             string `mapstructure:"ca_file" yaml:"ca_file" json:"ca_file"`
	InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify" yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
}

// KafkaSASLConfig defines Kafka SASL authentication
type KafkaSASLConfig struct {
	Mechanism string `mapstructure:"mechanism" yaml:"mechanism" json:"mechanism"` // PLAIN, SCRAM-SHA-256, SCRAM-SHA-512
	Username  string `mapstructure:"username" yaml:"username" json:"username"`
	Password  string `mapstructure:"password" yaml:"password" json:"password"`
}

// KafkaProducerConfig defines Kafka producer settings
type KafkaProducerConfig struct {
	RequiredAcks    int           `mapstructure:"required_acks" yaml:"required_acks" json:"required_acks"`
	Timeout         time.Duration `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
	Compression     string        `mapstructure:"compression" yaml:"compression" json:"compression"` // none, gzip, snappy, lz4, zstd
	MaxMessageBytes int           `mapstructure:"max_message_bytes" yaml:"max_message_bytes" json:"max_message_bytes"`
	RetryMax        int           `mapstructure:"retry_max" yaml:"retry_max" json:"retry_max"`
	RetryBackoff    time.Duration `mapstructure:"retry_backoff" yaml:"retry_backoff" json:"retry_backoff"`
	FlushFrequency  time.Duration `mapstructure:"flush_frequency" yaml:"flush_frequency" json:"flush_frequency"`
	FlushMessages   int           `mapstructure:"flush_messages" yaml:"flush_messages" json:"flush_messages"`
	FlushBytes      int           `mapstructure:"flush_bytes" yaml:"flush_bytes" json:"flush_bytes"`
	Idempotent      bool          `mapstructure:"idempotent" yaml:"idempotent" json:"idempotent"`
	TransactionID   string        `mapstructure:"transaction_id" yaml:"transaction_id" json:"transaction_id"`
}

// KafkaConsumerConfig defines Kafka consumer settings
type KafkaConsumerConfig struct {
	GroupID            string        `mapstructure:"group_id" yaml:"group_id" json:"group_id"`
	AutoOffsetReset    string        `mapstructure:"auto_offset_reset" yaml:"auto_offset_reset" json:"auto_offset_reset"` // earliest, latest
	EnableAutoCommit   bool          `mapstructure:"enable_auto_commit" yaml:"enable_auto_commit" json:"enable_auto_commit"`
	AutoCommitInterval time.Duration `mapstructure:"auto_commit_interval" yaml:"auto_commit_interval" json:"auto_commit_interval"`
	SessionTimeout     time.Duration `mapstructure:"session_timeout" yaml:"session_timeout" json:"session_timeout"`
	HeartbeatInterval  time.Duration `mapstructure:"heartbeat_interval" yaml:"heartbeat_interval" json:"heartbeat_interval"`
	MaxProcessingTime  time.Duration `mapstructure:"max_processing_time" yaml:"max_processing_time" json:"max_processing_time"`
	FetchMin           int32         `mapstructure:"fetch_min" yaml:"fetch_min" json:"fetch_min"`
	FetchDefault       int32         `mapstructure:"fetch_default" yaml:"fetch_default" json:"fetch_default"`
	FetchMax           int32         `mapstructure:"fetch_max" yaml:"fetch_max" json:"fetch_max"`
	MaxWaitTime        time.Duration `mapstructure:"max_wait_time" yaml:"max_wait_time" json:"max_wait_time"`
	ChannelBufferSize  int           `mapstructure:"channel_buffer_size" yaml:"channel_buffer_size" json:"channel_buffer_size"`
	IsolationLevel     string        `mapstructure:"isolation_level" yaml:"isolation_level" json:"isolation_level"` // ReadUncommitted, ReadCommitted
	ReturnErrors       bool          `mapstructure:"return_errors" yaml:"return_errors" json:"return_errors"`
	OffsetsInitial     int64         `mapstructure:"offsets_initial" yaml:"offsets_initial" json:"offsets_initial"`
	OffsetsRetention   time.Duration `mapstructure:"offsets_retention" yaml:"offsets_retention" json:"offsets_retention"`
}

// KafkaAdminConfig defines Kafka admin client settings
type KafkaAdminConfig struct {
	Timeout time.Duration `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
}

// SchemaRegistryConfig defines Confluent Schema Registry configuration
type SchemaRegistryConfig struct {
	Enabled bool     `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	URLs    []string `mapstructure:"urls" yaml:"urls" json:"urls"`
	Auth    struct {
		Username string `mapstructure:"username" yaml:"username" json:"username"`
		Password string `mapstructure:"password" yaml:"password" json:"password"`
	} `mapstructure:"auth" yaml:"auth" json:"auth"`
}

// DebeziumConfig defines Debezium Change Data Capture configuration
type DebeziumConfig struct {
	Enabled bool `mapstructure:"enabled" yaml:"enabled" json:"enabled"`

	// Debezium Connect configuration
	Connect DebeziumConnectConfig `mapstructure:"connect" yaml:"connect" json:"connect"`

	// Database connectors configuration
	Connectors []DebeziumConnectorConfig `mapstructure:"connectors" yaml:"connectors" json:"connectors"`

	// Monitoring and health configuration
	Monitoring DebeziumMonitoringConfig `mapstructure:"monitoring" yaml:"monitoring" json:"monitoring"`
}

// DebeziumConnectConfig defines Kafka Connect configuration for Debezium
type DebeziumConnectConfig struct {
	URL       string        `mapstructure:"url" yaml:"url" json:"url"`
	Timeout   time.Duration `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
	Username  string        `mapstructure:"username" yaml:"username" json:"username"`
	Password  string        `mapstructure:"password" yaml:"password" json:"password"`
	TLSConfig TLSConfig     `mapstructure:"tls" yaml:"tls" json:"tls"`
}

// DebeziumConnectorConfig defines individual connector configuration
type DebeziumConnectorConfig struct {
	Name     string            `mapstructure:"name" yaml:"name" json:"name"`
	Type     string            `mapstructure:"type" yaml:"type" json:"type"` // postgres, mysql, mongodb, etc.
	Config   map[string]string `mapstructure:"config" yaml:"config" json:"config"`
	Database DatabaseConfig    `mapstructure:"database" yaml:"database" json:"database"`
	Topics   TopicConfig       `mapstructure:"topics" yaml:"topics" json:"topics"`
}

// TopicConfig defines topic naming and configuration
type TopicConfig struct {
	Prefix          string `mapstructure:"prefix" yaml:"prefix" json:"prefix"`
	HeartbeatName   string `mapstructure:"heartbeat_name" yaml:"heartbeat_name" json:"heartbeat_name"`
	TransactionName string `mapstructure:"transaction_name" yaml:"transaction_name" json:"transaction_name"`
}

// DebeziumMonitoringConfig defines monitoring configuration for Debezium
type DebeziumMonitoringConfig struct {
	Enabled        bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	HealthInterval time.Duration `mapstructure:"health_interval" yaml:"health_interval" json:"health_interval"`
	MetricsEnabled bool          `mapstructure:"metrics_enabled" yaml:"metrics_enabled" json:"metrics_enabled"`
}

// DatabasesConfig defines multiple database connections
type DatabasesConfig struct {
	// Primary databases for each microservice
	AuthDB     DatabaseConfig `mapstructure:"auth_db" yaml:"auth_db" json:"auth_db"`
	FormDB     DatabaseConfig `mapstructure:"form_db" yaml:"form_db" json:"form_db"`
	ResponseDB DatabaseConfig `mapstructure:"response_db" yaml:"response_db" json:"response_db"`

	// Event store database for event sourcing
	EventStore DatabaseConfig `mapstructure:"event_store" yaml:"event_store" json:"event_store"`

	// Default database for service-specific data
	Default DatabaseConfig `mapstructure:"default" yaml:"default" json:"default"`
}

// DatabaseConfig defines individual database configuration
type DatabaseConfig struct {
	Type                string        `mapstructure:"type" yaml:"type" json:"type"` // postgres, mysql, mongodb
	Host                string        `mapstructure:"host" yaml:"host" json:"host"`
	Port                int           `mapstructure:"port" yaml:"port" json:"port"`
	Name                string        `mapstructure:"name" yaml:"name" json:"name"`
	Username            string        `mapstructure:"username" yaml:"username" json:"username"`
	Password            string        `mapstructure:"password" yaml:"password" json:"password"`
	SSLMode             string        `mapstructure:"ssl_mode" yaml:"ssl_mode" json:"ssl_mode"`
	ConnectTimeout      time.Duration `mapstructure:"connect_timeout" yaml:"connect_timeout" json:"connect_timeout"`
	MaxOpenConns        int           `mapstructure:"max_open_conns" yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns        int           `mapstructure:"max_idle_conns" yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime     time.Duration `mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
	ConnMaxIdleTime     time.Duration `mapstructure:"conn_max_idle_time" yaml:"conn_max_idle_time" json:"conn_max_idle_time"`
	EnableWAL           bool          `mapstructure:"enable_wal" yaml:"enable_wal" json:"enable_wal"`
	WALLevel            string        `mapstructure:"wal_level" yaml:"wal_level" json:"wal_level"` // replica, logical
	ReplicationSlotName string        `mapstructure:"replication_slot_name" yaml:"replication_slot_name" json:"replication_slot_name"`
}

// RedisConfig defines Redis configuration for caching and pub/sub
type RedisConfig struct {
	Enabled      bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Host         string        `mapstructure:"host" yaml:"host" json:"host"`
	Port         int           `mapstructure:"port" yaml:"port" json:"port"`
	Password     string        `mapstructure:"password" yaml:"password" json:"password"`
	DB           int           `mapstructure:"db" yaml:"db" json:"db"`
	PoolSize     int           `mapstructure:"pool_size" yaml:"pool_size" json:"pool_size"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout" yaml:"dial_timeout" json:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" yaml:"write_timeout" json:"write_timeout"`
	TLS          TLSConfig     `mapstructure:"tls" yaml:"tls" json:"tls"`
}

// SecurityConfig defines security and authentication configuration
type SecurityConfig struct {
	// JWT configuration for API authentication
	JWT JWTConfig `mapstructure:"jwt" yaml:"jwt" json:"jwt"`

	// API key configuration for service-to-service communication
	APIKeys APIKeysConfig `mapstructure:"api_keys" yaml:"api_keys" json:"api_keys"`

	// Event signing configuration for message integrity
	EventSigning EventSigningConfig `mapstructure:"event_signing" yaml:"event_signing" json:"event_signing"`
}

// JWTConfig defines JWT authentication configuration
type JWTConfig struct {
	Secret    string        `mapstructure:"secret" yaml:"secret" json:"secret"`
	Issuer    string        `mapstructure:"issuer" yaml:"issuer" json:"issuer"`
	ExpiresIn time.Duration `mapstructure:"expires_in" yaml:"expires_in" json:"expires_in"`
}

// APIKeysConfig defines API key authentication configuration
type APIKeysConfig struct {
	Enabled bool              `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Keys    map[string]string `mapstructure:"keys" yaml:"keys" json:"keys"` // service_name -> api_key
}

// EventSigningConfig defines event message signing configuration
type EventSigningConfig struct {
	Enabled   bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Algorithm string `mapstructure:"algorithm" yaml:"algorithm" json:"algorithm"` // HMAC-SHA256, RSA, etc.
	SecretKey string `mapstructure:"secret_key" yaml:"secret_key" json:"secret_key"`
	PublicKey string `mapstructure:"public_key" yaml:"public_key" json:"public_key"`
}

// ObservabilityConfig defines monitoring, logging, and tracing configuration
type ObservabilityConfig struct {
	// Metrics configuration
	Metrics MetricsConfig `mapstructure:"metrics" yaml:"metrics" json:"metrics"`

	// Logging configuration
	Logging LoggingConfig `mapstructure:"logging" yaml:"logging" json:"logging"`

	// Tracing configuration
	Tracing TracingConfig `mapstructure:"tracing" yaml:"tracing" json:"tracing"`

	// Health check configuration
	Health HealthConfig `mapstructure:"health" yaml:"health" json:"health"`
}

// MetricsConfig defines Prometheus metrics configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Path    string `mapstructure:"path" yaml:"path" json:"path"`
	Port    string `mapstructure:"port" yaml:"port" json:"port"`
}

// LoggingConfig defines logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level" yaml:"level" json:"level"`    // debug, info, warn, error
	Format     string `mapstructure:"format" yaml:"format" json:"format"` // json, text
	Output     string `mapstructure:"output" yaml:"output" json:"output"` // stdout, file
	Filename   string `mapstructure:"filename" yaml:"filename" json:"filename"`
	MaxSize    int    `mapstructure:"max_size" yaml:"max_size" json:"max_size"` // MB
	MaxBackups int    `mapstructure:"max_backups" yaml:"max_backups" json:"max_backups"`
	MaxAge     int    `mapstructure:"max_age" yaml:"max_age" json:"max_age"` // days
	Compress   bool   `mapstructure:"compress" yaml:"compress" json:"compress"`
}

// TracingConfig defines distributed tracing configuration
type TracingConfig struct {
	Enabled     bool    `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	ServiceName string  `mapstructure:"service_name" yaml:"service_name" json:"service_name"`
	Endpoint    string  `mapstructure:"endpoint" yaml:"endpoint" json:"endpoint"`
	SampleRate  float64 `mapstructure:"sample_rate" yaml:"sample_rate" json:"sample_rate"`
}

// HealthConfig defines health check configuration
type HealthConfig struct {
	CheckInterval time.Duration `mapstructure:"check_interval" yaml:"check_interval" json:"check_interval"`
	Timeout       time.Duration `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
}

// EventProcessingConfig defines event processing behavior
type EventProcessingConfig struct {
	// Worker configuration
	Workers        int           `mapstructure:"workers" yaml:"workers" json:"workers"`
	BatchSize      int           `mapstructure:"batch_size" yaml:"batch_size" json:"batch_size"`
	ProcessTimeout time.Duration `mapstructure:"process_timeout" yaml:"process_timeout" json:"process_timeout"`
	RetryAttempts  int           `mapstructure:"retry_attempts" yaml:"retry_attempts" json:"retry_attempts"`
	RetryBackoff   time.Duration `mapstructure:"retry_backoff" yaml:"retry_backoff" json:"retry_backoff"`

	// Dead letter queue configuration
	DeadLetterQueue DeadLetterConfig `mapstructure:"dead_letter_queue" yaml:"dead_letter_queue" json:"dead_letter_queue"`

	// Event deduplication configuration
	Deduplication DeduplicationConfig `mapstructure:"deduplication" yaml:"deduplication" json:"deduplication"`

	// Event ordering configuration
	Ordering OrderingConfig `mapstructure:"ordering" yaml:"ordering" json:"ordering"`
}

// DeadLetterConfig defines dead letter queue configuration
type DeadLetterConfig struct {
	Enabled   bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	TopicName string        `mapstructure:"topic_name" yaml:"topic_name" json:"topic_name"`
	TTL       time.Duration `mapstructure:"ttl" yaml:"ttl" json:"ttl"`
}

// DeduplicationConfig defines event deduplication configuration
type DeduplicationConfig struct {
	Enabled bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Window  time.Duration `mapstructure:"window" yaml:"window" json:"window"`
	Storage string        `mapstructure:"storage" yaml:"storage" json:"storage"` // memory, redis
}

// OrderingConfig defines event ordering configuration
type OrderingConfig struct {
	Enabled      bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	BufferSize   int           `mapstructure:"buffer_size" yaml:"buffer_size" json:"buffer_size"`
	MaxWaitTime  time.Duration `mapstructure:"max_wait_time" yaml:"max_wait_time" json:"max_wait_time"`
	PartitionKey string        `mapstructure:"partition_key" yaml:"partition_key" json:"partition_key"`
}

// ServicesConfig defines microservice integration configuration
type ServicesConfig struct {
	AuthService          ServiceConfig `mapstructure:"auth_service" yaml:"auth_service" json:"auth_service"`
	FormService          ServiceConfig `mapstructure:"form_service" yaml:"form_service" json:"form_service"`
	ResponseService      ServiceConfig `mapstructure:"response_service" yaml:"response_service" json:"response_service"`
	AnalyticsService     ServiceConfig `mapstructure:"analytics_service" yaml:"analytics_service" json:"analytics_service"`
	CollaborationService ServiceConfig `mapstructure:"collaboration_service" yaml:"collaboration_service" json:"collaboration_service"`
	RealtimeService      ServiceConfig `mapstructure:"realtime_service" yaml:"realtime_service" json:"realtime_service"`
	FileUploadService    ServiceConfig `mapstructure:"file_upload_service" yaml:"file_upload_service" json:"file_upload_service"`
	APIGateway           ServiceConfig `mapstructure:"api_gateway" yaml:"api_gateway" json:"api_gateway"`
}

// ServiceConfig defines individual service configuration
type ServiceConfig struct {
	URL            string               `mapstructure:"url" yaml:"url" json:"url"`
	Timeout        time.Duration        `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
	HealthPath     string               `mapstructure:"health_path" yaml:"health_path" json:"health_path"`
	APIKey         string               `mapstructure:"api_key" yaml:"api_key" json:"api_key"`
	RetryCount     int                  `mapstructure:"retry_count" yaml:"retry_count" json:"retry_count"`
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuit_breaker" yaml:"circuit_breaker" json:"circuit_breaker"`
}

// CircuitBreakerConfig defines circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled              bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	FailureThreshold     int           `mapstructure:"failure_threshold" yaml:"failure_threshold" json:"failure_threshold"`
	RecoveryTimeout      time.Duration `mapstructure:"recovery_timeout" yaml:"recovery_timeout" json:"recovery_timeout"`
	ExpectedRecoveryTime time.Duration `mapstructure:"expected_recovery_time" yaml:"expected_recovery_time" json:"expected_recovery_time"`
}

// RateLimitingConfig defines rate limiting configuration
type RateLimitingConfig struct {
	Enabled           bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	RequestsPerSecond int           `mapstructure:"requests_per_second" yaml:"requests_per_second" json:"requests_per_second"`
	BurstSize         int           `mapstructure:"burst_size" yaml:"burst_size" json:"burst_size"`
	WindowSize        time.Duration `mapstructure:"window_size" yaml:"window_size" json:"window_size"`
	Storage           string        `mapstructure:"storage" yaml:"storage" json:"storage"` // memory, redis
}

// Load loads configuration from multiple sources with the following precedence:
// 1. Environment variables (highest priority)
// 2. Configuration file
// 3. Default values (lowest priority)
//
// This function implements enterprise-grade configuration loading with
// comprehensive error handling and validation.
func Load() *Config {
	cfg := &Config{}

	// Set up Viper configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/event-bus-service")

	// Enable environment variable overrides
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Load configuration file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			zap.L().Error("Error reading config file", zap.Error(err))
		}
	}

	// Set default values
	setDefaults()

	// Unmarshal configuration into struct
	if err := viper.Unmarshal(cfg); err != nil {
		zap.L().Fatal("Unable to decode configuration", zap.Error(err))
	}

	// Apply environment variable overrides
	applyEnvironmentOverrides(cfg)

	// Validate configuration
	if err := validateConfig(cfg); err != nil {
		zap.L().Fatal("Configuration validation failed", zap.Error(err))
	}

	return cfg
}

// setDefaults sets default values for all configuration options
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", "8090")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "60s")

	// Environment defaults
	viper.SetDefault("environment", "development")
	viper.SetDefault("version", "1.0.0")

	// Kafka defaults
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.client_id", "event-bus-service")
	viper.SetDefault("kafka.version", "2.8.0")
	viper.SetDefault("kafka.producer.required_acks", 1)
	viper.SetDefault("kafka.producer.timeout", "30s")
	viper.SetDefault("kafka.producer.compression", "snappy")
	viper.SetDefault("kafka.producer.max_message_bytes", 1000000)
	viper.SetDefault("kafka.producer.retry_max", 3)
	viper.SetDefault("kafka.producer.retry_backoff", "100ms")
	viper.SetDefault("kafka.producer.flush_frequency", "5s")
	viper.SetDefault("kafka.producer.flush_messages", 100)
	viper.SetDefault("kafka.producer.idempotent", true)
	viper.SetDefault("kafka.consumer.group_id", "event-bus-service-group")
	viper.SetDefault("kafka.consumer.auto_offset_reset", "earliest")
	viper.SetDefault("kafka.consumer.enable_auto_commit", true)
	viper.SetDefault("kafka.consumer.auto_commit_interval", "1s")
	viper.SetDefault("kafka.consumer.session_timeout", "30s")
	viper.SetDefault("kafka.consumer.heartbeat_interval", "3s")
	viper.SetDefault("kafka.consumer.max_processing_time", "5m")
	viper.SetDefault("kafka.consumer.fetch_min", 1)
	viper.SetDefault("kafka.consumer.fetch_default", 1024*1024)
	viper.SetDefault("kafka.consumer.fetch_max", 50*1024*1024)
	viper.SetDefault("kafka.consumer.max_wait_time", "250ms")
	viper.SetDefault("kafka.consumer.channel_buffer_size", 256)
	viper.SetDefault("kafka.consumer.return_errors", true)

	// Debezium defaults
	viper.SetDefault("debezium.enabled", false)
	viper.SetDefault("debezium.connect.url", "http://localhost:8083")
	viper.SetDefault("debezium.connect.timeout", "30s")

	// Database defaults
	viper.SetDefault("databases.default.type", "postgres")
	viper.SetDefault("databases.default.host", "localhost")
	viper.SetDefault("databases.default.port", 5432)
	viper.SetDefault("databases.default.ssl_mode", "disable")
	viper.SetDefault("databases.default.connect_timeout", "30s")
	viper.SetDefault("databases.default.max_open_conns", 25)
	viper.SetDefault("databases.default.max_idle_conns", 5)
	viper.SetDefault("databases.default.conn_max_lifetime", "1h")
	viper.SetDefault("databases.default.conn_max_idle_time", "30m")

	// Redis defaults
	viper.SetDefault("redis.enabled", false)
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.dial_timeout", "5s")
	viper.SetDefault("redis.read_timeout", "3s")
	viper.SetDefault("redis.write_timeout", "3s")

	// Security defaults
	viper.SetDefault("security.jwt.issuer", "event-bus-service")
	viper.SetDefault("security.jwt.expires_in", "24h")
	viper.SetDefault("security.api_keys.enabled", false)
	viper.SetDefault("security.event_signing.enabled", false)
	viper.SetDefault("security.event_signing.algorithm", "HMAC-SHA256")

	// Observability defaults
	viper.SetDefault("observability.metrics.enabled", true)
	viper.SetDefault("observability.metrics.path", "/metrics")
	viper.SetDefault("observability.metrics.port", "9090")
	viper.SetDefault("observability.logging.level", "info")
	viper.SetDefault("observability.logging.format", "json")
	viper.SetDefault("observability.logging.output", "stdout")
	viper.SetDefault("observability.tracing.enabled", false)
	viper.SetDefault("observability.tracing.service_name", "event-bus-service")
	viper.SetDefault("observability.tracing.sample_rate", 0.1)
	viper.SetDefault("observability.health.check_interval", "30s")
	viper.SetDefault("observability.health.timeout", "10s")

	// Event processing defaults
	viper.SetDefault("event_processing.workers", 5)
	viper.SetDefault("event_processing.batch_size", 100)
	viper.SetDefault("event_processing.process_timeout", "30s")
	viper.SetDefault("event_processing.retry_attempts", 3)
	viper.SetDefault("event_processing.retry_backoff", "1s")
	viper.SetDefault("event_processing.dead_letter_queue.enabled", true)
	viper.SetDefault("event_processing.dead_letter_queue.topic_name", "dead-letter-queue")
	viper.SetDefault("event_processing.dead_letter_queue.ttl", "7d")
	viper.SetDefault("event_processing.deduplication.enabled", true)
	viper.SetDefault("event_processing.deduplication.window", "5m")
	viper.SetDefault("event_processing.deduplication.storage", "memory")
	viper.SetDefault("event_processing.ordering.enabled", false)
	viper.SetDefault("event_processing.ordering.buffer_size", 1000)
	viper.SetDefault("event_processing.ordering.max_wait_time", "1s")

	// Rate limiting defaults
	viper.SetDefault("rate_limiting.enabled", true)
	viper.SetDefault("rate_limiting.requests_per_second", 100)
	viper.SetDefault("rate_limiting.burst_size", 10)
	viper.SetDefault("rate_limiting.window_size", "1m")
	viper.SetDefault("rate_limiting.storage", "memory")

	// Service defaults
	serviceDefaults := map[string]interface{}{
		"timeout":                                "30s",
		"health_path":                            "/health",
		"retry_count":                            3,
		"circuit_breaker.enabled":                true,
		"circuit_breaker.failure_threshold":      5,
		"circuit_breaker.recovery_timeout":       "30s",
		"circuit_breaker.expected_recovery_time": "10s",
	}

	services := []string{
		"auth_service", "form_service", "response_service", "analytics_service",
		"collaboration_service", "realtime_service", "file_upload_service", "api_gateway",
	}

	for _, service := range services {
		for key, value := range serviceDefaults {
			viper.SetDefault(fmt.Sprintf("services.%s.%s", service, key), value)
		}
	}

	// Service-specific URL defaults
	viper.SetDefault("services.auth_service.url", "http://localhost:3001")
	viper.SetDefault("services.form_service.url", "http://localhost:3002")
	viper.SetDefault("services.response_service.url", "http://localhost:3003")
	viper.SetDefault("services.analytics_service.url", "http://localhost:3004")
	viper.SetDefault("services.collaboration_service.url", "http://localhost:3005")
	viper.SetDefault("services.realtime_service.url", "http://localhost:3006")
	viper.SetDefault("services.file_upload_service.url", "http://localhost:3007")
	viper.SetDefault("services.api_gateway.url", "http://localhost:8080")
}

// applyEnvironmentOverrides applies environment variable overrides to configuration
func applyEnvironmentOverrides(cfg *Config) {
	// Server configuration overrides
	if port := os.Getenv("SERVER_PORT"); port != "" {
		cfg.Server.Port = port
	}
	if host := os.Getenv("SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}

	// Environment override
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		cfg.Environment = env
	}
	if version := os.Getenv("VERSION"); version != "" {
		cfg.Version = version
	}

	// Kafka configuration overrides
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		cfg.Kafka.Brokers = strings.Split(brokers, ",")
	}
	if clientID := os.Getenv("KAFKA_CLIENT_ID"); clientID != "" {
		cfg.Kafka.ClientID = clientID
	}
	if groupID := os.Getenv("KAFKA_CONSUMER_GROUP_ID"); groupID != "" {
		cfg.Kafka.Consumer.GroupID = groupID
	}

	// Security overrides
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.Security.JWT.Secret = secret
	}

	// Database overrides
	applyDatabaseOverrides(&cfg.Databases.Default, "DATABASE")
	applyDatabaseOverrides(&cfg.Databases.AuthDB, "AUTH_DATABASE")
	applyDatabaseOverrides(&cfg.Databases.FormDB, "FORM_DATABASE")
	applyDatabaseOverrides(&cfg.Databases.ResponseDB, "RESPONSE_DATABASE")
	applyDatabaseOverrides(&cfg.Databases.EventStore, "EVENTSTORE_DATABASE")

	// Redis overrides
	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		cfg.Redis.Host = redisHost
	}
	if redisPort := os.Getenv("REDIS_PORT"); redisPort != "" {
		if port, err := strconv.Atoi(redisPort); err == nil {
			cfg.Redis.Port = port
		}
	}
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		cfg.Redis.Password = redisPassword
	}

	// Debezium overrides
	if debeziumURL := os.Getenv("DEBEZIUM_CONNECT_URL"); debeziumURL != "" {
		cfg.Debezium.Connect.URL = debeziumURL
	}
	if debeziumEnabled := os.Getenv("DEBEZIUM_ENABLED"); debeziumEnabled != "" {
		cfg.Debezium.Enabled = debeziumEnabled == "true"
	}
}

// applyDatabaseOverrides applies database-specific environment overrides
func applyDatabaseOverrides(dbConfig *DatabaseConfig, prefix string) {
	if host := os.Getenv(prefix + "_HOST"); host != "" {
		dbConfig.Host = host
	}
	if port := os.Getenv(prefix + "_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			dbConfig.Port = p
		}
	}
	if name := os.Getenv(prefix + "_NAME"); name != "" {
		dbConfig.Name = name
	}
	if username := os.Getenv(prefix + "_USERNAME"); username != "" {
		dbConfig.Username = username
	}
	if password := os.Getenv(prefix + "_PASSWORD"); password != "" {
		dbConfig.Password = password
	}
	if sslMode := os.Getenv(prefix + "_SSL_MODE"); sslMode != "" {
		dbConfig.SSLMode = sslMode
	}

	// Handle full DATABASE_URL format
	if url := os.Getenv(prefix + "_URL"); url != "" {
		// Parse PostgreSQL URL format: postgres://user:pass@host:port/dbname
		if strings.HasPrefix(url, "postgres://") || strings.HasPrefix(url, "postgresql://") {
			// This is a simplified parser - in production, use a proper URL parser
			parts := strings.Split(strings.TrimPrefix(strings.TrimPrefix(url, "postgres://"), "postgresql://"), "@")
			if len(parts) == 2 {
				// Extract user:pass
				userPass := strings.Split(parts[0], ":")
				if len(userPass) == 2 {
					dbConfig.Username = userPass[0]
					dbConfig.Password = userPass[1]
				}

				// Extract host:port/dbname
				hostParts := strings.Split(parts[1], "/")
				if len(hostParts) == 2 {
					hostPort := strings.Split(hostParts[0], ":")
					dbConfig.Host = hostPort[0]
					if len(hostPort) == 2 {
						if p, err := strconv.Atoi(hostPort[1]); err == nil {
							dbConfig.Port = p
						}
					}
					// Extract database name and remove query parameters
					dbName := strings.Split(hostParts[1], "?")[0]
					dbConfig.Name = dbName
				}
			}
		}
	}
}

// validateConfig validates the loaded configuration
func validateConfig(cfg *Config) error {
	// Validate required fields
	if cfg.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	if len(cfg.Kafka.Brokers) == 0 {
		return fmt.Errorf("kafka brokers are required")
	}

	if cfg.Kafka.Consumer.GroupID == "" {
		return fmt.Errorf("kafka consumer group ID is required")
	}

	if cfg.Security.JWT.Secret == "" && cfg.Environment == "production" {
		return fmt.Errorf("JWT secret is required in production environment")
	}

	// Validate event processing configuration
	if cfg.EventProcessing.Workers < 1 {
		return fmt.Errorf("event processing workers must be at least 1")
	}

	if cfg.EventProcessing.BatchSize < 1 {
		return fmt.Errorf("event processing batch size must be at least 1")
	}

	// Validate database configurations
	if err := validateDatabaseConfig(&cfg.Databases.Default, "default database"); err != nil {
		return err
	}

	return nil
}

// validateDatabaseConfig validates individual database configuration
func validateDatabaseConfig(dbConfig *DatabaseConfig, name string) error {
	if dbConfig.Host == "" {
		return fmt.Errorf("%s host is required", name)
	}

	if dbConfig.Port <= 0 {
		return fmt.Errorf("%s port must be greater than 0", name)
	}

	if dbConfig.Name == "" {
		return fmt.Errorf("%s name is required", name)
	}

	if dbConfig.Username == "" {
		return fmt.Errorf("%s username is required", name)
	}

	return nil
}

// GetAddress returns the full server address
func (s *ServerConfig) GetAddress() string {
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}

// GetConnectionString returns the database connection string
func (d *DatabaseConfig) GetConnectionString() string {
	switch d.Type {
	case "postgres", "postgresql":
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			d.Username, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			d.Username, d.Password, d.Host, d.Port, d.Name)
	default:
		return ""
	}
}

// GetRedisAddress returns the Redis connection address
func (r *RedisConfig) GetRedisAddress() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// GetKafkaBrokerAddresses returns Kafka broker addresses as a slice
func (k *KafkaConfig) GetKafkaBrokerAddresses() []string {
	return k.Brokers
}
