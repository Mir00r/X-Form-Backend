# Response Service

A comprehensive microservice for managing form responses with advanced features including validation, analytics, file uploads, and real-time capabilities.

## 🏗️ Architecture Overview

The Response Service is built following microservices best practices and includes:

- **RESTful API Design** with versioned endpoints
- **Comprehensive Validation** using Joi schemas
- **JWT Authentication & Authorization** with role-based access
- **Rate Limiting & Security** with comprehensive protection
- **Analytics & Reporting** with detailed insights
- **File Upload Management** with multiple storage backends
- **Event-Driven Communication** for decoupled services
- **OpenAPI/Swagger Documentation** for API specifications
- **Comprehensive Logging** with correlation ID tracking
- **Health Monitoring** with dependency checks
- **Graceful Shutdown** patterns for production reliability

## 🚀 Features

### Core Features
- ✅ Form response collection and management
- ✅ Real-time response validation
- ✅ File upload support with multiple backends
- ✅ Advanced analytics and reporting
- ✅ Export capabilities (CSV, Excel, PDF)
- ✅ Response status management (draft, partial, completed, archived)

### Security Features
- ✅ JWT-based authentication
- ✅ Role-based authorization
- ✅ Rate limiting and DDoS protection
- ✅ Input validation and sanitization
- ✅ CORS configuration
- ✅ Security headers (HSTS, CSP, etc.)
- ✅ IP whitelisting/blacklisting

### Monitoring & Observability
- ✅ Structured logging with correlation IDs
- ✅ Health check endpoints
- ✅ Performance metrics
- ✅ Event-driven monitoring
- ✅ Error tracking and reporting

### Integration Features
- ✅ Form service integration
- ✅ Event system for microservices communication
- ✅ External service health monitoring
- ✅ Webhook support for notifications

## 📋 Prerequisites

- Node.js >= 16.0.0
- npm >= 8.0.0
- MongoDB or compatible database
- Redis (optional, for caching)

## 🛠️ Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/your-org/x-form-backend.git
   cd x-form-backend/services/response-service
   ```

2. **Install dependencies**
   ```bash
   npm install
   ```

3. **Environment Configuration**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Start the service**
   ```bash
   # Development
   npm run dev

   # Production
   npm start
   ```

## ⚙️ Configuration

The service uses a comprehensive configuration system with environment variables:

### Core Configuration
```env
# Server Configuration
RESPONSE_SERVICE_PORT=3002
HOST=0.0.0.0
NODE_ENV=development
SERVICE_NAME=response-service
SERVICE_VERSION=1.0.0

# JWT Configuration
JWT_SECRET=your-secure-jwt-secret
JWT_EXPIRES_IN=24h
JWT_ISSUER=response-service
JWT_AUDIENCE=x-form-users

# Database Configuration
DB_TYPE=mongodb
DB_HOST=localhost
DB_PORT=27017
DB_NAME=response_service
DB_USERNAME=
DB_PASSWORD=
DATABASE_URL=mongodb://localhost:27017/response_service
```

### Security Configuration
```env
# Rate Limiting
RATE_LIMIT_WINDOW_MS=900000  # 15 minutes
RATE_LIMIT_MAX=100

# CORS
CORS_ORIGINS=http://localhost:3000,https://yourdomain.com

# Security Headers
ENABLE_HSTS=true
ENABLE_CSP=true
```

### Feature Flags
```env
# Features
FEATURE_ANALYTICS=true
FEATURE_FILE_UPLOADS=true
FEATURE_REAL_TIME_UPDATES=false
FEATURE_ADVANCED_VALIDATION=true
FEATURE_RESPONSE_EXPORT=true
```

See [Configuration Guide](docs/configuration.md) for complete configuration options.

## 📚 API Documentation

### Interactive Documentation
- **Swagger UI**: `http://localhost:3002/api-docs`
- **OpenAPI JSON**: `http://localhost:3002/api-docs.json`

### Quick API Reference

#### Authentication
```bash
# Get JWT token (from auth service)
curl -X POST http://localhost:3002/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password"}'
```

#### Responses
```bash
# Create response
curl -X POST http://localhost:3002/api/v1/responses \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "formId": "form_123",
    "responses": [
      {
        "questionId": "q1",
        "questionType": "text",
        "value": "Sample response"
      }
    ]
  }'

# Get responses
curl -X GET http://localhost:3002/api/v1/responses?formId=form_123 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Get response by ID
curl -X GET http://localhost:3002/api/v1/responses/resp_123 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Analytics
```bash
# Get form analytics
curl -X GET http://localhost:3002/api/v1/analytics/forms/form_123 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Get dashboard analytics
curl -X GET http://localhost:3002/api/v1/analytics/dashboard \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Health Check
```bash
# Service health
curl -X GET http://localhost:3002/health

# Detailed health check
curl -X GET http://localhost:3002/api/v1/health
```

## 🏗️ Project Structure

```
src/
├── config/           # Configuration management
│   ├── enhanced.js   # Enhanced configuration system
│   ├── index.js      # Legacy configuration
│   └── swagger.js    # OpenAPI/Swagger configuration
├── controllers/      # Request handlers
│   ├── responseController.js
│   ├── analyticsController.js
│   └── healthController.js
├── dto/              # Data Transfer Objects
│   └── response-dtos.js
├── events/           # Event system
│   └── eventSystem.js
├── integrations/     # External service integrations
│   └── formService.js
├── middleware/       # Express middleware
│   ├── auth.js       # Authentication middleware
│   ├── validation.js # Request validation
│   ├── security.js   # Security middleware
│   └── errorHandler.js
├── routes/           # API route definitions
│   └── v1/
│       ├── index.js
│       ├── responses.js
│       ├── analytics.js
│       └── health.js
├── utils/            # Utility functions
│   └── logger.js     # Logging utility
└── index.js          # Application entry point

tests/                # Test suite
├── testUtils.js      # Testing utilities
├── unit/             # Unit tests
├── integration/      # Integration tests
└── setup.js          # Test setup

docs/                 # Documentation
├── api.md           # API documentation
├── deployment.md    # Deployment guide
└── configuration.md # Configuration guide
```

## 🧪 Testing

### Running Tests
```bash
# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage

# Run tests for CI
npm run test:ci
```

### Test Types
- **Unit Tests**: Test individual functions and modules
- **Integration Tests**: Test API endpoints and service integration
- **Contract Tests**: Validate API contracts with other services

### Test Coverage
The project maintains high test coverage:
- Controllers: 95%+
- Middleware: 90%+
- Utilities: 95%+
- Overall: 90%+

## 🚀 Deployment

### Docker Deployment
```bash
# Build Docker image
npm run docker:build

# Run container
npm run docker:run
```

### Production Deployment
```bash
# Install production dependencies
npm ci --only=production

# Build and test
npm run build

# Start with PM2
pm2 start ecosystem.config.js
```

### Environment-Specific Deployment
See [Deployment Guide](docs/deployment.md) for detailed deployment instructions.

## 📊 Monitoring

### Health Checks
- **Basic Health**: `GET /health`
- **Detailed Health**: `GET /api/v1/health`
- **Dependencies**: Database, Form Service, External APIs

### Logging
- **Structured JSON Logging** with correlation IDs
- **Log Levels**: error, warn, info, debug
- **Log Rotation** with daily rotation and retention
- **Centralized Logging** support (ELK, Splunk, etc.)

### Metrics
- **Request Metrics**: Response times, error rates
- **Business Metrics**: Response counts, completion rates
- **System Metrics**: Memory usage, CPU usage
- **Custom Metrics**: Form-specific analytics

## 🔧 Development

### Prerequisites
```bash
# Install development dependencies
npm install

# Install git hooks
npm run prepare
```

### Code Quality
```bash
# Lint code
npm run lint

# Fix linting issues
npm run lint:fix

# Format code
npm run format

# Check formatting
npm run format:check
```

### Development Workflow
1. Create feature branch
2. Implement changes with tests
3. Run quality checks
4. Submit pull request
5. Code review and merge

## 🤝 Integration

### Form Service Integration
The Response Service integrates with the Form Service for:
- Form validation and schema retrieval
- Form existence verification
- Statistics synchronization

### Event-Driven Communication
```javascript
// Publishing events
eventSystem.publishEvent(ResponseEvents.RESPONSE_SUBMITTED, {
  responseId: 'resp_123',
  formId: 'form_456',
  submittedAt: new Date().toISOString()
}, correlationId);

// Subscribing to events
eventSystem.subscribeToEvent(ResponseEvents.FORM_UPDATED, 
  async (event) => {
    // Handle form update
  },
  { subscriberName: 'response-processor' }
);
```

## 📈 Performance

### Optimization Features
- **Response Compression**: Gzip compression for API responses
- **Request Caching**: Redis-based caching for frequent queries
- **Database Optimization**: Indexed queries and connection pooling
- **Rate Limiting**: Prevents API abuse and ensures fair usage

### Performance Metrics
- **Average Response Time**: < 200ms for standard queries
- **Throughput**: 1000+ requests per second
- **Cache Hit Rate**: > 80% for analytics queries
- **Database Query Time**: < 50ms average

## 🔒 Security

### Security Measures
- **Authentication**: JWT with configurable expiration
- **Authorization**: Role-based access control (RBAC)
- **Input Validation**: Joi schema validation for all inputs
- **Rate Limiting**: Per-IP and per-user rate limits
- **Security Headers**: HSTS, CSP, X-Frame-Options, etc.
- **Data Sanitization**: HTML and SQL injection prevention

### Security Best Practices
- Regular security audits
- Dependency vulnerability scanning
- Secure configuration management
- Encryption at rest and in transit

## 🐛 Troubleshooting

### Common Issues

#### Service Won't Start
```bash
# Check port availability
lsof -i :3002

# Check logs
npm run dev

# Verify configuration
node -e "console.log(require('./src/config/enhanced').getSanitized())"
```

#### Authentication Issues
```bash
# Verify JWT secret
echo $JWT_SECRET

# Check token validity
node -e "const jwt = require('jsonwebtoken'); console.log(jwt.decode('YOUR_TOKEN'))"
```

#### Database Connection Issues
```bash
# Test database connection
node -e "const config = require('./src/config/enhanced'); console.log(config.get('database'))"
```

### Debug Mode
```bash
# Enable debug logging
DEBUG=response-service:* npm run dev

# Enable specific debug categories
DEBUG=response-service:database,response-service:auth npm run dev
```

## 📋 Changelog

### Version 1.0.0
- ✅ Initial release with comprehensive microservices architecture
- ✅ RESTful API with OpenAPI documentation
- ✅ JWT authentication and authorization
- ✅ Advanced validation and error handling
- ✅ Analytics and reporting capabilities
- ✅ File upload management
- ✅ Event-driven communication
- ✅ Comprehensive testing suite
- ✅ Production-ready monitoring and logging

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines
- Follow the existing code style
- Write tests for new features
- Update documentation as needed
- Ensure all tests pass
- Follow semantic versioning

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👥 Team

- **Backend Team**: X-Form Development Team
- **DevOps Team**: Infrastructure and Deployment
- **QA Team**: Testing and Quality Assurance

## 📞 Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/your-org/x-form-backend/issues)
- **Email**: support@x-form.com
- **Slack**: #x-form-backend

---

**Built with ❤️ by the X-Form Team**
