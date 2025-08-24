const { ResponseFormatter, ErrorHelper } = require('../utils/helpers');
const logger = require('../utils/logger');
const config = require('../config');

/**
 * Global error handler middleware
 */
const errorHandler = (error, req, res, next) => {
  // Log the error
  logger.logError(error, req, {
    stack: error.stack,
    requestBody: req.body,
    requestParams: req.params,
    requestQuery: req.query,
  });

  // Handle different types of errors
  let statusCode = error.statusCode || 500;
  let message = error.message || 'Internal Server Error';
  let code = error.code || 'INTERNAL_ERROR';

  // Validation errors
  if (error.name === 'ValidationError') {
    statusCode = 400;
    message = 'Validation failed';
    code = 'VALIDATION_ERROR';
    
    const validationErrors = Object.values(error.errors).map(err => ({
      field: err.path,
      message: err.message,
      value: err.value,
    }));

    return res.status(statusCode).json({
      success: false,
      message,
      code,
      errors: validationErrors,
      timestamp: new Date().toISOString(),
    });
  }

  // Firestore errors
  if (error.code && error.code.includes('firestore')) {
    statusCode = 500;
    message = 'Database operation failed';
    code = 'DATABASE_ERROR';
    
    // Don't expose internal database errors in production
    if (config.server.environment === 'production') {
      message = 'An internal error occurred';
    }
  }

  // JWT errors
  if (error.name === 'JsonWebTokenError') {
    statusCode = 401;
    message = 'Invalid token';
    code = 'INVALID_TOKEN';
  }

  if (error.name === 'TokenExpiredError') {
    statusCode = 401;
    message = 'Token expired';
    code = 'TOKEN_EXPIRED';
  }

  // Multer errors (file upload)
  if (error.code === 'LIMIT_FILE_SIZE') {
    statusCode = 413;
    message = 'File too large';
    code = 'FILE_TOO_LARGE';
  }

  if (error.code === 'LIMIT_FILE_COUNT') {
    statusCode = 413;
    message = 'Too many files';
    code = 'TOO_MANY_FILES';
  }

  if (error.code === 'LIMIT_UNEXPECTED_FILE') {
    statusCode = 400;
    message = 'Unexpected file field';
    code = 'UNEXPECTED_FILE';
  }

  // Network/timeout errors
  if (error.code === 'ECONNREFUSED' || error.code === 'ENOTFOUND') {
    statusCode = 503;
    message = 'Service unavailable';
    code = 'SERVICE_UNAVAILABLE';
  }

  if (error.code === 'ETIMEDOUT') {
    statusCode = 504;
    message = 'Request timeout';
    code = 'REQUEST_TIMEOUT';
  }

  // Google Sheets API errors
  if (error.code && error.code.toString().startsWith('4')) {
    statusCode = parseInt(error.code);
    if (error.message.includes('PERMISSION_DENIED')) {
      message = 'Google Sheets access denied';
      code = 'SHEETS_PERMISSION_DENIED';
    } else if (error.message.includes('NOT_FOUND')) {
      message = 'Google Sheets resource not found';
      code = 'SHEETS_NOT_FOUND';
    }
  }

  // Kafka errors
  if (error.message && error.message.includes('kafka')) {
    statusCode = 503;
    message = 'Message queue unavailable';
    code = 'QUEUE_UNAVAILABLE';
    
    if (config.server.environment === 'production') {
      message = 'Service temporarily unavailable';
    }
  }

  // Rate limiting errors
  if (error.message && error.message.includes('rate limit')) {
    statusCode = 429;
    message = 'Rate limit exceeded';
    code = 'RATE_LIMIT_EXCEEDED';
  }

  // Permission errors
  if (error.message && error.message.includes('permission')) {
    statusCode = 403;
    message = 'Insufficient permissions';
    code = 'INSUFFICIENT_PERMISSIONS';
  }

  // In production, don't expose internal error details
  if (config.server.environment === 'production' && statusCode === 500) {
    message = 'An internal error occurred';
  }

  // Send error response
  const errorResponse = {
    success: false,
    message,
    code,
    timestamp: new Date().toISOString(),
  };

  // Add error ID for tracking
  if (statusCode >= 500) {
    errorResponse.errorId = `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  // Add stack trace in development
  if (config.server.environment === 'development') {
    errorResponse.stack = error.stack;
    errorResponse.details = {
      originalMessage: error.message,
      originalCode: error.code,
    };
  }

  res.status(statusCode).json(errorResponse);
};

/**
 * 404 Not Found handler
 */
const notFoundHandler = (req, res) => {
  logger.warn('404 Not Found', {
    url: req.url,
    method: req.method,
    ip: req.ip,
    userAgent: req.get('User-Agent'),
    userId: req.user?.id,
  });

  ResponseFormatter.notFound(res, 'Endpoint not found');
};

/**
 * Async wrapper to catch async errors
 */
const asyncHandler = (fn) => {
  return (req, res, next) => {
    Promise.resolve(fn(req, res, next)).catch(next);
  };
};

/**
 * Request timeout handler
 */
const timeoutHandler = (timeout = 30000) => {
  return (req, res, next) => {
    // Set a timeout for the request
    const timer = setTimeout(() => {
      if (!res.headersSent) {
        logger.warn('Request timeout', {
          url: req.url,
          method: req.method,
          timeout,
          ip: req.ip,
          userId: req.user?.id,
        });

        res.status(504).json({
          success: false,
          message: 'Request timeout',
          code: 'REQUEST_TIMEOUT',
          timestamp: new Date().toISOString(),
        });
      }
    }, timeout);

    // Clear timeout when response finishes
    res.on('finish', () => {
      clearTimeout(timer);
    });

    next();
  };
};

/**
 * Validation error formatter
 */
const formatValidationError = (errors) => {
  if (Array.isArray(errors)) {
    return errors.map(error => ({
      field: error.field || error.path,
      message: error.message,
      value: error.value,
    }));
  }

  if (errors.details) {
    return errors.details.map(detail => ({
      field: detail.path.join('.'),
      message: detail.message,
      value: detail.context?.value,
    }));
  }

  return [{ message: 'Validation failed' }];
};

/**
 * Database error handler
 */
const handleDatabaseError = (error) => {
  logger.error('Database error:', error);

  if (error.code === 'ECONNREFUSED') {
    return ErrorHelper.createError('Database connection failed', 503, 'DB_CONNECTION_ERROR');
  }

  if (error.message && error.message.includes('duplicate')) {
    return ErrorHelper.createError('Resource already exists', 409, 'DUPLICATE_RESOURCE');
  }

  if (error.message && error.message.includes('not found')) {
    return ErrorHelper.createError('Resource not found', 404, 'RESOURCE_NOT_FOUND');
  }

  return ErrorHelper.createError('Database operation failed', 500, 'DB_OPERATION_ERROR');
};

/**
 * External service error handler
 */
const handleExternalServiceError = (error, serviceName) => {
  logger.error(`${serviceName} service error:`, error);

  if (error.code === 'ECONNREFUSED' || error.code === 'ENOTFOUND') {
    return ErrorHelper.createError(`${serviceName} service unavailable`, 503, 'EXTERNAL_SERVICE_UNAVAILABLE');
  }

  if (error.code === 'ETIMEDOUT') {
    return ErrorHelper.createError(`${serviceName} service timeout`, 504, 'EXTERNAL_SERVICE_TIMEOUT');
  }

  if (error.response && error.response.status >= 400) {
    const statusCode = error.response.status;
    const message = error.response.data?.message || `${serviceName} service error`;
    return ErrorHelper.createError(message, statusCode, 'EXTERNAL_SERVICE_ERROR');
  }

  return ErrorHelper.createError(`${serviceName} service error`, 500, 'EXTERNAL_SERVICE_ERROR');
};

/**
 * File operation error handler
 */
const handleFileError = (error) => {
  logger.error('File operation error:', error);

  if (error.code === 'ENOENT') {
    return ErrorHelper.createError('File not found', 404, 'FILE_NOT_FOUND');
  }

  if (error.code === 'EACCES') {
    return ErrorHelper.createError('File access denied', 403, 'FILE_ACCESS_DENIED');
  }

  if (error.code === 'ENOSPC') {
    return ErrorHelper.createError('No space left on device', 507, 'INSUFFICIENT_STORAGE');
  }

  return ErrorHelper.createError('File operation failed', 500, 'FILE_OPERATION_ERROR');
};

/**
 * Graceful shutdown handler
 */
const gracefulShutdown = (server) => {
  const shutdown = (signal) => {
    logger.info(`Received ${signal}. Shutting down gracefully...`);

    server.close(() => {
      logger.info('HTTP server closed');
      
      // Close database connections, clear timers, etc.
      process.exit(0);
    });

    // Force shutdown after 10 seconds
    setTimeout(() => {
      logger.error('Could not close connections in time, forcefully shutting down');
      process.exit(1);
    }, 10000);
  };

  // Listen for termination signals
  process.on('SIGTERM', () => shutdown('SIGTERM'));
  process.on('SIGINT', () => shutdown('SIGINT'));
};

/**
 * Unhandled rejection handler
 */
const handleUnhandledRejection = () => {
  process.on('unhandledRejection', (reason, promise) => {
    logger.error('Unhandled Rejection at:', promise, 'reason:', reason);
    
    // Optionally exit the process
    if (config.server.environment === 'production') {
      process.exit(1);
    }
  });
};

/**
 * Uncaught exception handler
 */
const handleUncaughtException = () => {
  process.on('uncaughtException', (error) => {
    logger.error('Uncaught Exception:', error);
    
    // In production, exit gracefully
    if (config.server.environment === 'production') {
      process.exit(1);
    }
  });
};

module.exports = {
  errorHandler,
  notFoundHandler,
  asyncHandler,
  timeoutHandler,
  formatValidationError,
  handleDatabaseError,
  handleExternalServiceError,
  handleFileError,
  gracefulShutdown,
  handleUnhandledRejection,
  handleUncaughtException,
};
