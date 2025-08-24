const express = require('express');
const passport = require('passport');
const rateLimit = require('express-rate-limit');
const { body, validationResult } = require('express-validator');

const AuthService = require('../services/AuthService');
const GoogleAuthService = require('../services/GoogleAuthService');
const { userSchema } = require('../models/User');
const { 
  authenticateToken, 
  sensitiveOperationLimiter, 
  addRequestContext,
  requireEmailVerification 
} = require('../middleware/auth');

const router = express.Router();

// Initialize services
const authService = new AuthService();
const googleAuthService = new GoogleAuthService();

// Add request context to all routes
router.use(addRequestContext);

// Rate limiting for auth endpoints
const authLimiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 5, // 5 attempts per window
  message: {
    error: 'Too many authentication attempts. Please try again later.',
    code: 'AUTH_RATE_LIMIT_EXCEEDED'
  },
  standardHeaders: true,
  legacyHeaders: false
});

const generalLimiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100, // 100 requests per window
  message: {
    error: 'Too many requests. Please try again later.',
    code: 'GENERAL_RATE_LIMIT_EXCEEDED'
  }
});

// Validation middleware
const validateRequest = (schema) => {
  return async (req, res, next) => {
    try {
      const { error, value } = schema.validate(req.body, { abortEarly: false });
      
      if (error) {
        return res.status(400).json({
          error: 'Validation failed',
          code: 'VALIDATION_ERROR',
          details: error.details.map(detail => ({
            field: detail.path.join('.'),
            message: detail.message
          }))
        });
      }
      
      req.validatedBody = value;
      next();
    } catch (err) {
      return res.status(400).json({
        error: 'Invalid request data',
        code: 'INVALID_REQUEST_DATA'
      });
    }
  };
};

// Health check
router.get('/health', (req, res) => {
  res.json({
    status: 'healthy',
    service: 'auth-service',
    timestamp: new Date().toISOString(),
    version: '1.0.0'
  });
});

// POST /auth/signup - Register new user
router.post('/signup', 
  generalLimiter,
  sensitiveOperationLimiter(3, 60 * 60 * 1000), // 3 attempts per hour
  validateRequest(userSchema.signup),
  async (req, res) => {
    try {
      const result = await authService.register(
        req.validatedBody,
        req.clientIP,
        req.userAgent
      );

      res.status(201).json({
        message: 'User registered successfully',
        user: result.user,
        emailVerificationRequired: true
      });

    } catch (error) {
      if (error.message.includes('already exists')) {
        return res.status(409).json({
          error: 'User already exists with this email',
          code: 'USER_ALREADY_EXISTS'
        });
      }

      console.error('Signup error:', error);
      res.status(500).json({
        error: 'Registration failed',
        code: 'REGISTRATION_FAILED'
      });
    }
  }
);

// POST /auth/login - User login
router.post('/login',
  authLimiter,
  validateRequest(userSchema.login),
  async (req, res) => {
    try {
      const { email, password, rememberMe } = req.validatedBody;
      
      const result = await authService.login(
        email,
        password,
        rememberMe,
        req.clientIP,
        req.userAgent
      );

      // Set secure HTTP-only cookie for refresh token
      res.cookie('refreshToken', result.refreshToken, {
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        sameSite: 'strict',
        maxAge: rememberMe ? 30 * 24 * 60 * 60 * 1000 : 7 * 24 * 60 * 60 * 1000
      });

      res.json({
        message: 'Login successful',
        user: result.user,
        accessToken: result.accessToken,
        expiresIn: result.expiresIn
      });

    } catch (error) {
      if (error.message.includes('Invalid email or password')) {
        return res.status(401).json({
          error: 'Invalid email or password',
          code: 'INVALID_CREDENTIALS'
        });
      }

      if (error.message.includes('locked')) {
        return res.status(423).json({
          error: 'Account is locked due to multiple failed login attempts',
          code: 'ACCOUNT_LOCKED'
        });
      }

      if (error.message.includes('not active')) {
        return res.status(403).json({
          error: 'Account is not active',
          code: 'ACCOUNT_INACTIVE'
        });
      }

      console.error('Login error:', error);
      res.status(500).json({
        error: 'Login failed',
        code: 'LOGIN_FAILED'
      });
    }
  }
);

// POST /auth/refresh - Refresh access token
router.post('/refresh',
  generalLimiter,
  validateRequest(userSchema.refreshToken),
  async (req, res) => {
    try {
      const { refreshToken } = req.validatedBody;
      
      const result = await authService.refreshAccessToken(
        refreshToken,
        req.clientIP,
        req.userAgent
      );

      res.json({
        message: 'Token refreshed successfully',
        user: result.user,
        accessToken: result.accessToken,
        expiresIn: result.expiresIn
      });

    } catch (error) {
      if (error.message.includes('Invalid or expired')) {
        return res.status(401).json({
          error: 'Invalid or expired refresh token',
          code: 'INVALID_REFRESH_TOKEN'
        });
      }

      console.error('Token refresh error:', error);
      res.status(500).json({
        error: 'Token refresh failed',
        code: 'TOKEN_REFRESH_FAILED'
      });
    }
  }
);

// POST /auth/logout - Logout user
router.post('/logout',
  generalLimiter,
  validateRequest(userSchema.logout),
  async (req, res) => {
    try {
      const { refreshToken } = req.validatedBody;
      
      await authService.logout(
        refreshToken,
        req.clientIP,
        req.userAgent
      );

      // Clear refresh token cookie
      res.clearCookie('refreshToken');

      res.json({
        message: 'Logged out successfully'
      });

    } catch (error) {
      console.error('Logout error:', error);
      res.status(500).json({
        error: 'Logout failed',
        code: 'LOGOUT_FAILED'
      });
    }
  }
);

// GET /auth/me - Get current user profile
router.get('/me',
  generalLimiter,
  authenticateToken,
  async (req, res) => {
    try {
      // Get fresh user data
      const user = await authService.getUserById(req.user.id);
      
      res.json({
        user: user
      });

    } catch (error) {
      console.error('Get profile error:', error);
      res.status(500).json({
        error: 'Failed to get user profile',
        code: 'PROFILE_FETCH_FAILED'
      });
    }
  }
);

// Google OAuth routes
// GET /auth/google - Initiate Google OAuth
router.get('/google',
  generalLimiter,
  passport.authenticate('google', { 
    scope: ['profile', 'email'],
    prompt: 'select_account' // Force account selection
  })
);

// GET /auth/google/callback - Handle Google OAuth callback
router.get('/google/callback',
  generalLimiter,
  passport.authenticate('google', { session: false }),
  async (req, res) => {
    try {
      const result = await googleAuthService.loginWithGoogle(
        req.user,
        req.clientIP,
        req.userAgent
      );

      // Set secure HTTP-only cookie for refresh token
      res.cookie('refreshToken', result.refreshToken, {
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        sameSite: 'strict',
        maxAge: 7 * 24 * 60 * 60 * 1000 // 7 days
      });

      // Redirect to frontend with success
      const frontendUrl = process.env.FRONTEND_URL || 'http://localhost:3000';
      const redirectUrl = `${frontendUrl}/auth/callback?success=true&token=${result.accessToken}`;
      
      res.redirect(redirectUrl);

    } catch (error) {
      console.error('Google OAuth callback error:', error);
      
      // Redirect to frontend with error
      const frontendUrl = process.env.FRONTEND_URL || 'http://localhost:3000';
      const redirectUrl = `${frontendUrl}/auth/callback?error=oauth_failed`;
      
      res.redirect(redirectUrl);
    }
  }
);

// PUT /auth/profile - Update user profile
router.put('/profile',
  generalLimiter,
  authenticateToken,
  validateRequest(userSchema.updateProfile),
  async (req, res) => {
    try {
      const userId = req.user.id;
      const updates = req.validatedBody;

      // Update user profile
      const updateFields = [];
      const updateValues = [];
      let paramIndex = 1;

      Object.keys(updates).forEach(key => {
        const dbField = key === 'firstName' ? 'first_name' : 
                       key === 'lastName' ? 'last_name' : 
                       key === 'avatarUrl' ? 'avatar_url' : key;
        
        updateFields.push(`${dbField} = $${paramIndex}`);
        updateValues.push(updates[key]);
        paramIndex++;
      });

      updateValues.push(userId);

      await authService.pool.query(`
        UPDATE users 
        SET ${updateFields.join(', ')}, updated_at = CURRENT_TIMESTAMP
        WHERE id = $${paramIndex}
      `, updateValues);

      // Get updated user
      const updatedUser = await authService.getUserById(userId);

      res.json({
        message: 'Profile updated successfully',
        user: updatedUser
      });

    } catch (error) {
      console.error('Profile update error:', error);
      res.status(500).json({
        error: 'Failed to update profile',
        code: 'PROFILE_UPDATE_FAILED'
      });
    }
  }
);

// POST /auth/change-password - Change password
router.post('/change-password',
  generalLimiter,
  authenticateToken,
  sensitiveOperationLimiter(3, 60 * 60 * 1000), // 3 attempts per hour
  validateRequest(userSchema.changePassword),
  async (req, res) => {
    try {
      const { currentPassword, newPassword } = req.validatedBody;
      const userId = req.user.id;

      // Get current user with password hash
      const userResult = await authService.pool.query(
        'SELECT password_hash FROM users WHERE id = $1',
        [userId]
      );

      if (userResult.rows.length === 0) {
        return res.status(404).json({
          error: 'User not found',
          code: 'USER_NOT_FOUND'
        });
      }

      // Verify current password
      const isValidPassword = await authService.comparePassword(
        currentPassword, 
        userResult.rows[0].password_hash
      );

      if (!isValidPassword) {
        return res.status(400).json({
          error: 'Current password is incorrect',
          code: 'INVALID_CURRENT_PASSWORD'
        });
      }

      // Hash new password
      const newPasswordHash = await authService.hashPassword(newPassword);

      // Update password
      await authService.pool.query(`
        UPDATE users 
        SET password_hash = $1, updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
      `, [newPasswordHash, userId]);

      // Revoke all refresh tokens to force re-login
      await authService.pool.query(`
        UPDATE refresh_tokens 
        SET is_revoked = TRUE 
        WHERE user_id = $1
      `, [userId]);

      // Log password change
      await authService.logAuthEvent(userId, 'password_change', req.clientIP, req.userAgent, {}, true);

      res.json({
        message: 'Password changed successfully. Please log in again.'
      });

    } catch (error) {
      console.error('Password change error:', error);
      res.status(500).json({
        error: 'Failed to change password',
        code: 'PASSWORD_CHANGE_FAILED'
      });
    }
  }
);

// Error handling middleware
router.use((error, req, res, next) => {
  console.error('Auth route error:', error);
  
  if (error.type === 'validation') {
    return res.status(400).json({
      error: 'Validation error',
      code: 'VALIDATION_ERROR',
      details: error.details
    });
  }

  res.status(500).json({
    error: 'Internal server error',
    code: 'INTERNAL_SERVER_ERROR'
  });
});

module.exports = router;
