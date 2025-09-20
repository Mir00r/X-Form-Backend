// Application Layer - Use cases and business logic orchestration
// Following Single Responsibility Principle and Dependency Inversion Principle

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
  UserAlreadyExistsError,
  InvalidCredentialsError,
  AccountLockedError,
  EmailNotVerifiedError,
  TokenExpiredError,
  InvalidTokenError,
  UserRegisteredEvent,
  UserLoginEvent,
  UserAccountLockedEvent,
  DomainEvent
} from '../domain/auth';

// DTOs for Application Layer (Data Transfer Objects)
export interface RegisterUserRequest {
  email: string;
  username: string;
  password: string;
  firstName: string;
  lastName: string;
  provider?: AuthProvider;
  providerId?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
  ipAddress: string;
  userAgent: string;
}

export interface AuthResponse {
  accessToken: string;
  refreshToken: string;
  user: UserProfile;
}

export interface UserProfile {
  id: string;
  email: string;
  username: string;
  firstName: string;
  lastName: string;
  role: UserRole;
  emailVerified: boolean;
  createdAt: Date;
}

export interface RefreshTokenRequest {
  refreshToken: string;
}

export interface VerifyEmailRequest {
  token: string;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface ResetPasswordRequest {
  token: string;
  newPassword: string;
}

// Event Publisher Interface following Open/Closed Principle
export interface EventPublisher {
  publish(event: DomainEvent): Promise<void>;
}

// Application Service following Single Responsibility Principle
export class AuthApplicationService {
  private readonly MAX_LOGIN_ATTEMPTS = 5;
  
  constructor(
    private readonly userRepository: UserRepository,
    private readonly tokenRepository: TokenRepository,
    private readonly emailVerificationRepository: EmailVerificationRepository,
    private readonly passwordHashingService: PasswordHashingService,
    private readonly tokenService: TokenService,
    private readonly emailService: EmailService,
    private readonly eventPublisher: EventPublisher
  ) {}

  // Use Case: Register User
  async registerUser(request: RegisterUserRequest): Promise<AuthResponse> {
    // Validate email uniqueness (Business Rule)
    const email = new Email(request.email);
    const existingUser = await this.userRepository.findByEmail(email);
    if (existingUser) {
      throw new UserAlreadyExistsError(request.email);
    }

    // Validate username uniqueness (Business Rule)
    const existingUsername = await this.userRepository.findByUsername(request.username);
    if (existingUsername) {
      throw new UserAlreadyExistsError(request.username);
    }

    // Hash password
    const password = new Password(request.password);
    const hashedPassword = await this.passwordHashingService.hash(password);

    // Create user entity
    const userId = this.generateUserId();
    const user = new User(
      userId,
      email,
      request.username,
      request.firstName,
      request.lastName,
      new Password(hashedPassword, false), // false = already hashed
      UserRole.USER,
      false, // emailVerified
      false, // accountLocked
      0, // loginAttempts
      undefined, // lastLoginAt
      new Date(), // createdAt
      new Date(), // updatedAt
      undefined, // deletedAt
      'local', // provider
      undefined, // providerId
      {} // metadata
    );

    // Save user
    const savedUser = await this.userRepository.save(user);

    // Generate email verification token
    const verificationToken = this.tokenService.generateVerificationToken();
    const expiresAt = new Date(Date.now() + 24 * 60 * 60 * 1000); // 24 hours
    await this.emailVerificationRepository.saveVerificationToken(
      savedUser.id,
      verificationToken,
      expiresAt
    );

    // Send verification email
    await this.emailService.sendVerificationEmail(email, verificationToken);

    // Generate tokens
    const tokens = this.generateTokens(savedUser);
    await this.saveRefreshToken(savedUser.id, tokens.refreshToken);

    // Publish domain event
    await this.eventPublisher.publish(
      new UserRegisteredEvent(savedUser.id, savedUser.email.getValue(), savedUser.provider as AuthProvider)
    );

    return {
      ...tokens,
      user: this.mapToUserProfile(savedUser),
    };
  }

  // Use Case: Login User
  async loginUser(request: LoginRequest): Promise<AuthResponse> {
    const email = new Email(request.email);
    const user = await this.userRepository.findByEmail(email);

    if (!user) {
      throw new InvalidCredentialsError();
    }

    // Check if account is locked
    if (user.accountLocked) {
      throw new AccountLockedError();
    }

    // Verify password
    const password = new Password(request.password, true); // Password is plain text from request
    
    if (!user.password) {
      throw new InvalidCredentialsError();
    }
    
    const isValidPassword = await this.passwordHashingService.compare(
      new Password(request.password),
      user.password.getValue() // Get string value from Password value object
    );

    if (!isValidPassword) {
      await this.handleFailedLogin(user.id);
      throw new InvalidCredentialsError();
    }

    // Check email verification for local accounts
    if (user.provider === AuthProvider.LOCAL && !user.emailVerified) {
      throw new EmailNotVerifiedError();
    }

    // Reset login attempts on successful login
    await this.userRepository.resetLoginAttempts(user.id);

    // Generate tokens
    const tokens = this.generateTokens(user);
    await this.saveRefreshToken(user.id, tokens.refreshToken);

    // Update last login
    const updatedUser = {
      ...user,
      lastLoginAt: new Date(),
      updatedAt: new Date(),
    };
    await this.userRepository.update(updatedUser);

    // Publish domain event
    await this.eventPublisher.publish(
      new UserLoginEvent(user.id, request.ipAddress, request.userAgent)
    );

    return {
      ...tokens,
      user: this.mapToUserProfile(updatedUser),
    };
  }

  // Use Case: Refresh Token
  async refreshToken(request: RefreshTokenRequest): Promise<AuthResponse> {
    // Verify refresh token
    const payload = this.tokenService.verifyRefreshToken(request.refreshToken);
    
    // Check if token exists in database
    const tokenData = await this.tokenRepository.findRefreshToken(request.refreshToken);
    if (!tokenData) {
      throw new InvalidTokenError();
    }

    if (tokenData.expiresAt < new Date()) {
      await this.tokenRepository.revokeRefreshToken(request.refreshToken);
      throw new TokenExpiredError();
    }

    // Get user
    const user = await this.userRepository.findById(tokenData.userId);
    if (!user) {
      throw new UserNotFoundError(tokenData.userId);
    }

    // Generate new tokens
    const tokens = this.generateTokens(user);
    
    // Revoke old refresh token and save new one
    await this.tokenRepository.revokeRefreshToken(request.refreshToken);
    await this.saveRefreshToken(user.id, tokens.refreshToken);

    return {
      ...tokens,
      user: this.mapToUserProfile(user),
    };
  }

  // Use Case: Verify Email
  async verifyEmail(request: VerifyEmailRequest): Promise<void> {
    const tokenData = await this.emailVerificationRepository.findVerificationToken(request.token);
    
    if (!tokenData) {
      throw new InvalidTokenError();
    }

    if (tokenData.expiresAt < new Date()) {
      await this.emailVerificationRepository.deleteVerificationToken(request.token);
      throw new TokenExpiredError();
    }

    // Mark email as verified
    await this.emailVerificationRepository.markEmailAsVerified(tokenData.userId);
    await this.emailVerificationRepository.deleteVerificationToken(request.token);

    // Send welcome email
    const user = await this.userRepository.findById(tokenData.userId);
    if (user) {
      await this.emailService.sendWelcomeEmail(user.email, user.firstName);
    }
  }

  // Use Case: Forgot Password
  async forgotPassword(request: ForgotPasswordRequest): Promise<void> {
    const email = new Email(request.email);
    const user = await this.userRepository.findByEmail(email);

    if (!user) {
      // Don't reveal if user exists for security
      return;
    }

    // Generate password reset token
    const resetToken = this.tokenService.generatePasswordResetToken();
    const expiresAt = new Date(Date.now() + 60 * 60 * 1000); // 1 hour

    // Save token (using email verification repository for simplicity)
    await this.emailVerificationRepository.saveVerificationToken(
      user.id,
      resetToken,
      expiresAt
    );

    // Send password reset email
    await this.emailService.sendPasswordResetEmail(email, resetToken);
  }

  // Use Case: Reset Password
  async resetPassword(request: ResetPasswordRequest): Promise<void> {
    const tokenData = await this.emailVerificationRepository.findVerificationToken(request.token);
    
    if (!tokenData) {
      throw new InvalidTokenError();
    }

    if (tokenData.expiresAt < new Date()) {
      await this.emailVerificationRepository.deleteVerificationToken(request.token);
      throw new TokenExpiredError();
    }

    // Get user and update password
    const user = await this.userRepository.findById(tokenData.userId);
    if (!user) {
      throw new UserNotFoundError(tokenData.userId);
    }

    // Hash new password
    const newPassword = new Password(request.newPassword);
    const hashedPassword = await this.passwordHashingService.hash(newPassword);

    // Update user with new password
    const updatedUser = {
      ...user,
      password: hashedPassword as any,
      updatedAt: new Date(),
    };
    await this.userRepository.update(updatedUser);

    // Clean up token
    await this.emailVerificationRepository.deleteVerificationToken(request.token);

    // Revoke all refresh tokens for security
    await this.tokenRepository.revokeAllUserTokens(user.id);
  }

  // Use Case: Logout
  async logout(refreshToken: string): Promise<void> {
    await this.tokenRepository.revokeRefreshToken(refreshToken);
  }

  // Use Case: Get User Profile
  async getUserProfile(userId: string): Promise<UserProfile> {
    const user = await this.userRepository.findById(userId);
    if (!user) {
      throw new UserNotFoundError(userId);
    }
    return this.mapToUserProfile(user);
  }

  // Use Case: Update User Profile
  async updateUserProfile(userId: string, updateData: Partial<{ firstName: string; lastName: string; username: string }>): Promise<UserProfile> {
    const user = await this.userRepository.findById(userId);
    if (!user) {
      throw new UserNotFoundError(userId);
    }

    // Create updated user object
    const updatedUser = new User(
      user.id,
      user.email,
      updateData.username || user.username,
      updateData.firstName || user.firstName,
      updateData.lastName || user.lastName,
      user.password,
      user.role,
      user.emailVerified,
      user.accountLocked,
      user.loginAttempts,
      user.lastLoginAt,
      user.createdAt,
      new Date(), // Updated timestamp
      user.deletedAt,
      user.provider,
      user.providerId,
      user.metadata
    );

    const savedUser = await this.userRepository.update(updatedUser);
    return this.mapToUserProfile(savedUser);
  }

  // Use Case: Resend Verification Email
  async resendVerificationEmail(userId: string): Promise<void> {
    const user = await this.userRepository.findById(userId);
    if (!user) {
      throw new UserNotFoundError(userId);
    }

    if (user.emailVerified) {
      throw new Error('Email is already verified');
    }

    // Generate new verification token
    const verificationToken = this.tokenService.generateVerificationToken();

    // Save verification token
    const expiresAt = new Date(Date.now() + 24 * 60 * 60 * 1000); // 24 hours
    await this.emailVerificationRepository.saveVerificationToken(user.id, verificationToken, expiresAt);

    // Send verification email
    await this.emailService.sendVerificationEmail(
      user.email,
      verificationToken
    );
  }

  // Use Case: Change Password
  async changePassword(userId: string, currentPassword: string, newPassword: string): Promise<void> {
    const user = await this.userRepository.findById(userId);
    if (!user) {
      throw new UserNotFoundError(userId);
    }

    // Verify current password
    if (user.password) {
      const currentPasswordObj = new Password(currentPassword);
      const isCurrentPasswordValid = await this.passwordHashingService.compare(
        currentPasswordObj,
        user.password.getValue()
      );
      
      if (!isCurrentPasswordValid) {
        throw new InvalidCredentialsError();
      }
    } else {
      throw new Error('Cannot change password for OAuth users');
    }

    // Create new password hash
    const newPasswordObj = new Password(newPassword);
    const newPasswordHash = await this.passwordHashingService.hash(newPasswordObj);
    const hashedPasswordObj = new Password(newPasswordHash);

    // Update user with new password
    const updatedUser = new User(
      user.id,
      user.email,
      user.username,
      user.firstName,
      user.lastName,
      hashedPasswordObj,
      user.role,
      user.emailVerified,
      user.accountLocked,
      user.loginAttempts,
      user.lastLoginAt,
      user.createdAt,
      new Date(), // Updated timestamp
      user.deletedAt,
      user.provider,
      user.providerId,
      user.metadata
    );

    await this.userRepository.update(updatedUser);
  }

  // Private helper methods following Single Responsibility Principle
  private generateTokens(user: User): { accessToken: string; refreshToken: string } {
    const payload = {
      userId: user.id,
      email: user.email,
      role: user.role,
    };

    return {
      accessToken: this.tokenService.generateAccessToken(payload),
      refreshToken: this.tokenService.generateRefreshToken(payload),
    };
  }

  private async saveRefreshToken(userId: string, refreshToken: string): Promise<void> {
    const expiresAt = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000); // 7 days
    await this.tokenRepository.saveRefreshToken(userId, refreshToken, expiresAt);
  }

  private async handleFailedLogin(userId: string): Promise<void> {
    await this.userRepository.incrementLoginAttempts(userId);
    
    const user = await this.userRepository.findById(userId);
    if (user && user.loginAttempts >= this.MAX_LOGIN_ATTEMPTS) {
      await this.userRepository.lockAccount(userId);
      await this.eventPublisher.publish(
        new UserAccountLockedEvent(userId, 'Too many failed login attempts')
      );
    }
  }

  private mapToUserProfile(user: User): UserProfile {
    return {
      id: user.id,
      email: user.email.getValue(),
      username: user.username,
      firstName: user.firstName,
      lastName: user.lastName,
      role: user.role,
      emailVerified: user.emailVerified,
      createdAt: user.createdAt,
    };
  }

  private generateUserId(): string {
    return `user_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }
}
