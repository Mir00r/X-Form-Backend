"""
Unit Tests for Use Cases

Tests the application layer use cases with mocked dependencies
"""

import pytest
from unittest.mock import Mock, AsyncMock
from datetime import datetime, timedelta

from src.application.use_cases import (
    GenerateUploadUrlUseCase, DeleteFileUseCase,
    GetUploadStatusUseCase, CleanupExpiredUploadsUseCase
)
from src.domain.models import (
    UploadRequest, UploadResult, DeletionResult, FileMetadata,
    UploadPurpose, FileStatus, InvalidFileError, UnauthorizedAccessError
)


class TestGenerateUploadUrlUseCase:
    """Test cases for GenerateUploadUrlUseCase"""
    
    @pytest.fixture
    def use_case_dependencies(self):
        """Create mocked dependencies for use case"""
        upload_repo = Mock()
        storage_repo = Mock()
        event_publisher = Mock()
        cache_repo = Mock()
        auth_service = Mock()
        
        return {
            'upload_repo': upload_repo,
            'storage_repo': storage_repo,
            'event_publisher': event_publisher,
            'cache_repo': cache_repo,
            'auth_service': auth_service
        }
    
    @pytest.fixture
    def use_case(self, use_case_dependencies):
        """Create use case with mocked dependencies"""
        return GenerateUploadUrlUseCase(**use_case_dependencies)
    
    @pytest.mark.asyncio
    async def test_generate_upload_url_success(self, use_case, use_case_dependencies):
        """Test successful upload URL generation"""
        # Arrange
        filename = "test.jpg"
        content_type = "image/jpeg"
        
        # Mock storage repository response
        upload_result = UploadResult(
            upload_id="123",
            presigned_url="https://example.com/upload",
            s3_key="uploads/test.jpg",
            expires_at=datetime.utcnow() + timedelta(hours=1),
            upload_fields={"key": "value"}
        )
        use_case_dependencies['storage_repo'].generate_presigned_upload_url = AsyncMock(return_value=upload_result)
        
        # Mock repository save
        use_case_dependencies['upload_repo'].save = AsyncMock(side_effect=lambda x: x)
        
        # Mock cache and event publisher
        use_case_dependencies['cache_repo'].set = AsyncMock(return_value=True)
        use_case_dependencies['event_publisher'].publish_upload_started = AsyncMock()
        
        # Act
        result = await use_case.execute(filename=filename, content_type=content_type)
        
        # Assert
        assert result.presigned_url == "https://example.com/upload"
        assert result.s3_key == "uploads/test.jpg"
        use_case_dependencies['upload_repo'].save.assert_called_once()
        use_case_dependencies['event_publisher'].publish_upload_started.assert_called_once()
    
    @pytest.mark.asyncio
    async def test_generate_upload_url_with_authentication(self, use_case, use_case_dependencies):
        """Test upload URL generation with user authentication"""
        # Arrange
        filename = "test.jpg"
        content_type = "image/jpeg"
        user_token = "Bearer valid-token"
        
        # Mock authentication
        use_case_dependencies['auth_service'].get_user_id = AsyncMock(return_value="user123")
        
        # Mock storage repository
        upload_result = UploadResult(
            upload_id="123",
            presigned_url="https://example.com/upload",
            s3_key="uploads/test.jpg",
            expires_at=datetime.utcnow() + timedelta(hours=1),
            upload_fields={"key": "value"}
        )
        use_case_dependencies['storage_repo'].generate_presigned_upload_url = AsyncMock(return_value=upload_result)
        use_case_dependencies['upload_repo'].save = AsyncMock(side_effect=lambda x: x)
        use_case_dependencies['cache_repo'].set = AsyncMock(return_value=True)
        use_case_dependencies['event_publisher'].publish_upload_started = AsyncMock()
        
        # Act
        result = await use_case.execute(
            filename=filename,
            content_type=content_type,
            user_token=user_token
        )
        
        # Assert
        use_case_dependencies['auth_service'].get_user_id.assert_called_once_with(user_token)
        assert result.upload_id == "123"
    
    @pytest.mark.asyncio
    async def test_invalid_filename_raises_error(self, use_case):
        """Test that invalid filename raises InvalidFileError"""
        with pytest.raises(InvalidFileError, match="Filename cannot be empty"):
            await use_case.execute(filename="", content_type="image/jpeg")
    
    @pytest.mark.asyncio
    async def test_invalid_file_extension_raises_error(self, use_case):
        """Test that invalid file extension raises InvalidFileError"""
        with pytest.raises(InvalidFileError, match="File type .exe is not allowed"):
            await use_case.execute(filename="malware.exe", content_type="application/octet-stream")
    
    @pytest.mark.asyncio
    async def test_unauthorized_token_raises_error(self, use_case, use_case_dependencies):
        """Test that invalid token raises UnauthorizedAccessError"""
        # Arrange
        use_case_dependencies['auth_service'].get_user_id = AsyncMock(return_value=None)
        
        # Act & Assert
        with pytest.raises(UnauthorizedAccessError, match="Invalid or expired token"):
            await use_case.execute(
                filename="test.jpg",
                content_type="image/jpeg",
                user_token="invalid-token"
            )


class TestDeleteFileUseCase:
    """Test cases for DeleteFileUseCase"""
    
    @pytest.fixture
    def use_case_dependencies(self):
        """Create mocked dependencies for delete use case"""
        upload_repo = Mock()
        storage_repo = Mock()
        event_publisher = Mock()
        auth_service = Mock()
        
        return {
            'upload_repo': upload_repo,
            'storage_repo': storage_repo,
            'event_publisher': event_publisher,
            'auth_service': auth_service
        }
    
    @pytest.fixture
    def use_case(self, use_case_dependencies):
        """Create delete use case with mocked dependencies"""
        return DeleteFileUseCase(**use_case_dependencies)
    
    @pytest.mark.asyncio
    async def test_delete_file_success(self, use_case, use_case_dependencies):
        """Test successful file deletion"""
        # Arrange
        filename = "test.jpg"
        
        # Mock upload request
        upload_request = UploadRequest(
            filename=filename,
            s3_key="uploads/test.jpg",
            user_id="user123"
        )
        
        # Mock repository responses
        use_case_dependencies['upload_repo'].update = AsyncMock(return_value=upload_request)
        
        deletion_result = DeletionResult(
            filename=filename,
            s3_key="uploads/test.jpg",
            success=True,
            message="File deleted successfully"
        )
        use_case_dependencies['storage_repo'].delete_file = AsyncMock(return_value=deletion_result)
        
        # Mock event publisher
        use_case_dependencies['event_publisher'].publish_file_deleted = AsyncMock()
        
        # Mock finding upload request (this is a simplified mock)
        use_case._find_upload_request_by_filename = AsyncMock(return_value=upload_request)
        
        # Act
        result = await use_case.execute(filename=filename)
        
        # Assert
        assert result.success is True
        assert result.filename == filename
        use_case_dependencies['storage_repo'].delete_file.assert_called_once_with("uploads/test.jpg")
        use_case_dependencies['event_publisher'].publish_file_deleted.assert_called_once()


class TestGetUploadStatusUseCase:
    """Test cases for GetUploadStatusUseCase"""
    
    @pytest.fixture
    def use_case_dependencies(self):
        """Create mocked dependencies"""
        upload_repo = Mock()
        cache_repo = Mock()
        auth_service = Mock()
        
        return {
            'upload_repo': upload_repo,
            'cache_repo': cache_repo,
            'auth_service': auth_service
        }
    
    @pytest.fixture
    def use_case(self, use_case_dependencies):
        """Create use case with mocked dependencies"""
        return GetUploadStatusUseCase(**use_case_dependencies)
    
    @pytest.mark.asyncio
    async def test_get_upload_status_from_repository(self, use_case, use_case_dependencies):
        """Test getting upload status from repository"""
        # Arrange
        upload_id = "123"
        upload_request = UploadRequest(
            id=upload_id,
            filename="test.jpg",
            status=FileStatus.UPLOADED
        )
        
        # Mock cache miss and repository hit
        use_case_dependencies['cache_repo'].get = AsyncMock(return_value=None)
        use_case_dependencies['upload_repo'].find_by_id = AsyncMock(return_value=upload_request)
        
        # Act
        result = await use_case.execute(upload_id=upload_id)
        
        # Assert
        assert result.id == upload_id
        assert result.filename == "test.jpg"
        assert result.status == FileStatus.UPLOADED


class TestCleanupExpiredUploadsUseCase:
    """Test cases for CleanupExpiredUploadsUseCase"""
    
    @pytest.fixture
    def use_case_dependencies(self):
        """Create mocked dependencies"""
        upload_repo = Mock()
        storage_repo = Mock()
        event_publisher = Mock()
        
        return {
            'upload_repo': upload_repo,
            'storage_repo': storage_repo,
            'event_publisher': event_publisher
        }
    
    @pytest.fixture
    def use_case(self, use_case_dependencies):
        """Create cleanup use case with mocked dependencies"""
        return CleanupExpiredUploadsUseCase(**use_case_dependencies)
    
    @pytest.mark.asyncio
    async def test_cleanup_expired_uploads(self, use_case, use_case_dependencies):
        """Test cleanup of expired uploads"""
        # Arrange
        expired_request = UploadRequest(
            filename="expired.jpg",
            s3_key="uploads/expired.jpg",
            status=FileStatus.PENDING,
            expires_at=datetime.utcnow() - timedelta(hours=1)
        )
        
        # Mock repository responses
        use_case_dependencies['upload_repo'].find_expired_requests = AsyncMock(return_value=[expired_request])
        use_case_dependencies['storage_repo'].file_exists = AsyncMock(return_value=True)
        
        deletion_result = DeletionResult(
            filename="expired.jpg",
            s3_key="uploads/expired.jpg",
            success=True,
            message="File deleted"
        )
        use_case_dependencies['storage_repo'].delete_file = AsyncMock(return_value=deletion_result)
        use_case_dependencies['upload_repo'].update = AsyncMock(return_value=expired_request)
        
        # Act
        stats = await use_case.execute()
        
        # Assert
        assert stats["total_found"] == 1
        assert stats["deleted_from_storage"] == 1
        assert stats["updated_in_db"] == 1
        assert stats["errors"] == 0
