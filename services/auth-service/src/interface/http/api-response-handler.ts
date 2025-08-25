// Standardized API Response Handler for Microservices
// Implements consistent response format across all endpoints

import { Response } from 'express';
import { v4 as uuidv4 } from 'uuid';
import {
  BaseApiResponse,
  ApiError,
  ValidationError,
  ResponseMeta,
  RateLimitMeta,
  PaginationMeta,
} from '../dto/auth-dtos';

export class ApiResponseHandler {
  private static readonly API_VERSION = 'v1';

  /**
   * Send successful response with data
   */
  static success<T>(
    res: Response,
    data: T,
    statusCode: number = 200,
    meta?: Partial<ResponseMeta>
  ): void {
    const response: BaseApiResponse<T> = {
      success: true,
      timestamp: new Date().toISOString(),
      path: res.req.path,
      method: res.req.method,
      correlationId: this.getCorrelationId(res),
      data,
      meta: {
        version: this.API_VERSION,
        ...meta,
      },
    };

    res.status(statusCode).json(response);
  }

  /**
   * Send created response (201)
   */
  static created<T>(res: Response, data: T, meta?: Partial<ResponseMeta>): void {
    this.success(res, data, 201, meta);
  }

  /**
   * Send no content response (204)
   */
  static noContent(res: Response): void {
    res.status(204).send();
  }

  /**
   * Send error response
   */
  static error(
    res: Response,
    error: ApiError,
    statusCode: number = 500
  ): void {
    const response: BaseApiResponse = {
      success: false,
      timestamp: new Date().toISOString(),
      path: res.req.path,
      method: res.req.method,
      correlationId: this.getCorrelationId(res),
      error: {
        ...error,
        timestamp: new Date().toISOString(),
        path: res.req.path,
        correlationId: this.getCorrelationId(res),
      },
      meta: {
        version: this.API_VERSION,
      },
    };

    res.status(statusCode).json(response);
  }

  /**
   * Send validation error response (400)
   */
  static validationError(
    res: Response,
    errors: ValidationError[],
    message: string = 'Validation failed'
  ): void {
    this.error(
      res,
      {
        code: 'VALIDATION_ERROR',
        message,
        details: errors,
        timestamp: new Date().toISOString(),
        path: res.req.path,
        correlationId: this.getCorrelationId(res),
      },
      400
    );
  }

  /**
   * Send authentication error response (401)
   */
  static unauthorized(
    res: Response,
    message: string = 'Authentication required',
    code: string = 'AUTHENTICATION_REQUIRED'
  ): void {
    this.error(
      res,
      {
        code,
        message,
        timestamp: new Date().toISOString(),
        path: res.req.path,
        correlationId: this.getCorrelationId(res),
      },
      401
    );
  }

  /**
   * Send authorization error response (403)
   */
  static forbidden(
    res: Response,
    message: string = 'Insufficient permissions',
    code: string = 'ACCESS_DENIED'
  ): void {
    this.error(
      res,
      {
        code,
        message,
        timestamp: new Date().toISOString(),
        path: res.req.path,
        correlationId: this.getCorrelationId(res),
      },
      403
    );
  }

  /**
   * Send not found error response (404)
   */
  static notFound(
    res: Response,
    resource: string = 'Resource',
    identifier?: string
  ): void {
    this.error(
      res,
      {
        code: 'RESOURCE_NOT_FOUND',
        message: `${resource} not found${identifier ? ` with identifier: ${identifier}` : ''}`,
        details: { resource, identifier },
        timestamp: new Date().toISOString(),
        path: res.req.path,
        correlationId: this.getCorrelationId(res),
      },
      404
    );
  }

  /**
   * Send conflict error response (409)
   */
  static conflict(
    res: Response,
    message: string,
    code: string = 'RESOURCE_CONFLICT'
  ): void {
    this.error(
      res,
      {
        code,
        message,
        timestamp: new Date().toISOString(),
        path: res.req.path,
        correlationId: this.getCorrelationId(res),
      },
      409
    );
  }

  /**
   * Send rate limit error response (429)
   */
  static rateLimitExceeded(
    res: Response,
    retryAfter: number,
    limit: number,
    remaining: number = 0
  ): void {
    const resetTime = new Date(Date.now() + retryAfter * 1000).toISOString();
    
    // Set rate limit headers
    res.set({
      'X-RateLimit-Limit': limit.toString(),
      'X-RateLimit-Remaining': remaining.toString(),
      'X-RateLimit-Reset': Math.floor((Date.now() + retryAfter * 1000) / 1000).toString(),
      'Retry-After': retryAfter.toString(),
    });

    this.error(
      res,
      {
        code: 'RATE_LIMIT_EXCEEDED',
        message: 'Too many requests. Please try again later.',
        details: {
          retryAfter,
          limit,
          remaining,
          resetTime,
        },
        timestamp: new Date().toISOString(),
        path: res.req.path,
        correlationId: this.getCorrelationId(res),
      },
      429
    );
  }

  /**
   * Send internal server error response (500)
   */
  static internalError(
    res: Response,
    message: string = 'Internal server error',
    errorId?: string
  ): void {
    this.error(
      res,
      {
        code: 'INTERNAL_SERVER_ERROR',
        message,
        details: {
          errorId: errorId || uuidv4(),
          reportedAt: new Date().toISOString(),
        },
        timestamp: new Date().toISOString(),
        path: res.req.path,
        correlationId: this.getCorrelationId(res),
      },
      500
    );
  }

  /**
   * Send service unavailable error response (503)
   */
  static serviceUnavailable(
    res: Response,
    message: string = 'Service temporarily unavailable',
    retryAfter?: number
  ): void {
    if (retryAfter) {
      res.set('Retry-After', retryAfter.toString());
    }

    this.error(
      res,
      {
        code: 'SERVICE_UNAVAILABLE',
        message,
        details: { retryAfter },
        timestamp: new Date().toISOString(),
        path: res.req.path,
        correlationId: this.getCorrelationId(res),
      },
      503
    );
  }

  /**
   * Send paginated response
   */
  static paginated<T>(
    res: Response,
    data: T[],
    pagination: PaginationMeta,
    statusCode: number = 200
  ): void {
    this.success(res, data, statusCode, { pagination });
  }

  /**
   * Create response meta with rate limit info
   */
  static withRateLimit(
    limit: number,
    remaining: number,
    resetTime: string
  ): { rateLimit: RateLimitMeta } {
    return {
      rateLimit: {
        limit,
        remaining,
        resetTime,
      },
    };
  }

  /**
   * Get or generate correlation ID from request
   */
  private static getCorrelationId(res: Response): string {
    return res.locals.correlationId || res.req.headers['x-correlation-id'] as string || uuidv4();
  }

  /**
   * Set standard headers for all responses
   */
  static setStandardHeaders(res: Response): void {
    res.set({
      'X-API-Version': this.API_VERSION,
      'X-Powered-By': 'X-Form Auth Service',
      'X-Content-Type-Options': 'nosniff',
      'X-Frame-Options': 'DENY',
      'X-XSS-Protection': '1; mode=block',
    });
  }
}

/**
 * Middleware to set correlation ID and standard headers
 */
export const correlationMiddleware = (req: any, res: any, next: any): void => {
  // Generate or use existing correlation ID
  const correlationId = req.headers['x-correlation-id'] || uuidv4();
  res.locals.correlationId = correlationId;
  
  // Set correlation ID in response header
  res.set('X-Correlation-ID', correlationId);
  
  // Set standard headers
  ApiResponseHandler.setStandardHeaders(res);
  
  next();
};

/**
 * Helper to extract validation errors from express-validator
 */
export const extractValidationErrors = (errors: any[]): ValidationError[] => {
  return errors.map(error => ({
    field: error.path || error.param,
    message: error.msg,
    value: error.value,
    code: error.type || 'VALIDATION_ERROR',
  }));
};
