// Enhanced Swagger/OpenAPI Documentation Configuration for Auth Service
// Following OpenAPI 3.0.3 specification and industry best practices

import swaggerJSDoc from 'swagger-jsdoc';
import swaggerUi from 'swagger-ui-express';

// OpenAPI 3.0.3 Specification Configuration
const swaggerDefinition = {
  openapi: '3.0.3',
  info: {
    title: 'X-Form Auth Service API',
    version: '1.0.0',
    description: `
      # X-Form Authentication & User Management Service
      
      A production-ready authentication and user management microservice built with **Clean Architecture** and **SOLID principles**.
      
      ## 🏗️ Architecture Features
      - **Clean Architecture** with proper layer separation
      - **SOLID Principles** implementation
      - **Domain-Driven Design** with rich domain models
      - **Event-Driven Architecture** for extensibility
      - **Microservices patterns** with health monitoring
      
      ## 🔐 Security Features
      - **JWT-based authentication** with access and refresh tokens
      - **BCrypt password hashing** with 12 salt rounds
      - **Rate limiting** protection (global and endpoint-specific)
      - **Account lockout** after failed login attempts
      - **Input validation** with comprehensive error handling
      - **CORS and security headers** configuration
      
      ## 📊 API Features
      - **RESTful design** with proper HTTP methods and status codes
      - **Comprehensive error handling** with standardized error codes
      - **Request correlation tracking** for debugging and monitoring
      - **Health checks** and dependency monitoring
      - **Interactive API testing** with built-in authentication
      
      ## 🚀 Getting Started
      1. Use the **"Authorize"** button to authenticate with JWT tokens
      2. Start with **POST /api/v1/auth/register** to create a test user
      3. Use **POST /api/v1/auth/login** to get access tokens
      4. Test protected endpoints with the received tokens
    `,
    termsOfService: 'https://xform.com/terms',
    contact: {
      name: 'X-Form Development Team',
      email: 'dev@xform.com',
      url: 'https://github.com/Mir00r/X-Form-Backend',
    },
    license: {
      name: 'MIT',
      url: 'https://opensource.org/licenses/MIT',
    },
    'x-logo': {
      url: 'https://via.placeholder.com/200x80/2c3e50/ffffff?text=X-Form',
      altText: 'X-Form Logo'
    }
  },
  servers: [
    {
      url: 'http://localhost:3001',
      description: 'Development server',
      variables: {
        port: {
          default: '3001',
          description: 'Development port'
        }
      }
    },
    {
      url: 'https://auth-dev.xform.com',
      description: 'Development environment'
    },
    {
      url: 'https://auth-staging.xform.com',
      description: 'Staging environment'
    },
    {
      url: 'https://auth.xform.com',
      description: 'Production environment'
    }
  ],
  tags: [
    {
      name: 'Authentication',
      description: 'User authentication and token management operations',
      externalDocs: {
        description: 'Authentication Guide',
        url: 'https://docs.xform.com/auth'
      }
    },
    {
      name: 'User Management',
      description: 'User profile and account management operations'
    },
    {
      name: 'Email Verification',
      description: 'Email verification and related operations'
    },
    {
      name: 'Password Management',
      description: 'Password reset, change, and security operations'
    },
    {
      name: 'Health & Monitoring',
      description: 'Service health checks and monitoring endpoints'
    }
  ],
  components: {
    securitySchemes: {
      BearerAuth: {
        type: 'http',
        scheme: 'bearer',
        bearerFormat: 'JWT',
        description: 'JWT access token for API authentication. Format: `Bearer <token>`'
      },
      ApiKeyAuth: {
        type: 'apiKey',
        in: 'header',
        name: 'X-API-Key',
        description: 'API key for service-to-service authentication'
      }
    },
    parameters: {
      CorrelationId: {
        name: 'X-Correlation-ID',
        in: 'header',
        description: 'Unique identifier for request tracing and debugging',
        required: false,
        schema: {
          type: 'string',
          format: 'uuid',
          example: '550e8400-e29b-41d4-a716-446655440000'
        }
      },
      ApiVersion: {
        name: 'version',
        in: 'path',
        required: true,
        description: 'API version identifier',
        schema: {
          type: 'string',
          enum: ['v1'],
          default: 'v1'
        }
      }
    },
    schemas: {
      // Base Response Schemas
      SuccessResponse: {
        type: 'object',
        required: ['success', 'timestamp', 'path', 'method', 'correlationId'],
        properties: {
          success: {
            type: 'boolean',
            example: true,
            description: 'Indicates if the request was successful'
          },
          timestamp: {
            type: 'string',
            format: 'date-time',
            description: 'ISO 8601 timestamp of the response',
            example: '2024-01-15T10:30:00.000Z'
          },
          path: {
            type: 'string',
            description: 'Request path that was called',
            example: '/api/v1/auth/login'
          },
          method: {
            type: 'string',
            description: 'HTTP method used',
            enum: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH'],
            example: 'POST'
          },
          correlationId: {
            type: 'string',
            format: 'uuid',
            description: 'Unique identifier for request tracing',
            example: '550e8400-e29b-41d4-a716-446655440000'
          },
          data: {
            type: 'object',
            description: 'Response payload'
          },
          meta: {
            $ref: '#/components/schemas/ResponseMeta'
          }
        }
      },
      ErrorResponse: {
        type: 'object',
        required: ['success', 'timestamp', 'path', 'method', 'correlationId', 'error'],
        properties: {
          success: {
            type: 'boolean',
            example: false,
            description: 'Always false for error responses'
          },
          timestamp: {
            type: 'string',
            format: 'date-time',
            description: 'ISO 8601 timestamp of the error',
            example: '2024-01-15T10:30:00.000Z'
          },
          path: {
            type: 'string',
            description: 'Request path that caused the error',
            example: '/api/v1/auth/login'
          },
          method: {
            type: 'string',
            description: 'HTTP method used',
            example: 'POST'
          },
          correlationId: {
            type: 'string',
            format: 'uuid',
            description: 'Unique identifier for error tracing',
            example: '550e8400-e29b-41d4-a716-446655440000'
          },
          error: {
            $ref: '#/components/schemas/ApiError'
          }
        }
      },
      ApiError: {
        type: 'object',
        required: ['code', 'message', 'timestamp'],
        properties: {
          code: {
            type: 'string',
            description: 'Machine-readable error code',
            enum: [
              'VALIDATION_ERROR',
              'AUTHENTICATION_FAILED',
              'AUTHORIZATION_FAILED',
              'INVALID_CREDENTIALS',
              'ACCOUNT_LOCKED',
              'TOKEN_EXPIRED',
              'TOKEN_INVALID',
              'USER_NOT_FOUND',
              'USER_ALREADY_EXISTS',
              'EMAIL_NOT_VERIFIED',
              'RATE_LIMIT_EXCEEDED',
              'INTERNAL_SERVER_ERROR'
            ],
            example: 'VALIDATION_ERROR'
          },
          message: {
            type: 'string',
            description: 'Human-readable error message',
            example: 'Validation failed for the provided input'
          },
          timestamp: {
            type: 'string',
            format: 'date-time',
            description: 'When the error occurred',
            example: '2024-01-15T10:30:00.000Z'
          },
          path: {
            type: 'string',
            description: 'API path where error occurred',
            example: '/api/v1/auth/login'
          },
          correlationId: {
            type: 'string',
            format: 'uuid',
            description: 'Correlation ID for tracing',
            example: '550e8400-e29b-41d4-a716-446655440000'
          },
          details: {
            oneOf: [
              {
                type: 'array',
                items: { $ref: '#/components/schemas/ValidationError' },
                description: 'Detailed validation errors'
              },
              {
                type: 'object',
                description: 'Additional error details'
              }
            ]
          }
        }
      },
      ValidationError: {
        type: 'object',
        required: ['field', 'message', 'code'],
        properties: {
          field: {
            type: 'string',
            description: 'Field name that failed validation',
            example: 'email'
          },
          message: {
            type: 'string',
            description: 'Validation error message',
            example: 'Must be a valid email address'
          },
          value: {
            description: 'The invalid value that was provided',
            example: 'invalid-email'
          },
          code: {
            type: 'string',
            description: 'Validation error code',
            example: 'INVALID_EMAIL'
          }
        }
      },
      ResponseMeta: {
        type: 'object',
        properties: {
          version: {
            type: 'string',
            description: 'API version',
            example: 'v1'
          },
          rateLimit: {
            $ref: '#/components/schemas/RateLimitInfo'
          },
          pagination: {
            $ref: '#/components/schemas/PaginationInfo'
          }
        }
      },
      RateLimitInfo: {
        type: 'object',
        properties: {
          limit: {
            type: 'integer',
            description: 'Rate limit maximum requests',
            example: 100
          },
          remaining: {
            type: 'integer',
            description: 'Remaining requests in current window',
            example: 95
          },
          resetTime: {
            type: 'string',
            format: 'date-time',
            description: 'When the rate limit resets',
            example: '2024-01-15T10:45:00.000Z'
          }
        }
      },
      PaginationInfo: {
        type: 'object',
        properties: {
          page: {
            type: 'integer',
            minimum: 1,
            description: 'Current page number',
            example: 1
          },
          limit: {
            type: 'integer',
            minimum: 1,
            maximum: 100,
            description: 'Items per page',
            example: 20
          },
          total: {
            type: 'integer',
            minimum: 0,
            description: 'Total number of items',
            example: 150
          },
          totalPages: {
            type: 'integer',
            minimum: 1,
            description: 'Total number of pages',
            example: 8
          },
          hasNext: {
            type: 'boolean',
            description: 'Whether there are more pages',
            example: true
          },
          hasPrev: {
            type: 'boolean',
            description: 'Whether there are previous pages',
            example: false
          }
        }
      },
      
      // Authentication Request Schemas
      RegisterRequest: {
        type: 'object',
        required: ['email', 'username', 'password', 'confirmPassword', 'firstName', 'lastName', 'acceptTerms'],
        properties: {
          email: {
            type: 'string',
            format: 'email',
            maxLength: 100,
            description: 'User email address (must be unique)',
            example: 'john.doe@example.com'
          },
          username: {
            type: 'string',
            minLength: 3,
            maxLength: 30,
            pattern: '^[a-zA-Z0-9_-]+$',
            description: 'Unique username (alphanumeric, underscore, dash only)',
            example: 'johndoe'
          },
          password: {
            type: 'string',
            minLength: 8,
            maxLength: 128,
            description: 'Password (minimum 8 characters, must include uppercase, lowercase, number, special character)',
            example: 'SecurePass123!',
            format: 'password'
          },
          confirmPassword: {
            type: 'string',
            minLength: 8,
            maxLength: 128,
            description: 'Password confirmation (must match password)',
            example: 'SecurePass123!',
            format: 'password'
          },
          firstName: {
            type: 'string',
            minLength: 1,
            maxLength: 50,
            description: 'User first name',
            example: 'John'
          },
          lastName: {
            type: 'string',
            minLength: 1,
            maxLength: 50,
            description: 'User last name',
            example: 'Doe'
          },
          acceptTerms: {
            type: 'boolean',
            description: 'User must accept terms and conditions',
            example: true
          },
          marketingConsent: {
            type: 'boolean',
            description: 'Optional consent for marketing communications',
            example: false,
            default: false
          },
          referralCode: {
            type: 'string',
            minLength: 6,
            maxLength: 20,
            description: 'Optional referral code',
            example: 'REF12345'
          }
        }
      },
      LoginRequest: {
        type: 'object',
        required: ['email', 'password'],
        properties: {
          email: {
            type: 'string',
            format: 'email',
            description: 'User email address',
            example: 'john.doe@example.com'
          },
          password: {
            type: 'string',
            maxLength: 128,
            description: 'User password',
            example: 'SecurePass123!',
            format: 'password'
          },
          rememberMe: {
            type: 'boolean',
            description: 'Extended session duration',
            example: false,
            default: false
          },
          deviceId: {
            type: 'string',
            format: 'uuid',
            description: 'Unique device identifier for tracking',
            example: '550e8400-e29b-41d4-a716-446655440001'
          },
          deviceName: {
            type: 'string',
            maxLength: 100,
            description: 'Human-readable device name',
            example: 'iPhone 15 Pro'
          }
        }
      },
      RefreshTokenRequest: {
        type: 'object',
        required: ['refreshToken'],
        properties: {
          refreshToken: {
            type: 'string',
            description: 'Valid refresh token received from login',
            example: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.refresh.token.signature'
          },
          deviceId: {
            type: 'string',
            format: 'uuid',
            description: 'Device identifier for security validation',
            example: '550e8400-e29b-41d4-a716-446655440001'
          }
        }
      },
      
      // Authentication Response Schemas
      AuthResponse: {
        type: 'object',
        required: ['accessToken', 'refreshToken', 'tokenType', 'expiresIn', 'user'],
        properties: {
          accessToken: {
            type: 'string',
            description: 'JWT access token for API authentication',
            example: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.access.token.signature'
          },
          refreshToken: {
            type: 'string',
            description: 'JWT refresh token for obtaining new access tokens',
            example: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.refresh.token.signature'
          },
          tokenType: {
            type: 'string',
            description: 'Token type (always Bearer)',
            example: 'Bearer',
            enum: ['Bearer']
          },
          expiresIn: {
            type: 'integer',
            description: 'Access token lifetime in seconds',
            example: 900,
            minimum: 1
          },
          expiresAt: {
            type: 'string',
            format: 'date-time',
            description: 'Absolute expiration time of access token',
            example: '2024-01-15T10:45:00.000Z'
          },
          scope: {
            type: 'array',
            items: {
              type: 'string',
              enum: ['read', 'write', 'admin', 'user']
            },
            description: 'Token permissions scope',
            example: ['read', 'write']
          },
          user: {
            $ref: '#/components/schemas/UserProfile'
          }
        }
      },
      UserProfile: {
        type: 'object',
        required: ['id', 'email', 'username', 'firstName', 'lastName', 'role', 'emailVerified', 'accountStatus', 'createdAt', 'updatedAt'],
        properties: {
          id: {
            type: 'string',
            format: 'uuid',
            description: 'Unique user identifier',
            example: '550e8400-e29b-41d4-a716-446655440000'
          },
          email: {
            type: 'string',
            format: 'email',
            description: 'User email address',
            example: 'john.doe@example.com'
          },
          username: {
            type: 'string',
            description: 'Unique username',
            example: 'johndoe'
          },
          firstName: {
            type: 'string',
            description: 'User first name',
            example: 'John'
          },
          lastName: {
            type: 'string',
            description: 'User last name',
            example: 'Doe'
          },
          fullName: {
            type: 'string',
            description: 'Computed full name',
            example: 'John Doe'
          },
          role: {
            type: 'string',
            enum: ['USER', 'ADMIN', 'MODERATOR', 'SUPER_ADMIN'],
            description: 'User role and permissions level',
            example: 'USER'
          },
          emailVerified: {
            type: 'boolean',
            description: 'Whether email has been verified',
            example: true
          },
          phoneVerified: {
            type: 'boolean',
            description: 'Whether phone number has been verified',
            example: false
          },
          accountStatus: {
            type: 'string',
            enum: ['ACTIVE', 'SUSPENDED', 'PENDING_VERIFICATION', 'LOCKED', 'DEACTIVATED'],
            description: 'Current account status',
            example: 'ACTIVE'
          },
          lastLoginAt: {
            type: 'string',
            format: 'date-time',
            nullable: true,
            description: 'Last successful login timestamp',
            example: '2024-01-15T09:30:00.000Z'
          },
          createdAt: {
            type: 'string',
            format: 'date-time',
            description: 'Account creation timestamp',
            example: '2024-01-01T00:00:00.000Z'
          },
          updatedAt: {
            type: 'string',
            format: 'date-time',
            description: 'Last profile update timestamp',
            example: '2024-01-15T10:00:00.000Z'
          },
          preferences: {
            $ref: '#/components/schemas/UserPreferences'
          },
          metadata: {
            type: 'object',
            additionalProperties: true,
            description: 'Additional user metadata',
            example: {
              theme: 'dark',
              onboardingCompleted: true
            }
          }
        }
      },
      UserPreferences: {
        type: 'object',
        properties: {
          language: {
            type: 'string',
            description: 'Preferred language code (ISO 639-1)',
            example: 'en',
            default: 'en'
          },
          timezone: {
            type: 'string',
            description: 'User timezone (IANA timezone)',
            example: 'America/New_York',
            default: 'UTC'
          },
          emailNotifications: {
            type: 'boolean',
            description: 'Email notifications enabled',
            example: true,
            default: true
          },
          smsNotifications: {
            type: 'boolean',
            description: 'SMS notifications enabled',
            example: false,
            default: false
          },
          marketingEmails: {
            type: 'boolean',
            description: 'Marketing email consent',
            example: false,
            default: false
          },
          twoFactorEnabled: {
            type: 'boolean',
            description: 'Two-factor authentication status',
            example: false,
            default: false
          }
        }
      },
      
      // Health Check Schemas
      HealthCheck: {
        type: 'object',
        required: ['service', 'version', 'status', 'uptime', 'timestamp', 'environment'],
        properties: {
          service: {
            type: 'string',
            description: 'Service name',
            example: 'auth-service'
          },
          version: {
            type: 'string',
            description: 'Service version',
            example: '1.0.0'
          },
          status: {
            type: 'string',
            enum: ['HEALTHY', 'UNHEALTHY', 'DEGRADED'],
            description: 'Overall service health status',
            example: 'HEALTHY'
          },
          uptime: {
            type: 'number',
            description: 'Service uptime in seconds',
            example: 3600.5,
            minimum: 0
          },
          timestamp: {
            type: 'string',
            format: 'date-time',
            description: 'Health check timestamp',
            example: '2024-01-15T10:30:00.000Z'
          },
          environment: {
            type: 'string',
            enum: ['development', 'staging', 'production'],
            description: 'Current environment',
            example: 'development'
          },
          dependencies: {
            type: 'array',
            items: {
              $ref: '#/components/schemas/DependencyHealth'
            },
            description: 'Health status of service dependencies'
          },
          metrics: {
            $ref: '#/components/schemas/ServiceMetrics'
          }
        }
      },
      DependencyHealth: {
        type: 'object',
        required: ['name', 'type', 'status', 'lastChecked'],
        properties: {
          name: {
            type: 'string',
            description: 'Dependency name',
            example: 'postgresql'
          },
          type: {
            type: 'string',
            enum: ['DATABASE', 'CACHE', 'EMAIL', 'EXTERNAL_API', 'QUEUE'],
            description: 'Type of dependency',
            example: 'DATABASE'
          },
          status: {
            type: 'string',
            enum: ['HEALTHY', 'UNHEALTHY', 'DEGRADED'],
            description: 'Dependency health status',
            example: 'HEALTHY'
          },
          responseTime: {
            type: 'number',
            description: 'Response time in milliseconds',
            example: 25.5,
            minimum: 0
          },
          lastChecked: {
            type: 'string',
            format: 'date-time',
            description: 'Last health check timestamp',
            example: '2024-01-15T10:30:00.000Z'
          },
          error: {
            type: 'string',
            nullable: true,
            description: 'Error message if unhealthy',
            example: null
          }
        }
      },
      ServiceMetrics: {
        type: 'object',
        properties: {
          requestCount: {
            type: 'integer',
            description: 'Total requests processed',
            example: 1500,
            minimum: 0
          },
          errorRate: {
            type: 'number',
            description: 'Error rate percentage',
            example: 0.02,
            minimum: 0,
            maximum: 1
          },
          averageResponseTime: {
            type: 'number',
            description: 'Average response time in milliseconds',
            example: 120.5,
            minimum: 0
          },
          activeConnections: {
            type: 'integer',
            description: 'Current active connections',
            example: 15,
            minimum: 0
          },
          memoryUsage: {
            type: 'object',
            properties: {
              used: {
                type: 'number',
                description: 'Used memory in MB',
                example: 256
              },
              free: {
                type: 'number',
                description: 'Free memory in MB',
                example: 768
              },
              total: {
                type: 'number',
                description: 'Total memory in MB',
                example: 1024
              },
              percentage: {
                type: 'number',
                description: 'Memory usage percentage',
                example: 25.0,
                minimum: 0,
                maximum: 100
              }
            }
          },
          cpuUsage: {
            type: 'number',
            description: 'CPU usage percentage',
            example: 15.5,
            minimum: 0,
            maximum: 100
          }
        }
      }
    },
    responses: {
      '200': {
        description: 'Success',
        content: {
          'application/json': {
            schema: {
              $ref: '#/components/schemas/SuccessResponse'
            }
          }
        }
      },
      '201': {
        description: 'Created successfully',
        content: {
          'application/json': {
            schema: {
              $ref: '#/components/schemas/SuccessResponse'
            }
          }
        }
      },
      '400': {
        description: 'Bad Request - Validation Error',
        content: {
          'application/json': {
            schema: {
              $ref: '#/components/schemas/ErrorResponse'
            },
            example: {
              success: false,
              timestamp: '2024-01-15T10:30:00.000Z',
              path: '/api/v1/auth/register',
              method: 'POST',
              correlationId: '550e8400-e29b-41d4-a716-446655440000',
              error: {
                code: 'VALIDATION_ERROR',
                message: 'Validation failed',
                timestamp: '2024-01-15T10:30:00.000Z',
                path: '/api/v1/auth/register',
                correlationId: '550e8400-e29b-41d4-a716-446655440000',
                details: [
                  {
                    field: 'email',
                    message: 'Must be a valid email address',
                    value: 'invalid-email',
                    code: 'INVALID_EMAIL'
                  }
                ]
              }
            }
          }
        }
      },
      '401': {
        description: 'Unauthorized - Authentication Required',
        content: {
          'application/json': {
            schema: {
              $ref: '#/components/schemas/ErrorResponse'
            },
            example: {
              success: false,
              timestamp: '2024-01-15T10:30:00.000Z',
              path: '/api/v1/auth/profile',
              method: 'GET',
              correlationId: '550e8400-e29b-41d4-a716-446655440000',
              error: {
                code: 'AUTHENTICATION_FAILED',
                message: 'Valid authentication token required',
                timestamp: '2024-01-15T10:30:00.000Z',
                path: '/api/v1/auth/profile',
                correlationId: '550e8400-e29b-41d4-a716-446655440000'
              }
            }
          }
        }
      },
      '403': {
        description: 'Forbidden - Insufficient Permissions',
        content: {
          'application/json': {
            schema: {
              $ref: '#/components/schemas/ErrorResponse'
            }
          }
        }
      },
      '404': {
        description: 'Not Found - Resource Not Found',
        content: {
          'application/json': {
            schema: {
              $ref: '#/components/schemas/ErrorResponse'
            }
          }
        }
      },
      '409': {
        description: 'Conflict - Resource Already Exists',
        content: {
          'application/json': {
            schema: {
              $ref: '#/components/schemas/ErrorResponse'
            },
            example: {
              success: false,
              timestamp: '2024-01-15T10:30:00.000Z',
              path: '/api/v1/auth/register',
              method: 'POST',
              correlationId: '550e8400-e29b-41d4-a716-446655440000',
              error: {
                code: 'USER_ALREADY_EXISTS',
                message: 'User with this email already exists',
                timestamp: '2024-01-15T10:30:00.000Z',
                path: '/api/v1/auth/register',
                correlationId: '550e8400-e29b-41d4-a716-446655440000'
              }
            }
          }
        }
      },
      '423': {
        description: 'Locked - Account Locked',
        content: {
          'application/json': {
            schema: {
              $ref: '#/components/schemas/ErrorResponse'
            }
          }
        }
      },
      '429': {
        description: 'Too Many Requests - Rate Limit Exceeded',
        content: {
          'application/json': {
            schema: {
              $ref: '#/components/schemas/ErrorResponse'
            },
            example: {
              success: false,
              timestamp: '2024-01-15T10:30:00.000Z',
              path: '/api/v1/auth/login',
              method: 'POST',
              correlationId: '550e8400-e29b-41d4-a716-446655440000',
              error: {
                code: 'RATE_LIMIT_EXCEEDED',
                message: 'Too many requests. Please try again later.',
                timestamp: '2024-01-15T10:30:00.000Z',
                path: '/api/v1/auth/login',
                correlationId: '550e8400-e29b-41d4-a716-446655440000'
              }
            }
          }
        }
      },
      '500': {
        description: 'Internal Server Error',
        content: {
          'application/json': {
            schema: {
              $ref: '#/components/schemas/ErrorResponse'
            },
            example: {
              success: false,
              timestamp: '2024-01-15T10:30:00.000Z',
              path: '/api/v1/auth/register',
              method: 'POST',
              correlationId: '550e8400-e29b-41d4-a716-446655440000',
              error: {
                code: 'INTERNAL_SERVER_ERROR',
                message: 'An unexpected error occurred',
                timestamp: '2024-01-15T10:30:00.000Z',
                path: '/api/v1/auth/register',
                correlationId: '550e8400-e29b-41d4-a716-446655440000'
              }
            }
          }
        }
      }
    }
  },
  security: [
    {
      BearerAuth: []
    }
  ],
  externalDocs: {
    description: 'X-Form Documentation',
    url: 'https://docs.xform.com'
  }
};

// Swagger JSDoc Options
const swaggerOptions = {
  definition: swaggerDefinition,
  apis: [
    './src/interface/http/routes/*.ts',
    './src/interface/http/*.ts',
    './src/**/*.ts',
    './docs/api/*.yaml'
  ]
};

// Generate OpenAPI specification
export const swaggerSpec = swaggerJSDoc(swaggerOptions);

// Enhanced Swagger UI Options with Custom Styling
export const swaggerUiOptions: swaggerUi.SwaggerUiOptions = {
  customCss: `
    .swagger-ui .topbar { 
      background-color: #2c3e50; 
      border-bottom: 3px solid #3498db;
    }
    .swagger-ui .topbar .download-url-wrapper .select-label {
      color: #ffffff;
    }
    .swagger-ui .info .title { 
      color: #2c3e50; 
      font-size: 2.5em; 
      font-weight: bold;
      margin-bottom: 10px;
    }
    .swagger-ui .info .description p { 
      font-size: 1.1em; 
      line-height: 1.6; 
      color: #34495e;
    }
    .swagger-ui .info .description h1,
    .swagger-ui .info .description h2,
    .swagger-ui .info .description h3 {
      color: #2c3e50;
      margin-top: 25px;
      margin-bottom: 15px;
    }
    .swagger-ui .scheme-container { 
      background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%); 
      padding: 20px; 
      border-radius: 8px;
      border: 1px solid #dee2e6;
      margin: 20px 0;
    }
    .swagger-ui .auth-wrapper { 
      margin-top: 20px;
      background: #f8f9fa;
      padding: 15px;
      border-radius: 8px;
      border: 1px solid #dee2e6;
    }
    .swagger-ui .btn.authorize { 
      background-color: #3498db; 
      border-color: #3498db;
      color: white;
      font-weight: bold;
      padding: 8px 20px;
      border-radius: 5px;
    }
    .swagger-ui .btn.authorize:hover { 
      background-color: #2980b9; 
      border-color: #2980b9;
      transform: translateY(-1px);
      box-shadow: 0 4px 8px rgba(0,0,0,0.1);
    }
    .swagger-ui .model-title { 
      color: #2c3e50;
      font-weight: bold;
    }
    .swagger-ui .operation-tag-content { 
      font-size: 1.3em;
      font-weight: bold;
      color: #2c3e50;
    }
    .swagger-ui .opblock { 
      margin-bottom: 25px; 
      border-radius: 8px;
      box-shadow: 0 2px 4px rgba(0,0,0,0.1);
      border: 1px solid #dee2e6;
    }
    .swagger-ui .opblock-summary { 
      border-radius: 8px 8px 0 0;
      padding: 15px;
    }
    .swagger-ui .opblock.opblock-post .opblock-summary {
      background: rgba(73, 204, 144, 0.1);
      border-color: #49cc90;
    }
    .swagger-ui .opblock.opblock-get .opblock-summary {
      background: rgba(97, 175, 254, 0.1);
      border-color: #61affe;
    }
    .swagger-ui .opblock.opblock-put .opblock-summary {
      background: rgba(252, 161, 48, 0.1);
      border-color: #fca130;
    }
    .swagger-ui .opblock.opblock-delete .opblock-summary {
      background: rgba(249, 62, 62, 0.1);
      border-color: #f93e3e;
    }
    .swagger-ui .highlight-code { 
      background: #f8f9fa;
      border: 1px solid #e9ecef;
      border-radius: 4px;
    }
    .swagger-ui .model { 
      font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
      background: #f8f9fa;
      border: 1px solid #e9ecef;
      border-radius: 4px;
    }
    .swagger-ui .response-col_status { 
      font-weight: bold;
      font-size: 1.1em;
    }
    .swagger-ui .parameters-col_description p { 
      margin: 0;
      color: #6c757d;
    }
    .swagger-ui .tab { 
      border-radius: 4px 4px 0 0;
      border: 1px solid #dee2e6;
    }
    .swagger-ui .response-col_links { 
      font-size: 0.9em;
      color: #6c757d;
    }
    .swagger-ui .parameter__name {
      font-weight: bold;
      color: #2c3e50;
    }
    .swagger-ui .parameter__type {
      font-style: italic;
      color: #7f8c8d;
    }
    .swagger-ui .response-control-media-type {
      margin-top: 10px;
    }
    .swagger-ui .examples__example {
      margin-top: 15px;
      padding: 15px;
      background: #f8f9fa;
      border-radius: 4px;
      border: 1px solid #e9ecef;
    }
    .swagger-ui .operation-tag {
      margin: 30px 0 15px 0;
    }
    .swagger-ui .operation-tag:first-child {
      margin-top: 0;
    }
  `,
  customSiteTitle: 'X-Form Auth Service API - Interactive Documentation',
  customfavIcon: '/favicon.ico',
  swaggerOptions: {
    tryItOutEnabled: true,
    requestInterceptor: (req: any) => {
      // Add correlation ID to all requests for tracing
      if (!req.headers['X-Correlation-ID']) {
        req.headers['X-Correlation-ID'] = 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
          const r = Math.random() * 16 | 0;
          const v = c === 'x' ? r : (r & 0x3 | 0x8);
          return v.toString(16);
        });
      }
      return req;
    },
    responseInterceptor: (res: any) => {
      // Log API calls for development and debugging
      if (process.env.NODE_ENV === 'development') {
        console.log(`🌐 Swagger UI API Call: ${res.url} - Status: ${res.status} - Time: ${new Date().toISOString()}`);
      }
      return res;
    },
    docExpansion: 'list', // Show operations list expanded
    defaultModelsExpandDepth: 3, // Expand models to show more detail
    defaultModelExpandDepth: 3,
    defaultModelRendering: 'example', // Show examples by default
    displayRequestDuration: true, // Show request duration in UI
    filter: true, // Enable API filtering and search
    showExtensions: true,
    showCommonExtensions: true,
    persistAuthorization: true, // Persist auth tokens in browser storage
    displayOperationId: false,
    deepLinking: true, // Enable deep linking to operations
    validatorUrl: null, // Disable online spec validation
    supportedSubmitMethods: ['get', 'post', 'put', 'delete', 'patch'], // Enable all methods for testing
    oauth2RedirectUrl: `${process.env.BASE_URL || 'http://localhost:3001'}/api-docs/oauth2-redirect.html`,
    syntaxHighlight: {
      activate: true,
      theme: 'agate'
    }
  }
};

// Custom HTML template with enhanced features
export const getSwaggerHTML = (spec: any): string => {
  return `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>X-Form Auth Service API Documentation</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css" />
  <link rel="icon" type="image/png" href="https://via.placeholder.com/32x32/3498db/ffffff?text=XF" sizes="32x32" />
  <style>
    html {
      box-sizing: border-box;
      overflow: -moz-scrollbars-vertical;
      overflow-y: scroll;
    }
    *, *:before, *:after {
      box-sizing: inherit;
    }
    body {
      margin: 0;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    }
    .swagger-ui {
      max-width: 1400px;
      margin: 0 auto;
      background: white;
      box-shadow: 0 0 30px rgba(0,0,0,0.1);
    }
    .api-header {
      background: linear-gradient(135deg, #2c3e50 0%, #3498db 100%);
      color: white;
      padding: 20px;
      text-align: center;
      margin-bottom: 0;
    }
    .api-header h1 {
      margin: 0;
      font-size: 2.5em;
      font-weight: 300;
    }
    .api-header p {
      margin: 10px 0 0 0;
      opacity: 0.9;
      font-size: 1.1em;
    }
    .quick-actions {
      background: #f8f9fa;
      padding: 15px;
      border-bottom: 1px solid #dee2e6;
      text-align: center;
    }
    .quick-actions button {
      background: #3498db;
      color: white;
      border: none;
      padding: 8px 16px;
      margin: 0 5px;
      border-radius: 4px;
      cursor: pointer;
      font-size: 14px;
    }
    .quick-actions button:hover {
      background: #2980b9;
    }
    .footer {
      background: #2c3e50;
      color: white;
      padding: 20px;
      text-align: center;
      margin-top: 40px;
    }
    .footer a {
      color: #3498db;
      text-decoration: none;
    }
    .footer a:hover {
      text-decoration: underline;
    }
  </style>
</head>
<body>
  <div class="api-header">
    <h1>🔐 X-Form Auth Service API</h1>
    <p>Comprehensive Authentication & User Management with Clean Architecture</p>
  </div>
  
  <div class="quick-actions">
    <button onclick="expandAll()">📖 Expand All</button>
    <button onclick="collapseAll()">📄 Collapse All</button>
    <button onclick="downloadSpec()">💾 Download OpenAPI Spec</button>
    <button onclick="copyBaseUrl()">🔗 Copy Base URL</button>
  </div>

  <div id="swagger-ui"></div>
  
  <div class="footer">
    <p>Built with ❤️ using Clean Architecture and SOLID Principles</p>
    <p>
      <a href="https://github.com/Mir00r/X-Form-Backend">📚 Documentation</a> | 
      <a href="/health">🏥 Health Check</a> | 
      <a href="/api-docs.json">📋 OpenAPI Spec</a>
    </p>
  </div>

  <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        spec: ${JSON.stringify(spec)},
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout",
        tryItOutEnabled: true,
        requestInterceptor: function(request) {
          // Add correlation ID to all requests
          request.headers['X-Correlation-ID'] = request.headers['X-Correlation-ID'] || 
            'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
              var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
              return v.toString(16);
            });
          return request;
        },
        responseInterceptor: function(response) {
          console.log('API Response:', response.status, response.url);
          return response;
        },
        onComplete: function() {
          console.log('✅ Swagger UI loaded successfully');
        },
        docExpansion: 'list',
        defaultModelsExpandDepth: 2,
        filter: true,
        showExtensions: true,
        persistAuthorization: true
      });

      // Make UI globally accessible for custom functions
      window.swaggerUI = ui;
    };

    // Custom utility functions
    function expandAll() {
      document.querySelectorAll('.opblock').forEach(block => {
        if (!block.classList.contains('is-open')) {
          const button = block.querySelector('.opblock-summary');
          if (button) button.click();
        }
      });
    }

    function collapseAll() {
      document.querySelectorAll('.opblock.is-open').forEach(block => {
        const button = block.querySelector('.opblock-summary');
        if (button) button.click();
      });
    }

    function downloadSpec() {
      const spec = ${JSON.stringify(spec)};
      const blob = new Blob([JSON.stringify(spec, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'x-form-auth-service-openapi.json';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    }

    function copyBaseUrl() {
      const baseUrl = window.location.origin;
      navigator.clipboard.writeText(baseUrl).then(() => {
        alert('Base URL copied to clipboard: ' + baseUrl);
      }).catch(() => {
        prompt('Copy this URL:', baseUrl);
      });
    }
  </script>
</body>
</html>
  `;
};

export default swaggerSpec;
