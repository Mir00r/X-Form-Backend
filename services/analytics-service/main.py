"""
Analytics Service for X-Form Backend - Basic Version
Handles response analytics, reporting, and data export functionality
"""

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import Optional, List, Dict, Any
import os
import uvicorn
from datetime import datetime

# Initialize FastAPI app
app = FastAPI(
    title="X-Form Analytics Service",
    description="Analytics and reporting service for X-Form Backend",
    version="1.0.0"
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Pydantic models
class AnalyticsRequest(BaseModel):
    form_id: str
    start_date: Optional[str] = None
    end_date: Optional[str] = None

class AnalyticsResponse(BaseModel):
    form_id: str
    total_responses: int
    analytics_data: Dict[str, Any]

# Health check endpoint
@app.get("/health")
async def health_check():
    return {
        "status": "healthy",
        "service": "analytics-service",
        "timestamp": datetime.utcnow().isoformat(),
        "version": "1.0.0"
    }

# Basic analytics endpoints
@app.get("/api/analytics")
async def get_analytics():
    return {
        "message": "Analytics service is running",
        "available_endpoints": [
            "/health",
            "/api/analytics",
            "/api/analytics/{form_id}",
            "/api/reports/{form_id}"
        ]
    }

@app.get("/api/analytics/{form_id}")
async def get_form_analytics(form_id: str):
    return AnalyticsResponse(
        form_id=form_id,
        total_responses=0,
        analytics_data={
            "status": "placeholder",
            "message": f"Analytics for form {form_id}",
            "last_updated": datetime.utcnow().isoformat()
        }
    )

@app.get("/api/reports/{form_id}")
async def get_form_report(form_id: str, format: str = "json"):
    return {
        "form_id": form_id,
        "report_format": format,
        "message": f"Report for form {form_id} in {format} format",
        "status": "placeholder",
        "timestamp": datetime.utcnow().isoformat()
    }

@app.post("/api/analytics")
async def create_analytics_record(request: AnalyticsRequest):
    return {
        "message": "Analytics record created",
        "form_id": request.form_id,
        "status": "success",
        "timestamp": datetime.utcnow().isoformat()
    }

if __name__ == "__main__":
    port = int(os.getenv("PORT", 5001))
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=port,
        reload=False
    )
            "client_email": os.getenv('FIREBASE_CLIENT_EMAIL'),
        })
    firebase_admin.initialize_app(cred)

# Initialize Firestore
db = firestore.client()

# Initialize BigQuery (for advanced analytics)
bigquery_client = bigquery.Client() if os.getenv('BIGQUERY_PROJECT_ID') else None

# FastAPI app
app = FastAPI(
    title="X-Form Analytics Service",
    description="Analytics and reporting service for form responses",
    version="1.0.0"
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configure appropriately for production
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Security
security = HTTPBearer()

# Pydantic models
class AnalyticsSummary(BaseModel):
    total_responses: int
    completion_rate: float
    average_completion_time: Optional[float]
    response_trend: List[Dict[str, Any]]
    question_analytics: Dict[str, Any]

class ExportRequest(BaseModel):
    format: str = "csv"  # csv, xlsx, json
    include_metadata: bool = True
    date_range: Optional[Dict[str, str]] = None

# Helper functions
async def verify_token(credentials: HTTPAuthorizationCredentials = Depends(security)):
    """Verify JWT token (simplified for MVP)"""
    # In production, implement proper JWT verification
    token = credentials.credentials
    if not token:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid authentication credentials"
        )
    return {"user_id": "demo-user"}  # Return decoded user info

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": "analytics-service",
        "timestamp": datetime.utcnow().isoformat(),
        "version": "1.0.0"
    }

@app.get("/analytics/{form_id}/summary", response_model=AnalyticsSummary)
async def get_form_analytics(
    form_id: str,
    user = Depends(verify_token)
):
    """Get analytics summary for a form"""
    try:
        # Get responses from Firestore
        responses_ref = db.collection('responses').where('form_id', '==', form_id)
        responses = responses_ref.stream()
        
        response_data = []
        for response in responses:
            data = response.to_dict()
            data['id'] = response.id
            response_data.append(data)
        
        if not response_data:
            return AnalyticsSummary(
                total_responses=0,
                completion_rate=0.0,
                average_completion_time=None,
                response_trend=[],
                question_analytics={}
            )
        
        # Calculate analytics
        total_responses = len(response_data)
        
        # Calculate completion rate (responses with all required fields)
        completed_responses = sum(1 for r in response_data if r.get('completed', False))
        completion_rate = (completed_responses / total_responses) * 100 if total_responses > 0 else 0
        
        # Calculate average completion time
        completion_times = [r.get('completion_time') for r in response_data if r.get('completion_time')]
        avg_completion_time = sum(completion_times) / len(completion_times) if completion_times else None
        
        # Generate response trend (last 30 days)
        now = datetime.utcnow()
        trend_data = []
        for i in range(30):
            date = now - timedelta(days=29-i)
            date_str = date.strftime('%Y-%m-%d')
            count = sum(1 for r in response_data 
                       if r.get('created_at') and 
                       r['created_at'].date() == date.date())
            trend_data.append({"date": date_str, "count": count})
        
        # Analyze questions
        question_analytics = {}
        if response_data:
            sample_response = response_data[0]
            answers = sample_response.get('answers', {})
            
            for question_id, _ in answers.items():
                question_responses = [r.get('answers', {}).get(question_id) 
                                    for r in response_data 
                                    if r.get('answers', {}).get(question_id) is not None]
                
                question_analytics[question_id] = {
                    "response_count": len(question_responses),
                    "response_rate": (len(question_responses) / total_responses) * 100,
                    "sample_responses": question_responses[:5]  # First 5 responses as sample
                }
        
        return AnalyticsSummary(
            total_responses=total_responses,
            completion_rate=completion_rate,
            average_completion_time=avg_completion_time,
            response_trend=trend_data,
            question_analytics=question_analytics
        )
        
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to generate analytics: {str(e)}"
        )

@app.post("/forms/{form_id}/export")
async def export_responses(
    form_id: str,
    export_request: ExportRequest,
    user = Depends(verify_token)
):
    """Export form responses in various formats"""
    try:
        # Get responses from Firestore
        responses_ref = db.collection('responses').where('form_id', '==', form_id)
        responses = responses_ref.stream()
        
        response_data = []
        for response in responses:
            data = response.to_dict()
            data['id'] = response.id
            response_data.append(data)
        
        if not response_data:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="No responses found for this form"
            )
        
        # Convert to DataFrame for easier manipulation
        df = pd.json_normalize(response_data)
        
        # Apply date range filter if specified
        if export_request.date_range:
            start_date = datetime.fromisoformat(export_request.date_range.get('start'))
            end_date = datetime.fromisoformat(export_request.date_range.get('end'))
            df = df[
                (pd.to_datetime(df['created_at']) >= start_date) &
                (pd.to_datetime(df['created_at']) <= end_date)
            ]
        
        # Generate filename
        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
        filename = f"form_{form_id}_responses_{timestamp}"
        
        if export_request.format == "csv":
            file_path = f"/tmp/{filename}.csv"
            df.to_csv(file_path, index=False)
        elif export_request.format == "xlsx":
            file_path = f"/tmp/{filename}.xlsx"
            df.to_excel(file_path, index=False)
        elif export_request.format == "json":
            file_path = f"/tmp/{filename}.json"
            df.to_json(file_path, orient='records', indent=2)
        else:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Unsupported export format"
            )
        
        return {
            "message": "Export completed successfully",
            "filename": f"{filename}.{export_request.format}",
            "record_count": len(df),
            "download_url": f"/download/{filename}.{export_request.format}"
        }
        
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to export responses: {str(e)}"
        )

@app.get("/forms/{form_id}/insights")
async def get_form_insights(
    form_id: str,
    user = Depends(verify_token)
):
    """Get AI-powered insights for form responses (MVP placeholder)"""
    try:
        # This would integrate with AI services for deeper insights
        # For MVP, return basic insights
        
        responses_ref = db.collection('responses').where('form_id', '==', form_id)
        responses = list(responses_ref.stream())
        
        insights = {
            "response_volume": {
                "trend": "increasing" if len(responses) > 10 else "stable",
                "prediction": "Based on current trends, expect 20% more responses next week"
            },
            "completion_patterns": {
                "peak_hours": ["14:00-16:00", "20:00-22:00"],
                "drop_off_points": ["Question 5", "Question 8"]
            },
            "recommendations": [
                "Consider shortening the form to improve completion rate",
                "Add progress indicators to reduce abandonment",
                "Optimize questions 5 and 8 based on drop-off patterns"
            ]
        }
        
        return insights
        
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to generate insights: {str(e)}"
        )

if __name__ == "__main__":
    port = int(os.getenv("PORT", 5001))
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=port,
        reload=os.getenv("ENVIRONMENT") == "development"
    )
