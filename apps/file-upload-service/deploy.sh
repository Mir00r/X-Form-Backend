#!/bin/bash

# File Upload Service Deployment Script
# Builds and deploys the Lambda function to AWS

set -e  # Exit on any error

# Configuration
SERVICE_NAME="file-upload-service"
AWS_REGION=${AWS_REGION:-"us-east-1"}
ENVIRONMENT=${ENVIRONMENT:-"dev"}
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

echo_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

echo_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    echo_info "Checking prerequisites..."
    
    if ! command -v aws &> /dev/null; then
        echo_error "AWS CLI is not installed"
        exit 1
    fi
    
    if ! command -v docker &> /dev/null; then
        echo_error "Docker is not installed"
        exit 1
    fi
    
    if ! command -v terraform &> /dev/null; then
        echo_warning "Terraform is not installed (optional for manual deployment)"
    fi
    
    if ! aws sts get-caller-identity &> /dev/null; then
        echo_error "AWS credentials not configured"
        exit 1
    fi
    
    echo_success "Prerequisites check passed"
}

# Build Docker image
build_image() {
    echo_info "Building Docker image..."
    
    ECR_REPO="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${SERVICE_NAME}"
    
    # Build the image
    docker build -t ${SERVICE_NAME}:latest .
    
    # Tag for ECR
    docker tag ${SERVICE_NAME}:latest ${ECR_REPO}:latest
    
    echo_success "Docker image built successfully"
}

# Push to ECR
push_to_ecr() {
    echo_info "Pushing to ECR..."
    
    ECR_REPO="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${SERVICE_NAME}"
    
    # Create ECR repository if it doesn't exist
    aws ecr describe-repositories --repository-names ${SERVICE_NAME} --region ${AWS_REGION} 2>/dev/null || {
        echo_info "Creating ECR repository..."
        aws ecr create-repository --repository-name ${SERVICE_NAME} --region ${AWS_REGION}
    }
    
    # Get login token and login to ECR
    aws ecr get-login-password --region ${AWS_REGION} | docker login --username AWS --password-stdin ${ECR_REPO}
    
    # Push the image
    docker push ${ECR_REPO}:latest
    
    echo_success "Image pushed to ECR successfully"
}

# Deploy infrastructure with Terraform
deploy_infrastructure() {
    echo_info "Deploying infrastructure with Terraform..."
    
    if ! command -v terraform &> /dev/null; then
        echo_warning "Terraform not available, skipping infrastructure deployment"
        echo_warning "Please deploy infrastructure manually or install Terraform"
        return
    fi
    
    cd infrastructure/
    
    # Initialize Terraform
    terraform init
    
    # Plan the deployment
    terraform plan \
        -var="service_name=${SERVICE_NAME}" \
        -var="environment=${ENVIRONMENT}" \
        -var="aws_region=${AWS_REGION}" \
        -var="jwt_secret=${JWT_SECRET:-development-secret-key}"
    
    # Apply the configuration
    echo_info "Applying Terraform configuration..."
    terraform apply -auto-approve \
        -var="service_name=${SERVICE_NAME}" \
        -var="environment=${ENVIRONMENT}" \
        -var="aws_region=${AWS_REGION}" \
        -var="jwt_secret=${JWT_SECRET:-development-secret-key}"
    
    cd ..
    
    echo_success "Infrastructure deployed successfully"
}

# Deploy Lambda function (alternative to Terraform)
deploy_lambda() {
    echo_info "Deploying Lambda function..."
    
    FUNCTION_NAME="${SERVICE_NAME}-${ENVIRONMENT}"
    ECR_URI="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${SERVICE_NAME}:latest"
    
    # Check if function exists
    if aws lambda get-function --function-name ${FUNCTION_NAME} --region ${AWS_REGION} 2>/dev/null; then
        echo_info "Updating existing Lambda function..."
        aws lambda update-function-code \
            --function-name ${FUNCTION_NAME} \
            --image-uri ${ECR_URI} \
            --region ${AWS_REGION}
    else
        echo_info "Creating new Lambda function..."
        
        # Create execution role if it doesn't exist
        ROLE_ARN=$(aws iam get-role --role-name ${SERVICE_NAME}-lambda-role-${ENVIRONMENT} --query 'Role.Arn' --output text 2>/dev/null || echo "")
        
        if [ -z "$ROLE_ARN" ]; then
            echo_info "Creating IAM role..."
            aws iam create-role \
                --role-name ${SERVICE_NAME}-lambda-role-${ENVIRONMENT} \
                --assume-role-policy-document '{
                    "Version": "2012-10-17",
                    "Statement": [
                        {
                            "Effect": "Allow",
                            "Principal": {
                                "Service": "lambda.amazonaws.com"
                            },
                            "Action": "sts:AssumeRole"
                        }
                    ]
                }'
            
            # Attach basic execution policy
            aws iam attach-role-policy \
                --role-name ${SERVICE_NAME}-lambda-role-${ENVIRONMENT} \
                --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
            
            ROLE_ARN="arn:aws:iam::${AWS_ACCOUNT_ID}:role/${SERVICE_NAME}-lambda-role-${ENVIRONMENT}"
            
            echo_info "Waiting for role to be ready..."
            sleep 10
        fi
        
        # Create the function
        aws lambda create-function \
            --function-name ${FUNCTION_NAME} \
            --role ${ROLE_ARN} \
            --code ImageUri=${ECR_URI} \
            --package-type Image \
            --timeout 30 \
            --memory-size 512 \
            --region ${AWS_REGION} \
            --environment Variables="{
                AWS_REGION=${AWS_REGION},
                S3_BUCKET_NAME=${SERVICE_NAME}-uploads-${ENVIRONMENT}-${AWS_ACCOUNT_ID},
                DYNAMODB_TABLE_NAME=${SERVICE_NAME}-upload-requests-${ENVIRONMENT},
                JWT_SECRET=${JWT_SECRET:-development-secret-key},
                LOG_LEVEL=INFO
            }"
    fi
    
    echo_success "Lambda function deployed successfully"
}

# Run tests
run_tests() {
    echo_info "Running tests..."
    
    if [ -f "requirements-dev.txt" ]; then
        pip install -r requirements-dev.txt
    else
        pip install pytest pytest-asyncio pytest-mock
    fi
    
    python -m pytest tests/ -v
    
    echo_success "Tests completed"
}

# Main deployment flow
main() {
    echo_info "Starting deployment of ${SERVICE_NAME}..."
    echo_info "Environment: ${ENVIRONMENT}"
    echo_info "AWS Region: ${AWS_REGION}"
    echo_info "AWS Account: ${AWS_ACCOUNT_ID}"
    
    # Parse command line arguments
    SKIP_TESTS=false
    SKIP_INFRASTRUCTURE=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-tests)
                SKIP_TESTS=true
                shift
                ;;
            --skip-infrastructure)
                SKIP_INFRASTRUCTURE=true
                shift
                ;;
            --help)
                echo "Usage: $0 [--skip-tests] [--skip-infrastructure]"
                echo ""
                echo "Options:"
                echo "  --skip-tests           Skip running tests"
                echo "  --skip-infrastructure  Skip Terraform infrastructure deployment"
                echo ""
                echo "Environment variables:"
                echo "  AWS_REGION    AWS region (default: us-east-1)"
                echo "  ENVIRONMENT   Deployment environment (default: dev)"
                echo "  JWT_SECRET    JWT secret key for authentication"
                exit 0
                ;;
            *)
                echo_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Execute deployment steps
    check_prerequisites
    
    if [ "$SKIP_TESTS" = false ]; then
        run_tests
    fi
    
    build_image
    push_to_ecr
    
    if [ "$SKIP_INFRASTRUCTURE" = false ]; then
        deploy_infrastructure
    else
        deploy_lambda
    fi
    
    echo_success "Deployment completed successfully!"
    echo_info "Function ARN: arn:aws:lambda:${AWS_REGION}:${AWS_ACCOUNT_ID}:function:${SERVICE_NAME}-${ENVIRONMENT}"
    
    # Test the deployment
    echo_info "Testing deployment..."
    aws lambda invoke \
        --function-name ${SERVICE_NAME}-${ENVIRONMENT} \
        --payload '{"httpMethod": "GET", "path": "/health", "headers": {}}' \
        --region ${AWS_REGION} \
        response.json
    
    if [ $? -eq 0 ]; then
        echo_success "Health check passed!"
        cat response.json | python -m json.tool
        rm response.json
    else
        echo_error "Health check failed"
    fi
}

# Run main function
main "$@"
