/**
 * Response Service Main Application
 * Comprehensive microservice for managing form responses
 */

require('dotenv').config();
require('express-async-errors');

const express = require('express');
const cors = require('cors');
const helmet = require('helmet');
const compression = require('compression');
const rateLimit = require('express-rate-limit');

// Import configurations and utilities
const config = require('./config/enhanced');
const logger = require('./utils/logger');
const { eventSystem, ResponseEvents } = require('./events/eventSystem');

// Import middleware
const authMiddleware = require('./middleware/auth');
const validationMiddleware = require('./middleware/validation');
const securityMiddleware = require('./middleware/security');
const errorHandler = require('./middleware/errorHandler');

// Import routes
const v1Routes = require('./routes/v1');

// Import swagger configuration
const { specs, swaggerUi, swaggerUiOptions } = require('./config/swagger');

class ResponseServiceApplication {
  constructor() {
    this.app = express();
    this.server = null;
    this.setupMiddleware();
    this.setupRoutes();
    this.setupErrorHandling();
    this.setupEventListeners();
  }

  /**
   * Setup application middleware
   */
  setupMiddleware() {
    // Security middleware
    this.app.use(helmet({
      contentSecurityPolicy: config.get('security.enableCsp'),
      hsts: config.get('security.enableHsts')
    }));

    // CORS configuration
    this.app.use(cors({
      origin: config.get('security.corsOrigins'),
      credentials: true,
      methods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS'],
      allowedHeaders: ['Content-Type', 'Authorization', 'X-API-Key', 'X-Correlation-ID']
    }));

    // Compression
    if (config.get('api.enableCompression')) {
      this.app.use(compression());
    }

    // Body parsing
    this.app.use(express.json({ 
      limit: config.get('api.maxRequestSize'),
      strict: true
    }));
    this.app.use(express.urlencoded({ 
      extended: true, 
      limit: config.get('api.maxRequestSize') 
    }));

    // Rate limiting
    const limiter = rateLimit({
      windowMs: config.get('rateLimit.windowMs'),
      max: config.get('rateLimit.max'),
      message: config.get('rateLimit.message'),
      standardHeaders: config.get('rateLimit.standardHeaders'),
      legacyHeaders: config.get('rateLimit.legacyHeaders'),
      handler: (req, res) => {
        logger.warn('Rate limit exceeded', {
          ip: req.ip,
          userAgent: req.get('User-Agent'),
          path: req.path
        });
        
        res.status(429).json({
          success: false,
          error: {
            code: 'RATE_LIMIT_EXCEEDED',
            message: config.get('rateLimit.message')
          },
          correlationId: req.correlationId,
          timestamp: new Date().toISOString()
        });
      }
    });
    this.app.use(limiter);

    // Custom security middleware
    securityMiddleware.applySecurity(this.app);

    // Request logging
    this.app.use((req, res, next) => {
      // Generate correlation ID if not present
      req.correlationId = req.headers['x-correlation-id'] || 
        `cor_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

      logger.info('Request received', {
        method: req.method,
        url: req.url,
        ip: req.ip,
        userAgent: req.get('User-Agent'),
        correlationId: req.correlationId
      });

      // Add correlation ID to response headers
      res.set('X-Correlation-ID', req.correlationId);

      next();
    });

    // Health check endpoints are handled in the v1 routes
  }

  /**
   * Setup application routes
   */
  setupRoutes() {
    // API Documentation
    if (config.get('development.enableSwagger')) {
      this.app.use('/api-docs', swaggerUi.serve);
      this.app.get('/api-docs', swaggerUi.setup(specs, swaggerUiOptions));
      
      // Serve swagger JSON
      this.app.get('/api-docs.json', (req, res) => {
        res.setHeader('Content-Type', 'application/json');
        res.send(specs);
      });
    }

    // API Routes
    this.app.use('/api/v1', v1Routes);

    // Root endpoint
    this.app.get('/', (req, res) => {
      res.json({
        service: 'Response Service',
        version: config.get('server.version'),
        status: 'running',
        timestamp: new Date().toISOString(),
        documentation: config.get('development.enableSwagger') ? '/api-docs' : null
      });
    });

    // Handle 404 for undefined routes
    this.app.use('*', (req, res) => {
      logger.warn('Route not found', {
        method: req.method,
        url: req.url,
        ip: req.ip,
        correlationId: req.correlationId
      });

      res.status(404).json({
        success: false,
        error: {
          code: 'ROUTE_NOT_FOUND',
          message: `Route ${req.method} ${req.url} not found`
        },
        correlationId: req.correlationId,
        timestamp: new Date().toISOString()
      });
    });
  }

  /**
   * Setup error handling
   */
  setupErrorHandling() {
    // 404 handler for undefined routes
    this.app.use(errorHandler.notFoundHandler);
    
    // Global error handler
    this.app.use(errorHandler.globalErrorHandler);

    // Graceful shutdown handler
    process.on('SIGTERM', () => this.shutdown('SIGTERM'));
    process.on('SIGINT', () => this.shutdown('SIGINT'));

    // Unhandled rejection handler
    process.on('unhandledRejection', (reason, promise) => {
      logger.error('Unhandled Promise Rejection', {
        reason: reason.toString(),
        promise: promise.toString(),
        stack: reason.stack
      });
    });

    // Uncaught exception handler
    process.on('uncaughtException', (error) => {
      logger.error('Uncaught Exception', {
        error: error.message,
        stack: error.stack
      });
      
      // Graceful shutdown on uncaught exception
      this.shutdown('UNCAUGHT_EXCEPTION');
    });
  }

  /**
   * Setup event listeners
   */
  setupEventListeners() {
    // Listen for service events
    eventSystem.subscribeToEvent(ResponseEvents.SERVICE_STARTED, async (event) => {
      logger.info('Service started event received', {
        eventId: event.id,
        correlationId: event.correlationId
      });
    }, { subscriberName: 'main-application' });

    eventSystem.subscribeToEvent(ResponseEvents.SERVICE_STOPPED, async (event) => {
      logger.info('Service stopped event received', {
        eventId: event.id,
        correlationId: event.correlationId
      });
    }, { subscriberName: 'main-application' });

    // Listen for health check failures
    eventSystem.subscribeToEvent(ResponseEvents.HEALTH_CHECK_FAILED, async (event) => {
      logger.error('Health check failed', {
        eventId: event.id,
        data: event.data,
        correlationId: event.correlationId
      });
    }, { subscriberName: 'main-application' });
  }

  /**
   * Start the application server
   */
  async start() {
    try {
      const port = config.get('server.port');
      const host = config.get('server.host');

      this.server = this.app.listen(port, host, () => {
        logger.info('Response Service started successfully', {
          port,
          host,
          environment: config.get('server.environment'),
          version: config.get('server.version'),
          pid: process.pid,
          nodeVersion: process.version
        });

        // Publish service started event
        eventSystem.publishEvent(
          ResponseEvents.SERVICE_STARTED,
          {
            port,
            host,
            environment: config.get('server.environment'),
            version: config.get('server.version'),
            pid: process.pid
          },
          `startup_${Date.now()}`
        );
      });

      // Handle server errors
      this.server.on('error', (error) => {
        if (error.code === 'EADDRINUSE') {
          logger.error(`Port ${port} is already in use`, { error: error.message });
        } else {
          logger.error('Server error', { error: error.message, code: error.code });
        }
        process.exit(1);
      });

      // Set server timeouts
      this.server.keepAliveTimeout = config.get('performance.keepAliveTimeout') || 5000;
      this.server.headersTimeout = config.get('performance.headersTimeout') || 60000;

      return this.server;

    } catch (error) {
      logger.error('Failed to start Response Service', {
        error: error.message,
        stack: error.stack
      });
      process.exit(1);
    }
  }

  /**
   * Graceful shutdown
   */
  async shutdown(signal) {
    logger.info(`Received ${signal}, starting graceful shutdown...`);

    // Publish service stopping event
    await eventSystem.publishEvent(
      ResponseEvents.SERVICE_STOPPED,
      {
        signal,
        timestamp: new Date().toISOString(),
        pid: process.pid
      },
      `shutdown_${Date.now()}`
    );

    // Close server
    if (this.server) {
      this.server.close(() => {
        logger.info('HTTP server closed');
      });
    }

    // Close database connections, cleanup resources, etc.
    try {
      // Add cleanup logic here
      await this.cleanup();
      
      logger.info('Graceful shutdown completed');
      process.exit(0);
    } catch (error) {
      logger.error('Error during shutdown', {
        error: error.message,
        stack: error.stack
      });
      process.exit(1);
    }
  }

  /**
   * Cleanup resources
   */
  async cleanup() {
    // Cleanup event system
    if (eventSystem) {
      eventSystem.cleanupOldEvents();
    }

    // Close database connections
    // Close Redis connections
    // Close other external connections
    
    logger.info('Cleanup completed');
  }

  /**
   * Get Express application instance
   */
  getApp() {
    return this.app;
  }

  /**
   * Get server instance
   */
  getServer() {
    return this.server;
  }
}

// Create and export application instance
const application = new ResponseServiceApplication();

// Start the application if this file is run directly
if (require.main === module) {
  application.start().catch((error) => {
    logger.error('Failed to start application', {
      error: error.message,
      stack: error.stack
    });
    process.exit(1);
  });
}

module.exports = application;
