# Docker Commands Reference for X-Form Backend

## 📋 Table of Contents
- [🚀 Quick Start Commands](#-quick-start-commands)
- [🔧 Service Management](#-service-management)
- [🐛 Troubleshooting & Debugging](#-troubleshooting--debugging)
- [💾 Database Operations](#-database-operations)
- [📊 Monitoring & Health Checks](#-monitoring--health-checks)
- [🧹 Cleanup & Reset](#-cleanup--reset)
- [🏗️ Build & Development](#️-build--development)
- [🌐 Network Troubleshooting](#-network-troubleshooting)
- [📝 Log Analysis](#-log-analysis)
- [⚡ Performance & Resources](#-performance--resources)

---

## 🚀 Quick Start Commands

### Start the Complete Stack
```bash
# Start all services with Traefik (Production-like)
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml up -d

# Start specific services only
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml up -d postgres redis traefik

# Start with build (force rebuild)
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml up -d --build
```

### Stop the Stack
```bash
# Stop all services
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml down

# Stop and remove volumes (⚠️ DESTROYS DATA)
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml down -v

# Stop and remove everything (containers, networks, images)
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml down --rmi all -v
```

---

## 🔧 Service Management

### Check Service Status
```bash
# List all running containers
docker ps

# List containers for this project only
docker ps --filter "name=xform-"

# Show detailed service status with health
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml ps

# Show all containers (including stopped)
docker ps -a --filter "name=xform-"
```

### Start/Stop Individual Services
```bash
# Start specific service
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml up -d postgres

# Stop specific service
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml stop auth-service

# Restart specific service
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml restart form-service

# Remove and recreate service
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml up -d --force-recreate auth-service
```

### Service Dependencies
```bash
# Start services with dependencies (database first, then app services)
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml up -d postgres redis
sleep 10  # Wait for databases to be ready
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml up -d auth-service form-service response-service
```

---

## 🐛 Troubleshooting & Debugging

### Container Inspection
```bash
# Get detailed container information
docker inspect xform-auth

# Check container resource usage
docker stats xform-auth

# Get container IP address
docker inspect xform-auth | grep IPAddress

# Check container environment variables
docker inspect xform-auth | grep -A 20 "Env"
```

### Access Container Shell
```bash
# Access running container shell
docker exec -it xform-auth /bin/sh
docker exec -it xform-postgres /bin/bash

# Run one-off commands in container
docker exec xform-auth ls -la /app/
docker exec xform-auth cat /app/package.json
docker exec xform-auth ps aux
```

### Container Health Debugging
```bash
# Check health status
docker inspect xform-auth | grep -A 10 "Health"

# Run health check manually
docker exec xform-auth curl -f http://localhost:3001/health

# Check if service ports are listening
docker exec xform-auth netstat -tlnp
```

---

## 💾 Database Operations

### PostgreSQL Container Operations
```bash
# Connect to PostgreSQL
docker exec -it xform-postgres psql -U xform_user -d postgres

# List all databases
docker exec xform-postgres psql -U xform_user -d postgres -c "SELECT datname FROM pg_database;"

# Create new database
docker exec xform-postgres psql -U xform_user -d postgres -c "CREATE DATABASE xform_new;"

# Drop database (⚠️ DESTROYS DATA)
docker exec xform-postgres psql -U xform_user -d postgres -c "DROP DATABASE IF EXISTS xform_test;"

# Check database size
docker exec xform-postgres psql -U xform_user -d postgres -c "SELECT pg_database.datname, pg_size_pretty(pg_database_size(pg_database.datname)) AS size FROM pg_database;"

# Backup database
docker exec xform-postgres pg_dump -U xform_user xform_forms > backup_forms_$(date +%Y%m%d_%H%M%S).sql

# Restore database
docker exec -i xform-postgres psql -U xform_user xform_forms < backup_forms.sql
```

### Database Table Operations
```bash
# List tables in a database
docker exec xform-postgres psql -U xform_user -d xform_forms -c "\dt"

# Describe table structure
docker exec xform-postgres psql -U xform_user -d xform_forms -c "\d forms"

# Drop tables (⚠️ DESTROYS DATA)
docker exec xform-postgres psql -U xform_user -d xform_forms -c "DROP TABLE IF EXISTS forms CASCADE;"

# Check table row counts
docker exec xform-postgres psql -U xform_user -d xform_forms -c "SELECT 'forms' as table_name, COUNT(*) as row_count FROM forms UNION SELECT 'questions', COUNT(*) FROM questions;"
```

### Redis Container Operations
```bash
# Connect to Redis CLI
docker exec -it xform-redis redis-cli

# Check Redis info
docker exec xform-redis redis-cli info

# List all keys
docker exec xform-redis redis-cli keys "*"

# Clear all Redis data (⚠️ DESTROYS CACHE)
docker exec xform-redis redis-cli flushall

# Check Redis memory usage
docker exec xform-redis redis-cli info memory
```

---

## 📊 Monitoring & Health Checks

### Service Health Checks
```bash
# Check all service health via Docker
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# Test health endpoints manually
curl http://localhost:8080/ping  # Traefik health
curl http://localhost:3001/health  # Auth service (if exposed)
curl http://localhost:8001/health  # Form service (if exposed)
curl http://localhost:3002/health  # Response service (if exposed)

# Check PostgreSQL connectivity
docker exec xform-postgres pg_isready -U xform_user

# Check Redis connectivity
docker exec xform-redis redis-cli ping
```

### Traefik Monitoring
```bash
# Access Traefik dashboard
open http://localhost:8080

# Check Traefik configuration
curl -s http://localhost:8080/api/overview | jq

# List discovered services
curl -s http://localhost:8080/api/http/services | jq

# List active routers
curl -s http://localhost:8080/api/http/routers | jq

# Check Traefik logs for routing issues
docker logs xform-traefik --tail=50
```

---

## 🧹 Cleanup & Reset

### Remove Containers
```bash
# Stop and remove all project containers
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml down

# Force remove containers
docker rm -f $(docker ps -aq --filter "name=xform-")

# Remove containers and networks
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml down --remove-orphans
```

### Volume Management
```bash
# List project volumes
docker volume ls --filter "name=containers_"

# Remove specific volume (⚠️ DESTROYS DATA)
docker volume rm containers_postgres-data

# Remove all project volumes (⚠️ DESTROYS ALL DATA)
docker volume rm $(docker volume ls -q --filter "name=containers_")

# Backup volume data
docker run --rm -v containers_postgres-data:/data -v $(pwd):/backup alpine tar czf /backup/postgres-backup-$(date +%Y%m%d_%H%M%S).tar.gz -C /data .
```

### Network Cleanup
```bash
# List networks
docker network ls

# Remove project networks
docker network rm containers_xform-network

# Remove unused networks
docker network prune
```

### Complete System Cleanup
```bash
# ⚠️ NUCLEAR OPTION: Remove everything Docker
docker system prune -a --volumes

# Remove only unused resources
docker system prune

# Remove unused images
docker image prune -a

# Remove unused containers
docker container prune
```

---

## 🏗️ Build & Development

### Building Services
```bash
# Build all services
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml build

# Build specific service
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml build auth-service

# Build without cache (fresh build)
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml build --no-cache form-service

# Build and start immediately
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml up -d --build auth-service
```

### Development Mode
```bash
# Run with development overrides
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml -f docker-compose.dev.yml up -d

# Mount local code for hot reload (if configured)
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml -f docker-compose.dev.yml up -d --build
```

### Image Management
```bash
# List project images
docker images --filter "reference=containers-*"

# Remove specific image
docker rmi containers-auth-service

# Tag image for deployment
docker tag containers-auth-service:latest myregistry/auth-service:v1.0.0

# Check image layers and size
docker history containers-auth-service
```

---

## 🌐 Network Troubleshooting

### Network Inspection
```bash
# List networks
docker network ls

# Inspect project network
docker network inspect containers_xform-network

# Check which containers are connected to network
docker network inspect containers_xform-network | grep -A 5 "Containers"
```

### Connectivity Testing
```bash
# Test connectivity between containers
docker exec xform-auth ping xform-postgres
docker exec xform-auth nc -zv xform-postgres 5432

# Check DNS resolution
docker exec xform-auth nslookup xform-postgres
docker exec xform-auth dig xform-postgres

# Test HTTP connectivity between services
docker exec xform-auth wget -qO- http://xform-forms:8001/health
```

### Port Mapping Issues
```bash
# Check port conflicts
netstat -tulpn | grep :5432
lsof -i :5432

# Find which process is using a port
sudo lsof -i :8080

# Test port accessibility from host
telnet localhost 5432
nc -zv localhost 5432
```

---

## 📝 Log Analysis

### Container Logs
```bash
# View logs for specific service
docker logs xform-auth

# Follow logs in real-time
docker logs -f xform-auth

# Get last N lines
docker logs --tail=50 xform-auth

# Logs with timestamps
docker logs -t xform-auth

# Logs from specific time
docker logs --since="2025-09-21T10:00:00" xform-auth
docker logs --until="2025-09-21T11:00:00" xform-auth
```

### All Services Logs
```bash
# View all service logs
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml logs

# Follow all logs
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml logs -f

# Logs for specific services
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml logs auth-service form-service

# Export logs to file
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml logs > all-services-$(date +%Y%m%d_%H%M%S).log
```

### Log Analysis Commands
```bash
# Search for errors in logs
docker logs xform-auth 2>&1 | grep -i error

# Count error occurrences
docker logs xform-auth 2>&1 | grep -c "ERROR"

# Extract stack traces
docker logs xform-auth 2>&1 | grep -A 10 -B 5 "at "

# Monitor for specific patterns
docker logs -f xform-auth | grep --color=always -E "(ERROR|WARN|Failed)"
```

---

## ⚡ Performance & Resources

### Resource Monitoring
```bash
# Monitor container resource usage
docker stats

# Monitor specific containers
docker stats xform-auth xform-postgres xform-redis

# Get container resource limits
docker inspect xform-auth | grep -A 10 "Memory"

# Check disk usage
docker system df

# Check volume sizes
docker system df -v
```

### Performance Optimization
```bash
# Limit container memory (in docker-compose.yml)
# mem_limit: 512m
# mem_reservation: 256m

# Limit CPU usage
# cpus: '0.5'
# cpu_percent: 50

# Check container processes
docker exec xform-auth ps aux

# Monitor database performance
docker exec xform-postgres psql -U xform_user -d xform_forms -c "SELECT * FROM pg_stat_activity;"
```

---

## 🔗 Quick Reference Commands

### Most Used Commands
```bash
# Start everything
make start  # or manual command above

# Check status
docker ps --filter "name=xform-"

# View logs
docker logs -f xform-auth

# Access database
docker exec -it xform-postgres psql -U xform_user -d xform_forms

# Restart problematic service
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml restart auth-service

# Clean restart
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml down
docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml up -d --build
```

### Emergency Troubleshooting
```bash
# When services won't start
docker ps -a  # Check exit codes
docker logs xform-auth  # Check logs
docker inspect xform-auth  # Check configuration

# When database connection fails
docker exec xform-postgres pg_isready -U xform_user
docker exec -it xform-postgres psql -U xform_user -d postgres -c "\l"

# When network issues occur
docker network ls
docker network inspect containers_xform-network

# When ports are conflicting
netstat -tulpn | grep :5432
sudo lsof -i :8080
```

---

## 📚 Additional Resources

### Environment Files
- `.env` - Main environment configuration
- `apps/*/.*env` - Service-specific configurations
- `infrastructure/containers/docker-compose-traefik.yml` - Main compose file

### Service URLs (when running)
- **Traefik Dashboard**: http://localhost:8080
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379
- **Auth Service**: http://localhost:3001 (container) / http://localhost:3001 (local dev)
- **Form Service**: http://localhost:8001 (container) / http://localhost:8081 (local dev)
- **Response Service**: http://localhost:3002

### Helpful Aliases
Add these to your `.bashrc` or `.zshrc`:
```bash
alias dps='docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"'
alias dlogs='docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml logs -f'
alias dup='docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml up -d'
alias ddown='docker compose --env-file .env -f infrastructure/containers/docker-compose-traefik.yml down'
alias dpsql='docker exec -it xform-postgres psql -U xform_user -d xform_forms'
alias dredis='docker exec -it xform-redis redis-cli'
```

---

## ⚠️ Important Notes

1. **Always use `--env-file .env`** to ensure proper environment variable loading
2. **Database operations are destructive** - always backup before DROP/DELETE commands
3. **Network conflicts** can prevent startup - clean up conflicting networks
4. **Port conflicts** will cause services to fail - check with `netstat` or `lsof`
5. **Docker daemon must be running** - start Docker Desktop if commands fail
6. **Volume data persists** even after container removal unless explicitly deleted

---

*This reference was created based on actual troubleshooting sessions and covers common scenarios encountered during X-Form Backend development.*
