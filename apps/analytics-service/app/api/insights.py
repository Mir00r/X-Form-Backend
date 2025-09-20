"""
Insights API Routes with AI-Powered Analytics
"""
import logging
from datetime import datetime
from typing import Optional, List, Dict, Any
from fastapi import APIRouter, HTTPException, Query, Depends, status, Path
from fastapi.security import HTTPBearer
from pydantic import BaseModel

from app.models.analytics import AnalyticsResponse
from app.services.analytics_service import analytics_service
from app.utils.auth import get_current_user
from app.utils.rate_limiter import rate_limit

logger = logging.getLogger(__name__)
security = HTTPBearer()

router = APIRouter(prefix="/insights", tags=["insights"])

# Response Models
class InsightItem(BaseModel):
    """Individual insight item."""
    type: str
    title: str
    description: str
    confidence: float
    impact: str
    recommendation: str
    data: Dict[str, Any]
    
    class Config:
        schema_extra = {
            "example": {
                "type": "completion_rate_drop",
                "title": "Completion Rate Decline Detected",
                "description": "Form completion rate has dropped by 15% in the last week",
                "confidence": 0.87,
                "impact": "high",
                "recommendation": "Review questions 5-7 which show high abandonment rates",
                "data": {
                    "previous_rate": 84.6,
                    "current_rate": 71.8,
                    "change_percent": -15.1,
                    "affected_questions": ["q5", "q6", "q7"]
                }
            }
        }

class InsightsResponse(BaseModel):
    """Complete insights response."""
    form_id: str
    generated_at: datetime
    insights: List[InsightItem]
    summary: Dict[str, Any]
    
    class Config:
        schema_extra = {
            "example": {
                "form_id": "550e8400-e29b-41d4-a716-446655440000",
                "generated_at": "2025-09-06T12:00:00Z",
                "insights": [
                    {
                        "type": "performance_trend",
                        "title": "Response Volume Increasing",
                        "description": "Daily response volume has increased by 23% over the past month",
                        "confidence": 0.92,
                        "impact": "positive",
                        "recommendation": "Consider scaling infrastructure to handle increased load",
                        "data": {"growth_rate": 23.4, "trend": "upward"}
                    }
                ],
                "summary": {
                    "total_insights": 5,
                    "high_impact": 2,
                    "medium_impact": 2,
                    "low_impact": 1,
                    "overall_health_score": 8.7
                }
            }
        }

@router.get(
    "/{form_id}",
    response_model=AnalyticsResponse,
    status_code=status.HTTP_200_OK,
    summary="Get AI-Powered Form Insights",
    description="""
    **Get AI-powered insights and recommendations for form performance.**
    
    This endpoint uses machine learning algorithms to analyze form data and provide:
    
    ## Insight Types
    
    - **Performance Trends**: Response volume and completion rate changes
    - **User Behavior**: Patterns in how users interact with the form
    - **Abandonment Analysis**: Identification of drop-off points
    - **Question Optimization**: Suggestions for improving specific questions
    - **Timing Patterns**: Optimal times for form distribution
    - **Demographic Insights**: Response patterns across user segments
    - **Predictive Analytics**: Forecasts for future performance
    
    ## AI Features
    
    - **Anomaly Detection**: Automatically identify unusual patterns
    - **Trend Analysis**: Detect emerging trends in response data
    - **Predictive Modeling**: Forecast future form performance
    - **Recommendation Engine**: Actionable suggestions for improvement
    - **Sentiment Analysis**: Analyze text responses for sentiment
    - **Statistical Significance**: Ensure insights are statistically valid
    
    ## Confidence Scores
    
    Each insight includes a confidence score (0.0 to 1.0) indicating:
    - **0.9-1.0**: Very high confidence, strong statistical evidence
    - **0.7-0.9**: High confidence, good statistical support
    - **0.5-0.7**: Medium confidence, some evidence present
    - **0.3-0.5**: Low confidence, tentative insights
    - **0.0-0.3**: Very low confidence, insufficient data
    
    ## Impact Levels
    
    - **High**: Significant impact on form performance (>15% change)
    - **Medium**: Moderate impact (5-15% change)
    - **Low**: Minor impact (<5% change)
    
    ## Parameters
    
    - **form_id**: Unique identifier for the form
    - **insight_types**: Filter for specific types of insights
    - **min_confidence**: Minimum confidence threshold (0.0-1.0)
    - **include_predictions**: Whether to include predictive insights
    - **date_range_days**: Number of days to analyze (default: 30)
    """,
    responses={
        200: {
            "description": "Insights generated successfully",
            "content": {
                "application/json": {
                    "example": {
                        "success": True,
                        "message": "AI insights generated successfully",
                        "data": {
                            "form_id": "550e8400-e29b-41d4-a716-446655440000",
                            "insights": [
                                {
                                    "type": "completion_optimization",
                                    "title": "Question Length Optimization",
                                    "description": "Questions with >50 words show 25% higher abandonment",
                                    "confidence": 0.89,
                                    "impact": "high",
                                    "recommendation": "Shorten questions 3, 7, and 12 to improve completion rates"
                                }
                            ],
                            "summary": {
                                "total_insights": 8,
                                "actionable_recommendations": 5,
                                "predicted_improvement": "12-18% completion rate increase"
                            }
                        },
                        "timestamp": "2025-09-06T12:00:00Z"
                    }
                }
            }
        },
        400: {"description": "Invalid parameters"},
        401: {"description": "Authentication required"},
        403: {"description": "Access forbidden"},
        404: {"description": "Form not found"},
        429: {"description": "Rate limit exceeded"},
        500: {"description": "Insight generation failed"}
    },
    operation_id="getFormInsights",
    dependencies=[Depends(security)]
)
@rate_limit(max_requests=20, window_seconds=3600)  # 20 requests per hour
async def get_form_insights(
    form_id: str = Path(
        ..., 
        description="Unique identifier for the form",
        example="550e8400-e29b-41d4-a716-446655440000"
    ),
    insight_types: Optional[List[str]] = Query(
        None, 
        description="Filter for specific insight types",
        example=["performance_trends", "user_behavior", "optimization"]
    ),
    min_confidence: float = Query(
        0.5, 
        description="Minimum confidence threshold (0.0-1.0)",
        ge=0.0, 
        le=1.0
    ),
    include_predictions: bool = Query(
        True, 
        description="Whether to include predictive insights"
    ),
    date_range_days: int = Query(
        30, 
        description="Number of days to analyze (max 365)",
        ge=1, 
        le=365
    ),
    current_user: dict = Depends(get_current_user)
):
    """Generate AI-powered insights for form performance."""
    try:
        logger.info(f"Generating insights for form {form_id} by user {current_user.get('user_id')}")
        
        # Generate insights using AI service
        insights_data = await analytics_service.generate_ai_insights(
            form_id=form_id,
            insight_types=insight_types,
            min_confidence=min_confidence,
            include_predictions=include_predictions,
            date_range_days=date_range_days,
            user_id=current_user.get('user_id')
        )
        
        return AnalyticsResponse(
            success=True,
            message="AI insights generated successfully",
            data=insights_data,
            timestamp=datetime.utcnow()
        )
        
    except Exception as e:
        logger.error(f"Failed to generate insights for form {form_id}: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to generate insights: {str(e)}"
        )

@router.get(
    "/{form_id}/recommendations",
    status_code=status.HTTP_200_OK,
    summary="Get Optimization Recommendations",
    description="""
    **Get specific recommendations for optimizing form performance.**
    
    This endpoint provides actionable recommendations based on data analysis:
    
    ## Recommendation Categories
    
    - **Question Optimization**: Improve individual questions
    - **Flow Optimization**: Optimize question order and logic
    - **UI/UX Improvements**: Interface and usability enhancements
    - **Timing Optimization**: Best times for form distribution
    - **Content Optimization**: Improve text and messaging
    - **Technical Optimization**: Performance and loading improvements
    
    ## Implementation Priority
    
    - **Critical**: Immediate action required (>30% impact)
    - **High**: Should be implemented soon (15-30% impact)
    - **Medium**: Good to implement (5-15% impact)
    - **Low**: Nice to have (<5% impact)
    """,
    responses={
        200: {
            "description": "Recommendations generated successfully",
            "content": {
                "application/json": {
                    "example": {
                        "success": True,
                        "data": {
                            "recommendations": [
                                {
                                    "category": "question_optimization",
                                    "priority": "high",
                                    "title": "Simplify Question 7",
                                    "description": "Question 7 has a 40% abandonment rate",
                                    "action": "Reduce from 3 sub-questions to 1 main question",
                                    "expected_impact": "25% completion rate improvement",
                                    "implementation_effort": "low"
                                }
                            ],
                            "summary": {
                                "total_recommendations": 12,
                                "critical": 1,
                                "high": 4,
                                "medium": 5,
                                "low": 2
                            }
                        }
                    }
                }
            }
        }
    },
    operation_id="getOptimizationRecommendations",
    dependencies=[Depends(security)]
)
async def get_optimization_recommendations(
    form_id: str = Path(..., description="Unique identifier for the form"),
    categories: Optional[List[str]] = Query(
        None, 
        description="Filter by recommendation categories"
    ),
    min_impact: float = Query(
        0.05, 
        description="Minimum expected impact threshold (0.0-1.0)",
        ge=0.0, 
        le=1.0
    ),
    current_user: dict = Depends(get_current_user)
):
    """Get optimization recommendations for a form."""
    try:
        recommendations = await analytics_service.get_optimization_recommendations(
            form_id=form_id,
            categories=categories,
            min_impact=min_impact,
            user_id=current_user.get('user_id')
        )
        
        return {
            "success": True,
            "data": recommendations
        }
        
    except Exception as e:
        logger.error(f"Failed to get recommendations for form {form_id}: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to get recommendations: {str(e)}"
        )

@router.post(
    "/{form_id}/feedback",
    status_code=status.HTTP_201_CREATED,
    summary="Submit Insight Feedback",
    description="""
    **Provide feedback on the quality and usefulness of generated insights.**
    
    This feedback helps improve the AI model accuracy and relevance.
    
    ## Feedback Types
    
    - **Accuracy**: Was the insight accurate and correct?
    - **Usefulness**: Was the insight actionable and helpful?
    - **Relevance**: Was the insight relevant to your needs?
    - **Implementation**: Did you implement the recommendation?
    
    ## Feedback Scale
    
    - **5**: Excellent - Very accurate and extremely useful
    - **4**: Good - Mostly accurate and quite useful
    - **3**: Average - Somewhat accurate and moderately useful
    - **2**: Poor - Limited accuracy or usefulness
    - **1**: Very Poor - Inaccurate or not useful
    """,
    operation_id="submitInsightFeedback",
    dependencies=[Depends(security)]
)
async def submit_insight_feedback(
    form_id: str = Path(..., description="Unique identifier for the form"),
    insight_id: str = Query(..., description="Unique identifier for the insight"),
    rating: int = Query(..., description="Feedback rating (1-5)", ge=1, le=5),
    feedback_type: str = Query(..., description="Type of feedback"),
    comments: Optional[str] = Query(None, description="Additional comments"),
    current_user: dict = Depends(get_current_user)
):
    """Submit feedback on insight quality."""
    try:
        await analytics_service.submit_insight_feedback(
            form_id=form_id,
            insight_id=insight_id,
            rating=rating,
            feedback_type=feedback_type,
            comments=comments,
            user_id=current_user.get('user_id')
        )
        
        return {
            "success": True,
            "message": "Feedback submitted successfully"
        }
        
    except Exception as e:
        logger.error(f"Failed to submit feedback: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to submit feedback: {str(e)}"
        )
