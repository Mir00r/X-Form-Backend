#!/bin/bash

# Form Service API - Comprehensive Setup and Run Script
# This script provides easy setup and running of the Form Service with Swagger documentation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
DEFAULT_PORT=8080
SERVICE_NAME="Form Service API"

print_header() {
    echo -e "${PURPLE}"
    echo "╔══════════════════════════════════════════════════════════════════════════════╗"
    echo "║                        Form Service API Setup                               ║"
    echo "║              Comprehensive Swagger Documentation Setup                      ║"
    echo "╚══════════════════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

print_section() {
    echo -e "${CYAN}📋 $1${NC}"
    echo "────────────────────────────────────────────────────────────"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️ $1${NC}"
}

check_prerequisites() {
    print_section "Checking Prerequisites"
    
    # Check Go installation
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | cut -d' ' -f3)
        print_success "Go is installed: $GO_VERSION"
    else
        print_error "Go is not installed. Please install Go 1.23 or higher."
        exit 1
    fi
    
    # Check if we're in the right directory
    if [[ ! -f "go.mod" ]]; then
        print_error "go.mod not found. Please run this script from the form-service directory."
        exit 1
    fi
    
    print_success "All prerequisites met"
    echo
}

setup_environment() {
    print_section "Setting Up Environment"
    
    # Create .env file if it doesn't exist
    if [[ ! -f ".env" ]]; then
        print_info "Creating .env file..."
        cat > .env << EOF
# Form Service Configuration

# Server Configuration
PORT=${DEFAULT_PORT}
NODE_ENV=development

# Database Configuration (optional for demo)
DATABASE_URL=postgresql://xform_user:xform_password@localhost:5432/xform_db?sslmode=disable

# Redis Configuration (optional for demo)
REDIS_URL=redis://localhost:6379

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-for-development-only
EOF
        print_success ".env file created with default values"
    else
        print_success ".env file already exists"
    fi
    
    # Load environment variables
    if [[ -f ".env" ]]; then
        export $(cat .env | grep -v '#' | xargs)
    fi
    
    echo
}

install_dependencies() {
    print_section "Installing Dependencies"
    
    print_info "Running go mod tidy..."
    go mod tidy
    
    print_info "Installing Swagger CLI tool..."
    go install github.com/swaggo/swag/cmd/swag@latest
    
    print_success "Dependencies installed successfully"
    echo
}

generate_swagger_docs() {
    print_section "Generating Swagger Documentation"
    
    # Clean up any empty Go files that might cause issues
    print_info "Cleaning up empty Go files..."
    find . -name "*.go" -size 0 -delete 2>/dev/null || true
    
    # Generate Swagger documentation
    print_info "Generating Swagger docs..."
    SWAG_PATH=$(go env GOPATH)/bin/swag
    
    if [[ ! -f "$SWAG_PATH" ]]; then
        print_warning "swag not found in GOPATH, trying global installation..."
        if command -v swag &> /dev/null; then
            SWAG_PATH="swag"
        else
            print_error "swag command not found. Please ensure it's installed."
            exit 1
        fi
    fi
    
    $SWAG_PATH init -g cmd/demo-swagger-server/main.go -o docs
    
    if [[ -d "docs" && -f "docs/docs.go" ]]; then
        print_success "Swagger documentation generated successfully"
    else
        print_error "Failed to generate Swagger documentation"
        exit 1
    fi
    
    echo
}

build_application() {
    print_section "Building Application"
    
    # Create bin directory if it doesn't exist
    mkdir -p bin
    
    # Build demo server
    print_info "Building demo server..."
    go build -o bin/demo-swagger-server cmd/demo-swagger-server/main.go
    
    # Build full server
    print_info "Building full server..."
    go build -o bin/full-swagger-server cmd/full-swagger-server/main.go
    
    print_success "Application built successfully"
    echo
}

run_server() {
    local server_type=$1
    
    if [[ "$server_type" == "demo" ]]; then
        print_section "Starting Demo Server (No Database Required)"
        print_info "This server runs without external dependencies and provides full Swagger documentation"
        
        # Check if port is available
        if lsof -Pi :${PORT:-$DEFAULT_PORT} -sTCP:LISTEN -t >/dev/null ; then
            print_warning "Port ${PORT:-$DEFAULT_PORT} is already in use"
            read -p "Kill existing process and continue? (y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                lsof -ti:${PORT:-$DEFAULT_PORT} | xargs kill
                sleep 2
            else
                print_error "Cannot start server on port ${PORT:-$DEFAULT_PORT}"
                exit 1
            fi
        fi
        
        echo
        print_success "🚀 Starting Form Service API Demo Server..."
        echo
        print_info "📖 Swagger Documentation: http://localhost:${PORT:-$DEFAULT_PORT}/swagger/index.html"
        print_info "🔍 Health Check: http://localhost:${PORT:-$DEFAULT_PORT}/health"
        print_info "📊 Service Info: http://localhost:${PORT:-$DEFAULT_PORT}/"
        print_info "🌐 API Base URL: http://localhost:${PORT:-$DEFAULT_PORT}/api/v1"
        echo
        print_info "Press Ctrl+C to stop the server"
        echo
        
        ./bin/demo-swagger-server
        
    elif [[ "$server_type" == "full" ]]; then
        print_section "Starting Full Production Server"
        print_warning "This server requires PostgreSQL and Redis to be running"
        
        # Check for database URL
        if [[ -z "$DATABASE_URL" ]]; then
            print_error "DATABASE_URL not set. Please configure your database connection in .env"
            exit 1
        fi
        
        echo
        print_success "🚀 Starting Form Service API Production Server..."
        echo
        print_info "📖 Swagger Documentation: http://localhost:${PORT:-$DEFAULT_PORT}/swagger/index.html"
        print_info "🔍 Health Check: http://localhost:${PORT:-$DEFAULT_PORT}/health"
        print_info "📊 Service Info: http://localhost:${PORT:-$DEFAULT_PORT}/"
        print_info "🌐 API Base URL: http://localhost:${PORT:-$DEFAULT_PORT}/api/v1"
        echo
        print_info "Press Ctrl+C to stop the server"
        echo
        
        ./bin/full-swagger-server
        
    else
        print_error "Invalid server type. Use 'demo' or 'full'"
        exit 1
    fi
}

show_help() {
    print_header
    echo -e "${CYAN}Usage:${NC}"
    echo "  $0 [command]"
    echo
    echo -e "${CYAN}Commands:${NC}"
    echo "  setup     - Complete setup (install deps, generate docs, build)"
    echo "  demo      - Run demo server (no database required)"
    echo "  full      - Run full production server (requires database)"
    echo "  build     - Build application binaries"
    echo "  docs      - Generate Swagger documentation only"
    echo "  clean     - Clean build artifacts"
    echo "  help      - Show this help message"
    echo
    echo -e "${CYAN}Examples:${NC}"
    echo "  $0 setup     # Complete setup and build"
    echo "  $0 demo      # Run demo server"
    echo "  $0 full      # Run production server"
    echo
    echo -e "${CYAN}Environment Variables:${NC}"
    echo "  PORT              - Server port (default: 8080)"
    echo "  DATABASE_URL      - PostgreSQL connection string"
    echo "  REDIS_URL         - Redis connection string"
    echo "  JWT_SECRET        - JWT signing secret"
    echo
}

cleanup() {
    print_section "Cleaning Up"
    
    # Remove build artifacts
    rm -rf bin/
    rm -rf docs/
    
    # Remove generated files
    find . -name "*.log" -delete 2>/dev/null || true
    
    print_success "Cleanup completed"
}

main() {
    case "${1:-help}" in
        setup)
            print_header
            check_prerequisites
            setup_environment
            install_dependencies
            generate_swagger_docs
            build_application
            echo
            print_success "🎉 Setup completed successfully!"
            echo
            print_info "Next steps:"
            echo "  • Run demo server: $0 demo"
            echo "  • Run full server: $0 full"
            echo "  • View documentation: http://localhost:${PORT:-$DEFAULT_PORT}/swagger/index.html"
            ;;
        demo)
            check_prerequisites
            setup_environment
            if [[ ! -f "bin/demo-swagger-server" ]]; then
                print_warning "Demo server not built. Building now..."
                install_dependencies
                generate_swagger_docs
                build_application
            fi
            run_server "demo"
            ;;
        full)
            check_prerequisites
            setup_environment
            if [[ ! -f "bin/full-swagger-server" ]]; then
                print_warning "Full server not built. Building now..."
                install_dependencies
                generate_swagger_docs
                build_application
            fi
            run_server "full"
            ;;
        build)
            print_header
            check_prerequisites
            setup_environment
            install_dependencies
            generate_swagger_docs
            build_application
            print_success "Build completed successfully!"
            ;;
        docs)
            print_header
            check_prerequisites
            install_dependencies
            generate_swagger_docs
            print_success "Documentation generated successfully!"
            ;;
        clean)
            print_header
            cleanup
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "Unknown command: $1"
            echo
            show_help
            exit 1
            ;;
    esac
}

# Handle Ctrl+C gracefully
trap 'echo -e "\n${YELLOW}🛑 Shutting down gracefully...${NC}"; exit 0' INT

# Run main function
main "$@"
