/**
 * Enhanced Configuration Management System
 * Centralizes all environment variables and configuration with validation
 */

const path = require('path');

class ConfigurationManager {
  constructor() {
    this.config = {};
    this.loadConfiguration();
    this.validateConfiguration();
  }

  /**
   * Load configuration from environment variables with defaults
   */
  loadConfiguration() {
    this.config = {
      // Server Configuration
      server: {
        port: this.getEnvNumber('RESPONSE_SERVICE_PORT', 3002),
        host: this.getEnvString('HOST', '0.0.0.0'),
        environment: this.getEnvString('NODE_ENV', 'development'),
        serviceName: this.getEnvString('SERVICE_NAME', 'response-service'),
        version: this.getEnvString('SERVICE_VERSION', '1.0.0')
      },

      // Firebase Configuration (Legacy - maintained for compatibility)
      firebase: {
        projectId: this.getEnvString('FIREBASE_PROJECT_ID', 'xform-backend-dev'),
        keyFile: this.getEnvString('FIREBASE_KEY_FILE', './firebase-key.json'),
        databaseURL: this.getEnvString('FIREBASE_DATABASE_URL')
      },

      // Database Configuration
      database: {
        type: this.getEnvString('DB_TYPE', 'mongodb'),
        host: this.getEnvString('DB_HOST', 'localhost'),
        port: this.getEnvNumber('DB_PORT', 27017),
        name: this.getEnvString('DB_NAME', 'response_service'),
        username: this.getEnvString('DB_USERNAME'),
        password: this.getEnvString('DB_PASSWORD'),
        ssl: this.getEnvBoolean('DB_SSL', false),
        connectionTimeout: this.getEnvNumber('DB_CONNECTION_TIMEOUT', 10000),
        maxConnections: this.getEnvNumber('DB_MAX_CONNECTIONS', 10),
        url: this.getEnvString('DATABASE_URL') // Alternative connection string
      },

      // JWT Configuration
      jwt: {
        secret: this.getEnvString('JWT_SECRET', 'your-jwt-secret-key'),
        expiresIn: this.getEnvString('JWT_EXPIRES_IN', '24h'),
        issuer: this.getEnvString('JWT_ISSUER', 'response-service'),
        audience: this.getEnvString('JWT_AUDIENCE', 'x-form-users')
      },

      // Kafka Configuration
      kafka: {
        enabled: this.getEnvBoolean('KAFKA_ENABLED', false),
        clientId: this.getEnvString('KAFKA_CLIENT_ID', 'response-service'),
        brokers: this.getEnvArray('KAFKA_BROKERS', ['localhost:9092']),
        topics: {
          formUpdated: this.getEnvString('KAFKA_TOPIC_FORM_UPDATED', 'form.updated'),
          formDeleted: this.getEnvString('KAFKA_TOPIC_FORM_DELETED', 'form.deleted'),
          responseSubmitted: this.getEnvString('KAFKA_TOPIC_RESPONSE_SUBMITTED', 'response.submitted')
        }
      },

      // Google Sheets Configuration
      googleSheets: {
        enabled: this.getEnvBoolean('GOOGLE_SHEETS_ENABLED', false),
        credentials: {
          type: this.getEnvString('GOOGLE_ACCOUNT_TYPE', 'service_account'),
          project_id: this.getEnvString('GOOGLE_PROJECT_ID'),
          private_key_id: this.getEnvString('GOOGLE_PRIVATE_KEY_ID'),
          private_key: this.getEnvString('GOOGLE_PRIVATE_KEY')?.replace(/\\n/g, '\n'),
          client_email: this.getEnvString('GOOGLE_CLIENT_EMAIL'),
          client_id: this.getEnvString('GOOGLE_CLIENT_ID'),
          auth_uri: 'https://accounts.google.com/o/oauth2/auth',
          token_uri: 'https://oauth2.googleapis.com/token',
          auth_provider_x509_cert_url: 'https://www.googleapis.com/oauth2/v1/certs',
          client_x509_cert_url: this.getEnvString('GOOGLE_CLIENT_CERT_URL')
        },
        scopes: [
          'https://www.googleapis.com/auth/spreadsheets',
          'https://www.googleapis.com/auth/drive.file'
        ]
      },

      // Form Service Configuration
      formService: {
        baseUrl: this.getEnvString('FORM_SERVICE_URL', 'http://localhost:8082'),
        timeout: this.getEnvNumber('FORM_SERVICE_TIMEOUT', 5000),
        apiKey: this.getEnvString('FORM_SERVICE_API_KEY'),
        retryAttempts: this.getEnvNumber('FORM_SERVICE_RETRY_ATTEMPTS', 3),
        retryDelay: this.getEnvNumber('FORM_SERVICE_RETRY_DELAY', 1000),
        apiVersion: this.getEnvString('FORM_SERVICE_API_VERSION', 'v1')
      },

      // Rate Limiting Configuration
      rateLimit: {
        windowMs: this.getEnvNumber('RATE_LIMIT_WINDOW_MS', 15 * 60 * 1000), // 15 minutes
        max: this.getEnvNumber('RATE_LIMIT_MAX', 100), // limit each IP to 100 requests per windowMs
        message: 'Too many requests from this IP, please try again later.',
        standardHeaders: this.getEnvBoolean('RATE_LIMIT_STANDARD_HEADERS', true),
        legacyHeaders: this.getEnvBoolean('RATE_LIMIT_LEGACY_HEADERS', false)
      },

      // WebSocket Configuration
      websocket: {
        enabled: this.getEnvBoolean('WEBSOCKET_ENABLED', false),
        cors: {
          origin: this.getEnvString('WEBSOCKET_CORS_ORIGIN', '*'),
          methods: ['GET', 'POST'],
          credentials: true
        },
        pingTimeout: this.getEnvNumber('WEBSOCKET_PING_TIMEOUT', 60000),
        pingInterval: this.getEnvNumber('WEBSOCKET_PING_INTERVAL', 25000)
      },

      // Logging Configuration
      logging: {
        level: this.getEnvString('LOG_LEVEL', 'info'),
        format: this.getEnvString('LOG_FORMAT', 'json'),
        enableFileLogging: this.getEnvBoolean('ENABLE_FILE_LOGGING', false),
        logDirectory: this.getEnvString('LOG_DIRECTORY', './logs'),
        maxLogFiles: this.getEnvNumber('MAX_LOG_FILES', 5),
        maxLogSize: this.getEnvString('MAX_LOG_SIZE', '10m')
      },

      // Export Configuration
      export: {
        maxRecords: this.getEnvNumber('MAX_EXPORT_RECORDS', 10000),
        csvDelimiter: this.getEnvString('CSV_DELIMITER', ','),
        csvEncoding: this.getEnvString('CSV_ENCODING', 'utf8'),
        enablePdfExport: this.getEnvBoolean('ENABLE_PDF_EXPORT', true),
        enableExcelExport: this.getEnvBoolean('ENABLE_EXCEL_EXPORT', true)
      },

      // Validation Configuration
      validation: {
        maxResponseSize: this.getEnvNumber('MAX_RESPONSE_SIZE', 1024 * 1024), // 1MB
        maxFileSize: this.getEnvNumber('MAX_FILE_SIZE', 10 * 1024 * 1024), // 10MB
        allowedFileTypes: this.getEnvArray('ALLOWED_FILE_TYPES', [
          'image/jpeg',
          'image/png',
          'image/gif',
          'application/pdf',
          'text/plain',
          'application/msword',
          'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
        ]),
        enableStrictValidation: this.getEnvBoolean('ENABLE_STRICT_VALIDATION', true),
        enableSanitization: this.getEnvBoolean('ENABLE_SANITIZATION', true)
      },

      // Security Configuration
      security: {
        bcryptRounds: this.getEnvNumber('BCRYPT_ROUNDS', 12),
        sessionSecret: this.getEnvString('SESSION_SECRET', 'session-secret-change-in-production'),
        ipWhitelist: this.getEnvArray('IP_WHITELIST', []),
        ipBlacklist: this.getEnvArray('IP_BLACKLIST', []),
        enableHsts: this.getEnvBoolean('ENABLE_HSTS', true),
        enableCsp: this.getEnvBoolean('ENABLE_CSP', true),
        corsOrigins: this.getEnvArray('CORS_ORIGINS', ['http://localhost:3000'])
      },

      // Cache Configuration
      cache: {
        enabled: this.getEnvBoolean('CACHE_ENABLED', true),
        type: this.getEnvString('CACHE_TYPE', 'memory'), // memory, redis
        ttl: this.getEnvNumber('CACHE_TTL', 300), // 5 minutes
        redisHost: this.getEnvString('REDIS_HOST', 'localhost'),
        redisPort: this.getEnvNumber('REDIS_PORT', 6379),
        redisPassword: this.getEnvString('REDIS_PASSWORD'),
        redisDatabase: this.getEnvNumber('REDIS_DATABASE', 0),
        maxCacheSize: this.getEnvNumber('MAX_CACHE_SIZE', 100) // For memory cache
      },

      // Monitoring Configuration
      monitoring: {
        enableMetrics: this.getEnvBoolean('ENABLE_METRICS', true),
        metricsPort: this.getEnvNumber('METRICS_PORT', 9090),
        enableHealthCheck: this.getEnvBoolean('ENABLE_HEALTH_CHECK', true),
        healthCheckInterval: this.getEnvNumber('HEALTH_CHECK_INTERVAL', 30000),
        enableTracing: this.getEnvBoolean('ENABLE_TRACING', false),
        tracingEndpoint: this.getEnvString('TRACING_ENDPOINT')
      },

      // Event System Configuration
      events: {
        enabled: this.getEnvBoolean('EVENTS_ENABLED', true),
        maxListeners: this.getEnvNumber('EVENT_MAX_LISTENERS', 50),
        retryAttempts: this.getEnvNumber('EVENT_RETRY_ATTEMPTS', 3),
        retryDelay: this.getEnvNumber('EVENT_RETRY_DELAY', 1000),
        eventStoreSize: this.getEnvNumber('EVENT_STORE_SIZE', 1000),
        cleanupInterval: this.getEnvNumber('EVENT_CLEANUP_INTERVAL', 3600000) // 1 hour
      },

      // Feature Flags
      features: {
        analytics: this.getEnvBoolean('FEATURE_ANALYTICS', true),
        fileUploads: this.getEnvBoolean('FEATURE_FILE_UPLOADS', true),
        realTimeUpdates: this.getEnvBoolean('FEATURE_REAL_TIME_UPDATES', false),
        advancedValidation: this.getEnvBoolean('FEATURE_ADVANCED_VALIDATION', true),
        responseExport: this.getEnvBoolean('FEATURE_RESPONSE_EXPORT', true),
        multiLanguage: this.getEnvBoolean('FEATURE_MULTI_LANGUAGE', false)
      },

      // API Configuration
      api: {
        version: this.getEnvString('API_VERSION', 'v1'),
        prefix: this.getEnvString('API_PREFIX', '/api'),
        requestTimeout: this.getEnvNumber('API_REQUEST_TIMEOUT', 30000),
        maxRequestSize: this.getEnvString('API_MAX_REQUEST_SIZE', '10mb'),
        enableCompression: this.getEnvBoolean('API_ENABLE_COMPRESSION', true)
      },

      // Development Configuration
      development: {
        enableSwagger: this.getEnvBoolean('ENABLE_SWAGGER', true),
        enableDebugRoutes: this.getEnvBoolean('ENABLE_DEBUG_ROUTES', false),
        enableHotReload: this.getEnvBoolean('ENABLE_HOT_RELOAD', false),
        enableMockData: this.getEnvBoolean('ENABLE_MOCK_DATA', false)
      }
    };
  }

  /**
   * Validate configuration for required values and consistency
   */
  validateConfiguration() {
    const errors = [];

    // Validate required production settings
    if (this.config.server.environment === 'production') {
      if (this.config.jwt.secret === 'your-jwt-secret-key') {
        errors.push('JWT_SECRET must be set to a secure value in production');
      }

      if (this.config.security.sessionSecret === 'session-secret-change-in-production') {
        errors.push('SESSION_SECRET must be set to a secure value in production');
      }
    }

    // Validate port ranges
    if (this.config.server.port < 1 || this.config.server.port > 65535) {
      errors.push('RESPONSE_SERVICE_PORT must be between 1 and 65535');
    }

    // Validate file upload configuration
    if (this.config.features.fileUploads && this.config.validation.maxFileSize <= 0) {
      errors.push('MAX_FILE_SIZE must be greater than 0 when file uploads are enabled');
    }

    // Validate rate limiting
    if (this.config.rateLimit.max <= 0) {
      errors.push('RATE_LIMIT_MAX must be greater than 0');
    }

    // Validate logging configuration
    const validLogLevels = ['error', 'warn', 'info', 'debug'];
    if (!validLogLevels.includes(this.config.logging.level)) {
      errors.push(`LOG_LEVEL must be one of: ${validLogLevels.join(', ')}`);
    }

    // Report validation errors
    if (errors.length > 0) {
      const errorMessage = `Configuration validation failed:\n${errors.join('\n')}`;
      
      if (this.config.server.environment === 'production') {
        throw new Error(errorMessage);
      } else {
        console.warn(errorMessage);
      }
    }

    console.log('Configuration loaded and validated successfully', {
      environment: this.config.server.environment,
      serviceName: this.config.server.serviceName,
      version: this.config.server.version
    });
  }

  /**
   * Get string environment variable with default
   */
  getEnvString(key, defaultValue = undefined) {
    const value = process.env[key];
    return value !== undefined ? value : defaultValue;
  }

  /**
   * Get number environment variable with default
   */
  getEnvNumber(key, defaultValue = undefined) {
    const value = process.env[key];
    if (value === undefined) return defaultValue;
    
    const parsed = parseInt(value, 10);
    if (isNaN(parsed)) {
      throw new Error(`Environment variable ${key} must be a valid number, got: ${value}`);
    }
    return parsed;
  }

  /**
   * Get boolean environment variable with default
   */
  getEnvBoolean(key, defaultValue = undefined) {
    const value = process.env[key];
    if (value === undefined) return defaultValue;
    
    const lowerValue = value.toLowerCase();
    if (lowerValue === 'true' || lowerValue === '1') return true;
    if (lowerValue === 'false' || lowerValue === '0') return false;
    
    throw new Error(`Environment variable ${key} must be 'true' or 'false', got: ${value}`);
  }

  /**
   * Get array environment variable with default
   */
  getEnvArray(key, defaultValue = []) {
    const value = process.env[key];
    if (!value) return defaultValue;
    
    return value.split(',').map(item => item.trim()).filter(item => item.length > 0);
  }

  /**
   * Get configuration value by path
   */
  get(path) {
    const keys = path.split('.');
    let current = this.config;
    
    for (const key of keys) {
      if (current && typeof current === 'object' && key in current) {
        current = current[key];
      } else {
        return undefined;
      }
    }
    
    return current;
  }

  /**
   * Get all configuration
   */
  getAll() {
    return { ...this.config };
  }

  /**
   * Get sanitized configuration (removes sensitive data)
   */
  getSanitized() {
    const sanitized = JSON.parse(JSON.stringify(this.config));
    
    // Remove sensitive fields
    if (sanitized.jwt) sanitized.jwt.secret = '[REDACTED]';
    if (sanitized.security) sanitized.security.sessionSecret = '[REDACTED]';
    if (sanitized.database) {
      sanitized.database.password = '[REDACTED]';
      sanitized.database.url = '[REDACTED]';
    }
    if (sanitized.formService) {
      sanitized.formService.apiKey = '[REDACTED]';
    }
    if (sanitized.googleSheets?.credentials) {
      sanitized.googleSheets.credentials.private_key = '[REDACTED]';
    }
    if (sanitized.cache) {
      sanitized.cache.redisPassword = '[REDACTED]';
    }
    
    return sanitized;
  }

  /**
   * Check if running in production
   */
  isProduction() {
    return this.config.server.environment === 'production';
  }

  /**
   * Check if running in development
   */
  isDevelopment() {
    return this.config.server.environment === 'development';
  }

  /**
   * Check if running in test
   */
  isTest() {
    return this.config.server.environment === 'test';
  }

  /**
   * Check if feature is enabled
   */
  isFeatureEnabled(featureName) {
    return this.config.features[featureName] === true;
  }
}

// Create and export singleton instance
const enhancedConfig = new ConfigurationManager();

// Also export legacy config for backward compatibility
const legacyConfig = enhancedConfig.getAll();

module.exports = enhancedConfig;
module.exports.legacy = legacyConfig;
