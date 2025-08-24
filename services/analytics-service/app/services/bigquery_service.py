"""
BigQuery Service for Analytics
"""
import json
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any, Tuple
from google.cloud import bigquery
from google.cloud.exceptions import NotFound
import pandas as pd

from app.config import settings, BIGQUERY_TABLES, get_bigquery_table_name
from app.models.analytics import (
    FormSummary, QuestionAnalytics, TrendAnalysis, PeriodType,
    ResponseStatus, QuestionType, BigQueryResponse, BigQueryForm
)


logger = logging.getLogger(__name__)


class BigQueryService:
    """Service for interacting with BigQuery for analytics."""
    
    def __init__(self):
        self.client = bigquery.Client(project=settings.bigquery_project_id)
        self.dataset_id = settings.bigquery_dataset_id
        self.project_id = settings.bigquery_project_id
        
    async def initialize_tables(self) -> None:
        """Initialize BigQuery tables if they don't exist."""
        try:
            # Check if dataset exists, create if not
            dataset_ref = self.client.dataset(self.dataset_id)
            try:
                self.client.get_dataset(dataset_ref)
                logger.info(f"Dataset {self.dataset_id} already exists")
            except NotFound:
                dataset = bigquery.Dataset(dataset_ref)
                dataset.location = settings.bigquery_location
                dataset = self.client.create_dataset(dataset)
                logger.info(f"Created dataset {self.dataset_id}")
            
            # Create tables
            for table_config in BIGQUERY_TABLES.values():
                await self._create_table_if_not_exists(table_config)
                
        except Exception as e:
            logger.error(f"Error initializing BigQuery tables: {e}")
            raise
    
    async def _create_table_if_not_exists(self, table_config: Dict[str, Any]) -> None:
        """Create a BigQuery table if it doesn't exist."""
        table_id = table_config["name"]
        schema = [
            bigquery.SchemaField(
                field["name"],
                field["type"],
                mode=field.get("mode", "NULLABLE")
            )
            for field in table_config["schema"]
        ]
        
        table_ref = self.client.dataset(self.dataset_id).table(table_id.split(".")[-1])
        
        try:
            self.client.get_table(table_ref)
            logger.info(f"Table {table_id} already exists")
        except NotFound:
            table = bigquery.Table(table_ref, schema=schema)
            table = self.client.create_table(table)
            logger.info(f"Created table {table_id}")
    
    async def get_form_summary(self, form_id: str, start_date: Optional[datetime] = None, 
                             end_date: Optional[datetime] = None) -> FormSummary:
        """Get form analytics summary."""
        try:
            # Build date filter
            date_filter = ""
            if start_date or end_date:
                conditions = []
                if start_date:
                    conditions.append(f"submitted_at >= '{start_date.isoformat()}'")
                if end_date:
                    conditions.append(f"submitted_at <= '{end_date.isoformat()}'")
                date_filter = f"AND {' AND '.join(conditions)}"
            
            query = f"""
            WITH form_stats AS (
                SELECT 
                    r.form_id,
                    f.title,
                    COUNT(*) as total_responses,
                    COUNT(CASE WHEN r.completion_time_seconds IS NOT NULL THEN 1 END) as completed_responses,
                    COUNT(CASE WHEN r.completion_time_seconds IS NULL THEN 1 END) as partial_responses,
                    AVG(r.completion_time_seconds) as avg_completion_time,
                    MIN(r.submitted_at) as first_response_date,
                    MAX(r.submitted_at) as last_response_date,
                    COUNT(DISTINCT r.user_id) as unique_respondents
                FROM `{get_bigquery_table_name(settings.responses_table)}` r
                JOIN `{get_bigquery_table_name(settings.forms_table)}` f ON r.form_id = f.form_id
                WHERE r.form_id = @form_id {date_filter}
                GROUP BY r.form_id, f.title
            ),
            daily_trends AS (
                SELECT 
                    DATE(submitted_at) as response_date,
                    COUNT(*) as daily_count
                FROM `{get_bigquery_table_name(settings.responses_table)}`
                WHERE form_id = @form_id 
                AND submitted_at >= DATE_SUB(CURRENT_DATE(), INTERVAL 7 DAY)
                GROUP BY DATE(submitted_at)
                ORDER BY response_date
            )
            SELECT 
                fs.*,
                ARRAY_AGG(
                    STRUCT(
                        dt.response_date as date,
                        dt.daily_count as count
                    ) ORDER BY dt.response_date
                ) as trend_data
            FROM form_stats fs
            LEFT JOIN daily_trends dt ON TRUE
            GROUP BY fs.form_id, fs.title, fs.total_responses, fs.completed_responses, 
                     fs.partial_responses, fs.avg_completion_time, fs.first_response_date,
                     fs.last_response_date, fs.unique_respondents
            """
            
            job_config = bigquery.QueryJobConfig(
                query_parameters=[
                    bigquery.ScalarQueryParameter("form_id", "STRING", form_id)
                ]
            )
            
            query_job = self.client.query(query, job_config=job_config)
            results = query_job.result()
            
            for row in results:
                completion_rate = (row.completed_responses / row.total_responses * 100) if row.total_responses > 0 else 0
                
                trend_data = []
                if row.trend_data:
                    for trend_item in row.trend_data:
                        if trend_item.date:  # Skip null dates
                            trend_data.append({
                                "date": trend_item.date.isoformat(),
                                "count": trend_item.count
                            })
                
                return FormSummary(
                    form_id=form_id,
                    title=row.title,
                    total_responses=row.total_responses,
                    completed_responses=row.completed_responses,
                    partial_responses=row.partial_responses,
                    average_completion_time=row.avg_completion_time,
                    completion_rate=completion_rate,
                    first_response_date=row.first_response_date,
                    last_response_date=row.last_response_date,
                    unique_respondents=row.unique_respondents,
                    response_rate_trend=trend_data
                )
            
            # If no results, return empty summary
            return FormSummary(
                form_id=form_id,
                title="Unknown Form",
                total_responses=0,
                completed_responses=0,
                partial_responses=0,
                average_completion_time=None,
                completion_rate=0.0,
                first_response_date=None,
                last_response_date=None,
                unique_respondents=0,
                response_rate_trend=[]
            )
            
        except Exception as e:
            logger.error(f"Error getting form summary for {form_id}: {e}")
            raise
    
    async def get_question_analytics(self, form_id: str, question_id: str,
                                   start_date: Optional[datetime] = None,
                                   end_date: Optional[datetime] = None) -> Dict[str, Any]:
        """Get analytics for a specific question."""
        try:
            # Build date filter
            date_filter = ""
            if start_date or end_date:
                conditions = []
                if start_date:
                    conditions.append(f"submitted_at >= '{start_date.isoformat()}'")
                if end_date:
                    conditions.append(f"submitted_at <= '{end_date.isoformat()}'")
                date_filter = f"AND {' AND '.join(conditions)}"
            
            # Query for question response distribution
            query = f"""
            WITH question_responses AS (
                SELECT 
                    r.response_id,
                    r.submitted_at,
                    JSON_EXTRACT_SCALAR(r.responses, '$.{question_id}') as answer,
                    r.completion_time_seconds
                FROM `{get_bigquery_table_name(settings.responses_table)}` r
                WHERE r.form_id = @form_id {date_filter}
            ),
            answer_stats AS (
                SELECT 
                    answer,
                    COUNT(*) as response_count,
                    COUNT(*) * 100.0 / (SELECT COUNT(*) FROM question_responses WHERE answer IS NOT NULL) as percentage
                FROM question_responses
                WHERE answer IS NOT NULL AND answer != ''
                GROUP BY answer
                ORDER BY response_count DESC
            ),
            question_summary AS (
                SELECT 
                    COUNT(*) as total_responses,
                    COUNT(CASE WHEN answer IS NOT NULL AND answer != '' THEN 1 END) as answered_responses,
                    COUNT(CASE WHEN answer IS NULL OR answer = '' THEN 1 END) as skipped_responses,
                    AVG(completion_time_seconds) as avg_response_time
                FROM question_responses
            )
            SELECT 
                qs.total_responses,
                qs.answered_responses,
                qs.skipped_responses,
                qs.avg_response_time,
                ARRAY_AGG(
                    STRUCT(
                        answer_stats.answer as value,
                        answer_stats.response_count as count,
                        answer_stats.percentage as percentage
                    ) ORDER BY answer_stats.response_count DESC
                ) as distribution
            FROM question_summary qs
            LEFT JOIN answer_stats ON TRUE
            GROUP BY qs.total_responses, qs.answered_responses, qs.skipped_responses, qs.avg_response_time
            """
            
            job_config = bigquery.QueryJobConfig(
                query_parameters=[
                    bigquery.ScalarQueryParameter("form_id", "STRING", form_id)
                ]
            )
            
            query_job = self.client.query(query, job_config=job_config)
            results = query_job.result()
            
            for row in results:
                response_rate = (row.answered_responses / row.total_responses * 100) if row.total_responses > 0 else 0
                skip_rate = (row.skipped_responses / row.total_responses * 100) if row.total_responses > 0 else 0
                
                distribution = []
                if row.distribution and row.distribution[0].value is not None:
                    distribution = [
                        {
                            "value": item.value,
                            "count": item.count,
                            "percentage": item.percentage,
                            "label": str(item.value)
                        }
                        for item in row.distribution
                        if item.value is not None
                    ]
                
                return {
                    "question_id": question_id,
                    "total_responses": row.total_responses,
                    "answered_responses": row.answered_responses,
                    "response_rate": response_rate,
                    "skip_rate": skip_rate,
                    "average_response_time": row.avg_response_time,
                    "distribution": distribution
                }
            
            # Return empty stats if no data
            return {
                "question_id": question_id,
                "total_responses": 0,
                "answered_responses": 0,
                "response_rate": 0.0,
                "skip_rate": 0.0,
                "average_response_time": None,
                "distribution": []
            }
            
        except Exception as e:
            logger.error(f"Error getting question analytics for {form_id}/{question_id}: {e}")
            raise
    
    async def get_trend_analysis(self, form_id: str, period: PeriodType,
                               start_date: Optional[datetime] = None,
                               end_date: Optional[datetime] = None) -> Dict[str, Any]:
        """Get trend analysis for a form."""
        try:
            # Set default date range if not provided
            if not end_date:
                end_date = datetime.now()
            if not start_date:
                if period == PeriodType.HOUR:
                    start_date = end_date - timedelta(days=1)
                elif period == PeriodType.DAY:
                    start_date = end_date - timedelta(days=7)
                elif period == PeriodType.WEEK:
                    start_date = end_date - timedelta(days=30)
                elif period == PeriodType.MONTH:
                    start_date = end_date - timedelta(days=90)
                else:
                    start_date = end_date - timedelta(days=365)
            
            # Determine grouping based on period
            date_trunc_format = {
                PeriodType.HOUR: "HOUR",
                PeriodType.DAY: "DAY", 
                PeriodType.WEEK: "WEEK",
                PeriodType.MONTH: "MONTH",
                PeriodType.QUARTER: "QUARTER",
                PeriodType.YEAR: "YEAR"
            }.get(period, "DAY")
            
            query = f"""
            WITH time_series AS (
                SELECT 
                    DATETIME_TRUNC(submitted_at, {date_trunc_format}) as period_start,
                    COUNT(*) as response_count,
                    COUNT(CASE WHEN completion_time_seconds IS NOT NULL THEN 1 END) as completed_count,
                    AVG(completion_time_seconds) as avg_completion_time
                FROM `{get_bigquery_table_name(settings.responses_table)}`
                WHERE form_id = @form_id
                AND submitted_at >= @start_date
                AND submitted_at <= @end_date
                GROUP BY period_start
                ORDER BY period_start
            ),
            stats AS (
                SELECT 
                    COUNT(*) as total_responses,
                    MAX(response_count) as peak_responses,
                    AVG(response_count) as avg_responses_per_period,
                    STDDEV(response_count) as stddev_responses
                FROM time_series
            )
            SELECT 
                ts.period_start,
                ts.response_count,
                ts.completed_count,
                ts.avg_completion_time,
                s.total_responses,
                s.peak_responses,
                s.avg_responses_per_period,
                s.stddev_responses
            FROM time_series ts
            CROSS JOIN stats s
            ORDER BY ts.period_start
            """
            
            job_config = bigquery.QueryJobConfig(
                query_parameters=[
                    bigquery.ScalarQueryParameter("form_id", "STRING", form_id),
                    bigquery.ScalarQueryParameter("start_date", "DATETIME", start_date),
                    bigquery.ScalarQueryParameter("end_date", "DATETIME", end_date)
                ]
            )
            
            query_job = self.client.query(query, job_config=job_config)
            results = list(query_job.result())
            
            if not results:
                return {
                    "form_id": form_id,
                    "period": period,
                    "total_period_responses": 0,
                    "trend_data": [],
                    "statistics": {}
                }
            
            # Extract trend data
            trend_data = []
            statistics = {}
            
            for row in results:
                trend_data.append({
                    "timestamp": row.period_start,
                    "value": row.response_count,
                    "completed_count": row.completed_count,
                    "avg_completion_time": row.avg_completion_time
                })
                
                # Get statistics from last row (same for all)
                statistics = {
                    "total_responses": row.total_responses,
                    "peak_responses": row.peak_responses,
                    "avg_responses_per_period": row.avg_responses_per_period,
                    "stddev_responses": row.stddev_responses
                }
            
            # Calculate growth rate
            if len(trend_data) >= 2:
                first_value = trend_data[0]["value"]
                last_value = trend_data[-1]["value"]
                if first_value > 0:
                    growth_rate = ((last_value - first_value) / first_value) * 100
                    statistics["growth_rate"] = growth_rate
            
            return {
                "form_id": form_id,
                "period": period,
                "total_period_responses": statistics.get("total_responses", 0),
                "trend_data": trend_data,
                "statistics": statistics
            }
            
        except Exception as e:
            logger.error(f"Error getting trend analysis for {form_id}: {e}")
            raise
    
    async def insert_response(self, response_data: Dict[str, Any]) -> None:
        """Insert a response into BigQuery."""
        try:
            table_id = get_bigquery_table_name(settings.responses_table)
            table = self.client.get_table(table_id)
            
            rows_to_insert = [response_data]
            errors = self.client.insert_rows_json(table, rows_to_insert)
            
            if errors:
                logger.error(f"Error inserting response: {errors}")
                raise Exception(f"BigQuery insert errors: {errors}")
                
        except Exception as e:
            logger.error(f"Error inserting response into BigQuery: {e}")
            raise
    
    async def batch_insert_responses(self, responses: List[Dict[str, Any]]) -> None:
        """Batch insert responses into BigQuery."""
        try:
            table_id = get_bigquery_table_name(settings.responses_table)
            table = self.client.get_table(table_id)
            
            errors = self.client.insert_rows_json(table, responses)
            
            if errors:
                logger.error(f"Error batch inserting responses: {errors}")
                raise Exception(f"BigQuery batch insert errors: {errors}")
                
        except Exception as e:
            logger.error(f"Error batch inserting responses into BigQuery: {e}")
            raise
    
    async def query_custom(self, query: str, parameters: Optional[List] = None) -> List[Dict[str, Any]]:
        """Execute a custom BigQuery query."""
        try:
            job_config = bigquery.QueryJobConfig()
            if parameters:
                job_config.query_parameters = parameters
            
            query_job = self.client.query(query, job_config=job_config)
            results = query_job.result()
            
            return [dict(row) for row in results]
            
        except Exception as e:
            logger.error(f"Error executing custom query: {e}")
            raise
