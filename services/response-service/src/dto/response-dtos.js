/**
 * Data Transfer Objects (DTOs) for Response Service
 * Following microservices best practices for API contract stability
 */

const Joi = require('joi');

// =============================================================================
// Base Response DTOs
// =============================================================================

/**
 * Standard success response structure
 */
const createSuccessResponse = (data, message = 'Success', correlationId = null) => ({
  success: true,
  message,
  data,
  correlationId,
  timestamp: new Date().toISOString(),
  version: 'v1'
});

/**
 * Standard error response structure
 */
const createErrorResponse = (code, message, details = null, correlationId = null) => ({
  success: false,
  error: {
    code,
    message,
    details,
    timestamp: new Date().toISOString(),
    correlationId
  },
  version: 'v1'
});

// =============================================================================
// Response Request DTOs
// =============================================================================

/**
 * DTO for submitting a form response
 */
const createResponseRequestDTO = {
  formId: Joi.string()
    .uuid()
    .required()
    .description('UUID of the form being responded to')
    .example('f123e4567-e89b-12d3-a456-426614174000'),
  
  respondentId: Joi.string()
    .uuid()
    .optional()
    .description('UUID of the authenticated respondent (null for anonymous)')
    .example('u123e4567-e89b-12d3-a456-426614174000'),
  
  respondentEmail: Joi.string()
    .email()
    .optional()
    .max(255)
    .description('Email of the respondent for contact purposes')
    .example('respondent@example.com'),
  
  respondentName: Joi.string()
    .optional()
    .max(255)
    .trim()
    .description('Name of the respondent')
    .example('John Doe'),
  
  responses: Joi.array()
    .items(Joi.object({
      questionId: Joi.string()
        .uuid()
        .required()
        .description('UUID of the question being answered')
        .example('q123e4567-e89b-12d3-a456-426614174000'),
      
      questionType: Joi.string()
        .valid('text', 'textarea', 'number', 'email', 'date', 'checkbox', 'radio', 'select', 'file', 'rating')
        .required()
        .description('Type of the question'),
      
      value: Joi.alternatives()
        .try(
          Joi.string().max(5000),
          Joi.number(),
          Joi.boolean(),
          Joi.array().items(Joi.string()),
          Joi.object()
        )
        .required()
        .description('Answer value (type depends on question type)'),
      
      textValue: Joi.string()
        .optional()
        .max(5000)
        .description('Text representation of the answer for searching'),
      
      files: Joi.array()
        .items(Joi.object({
          fileName: Joi.string().required(),
          fileUrl: Joi.string().uri().required(),
          fileSize: Joi.number().required(),
          mimeType: Joi.string().required()
        }))
        .optional()
        .description('File attachments for file-type questions')
    }))
    .min(1)
    .required()
    .description('Array of question responses'),
  
  metadata: Joi.object({
    userAgent: Joi.string().optional(),
    ipAddress: Joi.string().ip().optional(),
    source: Joi.string().optional(),
    referrer: Joi.string().uri().optional(),
    sessionId: Joi.string().optional(),
    startedAt: Joi.date().iso().optional(),
    timeSpent: Joi.number().positive().optional().description('Time spent in seconds')
  }).optional().description('Additional metadata about the response'),
  
  isDraft: Joi.boolean()
    .default(false)
    .description('Whether this is a draft submission'),
  
  isPartial: Joi.boolean()
    .default(false)
    .description('Whether this is a partial submission (not all required questions answered)')
};

/**
 * DTO for updating an existing response
 */
const updateResponseRequestDTO = {
  responses: Joi.array()
    .items(Joi.object({
      questionId: Joi.string().uuid().required(),
      questionType: Joi.string()
        .valid('text', 'textarea', 'number', 'email', 'date', 'checkbox', 'radio', 'select', 'file', 'rating')
        .required(),
      value: Joi.alternatives()
        .try(
          Joi.string().max(5000),
          Joi.number(),
          Joi.boolean(),
          Joi.array().items(Joi.string()),
          Joi.object()
        )
        .required(),
      textValue: Joi.string().optional().max(5000),
      files: Joi.array()
        .items(Joi.object({
          fileName: Joi.string().required(),
          fileUrl: Joi.string().uri().required(),
          fileSize: Joi.number().required(),
          mimeType: Joi.string().required()
        }))
        .optional()
    }))
    .min(1)
    .optional(),
  
  respondentEmail: Joi.string()
    .email()
    .optional()
    .max(255),
  
  respondentName: Joi.string()
    .optional()
    .max(255)
    .trim(),
  
  metadata: Joi.object().optional(),
  
  isDraft: Joi.boolean().optional(),
  
  isPartial: Joi.boolean().optional()
};

// =============================================================================
// Response Response DTOs
// =============================================================================

/**
 * Complete response details DTO
 */
const responseResponseDTO = {
  id: {
    type: 'string',
    format: 'uuid',
    description: 'Unique identifier for the response',
    example: 'r123e4567-e89b-12d3-a456-426614174000'
  },
  
  formId: {
    type: 'string',
    format: 'uuid',
    description: 'UUID of the form this response belongs to',
    example: 'f123e4567-e89b-12d3-a456-426614174000'
  },
  
  formTitle: {
    type: 'string',
    description: 'Title of the form (for reference)',
    example: 'Customer Feedback Survey'
  },
  
  respondentId: {
    type: 'string',
    format: 'uuid',
    nullable: true,
    description: 'UUID of the authenticated respondent',
    example: 'u123e4567-e89b-12d3-a456-426614174000'
  },
  
  respondentEmail: {
    type: 'string',
    format: 'email',
    nullable: true,
    description: 'Email of the respondent',
    example: 'respondent@example.com'
  },
  
  respondentName: {
    type: 'string',
    nullable: true,
    description: 'Name of the respondent',
    example: 'John Doe'
  },
  
  responses: {
    type: 'array',
    items: {
      type: 'object',
      properties: {
        questionId: {
          type: 'string',
          format: 'uuid',
          description: 'UUID of the question'
        },
        questionTitle: {
          type: 'string',
          description: 'Title of the question'
        },
        questionType: {
          type: 'string',
          enum: ['text', 'textarea', 'number', 'email', 'date', 'checkbox', 'radio', 'select', 'file', 'rating'],
          description: 'Type of the question'
        },
        value: {
          description: 'Answer value (type varies by question type)'
        },
        textValue: {
          type: 'string',
          description: 'Text representation of the answer'
        },
        files: {
          type: 'array',
          items: {
            type: 'object',
            properties: {
              fileName: { type: 'string' },
              fileUrl: { type: 'string', format: 'uri' },
              fileSize: { type: 'number' },
              mimeType: { type: 'string' }
            }
          }
        }
      }
    }
  },
  
  status: {
    type: 'string',
    enum: ['draft', 'partial', 'completed', 'archived'],
    description: 'Status of the response',
    example: 'completed'
  },
  
  isDraft: {
    type: 'boolean',
    description: 'Whether this is a draft submission'
  },
  
  isPartial: {
    type: 'boolean',
    description: 'Whether this is a partial submission'
  },
  
  metadata: {
    type: 'object',
    description: 'Additional metadata',
    properties: {
      userAgent: { type: 'string' },
      ipAddress: { type: 'string' },
      source: { type: 'string' },
      referrer: { type: 'string' },
      sessionId: { type: 'string' },
      startedAt: { type: 'string', format: 'date-time' },
      timeSpent: { type: 'number', description: 'Time spent in seconds' }
    }
  },
  
  submittedAt: {
    type: 'string',
    format: 'date-time',
    description: 'When the response was submitted',
    example: '2024-01-01T12:00:00Z'
  },
  
  updatedAt: {
    type: 'string',
    format: 'date-time',
    description: 'When the response was last updated',
    example: '2024-01-01T12:30:00Z'
  },
  
  version: {
    type: 'number',
    description: 'Version number for optimistic locking',
    example: 1
  }
};

/**
 * Response summary DTO for listings
 */
const responseSummaryDTO = {
  id: responseResponseDTO.id,
  formId: responseResponseDTO.formId,
  formTitle: responseResponseDTO.formTitle,
  respondentEmail: responseResponseDTO.respondentEmail,
  respondentName: responseResponseDTO.respondentName,
  status: responseResponseDTO.status,
  submittedAt: responseResponseDTO.submittedAt,
  updatedAt: responseResponseDTO.updatedAt,
  responseCount: {
    type: 'number',
    description: 'Number of questions answered',
    example: 5
  }
};

// =============================================================================
// List and Filter DTOs
// =============================================================================

/**
 * Query parameters for listing responses
 */
const responseListQueryDTO = {
  formId: Joi.string()
    .uuid()
    .optional()
    .description('Filter by form ID'),
  
  respondentEmail: Joi.string()
    .email()
    .optional()
    .description('Filter by respondent email'),
  
  status: Joi.string()
    .valid('draft', 'partial', 'completed', 'archived')
    .optional()
    .description('Filter by response status'),
  
  startDate: Joi.date()
    .iso()
    .optional()
    .description('Filter responses submitted after this date'),
  
  endDate: Joi.date()
    .iso()
    .optional()
    .description('Filter responses submitted before this date'),
  
  page: Joi.number()
    .integer()
    .min(1)
    .default(1)
    .description('Page number for pagination'),
  
  limit: Joi.number()
    .integer()
    .min(1)
    .max(100)
    .default(20)
    .description('Number of items per page'),
  
  sortBy: Joi.string()
    .valid('submittedAt', 'updatedAt', 'respondentEmail', 'status')
    .default('submittedAt')
    .description('Field to sort by'),
  
  sortOrder: Joi.string()
    .valid('asc', 'desc')
    .default('desc')
    .description('Sort order'),
  
  search: Joi.string()
    .optional()
    .max(255)
    .description('Search in response text values')
};

/**
 * Paginated response list DTO
 */
const responseListResponseDTO = {
  responses: {
    type: 'array',
    items: responseSummaryDTO
  },
  pagination: {
    type: 'object',
    properties: {
      currentPage: { type: 'number', example: 1 },
      totalPages: { type: 'number', example: 5 },
      totalItems: { type: 'number', example: 95 },
      itemsPerPage: { type: 'number', example: 20 },
      hasNext: { type: 'boolean', example: true },
      hasPrevious: { type: 'boolean', example: false }
    }
  }
};

// =============================================================================
// Analytics DTOs
// =============================================================================

/**
 * Response analytics summary DTO
 */
const responseAnalyticsDTO = {
  formId: {
    type: 'string',
    format: 'uuid',
    description: 'Form ID for the analytics'
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
  averageCompletionTime: {
    type: 'number',
    description: 'Average completion time in seconds',
    example: 180
  },
  completionRate: {
    type: 'number',
    description: 'Completion rate as percentage',
    example: 80.5
  },
  responsesByDate: {
    type: 'array',
    items: {
      type: 'object',
      properties: {
        date: { type: 'string', format: 'date' },
        count: { type: 'number' }
      }
    }
  },
  lastResponse: {
    type: 'string',
    format: 'date-time',
    nullable: true,
    description: 'Timestamp of the last response'
  }
};

// =============================================================================
// Health Check DTOs
// =============================================================================

/**
 * Health check response DTO
 */
const healthResponseDTO = {
  status: {
    type: 'string',
    enum: ['healthy', 'degraded', 'unhealthy'],
    description: 'Overall health status'
  },
  timestamp: {
    type: 'string',
    format: 'date-time',
    description: 'Health check timestamp'
  },
  version: {
    type: 'string',
    description: 'Service version'
  },
  uptime: {
    type: 'string',
    description: 'Service uptime'
  },
  checks: {
    type: 'object',
    properties: {
      database: {
        type: 'object',
        properties: {
          status: { type: 'string', enum: ['healthy', 'unhealthy'] },
          responseTime: { type: 'number', description: 'Response time in ms' }
        }
      },
      externalServices: {
        type: 'object',
        properties: {
          formService: {
            type: 'object',
            properties: {
              status: { type: 'string', enum: ['healthy', 'unhealthy'] },
              responseTime: { type: 'number' }
            }
          }
        }
      }
    }
  }
};

// =============================================================================
// Validation Schemas
// =============================================================================

const validationSchemas = {
  createResponse: Joi.object(createResponseRequestDTO),
  updateResponse: Joi.object(updateResponseRequestDTO),
  listResponses: Joi.object(responseListQueryDTO),
  responseId: Joi.object({
    id: Joi.string().uuid().required().description('Response ID')
  }),
  formId: Joi.object({
    formId: Joi.string().uuid().required().description('Form ID')
  })
};

module.exports = {
  // Response functions
  createSuccessResponse,
  createErrorResponse,
  
  // Validation schemas
  validationSchemas,
  
  // DTO definitions for OpenAPI
  dtoSchemas: {
    CreateResponseRequest: createResponseRequestDTO,
    UpdateResponseRequest: updateResponseRequestDTO,
    ResponseResponse: responseResponseDTO,
    ResponseSummary: responseSummaryDTO,
    ResponseListQuery: responseListQueryDTO,
    ResponseListResponse: responseListResponseDTO,
    ResponseAnalytics: responseAnalyticsDTO,
    HealthResponse: healthResponseDTO
  }
};
