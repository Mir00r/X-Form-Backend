const Joi = require('joi');
const { ValidationHelper } = require('../utils/helpers');

/**
 * Common validation schemas
 */
const commonSchemas = {
  id: Joi.string().uuid().required(),
  optionalId: Joi.string().uuid().optional(),
  email: Joi.string().email().optional(),
  requiredEmail: Joi.string().email().required(),
  url: Joi.string().uri().optional(),
  pagination: Joi.object({
    page: Joi.number().integer().min(1).default(1),
    limit: Joi.number().integer().min(1).max(100).default(10),
  }),
  dateRange: Joi.object({
    startDate: Joi.date().iso().optional(),
    endDate: Joi.date().iso().min(Joi.ref('startDate')).optional(),
  }),
  status: Joi.string().valid('submitted', 'draft', 'flagged', 'deleted').default('submitted'),
};

/**
 * Response validation schemas
 */
const responseSchemas = {
  // Create response validation
  createResponse: Joi.object({
    formId: commonSchemas.id,
    responses: Joi.object().required(),
    submitterId: commonSchemas.optionalId,
    submitterEmail: commonSchemas.email,
    submitterName: Joi.string().max(100).optional(),
    isAnonymous: Joi.boolean().default(false),
    isDraft: Joi.boolean().default(false),
    isComplete: Joi.boolean().default(true),
    
    // Metadata
    startTime: Joi.date().iso().optional(),
    endTime: Joi.date().iso().min(Joi.ref('startTime')).optional(),
    sessionId: Joi.string().max(100).optional(),
    
    // Analytics data
    pageViews: Joi.array().items(Joi.object({
      page: Joi.string().required(),
      timestamp: Joi.date().iso().required(),
      duration: Joi.number().min(0).optional(),
    })).optional(),
    
    interactions: Joi.array().items(Joi.object({
      type: Joi.string().required(),
      questionId: Joi.string().optional(),
      timestamp: Joi.date().iso().required(),
      data: Joi.object().optional(),
    })).optional(),
    
    // Tags and scoring
    tags: Joi.array().items(Joi.string().max(50)).max(10).optional(),
    score: Joi.number().min(0).max(100).optional(),
  }),

  // Update response validation
  updateResponse: Joi.object({
    responses: Joi.object().optional(),
    submitterEmail: commonSchemas.email,
    submitterName: Joi.string().max(100).optional(),
    status: commonSchemas.status,
    isDraft: Joi.boolean().optional(),
    isComplete: Joi.boolean().optional(),
    tags: Joi.array().items(Joi.string().max(50)).max(10).optional(),
    score: Joi.number().min(0).max(100).optional(),
  }),

  // Query responses validation
  queryResponses: Joi.object({
    formId: commonSchemas.optionalId,
    submitterId: commonSchemas.optionalId,
    submitterEmail: commonSchemas.email,
    status: Joi.string().valid('submitted', 'draft', 'flagged', 'deleted', 'all').optional(),
    isComplete: Joi.boolean().optional(),
    isDraft: Joi.boolean().optional(),
    
    // Date filtering
    startDate: Joi.date().iso().optional(),
    endDate: Joi.date().iso().min(Joi.ref('startDate')).optional(),
    
    // Search and filtering
    search: Joi.string().max(100).optional(),
    tags: Joi.array().items(Joi.string()).optional(),
    minScore: Joi.number().min(0).max(100).optional(),
    maxScore: Joi.number().min(Joi.ref('minScore')).max(100).optional(),
    
    // Sorting
    sortBy: Joi.string().valid('submittedAt', 'updatedAt', 'score', 'duration').default('submittedAt'),
    sortOrder: Joi.string().valid('asc', 'desc').default('desc'),
    
    // Pagination
    ...commonSchemas.pagination.describe().keys,
  }),

  // Bulk operations validation
  bulkUpdate: Joi.object({
    responseIds: Joi.array().items(commonSchemas.id).min(1).max(100).required(),
    updates: Joi.object({
      status: commonSchemas.status,
      tags: Joi.array().items(Joi.string().max(50)).max(10).optional(),
      score: Joi.number().min(0).max(100).optional(),
    }).min(1).required(),
  }),

  bulkDelete: Joi.object({
    responseIds: Joi.array().items(commonSchemas.id).min(1).max(100).required(),
    permanent: Joi.boolean().default(false),
  }),
};

/**
 * Export validation schemas
 */
const exportSchemas = {
  // CSV export validation
  exportCsv: Joi.object({
    formId: commonSchemas.id,
    format: Joi.string().valid('csv', 'xlsx').default('csv'),
    includeMetadata: Joi.boolean().default(true),
    includeDrafts: Joi.boolean().default(false),
    includeDeleted: Joi.boolean().default(false),
    
    // Date range
    startDate: Joi.date().iso().optional(),
    endDate: Joi.date().iso().min(Joi.ref('startDate')).optional(),
    
    // Field selection
    fields: Joi.array().items(Joi.string()).optional(),
    excludeFields: Joi.array().items(Joi.string()).optional(),
    
    // Custom options
    filename: Joi.string().max(100).optional(),
    delimiter: Joi.string().max(1).default(','),
    encoding: Joi.string().valid('utf8', 'latin1').default('utf8'),
  }),

  // Google Sheets export validation
  exportGoogleSheets: Joi.object({
    formId: commonSchemas.id,
    spreadsheetId: Joi.string().required(),
    worksheetName: Joi.string().max(100).default('Responses'),
    includeMetadata: Joi.boolean().default(true),
    clearExisting: Joi.boolean().default(false),
    
    // Field mapping
    fieldMapping: Joi.object().pattern(
      Joi.string(),
      Joi.string()
    ).optional(),
    
    // Date range
    startDate: Joi.date().iso().optional(),
    endDate: Joi.date().iso().min(Joi.ref('startDate')).optional(),
  }),
};

/**
 * Analytics validation schemas
 */
const analyticsSchemas = {
  // Analytics query validation
  getAnalytics: Joi.object({
    formId: commonSchemas.optionalId,
    period: Joi.string().valid('daily', 'weekly', 'monthly').default('daily'),
    startDate: Joi.date().iso().required(),
    endDate: Joi.date().iso().min(Joi.ref('startDate')).required(),
    metrics: Joi.array().items(
      Joi.string().valid(
        'responses',
        'completion_rate',
        'average_time',
        'bounce_rate',
        'conversion_rate',
        'device_breakdown',
        'browser_breakdown',
        'location_breakdown'
      )
    ).default(['responses', 'completion_rate', 'average_time']),
  }),

  // Real-time analytics validation
  realtimeAnalytics: Joi.object({
    formId: commonSchemas.id,
    timeWindow: Joi.number().integer().min(1).max(1440).default(60), // minutes
  }),
};

/**
 * Integration validation schemas
 */
const integrationSchemas = {
  // Create integration validation
  createIntegration: Joi.object({
    formId: commonSchemas.id,
    type: Joi.string().valid('google_sheets', 'webhook', 'email', 'zapier').required(),
    isActive: Joi.boolean().default(true),
    config: Joi.object().required(),
  }).custom((value, helpers) => {
    // Type-specific config validation
    const { type, config } = value;
    
    switch (type) {
      case 'google_sheets':
        const sheetsSchema = Joi.object({
          spreadsheetId: Joi.string().required(),
          worksheetName: Joi.string().default('Responses'),
          includeTimestamp: Joi.boolean().default(true),
          fieldMapping: Joi.object().optional(),
        });
        const { error: sheetsError } = sheetsSchema.validate(config);
        if (sheetsError) {
          return helpers.error('custom.googleSheetsConfig', { error: sheetsError.message });
        }
        break;
        
      case 'webhook':
        const webhookSchema = Joi.object({
          url: Joi.string().uri().required(),
          method: Joi.string().valid('POST', 'PUT', 'PATCH').default('POST'),
          headers: Joi.object().optional(),
          includeMetadata: Joi.boolean().default(true),
          retryPolicy: Joi.object({
            maxRetries: Joi.number().integer().min(0).max(10).default(3),
            backoff: Joi.string().valid('linear', 'exponential').default('exponential'),
          }).optional(),
        });
        const { error: webhookError } = webhookSchema.validate(config);
        if (webhookError) {
          return helpers.error('custom.webhookConfig', { error: webhookError.message });
        }
        break;
        
      case 'email':
        const emailSchema = Joi.object({
          recipients: Joi.array().items(Joi.string().email()).min(1).required(),
          subject: Joi.string().max(200).default('New form response'),
          template: Joi.string().valid('default', 'detailed', 'summary').default('default'),
          includeAttachments: Joi.boolean().default(true),
        });
        const { error: emailError } = emailSchema.validate(config);
        if (emailError) {
          return helpers.error('custom.emailConfig', { error: emailError.message });
        }
        break;
    }
    
    return value;
  }, 'Integration config validation'),

  // Update integration validation
  updateIntegration: Joi.object({
    isActive: Joi.boolean().optional(),
    config: Joi.object().optional(),
  }),

  // Test integration validation
  testIntegration: Joi.object({
    integrationId: commonSchemas.id,
    sampleData: Joi.object().optional(),
  }),
};

/**
 * File upload validation schemas
 */
const fileSchemas = {
  // File upload validation
  uploadFile: Joi.object({
    questionId: Joi.string().required(),
    allowedTypes: Joi.array().items(Joi.string()).default(['jpg', 'jpeg', 'png', 'gif', 'pdf', 'doc', 'docx']),
    maxSize: Joi.number().integer().min(1).max(10485760).default(5242880), // 5MB default
  }),

  // Batch file upload validation
  batchUpload: Joi.object({
    files: Joi.array().items(Joi.object({
      questionId: Joi.string().required(),
      filename: Joi.string().required(),
      mimeType: Joi.string().required(),
      size: Joi.number().integer().min(1).required(),
    })).min(1).max(10).required(),
  }),
};

/**
 * WebSocket validation schemas
 */
const websocketSchemas = {
  // Subscribe to form updates
  subscribeForm: Joi.object({
    formId: commonSchemas.id,
    events: Joi.array().items(
      Joi.string().valid('response_created', 'response_updated', 'response_deleted', 'form_updated')
    ).default(['response_created']),
  }),

  // Real-time response validation
  realtimeResponse: Joi.object({
    formId: commonSchemas.id,
    questionId: Joi.string().required(),
    answer: Joi.any().required(),
    sessionId: Joi.string().optional(),
  }),
};

/**
 * Validation middleware factory
 */
const createValidator = (schema, property = 'body') => {
  return (req, res, next) => {
    const { error, value } = schema.validate(req[property], {
      abortEarly: false,
      allowUnknown: false,
      stripUnknown: true,
    });

    if (error) {
      const validationErrors = error.details.map(detail => ({
        field: detail.path.join('.'),
        message: detail.message,
        value: detail.context?.value,
      }));

      return res.status(400).json({
        success: false,
        message: 'Validation failed',
        errors: validationErrors,
        timestamp: new Date().toISOString(),
      });
    }

    // Replace validated data
    req[property] = value;
    next();
  };
};

/**
 * Custom validators
 */
const customValidators = {
  // Validate response answers against form schema
  validateResponseAnswers: async (responses, formSchema) => {
    const errors = [];
    
    if (!formSchema || !formSchema.questions) {
      return { isValid: true, errors: [] };
    }

    for (const question of formSchema.questions) {
      const answer = responses[question.id];
      
      // Check required questions
      if (question.required && (answer === undefined || answer === null || answer === '')) {
        errors.push({
          questionId: question.id,
          field: question.question || question.title,
          message: 'This field is required',
        });
        continue;
      }

      // Skip validation if answer is empty and not required
      if (!answer && !question.required) {
        continue;
      }

      // Type-specific validation
      switch (question.type) {
        case 'email':
          if (answer && !ValidationHelper.isValidEmail(answer)) {
            errors.push({
              questionId: question.id,
              field: question.question || question.title,
              message: 'Invalid email format',
              value: answer,
            });
          }
          break;

        case 'url':
          if (answer && !ValidationHelper.isValidUrl(answer)) {
            errors.push({
              questionId: question.id,
              field: question.question || question.title,
              message: 'Invalid URL format',
              value: answer,
            });
          }
          break;

        case 'number':
          const numValue = parseFloat(answer);
          if (isNaN(numValue)) {
            errors.push({
              questionId: question.id,
              field: question.question || question.title,
              message: 'Must be a valid number',
              value: answer,
            });
          } else {
            // Check min/max constraints
            if (question.validation?.min !== undefined && numValue < question.validation.min) {
              errors.push({
                questionId: question.id,
                field: question.question || question.title,
                message: `Must be at least ${question.validation.min}`,
                value: answer,
              });
            }
            if (question.validation?.max !== undefined && numValue > question.validation.max) {
              errors.push({
                questionId: question.id,
                field: question.question || question.title,
                message: `Must be at most ${question.validation.max}`,
                value: answer,
              });
            }
          }
          break;

        case 'text':
        case 'textarea':
          if (typeof answer !== 'string') {
            errors.push({
              questionId: question.id,
              field: question.question || question.title,
              message: 'Must be a text value',
              value: answer,
            });
          } else {
            // Check length constraints
            if (question.validation?.minLength && answer.length < question.validation.minLength) {
              errors.push({
                questionId: question.id,
                field: question.question || question.title,
                message: `Must be at least ${question.validation.minLength} characters`,
                value: answer,
              });
            }
            if (question.validation?.maxLength && answer.length > question.validation.maxLength) {
              errors.push({
                questionId: question.id,
                field: question.question || question.title,
                message: `Must be at most ${question.validation.maxLength} characters`,
                value: answer,
              });
            }
          }
          break;

        case 'multiple_choice':
          if (question.allowMultiple) {
            if (!Array.isArray(answer)) {
              errors.push({
                questionId: question.id,
                field: question.question || question.title,
                message: 'Must be an array of values',
                value: answer,
              });
            } else {
              // Check if all selected options are valid
              const validOptions = question.options?.map(opt => opt.value || opt.text) || [];
              const invalidOptions = answer.filter(val => !validOptions.includes(val));
              if (invalidOptions.length > 0) {
                errors.push({
                  questionId: question.id,
                  field: question.question || question.title,
                  message: `Invalid options: ${invalidOptions.join(', ')}`,
                  value: answer,
                });
              }
            }
          } else {
            // Single choice validation
            const validOptions = question.options?.map(opt => opt.value || opt.text) || [];
            if (!validOptions.includes(answer)) {
              errors.push({
                questionId: question.id,
                field: question.question || question.title,
                message: 'Invalid option selected',
                value: answer,
              });
            }
          }
          break;

        case 'date':
          if (!ValidationHelper.isValidDate(new Date(answer))) {
            errors.push({
              questionId: question.id,
              field: question.question || question.title,
              message: 'Invalid date format',
              value: answer,
            });
          }
          break;

        case 'file_upload':
          if (Array.isArray(answer)) {
            for (const file of answer) {
              if (!file.filename || !file.url) {
                errors.push({
                  questionId: question.id,
                  field: question.question || question.title,
                  message: 'Invalid file data',
                  value: file,
                });
              }
              
              // Check file type if specified
              if (question.validation?.allowedTypes) {
                const fileExtension = file.filename.split('.').pop().toLowerCase();
                if (!question.validation.allowedTypes.includes(fileExtension)) {
                  errors.push({
                    questionId: question.id,
                    field: question.question || question.title,
                    message: `File type not allowed. Allowed types: ${question.validation.allowedTypes.join(', ')}`,
                    value: file.filename,
                  });
                }
              }
            }
          }
          break;
      }
    }

    return {
      isValid: errors.length === 0,
      errors,
    };
  },
};

module.exports = {
  schemas: {
    common: commonSchemas,
    response: responseSchemas,
    export: exportSchemas,
    analytics: analyticsSchemas,
    integration: integrationSchemas,
    file: fileSchemas,
    websocket: websocketSchemas,
  },
  createValidator,
  customValidators,
};
