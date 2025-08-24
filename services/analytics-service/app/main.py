"""
Analytics Service - FastAPI Application
"""
import logging
from contextlib import asynccontextmanager
from fastapi import FastAPI, Request, status
from fastapi.middleware.cors import CORSMiddleware
from fastapi.middleware.trustedhost import TrustedHostMiddleware
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
import uvicorn

from app.config import settings
from app.api.analytics import router as analytics_router
from app.services.analytics_service import analytics_service

# Configure logging
logging.basicConfig(
    level=getattr(logging, settings.log_level.upper()),
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan manager."""
    # Startup
    logger.info("Starting Analytics Service...")
    try:
        await analytics_service.initialize()
        logger.info("Analytics Service initialized successfully")
    except Exception as e:
        logger.error(f"Failed to initialize Analytics Service: {e}")
        raise
    
    yield
    
    # Shutdown
    logger.info("Shutting down Analytics Service...")


# Create FastAPI application
app = FastAPI(
    title="X-Form Analytics Service",
    description="Analytics and reporting service for X-Form Backend with BigQuery integration",
    version="1.0.0",
    docs_url="/docs" if settings.environment != "production" else None,
    redoc_url="/redoc" if settings.environment != "production" else None,
    lifespan=lifespan
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origins,
    allow_credentials=True,
    allow_methods=settings.cors_methods,
    allow_headers=settings.cors_headers,
    max_age=3600,
)

# Add trusted host middleware for production
if settings.environment == "production":
    app.add_middleware(
        TrustedHostMiddleware,
        allowed_hosts=["localhost", "127.0.0.1", "analytics-service", "*.analytics.internal"]
    )


# Global exception handlers
@app.exception_handler(RequestValidationError)
async def validation_exception_handler(request: Request, exc: RequestValidationError):
    """Handle request validation errors."""
    logger.warning(f"Validation error on {request.url}: {exc}")
    return JSONResponse(
        status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
        content={
            "success": False,
            "error": "validation_error",
            "message": "Request validation failed",
            "details": exc.errors()
        }
    )


@app.exception_handler(Exception)
async def global_exception_handler(request: Request, exc: Exception):
    """Handle unexpected errors."""
    logger.error(f"Unexpected error on {request.url}: {exc}", exc_info=True)
    return JSONResponse(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        content={
            "success": False,
            "error": "internal_server_error",
            "message": "An unexpected error occurred"
        }
    )


# Health check endpoint
@app.get("/health")
async def health_check():
    """Basic health check endpoint."""
    return {
        "service": "analytics",
        "status": "healthy",
        "version": "1.0.0",
        "environment": settings.environment
    }


@app.get("/")
async def root():
    """Root endpoint with service information."""
    return {
        "service": "X-Form Analytics Service",
        "version": "1.0.0",
        "description": "Analytics and reporting service with BigQuery integration",
        "environment": settings.environment,
        "docs_url": "/docs" if settings.environment != "production" else None
    }


# Include routers
app.include_router(analytics_router)


# Request middleware for logging
@app.middleware("http")
async def log_requests(request: Request, call_next):
    """Log all requests for monitoring."""
    start_time = time.time()
    
    # Log request
    logger.info(f"{request.method} {request.url} - Client: {request.client.host if request.client else 'unknown'}")
    
    # Process request
    response = await call_next(request)
    
    # Log response
    process_time = time.time() - start_time
    logger.info(f"{request.method} {request.url} - Status: {response.status_code} - Time: {process_time:.3f}s")
    
    # Add timing header
    response.headers["X-Process-Time"] = str(process_time)
    
    return response


if __name__ == "__main__":
    import time
    
    # Run the application
    uvicorn.run(
        "app.main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.environment == "development",
        log_level=settings.log_level.lower(),
        access_log=True
    )
