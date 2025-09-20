#!/bin/bash

# Analytics Service Setup Script
# This script sets up the analytics service with all dependencies

set -e

echo "🚀 Setting up X-Form Analytics Service..."

# Check Python version
if ! command -v python3 &> /dev/null; then
    echo "❌ Python 3 is required but not installed"
    exit 1
fi

PYTHON_VERSION=$(python3 -c 'import sys; print(f"{sys.version_info.major}.{sys.version_info.minor}")')
echo "✅ Python $PYTHON_VERSION detected"

# Check if we're in the right directory
if [ ! -f "requirements.txt" ]; then
    echo "❌ Please run this script from the analytics-service directory"
    exit 1
fi

echo "📦 Installing Python dependencies..."

# Create virtual environment if it doesn't exist
if [ ! -d "venv" ]; then
    echo "🔧 Creating virtual environment..."
    python3 -m venv venv
fi

# Activate virtual environment
echo "🔧 Activating virtual environment..."
source venv/bin/activate

# Upgrade pip
echo "⬆️  Upgrading pip..."
pip install --upgrade pip

# Install dependencies
echo "📦 Installing dependencies..."
pip install -r requirements.txt

# Create necessary directories
echo "📁 Creating necessary directories..."
mkdir -p logs
mkdir -p exports
mkdir -p reports
mkdir -p tmp

# Set environment variables
echo "⚙️  Setting up environment..."
if [ ! -f ".env" ]; then
    echo "📝 Creating .env file from example..."
    cp .env.example .env
    echo "⚠️  Please update the .env file with your configuration"
else
    echo "✅ .env file already exists"
fi

# Check for required environment variables
echo "🔍 Checking environment configuration..."

# Create basic .env if it doesn't exist
if [ ! -f ".env" ]; then
    cat > .env << EOF
# Analytics Service Configuration
APP_NAME=Analytics Service
APP_VERSION=1.0.0
ENVIRONMENT=development
DEBUG=true

# Server Configuration  
HOST=0.0.0.0
PORT=8084
WORKERS=1

# Authentication
JWT_SECRET_KEY=your-super-secret-jwt-key-change-in-production
JWT_ALGORITHM=HS256
JWT_EXPIRATION_HOURS=24

# BigQuery Configuration (optional for development)
# GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
# BIGQUERY_PROJECT_ID=your-project-id
# BIGQUERY_DATASET_ID=xform_analytics

# Redis Configuration (optional for development)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_DB=0
CACHE_TTL=3600

# Logging
LOG_LEVEL=INFO

# CORS Configuration
CORS_ORIGINS=["http://localhost:3000", "http://localhost:8080"]
CORS_METHODS=["GET", "POST", "PUT", "DELETE", "OPTIONS"]
CORS_HEADERS=["*"]

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_PER_MINUTE=100

# File Upload
MAX_FILE_SIZE=50MB
UPLOAD_PATH=./uploads
EOF
    echo "✅ Created default .env file"
fi

echo "🔧 Checking service health..."

# Test import of main modules
python3 -c "
import sys
try:
    from app.main import app
    from app.config import settings
    print('✅ Main application imports successfully')
except ImportError as e:
    print(f'❌ Import error: {e}')
    sys.exit(1)
"

echo "🧪 Running basic tests..."

# Test basic endpoint availability
python3 -c "
import asyncio
import uvicorn
import threading
import time
import requests
from app.main import app

def run_server():
    uvicorn.run(app, host='127.0.0.1', port=8084, log_level='critical')

# Start server in background
server_thread = threading.Thread(target=run_server, daemon=True)
server_thread.start()

# Wait for server to start
time.sleep(3)

try:
    response = requests.get('http://127.0.0.1:8084/health', timeout=5)
    if response.status_code == 200:
        print('✅ Health check endpoint working')
    else:
        print(f'❌ Health check failed with status {response.status_code}')
except Exception as e:
    print(f'❌ Health check failed: {e}')
"

echo "📖 Generating API documentation..."

# Test Swagger docs
python3 -c "
from app.main import app
import json

# Get OpenAPI schema
openapi_schema = app.openapi()

# Save schema to file
with open('docs/openapi.json', 'w') as f:
    json.dump(openapi_schema, f, indent=2)

print('✅ OpenAPI schema generated')
print(f'📄 Endpoints documented: {len(openapi_schema.get(\"paths\", {}))}')
"

echo "🎉 Setup completed successfully!"
echo ""
echo "📋 Next steps:"
echo "1. Update the .env file with your configuration"
echo "2. Set up BigQuery credentials (optional)"
echo "3. Set up Redis server (optional)"
echo "4. Run the service with: python main.py"
echo "5. Access Swagger docs at: http://localhost:8084/docs"
echo ""
echo "🚀 To start the service:"
echo "   source venv/bin/activate"
echo "   python main.py"
echo ""
echo "📖 Documentation will be available at:"
echo "   Swagger UI: http://localhost:8084/docs"
echo "   ReDoc: http://localhost:8084/redoc"
echo "   OpenAPI JSON: http://localhost:8084/openapi.json"
