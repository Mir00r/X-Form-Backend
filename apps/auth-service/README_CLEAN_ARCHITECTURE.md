# Auth Service - Clean Architecture Implementation

## Overview
The Auth Service is a comprehensive authentication and user management system built using **Clean Architecture** principles and **SOLID design principles** with **TypeScript**. This service handles user registration, authentication, email verification, password management, and JWT token management.

## Architecture

### Clean Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    Interface Layer                          │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              HTTP Controllers                       │    │
│  │  • REST API endpoints                               │    │
│  │  • Request/Response mapping                         │    │
│  │  • HTTP status code handling                        │    │
│  │  • Authentication middleware                        │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  Application Layer                          │
│  ┌─────────────────────────────────────────────────────┐    │
│  │           Use Cases / Services                      │    │
│  │  • RegisterUser                                     │    │
│  │  • LoginUser                                        │    │
│  │  • RefreshToken                                     │    │
│  │  • VerifyEmail                                      │    │
│  │  • ForgotPassword                                   │    │
│  │  • ResetPassword                                    │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Domain Layer                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │         Business Entities & Rules                   │    │
│  │  • User entity                                      │    │
│  │  • Email/Password value objects                     │    │
│  │  • Domain events                                    │    │
│  │  • Business validation                              │    │
│  │  • Repository interfaces                            │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              ▲
                              │
┌─────────────────────────────────────────────────────────────┐
│                Infrastructure Layer                         │
│  ┌─────────────────────────────────────────────────────┐    │
│  │      External Concerns                              │    │
│  │  • PostgreSQL repositories                          │    │
│  │  • JWT token service                                │    │
│  │  • BCrypt password hashing                          │    │
│  │  • Email service                                    │    │
│  │  • Database connection                              │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

## SOLID Principles Implementation

### Single Responsibility Principle (SRP)
- **Domain entities**: Only contain business logic and rules
- **Application services**: Only orchestrate use cases
- **Repositories**: Only handle data persistence
- **Controllers**: Only handle HTTP request/response
- **Infrastructure services**: Only handle external concerns

### Open/Closed Principle (OCP)
- Interfaces allow extension without modification
- New authentication providers can be added via interfaces
- Event system allows extending functionality without changing core logic

### Liskov Substitution Principle (LSP)
- Repository implementations can be substituted (PostgreSQL, MongoDB, etc.)
- Authentication providers are interchangeable
- Service implementations follow interface contracts

### Interface Segregation Principle (ISP)
- Small, focused interfaces (UserRepository, TokenService, EmailService)
- Clients depend only on methods they use
- Separated concerns (authentication vs user management)

### Dependency Inversion Principle (DIP)
- High-level modules depend on abstractions
- Infrastructure implements domain interfaces
- Dependency injection throughout the application

## Project Structure

```
auth-service/
├── src/
│   ├── domain/                     # Domain Layer
│   │   └── auth.ts                 # Core entities, value objects, interfaces
│   ├── application/                # Application Layer
│   │   └── auth-service.ts         # Use cases and business logic
│   ├── infrastructure/             # Infrastructure Layer
│   │   ├── repositories.ts         # Repository implementations
│   │   └── container.ts            # Dependency injection container
│   ├── interface/
│   │   └── http/                   # Interface Layer
│   │       └── auth-controller.ts  # HTTP controllers and middleware
│   └── app.ts                      # Main application setup
├── dist/                           # Compiled TypeScript
├── tsconfig.json                   # TypeScript configuration
├── package.json                    # Dependencies and scripts
└── README.md                       # This file
```

## API Endpoints

### Authentication
```
POST   /api/v1/auth/register        # Register new user
POST   /api/v1/auth/login           # Login user
POST   /api/v1/auth/refresh         # Refresh access token
POST   /api/v1/auth/logout          # Logout user
```

### Email Verification
```
POST   /api/v1/auth/verify-email    # Verify email address
```

### Password Management
```
POST   /api/v1/auth/forgot-password # Request password reset
POST   /api/v1/auth/reset-password  # Reset password with token
```

### User Profile
```
GET    /api/v1/auth/profile         # Get user profile (authenticated)
```

### Health Check
```
GET    /health                      # Service health status
```

## Key Features

### 1. Clean Architecture Benefits
- **Testable**: Business rules can be tested independently
- **Independent of UI**: Can support multiple interfaces (REST, GraphQL)
- **Independent of Database**: Can swap between different databases
- **Independent of Frameworks**: Business logic not coupled to Express.js

### 2. TypeScript Benefits
- **Type Safety**: Compile-time error checking
- **Better IDE Support**: IntelliSense and refactoring
- **Self-Documenting**: Types serve as documentation
- **Maintainable**: Easier to refactor and maintain

### 3. Security Features
- **Password Hashing**: BCrypt with configurable salt rounds
- **JWT Tokens**: Secure access and refresh token system
- **Rate Limiting**: Protect against brute force attacks
- **Input Validation**: Comprehensive request validation
- **Account Locking**: Automatic account locking after failed attempts

### 4. Domain-Driven Design
- **Value Objects**: Email and Password with built-in validation
- **Domain Events**: User registration, login, account locked events
- **Rich Domain Models**: Business logic encapsulated in entities
- **Repository Pattern**: Abstracted data access

### 5. Event-Driven Architecture
- **Domain Events**: Publish events for business operations
- **Event Handlers**: Decoupled event processing
- **Extensible**: Easy to add new event handlers

## Getting Started

### Prerequisites
- Node.js 18+
- TypeScript 5+
- PostgreSQL 13+
- Redis (optional, for session management)

### Installation
```bash
# Clone the repository
git clone <repository-url>

# Navigate to auth service
cd X-Form-Backend/services/auth-service

# Install dependencies
npm install

# Build TypeScript
npm run build

# Set up environment variables
cp .env.example .env
```

### Configuration
Environment variables in `.env`:
```env
# Server
PORT=3001
NODE_ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=xform_auth
DB_USERNAME=postgres
DB_PASSWORD=password
DB_POOL_SIZE=20

# JWT
JWT_ACCESS_SECRET=your-access-secret
JWT_REFRESH_SECRET=your-refresh-secret
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

# Email (for production)
EMAIL_PROVIDER=sendgrid
EMAIL_API_KEY=your-sendgrid-api-key
FROM_EMAIL=noreply@xform.com

# CORS
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
```

### Database Setup
```sql
-- Create database
CREATE DATABASE xform_auth;

-- Create users table
CREATE TABLE users (
  id VARCHAR(255) PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  username VARCHAR(255) UNIQUE NOT NULL,
  first_name VARCHAR(255) NOT NULL,
  last_name VARCHAR(255) NOT NULL,
  password_hash VARCHAR(255),
  role VARCHAR(50) DEFAULT 'user',
  email_verified BOOLEAN DEFAULT FALSE,
  account_locked BOOLEAN DEFAULT FALSE,
  login_attempts INTEGER DEFAULT 0,
  last_login_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  provider VARCHAR(50) DEFAULT 'local',
  provider_id VARCHAR(255),
  metadata JSONB DEFAULT '{}'
);

-- Create refresh tokens table
CREATE TABLE refresh_tokens (
  id SERIAL PRIMARY KEY,
  user_id VARCHAR(255) REFERENCES users(id),
  token VARCHAR(512) UNIQUE NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  revoked BOOLEAN DEFAULT FALSE,
  revoked_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create email verification tokens table
CREATE TABLE email_verification_tokens (
  id SERIAL PRIMARY KEY,
  user_id VARCHAR(255) REFERENCES users(id),
  token VARCHAR(255) UNIQUE NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_verification_tokens_token ON email_verification_tokens(token);
```

### Development

#### Start Development Server
```bash
# With auto-reload
npm run dev:watch

# Single run
npm run dev
```

#### Build for Production
```bash
npm run build
npm start
```

#### Testing
```bash
# Run tests
npm test

# Run tests with coverage
npm run test:coverage

# Run tests in watch mode
npm run test:watch
```

#### Code Quality
```bash
# Lint code
npm run lint

# Fix lint issues
npm run lint:fix
```

## API Usage Examples

### Register User
```bash
curl -X POST http://localhost:3001/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "johndoe",
    "password": "SecurePass123",
    "firstName": "John",
    "lastName": "Doe"
  }'
```

### Login
```bash
curl -X POST http://localhost:3001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123"
  }'
```

### Get Profile (Authenticated)
```bash
curl -X GET http://localhost:3001/api/v1/auth/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Monitoring and Observability

### Health Checks
```json
{
  "success": true,
  "data": {
    "service": "auth-service",
    "version": "1.0.0",
    "architecture": "Clean Architecture with SOLID Principles",
    "status": "healthy",
    "components": {
      "database": "healthy",
      "jwt": "healthy",
      "email": "healthy"
    }
  }
}
```

### Logging
- Structured logging with request IDs
- Security event logging (failed logins, account locks)
- Performance metrics logging

## Testing Strategy

### Unit Tests
- Domain entity validation
- Use case logic testing
- Repository interface testing

### Integration Tests
- API endpoint testing
- Database integration testing
- Authentication flow testing

### End-to-End Tests
- Full user journey testing
- Cross-service integration testing

## Deployment

### Docker
```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production && npm cache clean --force

FROM node:18-alpine
WORKDIR /app
COPY --from=builder /app/node_modules ./node_modules
COPY dist ./dist
EXPOSE 3001
CMD ["node", "dist/app.js"]
```

### Kubernetes
Use the provided Kubernetes manifests for production deployment.

## Security Considerations

1. **Password Security**: BCrypt hashing with 12 salt rounds
2. **JWT Security**: Separate access and refresh tokens
3. **Rate Limiting**: Prevents brute force attacks
4. **Input Validation**: Comprehensive request validation
5. **CORS**: Configurable allowed origins
6. **Helmet**: Security headers for HTTP responses

## Contributing

1. Follow Clean Architecture principles
2. Maintain SOLID design patterns
3. Write comprehensive tests
4. Document new features
5. Follow TypeScript best practices

## License

[Your License Here]
