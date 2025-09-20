#!/bin/bash

# X-Form Backend Development Environment Verification Script
# This script verifies that all tools are installed and the development environment is working

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 X-Form Backend Development Environment Verification${NC}"
echo "=================================================================="
echo ""

# Function to check if command exists
check_command() {
    if command -v $1 &> /dev/null; then
        echo -e "${GREEN}✅ $1: $(command -v $1)${NC}"
        return 0
    else
        echo -e "${RED}❌ $1: Not found${NC}"
        return 1
    fi
}

# Function to check version
check_version() {
    local cmd=$1
    local version_cmd=$2
    local expected=$3
    
    if command -v $cmd &> /dev/null; then
        local version=$($version_cmd 2>/dev/null || echo "unknown")
        echo -e "${GREEN}✅ $cmd: $version${NC}"
        return 0
    else
        echo -e "${RED}❌ $cmd: Not installed${NC}"
        return 1
    fi
}

# Check required tools
echo -e "${YELLOW}📋 Required Tools:${NC}"
echo "==================="

# Core tools
check_version "node" "node --version" "v18+"
check_version "npm" "npm --version" "9+"
check_version "go" "go version" "1.21+"
check_version "python3" "python3 --version" "3.8+"
check_version "docker" "docker --version" "20+"
check_version "git" "git --version" "2+"
check_version "make" "make --version | head -n1" "4+"

# Check Docker Compose
echo -n "🐳 Docker Compose: "
if docker compose version &> /dev/null; then
    echo -e "${GREEN}✅ $(docker compose version)${NC}"
elif docker-compose --version &> /dev/null; then
    echo -e "${GREEN}✅ $(docker-compose --version)${NC}"
else
    echo -e "${RED}❌ Not found${NC}"
fi

echo ""

# Check optional development tools
echo -e "${YELLOW}🚀 Development Tools:${NC}"
echo "====================="

check_command "code"
check_command "curl"
check_command "jq"
check_command "http"
check_command "hey"

echo ""

# Check Go tools
echo -e "${YELLOW}🔧 Go Development Tools:${NC}"
echo "========================"

if command -v go &> /dev/null; then
    # Check for air (hot reload)
    if [ -f "$HOME/go/bin/air" ] || command -v air &> /dev/null; then
        echo -e "${GREEN}✅ air: Hot reload tool available${NC}"
    else
        echo -e "${YELLOW}⚠️  air: Not installed (install with: go install github.com/cosmtrek/air@latest)${NC}"
    fi
    
    # Check for golangci-lint
    if command -v golangci-lint &> /dev/null; then
        echo -e "${GREEN}✅ golangci-lint: Available${NC}"
    else
        echo -e "${YELLOW}⚠️  golangci-lint: Not installed (recommended for code quality)${NC}"
    fi
fi

echo ""

# Check Node.js tools
echo -e "${YELLOW}📦 Node.js Development Tools:${NC}"
echo "============================="

if command -v node &> /dev/null; then
    # Check for nodemon
    if npm list -g nodemon &> /dev/null; then
        echo -e "${GREEN}✅ nodemon: Available globally${NC}"
    else
        echo -e "${YELLOW}⚠️  nodemon: Not installed globally (install with: npm install -g nodemon)${NC}"
    fi
    
    # Check for TypeScript
    if npm list -g typescript &> /dev/null; then
        echo -e "${GREEN}✅ typescript: Available globally${NC}"
    else
        echo -e "${YELLOW}⚠️  typescript: Not installed globally (install with: npm install -g typescript)${NC}"
    fi
fi

echo ""

# Check Python tools
echo -e "${YELLOW}🐍 Python Development Tools:${NC}"
echo "============================"

if command -v python3 &> /dev/null; then
    # Check for virtual environment
    if python3 -c "import venv" &> /dev/null; then
        echo -e "${GREEN}✅ venv: Virtual environment support available${NC}"
    else
        echo -e "${RED}❌ venv: Virtual environment not available${NC}"
    fi
    
    # Check for pip
    if command -v pip3 &> /dev/null; then
        echo -e "${GREEN}✅ pip3: $(pip3 --version)${NC}"
    else
        echo -e "${RED}❌ pip3: Not available${NC}"
    fi
fi

echo ""

# Test Docker setup
echo -e "${YELLOW}🐳 Docker Environment Test:${NC}"
echo "=========================="

if command -v docker &> /dev/null; then
    # Test Docker daemon
    if docker info &> /dev/null; then
        echo -e "${GREEN}✅ Docker daemon: Running${NC}"
        
        # Test Docker run
        if docker run --rm hello-world &> /dev/null; then
            echo -e "${GREEN}✅ Docker run: Working${NC}"
        else
            echo -e "${RED}❌ Docker run: Failed${NC}"
        fi
    else
        echo -e "${RED}❌ Docker daemon: Not running${NC}"
        echo -e "${YELLOW}💡 Please start Docker Desktop or Docker daemon${NC}"
    fi
else
    echo -e "${RED}❌ Docker: Not installed${NC}"
fi

echo ""

# Check project structure
echo -e "${YELLOW}📁 Project Structure Check:${NC}"
echo "=========================="

required_dirs=(
    "apps"
    "infrastructure"
    "docs"
    "tools"
    "configs"
)

for dir in "${required_dirs[@]}"; do
    if [ -d "$dir" ]; then
        echo -e "${GREEN}✅ $dir/: Present${NC}"
    else
        echo -e "${RED}❌ $dir/: Missing${NC}"
    fi
done

# Check key files
key_files=(
    "Makefile"
    "docker-compose.yml"
    ".env.example"
    "README.md"
)

for file in "${key_files[@]}"; do
    if [ -f "$file" ]; then
        echo -e "${GREEN}✅ $file: Present${NC}"
    else
        echo -e "${RED}❌ $file: Missing${NC}"
    fi
done

echo ""

# Test environment setup
echo -e "${YELLOW}🔧 Environment Setup Test:${NC}"
echo "========================="

# Check if .env exists
if [ -f ".env" ]; then
    echo -e "${GREEN}✅ .env: Configuration file exists${NC}"
else
    echo -e "${YELLOW}⚠️  .env: Not found (copy from .env.example)${NC}"
fi

# Test make commands
if command -v make &> /dev/null && [ -f "Makefile" ]; then
    echo -e "${GREEN}✅ Makefile: Available${NC}"
    echo -e "${BLUE}📋 Available make commands:${NC}"
    make help 2>/dev/null | head -10 || echo "  Run 'make help' to see available commands"
else
    echo -e "${RED}❌ Make setup: Not working${NC}"
fi

echo ""

# Final recommendations
echo -e "${YELLOW}💡 Next Steps:${NC}"
echo "============="

if [ ! -f ".env" ]; then
    echo -e "${BLUE}1.${NC} Copy environment file: ${YELLOW}cp .env.example .env${NC}"
fi

echo -e "${BLUE}2.${NC} Setup development environment: ${YELLOW}make setup${NC}"
echo -e "${BLUE}3.${NC} Start development services: ${YELLOW}make dev${NC}"
echo -e "${BLUE}4.${NC} Check service health: ${YELLOW}make health${NC}"
echo -e "${BLUE}5.${NC} Access API documentation: ${YELLOW}http://localhost:8080/swagger/${NC}"

echo ""

# Installation suggestions
echo -e "${YELLOW}🔗 Installation Resources:${NC}"
echo "========================="
echo -e "${BLUE}📖 Complete setup guide:${NC} docs/development/TOOLS_INSTALLATION_GUIDE.md"
echo -e "${BLUE}🚀 Development guide:${NC} docs/development/LOCAL_DEVELOPMENT_COMPLETE_GUIDE.md"
echo -e "${BLUE}⚡ Quick reference:${NC} docs/development/DEVELOPER_QUICK_REFERENCE.md"

echo ""
echo -e "${GREEN}🎉 Verification complete!${NC}"
echo ""

# Summary
missing_tools=0
if ! command -v node &> /dev/null; then ((missing_tools++)); fi
if ! command -v go &> /dev/null; then ((missing_tools++)); fi
if ! command -v python3 &> /dev/null; then ((missing_tools++)); fi
if ! command -v docker &> /dev/null; then ((missing_tools++)); fi

if [ $missing_tools -eq 0 ]; then
    echo -e "${GREEN}✅ All core tools are installed! You're ready for development.${NC}"
else
    echo -e "${YELLOW}⚠️  $missing_tools core tool(s) missing. Please install them before proceeding.${NC}"
fi

echo -e "${BLUE}Happy coding! 🚀${NC}"
