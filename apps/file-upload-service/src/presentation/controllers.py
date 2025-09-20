"""
FastAPI Controllers for File Upload Service

This module contains the HTTP presentation layer using FastAPI.
Controllers handle HTTP requests/responses and delegate to use cases.
"""

from fastapi import FastAPI, HTTPException, Depends, UploadFile, File, Header
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field
from typing import Optional, Dict, Any
from datetime import datetime
import structlog

from ..application.use_cases import (
    GenerateUploadUrlUseCase, DeleteFileUseCase, 
    GetUploadStatusUseCase, CleanupExpiredUploadsUseCase
)
from ..domain.models import UploadPurpose, FileUploadDomainError, InvalidFileError, FileNotFoundError, UnauthorizedAccessError

logger = structlog.get_logger()


# Request/Response Models (DTOs)
class UploadRequestDTO(BaseModel):
    """Data Transfer Object for upload request"""
    filename: str = Field(..., description="Original filename", min_length=1, max_length=255)
    content_type: str = Field(..., description="MIME type of the file")
    purpose: UploadPurpose = Field(default=UploadPurpose.TEMPORARY, description="Purpose of the upload")
    form_id: Optional[str] = Field(None, description="Associated form ID")
    expires_in_seconds: int = Field(default=3600, ge=300, le=86400, description="URL expiration time (5 min to 24 hours)")


class UploadResponseDTO(BaseModel):
    """Data Transfer Object for upload response"""
    upload_id: str
    presigned_url: str
    s3_key: str
    expires_at: datetime
    upload_fields: Dict[str, Any]


class UploadStatusResponseDTO(BaseModel):
    """Data Transfer Object for upload status"""
    upload_id: str
    filename: str
    status: str
    created_at: datetime
    expires_at: datetime
    s3_key: Optional[str]


class DeletionResponseDTO(BaseModel):
    """Data Transfer Object for deletion response"""
    filename: str
    success: bool
    message: str


class ErrorResponseDTO(BaseModel):
    """Standardized error response"""
    error: str
    message: str
    details: Optional[Dict[str, Any]] = None


class FileUploadController:
    """
    Controller for file upload operations
    
    Handles HTTP requests and delegates to use cases
    """
    
    def __init__(
        self,
        generate_upload_url_use_case: GenerateUploadUrlUseCase,
        delete_file_use_case: DeleteFileUseCase,
        get_upload_status_use_case: GetUploadStatusUseCase,
        cleanup_use_case: CleanupExpiredUploadsUseCase
    ):
        self.generate_upload_url_use_case = generate_upload_url_use_case
        self.delete_file_use_case = delete_file_use_case
        self.get_upload_status_use_case = get_upload_status_use_case
        self.cleanup_use_case = cleanup_use_case
    
    async def generate_upload_url(
        self,
        request: UploadRequestDTO,
        authorization: Optional[str] = Header(None)
    ) -> UploadResponseDTO:
        """
        Generate presigned URL for file upload
        
        Args:
            request: Upload request data
            authorization: Authorization header with JWT token
            
        Returns:
            UploadResponseDTO with presigned URL
            
        Raises:
            HTTPException: For various error conditions
        """
        try:
            logger.info("Generating upload URL", filename=request.filename, purpose=request.purpose.value)
            
            result = await self.generate_upload_url_use_case.execute(
                filename=request.filename,
                content_type=request.content_type,
                purpose=request.purpose,
                user_token=authorization,
                form_id=request.form_id,
                expires_in_seconds=request.expires_in_seconds
            )
            
            return UploadResponseDTO(
                upload_id=result.upload_id,
                presigned_url=result.presigned_url,
                s3_key=result.s3_key,
                expires_at=result.expires_at,
                upload_fields=result.upload_fields
            )
            
        except InvalidFileError as e:
            logger.warning("Invalid file upload request", error=str(e))
            raise HTTPException(status_code=400, detail=str(e))
        except UnauthorizedAccessError as e:
            logger.warning("Unauthorized upload request", error=str(e))
            raise HTTPException(status_code=401, detail=str(e))
        except Exception as e:
            logger.error("Unexpected error generating upload URL", error=str(e))
            raise HTTPException(status_code=500, detail="Internal server error")
    
    async def delete_file(
        self,
        filename: str,
        authorization: Optional[str] = Header(None)
    ) -> DeletionResponseDTO:
        """
        Delete an uploaded file
        
        Args:
            filename: Name of the file to delete
            authorization: Authorization header with JWT token
            
        Returns:
            DeletionResponseDTO with operation result
        """
        try:
            logger.info("Deleting file", filename=filename)
            
            result = await self.delete_file_use_case.execute(
                filename=filename,
                user_token=authorization
            )
            
            return DeletionResponseDTO(
                filename=result.filename,
                success=result.success,
                message=result.message
            )
            
        except FileNotFoundError as e:
            logger.warning("File not found for deletion", filename=filename, error=str(e))
            raise HTTPException(status_code=404, detail=str(e))
        except UnauthorizedAccessError as e:
            logger.warning("Unauthorized deletion request", filename=filename, error=str(e))
            raise HTTPException(status_code=403, detail=str(e))
        except Exception as e:
            logger.error("Unexpected error deleting file", filename=filename, error=str(e))
            raise HTTPException(status_code=500, detail="Internal server error")
    
    async def get_upload_status(
        self,
        upload_id: str,
        authorization: Optional[str] = Header(None)
    ) -> UploadStatusResponseDTO:
        """
        Get upload status by ID
        
        Args:
            upload_id: Upload request ID
            authorization: Authorization header with JWT token
            
        Returns:
            UploadStatusResponseDTO with current status
        """
        try:
            logger.info("Getting upload status", upload_id=upload_id)
            
            upload_request = await self.get_upload_status_use_case.execute(
                upload_id=upload_id,
                user_token=authorization
            )
            
            return UploadStatusResponseDTO(
                upload_id=upload_request.id,
                filename=upload_request.filename,
                status=upload_request.status.value,
                created_at=upload_request.created_at,
                expires_at=upload_request.expires_at,
                s3_key=upload_request.s3_key
            )
            
        except FileNotFoundError as e:
            logger.warning("Upload request not found", upload_id=upload_id, error=str(e))
            raise HTTPException(status_code=404, detail=str(e))
        except UnauthorizedAccessError as e:
            logger.warning("Unauthorized status request", upload_id=upload_id, error=str(e))
            raise HTTPException(status_code=403, detail=str(e))
        except Exception as e:
            logger.error("Unexpected error getting upload status", upload_id=upload_id, error=str(e))
            raise HTTPException(status_code=500, detail="Internal server error")
    
    async def cleanup_expired_uploads(self) -> Dict[str, int]:
        """
        Cleanup expired upload requests (admin endpoint)
        
        Returns:
            Dictionary with cleanup statistics
        """
        try:
            logger.info("Starting cleanup of expired uploads")
            
            stats = await self.cleanup_use_case.execute()
            
            logger.info("Cleanup completed", **stats)
            return stats
            
        except Exception as e:
            logger.error("Unexpected error during cleanup", error=str(e))
            raise HTTPException(status_code=500, detail="Internal server error")


def create_file_upload_app(controller: FileUploadController) -> FastAPI:
    """
    Create FastAPI application with file upload routes
    
    Args:
        controller: FileUploadController instance
        
    Returns:
        Configured FastAPI application
    """
    app = FastAPI(
        title="File Upload Service",
        description="Microservice for handling file uploads to AWS S3",
        version="1.0.0",
        docs_url="/docs",
        redoc_url="/redoc"
    )
    
    # CORS middleware
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],  # Configure appropriately for production
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
    
    # Health check endpoint
    @app.get("/health")
    async def health_check():
        """Health check endpoint"""
        return {"status": "healthy", "timestamp": datetime.utcnow().isoformat()}
    
    # Upload endpoints
    @app.post("/upload", response_model=UploadResponseDTO)
    async def generate_upload_url(
        request: UploadRequestDTO,
        authorization: Optional[str] = Header(None)
    ):
        """Generate presigned URL for direct upload to S3"""
        return await controller.generate_upload_url(request, authorization)
    
    @app.delete("/upload/{filename}", response_model=DeletionResponseDTO)
    async def delete_file(
        filename: str,
        authorization: Optional[str] = Header(None)
    ):
        """Delete an uploaded file"""
        return await controller.delete_file(filename, authorization)
    
    @app.get("/upload/{upload_id}/status", response_model=UploadStatusResponseDTO)
    async def get_upload_status(
        upload_id: str,
        authorization: Optional[str] = Header(None)
    ):
        """Get upload status by ID"""
        return await controller.get_upload_status(upload_id, authorization)
    
    # Admin endpoints
    @app.post("/admin/cleanup")
    async def cleanup_expired_uploads():
        """Cleanup expired upload requests (admin only)"""
        return await controller.cleanup_expired_uploads()
    
    # Exception handlers
    @app.exception_handler(FileUploadDomainError)
    async def domain_error_handler(request, exc: FileUploadDomainError):
        """Handle domain-specific errors"""
        return JSONResponse(
            status_code=400,
            content=ErrorResponseDTO(
                error="domain_error",
                message=str(exc)
            ).model_dump()
        )
    
    @app.exception_handler(HTTPException)
    async def http_exception_handler(request, exc: HTTPException):
        """Handle HTTP exceptions with consistent format"""
        return JSONResponse(
            status_code=exc.status_code,
            content=ErrorResponseDTO(
                error="http_error",
                message=exc.detail
            ).model_dump()
        )
    
    return app
