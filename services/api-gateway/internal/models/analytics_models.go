package models

import (
	"time"
)

// AnalyticsRequest represents analytics query parameters
type AnalyticsRequest struct {
	FormIDs     []string           `form:"form_ids" json:"form_ids,omitempty" description:"Form IDs to analyze"`
	DateRange   DateRange          `json:"date_range,omitempty" description:"Date range for analytics"`
	Granularity string             `form:"granularity" json:"granularity" example:"day" enums:"hour,day,week,month" description:"Data granularity"`
	Metrics     []string           `form:"metrics" json:"metrics,omitempty" example:"views,submissions,completion_rate" description:"Metrics to include"`
	Dimensions  []string           `form:"dimensions" json:"dimensions,omitempty" example:"device,country,source" description:"Dimensions to group by"`
	Filters     AnalyticsFilters   `json:"filters,omitempty" description:"Analytics filters"`
	Segments    []AnalyticsSegment `json:"segments,omitempty" description:"User segments"`
	Comparison  *ComparisonPeriod  `json:"comparison,omitempty" description:"Comparison period"`
}

// AnalyticsFilters represents filters for analytics queries
type AnalyticsFilters struct {
	Countries    []string            `json:"countries,omitempty" description:"Filter by countries"`
	Devices      []string            `json:"devices,omitempty" description:"Filter by device types"`
	Sources      []string            `json:"sources,omitempty" description:"Filter by traffic sources"`
	UserTypes    []string            `json:"user_types,omitempty" description:"Filter by user types"`
	Languages    []string            `json:"languages,omitempty" description:"Filter by languages"`
	Browsers     []string            `json:"browsers,omitempty" description:"Filter by browsers"`
	OSes         []string            `json:"oses,omitempty" description:"Filter by operating systems"`
	CustomFields map[string][]string `json:"custom_fields,omitempty" description:"Filter by custom field values"`
}

// AnalyticsSegment represents a user segment for analytics
type AnalyticsSegment struct {
	Name        string             `json:"name" example:"Mobile Users" description:"Segment name"`
	Conditions  []SegmentCondition `json:"conditions" description:"Segment conditions"`
	Operator    string             `json:"operator" example:"AND" enums:"AND,OR" description:"Logical operator"`
	Color       string             `json:"color,omitempty" example:"#007bff" description:"Segment color"`
	Description string             `json:"description,omitempty" description:"Segment description"`
}

// SegmentCondition represents a condition for user segmentation
type SegmentCondition struct {
	Dimension string      `json:"dimension" example:"device_type" description:"Dimension to check"`
	Operator  string      `json:"operator" example:"equals" enums:"equals,not_equals,contains,in,not_in,greater_than,less_than" description:"Comparison operator"`
	Value     interface{} `json:"value" example:"mobile" description:"Value to compare"`
}

// ComparisonPeriod represents a comparison period for analytics
type ComparisonPeriod struct {
	Type      string     `json:"type" example:"previous_period" enums:"previous_period,same_period_last_year,custom" description:"Comparison type"`
	StartDate *time.Time `json:"start_date,omitempty" example:"2025-08-06T00:00:00Z" description:"Custom start date"`
	EndDate   *time.Time `json:"end_date,omitempty" example:"2025-09-05T23:59:59Z" description:"Custom end date"`
}

// AnalyticsResponse represents comprehensive analytics data
type AnalyticsResponse struct {
	Summary    AnalyticsSummary         `json:"summary" description:"High-level summary metrics"`
	TimeSeries []TimeSeriesData         `json:"time_series" description:"Time-based data points"`
	Dimensions map[string]DimensionData `json:"dimensions" description:"Dimensional breakdown"`
	Segments   map[string]SegmentData   `json:"segments,omitempty" description:"Segment-based analytics"`
	Comparison *ComparisonData          `json:"comparison,omitempty" description:"Comparison period data"`
	Insights   []AnalyticsInsight       `json:"insights" description:"AI-generated insights"`
	Funnel     *FunnelAnalytics         `json:"funnel,omitempty" description:"Funnel analysis"`
	Cohort     *CohortAnalytics         `json:"cohort,omitempty" description:"Cohort analysis"`
	Metadata   AnalyticsMetadata        `json:"metadata" description:"Query metadata"`
}

// AnalyticsSummary represents high-level summary metrics
type AnalyticsSummary struct {
	TotalViews           int     `json:"total_views" example:"1500" description:"Total form views"`
	UniqueViews          int     `json:"unique_views" example:"1200" description:"Unique form views"`
	TotalSubmissions     int     `json:"total_submissions" example:"450" description:"Total submissions"`
	CompletedSubmissions int     `json:"completed_submissions" example:"400" description:"Completed submissions"`
	DraftSubmissions     int     `json:"draft_submissions" example:"35" description:"Draft submissions"`
	AbandonedSubmissions int     `json:"abandoned_submissions" example:"15" description:"Abandoned submissions"`
	CompletionRate       float64 `json:"completion_rate" example:"0.89" description:"Overall completion rate"`
	ConversionRate       float64 `json:"conversion_rate" example:"0.30" description:"View to submission rate"`
	AverageTime          int     `json:"average_time" example:"240" description:"Average completion time (seconds)"`
	BounceRate           float64 `json:"bounce_rate" example:"0.25" description:"Bounce rate"`
	ReturnVisitorRate    float64 `json:"return_visitor_rate" example:"0.15" description:"Return visitor rate"`
	TopExitPage          string  `json:"top_exit_page,omitempty" example:"page_2" description:"Most common exit point"`
	PeakHour             int     `json:"peak_hour" example:"14" description:"Peak traffic hour (0-23)"`
	PeakDay              string  `json:"peak_day" example:"Tuesday" description:"Peak traffic day"`
	TopCountry           string  `json:"top_country,omitempty" example:"United States" description:"Top country by traffic"`
	TopDevice            string  `json:"top_device,omitempty" example:"desktop" description:"Top device type"`
	TopSource            string  `json:"top_source,omitempty" example:"google.com" description:"Top traffic source"`
}

// TimeSeriesData represents time-based analytics data
type TimeSeriesData struct {
	Timestamp         time.Time `json:"timestamp" example:"2025-09-06T00:00:00Z" description:"Data point timestamp"`
	Views             int       `json:"views" example:"100" description:"Views count"`
	UniqueViews       int       `json:"unique_views" example:"85" description:"Unique views count"`
	Submissions       int       `json:"submissions" example:"30" description:"Submissions count"`
	Completions       int       `json:"completions" example:"28" description:"Completions count"`
	CompletionRate    float64   `json:"completion_rate" example:"0.93" description:"Completion rate"`
	ConversionRate    float64   `json:"conversion_rate" example:"0.30" description:"Conversion rate"`
	AverageTime       int       `json:"average_time" example:"235" description:"Average time (seconds)"`
	BounceRate        float64   `json:"bounce_rate" example:"0.20" description:"Bounce rate"`
	NewVisitors       int       `json:"new_visitors" example:"75" description:"New visitors"`
	ReturningVisitors int       `json:"returning_visitors" example:"10" description:"Returning visitors"`
}

// DimensionData represents dimensional breakdown data
type DimensionData struct {
	Name       string           `json:"name" example:"United States" description:"Dimension value"`
	Value      interface{}      `json:"value" description:"Dimension identifier"`
	Metrics    DimensionMetrics `json:"metrics" description:"Metrics for this dimension"`
	Percentage float64          `json:"percentage" example:"45.5" description:"Percentage of total"`
	Rank       int              `json:"rank" example:"1" description:"Rank within dimension"`
	Trend      string           `json:"trend,omitempty" example:"up" enums:"up,down,stable" description:"Trend direction"`
	TrendValue float64          `json:"trend_value,omitempty" example:"15.5" description:"Trend percentage change"`
	Children   []DimensionData  `json:"children,omitempty" description:"Sub-dimensions"`
}

// DimensionMetrics represents metrics for a dimension
type DimensionMetrics struct {
	Views          int     `json:"views" example:"680" description:"Views count"`
	UniqueViews    int     `json:"unique_views" example:"545" description:"Unique views count"`
	Submissions    int     `json:"submissions" example:"205" description:"Submissions count"`
	Completions    int     `json:"completions" example:"182" description:"Completions count"`
	CompletionRate float64 `json:"completion_rate" example:"0.89" description:"Completion rate"`
	ConversionRate float64 `json:"conversion_rate" example:"0.30" description:"Conversion rate"`
	AverageTime    int     `json:"average_time" example:"255" description:"Average time (seconds)"`
	BounceRate     float64 `json:"bounce_rate" example:"0.23" description:"Bounce rate"`
}

// SegmentData represents analytics data for a specific segment
type SegmentData struct {
	Name       string           `json:"name" example:"Mobile Users" description:"Segment name"`
	Size       int              `json:"size" example:"340" description:"Segment size"`
	Percentage float64          `json:"percentage" example:"28.3" description:"Percentage of total"`
	Metrics    DimensionMetrics `json:"metrics" description:"Segment metrics"`
	Growth     *GrowthData      `json:"growth,omitempty" description:"Segment growth data"`
}

// GrowthData represents growth/trend information
type GrowthData struct {
	Rate      float64 `json:"rate" example:"15.5" description:"Growth rate percentage"`
	Direction string  `json:"direction" example:"up" enums:"up,down,stable" description:"Growth direction"`
	Period    string  `json:"period" example:"month" description:"Growth period"`
	Absolute  int     `json:"absolute" example:"45" description:"Absolute growth number"`
}

// ComparisonData represents comparison period analytics
type ComparisonData struct {
	Period       ComparisonPeriod      `json:"period" description:"Comparison period"`
	Summary      AnalyticsSummary      `json:"summary" description:"Comparison summary"`
	Changes      map[string]ChangeData `json:"changes" description:"Metric changes"`
	Significance map[string]bool       `json:"significance" description:"Statistical significance"`
}

// ChangeData represents change information between periods
type ChangeData struct {
	Current     float64 `json:"current" example:"450" description:"Current period value"`
	Previous    float64 `json:"previous" example:"380" description:"Previous period value"`
	Absolute    float64 `json:"absolute" example:"70" description:"Absolute change"`
	Percentage  float64 `json:"percentage" example:"18.42" description:"Percentage change"`
	Direction   string  `json:"direction" example:"up" enums:"up,down,stable" description:"Change direction"`
	Significant bool    `json:"significant" example:"true" description:"Statistically significant"`
}

// AnalyticsInsight represents AI-generated insights
type AnalyticsInsight struct {
	Type        string                 `json:"type" example:"trend" enums:"trend,anomaly,opportunity,alert" description:"Insight type"`
	Title       string                 `json:"title" example:"Mobile Traffic Increasing" description:"Insight title"`
	Description string                 `json:"description" example:"Mobile traffic has increased 25% over the past week" description:"Insight description"`
	Severity    string                 `json:"severity" example:"medium" enums:"low,medium,high,critical" description:"Insight severity"`
	Confidence  float64                `json:"confidence" example:"0.85" description:"Confidence score (0-1)"`
	Impact      string                 `json:"impact" example:"positive" enums:"positive,negative,neutral" description:"Impact assessment"`
	Metric      string                 `json:"metric,omitempty" example:"mobile_views" description:"Related metric"`
	Value       float64                `json:"value,omitempty" example:"25.5" description:"Insight value"`
	Dimension   string                 `json:"dimension,omitempty" example:"device_type" description:"Related dimension"`
	Timestamp   time.Time              `json:"timestamp" example:"2025-09-06T12:00:00Z" description:"Insight generation time"`
	Actions     []RecommendedAction    `json:"actions,omitempty" description:"Recommended actions"`
	Context     map[string]interface{} `json:"context,omitempty" description:"Additional context"`
}

// RecommendedAction represents a recommended action based on insights
type RecommendedAction struct {
	Title           string   `json:"title" example:"Optimize for Mobile" description:"Action title"`
	Description     string   `json:"description" example:"Consider improving mobile user experience" description:"Action description"`
	Priority        string   `json:"priority" example:"high" enums:"low,medium,high" description:"Action priority"`
	Category        string   `json:"category" example:"optimization" description:"Action category"`
	EstimatedImpact string   `json:"estimated_impact,omitempty" example:"15% improvement in mobile conversion" description:"Estimated impact"`
	Resources       []string `json:"resources,omitempty" description:"Required resources or links"`
}

// FunnelAnalytics represents funnel analysis data
type FunnelAnalytics struct {
	Steps      []FunnelStep          `json:"steps" description:"Funnel steps"`
	Overview   FunnelOverview        `json:"overview" description:"Funnel overview"`
	Segments   map[string]FunnelData `json:"segments,omitempty" description:"Segment-based funnel data"`
	Conversion ConversionAnalysis    `json:"conversion" description:"Conversion analysis"`
}

// FunnelStep represents a step in the funnel
type FunnelStep struct {
	Name           string  `json:"name" example:"Form View" description:"Step name"`
	Users          int     `json:"users" example:"1200" description:"Users at this step"`
	Percentage     float64 `json:"percentage" example:"100.0" description:"Percentage of funnel start"`
	DropoffRate    float64 `json:"dropoff_rate" example:"0.0" description:"Dropoff rate from previous step"`
	ConversionRate float64 `json:"conversion_rate" example:"1.0" description:"Conversion rate to next step"`
	AverageTime    int     `json:"average_time" example:"30" description:"Average time at step (seconds)"`
	Order          int     `json:"order" example:"1" description:"Step order"`
}

// FunnelOverview represents funnel overview metrics
type FunnelOverview struct {
	TotalSteps        int     `json:"total_steps" example:"4" description:"Total funnel steps"`
	OverallConversion float64 `json:"overall_conversion" example:"0.33" description:"Overall conversion rate"`
	BiggestDropoff    string  `json:"biggest_dropoff" example:"Form Start to First Field" description:"Biggest dropoff point"`
	DropoffRate       float64 `json:"dropoff_rate" example:"0.35" description:"Biggest dropoff rate"`
	AverageTime       int     `json:"average_time" example:"240" description:"Average funnel completion time"`
}

// FunnelData represents funnel data for a specific segment
type FunnelData struct {
	Steps    []FunnelStep   `json:"steps" description:"Segment funnel steps"`
	Overview FunnelOverview `json:"overview" description:"Segment funnel overview"`
}

// ConversionAnalysis represents conversion analysis
type ConversionAnalysis struct {
	TopPerformingPages []PagePerformance `json:"top_performing_pages" description:"Best performing pages"`
	BottleneckPages    []PagePerformance `json:"bottleneck_pages" description:"Pages with highest dropoff"`
	OptimizationTips   []OptimizationTip `json:"optimization_tips" description:"Optimization recommendations"`
}

// PagePerformance represents page performance metrics
type PagePerformance struct {
	PageID         string   `json:"page_id" example:"page_1" description:"Page identifier"`
	PageName       string   `json:"page_name" example:"Contact Information" description:"Page name"`
	Views          int      `json:"views" example:"800" description:"Page views"`
	Completions    int      `json:"completions" example:"720" description:"Page completions"`
	CompletionRate float64  `json:"completion_rate" example:"0.90" description:"Page completion rate"`
	AverageTime    int      `json:"average_time" example:"45" description:"Average time on page"`
	ExitRate       float64  `json:"exit_rate" example:"0.10" description:"Page exit rate"`
	Issues         []string `json:"issues,omitempty" description:"Identified issues"`
}

// OptimizationTip represents an optimization recommendation
type OptimizationTip struct {
	Category    string   `json:"category" example:"form_design" description:"Tip category"`
	Title       string   `json:"title" example:"Reduce Form Length" description:"Tip title"`
	Description string   `json:"description" example:"Consider breaking long forms into multiple pages" description:"Tip description"`
	Impact      string   `json:"impact" example:"high" enums:"low,medium,high" description:"Expected impact"`
	Difficulty  string   `json:"difficulty" example:"medium" enums:"easy,medium,hard" description:"Implementation difficulty"`
	Resources   []string `json:"resources,omitempty" description:"Helpful resources"`
}

// CohortAnalytics represents cohort analysis data
type CohortAnalytics struct {
	Cohorts    []CohortData  `json:"cohorts" description:"Cohort data"`
	MetricType string        `json:"metric_type" example:"retention" description:"Cohort metric type"`
	Periods    []string      `json:"periods" description:"Analysis periods"`
	Summary    CohortSummary `json:"summary" description:"Cohort summary"`
}

// CohortData represents data for a specific cohort
type CohortData struct {
	Name      string             `json:"name" example:"Week of 2025-09-01" description:"Cohort name"`
	StartDate time.Time          `json:"start_date" example:"2025-09-01T00:00:00Z" description:"Cohort start date"`
	Size      int                `json:"size" example:"150" description:"Cohort size"`
	Values    map[string]float64 `json:"values" description:"Metric values by period"`
	Trend     string             `json:"trend" example:"declining" enums:"improving,stable,declining" description:"Cohort trend"`
}

// CohortSummary represents cohort analysis summary
type CohortSummary struct {
	AverageRetention map[string]float64 `json:"average_retention" description:"Average retention by period"`
	BestCohort       string             `json:"best_cohort" example:"Week of 2025-08-15" description:"Best performing cohort"`
	WorstCohort      string             `json:"worst_cohort" example:"Week of 2025-07-20" description:"Worst performing cohort"`
	TrendDirection   string             `json:"trend_direction" example:"stable" description:"Overall trend direction"`
}

// AnalyticsMetadata represents query metadata
type AnalyticsMetadata struct {
	QueryTime     time.Duration `json:"query_time" example:"150ms" description:"Query execution time"`
	DataFreshness time.Time     `json:"data_freshness" example:"2025-09-06T11:55:00Z" description:"Data last updated"`
	RecordCount   int           `json:"record_count" example:"1500" description:"Number of records processed"`
	SampleRate    float64       `json:"sample_rate,omitempty" example:"1.0" description:"Data sampling rate"`
	Filters       []string      `json:"filters,omitempty" description:"Applied filters"`
	Segments      []string      `json:"segments,omitempty" description:"Applied segments"`
	CacheHit      bool          `json:"cache_hit" example:"false" description:"Whether data was served from cache"`
	QueryID       string        `json:"query_id" example:"q_123456" description:"Unique query identifier"`
}

// DashboardRequest represents dashboard configuration request
type DashboardRequest struct {
	Name        string            `json:"name" binding:"required" example:"Marketing Dashboard" description:"Dashboard name"`
	Description string            `json:"description,omitempty" example:"Overview of marketing metrics" description:"Dashboard description"`
	Layout      string            `json:"layout" example:"grid" enums:"grid,list,custom" description:"Dashboard layout"`
	Widgets     []DashboardWidget `json:"widgets" description:"Dashboard widgets"`
	Filters     DashboardFilters  `json:"filters,omitempty" description:"Global dashboard filters"`
	Settings    DashboardSettings `json:"settings" description:"Dashboard settings"`
	Sharing     DashboardSharing  `json:"sharing" description:"Sharing settings"`
}

// DashboardWidget represents a dashboard widget
type DashboardWidget struct {
	ID          string                 `json:"id" example:"widget_1" description:"Widget identifier"`
	Type        string                 `json:"type" example:"chart" enums:"chart,metric,table,text,funnel,heatmap" description:"Widget type"`
	Title       string                 `json:"title" example:"Daily Submissions" description:"Widget title"`
	Description string                 `json:"description,omitempty" description:"Widget description"`
	Position    WidgetPosition         `json:"position" description:"Widget position and size"`
	Config      WidgetConfig           `json:"config" description:"Widget configuration"`
	Data        interface{}            `json:"data,omitempty" description:"Widget data"`
	Filters     map[string]interface{} `json:"filters,omitempty" description:"Widget-specific filters"`
	Refresh     int                    `json:"refresh,omitempty" example:"300" description:"Auto-refresh interval (seconds)"`
}

// WidgetPosition represents widget position and size
type WidgetPosition struct {
	X      int `json:"x" example:"0" description:"X position"`
	Y      int `json:"y" example:"0" description:"Y position"`
	Width  int `json:"width" example:"6" description:"Widget width"`
	Height int `json:"height" example:"4" description:"Widget height"`
}

// WidgetConfig represents widget configuration
type WidgetConfig struct {
	ChartType   string                 `json:"chart_type,omitempty" example:"line" enums:"line,bar,pie,area,scatter" description:"Chart type"`
	Metrics     []string               `json:"metrics,omitempty" description:"Metrics to display"`
	Dimensions  []string               `json:"dimensions,omitempty" description:"Dimensions to group by"`
	TimeRange   string                 `json:"time_range,omitempty" example:"7d" description:"Time range"`
	Granularity string                 `json:"granularity,omitempty" example:"day" description:"Data granularity"`
	Colors      []string               `json:"colors,omitempty" description:"Custom colors"`
	ShowLegend  bool                   `json:"show_legend" example:"true" description:"Show chart legend"`
	ShowGrid    bool                   `json:"show_grid" example:"true" description:"Show grid lines"`
	Threshold   *ThresholdConfig       `json:"threshold,omitempty" description:"Threshold configuration"`
	Format      map[string]interface{} `json:"format,omitempty" description:"Number formatting options"`
	Sorting     *SortConfig            `json:"sorting,omitempty" description:"Sorting configuration"`
	Limit       int                    `json:"limit,omitempty" example:"10" description:"Data limit"`
}

// ThresholdConfig represents threshold configuration
type ThresholdConfig struct {
	Value float64 `json:"value" example:"80.0" description:"Threshold value"`
	Color string  `json:"color" example:"#ff0000" description:"Threshold color"`
	Label string  `json:"label,omitempty" example:"Target" description:"Threshold label"`
	Type  string  `json:"type" example:"line" enums:"line,area,band" description:"Threshold type"`
}

// SortConfig represents sorting configuration
type SortConfig struct {
	Field string `json:"field" example:"submissions" description:"Sort field"`
	Order string `json:"order" example:"desc" enums:"asc,desc" description:"Sort order"`
}

// DashboardFilters represents global dashboard filters
type DashboardFilters struct {
	DateRange   DateRange              `json:"date_range,omitempty" description:"Global date range"`
	FormIDs     []string               `json:"form_ids,omitempty" description:"Form filter"`
	UserSegment string                 `json:"user_segment,omitempty" description:"User segment filter"`
	Custom      map[string]interface{} `json:"custom,omitempty" description:"Custom filters"`
}

// DashboardSettings represents dashboard settings
type DashboardSettings struct {
	AutoRefresh     bool   `json:"auto_refresh" example:"true" description:"Enable auto-refresh"`
	RefreshInterval int    `json:"refresh_interval" example:"300" description:"Refresh interval (seconds)"`
	Theme           string `json:"theme" example:"light" enums:"light,dark,auto" description:"Dashboard theme"`
	Timezone        string `json:"timezone" example:"UTC" description:"Dashboard timezone"`
	Currency        string `json:"currency,omitempty" example:"USD" description:"Currency for monetary values"`
	Locale          string `json:"locale" example:"en-US" description:"Locale for formatting"`
}

// DashboardSharing represents dashboard sharing settings
type DashboardSharing struct {
	IsPublic    bool       `json:"is_public" example:"false" description:"Public access"`
	SharedWith  []string   `json:"shared_with,omitempty" description:"Users with access"`
	Permissions string     `json:"permissions" example:"view" enums:"view,edit,admin" description:"Default permissions"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" description:"Share expiration"`
	Password    string     `json:"password,omitempty" description:"Access password"`
}

// DashboardResponse represents a complete dashboard
type DashboardResponse struct {
	ID          string            `json:"id" example:"dashboard_123" description:"Dashboard ID"`
	Name        string            `json:"name" example:"Marketing Dashboard" description:"Dashboard name"`
	Description string            `json:"description,omitempty" description:"Dashboard description"`
	Layout      string            `json:"layout" example:"grid" description:"Dashboard layout"`
	Widgets     []DashboardWidget `json:"widgets" description:"Dashboard widgets"`
	Filters     DashboardFilters  `json:"filters,omitempty" description:"Global filters"`
	Settings    DashboardSettings `json:"settings" description:"Dashboard settings"`
	Sharing     DashboardSharing  `json:"sharing" description:"Sharing settings"`
	Owner       BasicUser         `json:"owner" description:"Dashboard owner"`
	CreatedAt   time.Time         `json:"created_at" example:"2025-09-01T00:00:00Z" description:"Creation time"`
	UpdatedAt   time.Time         `json:"updated_at" example:"2025-09-06T12:00:00Z" description:"Last update time"`
	LastViewed  *time.Time        `json:"last_viewed,omitempty" description:"Last viewed time"`
	ViewCount   int               `json:"view_count" example:"45" description:"View count"`
}
