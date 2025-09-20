// Simplified Auth Service Application with Swagger Documentation
// Clean implementation focusing on core functionality

import express from 'express';
import cors from 'cors';
import helmet from 'helmet';
import compression from 'compression';
import cookieParser from 'cookie-parser';
import rateLimit from 'express-rate-limit';
const swaggerJSDoc = require('swagger-jsdoc');
const swaggerUI = require('swagger-ui-express');

// Swagger Configuration
const swaggerDefinition = {
  openapi: '3.0.3',
  info: {
    title: 'X-Form Auth Service API',
    version: '1.0.0',
    description: `
      # X-Form Authentication & User Management Service
      
      A comprehensive authentication and user management service built with Clean Architecture and SOLID principles.
      
      ## Features
      - 🔐 JWT-based authentication with refresh tokens
      - 👤 User registration and profile management
      - 📧 Email verification and password reset
      - 🛡️ Rate limiting and security features
      - 📊 Health checks and monitoring
      - 🎯 RESTful API design
      
      ## Architecture
      - **Clean Architecture** with proper layer separation
      - **SOLID Principles** implementation
      - **Domain-Driven Design** with rich domain models
      - **Event-Driven Architecture** for extensibility
    `,
    contact: {
      name: 'X-Form Development Team',
      email: 'dev@xform.com',
      url: 'https://xform.com',
    },
    license: {
      name: 'MIT',
      url: 'https://opensource.org/licenses/MIT',
    },
  },
  servers: [
    {
      url: 'http://localhost:3002',
      description: 'Development server',
    },
    {
      url: 'https://auth-dev.xform.com',
      description: 'Development environment',
    },
  ],
  tags: [
    {
      name: 'Authentication',
      description: 'User authentication and token management',
    },
    {
      name: 'User Management',
      description: 'User registration and profile management',
    },
    {
      name: 'Health & Monitoring',
      description: 'Service health checks and monitoring endpoints',
    },
  ],
  components: {
    securitySchemes: {
      BearerAuth: {
        type: 'http',
        scheme: 'bearer',
        bearerFormat: 'JWT',
        description: 'JWT access token for authentication',
      },
    },
    schemas: {
      SuccessResponse: {
        type: 'object',
        properties: {
          success: { type: 'boolean', example: true },
          timestamp: { type: 'string', format: 'date-time' },
          path: { type: 'string', example: '/api/v1/auth/login' },
          method: { type: 'string', example: 'POST' },
          correlationId: { type: 'string', format: 'uuid' },
          data: { type: 'object' },
        },
      },
      ErrorResponse: {
        type: 'object',
        properties: {
          success: { type: 'boolean', example: false },
          timestamp: { type: 'string', format: 'date-time' },
          path: { type: 'string', example: '/api/v1/auth/login' },
          method: { type: 'string', example: 'POST' },
          correlationId: { type: 'string', format: 'uuid' },
          error: {
            type: 'object',
            properties: {
              code: { type: 'string', example: 'VALIDATION_ERROR' },
              message: { type: 'string', example: 'Validation failed' },
            },
          },
        },
      },
      RegisterRequest: {
        type: 'object',
        required: ['email', 'username', 'password', 'confirmPassword', 'firstName', 'lastName', 'acceptTerms'],
        properties: {
          email: { type: 'string', format: 'email', example: 'john.doe@example.com' },
          username: { type: 'string', minLength: 3, maxLength: 30, example: 'johndoe' },
          password: { type: 'string', minLength: 8, maxLength: 128, example: 'SecurePass123!' },
          confirmPassword: { type: 'string', example: 'SecurePass123!' },
          firstName: { type: 'string', maxLength: 50, example: 'John' },
          lastName: { type: 'string', maxLength: 50, example: 'Doe' },
          acceptTerms: { type: 'boolean', example: true },
        },
      },
      LoginRequest: {
        type: 'object',
        required: ['email', 'password'],
        properties: {
          email: { type: 'string', format: 'email', example: 'john.doe@example.com' },
          password: { type: 'string', example: 'SecurePass123!' },
          rememberMe: { type: 'boolean', example: false },
        },
      },
      AuthResponse: {
        type: 'object',
        properties: {
          accessToken: { type: 'string', example: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...' },
          refreshToken: { type: 'string', example: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...' },
          tokenType: { type: 'string', example: 'Bearer' },
          expiresIn: { type: 'integer', example: 900 },
          user: { $ref: '#/components/schemas/UserProfile' },
        },
      },
      UserProfile: {
        type: 'object',
        properties: {
          id: { type: 'string', format: 'uuid' },
          email: { type: 'string', format: 'email', example: 'john.doe@example.com' },
          username: { type: 'string', example: 'johndoe' },
          firstName: { type: 'string', example: 'John' },
          lastName: { type: 'string', example: 'Doe' },
          role: { type: 'string', enum: ['USER', 'ADMIN'], example: 'USER' },
          emailVerified: { type: 'boolean', example: true },
          createdAt: { type: 'string', format: 'date-time' },
          updatedAt: { type: 'string', format: 'date-time' },
        },
      },
    },
    responses: {
      '400': {
        description: 'Bad Request - Validation Error',
        content: {
          'application/json': {
            schema: { $ref: '#/components/schemas/ErrorResponse' },
          },
        },
      },
      '401': {
        description: 'Unauthorized - Authentication Required',
        content: {
          'application/json': {
            schema: { $ref: '#/components/schemas/ErrorResponse' },
          },
        },
      },
      '500': {
        description: 'Internal Server Error',
        content: {
          'application/json': {
            schema: { $ref: '#/components/schemas/ErrorResponse' },
          },
        },
      },
    },
  },
  security: [{ BearerAuth: [] }],
};

const swaggerOptions = {
  definition: swaggerDefinition,
  apis: ['./src/simple-app.ts'], // Look for swagger comments in this file
};

const swaggerSpec = swaggerJSDoc(swaggerOptions);

// Simple Auth Service Application
export class SimpleAuthServiceApp {
  private readonly app: express.Application;
  private readonly port: number;

  constructor() {
    this.app = express();
    this.port = parseInt(process.env.PORT || '3002'); // Changed from 3001 to 3002
    
    this.setupMiddleware();
    this.setupSwagger();
    this.setupRoutes();
    this.setupErrorHandling();
  }

  private setupMiddleware(): void {
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
        },
      },
      crossOriginEmbedderPolicy: false,
    }));

    // CORS
    this.app.use(cors({
      origin: ['http://localhost:3000', 'http://localhost:3001'],
      credentials: true,
      methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
      allowedHeaders: ['Content-Type', 'Authorization', 'X-Requested-With'],
    }));

    // Rate limiting
    const limiter = rateLimit({
      windowMs: 15 * 60 * 1000, // 15 minutes
      max: 100, // limit each IP to 100 requests per windowMs
    });
    this.app.use(limiter);

    // Body parsing
    this.app.use(express.json());
    this.app.use(express.urlencoded({ extended: true }));
    this.app.use(cookieParser());
    this.app.use(compression());

    // Request logging
    this.app.use((req, res, next) => {
      console.log(`${new Date().toISOString()} ${req.method} ${req.path}`);
      next();
    });
  }

  private setupSwagger(): void {
    // Swagger UI with custom styling
    const swaggerUiOptions = {
      customCss: `
        .swagger-ui .topbar { background-color: #2c3e50; }
        .swagger-ui .info .title { color: #2c3e50; font-size: 2em; }
        .swagger-ui .btn.authorize { background-color: #2c3e50; border-color: #2c3e50; }
      `,
      customSiteTitle: 'X-Form Auth Service API Documentation',
      swaggerOptions: {
        tryItOutEnabled: true,
        displayRequestDuration: true,
        filter: true,
        showExtensions: true,
        persistAuthorization: true,
      },
    };

    // Swagger routes
    this.app.use('/api-docs', swaggerUI.serve);
    this.app.get('/api-docs', swaggerUI.setup(swaggerSpec, swaggerUiOptions));
    
    // Alternative routes
    this.app.get('/docs', (req, res) => res.redirect('/api-docs'));
    this.app.get('/api-docs.json', (req, res) => {
      res.setHeader('Content-Type', 'application/json');
      res.send(swaggerSpec);
    });

    console.log('📚 Swagger documentation available at:');
    console.log(`   • http://localhost:${this.port}/api-docs`);
    console.log(`   • http://localhost:${this.port}/docs`);
    console.log(`   • http://localhost:${this.port}/api-docs.json (OpenAPI spec)`);
  }

  private setupRoutes(): void {
    /**
     * @swagger
     * /health:
     *   get:
     *     tags: [Health & Monitoring]
     *     summary: Health check endpoint
     *     description: Check if the service is running and healthy
     *     responses:
     *       200:
     *         description: Service is healthy
     *         content:
     *           application/json:
     *             schema:
     *               allOf:
     *                 - $ref: '#/components/schemas/SuccessResponse'
     *                 - type: object
     *                   properties:
     *                     data:
     *                       type: object
     *                       properties:
     *                         service: { type: string, example: auth-service }
     *                         version: { type: string, example: 1.0.0 }
     *                         status: { type: string, example: healthy }
     *                         uptime: { type: number, example: 3600 }
     */
    this.app.get('/health', (req, res) => {
      res.json({
        success: true,
        timestamp: new Date().toISOString(),
        path: req.path,
        method: req.method,
        correlationId: 'health-check-' + Date.now(),
        data: {
          service: 'auth-service',
          version: '1.0.0',
          status: 'healthy',
          uptime: process.uptime(),
          architecture: 'Clean Architecture with SOLID Principles',
          swagger: 'Enabled with comprehensive documentation',
        },
      });
    });

    /**
     * @swagger
     * /api/v1/auth/register:
     *   post:
     *     tags: [Authentication]
     *     summary: Register a new user account
     *     description: Register a new user with email verification
     *     requestBody:
     *       required: true
     *       content:
     *         application/json:
     *           schema:
     *             $ref: '#/components/schemas/RegisterRequest'
     *     responses:
     *       201:
     *         description: User registered successfully
     *         content:
     *           application/json:
     *             schema:
     *               allOf:
     *                 - $ref: '#/components/schemas/SuccessResponse'
     *                 - type: object
     *                   properties:
     *                     data:
     *                       type: object
     *                       properties:
     *                         message: { type: string, example: Registration successful }
     *                         userId: { type: string, format: uuid }
     *                         verificationRequired: { type: boolean, example: true }
     *       400:
     *         $ref: '#/components/responses/400'
     *       409:
     *         description: User already exists
     *       500:
     *         $ref: '#/components/responses/500'
     */
    this.app.post('/api/v1/auth/register', (req, res) => {
      // Mock implementation for demonstration
      const { email, username, password, firstName, lastName, acceptTerms } = req.body;
      
      if (!email || !username || !password || !firstName || !lastName || !acceptTerms) {
        return res.status(400).json({
          success: false,
          timestamp: new Date().toISOString(),
          path: req.path,
          method: req.method,
          correlationId: 'reg-' + Date.now(),
          error: {
            code: 'VALIDATION_ERROR',
            message: 'Missing required fields',
          },
        });
      }

      res.status(201).json({
        success: true,
        timestamp: new Date().toISOString(),
        path: req.path,
        method: req.method,
        correlationId: 'reg-' + Date.now(),
        data: {
          message: 'Registration successful',
          userId: '123e4567-e89b-12d3-a456-426614174000',
          verificationRequired: true,
        },
      });
    });

    /**
     * @swagger
     * /api/v1/auth/login:
     *   post:
     *     tags: [Authentication]
     *     summary: Authenticate user and get tokens
     *     description: Authenticate user with email and password, returns JWT tokens
     *     requestBody:
     *       required: true
     *       content:
     *         application/json:
     *           schema:
     *             $ref: '#/components/schemas/LoginRequest'
     *     responses:
     *       200:
     *         description: Login successful
     *         content:
     *           application/json:
     *             schema:
     *               allOf:
     *                 - $ref: '#/components/schemas/SuccessResponse'
     *                 - type: object
     *                   properties:
     *                     data:
     *                       $ref: '#/components/schemas/AuthResponse'
     *       400:
     *         $ref: '#/components/responses/400'
     *       401:
     *         description: Invalid credentials
     *       500:
     *         $ref: '#/components/responses/500'
     */
    this.app.post('/api/v1/auth/login', (req, res) => {
      // Mock implementation for demonstration
      const { email, password } = req.body;
      
      if (!email || !password) {
        return res.status(400).json({
          success: false,
          timestamp: new Date().toISOString(),
          path: req.path,
          method: req.method,
          correlationId: 'login-' + Date.now(),
          error: {
            code: 'VALIDATION_ERROR',
            message: 'Email and password are required',
          },
        });
      }

      if (email !== 'john.doe@example.com' || password !== 'SecurePass123!') {
        return res.status(401).json({
          success: false,
          timestamp: new Date().toISOString(),
          path: req.path,
          method: req.method,
          correlationId: 'login-' + Date.now(),
          error: {
            code: 'INVALID_CREDENTIALS',
            message: 'Invalid email or password',
          },
        });
      }

      res.json({
        success: true,
        timestamp: new Date().toISOString(),
        path: req.path,
        method: req.method,
        correlationId: 'login-' + Date.now(),
        data: {
          accessToken: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMiLCJuYW1lIjoiSm9obiBEb2UiLCJpYXQiOjE1MTYyMzkwMjJ9.mock-token',
          refreshToken: 'refresh-token-mock',
          tokenType: 'Bearer',
          expiresIn: 900,
          user: {
            id: '123e4567-e89b-12d3-a456-426614174000',
            email: 'john.doe@example.com',
            username: 'johndoe',
            firstName: 'John',
            lastName: 'Doe',
            role: 'USER',
            emailVerified: true,
            createdAt: '2024-01-01T00:00:00.000Z',
            updatedAt: '2024-01-01T00:00:00.000Z',
          },
        },
      });
    });

    /**
     * @swagger
     * /api/v1/auth/profile:
     *   get:
     *     tags: [User Management]
     *     summary: Get user profile
     *     description: Get authenticated user's profile information
     *     security:
     *       - BearerAuth: []
     *     responses:
     *       200:
     *         description: User profile retrieved successfully
     *         content:
     *           application/json:
     *             schema:
     *               allOf:
     *                 - $ref: '#/components/schemas/SuccessResponse'
     *                 - type: object
     *                   properties:
     *                     data:
     *                       $ref: '#/components/schemas/UserProfile'
     *       401:
     *         $ref: '#/components/responses/401'
     *       500:
     *         $ref: '#/components/responses/500'
     */
    this.app.get('/api/v1/auth/profile', this.authenticateToken, (req, res) => {
      res.json({
        success: true,
        timestamp: new Date().toISOString(),
        path: req.path,
        method: req.method,
        correlationId: 'profile-' + Date.now(),
        data: {
          id: '123e4567-e89b-12d3-a456-426614174000',
          email: 'john.doe@example.com',
          username: 'johndoe',
          firstName: 'John',
          lastName: 'Doe',
          role: 'USER',
          emailVerified: true,
          createdAt: '2024-01-01T00:00:00.000Z',
          updatedAt: '2024-01-01T00:00:00.000Z',
        },
      });
    });

    // 404 handler
    this.app.use('*', (req, res) => {
      res.status(404).json({
        success: false,
        timestamp: new Date().toISOString(),
        path: req.path,
        method: req.method,
        correlationId: '404-' + Date.now(),
        error: {
          code: 'NOT_FOUND',
          message: `Route ${req.method} ${req.originalUrl} not found`,
        },
      });
    });
  }

  private authenticateToken(req: express.Request, res: express.Response, next: express.NextFunction): void {
    const authHeader = req.headers.authorization;
    
    if (!authHeader?.startsWith('Bearer ')) {
      res.status(401).json({
        success: false,
        timestamp: new Date().toISOString(),
        path: req.path,
        method: req.method,
        correlationId: 'auth-' + Date.now(),
        error: {
          code: 'MISSING_TOKEN',
          message: 'Authorization token is required',
        },
      });
      return;
    }

    const token = authHeader.substring(7);
    
    // For demo purposes, we'll just check if token exists
    if (token === 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMiLCJuYW1lIjoiSm9obiBEb2UiLCJpYXQiOjE1MTYyMzkwMjJ9.mock-token') {
      (req as any).user = { userId: '123', email: 'john.doe@example.com' };
      next();
    } else {
      res.status(401).json({
        success: false,
        timestamp: new Date().toISOString(),
        path: req.path,
        method: req.method,
        correlationId: 'auth-' + Date.now(),
        error: {
          code: 'INVALID_TOKEN',
          message: 'Invalid authorization token',
        },
      });
    }
  }

  private setupErrorHandling(): void {
    this.app.use((error: Error, req: express.Request, res: express.Response, next: express.NextFunction) => {
      console.error('Unhandled error:', error);

      res.status(500).json({
        success: false,
        timestamp: new Date().toISOString(),
        path: req.path,
        method: req.method,
        correlationId: 'error-' + Date.now(),
        error: {
          code: 'INTERNAL_SERVER_ERROR',
          message: 'An unexpected error occurred',
          ...(process.env.NODE_ENV === 'development' && { stack: error.stack }),
        },
      });
    });
  }

  async start(): Promise<void> {
    try {
      this.app.listen(this.port, () => {
        console.log('🚀 Auth Service starting...');
        console.log(`📊 Environment: ${process.env.NODE_ENV || 'development'}`);
        console.log(`🌐 Server running on port ${this.port}`);
        console.log(`🏗️  Architecture: Clean Architecture with SOLID Principles`);
        console.log('📚 API Documentation:');
        console.log(`   • Swagger UI: http://localhost:${this.port}/api-docs`);
        console.log(`   • OpenAPI Spec: http://localhost:${this.port}/api-docs.json`);
        console.log('🔧 Available Endpoints:');
        console.log(`   • GET  /health - Health check`);
        console.log(`   • POST /api/v1/auth/register - User registration`);
        console.log(`   • POST /api/v1/auth/login - User authentication`);
        console.log(`   • GET  /api/v1/auth/profile - User profile (authenticated)`);
        console.log('✅ Auth Service ready to handle requests');
      });

      process.on('SIGTERM', this.gracefulShutdown.bind(this));
      process.on('SIGINT', this.gracefulShutdown.bind(this));

    } catch (error) {
      console.error('❌ Failed to start Auth Service:', error);
      process.exit(1);
    }
  }

  private async gracefulShutdown(signal: string): Promise<void> {
    console.log(`\n🛑 Received ${signal}. Starting graceful shutdown...`);
    console.log('✅ Auth Service shutdown complete');
    process.exit(0);
  }
}

// Start the application if this file is run directly
if (require.main === module) {
  const app = new SimpleAuthServiceApp();
  app.start().catch((error) => {
    console.error('Failed to start application:', error);
    process.exit(1);
  });
}
