"""
Analytics Service Configuration
"""
import os
from typing import List, Optional
from pydantic import BaseSettings, Field
from pydantic_settings import BaseSettings as PydanticBaseSettings


class Settings(PydanticBaseSettings):
    """Application settings loaded from environment variables."""
    
    # Application
    app_name: str = Field(default="Analytics Service", env="APP_NAME")
    app_version: str = Field(default="1.0.0", env="APP_VERSION")
    environment: str = Field(default="development", env="ENVIRONMENT")
    debug: bool = Field(default=True, env="DEBUG")
    
    # Server
    host: str = Field(default="0.0.0.0", env="HOST")
    port: int = Field(default=8084, env="PORT")
    workers: int = Field(default=1, env="WORKERS")
    
    # Authentication
    jwt_secret_key: str = Field(env="JWT_SECRET_KEY")
    jwt_algorithm: str = Field(default="HS256", env="JWT_ALGORITHM")
    jwt_expiration_hours: int = Field(default=24, env="JWT_EXPIRATION_HOURS")
    
    # BigQuery Configuration
    google_application_credentials: Optional[str] = Field(
        default=None, env="GOOGLE_APPLICATION_CREDENTIALS"
    )
    bigquery_project_id: str = Field(env="BIGQUERY_PROJECT_ID")
    bigquery_dataset_id: str = Field(default="xform_analytics", env="BIGQUERY_DATASET_ID")
    bigquery_location: str = Field(default="US", env="BIGQUERY_LOCATION")
    
    # BigQuery Tables
    responses_table: str = Field(default="form_responses", env="RESPONSES_TABLE")
    forms_table: str = Field(default="forms", env="FORMS_TABLE")
    users_table: str = Field(default="users", env="USERS_TABLE")
    events_table: str = Field(default="events", env="EVENTS_TABLE")
    
    # Redis Configuration
    redis_host: str = Field(default="localhost", env="REDIS_HOST")
    redis_port: int = Field(default=6379, env="REDIS_PORT")
    redis_password: Optional[str] = Field(default=None, env="REDIS_PASSWORD")
    redis_db: int = Field(default=0, env="REDIS_DB")
    redis_ssl: bool = Field(default=False, env="REDIS_SSL")
    cache_ttl: int = Field(default=3600, env="CACHE_TTL")  # 1 hour
    
    # External Services
    form_service_url: str = Field(default="http://localhost:8081", env="FORM_SERVICE_URL")
    response_service_url: str = Field(default="http://localhost:8082", env="RESPONSE_SERVICE_URL")
    auth_service_url: str = Field(default="http://localhost:8080", env="AUTH_SERVICE_URL")
    
    # Analytics Configuration
    max_query_results: int = Field(default=10000, env="MAX_QUERY_RESULTS")
    query_timeout_seconds: int = Field(default=30, env="QUERY_TIMEOUT_SECONDS")
    enable_query_caching: bool = Field(default=True, env="ENABLE_QUERY_CACHING")
    
    # Rate Limiting
    rate_limit_requests: int = Field(default=100, env="RATE_LIMIT_REQUESTS")
    rate_limit_window: int = Field(default=60, env="RATE_LIMIT_WINDOW")  # seconds
    
    # CORS
    allowed_origins: List[str] = Field(
        default=["http://localhost:3000", "http://localhost:3001"],
        env="ALLOWED_ORIGINS"
    )
    allowed_methods: List[str] = Field(
        default=["GET", "POST", "PUT", "DELETE", "OPTIONS"],
        env="ALLOWED_METHODS"
    )
    allowed_headers: List[str] = Field(
        default=["*"],
        env="ALLOWED_HEADERS"
    )
    
    # Logging
    log_level: str = Field(default="INFO", env="LOG_LEVEL")
    log_format: str = Field(default="json", env="LOG_FORMAT")
    
    # Monitoring
    enable_metrics: bool = Field(default=True, env="ENABLE_METRICS")
    metrics_port: int = Field(default=9090, env="METRICS_PORT")
    
    # Batch Processing
    batch_size: int = Field(default=1000, env="BATCH_SIZE")
    batch_timeout_seconds: int = Field(default=300, env="BATCH_TIMEOUT_SECONDS")
    
    # Data Retention
    data_retention_days: int = Field(default=365, env="DATA_RETENTION_DAYS")
    cleanup_interval_hours: int = Field(default=24, env="CLEANUP_INTERVAL_HOURS")

    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"
        case_sensitive = False


# Global settings instance
settings = Settings()


def get_settings() -> Settings:
    """Get application settings."""
    return settings


def get_bigquery_table_name(table_name: str) -> str:
    """Get fully qualified BigQuery table name."""
    return f"{settings.bigquery_project_id}.{settings.bigquery_dataset_id}.{table_name}"


def get_redis_url() -> str:
    """Get Redis connection URL."""
    if settings.redis_password:
        return f"redis://:{settings.redis_password}@{settings.redis_host}:{settings.redis_port}/{settings.redis_db}"
    return f"redis://{settings.redis_host}:{settings.redis_port}/{settings.redis_db}"


# BigQuery table configurations
BIGQUERY_TABLES = {
    "responses": {
        "name": get_bigquery_table_name(settings.responses_table),
        "schema": [
            {"name": "response_id", "type": "STRING", "mode": "REQUIRED"},
            {"name": "form_id", "type": "STRING", "mode": "REQUIRED"},
            {"name": "user_id", "type": "STRING", "mode": "NULLABLE"},
            {"name": "submitted_at", "type": "TIMESTAMP", "mode": "REQUIRED"},
            {"name": "completion_time_seconds", "type": "INTEGER", "mode": "NULLABLE"},
            {"name": "responses", "type": "JSON", "mode": "REQUIRED"},
            {"name": "metadata", "type": "JSON", "mode": "NULLABLE"},
            {"name": "ip_address", "type": "STRING", "mode": "NULLABLE"},
            {"name": "user_agent", "type": "STRING", "mode": "NULLABLE"},
            {"name": "created_at", "type": "TIMESTAMP", "mode": "REQUIRED"},
            {"name": "updated_at", "type": "TIMESTAMP", "mode": "REQUIRED"},
        ]
    },
    "forms": {
        "name": get_bigquery_table_name(settings.forms_table),
        "schema": [
            {"name": "form_id", "type": "STRING", "mode": "REQUIRED"},
            {"name": "title", "type": "STRING", "mode": "REQUIRED"},
            {"name": "description", "type": "STRING", "mode": "NULLABLE"},
            {"name": "owner_id", "type": "STRING", "mode": "REQUIRED"},
            {"name": "questions", "type": "JSON", "mode": "REQUIRED"},
            {"name": "settings", "type": "JSON", "mode": "NULLABLE"},
            {"name": "status", "type": "STRING", "mode": "REQUIRED"},
            {"name": "created_at", "type": "TIMESTAMP", "mode": "REQUIRED"},
            {"name": "updated_at", "type": "TIMESTAMP", "mode": "REQUIRED"},
        ]
    },
    "events": {
        "name": get_bigquery_table_name(settings.events_table),
        "schema": [
            {"name": "event_id", "type": "STRING", "mode": "REQUIRED"},
            {"name": "event_type", "type": "STRING", "mode": "REQUIRED"},
            {"name": "form_id", "type": "STRING", "mode": "REQUIRED"},
            {"name": "user_id", "type": "STRING", "mode": "NULLABLE"},
            {"name": "event_data", "type": "JSON", "mode": "REQUIRED"},
            {"name": "timestamp", "type": "TIMESTAMP", "mode": "REQUIRED"},
            {"name": "session_id", "type": "STRING", "mode": "NULLABLE"},
            {"name": "ip_address", "type": "STRING", "mode": "NULLABLE"},
        ]
    }
}
