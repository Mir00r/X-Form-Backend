"""
Chart Service for Analytics Visualization
"""
import json
import logging
from typing import Dict, List, Any, Optional
from datetime import datetime
import plotly.graph_objects as go
import plotly.express as px
from plotly.utils import PlotlyJSONEncoder

from app.models.analytics import ChartData, ChartType, PeriodType

logger = logging.getLogger(__name__)


class ChartService:
    """Service for generating charts and visualizations."""
    
    def __init__(self):
        self.default_colors = [
            '#3366CC', '#DC3912', '#FF9900', '#109618', '#990099',
            '#3B3EAC', '#0099C6', '#DD4477', '#66AA00', '#B82E2E'
        ]
    
    def create_bar_chart(self, data: List[Dict[str, Any]], title: str,
                        x_field: str = "label", y_field: str = "value",
                        color_field: Optional[str] = None) -> ChartData:
        """Create a bar chart."""
        try:
            x_values = [item[x_field] for item in data]
            y_values = [item[y_field] for item in data]
            
            if color_field and all(color_field in item for item in data):
                colors = [item[color_field] for item in data]
                fig = px.bar(
                    x=x_values, y=y_values, color=colors,
                    title=title,
                    labels={x_field: x_field.title(), y_field: y_field.title()}
                )
            else:
                fig = go.Figure(data=[
                    go.Bar(
                        x=x_values,
                        y=y_values,
                        marker_color=self.default_colors[0],
                        name=title
                    )
                ])
                fig.update_layout(title=title)
            
            # Update layout
            fig.update_layout(
                xaxis_title=x_field.replace('_', ' ').title(),
                yaxis_title=y_field.replace('_', ' ').title(),
                template="plotly_white",
                height=400
            )
            
            return ChartData(
                type=ChartType.BAR,
                title=title,
                data=json.loads(json.dumps(fig, cls=PlotlyJSONEncoder)),
                config={"displayModeBar": True, "responsive": True}
            )
            
        except Exception as e:
            logger.error(f"Error creating bar chart: {e}")
            raise
    
    def create_pie_chart(self, data: List[Dict[str, Any]], title: str,
                        label_field: str = "label", value_field: str = "value") -> ChartData:
        """Create a pie chart."""
        try:
            labels = [item[label_field] for item in data]
            values = [item[value_field] for item in data]
            
            fig = go.Figure(data=[
                go.Pie(
                    labels=labels,
                    values=values,
                    hole=0.3,
                    marker_colors=self.default_colors[:len(labels)]
                )
            ])
            
            fig.update_layout(
                title=title,
                template="plotly_white",
                height=400,
                showlegend=True
            )
            
            return ChartData(
                type=ChartType.PIE,
                title=title,
                data=json.loads(json.dumps(fig, cls=PlotlyJSONEncoder)),
                config={"displayModeBar": True, "responsive": True}
            )
            
        except Exception as e:
            logger.error(f"Error creating pie chart: {e}")
            raise
    
    def create_line_chart(self, data: List[Dict[str, Any]], title: str,
                         x_field: str = "timestamp", y_field: str = "value",
                         group_field: Optional[str] = None) -> ChartData:
        """Create a line chart for trends."""
        try:
            if group_field and all(group_field in item for item in data):
                # Multiple lines grouped by field
                fig = px.line(
                    data, x=x_field, y=y_field, color=group_field,
                    title=title,
                    markers=True
                )
            else:
                # Single line
                x_values = [item[x_field] for item in data]
                y_values = [item[y_field] for item in data]
                
                fig = go.Figure(data=[
                    go.Scatter(
                        x=x_values,
                        y=y_values,
                        mode='lines+markers',
                        name=title,
                        line=dict(color=self.default_colors[0], width=2),
                        marker=dict(size=6)
                    )
                ])
                fig.update_layout(title=title)
            
            # Update layout
            fig.update_layout(
                xaxis_title=x_field.replace('_', ' ').title(),
                yaxis_title=y_field.replace('_', ' ').title(),
                template="plotly_white",
                height=400,
                hovermode='x unified'
            )
            
            return ChartData(
                type=ChartType.LINE,
                title=title,
                data=json.loads(json.dumps(fig, cls=PlotlyJSONEncoder)),
                config={"displayModeBar": True, "responsive": True}
            )
            
        except Exception as e:
            logger.error(f"Error creating line chart: {e}")
            raise
    
    def create_histogram(self, data: List[Dict[str, Any]], title: str,
                        value_field: str = "value", bins: int = 20) -> ChartData:
        """Create a histogram."""
        try:
            values = [item[value_field] for item in data if item[value_field] is not None]
            
            fig = go.Figure(data=[
                go.Histogram(
                    x=values,
                    nbinsx=bins,
                    marker_color=self.default_colors[0],
                    name=title
                )
            ])
            
            fig.update_layout(
                title=title,
                xaxis_title=value_field.replace('_', ' ').title(),
                yaxis_title="Frequency",
                template="plotly_white",
                height=400
            )
            
            return ChartData(
                type=ChartType.HISTOGRAM,
                title=title,
                data=json.loads(json.dumps(fig, cls=PlotlyJSONEncoder)),
                config={"displayModeBar": True, "responsive": True}
            )
            
        except Exception as e:
            logger.error(f"Error creating histogram: {e}")
            raise
    
    def create_heatmap(self, data: List[List[float]], title: str,
                      x_labels: List[str], y_labels: List[str]) -> ChartData:
        """Create a heatmap."""
        try:
            fig = go.Figure(data=go.Heatmap(
                z=data,
                x=x_labels,
                y=y_labels,
                colorscale='Viridis',
                hoverongaps=False
            ))
            
            fig.update_layout(
                title=title,
                template="plotly_white",
                height=400
            )
            
            return ChartData(
                type=ChartType.HEATMAP,
                title=title,
                data=json.loads(json.dumps(fig, cls=PlotlyJSONEncoder)),
                config={"displayModeBar": True, "responsive": True}
            )
            
        except Exception as e:
            logger.error(f"Error creating heatmap: {e}")
            raise
    
    def create_response_distribution_chart(self, distribution_data: List[Dict[str, Any]],
                                         question_type: str = "multiple_choice") -> ChartData:
        """Create chart for question response distribution."""
        if not distribution_data:
            return self._create_empty_chart("No Response Data")
        
        if question_type in ["multiple_choice", "single_choice", "dropdown"]:
            return self.create_bar_chart(
                data=distribution_data,
                title="Response Distribution",
                x_field="label",
                y_field="count"
            )
        elif question_type in ["rating", "scale"]:
            return self.create_bar_chart(
                data=distribution_data,
                title="Rating Distribution",
                x_field="value",
                y_field="count"
            )
        elif question_type == "boolean":
            return self.create_pie_chart(
                data=distribution_data,
                title="Yes/No Distribution",
                label_field="label",
                value_field="count"
            )
        else:
            # Default to bar chart for text and other types
            return self.create_bar_chart(
                data=distribution_data,
                title="Response Distribution",
                x_field="label",
                y_field="count"
            )
    
    def create_trend_chart(self, trend_data: List[Dict[str, Any]],
                          period: PeriodType = PeriodType.DAY) -> ChartData:
        """Create trend chart based on period type."""
        if not trend_data:
            return self._create_empty_chart("No Trend Data")
        
        # Format data for line chart
        formatted_data = []
        for item in trend_data:
            timestamp = item.get("timestamp")
            if isinstance(timestamp, str):
                timestamp = datetime.fromisoformat(timestamp.replace('Z', '+00:00'))
            
            formatted_data.append({
                "timestamp": timestamp,
                "value": item.get("value", 0),
                "completed_count": item.get("completed_count", 0)
            })
        
        # Sort by timestamp
        formatted_data.sort(key=lambda x: x["timestamp"])
        
        title = f"Response Trend ({period.value.title()})"
        return self.create_line_chart(
            data=formatted_data,
            title=title,
            x_field="timestamp",
            y_field="value"
        )
    
    def create_completion_rate_chart(self, summary_data: Dict[str, Any]) -> ChartData:
        """Create completion rate visualization."""
        total = summary_data.get("total_responses", 0)
        completed = summary_data.get("completed_responses", 0)
        partial = summary_data.get("partial_responses", 0)
        
        if total == 0:
            return self._create_empty_chart("No Response Data")
        
        data = [
            {"label": "Completed", "value": completed},
            {"label": "Partial", "value": partial}
        ]
        
        return self.create_pie_chart(
            data=data,
            title="Response Completion Status",
            label_field="label",
            value_field="value"
        )
    
    def create_response_time_chart(self, response_times: List[float]) -> ChartData:
        """Create response time distribution chart."""
        if not response_times:
            return self._create_empty_chart("No Response Time Data")
        
        # Convert to chart data format
        data = [{"value": time} for time in response_times if time is not None]
        
        return self.create_histogram(
            data=data,
            title="Response Time Distribution",
            value_field="value",
            bins=20
        )
    
    def create_multi_metric_chart(self, metrics: Dict[str, List[Dict[str, Any]]],
                                 title: str = "Multi-Metric Analysis") -> ChartData:
        """Create a chart with multiple metrics."""
        try:
            fig = go.Figure()
            
            colors_iter = iter(self.default_colors)
            
            for metric_name, metric_data in metrics.items():
                if metric_data:
                    x_values = [item.get("timestamp") or item.get("label", "") for item in metric_data]
                    y_values = [item.get("value", 0) for item in metric_data]
                    
                    fig.add_trace(go.Scatter(
                        x=x_values,
                        y=y_values,
                        mode='lines+markers',
                        name=metric_name,
                        line=dict(color=next(colors_iter, self.default_colors[0]))
                    ))
            
            fig.update_layout(
                title=title,
                xaxis_title="Time",
                yaxis_title="Value",
                template="plotly_white",
                height=400,
                hovermode='x unified'
            )
            
            return ChartData(
                type=ChartType.LINE,
                title=title,
                data=json.loads(json.dumps(fig, cls=PlotlyJSONEncoder)),
                config={"displayModeBar": True, "responsive": True}
            )
            
        except Exception as e:
            logger.error(f"Error creating multi-metric chart: {e}")
            raise
    
    def _create_empty_chart(self, message: str = "No Data Available") -> ChartData:
        """Create an empty chart with a message."""
        fig = go.Figure()
        fig.add_annotation(
            text=message,
            xref="paper", yref="paper",
            x=0.5, y=0.5,
            xanchor='center', yanchor='middle',
            font=dict(size=16)
        )
        fig.update_layout(
            template="plotly_white",
            height=400,
            showlegend=False
        )
        
        return ChartData(
            type=ChartType.BAR,
            title=message,
            data=json.loads(json.dumps(fig, cls=PlotlyJSONEncoder)),
            config={"displayModeBar": False}
        )


# Global chart service instance
chart_service = ChartService()
