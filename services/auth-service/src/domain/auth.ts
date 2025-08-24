// Domain Layer - Core business entities and rules for Authentication
// Following Single Responsibility Principle: This layer only contains business logic

export class User {
  constructor(
    public readonly id: string,
    public readonly email: Email,
    public readonly username: string,
    public readonly firstName: string,
    public readonly lastName: string,
    public readonly password?: Password, // Optional for OAuth users
    public readonly role: UserRole = UserRole.USER,
    public readonly emailVerified: boolean = false,
    public readonly accountLocked: boolean = false,
    public readonly loginAttempts: number = 0,
    public readonly lastLoginAt?: Date,
    public readonly createdAt: Date = new Date(),
    public readonly updatedAt: Date = new Date(),
    public readonly deletedAt?: Date,
    public readonly provider: string = 'local',
    public readonly providerId?: string,
    public readonly metadata: Record<string, any> = {}
  ) {}
}

export enum UserRole {
  USER = 'user',
  ADMIN = 'admin',
  MODERATOR = 'moderator'
}

export enum AuthProvider {
  LOCAL = 'local',
  GOOGLE = 'google',
  GITHUB = 'github'
}

// Value Objects following Domain-Driven Design
export class Email {
  private readonly value: string;

  constructor(email: string) {
    if (!this.isValid(email)) {
      throw new InvalidEmailError('Invalid email format');
    }
    this.value = email.toLowerCase().trim();
  }

  getValue(): string {
    return this.value;
  }

  private isValid(email: string): boolean {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email) && email.length <= 255;
  }

  equals(other: Email): boolean {
    return this.value === other.value;
  }
}

export class Password {
  private readonly value: string;

  constructor(password: string, isHashed: boolean = false) {
    if (!isHashed) {
      this.validatePassword(password);
    }
    this.value = password;
  }

  getValue(): string {
    return this.value;
  }

  private validatePassword(password: string): void {
    if (password.length < 8) {
      throw new WeakPasswordError('Password must be at least 8 characters long');
    }
    if (!/(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/.test(password)) {
      throw new WeakPasswordError('Password must contain at least one lowercase letter, one uppercase letter, and one number');
    }
  }
}

// Domain Events following Event-Driven Architecture
export abstract class DomainEvent {
  public readonly occurredOn: Date;
  public readonly eventId: string;

  constructor() {
    this.occurredOn = new Date();
    this.eventId = Math.random().toString(36).substr(2, 9);
  }
}

export class UserRegisteredEvent extends DomainEvent {
  constructor(
    public readonly userId: string,
    public readonly email: string,
    public readonly provider: AuthProvider
  ) {
    super();
  }
}

export class UserLoginEvent extends DomainEvent {
  constructor(
    public readonly userId: string,
    public readonly ipAddress: string,
    public readonly userAgent: string
  ) {
    super();
  }
}

export class UserAccountLockedEvent extends DomainEvent {
  constructor(
    public readonly userId: string,
    public readonly reason: string
  ) {
    super();
  }
}

// Repository Interfaces following Dependency Inversion Principle
export interface UserRepository {
  findById(id: string): Promise<User | null>;
  findByEmail(email: Email): Promise<User | null>;
  findByUsername(username: string): Promise<User | null>;
  findByProviderId(provider: AuthProvider, providerId: string): Promise<User | null>;
  save(user: User): Promise<User>;
  update(user: User): Promise<User>;
  delete(id: string): Promise<void>;
  incrementLoginAttempts(id: string): Promise<void>;
  resetLoginAttempts(id: string): Promise<void>;
  lockAccount(id: string): Promise<void>;
  unlockAccount(id: string): Promise<void>;
}

export interface TokenRepository {
  saveRefreshToken(userId: string, token: string, expiresAt: Date): Promise<void>;
  findRefreshToken(token: string): Promise<{ userId: string; expiresAt: Date } | null>;
  revokeRefreshToken(token: string): Promise<void>;
  revokeAllUserTokens(userId: string): Promise<void>;
  cleanExpiredTokens(): Promise<void>;
}

export interface EmailVerificationRepository {
  saveVerificationToken(userId: string, token: string, expiresAt: Date): Promise<void>;
  findVerificationToken(token: string): Promise<{ userId: string; expiresAt: Date } | null>;
  markEmailAsVerified(userId: string): Promise<void>;
  deleteVerificationToken(token: string): Promise<void>;
}

// Domain Services following Domain-Driven Design
export interface PasswordHashingService {
  hash(password: Password): Promise<string>;
  compare(password: Password, hash: string): Promise<boolean>;
}

export interface TokenService {
  generateAccessToken(payload: any): string;
  generateRefreshToken(payload: any): string;
  verifyAccessToken(token: string): any;
  verifyRefreshToken(token: string): any;
  generateVerificationToken(): string;
  generatePasswordResetToken(): string;
}

export interface EmailService {
  sendVerificationEmail(email: Email, token: string): Promise<void>;
  sendPasswordResetEmail(email: Email, token: string): Promise<void>;
  sendWelcomeEmail(email: Email, firstName: string): Promise<void>;
}

// Domain Exceptions following Single Responsibility Principle
export class DomainError extends Error {
  constructor(message: string, public readonly code: string) {
    super(message);
    this.name = 'DomainError';
  }
}

export class UserNotFoundError extends DomainError {
  constructor(identifier: string) {
    super(`User not found: ${identifier}`, 'USER_NOT_FOUND');
  }
}

export class UserAlreadyExistsError extends DomainError {
  constructor(email: string) {
    super(`User already exists with email: ${email}`, 'USER_ALREADY_EXISTS');
  }
}

export class InvalidCredentialsError extends DomainError {
  constructor() {
    super('Invalid email or password', 'INVALID_CREDENTIALS');
  }
}

export class AccountLockedError extends DomainError {
  constructor() {
    super('Account is locked due to too many failed login attempts', 'ACCOUNT_LOCKED');
  }
}

export class EmailNotVerifiedError extends DomainError {
  constructor() {
    super('Email address is not verified', 'EMAIL_NOT_VERIFIED');
  }
}

export class InvalidEmailError extends DomainError {
  constructor(message: string = 'Invalid email format') {
    super(message, 'INVALID_EMAIL');
  }
}

export class WeakPasswordError extends DomainError {
  constructor(message: string) {
    super(message, 'WEAK_PASSWORD');
  }
}

export class TokenExpiredError extends DomainError {
  constructor() {
    super('Token has expired', 'TOKEN_EXPIRED');
  }
}

export class InvalidTokenError extends DomainError {
  constructor() {
    super('Invalid token', 'INVALID_TOKEN');
  }
}
