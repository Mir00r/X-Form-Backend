#!/bin/bash

# Event Bus Service Observability Stack Deployment Script
# This script deploys the complete observability stack for the Event Bus Service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="event-bus-observability"
COMPOSE_FILE="docker-compose.observability.yml"
GRAFANA_DASHBOARDS_DIR="deployments/grafana/dashboards"
GRAFANA_PROVISIONING_DIR="deployments/grafana/provisioning"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    if ! command -v curl &> /dev/null; then
        log_error "curl is not installed. Please install curl first."
        exit 1
    fi
    
    log_success "All prerequisites are met!"
}

# Create necessary directories
create_directories() {
    log_info "Creating necessary directories..."
    
    mkdir -p $GRAFANA_DASHBOARDS_DIR
    mkdir -p $GRAFANA_PROVISIONING_DIR/dashboards
    mkdir -p $GRAFANA_PROVISIONING_DIR/datasources
    mkdir -p deployments/grafana/templates
    
    log_success "Directories created!"
}

# Create Grafana provisioning files
create_grafana_config() {
    log_info "Creating Grafana configuration files..."
    
    # Datasources configuration
    cat > $GRAFANA_PROVISIONING_DIR/datasources/datasources.yml << EOF
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true

  - name: Tempo
    type: tempo
    access: proxy
    url: http://tempo:3200
    editable: true
    jsonData:
      tracesToLogs:
        datasourceUid: 'loki'
        tags: ['job', 'instance', 'pod', 'namespace']
        mappedTags: [{ key: 'service.name', value: 'service' }]
        mapTagNamesEnabled: false
        spanStartTimeShift: '1h'
        spanEndTimeShift: '1h'
        filterByTraceID: false
        filterBySpanID: false
      serviceMap:
        datasourceUid: 'prometheus'
      search:
        hide: false
      nodeGraph:
        enabled: true
EOF

    # Dashboard provisioning
    cat > $GRAFANA_PROVISIONING_DIR/dashboards/dashboards.yml << EOF
apiVersion: 1

providers:
  - name: 'Event Bus Dashboards'
    orgId: 1
    folder: 'Event Bus Service'
    type: file
    disableDeletion: false
    editable: true
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /var/lib/grafana/dashboards
EOF

    log_success "Grafana configuration created!"
}

# Build the application
build_application() {
    log_info "Building Event Bus Service application..."
    
    if [ ! -f "go.mod" ]; then
        log_error "go.mod not found. Please run this script from the event-bus-service directory."
        exit 1
    fi
    
    # Build the application
    go mod tidy
    go build -o bin/event-bus-service ./cmd/demo/
    
    log_success "Application built successfully!"
}

# Create Dockerfile if it doesn't exist
create_dockerfile() {
    if [ ! -f "Dockerfile" ]; then
        log_info "Creating Dockerfile..."
        
        cat > Dockerfile << EOF
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o event-bus-service ./cmd/demo/

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl
WORKDIR /root/

COPY --from=builder /app/event-bus-service .

EXPOSE 8080 9090

CMD ["./event-bus-service"]
EOF
        
        log_success "Dockerfile created!"
    fi
}

# Deploy the stack
deploy_stack() {
    log_info "Deploying observability stack..."
    
    # Stop existing containers
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME down 2>/dev/null || true
    
    # Start the stack
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d
    
    log_success "Observability stack deployed!"
}

# Wait for services to be ready
wait_for_services() {
    log_info "Waiting for services to be ready..."
    
    # Wait for Prometheus
    log_info "Waiting for Prometheus..."
    timeout 120 bash -c 'until curl -s http://localhost:9091/-/ready; do sleep 2; done' || {
        log_error "Prometheus failed to start"
        exit 1
    }
    
    # Wait for Grafana
    log_info "Waiting for Grafana..."
    timeout 120 bash -c 'until curl -s http://localhost:3000/api/health; do sleep 2; done' || {
        log_error "Grafana failed to start"
        exit 1
    }
    
    # Wait for Event Bus Service
    log_info "Waiting for Event Bus Service..."
    timeout 120 bash -c 'until curl -s http://localhost:8080/health; do sleep 2; done' || {
        log_error "Event Bus Service failed to start"
        exit 1
    }
    
    # Wait for Jaeger
    log_info "Waiting for Jaeger..."
    timeout 120 bash -c 'until curl -s http://localhost:16686; do sleep 2; done' || {
        log_warning "Jaeger might not be ready yet"
    }
    
    log_success "All services are ready!"
}

# Generate sample data
generate_sample_data() {
    log_info "Generating sample data..."
    
    # Generate some sample requests
    for i in {1..10}; do
        curl -s "http://localhost:8080/api/events?type=sample&source=deployment" > /dev/null
        curl -s "http://localhost:8080/api/kafka?operation=produce&topic=test-topic" > /dev/null
        curl -s "http://localhost:8080/api/cdc?connector=postgres&table=users&operation=insert" > /dev/null
        sleep 1
    done
    
    # Generate an error for testing
    curl -s "http://localhost:8080/api/error?type=demo" > /dev/null
    
    log_success "Sample data generated!"
}

# Display service URLs
show_service_urls() {
    log_success "Observability stack is ready!"
    echo ""
    echo -e "${GREEN}Service URLs:${NC}"
    echo -e "${BLUE}Event Bus Service:${NC}     http://localhost:8080"
    echo -e "${BLUE}Health Check:${NC}          http://localhost:8080/health"
    echo -e "${BLUE}Metrics:${NC}               http://localhost:8080/metrics"
    echo -e "${BLUE}Prometheus:${NC}            http://localhost:9091"
    echo -e "${BLUE}Grafana:${NC}               http://localhost:3000 (admin/admin)"
    echo -e "${BLUE}Jaeger UI:${NC}             http://localhost:16686"
    echo -e "${BLUE}AlertManager:${NC}          http://localhost:9093"
    echo ""
    echo -e "${GREEN}API Test Endpoints:${NC}"
    echo -e "${BLUE}Events:${NC}                curl 'http://localhost:8080/api/events?type=demo&source=test'"
    echo -e "${BLUE}Kafka:${NC}                 curl 'http://localhost:8080/api/kafka?operation=produce&topic=test'"
    echo -e "${BLUE}CDC:${NC}                   curl 'http://localhost:8080/api/cdc?connector=postgres&table=users&operation=insert'"
    echo -e "${BLUE}Error Test:${NC}            curl 'http://localhost:8080/api/error?type=demo'"
    echo ""
    echo -e "${YELLOW}Quick Test:${NC}"
    echo -e "Run: ${BLUE}./test-observability.sh${NC} (if available) to generate test data"
}

# Create test script
create_test_script() {
    log_info "Creating test script..."
    
    cat > test-observability.sh << 'EOF'
#!/bin/bash

echo "Testing Event Bus Service Observability..."

# Test different endpoints
echo "Testing event processing..."
for i in {1..5}; do
    curl -s "http://localhost:8080/api/events?type=load_test&source=script" > /dev/null
    echo -n "."
done
echo " Done!"

echo "Testing Kafka operations..."
for i in {1..5}; do
    curl -s "http://localhost:8080/api/kafka?operation=produce&topic=load-test" > /dev/null
    echo -n "."
done
echo " Done!"

echo "Testing CDC operations..."
for i in {1..3}; do
    curl -s "http://localhost:8080/api/cdc?connector=postgres&table=load_test&operation=update" > /dev/null
    echo -n "."
done
echo " Done!"

echo "Testing error handling..."
curl -s "http://localhost:8080/api/error?type=load_test" > /dev/null
echo " Done!"

echo "Load test completed! Check your observability dashboards:"
echo "- Prometheus: http://localhost:9091"
echo "- Grafana: http://localhost:3000"
echo "- Jaeger: http://localhost:16686"
echo "- Metrics: http://localhost:8080/metrics"
EOF

    chmod +x test-observability.sh
    log_success "Test script created!"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME down -v 2>/dev/null || true
    log_success "Cleanup completed!"
}

# Show help
show_help() {
    echo "Event Bus Service Observability Deployment Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  deploy    Deploy the complete observability stack (default)"
    echo "  cleanup   Stop and remove all containers and volumes"
    echo "  restart   Restart the observability stack"
    echo "  logs      Show logs from all services"
    echo "  status    Show status of all services"
    echo "  help      Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 deploy     # Deploy the stack"
    echo "  $0 cleanup    # Clean up everything"
    echo "  $0 restart    # Restart the stack"
}

# Show logs
show_logs() {
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f
}

# Show status
show_status() {
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME ps
}

# Main execution
main() {
    case "${1:-deploy}" in
        "deploy")
            check_prerequisites
            create_directories
            create_grafana_config
            create_dockerfile
            build_application
            deploy_stack
            wait_for_services
            generate_sample_data
            create_test_script
            show_service_urls
            ;;
        "cleanup")
            cleanup
            ;;
        "restart")
            cleanup
            main deploy
            ;;
        "logs")
            show_logs
            ;;
        "status")
            show_status
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            log_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
