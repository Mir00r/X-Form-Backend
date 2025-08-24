"""
Redis Cache Service for Analytics
"""
import json
import logging
from typing import Any, Optional, Dict, List
import redis
from datetime import datetime, timedelta

from app.config import settings

logger = logging.getLogger(__name__)


class CacheService:
    """Service for managing Redis cache for analytics data."""
    
    def __init__(self):
        self.redis_client = redis.Redis(
            host=settings.redis_host,
            port=settings.redis_port,
            password=settings.redis_password,
            db=settings.redis_db,
            decode_responses=True,
            socket_connect_timeout=5,
            socket_timeout=5,
            retry_on_timeout=True
        )
        self.default_ttl = settings.cache_ttl
    
    def _get_cache_key(self, key_type: str, **kwargs) -> str:
        """Generate a cache key based on type and parameters."""
        key_parts = [settings.cache_prefix, key_type]
        
        for key, value in sorted(kwargs.items()):
            if value is not None:
                if isinstance(value, datetime):
                    value = value.isoformat()
                key_parts.append(f"{key}:{value}")
        
        return ":".join(key_parts)
    
    async def get(self, key: str) -> Optional[Any]:
        """Get value from cache."""
        try:
            value = self.redis_client.get(key)
            if value:
                return json.loads(value)
            return None
        except Exception as e:
            logger.warning(f"Cache get error for key {key}: {e}")
            return None
    
    async def set(self, key: str, value: Any, ttl: Optional[int] = None) -> bool:
        """Set value in cache with TTL."""
        try:
            ttl = ttl or self.default_ttl
            serialized_value = json.dumps(value, default=str)
            return self.redis_client.setex(key, ttl, serialized_value)
        except Exception as e:
            logger.warning(f"Cache set error for key {key}: {e}")
            return False
    
    async def delete(self, key: str) -> bool:
        """Delete key from cache."""
        try:
            return bool(self.redis_client.delete(key))
        except Exception as e:
            logger.warning(f"Cache delete error for key {key}: {e}")
            return False
    
    async def delete_pattern(self, pattern: str) -> int:
        """Delete all keys matching pattern."""
        try:
            keys = self.redis_client.keys(pattern)
            if keys:
                return self.redis_client.delete(*keys)
            return 0
        except Exception as e:
            logger.warning(f"Cache delete pattern error for {pattern}: {e}")
            return 0
    
    # Form summary cache methods
    async def get_form_summary(self, form_id: str, start_date: Optional[datetime] = None,
                             end_date: Optional[datetime] = None) -> Optional[Dict[str, Any]]:
        """Get cached form summary."""
        key = self._get_cache_key("form_summary", form_id=form_id, 
                                 start_date=start_date, end_date=end_date)
        return await self.get(key)
    
    async def set_form_summary(self, form_id: str, data: Dict[str, Any],
                             start_date: Optional[datetime] = None,
                             end_date: Optional[datetime] = None,
                             ttl: Optional[int] = None) -> bool:
        """Cache form summary."""
        key = self._get_cache_key("form_summary", form_id=form_id,
                                 start_date=start_date, end_date=end_date)
        return await self.set(key, data, ttl)
    
    # Question analytics cache methods
    async def get_question_analytics(self, form_id: str, question_id: str,
                                   start_date: Optional[datetime] = None,
                                   end_date: Optional[datetime] = None) -> Optional[Dict[str, Any]]:
        """Get cached question analytics."""
        key = self._get_cache_key("question_analytics", form_id=form_id,
                                 question_id=question_id, start_date=start_date,
                                 end_date=end_date)
        return await self.get(key)
    
    async def set_question_analytics(self, form_id: str, question_id: str,
                                   data: Dict[str, Any],
                                   start_date: Optional[datetime] = None,
                                   end_date: Optional[datetime] = None,
                                   ttl: Optional[int] = None) -> bool:
        """Cache question analytics."""
        key = self._get_cache_key("question_analytics", form_id=form_id,
                                 question_id=question_id, start_date=start_date,
                                 end_date=end_date)
        return await self.set(key, data, ttl)
    
    # Trend analysis cache methods
    async def get_trend_analysis(self, form_id: str, period: str,
                               start_date: Optional[datetime] = None,
                               end_date: Optional[datetime] = None) -> Optional[Dict[str, Any]]:
        """Get cached trend analysis."""
        key = self._get_cache_key("trend_analysis", form_id=form_id,
                                 period=period, start_date=start_date,
                                 end_date=end_date)
        return await self.get(key)
    
    async def set_trend_analysis(self, form_id: str, period: str,
                               data: Dict[str, Any],
                               start_date: Optional[datetime] = None,
                               end_date: Optional[datetime] = None,
                               ttl: Optional[int] = None) -> bool:
        """Cache trend analysis."""
        key = self._get_cache_key("trend_analysis", form_id=form_id,
                                 period=period, start_date=start_date,
                                 end_date=end_date)
        return await self.set(key, data, ttl)
    
    # Cache invalidation methods
    async def invalidate_form_cache(self, form_id: str) -> int:
        """Invalidate all cache entries for a form."""
        pattern = f"{settings.cache_prefix}:*:form_id:{form_id}*"
        return await self.delete_pattern(pattern)
    
    async def invalidate_question_cache(self, form_id: str, question_id: str) -> int:
        """Invalidate cache entries for a specific question."""
        pattern = f"{settings.cache_prefix}:*:form_id:{form_id}*:question_id:{question_id}*"
        return await self.delete_pattern(pattern)
    
    async def invalidate_all_analytics_cache(self) -> int:
        """Invalidate all analytics cache."""
        pattern = f"{settings.cache_prefix}:*"
        return await self.delete_pattern(pattern)
    
    # Health check
    async def health_check(self) -> Dict[str, Any]:
        """Check Redis connection health."""
        try:
            info = self.redis_client.info()
            return {
                "status": "healthy",
                "connected_clients": info.get("connected_clients", 0),
                "used_memory": info.get("used_memory_human", "unknown"),
                "uptime": info.get("uptime_in_seconds", 0)
            }
        except Exception as e:
            return {
                "status": "unhealthy",
                "error": str(e)
            }
    
    # Statistics
    async def get_cache_stats(self) -> Dict[str, Any]:
        """Get cache statistics."""
        try:
            info = self.redis_client.info()
            keyspace_info = self.redis_client.info("keyspace")
            
            # Count keys by pattern
            analytics_keys = len(self.redis_client.keys(f"{settings.cache_prefix}:*"))
            
            return {
                "total_keys": analytics_keys,
                "memory_usage": info.get("used_memory_human", "unknown"),
                "cache_hits": info.get("keyspace_hits", 0),
                "cache_misses": info.get("keyspace_misses", 0),
                "hit_rate": self._calculate_hit_rate(
                    info.get("keyspace_hits", 0),
                    info.get("keyspace_misses", 0)
                ),
                "uptime_seconds": info.get("uptime_in_seconds", 0),
                "connected_clients": info.get("connected_clients", 0)
            }
        except Exception as e:
            logger.error(f"Error getting cache stats: {e}")
            return {"error": str(e)}
    
    def _calculate_hit_rate(self, hits: int, misses: int) -> float:
        """Calculate cache hit rate percentage."""
        total = hits + misses
        if total == 0:
            return 0.0
        return (hits / total) * 100
    
    async def clear_expired_keys(self) -> int:
        """Clear expired keys (Redis handles this automatically, but manual cleanup)."""
        try:
            # Get all keys with our prefix
            pattern = f"{settings.cache_prefix}:*"
            keys = self.redis_client.keys(pattern)
            
            expired_count = 0
            for key in keys:
                ttl = self.redis_client.ttl(key)
                if ttl == -2:  # Key doesn't exist (expired)
                    expired_count += 1
            
            return expired_count
        except Exception as e:
            logger.error(f"Error clearing expired keys: {e}")
            return 0


# Global cache service instance
cache_service = CacheService()
