# AWS Lambda File Upload Service

A production-ready microservice for handling file uploads to AWS S3 using clean architecture principles.

## 🏗️ Architecture

This service follows **Clean Architecture** and **SOLID principles**:

### Layer Structure
```
├── domain/          # Business logic and entities (innermost layer)
│   ├── models.py    # Domain entities and value objects
│   └── repositories.py  # Repository interfaces (DIP)
├── application/     # Use cases and business workflows
│   └── use_cases.py # Application services (SRP)
├── infrastructure/ # External dependencies (outermost layer)
│   ├── s3_repository.py     # S3 storage implementation
│   ├── dynamodb_repository.py # DynamoDB implementation
│   └── auth_service.py      # JWT authentication
├── presentation/   # HTTP/API layer
│   └── controllers.py # FastAPI controllers
├── configuration.py # Dependency injection (DIP)
└── main.py         # Lambda entry point
```

### SOLID Principles Applied

1. **Single Responsibility Principle (SRP)**
   - Each use case handles one specific workflow
   - Controllers only handle HTTP concerns
   - Repositories only handle data access

2. **Open/Closed Principle (OCP)**
   - Use cases depend on interfaces, not implementations
   - Easy to extend with new storage backends

3. **Liskov Substitution Principle (LSP)**
   - All repository implementations are interchangeable
   - Mock implementations for testing

4. **Interface Segregation Principle (ISP)**
   - Small, focused interfaces (IFileStorageRepository, IAuthenticationService)
   - Clients depend only on methods they use

5. **Dependency Inversion Principle (DIP)**
   - High-level modules depend on abstractions
   - Concrete implementations injected at runtime

## 🚀 Features

- **Presigned S3 URLs** - Direct upload to S3 without proxy
- **JWT Authentication** - Secure token-based authentication
- **File Validation** - Type and size validation
- **Clean URLs** - Organized S3 key structure
- **Comprehensive Logging** - Structured logging with correlation IDs
- **Error Handling** - Consistent error responses
- **Health Checks** - Built-in health monitoring
- **Background Cleanup** - Automated cleanup of expired uploads

## 📡 API Endpoints

### Core Operations
- `POST /upload` - Generate presigned upload URL
- `DELETE /upload/{filename}` - Delete uploaded file
- `GET /upload/{upload_id}/status` - Check upload status
- `GET /health` - Health check

### Request/Response Examples

#### Generate Upload URL
```http
POST /upload
Content-Type: application/json
Authorization: Bearer <jwt-token>

{
  "filename": "document.pdf",
  "content_type": "application/pdf",
  "purpose": "form_attachment",
  "form_id": "form-123",
  "expires_in_seconds": 3600
}
```

Response:
```json
{
  "upload_id": "uuid-here",
  "presigned_url": "https://bucket.s3.amazonaws.com/...",
  "s3_key": "form_attachment/user123/2024/01/15/uuid_document.pdf",
  "expires_at": "2024-01-15T15:30:00Z",
  "upload_fields": {
    "key": "form_attachment/user123/2024/01/15/uuid_document.pdf",
    "Content-Type": "application/pdf",
    "x-amz-meta-upload-id": "uuid-here"
  }
}
```

#### Delete File
```http
DELETE /upload/document.pdf
Authorization: Bearer <jwt-token>
```

Response:
```json
{
  "filename": "document.pdf",
  "success": true,
  "message": "File deleted successfully"
}
```

## 🛠️ Installation & Deployment

### Prerequisites
- Python 3.9+
- AWS CLI configured
- Terraform (for infrastructure)

### Local Development

1. **Install Dependencies**
```bash
pip install -r requirements.txt
```

2. **Set Environment Variables**
```bash
export AWS_REGION=us-east-1
export S3_BUCKET_NAME=your-upload-bucket
export DYNAMODB_TABLE_NAME=upload-requests
export JWT_SECRET=your-jwt-secret
export USE_MOCK_AUTH=true  # For development
```

3. **Run Locally**
```bash
python src/main.py
```

4. **Test API**
```bash
curl -X POST http://localhost:8000/upload \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "test.jpg",
    "content_type": "image/jpeg",
    "purpose": "temporary"
  }'
```

### AWS Lambda Deployment

1. **Package Dependencies**
```bash
pip install -r requirements.txt -t ./package
cp -r src/* ./package/
cd package && zip -r ../lambda-deployment.zip .
```

2. **Deploy with AWS CLI**
```bash
aws lambda create-function \
  --function-name file-upload-service \
  --runtime python3.9 \
  --role arn:aws:iam::account:role/lambda-execution-role \
  --handler main.lambda_handler \
  --zip-file fileb://lambda-deployment.zip \
  --timeout 30 \
  --memory-size 512
```

3. **Set Environment Variables**
```bash
aws lambda update-function-configuration \
  --function-name file-upload-service \
  --environment Variables='{
    "S3_BUCKET_NAME":"your-bucket",
    "DYNAMODB_TABLE_NAME":"upload-requests",
    "JWT_SECRET":"your-secret"
  }'
```

### Infrastructure as Code

Use the provided Terraform configuration:

```bash
cd infrastructure/
terraform init
terraform plan
terraform apply
```

## 🧪 Testing

### Unit Tests
```bash
pytest tests/test_domain_models.py -v
pytest tests/test_use_cases.py -v
```

### Integration Tests
```bash
pytest tests/test_integration.py -v
```

### Test Coverage
```bash
coverage run -m pytest
coverage report
coverage html
```

## 🔒 Security Features

- **JWT Token Validation** - Secure authentication
- **File Type Validation** - Prevent malicious uploads
- **Size Limits** - Prevent abuse (100MB max)
- **Presigned URLs** - Time-limited access
- **User Authorization** - Users can only access their files
- **CORS Configuration** - Proper cross-origin handling

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `AWS_REGION` | AWS region | `us-east-1` |
| `S3_BUCKET_NAME` | S3 bucket for uploads | Required |
| `DYNAMODB_TABLE_NAME` | DynamoDB table | `upload-requests` |
| `JWT_SECRET` | JWT signing key | Required |
| `JWT_ALGORITHM` | JWT algorithm | `HS256` |
| `USE_MOCK_AUTH` | Use mock auth for dev | `false` |
| `ENABLE_CACHING` | Enable caching | `true` |
| `LOG_LEVEL` | Logging level | `INFO` |

### DynamoDB Table Schema

```json
{
  "TableName": "upload-requests",
  "AttributeDefinitions": [
    {"AttributeName": "id", "AttributeType": "S"},
    {"AttributeName": "user_id", "AttributeType": "S"}
  ],
  "KeySchema": [
    {"AttributeName": "id", "KeyType": "HASH"}
  ],
  "GlobalSecondaryIndexes": [
    {
      "IndexName": "user-id-index",
      "KeySchema": [
        {"AttributeName": "user_id", "KeyType": "HASH"}
      ]
    }
  ]
}
```

## 📊 Monitoring & Logging

### CloudWatch Metrics
- Request count and latency
- Error rates by endpoint
- S3 operation success rates
- Authentication failures

### Structured Logging
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "message": "Upload URL generated",
  "upload_id": "uuid-here",
  "user_id": "user123",
  "filename": "document.pdf",
  "s3_key": "uploads/user123/document.pdf"
}
```

## 🔄 Background Jobs

### Cleanup Expired Uploads
Automatically cleans up expired upload requests:

```bash
# Schedule with CloudWatch Events
aws events put-rule \
  --name cleanup-expired-uploads \
  --schedule-expression "cron(0 2 * * ? *)"  # Daily at 2 AM
```

## 🚦 Error Handling

### Error Response Format
```json
{
  "error": "validation_error",
  "message": "File type .exe is not allowed",
  "details": {
    "field": "filename",
    "allowed_types": [".jpg", ".png", ".pdf"]
  }
}
```

### HTTP Status Codes
- `200` - Success
- `400` - Bad Request (validation error)
- `401` - Unauthorized (invalid token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `500` - Internal Server Error

## 🔄 Future Enhancements

### TODO Items in Code
- Implement Redis caching for better performance
- Add AWS EventBridge for event publishing
- Implement virus scanning integration
- Add file compression for images
- Implement multi-part upload for large files
- Add file preview generation

### Planned Features
- File versioning
- Bulk upload operations
- File sharing with expiration
- Image resizing and optimization
- Integration with form builder service

## 🤝 Contributing

1. Follow Clean Architecture principles
2. Write comprehensive tests
3. Use type hints throughout
4. Follow PEP 8 style guide
5. Add logging for important operations
6. Update documentation

## 📄 License

This project is part of the X-Form Backend microservices architecture.
