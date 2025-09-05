// Dependency Injection Container for Auth Service
// Following Inversion of Control (IoC) pattern and Dependency Inversion Principle

import { Pool } from 'pg';
import { AuthApplicationService, EventPublisher } from '../application/auth-service';
import {
  PostgreSQLUserRepository,
  PostgreSQLTokenRepository,
  PostgreSQLEmailVerificationRepository,
  BCryptPasswordHashingService,
  JWTTokenService,
  MockEmailService,
} from '../infrastructure/repositories';
import { AuthController } from '../interface/http/auth-controller';
import { 
  DomainEvent, 
  UserRegisteredEvent, 
  UserLoginEvent, 
  UserAccountLockedEvent 
} from '../domain/auth';

// Configuration interface
export interface AuthServiceConfig {
  database: {
    host: string;
    port: number;
    database: string;
    username: string;
    password: string;
    ssl?: boolean;
    poolSize?: number;
  };
  jwt: {
    accessTokenSecret: string;
    refreshTokenSecret: string;
    accessTokenExpiry?: string;
    refreshTokenExpiry?: string;
  };
  email?: {
    provider: 'sendgrid' | 'ses' | 'mock';
    apiKey?: string;
    fromEmail: string;
  };
}

// Simple event publisher implementation
class InMemoryEventPublisher implements EventPublisher {
  private handlers: Map<string, Array<(event: DomainEvent) => Promise<void>>> = new Map();

  async publish(event: DomainEvent): Promise<void> {
    const eventType = event.constructor.name;
    const handlers = this.handlers.get(eventType) || [];
    
    // Execute all handlers for this event type
    await Promise.all(handlers.map(handler => handler(event)));
    
    // Log event for debugging
    console.log(`Published event: ${eventType}`, {
      eventId: event.eventId,
      occurredOn: event.occurredOn,
    });
  }

  subscribe<T extends DomainEvent>(
    eventType: new (...args: any[]) => T,
    handler: (event: T) => Promise<void>
  ): void {
    const eventTypeName = eventType.name;
    if (!this.handlers.has(eventTypeName)) {
      this.handlers.set(eventTypeName, []);
    }
    this.handlers.get(eventTypeName)!.push(handler as any);
  }
}

// Main dependency injection container
export class AuthServiceContainer {
  private readonly config: AuthServiceConfig;
  private readonly pool: Pool;
  private readonly eventPublisher: EventPublisher;

  // Repositories
  private readonly userRepository: PostgreSQLUserRepository;
  private readonly tokenRepository: PostgreSQLTokenRepository;
  private readonly emailVerificationRepository: PostgreSQLEmailVerificationRepository;

  // Services
  private readonly passwordHashingService: BCryptPasswordHashingService;
  private readonly tokenService: JWTTokenService;
  private readonly emailService: MockEmailService;

  // Application services
  private readonly authApplicationService: AuthApplicationService;

  // Controllers
  private readonly authController: AuthController;

  constructor(config: AuthServiceConfig) {
    this.config = config;

    // Initialize database pool
    this.pool = new Pool({
      host: config.database.host,
      port: config.database.port,
      database: config.database.database,
      user: config.database.username,
      password: config.database.password,
      ssl: config.database.ssl,
      max: config.database.poolSize || 20,
      idleTimeoutMillis: 30000,
      connectionTimeoutMillis: 10000,
    });

    // Initialize event publisher
    this.eventPublisher = new InMemoryEventPublisher();

    // Initialize repositories (Infrastructure Layer)
    this.userRepository = new PostgreSQLUserRepository(this.pool);
    this.tokenRepository = new PostgreSQLTokenRepository(this.pool);
    this.emailVerificationRepository = new PostgreSQLEmailVerificationRepository(this.pool);

    // Initialize domain services (Infrastructure Layer)
    this.passwordHashingService = new BCryptPasswordHashingService();
    this.tokenService = new JWTTokenService(
      config.jwt.accessTokenSecret,
      config.jwt.refreshTokenSecret,
      config.jwt.accessTokenExpiry,
      config.jwt.refreshTokenExpiry
    );
    this.emailService = new MockEmailService(); // TODO: Replace with real email service

    // Initialize application service (Application Layer)
    this.authApplicationService = new AuthApplicationService(
      this.userRepository,
      this.tokenRepository,
      this.emailVerificationRepository,
      this.passwordHashingService,
      this.tokenService,
      this.emailService,
      this.eventPublisher
    );

    // Initialize controllers (Interface Layer)
    this.authController = new AuthController(this.authApplicationService);

    // Setup event handlers
    this.setupEventHandlers();
  }

  // Getters for accessing dependencies (following Dependency Injection pattern)
  getAuthController(): AuthController {
    return this.authController;
  }

  getAuthApplicationService(): AuthApplicationService {
    return this.authApplicationService;
  }

  getDatabasePool(): Pool {
    return this.pool;
  }

  getEventPublisher(): EventPublisher {
    return this.eventPublisher;
  }

  // Setup domain event handlers
  private setupEventHandlers(): void {
    const eventPublisher = this.eventPublisher as InMemoryEventPublisher;

    // Example: Log user registration events
    eventPublisher.subscribe(
      UserRegisteredEvent,
      async (event: UserRegisteredEvent) => {
        console.log(`User registered: ${event.email} with provider: ${event.provider}`);
        // Here you could trigger additional actions like:
        // - Send analytics event
        // - Update user metrics
        // - Trigger welcome email workflow
      }
    );

    // Example: Log user login events for security monitoring
    eventPublisher.subscribe(
      UserLoginEvent,
      async (event: UserLoginEvent) => {
        console.log(`User login: ${event.userId} from IP: ${event.ipAddress}`);
        // Here you could:
        // - Log security events
        // - Update last login timestamp
        // - Detect suspicious login patterns
      }
    );

    // Example: Handle account lockout events
    eventPublisher.subscribe(
      UserAccountLockedEvent,
      async (event: UserAccountLockedEvent) => {
        console.log(`Account locked: ${event.userId} - ${event.reason}`);
        // Here you could:
        // - Send security alert email
        // - Log security incident
        // - Trigger admin notification
      }
    );
  }

  // Graceful shutdown
  async close(): Promise<void> {
    console.log('Closing Auth Service Container...');
    
    // Close database pool
    await this.pool.end();
    
    console.log('Auth Service Container closed successfully');
  }

  // Health check method
  async healthCheck(): Promise<{ status: string; components: Record<string, string> }> {
    const health = {
      status: 'healthy',
      components: {} as Record<string, string>,
    };

    try {
      // Check database connection
      const client = await this.pool.connect();
      await client.query('SELECT 1');
      client.release();
      health.components.database = 'healthy';
    } catch (error) {
      health.components.database = 'unhealthy';
      health.status = 'unhealthy';
    }

    // Check other components as needed
    health.components.jwt = 'healthy'; // JWT service is always available
    health.components.email = 'healthy'; // Mock service is always available

    return health;
  }
}

// Factory function to create configured container
export function createAuthServiceContainer(): AuthServiceContainer {
  const config: AuthServiceConfig = {
    database: {
      host: process.env.DB_HOST || 'localhost',
      port: parseInt(process.env.DB_PORT || '5432'),
      database: process.env.DB_NAME || 'xform_auth',
      username: process.env.DB_USERNAME || 'postgres',
      password: process.env.DB_PASSWORD || 'password',
      ssl: process.env.NODE_ENV === 'production',
      poolSize: parseInt(process.env.DB_POOL_SIZE || '20'),
    },
    jwt: {
      accessTokenSecret: process.env.JWT_ACCESS_SECRET || 'default-access-secret',
      refreshTokenSecret: process.env.JWT_REFRESH_SECRET || 'default-refresh-secret',
      accessTokenExpiry: process.env.JWT_ACCESS_EXPIRY || '15m',
      refreshTokenExpiry: process.env.JWT_REFRESH_EXPIRY || '7d',
    },
    email: {
      provider: 'mock',
      fromEmail: process.env.FROM_EMAIL || 'noreply@xform.com',
    },
  };

  return new AuthServiceContainer(config);
}
