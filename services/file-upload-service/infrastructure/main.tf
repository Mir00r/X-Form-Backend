"""
Terraform Configuration for File Upload Service Infrastructure

This creates all necessary AWS resources for the file upload service
"""

# Configure the AWS Provider
terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# Variables
variable "service_name" {
  description = "Name of the service"
  type        = string
  default     = "file-upload-service"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "jwt_secret" {
  description = "JWT secret key"
  type        = string
  sensitive   = true
}

# Data sources
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# S3 Bucket for file uploads
resource "aws_s3_bucket" "upload_bucket" {
  bucket = "${var.service_name}-uploads-${var.environment}-${data.aws_caller_identity.current.account_id}"

  tags = {
    Environment = var.environment
    Service     = var.service_name
  }
}

# S3 Bucket versioning
resource "aws_s3_bucket_versioning" "upload_bucket_versioning" {
  bucket = aws_s3_bucket.upload_bucket.id
  versioning_configuration {
    status = "Enabled"
  }
}

# S3 Bucket server-side encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "upload_bucket_encryption" {
  bucket = aws_s3_bucket.upload_bucket.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# S3 Bucket CORS configuration
resource "aws_s3_bucket_cors_configuration" "upload_bucket_cors" {
  bucket = aws_s3_bucket.upload_bucket.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "POST", "PUT", "DELETE"]
    allowed_origins = ["*"]  # Configure appropriately for production
    max_age_seconds = 3000
  }
}

# S3 Bucket lifecycle configuration
resource "aws_s3_bucket_lifecycle_configuration" "upload_bucket_lifecycle" {
  bucket = aws_s3_bucket.upload_bucket.id

  rule {
    id     = "cleanup_incomplete_uploads"
    status = "Enabled"

    abort_incomplete_multipart_upload {
      days_after_initiation = 1
    }
  }

  rule {
    id     = "cleanup_temporary_files"
    status = "Enabled"

    filter {
      prefix = "temporary/"
    }

    expiration {
      days = 7
    }
  }
}

# DynamoDB table for upload requests
resource "aws_dynamodb_table" "upload_requests" {
  name           = "${var.service_name}-upload-requests-${var.environment}"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "user_id"
    type = "S"
  }

  global_secondary_index {
    name     = "user-id-index"
    hash_key = "user_id"
  }

  tags = {
    Environment = var.environment
    Service     = var.service_name
  }
}

# IAM role for Lambda execution
resource "aws_iam_role" "lambda_execution_role" {
  name = "${var.service_name}-lambda-role-${var.environment}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Environment = var.environment
    Service     = var.service_name
  }
}

# IAM policy for Lambda execution
resource "aws_iam_role_policy" "lambda_execution_policy" {
  name = "${var.service_name}-lambda-policy-${var.environment}"
  role = aws_iam_role.lambda_execution_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
      },
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:GetObjectVersion",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.upload_bucket.arn,
          "${aws_s3_bucket.upload_bucket.arn}/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Query",
          "dynamodb:Scan"
        ]
        Resource = [
          aws_dynamodb_table.upload_requests.arn,
          "${aws_dynamodb_table.upload_requests.arn}/index/*"
        ]
      }
    ]
  })
}

# Lambda function
resource "aws_lambda_function" "file_upload_service" {
  function_name = "${var.service_name}-${var.environment}"
  role         = aws_iam_role.lambda_execution_role.arn
  
  # Using container image deployment
  package_type = "Image"
  image_uri    = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${data.aws_region.current.name}.amazonaws.com/${var.service_name}:latest"
  
  timeout     = 30
  memory_size = 512

  environment {
    variables = {
      AWS_REGION            = data.aws_region.current.name
      S3_BUCKET_NAME       = aws_s3_bucket.upload_bucket.id
      DYNAMODB_TABLE_NAME  = aws_dynamodb_table.upload_requests.name
      JWT_SECRET           = var.jwt_secret
      LOG_LEVEL            = "INFO"
      ENABLE_CACHING       = "true"
    }
  }

  tags = {
    Environment = var.environment
    Service     = var.service_name
  }
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "lambda_logs" {
  name              = "/aws/lambda/${aws_lambda_function.file_upload_service.function_name}"
  retention_in_days = 14

  tags = {
    Environment = var.environment
    Service     = var.service_name
  }
}

# API Gateway (if needed for HTTP access)
resource "aws_api_gateway_rest_api" "file_upload_api" {
  name        = "${var.service_name}-api-${var.environment}"
  description = "API for file upload service"

  endpoint_configuration {
    types = ["REGIONAL"]
  }

  tags = {
    Environment = var.environment
    Service     = var.service_name
  }
}

# API Gateway Lambda integration
resource "aws_api_gateway_integration" "lambda_integration" {
  rest_api_id = aws_api_gateway_rest_api.file_upload_api.id
  resource_id = aws_api_gateway_rest_api.file_upload_api.root_resource_id
  http_method = "ANY"

  integration_http_method = "POST"
  type                   = "AWS_PROXY"
  uri                    = aws_lambda_function.file_upload_service.invoke_arn
}

# EventBridge rule for cleanup (runs daily at 2 AM)
resource "aws_cloudwatch_event_rule" "cleanup_schedule" {
  name                = "${var.service_name}-cleanup-${var.environment}"
  description         = "Trigger cleanup of expired uploads"
  schedule_expression = "cron(0 2 * * ? *)"

  tags = {
    Environment = var.environment
    Service     = var.service_name
  }
}

# EventBridge target for cleanup
resource "aws_cloudwatch_event_target" "cleanup_target" {
  rule      = aws_cloudwatch_event_rule.cleanup_schedule.name
  target_id = "CleanupLambdaTarget"
  arn       = aws_lambda_function.file_upload_service.arn

  input = jsonencode({
    "action" = "cleanup"
  })
}

# Lambda permission for EventBridge
resource "aws_lambda_permission" "allow_eventbridge" {
  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.file_upload_service.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.cleanup_schedule.arn
}

# Outputs
output "s3_bucket_name" {
  description = "Name of the S3 bucket for uploads"
  value       = aws_s3_bucket.upload_bucket.id
}

output "dynamodb_table_name" {
  description = "Name of the DynamoDB table"
  value       = aws_dynamodb_table.upload_requests.name
}

output "lambda_function_arn" {
  description = "ARN of the Lambda function"
  value       = aws_lambda_function.file_upload_service.arn
}

output "api_gateway_url" {
  description = "URL of the API Gateway"
  value       = "https://${aws_api_gateway_rest_api.file_upload_api.id}.execute-api.${data.aws_region.current.name}.amazonaws.com"
}
