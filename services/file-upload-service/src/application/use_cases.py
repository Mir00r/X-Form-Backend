"""
Use Cases for File Upload Service

This module contains the application layer use cases following Clean Architecture.
Each use case has a single responsibility and orchestrates domain objects and repositories.
"""

from abc import ABC, abstractmethod
from typing import Optional, List, Dict, Any
from datetime import datetime, timedelta
import structlog

from ..domain.models import (
    UploadRequest, UploadResult, DeletionResult, FileMetadata,
    UploadPurpose, FileStatus, InvalidFileError, FileNotFoundError,
    UploadExpiredError, UnauthorizedAccessError
)
from ..domain.repositories import (
    IUploadRequestRepository, IFileStorageRepository, IEventPublisher,
    ICacheRepository, IAuthenticationService
)

logger = structlog.get_logger()


class IUseCase(ABC):
    """Base interface for all use cases"""
    pass


class GenerateUploadUrlUseCase(IUseCase):
    """
    Use case for generating presigned upload URLs
    
    Follows SRP - Single responsibility for URL generation workflow
    Uses DI for all dependencies to improve testability
    """
    
    def __init__(
        self,
        upload_repo: IUploadRequestRepository,
        storage_repo: IFileStorageRepository,
        event_publisher: IEventPublisher,
        cache_repo: ICacheRepository,
        auth_service: IAuthenticationService
    ):
        self._upload_repo = upload_repo
        self._storage_repo = storage_repo
        self._event_publisher = event_publisher
        self._cache_repo = cache_repo
        self._auth_service = auth_service
    
    async def execute(
        self,
        filename: str,
        content_type: str,
        purpose: UploadPurpose = UploadPurpose.TEMPORARY,
        user_token: Optional[str] = None,
        form_id: Optional[str] = None,
        expires_in_seconds: int = 3600
    ) -> UploadResult:
        """
        Generate a presigned URL for file upload
        
        Args:
            filename: Original filename
            content_type: MIME type of the file
            purpose: Purpose of the upload
            user_token: JWT token for authentication
            form_id: Associated form ID (if applicable)
            expires_in_seconds: URL expiration time
            
        Returns:
            UploadResult with presigned URL and metadata
            
        Raises:
            InvalidFileError: If file validation fails
            UnauthorizedAccessError: If authentication fails
        """
        logger.info("Generating upload URL", filename=filename, purpose=purpose.value)
        
        # Validate input
        self._validate_upload_request(filename, content_type)
        
        # Authenticate user if token provided
        user_id = None
        if user_token:
            user_id = await self._authenticate_user(user_token)
        
        # Create file metadata
        metadata = FileMetadata(
            content_type=content_type,
            size_bytes=0,  # Will be validated on actual upload
            original_filename=filename
        )
        
        # Create upload request entity
        upload_request = UploadRequest(
            filename=filename,
            purpose=purpose,
            metadata=metadata,
            user_id=user_id,
            form_id=form_id,
            expires_at=datetime.utcnow() + timedelta(seconds=expires_in_seconds)
        )
        
        # Generate presigned URL
        upload_result = await self._storage_repo.generate_presigned_upload_url(
            s3_key=upload_request.s3_key,
            content_type=content_type,
            expires_in_seconds=expires_in_seconds
        )
        
        # Update request with presigned URL
        upload_request.presigned_url = upload_result.presigned_url
        
        # Save upload request
        saved_request = await self._upload_repo.save(upload_request)
        
        # Cache for quick access
        await self._cache_upload_request(saved_request)
        
        # Publish event
        await self._event_publisher.publish_upload_started(saved_request)
        
        logger.info(
            "Upload URL generated successfully",
            upload_id=saved_request.id,
            s3_key=saved_request.s3_key
        )
        
        return upload_result
    
    def _validate_upload_request(self, filename: str, content_type: str) -> None:
        """Validate upload request parameters"""
        if not filename or not filename.strip():
            raise InvalidFileError("Filename cannot be empty")
        
        if not content_type:
            raise InvalidFileError("Content type is required")
        
        # Validate file extension
        allowed_extensions = {
            '.jpg', '.jpeg', '.png', '.gif', '.pdf', '.doc', '.docx',
            '.txt', '.csv', '.xlsx', '.zip', '.mp4', '.mov'
        }
        
        file_ext = filename.lower().split('.')[-1] if '.' in filename else ''
        if f'.{file_ext}' not in allowed_extensions:
            raise InvalidFileError(f"File type .{file_ext} is not allowed")
    
    async def _authenticate_user(self, token: str) -> str:
        """Authenticate user and return user ID"""
        user_id = await self._auth_service.get_user_id(token)
        if not user_id:
            raise UnauthorizedAccessError("Invalid or expired token")
        return user_id
    
    async def _cache_upload_request(self, request: UploadRequest) -> None:
        """Cache upload request for quick access"""
        try:
            cache_key = f"upload_request:{request.id}"
            # TODO: Serialize request object properly
            await self._cache_repo.set(cache_key, request.id, ttl_seconds=3600)
        except Exception as e:
            logger.warning("Failed to cache upload request", error=str(e))


class DeleteFileUseCase(IUseCase):
    """
    Use case for deleting uploaded files
    
    Handles business logic for file deletion including authorization
    """
    
    def __init__(
        self,
        upload_repo: IUploadRequestRepository,
        storage_repo: IFileStorageRepository,
        event_publisher: IEventPublisher,
        auth_service: IAuthenticationService
    ):
        self._upload_repo = upload_repo
        self._storage_repo = storage_repo
        self._event_publisher = event_publisher
        self._auth_service = auth_service
    
    async def execute(
        self,
        filename: str,
        user_token: Optional[str] = None
    ) -> DeletionResult:
        """
        Delete a file from storage
        
        Args:
            filename: Name of the file to delete
            user_token: JWT token for authentication
            
        Returns:
            DeletionResult with operation status
            
        Raises:
            FileNotFoundError: If file doesn't exist
            UnauthorizedAccessError: If user doesn't have permission
        """
        logger.info("Deleting file", filename=filename)
        
        # Find upload request by filename
        upload_request = await self._find_upload_request_by_filename(filename)
        if not upload_request:
            raise FileNotFoundError(f"File '{filename}' not found")
        
        # Authorize deletion
        if user_token:
            await self._authorize_deletion(upload_request, user_token)
        
        # Delete from storage
        deletion_result = await self._storage_repo.delete_file(upload_request.s3_key)
        
        if deletion_result.success:
            # Update request status
            upload_request.mark_as_deleted()
            await self._upload_repo.update(upload_request)
            
            # Publish event
            await self._event_publisher.publish_file_deleted(
                upload_request.s3_key,
                upload_request.user_id
            )
            
            logger.info("File deleted successfully", filename=filename, s3_key=upload_request.s3_key)
        else:
            logger.error("Failed to delete file", filename=filename, error=deletion_result.message)
        
        return deletion_result
    
    async def _find_upload_request_by_filename(self, filename: str) -> Optional[UploadRequest]:
        """Find upload request by original filename"""
        # TODO: Implement efficient filename search
        # For now, this is a simplified implementation
        # In production, you might need to index by filename or use a different approach
        return None
    
    async def _authorize_deletion(self, upload_request: UploadRequest, token: str) -> None:
        """Authorize user to delete the file"""
        user_id = await self._auth_service.get_user_id(token)
        if not user_id:
            raise UnauthorizedAccessError("Invalid or expired token")
        
        # Users can only delete their own files
        if upload_request.user_id != user_id:
            has_admin_permission = await self._auth_service.has_permission(
                user_id, "files", "delete_any"
            )
            if not has_admin_permission:
                raise UnauthorizedAccessError("You can only delete your own files")


class GetUploadStatusUseCase(IUseCase):
    """
    Use case for checking upload status
    
    Provides status information for upload requests
    """
    
    def __init__(
        self,
        upload_repo: IUploadRequestRepository,
        cache_repo: ICacheRepository,
        auth_service: IAuthenticationService
    ):
        self._upload_repo = upload_repo
        self._cache_repo = cache_repo
        self._auth_service = auth_service
    
    async def execute(
        self,
        upload_id: str,
        user_token: Optional[str] = None
    ) -> UploadRequest:
        """
        Get upload status by ID
        
        Args:
            upload_id: Upload request ID
            user_token: JWT token for authentication
            
        Returns:
            UploadRequest with current status
            
        Raises:
            FileNotFoundError: If upload request doesn't exist
            UnauthorizedAccessError: If user doesn't have permission
        """
        logger.info("Getting upload status", upload_id=upload_id)
        
        # Try cache first
        upload_request = await self._get_from_cache(upload_id)
        
        if not upload_request:
            # Fallback to repository
            upload_request = await self._upload_repo.find_by_id(upload_id)
        
        if not upload_request:
            raise FileNotFoundError(f"Upload request '{upload_id}' not found")
        
        # Authorize access
        if user_token:
            await self._authorize_access(upload_request, user_token)
        
        return upload_request
    
    async def _get_from_cache(self, upload_id: str) -> Optional[UploadRequest]:
        """Get upload request from cache"""
        try:
            cache_key = f"upload_request:{upload_id}"
            cached_data = await self._cache_repo.get(cache_key)
            if cached_data:
                # TODO: Deserialize properly
                return await self._upload_repo.find_by_id(upload_id)
        except Exception as e:
            logger.warning("Failed to get from cache", error=str(e))
        return None
    
    async def _authorize_access(self, upload_request: UploadRequest, token: str) -> None:
        """Authorize user to access upload status"""
        user_id = await self._auth_service.get_user_id(token)
        if not user_id:
            raise UnauthorizedAccessError("Invalid or expired token")
        
        # Users can only access their own uploads
        if upload_request.user_id != user_id:
            has_read_permission = await self._auth_service.has_permission(
                user_id, "files", "read_any"
            )
            if not has_read_permission:
                raise UnauthorizedAccessError("You can only access your own uploads")


class CleanupExpiredUploadsUseCase(IUseCase):
    """
    Use case for cleaning up expired upload requests
    
    Background task to maintain system hygiene
    """
    
    def __init__(
        self,
        upload_repo: IUploadRequestRepository,
        storage_repo: IFileStorageRepository,
        event_publisher: IEventPublisher
    ):
        self._upload_repo = upload_repo
        self._storage_repo = storage_repo
        self._event_publisher = event_publisher
    
    async def execute(self, before_date: Optional[datetime] = None) -> Dict[str, int]:
        """
        Clean up expired upload requests
        
        Args:
            before_date: Clean up requests expired before this date
            
        Returns:
            Dictionary with cleanup statistics
        """
        if not before_date:
            before_date = datetime.utcnow()
        
        logger.info("Starting cleanup of expired uploads", before_date=before_date)
        
        expired_requests = await self._upload_repo.find_expired_requests(before_date)
        
        stats = {
            "total_found": len(expired_requests),
            "deleted_from_storage": 0,
            "updated_in_db": 0,
            "errors": 0
        }
        
        for request in expired_requests:
            try:
                # Only delete if file was never uploaded or is still pending
                if request.status in [FileStatus.PENDING, FileStatus.FAILED]:
                    # Delete from storage if exists
                    if await self._storage_repo.file_exists(request.s3_key):
                        deletion_result = await self._storage_repo.delete_file(request.s3_key)
                        if deletion_result.success:
                            stats["deleted_from_storage"] += 1
                    
                    # Mark as deleted in database
                    request.mark_as_deleted()
                    await self._upload_repo.update(request)
                    stats["updated_in_db"] += 1
                    
            except Exception as e:
                logger.error("Error cleaning up expired upload", upload_id=request.id, error=str(e))
                stats["errors"] += 1
        
        logger.info("Cleanup completed", **stats)
        return stats
