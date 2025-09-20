/**
 * Enhanced Swagger/OpenAPI Documentation Configuration
 * Professional-grade API documentation following current industry best practices
 */

const swaggerJsdoc = require('swagger-jsdoc');
const swaggerUi = require('swagger-ui-express');
const config = require('./enhanced');

// Enhanced OpenAPI 3.0 specification
const swaggerDefinition = {
  openapi: '3.0.3',
  info: {
    title: '🚀 X-Form Response Service API',
    version: config.get('server.version') || '1.0.0',
    description: `
# X-Form Response Service API

A comprehensive microservice for managing form responses with advanced features including:

## 🌟 Key Features

- **Response Management**: Complete CRUD operations for form responses
- **File Upload Support**: Handle multiple file types with validation
- **Real-time Analytics**: Live analytics and reporting capabilities  
- **Advanced Validation**: Multi-layer validation with custom rules
- **Security Features**: JWT authentication, rate limiting, and data sanitization
- **Export Capabilities**: Export responses in multiple formats (JSON, CSV, Excel, PDF)
- **Webhook Integration**: Real-time notifications and integrations

## 🔐 Security Features

- JWT-based authentication with refresh tokens
- API key support for service-to-service communication
- Rate limiting and DDoS protection
- Input validation and sanitization
- CORS and security headers
- Request/response encryption

## 📊 Analytics & Reporting

- Real-time response analytics and metrics
- Completion rate tracking and analysis
- Response time and performance analytics
- Geographic distribution analysis
- Device and browser analytics
- Custom dashboard creation

## 📁 File Management

- Multiple file upload support with drag & drop
- Image processing and optimization
- File type validation and virus scanning
- Secure file storage with CDN integration
- File compression and thumbnail generation

## 🔄 Integration Capabilities

- Webhook notifications with retry logic
- Real-time events via Socket.IO
- External service integrations (Zapier, Slack, etc.)
- Message queuing with Kafka/RabbitMQ
- Redis caching for performance optimization

## 📖 API Standards

This API follows REST principles and includes:
- Consistent error handling with detailed error codes
- Proper HTTP status codes and response formats
- Comprehensive input validation with sanitization
- Rate limiting with different tiers
- Detailed logging and monitoring
- Health check endpoints for monitoring
- OpenAPI 3.0 specification compliance
    `,
    contact: {
      name: 'X-Form API Support Team',
      email: 'api-support@x-form.com',
      url: 'https://docs.x-form.com/support'
    },
    license: {
      name: 'MIT License',
      url: 'https://opensource.org/licenses/MIT'
    },
    termsOfService: 'https://x-form.com/terms-of-service'
  },
  externalDocs: {
    description: 'Complete API Documentation',
    url: 'https://docs.x-form.com/response-service'
  },
  servers: [
    {
      url: `http://localhost:${config.get('server.port') || 3002}/api/v1`,
      description: '🏠 Development server (localhost)'
    },
    {
      url: 'https://api.x-form.com/response-service/api/v1',
      description: '🌐 Production server'
    },
    {
      url: 'https://staging-api.x-form.com/response-service/api/v1', 
      description: '🧪 Staging server'
    },
    {
      url: 'https://dev-api.x-form.com/response-service/api/v1',
      description: '🔧 Development server (shared)'
    }
  ],
  tags: [
    {
      name: 'Health',
      description: '💓 Service health and monitoring endpoints',
      externalDocs: {
        description: 'Health Check Best Practices',
        url: 'https://docs.x-form.com/health-checks'
      }
    },
    {
      name: 'Responses',
      description: '📝 Form response management operations',
      externalDocs: {
        description: 'Response Management Guide',
        url: 'https://docs.x-form.com/responses'
      }
    },
    {
      name: 'Analytics', 
      description: '📊 Response analytics and reporting',
      externalDocs: {
        description: 'Analytics Documentation',
        url: 'https://docs.x-form.com/analytics'
      }
    },
    {
      name: 'Files',
      description: '📁 File upload and management',
      externalDocs: {
        description: 'File Upload Guide',
        url: 'https://docs.x-form.com/file-uploads'
      }
    },
    {
      name: 'Export',
      description: '📤 Data export functionality',
      externalDocs: {
        description: 'Export Documentation',
        url: 'https://docs.x-form.com/exports'
      }
    },
    {
      name: 'Webhooks',
      description: '🔗 Webhook management and notifications',
      externalDocs: {
        description: 'Webhook Integration Guide',
        url: 'https://docs.x-form.com/webhooks'
      }
    }
  ],
  paths: {},
  components: {
    securitySchemes: {
      BearerAuth: {
        type: 'http',
        scheme: 'bearer',
        bearerFormat: 'JWT',
        description: 'JWT token obtained from authentication service. Format: `Authorization: Bearer <token>`'
      },
      ApiKeyAuth: {
        type: 'apiKey',
        in: 'header',
        name: 'X-API-Key',
        description: 'API key for service-to-service authentication'
      },
      BasicAuth: {
        type: 'http',
        scheme: 'basic',
        description: 'Basic authentication for development and testing'
      }
    },
    parameters: {
      FormIdParam: {
        name: 'formId',
        in: 'path',
        required: true,
        description: 'Unique identifier for the form',
        schema: {
          type: 'string',
          pattern: '^[a-zA-Z0-9_-]+$',
          minLength: 5,
          maxLength: 50
        },
        example: 'form_user_feedback_2023'
      },
      ResponseIdParam: {
        name: 'responseId', 
        in: 'path',
        required: true,
        description: 'Unique identifier for the response',
        schema: {
          type: 'string',
          pattern: '^resp_[a-zA-Z0-9_-]+$',
          minLength: 8,
          maxLength: 50
        },
        example: 'resp_1234567890abcdef'
      },
      PaginationLimit: {
        name: 'limit',
        in: 'query',
        description: 'Maximum number of items to return',
        schema: {
          type: 'integer',
          minimum: 1,
          maximum: 1000,
          default: 50
        }
      },
      PaginationOffset: {
        name: 'offset',
        in: 'query',
        description: 'Number of items to skip',
        schema: {
          type: 'integer',
          minimum: 0,
          default: 0
        }
      },
      SortBy: {
        name: 'sortBy',
        in: 'query',
        description: 'Field to sort by',
        schema: {
          type: 'string',
          enum: ['createdAt', 'updatedAt', 'submittedAt', 'status'],
          default: 'createdAt'
        }
      },
      SortOrder: {
        name: 'sortOrder',
        in: 'query',
        description: 'Sort order',
        schema: {
          type: 'string',
          enum: ['asc', 'desc'],
          default: 'desc'
        }
      },
      CorrelationId: {
        name: 'X-Correlation-ID',
        in: 'header',
        description: 'Unique request identifier for tracing',
        required: false,
        schema: {
          type: 'string',
          pattern: '^cor_[a-zA-Z0-9_-]+$'
        },
        example: 'cor_1234567890abcdef'
      }
    },
    schemas: {
      // =====================================================
      // CORE RESPONSE SCHEMAS  
      // =====================================================
      
      Response: {
        type: 'object',
        required: ['formId', 'responses'],
        properties: {
          id: {
            type: 'string',
            description: 'Unique response identifier',
            pattern: '^resp_[a-zA-Z0-9_-]+$',
            example: 'resp_1234567890abcdef'
          },
          formId: {
            type: 'string',
            description: 'Associated form identifier',
            pattern: '^form_[a-zA-Z0-9_-]+$',
            example: 'form_user_feedback_2023'
          },
          formTitle: {
            type: 'string',
            description: 'Title of the associated form',
            maxLength: 200,
            example: 'Customer Feedback Survey 2023'
          },
          respondentEmail: {
            type: 'string',
            format: 'email',
            description: 'Email of the respondent (optional)',
            example: 'customer@example.com'
          },
          respondentName: {
            type: 'string',
            description: 'Name of the respondent (optional)',
            maxLength: 100,
            example: 'John Doe'
          },
          status: {
            type: 'string',
            enum: ['draft', 'partial', 'completed', 'archived', 'deleted'],
            description: 'Current status of the response',
            example: 'completed'
          },
          priority: {
            type: 'string',
            enum: ['low', 'normal', 'high', 'urgent'],
            description: 'Priority level of the response',
            default: 'normal',
            example: 'normal'
          },
          responses: {
            type: 'array',
            description: 'Array of question responses',
            items: {
              $ref: '#/components/schemas/QuestionResponse'
            },
            minItems: 1
          },
          metadata: {
            $ref: '#/components/schemas/ResponseMetadata'
          },
          tags: {
            type: 'array',
            description: 'Tags for categorization',
            items: {
              type: 'string',
              maxLength: 50
            },
            example: ['feedback', 'customer-service', 'product']
          },
          submittedAt: {
            type: 'string',
            format: 'date-time',
            description: 'Timestamp when response was submitted',
            example: '2023-12-07T10:30:00.000Z'
          },
          createdAt: {
            type: 'string',
            format: 'date-time',
            description: 'Timestamp when response was created',
            example: '2023-12-07T10:25:00.000Z'
          },
          updatedAt: {
            type: 'string',
            format: 'date-time',
            description: 'Timestamp when response was last updated',
            example: '2023-12-07T10:35:00.000Z'
          },
          version: {
            type: 'integer',
            description: 'Version number for optimistic locking',
            minimum: 1,
            example: 1
          }
        }
      },

      CreateResponseRequest: {
        type: 'object',
        required: ['formId', 'responses'],
        properties: {
          formId: {
            type: 'string',
            description: 'Associated form identifier',
            pattern: '^form_[a-zA-Z0-9_-]+$',
            example: 'form_user_feedback_2023'
          },
          respondentEmail: {
            type: 'string',
            format: 'email',
            description: 'Email of the respondent (optional)',
            example: 'customer@example.com'
          },
          respondentName: {
            type: 'string',
            description: 'Name of the respondent (optional)',
            maxLength: 100,
            example: 'John Doe'
          },
          status: {
            type: 'string',
            enum: ['draft', 'partial', 'completed'],
            description: 'Initial status of the response',
            default: 'draft',
            example: 'completed'
          },
          responses: {
            type: 'array',
            description: 'Array of question responses',
            items: {
              $ref: '#/components/schemas/QuestionResponseInput'
            },
            minItems: 1
          },
          tags: {
            type: 'array',
            description: 'Tags for categorization',
            items: {
              type: 'string',
              maxLength: 50
            },
            example: ['feedback', 'customer-service']
          },
          metadata: {
            $ref: '#/components/schemas/ResponseMetadataInput'
          }
        }
      },

      UpdateResponseRequest: {
        type: 'object',
        properties: {
          status: {
            type: 'string',
            enum: ['draft', 'partial', 'completed', 'archived'],
            description: 'Updated status of the response',
            example: 'completed'
          },
          responses: {
            type: 'array',
            description: 'Updated question responses',
            items: {
              $ref: '#/components/schemas/QuestionResponseInput'
            }
          },
          tags: {
            type: 'array',
            description: 'Updated tags',
            items: {
              type: 'string',
              maxLength: 50
            }
          },
          metadata: {
            $ref: '#/components/schemas/ResponseMetadataInput'
          },
          version: {
            type: 'integer',
            description: 'Current version for optimistic locking',
            minimum: 1,
            example: 1
          }
        }
      },

      QuestionResponse: {
        type: 'object',
        required: ['questionId', 'questionType'],
        properties: {
          questionId: {
            type: 'string',
            description: 'Unique question identifier',
            pattern: '^q_[a-zA-Z0-9_-]+$',
            example: 'q_rating'
          },
          questionType: {
            type: 'string',
            enum: ['text', 'textarea', 'radio', 'checkbox', 'select', 'multiselect', 'number', 'email', 'url', 'tel', 'date', 'datetime', 'time', 'file', 'image', 'rating', 'slider', 'boolean'],
            description: 'Type of the question',
            example: 'radio'
          },
          label: {
            type: 'string',
            description: 'Question label/text',
            maxLength: 500,
            example: 'How would you rate our service?'
          },
          value: {
            oneOf: [
              { type: 'string' },
              { type: 'number' },
              { type: 'boolean' },
              { type: 'array', items: { type: 'string' } },
              { type: 'object' }
            ],
            description: 'Response value (type depends on question type)',
            example: 'Excellent'
          },
          files: {
            type: 'array',
            description: 'Uploaded files for file-type questions',
            items: {
              $ref: '#/components/schemas/FileUpload'
            }
          },
          validationErrors: {
            type: 'array',
            description: 'Validation errors for this response',
            items: {
              type: 'object',
              properties: {
                code: { type: 'string', example: 'REQUIRED_FIELD' },
                message: { type: 'string', example: 'This field is required' }
              }
            }
          },
          answeredAt: {
            type: 'string',
            format: 'date-time',
            description: 'When this question was answered',
            example: '2023-12-07T10:28:00.000Z'
          }
        }
      },

      QuestionResponseInput: {
        type: 'object',
        required: ['questionId', 'questionType'],
        properties: {
          questionId: {
            type: 'string',
            description: 'Unique question identifier',
            pattern: '^q_[a-zA-Z0-9_-]+$',
            example: 'q_rating'
          },
          questionType: {
            type: 'string',
            enum: ['text', 'textarea', 'radio', 'checkbox', 'select', 'multiselect', 'number', 'email', 'url', 'tel', 'date', 'datetime', 'time', 'file', 'image', 'rating', 'slider', 'boolean'],
            description: 'Type of the question',
            example: 'radio'
          },
          value: {
            oneOf: [
              { type: 'string' },
              { type: 'number' },
              { type: 'boolean' },
              { type: 'array', items: { type: 'string' } },
              { type: 'object' }
            ],
            description: 'Response value',
            example: 'Excellent'
          },
          files: {
            type: 'array',
            description: 'File uploads for file-type questions',
            items: {
              type: 'string',
              description: 'File upload ID',
              example: 'upload_1234567890'
            }
          }
        }
      },

      // =====================================================
      // METADATA & FILE SCHEMAS
      // =====================================================

      ResponseMetadata: {
        type: 'object',
        properties: {
          userAgent: {
            type: 'string',
            description: 'Browser user agent string',
            maxLength: 1000,
            example: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
          },
          ipAddress: {
            type: 'string',
            description: 'IP address of the respondent (anonymized)',
            pattern: '^(?:[0-9]{1,3}\\.){3}[0-9]{1,3}$|^[0-9a-fA-F:]+$',
            example: '192.168.1.***'
          },
          referrer: {
            type: 'string',
            format: 'uri',
            description: 'Referring URL',
            maxLength: 2000,
            example: 'https://example.com/survey-page'
          },
          timeSpent: {
            type: 'number',
            description: 'Time spent filling the form (in seconds)',
            minimum: 0,
            example: 180
          },
          deviceInfo: {
            type: 'object',
            properties: {
              type: {
                type: 'string',
                enum: ['desktop', 'mobile', 'tablet', 'unknown'],
                example: 'desktop'
              },
              os: {
                type: 'string',
                maxLength: 100,
                example: 'Windows 10'
              },
              browser: {
                type: 'string',
                maxLength: 100,
                example: 'Chrome 120.0.0.0'
              },
              screenResolution: {
                type: 'string',
                pattern: '^\\d+x\\d+$',
                example: '1920x1080'
              }
            }
          },
          location: {
            type: 'object',
            properties: {
              country: { 
                type: 'string', 
                maxLength: 100,
                example: 'United States' 
              },
              countryCode: {
                type: 'string',
                pattern: '^[A-Z]{2}$',
                example: 'US'
              },
              region: { 
                type: 'string', 
                maxLength: 100,
                example: 'New York' 
              },
              city: { 
                type: 'string', 
                maxLength: 100,
                example: 'New York City' 
              },
              timezone: {
                type: 'string',
                maxLength: 50,
                example: 'America/New_York'
              }
            }
          }
        }
      },

      ResponseMetadataInput: {
        type: 'object',
        properties: {
          timeSpent: {
            type: 'number',
            description: 'Time spent filling the form (in seconds)',
            minimum: 0,
            example: 180
          },
          referrer: {
            type: 'string',
            format: 'uri',
            description: 'Referring URL',
            maxLength: 2000,
            example: 'https://example.com/survey-page'
          },
          sessionId: {
            type: 'string',
            description: 'User session identifier',
            example: 'sess_1234567890abcdef'
          }
        }
      },

      FileUpload: {
        type: 'object',
        properties: {
          id: {
            type: 'string',
            description: 'File upload identifier',
            example: 'upload_1234567890'
          },
          fileName: {
            type: 'string',
            description: 'Original file name',
            maxLength: 255,
            example: 'customer_feedback_document.pdf'
          },
          fileSize: {
            type: 'number',
            description: 'File size in bytes',
            minimum: 0,
            maximum: 104857600,
            example: 1024000
          },
          mimeType: {
            type: 'string',
            description: 'MIME type of the file',
            example: 'application/pdf'
          },
          fileUrl: {
            type: 'string',
            format: 'uri',
            description: 'URL to access the uploaded file',
            example: 'https://storage.x-form.com/files/uuid-filename.pdf'
          },
          thumbnailUrl: {
            type: 'string',
            format: 'uri',
            description: 'URL to file thumbnail (for images)',
            example: 'https://storage.x-form.com/thumbnails/uuid-filename.jpg'
          },
          uploadedAt: {
            type: 'string',
            format: 'date-time',
            description: 'Timestamp when file was uploaded',
            example: '2023-12-07T10:30:00.000Z'
          },
          virusScanned: {
            type: 'boolean',
            description: 'Whether file has been virus scanned',
            example: true
          },
          scanResult: {
            type: 'string',
            enum: ['clean', 'infected', 'pending'],
            description: 'Virus scan result',
            example: 'clean'
          }
        }
      },

      // =====================================================
      // ANALYTICS SCHEMAS
      // =====================================================

      Analytics: {
        type: 'object',
        properties: {
          formId: {
            type: 'string',
            description: 'Form identifier',
            example: 'form_user_feedback_2023'
          },
          totalResponses: {
            type: 'number',
            description: 'Total number of responses',
            minimum: 0,
            example: 150
          },
          completedResponses: {
            type: 'number',
            description: 'Number of completed responses',
            minimum: 0,
            example: 120
          },
          draftResponses: {
            type: 'number',
            description: 'Number of draft responses',
            minimum: 0,
            example: 20
          },
          partialResponses: {
            type: 'number',
            description: 'Number of partial responses',
            minimum: 0,
            example: 10
          },
          completionRate: {
            type: 'number',
            description: 'Completion rate percentage',
            minimum: 0,
            maximum: 100,
            example: 80.0
          },
          averageCompletionTime: {
            type: 'number',
            description: 'Average time to complete in seconds',
            minimum: 0,
            example: 240
          },
          responsesByDate: {
            type: 'array',
            description: 'Responses grouped by date',
            items: {
              type: 'object',
              properties: {
                date: { type: 'string', format: 'date', example: '2023-12-07' },
                count: { type: 'number', minimum: 0, example: 15 },
                completed: { type: 'number', minimum: 0, example: 12 },
                draft: { type: 'number', minimum: 0, example: 2 },
                partial: { type: 'number', minimum: 0, example: 1 }
              }
            }
          },
          questionAnalytics: {
            type: 'array',
            description: 'Analytics for individual questions',
            items: {
              $ref: '#/components/schemas/QuestionAnalytics'
            }
          },
          deviceDistribution: {
            type: 'object',
            description: 'Distribution by device type',
            properties: {
              desktop: { type: 'number', example: 60 },
              mobile: { type: 'number', example: 35 },
              tablet: { type: 'number', example: 5 }
            }
          },
          geographicDistribution: {
            type: 'array',
            description: 'Geographic distribution of responses',
            items: {
              type: 'object',
              properties: {
                country: { type: 'string', example: 'United States' },
                countryCode: { type: 'string', example: 'US' },
                count: { type: 'number', example: 45 },
                percentage: { type: 'number', example: 30.0 }
              }
            }
          }
        }
      },

      QuestionAnalytics: {
        type: 'object',
        properties: {
          questionId: {
            type: 'string',
            description: 'Question identifier',
            example: 'q_rating'
          },
          questionType: {
            type: 'string',
            description: 'Question type',
            example: 'radio'
          },
          questionText: {
            type: 'string',
            description: 'Question text/label',
            example: 'How would you rate our service?'
          },
          totalResponses: {
            type: 'number',
            description: 'Total responses to this question',
            minimum: 0,
            example: 100
          },
          responseRate: {
            type: 'number',
            description: 'Response rate percentage',
            minimum: 0,
            maximum: 100,
            example: 95.5
          },
          averageResponseTime: {
            type: 'number',
            description: 'Average time to answer in seconds',
            minimum: 0,
            example: 15.2
          },
          valueDistribution: {
            type: 'object',
            description: 'Distribution of response values',
            additionalProperties: {
              type: 'number'
            },
            example: {
              'Excellent': 45,
              'Good': 35,
              'Average': 15,
              'Poor': 5
            }
          },
          sentiment: {
            type: 'object',
            description: 'Sentiment analysis for text responses',
            properties: {
              positive: { type: 'number', example: 65.5 },
              neutral: { type: 'number', example: 25.0 },
              negative: { type: 'number', example: 9.5 }
            }
          }
        }
      },

      // =====================================================
      // SYSTEM SCHEMAS
      // =====================================================

      HealthResponse: {
        type: 'object',
        properties: {
          status: {
            type: 'string',
            enum: ['healthy', 'unhealthy', 'degraded'],
            description: 'Overall health status',
            example: 'healthy'
          },
          timestamp: {
            type: 'string',
            format: 'date-time',
            description: 'Health check timestamp',
            example: '2023-12-07T10:30:00.000Z'
          },
          uptime: {
            type: 'number',
            description: 'Service uptime in seconds',
            minimum: 0,
            example: 86400
          },
          version: {
            type: 'string',
            description: 'Service version',
            example: '1.0.0'
          },
          environment: {
            type: 'string',
            enum: ['development', 'staging', 'production'],
            description: 'Current environment',
            example: 'production'
          },
          dependencies: {
            type: 'object',
            description: 'Health status of dependencies',
            properties: {
              database: {
                type: 'object',
                properties: {
                  status: { type: 'string', enum: ['healthy', 'unhealthy'], example: 'healthy' },
                  responseTime: { type: 'number', example: 15 },
                  lastChecked: { type: 'string', format: 'date-time' }
                }
              },
              cache: {
                type: 'object',
                properties: {
                  status: { type: 'string', enum: ['healthy', 'unhealthy'], example: 'healthy' },
                  responseTime: { type: 'number', example: 5 },
                  lastChecked: { type: 'string', format: 'date-time' }
                }
              },
              formService: {
                type: 'object',
                properties: {
                  status: { type: 'string', enum: ['healthy', 'unhealthy'], example: 'healthy' },
                  responseTime: { type: 'number', example: 25 },
                  lastChecked: { type: 'string', format: 'date-time' }
                }
              }
            }
          },
          metrics: {
            type: 'object',
            properties: {
              memoryUsage: { type: 'number', example: 67.5 },
              cpuUsage: { type: 'number', example: 23.1 },
              requestsPerMinute: { type: 'number', example: 150 },
              errorRate: { type: 'number', example: 0.5 }
            }
          }
        }
      },

      ApiResponse: {
        type: 'object',
        properties: {
          success: {
            type: 'boolean',
            description: 'Indicates if the request was successful',
            example: true
          },
          message: {
            type: 'string',
            description: 'Human-readable message',
            example: 'Operation completed successfully'
          },
          data: {
            description: 'Response data (varies by endpoint)',
            oneOf: [
              { type: 'object' },
              { type: 'array' },
              { type: 'string' },
              { type: 'number' },
              { type: 'boolean' }
            ]
          },
          meta: {
            type: 'object',
            description: 'Metadata about the response',
            properties: {
              pagination: {
                type: 'object',
                properties: {
                  page: { type: 'number', example: 1 },
                  limit: { type: 'number', example: 50 },
                  total: { type: 'number', example: 150 },
                  totalPages: { type: 'number', example: 3 }
                }
              },
              timing: {
                type: 'object',
                properties: {
                  requestTime: { type: 'string', format: 'date-time' },
                  processingTime: { type: 'number', example: 0.045 }
                }
              }
            }
          },
          correlationId: {
            type: 'string',
            description: 'Request correlation ID for tracing',
            example: 'cor_1234567890abcdef'
          },
          timestamp: {
            type: 'string',
            format: 'date-time',
            description: 'Response timestamp',
            example: '2023-12-07T10:30:00.000Z'
          }
        }
      },

      ErrorResponse: {
        type: 'object',
        properties: {
          success: {
            type: 'boolean',
            description: 'Always false for error responses',
            example: false
          },
          error: {
            type: 'object',
            properties: {
              code: {
                type: 'string',
                description: 'Error code for programmatic handling',
                example: 'VALIDATION_ERROR'
              },
              message: {
                type: 'string',
                description: 'Human-readable error message',
                example: 'The provided data is invalid'
              },
              details: {
                type: 'array',
                description: 'Detailed error information',
                items: {
                  type: 'object',
                  properties: {
                    field: { type: 'string', example: 'email' },
                    message: { type: 'string', example: 'Invalid email format' },
                    code: { type: 'string', example: 'INVALID_FORMAT' },
                    value: { type: 'string', example: 'invalid-email' }
                  }
                }
              },
              stack: {
                type: 'string',
                description: 'Stack trace (development only)',
                example: 'Error: Validation failed\\n    at validateRequest...'
              }
            }
          },
          correlationId: {
            type: 'string',
            description: 'Request correlation ID for tracing',
            example: 'cor_1234567890abcdef'
          },
          timestamp: {
            type: 'string',
            format: 'date-time',
            description: 'Error timestamp',
            example: '2023-12-07T10:30:00.000Z'
          },
          requestId: {
            type: 'string',
            description: 'Unique request identifier',
            example: 'req_1234567890abcdef'
          }
        }
      },

      PaginatedResponse: {
        type: 'object',
        properties: {
          data: {
            type: 'array',
            description: 'Array of response items',
            items: {
              $ref: '#/components/schemas/Response'
            }
          },
          pagination: {
            type: 'object',
            properties: {
              page: {
                type: 'number',
                description: 'Current page number',
                minimum: 1,
                example: 1
              },
              limit: {
                type: 'number',
                description: 'Items per page',
                minimum: 1,
                maximum: 1000,
                example: 50
              },
              total: {
                type: 'number',
                description: 'Total number of items',
                minimum: 0,
                example: 150
              },
              totalPages: {
                type: 'number',
                description: 'Total number of pages',
                minimum: 0,
                example: 3
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
          }
        }
      }
    }
  }
};

// Configure swagger-jsdoc options
const swaggerOptions = {
  definition: swaggerDefinition,
  apis: [
    './src/routes/v1/*.js',
    './src/controllers/*.js',
    './src/models/*.js',
    './src/index.js'
  ]
};

// Generate OpenAPI specification
const specs = swaggerJsdoc(swaggerOptions);

// Enhanced Swagger UI options with professional styling
const swaggerUiOptions = {
  customCss: `
    .swagger-ui .topbar { display: none }
    .swagger-ui .info { margin: 20px 0 }
    .swagger-ui .info .title { 
      color: #2563eb; 
      font-size: 2.5rem; 
      font-weight: bold;
    }
    .swagger-ui .info .description { 
      font-size: 1.1rem; 
      line-height: 1.6;
      color: #374151;
    }
    .swagger-ui .scheme-container { 
      background: #f8fafc; 
      padding: 20px;
      border-radius: 8px;
      margin: 20px 0;
    }
    .swagger-ui .opblock { 
      border-radius: 8px;
      margin-bottom: 15px;
      box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    }
    .swagger-ui .opblock.opblock-get .opblock-summary {
      background: linear-gradient(90deg, #059669, #047857);
      color: white;
    }
    .swagger-ui .opblock.opblock-post .opblock-summary {
      background: linear-gradient(90deg, #2563eb, #1d4ed8);
      color: white;
    }
    .swagger-ui .opblock.opblock-put .opblock-summary {
      background: linear-gradient(90deg, #d97706, #b45309);
      color: white;
    }
    .swagger-ui .opblock.opblock-delete .opblock-summary {
      background: linear-gradient(90deg, #dc2626, #b91c1c);
      color: white;
    }
    .swagger-ui .opblock .opblock-summary-description {
      font-weight: 500;
    }
    .swagger-ui .btn.authorize {
      background: #2563eb;
      color: white;
      border: none;
      padding: 10px 20px;
      border-radius: 6px;
      font-weight: 500;
    }
    .swagger-ui .btn.authorize:hover {
      background: #1d4ed8;
    }
    .swagger-ui .servers select {
      background: #f3f4f6;
      border: 1px solid #d1d5db;
      border-radius: 6px;
      padding: 8px 12px;
    }
    .swagger-ui .response-col_status {
      font-weight: bold;
    }
    .swagger-ui .response.valid {
      background: #f0f9ff;
      border-left: 4px solid #059669;
    }
    .swagger-ui .response.error {
      background: #fef2f2;
      border-left: 4px solid #dc2626;
    }
  `,
  customSiteTitle: 'X-Form Response Service API Documentation',
  customfavIcon: '/favicon.ico',
  swaggerOptions: {
    persistAuthorization: true,
    displayRequestDuration: true,
    docExpansion: 'list',
    filter: true,
    showExtensions: true,
    showCommonExtensions: true,
    tryItOutEnabled: true,
    defaultModelsExpandDepth: 2,
    defaultModelExpandDepth: 2,
    displayOperationId: false,
    operationsSorter: 'alpha',
    tagsSorter: 'alpha'
  }
};

module.exports = {
  specs,
  swaggerUi,
  swaggerUiOptions,
  swaggerDefinition
};
