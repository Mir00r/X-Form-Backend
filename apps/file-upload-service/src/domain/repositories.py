"""
Repository Interfaces for File Upload Service

This module defines abstract interfaces for data access following the Repository pattern
and Dependency Inversion Principle. Business logic depends on these abstractions,
not concrete implementations.
"""

from abc import ABC, abstractmethod
from typing import List, Optional, Dict, Any, TypeVar, Generic
from datetime import datetime

from .models import UploadRequest, UploadResult, DeletionResult, FileStatus

# Generic type for repository operations
T = TypeVar('T')


class IRepository(ABC, Generic[T]):
    """
    Generic repository interface following DIP
    
    Provides common CRUD operations that can be implemented
    by different storage backends (DynamoDB, MongoDB, etc.)
    """
    
    @abstractmethod
    async def save(self, entity: T) -> T:
        """Save an entity to the repository"""
        pass
    
    @abstractmethod
    async def find_by_id(self, entity_id: str) -> Optional[T]:
        """Find an entity by its unique identifier"""
        pass
    
    @abstractmethod
    async def update(self, entity: T) -> T:
        """Update an existing entity"""
        pass
    
    @abstractmethod
    async def delete(self, entity_id: str) -> bool:
        """Delete an entity by its identifier"""
        pass


class IUploadRequestRepository(IRepository[UploadRequest]):
    """
    Repository interface for UploadRequest entities
    
    Extends the generic repository with upload-specific operations
    """
    
    @abstractmethod
    async def find_by_user_id(self, user_id: str, limit: int = 50) -> List[UploadRequest]:
        """Find upload requests for a specific user"""
        pass
    
    @abstractmethod
    async def find_by_status(self, status: FileStatus, limit: int = 100) -> List[UploadRequest]:
        """Find upload requests by status"""
        pass
    
    @abstractmethod
    async def find_expired_requests(self, before_date: datetime) -> List[UploadRequest]:
        """Find requests that have expired before the given date"""
        pass
    
    @abstractmethod
    async def find_by_s3_key(self, s3_key: str) -> Optional[UploadRequest]:
        """Find upload request by S3 key"""
        pass


class IFileStorageRepository(ABC):
    """
    Repository interface for file storage operations
    
    Abstracts S3 operations to allow for different storage backends
    """
    
    @abstractmethod
    async def generate_presigned_upload_url(
        self, 
        s3_key: str, 
        content_type: str,
        expires_in_seconds: int = 3600
    ) -> UploadResult:
        """
        Generate a presigned URL for direct upload to storage
        
        Args:
            s3_key: The storage key for the file
            content_type: MIME type of the file
            expires_in_seconds: URL expiration time
            
        Returns:
            UploadResult with presigned URL and upload fields
        """
        pass
    
    @abstractmethod
    async def delete_file(self, s3_key: str) -> DeletionResult:
        """
        Delete a file from storage
        
        Args:
            s3_key: The storage key of the file to delete
            
        Returns:
            DeletionResult with operation status
        """
        pass
    
    @abstractmethod
    async def file_exists(self, s3_key: str) -> bool:
        """Check if a file exists in storage"""
        pass
    
    @abstractmethod
    async def get_file_metadata(self, s3_key: str) -> Optional[Dict[str, Any]]:
        """Get metadata for a file"""
        pass
    
    @abstractmethod
    async def copy_file(self, source_key: str, destination_key: str) -> bool:
        """Copy a file within storage"""
        pass


class IEventPublisher(ABC):
    """
    Interface for publishing domain events
    
    Allows the application to notify other services of file operations
    """
    
    @abstractmethod
    async def publish_upload_started(self, upload_request: UploadRequest) -> None:
        """Publish event when upload URL is generated"""
        pass
    
    @abstractmethod
    async def publish_upload_completed(self, upload_request: UploadRequest) -> None:
        """Publish event when file upload is completed"""
        pass
    
    @abstractmethod
    async def publish_file_deleted(self, s3_key: str, user_id: Optional[str] = None) -> None:
        """Publish event when file is deleted"""
        pass
    
    @abstractmethod
    async def publish_upload_failed(self, upload_request: UploadRequest, error: str) -> None:
        """Publish event when upload fails"""
        pass


class ICacheRepository(ABC):
    """
    Interface for caching operations
    
    Allows caching of frequently accessed data like upload requests
    """
    
    @abstractmethod
    async def get(self, key: str) -> Optional[str]:
        """Get a value from cache"""
        pass
    
    @abstractmethod
    async def set(self, key: str, value: str, ttl_seconds: int = 3600) -> bool:
        """Set a value in cache with TTL"""
        pass
    
    @abstractmethod
    async def delete(self, key: str) -> bool:
        """Delete a value from cache"""
        pass
    
    @abstractmethod
    async def exists(self, key: str) -> bool:
        """Check if a key exists in cache"""
        pass


class IAuthenticationService(ABC):
    """
    Interface for authentication operations
    
    Abstracts JWT token validation and user authorization
    """
    
    @abstractmethod
    async def validate_token(self, token: str) -> Optional[Dict[str, Any]]:
        """
        Validate JWT token and return user claims
        
        Returns:
            User claims if token is valid, None otherwise
        """
        pass
    
    @abstractmethod
    async def get_user_id(self, token: str) -> Optional[str]:
        """Extract user ID from valid token"""
        pass
    
    @abstractmethod
    async def has_permission(self, user_id: str, resource: str, action: str) -> bool:
        """Check if user has permission to perform action on resource"""
        pass
