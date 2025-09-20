// DTOs (Data Transfer Objects) for Auth Service API
// Following Interface Segregation Principle and API contract stability

// Base DTOs
export interface BaseApiResponse<T = any> {
  success: boolean;
  timestamp: string;
  path: string;
  method: string;
  correlationId: string;
  data?: T;
  error?: ApiError;
  meta?: ResponseMeta;
}

export interface ApiError {
  code: string;
  message: string;
  details?: ValidationError[] | Record<string, any>;
  timestamp: string;
  path: string;
  correlationId: string;
}

export interface ValidationError {
  field: string;
  message: string;
  value?: any;
  code: string;
}

export interface ResponseMeta {
  version: string;
  pagination?: PaginationMeta;
  rateLimit?: RateLimitMeta;
}

export interface PaginationMeta {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
  hasNext: boolean;
  hasPrev: boolean;
}

export interface RateLimitMeta {
  limit: number;
  remaining: number;
  resetTime: string;
}

// Auth-specific Request DTOs
export interface RegisterUserRequestDTO {
  email: string;
  username: string;
  password: string;
  confirmPassword: string;
  firstName: string;
  lastName: string;
  acceptTerms: boolean;
  marketingConsent?: boolean;
  referralCode?: string;
}

export interface LoginRequestDTO {
  email: string;
  password: string;
  rememberMe?: boolean;
  deviceId?: string;
  deviceName?: string;
}

export interface RefreshTokenRequestDTO {
  refreshToken: string;
  deviceId?: string;
}

export interface VerifyEmailRequestDTO {
  token: string;
  email?: string;
}

export interface ForgotPasswordRequestDTO {
  email: string;
  callbackUrl?: string;
}

export interface ResetPasswordRequestDTO {
  token: string;
  email: string;
  newPassword: string;
  confirmPassword: string;
}

export interface ChangePasswordRequestDTO {
  currentPassword: string;
  newPassword: string;
  confirmPassword: string;
}

export interface UpdateProfileRequestDTO {
  firstName?: string;
  lastName?: string;
  phoneNumber?: string;
  dateOfBirth?: string;
  timezone?: string;
  language?: string;
}

// Auth-specific Response DTOs
export interface AuthResponseDTO {
  accessToken: string;
  refreshToken: string;
  tokenType: string;
  expiresIn: number;
  expiresAt: string;
  scope: string[];
  user: UserProfileDTO;
}

export interface UserProfileDTO {
  id: string;
  email: string;
  username: string;
  firstName: string;
  lastName: string;
  fullName: string;
  role: string;
  emailVerified: boolean;
  phoneVerified: boolean;
  accountStatus: 'ACTIVE' | 'SUSPENDED' | 'PENDING_VERIFICATION' | 'LOCKED';
  lastLoginAt?: string;
  createdAt: string;
  updatedAt: string;
  preferences: UserPreferencesDTO;
  metadata: Record<string, any>;
}

export interface UserPreferencesDTO {
  language: string;
  timezone: string;
  emailNotifications: boolean;
  smsNotifications: boolean;
  marketingEmails: boolean;
  twoFactorEnabled: boolean;
}

export interface TokenInfoDTO {
  tokenId: string;
  type: 'ACCESS' | 'REFRESH' | 'EMAIL_VERIFICATION' | 'PASSWORD_RESET';
  issuedAt: string;
  expiresAt: string;
  isValid: boolean;
  deviceId?: string;
  deviceName?: string;
  ipAddress?: string;
  userAgent?: string;
}

export interface PasswordPolicyDTO {
  minLength: number;
  requireUppercase: boolean;
  requireLowercase: boolean;
  requireNumbers: boolean;
  requireSpecialChars: boolean;
  maxAge: number;
  historyCount: number;
  complexity: 'LOW' | 'MEDIUM' | 'HIGH';
}

// Success Response DTOs
export interface RegisterResponseDTO {
  message: string;
  userId: string;
  verificationRequired: boolean;
  verificationMethod: 'EMAIL' | 'SMS' | 'MANUAL';
  nextSteps: string[];
}

export interface LoginResponseDTO extends AuthResponseDTO {
  loginAttempts: number;
  accountStatus: string;
  mustChangePassword: boolean;
  twoFactorRequired: boolean;
  deviceTrusted: boolean;
}

export interface LogoutResponseDTO {
  message: string;
  loggedOutAt: string;
  devicesLoggedOut: number;
}

export interface VerifyEmailResponseDTO {
  message: string;
  emailVerified: boolean;
  verifiedAt: string;
  canProceed: boolean;
}

export interface ForgotPasswordResponseDTO {
  message: string;
  resetTokenSent: boolean;
  expiresIn: number;
  nextSteps: string[];
}

export interface ResetPasswordResponseDTO {
  message: string;
  passwordReset: boolean;
  resetAt: string;
  autoLogin: boolean;
  tokens?: AuthResponseDTO;
}

// Health Check DTOs
export interface HealthCheckResponseDTO {
  service: string;
  version: string;
  status: 'HEALTHY' | 'UNHEALTHY' | 'DEGRADED';
  uptime: number;
  timestamp: string;
  environment: string;
  dependencies: DependencyHealthDTO[];
  metrics: ServiceMetricsDTO;
}

export interface DependencyHealthDTO {
  name: string;
  type: 'DATABASE' | 'CACHE' | 'EMAIL' | 'EXTERNAL_API';
  status: 'HEALTHY' | 'UNHEALTHY' | 'DEGRADED';
  responseTime?: number;
  lastChecked: string;
  error?: string;
}

export interface ServiceMetricsDTO {
  requestCount: number;
  errorRate: number;
  averageResponseTime: number;
  activeConnections: number;
  memoryUsage: {
    used: number;
    free: number;
    total: number;
    percentage: number;
  };
  cpuUsage: number;
}

// Audit and Security DTOs
export interface SecurityEventDTO {
  eventId: string;
  type: 'LOGIN_SUCCESS' | 'LOGIN_FAILED' | 'PASSWORD_CHANGED' | 'ACCOUNT_LOCKED' | 'SUSPICIOUS_ACTIVITY';
  userId?: string;
  ipAddress: string;
  userAgent: string;
  timestamp: string;
  severity: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  details: Record<string, any>;
  location?: {
    country?: string;
    city?: string;
    latitude?: number;
    longitude?: number;
  };
}

export interface AuditLogDTO {
  id: string;
  action: string;
  resource: string;
  userId?: string;
  changes: Record<string, { old: any; new: any }>;
  timestamp: string;
  ipAddress: string;
  userAgent: string;
  correlationId: string;
}

// API Documentation DTOs
export interface ApiInfoDTO {
  title: string;
  version: string;
  description: string;
  contact: {
    name: string;
    email: string;
    url: string;
  };
  license: {
    name: string;
    url: string;
  };
  servers: Array<{
    url: string;
    description: string;
    environment: string;
  }>;
}

// Error Response DTOs
export interface NotFoundErrorDTO extends ApiError {
  code: 'RESOURCE_NOT_FOUND';
  resource: string;
  identifier: string;
}

export interface ValidationErrorDTO extends ApiError {
  code: 'VALIDATION_ERROR';
  details: ValidationError[];
}

export interface AuthenticationErrorDTO extends ApiError {
  code: 'AUTHENTICATION_FAILED' | 'TOKEN_EXPIRED' | 'TOKEN_INVALID' | 'CREDENTIALS_INVALID';
  authenticationMethod: string;
  retryAfter?: number;
}

export interface AuthorizationErrorDTO extends ApiError {
  code: 'INSUFFICIENT_PERMISSIONS' | 'ACCESS_DENIED' | 'ROLE_REQUIRED';
  requiredPermissions: string[];
  userPermissions: string[];
}

export interface RateLimitErrorDTO extends ApiError {
  code: 'RATE_LIMIT_EXCEEDED';
  retryAfter: number;
  limit: number;
  remaining: number;
  resetTime: string;
}

export interface InternalErrorDTO extends ApiError {
  code: 'INTERNAL_SERVER_ERROR' | 'DATABASE_ERROR' | 'EXTERNAL_SERVICE_ERROR';
  errorId: string;
  reportedAt: string;
}
