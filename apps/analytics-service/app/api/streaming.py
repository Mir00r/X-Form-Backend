"""
Streaming API Routes for Real-time Analytics
"""
import logging
import json
from datetime import datetime
from typing import Optional, AsyncGenerator
from fastapi import APIRouter, HTTPException, Query, Depends, status, Path
from fastapi.responses import StreamingResponse
from fastapi.security import HTTPBearer
from pydantic import BaseModel

from app.models.analytics import AnalyticsResponse
from app.services.analytics_service import analytics_service
from app.utils.auth import get_current_user
from app.utils.rate_limiter import rate_limit

logger = logging.getLogger(__name__)
security = HTTPBearer()

router = APIRouter(prefix="/streaming", tags=["streaming"])

class StreamConfig(BaseModel):
    """Configuration for streaming analytics."""
    interval_seconds: int = 5
    include_metadata: bool = True
    format: str = "json"
    
    class Config:
        schema_extra = {
            "example": {
                "interval_seconds": 10,
                "include_metadata": True,
                "format": "json"
            }
        }

@router.get(
    "/{form_id}/live",
    summary="Live Analytics Stream",
    description="""
    **Stream real-time analytics data for a form.**
    
    This endpoint provides real-time analytics data using Server-Sent Events (SSE).
    
    ## Stream Features
    
    - **Real-time Updates**: Get analytics data as responses are submitted
    - **Configurable Intervals**: Set update frequency (1-60 seconds)
    - **Multiple Formats**: JSON, CSV, or structured text
    - **Filtered Data**: Stream specific metrics or data points
    - **Connection Management**: Automatic reconnection and error handling
    
    ## Use Cases
    
    - **Live Dashboards**: Real-time analytics dashboards
    - **Monitoring Systems**: Alert systems for form performance
    - **Data Integration**: Feed live data to external systems
    - **Research Applications**: Real-time data collection monitoring
    
    ## Data Structure
    
    Each stream message contains:
    - **timestamp**: When the data was generated
    - **event_type**: Type of update (new_response, analytics_update, etc.)
    - **data**: The actual analytics data
    - **metadata**: Additional context information
    
    ## Rate Limiting
    
    This endpoint is rate limited to 50 requests per hour per user.
    """,
    responses={
        200: {
            "description": "Analytics stream established",
            "content": {
                "text/event-stream": {
                    "example": "data: {\"timestamp\": \"2025-09-06T12:00:00Z\", \"event_type\": \"analytics_update\", \"data\": {\"total_responses\": 1524, \"completion_rate\": 84.7}}\n\n"
                }
            }
        },
        400: {"description": "Invalid stream parameters"},
        401: {"description": "Authentication required"},
        403: {"description": "Access forbidden"},
        404: {"description": "Form not found"},
        429: {"description": "Rate limit exceeded"},
        500: {"description": "Stream initialization failed"}
    },
    operation_id="streamLiveAnalytics",
    dependencies=[Depends(security)]
)
@rate_limit(max_requests=50, window_seconds=3600)  # 50 requests per hour
async def stream_live_analytics(
    form_id: str = Path(
        ..., 
        description="Unique identifier for the form",
        example="550e8400-e29b-41d4-a716-446655440000"
    ),
    interval: int = Query(
        5, 
        description="Update interval in seconds (1-60)",
        ge=1, 
        le=60
    ),
    metrics: Optional[str] = Query(
        None, 
        description="Comma-separated list of specific metrics to stream",
        example="total_responses,completion_rate,response_trend"
    ),
    format: str = Query(
        "json", 
        description="Stream data format",
        enum=["json", "csv", "text"]
    ),
    current_user: dict = Depends(get_current_user)
):
    """Stream live analytics data for a form."""
    try:
        logger.info(f"Starting live stream for form {form_id} by user {current_user.get('user_id')}")
        
        async def generate_stream() -> AsyncGenerator[str, None]:
            """Generate real-time analytics stream."""
            try:
                async for analytics_data in analytics_service.stream_live_analytics(
                    form_id=form_id,
                    interval_seconds=interval,
                    metrics=metrics.split(',') if metrics else None,
                    user_id=current_user.get('user_id')
                ):
                    # Format data based on requested format
                    if format == "json":
                        formatted_data = json.dumps(analytics_data)
                    elif format == "csv":
                        formatted_data = analytics_service.format_as_csv(analytics_data)
                    else:  # text
                        formatted_data = analytics_service.format_as_text(analytics_data)
                    
                    # Send as Server-Sent Event
                    yield f"data: {formatted_data}\n\n"
                    
            except Exception as e:
                logger.error(f"Stream error for form {form_id}: {e}")
                error_data = {
                    "error": "stream_error",
                    "message": str(e),
                    "timestamp": datetime.utcnow().isoformat()
                }
                yield f"data: {json.dumps(error_data)}\n\n"
        
        return StreamingResponse(
            generate_stream(),
            media_type="text/event-stream",
            headers={
                "Cache-Control": "no-cache",
                "Connection": "keep-alive",
                "X-Accel-Buffering": "no"  # Disable nginx buffering
            }
        )
        
    except Exception as e:
        logger.error(f"Failed to start stream for form {form_id}: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to start analytics stream: {str(e)}"
        )

@router.get(
    "/{form_id}/events",
    summary="Form Events Stream", 
    description="""
    **Stream real-time form events as they occur.**
    
    This endpoint streams individual form events in real-time:
    
    ## Event Types
    
    - **response_started**: User began filling out the form
    - **response_submitted**: User completed and submitted the form
    - **response_abandoned**: User left without completing
    - **question_answered**: Individual question was answered
    - **page_changed**: User moved to different form page
    - **validation_error**: Form validation error occurred
    
    ## Event Data
    
    Each event includes:
    - **event_id**: Unique identifier for the event
    - **event_type**: Type of event that occurred
    - **timestamp**: When the event occurred
    - **form_id**: Form identifier
    - **session_id**: User session identifier
    - **data**: Event-specific data payload
    - **metadata**: Additional context (IP, user agent, etc.)
    
    ## Use Cases
    
    - **Real-time Monitoring**: Monitor form usage in real-time
    - **User Behavior Analysis**: Track user interactions
    - **Abandonment Alerts**: Get notified of form abandonment
    - **A/B Testing**: Real-time comparison of form variants
    """,
    responses={
        200: {
            "description": "Event stream established",
            "content": {
                "text/event-stream": {
                    "example": "data: {\"event_id\": \"evt_123\", \"event_type\": \"response_submitted\", \"timestamp\": \"2025-09-06T12:05:00Z\", \"data\": {\"completion_time\": 145}}\n\n"
                }
            }
        },
        400: {"description": "Invalid parameters"},
        401: {"description": "Authentication required"},
        404: {"description": "Form not found"},
        429: {"description": "Rate limit exceeded"},
        500: {"description": "Stream initialization failed"}
    },
    operation_id="streamFormEvents",
    dependencies=[Depends(security)]
)
async def stream_form_events(
    form_id: str = Path(..., description="Unique identifier for the form"),
    event_types: Optional[str] = Query(
        None, 
        description="Comma-separated list of event types to stream",
        example="response_submitted,response_abandoned"
    ),
    include_metadata: bool = Query(
        True, 
        description="Whether to include event metadata"
    ),
    current_user: dict = Depends(get_current_user)
):
    """Stream real-time form events."""
    try:
        logger.info(f"Starting event stream for form {form_id} by user {current_user.get('user_id')}")
        
        async def generate_event_stream() -> AsyncGenerator[str, None]:
            """Generate real-time event stream."""
            try:
                async for event_data in analytics_service.stream_form_events(
                    form_id=form_id,
                    event_types=event_types.split(',') if event_types else None,
                    include_metadata=include_metadata,
                    user_id=current_user.get('user_id')
                ):
                    formatted_event = json.dumps(event_data)
                    yield f"data: {formatted_event}\n\n"
                    
            except Exception as e:
                logger.error(f"Event stream error for form {form_id}: {e}")
                error_event = {
                    "event_type": "stream_error",
                    "error": str(e),
                    "timestamp": datetime.utcnow().isoformat()
                }
                yield f"data: {json.dumps(error_event)}\n\n"
        
        return StreamingResponse(
            generate_event_stream(),
            media_type="text/event-stream",
            headers={
                "Cache-Control": "no-cache",
                "Connection": "keep-alive"
            }
        )
        
    except Exception as e:
        logger.error(f"Failed to start event stream for form {form_id}: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to start event stream: {str(e)}"
        )

@router.post(
    "/{form_id}/webhooks",
    status_code=status.HTTP_201_CREATED,
    summary="Configure Analytics Webhook",
    description="""
    **Configure a webhook to receive analytics updates.**
    
    Webhooks provide a way to receive analytics data pushed to your own endpoints.
    
    ## Webhook Configuration
    
    - **URL**: Your endpoint URL to receive data
    - **Events**: Types of analytics events to receive
    - **Format**: Data format (JSON, XML, form-data)
    - **Authentication**: Optional authentication headers
    - **Frequency**: How often to send updates
    
    ## Webhook Events
    
    - **analytics_updated**: Analytics data has been recalculated
    - **threshold_exceeded**: Metric exceeded configured threshold
    - **report_generated**: Scheduled report is ready
    - **insight_generated**: New AI insight is available
    
    ## Security
    
    - **HTTPS Required**: All webhook URLs must use HTTPS
    - **Signature Verification**: Webhooks include signature for verification
    - **Retry Logic**: Failed deliveries are retried with exponential backoff
    """,
    operation_id="configureAnalyticsWebhook",
    dependencies=[Depends(security)]
)
async def configure_webhook(
    form_id: str = Path(..., description="Unique identifier for the form"),
    webhook_url: str = Query(..., description="Your webhook endpoint URL"),
    events: str = Query(..., description="Comma-separated list of events"),
    secret: Optional[str] = Query(None, description="Secret for webhook signature"),
    current_user: dict = Depends(get_current_user)
):
    """Configure a webhook for analytics updates."""
    try:
        webhook_id = await analytics_service.configure_webhook(
            form_id=form_id,
            webhook_url=webhook_url,
            events=events.split(','),
            secret=secret,
            user_id=current_user.get('user_id')
        )
        
        return {
            "success": True,
            "message": "Webhook configured successfully",
            "data": {
                "webhook_id": webhook_id,
                "url": webhook_url,
                "events": events.split(',')
            }
        }
        
    except Exception as e:
        logger.error(f"Failed to configure webhook for form {form_id}: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to configure webhook: {str(e)}"
        )
