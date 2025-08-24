const Joi = require('joi');

// User model schema validation
const userSchema = {
  signup: Joi.object({
    email: Joi.string().email().required().lowercase().trim(),
    password: Joi.string()
      .min(8)
      .max(128)
      .pattern(new RegExp('^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&])[A-Za-z\\d@$!%*?&]'))
      .required()
      .messages({
        'string.pattern.base': 'Password must contain at least one lowercase letter, one uppercase letter, one number, and one special character'
      }),
    firstName: Joi.string().min(1).max(50).required().trim(),
    lastName: Joi.string().min(1).max(50).required().trim(),
    timezone: Joi.string().optional(),
    language: Joi.string().length(2).optional()
  }),

  login: Joi.object({
    email: Joi.string().email().required().lowercase().trim(),
    password: Joi.string().required(),
    rememberMe: Joi.boolean().optional()
  }),

  refreshToken: Joi.object({
    refreshToken: Joi.string().required()
  }),

  logout: Joi.object({
    refreshToken: Joi.string().required()
  }),

  updateProfile: Joi.object({
    firstName: Joi.string().min(1).max(50).optional().trim(),
    lastName: Joi.string().min(1).max(50).optional().trim(),
    timezone: Joi.string().optional(),
    language: Joi.string().length(2).optional(),
    avatarUrl: Joi.string().uri().optional()
  }),

  changePassword: Joi.object({
    currentPassword: Joi.string().required(),
    newPassword: Joi.string()
      .min(8)
      .max(128)
      .pattern(new RegExp('^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&])[A-Za-z\\d@$!%*?&]'))
      .required()
      .messages({
        'string.pattern.base': 'Password must contain at least one lowercase letter, one uppercase letter, one number, and one special character'
      })
  }),

  requestPasswordReset: Joi.object({
    email: Joi.string().email().required().lowercase().trim()
  }),

  resetPassword: Joi.object({
    token: Joi.string().required(),
    password: Joi.string()
      .min(8)
      .max(128)
      .pattern(new RegExp('^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&])[A-Za-z\\d@$!%*?&]'))
      .required()
      .messages({
        'string.pattern.base': 'Password must contain at least one lowercase letter, one uppercase letter, one number, and one special character'
      })
  }),

  verifyEmail: Joi.object({
    token: Joi.string().required()
  })
};

// Response schemas
const responseSchema = {
  user: {
    id: Joi.string().uuid().required(),
    email: Joi.string().email().required(),
    firstName: Joi.string().required(),
    lastName: Joi.string().required(),
    avatarUrl: Joi.string().uri().allow(null),
    emailVerified: Joi.boolean().required(),
    provider: Joi.string().required(),
    timezone: Joi.string().required(),
    language: Joi.string().required(),
    twoFactorEnabled: Joi.boolean().required(),
    createdAt: Joi.date().required(),
    updatedAt: Joi.date().required(),
    lastLogin: Joi.date().allow(null)
  },

  authResponse: {
    user: Joi.object().keys(responseSchema.user),
    accessToken: Joi.string().required(),
    refreshToken: Joi.string().required(),
    expiresIn: Joi.number().required()
  }
};

// User model class
class User {
  constructor(userData) {
    this.id = userData.id;
    this.email = userData.email;
    this.firstName = userData.first_name;
    this.lastName = userData.last_name;
    this.avatarUrl = userData.avatar_url;
    this.googleId = userData.google_id;
    this.provider = userData.provider;
    this.emailVerified = userData.email_verified;
    this.isActive = userData.is_active;
    this.timezone = userData.timezone;
    this.language = userData.language;
    this.twoFactorEnabled = userData.two_factor_enabled;
    this.createdAt = userData.created_at;
    this.updatedAt = userData.updated_at;
    this.lastLogin = userData.last_login;
  }

  // Return safe user object (no sensitive data)
  toSafeObject() {
    return {
      id: this.id,
      email: this.email,
      firstName: this.firstName,
      lastName: this.lastName,
      avatarUrl: this.avatarUrl,
      provider: this.provider,
      emailVerified: this.emailVerified,
      timezone: this.timezone,
      language: this.language,
      twoFactorEnabled: this.twoFactorEnabled,
      createdAt: this.createdAt,
      updatedAt: this.updatedAt,
      lastLogin: this.lastLogin
    };
  }

  // Return minimal user object for JWT payload
  toJWTPayload() {
    return {
      id: this.id,
      email: this.email,
      firstName: this.firstName,
      lastName: this.lastName,
      emailVerified: this.emailVerified
    };
  }

  // Get full name
  getFullName() {
    return `${this.firstName} ${this.lastName}`.trim();
  }

  // Check if user can login
  canLogin() {
    return this.isActive && !this.isLocked;
  }
}

module.exports = {
  userSchema,
  responseSchema,
  User
};
