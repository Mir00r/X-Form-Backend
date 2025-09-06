"""
Reports API Routes with Comprehensive Swagger Documentation
"""
import logging
from datetime import datetime
from typing import Optional, List
from fastapi import APIRouter, HTTPException, Query, Depends, status, Path, BackgroundTasks
from fastapi.responses import FileResponse, StreamingResponse
from fastapi.security import HTTPBearer
from pydantic import BaseModel

from app.models.analytics import AnalyticsResponse, ErrorResponse
from app.services.analytics_service import analytics_service
from app.utils.auth import get_current_user
from app.utils.rate_limiter import rate_limit

logger = logging.getLogger(__name__)
security = HTTPBearer()

router = APIRouter(prefix="/reports", tags=["reports"])

# Request Models
class ExportRequest(BaseModel):
    """Request model for data export operations."""
    format: str = "csv"
    include_metadata: bool = True
    date_range: Optional[dict] = None
    filters: Optional[dict] = None
    columns: Optional[List[str]] = None
    
    class Config:
        schema_extra = {
            "example": {
                "format": "xlsx",
                "include_metadata": True,
                "date_range": {
                    "start": "2025-09-01T00:00:00Z",
                    "end": "2025-09-06T23:59:59Z"
                },
                "filters": {
                    "completed_only": True,
                    "min_completion_time": 30
                },
                "columns": ["response_id", "submitted_at", "completion_time", "answers"]
            }
        }

class ReportRequest(BaseModel):
    """Request model for custom report generation."""
    report_type: str
    parameters: dict
    delivery_method: str = "download"
    schedule: Optional[str] = None
    
    class Config:
        schema_extra = {
            "example": {
                "report_type": "comprehensive_analytics",
                "parameters": {
                    "include_charts": True,
                    "include_raw_data": False,
                    "group_by": "date"
                },
                "delivery_method": "email",
                "schedule": "weekly"
            }
        }

@router.post(
    "/{form_id}/export",
    status_code=status.HTTP_200_OK,
    summary="Export Form Data",
    description="""
    **Export form response data in various formats.**
    
    This endpoint allows exporting form data in multiple formats:
    - **CSV**: Comma-separated values for spreadsheet applications
    - **Excel (XLSX)**: Microsoft Excel format with formatting
    - **JSON**: JavaScript Object Notation for API integration
    - **PDF**: Formatted report for presentation
    
    ## Export Features
    
    - **Flexible Formatting**: Multiple output formats supported
    - **Custom Date Ranges**: Export data for specific time periods
    - **Advanced Filtering**: Filter by completion status, response time, etc.
    - **Column Selection**: Choose specific fields to include
    - **Metadata Options**: Include or exclude response metadata
    - **Large Dataset Support**: Streaming for large exports
    
    ## Parameters
    
    - **form_id**: Unique identifier for the form
    - **export_request**: Export configuration and filters
    
    ## Rate Limiting
    
    This endpoint is rate limited to 10 requests per hour per user due to resource intensity.
    
    ## File Size Limits
    
    - **CSV/JSON**: Up to 100MB
    - **Excel**: Up to 50MB  
    - **PDF**: Up to 25MB
    
    For larger datasets, consider using date range filtering or the streaming export endpoint.
    """,
    responses={
        200: {
            "description": "Export completed successfully",
            "content": {
                "application/json": {
                    "example": {
                        "success": True,
                        "message": "Export completed successfully",
                        "data": {
                            "export_id": "exp_550e8400e29b41d4a716446655440000",
                            "filename": "form_responses_20250906_120000.xlsx",
                            "format": "xlsx",
                            "size_bytes": 2458720,
                            "record_count": 1289,
                            "download_url": "/reports/download/exp_550e8400e29b41d4a716446655440000",
                            "expires_at": "2025-09-07T12:00:00Z"
                        },
                        "timestamp": "2025-09-06T12:00:00Z"
                    }
                }
            }
        },
        400: {"description": "Invalid export parameters"},
        401: {"description": "Authentication required"},
        403: {"description": "Access forbidden"},
        404: {"description": "Form not found"},
        413: {"description": "Export too large"},
        429: {"description": "Rate limit exceeded"},
        500: {"description": "Export failed"}
    },
    operation_id="exportFormData",
    dependencies=[Depends(security)]
)
@rate_limit(max_requests=10, window_seconds=3600)  # 10 requests per hour
async def export_form_data(
    background_tasks: BackgroundTasks,
    form_id: str = Path(
        ..., 
        description="Unique identifier for the form",
        example="550e8400-e29b-41d4-a716-446655440000"
    ),
    export_request: ExportRequest = None,
    current_user: dict = Depends(get_current_user)
):
    """Export form response data in the specified format."""
    try:
        logger.info(f"Starting export for form {form_id} by user {current_user.get('user_id')}")
        
        # Validate export format
        supported_formats = ["csv", "xlsx", "json", "pdf"]
        if export_request.format not in supported_formats:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Unsupported export format. Supported formats: {supported_formats}"
            )
        
        # Start background export task
        export_id = await analytics_service.start_export_task(
            form_id=form_id,
            export_config=export_request,
            user_id=current_user.get('user_id')
        )
        
        return {
            "success": True,
            "message": "Export started successfully",
            "data": {
                "export_id": export_id,
                "status": "processing",
                "estimated_completion": "2-5 minutes"
            }
        }
        
    except Exception as e:
        logger.error(f"Export failed for form {form_id}: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Export failed: {str(e)}"
        )

@router.get(
    "/{form_id}/export/{export_id}/status",
    status_code=status.HTTP_200_OK,
    summary="Get Export Status",
    description="""
    **Check the status of an ongoing export operation.**
    
    Use this endpoint to monitor the progress of export tasks.
    
    ## Export Statuses
    
    - **pending**: Export request received, waiting to start
    - **processing**: Export is currently being generated
    - **completed**: Export completed successfully, ready for download
    - **failed**: Export failed due to an error
    - **expired**: Export file has expired and been deleted
    
    ## Parameters
    
    - **form_id**: Unique identifier for the form
    - **export_id**: Unique identifier for the export task
    """,
    responses={
        200: {
            "description": "Export status retrieved successfully",
            "content": {
                "application/json": {
                    "examples": {
                        "processing": {
                            "summary": "Export in progress",
                            "value": {
                                "success": True,
                                "data": {
                                    "export_id": "exp_550e8400e29b41d4a716446655440000",
                                    "status": "processing",
                                    "progress": 65,
                                    "estimated_completion": "2025-09-06T12:05:00Z",
                                    "records_processed": 837
                                }
                            }
                        },
                        "completed": {
                            "summary": "Export completed",
                            "value": {
                                "success": True,
                                "data": {
                                    "export_id": "exp_550e8400e29b41d4a716446655440000",
                                    "status": "completed",
                                    "progress": 100,
                                    "filename": "form_responses_20250906_120000.xlsx",
                                    "size_bytes": 2458720,
                                    "download_url": "/reports/download/exp_550e8400e29b41d4a716446655440000",
                                    "expires_at": "2025-09-07T12:00:00Z"
                                }
                            }
                        }
                    }
                }
            }
        },
        404: {"description": "Export not found"},
        401: {"description": "Authentication required"},
        500: {"description": "Server error"}
    },
    operation_id="getExportStatus"
)
async def get_export_status(
    form_id: str = Path(..., description="Unique identifier for the form"),
    export_id: str = Path(..., description="Unique identifier for the export task"),
    current_user: dict = Depends(get_current_user)
):
    """Get the status of an export operation."""
    try:
        status_info = await analytics_service.get_export_status(export_id, current_user.get('user_id'))
        return {
            "success": True,
            "data": status_info
        }
    except Exception as e:
        logger.error(f"Failed to get export status {export_id}: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to get export status: {str(e)}"
        )

@router.get(
    "/download/{export_id}",
    summary="Download Export File",
    description="""
    **Download a completed export file.**
    
    This endpoint serves the generated export file for download.
    
    ## File Types
    
    - **CSV**: text/csv content type
    - **Excel**: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
    - **JSON**: application/json
    - **PDF**: application/pdf
    
    ## Security
    
    - Files are automatically deleted after 24 hours
    - Only the user who initiated the export can download
    - Secure temporary URLs with limited access
    """,
    responses={
        200: {
            "description": "File download",
            "content": {
                "application/octet-stream": {}
            }
        },
        404: {"description": "File not found or expired"},
        401: {"description": "Authentication required"},
        403: {"description": "Access forbidden"}
    },
    operation_id="downloadExportFile"
)
async def download_export_file(
    export_id: str = Path(..., description="Unique identifier for the export"),
    current_user: dict = Depends(get_current_user)
):
    """Download an export file."""
    try:
        file_path = await analytics_service.get_export_file_path(export_id, current_user.get('user_id'))
        if not file_path:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Export file not found or expired"
            )
        
        return FileResponse(
            path=file_path,
            filename=f"export_{export_id}",
            media_type="application/octet-stream"
        )
        
    except Exception as e:
        logger.error(f"Download failed for export {export_id}: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Download failed: {str(e)}"
        )

@router.post(
    "/{form_id}/generate",
    status_code=status.HTTP_201_CREATED,
    summary="Generate Custom Report",
    description="""
    **Generate a custom analytics report with specific parameters.**
    
    This endpoint creates comprehensive reports with custom formatting and content.
    
    ## Report Types
    
    - **summary**: Executive summary with key metrics
    - **detailed**: Comprehensive analysis with all data points
    - **comparison**: Compare multiple time periods or segments
    - **trend**: Focus on trends and patterns over time
    - **custom**: User-defined report with specific parameters
    
    ## Delivery Methods
    
    - **download**: Generate file for immediate download
    - **email**: Send report via email
    - **webhook**: POST report to specified URL
    - **scheduled**: Set up recurring report generation
    
    ## Parameters
    
    - **form_id**: Unique identifier for the form
    - **report_request**: Report configuration and parameters
    """,
    responses={
        201: {
            "description": "Report generation started",
            "content": {
                "application/json": {
                    "example": {
                        "success": True,
                        "message": "Report generation started",
                        "data": {
                            "report_id": "rpt_550e8400e29b41d4a716446655440000",
                            "status": "generating",
                            "estimated_completion": "2025-09-06T12:10:00Z",
                            "report_type": "comprehensive_analytics"
                        }
                    }
                }
            }
        },
        400: {"description": "Invalid report parameters"},
        401: {"description": "Authentication required"},
        404: {"description": "Form not found"},
        429: {"description": "Rate limit exceeded"},
        500: {"description": "Report generation failed"}
    },
    operation_id="generateCustomReport",
    dependencies=[Depends(security)]
)
@rate_limit(max_requests=5, window_seconds=3600)  # 5 reports per hour
async def generate_custom_report(
    background_tasks: BackgroundTasks,
    form_id: str = Path(..., description="Unique identifier for the form"),
    report_request: ReportRequest = None,
    current_user: dict = Depends(get_current_user)
):
    """Generate a custom analytics report."""
    try:
        logger.info(f"Generating custom report for form {form_id} by user {current_user.get('user_id')}")
        
        # Start background report generation
        report_id = await analytics_service.start_report_generation(
            form_id=form_id,
            report_config=report_request,
            user_id=current_user.get('user_id')
        )
        
        return {
            "success": True,
            "message": "Report generation started",
            "data": {
                "report_id": report_id,
                "status": "generating",
                "report_type": report_request.report_type
            }
        }
        
    except Exception as e:
        logger.error(f"Report generation failed for form {form_id}: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Report generation failed: {str(e)}"
        )

@router.get(
    "/templates",
    summary="List Report Templates",
    description="""
    **Get a list of available report templates.**
    
    Report templates provide predefined configurations for common analytics reports.
    
    ## Template Categories
    
    - **Executive**: High-level summaries for leadership
    - **Operational**: Detailed metrics for operations teams  
    - **Marketing**: Campaign and engagement analytics
    - **Research**: Academic and research-focused reports
    - **Custom**: User-created template library
    """,
    responses={
        200: {
            "description": "Report templates retrieved successfully",
            "content": {
                "application/json": {
                    "example": {
                        "success": True,
                        "data": {
                            "templates": [
                                {
                                    "id": "executive_summary",
                                    "name": "Executive Summary",
                                    "description": "High-level overview with key metrics",
                                    "category": "executive",
                                    "fields": ["completion_rate", "response_count", "trends"]
                                },
                                {
                                    "id": "detailed_analytics",
                                    "name": "Detailed Analytics Report",
                                    "description": "Comprehensive analysis with all metrics",
                                    "category": "operational",
                                    "fields": ["all"]
                                }
                            ]
                        }
                    }
                }
            }
        },
        401: {"description": "Authentication required"},
        500: {"description": "Server error"}
    },
    operation_id="listReportTemplates"
)
async def list_report_templates(
    category: Optional[str] = Query(None, description="Filter templates by category"),
    current_user: dict = Depends(get_current_user)
):
    """Get a list of available report templates."""
    try:
        templates = await analytics_service.get_report_templates(category)
        return {
            "success": True,
            "data": {"templates": templates}
        }
    except Exception as e:
        logger.error(f"Failed to get report templates: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to get report templates: {str(e)}"
        )
