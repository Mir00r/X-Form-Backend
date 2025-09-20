#!/bin/bash

# X-Form Backend Startup Script
# Ensures Docker is running and starts the Traefik All-in-One stack

set -e

echo "🚀 X-Form Backend Startup Script"
echo ""

# Function to check if Docker is running
check_docker() {
    if ! docker info >/dev/null 2>&1; then
        return 1
    fi
    return 0
}

# Function to start Docker Desktop on macOS
start_docker() {
    echo "📦 Starting Docker Desktop..."
    open -a Docker
    
    # Wait for Docker to start
    echo "⏳ Waiting for Docker to start..."
    local count=0
    while ! check_docker; do
        if [ $count -ge 30 ]; then
            echo "❌ Docker failed to start within 30 seconds"
            echo ""
            echo "Please:"
            echo "1. Make sure Docker Desktop is installed"
            echo "2. Start Docker Desktop manually"
            echo "3. Wait for it to fully start"
            echo "4. Run this script again"
            exit 1
        fi
        sleep 2
        count=$((count + 1))
        echo -n "."
    done
    echo ""
    echo "✅ Docker is now running"
}

# Check if Docker is running
echo "🔍 Checking Docker status..."
if ! check_docker; then
    echo "⚠️  Docker is not running"
    start_docker
else
    echo "✅ Docker is already running"
fi

echo ""

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "⚠️  .env file not found"
    echo "📝 Creating .env from .env.example..."
    cp .env.example .env
    echo "✅ .env file created"
    echo ""
    echo "🔧 Please edit .env file with your configuration:"
    echo "   • JWT_SECRET: Change the default secret"
    echo "   • Firebase credentials (if using Firestore)"
    echo "   • BigQuery credentials (if using analytics)"
    echo "   • Google OAuth credentials (if using OAuth)"
    echo ""
    read -p "Press Enter to continue with default values or Ctrl+C to edit .env first..."
fi

# Validate required environment variables
echo "🔍 Validating environment variables..."
source .env

if [ -z "$JWT_SECRET" ] || [ "$JWT_SECRET" = "your-super-secret-jwt-key-change-this-in-production" ]; then
    echo "⚠️  WARNING: JWT_SECRET is using default value"
    echo "   This is insecure for production use"
fi

if [ -z "$BIGQUERY_DATASET" ]; then
    echo "⚠️  BIGQUERY_DATASET is not set, using default: xform_analytics"
    export BIGQUERY_DATASET="xform_analytics"
fi

echo "✅ Environment validation completed"
echo ""

# Start the stack
echo "🚀 Starting X-Form Backend with Traefik All-in-One..."
docker-compose -f docker-compose-traefik.yml up -d

echo ""
echo "⏳ Waiting for services to start..."
sleep 10

# Check service health
echo ""
echo "🏥 Checking service health..."

# Check Traefik
if curl -s http://localhost:8080/ping >/dev/null 2>&1; then
    echo "✅ Traefik: Running"
else
    echo "❌ Traefik: Not responding"
fi

# Check if containers are running
echo ""
echo "📊 Container Status:"
docker-compose -f docker-compose-traefik.yml ps

echo ""
echo "🎉 X-Form Backend started successfully!"
echo ""
echo "🌐 Access Points:"
echo "   📡 Main API:           http://api.localhost"
echo "   🔌 WebSocket:          ws://ws.localhost"  
echo "   📊 Traefik Dashboard:  http://traefik.localhost:8080"
echo "   📈 Grafana:            http://grafana.localhost:3000"
echo "   🔍 Prometheus:         http://prometheus.localhost:9091"
echo "   🔎 Jaeger:             http://jaeger.localhost:16686"
echo ""
echo "💡 Next steps:"
echo "   • Run 'make health' to check detailed service health"
echo "   • Run 'make api-test' to test API endpoints"
echo "   • Run 'make monitor' to open all dashboards"
echo "   • Check logs with 'make logs' if any issues"
echo ""
echo "📚 Documentation:"
echo "   • README.md - Overview and quick start"
echo "   • IMPLEMENTATION_GUIDE.md - Detailed setup guide"
echo "   • ARCHITECTURE_V2.md - Architecture details"
