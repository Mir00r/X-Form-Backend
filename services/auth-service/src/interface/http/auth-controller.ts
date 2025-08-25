// Interface Layer - HTTP handlers and controllers
// Following Single Responsibility Principle: Only handles HTTP concerns

import { Request, Response, NextFunction } from 'express';
import { AuthApplicationService } from '../../application/auth-service';
import {
  RegisterUserRequest,
  LoginRequest,
  RefreshTokenRequest,
  VerifyEmailRequest,
  ForgotPasswordRequest,
  ResetPasswordRequest,
} from '../../application/auth-service';
import {
  DomainError,
  UserNotFoundError,
  UserAlreadyExistsError,
  InvalidCredentialsError,
  AccountLockedError,
  EmailNotVerifiedError,
  TokenExpiredError,
  InvalidTokenError,
} from '../../domain/auth';

// HTTP Response interfaces following Interface Segregation Principle
interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: any;
  };
}

interface ValidationError {
  field: string;
  message: string;
}

// Auth Controller following Single Responsibility Principle
export class AuthController {
  constructor(private readonly authService: AuthApplicationService) {}

  // POST /auth/register
  register = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const validationErrors = this.validateRegisterRequest(req.body);
      if (validationErrors.length > 0) {
        res.status(400).json(this.createErrorResponse('VALIDATION_ERROR', 'Validation failed', validationErrors));
        return;
      }

      const registerRequest: RegisterUserRequest = {
        email: req.body.email,
        username: req.body.username,
        password: req.body.password,
        firstName: req.body.firstName,
        lastName: req.body.lastName,
      };

      const result = await this.authService.registerUser(registerRequest);

      res.status(201).json(this.createSuccessResponse(result, 'User registered successfully'));
    } catch (error) {
      next(error);
    }
  };

  // POST /auth/login
  login = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const validationErrors = this.validateLoginRequest(req.body);
      if (validationErrors.length > 0) {
        res.status(400).json(this.createErrorResponse('VALIDATION_ERROR', 'Validation failed', validationErrors));
        return;
      }

      const loginRequest: LoginRequest = {
        email: req.body.email,
        password: req.body.password,
        ipAddress: this.getClientIP(req),
        userAgent: req.get('User-Agent') || '',
      };

      const result = await this.authService.loginUser(loginRequest);

      // Set refresh token as httpOnly cookie
      res.cookie('refreshToken', result.refreshToken, {
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        sameSite: 'strict',
        maxAge: 7 * 24 * 60 * 60 * 1000, // 7 days
      });

      // Don't send refresh token in response body
      const response = {
        accessToken: result.accessToken,
        user: result.user,
      };

      res.status(200).json(this.createSuccessResponse(response, 'Login successful'));
    } catch (error) {
      next(error);
    }
  };

  // POST /auth/refresh
  refreshToken = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const refreshToken = req.cookies.refreshToken || req.body.refreshToken;
      
      if (!refreshToken) {
        res.status(401).json(this.createErrorResponse('MISSING_TOKEN', 'Refresh token is required'));
        return;
      }

      const refreshRequest: RefreshTokenRequest = { refreshToken };
      const result = await this.authService.refreshToken(refreshRequest);

      // Update refresh token cookie
      res.cookie('refreshToken', result.refreshToken, {
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        sameSite: 'strict',
        maxAge: 7 * 24 * 60 * 60 * 1000, // 7 days
      });

      const response = {
        accessToken: result.accessToken,
        user: result.user,
      };

      res.status(200).json(this.createSuccessResponse(response, 'Token refreshed successfully'));
    } catch (error) {
      next(error);
    }
  };

  // POST /auth/verify-email
  verifyEmail = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const { token } = req.body;
      
      if (!token) {
        res.status(400).json(this.createErrorResponse('MISSING_TOKEN', 'Verification token is required'));
        return;
      }

      const verifyRequest: VerifyEmailRequest = { token };
      await this.authService.verifyEmail(verifyRequest);

      res.status(200).json(this.createSuccessResponse(null, 'Email verified successfully'));
    } catch (error) {
      next(error);
    }
  };

  // POST /auth/forgot-password
  forgotPassword = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const { email } = req.body;
      
      if (!email) {
        res.status(400).json(this.createErrorResponse('MISSING_EMAIL', 'Email is required'));
        return;
      }

      const forgotRequest: ForgotPasswordRequest = { email };
      await this.authService.forgotPassword(forgotRequest);

      // Always return success for security (don't reveal if email exists)
      res.status(200).json(this.createSuccessResponse(null, 'Password reset email sent if account exists'));
    } catch (error) {
      next(error);
    }
  };

  // POST /auth/reset-password
  resetPassword = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const validationErrors = this.validateResetPasswordRequest(req.body);
      if (validationErrors.length > 0) {
        res.status(400).json(this.createErrorResponse('VALIDATION_ERROR', 'Validation failed', validationErrors));
        return;
      }

      const resetRequest: ResetPasswordRequest = {
        token: req.body.token,
        newPassword: req.body.newPassword,
      };

      await this.authService.resetPassword(resetRequest);

      res.status(200).json(this.createSuccessResponse(null, 'Password reset successfully'));
    } catch (error) {
      next(error);
    }
  };

  // POST /auth/logout
  logout = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const refreshToken = req.cookies.refreshToken;
      
      if (refreshToken) {
        await this.authService.logout(refreshToken);
      }

      // Clear refresh token cookie
      res.clearCookie('refreshToken');

      res.status(200).json(this.createSuccessResponse(null, 'Logout successful'));
    } catch (error) {
      next(error);
    }
  };

  // GET /auth/profile
  getProfile = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const userId = (req as any).user?.userId; // Set by auth middleware
      
      if (!userId) {
        res.status(401).json(this.createErrorResponse('UNAUTHORIZED', 'User not authenticated'));
        return;
      }

      const profile = await this.authService.getUserProfile(userId);

      res.status(200).json(this.createSuccessResponse(profile, 'Profile retrieved successfully'));
    } catch (error) {
      next(error);
    }
  };

  // POST /auth/profile
  updateProfile = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const userId = (req as any).user?.userId;
      
      if (!userId) {
        res.status(401).json(this.createErrorResponse('UNAUTHORIZED', 'User not authenticated'));
        return;
      }

      const updateData = {
        firstName: req.body.firstName,
        lastName: req.body.lastName,
        bio: req.body.bio,
        avatar: req.body.avatar,
      };

      const updatedProfile = await this.authService.updateUserProfile(userId, updateData);

      res.status(200).json(this.createSuccessResponse(updatedProfile, 'Profile updated successfully'));
    } catch (error) {
      next(error);
    }
  };

  // POST /auth/resend-verification
  resendVerification = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const userId = (req as any).user?.userId;
      
      if (!userId) {
        res.status(401).json(this.createErrorResponse('UNAUTHORIZED', 'User not authenticated'));
        return;
      }

      await this.authService.resendVerificationEmail(userId);

      res.status(200).json(this.createSuccessResponse(null, 'Verification email sent successfully'));
    } catch (error) {
      next(error);
    }
  };

  // POST /auth/change-password
  changePassword = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const userId = (req as any).user?.userId;
      
      if (!userId) {
        res.status(401).json(this.createErrorResponse('UNAUTHORIZED', 'User not authenticated'));
        return;
      }

      const { currentPassword, newPassword } = req.body;

      await this.authService.changePassword(userId, currentPassword, newPassword);

      res.status(200).json(this.createSuccessResponse(null, 'Password changed successfully'));
    } catch (error) {
      next(error);
    }
  };

  // Private validation methods following Single Responsibility Principle
  private validateRegisterRequest(body: any): ValidationError[] {
    const errors: ValidationError[] = [];

    if (!body.email) {
      errors.push({ field: 'email', message: 'Email is required' });
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(body.email)) {
      errors.push({ field: 'email', message: 'Invalid email format' });
    }

    if (!body.username) {
      errors.push({ field: 'username', message: 'Username is required' });
    } else if (body.username.length < 3) {
      errors.push({ field: 'username', message: 'Username must be at least 3 characters' });
    }

    if (!body.password) {
      errors.push({ field: 'password', message: 'Password is required' });
    } else if (body.password.length < 8) {
      errors.push({ field: 'password', message: 'Password must be at least 8 characters' });
    }

    if (!body.firstName) {
      errors.push({ field: 'firstName', message: 'First name is required' });
    }

    if (!body.lastName) {
      errors.push({ field: 'lastName', message: 'Last name is required' });
    }

    return errors;
  }

  private validateLoginRequest(body: any): ValidationError[] {
    const errors: ValidationError[] = [];

    if (!body.email) {
      errors.push({ field: 'email', message: 'Email is required' });
    }

    if (!body.password) {
      errors.push({ field: 'password', message: 'Password is required' });
    }

    return errors;
  }

  private validateResetPasswordRequest(body: any): ValidationError[] {
    const errors: ValidationError[] = [];

    if (!body.token) {
      errors.push({ field: 'token', message: 'Reset token is required' });
    }

    if (!body.newPassword) {
      errors.push({ field: 'newPassword', message: 'New password is required' });
    } else if (body.newPassword.length < 8) {
      errors.push({ field: 'newPassword', message: 'Password must be at least 8 characters' });
    }

    return errors;
  }

  private getClientIP(req: Request): string {
    return (req.ip || 
            req.connection.remoteAddress || 
            req.socket.remoteAddress || 
            (req.connection as any)?.socket?.remoteAddress || 
            '0.0.0.0');
  }

  private createSuccessResponse<T>(data: T, message?: string): ApiResponse<T> {
    return {
      success: true,
      data,
      ...(message && { message }),
    };
  }

  private createErrorResponse(code: string, message: string, details?: any): ApiResponse {
    return {
      success: false,
      error: {
        code,
        message,
        ...(details && { details }),
      },
    };
  }
}

// Error handler middleware following Open/Closed Principle
export const authErrorHandler = (error: Error, req: Request, res: Response, next: NextFunction): void => {
  console.error('Auth error:', error);

  if (error instanceof DomainError) {
    const statusCode = getStatusCodeForDomainError(error);
    res.status(statusCode).json({
      success: false,
      error: {
        code: (error as any).code || 'INTERNAL_ERROR',
        message: error.message,
      },
    });
    return;
  }

  // JWT errors
  if (error.name === 'JsonWebTokenError') {
    res.status(401).json({
      success: false,
      error: {
        code: 'INVALID_TOKEN',
        message: 'Invalid token',
      },
    });
    return;
  }

  if (error.name === 'TokenExpiredError') {
    res.status(401).json({
      success: false,
      error: {
        code: 'TOKEN_EXPIRED',
        message: 'Token has expired',
      },
    });
    return;
  }

  // Default error
  res.status(500).json({
    success: false,
    error: {
      code: 'INTERNAL_SERVER_ERROR',
      message: 'An unexpected error occurred',
    },
  });
};

// Helper function to map domain errors to HTTP status codes
function getStatusCodeForDomainError(error: DomainError): number {
  switch (error.constructor) {
    case UserNotFoundError:
      return 404;
    case UserAlreadyExistsError:
      return 409;
    case InvalidCredentialsError:
      return 401;
    case AccountLockedError:
      return 423;
    case EmailNotVerifiedError:
      return 403;
    case TokenExpiredError:
      return 401;
    case InvalidTokenError:
      return 401;
    default:
      return 400;
  }
}
