#!/bin/bash

# Enhanced Response Service Quick Start Script
# This script helps you quickly start the response service with proper setup

echo "🚀 Enhanced Response Service - Quick Start"
echo "=========================================="

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 18+ first."
    exit 1
fi

# Check Node.js version
NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 18 ]; then
    echo "❌ Node.js version 18+ is required. Current version: $(node -v)"
    exit 1
fi

echo "✅ Node.js version: $(node -v)"

# Check if we're in the right directory
if [ ! -f "package.json" ]; then
    echo "❌ Please run this script from the response-service directory"
    exit 1
fi

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
    echo "📦 Installing dependencies..."
    npm install
    if [ $? -ne 0 ]; then
        echo "❌ Failed to install dependencies"
        exit 1
    fi
fi

# Create .env file if it doesn't exist
if [ ! -f ".env" ]; then
    echo "⚙️  Creating default .env file..."
    cat > .env << EOF
# Server Configuration
PORT=3002
NODE_ENV=development
HOST=0.0.0.0

# Service Information
SERVICE_NAME=response-service
SERVICE_VERSION=1.0.0

# Authentication
JWT_SECRET=development-secret-key-change-in-production
API_KEY=dev-api-key-12345

# Security
RATE_LIMIT_WINDOW_MS=900000
RATE_LIMIT_MAX_REQUESTS=100
EOF
    echo "✅ Created .env file with default settings"
fi

echo ""
echo "🎯 Starting Enhanced Response Service..."
echo "📖 Swagger Documentation will be available at: http://localhost:3002/api-docs"
echo "🏥 Health Check: http://localhost:3002/api/v1/health"
echo ""
echo "Press Ctrl+C to stop the service"
echo ""

# Start the service
node src/index.js
