#!/bin/bash

# X-Form-Backend Infrastructure Deployment Script
# This script deploys the complete infrastructure and applications

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT="dev"
DEPLOY_INFRA="false"
DEPLOY_APPS="true"
AWS_REGION="us-west-2"
DRY_RUN="false"

# Help function
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -e, --environment ENV     Target environment (dev, staging, production)"
    echo "  -i, --infrastructure      Deploy infrastructure (Terraform)"
    echo "  -a, --apps-only          Deploy applications only (skip infrastructure)"
    echo "  -r, --region REGION       AWS region (default: us-west-2)"
    echo "  -d, --dry-run            Dry run mode (plan only, no apply)"
    echo "  -h, --help               Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 -e dev -i                    # Deploy dev infrastructure and apps"
    echo "  $0 -e production -a             # Deploy production apps only"
    echo "  $0 -e staging -i -d             # Dry run for staging infrastructure"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -i|--infrastructure)
            DEPLOY_INFRA="true"
            shift
            ;;
        -a|--apps-only)
            DEPLOY_INFRA="false"
            DEPLOY_APPS="true"
            shift
            ;;
        -r|--region)
            AWS_REGION="$2"
            shift 2
            ;;
        -d|--dry-run)
            DRY_RUN="true"
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "Unknown option $1"
            show_help
            exit 1
            ;;
    esac
done

# Validate environment
if [[ ! "$ENVIRONMENT" =~ ^(dev|staging|production)$ ]]; then
    echo -e "${RED}❌ Invalid environment: $ENVIRONMENT${NC}"
    echo "Valid environments: dev, staging, production"
    exit 1
fi

echo -e "${BLUE}🚀 Starting X-Form-Backend deployment...${NC}"
echo -e "${BLUE}Environment: ${GREEN}$ENVIRONMENT${NC}"
echo -e "${BLUE}AWS Region: ${GREEN}$AWS_REGION${NC}"
echo -e "${BLUE}Deploy Infrastructure: ${GREEN}$DEPLOY_INFRA${NC}"
echo -e "${BLUE}Deploy Applications: ${GREEN}$DEPLOY_APPS${NC}"
echo -e "${BLUE}Dry Run: ${GREEN}$DRY_RUN${NC}"
echo ""

# Check prerequisites
echo -e "${BLUE}📋 Checking prerequisites...${NC}"

# Check AWS CLI
if ! command -v aws &> /dev/null; then
    echo -e "${RED}❌ AWS CLI is not installed${NC}"
    exit 1
fi

# Check Terraform
if [[ "$DEPLOY_INFRA" == "true" ]] && ! command -v terraform &> /dev/null; then
    echo -e "${RED}❌ Terraform is not installed${NC}"
    exit 1
fi

# Check Helm
if [[ "$DEPLOY_APPS" == "true" ]] && ! command -v helm &> /dev/null; then
    echo -e "${RED}❌ Helm is not installed${NC}"
    exit 1
fi

# Check kubectl
if [[ "$DEPLOY_APPS" == "true" ]] && ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl is not installed${NC}"
    exit 1
fi

# Check AWS credentials
if ! aws sts get-caller-identity &> /dev/null; then
    echo -e "${RED}❌ AWS credentials not configured${NC}"
    exit 1
fi

echo -e "${GREEN}✅ All prerequisites satisfied${NC}"

# Deploy infrastructure
if [[ "$DEPLOY_INFRA" == "true" ]]; then
    echo -e "${BLUE}🏗️  Deploying infrastructure...${NC}"
    
    cd infrastructure/terraform
    
    # Initialize Terraform
    echo -e "${BLUE}🔧 Initializing Terraform...${NC}"
    terraform init
    
    # Plan infrastructure
    echo -e "${BLUE}📋 Planning infrastructure changes...${NC}"
    terraform plan \
        -var-file="${ENVIRONMENT}.tfvars" \
        -var="aws_region=${AWS_REGION}" \
        -out=tfplan
    
    if [[ "$DRY_RUN" == "false" ]]; then
        # Apply infrastructure
        echo -e "${BLUE}🚀 Applying infrastructure changes...${NC}"
        terraform apply tfplan
        
        # Output infrastructure details
        echo -e "${GREEN}✅ Infrastructure deployed successfully${NC}"
        terraform output
    else
        echo -e "${YELLOW}🔍 Dry run completed - no changes applied${NC}"
    fi
    
    cd ../..
fi

# Deploy applications
if [[ "$DEPLOY_APPS" == "true" && "$DRY_RUN" == "false" ]]; then
    echo -e "${BLUE}📦 Deploying applications...${NC}"
    
    # Update kubeconfig
    echo -e "${BLUE}🔧 Updating kubeconfig...${NC}"
    aws eks update-kubeconfig \
        --region "${AWS_REGION}" \
        --name "x-form-backend-${ENVIRONMENT}"
    
    # Create namespace
    echo -e "${BLUE}📁 Creating namespace...${NC}"
    kubectl create namespace "x-form-backend-${ENVIRONMENT}" --dry-run=client -o yaml | kubectl apply -f -
    
    cd infrastructure/helm/x-form-backend
    
    # Add helm repositories
    echo -e "${BLUE}📚 Adding Helm repositories...${NC}"
    helm repo add bitnami https://charts.bitnami.com/bitnami
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo update
    
    # Deploy observability stack
    echo -e "${BLUE}📊 Deploying observability stack...${NC}"
    
    # Deploy Prometheus
    helm upgrade --install prometheus prometheus-community/kube-prometheus-stack \
        --namespace monitoring \
        --create-namespace \
        --set grafana.adminPassword="${GRAFANA_ADMIN_PASSWORD:-admin123}" \
        --wait
    
    # Deploy application
    echo -e "${BLUE}🚀 Deploying X-Form-Backend application...${NC}"
    
    helm upgrade --install "x-form-backend-${ENVIRONMENT}" . \
        --namespace "x-form-backend-${ENVIRONMENT}" \
        --values values.yaml \
        --values "values-${ENVIRONMENT}.yaml" \
        --set image.tag="${IMAGE_TAG:-latest}" \
        --wait --timeout=600s
    
    # Verify deployment
    echo -e "${BLUE}🔍 Verifying deployment...${NC}"
    kubectl get pods -n "x-form-backend-${ENVIRONMENT}"
    kubectl get services -n "x-form-backend-${ENVIRONMENT}"
    kubectl get ingress -n "x-form-backend-${ENVIRONMENT}"
    
    # Wait for pods to be ready
    echo -e "${BLUE}⏳ Waiting for pods to be ready...${NC}"
    kubectl wait --for=condition=ready pod \
        -l app.kubernetes.io/name=x-form-backend \
        -n "x-form-backend-${ENVIRONMENT}" \
        --timeout=300s
    
    echo -e "${GREEN}✅ Application deployed successfully${NC}"
    
    cd ../../..
fi

# Display access information
if [[ "$DEPLOY_APPS" == "true" && "$DRY_RUN" == "false" ]]; then
    echo -e "${BLUE}🌐 Access Information:${NC}"
    
    # Get ingress information
    INGRESS_HOST=$(kubectl get ingress -n "x-form-backend-${ENVIRONMENT}" -o jsonpath='{.items[0].spec.rules[0].host}')
    if [[ -n "$INGRESS_HOST" ]]; then
        echo -e "${GREEN}🔗 Application URL: https://${INGRESS_HOST}${NC}"
    fi
    
    # Get Grafana information
    GRAFANA_HOST=$(kubectl get ingress -n monitoring -o jsonpath='{.items[?(@.metadata.name=="prometheus-grafana")].spec.rules[0].host}' 2>/dev/null || echo "")
    if [[ -n "$GRAFANA_HOST" ]]; then
        echo -e "${GREEN}📊 Grafana Dashboard: https://${GRAFANA_HOST}${NC}"
    else
        echo -e "${GREEN}📊 Grafana Dashboard: kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80${NC}"
    fi
    
    echo -e "${GREEN}🔍 Prometheus: kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090${NC}"
fi

echo ""
echo -e "${GREEN}🎉 Deployment completed successfully!${NC}"

# Show next steps
echo -e "${BLUE}📚 Next Steps:${NC}"
echo "1. Verify all services are running: kubectl get pods -n x-form-backend-${ENVIRONMENT}"
echo "2. Check logs: kubectl logs -f deployment/x-form-backend-${ENVIRONMENT}-api-gateway -n x-form-backend-${ENVIRONMENT}"
echo "3. Test API endpoints: curl https://${INGRESS_HOST:-api.example.com}/health"
echo "4. Monitor with Grafana and Prometheus"
echo "5. Set up alerts and notifications"
