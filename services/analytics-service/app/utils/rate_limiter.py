"""
Rate limiting utilities for Analytics Service
"""
import time
import logging
from typing import Dict, Any, Optional
from functools import wraps
from collections import defaultdict, deque
import asyncio
from fastapi import HTTPException, status, Request
import redis

from app.config import settings

logger = logging.getLogger(__name__)


class InMemoryRateLimiter:
    """In-memory rate limiter for development/testing."""
    
    def __init__(self):
        self.requests = defaultdict(deque)
        self.lock = asyncio.Lock()
    
    async def is_allowed(self, key: str, max_requests: int, window_seconds: int) -> bool:
        """Check if request is allowed based on rate limits."""
        async with self.lock:
            now = time.time()
            window_start = now - window_seconds
            
            # Remove old requests outside the window
            while self.requests[key] and self.requests[key][0] < window_start:
                self.requests[key].popleft()
            
            # Check if under limit
            if len(self.requests[key]) < max_requests:
                self.requests[key].append(now)
                return True
            
            return False
    
    async def get_remaining(self, key: str, max_requests: int, window_seconds: int) -> int:
        """Get remaining requests in the current window."""
        now = time.time()
        window_start = now - window_seconds
        
        # Count requests in current window
        current_requests = sum(1 for req_time in self.requests[key] if req_time >= window_start)
        return max(0, max_requests - current_requests)
    
    async def get_reset_time(self, key: str, window_seconds: int) -> float:
        """Get time when the rate limit window resets."""
        if not self.requests[key]:
            return time.time()
        
        return self.requests[key][0] + window_seconds


class RedisRateLimiter:
    """Redis-based rate limiter for production."""
    
    def __init__(self):
        try:
            self.redis_client = redis.Redis(
                host=settings.redis_host,
                port=settings.redis_port,
                password=settings.redis_password,
                db=settings.redis_db + 1,  # Use different DB for rate limiting
                decode_responses=True,
                socket_connect_timeout=5,
                socket_timeout=5
            )
            # Test connection
            self.redis_client.ping()
            self.available = True
            logger.info("Redis rate limiter initialized successfully")
        except Exception as e:
            logger.warning(f"Redis rate limiter unavailable, falling back to in-memory: {e}")
            self.available = False
            self.fallback = InMemoryRateLimiter()
    
    async def is_allowed(self, key: str, max_requests: int, window_seconds: int) -> bool:
        """Check if request is allowed based on rate limits."""
        if not self.available:
            return await self.fallback.is_allowed(key, max_requests, window_seconds)
        
        try:
            pipeline = self.redis_client.pipeline()
            now = time.time()
            window_start = now - window_seconds
            
            # Use sorted set to track requests with timestamps
            rate_key = f"rate_limit:{key}"
            
            # Remove old entries
            pipeline.zremrangebyscore(rate_key, 0, window_start)
            
            # Count current requests
            pipeline.zcard(rate_key)
            
            # Add current request timestamp
            pipeline.zadd(rate_key, {str(now): now})
            
            # Set expiration
            pipeline.expire(rate_key, window_seconds)
            
            results = pipeline.execute()
            current_count = results[1]  # Count after cleanup
            
            # Check if under limit (subtract 1 because we already added current request)
            return current_count < max_requests
            
        except Exception as e:
            logger.error(f"Redis rate limiter error: {e}")
            # Fallback to allowing the request
            return True
    
    async def get_remaining(self, key: str, max_requests: int, window_seconds: int) -> int:
        """Get remaining requests in the current window."""
        if not self.available:
            return await self.fallback.get_remaining(key, max_requests, window_seconds)
        
        try:
            now = time.time()
            window_start = now - window_seconds
            rate_key = f"rate_limit:{key}"
            
            # Clean up and count
            pipeline = self.redis_client.pipeline()
            pipeline.zremrangebyscore(rate_key, 0, window_start)
            pipeline.zcard(rate_key)
            results = pipeline.execute()
            
            current_count = results[1]
            return max(0, max_requests - current_count)
            
        except Exception as e:
            logger.error(f"Redis rate limiter error getting remaining: {e}")
            return max_requests  # Conservative fallback
    
    async def get_reset_time(self, key: str, window_seconds: int) -> float:
        """Get time when the rate limit window resets."""
        if not self.available:
            return await self.fallback.get_reset_time(key, window_seconds)
        
        try:
            rate_key = f"rate_limit:{key}"
            oldest_request = self.redis_client.zrange(rate_key, 0, 0, withscores=True)
            
            if oldest_request:
                return oldest_request[0][1] + window_seconds
            else:
                return time.time()
                
        except Exception as e:
            logger.error(f"Redis rate limiter error getting reset time: {e}")
            return time.time()


class RateLimiter:
    """Main rate limiter class that chooses implementation based on configuration."""
    
    def __init__(self):
        if settings.redis_host and settings.enable_rate_limiting:
            self.limiter = RedisRateLimiter()
        else:
            self.limiter = InMemoryRateLimiter()
            logger.info("Using in-memory rate limiter")
    
    async def check_rate_limit(
        self,
        key: str,
        max_requests: int,
        window_seconds: int
    ) -> Dict[str, Any]:
        """
        Check rate limit and return status information.
        
        Returns:
            Dict with keys: allowed, remaining, reset_time, retry_after
        """
        allowed = await self.limiter.is_allowed(key, max_requests, window_seconds)
        remaining = await self.limiter.get_remaining(key, max_requests, window_seconds)
        reset_time = await self.limiter.get_reset_time(key, window_seconds)
        retry_after = max(0, int(reset_time - time.time())) if not allowed else 0
        
        return {
            "allowed": allowed,
            "remaining": remaining,
            "reset_time": reset_time,
            "retry_after": retry_after
        }


# Global rate limiter instance
rate_limiter = RateLimiter()


def rate_limit(max_requests: int, window_seconds: int, key_func: Optional[callable] = None):
    """
    Decorator to apply rate limiting to FastAPI routes.
    
    Args:
        max_requests: Maximum number of requests allowed
        window_seconds: Time window in seconds
        key_func: Optional function to generate rate limit key from request
    
    Usage:
        @rate_limit(max_requests=100, window_seconds=3600)
        async def my_route():
            ...
    """
    def decorator(func):
        @wraps(func)
        async def wrapper(*args, **kwargs):
            # Find request object in arguments
            request = None
            for arg in args:
                if hasattr(arg, 'client') and hasattr(arg, 'headers'):
                    request = arg
                    break
            
            if not request:
                # If no request found, proceed without rate limiting
                logger.warning("Rate limit decorator: No request object found")
                return await func(*args, **kwargs)
            
            # Generate rate limit key
            if key_func:
                key = key_func(request)
            else:
                # Default key: IP address + route
                client_ip = request.client.host if request.client else "unknown"
                route = request.url.path
                key = f"{client_ip}:{route}"
            
            # Check rate limit
            limit_info = await rate_limiter.check_rate_limit(
                key, max_requests, window_seconds
            )
            
            if not limit_info["allowed"]:
                # Rate limit exceeded
                raise HTTPException(
                    status_code=status.HTTP_429_TOO_MANY_REQUESTS,
                    detail="Rate limit exceeded",
                    headers={
                        "X-RateLimit-Limit": str(max_requests),
                        "X-RateLimit-Remaining": str(limit_info["remaining"]),
                        "X-RateLimit-Reset": str(int(limit_info["reset_time"])),
                        "Retry-After": str(limit_info["retry_after"])
                    }
                )
            
            # Add rate limit headers to response
            response = await func(*args, **kwargs)
            
            if hasattr(response, 'headers'):
                response.headers["X-RateLimit-Limit"] = str(max_requests)
                response.headers["X-RateLimit-Remaining"] = str(limit_info["remaining"])
                response.headers["X-RateLimit-Reset"] = str(int(limit_info["reset_time"]))
            
            return response
            
        return wrapper
    return decorator


def user_rate_limit(max_requests: int, window_seconds: int):
    """
    Rate limit decorator that uses user ID as the key.
    Requires authentication to work properly.
    """
    def key_func(request):
        # Extract user ID from JWT token if available
        auth_header = request.headers.get("Authorization", "")
        if auth_header.startswith("Bearer "):
            try:
                token = auth_header[7:]
                from app.utils.auth import verify_token
                user_data = verify_token(token)
                return f"user:{user_data['user_id']}"
            except Exception:
                pass
        
        # Fallback to IP if no valid token
        client_ip = request.client.host if request.client else "unknown"
        return f"ip:{client_ip}"
    
    return rate_limit(max_requests, window_seconds, key_func)


def ip_rate_limit(max_requests: int, window_seconds: int):
    """Rate limit decorator that uses IP address as the key."""
    def key_func(request):
        client_ip = request.client.host if request.client else "unknown"
        return f"ip:{client_ip}"
    
    return rate_limit(max_requests, window_seconds, key_func)


async def get_rate_limit_status(key: str, max_requests: int, window_seconds: int) -> Dict[str, Any]:
    """Get current rate limit status for a key."""
    return await rate_limiter.check_rate_limit(key, max_requests, window_seconds)
