# Analytics Service - Comprehensive Swagger Documentation

## Overview

The X-Form Analytics Service provides comprehensive analytics, reporting, and insights for form responses with BigQuery integration. This service follows industry best practices for API design and documentation.

## 🚀 Features

### Core Analytics
- **Real-time Analytics**: Instant insights on form performance and user behavior
- **Advanced Reporting**: Detailed reports with multiple export formats (CSV, Excel, JSON, PDF)
- **Data Visualization**: Interactive charts and graphs for data representation
- **Question Analytics**: Deep dive into individual question performance and patterns

### AI-Powered Insights
- **Machine Learning**: AI-powered insights and recommendations for form optimization
- **Predictive Analytics**: Forecast future form performance and trends
- **Anomaly Detection**: Automatically identify unusual patterns in response data
- **Sentiment Analysis**: Analyze text responses for sentiment and themes

### Real-time Features
- **Live Streaming**: Real-time analytics data streaming with Server-Sent Events
- **Event Streaming**: Monitor form events as they occur
- **Webhooks**: Configure webhooks to receive analytics updates
- **Push Notifications**: Real-time alerts for important metrics

### Data Management
- **BigQuery Integration**: Scalable analytics with Google BigQuery
- **Redis Caching**: High-performance caching for optimal response times
- **Data Export**: Multiple export formats with advanced filtering
- **Report Generation**: Automated and custom report generation

## 📊 API Endpoints

### Analytics Endpoints (`/analytics`)

| Method | Endpoint | Description | Rate Limit |
|--------|----------|-------------|------------|
| `GET` | `/{form_id}/summary` | Get comprehensive form analytics summary | 100/hour |
| `GET` | `/{form_id}/question/{question_id}` | Get detailed question analytics | 200/hour |
| `GET` | `/{form_id}/trend` | Get trend analysis over time | 50/hour |
| `GET` | `/compare` | Compare multiple forms or time periods | 30/hour |
| `POST` | `/{form_id}/cache/invalidate` | Invalidate analytics cache | 10/hour |

### Reports Endpoints (`/reports`)

| Method | Endpoint | Description | Rate Limit |
|--------|----------|-------------|------------|
| `POST` | `/{form_id}/export` | Export form data in various formats | 10/hour |
| `GET` | `/{form_id}/export/{export_id}/status` | Check export status | 100/hour |
| `GET` | `/download/{export_id}` | Download completed export | 50/hour |
| `POST` | `/{form_id}/generate` | Generate custom analytics report | 5/hour |
| `GET` | `/templates` | List available report templates | 50/hour |

### Insights Endpoints (`/insights`)

| Method | Endpoint | Description | Rate Limit |
|--------|----------|-------------|------------|
| `GET` | `/{form_id}` | Get AI-powered form insights | 20/hour |
| `GET` | `/{form_id}/recommendations` | Get optimization recommendations | 30/hour |
| `POST` | `/{form_id}/feedback` | Submit feedback on insights | 100/hour |

### Streaming Endpoints (`/streaming`)

| Method | Endpoint | Description | Rate Limit |
|--------|----------|-------------|------------|
| `GET` | `/{form_id}/live` | Stream live analytics data | 50/hour |
| `GET` | `/{form_id}/events` | Stream form events in real-time | 50/hour |
| `POST` | `/{form_id}/webhooks` | Configure analytics webhooks | 10/hour |

### System Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Service health check |
| `GET` | `/` | Service information |
| `GET` | `/docs` | Swagger UI documentation |
| `GET` | `/redoc` | ReDoc documentation |
| `GET` | `/openapi.json` | OpenAPI specification |

## 🔒 Authentication

The API uses JWT Bearer token authentication:

```bash
Authorization: Bearer <your-jwt-token>
```

Alternative API key authentication is also supported:

```bash
X-API-Key: <your-api-key>
```

## 📈 Rate Limiting

Rate limits are enforced per user and endpoint:

- **Analytics**: 100-200 requests per hour
- **Reports**: 5-50 requests per hour  
- **Insights**: 20-100 requests per hour
- **Streaming**: 50 requests per hour

Rate limit headers are included in responses:
- `X-RateLimit-Limit`: Rate limit window size
- `X-RateLimit-Remaining`: Remaining requests
- `X-RateLimit-Reset`: Reset time

## 📝 Request/Response Format

### Standard Response Format

```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {
    "form_id": "550e8400-e29b-41d4-a716-446655440000",
    "total_responses": 1523,
    "completion_rate": 84.6
  },
  "timestamp": "2025-09-06T12:00:00Z",
  "request_id": "req_abc123def456"
}
```

### Error Response Format

```json
{
  "success": false,
  "error": "validation_error",
  "message": "Invalid form ID format",
  "details": {
    "field": "form_id",
    "expected": "UUID format",
    "received": "invalid_string"
  },
  "timestamp": "2025-09-06T12:00:00Z",
  "request_id": "req_abc123def456"
}
```

## 🔧 Configuration

### Environment Variables

```bash
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

# BigQuery
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
BIGQUERY_PROJECT_ID=your-project-id
BIGQUERY_DATASET_ID=xform_analytics

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=optional-password
REDIS_DB=0
CACHE_TTL=3600

# CORS
CORS_ORIGINS=["http://localhost:3000"]
CORS_METHODS=["GET", "POST", "PUT", "DELETE"]
CORS_HEADERS=["*"]
```

## 🚀 Quick Start

### 1. Setup and Installation

```bash
# Clone the repository
cd services/analytics-service

# Run setup script
./setup.sh

# Or manual setup:
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### 2. Configuration

```bash
# Copy example environment file
cp .env.example .env

# Edit configuration
nano .env
```

### 3. Run the Service

```bash
# Activate virtual environment
source venv/bin/activate

# Start the service
python main.py

# Or using uvicorn directly
uvicorn app.main:app --host 0.0.0.0 --port 8084 --reload
```

### 4. Access Documentation

- **Swagger UI**: http://localhost:8084/docs
- **ReDoc**: http://localhost:8084/redoc
- **OpenAPI JSON**: http://localhost:8084/openapi.json

## 🧪 Testing the API

### Health Check

```bash
curl http://localhost:8084/health
```

### Get Form Analytics

```bash
curl -X GET "http://localhost:8084/analytics/550e8400-e29b-41d4-a716-446655440000/summary" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Export Form Data

```bash
curl -X POST "http://localhost:8084/reports/550e8400-e29b-41d4-a716-446655440000/export" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "format": "csv",
    "include_metadata": true,
    "date_range": {
      "start": "2025-09-01T00:00:00Z",
      "end": "2025-09-06T23:59:59Z"
    }
  }'
```

### Stream Live Analytics

```bash
curl -N -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  "http://localhost:8084/streaming/550e8400-e29b-41d4-a716-446655440000/live?interval=5"
```

## 📊 Data Models

### Form Summary Model

```json
{
  "form_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Customer Satisfaction Survey",
  "total_responses": 1523,
  "completed_responses": 1289,
  "partial_responses": 234,
  "average_completion_time": 145.7,
  "completion_rate": 84.6,
  "unique_respondents": 1456,
  "response_rate_trend": [
    {"date": "2025-09-05", "responses": 45, "completion_rate": 86.7},
    {"date": "2025-09-06", "responses": 52, "completion_rate": 84.6}
  ]
}
```

### Question Analytics Model

```json
{
  "question_id": "q_rating_satisfaction",
  "question_type": "rating",
  "total_responses": 1289,
  "response_rate": 94.8,
  "answer_distribution": {
    "1": 45, "2": 89, "3": 234, "4": 567, "5": 354
  },
  "statistics": {
    "mean": 4.1,
    "median": 4.0,
    "mode": 4,
    "std_dev": 1.2
  }
}
```

## 🔍 Advanced Features

### AI-Powered Insights

The service includes machine learning capabilities for:

- **Trend Detection**: Identify patterns and trends in response data
- **Anomaly Detection**: Detect unusual patterns or outliers
- **Predictive Analytics**: Forecast future form performance
- **Optimization Recommendations**: Suggest improvements for better completion rates

### Real-time Streaming

Server-Sent Events (SSE) for real-time data:

```javascript
const eventSource = new EventSource('/streaming/form-id/live');
eventSource.onmessage = function(event) {
  const data = JSON.parse(event.data);
  updateDashboard(data);
};
```

### Custom Reports

Generate custom reports with specific parameters:

```json
{
  "report_type": "comprehensive_analytics",
  "parameters": {
    "include_charts": true,
    "include_raw_data": false,
    "group_by": "date"
  },
  "delivery_method": "email",
  "schedule": "weekly"
}
```

## 🛠️ Development

### Project Structure

```
analytics-service/
├── app/
│   ├── api/
│   │   ├── analytics.py      # Analytics endpoints
│   │   ├── reports.py        # Report generation endpoints
│   │   ├── insights.py       # AI insights endpoints
│   │   └── streaming.py      # Real-time streaming endpoints
│   ├── models/
│   │   └── analytics.py      # Pydantic models
│   ├── services/
│   │   └── analytics_service.py
│   ├── utils/
│   │   ├── auth.py          # Authentication utilities
│   │   └── rate_limiter.py  # Rate limiting
│   ├── config.py            # Configuration
│   └── main.py              # FastAPI application
├── docs/                    # Generated documentation
├── tests/                   # Test suite
├── requirements.txt         # Dependencies
├── setup.sh                 # Setup script
└── README.md               # This file
```

### Adding New Endpoints

1. Create endpoint in appropriate router file
2. Add comprehensive Swagger documentation
3. Update models if needed
4. Add tests
5. Update this documentation

### Code Style

- Follow PEP 8 guidelines
- Use type hints for all functions
- Add comprehensive docstrings
- Include Swagger documentation for all endpoints

## 📚 Additional Resources

- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [Swagger/OpenAPI Specification](https://swagger.io/specification/)
- [Google BigQuery Documentation](https://cloud.google.com/bigquery/docs)
- [Redis Documentation](https://redis.io/documentation)

## 🆘 Support

For issues and questions:
1. Check the Swagger documentation at `/docs`
2. Review the application logs
3. Test endpoints using the provided examples
4. Check the GitHub repository for known issues

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.
