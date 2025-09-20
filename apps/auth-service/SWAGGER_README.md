# X-Form Auth Service - Comprehensive Swagger Documentation

## 🎯 Overview

The X-Form Auth Service is a production-ready authentication and user management microservice built with **Clean Architecture** and **SOLID principles**. This service includes comprehensive **OpenAPI 3.0/Swagger documentation** following current industry best practices.

## 🏗️ Architecture Features

### Clean Architecture Implementation
- **Domain Layer**: Core business logic and entities
- **Application Layer**: Use cases and business workflows  
- **Infrastructure Layer**: External concerns (database, email, JWT)
- **Interface Layer**: HTTP controllers and API endpoints

### SOLID Principles Applied
- ✅ **Single Responsibility**: Each class has one reason to change
- ✅ **Open/Closed**: Open for extension, closed for modification
- ✅ **Liskov Substitution**: Interfaces enable substitutability
- ✅ **Interface Segregation**: Small, focused interfaces
- ✅ **Dependency Inversion**: Depend on abstractions, not concretions

## 📚 API Documentation Features

### Comprehensive OpenAPI 3.0 Specification
- **Interactive Swagger UI** with Try-It-Out functionality
- **Detailed endpoint documentation** with examples
- **Request/Response schemas** with validation rules
- **Authentication flows** with JWT bearer tokens
- **Error responses** with standardized error codes
- **Rate limiting information** and security details

### Documentation Endpoints
```
🌐 http://localhost:3001/api-docs           # Interactive Swagger UI
🌐 http://localhost:3001/docs               # Alternative route
📋 http://localhost:3001/api-docs.json      # Raw OpenAPI specification
🎨 http://localhost:3001/api-docs-custom    # Custom styled documentation
```

## 🚀 Quick Start Guide

### Prerequisites
- Node.js 18+ 
- npm or yarn
- PostgreSQL 13+ (for full functionality)

### Installation

1. **Install Dependencies**
   ```bash
   cd /Users/mir00r/Developer/kamkaiz/X-Form-Backend/services/auth-service
   npm install
   ```

2. **Environment Configuration**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start Development Server**
   ```bash
   # Full application with Clean Architecture
   npm run dev
   
   # OR simplified version for quick testing
   npm run dev:simple
   ```

4. **Access Documentation**
   - Open browser to: http://localhost:3001/api-docs
   - Interactive API testing available immediately

## 📋 Available API Endpoints

### Authentication Endpoints
```
POST /api/v1/auth/register          # User registration
POST /api/v1/auth/login            # User authentication  
POST /api/v1/auth/refresh          # Token refresh
POST /api/v1/auth/logout           # User logout
POST /api/v1/auth/verify-email     # Email verification
```

### User Management Endpoints
```
GET  /api/v1/auth/profile          # Get user profile
PUT  /api/v1/auth/profile          # Update user profile
POST /api/v1/auth/resend-verification  # Resend verification email
```

### Password Management Endpoints
```
POST /api/v1/auth/forgot-password  # Request password reset
POST /api/v1/auth/reset-password   # Reset password with token
PUT  /api/v1/auth/change-password  # Change password (authenticated)
```

### Health & Monitoring
```
GET  /health                       # Service health check
GET  /api-docs/health             # Documentation health check
```

## 🔧 Development Scripts

### Core Commands
```bash
npm run dev                # Start development server with hot reload
npm run dev:simple         # Start simplified demo server
npm run build              # Build TypeScript to JavaScript
npm run start              # Start production server
npm run test               # Run test suite
npm run test:watch         # Run tests in watch mode
npm run lint               # Run ESLint code analysis
npm run lint:fix           # Fix auto-fixable lint issues
```

### Database Commands
```bash
npm run db:migrate         # Run database migrations
npm run db:seed            # Seed database with test data
```

## 🛡️ Security Features

### Authentication & Authorization
- **JWT Access Tokens** (15-minute expiry)
- **JWT Refresh Tokens** (7-day expiry)
- **Bearer token authentication** for protected endpoints
- **Role-based access control** (USER, ADMIN, MODERATOR)

### Rate Limiting
- **Global rate limit**: 1000 requests per 15 minutes per IP
- **Auth endpoints**: 10 attempts per 15 minutes per IP
- **Headers included**: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

### Password Security
- **BCrypt hashing** with 12 salt rounds
- **Password complexity requirements**
- **Account lockout** after 5 failed attempts
- **Password history** tracking

### Data Protection
- **Input validation** with express-validator
- **SQL injection protection** with parameterized queries
- **XSS protection** with helmet.js
- **CORS configuration** for allowed origins
- **Security headers** with helmet.js

## 📖 Swagger Documentation Details

### Interactive Features
- **Try It Out**: Test APIs directly from documentation
- **Authentication Testing**: Built-in JWT token management
- **Request Examples**: Pre-filled example requests
- **Response Examples**: Complete response structures
- **Schema Validation**: Real-time request validation

### API Documentation Standards
- **OpenAPI 3.0.3** specification
- **Comprehensive schemas** for all DTOs
- **Detailed error responses** with error codes
- **Security scheme definitions** for authentication
- **Request/response examples** for all endpoints
- **Parameter validation rules** and constraints

### Custom Features
- **Request correlation tracking** with X-Correlation-ID headers
- **Enhanced error responses** with correlation IDs
- **Custom CSS styling** for improved readability
- **Persistent authorization** across browser sessions
- **API filtering and search** functionality

## 🔍 Testing the APIs

### Using Swagger UI (Recommended)
1. Navigate to http://localhost:3001/api-docs
2. Click "Authorize" button
3. For protected endpoints, use format: `Bearer your-jwt-token`
4. Use "Try it out" on any endpoint

### Sample Authentication Flow
1. **Register User**:
   ```json
   POST /api/v1/auth/register
   {
     "email": "john.doe@example.com",
     "username": "johndoe",
     "password": "SecurePass123!",
     "confirmPassword": "SecurePass123!",
     "firstName": "John",
     "lastName": "Doe",
     "acceptTerms": true
   }
   ```

2. **Login User**:
   ```json
   POST /api/v1/auth/login
   {
     "email": "john.doe@example.com",
     "password": "SecurePass123!"
   }
   ```

3. **Use Access Token**:
   ```bash
   Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   ```

### Using cURL
```bash
# Register user
curl -X POST http://localhost:3001/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","username":"testuser","password":"SecurePass123!","confirmPassword":"SecurePass123!","firstName":"Test","lastName":"User","acceptTerms":true}'

# Login user  
curl -X POST http://localhost:3001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"SecurePass123!"}'

# Get profile (with token)
curl -X GET http://localhost:3001/api/v1/auth/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## 🏥 Health Checks & Monitoring

### Health Check Endpoint
```bash
GET /health
```

**Response Example**:
```json
{
  "success": true,
  "data": {
    "service": "auth-service",
    "version": "1.0.0",
    "status": "healthy",
    "uptime": 3600.5,
    "architecture": "Clean Architecture with SOLID Principles",
    "dependencies": [
      {
        "name": "postgresql",
        "status": "healthy",
        "responseTime": 25.5
      }
    ]
  }
}
```

## 🛠️ Troubleshooting

### Common Issues

1. **Module Not Found Errors**
   ```bash
   npm install
   npm install @types/swagger-ui-express @types/swagger-jsdoc --save-dev
   ```

2. **Port Already in Use**
   ```bash
   export PORT=3002
   npm run dev
   ```

3. **Database Connection Issues**
   - Verify PostgreSQL is running
   - Check connection string in .env
   - Run database migrations

4. **Swagger UI Not Loading**
   - Check console for CSP errors
   - Verify all swagger dependencies installed
   - Try accessing /api-docs.json directly

### Development Tips
- Use `npm run dev:simple` for quick API testing
- Check logs for detailed error information
- Use browser dev tools to debug Swagger UI issues
- Verify CORS settings for cross-origin requests

## 📦 Dependencies

### Core Dependencies
- **express**: Web framework
- **swagger-jsdoc**: OpenAPI specification generation
- **swagger-ui-express**: Swagger UI middleware
- **jsonwebtoken**: JWT token handling
- **bcryptjs**: Password hashing
- **express-validator**: Request validation
- **helmet**: Security headers
- **cors**: Cross-origin resource sharing

### Development Dependencies
- **typescript**: Type checking
- **ts-node**: TypeScript execution
- **@types/swagger-ui-express**: Swagger UI types
- **@types/swagger-jsdoc**: Swagger JSDoc types
- **eslint**: Code linting
- **jest**: Testing framework

## 🎯 Production Deployment

### Environment Variables
```bash
NODE_ENV=production
PORT=3001
DATABASE_URL=postgresql://user:pass@localhost:5432/authdb
JWT_SECRET=your-super-secret-jwt-key
JWT_REFRESH_SECRET=your-refresh-token-secret
EMAIL_SERVICE_URL=https://your-email-service.com
REDIS_URL=redis://localhost:6379
```

### Build and Deploy
```bash
npm run build
npm start
```

### Docker Deployment
```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY dist ./dist
EXPOSE 3001
CMD ["npm", "start"]
```

## 📋 API Documentation Standards Compliance

### OpenAPI 3.0.3 Features
- ✅ **Complete schema definitions** for all request/response objects
- ✅ **Security schemes** with JWT bearer token support
- ✅ **Comprehensive error responses** with standard HTTP status codes
- ✅ **Request validation** with parameter constraints
- ✅ **Example values** for all schema properties
- ✅ **Tag organization** for logical endpoint grouping
- ✅ **Server definitions** for multiple environments

### Industry Best Practices
- ✅ **RESTful API design** with proper HTTP methods
- ✅ **Consistent error response format** across all endpoints
- ✅ **Request correlation tracking** for debugging
- ✅ **Rate limiting documentation** with header information
- ✅ **Authentication flow documentation** with examples
- ✅ **Comprehensive field validation** with clear error messages

## 🤝 Contributing

### Development Workflow
1. Follow Clean Architecture principles
2. Maintain SOLID design patterns
3. Update Swagger documentation for new endpoints
4. Add comprehensive tests for new features
5. Follow TypeScript best practices

### Code Standards
- Use TypeScript for type safety
- Follow ESLint configuration
- Maintain 80%+ test coverage
- Document all public APIs
- Use meaningful commit messages

---

## 📞 Support

For questions or issues:
- 📧 Email: dev@xform.com
- 🌐 Documentation: http://localhost:3001/api-docs
- 📋 Health Check: http://localhost:3001/health

---

**Built with ❤️ using Clean Architecture and SOLID Principles**
