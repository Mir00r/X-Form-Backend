package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for the collaboration service
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Auth      AuthConfig      `mapstructure:"auth"`
	WebSocket WebSocketConfig `mapstructure:"websocket"`
	Kafka     KafkaConfig     `mapstructure:"kafka"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Metrics   MetricsConfig   `mapstructure:"metrics"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port            string        `mapstructure:"port"`
	Host            string        `mapstructure:"host"`
	Environment     string        `mapstructure:"environment"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	TLSEnabled      bool          `mapstructure:"tls_enabled"`
	CertFile        string        `mapstructure:"cert_file"`
	KeyFile         string        `mapstructure:"key_file"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         string        `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	MaxRetries   int           `mapstructure:"max_retries"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret           string        `mapstructure:"jwt_secret"`
	JWTExpiration       time.Duration `mapstructure:"jwt_expiration"`
	ServiceSecret       string        `mapstructure:"service_secret"`
	TokenValidationURL  string        `mapstructure:"token_validation_url"`
	PermissionCacheTime time.Duration `mapstructure:"permission_cache_time"`
}

// WebSocketConfig holds WebSocket configuration
type WebSocketConfig struct {
	MaxConnections    int           `mapstructure:"max_connections"`
	MaxMessageSize    int64         `mapstructure:"max_message_size"`
	WriteWait         time.Duration `mapstructure:"write_wait"`
	PongWait          time.Duration `mapstructure:"pong_wait"`
	PingPeriod        time.Duration `mapstructure:"ping_period"`
	ReadBufferSize    int           `mapstructure:"read_buffer_size"`
	WriteBufferSize   int           `mapstructure:"write_buffer_size"`
	EnableCompression bool          `mapstructure:"enable_compression"`
	CheckOrigin       bool          `mapstructure:"check_origin"`
	AllowedOrigins    []string      `mapstructure:"allowed_origins"`
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout"`
	MaxRoomsPerUser   int           `mapstructure:"max_rooms_per_user"`
	MaxUsersPerRoom   int           `mapstructure:"max_users_per_room"`
	MessageRateLimit  int           `mapstructure:"message_rate_limit"`
	RateLimitWindow   time.Duration `mapstructure:"rate_limit_window"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers        []string       `mapstructure:"brokers"`
	ConsumerGroup  string         `mapstructure:"consumer_group"`
	Topics         TopicsConfig   `mapstructure:"topics"`
	ProducerConfig ProducerConfig `mapstructure:"producer"`
	ConsumerConfig ConsumerConfig `mapstructure:"consumer"`
	RetryPolicy    RetryPolicy    `mapstructure:"retry_policy"`
}

// TopicsConfig holds topic configurations
type TopicsConfig struct {
	FormEvents          string `mapstructure:"form_events"`
	CollaborationEvents string `mapstructure:"collaboration_events"`
	UserEvents          string `mapstructure:"user_events"`
	SystemEvents        string `mapstructure:"system_events"`
}

// ProducerConfig holds Kafka producer configuration
type ProducerConfig struct {
	Timeout      time.Duration `mapstructure:"timeout"`
	RetryMax     int           `mapstructure:"retry_max"`
	BatchSize    int           `mapstructure:"batch_size"`
	BatchTimeout time.Duration `mapstructure:"batch_timeout"`
	RequiredAcks int           `mapstructure:"required_acks"`
	Compression  string        `mapstructure:"compression"`
}

// ConsumerConfig holds Kafka consumer configuration
type ConsumerConfig struct {
	SessionTimeout    time.Duration `mapstructure:"session_timeout"`
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
	MaxWait           time.Duration `mapstructure:"max_wait"`
	MinBytes          int           `mapstructure:"min_bytes"`
	MaxBytes          int           `mapstructure:"max_bytes"`
	StartOffset       int64         `mapstructure:"start_offset"`
}

// RetryPolicy holds retry configuration
type RetryPolicy struct {
	MaxRetries   int           `mapstructure:"max_retries"`
	BackoffDelay time.Duration `mapstructure:"backoff_delay"`
	MaxBackoff   time.Duration `mapstructure:"max_backoff"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Port      string `mapstructure:"port"`
	Path      string `mapstructure:"path"`
	Namespace string `mapstructure:"namespace"`
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Set defaults
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Override with environment variables
	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override with environment variables for sensitive data
	overrideWithEnv(&config)

	// Validate configuration
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8083")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "120s")
	viper.SetDefault("server.shutdown_timeout", "10s")
	viper.SetDefault("server.tls_enabled", false)

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 100)
	viper.SetDefault("redis.min_idle_conns", 10)
	viper.SetDefault("redis.max_retries", 3)
	viper.SetDefault("redis.dial_timeout", "5s")
	viper.SetDefault("redis.read_timeout", "3s")
	viper.SetDefault("redis.write_timeout", "3s")
	viper.SetDefault("redis.idle_timeout", "300s")

	// Auth defaults
	viper.SetDefault("auth.jwt_expiration", "24h")
	viper.SetDefault("auth.permission_cache_time", "5m")

	// WebSocket defaults
	viper.SetDefault("websocket.max_connections", 10000)
	viper.SetDefault("websocket.max_message_size", 1024)
	viper.SetDefault("websocket.write_wait", "10s")
	viper.SetDefault("websocket.pong_wait", "60s")
	viper.SetDefault("websocket.ping_period", "54s")
	viper.SetDefault("websocket.read_buffer_size", 1024)
	viper.SetDefault("websocket.write_buffer_size", 1024)
	viper.SetDefault("websocket.enable_compression", true)
	viper.SetDefault("websocket.check_origin", true)
	viper.SetDefault("websocket.heartbeat_interval", "30s")
	viper.SetDefault("websocket.connection_timeout", "60s")
	viper.SetDefault("websocket.max_rooms_per_user", 10)
	viper.SetDefault("websocket.max_users_per_room", 100)
	viper.SetDefault("websocket.message_rate_limit", 60)
	viper.SetDefault("websocket.rate_limit_window", "1m")

	// Kafka defaults
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.consumer_group", "collaboration-service")
	viper.SetDefault("kafka.topics.form_events", "form-events")
	viper.SetDefault("kafka.topics.collaboration_events", "collaboration-events")
	viper.SetDefault("kafka.topics.user_events", "user-events")
	viper.SetDefault("kafka.topics.system_events", "system-events")
	viper.SetDefault("kafka.producer.timeout", "10s")
	viper.SetDefault("kafka.producer.retry_max", 3)
	viper.SetDefault("kafka.producer.batch_size", 100)
	viper.SetDefault("kafka.producer.batch_timeout", "1s")
	viper.SetDefault("kafka.producer.required_acks", 1)
	viper.SetDefault("kafka.producer.compression", "snappy")
	viper.SetDefault("kafka.consumer.session_timeout", "30s")
	viper.SetDefault("kafka.consumer.heartbeat_interval", "3s")
	viper.SetDefault("kafka.consumer.max_wait", "1s")
	viper.SetDefault("kafka.consumer.min_bytes", 1)
	viper.SetDefault("kafka.consumer.max_bytes", 1048576)
	viper.SetDefault("kafka.consumer.start_offset", -1)
	viper.SetDefault("kafka.retry_policy.max_retries", 3)
	viper.SetDefault("kafka.retry_policy.backoff_delay", "1s")
	viper.SetDefault("kafka.retry_policy.max_backoff", "30s")

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")
	viper.SetDefault("logging.max_size", 100)
	viper.SetDefault("logging.max_backups", 3)
	viper.SetDefault("logging.max_age", 28)
	viper.SetDefault("logging.compress", true)

	// Metrics defaults
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.port", "9090")
	viper.SetDefault("metrics.path", "/metrics")
	viper.SetDefault("metrics.namespace", "collaboration_service")
}

// overrideWithEnv overrides configuration with environment variables
func overrideWithEnv(config *Config) {
	// Server
	if port := os.Getenv("PORT"); port != "" {
		config.Server.Port = port
	}
	if host := os.Getenv("HOST"); host != "" {
		config.Server.Host = host
	}
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		config.Server.Environment = env
	}

	// Redis
	if host := os.Getenv("REDIS_HOST"); host != "" {
		config.Redis.Host = host
	}
	if port := os.Getenv("REDIS_PORT"); port != "" {
		config.Redis.Port = port
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		config.Redis.Password = password
	}
	if db := os.Getenv("REDIS_DB"); db != "" {
		if dbInt, err := strconv.Atoi(db); err == nil {
			config.Redis.DB = dbInt
		}
	}

	// Auth
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		config.Auth.JWTSecret = secret
	}
	if serviceSecret := os.Getenv("SERVICE_SECRET"); serviceSecret != "" {
		config.Auth.ServiceSecret = serviceSecret
	}

	// Kafka
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		// Simple parsing - in production you might want more sophisticated parsing
		config.Kafka.Brokers = []string{brokers}
	}

	// WebSocket
	if maxConn := os.Getenv("WS_MAX_CONNECTIONS"); maxConn != "" {
		if maxConnInt, err := strconv.Atoi(maxConn); err == nil {
			config.WebSocket.MaxConnections = maxConnInt
		}
	}
}

// validate validates the configuration
func validate(config *Config) error {
	// Validate required fields
	if config.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if config.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	if len(config.Kafka.Brokers) == 0 {
		return fmt.Errorf("kafka brokers are required")
	}

	// Validate Redis connection
	if config.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}

	// Validate WebSocket configuration
	if config.WebSocket.MaxConnections <= 0 {
		return fmt.Errorf("websocket max_connections must be positive")
	}

	if config.WebSocket.MaxMessageSize <= 0 {
		return fmt.Errorf("websocket max_message_size must be positive")
	}

	// Validate timeouts
	if config.WebSocket.WriteWait <= 0 {
		return fmt.Errorf("websocket write_wait must be positive")
	}

	if config.WebSocket.PongWait <= 0 {
		return fmt.Errorf("websocket pong_wait must be positive")
	}

	return nil
}

// GetRedisAddr returns the Redis address
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}
