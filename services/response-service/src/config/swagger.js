/**
 * Swagger/OpenAPI Documentation Configuration
 * Comprehensive API documentation setup for the Response Service
 */

const swaggerJsdoc = require('swagger-jsdoc');
const swaggerUi = require('swagger-ui-express');
const config = require('./enhanced');

// Basic API information
const swaggerDefinition = {
  openapi: '3.0.0',
  info: {
    title: 'Response Service API',
    version: config.get('server.version') || '1.0.0',
    description: 'A comprehensive microservice for managing form responses with advanced features including validation, analytics, file uploads, and real-time capabilities.',
    contact: {
      name: 'API Support',
      email: 'support@x-form.com',
      url: 'https://x-form.com/support'
    },
    license: {
      name: 'MIT',
      url: 'https://opensource.org/licenses/MIT'
    },
    termsOfService: 'https://x-form.com/terms'
  },
  servers: [
    {
      url: `http://localhost:${config.get('server.port')}/api/v1`,
      description: 'Development server'
    },
    {
      url: `https://api.x-form.com/response-service/api/v1`,
      description: 'Production server'
    },
    {
      url: `https://staging-api.x-form.com/response-service/api/v1`,
      description: 'Staging server'
    }
  ],
  components: {
    securitySchemes: {
      BearerAuth: {
        type: 'http',
        scheme: 'bearer',
        bearerFormat: 'JWT',
        description: 'JWT token obtained from authentication service'
      },
      ApiKeyAuth: {
        type: 'apiKey',
        in: 'header',
        name: 'X-API-Key',
        description: 'API key for service-to-service authentication'
      }
    },
    schemas: {
      // Response Schemas
      Response: {
        type: 'object',
        required: ['formId', 'responses'],
        properties: {
          id: {
            type: 'string',
            description: 'Unique response identifier',
            example: 'resp_1234567890abcdef'
          },
          formId: {
            type: 'string',
            description: 'Associated form identifier',
            example: 'form_abcdef1234567890'
          },
          formTitle: {
            type: 'string',
            description: 'Title of the associated form',
            example: 'Customer Feedback Survey'
          },
          respondentEmail: {
            type: 'string',
            format: 'email',
            description: 'Email of the respondent (optional)',
            example: 'user@example.com'
          },
          status: {
            type: 'string',
            enum: ['draft', 'partial', 'completed', 'archived'],
            description: 'Current status of the response',
            example: 'completed'
          },
          responses: {
            type: 'array',
            description: 'Array of question responses',
            items: {
              $ref: '#/components/schemas/QuestionResponse'
            }
          },
          metadata: {
            $ref: '#/components/schemas/ResponseMetadata'
          },
          submittedAt: {
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
