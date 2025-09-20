import express, { Application, Request, Response } from 'express';
import cors from 'cors';
import helmet from 'helmet';
import compression from 'compression';
import swaggerUi from 'swagger-ui-express';
import swaggerJSDoc from 'swagger-jsdoc';
import { swaggerSpec, swaggerUiOptions, getSwaggerHTML } from './src/infrastructure/swagger/enhanced-swagger-config';

const app: Application = express();
const PORT = process.env.PORT || 3001;

// Enhanced middleware configuration
app.use(cors({
  origin: process.env.ALLOWED_ORIGINS?.split(',') || ['http://localhost:3000'],
  credentials: true
}));

app.use(helmet({
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      styleSrc: ["'self'", "'unsafe-inline'", "https://unpkg.com"],
      scriptSrc: ["'self'", "'unsafe-inline'", "https://unpkg.com"],
      imgSrc: ["'self'", "data:", "https:"],
      connectSrc: ["'self'"]
    }
  }
}));

app.use(compression());
app.use(express.json({ limit: '10mb' }));
app.use(express.urlencoded({ extended: true }));

// Request correlation middleware
app.use((req: Request, res: Response, next) => {
  const correlationId = req.headers['x-correlation-id'] as string || 
    'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
      const r = Math.random() * 16 | 0;
      const v = c === 'x' ? r : (r & 0x3 | 0x8);
      return v.toString(16);
    });
  
  req.correlationId = correlationId;
  res.setHeader('X-Correlation-ID', correlationId);
  next();
});

// Swagger Documentation Endpoints
app.get('/api-docs.json', (req: Request, res: Response) => {
  res.setHeader('Content-Type', 'application/json');
  res.send(swaggerSpec);
});

app.get('/api-docs', (req: Request, res: Response) => {
  const html = getSwaggerHTML(swaggerSpec);
  res.send(html);
});

app.use('/api-docs', swaggerUi.serve);
app.get('/api-docs/ui', swaggerUi.setup(swaggerSpec, swaggerUiOptions));

// Health check endpoint
/**
 * @swagger
 * /health:
 *   get:
 *     tags:
 *       - Health & Monitoring
 *     summary: Service health check
 *     description: Returns the current health status of the auth service and its dependencies
 *     operationId: getHealthCheck
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
 *                       $ref: '#/components/schemas/HealthCheck'
 *             example:
 *               success: true
 *               timestamp: '2024-01-15T10:30:00.000Z'
 *               path: '/health'
 *               method: 'GET'
 *               correlationId: '550e8400-e29b-41d4-a716-446655440000'
 *               data:
 *                 service: 'auth-service'
 *                 version: '1.0.0'
 *                 status: 'HEALTHY'
 *                 uptime: 3600.5
 *                 timestamp: '2024-01-15T10:30:00.000Z'
 *                 environment: 'development'
 *                 dependencies:
 *                   - name: 'postgresql'
 *                     type: 'DATABASE'
 *                     status: 'HEALTHY'
 *                     responseTime: 25.5
 *                     lastChecked: '2024-01-15T10:30:00.000Z'
 *                     error: null
 *                 metrics:
 *                   requestCount: 1500
 *                   errorRate: 0.02
 *                   averageResponseTime: 120.5
 *                   activeConnections: 15
 *                   memoryUsage:
 *                     used: 256
 *                     free: 768
 *                     total: 1024
 *                     percentage: 25.0
 *                   cpuUsage: 15.5
 *       503:
 *         description: Service is unhealthy
 *         content:
 *           application/json:
 *             schema:
 *               $ref: '#/components/schemas/ErrorResponse'
 */
app.get('/health', (req: Request, res: Response) => {
  const startTime = process.uptime();
  const memUsage = process.memoryUsage();
  
  const healthData = {
    service: 'auth-service',
    version: '1.0.0',
    status: 'HEALTHY',
    uptime: startTime,
    timestamp: new Date().toISOString(),
    environment: process.env.NODE_ENV || 'development',
    dependencies: [
      {
        name: 'postgresql',
        type: 'DATABASE',
        status: 'HEALTHY',
        responseTime: 25.5,
        lastChecked: new Date().toISOString(),
        error: null
      }
    ],
    metrics: {
      requestCount: 1500,
      errorRate: 0.02,
      averageResponseTime: 120.5,
      activeConnections: 15,
      memoryUsage: {
        used: Math.round(memUsage.heapUsed / 1024 / 1024),
        free: Math.round((memUsage.heapTotal - memUsage.heapUsed) / 1024 / 1024),
        total: Math.round(memUsage.heapTotal / 1024 / 1024),
        percentage: Math.round((memUsage.heapUsed / memUsage.heapTotal) * 100)
      },
      cpuUsage: Math.round(Math.random() * 50) // Mock CPU usage
    }
  };

  res.json({
    success: true,
    timestamp: new Date().toISOString(),
    path: req.path,
    method: req.method,
    correlationId: req.correlationId,
    data: healthData
  });
});

// Authentication Demo Endpoints

/**
 * @swagger
 * /api/v1/auth/register:
 *   post:
 *     tags:
 *       - Authentication
 *     summary: Register new user account
 *     description: |
 *       Creates a new user account with email verification. This endpoint implements comprehensive validation,
 *       password security requirements, and duplicate detection.
 *       
 *       **Security Features:**
 *       - Email uniqueness validation
 *       - Strong password requirements (8+ chars, uppercase, lowercase, number, special char)
 *       - Input sanitization and validation
 *       - Rate limiting protection
 *       
 *       **Business Rules:**
 *       - Email must be unique across the system
 *       - Username must be unique and follow naming conventions
 *       - Password confirmation must match
 *       - Terms acceptance is mandatory
 *       - Email verification email will be sent automatically
 *     operationId: registerUser
 *     requestBody:
 *       required: true
 *       content:
 *         application/json:
 *           schema:
 *             $ref: '#/components/schemas/RegisterRequest'
 *           examples:
 *             validUser:
 *               summary: Valid registration request
 *               value:
 *                 email: "john.doe@example.com"
 *                 username: "johndoe"
 *                 password: "SecurePass123!"
 *                 confirmPassword: "SecurePass123!"
 *                 firstName: "John"
 *                 lastName: "Doe"
 *                 acceptTerms: true
 *                 marketingConsent: false
 *             userWithReferral:
 *               summary: Registration with referral code
 *               value:
 *                 email: "jane.smith@example.com"
 *                 username: "janesmith"
 *                 password: "MySecure456!"
 *                 confirmPassword: "MySecure456!"
 *                 firstName: "Jane"
 *                 lastName: "Smith"
 *                 acceptTerms: true
 *                 marketingConsent: true
 *                 referralCode: "REF12345"
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
 *                         user:
 *                           $ref: '#/components/schemas/UserProfile'
 *                         message:
 *                           type: string
 *                           example: "Account created successfully. Please check your email to verify your account."
 *             example:
 *               success: true
 *               timestamp: '2024-01-15T10:30:00.000Z'
 *               path: '/api/v1/auth/register'
 *               method: 'POST'
 *               correlationId: '550e8400-e29b-41d4-a716-446655440000'
 *               data:
 *                 user:
 *                   id: '550e8400-e29b-41d4-a716-446655440000'
 *                   email: 'john.doe@example.com'
 *                   username: 'johndoe'
 *                   firstName: 'John'
 *                   lastName: 'Doe'
 *                   fullName: 'John Doe'
 *                   role: 'USER'
 *                   emailVerified: false
 *                   phoneVerified: false
 *                   accountStatus: 'PENDING_VERIFICATION'
 *                   lastLoginAt: null
 *                   createdAt: '2024-01-15T10:30:00.000Z'
 *                   updatedAt: '2024-01-15T10:30:00.000Z'
 *                 message: 'Account created successfully. Please check your email to verify your account.'
 *       400:
 *         $ref: '#/components/responses/400'
 *       409:
 *         $ref: '#/components/responses/409'
 *       429:
 *         $ref: '#/components/responses/429'
 *       500:
 *         $ref: '#/components/responses/500'
 */
app.post('/api/v1/auth/register', (req: Request, res: Response) => {
  // Demo implementation
  const { email, username, password, firstName, lastName } = req.body;
  
  const mockUser = {
    id: '550e8400-e29b-41d4-a716-446655440000',
    email,
    username,
    firstName,
    lastName,
    fullName: `${firstName} ${lastName}`,
    role: 'USER',
    emailVerified: false,
    phoneVerified: false,
    accountStatus: 'PENDING_VERIFICATION',
    lastLoginAt: null,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString()
  };

  res.status(201).json({
    success: true,
    timestamp: new Date().toISOString(),
    path: req.path,
    method: req.method,
    correlationId: req.correlationId,
    data: {
      user: mockUser,
      message: 'Account created successfully. Please check your email to verify your account.'
    }
  });
});

/**
 * @swagger
 * /api/v1/auth/login:
 *   post:
 *     tags:
 *       - Authentication
 *     summary: Authenticate user credentials
 *     description: |
 *       Authenticates user credentials and returns JWT access and refresh tokens.
 *       
 *       **Security Features:**
 *       - BCrypt password verification
 *       - Account lockout after failed attempts
 *       - Rate limiting protection
 *       - Device tracking and management
 *       - Session management with refresh tokens
 *       
 *       **Token Information:**
 *       - Access Token: Short-lived (15 minutes), used for API authentication
 *       - Refresh Token: Long-lived (7 days), used to obtain new access tokens
 *       - Both tokens are JWTs signed with secure algorithms
 *       
 *       **Rate Limiting:**
 *       - 5 login attempts per minute per IP
 *       - Account locked after 5 failed attempts within 15 minutes
 *     operationId: loginUser
 *     requestBody:
 *       required: true
 *       content:
 *         application/json:
 *           schema:
 *             $ref: '#/components/schemas/LoginRequest'
 *           examples:
 *             standardLogin:
 *               summary: Standard login
 *               value:
 *                 email: "john.doe@example.com"
 *                 password: "SecurePass123!"
 *                 rememberMe: false
 *             extendedSession:
 *               summary: Login with extended session
 *               value:
 *                 email: "john.doe@example.com"
 *                 password: "SecurePass123!"
 *                 rememberMe: true
 *                 deviceId: "550e8400-e29b-41d4-a716-446655440001"
 *                 deviceName: "iPhone 15 Pro"
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
 *             example:
 *               success: true
 *               timestamp: '2024-01-15T10:30:00.000Z'
 *               path: '/api/v1/auth/login'
 *               method: 'POST'
 *               correlationId: '550e8400-e29b-41d4-a716-446655440000'
 *               data:
 *                 accessToken: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.access.token.signature'
 *                 refreshToken: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.refresh.token.signature'
 *                 tokenType: 'Bearer'
 *                 expiresIn: 900
 *                 expiresAt: '2024-01-15T10:45:00.000Z'
 *                 scope: ['read', 'write']
 *                 user:
 *                   id: '550e8400-e29b-41d4-a716-446655440000'
 *                   email: 'john.doe@example.com'
 *                   username: 'johndoe'
 *                   firstName: 'John'
 *                   lastName: 'Doe'
 *                   fullName: 'John Doe'
 *                   role: 'USER'
 *                   emailVerified: true
 *                   phoneVerified: false
 *                   accountStatus: 'ACTIVE'
 *                   lastLoginAt: '2024-01-15T10:30:00.000Z'
 *                   createdAt: '2024-01-01T00:00:00.000Z'
 *                   updatedAt: '2024-01-15T10:30:00.000Z'
 *       400:
 *         $ref: '#/components/responses/400'
 *       401:
 *         $ref: '#/components/responses/401'
 *       423:
 *         $ref: '#/components/responses/423'
 *       429:
 *         $ref: '#/components/responses/429'
 *       500:
 *         $ref: '#/components/responses/500'
 */
app.post('/api/v1/auth/login', (req: Request, res: Response) => {
  const { email } = req.body;
  
  const mockAuthResponse = {
    accessToken: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.access.token.signature',
    refreshToken: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.refresh.token.signature',
    tokenType: 'Bearer',
    expiresIn: 900,
    expiresAt: new Date(Date.now() + 900000).toISOString(),
    scope: ['read', 'write'],
    user: {
      id: '550e8400-e29b-41d4-a716-446655440000',
      email,
      username: 'johndoe',
      firstName: 'John',
      lastName: 'Doe',
      fullName: 'John Doe',
      role: 'USER',
      emailVerified: true,
      phoneVerified: false,
      accountStatus: 'ACTIVE',
      lastLoginAt: new Date().toISOString(),
      createdAt: '2024-01-01T00:00:00.000Z',
      updatedAt: new Date().toISOString()
    }
  };

  res.json({
    success: true,
    timestamp: new Date().toISOString(),
    path: req.path,
    method: req.method,
    correlationId: req.correlationId,
    data: mockAuthResponse
  });
});

/**
 * @swagger
 * /api/v1/auth/refresh:
 *   post:
 *     tags:
 *       - Authentication
 *     summary: Refresh access token
 *     description: |
 *       Exchanges a valid refresh token for a new access token and optionally a new refresh token.
 *       
 *       **Security Features:**
 *       - Refresh token validation and rotation
 *       - Device binding verification
 *       - Automatic token revocation on suspicious activity
 *       
 *       **Token Rotation:**
 *       - New access token is always issued
 *       - Refresh token may be rotated for enhanced security
 *       - Old refresh token is invalidated upon successful refresh
 *     operationId: refreshToken
 *     requestBody:
 *       required: true
 *       content:
 *         application/json:
 *           schema:
 *             $ref: '#/components/schemas/RefreshTokenRequest'
 *           example:
 *             refreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.refresh.token.signature"
 *             deviceId: "550e8400-e29b-41d4-a716-446655440001"
 *     responses:
 *       200:
 *         description: Token refreshed successfully
 *         content:
 *           application/json:
 *             schema:
 *               allOf:
 *                 - $ref: '#/components/schemas/SuccessResponse'
 *                 - type: object
 *                   properties:
 *                     data:
 *                       $ref: '#/components/schemas/AuthResponse'
 *       401:
 *         $ref: '#/components/responses/401'
 *       500:
 *         $ref: '#/components/responses/500'
 */
app.post('/api/v1/auth/refresh', (req: Request, res: Response) => {
  const mockAuthResponse = {
    accessToken: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.new.access.token.signature',
    refreshToken: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.new.refresh.token.signature',
    tokenType: 'Bearer',
    expiresIn: 900,
    expiresAt: new Date(Date.now() + 900000).toISOString(),
    scope: ['read', 'write'],
    user: {
      id: '550e8400-e29b-41d4-a716-446655440000',
      email: 'john.doe@example.com',
      username: 'johndoe',
      firstName: 'John',
      lastName: 'Doe',
      fullName: 'John Doe',
      role: 'USER',
      emailVerified: true,
      phoneVerified: false,
      accountStatus: 'ACTIVE',
      lastLoginAt: '2024-01-15T09:30:00.000Z',
      createdAt: '2024-01-01T00:00:00.000Z',
      updatedAt: new Date().toISOString()
    }
  };

  res.json({
    success: true,
    timestamp: new Date().toISOString(),
    path: req.path,
    method: req.method,
    correlationId: req.correlationId,
    data: mockAuthResponse
  });
});

/**
 * @swagger
 * /api/v1/auth/profile:
 *   get:
 *     tags:
 *       - User Management
 *     summary: Get user profile
 *     description: |
 *       Retrieves the authenticated user's profile information including preferences and metadata.
 *       
 *       **Authentication Required:**
 *       This endpoint requires a valid JWT access token in the Authorization header.
 *       
 *       **Returned Information:**
 *       - Basic profile data (name, email, username)
 *       - Account status and verification states
 *       - User preferences and settings
 *       - Account metadata and timestamps
 *     operationId: getUserProfile
 *     security:
 *       - BearerAuth: []
 *     parameters:
 *       - $ref: '#/components/parameters/CorrelationId'
 *     responses:
 *       200:
 *         description: Profile retrieved successfully
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
app.get('/api/v1/auth/profile', (req: Request, res: Response) => {
  const mockUser = {
    id: '550e8400-e29b-41d4-a716-446655440000',
    email: 'john.doe@example.com',
    username: 'johndoe',
    firstName: 'John',
    lastName: 'Doe',
    fullName: 'John Doe',
    role: 'USER',
    emailVerified: true,
    phoneVerified: false,
    accountStatus: 'ACTIVE',
    lastLoginAt: '2024-01-15T09:30:00.000Z',
    createdAt: '2024-01-01T00:00:00.000Z',
    updatedAt: new Date().toISOString(),
    preferences: {
      language: 'en',
      timezone: 'America/New_York',
      emailNotifications: true,
      smsNotifications: false,
      marketingEmails: false,
      twoFactorEnabled: false
    },
    metadata: {
      theme: 'dark',
      onboardingCompleted: true
    }
  };

  res.json({
    success: true,
    timestamp: new Date().toISOString(),
    path: req.path,
    method: req.method,
    correlationId: req.correlationId,
    data: mockUser
  });
});

/**
 * @swagger
 * /api/v1/auth/logout:
 *   post:
 *     tags:
 *       - Authentication
 *     summary: Logout user session
 *     description: |
 *       Invalidates the current user session and optionally all sessions for the user.
 *       
 *       **Security Features:**
 *       - Token revocation and blacklisting
 *       - Session cleanup
 *       - Optional logout from all devices
 *       
 *       **Logout Options:**
 *       - Single device logout (default)
 *       - All devices logout (when logoutAllDevices=true)
 *     operationId: logoutUser
 *     security:
 *       - BearerAuth: []
 *     requestBody:
 *       required: false
 *       content:
 *         application/json:
 *           schema:
 *             type: object
 *             properties:
 *               logoutAllDevices:
 *                 type: boolean
 *                 description: Logout from all devices
 *                 default: false
 *               deviceId:
 *                 type: string
 *                 format: uuid
 *                 description: Specific device to logout
 *           example:
 *             logoutAllDevices: false
 *             deviceId: "550e8400-e29b-41d4-a716-446655440001"
 *     responses:
 *       200:
 *         description: Logout successful
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
 *                         message:
 *                           type: string
 *                           example: "Logout successful"
 *                         sessionsTerminated:
 *                           type: integer
 *                           example: 1
 *       401:
 *         $ref: '#/components/responses/401'
 *       500:
 *         $ref: '#/components/responses/500'
 */
app.post('/api/v1/auth/logout', (req: Request, res: Response) => {
  res.json({
    success: true,
    timestamp: new Date().toISOString(),
    path: req.path,
    method: req.method,
    correlationId: req.correlationId,
    data: {
      message: 'Logout successful',
      sessionsTerminated: 1
    }
  });
});

// Root redirect to API documentation
app.get('/', (req: Request, res: Response) => {
  res.redirect('/api-docs');
});

// Handle 404 errors
app.use('*', (req: Request, res: Response) => {
  res.status(404).json({
    success: false,
    timestamp: new Date().toISOString(),
    path: req.path,
    method: req.method,
    correlationId: req.correlationId,
    error: {
      code: 'NOT_FOUND',
      message: `Endpoint ${req.method} ${req.path} not found`,
      timestamp: new Date().toISOString(),
      path: req.path,
      correlationId: req.correlationId
    }
  });
});

// Global error handler
app.use((err: any, req: Request, res: Response, next: any) => {
  console.error('Error:', err);
  
  res.status(500).json({
    success: false,
    timestamp: new Date().toISOString(),
    path: req.path,
    method: req.method,
    correlationId: req.correlationId,
    error: {
      code: 'INTERNAL_SERVER_ERROR',
      message: 'An unexpected error occurred',
      timestamp: new Date().toISOString(),
      path: req.path,
      correlationId: req.correlationId
    }
  });
});

// Extend Express Request interface
declare global {
  namespace Express {
    interface Request {
      correlationId?: string;
    }
  }
}

// Start server
const server = app.listen(PORT, () => {
  console.log('🚀 X-Form Auth Service Demo Started Successfully!');
  console.log('================================================');
  console.log(`🌐 Server running on: http://localhost:${PORT}`);
  console.log(`📖 API Documentation: http://localhost:${PORT}/api-docs`);
  console.log(`📋 OpenAPI Spec: http://localhost:${PORT}/api-docs.json`);
  console.log(`🏥 Health Check: http://localhost:${PORT}/health`);
  console.log('================================================');
  console.log('🔐 Test the Authentication Flow:');
  console.log('1. Visit the Swagger UI for interactive testing');
  console.log('2. Try the register endpoint to create a user');
  console.log('3. Use login to get JWT tokens');
  console.log('4. Test protected endpoints with the Bearer token');
  console.log('================================================');
});

// Graceful shutdown
process.on('SIGTERM', () => {
  console.log('🛑 Received SIGTERM, shutting down gracefully...');
  server.close(() => {
    console.log('✅ Server closed successfully');
    process.exit(0);
  });
});

process.on('SIGINT', () => {
  console.log('🛑 Received SIGINT, shutting down gracefully...');
  server.close(() => {
    console.log('✅ Server closed successfully');
    process.exit(0);
  });
});

export default app;
