# X-Form Realtime Service - Complete API Documentation

## 🚀 Quick Start Guide

### Prerequisites
- Node.js 18.0.0 or higher
- npm or yarn package manager
- Redis (optional, for horizontal scaling)

### Installation & Setup

1. **Clone and Navigate**
   ```bash
   cd services/realtime-service
   ```

2. **Install Dependencies**
   ```bash
   npm install
   ```

3. **Environment Configuration**
   ```bash
   cp .env.example .env
   ```
   
   Edit `.env` file with your configuration:
   ```env
   REALTIME_SERVICE_PORT=8002
   NODE_ENV=development
   ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
   JWT_SECRET=your-super-secure-jwt-secret
   REQUIRE_AUTH=false
   ```

4. **Start the Service**
   ```bash
   # Development mode with auto-reload
   npm run dev
   
   # Production mode
   npm start
   ```

5. **Verify Installation**
   - Health Check: http://localhost:8002/health
   - API Documentation: http://localhost:8002/api-docs
   - Service Info: http://localhost:8002/

## 📖 API Documentation

### Base URL
- **Local Development**: `http://localhost:8002`
- **Swagger UI**: `http://localhost:8002/api-docs`

### Authentication
The service supports optional JWT authentication. Set `REQUIRE_AUTH=true` in your environment to enforce authentication.

**WebSocket Authentication:**
```javascript
const socket = io('ws://localhost:8002', {
  auth: {
    token: 'your-jwt-token-here'
  }
});
```

**REST API Authentication:**
```bash
curl -H "Authorization: Bearer your-jwt-token" \
     http://localhost:8002/ws/connections
```

## 🔌 WebSocket API

### Connection
```javascript
const io = require('socket.io-client');

const socket = io('ws://localhost:8002', {
  transports: ['websocket', 'polling'],
  timeout: 20000
});

socket.on('connect', () => {
  console.log('Connected:', socket.id);
});
```

### Events Reference

#### Client → Server Events

| Event | Description | Payload |
|-------|-------------|---------|
| `form:subscribe` | Subscribe to form updates | `{ formId: "form123" }` |
| `form:unsubscribe` | Unsubscribe from form | `{ formId: "form123" }` |
| `response:new` | Submit new response | `{ formId: "form123", responseId: "resp456", data: {...} }` |
| `response:update` | Update existing response | `{ formId: "form123", responseId: "resp456", data: {...} }` |
| `user:typing` | Indicate user is typing | `{ formId: "form123", field: "question1" }` |
| `user:stopped_typing` | User stopped typing | `{ formId: "form123", field: "question1" }` |
| `ping` | Health check ping | `{}` |
| `rooms:list` | Get subscribed rooms | `{}` |

#### Server → Client Events

| Event | Description | Payload |
|-------|-------------|---------|
| `form:subscribed` | Subscription confirmed | `{ formId, roomName, timestamp }` |
| `form:unsubscribed` | Unsubscription confirmed | `{ formId, roomName, timestamp }` |
| `response:update` | New response received | `{ formId, responseId, data, userId, timestamp }` |
| `response:updated` | Response was updated | `{ formId, responseId, data, userId, timestamp }` |
| `form:published` | Form was published | `{ formId, publishedBy, timestamp }` |
| `form:closed` | Form was closed | `{ formId, closedBy, timestamp }` |
| `user:typing` | Another user is typing | `{ formId, field, userId, timestamp }` |
| `user:stopped_typing` | User stopped typing | `{ formId, field, userId, timestamp }` |
| `user:disconnected` | User disconnected | `{ socketId, userId, reason, timestamp }` |
| `pong` | Response to ping | `{ timestamp, socketId }` |
| `error` | Error occurred | `{ message, timestamp }` |

### Example Usage

#### Form Subscription
```javascript
// Subscribe to form updates
socket.emit('form:subscribe', 'form123');

socket.on('form:subscribed', (data) => {
  console.log('Subscribed to form:', data.formId);
});

// Listen for new responses
socket.on('response:update', (data) => {
  console.log('New response:', data);
  updateFormUI(data);
});
```

#### Submit Response
```javascript
// Submit a new response
socket.emit('response:new', {
  formId: 'form123',
  responseId: 'resp456',
  data: {
    question1: 'John Doe',
    question2: 'john@example.com',
    question3: 'Great form!'
  }
});
```

#### Real-time Collaboration
```javascript
// Show typing indicator
socket.emit('user:typing', {
  formId: 'form123',
  field: 'question1'
});

// Listen for other users typing
socket.on('user:typing', (data) => {
  showTypingIndicator(data.userId, data.field);
});

// Stop typing
setTimeout(() => {
  socket.emit('user:stopped_typing', {
    formId: 'form123',
    field: 'question1'
  });
}, 2000);
```

## 🌐 REST API

### Health Endpoints

#### Basic Health Check
```bash
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00.000Z",
  "uptime": 3600,
  "service": "realtime-service",
  "version": "1.0.0",
  "connections": {
    "active": 25,
    "total": 150
  },
  "memory": {
    "used": 128,
    "total": 256,
    "external": 16
  },
  "environment": "development"
}
```

#### Detailed Health Check
```bash
GET /health/detailed
```

#### Kubernetes Probes
```bash
GET /health/live     # Liveness probe
GET /health/ready    # Readiness probe
```

### WebSocket Management

#### Get WebSocket Info
```bash
GET /ws/info
```

#### List Active Connections
```bash
GET /ws/connections
```

#### Get Active Rooms
```bash
GET /ws/rooms
```

#### Broadcast to Room (Admin)
```bash
POST /ws/broadcast
Authorization: Bearer <token>
Content-Type: application/json

{
  "room": "form:12345",
  "event": "admin:announcement",
  "data": {
    "message": "System maintenance in 10 minutes",
    "priority": "high"
  }
}
```

#### Disconnect Socket (Admin)
```bash
POST /ws/disconnect/{socketId}
Authorization: Bearer <token>
```

### Event Management

#### Send Form Notification
```bash
POST /events/form/{formId}/notify
Content-Type: application/json

{
  "event": "form:updated",
  "data": {
    "title": "New Form Title",
    "updatedBy": "admin"
  },
  "urgent": false
}
```

#### Broadcast New Response
```bash
POST /events/form/{formId}/response
Content-Type: application/json

{
  "responseId": "resp456",
  "data": {
    "question1": "John Doe",
    "question2": "john@example.com"
  },
  "metadata": {
    "submittedAt": "2024-01-15T10:30:00.000Z",
    "submittedBy": "user123"
  }
}
```

#### Update Form Status
```bash
POST /events/form/{formId}/status
Content-Type: application/json

{
  "status": "published",
  "reason": "Form is ready for responses",
  "updatedBy": "admin123"
}
```

#### Global Broadcast (Admin)
```bash
POST /events/broadcast
Authorization: Bearer <token>
Content-Type: application/json

{
  "event": "system:maintenance",
  "data": {
    "message": "System will be down for maintenance in 10 minutes"
  },
  "priority": "high"
}
```

### Monitoring & Metrics

#### Connection Statistics
```bash
GET /stats
```

#### Event Metrics
```bash
GET /events/metrics
```

## 🧪 Testing

### Unit Tests
```bash
npm test
```

### Integration Tests
```bash
npm run test:integration
```

### Load Testing
```bash
# Install artillery for load testing
npm install -g artillery

# Run load test
artillery run tests/load-test.yml
```

### Manual Testing with curl

#### Test Health Endpoint
```bash
curl -s http://localhost:8002/health | jq .
```

#### Test WebSocket Info
```bash
curl -s http://localhost:8002/ws/info | jq .
```

#### Test Form Notification
```bash
curl -X POST http://localhost:8002/events/form/test123/notify \
  -H "Content-Type: application/json" \
  -d '{
    "event": "form:updated",
    "data": {"message": "Test notification"},
    "urgent": false
  }' | jq .
```

## 🚀 Deployment

### Docker Deployment
```bash
# Build image
docker build -t xform-realtime-service .

# Run container
docker run -p 8002:8002 \
  -e NODE_ENV=production \
  -e REALTIME_SERVICE_PORT=8002 \
  xform-realtime-service
```

### Docker Compose
```yaml
version: '3.8'
services:
  realtime-service:
    build: .
    ports:
      - "8002:8002"
    environment:
      - NODE_ENV=production
      - REALTIME_SERVICE_PORT=8002
      - ALLOWED_ORIGINS=https://yourapp.com
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - redis
  
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: realtime-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: realtime-service
  template:
    metadata:
      labels:
        app: realtime-service
    spec:
      containers:
      - name: realtime-service
        image: xform-realtime-service:latest
        ports:
        - containerPort: 8002
        env:
        - name: NODE_ENV
          value: "production"
        - name: REALTIME_SERVICE_PORT
          value: "8002"
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8002
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8002
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: realtime-service-svc
spec:
  selector:
    app: realtime-service
  ports:
  - port: 8002
    targetPort: 8002
  type: LoadBalancer
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|-----------|
| `REALTIME_SERVICE_PORT` | Service port | `8002` | No |
| `NODE_ENV` | Environment | `development` | No |
| `ALLOWED_ORIGINS` | CORS origins | `http://localhost:3000` | No |
| `JWT_SECRET` | JWT secret key | - | If auth enabled |
| `REQUIRE_AUTH` | Require authentication | `false` | No |
| `REDIS_HOST` | Redis host | `localhost` | If using Redis |
| `REDIS_PORT` | Redis port | `6379` | If using Redis |
| `REDIS_PASSWORD` | Redis password | - | If using Redis |
| `LOG_LEVEL` | Logging level | `info` | No |
| `MAX_CONNECTIONS` | Max connections | `1000` | No |
| `EVENT_RATE_LIMIT` | Events per second limit | `100` | No |

### Performance Tuning

#### Connection Limits
```javascript
// In production, adjust these values based on your infrastructure
const io = socketIo(server, {
  maxHttpBufferSize: 1e6,  // 1MB
  pingTimeout: 60000,      // 60 seconds
  pingInterval: 25000,     // 25 seconds
  upgradeTimeout: 30000,   // 30 seconds
  allowUpgrades: true,
  transports: ['websocket', 'polling']
});
```

#### Memory Optimization
```javascript
// Monitor memory usage
setInterval(() => {
  const usage = process.memoryUsage();
  if (usage.heapUsed > 500 * 1024 * 1024) { // 500MB
    console.warn('High memory usage detected:', usage);
  }
}, 30000);
```

## 🐛 Troubleshooting

### Common Issues

#### 1. Connection Refused
```bash
# Check if service is running
curl http://localhost:8002/health

# Check logs
npm run dev

# Verify port is not in use
lsof -i :8002
```

#### 2. CORS Errors
```javascript
// Update ALLOWED_ORIGINS in .env
ALLOWED_ORIGINS=http://localhost:3000,https://yourapp.com
```

#### 3. Authentication Issues
```javascript
// Verify JWT token format
const token = jwt.sign({ userId: 'user123' }, process.env.JWT_SECRET);

// Check token in WebSocket connection
const socket = io('ws://localhost:8002', {
  auth: { token }
});
```

#### 4. High Memory Usage
```javascript
// Enable garbage collection logs
node --expose-gc --trace-gc src/index.js

// Monitor connections
setInterval(() => {
  console.log('Active connections:', io.engine.clientsCount);
}, 10000);
```

### Debug Mode
```bash
DEBUG=socket.io:* npm run dev
```

### Logging
```javascript
// Custom logging configuration
const winston = require('winston');

const logger = winston.createLogger({
  level: process.env.LOG_LEVEL || 'info',
  format: winston.format.combine(
    winston.format.timestamp(),
    winston.format.json()
  ),
  transports: [
    new winston.transports.Console(),
    new winston.transports.File({ filename: 'realtime-service.log' })
  ]
});
```

## 📊 Monitoring & Observability

### Health Checks
- **Liveness**: `/health/live` - Basic service availability
- **Readiness**: `/health/ready` - Service ready to accept traffic
- **Detailed**: `/health/detailed` - Comprehensive health information

### Metrics
- **Connection Metrics**: Active/total connections, rooms, subscriptions
- **Performance Metrics**: Events per second, memory usage, uptime
- **Error Metrics**: Connection errors, authentication failures

### Alerting
Set up alerts for:
- High memory usage (>80% of available)
- High connection count (>90% of max)
- Service downtime
- Authentication failures
- WebSocket connection errors

## 🔒 Security Best Practices

### Authentication
- Always use HTTPS in production
- Implement proper JWT validation
- Use strong, rotating secrets
- Implement rate limiting

### CORS Configuration
```javascript
// Restrictive CORS for production
app.use(cors({
  origin: ['https://yourapp.com'],
  credentials: true,
  methods: ['GET', 'POST'],
  allowedHeaders: ['Content-Type', 'Authorization']
}));
```

### Security Headers
```javascript
// Enhanced security headers
app.use(helmet({
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      styleSrc: ["'self'", "'unsafe-inline'"],
      scriptSrc: ["'self'"],
      imgSrc: ["'self'", "data:", "https:"],
    },
  },
  hsts: {
    maxAge: 31536000,
    includeSubDomains: true,
    preload: true
  }
}));
```

## 📝 API Changelog

### v1.0.0 (Current)
- Initial release with complete WebSocket and REST API
- JWT authentication support
- Real-time form subscriptions and responses
- Comprehensive health checks and monitoring
- Professional Swagger documentation
- Docker and Kubernetes deployment support

---

## 🆘 Support

For issues and questions:
1. Check this documentation
2. Review the Swagger API documentation at `/api-docs`
3. Check service logs and health endpoints
4. Create an issue with detailed error information

---

**© 2024 X-Form Realtime Service - Built with ❤️ using Node.js and Socket.IO**
