"""
Analytics Service - Main service layer
"""
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional
import asyncio

from app.config import settings
from app.models.analytics import (
    FormSummary, QuestionAnalytics, TrendAnalysis, PeriodType,
    AnalyticsResponse, ChartData, ErrorResponse
)
from app.services.bigquery_service import BigQueryService
from app.services.cache_service import CacheService
from app.services.chart_service import ChartService

logger = logging.getLogger(__name__)


class AnalyticsService:
    """Main analytics service that orchestrates all analytics operations."""
    
    def __init__(self):
        self.bigquery_service = BigQueryService()
        self.cache_service = CacheService()
        self.chart_service = ChartService()
    
    async def initialize(self) -> None:
        """Initialize the analytics service."""
        try:
            await self.bigquery_service.initialize_tables()
            logger.info("Analytics service initialized successfully")
        except Exception as e:
            logger.error(f"Failed to initialize analytics service: {e}")
            raise
    
    async def get_form_analytics_summary(
        self,
        form_id: str,
        start_date: Optional[datetime] = None,
        end_date: Optional[datetime] = None,
        use_cache: bool = True
    ) -> Dict[str, Any]:
        """Get comprehensive form analytics summary with charts."""
        try:
            # Check cache first
            if use_cache:
                cached_data = await self.cache_service.get_form_summary(
                    form_id, start_date, end_date
                )
                if cached_data:
                    logger.info(f"Form summary cache hit for {form_id}")
                    return cached_data
            
            # Get data from BigQuery
            summary = await self.bigquery_service.get_form_summary(
                form_id, start_date, end_date
            )
            
            # Create charts
            charts = {}
            
            # Completion rate chart
            if summary.total_responses > 0:
                charts["completion_rate"] = self.chart_service.create_completion_rate_chart({
                    "total_responses": summary.total_responses,
                    "completed_responses": summary.completed_responses,
                    "partial_responses": summary.partial_responses
                })
            
            # Response trend chart
            if summary.response_rate_trend:
                trend_data = [
                    {
                        "timestamp": item["date"],
                        "value": item["count"]
                    }
                    for item in summary.response_rate_trend
                ]
                charts["trend"] = self.chart_service.create_trend_chart(
                    trend_data, PeriodType.DAY
                )
            
            # Prepare response
            response_data = {
                "form_id": form_id,
                "summary": summary.dict(),
                "charts": {key: chart.dict() for key, chart in charts.items()},
                "generated_at": datetime.now().isoformat(),
                "cache_info": {
                    "cached": False,
                    "ttl": settings.cache_ttl
                }
            }
            
            # Cache the result
            if use_cache:
                await self.cache_service.set_form_summary(
                    form_id, response_data, start_date, end_date
                )
            
            return response_data
            
        except Exception as e:
            logger.error(f"Error getting form analytics summary for {form_id}: {e}")
            raise
    
    async def get_question_analytics(
        self,
        form_id: str,
        question_id: str,
        start_date: Optional[datetime] = None,
        end_date: Optional[datetime] = None,
        question_type: str = "multiple_choice",
        use_cache: bool = True
    ) -> Dict[str, Any]:
        """Get analytics for a specific question with visualizations."""
        try:
            # Check cache first
            if use_cache:
                cached_data = await self.cache_service.get_question_analytics(
                    form_id, question_id, start_date, end_date
                )
                if cached_data:
                    logger.info(f"Question analytics cache hit for {form_id}/{question_id}")
                    return cached_data
            
            # Get data from BigQuery
            analytics_data = await self.bigquery_service.get_question_analytics(
                form_id, question_id, start_date, end_date
            )
            
            # Create charts
            charts = {}
            
            # Response distribution chart
            if analytics_data["distribution"]:
                charts["distribution"] = self.chart_service.create_response_distribution_chart(
                    analytics_data["distribution"], question_type
                )
            
            # Response rate visualization
            if analytics_data["total_responses"] > 0:
                rate_data = [
                    {"label": "Answered", "value": analytics_data["answered_responses"]},
                    {"label": "Skipped", "value": analytics_data["total_responses"] - analytics_data["answered_responses"]}
                ]
                charts["response_rate"] = self.chart_service.create_pie_chart(
                    rate_data, "Response Rate", "label", "value"
                )
            
            # Prepare response
            response_data = {
                "form_id": form_id,
                "question_id": question_id,
                "analytics": analytics_data,
                "charts": {key: chart.dict() for key, chart in charts.items()},
                "generated_at": datetime.now().isoformat(),
                "cache_info": {
                    "cached": False,
                    "ttl": settings.cache_ttl
                }
            }
            
            # Cache the result
            if use_cache:
                await self.cache_service.set_question_analytics(
                    form_id, question_id, response_data, start_date, end_date
                )
            
            return response_data
            
        except Exception as e:
            logger.error(f"Error getting question analytics for {form_id}/{question_id}: {e}")
            raise
    
    async def get_trend_analysis(
        self,
        form_id: str,
        period: PeriodType = PeriodType.DAY,
        start_date: Optional[datetime] = None,
        end_date: Optional[datetime] = None,
        use_cache: bool = True
    ) -> Dict[str, Any]:
        """Get trend analysis with visualizations."""
        try:
            # Check cache first
            if use_cache:
                cached_data = await self.cache_service.get_trend_analysis(
                    form_id, period.value, start_date, end_date
                )
                if cached_data:
                    logger.info(f"Trend analysis cache hit for {form_id}")
                    return cached_data
            
            # Get data from BigQuery
            trend_data = await self.bigquery_service.get_trend_analysis(
                form_id, period, start_date, end_date
            )
            
            # Create charts
            charts = {}
            
            # Main trend chart
            if trend_data["trend_data"]:
                charts["trend"] = self.chart_service.create_trend_chart(
                    trend_data["trend_data"], period
                )
                
                # Completion trend if available
                completion_trend = [
                    {
                        "timestamp": item["timestamp"],
                        "value": item["completed_count"]
                    }
                    for item in trend_data["trend_data"]
                    if item.get("completed_count") is not None
                ]
                
                if completion_trend:
                    charts["completion_trend"] = self.chart_service.create_line_chart(
                        completion_trend,
                        f"Completion Trend ({period.value.title()})",
                        "timestamp",
                        "value"
                    )
            
            # Statistics visualization
            stats = trend_data.get("statistics", {})
            if stats:
                stats_data = [
                    {"label": "Total Responses", "value": stats.get("total_responses", 0)},
                    {"label": "Peak Responses", "value": stats.get("peak_responses", 0)},
                    {"label": "Avg per Period", "value": int(stats.get("avg_responses_per_period", 0))}
                ]
                charts["statistics"] = self.chart_service.create_bar_chart(
                    stats_data, "Statistics Overview", "label", "value"
                )
            
            # Prepare response
            response_data = {
                "form_id": form_id,
                "period": period.value,
                "trend_analysis": trend_data,
                "charts": {key: chart.dict() for key, chart in charts.items()},
                "generated_at": datetime.now().isoformat(),
                "cache_info": {
                    "cached": False,
                    "ttl": settings.cache_ttl
                }
            }
            
            # Cache the result
            if use_cache:
                await self.cache_service.set_trend_analysis(
                    form_id, period.value, response_data, start_date, end_date
                )
            
            return response_data
            
        except Exception as e:
            logger.error(f"Error getting trend analysis for {form_id}: {e}")
            raise
    
    async def get_comparative_analytics(
        self,
        form_ids: List[str],
        metric: str = "response_count",
        period: PeriodType = PeriodType.DAY,
        start_date: Optional[datetime] = None,
        end_date: Optional[datetime] = None
    ) -> Dict[str, Any]:
        """Get comparative analytics across multiple forms."""
        try:
            # Get data for all forms concurrently
            tasks = []
            for form_id in form_ids:
                if metric == "response_count":
                    task = self.bigquery_service.get_trend_analysis(
                        form_id, period, start_date, end_date
                    )
                else:
                    task = self.bigquery_service.get_form_summary(
                        form_id, start_date, end_date
                    )
                tasks.append(task)
            
            results = await asyncio.gather(*tasks, return_exceptions=True)
            
            # Process results
            comparative_data = {}
            chart_data = {}
            
            for i, result in enumerate(results):
                form_id = form_ids[i]
                if isinstance(result, Exception):
                    logger.warning(f"Error getting data for form {form_id}: {result}")
                    continue
                
                if metric == "response_count" and "trend_data" in result:
                    chart_data[form_id] = result["trend_data"]
                    comparative_data[form_id] = {
                        "total_responses": result.get("total_period_responses", 0),
                        "statistics": result.get("statistics", {})
                    }
                elif hasattr(result, 'total_responses'):
                    comparative_data[form_id] = {
                        "total_responses": result.total_responses,
                        "completion_rate": result.completion_rate,
                        "average_completion_time": result.average_completion_time
                    }
            
            # Create comparative chart
            charts = {}
            if chart_data:
                charts["comparison"] = self.chart_service.create_multi_metric_chart(
                    chart_data, f"Form Comparison - {metric.replace('_', ' ').title()}"
                )
            
            return {
                "forms": form_ids,
                "metric": metric,
                "period": period.value,
                "comparative_data": comparative_data,
                "charts": {key: chart.dict() for key, chart in charts.items()},
                "generated_at": datetime.now().isoformat()
            }
            
        except Exception as e:
            logger.error(f"Error getting comparative analytics: {e}")
            raise
    
    async def invalidate_cache(self, form_id: str, question_id: Optional[str] = None) -> Dict[str, Any]:
        """Invalidate cache for a form or specific question."""
        try:
            if question_id:
                deleted_count = await self.cache_service.invalidate_question_cache(form_id, question_id)
                return {
                    "action": "cache_invalidation",
                    "scope": "question",
                    "form_id": form_id,
                    "question_id": question_id,
                    "deleted_keys": deleted_count
                }
            else:
                deleted_count = await self.cache_service.invalidate_form_cache(form_id)
                return {
                    "action": "cache_invalidation",
                    "scope": "form",
                    "form_id": form_id,
                    "deleted_keys": deleted_count
                }
                
        except Exception as e:
            logger.error(f"Error invalidating cache: {e}")
            raise
    
    async def get_service_health(self) -> Dict[str, Any]:
        """Get health status of all analytics services."""
        try:
            # Check cache health
            cache_health = await self.cache_service.health_check()
            
            # Check BigQuery (basic connection test)
            try:
                await self.bigquery_service.query_custom("SELECT 1 as test_query")
                bigquery_health = {"status": "healthy"}
            except Exception as e:
                bigquery_health = {"status": "unhealthy", "error": str(e)}
            
            return {
                "service": "analytics",
                "status": "healthy" if all(
                    service["status"] == "healthy" 
                    for service in [cache_health, bigquery_health]
                ) else "degraded",
                "components": {
                    "cache": cache_health,
                    "bigquery": bigquery_health,
                    "charts": {"status": "healthy"}  # Chart service is stateless
                },
                "timestamp": datetime.now().isoformat()
            }
            
        except Exception as e:
            logger.error(f"Error checking service health: {e}")
            return {
                "service": "analytics",
                "status": "unhealthy",
                "error": str(e),
                "timestamp": datetime.now().isoformat()
            }
    
    async def get_cache_statistics(self) -> Dict[str, Any]:
        """Get cache performance statistics."""
        try:
            return await self.cache_service.get_cache_stats()
        except Exception as e:
            logger.error(f"Error getting cache statistics: {e}")
            raise


# Global analytics service instance
analytics_service = AnalyticsService()
