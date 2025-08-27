/**
 * Security Middleware for Response Service
 * Implements comprehensive security measures including rate limiting, CORS, and security headers
 */

const rateLimit = require('express-rate-limit');
const slowDown = require('express-slow-down');
const helmet = require('helmet');
const cors = require('cors');
const { createErrorResponse } = require('../dto/response-dtos');
const logger = require('../utils/logger');

/**
 * CORS Configuration
 */
const corsOptions = {
  origin: function (origin, callback) {
    // Allow requests with no origin (like mobile apps or curl requests)
    if (!origin) return callback(null, true);
    
    const allowedOrigins = (process.env.ALLOWED_ORIGINS || 'http://localhost:3000,http://localhost:3001').split(',');
    
    if (allowedOrigins.indexOf(origin) !== -1 || process.env.NODE_ENV === 'development') {
      callback(null, true);
    } else {
      logger.logSecurity('CORS_BLOCKED', { origin });
      callback(new Error('Not allowed by CORS'));
    }
  },
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS'],
  allowedHeaders: [
    'Origin',
    'X-Requested-With',
    'Content-Type',
    'Accept',
    'Authorization',
    'X-API-Key',
    'X-Correlation-ID',
    'X-Request-ID'
  ],
  exposedHeaders: [
    'X-Correlation-ID',
    'X-Request-ID',
    'X-RateLimit-Limit',
    'X-RateLimit-Remaining',
    'X-RateLimit-Reset'
  ],
  maxAge: 86400 // 24 hours
};

/**
 * Rate Limiting Configuration
 */

// General API rate limit
const apiRateLimit = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: process.env.NODE_ENV === 'production' ? 100 : 1000, // Limit each IP to 100 requests per windowMs in production
  message: (req) => {
    const correlationId = req.headers['x-correlation-id'] || req.correlationId;
    logger.logSecurity('RATE_LIMIT_EXCEEDED', {
      ip: req.ip,
      userAgent: req.get('User-Agent'),
      endpoint: req.path
    }, { correlationId });
    
    return createErrorResponse(
      'RATE_LIMIT_EXCEEDED',
      'Too many requests from this IP, please try again later',
      {
        retryAfter: Math.round(this.windowMs / 1000),
        limit: this.max
      },
      correlationId
    );
  },
  standardHeaders: true, // Return rate limit info in the `RateLimit-*` headers
  legacyHeaders: false, // Disable the `X-RateLimit-*` headers
  handler: (req, res) => {
    const correlationId = req.headers['x-correlation-id'] || req.correlationId;
    const response = createErrorResponse(
      'RATE_LIMIT_EXCEEDED',
      'Too many requests from this IP, please try again later',
      {
        retryAfter: Math.round(req.rateLimit.resetTime / 1000),
        limit: req.rateLimit.limit
      },
      correlationId
    );
    
    res.status(429).json(response);
  },
  skip: (req) => {
    // Skip rate limiting for health checks
    return req.path === '/health' || req.path === '/api/v1/health';
  }
});

// Strict rate limit for form submission endpoints
const submissionRateLimit = rateLimit({
  windowMs: 5 * 60 * 1000, // 5 minutes
  max: process.env.NODE_ENV === 'production' ? 20 : 100, // 20 submissions per 5 minutes in production
  keyGenerator: (req) => {
    // Rate limit by user ID if authenticated, otherwise by IP
    return req.user?.id || req.ip;
  },
  message: (req) => {
    const correlationId = req.headers['x-correlation-id'] || req.correlationId;
    logger.logSecurity('SUBMISSION_RATE_LIMIT_EXCEEDED', {
      userId: req.user?.id,
      ip: req.ip,
      userAgent: req.get('User-Agent'),
      endpoint: req.path
    }, { correlationId });
    
    return createErrorResponse(
      'SUBMISSION_RATE_LIMIT_EXCEEDED',
      'Too many form submissions, please slow down',
      {
        retryAfter: Math.round(this.windowMs / 1000),
        limit: this.max
      },
      correlationId
    );
  },
  standardHeaders: true,
  legacyHeaders: false
});

// Slow down middleware for repeated requests
const speedLimiter = slowDown({
  windowMs: 15 * 60 * 1000, // 15 minutes
  delayAfter: 50, // Allow 50 requests per windowMs without delay
  delayMs: 500, // Add 500ms delay per request after delayAfter
  maxDelayMs: 5000, // Maximum delay of 5 seconds
  skip: (req) => {
    return req.path === '/health' || req.path === '/api/v1/health';
  }
});

/**
 * Security Headers Configuration
 */
const securityHeaders = helmet({
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      styleSrc: ["'self'", "'unsafe-inline'"],
      scriptSrc: ["'self'"],
      imgSrc: ["'self'", "data:", "https:"],
      connectSrc: ["'self'"],
      fontSrc: ["'self'"],
      objectSrc: ["'none'"],
      mediaSrc: ["'self'"],
      frameSrc: ["'none'"],
      baseUri: ["'self'"],
      formAction: ["'self'"]
    }
  },
  crossOriginEmbedderPolicy: false, // Disable for API service
  hsts: {
    maxAge: 31536000, // 1 year
    includeSubDomains: true,
    preload: true
  },
  noSniff: true,
  frameguard: { action: 'deny' },
  xssFilter: true,
  referrerPolicy: { policy: "strict-origin-when-cross-origin" }
});

/**
 * Request size limiting middleware
 */
const requestSizeLimit = (req, res, next) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  const contentLength = parseInt(req.get('Content-Length') || '0', 10);
  const maxSize = 10 * 1024 * 1024; // 10MB
  
  if (contentLength > maxSize) {
    logger.logSecurity('REQUEST_SIZE_EXCEEDED', {
      contentLength,
      maxSize,
      ip: req.ip,
      endpoint: req.path
    }, { correlationId });
    
    return res.status(413).json(
      createErrorResponse(
        'REQUEST_TOO_LARGE',
        'Request payload too large',
        {
          maxSize: `${maxSize / 1024 / 1024}MB`,
          received: `${contentLength / 1024 / 1024}MB`
        },
        correlationId
      )
    );
  }
  
  next();
};

/**
 * IP Whitelist/Blacklist middleware
 */
const ipFilter = (req, res, next) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  const clientIP = req.ip;
  
  // IP blacklist check
  const blacklistedIPs = (process.env.BLACKLISTED_IPS || '').split(',').filter(Boolean);
  if (blacklistedIPs.includes(clientIP)) {
    logger.logSecurity('BLACKLISTED_IP_ACCESS', {
      ip: clientIP,
      userAgent: req.get('User-Agent'),
      endpoint: req.path
    }, { correlationId });
    
    return res.status(403).json(
      createErrorResponse(
        'ACCESS_DENIED',
        'Access denied from this IP address',
        null,
        correlationId
      )
    );
  }
  
  // IP whitelist check (if configured)
  const whitelistedIPs = (process.env.WHITELISTED_IPS || '').split(',').filter(Boolean);
  if (whitelistedIPs.length > 0 && !whitelistedIPs.includes(clientIP)) {
    logger.logSecurity('NON_WHITELISTED_IP_ACCESS', {
      ip: clientIP,
      userAgent: req.get('User-Agent'),
      endpoint: req.path
    }, { correlationId });
    
    return res.status(403).json(
      createErrorResponse(
        'ACCESS_DENIED',
        'Access denied - IP not whitelisted',
        null,
        correlationId
      )
    );
  }
  
  next();
};

/**
 * User-Agent validation middleware
 */
const userAgentFilter = (req, res, next) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  const userAgent = req.get('User-Agent');
  
  // Block requests without User-Agent (potential bots)
  if (!userAgent && process.env.REQUIRE_USER_AGENT === 'true') {
    logger.logSecurity('MISSING_USER_AGENT', {
      ip: req.ip,
      endpoint: req.path
    }, { correlationId });
    
    return res.status(400).json(
      createErrorResponse(
        'MISSING_USER_AGENT',
        'User-Agent header is required',
        null,
        correlationId
      )
    );
  }
  
  // Block suspicious User-Agents
  const suspiciousPatterns = [
    /bot/i,
    /crawler/i,
    /spider/i,
    /scraper/i,
    /scanner/i
  ];
  
  if (userAgent && process.env.BLOCK_BOTS === 'true') {
    const isSuspicious = suspiciousPatterns.some(pattern => pattern.test(userAgent));
    if (isSuspicious) {
      logger.logSecurity('SUSPICIOUS_USER_AGENT', {
        ip: req.ip,
        userAgent,
        endpoint: req.path
      }, { correlationId });
      
      return res.status(403).json(
        createErrorResponse(
          'ACCESS_DENIED',
          'Access denied',
          null,
          correlationId
        )
      );
    }
  }
  
  next();
};

/**
 * Referrer validation middleware
 */
const referrerFilter = (req, res, next) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  const referrer = req.get('Referer') || req.get('Referrer');
  
  // Skip validation for API endpoints and development
  if (req.path.startsWith('/api/') || process.env.NODE_ENV === 'development') {
    return next();
  }
  
  const allowedReferrers = (process.env.ALLOWED_REFERRERS || '').split(',').filter(Boolean);
  
  if (allowedReferrers.length > 0 && referrer) {
    const isAllowed = allowedReferrers.some(allowed => referrer.includes(allowed));
    
    if (!isAllowed) {
      logger.logSecurity('INVALID_REFERRER', {
        ip: req.ip,
        referrer,
        endpoint: req.path
      }, { correlationId });
      
      return res.status(403).json(
        createErrorResponse(
          'ACCESS_DENIED',
          'Access denied - invalid referrer',
          null,
          correlationId
        )
      );
    }
  }
  
  next();
};

/**
 * Request method validation middleware
 */
const methodFilter = (req, res, next) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  const allowedMethods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS', 'HEAD'];
  
  if (!allowedMethods.includes(req.method)) {
    logger.logSecurity('INVALID_HTTP_METHOD', {
      ip: req.ip,
      method: req.method,
      endpoint: req.path
    }, { correlationId });
    
    return res.status(405).json(
      createErrorResponse(
        'METHOD_NOT_ALLOWED',
        `HTTP method ${req.method} is not allowed`,
        { allowedMethods },
        correlationId
      )
    );
  }
  
  next();
};

/**
 * SQL Injection detection middleware
 */
const sqlInjectionFilter = (req, res, next) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  
  // Common SQL injection patterns
  const sqlPatterns = [
    /(\b(SELECT|INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|EXEC|UNION|SCRIPT)\b)/i,
    /((\%27)|(\'))/i, // SQL meta-characters
    /((\%6F)|o|(\%4F))((\%72)|r|(\%52))/i, // "or" variations
    /((\%27)|(\'))union/i,
    /(exec(\s|\+)+(s|x)p\w+)/i
  ];
  
  const checkForSQLInjection = (value) => {
    if (typeof value === 'string') {
      return sqlPatterns.some(pattern => pattern.test(value));
    }
    if (typeof value === 'object' && value !== null) {
      return Object.values(value).some(checkForSQLInjection);
    }
    return false;
  };
  
  // Check query parameters
  if (checkForSQLInjection(req.query)) {
    logger.logSecurity('SQL_INJECTION_ATTEMPT', {
      ip: req.ip,
      userAgent: req.get('User-Agent'),
      query: req.query,
      endpoint: req.path
    }, { correlationId });
    
    return res.status(400).json(
      createErrorResponse(
        'INVALID_INPUT',
        'Invalid characters detected in request',
        null,
        correlationId
      )
    );
  }
  
  // Check request body
  if (req.body && checkForSQLInjection(req.body)) {
    logger.logSecurity('SQL_INJECTION_ATTEMPT', {
      ip: req.ip,
      userAgent: req.get('User-Agent'),
      body: req.body,
      endpoint: req.path
    }, { correlationId });
    
    return res.status(400).json(
      createErrorResponse(
        'INVALID_INPUT',
        'Invalid characters detected in request body',
        null,
        correlationId
      )
    );
  }
  
  next();
};

module.exports = {
  corsOptions,
  apiRateLimit,
  submissionRateLimit,
  speedLimiter,
  securityHeaders,
  requestSizeLimit,
  ipFilter,
  userAgentFilter,
  referrerFilter,
  methodFilter,
  sqlInjectionFilter,
  
  // Combined security middleware
  applySecurity: (app) => {
    app.use(securityHeaders);
    app.use(cors(corsOptions));
    app.use(methodFilter);
    app.use(ipFilter);
    app.use(userAgentFilter);
    app.use(referrerFilter);
    app.use(requestSizeLimit);
    app.use(sqlInjectionFilter);
    app.use(speedLimiter);
    app.use(apiRateLimit);
  }
};
