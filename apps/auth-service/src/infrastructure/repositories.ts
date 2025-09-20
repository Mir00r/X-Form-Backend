// Infrastructure Layer - External concerns and repository implementations
// Following Dependency Inversion Principle: Implements domain interfaces

import { Pool, PoolClient } from 'pg';
import bcrypt from 'bcryptjs';
import jwt from 'jsonwebtoken';
import crypto from 'crypto';
import {
  User,
  UserRole,
  AuthProvider,
  Email,
  Password,
  UserRepository,
  TokenRepository,
  EmailVerificationRepository,
  PasswordHashingService,
  TokenService,
  EmailService,
  UserNotFoundError,
} from '../domain/auth';

// PostgreSQL User Repository Implementation
export class PostgreSQLUserRepository implements UserRepository {
  constructor(private readonly pool: Pool) {}

  async findById(id: string): Promise<User | null> {
    const client = await this.pool.connect();
    try {
      const query = `
        SELECT id, email, username, first_name, last_name, password_hash,
               role, email_verified, account_locked, login_attempts,
               last_login_at, created_at, updated_at, provider, provider_id, metadata
        FROM users 
        WHERE id = $1 AND deleted_at IS NULL
      `;
      const result = await client.query(query, [id]);
      
      return result.rows.length > 0 ? this.mapRowToUser(result.rows[0]) : null;
    } finally {
      client.release();
    }
  }

  async findByEmail(email: Email): Promise<User | null> {
    const client = await this.pool.connect();
    try {
      const query = `
        SELECT id, email, username, first_name, last_name, password_hash,
               role, email_verified, account_locked, login_attempts,
               last_login_at, created_at, updated_at, provider, provider_id, metadata
        FROM users 
        WHERE email = $1 AND deleted_at IS NULL
      `;
      const result = await client.query(query, [email.getValue()]);
      
      return result.rows.length > 0 ? this.mapRowToUser(result.rows[0]) : null;
    } finally {
      client.release();
    }
  }

  async findByUsername(username: string): Promise<User | null> {
    const client = await this.pool.connect();
    try {
      const query = `
        SELECT id, email, username, first_name, last_name, password_hash,
               role, email_verified, account_locked, login_attempts,
               last_login_at, created_at, updated_at, provider, provider_id, metadata
        FROM users 
        WHERE username = $1 AND deleted_at IS NULL
      `;
      const result = await client.query(query, [username]);
      
      return result.rows.length > 0 ? this.mapRowToUser(result.rows[0]) : null;
    } finally {
      client.release();
    }
  }

  async findByProviderId(provider: AuthProvider, providerId: string): Promise<User | null> {
    const client = await this.pool.connect();
    try {
      const query = `
        SELECT id, email, username, first_name, last_name, password_hash,
               role, email_verified, account_locked, login_attempts,
               last_login_at, created_at, updated_at, provider, provider_id, metadata
        FROM users 
        WHERE provider = $1 AND provider_id = $2 AND deleted_at IS NULL
      `;
      const result = await client.query(query, [provider, providerId]);
      
      return result.rows.length > 0 ? this.mapRowToUser(result.rows[0]) : null;
    } finally {
      client.release();
    }
  }

  async save(user: User): Promise<User> {
    const client = await this.pool.connect();
    try {
      const query = `
        INSERT INTO users (
          id, email, username, first_name, last_name, password_hash,
          role, email_verified, account_locked, login_attempts,
          created_at, updated_at, provider, provider_id, metadata
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
        RETURNING *
      `;
      
      const values = [
        user.id,
        user.email.getValue(),
        user.username,
        user.firstName,
        user.lastName,
        user.password ? user.password.getValue() : null, // Password value from Password object
        user.role,
        user.emailVerified,
        user.accountLocked,
        user.loginAttempts,
        user.createdAt,
        user.updatedAt,
        user.provider,
        user.providerId,
        JSON.stringify(user.metadata || {})
      ];
      
      const result = await client.query(query, values);
      return this.mapRowToUser(result.rows[0]);
    } finally {
      client.release();
    }
  }

  async update(user: User): Promise<User> {
    const client = await this.pool.connect();
    try {
      const query = `
        UPDATE users SET
          email = $2, username = $3, first_name = $4, last_name = $5,
          password_hash = $6, role = $7, email_verified = $8, account_locked = $9,
          login_attempts = $10, last_login_at = $11, updated_at = $12,
          provider = $13, provider_id = $14, metadata = $15
        WHERE id = $1 AND deleted_at IS NULL
        RETURNING *
      `;
      
      const values = [
        user.id,
        user.email.getValue(),
        user.username,
        user.firstName,
        user.lastName,
        user.password ? user.password.getValue() : null,
        user.role,
        user.emailVerified,
        user.accountLocked,
        user.loginAttempts,
        user.lastLoginAt,
        user.updatedAt,
        user.provider,
        user.providerId,
        JSON.stringify(user.metadata || {})
      ];
      
      const result = await client.query(query, values);
      if (result.rows.length === 0) {
        throw new UserNotFoundError(user.id);
      }
      
      return this.mapRowToUser(result.rows[0]);
    } finally {
      client.release();
    }
  }

  async delete(id: string): Promise<void> {
    const client = await this.pool.connect();
    try {
      const query = `
        UPDATE users SET deleted_at = CURRENT_TIMESTAMP 
        WHERE id = $1 AND deleted_at IS NULL
      `;
      await client.query(query, [id]);
    } finally {
      client.release();
    }
  }

  async incrementLoginAttempts(id: string): Promise<void> {
    const client = await this.pool.connect();
    try {
      const query = `
        UPDATE users SET 
          login_attempts = login_attempts + 1,
          updated_at = CURRENT_TIMESTAMP
        WHERE id = $1 AND deleted_at IS NULL
      `;
      await client.query(query, [id]);
    } finally {
      client.release();
    }
  }

  async resetLoginAttempts(id: string): Promise<void> {
    const client = await this.pool.connect();
    try {
      const query = `
        UPDATE users SET 
          login_attempts = 0,
          updated_at = CURRENT_TIMESTAMP
        WHERE id = $1 AND deleted_at IS NULL
      `;
      await client.query(query, [id]);
    } finally {
      client.release();
    }
  }

  async lockAccount(id: string): Promise<void> {
    const client = await this.pool.connect();
    try {
      const query = `
        UPDATE users SET 
          account_locked = true,
          updated_at = CURRENT_TIMESTAMP
        WHERE id = $1 AND deleted_at IS NULL
      `;
      await client.query(query, [id]);
    } finally {
      client.release();
    }
  }

  async unlockAccount(id: string): Promise<void> {
    const client = await this.pool.connect();
    try {
      const query = `
        UPDATE users SET 
          account_locked = false,
          login_attempts = 0,
          updated_at = CURRENT_TIMESTAMP
        WHERE id = $1 AND deleted_at IS NULL
      `;
      await client.query(query, [id]);
    } finally {
      client.release();
    }
  }

  private mapRowToUser(row: any): User {
    return new User(
      row.id,
      new Email(row.email),
      row.username,
      row.first_name,
      row.last_name,
      row.password_hash ? new Password(row.password_hash, false) : undefined, // false = already hashed
      row.role as UserRole,
      row.email_verified,
      row.account_locked,
      row.login_attempts,
      row.last_login_at,
      row.created_at,
      row.updated_at,
      row.deleted_at,
      row.provider as AuthProvider,
      row.provider_id,
      JSON.parse(row.metadata || '{}')
    );
  }
}

// PostgreSQL Token Repository Implementation
export class PostgreSQLTokenRepository implements TokenRepository {
  constructor(private readonly pool: Pool) {}

  async saveRefreshToken(userId: string, token: string, expiresAt: Date): Promise<void> {
    const client = await this.pool.connect();
    try {
      const query = `
        INSERT INTO refresh_tokens (user_id, token, expires_at, created_at)
        VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
        ON CONFLICT (token) DO UPDATE SET
          expires_at = EXCLUDED.expires_at,
          created_at = EXCLUDED.created_at
      `;
      await client.query(query, [userId, token, expiresAt]);
    } finally {
      client.release();
    }
  }

  async findRefreshToken(token: string): Promise<{ userId: string; expiresAt: Date } | null> {
    const client = await this.pool.connect();
    try {
      const query = `
        SELECT user_id, expires_at 
        FROM refresh_tokens 
        WHERE token = $1 AND revoked = false
      `;
      const result = await client.query(query, [token]);
      
      return result.rows.length > 0 
        ? { userId: result.rows[0].user_id, expiresAt: result.rows[0].expires_at }
        : null;
    } finally {
      client.release();
    }
  }

  async revokeRefreshToken(token: string): Promise<void> {
    const client = await this.pool.connect();
    try {
      const query = `
        UPDATE refresh_tokens SET 
          revoked = true,
          revoked_at = CURRENT_TIMESTAMP
        WHERE token = $1
      `;
      await client.query(query, [token]);
    } finally {
      client.release();
    }
  }

  async revokeAllUserTokens(userId: string): Promise<void> {
    const client = await this.pool.connect();
    try {
      const query = `
        UPDATE refresh_tokens SET 
          revoked = true,
          revoked_at = CURRENT_TIMESTAMP
        WHERE user_id = $1 AND revoked = false
      `;
      await client.query(query, [userId]);
    } finally {
      client.release();
    }
  }

  async cleanExpiredTokens(): Promise<void> {
    const client = await this.pool.connect();
    try {
      const query = `
        DELETE FROM refresh_tokens 
        WHERE expires_at < CURRENT_TIMESTAMP OR revoked = true
      `;
      await client.query(query);
    } finally {
      client.release();
    }
  }
}

// PostgreSQL Email Verification Repository Implementation
export class PostgreSQLEmailVerificationRepository implements EmailVerificationRepository {
  constructor(private readonly pool: Pool) {}

  async saveVerificationToken(userId: string, token: string, expiresAt: Date): Promise<void> {
    const client = await this.pool.connect();
    try {
      const query = `
        INSERT INTO email_verification_tokens (user_id, token, expires_at, created_at)
        VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
        ON CONFLICT (user_id) DO UPDATE SET
          token = EXCLUDED.token,
          expires_at = EXCLUDED.expires_at,
          created_at = EXCLUDED.created_at
      `;
      await client.query(query, [userId, token, expiresAt]);
    } finally {
      client.release();
    }
  }

  async findVerificationToken(token: string): Promise<{ userId: string; expiresAt: Date } | null> {
    const client = await this.pool.connect();
    try {
      const query = `
        SELECT user_id, expires_at 
        FROM email_verification_tokens 
        WHERE token = $1
      `;
      const result = await client.query(query, [token]);
      
      return result.rows.length > 0 
        ? { userId: result.rows[0].user_id, expiresAt: result.rows[0].expires_at }
        : null;
    } finally {
      client.release();
    }
  }

  async markEmailAsVerified(userId: string): Promise<void> {
    const client = await this.pool.connect();
    try {
      const query = `
        UPDATE users SET 
          email_verified = true,
          updated_at = CURRENT_TIMESTAMP
        WHERE id = $1
      `;
      await client.query(query, [userId]);
    } finally {
      client.release();
    }
  }

  async deleteVerificationToken(token: string): Promise<void> {
    const client = await this.pool.connect();
    try {
      const query = `DELETE FROM email_verification_tokens WHERE token = $1`;
      await client.query(query, [token]);
    } finally {
      client.release();
    }
  }
}

// BCrypt Password Hashing Service Implementation
export class BCryptPasswordHashingService implements PasswordHashingService {
  private readonly saltRounds = 12;

  async hash(password: Password): Promise<string> {
    return await bcrypt.hash(password.getValue(), this.saltRounds);
  }

  async compare(password: Password, hash: string): Promise<boolean> {
    return await bcrypt.compare(password.getValue(), hash);
  }
}

// JWT Token Service Implementation
export class JWTTokenService implements TokenService {
  constructor(
    private readonly accessTokenSecret: string,
    private readonly refreshTokenSecret: string,
    private readonly accessTokenExpiry: string = '15m',
    private readonly refreshTokenExpiry: string = '7d'
  ) {}

  generateAccessToken(payload: Record<string, any>): string {
    return jwt.sign(payload, this.accessTokenSecret, {
      expiresIn: this.accessTokenExpiry,
      issuer: 'auth-service',
      audience: 'xform-api'
    } as jwt.SignOptions);
  }

  generateRefreshToken(payload: Record<string, any>): string {
    return jwt.sign(payload, this.refreshTokenSecret, {
      expiresIn: this.refreshTokenExpiry,
      issuer: 'auth-service',
      audience: 'xform-api'
    } as jwt.SignOptions);
  }

  verifyAccessToken(token: string): any {
    return jwt.verify(token, this.accessTokenSecret, {
      issuer: 'xform-auth-service',
      audience: 'xform-api',
    });
  }

  verifyRefreshToken(token: string): any {
    return jwt.verify(token, this.refreshTokenSecret, {
      issuer: 'xform-auth-service',
      audience: 'xform-api',
    });
  }

  generateVerificationToken(): string {
    return crypto.randomBytes(32).toString('hex');
  }

  generatePasswordResetToken(): string {
    return crypto.randomBytes(32).toString('hex');
  }
}

// Email Service Implementation (Mock for now - would use SendGrid, SES, etc.)
export class MockEmailService implements EmailService {
  async sendVerificationEmail(email: Email, token: string): Promise<void> {
    console.log(`Sending verification email to ${email.getValue()} with token: ${token}`);
    // In production, integrate with email service provider
  }

  async sendPasswordResetEmail(email: Email, token: string): Promise<void> {
    console.log(`Sending password reset email to ${email.getValue()} with token: ${token}`);
    // In production, integrate with email service provider
  }

  async sendWelcomeEmail(email: Email, firstName: string): Promise<void> {
    console.log(`Sending welcome email to ${email.getValue()} for ${firstName}`);
    // In production, integrate with email service provider
  }
}
