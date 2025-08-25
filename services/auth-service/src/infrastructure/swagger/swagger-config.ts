// Swagger/OpenAPI Documentation Configuration for Auth Service
// Comprehensive API documentation following microservices best practices

import swaggerJSDoc from 'swagger-jsdoc';
import { SwaggerDefinition } from 'swagger-jsdoc';

const swaggerDefinition: SwaggerDefinition = {
  openapi: '3.0.3',
  info: {
    title: 'X-Form Auth Service API',
    version: '1.0.0',
    description: `
      # X-Form Authentication & User Management Service
      
      A comprehensive authentication and user management service built with Clean Architecture and SOLID principles.
      
      ## Features
      - 🔐 JWT-based authentication with refresh tokens
      - 👤 User registration and profile management
      - 📧 Email verification and password reset
      - 🛡️ Rate limiting and security features
      - 📊 Health checks and monitoring
      - 🎯 RESTful API design
      
      ## Architecture
      - **Clean Architecture** with proper layer separation
      - **SOLID Principles** implementation
      - **Domain-Driven Design** with rich domain models
      - **Event-Driven Architecture** for extensibility
      
      ## Security
      - BCrypt password hashing (12 salt rounds)
      - JWT access & refresh tokens
      - Rate limiting (global and auth-specific)
      - Account locking after failed attempts
      - Comprehensive input validation
      - CORS and security headers
      
      ## Error Handling
      All errors follow a standardized format with:
      - Consistent error codes
      - Detailed error messages
      - Correlation IDs for tracing
      - Proper HTTP status codes
    `,
    contact: {
      name: 'X-Form Development Team',
      email: 'dev@xform.com',
      url: 'https://xform.com',
    },
    license: {
      name: 'MIT',
      url: 'https://opensource.org/licenses/MIT',
    },
    termsOfService: 'https://xform.com/terms',
  },
  servers: [
    {
      url: 'http://localhost:3001',
      description: 'Development server',
    },
    {
      url: 'https://auth-dev.xform.com',
      description: 'Development environment',
    },
    {
      url: 'https://auth-staging.xform.com',
      description: 'Staging environment',
    },
    {
      url: 'https://auth.xform.com',
      description: 'Production environment',
    },
  ],
  tags: [
    {
      name: 'Authentication',
      description: 'User authentication and token management',
    },
    {
      name: 'User Management',
      description: 'User registration and profile management',
    },
    {
      name: 'Email Verification',
      description: 'Email verification and related operations',
    },
    {
      name: 'Password Management',
      description: 'Password reset and change operations',
    },
    {
      name: 'Health & Monitoring',
      description: 'Service health checks and monitoring endpoints',
    },
  ],
  components: {
    securitySchemes: {
      BearerAuth: {
        type: 'http',
        scheme: 'bearer',
        bearerFormat: 'JWT',
        description: 'JWT access token for authentication',
      },
      ApiKeyAuth: {
        type: 'apiKey',
        in: 'header',
        name: 'X-API-Key',
        description: 'API key for service-to-service authentication',
      },
    },
    schemas: {
      // Base response schemas
      SuccessResponse: {
        type: 'object',
        properties: {
          success: { type: 'boolean', example: true },
          timestamp: { type: 'string', format: 'date-time' },
          path: { type: 'string', example: '/api/v1/auth/login' },
          method: { type: 'string', example: 'POST' },
          correlationId: { type: 'string', format: 'uuid' },
          data: { type: 'object' },
          meta: {
            type: 'object',
            properties: {
              version: { type: 'string', example: 'v1' },
              rateLimit: {
                type: 'object',
                properties: {
                  limit: { type: 'integer', example: 100 },
                  remaining: { type: 'integer', example: 95 },
                  resetTime: { type: 'string', format: 'date-time' },
                },
              },
            },
          },
        },
      },
      ErrorResponse: {
        type: 'object',
        properties: {
          success: { type: 'boolean', example: false },
          timestamp: { type: 'string', format: 'date-time' },
          path: { type: 'string', example: '/api/v1/auth/login' },
          method: { type: 'string', example: 'POST' },
          correlationId: { type: 'string', format: 'uuid' },
          error: {
            type: 'object',
            properties: {
              code: { type: 'string', example: 'VALIDATION_ERROR' },
              message: { type: 'string', example: 'Validation failed' },
              timestamp: { type: 'string', format: 'date-time' },
              path: { type: 'string', example: '/api/v1/auth/login' },
              correlationId: { type: 'string', format: 'uuid' },
              details: {
                oneOf: [
                  { type: 'array', items: { $ref: '#/components/schemas/ValidationError' } },
                  { type: 'object' },
                ],
              },
            },
          },
        },
      },
      ValidationError: {
        type: 'object',
        properties: {
          field: { type: 'string', example: 'email' },
          message: { type: 'string', example: 'Must be a valid email address' },
          value: { type: 'string', example: 'invalid-email' },
          code: { type: 'string', example: 'INVALID_EMAIL' },
        },
      },
      // Request schemas
      RegisterRequest: {
        type: 'object',
        required: ['email', 'username', 'password', 'confirmPassword', 'firstName', 'lastName', 'acceptTerms'],
        properties: {
          email: { type: 'string', format: 'email', example: 'john.doe@example.com' },
          username: { type: 'string', minLength: 3, maxLength: 30, example: 'johndoe' },
          password: { type: 'string', minLength: 8, maxLength: 128, example: 'SecurePass123!' },
          confirmPassword: { type: 'string', minLength: 8, maxLength: 128, example: 'SecurePass123!' },
          firstName: { type: 'string', maxLength: 50, example: 'John' },
          lastName: { type: 'string', maxLength: 50, example: 'Doe' },
          acceptTerms: { type: 'boolean', example: true },
          marketingConsent: { type: 'boolean', example: false },
          referralCode: { type: 'string', minLength: 6, maxLength: 20, example: 'REF123' },
        },
      },
      LoginRequest: {
        type: 'object',
        required: ['email', 'password'],
        properties: {
          email: { type: 'string', format: 'email', example: 'john.doe@example.com' },
          password: { type: 'string', maxLength: 128, example: 'SecurePass123!' },
          rememberMe: { type: 'boolean', example: false },
          deviceId: { type: 'string', format: 'uuid' },
          deviceName: { type: 'string', maxLength: 100, example: 'iPhone 13' },
        },
      },
      RefreshTokenRequest: {
        type: 'object',
        required: ['refreshToken'],
        properties: {
          refreshToken: { type: 'string', example: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...' },
          deviceId: { type: 'string', format: 'uuid' },
        },
      },
      // Response schemas
      AuthResponse: {
        type: 'object',
        properties: {
          accessToken: { type: 'string', example: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...' },
          refreshToken: { type: 'string', example: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...' },
          tokenType: { type: 'string', example: 'Bearer' },
          expiresIn: { type: 'integer', example: 900 },
          expiresAt: { type: 'string', format: 'date-time' },
          scope: { type: 'array', items: { type: 'string' }, example: ['read', 'write'] },
          user: { $ref: '#/components/schemas/UserProfile' },
        },
      },
      UserProfile: {
        type: 'object',
        properties: {
          id: { type: 'string', format: 'uuid' },
          email: { type: 'string', format: 'email', example: 'john.doe@example.com' },
          username: { type: 'string', example: 'johndoe' },
          firstName: { type: 'string', example: 'John' },
          lastName: { type: 'string', example: 'Doe' },
          fullName: { type: 'string', example: 'John Doe' },
          role: { type: 'string', enum: ['USER', 'ADMIN', 'MODERATOR'], example: 'USER' },
          emailVerified: { type: 'boolean', example: true },
          phoneVerified: { type: 'boolean', example: false },
          accountStatus: { type: 'string', enum: ['ACTIVE', 'SUSPENDED', 'PENDING_VERIFICATION', 'LOCKED'], example: 'ACTIVE' },
          lastLoginAt: { type: 'string', format: 'date-time', nullable: true },
          createdAt: { type: 'string', format: 'date-time' },
          updatedAt: { type: 'string', format: 'date-time' },
          preferences: { $ref: '#/components/schemas/UserPreferences' },
          metadata: { type: 'object', additionalProperties: true },
        },
      },
      UserPreferences: {
        type: 'object',
        properties: {
          language: { type: 'string', example: 'en' },
          timezone: { type: 'string', example: 'UTC' },
          emailNotifications: { type: 'boolean', example: true },
          smsNotifications: { type: 'boolean', example: false },
          marketingEmails: { type: 'boolean', example: false },
          twoFactorEnabled: { type: 'boolean', example: false },
        },
      },
      HealthCheck: {
        type: 'object',
        properties: {
          service: { type: 'string', example: 'auth-service' },
          version: { type: 'string', example: '1.0.0' },
          status: { type: 'string', enum: ['HEALTHY', 'UNHEALTHY', 'DEGRADED'], example: 'HEALTHY' },
          uptime: { type: 'number', example: 3600.5 },
          timestamp: { type: 'string', format: 'date-time' },
          environment: { type: 'string', example: 'development' },
          dependencies: {
            type: 'array',
            items: { $ref: '#/components/schemas/DependencyHealth' },
          },
          metrics: { $ref: '#/components/schemas/ServiceMetrics' },
        },
      },
      DependencyHealth: {
        type: 'object',
        properties: {
          name: { type: 'string', example: 'postgresql' },
          type: { type: 'string', enum: ['DATABASE', 'CACHE', 'EMAIL', 'EXTERNAL_API'], example: 'DATABASE' },
          status: { type: 'string', enum: ['HEALTHY', 'UNHEALTHY', 'DEGRADED'], example: 'HEALTHY' },
          responseTime: { type: 'number', example: 25.5 },
          lastChecked: { type: 'string', format: 'date-time' },
          error: { type: 'string', nullable: true },
        },
      },
      ServiceMetrics: {
        type: 'object',
        properties: {
          requestCount: { type: 'integer', example: 1500 },
          errorRate: { type: 'number', example: 0.02 },
          averageResponseTime: { type: 'number', example: 120.5 },
          activeConnections: { type: 'integer', example: 15 },
          memoryUsage: {
            type: 'object',
            properties: {
              used: { type: 'number', example: 256 },
              free: { type: 'number', example: 768 },
              total: { type: 'number', example: 1024 },
              percentage: { type: 'number', example: 25.0 },
            },
          },
          cpuUsage: { type: 'number', example: 15.5 },
        },
      },
    },
    responses: {
      '400': {
        description: 'Bad Request - Validation Error',
        content: {
          'application/json': {
            schema: { $ref: '#/components/schemas/ErrorResponse' },
            example: {
              success: false,
              timestamp: '2024-01-01T12:00:00.000Z',
              path: '/api/v1/auth/register',
              method: 'POST',
              correlationId: '550e8400-e29b-41d4-a716-446655440000',
              error: {
                code: 'VALIDATION_ERROR',
                message: 'Validation failed',
                timestamp: '2024-01-01T12:00:00.000Z',
                path: '/api/v1/auth/register',
                correlationId: '550e8400-e29b-41d4-a716-446655440000',
                details: [
                  {
                    field: 'email',
                    message: 'Must be a valid email address',
                    value: 'invalid-email',
                    code: 'INVALID_EMAIL',
                  },
                ],
              },
            },
          },
        },
      },
      '401': {
        description: 'Unauthorized - Authentication Required',
        content: {
          'application/json': {
            schema: { $ref: '#/components/schemas/ErrorResponse' },
          },
        },
      },
      '403': {
        description: 'Forbidden - Insufficient Permissions',
        content: {
          'application/json': {
            schema: { $ref: '#/components/schemas/ErrorResponse' },
          },
        },
      },
      '404': {
        description: 'Not Found - Resource Not Found',
        content: {
          'application/json': {
            schema: { $ref: '#/components/schemas/ErrorResponse' },
          },
        },
      },
      '429': {
        description: 'Too Many Requests - Rate Limit Exceeded',
        content: {
          'application/json': {
            schema: { $ref: '#/components/schemas/ErrorResponse' },
          },
        },
      },
      '500': {
        description: 'Internal Server Error',
        content: {
          'application/json': {
            schema: { $ref: '#/components/schemas/ErrorResponse' },
          },
        },
      },
    },
    parameters: {
      CorrelationId: {
        name: 'X-Correlation-ID',
        in: 'header',
        description: 'Unique identifier for request tracing',
        schema: { type: 'string', format: 'uuid' },
      },
      ApiVersion: {
        name: 'version',
        in: 'path',
        required: true,
        description: 'API version',
        schema: { type: 'string', enum: ['v1'], default: 'v1' },
      },
    },
  },
  security: [
    { BearerAuth: [] },
  ],
};

const options = {
  definition: swaggerDefinition,
  apis: [
    './src/interface/http/*.ts',
    './src/interface/http/routes/*.ts',
    './docs/api/*.yaml',
  ],
};

export const swaggerSpec = swaggerJSDoc(options);

// Helper function to generate swagger HTML with custom styling
export const getSwaggerHTML = (swaggerSpec: any): string => {
  return `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>X-Form Auth Service API Documentation</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css" />
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
      margin:0;
      background: #fafafa;
    }
    .swagger-ui .topbar {
      background-color: #2c3e50;
    }
    .swagger-ui .topbar .download-url-wrapper .select-label {
      color: #ffffff;
    }
    .swagger-ui .info .title {
      color: #2c3e50;
    }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        spec: ${JSON.stringify(swaggerSpec)},
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
        }
      });
    };
  </script>
</body>
</html>
  `;
};
