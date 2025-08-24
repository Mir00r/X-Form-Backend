"""
Authentication utilities for Analytics Service
"""
import logging
from datetime import datetime, timedelta
from typing import Dict, Any, Optional
import jwt
from fastapi import HTTPException, status, Depends
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials

from app.config import settings

logger = logging.getLogger(__name__)

security = HTTPBearer()


class TokenValidator:
    """JWT Token validator for authentication."""
    
    def __init__(self):
        self.jwt_secret = settings.jwt_secret
        self.jwt_algorithm = settings.jwt_algorithm
        self.jwt_expiration_hours = settings.jwt_expiration_hours
    
    def verify_token(self, token: str) -> Dict[str, Any]:
        """Verify and decode JWT token."""
        try:
            payload = jwt.decode(
                token,
                self.jwt_secret,
                algorithms=[self.jwt_algorithm]
            )
            
            # Check if token is expired
            exp_timestamp = payload.get("exp")
            if exp_timestamp:
                exp_datetime = datetime.fromtimestamp(exp_timestamp)
                if exp_datetime < datetime.now():
                    raise HTTPException(
                        status_code=status.HTTP_401_UNAUTHORIZED,
                        detail="Token has expired"
                    )
            
            # Validate required fields
            if not payload.get("user_id"):
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="Invalid token: missing user_id"
                )
            
            return payload
            
        except jwt.ExpiredSignatureError:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Token has expired"
            )
        except jwt.InvalidTokenError as e:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail=f"Invalid token: {str(e)}"
            )
        except Exception as e:
            logger.error(f"Token verification error: {e}")
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Token verification failed"
            )
    
    def create_token(self, user_data: Dict[str, Any]) -> str:
        """Create a new JWT token."""
        try:
            payload = {
                "user_id": user_data["user_id"],
                "email": user_data.get("email"),
                "role": user_data.get("role", "user"),
                "exp": datetime.now() + timedelta(hours=self.jwt_expiration_hours),
                "iat": datetime.now(),
                "iss": "analytics-service"
            }
            
            token = jwt.encode(payload, self.jwt_secret, algorithm=self.jwt_algorithm)
            return token
            
        except Exception as e:
            logger.error(f"Token creation error: {e}")
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail="Failed to create token"
            )


# Global token validator instance
token_validator = TokenValidator()


def verify_token(token: str) -> Dict[str, Any]:
    """Verify JWT token and return payload."""
    return token_validator.verify_token(token)


async def get_current_user(credentials: HTTPAuthorizationCredentials = Depends(security)) -> Dict[str, Any]:
    """
    Dependency to get current authenticated user from JWT token.
    
    Usage in routes:
    @router.get("/protected")
    async def protected_route(current_user: dict = Depends(get_current_user)):
        user_id = current_user["user_id"]
        ...
    """
    try:
        token = credentials.credentials
        user_data = verify_token(token)
        
        logger.info(f"Authenticated user: {user_data.get('user_id')}")
        return user_data
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Authentication error: {e}")
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Authentication failed"
        )


async def get_optional_user(credentials: Optional[HTTPAuthorizationCredentials] = Depends(security)) -> Optional[Dict[str, Any]]:
    """
    Optional authentication dependency.
    Returns user data if token is provided and valid, None otherwise.
    """
    if not credentials:
        return None
    
    try:
        token = credentials.credentials
        return verify_token(token)
    except HTTPException:
        return None
    except Exception:
        return None


def require_role(required_role: str):
    """
    Decorator to require specific role for route access.
    
    Usage:
    @router.get("/admin")
    @require_role("admin")
    async def admin_route(current_user: dict = Depends(get_current_user)):
        ...
    """
    def decorator(func):
        async def wrapper(*args, current_user: dict = Depends(get_current_user), **kwargs):
            user_role = current_user.get("role", "user")
            if user_role != required_role and user_role != "admin":
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail=f"Access denied. Required role: {required_role}"
                )
            return await func(*args, current_user=current_user, **kwargs)
        return wrapper
    return decorator


def require_permission(permission: str):
    """
    Decorator to require specific permission for route access.
    
    Usage:
    @router.get("/analytics")
    @require_permission("analytics:read")
    async def analytics_route(current_user: dict = Depends(get_current_user)):
        ...
    """
    def decorator(func):
        async def wrapper(*args, current_user: dict = Depends(get_current_user), **kwargs):
            user_permissions = current_user.get("permissions", [])
            if permission not in user_permissions and "admin" not in current_user.get("role", ""):
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail=f"Access denied. Required permission: {permission}"
                )
            return await func(*args, current_user=current_user, **kwargs)
        return wrapper
    return decorator


class AuthenticationError(Exception):
    """Custom authentication error."""
    pass


class AuthorizationError(Exception):
    """Custom authorization error."""
    pass
