"""
AWS Lambda Handler for File Upload Service

This module provides the entry point for AWS Lambda execution.
It configures the FastAPI application and uses Mangum for ASGI adaptation.
"""

import os
import json
from typing import Dict, Any
from mangum import Mangum
import structlog

from .configuration import create_configured_app


# Configure structured logging
structlog.configure(
    processors=[
        structlog.stdlib.filter_by_level,
        structlog.stdlib.add_logger_name,
        structlog.stdlib.add_log_level,
        structlog.stdlib.PositionalArgumentsFormatter(),
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.StackInfoRenderer(),
        structlog.processors.format_exc_info,
        structlog.processors.UnicodeDecoder(),
        structlog.processors.JSONRenderer()
    ],
    context_class=dict,
    logger_factory=structlog.stdlib.LoggerFactory(),
    wrapper_class=structlog.stdlib.BoundLogger,
    cache_logger_on_first_use=True,
)

logger = structlog.get_logger()

# Create the FastAPI app
app = create_configured_app()

# Create the Lambda handler using Mangum
handler = Mangum(app, lifespan="off")


def lambda_handler(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """
    AWS Lambda entry point
    
    Args:
        event: Lambda event data
        context: Lambda context
        
    Returns:
        HTTP response in Lambda format
    """
    try:
        # Log the incoming request
        logger.info(
            "Processing Lambda request",
            http_method=event.get("httpMethod"),
            path=event.get("path"),
            source_ip=event.get("requestContext", {}).get("identity", {}).get("sourceIp"),
            user_agent=event.get("headers", {}).get("User-Agent")
        )
        
        # Process the request through Mangum
        response = handler(event, context)
        
        # Log the response
        logger.info(
            "Lambda request processed",
            status_code=response.get("statusCode"),
            response_size=len(json.dumps(response))
        )
        
        return response
        
    except Exception as e:
        logger.error(
            "Unhandled error in Lambda handler",
            error=str(e),
            error_type=type(e).__name__
        )
        
        # Return a generic error response
        return {
            "statusCode": 500,
            "headers": {
                "Content-Type": "application/json",
                "Access-Control-Allow-Origin": "*"
            },
            "body": json.dumps({
                "error": "internal_server_error",
                "message": "An unexpected error occurred"
            })
        }


# Health check handler for Lambda container reuse
def health_check_handler(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """
    Simple health check handler
    
    Can be used for Lambda warmup or container health checks
    """
    return {
        "statusCode": 200,
        "headers": {
            "Content-Type": "application/json"
        },
        "body": json.dumps({
            "status": "healthy",
            "service": "file-upload-service",
            "version": "1.0.0",
            "timestamp": context.aws_request_id if hasattr(context, 'aws_request_id') else None
        })
    }


# For local development with uvicorn
if __name__ == "__main__":
    import uvicorn
    
    # Run locally for development
    uvicorn.run(
        app,
        host="0.0.0.0",
        port=int(os.getenv("PORT", "8000")),
        log_level=os.getenv("LOG_LEVEL", "info").lower(),
        reload=os.getenv("RELOAD", "false").lower() == "true"
    )
