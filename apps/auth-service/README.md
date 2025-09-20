# Auth Service - Industry Standard Authentication & User Management

A comprehensive, enterprise-grade authentication service built with Node.js, Express, and PostgreSQL, implementing modern security practices and industry standards.

## 🚀 Features

### Core Authentication
- **JWT Access Tokens** with short expiration (15 minutes)
- **Refresh Tokens** with secure HTTP-only cookies
- **Password Security** with bcrypt hashing and strength requirements
- **Account Lockout** protection against brute force attacks
- **Rate Limiting** with configurable thresholds

### OAuth Integration
- **Google OAuth 2.0** with account linking
- **Extensible OAuth** framework for additional providers

### Security Features
- **Comprehensive Audit Logging** for all auth events
- **IP and User Agent Tracking** for session security
- **CORS Protection** with configurable origins
- **Helmet Security Headers** for XSS and clickjacking protection
- **Input Validation** with Joi schemas
- **SQL Injection Protection** with parameterized queries

### User Management
- **Email Verification** workflow
- **Password Reset** with secure tokens
- **Profile Management** with validation
- **Account Status** management (active/inactive)
- **Session Management** with multiple device support

## 📋 Prerequisites

- **Node.js** >= 18.0.0
- **PostgreSQL** >= 13.0
- **Redis** (optional, for advanced rate limiting)

## 🛠️ Installation

1. **Clone the repository**
   ```bash
   cd services/auth-service
   ```

2. **Install dependencies**
   ```bash
   npm install
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Set up PostgreSQL database**
   ```bash
   # Create database
   createdb xform_db
   
   # Run database schema
   psql -d xform_db -f schema.sql
   ```

5. **Start the service**
   ```bash
   # Development
   npm run dev
   
   # Production
   npm start
   ```

## 📊 API Endpoints

### Health Check
```http
GET /health
```

### Authentication Endpoints

#### User Registration
```http
POST /auth/signup
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "firstName": "John",
  "lastName": "Doe",
  "acceptTerms": true
}
```

#### User Login
```http
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "rememberMe": false
}
```

#### Token Refresh
```http
POST /auth/refresh
Content-Type: application/json

{
  "refreshToken": "your-refresh-token"
}
```

#### User Logout
```http
POST /auth/logout
Content-Type: application/json

{
  "refreshToken": "your-refresh-token"
}
```

#### Get User Profile
```http
GET /auth/me
Authorization: Bearer your-access-token
```

#### Update Profile
```http
PUT /auth/profile
Authorization: Bearer your-access-token
Content-Type: application/json

{
  "firstName": "John",
  "lastName": "Smith",
  "avatarUrl": "https://example.com/avatar.jpg"
}
```

#### Change Password
```http
POST /auth/change-password
Authorization: Bearer your-access-token
Content-Type: application/json

{
  "currentPassword": "OldPass123!",
  "newPassword": "NewSecurePass123!",
  "confirmPassword": "NewSecurePass123!"
}
```

### Google OAuth
```http
GET /auth/google
# Redirects to Google OAuth consent screen

GET /auth/google/callback
# Google OAuth callback endpoint
```

## 🔒 Security Implementation

### Password Requirements
- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one number
- At least one special character (@$!%*?&)

### Rate Limiting
- **General API**: 100 requests per 15 minutes
- **Auth Endpoints**: 5 attempts per 15 minutes
- **Sensitive Operations**: 3 attempts per hour

### Account Security
- **Lockout**: Account locked after 5 failed login attempts
- **Lockout Duration**: 1 hour (configurable)
- **Session Management**: Automatic cleanup of expired sessions
- **Token Security**: JWT with RS256 signing (recommended for production)

### Audit Logging
All authentication events are logged with:
- User ID and email
- Event type (login, logout, failed_login, etc.)
- IP address and User Agent
- Timestamp and outcome
- Additional metadata

## 🌍 Environment Variables

See `.env.example` for all available configuration options.

### Required Variables
```bash
DATABASE_URL=postgresql://user:pass@localhost:5432/xform_db
JWT_SECRET=your-jwt-secret-minimum-32-characters
JWT_REFRESH_SECRET=your-refresh-secret-different-from-jwt
```

### Optional Variables
```bash
GOOGLE_CLIENT_ID=your-google-oauth-client-id
GOOGLE_CLIENT_SECRET=your-google-oauth-client-secret
FRONTEND_URL=http://localhost:3000
REDIS_URL=redis://localhost:6379
```

## 🧪 Testing

```bash
# Run tests
npm test

# Run tests with coverage
npm run test:coverage

# Run tests in watch mode
npm run test:watch
```

## 🔧 Development

```bash
# Start development server with auto-reload
npm run dev

# Lint code
npm run lint

# Fix linting issues
npm run lint:fix
```

## 🐳 Docker Deployment

```dockerfile
# Build image
docker build -t auth-service .

# Run container
docker run -p 3001:3001 \
  -e DATABASE_URL=postgresql://user:pass@host:5432/db \
  -e JWT_SECRET=your-secret \
  -e JWT_REFRESH_SECRET=your-refresh-secret \
  auth-service
```

## 📈 Monitoring & Logging

### Health Checks
- **Health Endpoint**: `/health`
- **Database Connection**: Checked on startup
- **Memory Usage**: Logged periodically

### Logging Levels
- **Error**: Authentication failures, security events
- **Warn**: Rate limit violations, account lockouts
- **Info**: Successful logins, user registrations
- **Debug**: Token operations, middleware execution

### Metrics (Recommended)
- Request rate and response times
- Authentication success/failure rates
- Active user sessions
- Database connection health

## 🔐 Production Recommendations

### Security
1. **Use HTTPS only** in production
2. **Set secure JWT secrets** (minimum 32 characters)
3. **Configure CORS** for your domain only
4. **Enable rate limiting** with Redis
5. **Use environment-specific** database credentials
6. **Implement log monitoring** and alerting

### Performance
1. **Use connection pooling** for PostgreSQL
2. **Implement Redis caching** for rate limiting
3. **Configure proper indexes** on database tables
4. **Use CDN** for static assets
5. **Implement database monitoring**

### Monitoring
1. **Set up application metrics** (Prometheus/Grafana)
2. **Configure error tracking** (Sentry)
3. **Implement log aggregation** (ELK Stack)
4. **Set up uptime monitoring**
5. **Configure database monitoring**

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support and questions:
- Create an issue in the repository
- Contact the development team
- Check the documentation wiki

---

**Built with ❤️ by the X-Form Team**
