# Real-Time Collaboration Service

A WebSocket-based real-time collaboration service for the X-Form Backend, enabling multiple users to collaboratively edit forms in real-time with cursor tracking, question updates, and session management.

## Features

### Core Functionality
- **Real-time WebSocket Communication**: Pure WebSocket implementation without REST endpoints
- **Form Collaboration**: Multiple users can edit forms simultaneously
- **Cursor Tracking**: Real-time cursor position sharing between users
- **Question Management**: Live question creation, updates, and deletion
- **Session Management**: User session tracking with Redis persistence
- **Rate Limiting**: Per-user rate limiting for WebSocket messages
- **JWT Authentication**: Secure authentication using JWT tokens

### WebSocket Events

#### Client → Server Events
- `join:form` - Join a form collaboration session
- `leave:form` - Leave a form collaboration session
- `cursor:update` - Update cursor position
- `question:update` - Update existing question
- `question:create` - Create new question
- `question:delete` - Delete question
- `ping` - Keep-alive ping

#### Server → Client Events
- `join:form:response` - Response to join request
- `leave:form:response` - Response to leave request
- `user:joined` - Notify when user joins
- `user:left` - Notify when user leaves
- `cursor:update` - Broadcast cursor updates
- `question:update` - Broadcast question updates
- `question:create` - Broadcast question creation
- `question:delete` - Broadcast question deletion
- `pong` - Response to ping
- `error` - Error notifications

### Architecture

```
┌─────────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│                     │    │                  │    │                 │
│   WebSocket Client  │◄──►│  Collaboration   │◄──►│     Redis       │
│                     │    │     Service      │    │   (Pub/Sub +    │
└─────────────────────┘    │                  │    │   Sessions)     │
                           │                  │    │                 │
┌─────────────────────┐    │                  │    └─────────────────┘
│                     │    │                  │    
│   WebSocket Client  │◄──►│   - Room Mgmt    │    ┌─────────────────┐
│                     │    │   - Session Mgmt │    │                 │
└─────────────────────┘    │   - Auth Service │◄──►│   Auth Service  │
                           │   - Rate Limiting│    │   (JWT Tokens)  │
┌─────────────────────┐    │   - Pub/Sub      │    │                 │
│                     │    │                  │    └─────────────────┘
│   WebSocket Client  │◄──►│                  │    
│                     │    │                  │    ┌─────────────────┐
└─────────────────────┘    └──────────────────┘    │                 │
                                                  │   Kafka         │
                                                  │ (Future: Inter- │
                                                  │ service comms)  │
                                                  │                 │
                                                  └─────────────────┘
```

## Quick Start

### Prerequisites
- Go 1.21+
- Redis 6.0+
- Docker (optional)

### Environment Setup

1. **Clone and Navigate**:
   ```bash
   cd services/collaboration-service
   ```

2. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

3. **Configure Environment**:
   ```bash
   cp .env.example .env
   # Edit .env with your Redis and auth configuration
   ```

4. **Start Redis** (if not already running):
   ```bash
   docker run -d --name redis -p 6379:6379 redis:7-alpine
   ```

### Running the Service

#### Development Mode
```bash
go run cmd/server/main.go
```

#### Production Build
```bash
go build -o collaboration-service cmd/server/main.go
./collaboration-service
```

#### Docker
```bash
docker build -t collaboration-service .
docker run -p 8083:8083 --env-file .env collaboration-service
```

## API Documentation

### WebSocket Connection

#### Connection URL
```
ws://localhost:8083/ws?token=YOUR_JWT_TOKEN
```

#### Alternative Authentication
Include JWT token in the `Authorization` header:
```
Authorization: Bearer YOUR_JWT_TOKEN
```

### Message Format

All WebSocket messages follow this JSON structure:

```json
{
  "type": "event_type",
  "payload": {},
  "timestamp": "2024-01-01T00:00:00Z",
  "messageId": "uuid",
  "userId": "user_id",
  "formId": "form_id"
}
```

### Event Examples

#### Join Form
```json
{
  "type": "join:form",
  "payload": {
    "formId": "form_123"
  }
}
```

#### Update Cursor
```json
{
  "type": "cursor:update",
  "payload": {
    "formId": "form_123",
    "position": {
      "x": 100,
      "y": 200,
      "questionId": "q1",
      "section": "title"
    },
    "color": "#ff6b6b"
  }
}
```

#### Update Question
```json
{
  "type": "question:update",
  "payload": {
    "formId": "form_123",
    "questionId": "q1",
    "changes": {
      "title": "Updated question title",
      "type": "multiple_choice"
    },
    "version": 2
  }
}
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | Service port | `8083` |
| `REDIS_HOST` | Redis host | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |
| `AUTH_JWT_SECRET` | JWT secret key | Required |
| `WEBSOCKET_MAX_USERS_PER_ROOM` | Max users per form | `50` |
| `WEBSOCKET_MESSAGE_RATE_LIMIT` | Messages per minute | `100` |

### Redis Configuration

The service uses Redis for:
- Session management
- Pub/Sub messaging
- Rate limiting
- Cursor position caching
- Room state persistence

### Authentication

- JWT token required for WebSocket connections
- Token can be provided via query parameter or Authorization header
- User permissions checked for form access and editing

## Monitoring

### Health Check
```bash
curl http://localhost:8083/health
```

### Metrics
```bash
curl http://localhost:8083/metrics
```

### Available Metrics
- `totalConnections` - Total connections since startup
- `activeConnections` - Currently active connections
- `totalRooms` - Total rooms created
- `activeRooms` - Currently active rooms
- `messagesPerSecond` - Message throughput
- `errorsPerSecond` - Error rate

## Development

### Project Structure
```
├── cmd/
│   └── server/           # Application entry point
├── internal/
│   ├── auth/            # JWT authentication
│   ├── config/          # Configuration management
│   ├── models/          # Data models
│   ├── redis/           # Redis service layer
│   └── websocket/       # WebSocket hub and handlers
├── Dockerfile           # Container build
├── go.mod              # Go dependencies
└── README.md           # This file
```

### Adding New Event Types

1. **Define Event Type** in `models/models.go`:
   ```go
   EventNewFeature EventType = "new:feature"
   ```

2. **Create Payload Struct**:
   ```go
   type NewFeaturePayload struct {
       Data string `json:"data"`
   }
   ```

3. **Implement Handler** in `websocket/handlers.go`:
   ```go
   type NewFeatureHandler struct {
       hub *Hub
   }

   func (h *NewFeatureHandler) Handle(ctx context.Context, client *Client, message *models.Message) error {
       // Implementation
   }
   ```

4. **Register Handler** in `websocket/hub.go`:
   ```go
   h.eventHandlers[models.EventNewFeature] = &NewFeatureHandler{hub: h}
   ```

### Testing

#### Unit Tests
```bash
go test ./...
```

#### Integration Tests
```bash
go test ./... -tags=integration
```

#### WebSocket Testing
Use tools like:
- [websocat](https://github.com/vi/websocat)
- Browser WebSocket API
- [wscat](https://github.com/websockets/wscat)

Example with websocat:
```bash
websocat ws://localhost:8083/ws -H="Authorization: Bearer YOUR_JWT_TOKEN"
```

## Security

### Authentication
- JWT tokens required for all connections
- Token validation on connection establishment
- User permissions checked for each action

### Rate Limiting
- Per-user message rate limiting
- Configurable limits and windows
- Redis-backed rate limiting

### CORS
- Configurable allowed origins
- Secure defaults for production

## Performance

### Scalability
- Designed for horizontal scaling
- Redis-backed state sharing
- Stateless WebSocket handlers

### Resource Management
- Connection pooling
- Automatic cleanup of inactive sessions
- Configurable connection limits

### Optimization Tips
1. **Redis Configuration**: Tune Redis memory and persistence settings
2. **Connection Limits**: Set appropriate `MAX_CONNECTIONS`
3. **Rate Limiting**: Adjust limits based on usage patterns
4. **Buffer Sizes**: Tune WebSocket buffer sizes for your payload sizes

## Troubleshooting

### Common Issues

#### Connection Refused
- Check if Redis is running
- Verify Redis connection parameters
- Check firewall settings

#### Authentication Errors
- Verify JWT secret configuration
- Check token expiration
- Ensure proper token format

#### High Memory Usage
- Monitor Redis memory usage
- Check for connection leaks
- Review cleanup intervals

### Debugging

#### Enable Debug Logging
```bash
export LOG_LEVEL=debug
```

#### Redis Debugging
```bash
redis-cli monitor  # Monitor Redis commands
redis-cli info     # Check Redis stats
```

#### WebSocket Debugging
Use browser developer tools or specialized WebSocket debugging tools.

## Deployment

### Docker Compose Example
```yaml
version: '3.8'
services:
  collaboration-service:
    build: .
    ports:
      - "8083:8083"
    environment:
      - REDIS_HOST=redis
      - AUTH_JWT_SECRET=your-secret
    depends_on:
      - redis
  
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

### Kubernetes Deployment
See `k8s/` directory for Kubernetes manifests.

### Production Considerations
- Use Redis Cluster for high availability
- Configure proper secrets management
- Set up monitoring and alerting
- Use load balancers for multiple instances
- Configure log aggregation

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

### Code Style
- Follow Go conventions
- Use `gofmt` for formatting
- Include unit tests for new features
- Update documentation

## License

This project is part of the X-Form Backend system.

## Support

For issues and questions:
1. Check the troubleshooting section
2. Review existing GitHub issues
3. Create a new issue with detailed information

---

**Built with ❤️ for real-time collaboration**
