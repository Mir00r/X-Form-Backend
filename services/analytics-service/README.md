# Analytics Service

Analytics and reporting service for X-Form Backend with BigQuery integration for large-scale data analysis.

## Features

### 📊 Core Analytics
- **Form Summary Analytics**: Response statistics, completion rates, and performance metrics
- **Question-Level Analytics**: Detailed analysis of individual question responses and patterns
- **Trend Analysis**: Time-series analysis with configurable periods (hour, day, week, month, quarter, year)
- **Comparative Analytics**: Side-by-side comparison of multiple forms

### 🗄️ Data Processing
- **BigQuery Integration**: Large-scale data warehousing and analytics with Google BigQuery
- **Real-time Data Ingestion**: Automatic processing of form responses and user interactions
- **Custom Query Support**: Flexible SQL queries for advanced analytics
- **Data Visualization**: Interactive charts and graphs using Plotly

### ⚡ Performance & Caching
- **Redis Caching**: High-performance caching layer with configurable TTL
- **Rate Limiting**: Request throttling to prevent abuse and ensure fair usage
- **Batch Processing**: Efficient bulk operations for large datasets
- **Query Optimization**: Optimized BigQuery queries for fast response times

### 🔐 Security & Authentication
- **JWT Authentication**: Secure token-based authentication
- **Role-based Access Control**: Fine-grained permissions for different user types
- **Rate Limiting**: Protection against abuse and DDoS attacks
- **CORS Support**: Secure cross-origin resource sharing

## API Endpoints

### Form Analytics
```
GET /analytics/{form_id}/summary
```
Get comprehensive analytics summary for a form including:
- Total responses, completion rates
- Response trends over time
- Performance metrics
- Interactive visualizations

**Query Parameters:**
- `start_date` (optional): Start date for analytics range
- `end_date` (optional): End date for analytics range
- `use_cache` (optional): Whether to use cached data (default: true)

### Question Analytics
```
GET /analytics/{form_id}/question/{question_id}
```
Get detailed analytics for a specific question including:
- Response distribution
- Answer patterns and trends
- Response and skip rates
- Question-specific visualizations

**Query Parameters:**
- `start_date` (optional): Start date for analytics range
- `end_date` (optional): End date for analytics range
- `question_type` (optional): Type of question for appropriate visualization
- `use_cache` (optional): Whether to use cached data (default: true)

### Trend Analysis
```
GET /analytics/{form_id}/trend
```
Get trend analysis for form responses over time including:
- Time-series data with configurable periods
- Growth rates and statistical analysis
- Trend visualizations
- Performance patterns

**Query Parameters:**
- `period` (optional): Time period (hour, day, week, month, quarter, year)
- `start_date` (optional): Start date for trend analysis
- `end_date` (optional): End date for trend analysis
- `use_cache` (optional): Whether to use cached data (default: true)

### Comparative Analytics
```
GET /analytics/compare
```
Compare analytics across multiple forms including:
- Side-by-side metrics comparison
- Comparative performance analysis
- Multi-form visualizations

**Query Parameters:**
- `form_ids` (required): List of form IDs to compare
- `metric` (optional): Metric to compare (default: response_count)
- `period` (optional): Time period for comparison
- `start_date` (optional): Start date for comparison
- `end_date` (optional): End date for comparison

### Cache Management
```
POST /analytics/{form_id}/cache/invalidate
```
Invalidate analytics cache for a form or specific question.

### Service Health
```
GET /analytics/health
```
Check the health status of all analytics service components.

```
GET /analytics/cache/stats
```
Get cache performance statistics and metrics.

## Configuration

### Environment Variables

#### Service Configuration
```bash
ENVIRONMENT=development          # Environment (development/staging/production)
HOST=0.0.0.0                    # Service host
PORT=8080                       # Service port
LOG_LEVEL=INFO                  # Logging level
```

#### BigQuery Configuration
```bash
BIGQUERY_PROJECT_ID=your-project-id           # GCP Project ID
BIGQUERY_DATASET_ID=x_form_analytics          # BigQuery dataset
BIGQUERY_LOCATION=US                          # BigQuery location
GOOGLE_APPLICATION_CREDENTIALS=/path/to/key   # GCP service account key
```

#### Redis Configuration
```bash
REDIS_HOST=localhost            # Redis host
REDIS_PORT=6379                # Redis port
REDIS_PASSWORD=                # Redis password (optional)
REDIS_DB=0                     # Redis database number
```

#### Cache Configuration
```bash
CACHE_TTL=3600                 # Cache TTL in seconds (1 hour)
CACHE_PREFIX=analytics         # Cache key prefix
ENABLE_CACHE=true              # Enable/disable caching
```

#### Authentication Configuration
```bash
JWT_SECRET=your-super-secret-jwt-key    # JWT signing secret
JWT_ALGORITHM=HS256                     # JWT algorithm
JWT_EXPIRATION_HOURS=24                 # Token expiration time
```

#### Rate Limiting
```bash
ENABLE_RATE_LIMITING=true      # Enable/disable rate limiting
```

#### CORS Configuration
```bash
CORS_ORIGINS=["http://localhost:3000"]     # Allowed origins
CORS_METHODS=["GET","POST","PUT","DELETE"] # Allowed methods
CORS_HEADERS=["*"]                         # Allowed headers
```

## Development Setup

### Prerequisites
- Python 3.11+
- Redis 6.0+
- Google Cloud Platform account with BigQuery enabled
- Docker and Docker Compose (optional)

### Local Development

1. **Clone and setup**:
```bash
cd services/analytics-service
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install -r requirements.txt
```

2. **Configure environment**:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. **Setup BigQuery**:
- Create a GCP project and enable BigQuery API
- Create a service account with BigQuery permissions
- Download the service account key JSON file
- Set GOOGLE_APPLICATION_CREDENTIALS environment variable

4. **Setup Redis**:
```bash
# Using Docker
docker run -d --name redis -p 6379:6379 redis:7-alpine

# Or install locally (macOS)
brew install redis
redis-server
```

5. **Run the service**:
```bash
python -m uvicorn app.main:app --reload --host 0.0.0.0 --port 8080
```

### Docker Development

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f analytics-service

# Stop services
docker-compose down
```

### Testing

```bash
# Install test dependencies
pip install pytest pytest-asyncio httpx

# Run tests
pytest tests/ -v

# Run with coverage
pytest tests/ --cov=app --cov-report=html
```

## BigQuery Schema

The service uses the following BigQuery tables:

### Responses Table
```sql
CREATE TABLE responses (
  response_id STRING NOT NULL,
  form_id STRING NOT NULL,
  user_id STRING,
  responses JSON NOT NULL,
  submitted_at TIMESTAMP NOT NULL,
  completion_time_seconds FLOAT64,
  user_agent STRING,
  ip_address STRING,
  metadata JSON
);
```

### Forms Table
```sql
CREATE TABLE forms (
  form_id STRING NOT NULL,
  title STRING NOT NULL,
  description STRING,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP,
  metadata JSON
);
```

## Monitoring and Logging

### Health Checks
- Service health endpoint: `/health`
- Component health: `/analytics/health`
- Cache statistics: `/analytics/cache/stats`

### Logging
- Structured JSON logging
- Request/response logging
- Error tracking and monitoring
- Performance metrics

### Metrics
- Request count and duration
- Cache hit/miss rates
- BigQuery query performance
- Error rates by endpoint

## Production Deployment

### Docker Deployment
```bash
# Build production image
docker build -t analytics-service:latest .

# Run with production config
docker run -d \
  --name analytics-service \
  -p 8080:8080 \
  -e ENVIRONMENT=production \
  -e BIGQUERY_PROJECT_ID=your-project \
  -v /path/to/gcp-key.json:/app/credentials/gcp-key.json:ro \
  analytics-service:latest
```

### Environment Considerations
- Use managed Redis (Cloud Memorystore, ElastiCache)
- Set up BigQuery datasets with appropriate permissions
- Configure monitoring and alerting
- Enable HTTPS with proper TLS certificates
- Set up log aggregation and monitoring

## Security Considerations

- Store sensitive credentials securely (GCP key, JWT secret)
- Use HTTPS in production
- Implement proper CORS policies
- Enable rate limiting
- Monitor for suspicious activity
- Regular security updates

## Performance Optimization

- Use BigQuery partitioning for large tables
- Implement efficient caching strategies
- Optimize SQL queries for BigQuery
- Use connection pooling
- Monitor and tune cache TTL values
- Implement query result pagination for large datasets

## Contributing

1. Follow Python PEP 8 style guidelines
2. Add tests for new features
3. Update documentation
4. Ensure proper error handling
5. Add logging for debugging

## License

This project is part of the X-Form Backend system.
