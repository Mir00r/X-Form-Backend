const { errorHandler, notFoundHandler, asyncHandler, timeoutHandler } = require('./errorHandler');
const { 
  authenticate, 
  optionalAuth, 
  authorize, 
  requireRole, 
  authenticateApiKey, 
  authenticateService,
  rateLimitContext,
  securityHeaders 
} = require('./auth');
const { 
  dynamicRateLimit, 
  uploadRateLimit, 
  exportRateLimit, 
  analyticsRateLimit,
  bulkRateLimit,
  integrationRateLimit,
  websocketRateLimit,
  rateLimitBypass,
  rateLimitMonitor 
} = require('./rateLimit');

/**
 * Export all middleware modules
 */
module.exports = {
  // Error handling
  errorHandler,
  notFoundHandler,
  asyncHandler,
  timeoutHandler,
  
  // Authentication & Authorization
  authenticate,
  optionalAuth,
  authorize,
  requireRole,
  authenticateApiKey,
  authenticateService,
  rateLimitContext,
  securityHeaders,
  
  // Rate limiting
  dynamicRateLimit,
  uploadRateLimit,
  exportRateLimit,
  analyticsRateLimit,
  bulkRateLimit,
  integrationRateLimit,
  websocketRateLimit,
  rateLimitBypass,
  rateLimitMonitor,
};
