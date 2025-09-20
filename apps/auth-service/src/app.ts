// Main application entry point for Auth Service
// Following Clean Architecture and SOLID principles

import express from 'express';
import cors from 'cors';
import helmet from 'helmet';
import compression from 'compression';
import cookieParser from 'cookie-parser';
import rateLimit from 'express-rate-limit';
import swaggerUi from 'swagger-ui-express';
import { createAuthServiceContainer } from './infrastructure/container';
import { authErrorHandler } from './interface/http/auth-controller';
import { swaggerSpec, getSwaggerHTML } from './infrastructure/swagger/swagger-config';

// Application class following Single Responsibility Principle
export class AuthServiceApp {
  private readonly app: express.Application;
  private readonly container: ReturnType<typeof createAuthServiceContainer>;
  private readonly port: number;

  constructor() {
    this.app = express();
    this.port = parseInt(process.env.PORT || '3001');
    
    // Initialize dependency injection container
    this.container = createAuthServiceContainer();
    
    // Setup middleware and routes
    this.setupMiddleware();
    this.setupRoutes();
    this.setupErrorHandling();
  }

  private setupMiddleware(): void {
    // Trust proxy for rate limiting and IP detection
    this.app.set('trust proxy', 1);

    // Security middleware
    this.app.use(helmet({
      contentSecurityPolicy: {
        directives: {
          defaultSrc: ["'self'"],
          styleSrc: ["'self'", "'unsafe-inline'", "https://unpkg.com"],
          scriptSrc: ["'self'", "https://unpkg.com"],
          imgSrc: ["'self'", "data:", "https:"],
          connectSrc: ["'self'"],
          fontSrc: ["'self'", "https://unpkg.com"],
          objectSrc: ["'none'"],
          mediaSrc: ["'self'"],
          frameSrc: ["'none'"],
        },
      },
      crossOriginEmbedderPolicy: false,
    }));

    // CORS configuration
    this.app.use(cors({
      origin: this.getAllowedOrigins(),
      credentials: true,
      methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
      allowedHeaders: ['Content-Type', 'Authorization', 'X-Requested-With'],
    }));

    // Rate limiting
    const globalLimiter = rateLimit({
      windowMs: 15 * 60 * 1000, // 15 minutes
      max: 1000, // 1000 requests per window per IP
      message: {
        success: false,
        error: {
          code: 'RATE_LIMIT_EXCEEDED',
          message: 'Too many requests from this IP, please try again later.',
        },
      },
      standardHeaders: true,
      legacyHeaders: false,
    });

    const authLimiter = rateLimit({
      windowMs: 15 * 60 * 1000, // 15 minutes
      max: 10, // 10 auth attempts per window per IP
      message: {
        success: false,
        error: {
          code: 'AUTH_RATE_LIMIT_EXCEEDED',
          message: 'Too many authentication attempts, please try again later.',
        },
      },
      standardHeaders: true,
      legacyHeaders: false,
    });

    // Body parsing middleware
    this.app.use(express.json({ limit: '10mb' }));
    this.app.use(express.urlencoded({ extended: true, limit: '10mb' }));
    this.app.use(cookieParser());
    this.app.use(compression());

    // Request logging middleware
    this.app.use((req, res, next) => {
      console.log(`${new Date().toISOString()} ${req.method} ${req.path} - ${req.ip}`);
      next();
    });

    // Apply rate limiting
    this.app.use(globalLimiter);
    this.app.use('/auth/login', authLimiter);
    this.app.use('/auth/register', authLimiter);
  }

  private setupSwaggerDocumentation(): void {
    // Custom swagger options with enhanced UI
    const swaggerOptions: swaggerUi.SwaggerUiOptions = {
      customCss: `
        .swagger-ui .topbar { background-color: #2c3e50; }
        .swagger-ui .info .title { color: #2c3e50; font-size: 2em; }
        .swagger-ui .info .description p { font-size: 1.1em; line-height: 1.6; }
        .swagger-ui .scheme-container { background: #f8f9fa; padding: 15px; border-radius: 5px; }
        .swagger-ui .auth-wrapper { margin-top: 20px; }
        .swagger-ui .btn.authorize { background-color: #2c3e50; border-color: #2c3e50; }
        .swagger-ui .btn.authorize:hover { background-color: #34495e; border-color: #34495e; }
        .swagger-ui .model-title { color: #2c3e50; }
        .swagger-ui .operation-tag-content { font-size: 1.2em; }
        .swagger-ui .opblock { margin-bottom: 20px; border-radius: 8px; }
        .swagger-ui .opblock-summary { border-radius: 8px 8px 0 0; }
        .swagger-ui .highlight-code { background: #f8f9fa; }
        .swagger-ui .model { font-family: 'Courier New', monospace; }
        .swagger-ui .response-col_status { font-weight: bold; }
        .swagger-ui .parameters-col_description p { margin: 0; }
        .swagger-ui .tab { border-radius: 4px 4px 0 0; }
        .swagger-ui .response-col_links { font-size: 0.9em; }
      `,
      customSiteTitle: 'X-Form Auth Service API Documentation',
      customfavIcon: '/favicon.ico',
      swaggerOptions: {
        tryItOutEnabled: true,
        requestInterceptor: (req: any) => {
          // Add correlation ID to all requests
          if (!req.headers['X-Correlation-ID']) {
            req.headers['X-Correlation-ID'] = 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
              const r = Math.random() * 16 | 0;
              const v = c === 'x' ? r : (r & 0x3 | 0x8);
              return v.toString(16);
            });
          }
          return req;
        },
        responseInterceptor: (res: any) => {
          // Log API calls for development
          if (process.env.NODE_ENV === 'development') {
            console.log(`Swagger UI API Call: ${res.url} - Status: ${res.status}`);
          }
          return res;
        },
        docExpansion: 'list', // Show operations list expanded
        defaultModelsExpandDepth: 2, // Expand models to 2 levels
        defaultModelExpandDepth: 2,
        displayRequestDuration: true, // Show request duration
        filter: true, // Enable API filtering
        showExtensions: true,
        showCommonExtensions: true,
        persistAuthorization: true, // Persist auth tokens in browser
      },
    };

    // Main Swagger UI route with enhanced features
    this.app.use('/api-docs', swaggerUi.serve);
    this.app.get('/api-docs', swaggerUi.setup(swaggerSpec, swaggerOptions));

    // Alternative Swagger routes for different use cases
    this.app.get('/docs', (req, res) => {
      res.redirect('/api-docs');
    });

    // Raw OpenAPI spec endpoint
    this.app.get('/api-docs.json', (req, res) => {
      res.setHeader('Content-Type', 'application/json');
      res.setHeader('Access-Control-Allow-Origin', '*');
      res.setHeader('Access-Control-Allow-Methods', 'GET');
      res.setHeader('Access-Control-Allow-Headers', 'Content-Type');
      res.send(swaggerSpec);
    });

    // Custom HTML documentation with additional features
    this.app.get('/api-docs-custom', (req, res) => {
      const customHTML = getSwaggerHTML(swaggerSpec);
      res.send(customHTML);
    });

    // Health check for documentation
    this.app.get('/api-docs/health', (req, res) => {
      res.json({
        success: true,
        message: 'API Documentation is available',
        endpoints: {
          swaggerUI: '/api-docs',
          openAPISpec: '/api-docs.json',
          customDocs: '/api-docs-custom',
          alternateRoute: '/docs',
        },
        features: [
          'Interactive API testing',
          'Request/Response examples',
          'Authentication testing',
          'Schema validation',
          'Request correlation tracking',
          'Enhanced UI with custom styling',
        ],
        timestamp: new Date().toISOString(),
      });
    });
  }

  private setupRoutes(): void {
    const authController = this.container.getAuthController();

    // Swagger API Documentation
    this.setupSwaggerDocumentation();

    // Health check endpoint
    this.app.get('/health', async (req, res) => {
      try {
        const health = await this.container.healthCheck();
        const statusCode = health.status === 'healthy' ? 200 : 503;
        res.status(statusCode).json({
          success: true,
          data: {
            service: 'auth-service',
            version: '1.0.0',
            timestamp: new Date().toISOString(),
            architecture: 'Clean Architecture with SOLID Principles',
            ...health,
          },
        });
      } catch (error) {
        res.status(503).json({
          success: false,
          error: {
            code: 'HEALTH_CHECK_FAILED',
            message: 'Health check failed',
          },
        });
      }
    });

    // API routes
    const apiRouter = express.Router();

    // Authentication routes
    const authRouter = express.Router();
    authRouter.post('/register', authController.register);
    authRouter.post('/login', authController.login);
    authRouter.post('/refresh', authController.refreshToken);
    authRouter.post('/verify-email', authController.verifyEmail);
    authRouter.post('/forgot-password', authController.forgotPassword);
    authRouter.post('/reset-password', authController.resetPassword);
    authRouter.post('/logout', authController.logout);
    authRouter.get('/profile', this.authMiddleware.bind(this), authController.getProfile);

    // Mount routers
    apiRouter.use('/auth', authRouter);
    this.app.use('/api/v1', apiRouter);

    // 404 handler
    this.app.use('*', (req, res) => {
      res.status(404).json({
        success: false,
        error: {
          code: 'NOT_FOUND',
          message: `Route ${req.method} ${req.originalUrl} not found`,
        },
      });
    });
  }

  private setupErrorHandling(): void {
    // Domain error handler
    this.app.use(authErrorHandler);

    // Global error handler
    this.app.use((error: Error, req: express.Request, res: express.Response, next: express.NextFunction) => {
      console.error('Unhandled error:', error);

      res.status(500).json({
        success: false,
        error: {
          code: 'INTERNAL_SERVER_ERROR',
          message: 'An unexpected error occurred',
          ...(process.env.NODE_ENV === 'development' && { stack: error.stack }),
        },
      });
    });
  }

  // Authentication middleware
  private authMiddleware(req: express.Request, res: express.Response, next: express.NextFunction): void {
    const authHeader = req.headers.authorization;
    
    if (!authHeader?.startsWith('Bearer ')) {
      res.status(401).json({
        success: false,
        error: {
          code: 'MISSING_TOKEN',
          message: 'Authorization token is required',
        },
      });
      return;
    }

    const token = authHeader.substring(7);
    
    try {
      // Get token service from container for verification
      const authService = this.container.getAuthApplicationService();
      // For now, we'll add a simple token verification
      // In a full implementation, we'd extract this to a middleware service
      
      // Decode JWT manually (simplified)
      const payload = JSON.parse(Buffer.from(token.split('.')[1], 'base64').toString());
      
      // Check expiration
      if (payload.exp && payload.exp < Date.now() / 1000) {
        res.status(401).json({
          success: false,
          error: {
            code: 'TOKEN_EXPIRED',
            message: 'Token has expired',
          },
        });
        return;
      }

      // Add user info to request
      (req as any).user = {
        userId: payload.userId,
        email: payload.email,
        role: payload.role,
      };

      next();
    } catch (error) {
      res.status(401).json({
        success: false,
        error: {
          code: 'INVALID_TOKEN',
          message: 'Invalid authorization token',
        },
      });
    }
  }

  private getAllowedOrigins(): string[] {
    const origins = process.env.ALLOWED_ORIGINS?.split(',') || [];
    
    // Default allowed origins for development
    if (process.env.NODE_ENV === 'development') {
      origins.push('http://localhost:3000', 'http://localhost:3001', 'http://localhost:8080');
    }

    return origins;
  }

  // Start the server
  async start(): Promise<void> {
    try {
      // Test database connection
      await this.container.healthCheck();
      
      this.app.listen(this.port, () => {
        console.log('🚀 Auth Service starting...');
        console.log(`📊 Environment: ${process.env.NODE_ENV || 'development'}`);
        console.log(`🌐 Server running on port ${this.port}`);
        console.log(`🏗️  Architecture: Clean Architecture with SOLID Principles`);
        console.log('📋 SOLID Principles Applied:');
        console.log('   • Single Responsibility: Each layer has one responsibility');
        console.log('   • Open/Closed: Open for extension, closed for modification');
        console.log('   • Liskov Substitution: Interfaces enable substitutability');
        console.log('   • Interface Segregation: Small, focused interfaces');
        console.log('   • Dependency Inversion: Depend on abstractions, not concretions');
        console.log('✅ Auth Service ready to handle requests');
      });

      // Graceful shutdown
      process.on('SIGTERM', this.gracefulShutdown.bind(this));
      process.on('SIGINT', this.gracefulShutdown.bind(this));

    } catch (error) {
      console.error('❌ Failed to start Auth Service:', error);
      process.exit(1);
    }
  }

  private async gracefulShutdown(signal: string): Promise<void> {
    console.log(`\n🛑 Received ${signal}. Starting graceful shutdown...`);
    
    try {
      // Close container and cleanup resources
      await this.container.close();
      console.log('✅ Auth Service shutdown complete');
      process.exit(0);
    } catch (error) {
      console.error('❌ Error during shutdown:', error);
      process.exit(1);
    }
  }
}

// Export for testing
export { createAuthServiceContainer };

// Start the application if this file is run directly
if (require.main === module) {
  const app = new AuthServiceApp();
  app.start().catch((error) => {
    console.error('Failed to start application:', error);
    process.exit(1);
  });
}
