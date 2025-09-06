#!/bin/bash

# Event Bus Service - Development Startup Script
# This script provides an easy way to start the Event Bus Service for development

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if a port is available
port_available() {
    ! nc -z localhost "$1" >/dev/null 2>&1
}

# Function to wait for service to be ready
wait_for_service() {
    local url=$1
    local name=$2
    local timeout=${3:-60}
    local count=0
    
    print_status "Waiting for $name to be ready..."
    
    while [ $count -lt $timeout ]; do
        if curl -s "$url" >/dev/null 2>&1; then
            print_success "$name is ready!"
            return 0
        fi
        
        sleep 2
        count=$((count + 2))
        echo -n "."
    done
    
    print_error "$name failed to start within $timeout seconds"
    return 1
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check required commands
    local required_commands=("docker" "docker-compose" "go" "curl")
    for cmd in "${required_commands[@]}"; do
        if ! command_exists "$cmd"; then
            print_error "$cmd is not installed or not in PATH"
            exit 1
        fi
    done
    
    # Check Go version
    local go_version=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
    local required_version="1.21"
    
    if ! printf '%s\n%s\n' "$required_version" "$go_version" | sort -V -C; then
        print_error "Go $required_version or higher is required (found: $go_version)"
        exit 1
    fi
    
    # Check available ports
    local required_ports=(8080 9090 5432 6379 9092 8083 8081 9091 3000)
    for port in "${required_ports[@]}"; do
        if ! port_available "$port"; then
            print_warning "Port $port is already in use"
        fi
    done
    
    print_success "Prerequisites check completed"
}

# Function to setup environment
setup_environment() {
    print_status "Setting up environment..."
    
    # Create necessary directories
    mkdir -p logs config/debezium monitoring/grafana scripts
    
    # Set environment variables
    export GO_ENV=${GO_ENV:-development}
    export LOG_LEVEL=${LOG_LEVEL:-info}
    export LOG_FORMAT=${LOG_FORMAT:-json}
    
    print_success "Environment setup completed"
}

# Function to start infrastructure services
start_infrastructure() {
    print_status "Starting infrastructure services..."
    
    # Start PostgreSQL, Redis, Kafka, and Debezium
    docker-compose up -d postgres redis zookeeper kafka debezium
    
    # Wait for services to be ready
    wait_for_service "http://localhost:5432" "PostgreSQL" 30 || true
    wait_for_service "http://localhost:6379" "Redis" 30 || true
    wait_for_service "http://localhost:9092" "Kafka" 60 || true
    wait_for_service "http://localhost:8083/connectors" "Debezium" 90
    
    print_success "Infrastructure services started"
}

# Function to setup Debezium connector
setup_debezium_connector() {
    print_status "Setting up Debezium PostgreSQL connector..."
    
    # Wait a bit more for Debezium to be fully ready
    sleep 10
    
    # Create connector configuration
    local connector_config='{
        "name": "postgres-connector",
        "config": {
            "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
            "database.hostname": "postgres",
            "database.port": "5432",
            "database.user": "eventbus",
            "database.password": "eventbus_password",
            "database.dbname": "eventbus",
            "database.server.name": "eventbus",
            "table.include.list": "public.forms,public.responses,public.analytics",
            "plugin.name": "pgoutput",
            "slot.name": "eventbus_slot",
            "publication.name": "eventbus_publication",
            "key.converter": "org.apache.kafka.connect.json.JsonConverter",
            "value.converter": "org.apache.kafka.connect.json.JsonConverter",
            "key.converter.schemas.enable": "false",
            "value.converter.schemas.enable": "false",
            "transforms": "route",
            "transforms.route.type": "org.apache.kafka.connect.transforms.RegexRouter",
            "transforms.route.regex": "([^.]+)\\.([^.]+)\\.([^.]+)",
            "transforms.route.replacement": "cdc.$3"
        }
    }'
    
    # Create the connector
    if curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$connector_config" \
        http://localhost:8083/connectors; then
        print_success "Debezium connector created successfully"
    else
        print_warning "Failed to create Debezium connector (may already exist)"
    fi
}

# Function to build the service
build_service() {
    print_status "Building Event Bus Service..."
    
    # Download dependencies
    go mod download
    
    # Build the service
    if go build -o bin/event-bus-service cmd/server/main.go; then
        print_success "Event Bus Service built successfully"
    else
        print_error "Failed to build Event Bus Service"
        exit 1
    fi
}

# Function to start the service
start_service() {
    print_status "Starting Event Bus Service..."
    
    # Start the service in background
    nohup ./bin/event-bus-service > logs/event-bus.log 2>&1 &
    local service_pid=$!
    
    # Save PID for later cleanup
    echo $service_pid > logs/event-bus.pid
    
    # Wait for service to be ready
    if wait_for_service "http://localhost:8080/health" "Event Bus Service" 30; then
        print_success "Event Bus Service started successfully (PID: $service_pid)"
        print_status "Service logs: tail -f logs/event-bus.log"
    else
        print_error "Event Bus Service failed to start"
        exit 1
    fi
}

# Function to start monitoring services
start_monitoring() {
    print_status "Starting monitoring services..."
    
    # Start Prometheus and Grafana
    docker-compose up -d prometheus grafana kafka-ui
    
    # Wait for services
    wait_for_service "http://localhost:9091" "Prometheus" 30 || true
    wait_for_service "http://localhost:3000" "Grafana" 30 || true
    wait_for_service "http://localhost:8081" "Kafka UI" 30 || true
    
    print_success "Monitoring services started"
    print_status "Kafka UI: http://localhost:8081"
    print_status "Prometheus: http://localhost:9091"
    print_status "Grafana: http://localhost:3000 (admin/admin)"
}

# Function to show service status
show_status() {
    print_status "Service Status:"
    echo ""
    
    # Check Event Bus Service
    if curl -s http://localhost:8080/health >/dev/null 2>&1; then
        echo "✅ Event Bus Service: http://localhost:8080"
    else
        echo "❌ Event Bus Service: Not running"
    fi
    
    # Check infrastructure services
    if docker-compose ps postgres | grep -q "Up"; then
        echo "✅ PostgreSQL: localhost:5432"
    else
        echo "❌ PostgreSQL: Not running"
    fi
    
    if docker-compose ps redis | grep -q "Up"; then
        echo "✅ Redis: localhost:6379"
    else
        echo "❌ Redis: Not running"
    fi
    
    if docker-compose ps kafka | grep -q "Up"; then
        echo "✅ Kafka: localhost:9092"
    else
        echo "❌ Kafka: Not running"
    fi
    
    if curl -s http://localhost:8083/connectors >/dev/null 2>&1; then
        echo "✅ Debezium: http://localhost:8083"
    else
        echo "❌ Debezium: Not running"
    fi
    
    # Check monitoring services
    if curl -s http://localhost:8081 >/dev/null 2>&1; then
        echo "✅ Kafka UI: http://localhost:8081"
    else
        echo "❌ Kafka UI: Not running"
    fi
    
    if curl -s http://localhost:9091 >/dev/null 2>&1; then
        echo "✅ Prometheus: http://localhost:9091"
    else
        echo "❌ Prometheus: Not running"
    fi
    
    if curl -s http://localhost:3000 >/dev/null 2>&1; then
        echo "✅ Grafana: http://localhost:3000"
    else
        echo "❌ Grafana: Not running"
    fi
    
    echo ""
    print_status "API Endpoints:"
    echo "  Health Check: curl http://localhost:8080/health"
    echo "  Version Info: curl http://localhost:8080/version"
    echo "  Publish Event: curl -X POST http://localhost:8080/events -d '{...}'"
    echo "  Metrics: curl http://localhost:9090/metrics"
}

# Function to stop services
stop_services() {
    print_status "Stopping services..."
    
    # Stop Event Bus Service
    if [ -f logs/event-bus.pid ]; then
        local pid=$(cat logs/event-bus.pid)
        if kill -0 "$pid" 2>/dev/null; then
            kill "$pid"
            print_success "Event Bus Service stopped"
        fi
        rm -f logs/event-bus.pid
    fi
    
    # Stop Docker services
    docker-compose down
    
    print_success "All services stopped"
}

# Function to clean up
cleanup() {
    print_status "Cleaning up..."
    
    # Stop services
    stop_services
    
    # Remove volumes (optional)
    if [ "$1" = "--volumes" ]; then
        docker-compose down -v
        print_status "Docker volumes removed"
    fi
    
    # Clean build artifacts
    rm -rf bin/ logs/*.log
    
    print_success "Cleanup completed"
}

# Function to run tests
run_tests() {
    print_status "Running tests..."
    
    # Unit tests
    go test -v ./...
    
    # Integration tests (if infrastructure is running)
    if curl -s http://localhost:8080/health >/dev/null 2>&1; then
        go test -tags=integration -v ./...
    else
        print_warning "Skipping integration tests (service not running)"
    fi
    
    print_success "Tests completed"
}

# Function to show help
show_help() {
    echo "Event Bus Service - Development Script"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  start       Start all services (default)"
    echo "  stop        Stop all services"
    echo "  restart     Restart all services"
    echo "  status      Show service status"
    echo "  logs        Show service logs"
    echo "  test        Run tests"
    echo "  build       Build the service only"
    echo "  clean       Clean up (add --volumes to remove data)"
    echo "  help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 start          # Start all services"
    echo "  $0 status         # Check service status"
    echo "  $0 clean --volumes # Clean up including data volumes"
}

# Main function
main() {
    local command=${1:-start}
    
    case $command in
        start)
            check_prerequisites
            setup_environment
            start_infrastructure
            setup_debezium_connector
            build_service
            start_service
            start_monitoring
            echo ""
            show_status
            ;;
        stop)
            stop_services
            ;;
        restart)
            stop_services
            sleep 2
            main start
            ;;
        status)
            show_status
            ;;
        logs)
            if [ -f logs/event-bus.log ]; then
                tail -f logs/event-bus.log
            else
                print_error "Log file not found. Is the service running?"
            fi
            ;;
        test)
            run_tests
            ;;
        build)
            check_prerequisites
            build_service
            ;;
        clean)
            cleanup "$2"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

# Trap to handle script interruption
trap 'print_warning "Script interrupted"; exit 1' INT TERM

# Run main function with all arguments
main "$@"
