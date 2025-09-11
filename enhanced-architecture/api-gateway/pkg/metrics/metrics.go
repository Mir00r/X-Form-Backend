// Package metrics provides application metrics collection and reporting
// Implements Prometheus metrics following monitoring best practices
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Collector holds all application metrics
type Collector struct {
	// HTTP metrics
	RequestsTotal    *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	ResponseSize     *prometheus.HistogramVec
	RequestsInFlight prometheus.Gauge

	// Authentication metrics
	AuthAttempts     *prometheus.CounterVec
	AuthDuration     *prometheus.HistogramVec
	TokenValidations *prometheus.CounterVec

	// Rate limiting metrics
	RateLimitHits      *prometheus.CounterVec
	RateLimitRemaining *prometheus.GaugeVec

	// Upstream service metrics
	UpstreamRequests *prometheus.CounterVec
	UpstreamLatency  *prometheus.HistogramVec
	UpstreamErrors   *prometheus.CounterVec

	// Circuit breaker metrics
	CircuitBreakerState *prometheus.GaugeVec
	CircuitBreakerTrips *prometheus.CounterVec

	// System metrics
	MemoryUsage    prometheus.Gauge
	CPUUsage       prometheus.Gauge
	GoroutineCount prometheus.Gauge

	// Business metrics
	ActiveSessions    prometheus.Gauge
	FormSubmissions   *prometheus.CounterVec
	UserRegistrations *prometheus.CounterVec

	// Error metrics
	ErrorsTotal *prometheus.CounterVec
	PanicsTotal *prometheus.CounterVec

	registry *prometheus.Registry
}

// Config holds metrics configuration
type Config struct {
	// Enabled controls whether metrics collection is enabled
	Enabled bool `json:"enabled"`

	// Path is the HTTP path for metrics endpoint
	Path string `json:"path"`

	// Port is the port to serve metrics on (0 = use main server port)
	Port int `json:"port"`

	// Namespace is the prefix for all metric names
	Namespace string `json:"namespace"`

	// Subsystem is the subsystem name for metrics
	Subsystem string `json:"subsystem"`

	// Labels are additional labels to add to all metrics
	Labels map[string]string `json:"labels"`

	// HistogramBuckets defines custom buckets for histograms
	HistogramBuckets []float64 `json:"histogram_buckets"`

	// EnableGoMetrics includes Go runtime metrics
	EnableGoMetrics bool `json:"enable_go_metrics"`

	// EnableProcessMetrics includes process metrics
	EnableProcessMetrics bool `json:"enable_process_metrics"`
}

// NewCollector creates a new metrics collector
func NewCollector(config Config) *Collector {
	// Create custom registry if needed
	var registry *prometheus.Registry
	if config.Enabled {
		registry = prometheus.NewRegistry()
	} else {
		registry = prometheus.NewRegistry()
	}

	// Default histogram buckets for request duration (in seconds)
	histogramBuckets := config.HistogramBuckets
	if len(histogramBuckets) == 0 {
		histogramBuckets = []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}
	}

	// Size buckets for response size (in bytes)
	sizeBuckets := []float64{100, 1000, 10000, 100000, 1000000, 10000000}

	collector := &Collector{
		registry: registry,

		// HTTP metrics
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status_code"},
		),

		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   histogramBuckets,
			},
			[]string{"method", "path", "status_code"},
		),

		ResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "http_response_size_bytes",
				Help:      "HTTP response size in bytes",
				Buckets:   sizeBuckets,
			},
			[]string{"method", "path", "status_code"},
		),

		RequestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "http_requests_in_flight",
				Help:      "Current number of HTTP requests being processed",
			},
		),

		// Authentication metrics
		AuthAttempts: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "auth_attempts_total",
				Help:      "Total number of authentication attempts",
			},
			[]string{"type", "result"},
		),

		AuthDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "auth_duration_seconds",
				Help:      "Authentication operation duration in seconds",
				Buckets:   histogramBuckets,
			},
			[]string{"type"},
		),

		TokenValidations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "token_validations_total",
				Help:      "Total number of token validations",
			},
			[]string{"result"},
		),

		// Rate limiting metrics
		RateLimitHits: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "rate_limit_hits_total",
				Help:      "Total number of rate limit hits",
			},
			[]string{"client_type"},
		),

		RateLimitRemaining: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "rate_limit_remaining",
				Help:      "Remaining rate limit quota",
			},
			[]string{"client_id"},
		),

		// Upstream service metrics
		UpstreamRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "upstream_requests_total",
				Help:      "Total number of upstream service requests",
			},
			[]string{"service", "method", "status_code"},
		),

		UpstreamLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "upstream_latency_seconds",
				Help:      "Upstream service request latency in seconds",
				Buckets:   histogramBuckets,
			},
			[]string{"service", "method"},
		),

		UpstreamErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "upstream_errors_total",
				Help:      "Total number of upstream service errors",
			},
			[]string{"service", "error_type"},
		),

		// Circuit breaker metrics
		CircuitBreakerState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "circuit_breaker_state",
				Help:      "Circuit breaker state (0=closed, 1=open, 2=half-open)",
			},
			[]string{"service"},
		),

		CircuitBreakerTrips: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "circuit_breaker_trips_total",
				Help:      "Total number of circuit breaker trips",
			},
			[]string{"service"},
		),

		// System metrics
		MemoryUsage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "memory_usage_bytes",
				Help:      "Current memory usage in bytes",
			},
		),

		CPUUsage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "cpu_usage_percent",
				Help:      "Current CPU usage percentage",
			},
		),

		GoroutineCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "goroutines_count",
				Help:      "Current number of goroutines",
			},
		),

		// Business metrics
		ActiveSessions: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "active_sessions",
				Help:      "Current number of active sessions",
			},
		),

		FormSubmissions: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "form_submissions_total",
				Help:      "Total number of form submissions",
			},
			[]string{"form_type", "status"},
		),

		UserRegistrations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "user_registrations_total",
				Help:      "Total number of user registrations",
			},
			[]string{"source"},
		),

		// Error metrics
		ErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "errors_total",
				Help:      "Total number of errors",
			},
			[]string{"type", "component"},
		),

		PanicsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "panics_total",
				Help:      "Total number of panics",
			},
			[]string{"component"},
		),
	}

	// Register metrics
	if config.Enabled {
		collector.registerMetrics(config)
	}

	return collector
}

// registerMetrics registers all metrics with the registry
func (c *Collector) registerMetrics(config Config) {
	// Register HTTP metrics
	c.registry.MustRegister(c.RequestsTotal)
	c.registry.MustRegister(c.RequestDuration)
	c.registry.MustRegister(c.ResponseSize)
	c.registry.MustRegister(c.RequestsInFlight)

	// Register authentication metrics
	c.registry.MustRegister(c.AuthAttempts)
	c.registry.MustRegister(c.AuthDuration)
	c.registry.MustRegister(c.TokenValidations)

	// Register rate limiting metrics
	c.registry.MustRegister(c.RateLimitHits)
	c.registry.MustRegister(c.RateLimitRemaining)

	// Register upstream metrics
	c.registry.MustRegister(c.UpstreamRequests)
	c.registry.MustRegister(c.UpstreamLatency)
	c.registry.MustRegister(c.UpstreamErrors)

	// Register circuit breaker metrics
	c.registry.MustRegister(c.CircuitBreakerState)
	c.registry.MustRegister(c.CircuitBreakerTrips)

	// Register system metrics
	c.registry.MustRegister(c.MemoryUsage)
	c.registry.MustRegister(c.CPUUsage)
	c.registry.MustRegister(c.GoroutineCount)

	// Register business metrics
	c.registry.MustRegister(c.ActiveSessions)
	c.registry.MustRegister(c.FormSubmissions)
	c.registry.MustRegister(c.UserRegistrations)

	// Register error metrics
	c.registry.MustRegister(c.ErrorsTotal)
	c.registry.MustRegister(c.PanicsTotal)

	// Register Go metrics if enabled
	if config.EnableGoMetrics {
		c.registry.MustRegister(prometheus.NewGoCollector())
	}

	// Register process metrics if enabled
	if config.EnableProcessMetrics {
		c.registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	}
}

// Handler returns the HTTP handler for metrics endpoint
func (c *Collector) Handler() http.Handler {
	return promhttp.HandlerFor(c.registry, promhttp.HandlerOpts{
		Registry:          c.registry,
		EnableOpenMetrics: true,
	})
}

// RecordHTTPRequest records HTTP request metrics
func (c *Collector) RecordHTTPRequest(method, path string, statusCode int, duration time.Duration, responseSize int64) {
	statusStr := strconv.Itoa(statusCode)

	c.RequestsTotal.WithLabelValues(method, path, statusStr).Inc()
	c.RequestDuration.WithLabelValues(method, path, statusStr).Observe(duration.Seconds())
	c.ResponseSize.WithLabelValues(method, path, statusStr).Observe(float64(responseSize))
}

// IncrementRequestsInFlight increments in-flight requests counter
func (c *Collector) IncrementRequestsInFlight() {
	c.RequestsInFlight.Inc()
}

// DecrementRequestsInFlight decrements in-flight requests counter
func (c *Collector) DecrementRequestsInFlight() {
	c.RequestsInFlight.Dec()
}

// RecordAuthAttempt records authentication attempt
func (c *Collector) RecordAuthAttempt(authType, result string, duration time.Duration) {
	c.AuthAttempts.WithLabelValues(authType, result).Inc()
	c.AuthDuration.WithLabelValues(authType).Observe(duration.Seconds())
}

// RecordTokenValidation records token validation result
func (c *Collector) RecordTokenValidation(result string) {
	c.TokenValidations.WithLabelValues(result).Inc()
}

// RecordRateLimitHit records rate limit hit
func (c *Collector) RecordRateLimitHit(clientType string) {
	c.RateLimitHits.WithLabelValues(clientType).Inc()
}

// SetRateLimitRemaining sets remaining rate limit quota
func (c *Collector) SetRateLimitRemaining(clientID string, remaining float64) {
	c.RateLimitRemaining.WithLabelValues(clientID).Set(remaining)
}

// RecordUpstreamRequest records upstream service request
func (c *Collector) RecordUpstreamRequest(service, method string, statusCode int, duration time.Duration) {
	statusStr := strconv.Itoa(statusCode)
	c.UpstreamRequests.WithLabelValues(service, method, statusStr).Inc()
	c.UpstreamLatency.WithLabelValues(service, method).Observe(duration.Seconds())
}

// RecordUpstreamError records upstream service error
func (c *Collector) RecordUpstreamError(service, errorType string) {
	c.UpstreamErrors.WithLabelValues(service, errorType).Inc()
}

// SetCircuitBreakerState sets circuit breaker state
func (c *Collector) SetCircuitBreakerState(service string, state CircuitBreakerState) {
	c.CircuitBreakerState.WithLabelValues(service).Set(float64(state))
}

// RecordCircuitBreakerTrip records circuit breaker trip
func (c *Collector) RecordCircuitBreakerTrip(service string) {
	c.CircuitBreakerTrips.WithLabelValues(service).Inc()
}

// SetMemoryUsage sets current memory usage
func (c *Collector) SetMemoryUsage(bytes float64) {
	c.MemoryUsage.Set(bytes)
}

// SetCPUUsage sets current CPU usage percentage
func (c *Collector) SetCPUUsage(percent float64) {
	c.CPUUsage.Set(percent)
}

// SetGoroutineCount sets current goroutine count
func (c *Collector) SetGoroutineCount(count float64) {
	c.GoroutineCount.Set(count)
}

// SetActiveSessions sets current active sessions count
func (c *Collector) SetActiveSessions(count float64) {
	c.ActiveSessions.Set(count)
}

// RecordFormSubmission records form submission
func (c *Collector) RecordFormSubmission(formType, status string) {
	c.FormSubmissions.WithLabelValues(formType, status).Inc()
}

// RecordUserRegistration records user registration
func (c *Collector) RecordUserRegistration(source string) {
	c.UserRegistrations.WithLabelValues(source).Inc()
}

// RecordError records an error
func (c *Collector) RecordError(errorType, component string) {
	c.ErrorsTotal.WithLabelValues(errorType, component).Inc()
}

// RecordPanic records a panic
func (c *Collector) RecordPanic(component string) {
	c.PanicsTotal.WithLabelValues(component).Inc()
}

// CircuitBreakerState represents circuit breaker states
type CircuitBreakerState int

const (
	// CircuitBreakerClosed indicates the circuit breaker is closed (normal operation)
	CircuitBreakerClosed CircuitBreakerState = 0
	// CircuitBreakerOpen indicates the circuit breaker is open (failing fast)
	CircuitBreakerOpen CircuitBreakerState = 1
	// CircuitBreakerHalfOpen indicates the circuit breaker is half-open (testing)
	CircuitBreakerHalfOpen CircuitBreakerState = 2
)

// DefaultConfig returns default metrics configuration
func DefaultConfig() Config {
	return Config{
		Enabled:              true,
		Path:                 "/metrics",
		Port:                 0, // Use main server port
		Namespace:            "xform",
		Subsystem:            "api_gateway",
		Labels:               make(map[string]string),
		HistogramBuckets:     []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		EnableGoMetrics:      true,
		EnableProcessMetrics: true,
	}
}

// HealthMetrics tracks health check metrics
type HealthMetrics struct {
	ServiceStatus *prometheus.GaugeVec
	HealthChecks  *prometheus.CounterVec
}

// NewHealthMetrics creates new health metrics
func NewHealthMetrics(registry *prometheus.Registry, namespace, subsystem string) *HealthMetrics {
	metrics := &HealthMetrics{
		ServiceStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "service_status",
				Help:      "Service health status (1=healthy, 0=unhealthy)",
			},
			[]string{"service", "check_type"},
		),

		HealthChecks: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "health_checks_total",
				Help:      "Total number of health checks performed",
			},
			[]string{"service", "status"},
		),
	}

	if registry != nil {
		registry.MustRegister(metrics.ServiceStatus)
		registry.MustRegister(metrics.HealthChecks)
	}

	return metrics
}

// SetServiceStatus sets service health status
func (h *HealthMetrics) SetServiceStatus(service, checkType string, healthy bool) {
	status := 0.0
	if healthy {
		status = 1.0
	}
	h.ServiceStatus.WithLabelValues(service, checkType).Set(status)
}

// RecordHealthCheck records health check execution
func (h *HealthMetrics) RecordHealthCheck(service, status string) {
	h.HealthChecks.WithLabelValues(service, status).Inc()
}

// Middleware creates a metrics middleware function
func (c *Collector) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			c.IncrementRequestsInFlight()
			defer c.DecrementRequestsInFlight()

			// Create response writer wrapper to capture status and size
			rw := &responseWriter{ResponseWriter: w, statusCode: 200}

			// Process request
			next.ServeHTTP(rw, r)

			// Record metrics
			duration := time.Since(start)
			c.RecordHTTPRequest(r.Method, r.URL.Path, rw.statusCode, duration, rw.bytesWritten)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture response metrics
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// StartSystemMetricsCollector starts a goroutine to collect system metrics
func (c *Collector) StartSystemMetricsCollector(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			c.collectSystemMetrics()
		}
	}()
}

// collectSystemMetrics collects system metrics
func (c *Collector) collectSystemMetrics() {
	// This is a simplified implementation
	// In production, you would use proper system monitoring libraries

	// Memory usage (placeholder)
	// m := &runtime.MemStats{}
	// runtime.ReadMemStats(m)
	// c.SetMemoryUsage(float64(m.Alloc))

	// Goroutine count
	// c.SetGoroutineCount(float64(runtime.NumGoroutine()))

	// CPU usage would require external library like gopsutil
}
