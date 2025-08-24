const rateLimit = require('express-rate-limit');
const RedisStore = require('rate-limit-redis');
const redis = require('redis');
const config = require('../config');
const { ResponseFormatter } = require('../utils/helpers');
const logger = require('../utils/logger');

// Create Redis client for rate limiting if Redis is configured
let redisClient = null;
if (config.redis.enabled) {
  try {
    redisClient = redis.createClient({
      host: config.redis.host,
      port: config.redis.port,
      password: config.redis.password,
      db: config.redis.db,
    });

    redisClient.on('error', (err) => {
      logger.error('Redis rate limiting error:', err);
    });
  } catch (error) {
    logger.error('Failed to create Redis client for rate limiting:', error);
  }
}

/**
 * Custom rate limit key generator
 */
const keyGenerator = (req) => {
  // Use different keys for different authentication types
  if (req.user) {
    if (req.user.type === 'api_key') {
      return `api_key:${req.user.id}`;
    }
    return `user:${req.user.id}`;
  }
  
  // For anonymous users, use IP address
  return `ip:${req.ip}`;
};

/**
 * Custom rate limit message
 */
const rateLimitMessage = (req, res) => {
  const resetTime = new Date(Date.now() + req.rateLimit.resetTime);
  
  ResponseFormatter.tooManyRequests(res, 'Rate limit exceeded. Please try again later.');
  
  // Add additional headers
  res.setHeader('X-RateLimit-Limit', req.rateLimit.limit);
  res.setHeader('X-RateLimit-Remaining', req.rateLimit.remaining);
  res.setHeader('X-RateLimit-Reset', resetTime.toISOString());
};

/**
 * Rate limit configuration for different tiers
 */
const rateLimitConfigs = {
  // Anonymous users - most restrictive
  anonymous: {
    windowMs: 15 * 60 * 1000, // 15 minutes
    max: 100, // 100 requests per window
    message: rateLimitMessage,
  },
  
  // Standard authenticated users
  standard: {
    windowMs: 15 * 60 * 1000, // 15 minutes
    max: 1000, // 1000 requests per window
    message: rateLimitMessage,
  },
  
  // Premium users
  premium: {
    windowMs: 15 * 60 * 1000, // 15 minutes
    max: 5000, // 5000 requests per window
    message: rateLimitMessage,
  },
  
  // Admin users
  admin: {
    windowMs: 15 * 60 * 1000, // 15 minutes
    max: 10000, // 10000 requests per window
    message: rateLimitMessage,
  },
  
  // API keys - separate limits
  api_key: {
    windowMs: 15 * 60 * 1000, // 15 minutes
    max: 2000, // 2000 requests per window
    message: rateLimitMessage,
  },
};

/**
 * Create rate limiter with Redis store if available
 */
const createRateLimiter = (config) => {
  const limiterConfig = {
    ...config,
    keyGenerator,
    standardHeaders: true,
    legacyHeaders: false,
  };

  // Use Redis store if available
  if (redisClient) {
    limiterConfig.store = new RedisStore({
      sendCommand: (...args) => redisClient.call(...args),
    });
  }

  return rateLimit(limiterConfig);
};

/**
 * Dynamic rate limiter that adjusts based on user tier
 */
const dynamicRateLimit = (req, res, next) => {
  // Determine rate limit tier
  let tier = 'anonymous';
  
  if (req.user) {
    if (req.user.type === 'api_key') {
      tier = 'api_key';
    } else {
      tier = req.rateLimitTier || 'standard';
    }
  }

  // Get rate limit config for tier
  const config = rateLimitConfigs[tier] || rateLimitConfigs.anonymous;
  
  // Create and apply rate limiter
  const limiter = createRateLimiter(config);
  limiter(req, res, (err) => {
    if (err) {
      logger.error('Rate limiting error:', err);
    }
    
    // Log rate limit info
    if (req.rateLimit) {
      logger.debug('Rate limit applied', {
        tier,
        key: keyGenerator(req),
        limit: req.rateLimit.limit,
        remaining: req.rateLimit.remaining,
        resetTime: req.rateLimit.resetTime,
      });
    }
    
    next(err);
  });
};

/**
 * Specific rate limiters for different endpoints
 */

// File upload rate limiter - more restrictive
const uploadRateLimit = createRateLimiter({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 50, // 50 uploads per window
  message: rateLimitMessage,
  skipSuccessfulRequests: false,
});

// Export rate limiter - very restrictive
const exportRateLimit = createRateLimiter({
  windowMs: 60 * 60 * 1000, // 1 hour
  max: 10, // 10 exports per hour
  message: rateLimitMessage,
  skipSuccessfulRequests: true,
});

// Analytics rate limiter
const analyticsRateLimit = createRateLimiter({
  windowMs: 60 * 1000, // 1 minute
  max: 60, // 60 requests per minute
  message: rateLimitMessage,
});

// Bulk operations rate limiter
const bulkRateLimit = createRateLimiter({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 20, // 20 bulk operations per window
  message: rateLimitMessage,
});

// Integration operations rate limiter
const integrationRateLimit = createRateLimiter({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100, // 100 integration calls per window
  message: rateLimitMessage,
});

// WebSocket connection rate limiter
const websocketRateLimit = createRateLimiter({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 50, // 50 WebSocket connections per window
  message: rateLimitMessage,
});

/**
 * Rate limit bypass for specific conditions
 */
const rateLimitBypass = (req, res, next) => {
  // Bypass rate limiting for certain conditions
  const bypassConditions = [
    // Internal service calls
    req.service && req.service.name === 'form-service',
    
    // Emergency access (if implemented)
    req.headers['x-emergency-access'] && config.security.emergencyBypass,
    
    // Health checks
    req.path === '/health' || req.path === '/ping',
  ];

  if (bypassConditions.some(condition => condition)) {
    req.skipRateLimit = true;
  }

  next();
};

/**
 * Rate limit monitoring and alerting
 */
const rateLimitMonitor = (req, res, next) => {
  // Skip if no rate limit info
  if (!req.rateLimit) {
    return next();
  }

  const { limit, remaining, resetTime } = req.rateLimit;
  const usagePercent = ((limit - remaining) / limit) * 100;

  // Log high usage
  if (usagePercent > 80) {
    logger.warn('High rate limit usage', {
      key: keyGenerator(req),
      usagePercent: usagePercent.toFixed(2),
      remaining,
      limit,
      resetTime,
      userAgent: req.get('User-Agent'),
      ip: req.ip,
    });
  }

  // Alert on rate limit hit
  if (remaining === 0) {
    logger.error('Rate limit exceeded', {
      key: keyGenerator(req),
      limit,
      resetTime,
      userAgent: req.get('User-Agent'),
      ip: req.ip,
      url: req.url,
      method: req.method,
    });

    // Could trigger additional alerting here
    // e.g., send to monitoring service, Slack, email, etc.
  }

  next();
};

/**
 * Adaptive rate limiting based on system load
 */
const adaptiveRateLimit = (req, res, next) => {
  // This could be enhanced to adjust rate limits based on:
  // - System CPU/memory usage
  // - Database response times
  // - Queue lengths
  // - Error rates
  
  // For now, just apply standard rate limiting
  next();
};

/**
 * Rate limit status endpoint
 */
const getRateLimitStatus = (req, res) => {
  const key = keyGenerator(req);
  
  // This would typically query the rate limit store
  // For now, return current request's rate limit info
  if (req.rateLimit) {
    return ResponseFormatter.success(res, {
      key,
      limit: req.rateLimit.limit,
      remaining: req.rateLimit.remaining,
      resetTime: new Date(Date.now() + req.rateLimit.resetTime).toISOString(),
      tier: req.rateLimitTier || 'anonymous',
    });
  }

  return ResponseFormatter.success(res, {
    key,
    message: 'No rate limit information available',
  });
};

module.exports = {
  dynamicRateLimit,
  uploadRateLimit,
  exportRateLimit,
  analyticsRateLimit,
  bulkRateLimit,
  integrationRateLimit,
  websocketRateLimit,
  rateLimitBypass,
  rateLimitMonitor,
  adaptiveRateLimit,
  getRateLimitStatus,
  rateLimitConfigs,
};
