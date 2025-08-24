const bcrypt = require('bcryptjs');
const jwt = require('jsonwebtoken');
const crypto = require('crypto');
const { v4: uuidv4 } = require('uuid');
const { getPool } = require('../config/database');
const { User } = require('../models/User');

class AuthService {
  constructor() {
    this.pool = getPool();
    this.JWT_SECRET = process.env.JWT_SECRET;
    this.JWT_REFRESH_SECRET = process.env.JWT_REFRESH_SECRET || process.env.JWT_SECRET + '_refresh';
    this.ACCESS_TOKEN_EXPIRY = process.env.ACCESS_TOKEN_EXPIRY || '15m';
    this.REFRESH_TOKEN_EXPIRY = process.env.REFRESH_TOKEN_EXPIRY || '7d';
    this.MAX_LOGIN_ATTEMPTS = parseInt(process.env.MAX_LOGIN_ATTEMPTS) || 5;
    this.LOCK_TIME = parseInt(process.env.LOCK_TIME) || 30 * 60 * 1000; // 30 minutes
  }

  // Hash password with bcrypt
  async hashPassword(password) {
    const saltRounds = 12;
    return await bcrypt.hash(password, saltRounds);
  }

  // Compare password with hash
  async comparePassword(password, hash) {
    return await bcrypt.compare(password, hash);
  }

  // Generate JWT access token
  generateAccessToken(payload) {
    return jwt.sign(payload, this.JWT_SECRET, {
      expiresIn: this.ACCESS_TOKEN_EXPIRY,
      issuer: 'xform-auth-service',
      audience: 'xform-api'
    });
  }

  // Generate JWT refresh token
  generateRefreshToken(payload) {
    return jwt.sign(payload, this.JWT_REFRESH_SECRET, {
      expiresIn: this.REFRESH_TOKEN_EXPIRY,
      issuer: 'xform-auth-service',
      audience: 'xform-api'
    });
  }

  // Verify JWT token
  verifyAccessToken(token) {
    try {
      return jwt.verify(token, this.JWT_SECRET, {
        issuer: 'xform-auth-service',
        audience: 'xform-api'
      });
    } catch (error) {
      throw new Error('Invalid or expired access token');
    }
  }

  // Verify refresh token
  verifyRefreshToken(token) {
    try {
      return jwt.verify(token, this.JWT_REFRESH_SECRET, {
        issuer: 'xform-auth-service',
        audience: 'xform-api'
      });
    } catch (error) {
      throw new Error('Invalid or expired refresh token');
    }
  }

  // Generate secure random token
  generateSecureToken() {
    return crypto.randomBytes(32).toString('hex');
  }

  // Register new user
  async register(userData, ipAddress, userAgent) {
    const client = await this.pool.connect();
    
    try {
      await client.query('BEGIN');

      // Check if user already exists
      const existingUser = await client.query(
        'SELECT id FROM users WHERE email = $1',
        [userData.email]
      );

      if (existingUser.rows.length > 0) {
        throw new Error('User already exists with this email');
      }

      // Hash password
      const passwordHash = await this.hashPassword(userData.password);

      // Generate email verification token
      const emailVerificationToken = this.generateSecureToken();
      const emailVerificationExpires = new Date(Date.now() + 24 * 60 * 60 * 1000); // 24 hours

      // Insert user
      const userResult = await client.query(`
        INSERT INTO users (
          email, password_hash, first_name, last_name, 
          email_verification_token, email_verification_expires,
          timezone, language, provider
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING *
      `, [
        userData.email,
        passwordHash,
        userData.firstName,
        userData.lastName,
        emailVerificationToken,
        emailVerificationExpires,
        userData.timezone || 'UTC',
        userData.language || 'en',
        'local'
      ]);

      const user = new User(userResult.rows[0]);

      // Log registration event
      await this.logAuthEvent(user.id, 'register', ipAddress, userAgent, {
        provider: 'local'
      }, true);

      await client.query('COMMIT');

      return {
        user: user.toSafeObject(),
        emailVerificationToken // Return for testing purposes
      };

    } catch (error) {
      await client.query('ROLLBACK');
      throw error;
    } finally {
      client.release();
    }
  }

  // Login user
  async login(email, password, rememberMe = false, ipAddress, userAgent) {
    const client = await this.pool.connect();
    
    try {
      await client.query('BEGIN');

      // Get user with password hash
      const userResult = await client.query(`
        SELECT * FROM users WHERE email = $1
      `, [email]);

      if (userResult.rows.length === 0) {
        await this.logAuthEvent(null, 'login_failed', ipAddress, userAgent, {
          email,
          reason: 'user_not_found'
        }, false);
        throw new Error('Invalid email or password');
      }

      const userData = userResult.rows[0];
      const user = new User(userData);

      // Check if account is locked
      if (userData.is_locked && userData.lock_until && userData.lock_until > new Date()) {
        await this.logAuthEvent(user.id, 'login_failed', ipAddress, userAgent, {
          reason: 'account_locked'
        }, false);
        throw new Error('Account is locked. Please try again later');
      }

      // Check if account is active
      if (!userData.is_active) {
        await this.logAuthEvent(user.id, 'login_failed', ipAddress, userAgent, {
          reason: 'account_inactive'
        }, false);
        throw new Error('Account is not active');
      }

      // Verify password
      const isValidPassword = await this.comparePassword(password, userData.password_hash);

      if (!isValidPassword) {
        // Increment failed login attempts
        const failedAttempts = (userData.failed_login_attempts || 0) + 1;
        const lockUntil = failedAttempts >= this.MAX_LOGIN_ATTEMPTS 
          ? new Date(Date.now() + this.LOCK_TIME) 
          : null;

        await client.query(`
          UPDATE users 
          SET failed_login_attempts = $1, 
              last_failed_login = CURRENT_TIMESTAMP,
              is_locked = $2,
              lock_until = $3
          WHERE id = $4
        `, [failedAttempts, lockUntil !== null, lockUntil, user.id]);

        await this.logAuthEvent(user.id, 'login_failed', ipAddress, userAgent, {
          reason: 'invalid_password',
          attempts: failedAttempts
        }, false);

        throw new Error('Invalid email or password');
      }

      // Clear failed login attempts and update last login
      await client.query(`
        UPDATE users 
        SET failed_login_attempts = 0, 
            is_locked = FALSE,
            lock_until = NULL,
            last_login = CURRENT_TIMESTAMP
        WHERE id = $1
      `, [user.id]);

      // Generate tokens
      const jwtPayload = user.toJWTPayload();
      const accessToken = this.generateAccessToken(jwtPayload);
      const refreshToken = this.generateRefreshToken({ id: user.id });

      // Store refresh token in database
      const refreshTokenExpiry = rememberMe 
        ? new Date(Date.now() + 30 * 24 * 60 * 60 * 1000) // 30 days
        : new Date(Date.now() + 7 * 24 * 60 * 60 * 1000);  // 7 days

      const refreshTokenHash = crypto.createHash('sha256').update(refreshToken).digest('hex');

      await client.query(`
        INSERT INTO refresh_tokens (user_id, token_hash, expires_at, ip_address, device_info)
        VALUES ($1, $2, $3, $4, $5)
      `, [
        user.id,
        refreshTokenHash,
        refreshTokenExpiry,
        ipAddress,
        JSON.stringify({ userAgent, rememberMe })
      ]);

      // Create user session
      const sessionToken = uuidv4();
      await client.query(`
        INSERT INTO user_sessions (user_id, session_token, ip_address, user_agent, expires_at)
        VALUES ($1, $2, $3, $4, $5)
      `, [
        user.id,
        sessionToken,
        ipAddress,
        userAgent,
        refreshTokenExpiry
      ]);

      // Log successful login
      await this.logAuthEvent(user.id, 'login', ipAddress, userAgent, {
        provider: 'local',
        rememberMe
      }, true);

      await client.query('COMMIT');

      // Get updated user data
      const updatedUserResult = await client.query('SELECT * FROM users WHERE id = $1', [user.id]);
      const updatedUser = new User(updatedUserResult.rows[0]);

      return {
        user: updatedUser.toSafeObject(),
        accessToken,
        refreshToken,
        expiresIn: this.getTokenExpiryTime(this.ACCESS_TOKEN_EXPIRY)
      };

    } catch (error) {
      await client.query('ROLLBACK');
      throw error;
    } finally {
      client.release();
    }
  }

  // Refresh access token
  async refreshAccessToken(refreshToken, ipAddress, userAgent) {
    try {
      // Verify refresh token
      const decoded = this.verifyRefreshToken(refreshToken);
      const refreshTokenHash = crypto.createHash('sha256').update(refreshToken).digest('hex');

      // Check if refresh token exists and is not revoked
      const tokenResult = await this.pool.query(`
        SELECT rt.*, u.* FROM refresh_tokens rt
        JOIN users u ON rt.user_id = u.id
        WHERE rt.token_hash = $1 
        AND rt.expires_at > CURRENT_TIMESTAMP 
        AND rt.is_revoked = FALSE
      `, [refreshTokenHash]);

      if (tokenResult.rows.length === 0) {
        throw new Error('Invalid or expired refresh token');
      }

      const userData = tokenResult.rows[0];
      const user = new User(userData);

      // Check if user is still active
      if (!user.canLogin()) {
        throw new Error('User account is not active');
      }

      // Update last used timestamp
      await this.pool.query(`
        UPDATE refresh_tokens 
        SET last_used = CURRENT_TIMESTAMP 
        WHERE token_hash = $1
      `, [refreshTokenHash]);

      // Generate new access token
      const jwtPayload = user.toJWTPayload();
      const accessToken = this.generateAccessToken(jwtPayload);

      // Log token refresh
      await this.logAuthEvent(user.id, 'token_refresh', ipAddress, userAgent, {}, true);

      return {
        user: user.toSafeObject(),
        accessToken,
        expiresIn: this.getTokenExpiryTime(this.ACCESS_TOKEN_EXPIRY)
      };

    } catch (error) {
      throw error;
    }
  }

  // Logout user
  async logout(refreshToken, ipAddress, userAgent) {
    try {
      const refreshTokenHash = crypto.createHash('sha256').update(refreshToken).digest('hex');

      // Get user info before revoking token
      const tokenResult = await this.pool.query(`
        SELECT rt.user_id FROM refresh_tokens rt
        WHERE rt.token_hash = $1 AND rt.is_revoked = FALSE
      `, [refreshTokenHash]);

      if (tokenResult.rows.length > 0) {
        const userId = tokenResult.rows[0].user_id;

        // Revoke refresh token
        await this.pool.query(`
          UPDATE refresh_tokens 
          SET is_revoked = TRUE 
          WHERE token_hash = $1
        `, [refreshTokenHash]);

        // Deactivate session
        await this.pool.query(`
          UPDATE user_sessions 
          SET is_active = FALSE 
          WHERE user_id = $1 AND ip_address = $2
        `, [userId, ipAddress]);

        // Log logout
        await this.logAuthEvent(userId, 'logout', ipAddress, userAgent, {}, true);
      }

      return { message: 'Logged out successfully' };

    } catch (error) {
      throw error;
    }
  }

  // Get user by ID
  async getUserById(userId) {
    try {
      const result = await this.pool.query('SELECT * FROM users WHERE id = $1', [userId]);
      
      if (result.rows.length === 0) {
        throw new Error('User not found');
      }

      const user = new User(result.rows[0]);
      return user.toSafeObject();
    } catch (error) {
      throw error;
    }
  }

  // Log authentication events
  async logAuthEvent(userId, action, ipAddress, userAgent, details = {}, success = true) {
    try {
      await this.pool.query(`
        INSERT INTO auth_audit_log (user_id, action, ip_address, user_agent, details, success)
        VALUES ($1, $2, $3, $4, $5, $6)
      `, [userId, action, ipAddress, userAgent, JSON.stringify(details), success]);
    } catch (error) {
      console.error('Failed to log auth event:', error);
    }
  }

  // Get token expiry time in seconds
  getTokenExpiryTime(expiry) {
    if (expiry.endsWith('m')) {
      return parseInt(expiry) * 60;
    } else if (expiry.endsWith('h')) {
      return parseInt(expiry) * 60 * 60;
    } else if (expiry.endsWith('d')) {
      return parseInt(expiry) * 24 * 60 * 60;
    }
    return parseInt(expiry);
  }

  // Cleanup expired tokens (should be run periodically)
  async cleanupExpiredTokens() {
    try {
      await this.pool.query('SELECT cleanup_expired_tokens()');
    } catch (error) {
      console.error('Failed to cleanup expired tokens:', error);
    }
  }
}

module.exports = AuthService;
