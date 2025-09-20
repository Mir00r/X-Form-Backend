// Comprehensive Input Validation for Auth Service
// Using express-validator for robust validation following microservices best practices

import { body, param, query, ValidationChain } from 'express-validator';

// Common validation rules
const emailValidation = body('email')
  .isEmail()
  .withMessage('Must be a valid email address')
  .normalizeEmail()
  .isLength({ max: 320 })
  .withMessage('Email must not exceed 320 characters');

const passwordValidation = body('password')
  .isLength({ min: 8, max: 128 })
  .withMessage('Password must be between 8 and 128 characters')
  .matches(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]/)
  .withMessage('Password must contain at least one uppercase letter, one lowercase letter, one number, and one special character');

const usernameValidation = body('username')
  .isLength({ min: 3, max: 30 })
  .withMessage('Username must be between 3 and 30 characters')
  .matches(/^[a-zA-Z0-9_.-]+$/)
  .withMessage('Username can only contain letters, numbers, underscores, dots, and hyphens')
  .custom((value) => {
    if (value.startsWith('.') || value.endsWith('.') || value.includes('..')) {
      throw new Error('Username cannot start or end with a dot, or contain consecutive dots');
    }
    return true;
  });

const nameValidation = (field: string) => 
  body(field)
    .isLength({ min: 1, max: 50 })
    .withMessage(`${field} must be between 1 and 50 characters`)
    .matches(/^[a-zA-Z\s'-]+$/)
    .withMessage(`${field} can only contain letters, spaces, apostrophes, and hyphens`)
    .trim();

// Registration validation
export const validateRegisterRequest: ValidationChain[] = [
  emailValidation,
  usernameValidation,
  passwordValidation,
  body('confirmPassword')
    .custom((value, { req }) => {
      if (value !== req.body.password) {
        throw new Error('Password confirmation does not match password');
      }
      return true;
    }),
  nameValidation('firstName'),
  nameValidation('lastName'),
  body('acceptTerms')
    .isBoolean()
    .withMessage('Terms acceptance must be a boolean')
    .custom((value) => {
      if (!value) {
        throw new Error('You must accept the terms and conditions');
      }
      return true;
    }),
  body('marketingConsent')
    .optional()
    .isBoolean()
    .withMessage('Marketing consent must be a boolean'),
  body('referralCode')
    .optional()
    .isLength({ min: 6, max: 20 })
    .withMessage('Referral code must be between 6 and 20 characters')
    .matches(/^[a-zA-Z0-9]+$/)
    .withMessage('Referral code can only contain letters and numbers'),
];

// Login validation
export const validateLoginRequest: ValidationChain[] = [
  emailValidation,
  body('password')
    .notEmpty()
    .withMessage('Password is required')
    .isLength({ max: 128 })
    .withMessage('Password must not exceed 128 characters'),
  body('rememberMe')
    .optional()
    .isBoolean()
    .withMessage('Remember me must be a boolean'),
  body('deviceId')
    .optional()
    .isUUID(4)
    .withMessage('Device ID must be a valid UUID'),
  body('deviceName')
    .optional()
    .isLength({ max: 100 })
    .withMessage('Device name must not exceed 100 characters')
    .trim(),
];

// Refresh token validation
export const validateRefreshTokenRequest: ValidationChain[] = [
  body('refreshToken')
    .notEmpty()
    .withMessage('Refresh token is required')
    .isJWT()
    .withMessage('Refresh token must be a valid JWT'),
  body('deviceId')
    .optional()
    .isUUID(4)
    .withMessage('Device ID must be a valid UUID'),
];

// Email verification validation
export const validateVerifyEmailRequest: ValidationChain[] = [
  body('token')
    .notEmpty()
    .withMessage('Verification token is required')
    .isLength({ min: 32, max: 512 })
    .withMessage('Verification token must be between 32 and 512 characters')
    .matches(/^[a-zA-Z0-9+/=]+$/)
    .withMessage('Invalid verification token format'),
  body('email')
    .optional()
    .isEmail()
    .withMessage('Must be a valid email address')
    .normalizeEmail(),
];

// Forgot password validation
export const validateForgotPasswordRequest: ValidationChain[] = [
  emailValidation,
  body('callbackUrl')
    .optional()
    .isURL({ protocols: ['http', 'https'] })
    .withMessage('Callback URL must be a valid HTTP or HTTPS URL')
    .isLength({ max: 2048 })
    .withMessage('Callback URL must not exceed 2048 characters'),
];

// Reset password validation
export const validateResetPasswordRequest: ValidationChain[] = [
  body('token')
    .notEmpty()
    .withMessage('Reset token is required')
    .isLength({ min: 32, max: 512 })
    .withMessage('Reset token must be between 32 and 512 characters'),
  emailValidation,
  passwordValidation,
  body('confirmPassword')
    .custom((value, { req }) => {
      if (value !== req.body.newPassword) {
        throw new Error('Password confirmation does not match new password');
      }
      return true;
    }),
];

// Change password validation
export const validateChangePasswordRequest: ValidationChain[] = [
  body('currentPassword')
    .notEmpty()
    .withMessage('Current password is required')
    .isLength({ max: 128 })
    .withMessage('Current password must not exceed 128 characters'),
  body('newPassword')
    .isLength({ min: 8, max: 128 })
    .withMessage('New password must be between 8 and 128 characters')
    .matches(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]/)
    .withMessage('New password must contain at least one uppercase letter, one lowercase letter, one number, and one special character')
    .custom((value, { req }) => {
      if (value === req.body.currentPassword) {
        throw new Error('New password must be different from current password');
      }
      return true;
    }),
  body('confirmPassword')
    .custom((value, { req }) => {
      if (value !== req.body.newPassword) {
        throw new Error('Password confirmation does not match new password');
      }
      return true;
    }),
];

// Update profile validation
export const validateUpdateProfileRequest: ValidationChain[] = [
  body('firstName')
    .optional()
    .isLength({ min: 1, max: 50 })
    .withMessage('First name must be between 1 and 50 characters')
    .matches(/^[a-zA-Z\s'-]+$/)
    .withMessage('First name can only contain letters, spaces, apostrophes, and hyphens')
    .trim(),
  body('lastName')
    .optional()
    .isLength({ min: 1, max: 50 })
    .withMessage('Last name must be between 1 and 50 characters')
    .matches(/^[a-zA-Z\s'-]+$/)
    .withMessage('Last name can only contain letters, spaces, apostrophes, and hyphens')
    .trim(),
  body('phoneNumber')
    .optional()
    .matches(/^\+?[1-9]\d{1,14}$/)
    .withMessage('Phone number must be a valid international format (E.164)'),
  body('dateOfBirth')
    .optional()
    .isISO8601({ strict: true })
    .withMessage('Date of birth must be a valid ISO 8601 date')
    .custom((value) => {
      const birthDate = new Date(value);
      const today = new Date();
      const age = today.getFullYear() - birthDate.getFullYear();
      if (age < 13 || age > 120) {
        throw new Error('Age must be between 13 and 120 years');
      }
      return true;
    }),
  body('timezone')
    .optional()
    .isIn([
      'UTC', 'America/New_York', 'America/Chicago', 'America/Denver', 'America/Los_Angeles',
      'Europe/London', 'Europe/Paris', 'Europe/Berlin', 'Asia/Tokyo', 'Asia/Shanghai',
      'Asia/Kolkata', 'Australia/Sydney', 'Pacific/Auckland'
    ])
    .withMessage('Timezone must be a valid timezone identifier'),
  body('language')
    .optional()
    .isIn(['en', 'es', 'fr', 'de', 'it', 'pt', 'ru', 'ja', 'ko', 'zh'])
    .withMessage('Language must be a supported language code'),
];

// Parameter validation
export const validateUserIdParam: ValidationChain[] = [
  param('id')
    .isUUID(4)
    .withMessage('User ID must be a valid UUID'),
];

export const validateTokenParam: ValidationChain[] = [
  param('token')
    .notEmpty()
    .withMessage('Token is required')
    .isLength({ min: 32, max: 512 })
    .withMessage('Token must be between 32 and 512 characters'),
];

// Query parameter validation
export const validatePaginationQuery: ValidationChain[] = [
  query('page')
    .optional()
    .isInt({ min: 1, max: 10000 })
    .withMessage('Page must be a positive integer not greater than 10000')
    .toInt(),
  query('limit')
    .optional()
    .isInt({ min: 1, max: 100 })
    .withMessage('Limit must be a positive integer not greater than 100')
    .toInt(),
  query('sort')
    .optional()
    .isIn(['createdAt', 'updatedAt', 'email', 'username', 'lastName'])
    .withMessage('Sort field must be one of: createdAt, updatedAt, email, username, lastName'),
  query('order')
    .optional()
    .isIn(['asc', 'desc'])
    .withMessage('Order must be either asc or desc'),
];

export const validateSearchQuery: ValidationChain[] = [
  query('q')
    .optional()
    .isLength({ min: 1, max: 100 })
    .withMessage('Search query must be between 1 and 100 characters')
    .trim()
    .escape(),
  query('filter')
    .optional()
    .isIn(['active', 'suspended', 'pending', 'verified', 'unverified'])
    .withMessage('Filter must be one of: active, suspended, pending, verified, unverified'),
];

// Custom validation for API versioning
export const validateApiVersion: ValidationChain[] = [
  param('version')
    .isIn(['v1', 'v2'])
    .withMessage('API version must be v1 or v2'),
];

// Security header validation
export const validateSecurityHeaders = [
  body()
    .custom((value, { req }) => {
      // Check for required security headers
      if (!req.headers) {
        throw new Error('Request headers are missing');
      }
      
      const requiredHeaders = ['user-agent', 'accept'];
      for (const header of requiredHeaders) {
        if (!req.headers[header]) {
          throw new Error(`Missing required header: ${header}`);
        }
      }
      
      // Validate user agent format
      const userAgent = req.headers['user-agent'] as string;
      if (userAgent && userAgent.length > 512) {
        throw new Error('User-Agent header too long');
      }
      
      return true;
    }),
];

// Rate limiting bypass validation (for admin endpoints)
export const validateRateLimitBypass: ValidationChain[] = [
  body('bypassRateLimit')
    .optional()
    .isBoolean()
    .withMessage('Rate limit bypass must be a boolean'),
  body('adminKey')
    .if(body('bypassRateLimit').equals('true'))
    .notEmpty()
    .withMessage('Admin key is required when bypassing rate limit')
    .isLength({ min: 32, max: 64 })
    .withMessage('Admin key must be between 32 and 64 characters'),
];
