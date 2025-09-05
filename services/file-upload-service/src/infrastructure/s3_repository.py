"""
AWS S3 Storage Repository Implementation

Concrete implementation of IFileStorageRepository using AWS S3
"""

import boto3
from botocore.exceptions import ClientError, NoCredentialsError
from typing import Dict, Any, Optional
import uuid
from datetime import datetime, timedelta

from ..domain.repositories import IFileStorageRepository
from ..domain.models import UploadResult, DeletionResult


class S3StorageRepository(IFileStorageRepository):
    """
    S3 implementation of file storage repository
    
    Handles all S3 operations including presigned URLs and file management
    """
    
    def __init__(
        self,
        bucket_name: str,
        aws_region: str = "us-east-1",
        s3_client: Optional[boto3.client] = None
    ):
        """
        Initialize S3 repository
        
        Args:
            bucket_name: S3 bucket name
            aws_region: AWS region
            s3_client: Optional S3 client (for testing)
        """
        self.bucket_name = bucket_name
        self.aws_region = aws_region
        self._s3_client = s3_client or boto3.client('s3', region_name=aws_region)
    
    async def generate_presigned_upload_url(
        self,
        s3_key: str,
        content_type: str,
        expires_in_seconds: int = 3600
    ) -> UploadResult:
        """
        Generate presigned URL for S3 upload
        
        Uses POST presigned URL for better security and control
        """
        try:
            # Generate presigned POST for direct upload
            post_data = self._s3_client.generate_presigned_post(
                Bucket=self.bucket_name,
                Key=s3_key,
                Fields={
                    'Content-Type': content_type,
                    'x-amz-meta-upload-id': str(uuid.uuid4())
                },
                Conditions=[
                    ['content-length-range', 1, 100 * 1024 * 1024],  # 1 byte to 100MB
                    {'Content-Type': content_type}
                ],
                ExpiresIn=expires_in_seconds
            )
            
            return UploadResult(
                upload_id=str(uuid.uuid4()),
                presigned_url=post_data['url'],
                s3_key=s3_key,
                expires_at=datetime.utcnow() + timedelta(seconds=expires_in_seconds),
                upload_fields=post_data['fields']
            )
            
        except ClientError as e:
            error_code = e.response['Error']['Code']
            raise Exception(f"S3 error generating presigned URL: {error_code}")
        except NoCredentialsError:
            raise Exception("AWS credentials not found")
    
    async def delete_file(self, s3_key: str) -> DeletionResult:
        """Delete file from S3"""
        try:
            self._s3_client.delete_object(
                Bucket=self.bucket_name,
                Key=s3_key
            )
            
            return DeletionResult(
                filename=s3_key.split('/')[-1],
                s3_key=s3_key,
                success=True,
                message="File deleted successfully"
            )
            
        except ClientError as e:
            error_code = e.response['Error']['Code']
            return DeletionResult(
                filename=s3_key.split('/')[-1],
                s3_key=s3_key,
                success=False,
                message=f"S3 error: {error_code}"
            )
        except Exception as e:
            return DeletionResult(
                filename=s3_key.split('/')[-1],
                s3_key=s3_key,
                success=False,
                message=f"Unexpected error: {str(e)}"
            )
    
    async def file_exists(self, s3_key: str) -> bool:
        """Check if file exists in S3"""
        try:
            self._s3_client.head_object(Bucket=self.bucket_name, Key=s3_key)
            return True
        except ClientError as e:
            if e.response['Error']['Code'] == '404':
                return False
            raise
    
    async def get_file_metadata(self, s3_key: str) -> Optional[Dict[str, Any]]:
        """Get file metadata from S3"""
        try:
            response = self._s3_client.head_object(Bucket=self.bucket_name, Key=s3_key)
            return {
                'size': response['ContentLength'],
                'content_type': response['ContentType'],
                'last_modified': response['LastModified'],
                'etag': response['ETag'].strip('"'),
                'metadata': response.get('Metadata', {})
            }
        except ClientError as e:
            if e.response['Error']['Code'] == '404':
                return None
            raise
    
    async def copy_file(self, source_key: str, destination_key: str) -> bool:
        """Copy file within S3"""
        try:
            copy_source = {'Bucket': self.bucket_name, 'Key': source_key}
            self._s3_client.copy_object(
                CopySource=copy_source,
                Bucket=self.bucket_name,
                Key=destination_key
            )
            return True
        except ClientError:
            return False
