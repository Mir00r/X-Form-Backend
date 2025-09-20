"""
Unit Tests for Domain Models

Tests the core business logic without external dependencies
"""

import pytest
from datetime import datetime, timedelta
from unittest.mock import Mock

from src.domain.models import (
    UploadRequest, FileMetadata, UploadPurpose, FileStatus,
    UploadResult, DeletionResult, InvalidFileError
)


class TestFileMetadata:
    """Test cases for FileMetadata value object"""
    
    def test_valid_file_metadata_creation(self):
        """Test creating valid file metadata"""
        metadata = FileMetadata(
            content_type="image/jpeg",
            size_bytes=1024,
            checksum="abc123",
            original_filename="test.jpg"
        )
        
        assert metadata.content_type == "image/jpeg"
        assert metadata.size_bytes == 1024
        assert metadata.checksum == "abc123"
        assert metadata.original_filename == "test.jpg"
    
    def test_negative_file_size_raises_error(self):
        """Test that negative file size raises ValueError"""
        with pytest.raises(ValueError, match="File size cannot be negative"):
            FileMetadata(content_type="image/jpeg", size_bytes=-1)
    
    def test_oversized_file_raises_error(self):
        """Test that oversized file raises ValueError"""
        with pytest.raises(ValueError, match="File size exceeds maximum"):
            FileMetadata(content_type="image/jpeg", size_bytes=101 * 1024 * 1024)
    
    def test_metadata_immutability(self):
        """Test that FileMetadata is immutable"""
        metadata = FileMetadata(content_type="image/jpeg", size_bytes=1024)
        
        with pytest.raises(AttributeError):
            metadata.size_bytes = 2048


class TestUploadRequest:
    """Test cases for UploadRequest entity"""
    
    def test_valid_upload_request_creation(self):
        """Test creating valid upload request"""
        request = UploadRequest(
            filename="test.jpg",
            purpose=UploadPurpose.FORM_ATTACHMENT
        )
        
        assert request.filename == "test.jpg"
        assert request.purpose == UploadPurpose.FORM_ATTACHMENT
        assert request.status == FileStatus.PENDING
        assert request.s3_key is not None
        assert request.id is not None
    
    def test_empty_filename_raises_error(self):
        """Test that empty filename raises ValueError"""
        with pytest.raises(ValueError, match="Filename is required"):
            UploadRequest(filename="")
    
    def test_s3_key_generation_with_user_id(self):
        """Test S3 key generation with user ID"""
        request = UploadRequest(
            filename="test.jpg",
            user_id="user123",
            purpose=UploadPurpose.USER_AVATAR
        )
        
        assert "user_avatar/user123/" in request.s3_key
        assert "test.jpg" in request.s3_key
    
    def test_s3_key_generation_without_user_id(self):
        """Test S3 key generation without user ID"""
        request = UploadRequest(
            filename="test.jpg",
            purpose=UploadPurpose.TEMPORARY
        )
        
        assert "temporary/anonymous/" in request.s3_key
        assert "test.jpg" in request.s3_key
    
    def test_expiration_check(self):
        """Test upload request expiration logic"""
        # Create expired request
        expired_request = UploadRequest(
            filename="test.jpg",
            expires_at=datetime.utcnow() - timedelta(hours=1)
        )
        assert expired_request.is_expired()
        
        # Create non-expired request
        valid_request = UploadRequest(
            filename="test.jpg",
            expires_at=datetime.utcnow() + timedelta(hours=1)
        )
        assert not valid_request.is_expired()
    
    def test_status_transitions(self):
        """Test valid status transitions"""
        request = UploadRequest(filename="test.jpg")
        
        # Test marking as uploaded
        request.mark_as_uploaded()
        assert request.status == FileStatus.UPLOADED
        
        # Test that cannot mark as uploaded again
        with pytest.raises(ValueError, match="Cannot mark as uploaded"):
            request.mark_as_uploaded()
        
        # Create new request for other transitions
        request2 = UploadRequest(filename="test2.jpg")
        request2.mark_as_failed()
        assert request2.status == FileStatus.FAILED
        
        request3 = UploadRequest(filename="test3.jpg")
        request3.mark_as_deleted()
        assert request3.status == FileStatus.DELETED


class TestUploadResult:
    """Test cases for UploadResult value object"""
    
    def test_upload_result_creation(self):
        """Test creating upload result"""
        result = UploadResult(
            upload_id="123",
            presigned_url="https://example.com/upload",
            s3_key="test/file.jpg",
            expires_at=datetime.utcnow(),
            upload_fields={"key": "value"}
        )
        
        assert result.upload_id == "123"
        assert result.presigned_url == "https://example.com/upload"
        assert result.s3_key == "test/file.jpg"
        assert result.upload_fields == {"key": "value"}
    
    def test_to_dict_conversion(self):
        """Test converting UploadResult to dictionary"""
        expires_at = datetime.utcnow()
        result = UploadResult(
            upload_id="123",
            presigned_url="https://example.com/upload",
            s3_key="test/file.jpg",
            expires_at=expires_at,
            upload_fields={"key": "value"}
        )
        
        result_dict = result.to_dict()
        
        assert result_dict["upload_id"] == "123"
        assert result_dict["presigned_url"] == "https://example.com/upload"
        assert result_dict["s3_key"] == "test/file.jpg"
        assert result_dict["expires_at"] == expires_at.isoformat()
        assert result_dict["upload_fields"] == {"key": "value"}


class TestDeletionResult:
    """Test cases for DeletionResult value object"""
    
    def test_deletion_result_creation(self):
        """Test creating deletion result"""
        result = DeletionResult(
            filename="test.jpg",
            s3_key="uploads/test.jpg",
            success=True,
            message="File deleted successfully"
        )
        
        assert result.filename == "test.jpg"
        assert result.s3_key == "uploads/test.jpg"
        assert result.success is True
        assert result.message == "File deleted successfully"
    
    def test_to_dict_conversion(self):
        """Test converting DeletionResult to dictionary"""
        result = DeletionResult(
            filename="test.jpg",
            s3_key="uploads/test.jpg",
            success=False,
            message="File not found"
        )
        
        result_dict = result.to_dict()
        
        assert result_dict["filename"] == "test.jpg"
        assert result_dict["s3_key"] == "uploads/test.jpg"
        assert result_dict["success"] is False
        assert result_dict["message"] == "File not found"


# Test fixtures for common objects
@pytest.fixture
def valid_file_metadata():
    """Fixture for valid file metadata"""
    return FileMetadata(
        content_type="image/jpeg",
        size_bytes=1024,
        original_filename="test.jpg"
    )


@pytest.fixture
def valid_upload_request():
    """Fixture for valid upload request"""
    return UploadRequest(
        filename="test.jpg",
        purpose=UploadPurpose.FORM_ATTACHMENT,
        user_id="user123"
    )
