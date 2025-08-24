const logger = require('./logger');

/**
 * Standard response format for API responses
 */
class ResponseFormatter {
  /**
   * Success response
   * @param {Object} res - Express response object
   * @param {*} data - Response data
   * @param {string} message - Success message
   * @param {number} statusCode - HTTP status code
   */
  static success(res, data = null, message = 'Success', statusCode = 200) {
    const response = {
      success: true,
      message,
      data,
      timestamp: new Date().toISOString(),
    };

    if (data && typeof data === 'object' && data.pagination) {
      response.pagination = data.pagination;
      response.data = data.items || data.data;
    }

    res.status(statusCode).json(response);
  }

  /**
   * Error response
   * @param {Object} res - Express response object
   * @param {string} message - Error message
   * @param {number} statusCode - HTTP status code
   * @param {*} errors - Detailed error information
   */
  static error(res, message = 'Internal Server Error', statusCode = 500, errors = null) {
    const response = {
      success: false,
      message,
      timestamp: new Date().toISOString(),
    };

    if (errors) {
      response.errors = errors;
    }

    res.status(statusCode).json(response);
  }

  /**
   * Validation error response
   * @param {Object} res - Express response object
   * @param {Array} validationErrors - Array of validation errors
   */
  static validationError(res, validationErrors) {
    const formattedErrors = validationErrors.map(error => ({
      field: error.path || error.field,
      message: error.message,
      value: error.value,
    }));

    this.error(res, 'Validation failed', 400, formattedErrors);
  }

  /**
   * Not found response
   * @param {Object} res - Express response object
   * @param {string} resource - Resource name
   */
  static notFound(res, resource = 'Resource') {
    this.error(res, `${resource} not found`, 404);
  }

  /**
   * Unauthorized response
   * @param {Object} res - Express response object
   * @param {string} message - Custom message
   */
  static unauthorized(res, message = 'Unauthorized') {
    this.error(res, message, 401);
  }

  /**
   * Forbidden response
   * @param {Object} res - Express response object
   * @param {string} message - Custom message
   */
  static forbidden(res, message = 'Forbidden') {
    this.error(res, message, 403);
  }

  /**
   * Conflict response
   * @param {Object} res - Express response object
   * @param {string} message - Custom message
   */
  static conflict(res, message = 'Conflict') {
    this.error(res, message, 409);
  }

  /**
   * Too Many Requests response
   * @param {Object} res - Express response object
   * @param {string} message - Custom message
   */
  static tooManyRequests(res, message = 'Too many requests') {
    this.error(res, message, 429);
  }
}

/**
 * Pagination helper
 */
class PaginationHelper {
  /**
   * Create pagination metadata
   * @param {number} page - Current page
   * @param {number} limit - Items per page
   * @param {number} total - Total items
   */
  static createPagination(page, limit, total) {
    const totalPages = Math.ceil(total / limit);
    const hasNext = page < totalPages;
    const hasPrev = page > 1;

    return {
      currentPage: page,
      totalPages,
      totalItems: total,
      itemsPerPage: limit,
      hasNext,
      hasPrev,
    };
  }

  /**
   * Get offset for database queries
   * @param {number} page - Current page
   * @param {number} limit - Items per page
   */
  static getOffset(page, limit) {
    return (page - 1) * limit;
  }

  /**
   * Validate pagination parameters
   * @param {number} page - Page number
   * @param {number} limit - Items per page
   */
  static validateParams(page, limit) {
    const validatedPage = Math.max(1, parseInt(page) || 1);
    const validatedLimit = Math.min(100, Math.max(1, parseInt(limit) || 10));

    return { page: validatedPage, limit: validatedLimit };
  }
}

/**
 * Cache helper utilities
 */
class CacheHelper {
  /**
   * Generate cache key
   * @param {string} prefix - Key prefix
   * @param {string|Object} identifier - Unique identifier
   */
  static generateKey(prefix, identifier) {
    if (typeof identifier === 'object') {
      identifier = JSON.stringify(identifier);
    }
    return `${prefix}:${identifier}`;
  }

  /**
   * Generate TTL based on type
   * @param {string} type - Cache type
   */
  static getTTL(type) {
    const ttlMap = {
      form: 300, // 5 minutes
      response: 60, // 1 minute
      user: 900, // 15 minutes
      analytics: 1800, // 30 minutes
    };

    return ttlMap[type] || 300;
  }
}

/**
 * Date and time utilities
 */
class DateHelper {
  /**
   * Get start and end of day
   * @param {Date} date - Input date
   */
  static getDayRange(date = new Date()) {
    const start = new Date(date);
    start.setHours(0, 0, 0, 0);

    const end = new Date(date);
    end.setHours(23, 59, 59, 999);

    return { start, end };
  }

  /**
   * Get start and end of week
   * @param {Date} date - Input date
   */
  static getWeekRange(date = new Date()) {
    const start = new Date(date);
    const day = start.getDay();
    const diff = start.getDate() - day + (day === 0 ? -6 : 1); // Monday as first day
    start.setDate(diff);
    start.setHours(0, 0, 0, 0);

    const end = new Date(start);
    end.setDate(start.getDate() + 6);
    end.setHours(23, 59, 59, 999);

    return { start, end };
  }

  /**
   * Get start and end of month
   * @param {Date} date - Input date
   */
  static getMonthRange(date = new Date()) {
    const start = new Date(date.getFullYear(), date.getMonth(), 1);
    const end = new Date(date.getFullYear(), date.getMonth() + 1, 0, 23, 59, 59, 999);

    return { start, end };
  }

  /**
   * Format date for filename
   * @param {Date} date - Input date
   */
  static formatForFilename(date = new Date()) {
    return date.toISOString().split('T')[0].replace(/-/g, '');
  }

  /**
   * Check if date is valid
   * @param {*} date - Date to validate
   */
  static isValidDate(date) {
    return date instanceof Date && !isNaN(date.getTime());
  }
}

/**
 * Validation utilities
 */
class ValidationHelper {
  /**
   * Validate email format
   * @param {string} email - Email to validate
   */
  static isValidEmail(email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  }

  /**
   * Validate URL format
   * @param {string} url - URL to validate
   */
  static isValidUrl(url) {
    try {
      new URL(url);
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Validate UUID format
   * @param {string} uuid - UUID to validate
   */
  static isValidUUID(uuid) {
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
    return uuidRegex.test(uuid);
  }

  /**
   * Sanitize string input
   * @param {string} input - Input to sanitize
   */
  static sanitizeString(input) {
    if (typeof input !== 'string') return input;
    return input.trim().replace(/[<>]/g, '');
  }

  /**
   * Validate file type
   * @param {string} filename - Filename to validate
   * @param {Array} allowedTypes - Allowed file types
   */
  static isValidFileType(filename, allowedTypes = ['csv', 'xlsx', 'json']) {
    const extension = filename.split('.').pop().toLowerCase();
    return allowedTypes.includes(extension);
  }
}

/**
 * Error handling utilities
 */
class ErrorHelper {
  /**
   * Handle async errors
   * @param {Function} fn - Async function to wrap
   */
  static asyncHandler(fn) {
    return (req, res, next) => {
      Promise.resolve(fn(req, res, next)).catch(next);
    };
  }

  /**
   * Create application error
   * @param {string} message - Error message
   * @param {number} statusCode - HTTP status code
   * @param {string} code - Error code
   */
  static createError(message, statusCode = 500, code = null) {
    const error = new Error(message);
    error.statusCode = statusCode;
    error.code = code;
    return error;
  }

  /**
   * Handle database errors
   * @param {Error} error - Database error
   */
  static handleDbError(error) {
    logger.error('Database error:', error);

    if (error.code === 'ECONNREFUSED') {
      return this.createError('Database connection failed', 503, 'DB_CONNECTION_ERROR');
    }

    if (error.message.includes('duplicate')) {
      return this.createError('Resource already exists', 409, 'DUPLICATE_RESOURCE');
    }

    return this.createError('Database operation failed', 500, 'DB_OPERATION_ERROR');
  }
}

module.exports = {
  ResponseFormatter,
  PaginationHelper,
  CacheHelper,
  DateHelper,
  ValidationHelper,
  ErrorHelper,
};
