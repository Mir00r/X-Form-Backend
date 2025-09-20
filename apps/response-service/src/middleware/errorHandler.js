/**
 * Error Handler Middleware for Response Service
 * Comprehensive error handling with proper logging and response formatting
 */

const { createErrorResponse } = require('../dto/response-dtos');
const logger = require('../utils/logger');

/**
 * Async handler wrapper to catch async errors
 * @param {Function} fn - Async function to wrap
 * @returns {Function} Express middleware function
 */
const asyncHandler = (fn) => {
  return (req, res, next) => {
    Promise.resolve(fn(req, res, next)).catch(next);
  };
};

/**
 * Not Found Error Handler
 * Handles 404 errors for undefined routes
 */
const notFoundHandler = (req, res, next) => {
  const error = new Error(`Route not found: ${req.method} ${req.originalUrl}`);
  error.status = 404;
  error.code = 'ROUTE_NOT_FOUND';
  next(error);
};

/**
 * Global Error Handler
 * Centralized error handling for all application errors
 */
const globalErrorHandler = (err, req, res, next) => {
  // Set default error properties
  err.status = err.status || 500;
  err.code = err.code || 'INTERNAL_SERVER_ERROR';

  // Log error details
  logger.error('Error occurred:', {
    error: err.message,
    code: err.code,
    status: err.status,
    stack: err.stack,
    url: req.originalUrl,
    method: req.method,
    ip: req.ip,
    userAgent: req.get('User-Agent'),
    correlationId: req.headers['x-correlation-id']
  });

  // Don't expose internal errors in production
  const isDevelopment = process.env.NODE_ENV === 'development';
  
  // Create error response
  const errorResponse = createErrorResponse(
    err.message || 'An unexpected error occurred',
    err.code,
    err.status,
    isDevelopment ? err.stack : undefined,
    req.headers['x-correlation-id']
  );

  // Handle specific error types
  if (err.name === 'ValidationError') {
    err.status = 400;
    err.code = 'VALIDATION_ERROR';
  } else if (err.name === 'JsonWebTokenError') {
    err.status = 401;
    err.code = 'INVALID_TOKEN';
  } else if (err.name === 'TokenExpiredError') {
    err.status = 401;
    err.code = 'TOKEN_EXPIRED';
  } else if (err.name === 'CastError') {
    err.status = 400;
    err.code = 'INVALID_ID_FORMAT';
  }

  // Send error response
  res.status(err.status).json(errorResponse);
};

/**
 * Request validation error handler
 * Handles express-validator errors
 */
const validationErrorHandler = (errors) => {
  const errorDetails = errors.array().map(error => ({
    field: error.path || error.param,
    message: error.msg,
    value: error.value,
    code: 'VALIDATION_ERROR'
  }));

  const error = new Error('Validation failed');
  error.status = 400;
  error.code = 'VALIDATION_ERROR';
  error.details = errorDetails;
  
  return error;
};

/**
 * Create custom error with specific properties
 * @param {string} message - Error message
 * @param {number} status - HTTP status code
 * @param {string} code - Error code
 * @returns {Error} Custom error object
 */
const createError = (message, status = 500, code = 'INTERNAL_SERVER_ERROR') => {
  const error = new Error(message);
  error.status = status;
  error.code = code;
  return error;
};

module.exports = {
  asyncHandler,
  notFoundHandler,
  globalErrorHandler,
  validationErrorHandler,
  createError
};
