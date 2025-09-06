"""
Analytics Service Data Models with Comprehensive Swagger Documentation
"""
from typing import Any, Dict, List, Optional, Union
from datetime import datetime, date
from pydantic import BaseModel, Field, validator
from enum import Enum


class ResponseStatus(str, Enum):
    """Response submission status."""
    COMPLETED = "completed"
    PARTIAL = "partial"
    DRAFT = "draft"


class QuestionType(str, Enum):
    """Question types for analytics."""
    TEXT = "text"
    EMAIL = "email"
    NUMBER = "number"
    TEXTAREA = "textarea"
    SELECT = "select"
    MULTISELECT = "multiselect"
    RADIO = "radio"
    CHECKBOX = "checkbox"
    DATE = "date"
    TIME = "time"
    DATETIME = "datetime"
    FILE = "file"
    URL = "url"
    RATING = "rating"
    SCALE = "scale"
    MATRIX = "matrix"


class PeriodType(str, Enum):
    """Time period types for trend analysis."""
    HOUR = "1h"
    DAY = "1d"
    WEEK = "7d"
    MONTH = "30d"
    QUARTER = "90d"
    YEAR = "365d"


class ChartType(str, Enum):
    """Chart types for analytics visualization."""
    BAR = "bar"
    PIE = "pie"
    LINE = "line"
    HISTOGRAM = "histogram"
    SCATTER = "scatter"
    HEATMAP = "heatmap"
    TABLE = "table"


# Request Models
class AnalyticsQueryParams(BaseModel):
    """Base analytics query parameters."""
    start_date: Optional[datetime] = None
    end_date: Optional[datetime] = None
    user_id: Optional[str] = None
    status: Optional[ResponseStatus] = None


class TrendQueryParams(AnalyticsQueryParams):
    """Trend analysis query parameters."""
    period: PeriodType = Field(default=PeriodType.DAY)
    group_by: Optional[str] = None


class QuestionAnalyticsParams(AnalyticsQueryParams):
    """Question-specific analytics parameters."""
    chart_type: Optional[ChartType] = None
    include_metadata: bool = False
    limit: int = Field(default=100, le=1000)


# Response Models
class AnalyticsResponse(BaseModel):
    """Standard analytics API response model."""
    success: bool = Field(description="Whether the request was successful")
    message: str = Field(description="Human-readable response message")
    data: Optional[Dict[str, Any]] = Field(description="Response data payload")
    timestamp: datetime = Field(description="Response timestamp in UTC")
    request_id: Optional[str] = Field(description="Unique request identifier for tracking")
    
    class Config:
        schema_extra = {
            "example": {
                "success": True,
                "message": "Analytics data retrieved successfully",
                "data": {
                    "form_id": "550e8400-e29b-41d4-a716-446655440000",
                    "total_responses": 1523,
                    "completion_rate": 84.6
                },
                "timestamp": "2025-09-06T12:00:00Z",
                "request_id": "req_abc123def456"
            }
        }

class ErrorResponse(BaseModel):
    """Standard error response model."""
    success: bool = Field(default=False, description="Always false for error responses")
    error: str = Field(description="Error type or code")
    message: str = Field(description="Human-readable error message")
    details: Optional[Dict[str, Any]] = Field(description="Additional error details")
    timestamp: datetime = Field(description="Error timestamp in UTC")
    request_id: Optional[str] = Field(description="Request identifier for troubleshooting")
    
    class Config:
        schema_extra = {
            "example": {
                "success": False,
                "error": "validation_error",
                "message": "Invalid form ID format",
                "details": {
                    "field": "form_id",
                    "expected": "UUID format",
                    "received": "invalid_string"
                },
                "timestamp": "2025-09-06T12:00:00Z",
                "request_id": "req_abc123def456"
            }
        }

class FormSummary(BaseModel):
    """Form analytics summary with comprehensive metrics."""
    form_id: str = Field(description="Unique form identifier")
    title: str = Field(description="Form title")
    total_responses: int = Field(description="Total number of responses received", ge=0)
    completed_responses: int = Field(description="Number of completed responses", ge=0)
    partial_responses: int = Field(description="Number of partial responses", ge=0)
    average_completion_time: Optional[float] = Field(description="Average completion time in seconds", ge=0)
    completion_rate: float = Field(description="Completion rate as percentage", ge=0, le=100)
    first_response_date: Optional[datetime] = Field(description="Date of first response")
    last_response_date: Optional[datetime] = Field(description="Date of most recent response")
    unique_respondents: int = Field(description="Number of unique respondents", ge=0)
    response_rate_trend: List[Dict[str, Any]] = Field(description="Response rate trend over last 7 days")
    
    class Config:
        schema_extra = {
            "example": {
                "form_id": "550e8400-e29b-41d4-a716-446655440000",
                "title": "Customer Satisfaction Survey",
                "total_responses": 1523,
                "completed_responses": 1289,
                "partial_responses": 234,
                "average_completion_time": 145.7,
                "completion_rate": 84.6,
                "first_response_date": "2025-08-01T09:30:00Z",
                "last_response_date": "2025-09-06T11:45:00Z",
                "unique_respondents": 1456,
                "response_rate_trend": [
                    {"date": "2025-09-05", "responses": 45, "completion_rate": 86.7},
                    {"date": "2025-09-06", "responses": 52, "completion_rate": 84.6}
                ]
            }
        }


class QuestionSummary(BaseModel):
    """Question analytics summary."""
    question_id: str
    question_type: QuestionType
    question_text: str
    total_responses: int
    response_rate: float  # percentage of form responses that answered this question
    skip_rate: float  # percentage that skipped
    average_response_time: Optional[float]  # in seconds


class ResponseDistribution(BaseModel):
    """Response value distribution for a question."""
    value: Union[str, int, float, bool]
    count: int
    percentage: float
    label: Optional[str] = None


class ChartData(BaseModel):
    """Chart data for visualization."""
    chart_type: ChartType
    title: str
    labels: List[str]
    datasets: List[Dict[str, Any]]
    options: Dict[str, Any] = {}


class QuestionAnalytics(BaseModel):
    """Detailed analytics for a specific question."""
    question_summary: QuestionSummary
    distribution: List[ResponseDistribution]
    chart_data: ChartData
    statistics: Dict[str, Any]  # mean, median, mode, std_dev, etc.
    correlations: List[Dict[str, Any]] = []  # correlations with other questions
    trends: List[Dict[str, Any]] = []  # time-based trends


class TrendDataPoint(BaseModel):
    """Single data point in a trend."""
    timestamp: datetime
    value: Union[int, float]
    label: Optional[str] = None
    metadata: Dict[str, Any] = {}


class TrendAnalysis(BaseModel):
    """Trend analysis result."""
    form_id: str
    period: PeriodType
    total_period_responses: int
    trend_data: List[TrendDataPoint]
    statistics: Dict[str, Any]  # growth_rate, peak_day, etc.
    forecasting: Optional[Dict[str, Any]] = None


class ResponseMetadata(BaseModel):
    """Response metadata for analytics."""
    user_agent: Optional[str] = None
    ip_address: Optional[str] = None
    referrer: Optional[str] = None
    location: Optional[Dict[str, str]] = None  # country, city, etc.
    device_type: Optional[str] = None
    browser: Optional[str] = None


class FormResponse(BaseModel):
    """Form response data model."""
    response_id: str
    form_id: str
    user_id: Optional[str] = None
    submitted_at: datetime
    completion_time_seconds: Optional[int] = None
    status: ResponseStatus
    responses: Dict[str, Any]  # question_id -> answer
    metadata: Optional[ResponseMetadata] = None


class FormMetadata(BaseModel):
    """Form metadata for analytics."""
    form_id: str
    title: str
    description: Optional[str] = None
    owner_id: str
    created_at: datetime
    updated_at: datetime
    status: str
    questions: List[Dict[str, Any]]
    settings: Dict[str, Any] = {}


# Analytics Result Models
class AnalyticsResult(BaseModel):
    """Base analytics result."""
    query_id: str = Field(default_factory=lambda: f"query_{datetime.now().timestamp()}")
    executed_at: datetime = Field(default_factory=datetime.now)
    execution_time_ms: Optional[int] = None
    cached: bool = False


class FormAnalyticsResult(AnalyticsResult):
    """Complete form analytics result."""
    form_summary: FormSummary
    question_analytics: List[QuestionAnalytics]
    response_trends: TrendAnalysis
    demographic_insights: Dict[str, Any] = {}
    recommendations: List[str] = []


class QuestionAnalyticsResult(AnalyticsResult):
    """Question-specific analytics result."""
    question_analytics: QuestionAnalytics


class TrendAnalyticsResult(AnalyticsResult):
    """Trend analysis result."""
    trend_analysis: TrendAnalysis


# Database Models
class BigQueryResponse(BaseModel):
    """BigQuery response model."""
    response_id: str
    form_id: str
    user_id: Optional[str]
    submitted_at: datetime
    completion_time_seconds: Optional[int]
    responses: str  # JSON string
    metadata: Optional[str] = None  # JSON string
    ip_address: Optional[str] = None
    user_agent: Optional[str] = None
    created_at: datetime
    updated_at: datetime


class BigQueryForm(BaseModel):
    """BigQuery form model."""
    form_id: str
    title: str
    description: Optional[str]
    owner_id: str
    questions: str  # JSON string
    settings: Optional[str] = None  # JSON string
    status: str
    created_at: datetime
    updated_at: datetime


class BigQueryEvent(BaseModel):
    """BigQuery event model."""
    event_id: str
    event_type: str
    form_id: str
    user_id: Optional[str]
    event_data: str  # JSON string
    timestamp: datetime
    session_id: Optional[str] = None
    ip_address: Optional[str] = None


# Error Models
class AnalyticsError(BaseModel):
    """Analytics error response."""
    error_code: str
    message: str
    details: Optional[Dict[str, Any]] = None
    timestamp: datetime = Field(default_factory=datetime.now)


class ValidationError(BaseModel):
    """Validation error response."""
    field: str
    message: str
    value: Optional[Any] = None


# Cache Models
class CacheKey(BaseModel):
    """Cache key structure."""
    service: str = "analytics"
    resource: str
    identifier: str
    params_hash: Optional[str] = None
    
    def to_string(self) -> str:
        """Convert to cache key string."""
        parts = [self.service, self.resource, self.identifier]
        if self.params_hash:
            parts.append(self.params_hash)
        return ":".join(parts)


class CachedResult(BaseModel):
    """Cached analytics result."""
    key: str
    data: Dict[str, Any]
    created_at: datetime
    expires_at: datetime
    hit_count: int = 0


# Export all models
__all__ = [
    "ResponseStatus",
    "QuestionType", 
    "PeriodType",
    "ChartType",
    "AnalyticsQueryParams",
    "TrendQueryParams",
    "QuestionAnalyticsParams",
    "FormSummary",
    "QuestionSummary",
    "ResponseDistribution",
    "ChartData",
    "QuestionAnalytics",
    "TrendDataPoint",
    "TrendAnalysis",
    "ResponseMetadata",
    "FormResponse",
    "FormMetadata",
    "AnalyticsResult",
    "FormAnalyticsResult",
    "QuestionAnalyticsResult",
    "TrendAnalyticsResult",
    "BigQueryResponse",
    "BigQueryForm",
    "BigQueryEvent",
    "AnalyticsError",
    "ValidationError",
    "CacheKey",
    "CachedResult"
]
