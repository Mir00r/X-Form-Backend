"""
DynamoDB Repository Implementation

Concrete implementation of IUploadRequestRepository using AWS DynamoDB
"""

import boto3
from boto3.dynamodb.conditions import Key, Attr
from botocore.exceptions import ClientError
from typing import List, Optional
from datetime import datetime
import json

from ..domain.repositories import IUploadRequestRepository
from ..domain.models import UploadRequest, FileStatus, UploadPurpose, FileMetadata


class DynamoDBUploadRequestRepository(IUploadRequestRepository):
    """
    DynamoDB implementation of upload request repository
    
    Provides efficient storage and querying of upload requests
    """
    
    def __init__(
        self,
        table_name: str,
        aws_region: str = "us-east-1",
        dynamodb_resource: Optional[boto3.resource] = None
    ):
        """
        Initialize DynamoDB repository
        
        Args:
            table_name: DynamoDB table name
            aws_region: AWS region
            dynamodb_resource: Optional DynamoDB resource (for testing)
        """
        self.table_name = table_name
        self._dynamodb = dynamodb_resource or boto3.resource('dynamodb', region_name=aws_region)
        self._table = self._dynamodb.Table(table_name)
    
    async def save(self, entity: UploadRequest) -> UploadRequest:
        """Save upload request to DynamoDB"""
        try:
            item = self._entity_to_item(entity)
            self._table.put_item(Item=item)
            return entity
        except ClientError as e:
            raise Exception(f"DynamoDB error saving upload request: {e.response['Error']['Code']}")
    
    async def find_by_id(self, entity_id: str) -> Optional[UploadRequest]:
        """Find upload request by ID"""
        try:
            response = self._table.get_item(Key={'id': entity_id})
            if 'Item' in response:
                return self._item_to_entity(response['Item'])
            return None
        except ClientError as e:
            raise Exception(f"DynamoDB error finding upload request: {e.response['Error']['Code']}")
    
    async def update(self, entity: UploadRequest) -> UploadRequest:
        """Update existing upload request"""
        try:
            item = self._entity_to_item(entity)
            self._table.put_item(Item=item)
            return entity
        except ClientError as e:
            raise Exception(f"DynamoDB error updating upload request: {e.response['Error']['Code']}")
    
    async def delete(self, entity_id: str) -> bool:
        """Delete upload request by ID"""
        try:
            self._table.delete_item(Key={'id': entity_id})
            return True
        except ClientError:
            return False
    
    async def find_by_user_id(self, user_id: str, limit: int = 50) -> List[UploadRequest]:
        """Find upload requests for a specific user"""
        try:
            response = self._table.query(
                IndexName='user-id-index',  # Assumes GSI exists
                KeyConditionExpression=Key('user_id').eq(user_id),
                Limit=limit,
                ScanIndexForward=False  # Most recent first
            )
            return [self._item_to_entity(item) for item in response['Items']]
        except ClientError as e:
            raise Exception(f"DynamoDB error querying by user ID: {e.response['Error']['Code']}")
    
    async def find_by_status(self, status: FileStatus, limit: int = 100) -> List[UploadRequest]:
        """Find upload requests by status"""
        try:
            response = self._table.scan(
                FilterExpression=Attr('status').eq(status.value),
                Limit=limit
            )
            return [self._item_to_entity(item) for item in response['Items']]
        except ClientError as e:
            raise Exception(f"DynamoDB error querying by status: {e.response['Error']['Code']}")
    
    async def find_expired_requests(self, before_date: datetime) -> List[UploadRequest]:
        """Find requests that have expired before the given date"""
        try:
            response = self._table.scan(
                FilterExpression=Attr('expires_at').lt(before_date.isoformat()),
                ProjectionExpression='id, s3_key, #status, expires_at',
                ExpressionAttributeNames={'#status': 'status'}
            )
            return [self._item_to_entity(item) for item in response['Items']]
        except ClientError as e:
            raise Exception(f"DynamoDB error finding expired requests: {e.response['Error']['Code']}")
    
    async def find_by_s3_key(self, s3_key: str) -> Optional[UploadRequest]:
        """Find upload request by S3 key"""
        try:
            response = self._table.scan(
                FilterExpression=Attr('s3_key').eq(s3_key),
                Limit=1
            )
            if response['Items']:
                return self._item_to_entity(response['Items'][0])
            return None
        except ClientError as e:
            raise Exception(f"DynamoDB error finding by S3 key: {e.response['Error']['Code']}")
    
    def _entity_to_item(self, entity: UploadRequest) -> dict:
        """Convert UploadRequest entity to DynamoDB item"""
        item = {
            'id': entity.id,
            'filename': entity.filename,
            'purpose': entity.purpose.value,
            'user_id': entity.user_id,
            'form_id': entity.form_id,
            'expires_at': entity.expires_at.isoformat(),
            'created_at': entity.created_at.isoformat(),
            'status': entity.status.value,
            's3_key': entity.s3_key,
            'presigned_url': entity.presigned_url
        }
        
        if entity.metadata:
            item['metadata'] = {
                'content_type': entity.metadata.content_type,
                'size_bytes': entity.metadata.size_bytes,
                'checksum': entity.metadata.checksum,
                'original_filename': entity.metadata.original_filename
            }
        
        # Remove None values
        return {k: v for k, v in item.items() if v is not None}
    
    def _item_to_entity(self, item: dict) -> UploadRequest:
        """Convert DynamoDB item to UploadRequest entity"""
        metadata = None
        if 'metadata' in item:
            metadata = FileMetadata(
                content_type=item['metadata']['content_type'],
                size_bytes=item['metadata']['size_bytes'],
                checksum=item['metadata'].get('checksum'),
                original_filename=item['metadata'].get('original_filename')
            )
        
        return UploadRequest(
            id=item['id'],
            filename=item['filename'],
            purpose=UploadPurpose(item['purpose']),
            metadata=metadata,
            user_id=item.get('user_id'),
            form_id=item.get('form_id'),
            expires_at=datetime.fromisoformat(item['expires_at']),
            created_at=datetime.fromisoformat(item['created_at']),
            status=FileStatus(item['status']),
            s3_key=item.get('s3_key'),
            presigned_url=item.get('presigned_url')
        )
