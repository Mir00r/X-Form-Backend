"""
Analytics API Routes with Comprehensive Swagger Documentation
"""
import logging
from datetime import datetime
from typing import Optional, List
from fastapi import APIRouter, HTTPException, Query, Depends, status, Path
from fastapi.responses import JSONResponse
from fastapi.security import HTTPBearer

from app.models.analytics import (
    AnalyticsResponse, ErrorResponse, PeriodType, 
    FormSummary, QuestionAnalytics, TrendAnalysis
)
from app.services.analytics_service import analytics_service
from app.utils.auth import verify_token, get_current_user
from app.utils.rate_limiter import rate_limit

logger = logging.getLogger(__name__)
security = HTTPBearer()

router = APIRouter(prefix="/analytics", tags=["analytics"])


@router.get(
    "/{form_id}/summary", 
    response_model=AnalyticsResponse,
    status_code=status.HTTP_200_OK,
    summary="Get Form Analytics Summary",
    description="""
    **Get comprehensive analytics summary for a specific form.**
    
    This endpoint provides detailed analytics including:
    - Total and completed response counts
    - Completion rates and trends
    - Average completion time
    - Response patterns and distribution
    - Performance benchmarks
    
    ## Parameters
    
    - **form_id**: Unique identifier for the form
    - **start_date**: Optional start date for analytics range (ISO format)
    - **end_date**: Optional end date for analytics range (ISO format)  
    - **use_cache**: Whether to use cached data for faster response
    
    ## Response
    
    Returns a comprehensive analytics summary with:
    - Form metadata and basic statistics
    - Completion rate analysis
    - Response trend data (last 30 days)
    - Question-level analytics summary
    - Performance indicators
    
    ## Rate Limiting
    
    This endpoint is rate limited to 100 requests per hour per user.
    
    ## Authentication
    
    Requires valid JWT Bearer token or API key authentication.
    """,
    responses={
        200: {
            "description": "Analytics summary retrieved successfully",
            "content": {
                "application/json": {
                    "example": {
                        "success": True,
                        "message": "Form analytics summary retrieved successfully",
                        "data": {
                            "form_id": "550e8400-e29b-41d4-a716-446655440000",
                            "title": "Customer Satisfaction Survey",
                            "total_responses": 1523,
                            "completed_responses": 1289,
                            "completion_rate": 84.6,
                            "average_completion_time": 145.7,
                            "response_trend": [
                                {"date": "2025-09-01", "count": 45},
                                {"date": "2025-09-02", "count": 52}
                            ],
                            "performance_score": 8.7
                        },
                        "timestamp": "2025-09-06T12:00:00Z"
                    }
                }
            }
        },
        400: {
            "description": "Invalid request parameters",
            "content": {
                "application/json": {
                    "example": {
                        "success": False,
                        "error": "validation_error",
                        "message": "Start date must be before end date",
                        "timestamp": "2025-09-06T12:00:00Z"
                    }
                }
            }
        },
        401: {
            "description": "Authentication required",
            "content": {
                "application/json": {
                    "example": {
                        "success": False,
                        "error": "authentication_required",
                        "message": "Valid JWT token required",
                        "timestamp": "2025-09-06T12:00:00Z"
                    }
                }
            }
        },
        403: {
            "description": "Access forbidden",
            "content": {
                "application/json": {
                    "example": {
                        "success": False,
                        "error": "access_forbidden",
                        "message": "Insufficient permissions to access this form",
                        "timestamp": "2025-09-06T12:00:00Z"
                    }
                }
            }
        },
        404: {
            "description": "Form not found",
            "content": {
                "application/json": {
                    "example": {
                        "success": False,
                        "error": "form_not_found",
                        "message": "Form with specified ID does not exist",
                        "timestamp": "2025-09-06T12:00:00Z"
                    }
                }
            }
        },
        429: {
            "description": "Rate limit exceeded",
            "content": {
                "application/json": {
                    "example": {
                        "success": False,
                        "error": "rate_limit_exceeded",
                        "message": "Rate limit of 100 requests per hour exceeded",
                        "timestamp": "2025-09-06T12:00:00Z"
                    }
                }
            }
        },
        500: {
            "description": "Internal server error",
            "content": {
                "application/json": {
                    "example": {
                        "success": False,
                        "error": "internal_server_error", 
                        "message": "Failed to retrieve form analytics",
                        "timestamp": "2025-09-06T12:00:00Z"
                    }
                }
            }
        }
    },
    operation_id="getFormAnalyticsSummary",
    dependencies=[Depends(security)]
)
@rate_limit(max_requests=100, window_seconds=3600)  # 100 requests per hour
async def get_form_summary(
    form_id: str = Path(
        ..., 
        description="Unique identifier for the form",
        example="550e8400-e29b-41d4-a716-446655440000",
        regex="^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
    ),
    start_date: Optional[datetime] = Query(
        None, 
        description="Start date for analytics range (ISO 8601 format)",
        example="2025-09-01T00:00:00Z"
    ),
    end_date: Optional[datetime] = Query(
        None, 
        description="End date for analytics range (ISO 8601 format)",
        example="2025-09-06T23:59:59Z"
    ),
    use_cache: bool = Query(
        True, 
        description="Whether to use cached data for faster response"
    ),
    current_user: dict = Depends(get_current_user)
):
    """
    Get comprehensive analytics summary for a form.
    
    Returns:
    - Response statistics (total, completed, partial)
    - Completion rates
    - Response trends
    - Performance metrics
    - Visualizations
    """
    try:
        logger.info(f"Getting form summary for {form_id} by user {current_user.get('user_id')}")
        
        # Validate date range
        if start_date and end_date and start_date > end_date:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Start date must be before end date"
            )
        
        # Get analytics data
        result = await analytics_service.get_form_analytics_summary(
            form_id=form_id,
            start_date=start_date,
            end_date=end_date,
            use_cache=use_cache
        )
        
        return AnalyticsResponse(
            success=True,
            data=result,
            message="Form analytics summary retrieved successfully"
        )
        
    except Exception as e:
        logger.error(f"Error getting form summary for {form_id}: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to retrieve form analytics: {str(e)}"
        )


@router.get(
    "/{form_id}/question/{question_id}", 
    response_model=AnalyticsResponse,
    status_code=status.HTTP_200_OK,
    summary="Get Question Analytics",
    description="""
    **Get detailed analytics for a specific question within a form.**
    
    This endpoint provides question-specific analytics including:
    - Response distribution and patterns
    - Answer frequency analysis
    - Skip rates and completion metrics
    - Visual representations based on question type
    - Statistical analysis (mean, median, mode for numeric data)
    
    ## Question Type Support
    
    - **Multiple Choice**: Response distribution with percentages
    - **Text/Textarea**: Word clouds and sentiment analysis
    - **Rating/Scale**: Average scores and distribution
    - **Number**: Statistical summaries and ranges
    - **Date/Time**: Temporal patterns and trends
    
    ## Parameters
    
    - **form_id**: Unique identifier for the form
    - **question_id**: Unique identifier for the question
    - **start_date**: Optional start date for analytics range
    - **end_date**: Optional end date for analytics range
    - **question_type**: Type of question for appropriate visualization
    - **use_cache**: Whether to use cached data for faster response
    
    ## Response
    
    Returns detailed question analytics with:
    - Response count and completion rate
    - Answer distribution and frequency
    - Statistical analysis where applicable
    - Visualization data for charts
    - Insights and recommendations
    
    ## Rate Limiting
    
    This endpoint is rate limited to 200 requests per hour per user.
    """,
    responses={
        200: {
            "description": "Question analytics retrieved successfully",
            "content": {
                "application/json": {
                    "example": {
                        "success": True,
                        "message": "Question analytics retrieved successfully",
                        "data": {
                            "question_id": "q_rating_satisfaction",
                            "question_type": "rating",
                            "total_responses": 1289,
                            "response_rate": 94.8,
                            "answer_distribution": {
                                "1": 45, "2": 89, "3": 234, "4": 567, "5": 354
                            },
                            "statistics": {
                                "mean": 4.1,
                                "median": 4.0,
                                "mode": 4,
                                "std_dev": 1.2
                            },
                            "visualization": {
                                "type": "bar_chart",
                                "data": [45, 89, 234, 567, 354],
                                "labels": ["1", "2", "3", "4", "5"]
                            }
                        },
                        "timestamp": "2025-09-06T12:00:00Z"
                    }
                }
            }
        },
        400: {"description": "Invalid request parameters"},
        401: {"description": "Authentication required"},
        403: {"description": "Access forbidden"},
        404: {"description": "Form or question not found"},
        429: {"description": "Rate limit exceeded"},
        500: {"description": "Internal server error"}
    },
    operation_id="getQuestionAnalytics",
    dependencies=[Depends(security)]
)
@rate_limit(max_requests=200, window_seconds=3600)  # 200 requests per hour
async def get_question_analytics(
    form_id: str = Path(
        ..., 
        description="Unique identifier for the form",
        example="550e8400-e29b-41d4-a716-446655440000"
    ),
    question_id: str = Path(
        ..., 
        description="Unique identifier for the question",
        example="q_rating_satisfaction"
    ),
    start_date: Optional[datetime] = Query(
        None, 
        description="Start date for analytics range (ISO 8601 format)",
        example="2025-09-01T00:00:00Z"
    ),
    end_date: Optional[datetime] = Query(
        None, 
        description="End date for analytics range (ISO 8601 format)",
        example="2025-09-06T23:59:59Z"
    ),
    question_type: str = Query(
        "multiple_choice", 
        description="Type of question for appropriate visualization",
        enum=["text", "textarea", "number", "rating", "scale", "multiple_choice", "checkbox", "radio", "date", "time"]
    ),
    use_cache: bool = Query(
        True, 
        description="Whether to use cached data for faster response"
    ),
    current_user: dict = Depends(get_current_user)
):
    """
    Get detailed analytics for a specific question.
    
    Returns:
    - Response distribution
    - Answer patterns
    - Response rates
    - Skip rates
    - Visualizations based on question type
    """
    try:
        logger.info(f"Getting question analytics for {form_id}/{question_id} by user {current_user.get('user_id')}")
        
        # Validate date range
        if start_date and end_date and start_date > end_date:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Start date must be before end date"
            )
        
        # Get analytics data
        result = await analytics_service.get_question_analytics(
            form_id=form_id,
            question_id=question_id,
            start_date=start_date,
            end_date=end_date,
            question_type=question_type,
            use_cache=use_cache
        )
        
        return AnalyticsResponse(
            success=True,
            data=result,
            message="Question analytics retrieved successfully"
        )
        
    except Exception as e:
        logger.error(f"Error getting question analytics for {form_id}/{question_id}: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to retrieve question analytics: {str(e)}"
        )


@router.get("/{form_id}/trend", response_model=AnalyticsResponse)
@rate_limit(max_requests=50, window_seconds=3600)  # 50 requests per hour
async def get_trend_analysis(
    form_id: str,
    period: PeriodType = Query(PeriodType.DAY, description="Time period for trend analysis"),
    start_date: Optional[datetime] = Query(None, description="Start date for trend analysis"),
    end_date: Optional[datetime] = Query(None, description="End date for trend analysis"),
    use_cache: bool = Query(True, description="Whether to use cached data"),
    current_user: dict = Depends(get_current_user)
):
    """
    Get trend analysis for form responses over time.
    
    Returns:
    - Time-series data
    - Trend patterns
    - Growth rates
    - Statistical analysis
    - Trend visualizations
    """
    try:
        logger.info(f"Getting trend analysis for {form_id} by user {current_user.get('user_id')}")
        
        # Validate date range
        if start_date and end_date and start_date > end_date:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Start date must be before end date"
            )
        
        # Get trend data
        result = await analytics_service.get_trend_analysis(
            form_id=form_id,
            period=period,
            start_date=start_date,
            end_date=end_date,
            use_cache=use_cache
        )
        
        return AnalyticsResponse(
            success=True,
            data=result,
            message="Trend analysis retrieved successfully"
        )
        
    except Exception as e:
        logger.error(f"Error getting trend analysis for {form_id}: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to retrieve trend analysis: {str(e)}"
        )


@router.get("/compare", response_model=AnalyticsResponse)
@rate_limit(max_requests=20, window_seconds=3600)  # 20 requests per hour
async def get_comparative_analytics(
    form_ids: List[str] = Query(..., description="List of form IDs to compare"),
    metric: str = Query("response_count", description="Metric to compare"),
    period: PeriodType = Query(PeriodType.DAY, description="Time period for comparison"),
    start_date: Optional[datetime] = Query(None, description="Start date for comparison"),
    end_date: Optional[datetime] = Query(None, description="End date for comparison"),
    current_user: dict = Depends(get_current_user)
):
    """
    Compare analytics across multiple forms.
    
    Returns:
    - Comparative metrics
    - Side-by-side analysis
    - Performance comparison
    - Comparative visualizations
    """
    try:
        logger.info(f"Getting comparative analytics for forms {form_ids} by user {current_user.get('user_id')}")
        
        # Validate inputs
        if len(form_ids) < 2:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="At least 2 forms required for comparison"
            )
        
        if len(form_ids) > 10:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Maximum 10 forms allowed for comparison"
            )
        
        if start_date and end_date and start_date > end_date:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Start date must be before end date"
            )
        
        # Get comparative data
        result = await analytics_service.get_comparative_analytics(
            form_ids=form_ids,
            metric=metric,
            period=period,
            start_date=start_date,
            end_date=end_date
        )
        
        return AnalyticsResponse(
            success=True,
            data=result,
            message="Comparative analytics retrieved successfully"
        )
        
    except Exception as e:
        logger.error(f"Error getting comparative analytics: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to retrieve comparative analytics: {str(e)}"
        )


@router.post("/{form_id}/cache/invalidate")
@rate_limit(max_requests=10, window_seconds=300)  # 10 requests per 5 minutes
async def invalidate_cache(
    form_id: str,
    question_id: Optional[str] = Query(None, description="Specific question ID to invalidate"),
    current_user: dict = Depends(get_current_user)
):
    """
    Invalidate analytics cache for a form or specific question.
    
    Use this endpoint when form data has been updated and you need fresh analytics.
    """
    try:
        logger.info(f"Invalidating cache for {form_id} by user {current_user.get('user_id')}")
        
        result = await analytics_service.invalidate_cache(
            form_id=form_id,
            question_id=question_id
        )
        
        return JSONResponse(
            status_code=status.HTTP_200_OK,
            content={
                "success": True,
                "data": result,
                "message": "Cache invalidated successfully"
            }
        )
        
    except Exception as e:
        logger.error(f"Error invalidating cache for {form_id}: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to invalidate cache: {str(e)}"
        )


@router.get("/health")
async def health_check():
    """
    Check the health status of the analytics service.
    
    Returns status of all service components.
    """
    try:
        health_status = await analytics_service.get_service_health()
        
        status_code = status.HTTP_200_OK if health_status["status"] == "healthy" else status.HTTP_503_SERVICE_UNAVAILABLE
        
        return JSONResponse(
            status_code=status_code,
            content=health_status
        )
        
    except Exception as e:
        logger.error(f"Error checking service health: {e}")
        return JSONResponse(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            content={
                "service": "analytics",
                "status": "unhealthy",
                "error": str(e),
                "timestamp": datetime.now().isoformat()
            }
        )


@router.get("/cache/stats")
async def get_cache_statistics(
    current_user: dict = Depends(get_current_user)
):
    """
    Get cache performance statistics.
    
    Returns cache hit rates, memory usage, and other performance metrics.
    """
    try:
        logger.info(f"Getting cache statistics by user {current_user.get('user_id')}")
        
        stats = await analytics_service.get_cache_statistics()
        
        return JSONResponse(
            status_code=status.HTTP_200_OK,
            content={
                "success": True,
                "data": stats,
                "message": "Cache statistics retrieved successfully"
            }
        )
        
    except Exception as e:
        logger.error(f"Error getting cache statistics: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to retrieve cache statistics: {str(e)}"
        )
