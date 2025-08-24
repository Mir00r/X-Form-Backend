"""
API module for Analytics Service
"""

from .analytics import router as analytics_router

__all__ = [
    "analytics_router"
]
