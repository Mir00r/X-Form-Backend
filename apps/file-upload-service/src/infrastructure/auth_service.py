"""
JWT Authentication Service Implementation

Concrete implementation of IAuthenticationService for JWT validation
"""

from typing import Optional, Dict, Any
import jwt
from datetime import datetime, timezone
import httpx

from ..domain.repositories import IAuthenticationService


class JWTAuthenticationService(IAuthenticationService):
    """
    JWT-based authentication service
    
    Validates JWT tokens and provides user authorization
    """
    
    def __init__(
        self,
        jwt_secret: str,
        jwt_algorithm: str = "HS256",
        auth_service_url: Optional[str] = None
    ):
        """
        Initialize JWT authentication service
        
        Args:
            jwt_secret: Secret key for JWT validation
            jwt_algorithm: JWT algorithm (default: HS256)
            auth_service_url: Optional URL for external auth service
        """
        self.jwt_secret = jwt_secret
        self.jwt_algorithm = jwt_algorithm
        self.auth_service_url = auth_service_url
    
    async def validate_token(self, token: str) -> Optional[Dict[str, Any]]:
        """
        Validate JWT token and return user claims
        
        Args:
            token: JWT token string
            
        Returns:
            User claims if token is valid, None otherwise
        """
        try:
            # Remove Bearer prefix if present
            if token.startswith('Bearer '):
                token = token[7:]
            
            # Decode and validate token
            payload = jwt.decode(
                token,
                self.jwt_secret,
                algorithms=[self.jwt_algorithm]
            )
            
            # Check expiration
            if 'exp' in payload:
                exp_timestamp = payload['exp']
                if datetime.fromtimestamp(exp_timestamp, tz=timezone.utc) < datetime.now(timezone.utc):
                    return None
            
            return payload
            
        except jwt.InvalidTokenError:
            return None
        except Exception:
            return None
    
    async def get_user_id(self, token: str) -> Optional[str]:
        """
        Extract user ID from valid token
        
        Args:
            token: JWT token string
            
        Returns:
            User ID if token is valid, None otherwise
        """
        claims = await self.validate_token(token)
        if claims:
            # Try common user ID fields
            for field in ['user_id', 'sub', 'id', 'userId']:
                if field in claims:
                    return str(claims[field])
        return None
    
    async def has_permission(self, user_id: str, resource: str, action: str) -> bool:
        """
        Check if user has permission to perform action on resource
        
        Args:
            user_id: User identifier
            resource: Resource name (e.g., 'files')
            action: Action name (e.g., 'delete_any')
            
        Returns:
            True if user has permission, False otherwise
        """
        # If external auth service is configured, use it
        if self.auth_service_url:
            return await self._check_external_permission(user_id, resource, action)
        
        # Default implementation - basic role-based permissions
        return await self._check_default_permission(user_id, resource, action)
    
    async def _check_external_permission(self, user_id: str, resource: str, action: str) -> bool:
        """Check permission using external auth service"""
        try:
            async with httpx.AsyncClient() as client:
                response = await client.get(
                    f"{self.auth_service_url}/permissions/check",
                    params={
                        'user_id': user_id,
                        'resource': resource,
                        'action': action
                    },
                    timeout=5.0
                )
                
                if response.status_code == 200:
                    result = response.json()
                    return result.get('has_permission', False)
                
        except Exception:
            # Fail safely - deny permission if service is unavailable
            pass
        
        return False
    
    async def _check_default_permission(self, user_id: str, resource: str, action: str) -> bool:
        """Default permission checking logic"""
        # This is a simplified implementation
        # In production, you would check against a proper authorization system
        
        # For now, allow basic file operations for authenticated users
        if resource == "files":
            if action in ["read", "create", "delete_own"]:
                return True
            if action in ["delete_any", "read_any"]:
                # These would require admin role
                # TODO: Implement proper role checking
                return False
        
        return False


class MockAuthenticationService(IAuthenticationService):
    """
    Mock authentication service for testing and development
    
    Always returns successful authentication for development purposes
    """
    
    def __init__(self, mock_user_id: str = "test-user-123"):
        self.mock_user_id = mock_user_id
    
    async def validate_token(self, token: str) -> Optional[Dict[str, Any]]:
        """Always return valid claims for any token"""
        return {
            'user_id': self.mock_user_id,
            'sub': self.mock_user_id,
            'exp': datetime.now(timezone.utc).timestamp() + 3600,
            'roles': ['user']
        }
    
    async def get_user_id(self, token: str) -> Optional[str]:
        """Always return mock user ID"""
        return self.mock_user_id
    
    async def has_permission(self, user_id: str, resource: str, action: str) -> bool:
        """Always grant permission for development"""
        return True
