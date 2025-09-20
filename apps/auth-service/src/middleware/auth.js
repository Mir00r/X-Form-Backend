const AuthService = require('../services/AuthService');

const authService = new AuthService();

// Extract IP address from request
const getClientIP = (req) => {
  return req.ip || 
         req.connection.remoteAddress || 
         req.socket.remoteAddress ||
         (req.connection.socket ? req.connection.socket.remoteAddress : null) ||
         req.headers['x-forwarded-for']?.split(',')[0]?.trim() ||
         '127.0.0.1';
};

// Extract User Agent
const getUserAgent = (req) => {
  return req.headers['user-agent'] || 'Unknown';
};

// Authentication middleware
const authenticateToken = async (req, res, next) => {
  try {
    const authHeader = req.headers['authorization'];
    const token = authHeader && authHeader.split(' ')[1]; // Bearer TOKEN

    if (!token) {
      return res.status(401).json({
        error: 'Access token required',
        code: 'TOKEN_REQUIRED'
      });
    }

    // Verify token
    const decoded = authService.verifyAccessToken(token);
    
    // Get fresh user data
    const user = await authService.getUserById(decoded.id);
    
    if (!user) {
      return res.status(401).json({
        error: 'User not found',
        code: 'USER_NOT_FOUND'
      });
    }

    // Check if user is active
    if (!user.isActive) {
      return res.status(401).json({
        error: 'Account is not active',
        code: 'ACCOUNT_INACTIVE'
      });
    }

    // Add user to request object
    req.user = user;
    req.token = token;
    req.clientIP = getClientIP(req);
    req.userAgent = getUserAgent(req);

    next();
  } catch (error) {
    if (error.message.includes('expired')) {
      return res.status(401).json({
        error: 'Access token expired',
        code: 'TOKEN_EXPIRED'
      });
    }

    return res.status(401).json({
      error: 'Invalid access token',
      code: 'TOKEN_INVALID'
    });
  }
};

// Optional authentication middleware (doesn't fail if no token)
const optionalAuth = async (req, res, next) => {
  try {
    const authHeader = req.headers['authorization'];
    const token = authHeader && authHeader.split(' ')[1];

    if (!token) {
      req.user = null;
      req.clientIP = getClientIP(req);
      req.userAgent = getUserAgent(req);
      return next();
    }

    const decoded = authService.verifyAccessToken(token);
    const user = await authService.getUserById(decoded.id);
    
    req.user = user || null;
    req.token = token;
    req.clientIP = getClientIP(req);
    req.userAgent = getUserAgent(req);

    next();
  } catch (error) {
    // If token is invalid, continue without user
    req.user = null;
    req.clientIP = getClientIP(req);
    req.userAgent = getUserAgent(req);
    next();
  }
};

// Email verification required middleware
const requireEmailVerification = (req, res, next) => {
  if (!req.user.emailVerified) {
    return res.status(403).json({
      error: 'Email verification required',
      code: 'EMAIL_VERIFICATION_REQUIRED'
    });
  }
  next();
};

// Admin role middleware
const requireAdmin = (req, res, next) => {
  // This would check for admin role in user object
  // For now, we'll check if user is the admin user
  if (req.user.email !== 'admin@xform.dev') {
    return res.status(403).json({
      error: 'Admin access required',
      code: 'ADMIN_ACCESS_REQUIRED'
    });
  }
  next();
};

// Rate limiting middleware for sensitive operations
const sensitiveOperationLimiter = (maxAttempts = 5, windowMs = 15 * 60 * 1000) => {
  const attempts = new Map();

  return (req, res, next) => {
    const key = `${req.clientIP}-${req.route.path}`;
    const now = Date.now();
    const windowStart = now - windowMs;

    // Clean old entries
    const userAttempts = attempts.get(key) || [];
    const recentAttempts = userAttempts.filter(time => time > windowStart);

    if (recentAttempts.length >= maxAttempts) {
      return res.status(429).json({
        error: 'Too many attempts. Please try again later.',
        code: 'RATE_LIMIT_EXCEEDED',
        retryAfter: Math.ceil((recentAttempts[0] + windowMs - now) / 1000)
      });
    }

    // Record this attempt
    recentAttempts.push(now);
    attempts.set(key, recentAttempts);

    next();
  };
};

// Resource ownership middleware (for checking if user owns a resource)
const requireResourceOwnership = (resourceType, paramName = 'id') => {
  return async (req, res, next) => {
    try {
      const resourceId = req.params[paramName];
      const userId = req.user.id;

      // This would be implemented based on your resource tables
      // For example, checking if user owns a form
      switch (resourceType) {
        case 'form':
          // Check if user owns the form
          const formResult = await authService.pool.query(
            'SELECT user_id FROM forms WHERE id = $1',
            [resourceId]
          );
          
          if (formResult.rows.length === 0) {
            return res.status(404).json({
              error: 'Resource not found',
              code: 'RESOURCE_NOT_FOUND'
            });
          }

          if (formResult.rows[0].user_id !== userId) {
            return res.status(403).json({
              error: 'Access denied. You do not own this resource.',
              code: 'RESOURCE_ACCESS_DENIED'
            });
          }
          break;

        default:
          return res.status(500).json({
            error: 'Invalid resource type',
            code: 'INVALID_RESOURCE_TYPE'
          });
      }

      next();
    } catch (error) {
      return res.status(500).json({
        error: 'Failed to verify resource ownership',
        code: 'OWNERSHIP_CHECK_FAILED'
      });
    }
  };
};

// Request context middleware (adds common request data)
const addRequestContext = (req, res, next) => {
  req.clientIP = getClientIP(req);
  req.userAgent = getUserAgent(req);
  req.requestId = require('crypto').randomUUID();
  
  // Add request start time for performance monitoring
  req.startTime = Date.now();
  
  next();
};

module.exports = {
  authenticateToken,
  optionalAuth,
  requireEmailVerification,
  requireAdmin,
  sensitiveOperationLimiter,
  requireResourceOwnership,
  addRequestContext,
  getClientIP,
  getUserAgent
};

const optionalAuth = (req, res, next) => {
  const authHeader = req.headers['authorization'];
  const token = authHeader && authHeader.split(' ')[1];

  if (!token) {
    req.user = null;
    return next();
  }

  try {
    const decoded = jwt.verify(token, process.env.JWT_SECRET);
    req.user = decoded;
  } catch (error) {
    req.user = null;
  }
  
  next();
};

module.exports = {
  authenticateToken,
  optionalAuth
};
