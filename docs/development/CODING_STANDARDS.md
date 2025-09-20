# X-Form Backend - Coding Standards

## 📋 Overview

This document outlines the coding standards and best practices for the X-Form Backend project. Following these standards ensures code consistency, maintainability, and quality across all microservices.

## 🏗️ Architecture Standards

### **Clean Architecture Principles**

All services must follow Clean Architecture:

```
📁 Service Structure
├── 📁 application/     # Use cases, application services
│   ├── 📁 usecases/    # Business use cases
│   ├── 📁 services/    # Application services
│   └── 📁 ports/       # Interfaces for external dependencies
├── 📁 domain/          # Business logic
│   ├── 📁 entities/    # Business entities
│   ├── 📁 services/    # Domain services
│   └── 📁 events/      # Domain events
├── 📁 infrastructure/ # External concerns
│   ├── 📁 database/   # Database implementation
│   ├── 📁 http/       # HTTP clients
│   └── 📁 messaging/  # Message queues
└── 📁 interfaces/     # External interfaces
    ├── 📁 http/       # HTTP controllers
    ├── 📁 grpc/       # gRPC services
    └── 📁 cli/        # CLI commands
```

### **Dependency Direction**
- **Domain** → No dependencies (pure business logic)
- **Application** → Domain only
- **Infrastructure** → Application, Domain
- **Interfaces** → Application, Domain

## 🎯 Language-Specific Standards

### **TypeScript/Node.js Services**

#### **Project Structure**
```typescript
// src/application/usecases/CreateUser.ts
export interface CreateUserUseCase {
  execute(request: CreateUserRequest): Promise<CreateUserResponse>;
}

// src/domain/entities/User.ts
export class User {
  constructor(
    private readonly id: UserId,
    private readonly email: Email,
    private readonly profile: UserProfile
  ) {}
  
  // Domain methods
  public changeEmail(newEmail: Email): void {
    // Business logic
  }
}

// src/infrastructure/database/UserRepository.ts
export class PostgresUserRepository implements UserRepository {
  async save(user: User): Promise<void> {
    // Database implementation
  }
}
```

#### **Naming Conventions**
- **Files**: PascalCase for classes, camelCase for utilities
- **Classes**: PascalCase (`UserService`, `CreateUserUseCase`)
- **Interfaces**: PascalCase with descriptive names (`UserRepository`)
- **Methods**: camelCase (`createUser`, `findById`)
- **Constants**: SCREAMING_SNAKE_CASE (`MAX_RETRY_ATTEMPTS`)

#### **Error Handling**
```typescript
// Domain errors
export class UserNotFoundError extends Error {
  constructor(userId: string) {
    super(`User with ID ${userId} not found`);
    this.name = 'UserNotFoundError';
  }
}

// Result pattern for use cases
export type CreateUserResult = 
  | { success: true; user: User }
  | { success: false; error: UserValidationError };

// Async error handling
try {
  const result = await createUser(request);
  if (!result.success) {
    return errorResponse(result.error);
  }
  return successResponse(result.user);
} catch (error) {
  logger.error('Unexpected error', { error, request });
  return internalErrorResponse();
}
```

#### **Type Safety**
```typescript
// Use strict types
interface CreateUserRequest {
  readonly email: string;
  readonly password: string;
  readonly profile: {
    readonly firstName: string;
    readonly lastName: string;
  };
}

// Use branded types for IDs
type UserId = string & { readonly brand: unique symbol };
type Email = string & { readonly brand: unique symbol };

// Use discriminated unions
type UserEvent = 
  | { type: 'USER_CREATED'; user: User }
  | { type: 'USER_UPDATED'; user: User; changes: UserChanges }
  | { type: 'USER_DELETED'; userId: UserId };
```

### **Go Services**

#### **Project Structure**
```go
// internal/application/usecases/create_user.go
package usecases

type CreateUserUseCase interface {
    Execute(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error)
}

// internal/domain/entities/user.go
package entities

type User struct {
    id       UserID
    email    Email
    profile  UserProfile
}

func (u *User) ChangeEmail(email Email) error {
    // Business logic
    return nil
}

// internal/infrastructure/database/user_repository.go
package database

type PostgresUserRepository struct {
    db *sql.DB
}

func (r *PostgresUserRepository) Save(ctx context.Context, user *entities.User) error {
    // Database implementation
    return nil
}
```

#### **Naming Conventions**
- **Packages**: lowercase, single word (`user`, `auth`, `forms`)
- **Files**: snake_case (`user_service.go`, `create_user.go`)
- **Types**: PascalCase (`User`, `CreateUserRequest`)
- **Functions**: PascalCase for exported, camelCase for private
- **Constants**: PascalCase (`MaxRetryAttempts`)

#### **Error Handling**
```go
// Domain errors
type UserNotFoundError struct {
    UserID string
}

func (e UserNotFoundError) Error() string {
    return fmt.Sprintf("user with ID %s not found", e.UserID)
}

// Wrapped errors
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    user, err := s.userRepo.Save(ctx, newUser)
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    return user, nil
}

// Error handling in handlers
func (h *UserHandler) CreateUser(c *gin.Context) {
    user, err := h.userService.CreateUser(c.Request.Context(), req)
    if err != nil {
        var notFoundErr UserNotFoundError
        if errors.As(err, &notFoundErr) {
            c.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
            return
        }
        c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "internal error"})
        return
    }
    c.JSON(http.StatusCreated, user)
}
```

#### **Interfaces and Dependency Injection**
```go
// Define interfaces in application layer
type UserRepository interface {
    Save(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id UserID) (*User, error)
}

// Dependency injection container
type Container struct {
    UserRepository UserRepository
    UserService    UserService
}

func NewContainer(db *sql.DB) *Container {
    userRepo := database.NewPostgresUserRepository(db)
    userService := services.NewUserService(userRepo)
    
    return &Container{
        UserRepository: userRepo,
        UserService:    userService,
    }
}
```

### **Python Services**

#### **Project Structure**
```python
# src/application/usecases/create_user.py
from abc import ABC, abstractmethod
from dataclasses import dataclass

class CreateUserUseCase(ABC):
    @abstractmethod
    async def execute(self, request: CreateUserRequest) -> CreateUserResponse:
        pass

# src/domain/entities/user.py
@dataclass(frozen=True)
class User:
    id: UserId
    email: Email
    profile: UserProfile
    
    def change_email(self, new_email: Email) -> 'User':
        return dataclasses.replace(self, email=new_email)

# src/infrastructure/database/user_repository.py
class PostgresUserRepository(UserRepository):
    async def save(self, user: User) -> None:
        # Database implementation
        pass
```

#### **Naming Conventions**
- **Files**: snake_case (`user_service.py`, `create_user.py`)
- **Classes**: PascalCase (`User`, `CreateUserRequest`)
- **Functions**: snake_case (`create_user`, `find_by_id`)
- **Constants**: SCREAMING_SNAKE_CASE (`MAX_RETRY_ATTEMPTS`)

## 📊 API Standards

### **RESTful API Design**

#### **Resource Naming**
```bash
# Good
GET    /api/v1/users
POST   /api/v1/users
GET    /api/v1/users/{id}
PUT    /api/v1/users/{id}
DELETE /api/v1/users/{id}

# Bad
GET    /api/v1/getUsers
POST   /api/v1/createUser
GET    /api/v1/user/{id}
```

#### **HTTP Status Codes**
```typescript
// Success
200 OK          // Successful GET, PUT
201 Created     // Successful POST
204 No Content  // Successful DELETE

// Client Errors
400 Bad Request     // Invalid request format
401 Unauthorized    // Missing/invalid authentication
403 Forbidden       // Valid auth but insufficient permissions
404 Not Found       // Resource doesn't exist
409 Conflict        // Resource state conflict
422 Unprocessable   // Validation errors

// Server Errors
500 Internal Server Error  // Unexpected server error
502 Bad Gateway           // Upstream service error
503 Service Unavailable   // Service temporarily down
```

#### **Request/Response Format**
```typescript
// Request
interface CreateUserRequest {
  email: string;
  password: string;
  profile: {
    firstName: string;
    lastName: string;
  };
}

// Success Response
interface CreateUserResponse {
  data: {
    id: string;
    email: string;
    profile: {
      firstName: string;
      lastName: string;
    };
    createdAt: string;
    updatedAt: string;
  };
  meta: {
    requestId: string;
    timestamp: string;
  };
}

// Error Response
interface ErrorResponse {
  error: {
    code: string;
    message: string;
    details?: Record<string, any>;
  };
  meta: {
    requestId: string;
    timestamp: string;
  };
}
```

### **API Versioning**
```bash
# URL versioning (preferred)
/api/v1/users
/api/v2/users

# Header versioning (alternative)
Accept: application/vnd.api+json;version=1
```

## 🧪 Testing Standards

### **Test Structure**
```typescript
// Unit test structure
describe('UserService', () => {
  let userService: UserService;
  let mockUserRepository: jest.Mocked<UserRepository>;
  
  beforeEach(() => {
    mockUserRepository = createMockUserRepository();
    userService = new UserService(mockUserRepository);
  });
  
  describe('createUser', () => {
    it('should create user when data is valid', async () => {
      // Given
      const request = createValidUserRequest();
      mockUserRepository.save.mockResolvedValue(undefined);
      
      // When
      const result = await userService.createUser(request);
      
      // Then
      expect(result.success).toBe(true);
      expect(mockUserRepository.save).toHaveBeenCalledWith(
        expect.objectContaining({
          email: request.email
        })
      );
    });
    
    it('should return error when email already exists', async () => {
      // Given
      const request = createValidUserRequest();
      mockUserRepository.save.mockRejectedValue(new EmailAlreadyExistsError());
      
      // When
      const result = await userService.createUser(request);
      
      // Then
      expect(result.success).toBe(false);
      expect(result.error).toBeInstanceOf(EmailAlreadyExistsError);
    });
  });
});
```

### **Test Categories**
- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test component interactions
- **Contract Tests**: Test API contracts between services
- **E2E Tests**: Test complete user workflows

### **Test Naming**
```typescript
// Pattern: should_[expectedBehavior]_when_[condition]
it('should_create_user_when_data_is_valid')
it('should_return_error_when_email_already_exists')
it('should_hash_password_when_creating_user')
```

## 📝 Documentation Standards

### **Code Documentation**
```typescript
/**
 * Creates a new user account with the provided information.
 * 
 * @param request - The user creation request containing email, password, and profile
 * @returns Promise resolving to the creation result
 * @throws {EmailAlreadyExistsError} When the email is already registered
 * @throws {ValidationError} When the request data is invalid
 * 
 * @example
 * ```typescript
 * const result = await userService.createUser({
 *   email: 'user@example.com',
 *   password: 'securePassword123',
 *   profile: { firstName: 'John', lastName: 'Doe' }
 * });
 * ```
 */
async createUser(request: CreateUserRequest): Promise<CreateUserResult> {
  // Implementation
}
```

### **README Structure**
```markdown
# Service Name

Brief description of the service

## Features
- Feature 1
- Feature 2

## Quick Start
\`\`\`bash
# Installation and basic usage
\`\`\`

## API Documentation
Link to detailed API docs

## Configuration
Environment variables and configuration options

## Development
Setup and development instructions

## Testing
How to run tests

## Deployment
Deployment instructions
```

## 🔒 Security Standards

### **Authentication & Authorization**
```typescript
// JWT token validation
interface JWTPayload {
  sub: string;  // User ID
  email: string;
  roles: string[];
  iat: number;  // Issued at
  exp: number;  // Expires at
}

// Role-based access control
const requireRole = (roles: string[]) => {
  return (req: AuthenticatedRequest, res: Response, next: NextFunction) => {
    if (!req.user.roles.some(role => roles.includes(role))) {
      return res.status(403).json({ error: 'Insufficient permissions' });
    }
    next();
  };
};
```

### **Input Validation**
```typescript
// Use schema validation
const createUserSchema = Joi.object({
  email: Joi.string().email().required(),
  password: Joi.string().min(8).pattern(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/).required(),
  profile: Joi.object({
    firstName: Joi.string().min(1).max(50).required(),
    lastName: Joi.string().min(1).max(50).required()
  }).required()
});

// Sanitize inputs
const sanitizedInput = DOMPurify.sanitize(userInput);
```

### **Database Security**
```typescript
// Use parameterized queries
const query = 'SELECT * FROM users WHERE email = $1 AND active = $2';
const result = await db.query(query, [email, true]);

// Never concatenate user input
// ❌ BAD: `SELECT * FROM users WHERE email = '${email}'`
// ✅ GOOD: Use parameterized queries
```

## 📊 Performance Standards

### **Response Time Targets**
- **Health checks**: < 10ms
- **Simple queries**: < 100ms
- **Complex operations**: < 500ms
- **File uploads**: < 5s (depending on size)

### **Database Optimization**
```sql
-- Use proper indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_forms_user_id ON forms(user_id);

-- Use explain analyze for query optimization
EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'user@example.com';
```

### **Caching Strategy**
```typescript
// Redis caching
const cacheKey = `user:${userId}`;
const cachedUser = await redis.get(cacheKey);

if (cachedUser) {
  return JSON.parse(cachedUser);
}

const user = await userRepository.findById(userId);
await redis.setex(cacheKey, 300, JSON.stringify(user)); // 5 min TTL
return user;
```

## 🔧 Configuration Standards

### **Environment Variables**
```bash
# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/dbname
REDIS_URL=redis://localhost:6379

# Authentication
JWT_SECRET=your-secret-key
JWT_EXPIRE=24h

# External Services
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_REGION=us-east-1

# Application
NODE_ENV=development
LOG_LEVEL=debug
PORT=3001
```

### **Configuration Validation**
```typescript
// Validate configuration on startup
const config = {
  database: {
    url: process.env.DATABASE_URL!,
  },
  jwt: {
    secret: process.env.JWT_SECRET!,
    expire: process.env.JWT_EXPIRE || '24h',
  },
  port: parseInt(process.env.PORT || '3001'),
};

// Validate required environment variables
const requiredEnvVars = ['DATABASE_URL', 'JWT_SECRET'];
const missing = requiredEnvVars.filter(env => !process.env[env]);

if (missing.length > 0) {
  throw new Error(`Missing required environment variables: ${missing.join(', ')}`);
}
```

## 📏 Code Quality Metrics

### **Complexity Limits**
- **Cyclomatic complexity**: < 10 per function
- **Function length**: < 50 lines
- **File length**: < 500 lines
- **Parameter count**: < 5 parameters

### **Test Coverage Targets**
- **Overall coverage**: > 80%
- **Critical paths**: > 95%
- **New code**: > 90%

### **Code Review Checklist**
- [ ] Follows clean architecture principles
- [ ] Has appropriate test coverage
- [ ] Includes proper error handling
- [ ] Uses consistent naming conventions
- [ ] Has necessary documentation
- [ ] Passes all linting rules
- [ ] Has no security vulnerabilities

---

These standards ensure high code quality, maintainability, and consistency across the X-Form Backend project. All team members should follow these guidelines and update them as the project evolves.
