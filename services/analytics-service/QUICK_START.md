# Analytics Service - Quick Start Guide

## 🚀 Get Started in 5 Minutes

### Prerequisites

- Python 3.8 or higher
- pip (Python package installer)
- Optional: Redis server for caching
- Optional: Google BigQuery for advanced analytics

### Step 1: Clone and Navigate

```bash
cd /Users/mir00r/Developer/kamkaiz/X-Form-Backend/services/analytics-service
```

### Step 2: Automated Setup

```bash
# Run the automated setup script
./setup.sh
```

This script will:
- Create virtual environment
- Install all dependencies
- Create configuration files
- Test the installation
- Generate API documentation

### Step 3: Manual Setup (Alternative)

If you prefer manual setup:

```bash
# Create virtual environment
python3 -m venv venv

# Activate virtual environment
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Create environment file
cp .env.example .env
```

### Step 4: Configure Environment

Edit the `.env` file with your settings:

```bash
nano .env
```

Minimum required configuration:
```env
JWT_SECRET_KEY=your-super-secret-jwt-key-change-in-production
PORT=8084
ENVIRONMENT=development
```

### Step 5: Start the Service

```bash
# Activate virtual environment (if not already active)
source venv/bin/activate

# Start the service
python main.py
```

Alternative start methods:
```bash
# Using uvicorn directly
uvicorn app.main:app --host 0.0.0.0 --port 8084 --reload

# Using the Python module
python -m uvicorn app.main:app --host 0.0.0.0 --port 8084 --reload
```

### Step 6: Access Swagger Documentation

Open your browser and navigate to:
- **Swagger UI**: http://localhost:8084/docs
- **ReDoc**: http://localhost:8084/redoc
- **Health Check**: http://localhost:8084/health

## ✅ Verification

### Check Service Health

```bash
curl http://localhost:8084/health
```

Expected response:
```json
{
  "service": "analytics",
  "status": "healthy",
  "version": "1.0.0",
  "environment": "development"
}
```

### Test Authentication

```bash
# This will return 401 Unauthorized (expected)
curl http://localhost:8084/analytics/test-form-id/summary
```

### View API Documentation

Visit http://localhost:8084/docs to see:
- ✅ All API endpoints documented
- ✅ Interactive "Try it out" functionality
- ✅ Request/response examples
- ✅ Authentication setup

## 🧪 Testing Endpoints

### 1. Health Check

```bash
curl -X GET "http://localhost:8084/health"
```

### 2. Service Information

```bash
curl -X GET "http://localhost:8084/"
```

### 3. OpenAPI Schema

```bash
curl -X GET "http://localhost:8084/openapi.json"
```

### 4. Protected Endpoint (with mock token)

```bash
curl -X GET "http://localhost:8084/analytics/550e8400-e29b-41d4-a716-446655440000/summary" \
  -H "Authorization: Bearer mock-jwt-token-for-testing"
```

## 📊 Swagger Features Included

### ✅ Comprehensive Documentation
- **API Overview**: Complete service description
- **Endpoint Documentation**: All 20+ endpoints documented
- **Request/Response Examples**: Real examples for every endpoint
- **Error Handling**: Complete error response documentation

### ✅ Interactive Features
- **Try It Out**: Test endpoints directly from Swagger UI
- **Authentication**: JWT token input for protected endpoints
- **Parameter Validation**: Real-time validation of inputs
- **Response Preview**: See actual API responses

### ✅ Industry Best Practices
- **OpenAPI 3.0**: Latest specification standard
- **Security Schemes**: JWT and API key authentication
- **Rate Limiting**: Documented rate limits for all endpoints
- **Status Codes**: Comprehensive HTTP status code coverage
- **Data Models**: Strongly typed request/response models

### ✅ Advanced Features
- **Real-time Streaming**: Server-Sent Events documentation
- **File Downloads**: Export and download endpoints
- **Background Tasks**: Async operation tracking
- **Webhooks**: Real-time notification setup

## 🛠️ Development Mode

### Hot Reload Development

```bash
# Start with hot reload
uvicorn app.main:app --host 0.0.0.0 --port 8084 --reload

# Or set environment variable
export ENVIRONMENT=development
python main.py
```

### Adding New Endpoints

1. **Add to router**: Create new endpoint in appropriate router file
2. **Add documentation**: Include comprehensive Swagger annotations
3. **Test**: Use Swagger UI to test the endpoint
4. **Validate**: Ensure proper error handling and responses

Example new endpoint:
```python
@router.get(
    "/example",
    summary="Example Endpoint",
    description="This is an example endpoint with full documentation",
    responses={
        200: {"description": "Success response"},
        400: {"description": "Bad request"}
    }
)
async def example_endpoint():
    return {"message": "Hello World"}
```

## 🐛 Troubleshooting

### Common Issues

1. **Port already in use**
   ```bash
   # Kill process using port 8084
   lsof -ti:8084 | xargs kill -9
   ```

2. **Virtual environment issues**
   ```bash
   # Remove and recreate
   rm -rf venv
   python3 -m venv venv
   source venv/bin/activate
   pip install -r requirements.txt
   ```

3. **Import errors**
   ```bash
   # Ensure you're in the correct directory and virtual environment
   pwd  # Should show analytics-service directory
   which python  # Should show venv/bin/python
   ```

4. **Permission errors**
   ```bash
   # Make setup script executable
   chmod +x setup.sh
   ```

### Check Logs

```bash
# View application logs
tail -f logs/app.log

# Or run with debug logging
DEBUG=true python main.py
```

### Validate Configuration

```bash
# Test configuration loading
python -c "from app.config import settings; print('Config loaded successfully')"
```

## 🔧 Configuration Options

### Minimum Configuration (.env)

```env
JWT_SECRET_KEY=your-secret-key
PORT=8084
ENVIRONMENT=development
```

### Full Configuration (.env)

```env
# Application
APP_NAME=Analytics Service
APP_VERSION=1.0.0
ENVIRONMENT=development
DEBUG=true

# Server
HOST=0.0.0.0
PORT=8084
WORKERS=1

# Authentication
JWT_SECRET_KEY=your-super-secret-jwt-key
JWT_ALGORITHM=HS256
JWT_EXPIRATION_HOURS=24

# Database (Optional)
BIGQUERY_PROJECT_ID=your-project-id
REDIS_HOST=localhost
REDIS_PORT=6379

# Logging
LOG_LEVEL=INFO

# CORS
CORS_ORIGINS=["http://localhost:3000"]
```

## 📈 Next Steps

After getting the service running:

1. **Explore the API**: Use Swagger UI to test all endpoints
2. **Set up Authentication**: Configure proper JWT authentication
3. **Connect to BigQuery**: Set up BigQuery integration for production data
4. **Configure Redis**: Set up Redis for caching and performance
5. **Deploy**: Prepare for production deployment

## 🎯 Production Deployment

### Environment Variables for Production

```env
ENVIRONMENT=production
DEBUG=false
JWT_SECRET_KEY=secure-production-key
BIGQUERY_PROJECT_ID=production-project
REDIS_HOST=production-redis-host
LOG_LEVEL=WARNING
```

### Docker Deployment

```bash
# Build Docker image
docker build -t analytics-service .

# Run container
docker run -p 8084:8084 --env-file .env analytics-service
```

## 📞 Support

If you encounter any issues:
1. Check this guide first
2. Review the comprehensive documentation at `/docs`
3. Check the application logs
4. Verify your configuration
5. Test with the provided curl examples

---

**🎉 Congratulations!** You now have a fully functional Analytics Service with comprehensive Swagger documentation running locally.
