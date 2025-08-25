// Versioned API Routes for Auth Service
// Implementing microservices best practices with proper versioning

import { Router } from 'express';
import { validationResult } from 'express-validator';
import { AuthController } from '../auth-controller';
import { ApiResponseHandler, extractValidationErrors } from '../api-response-handler';
import { getHealthCheckService } from '../../../infrastructure/monitoring/health-check';
import { logger, logSecurityEvent, logAuditEvent } from '../../../infrastructure/logging/structured-logger';
import {
  validateRegisterRequest,
  validateLoginRequest,
  validateRefreshTokenRequest,
  validateVerifyEmailRequest,
  validateForgotPasswordRequest,
  validateResetPasswordRequest,
  validateChangePasswordRequest,
  validateUpdateProfileRequest,
  validateUserIdParam,
  validatePaginationQuery,
} from '../validation/auth-validators';

/**
 * @swagger
 * components:
 *   securitySchemes:
 *     BearerAuth:
 *       type: http
 *       scheme: bearer
 *       bearerFormat: JWT
 */

export class AuthRoutesV1 {
  private router: Router;
  private authController: AuthController;

  constructor(authController: AuthController) {
    this.router = Router();
    this.authController = authController;
    this.setupRoutes();
  }

  private setupRoutes(): void {
    // Authentication endpoints
    this.setupAuthenticationRoutes();
    
    // User management endpoints
    this.setupUserManagementRoutes();
    
    // Email verification endpoints
    this.setupEmailVerificationRoutes();
    
    // Password management endpoints
    this.setupPasswordManagementRoutes();
    
    // Health and monitoring endpoints
    this.setupHealthRoutes();
  }

  private setupAuthenticationRoutes(): void {
    /**
     * @swagger
     * /api/v1/auth/register:
     *   post:
     *     tags: [Authentication]
     *     summary: Register a new user account
     *     description: |
     *       Register a new user with email verification.
     *       
     *       **Security Features:**
     *       - Password strength validation
     *       - Email format validation
     *       - Username uniqueness check
     *       - Rate limiting protection
     *       
     *       **Business Rules:**
     *       - Email must be unique across the system
     *       - Username must be unique and follow naming conventions
     *       - Password must meet complexity requirements
     *       - Terms acceptance is mandatory
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
     *                       $ref: '#/components/schemas/RegisterResponseDTO'
     *       400:
     *         $ref: '#/components/responses/400'
     *       409:
     *         description: User already exists
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ErrorResponse'
     *       429:
     *         $ref: '#/components/responses/429'
     *       500:
     *         $ref: '#/components/responses/500'
     */
    this.router.post('/register', 
      validateRegisterRequest,
      this.handleValidation,
      this.authController.register
    );

    /**
     * @swagger
     * /api/v1/auth/login:
     *   post:
     *     tags: [Authentication]
     *     summary: Authenticate user and get tokens
     *     description: |
     *       Authenticate user with email and password, returns JWT tokens.
     *       
     *       **Security Features:**
     *       - Account lockout after failed attempts
     *       - Device tracking and trusted devices
     *       - Session management
     *       - Audit logging
     *       
     *       **Response includes:**
     *       - Access token (short-lived)
     *       - Refresh token (long-lived)
     *       - User profile information
     *       - Token expiration details
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
     *                       $ref: '#/components/schemas/LoginResponseDTO'
     *       400:
     *         $ref: '#/components/responses/400'
     *       401:
     *         description: Invalid credentials
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ErrorResponse'
     *       423:
     *         description: Account locked
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ErrorResponse'
     *       429:
     *         $ref: '#/components/responses/429'
     *       500:
     *         $ref: '#/components/responses/500'
     */
    this.router.post('/login',
      validateLoginRequest,
      this.handleValidation,
      this.authController.login
    );

    /**
     * @swagger
     * /api/v1/auth/refresh:
     *   post:
     *     tags: [Authentication]
     *     summary: Refresh access token
     *     description: |
     *       Use refresh token to get a new access token.
     *       
     *       **Security Features:**
     *       - Refresh token rotation
     *       - Device validation
     *       - Token family tracking
     *       - Automatic revocation on suspicious activity
     *     requestBody:
     *       required: true
     *       content:
     *         application/json:
     *           schema:
     *             $ref: '#/components/schemas/RefreshTokenRequest'
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
     *       400:
     *         $ref: '#/components/responses/400'
     *       401:
     *         description: Invalid or expired refresh token
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ErrorResponse'
     *       500:
     *         $ref: '#/components/responses/500'
     */
    this.router.post('/refresh',
      validateRefreshTokenRequest,
      this.handleValidation,
      this.authController.refreshToken
    );

    /**
     * @swagger
     * /api/v1/auth/logout:
     *   post:
     *     tags: [Authentication]
     *     summary: Logout user and revoke tokens
     *     description: |
     *       Logout user and revoke all tokens.
     *       
     *       **Security Features:**
     *       - Token revocation
     *       - Session termination
     *       - Device logout tracking
     *       - Audit logging
     *     security:
     *       - BearerAuth: []
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
     *                       $ref: '#/components/schemas/LogoutResponseDTO'
     *       401:
     *         $ref: '#/components/responses/401'
     *       500:
     *         $ref: '#/components/responses/500'
     */
    this.router.post('/logout',
      this.authenticateToken,
      this.authController.logout
    );
  }

  private setupUserManagementRoutes(): void {
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
    this.router.get('/profile',
      this.authenticateToken,
      this.authController.getProfile
    );

    /**
     * @swagger
     * /api/v1/auth/profile:
     *   put:
     *     tags: [User Management]
     *     summary: Update user profile
     *     description: Update authenticated user's profile information
     *     security:
     *       - BearerAuth: []
     *     requestBody:
     *       required: true
     *       content:
     *         application/json:
     *           schema:
     *             $ref: '#/components/schemas/UpdateProfileRequestDTO'
     *     responses:
     *       200:
     *         description: Profile updated successfully
     *         content:
     *           application/json:
     *             schema:
     *               allOf:
     *                 - $ref: '#/components/schemas/SuccessResponse'
     *                 - type: object
     *                   properties:
     *                     data:
     *                       $ref: '#/components/schemas/UserProfile'
     *       400:
     *         $ref: '#/components/responses/400'
     *       401:
     *         $ref: '#/components/responses/401'
     *       500:
     *         $ref: '#/components/responses/500'
     */
    this.router.put('/profile',
      this.authenticateToken,
      validateUpdateProfileRequest,
      this.handleValidation,
      this.authController.updateProfile
    );
  }

  private setupEmailVerificationRoutes(): void {
    /**
     * @swagger
     * /api/v1/auth/verify-email:
     *   post:
     *     tags: [Email Verification]
     *     summary: Verify email address
     *     description: Verify user's email address using verification token
     *     requestBody:
     *       required: true
     *       content:
     *         application/json:
     *           schema:
     *             $ref: '#/components/schemas/VerifyEmailRequestDTO'
     *     responses:
     *       200:
     *         description: Email verified successfully
     *         content:
     *           application/json:
     *             schema:
     *               allOf:
     *                 - $ref: '#/components/schemas/SuccessResponse'
     *                 - type: object
     *                   properties:
     *                     data:
     *                       $ref: '#/components/schemas/VerifyEmailResponseDTO'
     *       400:
     *         $ref: '#/components/responses/400'
     *       410:
     *         description: Verification token expired
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ErrorResponse'
     *       500:
     *         $ref: '#/components/responses/500'
     */
    this.router.post('/verify-email',
      validateVerifyEmailRequest,
      this.handleValidation,
      this.authController.verifyEmail
    );

    /**
     * @swagger
     * /api/v1/auth/resend-verification:
     *   post:
     *     tags: [Email Verification]
     *     summary: Resend verification email
     *     description: Resend email verification token
     *     security:
     *       - BearerAuth: []
     *     responses:
     *       200:
     *         description: Verification email sent successfully
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/SuccessResponse'
     *       401:
     *         $ref: '#/components/responses/401'
     *       429:
     *         $ref: '#/components/responses/429'
     *       500:
     *         $ref: '#/components/responses/500'
     */
    this.router.post('/resend-verification',
      this.authenticateToken,
      this.authController.resendVerification
    );
  }

  private setupPasswordManagementRoutes(): void {
    /**
     * @swagger
     * /api/v1/auth/forgot-password:
     *   post:
     *     tags: [Password Management]
     *     summary: Request password reset
     *     description: Send password reset token to user's email
     *     requestBody:
     *       required: true
     *       content:
     *         application/json:
     *           schema:
     *             $ref: '#/components/schemas/ForgotPasswordRequestDTO'
     *     responses:
     *       200:
     *         description: Password reset email sent
     *         content:
     *           application/json:
     *             schema:
     *               allOf:
     *                 - $ref: '#/components/schemas/SuccessResponse'
     *                 - type: object
     *                   properties:
     *                     data:
     *                       $ref: '#/components/schemas/ForgotPasswordResponseDTO'
     *       400:
     *         $ref: '#/components/responses/400'
     *       404:
     *         description: User not found
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ErrorResponse'
     *       429:
     *         $ref: '#/components/responses/429'
     *       500:
     *         $ref: '#/components/responses/500'
     */
    this.router.post('/forgot-password',
      validateForgotPasswordRequest,
      this.handleValidation,
      this.authController.forgotPassword
    );

    /**
     * @swagger
     * /api/v1/auth/reset-password:
     *   post:
     *     tags: [Password Management]
     *     summary: Reset password with token
     *     description: Reset user password using reset token
     *     requestBody:
     *       required: true
     *       content:
     *         application/json:
     *           schema:
     *             $ref: '#/components/schemas/ResetPasswordRequestDTO'
     *     responses:
     *       200:
     *         description: Password reset successfully
     *         content:
     *           application/json:
     *             schema:
     *               allOf:
     *                 - $ref: '#/components/schemas/SuccessResponse'
     *                 - type: object
     *                   properties:
     *                     data:
     *                       $ref: '#/components/schemas/ResetPasswordResponseDTO'
     *       400:
     *         $ref: '#/components/responses/400'
     *       410:
     *         description: Reset token expired
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ErrorResponse'
     *       500:
     *         $ref: '#/components/responses/500'
     */
    this.router.post('/reset-password',
      validateResetPasswordRequest,
      this.handleValidation,
      this.authController.resetPassword
    );

    /**
     * @swagger
     * /api/v1/auth/change-password:
     *   post:
     *     tags: [Password Management]
     *     summary: Change user password
     *     description: Change authenticated user's password
     *     security:
     *       - BearerAuth: []
     *     requestBody:
     *       required: true
     *       content:
     *         application/json:
     *           schema:
     *             $ref: '#/components/schemas/ChangePasswordRequestDTO'
     *     responses:
     *       200:
     *         description: Password changed successfully
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/SuccessResponse'
     *       400:
     *         $ref: '#/components/responses/400'
     *       401:
     *         $ref: '#/components/responses/401'
     *       500:
     *         $ref: '#/components/responses/500'
     */
    this.router.post('/change-password',
      this.authenticateToken,
      validateChangePasswordRequest,
      this.handleValidation,
      this.authController.changePassword
    );
  }

  private setupHealthRoutes(): void {
    /**
     * @swagger
     * /api/v1/auth/health:
     *   get:
     *     tags: [Health & Monitoring]
     *     summary: Get service health status
     *     description: |
     *       Comprehensive health check including dependencies.
     *       
     *       **Checks include:**
     *       - Database connectivity
     *       - External service availability
     *       - Circuit breaker status
     *       - Memory and CPU usage
     *       - Service metrics
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
     *       503:
     *         description: Service is unhealthy
     *         content:
     *           application/json:
     *             schema:
     *               allOf:
     *                 - $ref: '#/components/schemas/SuccessResponse'
     *                 - type: object
     *                   properties:
     *                     data:
     *                       $ref: '#/components/schemas/HealthCheck'
     */
    this.router.get('/health', async (req, res) => {
      try {
        const healthService = getHealthCheckService();
        const health = await healthService.performHealthCheck();
        
        const statusCode = health.status === 'UNHEALTHY' ? 503 : 200;
        ApiResponseHandler.success(res, health, statusCode);
      } catch (error) {
        logger.error('Health check failed', error as Error);
        ApiResponseHandler.internalError(res, 'Health check failed');
      }
    });

    /**
     * @swagger
     * /api/v1/auth/ready:
     *   get:
     *     tags: [Health & Monitoring]
     *     summary: Get service readiness status
     *     description: Check if service is ready to accept traffic
     *     responses:
     *       200:
     *         description: Service is ready
     *       503:
     *         description: Service is not ready
     */
    this.router.get('/ready', async (req, res) => {
      try {
        const healthService = getHealthCheckService();
        const isReady = await healthService.getReadiness();
        
        if (isReady) {
          ApiResponseHandler.success(res, { ready: true, message: 'Service is ready' });
        } else {
          ApiResponseHandler.serviceUnavailable(res, 'Service is not ready');
        }
      } catch (error) {
        logger.error('Readiness check failed', error as Error);
        ApiResponseHandler.serviceUnavailable(res, 'Readiness check failed');
      }
    });

    /**
     * @swagger
     * /api/v1/auth/live:
     *   get:
     *     tags: [Health & Monitoring]
     *     summary: Get service liveness status
     *     description: Check if service is alive and responsive
     *     responses:
     *       200:
     *         description: Service is alive
     *       503:
     *         description: Service is not responsive
     */
    this.router.get('/live', async (req, res) => {
      try {
        const healthService = getHealthCheckService();
        const isAlive = await healthService.getLiveness();
        
        if (isAlive) {
          ApiResponseHandler.success(res, { alive: true, message: 'Service is alive' });
        } else {
          ApiResponseHandler.serviceUnavailable(res, 'Service is not responsive');
        }
      } catch (error) {
        logger.error('Liveness check failed', error as Error);
        ApiResponseHandler.serviceUnavailable(res, 'Liveness check failed');
      }
    });
  }

  // Middleware functions
  private handleValidation = (req: any, res: any, next: any): void => {
    const errors = validationResult(req);
    if (!errors.isEmpty()) {
      const validationErrors = extractValidationErrors(errors.array());
      
      // Log validation failure for security monitoring
      logSecurityEvent(
        'LOGIN_FAILED',
        'LOW',
        'Validation failed',
        {
          ipAddress: req.ip,
          userAgent: req.get('User-Agent'),
          metadata: { validationErrors, path: req.path },
        }
      );
      
      ApiResponseHandler.validationError(res, validationErrors);
      return;
    }
    next();
  };

  private authenticateToken = (req: any, res: any, next: any): void => {
    const authHeader = req.headers.authorization;
    const token = authHeader && authHeader.split(' ')[1];

    if (!token) {
      return ApiResponseHandler.unauthorized(res, 'Access token required');
    }

    // Token validation would be implemented here
    // For now, we'll simulate successful authentication
    req.user = { id: 'user-123', email: 'user@example.com' };
    next();
  };

  getRouter(): Router {
    return this.router;
  }
}

export const createAuthRoutesV1 = (authController: AuthController): Router => {
  const authRoutes = new AuthRoutesV1(authController);
  return authRoutes.getRouter();
};
