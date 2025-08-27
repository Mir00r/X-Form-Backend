/**
 * Validation Middleware for Response Service
 * Implements comprehensive request validation with proper error handling
 */

const { validationSchemas, createErrorResponse } = require('../dto/response-dtos');
const logger = require('../utils/logger');

/**
 * Generic validation middleware factory
 */
const validate = (schema, source = 'body') => {
  return (req, res, next) => {
    const correlationId = req.headers['x-correlation-id'] || req.correlationId;
    
    let dataToValidate;
    switch (source) {
      case 'body':
        dataToValidate = req.body;
        break;
      case 'query':
        dataToValidate = req.query;
        break;
      case 'params':
        dataToValidate = req.params;
        break;
      case 'headers':
        dataToValidate = req.headers;
        break;
      default:
        dataToValidate = req.body;
    }

    const { error, value } = schema.validate(dataToValidate, {
      abortEarly: false, // Return all validation errors
      stripUnknown: true, // Remove unknown fields
      convert: true // Convert strings to appropriate types
    });

    if (error) {
      const validationErrors = error.details.map(detail => ({
        field: detail.path.join('.'),
        message: detail.message,
        value: detail.context?.value
      }));

      logger.warn('Validation failed', {
        correlationId,
        source,
        errors: validationErrors,
        path: req.path,
        method: req.method,
        userAgent: req.get('User-Agent'),
        ip: req.ip
      });

      return res.status(400).json(
        createErrorResponse(
          'VALIDATION_ERROR',
          'Request validation failed',
          {
            errors: validationErrors,
            source
          },
          correlationId
        )
      );
    }

    // Replace the original data with validated/sanitized data
    switch (source) {
      case 'body':
        req.body = value;
        break;
      case 'query':
        req.query = value;
        break;
      case 'params':
        req.params = value;
        break;
      case 'headers':
        req.headers = { ...req.headers, ...value };
        break;
    }

    next();
  };
};

/**
 * Specific validation middlewares for different endpoints
 */

// Request body validations
const validateCreateResponse = validate(validationSchemas.createResponse, 'body');
const validateUpdateResponse = validate(validationSchemas.updateResponse, 'body');

// Query parameter validations
const validateResponseListQuery = validate(validationSchemas.listResponses, 'query');

// URL parameter validations
const validateResponseId = validate(validationSchemas.responseId, 'params');
const validateFormId = validate(validationSchemas.formId, 'params');

/**
 * Combined parameter validation (for routes with multiple params)
 */
const validateResponseIdAndFormId = (req, res, next) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  
  const combinedSchema = validationSchemas.responseId.keys(
    validationSchemas.formId.describe().keys
  );
  
  const { error, value } = combinedSchema.validate(req.params, {
    abortEarly: false,
    stripUnknown: true,
    convert: true
  });

  if (error) {
    const validationErrors = error.details.map(detail => ({
      field: detail.path.join('.'),
      message: detail.message,
      value: detail.context?.value
    }));

    logger.warn('Parameter validation failed', {
      correlationId,
      errors: validationErrors,
      path: req.path,
      method: req.method
    });

    return res.status(400).json(
      createErrorResponse(
        'VALIDATION_ERROR',
        'URL parameter validation failed',
        { errors: validationErrors },
        correlationId
      )
    );
  }

  req.params = value;
  next();
};

/**
 * File upload validation middleware
 */
const validateFileUpload = (req, res, next) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  
  if (!req.files || req.files.length === 0) {
    return next(); // No files to validate
  }

  const errors = [];
  const allowedMimeTypes = [
    'image/jpeg',
    'image/png',
    'image/gif',
    'application/pdf',
    'text/plain',
    'application/msword',
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    'application/vnd.ms-excel',
    'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
  ];
  
  const maxFileSize = 10 * 1024 * 1024; // 10MB
  const maxTotalSize = 50 * 1024 * 1024; // 50MB total
  
  let totalSize = 0;

  req.files.forEach((file, index) => {
    // Check file size
    if (file.size > maxFileSize) {
      errors.push({
        field: `files[${index}]`,
        message: `File size exceeds maximum allowed size of ${maxFileSize / 1024 / 1024}MB`,
        value: file.originalname
      });
    }

    // Check mime type
    if (!allowedMimeTypes.includes(file.mimetype)) {
      errors.push({
        field: `files[${index}]`,
        message: `File type ${file.mimetype} is not allowed`,
        value: file.originalname
      });
    }

    // Check filename
    if (!file.originalname || file.originalname.length > 255) {
      errors.push({
        field: `files[${index}]`,
        message: 'Invalid filename or filename too long',
        value: file.originalname
      });
    }

    totalSize += file.size;
  });

  // Check total size
  if (totalSize > maxTotalSize) {
    errors.push({
      field: 'files',
      message: `Total file size exceeds maximum allowed size of ${maxTotalSize / 1024 / 1024}MB`,
      value: `${totalSize / 1024 / 1024}MB`
    });
  }

  if (errors.length > 0) {
    logger.warn('File upload validation failed', {
      correlationId,
      errors,
      fileCount: req.files.length,
      totalSize,
      path: req.path,
      method: req.method
    });

    return res.status(400).json(
      createErrorResponse(
        'FILE_VALIDATION_ERROR',
        'File upload validation failed',
        { errors },
        correlationId
      )
    );
  }

  next();
};

/**
 * Content-Type validation middleware
 */
const validateContentType = (expectedTypes = ['application/json']) => {
  return (req, res, next) => {
    const correlationId = req.headers['x-correlation-id'] || req.correlationId;
    const contentType = req.get('Content-Type');
    
    if (!contentType) {
      logger.warn('Missing Content-Type header', {
        correlationId,
        path: req.path,
        method: req.method
      });
      
      return res.status(400).json(
        createErrorResponse(
          'MISSING_CONTENT_TYPE',
          'Content-Type header is required',
          { expectedTypes },
          correlationId
        )
      );
    }

    const isValidType = expectedTypes.some(type => 
      contentType.toLowerCase().includes(type.toLowerCase())
    );

    if (!isValidType) {
      logger.warn('Invalid Content-Type', {
        correlationId,
        contentType,
        expectedTypes,
        path: req.path,
        method: req.method
      });
      
      return res.status(415).json(
        createErrorResponse(
          'UNSUPPORTED_MEDIA_TYPE',
          'Unsupported Content-Type',
          { 
            provided: contentType,
            expected: expectedTypes 
          },
          correlationId
        )
      );
    }

    next();
  };
};

/**
 * Request size validation middleware
 */
const validateRequestSize = (maxSize = '10mb') => {
  return (req, res, next) => {
    const correlationId = req.headers['x-correlation-id'] || req.correlationId;
    const contentLength = req.get('Content-Length');
    
    if (contentLength) {
      const sizeInBytes = parseInt(contentLength, 10);
      const maxSizeInBytes = typeof maxSize === 'string' 
        ? parseSize(maxSize) 
        : maxSize;
      
      if (sizeInBytes > maxSizeInBytes) {
        logger.warn('Request size exceeds limit', {
          correlationId,
          contentLength: sizeInBytes,
          maxSize: maxSizeInBytes,
          path: req.path,
          method: req.method
        });
        
        return res.status(413).json(
          createErrorResponse(
            'REQUEST_TOO_LARGE',
            'Request size exceeds maximum allowed size',
            { 
              size: sizeInBytes,
              maxSize: maxSizeInBytes 
            },
            correlationId
          )
        );
      }
    }

    next();
  };
};

/**
 * Helper function to parse size strings like "10mb" to bytes
 */
function parseSize(size) {
  const units = {
    b: 1,
    kb: 1024,
    mb: 1024 * 1024,
    gb: 1024 * 1024 * 1024
  };
  
  const match = size.toLowerCase().match(/^(\d+(?:\.\d+)?)(b|kb|mb|gb)$/);
  if (!match) {
    throw new Error(`Invalid size format: ${size}`);
  }
  
  const value = parseFloat(match[1]);
  const unit = match[2];
  
  return Math.floor(value * units[unit]);
}

/**
 * Sanitization middleware to prevent XSS and injection attacks
 */
const sanitizeInput = (req, res, next) => {
  if (req.body && typeof req.body === 'object') {
    req.body = sanitizeObject(req.body);
  }
  
  if (req.query && typeof req.query === 'object') {
    req.query = sanitizeObject(req.query);
  }
  
  next();
};

function sanitizeObject(obj) {
  if (Array.isArray(obj)) {
    return obj.map(item => 
      typeof item === 'object' ? sanitizeObject(item) : sanitizeValue(item)
    );
  }
  
  if (obj && typeof obj === 'object') {
    const sanitized = {};
    for (const [key, value] of Object.entries(obj)) {
      if (typeof value === 'object') {
        sanitized[key] = sanitizeObject(value);
      } else {
        sanitized[key] = sanitizeValue(value);
      }
    }
    return sanitized;
  }
  
  return sanitizeValue(obj);
}

function sanitizeValue(value) {
  if (typeof value !== 'string') {
    return value;
  }
  
  // Basic XSS prevention - remove script tags and javascript: protocols
  return value
    .replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '')
    .replace(/javascript:/gi, '')
    .replace(/on\w+\s*=/gi, '')
    .trim();
}

module.exports = {
  validate,
  validateCreateResponse,
  validateUpdateResponse,
  validateResponseListQuery,
  validateResponseId,
  validateFormId,
  validateResponseIdAndFormId,
  validateFileUpload,
  validateContentType,
  validateRequestSize,
  sanitizeInput
};
