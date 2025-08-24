const jwt = require('jsonwebtoken');
const config = require('../config');
const { ResponseFormatter } = require('../utils/helpers');
const logger = require('../utils/logger');

/**
 * Authentication middleware to verify JWT tokens
 */
const authenticate = async (req, res, next) => {
  try {
    const authHeader = req.headers.authorization;
    
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return ResponseFormatter.unauthorized(res, 'Access token required');
    }

    const token = authHeader.substring(7); // Remove 'Bearer ' prefix

    try {
      const decoded = jwt.verify(token, config.jwt.secret);
      
      // Add user info to request
      req.user = {
        id: decoded.sub || decoded.userId,
        email: decoded.email,
        role: decoded.role || 'user',
        permissions: decoded.permissions || [],
        sessionId: decoded.sessionId,
      };

      // Log successful authentication
      logger.info('User authenticated', {
        userId: req.user.id,
        email: req.user.email,
        ip: req.ip,
        userAgent: req.get('User-Agent'),
      });

      next();
    } catch (jwtError) {
      if (jwtError.name === 'TokenExpiredError') {
        return ResponseFormatter.unauthorized(res, 'Token expired');
      } else if (jwtError.name === 'JsonWebTokenError') {
        return ResponseFormatter.unauthorized(res, 'Invalid token');
      } else {
        throw jwtError;
      }
    }
  } catch (error) {
    logger.error('Authentication error:', error);
    return ResponseFormatter.error(res, 'Authentication failed', 500);
  }
};

/**
 * Optional authentication middleware (doesn't fail if no token)
 */
const optionalAuth = async (req, res, next) => {
  try {
    const authHeader = req.headers.authorization;
    
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      // No token provided, continue without user info
      req.user = null;
      return next();
    }

    const token = authHeader.substring(7);

    try {
      const decoded = jwt.verify(token, config.jwt.secret);
      
      req.user = {
        id: decoded.sub || decoded.userId,
        email: decoded.email,
        role: decoded.role || 'user',
        permissions: decoded.permissions || [],
        sessionId: decoded.sessionId,
      };
    } catch (jwtError) {
      // Invalid token, continue without user info
      req.user = null;
    }

    next();
  } catch (error) {
    logger.error('Optional authentication error:', error);
    req.user = null;
    next();
  }
};

/**
 * Authorization middleware to check user permissions
 */
const authorize = (requiredPermissions = [], options = {}) => {
  return (req, res, next) => {
    try {
      const {
        requireAll = false, // Whether all permissions are required
        allowOwner = true,   // Whether resource owner has access
        resourceParam = 'id', // Parameter to check for ownership
      } = options;

      if (!req.user) {
        return ResponseFormatter.unauthorized(res, 'Authentication required');
      }

      // Super admin has all permissions
      if (req.user.role === 'super_admin') {
        return next();
      }

      // Check if user has required permissions
      const userPermissions = req.user.permissions || [];
      
      let hasPermission = false;
      
      if (requireAll) {
        // User must have ALL required permissions
        hasPermission = requiredPermissions.every(perm => 
          userPermissions.includes(perm)
        );
      } else {
        // User must have at least ONE required permission
        hasPermission = requiredPermissions.length === 0 || 
          requiredPermissions.some(perm => userPermissions.includes(perm));
      }

      // Check ownership if allowed and no permission found
      if (!hasPermission && allowOwner && req.params[resourceParam]) {
        // This will be checked in the controller with actual resource data
        req.checkOwnership = true;
        return next();
      }

      if (!hasPermission) {
        logger.warn('Authorization failed', {
          userId: req.user.id,
          requiredPermissions,
          userPermissions,
          url: req.url,
          method: req.method,
        });
        
        return ResponseFormatter.forbidden(res, 'Insufficient permissions');
      }

      next();
    } catch (error) {
      logger.error('Authorization error:', error);
      return ResponseFormatter.error(res, 'Authorization failed', 500);
    }
  };
};

/**
 * Role-based authorization middleware
 */
const requireRole = (allowedRoles = []) => {
  return (req, res, next) => {
    try {
      if (!req.user) {
        return ResponseFormatter.unauthorized(res, 'Authentication required');
      }

      if (!allowedRoles.includes(req.user.role)) {
        logger.warn('Role authorization failed', {
          userId: req.user.id,
          userRole: req.user.role,
          allowedRoles,
          url: req.url,
          method: req.method,
        });
        
        return ResponseFormatter.forbidden(res, 'Insufficient role privileges');
      }

      next();
    } catch (error) {
      logger.error('Role authorization error:', error);
      return ResponseFormatter.error(res, 'Authorization failed', 500);
    }
  };
};

/**
 * API key authentication middleware
 */
const authenticateApiKey = (req, res, next) => {
  try {
    const apiKey = req.headers['x-api-key'] || req.query.api_key;
    
    if (!apiKey) {
      return ResponseFormatter.unauthorized(res, 'API key required');
    }

    // In a real implementation, you would validate the API key against a database
    // For now, we'll use a simple check against configured keys
    const validApiKeys = config.security.apiKeys || [];
    
    if (!validApiKeys.includes(apiKey)) {
      logger.warn('Invalid API key attempt', {
        apiKey: apiKey.substring(0, 8) + '...',
        ip: req.ip,
        userAgent: req.get('User-Agent'),
      });
      
      return ResponseFormatter.unauthorized(res, 'Invalid API key');
    }

    // Set API key user context
    req.user = {
      id: 'api-key-user',
      type: 'api_key',
      permissions: ['api_access'],
    };

    logger.info('API key authenticated', {
      apiKey: apiKey.substring(0, 8) + '...',
      ip: req.ip,
    });

    next();
  } catch (error) {
    logger.error('API key authentication error:', error);
    return ResponseFormatter.error(res, 'Authentication failed', 500);
  }
};

/**
 * Service-to-service authentication middleware
 */
const authenticateService = (req, res, next) => {
  try {
    const serviceToken = req.headers['x-service-token'];
    
    if (!serviceToken) {
      return ResponseFormatter.unauthorized(res, 'Service token required');
    }

    try {
      const decoded = jwt.verify(serviceToken, config.jwt.serviceSecret);
      
      // Verify service identity
      if (decoded.type !== 'service' || !decoded.service) {
        return ResponseFormatter.unauthorized(res, 'Invalid service token');
      }

      req.service = {
        name: decoded.service,
        permissions: decoded.permissions || [],
      };

      logger.info('Service authenticated', {
        service: req.service.name,
        ip: req.ip,
      });

      next();
    } catch (jwtError) {
      if (jwtError.name === 'TokenExpiredError') {
        return ResponseFormatter.unauthorized(res, 'Service token expired');
      } else if (jwtError.name === 'JsonWebTokenError') {
        return ResponseFormatter.unauthorized(res, 'Invalid service token');
      } else {
        throw jwtError;
      }
    }
  } catch (error) {
    logger.error('Service authentication error:', error);
    return ResponseFormatter.error(res, 'Service authentication failed', 500);
  }
};

/**
 * Ownership verification helper
 */
const verifyOwnership = async (resource, userId, ownerField = 'submitterId') => {
  try {
    if (!resource || !userId) {
      return false;
    }

    // Check direct ownership
    if (resource[ownerField] === userId) {
      return true;
    }

    // Check if user is the submitter (for responses)
    if (resource.submitterId === userId) {
      return true;
    }

    // Additional ownership checks can be added here
    // e.g., team membership, organization access, etc.

    return false;
  } catch (error) {
    logger.error('Ownership verification error:', error);
    return false;
  }
};

/**
 * Rate limiting context middleware
 */
const rateLimitContext = (req, res, next) => {
  // Add rate limiting context for different user types
  if (req.user) {
    if (req.user.role === 'premium') {
      req.rateLimitTier = 'premium';
    } else if (req.user.role === 'admin') {
      req.rateLimitTier = 'admin';
    } else {
      req.rateLimitTier = 'standard';
    }
  } else {
    req.rateLimitTier = 'anonymous';
  }

  next();
};

/**
 * Security headers middleware
 */
const securityHeaders = (req, res, next) => {
  // Set security headers
  res.setHeader('X-Content-Type-Options', 'nosniff');
  res.setHeader('X-Frame-Options', 'DENY');
  res.setHeader('X-XSS-Protection', '1; mode=block');
  res.setHeader('Referrer-Policy', 'strict-origin-when-cross-origin');
  
  // Remove potentially sensitive headers
  res.removeHeader('X-Powered-By');
  
  next();
};

module.exports = {
  authenticate,
  optionalAuth,
  authorize,
  requireRole,
  authenticateApiKey,
  authenticateService,
  verifyOwnership,
  rateLimitContext,
  securityHeaders,
};
