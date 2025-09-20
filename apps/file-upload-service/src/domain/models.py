"""
Domain Models for File Upload Service

This module defines the core business entities following Domain-Driven Design principles.
Models are pure Python classes with business logic and validation rules.
"""

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from enum import Enum
from typing import Dict, Optional, Any
import uuid


class FileStatus(Enum):
    """Enumeration of possible file statuses"""
    PENDING = "pending"
    UPLOADED = "uploaded"
    PROCESSING = "processing"
    COMPLETED = "completed"
    FAILED = "failed"
    DELETED = "deleted"


class UploadPurpose(Enum):
    """Enumeration of upload purposes for business logic"""
    FORM_ATTACHMENT = "form_attachment"
    USER_AVATAR = "user_avatar"
    DOCUMENT = "document"
    IMAGE = "image"
    TEMPORARY = "temporary"


@dataclass(frozen=True)
class FileMetadata:
    """
    Value object representing file metadata
    
    Follows SRP - Single responsibility for file metadata representation
    Immutable to ensure data integrity
    """
    content_type: str
    size_bytes: int
    checksum: Optional[str] = None
    original_filename: Optional[str] = None
    
    def __post_init__(self) -> None:
        """Validate file metadata on creation"""
        if self.size_bytes < 0:
            raise ValueError("File size cannot be negative")
        
        if self.size_bytes > 100 * 1024 * 1024:  # 100MB limit
            raise ValueError("File size exceeds maximum allowed (100MB)")


@dataclass
class UploadRequest:
    """
    Entity representing an upload request
    
    Contains business logic for upload validation and URL generation
    """
    id: str = field(default_factory=lambda: str(uuid.uuid4()))
    filename: str = field(default="")
    purpose: UploadPurpose = UploadPurpose.TEMPORARY
    metadata: Optional[FileMetadata] = None
    user_id: Optional[str] = None
    form_id: Optional[str] = None
    expires_at: datetime = field(default_factory=lambda: datetime.utcnow() + timedelta(hours=1))
    created_at: datetime = field(default_factory=datetime.utcnow)
    status: FileStatus = FileStatus.PENDING
    s3_key: Optional[str] = None
    presigned_url: Optional[str] = None
    
    def __post_init__(self) -> None:
        """Initialize derived fields and validate business rules"""
        if not self.filename:
            raise ValueError("Filename is required")
        
        # Generate S3 key if not provided
        if not self.s3_key:
            self.s3_key = self._generate_s3_key()
    
    def _generate_s3_key(self) -> str:
        """
        Generate a unique S3 key for the file
        
        Format: {purpose}/{user_id}/{date}/{uuid}_{filename}
        """
        date_prefix = self.created_at.strftime("%Y/%m/%d")
        unique_filename = f"{uuid.uuid4()}_{self.filename}"
        
        if self.user_id:
            return f"{self.purpose.value}/{self.user_id}/{date_prefix}/{unique_filename}"
        else:
            return f"{self.purpose.value}/anonymous/{date_prefix}/{unique_filename}"
    
    def is_expired(self) -> bool:
        """Check if the upload request has expired"""
        return datetime.utcnow() > self.expires_at
    
    def mark_as_uploaded(self) -> None:
        """Mark the request as successfully uploaded"""
        if self.status != FileStatus.PENDING:
            raise ValueError(f"Cannot mark as uploaded from status: {self.status}")
        self.status = FileStatus.UPLOADED
    
    def mark_as_failed(self) -> None:
        """Mark the request as failed"""
        self.status = FileStatus.FAILED
    
    def mark_as_deleted(self) -> None:
        """Mark the file as deleted"""
        self.status = FileStatus.DELETED


@dataclass
class UploadResult:
    """
    Value object representing the result of an upload operation
    
    Encapsulates all information needed by the client
    """
    upload_id: str
    presigned_url: str
    s3_key: str
    expires_at: datetime
    upload_fields: Dict[str, Any]
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for API response"""
        return {
            "upload_id": self.upload_id,
            "presigned_url": self.presigned_url,
            "s3_key": self.s3_key,
            "expires_at": self.expires_at.isoformat(),
            "upload_fields": self.upload_fields
        }


@dataclass
class DeletionResult:
    """
    Value object representing the result of a deletion operation
    """
    filename: str
    s3_key: str
    success: bool
    message: str
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for API response"""
        return {
            "filename": self.filename,
            "s3_key": self.s3_key,
            "success": self.success,
            "message": self.message
        }


# Domain Exceptions
class FileUploadDomainError(Exception):
    """Base exception for domain-related errors"""
    pass


class InvalidFileError(FileUploadDomainError):
    """Raised when file validation fails"""
    pass


class FileNotFoundError(FileUploadDomainError):
    """Raised when requested file doesn't exist"""
    pass


class UploadExpiredError(FileUploadDomainError):
    """Raised when upload request has expired"""
    pass


class UnauthorizedAccessError(FileUploadDomainError):
    """Raised when user doesn't have permission to access file"""
    pass
