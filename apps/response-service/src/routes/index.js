/**
 * Main Routes Index for Response Service
 * Implements API versioning and route organization
 */

const express = require('express');
const v1Routes = require('./v1');
const { createErrorResponse } = require('../dto/response-dtos');
const logger = require('../utils/logger');

const router = express.Router();

// =============================================================================
// API Version Routes
// =============================================================================

// Mount v1 routes
router.use('/v1', v1Routes);

// =============================================================================
// Root API Information
// =============================================================================

/**
 * @swagger
 * /api:
 *   get:
 *     tags: [Service Info]
 *     summary: API service information
 *     description: Returns basic information about the Response Service API
 *     responses:
 *       200:
 *         description: Service information
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 success:
 *                   type: boolean
 *                   example: true
 *                 message:
 *                   type: string
 *                   example: Response Service API
 *                 data:
 *                   type: object
 *                   properties:
 *                     service:
 *                       type: string
 *                       example: response-service
 *                     version:
 *                       type: string
 *                       example: 1.0.0
 *                     description:
 *                       type: string
 *                       example: Microservice for handling form responses
 *                     availableVersions:
 *                       type: array
 *                       items:
 *                         type: string
 *                       example: ['v1']
 *                     documentation:
 *                       type: string
 *                       example: /api/docs
 *                     health:
 *                       type: string
 *                       example: /api/v1/health
 *                 timestamp:
 *                   type: string
 *                   format: date-time
 *                 version:
 *                   type: string
 *                   example: v1
 */
router.get('/', (req, res) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  
  logger.debug('API root accessed', {
    correlationId,
    ip: req.ip,
    userAgent: req.get('User-Agent')
  });

  const serviceInfo = {
    service: 'response-service',
    version: process.env.SERVICE_VERSION || '1.0.0',
    description: 'Microservice for handling form responses with comprehensive validation and analytics',
    availableVersions: ['v1'],
    endpoints: {
      documentation: '/api/docs',
      health: '/api/v1/health',
      responses: '/api/v1/responses',
      analytics: '/api/v1/forms/{formId}/responses/analytics'
    },
    features: [
      'Form response submission',
      'Response validation',
      'File upload support', 
      'Response analytics',
      'Data export',
      'Real-time monitoring',
      'Security & rate limiting'
    ],
    environment: process.env.NODE_ENV || 'development',
    uptime: process.uptime(),
    timestamp: new Date().toISOString()
  };

  const response = {
    success: true,
    message: 'Response Service API',
    data: serviceInfo,
    correlationId,
    timestamp: new Date().toISOString(),
    version: 'v1'
  };

  res.json(response);
});

// =============================================================================
// Version Negotiation
// =============================================================================

/**
 * Handle requests to unsupported API versions
 */
router.use('/v*', (req, res) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  const requestedVersion = req.path.split('/')[1];
  
  logger.warn('Unsupported API version requested', {
    correlationId,
    requestedVersion,
    path: req.path,
    method: req.method,
    ip: req.ip,
    userAgent: req.get('User-Agent')
  });

  const errorResponse = createErrorResponse(
    'UNSUPPORTED_API_VERSION',
    `API version '${requestedVersion}' is not supported`,
    {
      requestedVersion,
      supportedVersions: ['v1'],
      upgradeInstructions: 'Please update your client to use a supported API version'
    },
    correlationId
  );

  res.status(400).json(errorResponse);
});

// =============================================================================
// API Documentation Route
// =============================================================================

/**
 * Redirect to API documentation
 */
router.get('/docs', (req, res) => {
  res.redirect('/api-docs');
});

// =============================================================================
// Health Check (Backward Compatibility)
// =============================================================================

/**
 * @swagger
 * /api/health:
 *   get:
 *     tags: [Health]
 *     summary: Service health check (deprecated)
 *     description: Legacy health check endpoint. Use /api/v1/health instead.
 *     deprecated: true
 *     responses:
 *       200:
 *         description: Service is healthy
 *       301:
 *         description: Redirected to versioned endpoint
 */
router.get('/health', (req, res) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  
  logger.debug('Legacy health endpoint accessed', {
    correlationId,
    ip: req.ip,
    userAgent: req.get('User-Agent')
  });

  // Redirect to versioned endpoint
  res.redirect(301, '/api/v1/health');
});

// =============================================================================
// Metrics Endpoint (for monitoring systems)
// =============================================================================

/**
 * @swagger
 * /api/metrics:
 *   get:
 *     tags: [Monitoring]
 *     summary: Service metrics
 *     description: Returns basic service metrics for monitoring systems
 *     responses:
 *       200:
 *         description: Service metrics
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 uptime:
 *                   type: number
 *                   description: Service uptime in seconds
 *                 memory:
 *                   type: object
 *                   properties:
 *                     used:
 *                       type: number
 *                     total:
 *                       type: number
 *                 version:
 *                   type: string
 *                 environment:
 *                   type: string
 */
router.get('/metrics', (req, res) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  
  const memoryUsage = process.memoryUsage();
  const metrics = {
    uptime: process.uptime(),
    memory: {
      used: Math.round(memoryUsage.heapUsed / 1024 / 1024 * 100) / 100, // MB
      total: Math.round(memoryUsage.heapTotal / 1024 / 1024 * 100) / 100, // MB
      external: Math.round(memoryUsage.external / 1024 / 1024 * 100) / 100, // MB
      rss: Math.round(memoryUsage.rss / 1024 / 1024 * 100) / 100 // MB
    },
    version: process.env.SERVICE_VERSION || '1.0.0',
    environment: process.env.NODE_ENV || 'development',
    nodeVersion: process.version,
    pid: process.pid,
    timestamp: new Date().toISOString(),
    correlationId
  };

  logger.debug('Metrics accessed', {
    correlationId,
    metrics: {
      uptime: metrics.uptime,
      memoryUsed: metrics.memory.used
    }
  });

  res.json(metrics);
});

module.exports = router;
