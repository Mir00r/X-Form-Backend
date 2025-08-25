// Enhanced Auth Service Application with Microservices Best Practices
// Comprehensive integration of all improvements

import express, { Application, Request, Response, NextFunction } from 'express';
import cors from 'cors';
import helmet from 'helmet';
import compression from 'compression';
import cookieParser from 'cookie-parser';
import rateLimit from 'express-rate-limit';
import swaggerUi from 'swagger-ui-express';
import { correlationIdMiddleware, requestMetricsMiddleware, logger } from './infrastructure/logging/structured-logger';
import { getHealthCheckService } from './infrastructure/monitoring/health-check';
import { swaggerSpec } from './infrastructure/swagger/swagger-config';
import { ApiResponseHandler } from './interface/http/api-response-handler';
import { createAuthServiceContainer } from './infrastructure/container';
import { authErrorHandler } from './interface/http/auth-controller';

// Version-specific route imports
import { createAuthRoutesV1 } from './interface/http/routes/auth-routes-v1';

/**
 * Enhanced Auth Service Application
 * 
 * Features implemented:
 * ✅ API Versioning (/api/v1/)
 * ✅ Comprehensive DTOs
 * ✅ Standardized API responses
 * ✅ Input validation with express-validator
 * ✅ OpenAPI/Swagger documentation
 * ✅ Circuit breaker patterns
 * ✅ Structured logging with correlation IDs
 * ✅ Health checks and monitoring
 * ✅ Security middleware
 * ✅ Rate limiting
 * ✅ CORS configuration
 * ✅ Error handling
 */
export class EnhancedAuthServiceApp {
  private readonly app: Application;
  private readonly container: ReturnType<typeof createAuthServiceContainer>;
  private readonly port: number;

  constructor() {
    this.app = express();
    this.port = parseInt(process.env.PORT || '3001');
    
    // Initialize dependency injection container
    this.container = createAuthServiceContainer();
    
    // Setup application
    this.setupMiddleware();
    this.setupSwagger();
    this.setupRoutes();
    this.setupErrorHandling();
    
    logger.info('Enhanced Auth Service initialized with microservices best practices');
  }

  private setupMiddleware(): void {
    // Trust proxy for rate limiting and IP detection
    this.app.set('trust proxy', 1);

    // Correlation ID middleware (must be first)
    this.app.use(correlationIdMiddleware);

    // Request metrics middleware
    this.app.use(requestMetricsMiddleware);

    // Security middleware
    this.app.use(helmet({
      contentSecurityPolicy: {
        directives: {
          defaultSrc: ["'self'"],
          styleSrc: ["'self'", "'unsafe-inline'", "https://fonts.googleapis.com"],
          scriptSrc: ["'self'", "'unsafe-inline'"],
          imgSrc: ["'self'", "data:", "https:"],
          connectSrc: ["'self'"],
          fontSrc: ["'self'", "https://fonts.gstatic.com"],
          objectSrc: ["'none'"],
          mediaSrc: ["'self'"],
          frameSrc: ["'none'"],
        },
      },
      crossOriginEmbedderPolicy: false,
    }));

    // CORS configuration
    this.app.use(cors({
      origin: process.env.NODE_ENV === 'production' 
        ? [process.env.CLIENT_URL || 'https://yourapp.com']
        : ['http://localhost:3000', 'http://localhost:3001'],
      credentials: true,
      methods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS'],
      allowedHeaders: ['Content-Type', 'Authorization', 'X-Correlation-ID'],
      exposedHeaders: ['X-Correlation-ID'],
    }));

    // Compression
    this.app.use(compression());

    // Body parsing
    this.app.use(express.json({ limit: '10mb' }));
    this.app.use(express.urlencoded({ extended: true, limit: '10mb' }));
    this.app.use(cookieParser());

    // Rate limiting
    const limiter = rateLimit({
      windowMs: 15 * 60 * 1000, // 15 minutes
      max: 100, // limit each IP to 100 requests per windowMs
      message: {
        success: false,
        error: {
          code: 'RATE_LIMIT_EXCEEDED',
          message: 'Too many requests from this IP, please try again later.',
        },
      },
      standardHeaders: true,
      legacyHeaders: false,
      handler: (req: Request, res: Response) => {
        logger.warn('Rate limit exceeded', {
          ip: req.ip,
          userAgent: req.get('User-Agent'),
          path: req.path,
          method: req.method,
        });
        ApiResponseHandler.rateLimitExceeded(res);
      },
    });

    this.app.use(limiter);

    // Health check route (no rate limiting)
    this.app.get('/health', async (req: Request, res: Response) => {
      try {
        const healthService = getHealthCheckService();
        const health = await healthService.performHealthCheck();
        
        const statusCode = health.status === 'UNHEALTHY' ? 503 : 200;
        res.status(statusCode).json({
          success: true,
          data: health,
          timestamp: new Date().toISOString(),
        });
      } catch (error) {
        logger.error('Health check failed', error as Error);
        res.status(503).json({
          success: false,
          error: {
            code: 'HEALTH_CHECK_FAILED',
            message: 'Health check failed',
          },
          timestamp: new Date().toISOString(),
        });
      }
    });

    logger.info('Middleware setup completed');
  }

  private setupSwagger(): void {
    // Swagger UI setup
    const swaggerOptions = {
      explorer: true,
      swaggerOptions: {
        urls: [
          {
            url: '/api-docs/swagger.json',
            name: 'Auth Service API v1'
          }
        ]
      }
    };

    this.app.use('/api-docs', swaggerUi.serve, swaggerUi.setup(swaggerSpec, swaggerOptions));
    
    // Serve swagger.json
    this.app.get('/api-docs/swagger.json', (req: Request, res: Response) => {
      res.setHeader('Content-Type', 'application/json');
      res.send(swaggerSpec);
    });

    logger.info('Swagger documentation setup completed at /api-docs');
  }

  private setupRoutes(): void {
    // API version prefix
    const API_VERSION = '/api/v1';

    // Root endpoint
    this.app.get('/', (req: Request, res: Response) => {
      ApiResponseHandler.success(res, {
        service: 'Auth Service',
        version: '1.0.0',
        status: 'running',
        apiVersion: 'v1',
        documentation: '/api-docs',
        endpoints: {
          health: '/health',
          ready: `${API_VERSION}/auth/ready`,
          live: `${API_VERSION}/auth/live`,
          api: `${API_VERSION}/auth`
        },
        timestamp: new Date().toISOString(),
      });
    });

    // Get auth controller from container
    const authController = this.container.authController;

    // Mount versioned routes
    this.app.use(`${API_VERSION}/auth`, createAuthRoutesV1(authController));

    // API documentation redirect
    this.app.get('/docs', (req: Request, res: Response) => {
      res.redirect('/api-docs');
    });

    // Handle 404 for API routes
    this.app.use('/api/*', (req: Request, res: Response) => {
      ApiResponseHandler.notFound(res, `API endpoint not found: ${req.originalUrl}`);
    });

    logger.info(`API routes setup completed with version ${API_VERSION}`);
  }

  private setupErrorHandling(): void {
    // Global error handler
    this.app.use((error: Error, req: Request, res: Response, next: NextFunction) => {
      // Log error with correlation ID
      logger.error('Unhandled error occurred', error, {
        path: req.path,
        method: req.method,
        ip: req.ip,
        userAgent: req.get('User-Agent'),
      });

      // Use auth-specific error handler
      authErrorHandler(error, req, res, next);
    });

    // Handle unhandled promise rejections
    process.on('unhandledRejection', (reason: unknown, promise: Promise<any>) => {
      logger.error('Unhandled Promise Rejection', new Error(String(reason)), {
        promise: promise.toString(),
      });
    });

    // Handle uncaught exceptions
    process.on('uncaughtException', (error: Error) => {
      logger.error('Uncaught Exception', error);
      
      // Graceful shutdown
      this.gracefulShutdown();
    });

    logger.info('Error handling setup completed');
  }

  public async start(): Promise<void> {
    try {
      // Initialize health check service
      const healthService = getHealthCheckService();
      await healthService.initialize();

      // Start server
      const server = this.app.listen(this.port, () => {
        logger.info(`🚀 Enhanced Auth Service started successfully`, {
          port: this.port,
          environment: process.env.NODE_ENV || 'development',
          features: [
            'API Versioning',
            'Swagger Documentation',
            'Circuit Breakers',
            'Structured Logging',
            'Health Monitoring',
            'Input Validation',
            'Rate Limiting',
            'Security Headers',
          ],
          endpoints: {
            api: `http://localhost:${this.port}/api/v1/auth`,
            docs: `http://localhost:${this.port}/api-docs`,
            health: `http://localhost:${this.port}/health`,
          },
        });
      });

      // Graceful shutdown
      process.on('SIGTERM', () => {
        logger.info('SIGTERM received, starting graceful shutdown');
        server.close(() => {
          this.gracefulShutdown();
        });
      });

      process.on('SIGINT', () => {
        logger.info('SIGINT received, starting graceful shutdown');
        server.close(() => {
          this.gracefulShutdown();
        });
      });

    } catch (error) {
      logger.error('Failed to start Enhanced Auth Service', error as Error);
      process.exit(1);
    }
  }

  private async gracefulShutdown(): Promise<void> {
    try {
      logger.info('Starting graceful shutdown');
      
      // Close database connections
      // await this.container.database.close();
      
      // Stop health check service
      const healthService = getHealthCheckService();
      await healthService.stop();
      
      logger.info('Graceful shutdown completed');
      process.exit(0);
    } catch (error) {
      logger.error('Error during graceful shutdown', error as Error);
      process.exit(1);
    }
  }

  public getApp(): Application {
    return this.app;
  }
}

// Export singleton instance
export const createEnhancedAuthService = (): EnhancedAuthServiceApp => {
  return new EnhancedAuthServiceApp();
};

// Auto-start if this file is run directly
if (require.main === module) {
  const authService = createEnhancedAuthService();
  authService.start().catch((error) => {
    console.error('Failed to start auth service:', error);
    process.exit(1);
  });
}
