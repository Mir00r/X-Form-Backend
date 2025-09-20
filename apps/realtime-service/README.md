# X-Form Realtime Service

## 📋 Overview

The X-Form Realtime Service is a comprehensive WebSocket-based real-time communication service built with Node.js, Express, and Socket.IO. It provides live communication capabilities for the X-Form Backend, enabling real-time form updates, response notifications, and collaborative features.

## 🚀 Features

- ✅ **Real-time WebSocket messaging** with Socket.IO
- ✅ **Form subscription management** for live updates
- ✅ **Event broadcasting** for form responses and lifecycle changes
- ✅ **Connection monitoring** and statistics
- ✅ **Comprehensive Swagger API documentation**
- ✅ **Health checks & metrics** with Kubernetes support
- ✅ **CORS & security headers** with Helmet.js
- ✅ **JWT authentication support** (optional)
- ✅ **Room-based communication** for organized messaging
- ✅ **Horizontal scaling ready** with Redis adapter support
- ✅ **Graceful shutdown** handling
- ✅ **Interactive demo client** for testing

## 🛠️ Technology Stack

- **Runtime**: Node.js 18+
- **Framework**: Express.js 4.18.2
- **WebSocket**: Socket.IO 4.7.4
- **Security**: Helmet.js, CORS
- **Authentication**: JWT (optional)
- **Documentation**: Swagger/OpenAPI 3.0
- **Environment**: dotenv
- **Scaling**: Redis (configurable)

## 📦 Installation & Setup

### Prerequisites

- Node.js 18 or higher
- npm or yarn package manager
- (Optional) Redis for horizontal scaling

### Quick Start

1. **Navigate to the service directory:**
   ```bash
   cd services/realtime-service
   ```

2. **Install dependencies:**
   ```bash
   npm install
   ```

3. **Configure environment:**
   ```bash
   cp .env.example .env
   # Edit .env file with your configurations
   ```

4. **Start the service:**
   ```bash
   # Development mode
   npm run dev
   
   # Production mode
   npm start
   ```

5. **Verify the service:**
   - Health check: http://localhost:8002/health
   - API documentation: http://localhost:8002/api-docs
   - Service info: http://localhost:8002/

## 🔧 Configuration

### Environment Variables

```bash
# Server Configuration
REALTIME_SERVICE_PORT=8002
NODE_ENV=development

# CORS Configuration
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001

# JWT Configuration (Optional)
JWT_SECRET=your-super-secure-jwt-secret-key-here
REQUIRE_AUTH=false

# Socket.IO Configuration
SOCKET_PING_TIMEOUT=5000
SOCKET_PING_INTERVAL=25000

# API Documentation
API_TITLE=X-Form Realtime Service API
API_VERSION=1.0.0
API_DESCRIPTION=Real-time WebSocket service for X-Form Backend
```

### Package.json Scripts

Add these scripts to your `package.json`:

```json
{
  "scripts": {
    "start": "node src/index.js",
    "dev": "nodemon src/index.js",
    "test": "jest",
    "test:watch": "jest --watch",
    "lint": "eslint src/",
    "lint:fix": "eslint src/ --fix"
  }
}
```

## 📡 API Endpoints

### REST API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | Service information and features |
| GET | `/health` | Basic health check |
| GET | `/health/detailed` | Detailed health information |
| GET | `/health/live` | Kubernetes liveness probe |
| GET | `/health/ready` | Kubernetes readiness probe |
| GET | `/api-docs` | Swagger API documentation |
| GET | `/stats` | Connection statistics |
| GET | `/ws/info` | WebSocket endpoint information |
| GET | `/ws/connections` | Active WebSocket connections |
| GET | `/ws/rooms` | Active WebSocket rooms |
| POST | `/ws/broadcast` | Broadcast message to room |
| POST | `/ws/disconnect/{socketId}` | Force disconnect socket |
| POST | `/events/form/{formId}/notify` | Send form notification |
| POST | `/events/form/{formId}/response` | Broadcast new response |
| POST | `/events/form/{formId}/status` | Update form status |
| POST | `/events/broadcast` | Global broadcast message |
| GET | `/events/metrics` | Get event metrics |

### WebSocket Events

#### Client to Server Events

| Event | Description | Data |
|-------|-------------|------|
| `form:subscribe` | Subscribe to form updates | `formId: string` |
| `form:unsubscribe` | Unsubscribe from form | `formId: string` |
| `response:new` | Submit new response | `{formId, responseId, data}` |
| `response:update` | Update existing response | `{formId, responseId, data}` |
| `form:published` | Notify form published | `{formId, reason}` |
| `form:closed` | Notify form closed | `{formId, reason}` |
| `user:typing` | User typing indicator | `{formId, fieldName}` |
| `user:stopped_typing` | User stopped typing | `{formId, fieldName}` |
| `ping` | Connection health check | `null` |
| `rooms:list` | Get subscribed rooms | `null` |

#### Server to Client Events

| Event | Description | Data |
|-------|-------------|------|
| `form:subscribed` | Subscription confirmed | `{formId, roomName, timestamp}` |
| `form:unsubscribed` | Unsubscription confirmed | `{formId, roomName, timestamp}` |
| `form:new_subscriber` | New subscriber joined | `{socketId, userId, formId}` |
| `form:subscriber_left` | Subscriber left | `{socketId, userId, formId}` |
| `response:update` | New response received | `{formId, responseId, data}` |
| `response:updated` | Response was updated | `{formId, responseId, data}` |
| `form:published` | Form was published | `{formId, publishedBy, timestamp}` |
| `form:closed` | Form was closed | `{formId, closedBy, timestamp}` |
| `user:typing` | User is typing | `{formId, fieldName, socketId}` |
| `user:stopped_typing` | User stopped typing | `{formId, fieldName, socketId}` |
| `user:disconnected` | User disconnected | `{socketId, userId, reason}` |
| `pong` | Ping response | `{timestamp, socketId}` |
| `rooms:list` | List of subscribed rooms | `{rooms: [], socketId}` |
| `error` | Error occurred | `{message, timestamp}` |

## 🧪 Testing

### Interactive Demo Client

The service includes a comprehensive HTML demo client for testing WebSocket functionality:

1. **Start the service:**
   ```bash
   npm run dev
   ```

2. **Open the demo client:**
   ```bash
   open demo/websocket-demo.html
   # Or visit: http://localhost:8002 and navigate to demo
   ```

3. **Test features:**
   - WebSocket connection/disconnection
   - Form subscriptions
   - Real-time messaging
   - Event broadcasting
   - Connection statistics

### API Testing with cURL

#### Health Check
```bash
curl http://localhost:8002/health
```

#### Service Information
```bash
curl http://localhost:8002/
```

#### Connection Statistics
```bash
curl http://localhost:8002/stats
```

#### Send Form Notification
```bash
curl -X POST http://localhost:8002/events/form/test123/notify \
  -H "Content-Type: application/json" \
  -d '{
    "event": "form:updated",
    "data": {"title": "New Form Title"},
    "urgent": false
  }'
```

#### Broadcast New Response
```bash
curl -X POST http://localhost:8002/events/form/test123/response \
  -H "Content-Type: application/json" \
  -d '{
    "responseId": "resp456",
    "data": "Sample response data",
    "userId": "user123"
  }'
```

### WebSocket Testing with Node.js

```javascript
const io = require('socket.io-client');

// Connect to the service
const socket = io('http://localhost:8002');

// Listen for connection
socket.on('connect', () => {
  console.log('Connected:', socket.id);
  
  // Subscribe to a form
  socket.emit('form:subscribe', 'test123');
});

// Listen for form subscription confirmation
socket.on('form:subscribed', (data) => {
  console.log('Subscribed to form:', data);
  
  // Send a test response
  socket.emit('response:new', {
    formId: 'test123',
    responseId: 'resp789',
    data: 'Test response data'
  });
});

// Listen for new responses
socket.on('response:update', (data) => {
  console.log('New response:', data);
});

// Handle errors
socket.on('error', (error) => {
  console.error('Socket error:', error);
});
```

## 🔒 Security

### Authentication (Optional)

The service supports JWT authentication. When enabled:

1. **Set environment variables:**
   ```bash
   REQUIRE_AUTH=true
   JWT_SECRET=your-super-secure-secret-key
   ```

2. **Connect with token:**
   ```javascript
   const socket = io('http://localhost:8002', {
     auth: {
       token: 'your-jwt-token'
     }
   });
   ```

### CORS Configuration

Configure allowed origins in the environment:

```bash
ALLOWED_ORIGINS=http://localhost:3000,https://yourdomain.com
```

### Security Headers

The service automatically applies security headers using Helmet.js:
- Content Security Policy
- X-Frame-Options
- X-Content-Type-Options
- And more...

## 🚀 Deployment

### Docker Deployment

```dockerfile
FROM node:18-alpine

WORKDIR /app

COPY package*.json ./
RUN npm ci --only=production

COPY src/ ./src/
COPY .env ./

EXPOSE 8002

USER node

CMD ["npm", "start"]
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
        image: your-registry/realtime-service:latest
        ports:
        - containerPort: 8002
        env:
        - name: REALTIME_SERVICE_PORT
          value: "8002"
        - name: NODE_ENV
          value: "production"
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8002
          initialDelaySeconds: 30
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8002
          initialDelaySeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: realtime-service
spec:
  selector:
    app: realtime-service
  ports:
  - port: 8002
    targetPort: 8002
  type: ClusterIP
```

### Horizontal Scaling with Redis

For horizontal scaling across multiple instances:

1. **Install Redis adapter:**
   ```bash
   npm install @socket.io/redis-adapter redis
   ```

2. **Configure Redis:**
   ```bash
   REDIS_HOST=localhost
   REDIS_PORT=6379
   REDIS_PASSWORD=your-password
   ```

3. **Update application code:**
   ```javascript
   const { createAdapter } = require('@socket.io/redis-adapter');
   const { createClient } = require('redis');
   
   const pubClient = createClient({ url: 'redis://localhost:6379' });
   const subClient = pubClient.duplicate();
   
   io.adapter(createAdapter(pubClient, subClient));
   ```

## 📊 Monitoring

### Health Checks

- **Basic**: `/health` - Service health status
- **Detailed**: `/health/detailed` - Comprehensive health information
- **Liveness**: `/health/live` - Kubernetes liveness probe
- **Readiness**: `/health/ready` - Kubernetes readiness probe

### Metrics

- **Connection Statistics**: `/stats`
- **Event Metrics**: `/events/metrics`
- **WebSocket Information**: `/ws/info`
- **Active Connections**: `/ws/connections`
- **Active Rooms**: `/ws/rooms`

### Logging

The service provides structured logging for:
- Connection events
- Message broadcasting
- Error handling
- Performance metrics

## 🛠️ Development

### Project Structure

```
src/
├── app.js              # Main application setup
├── index.js            # Entry point
├── swagger.js          # Swagger configuration
└── routes/
    ├── health.js       # Health check routes
    ├── websocket.js    # WebSocket management routes
    └── realtime.js     # Real-time event routes

demo/
└── websocket-demo.html # Interactive demo client

.env                    # Environment configuration
.env.example           # Environment template
package.json           # Dependencies and scripts
README.md              # This documentation
```

### Adding New Events

1. **Define the event in Swagger documentation** (`src/swagger.js`)
2. **Add event handler in main application** (`src/app.js`)
3. **Update demo client** (`demo/websocket-demo.html`)
4. **Add tests** for the new functionality

### Code Style

- Use ES6+ features
- Follow Express.js best practices
- Implement error handling for all endpoints
- Add comprehensive logging
- Write clear documentation

## 🐛 Troubleshooting

### Common Issues

#### Connection Issues

1. **CORS errors:**
   - Check `ALLOWED_ORIGINS` environment variable
   - Ensure client origin is whitelisted

2. **Authentication errors:**
   - Verify JWT token is valid
   - Check `JWT_SECRET` configuration

3. **Connection timeouts:**
   - Increase timeout values in client
   - Check network connectivity

#### Performance Issues

1. **High memory usage:**
   - Monitor connection count
   - Implement connection limits
   - Use Redis for scaling

2. **Slow responses:**
   - Check event handler performance
   - Monitor server resources
   - Optimize message payloads

### Debugging

Enable debug logging:

```bash
DEBUG=socket.io:* node src/index.js
```

Check application logs:

```bash
# View real-time logs
tail -f logs/app.log

# Check error logs
grep "ERROR" logs/app.log
```

## 📞 Support

For issues, questions, or contributions:

1. Check the [troubleshooting section](#troubleshooting)
2. Review the [API documentation](http://localhost:8002/api-docs)
3. Test with the [demo client](demo/websocket-demo.html)
4. Check application logs for errors

## 📄 License

This project is part of the X-Form Backend system.

---

## 🎯 Quick Reference

### Start Commands
```bash
npm install          # Install dependencies
npm run dev         # Development mode
npm start           # Production mode
```

### Important URLs
```
Service Info:     http://localhost:8002/
Health Check:     http://localhost:8002/health
API Docs:         http://localhost:8002/api-docs
WebSocket:        ws://localhost:8002
Demo Client:      demo/websocket-demo.html
```

### Key Environment Variables
```bash
REALTIME_SERVICE_PORT=8002
ALLOWED_ORIGINS=http://localhost:3000
REQUIRE_AUTH=false
JWT_SECRET=your-secret
```
