"""
Dependency Injection Configuration

This module configures all dependencies following the Dependency Inversion Principle.
It provides factory functions for creating configured instances of all services.
"""

import os
from typing import Optional
import boto3

from .domain.repositories import (
    IUploadRequestRepository, IFileStorageRepository, IEventPublisher,
    ICacheRepository, IAuthenticationService
)
from .infrastructure.s3_repository import S3StorageRepository
from .infrastructure.dynamodb_repository import DynamoDBUploadRequestRepository
from .infrastructure.auth_service import JWTAuthenticationService, MockAuthenticationService
from .application.use_cases import (
    GenerateUploadUrlUseCase, DeleteFileUseCase,
    GetUploadStatusUseCase, CleanupExpiredUploadsUseCase
)
from .presentation.controllers import FileUploadController, create_file_upload_app


class ServiceConfiguration:
    """
    Configuration container for all service dependencies
    
    Centralizes dependency creation and configuration
    """
    
    def __init__(self):
        """Initialize configuration from environment variables"""
        # AWS Configuration
        self.aws_region = os.getenv("AWS_REGION", "us-east-1")
        self.s3_bucket = os.getenv("S3_BUCKET_NAME", "file-upload-bucket")
        self.dynamodb_table = os.getenv("DYNAMODB_TABLE_NAME", "upload-requests")
        
        # Authentication Configuration
        self.jwt_secret = os.getenv("JWT_SECRET", "development-secret-key")
        self.jwt_algorithm = os.getenv("JWT_ALGORITHM", "HS256")
        self.auth_service_url = os.getenv("AUTH_SERVICE_URL")
        
        # Feature Flags
        self.use_mock_auth = os.getenv("USE_MOCK_AUTH", "false").lower() == "true"
        self.enable_caching = os.getenv("ENABLE_CACHING", "true").lower() == "true"
        
        # Logging
        self.log_level = os.getenv("LOG_LEVEL", "INFO")
    
    def create_s3_repository(self) -> IFileStorageRepository:
        """Create S3 storage repository"""
        return S3StorageRepository(
            bucket_name=self.s3_bucket,
            aws_region=self.aws_region
        )
    
    def create_upload_repository(self) -> IUploadRequestRepository:
        """Create upload request repository"""
        return DynamoDBUploadRequestRepository(
            table_name=self.dynamodb_table,
            aws_region=self.aws_region
        )
    
    def create_auth_service(self) -> IAuthenticationService:
        """Create authentication service"""
        if self.use_mock_auth:
            return MockAuthenticationService()
        else:
            return JWTAuthenticationService(
                jwt_secret=self.jwt_secret,
                jwt_algorithm=self.jwt_algorithm,
                auth_service_url=self.auth_service_url
            )
    
    def create_event_publisher(self) -> IEventPublisher:
        """Create event publisher (stub implementation)"""
        return StubEventPublisher()
    
    def create_cache_repository(self) -> ICacheRepository:
        """Create cache repository (stub implementation)"""
        if self.enable_caching:
            return StubCacheRepository()
        else:
            return NullCacheRepository()


class StubEventPublisher(IEventPublisher):
    """
    Stub implementation of event publisher for now
    
    TODO: Implement with AWS EventBridge or SQS
    """
    
    async def publish_upload_started(self, upload_request) -> None:
        """Log upload started event"""
        print(f"Event: Upload started - {upload_request.id}")
    
    async def publish_upload_completed(self, upload_request) -> None:
        """Log upload completed event"""
        print(f"Event: Upload completed - {upload_request.id}")
    
    async def publish_file_deleted(self, s3_key: str, user_id: Optional[str] = None) -> None:
        """Log file deleted event"""
        print(f"Event: File deleted - {s3_key}")
    
    async def publish_upload_failed(self, upload_request, error: str) -> None:
        """Log upload failed event"""
        print(f"Event: Upload failed - {upload_request.id}: {error}")


class StubCacheRepository(ICacheRepository):
    """
    Stub implementation of cache repository
    
    TODO: Implement with Redis or DynamoDB
    """
    
    def __init__(self):
        self._cache = {}
    
    async def get(self, key: str) -> Optional[str]:
        """Get from in-memory cache"""
        return self._cache.get(key)
    
    async def set(self, key: str, value: str, ttl_seconds: int = 3600) -> bool:
        """Set in in-memory cache (ignoring TTL for simplicity)"""
        self._cache[key] = value
        return True
    
    async def delete(self, key: str) -> bool:
        """Delete from in-memory cache"""
        return self._cache.pop(key, None) is not None
    
    async def exists(self, key: str) -> bool:
        """Check if key exists in cache"""
        return key in self._cache


class NullCacheRepository(ICacheRepository):
    """Null object pattern for disabled caching"""
    
    async def get(self, key: str) -> Optional[str]:
        return None
    
    async def set(self, key: str, value: str, ttl_seconds: int = 3600) -> bool:
        return True
    
    async def delete(self, key: str) -> bool:
        return True
    
    async def exists(self, key: str) -> bool:
        return False


def create_configured_app():
    """
    Factory function to create fully configured FastAPI application
    
    This is the main composition root for dependency injection
    """
    # Create configuration
    config = ServiceConfiguration()
    
    # Create repositories
    storage_repo = config.create_s3_repository()
    upload_repo = config.create_upload_repository()
    auth_service = config.create_auth_service()
    event_publisher = config.create_event_publisher()
    cache_repo = config.create_cache_repository()
    
    # Create use cases
    generate_upload_url_use_case = GenerateUploadUrlUseCase(
        upload_repo=upload_repo,
        storage_repo=storage_repo,
        event_publisher=event_publisher,
        cache_repo=cache_repo,
        auth_service=auth_service
    )
    
    delete_file_use_case = DeleteFileUseCase(
        upload_repo=upload_repo,
        storage_repo=storage_repo,
        event_publisher=event_publisher,
        auth_service=auth_service
    )
    
    get_upload_status_use_case = GetUploadStatusUseCase(
        upload_repo=upload_repo,
        cache_repo=cache_repo,
        auth_service=auth_service
    )
    
    cleanup_use_case = CleanupExpiredUploadsUseCase(
        upload_repo=upload_repo,
        storage_repo=storage_repo,
        event_publisher=event_publisher
    )
    
    # Create controller
    controller = FileUploadController(
        generate_upload_url_use_case=generate_upload_url_use_case,
        delete_file_use_case=delete_file_use_case,
        get_upload_status_use_case=get_upload_status_use_case,
        cleanup_use_case=cleanup_use_case
    )
    
    # Create and configure FastAPI app
    app = create_file_upload_app(controller)
    
    return app
