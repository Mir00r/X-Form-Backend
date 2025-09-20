# X-Form Backend - Developer Quick Reference

> **⚡ Quick commands and endpoints for daily development**

## 🚀 Quick Start Commands

```bash
# Setup (first time only)
make setup

# Daily development
make dev                    # Start all services
make health                 # Check service status
make stop                   # Stop all services

# Individual services
cd apps/auth-service && npm run dev      # Node.js hot reload
cd apps/form-service && air             # Go hot reload  
cd apps/analytics-service && uvicorn main:app --reload --port 5001  # Python
```

## 🔗 Service URLs

| Service | URL | Swagger Docs |
|---------|-----|--------------|
| **API Gateway** | `http://localhost:8080` | `http://localhost:8080/swagger/` |
| **Auth Service** | `http://localhost:3001` | `http://localhost:3001/api-docs` |
| **Form Service** | `http://localhost:8001` | `http://localhost:8001/docs` |
| **Response Service** | `http://localhost:3002` | `http://localhost:3002/docs` |
| **Realtime Service** | `http://localhost:8002` | `http://localhost:8002/api-docs` |
| **Analytics Service** | `http://localhost:5001` | `http://localhost:5001/docs` |

## 🧪 Quick API Tests

### Authentication Flow
```bash
# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"dev@example.com","username":"dev","password":"SecurePass123!","firstName":"Dev","lastName":"User"}'

# Login and get token
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"dev@example.com","password":"SecurePass123!"}' | jq -r '.token')

# Use token in requests
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer $TOKEN"
```

### Form Management
```bash
# Create form
curl -X POST http://localhost:8080/api/v1/forms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Form","description":"A test form","questions":[{"type":"text","title":"Your name?","required":true}]}'

# List forms
curl -X GET http://localhost:8080/api/v1/forms \
  -H "Authorization: Bearer $TOKEN"
```

## 🔧 Development Tools

### Code Quality
```bash
# Node.js services
npm run lint && npm run test

# Go services  
go test ./... && go vet ./...

# Python services
black . && flake8 . && pytest
```

### Database
```bash
# Connect to PostgreSQL
psql $DATABASE_URL

# Connect to Redis
redis-cli -u $REDIS_URL

# Reset database
make db-reset
```

## 📊 Monitoring

### Health Checks
```bash
curl http://localhost:8080/health     # API Gateway
curl http://localhost:3001/health     # Auth Service
curl http://localhost:8001/health     # Form Service
```

### Dashboards
- **Grafana**: `http://grafana.localhost:3000` (admin/admin)
- **Prometheus**: `http://prometheus.localhost:9091`
- **Traefik**: `http://traefik.localhost:8080`

## 🐛 Troubleshooting

### Common Issues
```bash
# Port in use
lsof -ti:8080 | xargs kill -9

# Reset Docker
make clean && docker system prune -a

# Check logs
make logs
docker-compose logs [service-name]

# Database connection
make db-setup
```

### Environment Variables
```bash
# Essential .env settings
JWT_SECRET=your-secret-key
DATABASE_URL=postgresql://xform_user:xform_password@localhost:5432/xform_db
REDIS_URL=redis://localhost:6379
NODE_ENV=development
```

## 📚 More Resources

- **[Complete Development Guide](LOCAL_DEVELOPMENT_COMPLETE_GUIDE.md)** - Comprehensive setup and usage
- **[Architecture Guide](../architecture/ARCHITECTURE_V2.md)** - System overview
- **[API Documentation](http://localhost:8080/swagger/)** - Interactive API docs
- **[Troubleshooting Guide](LOCAL_DEVELOPMENT_COMPLETE_GUIDE.md#-troubleshooting)** - Common issues and solutions

---

**💡 Pro Tip**: Bookmark `http://localhost:8080/swagger/` for interactive API testing!
