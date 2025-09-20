// Structured Logging with Correlation IDs for Microservices
// Implements distributed tracing and comprehensive monitoring

import winston, { Logger } from 'winston';
import { v4 as uuidv4 } from 'uuid';

export interface LogContext {
  correlationId?: string;
  userId?: string;
  requestId?: string;
  sessionId?: string;
  operation?: string;
  component?: string;
  ipAddress?: string;
  userAgent?: string;
  method?: string;
  path?: string;
  statusCode?: number;
  responseTime?: number;
  errorCode?: string;
  contentLength?: string;
  query?: any;
  headers?: any;
  timestamp?: string;
  metadata?: Record<string, any>;
}

export interface SecurityLogContext extends LogContext {
  eventType: 'LOGIN_ATTEMPT' | 'LOGIN_SUCCESS' | 'LOGIN_FAILED' | 'PASSWORD_CHANGE' | 
             'ACCOUNT_LOCKED' | 'SUSPICIOUS_ACTIVITY' | 'TOKEN_ISSUED' | 'TOKEN_REVOKED';
  severity: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  riskScore?: number;
  location?: {
    country?: string;
    city?: string;
    latitude?: number;
    longitude?: number;
  };
}

export interface PerformanceLogContext extends LogContext {
  responseTime: number;
  memoryUsage?: number;
  cpuUsage?: number;
  activeConnections?: number;
  databaseQueryTime?: number;
  externalServiceTime?: number;
}

export interface AuditLogContext extends LogContext {
  action: string;
  resource: string;
  resourceId?: string;
  changes?: Record<string, { old: any; new: any }>;
  reason?: string;
}

export class StructuredLogger {
  private logger: Logger;
  private context: LogContext = {};

  constructor(options?: winston.LoggerOptions) {
    this.logger = winston.createLogger({
      level: process.env.LOG_LEVEL || 'info',
      format: winston.format.combine(
        winston.format.timestamp(),
        winston.format.errors({ stack: true }),
        winston.format.json(),
        winston.format.printf((info: any) => {
          const { timestamp, level, message, ...meta } = info;
          return JSON.stringify({
            timestamp,
            level: level.toUpperCase(),
            message,
            service: 'auth-service',
            version: process.env.npm_package_version || '1.0.0',
            environment: process.env.NODE_ENV || 'development',
            correlationId: this.context.correlationId,
            ...meta,
          });
        })
      ),
      transports: [
        new winston.transports.Console({
          format: winston.format.combine(
            winston.format.colorize(),
            winston.format.simple()
          ),
        }),
        new winston.transports.File({
          filename: 'logs/error.log',
          level: 'error',
        }),
        new winston.transports.File({
          filename: 'logs/combined.log',
        }),
      ],
      ...options,
    });

    // Add production transports
    if (process.env.NODE_ENV === 'production') {
      this.logger.add(new winston.transports.File({
        filename: 'logs/audit.log',
        level: 'info',
        format: winston.format.combine(
          winston.format.timestamp(),
          winston.format.json()
        ),
      }));
    }
  }

  setContext(context: Partial<LogContext>): void {
    this.context = { ...this.context, ...context };
  }

  clearContext(): void {
    this.context = {};
  }

  updateContext(updates: Partial<LogContext>): void {
    this.context = { ...this.context, ...updates };
  }

  info(message: string, context?: LogContext): void {
    this.logger.info(message, { ...this.context, ...context });
  }

  warn(message: string, context?: LogContext): void {
    this.logger.warn(message, { ...this.context, ...context });
  }

  error(message: string, error?: Error, context?: LogContext): void {
    const errorContext = error ? {
      error: {
        name: error.name,
        message: error.message,
        stack: error.stack,
      },
    } : {};

    this.logger.error(message, { 
      ...this.context, 
      ...context, 
      ...errorContext 
    });
  }

  debug(message: string, context?: LogContext): void {
    this.logger.debug(message, { ...this.context, ...context });
  }

  // Security-specific logging
  security(message: string, context: SecurityLogContext): void {
    this.logger.warn(message, {
      ...this.context,
      ...context,
      logType: 'SECURITY',
      timestamp: new Date().toISOString(),
    });
  }

  // Performance logging
  performance(message: string, context: PerformanceLogContext): void {
    this.logger.info(message, {
      ...this.context,
      ...context,
      logType: 'PERFORMANCE',
      timestamp: new Date().toISOString(),
    });
  }

  // Audit logging
  audit(message: string, context: AuditLogContext): void {
    this.logger.info(message, {
      ...this.context,
      ...context,
      logType: 'AUDIT',
      timestamp: new Date().toISOString(),
    });
  }

  // HTTP request logging
  httpRequest(req: any, res: any, responseTime: number): void {
    const context: LogContext = {
      method: req.method,
      path: req.path,
      statusCode: res.statusCode,
      responseTime,
      ipAddress: req.ip,
      userAgent: req.get('User-Agent'),
      contentLength: res.get('Content-Length'),
      metadata: {
        query: req.query,
        params: req.params,
      },
    };

    if (res.statusCode >= 400) {
      this.warn(`HTTP ${res.statusCode} - ${req.method} ${req.path}`, context);
    } else {
      this.info(`HTTP ${res.statusCode} - ${req.method} ${req.path}`, context);
    }
  }

  // Database operation logging
  databaseOperation(operation: string, table: string, duration: number, error?: Error): void {
    const context: LogContext = {
      operation: 'database',
      component: 'postgresql',
      metadata: {
        operation,
        table,
        duration,
      },
    };

    if (error) {
      this.error(`Database operation failed: ${operation} on ${table}`, error, context);
    } else {
      this.debug(`Database operation completed: ${operation} on ${table} (${duration}ms)`, context);
    }
  }

  // External service call logging
  externalServiceCall(service: string, endpoint: string, duration: number, success: boolean, error?: Error): void {
    const context: LogContext = {
      operation: 'external_service',
      component: service,
      metadata: {
        endpoint,
        duration,
        success,
      },
    };

    if (!success && error) {
      this.error(`External service call failed: ${service}${endpoint}`, error, context);
    } else {
      this.info(`External service call: ${service}${endpoint} (${duration}ms)`, context);
    }
  }

  // Business event logging
  businessEvent(event: string, entityType: string, entityId: string, details?: any): void {
    this.info(`Business event: ${event}`, {
      operation: 'business_event',
      metadata: {
        event,
        entityType,
        entityId,
        details,
      },
    });
  }
}

// Singleton logger instance
export const logger = new StructuredLogger();

// Express middleware for request logging and correlation ID
export const requestLoggingMiddleware = (req: any, res: any, next: any): void => {
  const start = Date.now();
  
  // Generate or extract correlation ID
  const correlationId = req.headers['x-correlation-id'] || uuidv4();
  req.correlationId = correlationId;
  res.set('X-Correlation-ID', correlationId);

  // Set correlation context for this request
  logger.setContext({
    correlationId,
    requestId: uuidv4(),
    ipAddress: req.ip,
    userAgent: req.get('User-Agent'),
  });

  // Log request start
  logger.info(`Request started: ${req.method} ${req.path}`, {
    method: req.method,
    path: req.path,
    query: req.query,
    headers: {
      'content-type': req.get('Content-Type'),
      'accept': req.get('Accept'),
      'user-agent': req.get('User-Agent'),
    },
  });

  // Override res.end to log response
  const originalEnd = res.end;
  res.end = function(chunk: any, encoding: any) {
    res.end = originalEnd;
    res.end(chunk, encoding);
    
    const responseTime = Date.now() - start;
    logger.httpRequest(req, res, responseTime);
    
    // Clear context after request
    logger.clearContext();
  };

  next();
};

// Error logging middleware
export const errorLoggingMiddleware = (error: Error, req: any, res: any, next: any): void => {
  logger.error('Unhandled error in request', error, {
    method: req.method,
    path: req.path,
    statusCode: res.statusCode,
    correlationId: req.correlationId,
  });
  
  next(error);
};

// Security event helpers
export const logSecurityEvent = (
  eventType: SecurityLogContext['eventType'],
  severity: SecurityLogContext['severity'],
  message: string,
  context?: Partial<SecurityLogContext>
): void => {
  logger.security(message, {
    eventType,
    severity,
    timestamp: new Date().toISOString(),
    ...context,
  });
};

export const logLoginAttempt = (email: string, success: boolean, req: any): void => {
  logSecurityEvent(
    success ? 'LOGIN_SUCCESS' : 'LOGIN_FAILED',
    success ? 'LOW' : 'MEDIUM',
    `Login ${success ? 'successful' : 'failed'} for ${email}`,
    {
      userId: success ? 'user-id' : undefined,
      ipAddress: req.ip,
      userAgent: req.get('User-Agent'),
      metadata: { email },
    }
  );
};

export const logPasswordChange = (userId: string, req: any): void => {
  logSecurityEvent(
    'PASSWORD_CHANGE',
    'MEDIUM',
    `Password changed for user ${userId}`,
    {
      userId,
      ipAddress: req.ip,
      userAgent: req.get('User-Agent'),
    }
  );
};

export const logAccountLocked = (userId: string, reason: string): void => {
  logSecurityEvent(
    'ACCOUNT_LOCKED',
    'HIGH',
    `Account locked for user ${userId}: ${reason}`,
    {
      userId,
      metadata: { reason },
    }
  );
};

export const logSuspiciousActivity = (description: string, req: any, riskScore?: number): void => {
  logSecurityEvent(
    'SUSPICIOUS_ACTIVITY',
    riskScore && riskScore > 7 ? 'CRITICAL' : 'HIGH',
    `Suspicious activity detected: ${description}`,
    {
      ipAddress: req.ip,
      userAgent: req.get('User-Agent'),
      riskScore,
      metadata: { description },
    }
  );
};

// Performance monitoring helpers
export const logPerformanceMetric = (
  operation: string,
  responseTime: number,
  additionalMetrics?: Partial<PerformanceLogContext>
): void => {
  logger.performance(`Performance metric: ${operation}`, {
    operation,
    responseTime,
    ...additionalMetrics,
  });
};

// Audit logging helpers
export const logAuditEvent = (
  action: string,
  resource: string,
  userId?: string,
  changes?: Record<string, { old: any; new: any }>,
  reason?: string
): void => {
  logger.audit(`Audit: ${action} on ${resource}`, {
    action,
    resource,
    userId,
    changes,
    reason,
  });
};
