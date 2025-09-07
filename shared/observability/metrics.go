package observability

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// MetricsProvider handles Prometheus metrics collection
type MetricsProvider struct {
	// HTTP Metrics
	httpRequestsTotal     prometheus.CounterVec
	httpRequestDuration   prometheus.HistogramVec
	httpActiveConnections prometheus.Gauge
	httpRequestSize       prometheus.HistogramVec
	httpResponseSize      prometheus.HistogramVec

	// Service Metrics
	serviceUptime          prometheus.Counter
	serviceHealthScore     prometheus.Gauge
	serviceErrorsTotal     prometheus.CounterVec
	serviceOperationsTotal prometheus.CounterVec

	// Database Metrics
	dbConnectionsActive prometheus.Gauge
	dbConnectionsIdle   prometheus.Gauge
	dbQueryDuration     prometheus.HistogramVec
	dbQueriesTotal      prometheus.CounterVec

	// External Service Metrics
	externalServiceCalls    prometheus.CounterVec
	externalServiceDuration prometheus.HistogramVec

	// Business Metrics
	businessMetrics prometheus.CounterVec
	businessGauges  prometheus.GaugeVec

	registry    prometheus.Registry
	serviceName string
	logger      *zap.Logger
}

// MetricsConfig holds configuration for metrics
type MetricsConfig struct {
	ServiceName string
	Environment string
	Version     string
	Namespace   string
}

// NewMetricsProvider creates a new metrics provider
func NewMetricsProvider(config MetricsConfig, logger *zap.Logger) *MetricsProvider {
	registry := prometheus.NewRegistry()

	// Common labels
	commonLabels := prometheus.Labels{
		"service":     config.ServiceName,
		"environment": config.Environment,
		"version":     config.Version,
	}

	mp := &MetricsProvider{
		registry:    *registry,
		serviceName: config.ServiceName,
		logger:      logger,

		// HTTP Metrics
		httpRequestsTotal: *prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   config.Namespace,
				Name:        "http_requests_total",
				Help:        "Total number of HTTP requests",
				ConstLabels: commonLabels,
			},
			[]string{"method", "endpoint", "status_code", "user_id"},
		),

		httpRequestDuration: *prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   config.Namespace,
				Name:        "http_request_duration_seconds",
				Help:        "HTTP request duration in seconds",
				ConstLabels: commonLabels,
				Buckets:     prometheus.DefBuckets,
			},
			[]string{"method", "endpoint", "status_code"},
		),

		httpActiveConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   config.Namespace,
				Name:        "http_active_connections",
				Help:        "Number of active HTTP connections",
				ConstLabels: commonLabels,
			},
		),

		httpRequestSize: *prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   config.Namespace,
				Name:        "http_request_size_bytes",
				Help:        "HTTP request size in bytes",
				ConstLabels: commonLabels,
				Buckets:     prometheus.ExponentialBuckets(100, 10, 7),
			},
			[]string{"method", "endpoint"},
		),

		httpResponseSize: *prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   config.Namespace,
				Name:        "http_response_size_bytes",
				Help:        "HTTP response size in bytes",
				ConstLabels: commonLabels,
				Buckets:     prometheus.ExponentialBuckets(100, 10, 7),
			},
			[]string{"method", "endpoint", "status_code"},
		),

		// Service Metrics
		serviceUptime: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   config.Namespace,
				Name:        "service_uptime_seconds_total",
				Help:        "Total service uptime in seconds",
				ConstLabels: commonLabels,
			},
		),

		serviceHealthScore: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   config.Namespace,
				Name:        "service_health_score",
				Help:        "Service health score (0-1)",
				ConstLabels: commonLabels,
			},
		),

		serviceErrorsTotal: *prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   config.Namespace,
				Name:        "service_errors_total",
				Help:        "Total number of service errors",
				ConstLabels: commonLabels,
			},
			[]string{"error_type", "component", "severity"},
		),

		serviceOperationsTotal: *prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   config.Namespace,
				Name:        "service_operations_total",
				Help:        "Total number of service operations",
				ConstLabels: commonLabels,
			},
			[]string{"operation", "status", "component"},
		),

		// Database Metrics
		dbConnectionsActive: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   config.Namespace,
				Name:        "db_connections_active",
				Help:        "Number of active database connections",
				ConstLabels: commonLabels,
			},
		),

		dbConnectionsIdle: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   config.Namespace,
				Name:        "db_connections_idle",
				Help:        "Number of idle database connections",
				ConstLabels: commonLabels,
			},
		),

		dbQueryDuration: *prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   config.Namespace,
				Name:        "db_query_duration_seconds",
				Help:        "Database query duration in seconds",
				ConstLabels: commonLabels,
				Buckets:     prometheus.DefBuckets,
			},
			[]string{"query_type", "table", "status"},
		),

		dbQueriesTotal: *prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   config.Namespace,
				Name:        "db_queries_total",
				Help:        "Total number of database queries",
				ConstLabels: commonLabels,
			},
			[]string{"query_type", "table", "status"},
		),

		// External Service Metrics
		externalServiceCalls: *prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   config.Namespace,
				Name:        "external_service_calls_total",
				Help:        "Total number of external service calls",
				ConstLabels: commonLabels,
			},
			[]string{"service", "method", "endpoint", "status_code"},
		),

		externalServiceDuration: *prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   config.Namespace,
				Name:        "external_service_duration_seconds",
				Help:        "External service call duration in seconds",
				ConstLabels: commonLabels,
				Buckets:     prometheus.DefBuckets,
			},
			[]string{"service", "method", "endpoint"},
		),

		// Business Metrics
		businessMetrics: *prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   config.Namespace,
				Name:        "business_events_total",
				Help:        "Total number of business events",
				ConstLabels: commonLabels,
			},
			[]string{"event_type", "category", "status"},
		),

		businessGauges: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   config.Namespace,
				Name:        "business_metrics",
				Help:        "Business metrics gauge",
				ConstLabels: commonLabels,
			},
			[]string{"metric_name", "category"},
		),
	}

	// Register all metrics
	registry.MustRegister(
		&mp.httpRequestsTotal,
		&mp.httpRequestDuration,
		mp.httpActiveConnections,
		&mp.httpRequestSize,
		&mp.httpResponseSize,
		mp.serviceUptime,
		mp.serviceHealthScore,
		&mp.serviceErrorsTotal,
		&mp.serviceOperationsTotal,
		mp.dbConnectionsActive,
		mp.dbConnectionsIdle,
		&mp.dbQueryDuration,
		&mp.dbQueriesTotal,
		&mp.externalServiceCalls,
		&mp.externalServiceDuration,
		&mp.businessMetrics,
		&mp.businessGauges,
	)

	// Start uptime counter
	go mp.trackUptime()

	logger.Info("Metrics provider initialized",
		zap.String("service", config.ServiceName),
		zap.String("namespace", config.Namespace),
	)

	return mp
}

// HTTP Metrics Methods

// RecordHTTPRequest records HTTP request metrics
func (mp *MetricsProvider) RecordHTTPRequest(method, endpoint, statusCode, userID string, duration time.Duration, requestSize, responseSize int64) {
	mp.httpRequestsTotal.WithLabelValues(method, endpoint, statusCode, userID).Inc()
	mp.httpRequestDuration.WithLabelValues(method, endpoint, statusCode).Observe(duration.Seconds())
	mp.httpRequestSize.WithLabelValues(method, endpoint).Observe(float64(requestSize))
	mp.httpResponseSize.WithLabelValues(method, endpoint, statusCode).Observe(float64(responseSize))
}

// IncrementActiveConnections increments active connections
func (mp *MetricsProvider) IncrementActiveConnections() {
	mp.httpActiveConnections.Inc()
}

// DecrementActiveConnections decrements active connections
func (mp *MetricsProvider) DecrementActiveConnections() {
	mp.httpActiveConnections.Dec()
}

// Service Metrics Methods

// RecordServiceError records service error
func (mp *MetricsProvider) RecordServiceError(errorType, component, severity string) {
	mp.serviceErrorsTotal.WithLabelValues(errorType, component, severity).Inc()
}

// RecordServiceOperation records service operation
func (mp *MetricsProvider) RecordServiceOperation(operation, status, component string) {
	mp.serviceOperationsTotal.WithLabelValues(operation, status, component).Inc()
}

// SetHealthScore sets service health score
func (mp *MetricsProvider) SetHealthScore(score float64) {
	mp.serviceHealthScore.Set(score)
}

// Database Metrics Methods

// RecordDBQuery records database query metrics
func (mp *MetricsProvider) RecordDBQuery(queryType, table, status string, duration time.Duration) {
	mp.dbQueriesTotal.WithLabelValues(queryType, table, status).Inc()
	mp.dbQueryDuration.WithLabelValues(queryType, table, status).Observe(duration.Seconds())
}

// SetDBConnections sets database connection counts
func (mp *MetricsProvider) SetDBConnections(active, idle int) {
	mp.dbConnectionsActive.Set(float64(active))
	mp.dbConnectionsIdle.Set(float64(idle))
}

// External Service Metrics Methods

// RecordExternalServiceCall records external service call metrics
func (mp *MetricsProvider) RecordExternalServiceCall(service, method, endpoint, statusCode string, duration time.Duration) {
	mp.externalServiceCalls.WithLabelValues(service, method, endpoint, statusCode).Inc()
	mp.externalServiceDuration.WithLabelValues(service, method, endpoint).Observe(duration.Seconds())
}

// Business Metrics Methods

// RecordBusinessEvent records business event
func (mp *MetricsProvider) RecordBusinessEvent(eventType, category, status string) {
	mp.businessMetrics.WithLabelValues(eventType, category, status).Inc()
}

// SetBusinessMetric sets business metric gauge
func (mp *MetricsProvider) SetBusinessMetric(metricName, category string, value float64) {
	mp.businessGauges.WithLabelValues(metricName, category).Set(value)
}

// trackUptime increments uptime counter every second
func (mp *MetricsProvider) trackUptime() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		mp.serviceUptime.Inc()
	}
}

// Handler returns Prometheus HTTP handler
func (mp *MetricsProvider) Handler() http.Handler {
	return promhttp.HandlerFor(&mp.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// DefaultMetricsConfig returns default metrics configuration
func DefaultMetricsConfig(serviceName string) MetricsConfig {
	return MetricsConfig{
		ServiceName: serviceName,
		Environment: getEnv("ENVIRONMENT", "development"),
		Version:     getEnv("SERVICE_VERSION", "1.0.0"),
		Namespace:   "xform",
	}
}
