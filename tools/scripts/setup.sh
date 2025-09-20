#!/bin/bash

# X-Form Backend Setup Script
# This script sets up the complete development environment

set -e

echo "🚀 Setting up X-Form Backend..."

# Check if required tools are installed
check_tool() {
    if ! command -v $1 &> /dev/null; then
        echo "❌ $1 is not installed. Please install it first."
        exit 1
    fi
}

echo "🔍 Checking required tools..."
check_tool "docker"
check_tool "docker-compose"
check_tool "node"
check_tool "npm"
check_tool "go"
check_tool "python3"

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file..."
    cp .env.example .env
    echo "⚠️  Please update the .env file with your actual configuration values"
fi

# Start infrastructure services
echo "🐳 Starting infrastructure services (PostgreSQL, Redis)..."
docker-compose up -d postgres redis

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 10

# Setup Auth Service
echo "🔐 Setting up Auth Service..."
cd apps/auth-service
npm install
cd ../..

# Setup Response Service  
echo "📝 Setting up Response Service..."
cd apps/response-service
npm install
cd ../..

# Setup Form Service (Go)
echo "📋 Setting up Form Service..."
cd apps/form-service
go mod tidy
cd ../..

# Setup Real-time Service (Go)
echo "⚡ Setting up Real-time Service..."
cd apps/realtime-service
go mod tidy
cd ../..

# Setup Analytics Service (Python)
echo "📊 Setting up Analytics Service..."
cd apps/analytics-service
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
cd ../..

echo "✅ Setup complete!"
echo ""
echo "🔧 Next steps:"
echo "1. Update .env file with your configuration"
echo "2. Run 'docker-compose up' to start all services"
echo "3. Or run individual services:"
echo "   - Auth Service: cd apps/auth-service && npm run dev"
echo "   - Form Service: cd apps/form-service && go run cmd/server/main.go"
echo "   - Real-time Service: cd apps/realtime-service && go run cmd/server/main.go"
echo "   - Response Service: cd apps/response-service && npm run dev"
echo "   - Analytics Service: cd apps/analytics-service && python main.py"
echo ""
echo "📚 API Documentation will be available at:"
echo "   - API Gateway: http://localhost:8080"
echo "   - Auth Service: http://localhost:3001"
echo "   - Form Service: http://localhost:8001"
echo "   - Real-time Service: http://localhost:8002"
echo "   - Response Service: http://localhost:3002"
echo "   - Analytics Service: http://localhost:5001"
