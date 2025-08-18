package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	MetricsPort string
	Environment string
	JWTSecret   string
	Version     string

	// Service URLs
	AuthServiceURL      string
	FormServiceURL      string
	ResponseServiceURL  string
	AnalyticsServiceURL string
	FileServiceURL      string
	RealtimeServiceURL  string

	// Redis for rate limiting and caching
	RedisURL string

	// API Management
	KongAdminURL string

	// Observability
	JaegerURL string
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
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

		// Infrastructure
		RedisURL: getEnv("REDIS_URL", "redis://redis:6379"),

		// API Management
		KongAdminURL: getEnv("KONG_ADMIN_URL", "http://kong:8001"),

		// Observability
		JaegerURL: getEnv("JAEGER_URL", "http://jaeger:14268/api/traces"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
