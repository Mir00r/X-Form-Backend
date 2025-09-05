/**
 * Authentication Middleware for Response Service
 * Handles JWT authentication and authorization
 */

const jwt = require('jsonwebtoken');
const { createErrorResponse } = require('../dto/response-dtos');
const logger = require('../utils/logger');

/**
 * JWT Secret - should be loaded from environment variables
 */
const JWT_SECRET = process.env.JWT_SECRET || 'your-super-secret-jwt-key-change-in-production';

/**
 * Extract token from request headers
 */
const extractToken = (req) => {
  const authHeader = req.headers.authorization;
  
  if (authHeader && authHeader.startsWith('Bearer ')) {
    return authHeader.substring(7);
  }
  
  // Check for API key in headers
  const apiKey = req.headers['x-api-key'];
  if (apiKey) {
    return apiKey;
  }
  
  return null;
};

/**
 * Verify JWT token
 */
const verifyToken = (token) => {
  try {
    return jwt.verify(token, JWT_SECRET);
  } catch (error) {
    throw new Error(`Token verification failed: ${error.message}`);
  }
};

/**
 * Required authentication middleware
 * Returns 401 if no valid token is provided
 */
const authenticate = (req, res, next) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  
  try {
    const token = extractToken(req);
    
    if (!token) {
      logger.warn('Authentication required - no token provided', {
        correlationId,
        path: req.path,
        method: req.method,
        ip: req.ip
      });
      
      return res.status(401).json(
        createErrorResponse(
          'AUTHENTICATION_REQUIRED',
          'Authentication token is required',
          null,
          correlationId
        )
      );
    }
    
    const decoded = verifyToken(token);
    req.user = decoded;
    req.authToken = token;
    
    logger.debug('Authentication successful', {
      correlationId,
      userId: decoded.userId || decoded.sub,
      path: req.path,
      method: req.method
    });
    
    next();
  } catch (error) {
    logger.error('Authentication failed', {
      correlationId,
      error: error.message,
      path: req.path,
      method: req.method,
      ip: req.ip
    });
    
    return res.status(401).json(
      createErrorResponse(
        'INVALID_TOKEN',
        'Invalid or expired authentication token',
        null,
        correlationId
      )
    );
  }
};

/**
 * Optional authentication middleware
 * Continues without error if no token is provided, but validates if present
 */
const optionalAuthentication = (req, res, next) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  
  try {
    const token = extractToken(req);
    
    if (!token) {
      // No token provided, continue without authentication
      return next();
    }
    
    const decoded = verifyToken(token);
    req.user = decoded;
    req.authToken = token;
    
    logger.debug('Optional authentication successful', {
      correlationId,
      userId: decoded.userId || decoded.sub,
      path: req.path,
      method: req.method
    });
    
    next();
  } catch (error) {
    // Token provided but invalid - return error
    logger.error('Optional authentication failed with invalid token', {
      correlationId,
      error: error.message,
      path: req.path,
      method: req.method,
      ip: req.ip
    });
    
    return res.status(401).json(
      createErrorResponse(
        'INVALID_TOKEN',
        'Invalid or expired authentication token',
        null,
        correlationId
      )
    );
  }
};

/**
 * Authorization middleware
 * Checks if authenticated user has required permissions
 */
const authorize = (requiredRoles = [], requiredPermissions = []) => {
  return (req, res, next) => {
    const correlationId = req.headers['x-correlation-id'] || req.correlationId;
    
    if (!req.user) {
      logger.warn('Authorization check failed - user not authenticated', {
        correlationId,
        path: req.path,
        method: req.method
      });
      
      return res.status(401).json(
        createErrorResponse(
          'AUTHENTICATION_REQUIRED',
          'Authentication is required for this action',
          null,
          correlationId
        )
      );
    }
    
    // Check roles if specified
    if (requiredRoles.length > 0) {
      const userRoles = req.user.roles || [];
      const hasRequiredRole = requiredRoles.some(role => 
        userRoles.includes(role)
      );
      
      if (!hasRequiredRole) {
        logger.warn('Authorization failed - insufficient role', {
          correlationId,
          userId: req.user.userId || req.user.sub,
          userRoles,
          requiredRoles,
          path: req.path,
          method: req.method
        });
        
        return res.status(403).json(
          createErrorResponse(
            'INSUFFICIENT_PERMISSIONS',
            'Insufficient permissions to access this resource',
            {
              required: requiredRoles,
              current: userRoles
            },
            correlationId
          )
        );
      }
    }
    
    // Check permissions if specified
    if (requiredPermissions.length > 0) {
      const userPermissions = req.user.permissions || [];
      const hasRequiredPermission = requiredPermissions.some(permission => 
        userPermissions.includes(permission)
      );
      
      if (!hasRequiredPermission) {
        logger.warn('Authorization failed - insufficient permissions', {
          correlationId,
          userId: req.user.userId || req.user.sub,
          userPermissions,
          requiredPermissions,
          path: req.path,
          method: req.method
        });
        
        return res.status(403).json(
          createErrorResponse(
            'INSUFFICIENT_PERMISSIONS',
            'Insufficient permissions to access this resource',
            {
              required: requiredPermissions,
              current: userPermissions
            },
            correlationId
          )
        );
      }
    }
    
    logger.debug('Authorization successful', {
      correlationId,
      userId: req.user.userId || req.user.sub,
      path: req.path,
      method: req.method
    });
    
    next();
  };
};

/**
 * API Key authentication middleware
 * For service-to-service communication
 */
const authenticateApiKey = (req, res, next) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  const providedApiKey = req.headers['x-api-key'];
  const validApiKey = process.env.API_KEY || 'your-api-key-for-service-to-service-communication';
  
  if (!providedApiKey) {
    logger.warn('API key authentication required', {
      correlationId,
      path: req.path,
      method: req.method,
      ip: req.ip
    });
    
    return res.status(401).json(
      createErrorResponse(
        'API_KEY_REQUIRED',
        'API key is required for this endpoint',
        null,
        correlationId
      )
    );
  }
  
  if (providedApiKey !== validApiKey) {
    logger.error('Invalid API key provided', {
      correlationId,
      path: req.path,
      method: req.method,
      ip: req.ip
    });
    
    return res.status(401).json(
      createErrorResponse(
        'INVALID_API_KEY',
        'Invalid API key provided',
        null,
        correlationId
      )
    );
  }
  
  logger.debug('API key authentication successful', {
    correlationId,
    path: req.path,
    method: req.method
  });
  
  req.authenticatedViaApiKey = true;
  next();
};

/**
 * Combined authentication middleware
 * Accepts either JWT token or API key
 */
const authenticateFlexible = (req, res, next) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  const token = extractToken(req);
  const apiKey = req.headers['x-api-key'];
  
  // Try API key first
  if (apiKey) {
    return authenticateApiKey(req, res, next);
  }
  
  // Fall back to JWT authentication
  if (token) {
    return authenticate(req, res, next);
  }
  
  // No authentication provided
  logger.warn('No authentication provided', {
    correlationId,
    path: req.path,
    method: req.method,
    ip: req.ip
  });
  
  return res.status(401).json(
    createErrorResponse(
      'AUTHENTICATION_REQUIRED',
      'Authentication is required (JWT token or API key)',
      null,
      correlationId
    )
  );
};

module.exports = {
  authenticate,
  optionalAuthentication,
  authorize,
  authenticateApiKey,
  authenticateFlexible,
  extractToken,
  verifyToken
};
