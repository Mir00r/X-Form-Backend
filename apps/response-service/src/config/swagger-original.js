/**
 * Swagger/OpenAPI Documentation Configuration
 * Enhanced with current industry best practices for Response Service
 */

const swaggerJsdoc = require('swagger-jsdoc');
const swaggerUi = require('swagger-ui-express');
const config = require('./enhanced');

// OpenAPI 3.0 specification with enhanced features
const swaggerDefinition = {
  openapi: '3.0.3',
  info: {
    title: 'X-Form Response Service API',
    version: config.get('server.version') || '1.0.0',
    description: `
# X-Form Response Service API

A comprehensive microservice for managing form responses with advanced features including:

- **Response Management**: Complete CRUD operations for form responses
- **File Upload Support**: Handle multiple file types with validation
- **Real-time Analytics**: Live analytics and reporting capabilities  
- **Advanced Validation**: Multi-layer validation with custom rules
- **Security Features**: JWT authentication, rate limiting, and data sanitization
- **Export Capabilities**: Export responses in multiple formats (JSON, CSV, Excel, PDF)
- **Webhook Integration**: Real-time notifications and integrations

## Features

### 🔐 Security
- JWT-based authentication
- API key support for service-to-service communication
- Rate limiting and DDoS protection
- Input validation and sanitization
- CORS and security headers

### 📊 Analytics
- Real-time response analytics
- Completion rate tracking
- Response time analysis
- Geographic distribution
- Device and browser analytics

### 📁 File Handling
- Multiple file upload support
- Image processing and optimization
- File type validation
- Secure file storage
- Virus scanning integration

### 🔄 Integrations
- Webhook notifications
- Real-time events via Socket.IO
- External service integrations
- Kafka message queuing
- Redis caching

## API Standards

This API follows REST principles and includes:
- Consistent error handling
- Proper HTTP status codes
- Comprehensive input validation
- Rate limiting
- Detailed logging
- Health check endpoints
    `,
    contact: {
      name: 'X-Form API Support',
      email: 'api-support@x-form.com',
      url: 'https://docs.x-form.com/support'
    },
    license: {
      name: 'MIT License',
      url: 'https://opensource.org/licenses/MIT'
    },
    termsOfService: 'https://x-form.com/terms-of-service'
  },
  servers: [
    {
      url: `http://localhost:${config.get('server.port') || 3002}/api/v1`,
      description: 'Development server (localhost)'
    },
    {
      url: 'https://api.x-form.com/response-service/api/v1',
      description: 'Production server'
    },
    {
      url: 'https://staging-api.x-form.com/response-service/api/v1', 
      description: 'Staging server'
    },
    {
      url: 'https://dev-api.x-form.com/response-service/api/v1',
      description: 'Development server (shared)'
    }
  ],
  tags: [
    {
      name: 'Health',
      description: 'Service health and monitoring endpoints',
      externalDocs: {
        description: 'Health Check Best Practices',
        url: 'https://docs.x-form.com/health-checks'
      }
    },
    {
      name: 'Responses',
      description: 'Form response management operations',
      externalDocs: {
        description: 'Response Management Guide',
        url: 'https://docs.x-form.com/responses'
      }
    },
    {
      name: 'Analytics',
      description: 'Response analytics and reporting',
      externalDocs: {
        description: 'Analytics Documentation',
        url: 'https://docs.x-form.com/analytics'
      }
    },
    {
      name: 'Files',
      description: 'File upload and management',
      externalDocs: {
        description: 'File Upload Guide',
        url: 'https://docs.x-form.com/file-uploads'
      }
    },
    {
      name: 'Export',
      description: 'Data export functionality',
      externalDocs: {
        description: 'Export Documentation',
        url: 'https://docs.x-form.com/exports'
      }
    },
    {
      name: 'Webhooks',
      description: 'Webhook management and notifications',
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
        },
        example: {
          id: 'resp_1234567890abcdef',
          formId: 'form_user_feedback_2023',
          formTitle: 'Customer Feedback Survey 2023',
          respondentEmail: 'customer@example.com',
          respondentName: 'John Doe',
          status: 'completed',
          priority: 'normal',
          responses: [
            {
              questionId: 'q_rating',
              questionType: 'radio',
              value: 'Excellent',
              label: 'How would you rate our service?'
            }
          ],
          tags: ['feedback', 'customer-service'],
          submittedAt: '2023-12-07T10:30:00.000Z',
          createdAt: '2023-12-07T10:25:00.000Z',
          updatedAt: '2023-12-07T10:35:00.000Z',
          version: 1
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
      // METADATA SCHEMAS
      // =====================================================

      ResponseMetadata: {
        type: 'object',
        properties: {
          userAgent: {
            type: 'string',
            description: 'Browser user agent string',
            maxLength: 1000,
            example: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
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
              },
              coordinates: {
                type: 'object',
                properties: {
                  latitude: { 
                    type: 'number', 
                    minimum: -90, 
                    maximum: 90,
                    example: 40.7128 
                  },
                  longitude: { 
                    type: 'number', 
                    minimum: -180, 
                    maximum: 180,
                    example: -74.0060 
                  }
                }
              }
            }
          },
          sessionInfo: {
            type: 'object',
            properties: {
              sessionId: {
                type: 'string',
                description: 'User session identifier',
                example: 'sess_1234567890abcdef'
              },
              isReturningUser: {
                type: 'boolean',
                description: 'Whether this is a returning user',
                example: false
              },
              pageViews: {
                type: 'integer',
                description: 'Number of page views in session',
                minimum: 1,
                example: 3
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
            type: 'string',
            format: 'date-time',
            description: 'Timestamp when response was submitted',
            example: '2023-12-07T10:30:00Z'
          },
          updatedAt: {
            type: 'string',
            format: 'date-time',
            description: 'Timestamp when response was last updated',
            example: '2023-12-07T10:35:00Z'
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
            example: 'q_1234567890'
          },
          questionType: {
            type: 'string',
            enum: ['text', 'textarea', 'radio', 'checkbox', 'select', 'number', 'email', 'url', 'date', 'file'],
            description: 'Type of the question',
            example: 'text'
          },
          value: {
            oneOf: [
              { type: 'string' },
              { type: 'number' },
              { type: 'array', items: { type: 'string' } },
              { type: 'object' }
            ],
            description: 'Response value (type depends on question type)',
            example: 'This is a sample text response'
          },
          files: {
            type: 'array',
            description: 'Uploaded files for file-type questions',
            items: {
              $ref: '#/components/schemas/FileUpload'
            }
          }
        }
      },
      ResponseMetadata: {
        type: 'object',
        properties: {
          userAgent: {
            type: 'string',
            description: 'Browser user agent string',
            example: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
          },
          ipAddress: {
            type: 'string',
            description: 'IP address of the respondent',
            example: '192.168.1.100'
          },
          referrer: {
            type: 'string',
            description: 'Referring URL',
            example: 'https://example.com/survey-page'
          },
          timeSpent: {
            type: 'number',
            description: 'Time spent filling the form (in seconds)',
            example: 180
          },
          deviceType: {
            type: 'string',
            enum: ['desktop', 'mobile', 'tablet'],
            description: 'Type of device used',
            example: 'desktop'
          },
          location: {
            type: 'object',
            properties: {
              country: { type: 'string', example: 'United States' },
              city: { type: 'string', example: 'New York' },
              latitude: { type: 'number', example: 40.7128 },
              longitude: { type: 'number', example: -74.0060 }
            }
          }
        }
      },
      FileUpload: {
        type: 'object',
        properties: {
          fileName: {
            type: 'string',
            description: 'Original file name',
            example: 'document.pdf'
          },
          fileSize: {
            type: 'number',
            description: 'File size in bytes',
            example: 1024000
          },
          mimeType: {
            type: 'string',
            description: 'MIME type of the file',
            example: 'application/pdf'
          },
          fileUrl: {
            type: 'string',
            description: 'URL to access the uploaded file',
            example: 'https://storage.example.com/files/uuid-filename.pdf'
          },
          uploadedAt: {
            type: 'string',
            format: 'date-time',
            description: 'Timestamp when file was uploaded',
            example: '2023-12-07T10:30:00Z'
          }
        }
      },
      Analytics: {
        type: 'object',
        properties: {
          formId: {
            type: 'string',
            description: 'Form identifier',
            example: 'form_abcdef1234567890'
          },
          totalResponses: {
            type: 'number',
            description: 'Total number of responses',
            example: 150
          },
          completedResponses: {
            type: 'number',
            description: 'Number of completed responses',
            example: 120
          },
          draftResponses: {
            type: 'number',
            description: 'Number of draft responses',
            example: 20
          },
          partialResponses: {
            type: 'number',
            description: 'Number of partial responses',
            example: 10
          },
          completionRate: {
            type: 'number',
            description: 'Completion rate percentage',
            example: 80.0
          },
          averageCompletionTime: {
            type: 'number',
            description: 'Average time to complete in seconds',
            example: 240
          },
          responsesByDate: {
            type: 'array',
            description: 'Responses grouped by date',
            items: {
              type: 'object',
              properties: {
                date: { type: 'string', format: 'date', example: '2023-12-07' },
                count: { type: 'number', example: 15 },
                completed: { type: 'number', example: 12 },
                draft: { type: 'number', example: 2 },
                partial: { type: 'number', example: 1 }
              }
            }
          },
          questionAnalytics: {
            type: 'array',
            description: 'Analytics for individual questions',
            items: {
              $ref: '#/components/schemas/QuestionAnalytics'
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
            example: 'q_1234567890'
          },
          questionType: {
            type: 'string',
            description: 'Question type',
            example: 'radio'
          },
          totalResponses: {
            type: 'number',
            description: 'Total responses to this question',
            example: 100
          },
          responseRate: {
            type: 'number',
            description: 'Response rate percentage',
            example: 95.5
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
          }
        }
      },
      HealthCheck: {
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
            example: '2023-12-07T10:30:00Z'
          },
          uptime: {
            type: 'number',
            description: 'Service uptime in seconds',
            example: 86400
          },
          version: {
            type: 'string',
            description: 'Service version',
            example: '1.0.0'
          },
          dependencies: {
            type: 'object',
            description: 'Health status of dependencies',
            properties: {
              database: {
                type: 'object',
                properties: {
                  status: { type: 'string', example: 'healthy' },
                  responseTime: { type: 'number', example: 15 }
                }
              },
              formService: {
                type: 'object',
                properties: {
                  status: { type: 'string', example: 'healthy' },
                  responseTime: { type: 'number', example: 25 }
                }
              }
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
          correlationId: {
            type: 'string',
            description: 'Request correlation ID for tracing',
            example: 'cor_1234567890abcdef'
          },
          timestamp: {
            type: 'string',
            format: 'date-time',
            description: 'Response timestamp',
            example: '2023-12-07T10:30:00Z'
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
                description: 'Error code',
                example: 'VALIDATION_ERROR'
              },
              message: {
                type: 'string',
                description: 'Error message',
                example: 'The provided data is invalid'
              },
              details: {
                type: 'array',
                description: 'Detailed error information',
                items: {
                  type: 'object',
                  properties: {
                    field: { type: 'string', example: 'email' },
                    message: { type: 'string', example: 'Invalid email format' }
                  }
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
            description: 'Error timestamp',
            example: '2023-12-07T10:30:00Z'
          }
        }
      }
    },
    parameters: {
      CorrelationId: {
        name: 'X-Correlation-ID',
        in: 'header',
        description: 'Unique request identifier for tracing',
        required: false,
        schema: {
          type: 'string',
          example: 'cor_1234567890abcdef'
        }
      },
      FormId: {
        name: 'formId',
        in: 'path',
        description: 'Form identifier',
        required: true,
        schema: {
          type: 'string',
          example: 'form_abcdef1234567890'
        }
      },
      ResponseId: {
        name: 'responseId',
        in: 'path',
        description: 'Response identifier',
        required: true,
        schema: {
          type: 'string',
          example: 'resp_1234567890abcdef'
        }
      },
      Page: {
        name: 'page',
        in: 'query',
        description: 'Page number for pagination',
        required: false,
        schema: {
          type: 'integer',
          minimum: 1,
          default: 1,
          example: 1
        }
      },
      Limit: {
        name: 'limit',
        in: 'query',
        description: 'Number of items per page',
        required: false,
        schema: {
          type: 'integer',
          minimum: 1,
          maximum: 100,
          default: 20,
          example: 20
        }
      },
      Status: {
        name: 'status',
        in: 'query',
        description: 'Filter by response status',
        required: false,
        schema: {
          type: 'string',
          enum: ['draft', 'partial', 'completed', 'archived'],
          example: 'completed'
        }
      }
    },
    responses: {
      UnauthorizedError: {
        description: 'Authentication token is missing or invalid',
        content: {
          'application/json': {
            schema: {
              $ref: '#/components/schemas/ErrorResponse'
            },
            example: {
              success: false,
              error: {
                code: 'UNAUTHORIZED',
                message: 'Authentication token is required'
              },
              correlationId: 'cor_1234567890abcdef',
              timestamp: '2023-12-07T10:30:00Z'
            }
          }
        }
      },
      ForbiddenError: {
        description: 'Access denied - insufficient permissions',
        content: {
          'application/json': {
            schema: {
              $ref: '#/components/schemas/ErrorResponse'
            }
          }
        }
      },
      NotFoundError: {
        description: 'Resource not found',
        content: {
          'application/json': {
            schema: {
              $ref: '#/components/schemas/ErrorResponse'
            }
          }
        }
      },
      ValidationError: {
        description: 'Invalid request data',
        content: {
          'application/json': {
            schema: {
              $ref: '#/components/schemas/ErrorResponse'
            }
          }
        }
      },
      RateLimitError: {
        description: 'Rate limit exceeded',
        content: {
          'application/json': {
            schema: {
              $ref: '#/components/schemas/ErrorResponse'
            }
          }
        }
      },
      ServerError: {
        description: 'Internal server error',
        content: {
          'application/json': {
            schema: {
              $ref: '#/components/schemas/ErrorResponse'
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
  tags: [
    {
      name: 'Responses',
      description: 'Form response management operations'
    },
    {
      name: 'Analytics',
      description: 'Analytics and reporting operations'
    },
    {
      name: 'Health',
      description: 'Service health and monitoring'
    },
    {
      name: 'Files',
      description: 'File upload and management'
    }
  ]
};

// Swagger JSDoc options
const options = {
  definition: swaggerDefinition,
  apis: [
    './src/routes/v1/*.js',
    './src/controllers/*.js',
    './src/middleware/*.js'
  ], // Path to the API docs
};

// Initialize swagger-jsdoc
const specs = swaggerJsdoc(options);

// Swagger UI options
const swaggerUiOptions = {
  explorer: true,
  swaggerOptions: {
    persistAuthorization: true,
    displayRequestDuration: true,
    docExpansion: 'none',
    filter: true,
    showExtensions: true,
    showCommonExtensions: true,
    defaultModelsExpandDepth: 2,
    defaultModelExpandDepth: 2
  },
  customJs: '/api-docs/custom.js',
  customCss: `
    .swagger-ui .topbar { display: none }
    .swagger-ui .info { margin: 20px 0 }
    .swagger-ui .scheme-container { background: #f7f7f7; padding: 10px; border-radius: 4px; }
  `,
  customSiteTitle: 'Response Service API Documentation'
};

module.exports = {
  specs,
  swaggerUi,
  swaggerUiOptions
};
