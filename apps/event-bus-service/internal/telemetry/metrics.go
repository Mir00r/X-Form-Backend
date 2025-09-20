// Package telemetry provides comprehensive observability implementation
// This file implements Prometheus metrics collection for the Event Bus Service
// following enterprise best practices for metrics instrumentation.
package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/config"
)

// MetricsProvider manages Prometheus metrics collection
type MetricsProvider struct {
	config   *config.Config
	logger   *zap.Logger
	registry *prometheus.Registry
	server   *http.Server

	// Application metrics
	requestsTotal     prometheus.CounterVec
	requestDuration   prometheus.HistogramVec
	activeConnections prometheus.Gauge

	// Event Bus specific metrics
	eventsProduced        prometheus.CounterVec
	eventsConsumed        prometheus.CounterVec
	eventProcessingTime   prometheus.HistogramVec
	eventProcessingErrors prometheus.CounterVec
	kafkaOperations       prometheus.CounterVec
	kafkaConnectionStatus prometheus.GaugeVec

	// CDC specific metrics
	debeziumConnectorStatus prometheus.GaugeVec
	cdcEventsProcessed      prometheus.CounterVec
	cdcLagSeconds           prometheus.GaugeVec

	// System metrics
	goRoutines  prometheus.Gauge
	memoryUsage prometheus.GaugeVec
	cpuUsage    prometheus.Gauge
	diskUsage   prometheus.GaugeVec

	// Business metrics
	formsProcessed     prometheus.CounterVec
	responsesProcessed prometheus.CounterVec
	analyticsEvents    prometheus.CounterVec
	errorsByType       prometheus.CounterVec
}

// MetricsConfig defines metrics configuration
type MetricsConfig struct {
	Namespace   string            `json:"namespace"`
	Subsystem   string            `json:"subsystem"`
	Port        int               `json:"port"`
	Path        string            `json:"path"`
	Labels      map[string]string `json:"labels"`
	Buckets     []float64         `json:"buckets"`
	Percentiles []float64         `json:"percentiles"`
}

// NewMetricsProvider creates a new Prometheus metrics provider
func NewMetricsProvider(cfg *config.Config, logger *zap.Logger) (*MetricsProvider, error) {
	mp := &MetricsProvider{
		config:   cfg,
		logger:   logger,
		registry: prometheus.NewRegistry(),
	}

	if err := mp.initializeMetrics(); err != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	if err := mp.setupMetricsServer(); err != nil {
		return nil, fmt.Errorf("failed to setup metrics server: %w", err)
	}

	return mp, nil
}

// initializeMetrics initializes all Prometheus metrics
func (mp *MetricsProvider) initializeMetrics() error {
	// Define metric labels
	commonLabels := []string{"service", "version", "environment"}
	httpLabels := append(commonLabels, "method", "endpoint", "status_code")
	eventLabels := append(commonLabels, "event_type", "topic", "source")
	kafkaLabels := append(commonLabels, "operation", "topic", "partition")
	cdcLabels := append(commonLabels, "connector", "table", "operation")

	// HTTP metrics
	mp.requestsTotal = *prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "event_bus",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests processed",
		},
		httpLabels,
	)

	mp.requestDuration = *prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "event_bus",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		httpLabels,
	)

	mp.activeConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "event_bus",
			Subsystem: "http",
			Name:      "active_connections",
			Help:      "Number of active HTTP connections",
		},
	)

	// Event processing metrics
	mp.eventsProduced = *prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "event_bus",
			Subsystem: "events",
			Name:      "produced_total",
			Help:      "Total number of events produced",
		},
		eventLabels,
	)

	mp.eventsConsumed = *prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "event_bus",
			Subsystem: "events",
			Name:      "consumed_total",
			Help:      "Total number of events consumed",
		},
		eventLabels,
	)

	mp.eventProcessingTime = *prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "event_bus",
			Subsystem: "events",
			Name:      "processing_duration_seconds",
			Help:      "Event processing duration in seconds",
			Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
		eventLabels,
	)

	mp.eventProcessingErrors = *prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "event_bus",
			Subsystem: "events",
			Name:      "processing_errors_total",
			Help:      "Total number of event processing errors",
		},
		append(eventLabels, "error_type"),
	)

	// Kafka metrics
	mp.kafkaOperations = *prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "event_bus",
			Subsystem: "kafka",
			Name:      "operations_total",
			Help:      "Total number of Kafka operations",
		},
		kafkaLabels,
	)

	mp.kafkaConnectionStatus = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "event_bus",
			Subsystem: "kafka",
			Name:      "connection_status",
			Help:      "Kafka connection status (1 = connected, 0 = disconnected)",
		},
		[]string{"broker", "client_id"},
	)

	// CDC/Debezium metrics
	mp.debeziumConnectorStatus = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "event_bus",
			Subsystem: "cdc",
			Name:      "connector_status",
			Help:      "Debezium connector status (1 = running, 0 = stopped)",
		},
		[]string{"connector_name", "task_id"},
	)

	mp.cdcEventsProcessed = *prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "event_bus",
			Subsystem: "cdc",
			Name:      "events_processed_total",
			Help:      "Total number of CDC events processed",
		},
		cdcLabels,
	)

	mp.cdcLagSeconds = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "event_bus",
			Subsystem: "cdc",
			Name:      "lag_seconds",
			Help:      "CDC lag in seconds",
		},
		[]string{"connector", "table"},
	)

	// System metrics
	mp.goRoutines = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "event_bus",
			Subsystem: "system",
			Name:      "goroutines",
			Help:      "Number of goroutines",
		},
	)

	mp.memoryUsage = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "event_bus",
			Subsystem: "system",
			Name:      "memory_usage_bytes",
			Help:      "Memory usage in bytes",
		},
		[]string{"type"}, // heap, stack, etc.
	)

	mp.cpuUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "event_bus",
			Subsystem: "system",
			Name:      "cpu_usage_percent",
			Help:      "CPU usage percentage",
		},
	)

	mp.diskUsage = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "event_bus",
			Subsystem: "system",
			Name:      "disk_usage_bytes",
			Help:      "Disk usage in bytes",
		},
		[]string{"mount_point", "type"}, // used, available
	)

	// Business metrics
	mp.formsProcessed = *prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "event_bus",
			Subsystem: "business",
			Name:      "forms_processed_total",
			Help:      "Total number of forms processed",
		},
		[]string{"service", "operation", "status"},
	)

	mp.responsesProcessed = *prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "event_bus",
			Subsystem: "business",
			Name:      "responses_processed_total",
			Help:      "Total number of responses processed",
		},
		[]string{"service", "form_id", "status"},
	)

	mp.analyticsEvents = *prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "event_bus",
			Subsystem: "business",
			Name:      "analytics_events_total",
			Help:      "Total number of analytics events processed",
		},
		[]string{"event_type", "source"},
	)

	mp.errorsByType = *prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "event_bus",
			Subsystem: "errors",
			Name:      "total",
			Help:      "Total number of errors by type",
		},
		[]string{"error_type", "component", "severity"},
	)

	// Register all metrics with the registry
	mp.registerMetrics()

	mp.logger.Info("Prometheus metrics initialized successfully")
	return nil
}

// registerMetrics registers all metrics with the Prometheus registry
func (mp *MetricsProvider) registerMetrics() {
	// HTTP metrics
	mp.registry.MustRegister(&mp.requestsTotal)
	mp.registry.MustRegister(&mp.requestDuration)
	mp.registry.MustRegister(mp.activeConnections)

	// Event metrics
	mp.registry.MustRegister(&mp.eventsProduced)
	mp.registry.MustRegister(&mp.eventsConsumed)
	mp.registry.MustRegister(&mp.eventProcessingTime)
	mp.registry.MustRegister(&mp.eventProcessingErrors)

	// Kafka metrics
	mp.registry.MustRegister(&mp.kafkaOperations)
	mp.registry.MustRegister(&mp.kafkaConnectionStatus)

	// CDC metrics
	mp.registry.MustRegister(&mp.debeziumConnectorStatus)
	mp.registry.MustRegister(&mp.cdcEventsProcessed)
	mp.registry.MustRegister(&mp.cdcLagSeconds)

	// System metrics
	mp.registry.MustRegister(mp.goRoutines)
	mp.registry.MustRegister(&mp.memoryUsage)
	mp.registry.MustRegister(mp.cpuUsage)
	mp.registry.MustRegister(&mp.diskUsage)

	// Business metrics
	mp.registry.MustRegister(&mp.formsProcessed)
	mp.registry.MustRegister(&mp.responsesProcessed)
	mp.registry.MustRegister(&mp.analyticsEvents)
	mp.registry.MustRegister(&mp.errorsByType)

	// Register Go runtime metrics
	mp.registry.MustRegister(prometheus.NewGoCollector())
	mp.registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
}

// setupMetricsServer sets up the HTTP server for metrics exposition
func (mp *MetricsProvider) setupMetricsServer() error {
	mux := http.NewServeMux()

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.HandlerFor(mp.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))

	// Health check endpoint for metrics server
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := mp.config.Observability.Metrics.Port
	mp.server = &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return nil
}

// Handler returns the Prometheus metrics handler
func (mp *MetricsProvider) Handler() http.Handler {
	return promhttp.HandlerFor(mp.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// Start starts the metrics server
func (mp *MetricsProvider) Start() error {
	go func() {
		mp.logger.Info("Starting metrics server",
			zap.String("addr", mp.server.Addr))

		if err := mp.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			mp.logger.Error("Metrics server error", zap.Error(err))
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the metrics server
func (mp *MetricsProvider) Shutdown(ctx context.Context) error {
	if mp.server == nil {
		return nil
	}

	mp.logger.Info("Shutting down metrics server")
	return mp.server.Shutdown(ctx)
}

// HTTP Metrics Methods

// RecordHTTPRequest records HTTP request metrics
func (mp *MetricsProvider) RecordHTTPRequest(method, endpoint string, statusCode int, duration time.Duration) {
	labels := prometheus.Labels{
		"service":     "event-bus-service",
		"version":     mp.config.Version,
		"environment": mp.config.Environment,
		"method":      method,
		"endpoint":    endpoint,
		"status_code": strconv.Itoa(statusCode),
	}

	mp.requestsTotal.With(labels).Inc()
	mp.requestDuration.With(labels).Observe(duration.Seconds())
}

// SetActiveConnections sets the number of active HTTP connections
func (mp *MetricsProvider) SetActiveConnections(count int) {
	mp.activeConnections.Set(float64(count))
}

// Event Metrics Methods

// RecordEventProduced records metrics for produced events
func (mp *MetricsProvider) RecordEventProduced(eventType, topic, source string) {
	labels := prometheus.Labels{
		"service":     "event-bus-service",
		"version":     mp.config.Version,
		"environment": mp.config.Environment,
		"event_type":  eventType,
		"topic":       topic,
		"source":      source,
	}

	mp.eventsProduced.With(labels).Inc()
}

// RecordEventConsumed records metrics for consumed events
func (mp *MetricsProvider) RecordEventConsumed(eventType, topic, source string) {
	labels := prometheus.Labels{
		"service":     "event-bus-service",
		"version":     mp.config.Version,
		"environment": mp.config.Environment,
		"event_type":  eventType,
		"topic":       topic,
		"source":      source,
	}

	mp.eventsConsumed.With(labels).Inc()
}

// RecordEventProcessingTime records event processing duration
func (mp *MetricsProvider) RecordEventProcessingTime(eventType, topic, source string, duration time.Duration) {
	labels := prometheus.Labels{
		"service":     "event-bus-service",
		"version":     mp.config.Version,
		"environment": mp.config.Environment,
		"event_type":  eventType,
		"topic":       topic,
		"source":      source,
	}

	mp.eventProcessingTime.With(labels).Observe(duration.Seconds())
}

// RecordEventProcessingError records event processing errors
func (mp *MetricsProvider) RecordEventProcessingError(eventType, topic, source, errorType string) {
	labels := prometheus.Labels{
		"service":     "event-bus-service",
		"version":     mp.config.Version,
		"environment": mp.config.Environment,
		"event_type":  eventType,
		"topic":       topic,
		"source":      source,
		"error_type":  errorType,
	}

	mp.eventProcessingErrors.With(labels).Inc()
}

// Kafka Metrics Methods

// RecordKafkaOperation records Kafka operation metrics
func (mp *MetricsProvider) RecordKafkaOperation(operation, topic string, partition int32) {
	labels := prometheus.Labels{
		"service":     "event-bus-service",
		"version":     mp.config.Version,
		"environment": mp.config.Environment,
		"operation":   operation,
		"topic":       topic,
		"partition":   strconv.Itoa(int(partition)),
	}

	mp.kafkaOperations.With(labels).Inc()
}

// SetKafkaConnectionStatus sets Kafka connection status
func (mp *MetricsProvider) SetKafkaConnectionStatus(broker, clientID string, connected bool) {
	labels := prometheus.Labels{
		"broker":    broker,
		"client_id": clientID,
	}

	status := 0.0
	if connected {
		status = 1.0
	}

	mp.kafkaConnectionStatus.With(labels).Set(status)
}

// CDC Metrics Methods

// SetDebeziumConnectorStatus sets Debezium connector status
func (mp *MetricsProvider) SetDebeziumConnectorStatus(connectorName, taskID string, running bool) {
	labels := prometheus.Labels{
		"connector_name": connectorName,
		"task_id":        taskID,
	}

	status := 0.0
	if running {
		status = 1.0
	}

	mp.debeziumConnectorStatus.With(labels).Set(status)
}

// RecordCDCEvent records CDC event processing
func (mp *MetricsProvider) RecordCDCEvent(connector, table, operation string) {
	labels := prometheus.Labels{
		"service":     "event-bus-service",
		"version":     mp.config.Version,
		"environment": mp.config.Environment,
		"connector":   connector,
		"table":       table,
		"operation":   operation,
	}

	mp.cdcEventsProcessed.With(labels).Inc()
}

// SetCDCLag sets CDC lag in seconds
func (mp *MetricsProvider) SetCDCLag(connector, table string, lagSeconds float64) {
	labels := prometheus.Labels{
		"connector": connector,
		"table":     table,
	}

	mp.cdcLagSeconds.With(labels).Set(lagSeconds)
}

// System Metrics Methods

// UpdateSystemMetrics updates system resource metrics
func (mp *MetricsProvider) UpdateSystemMetrics() {
	// Get runtime memory statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Get number of goroutines
	numGoroutines := runtime.NumGoroutine()

	mp.goRoutines.Set(float64(numGoroutines))
	mp.memoryUsage.WithLabelValues("heap").Set(float64(memStats.HeapAlloc))
	mp.memoryUsage.WithLabelValues("stack").Set(float64(memStats.StackInuse))
	mp.memoryUsage.WithLabelValues("total").Set(float64(memStats.Sys))
}

// Business Metrics Methods

// RecordFormProcessed records form processing metrics
func (mp *MetricsProvider) RecordFormProcessed(service, operation, status string) {
	labels := prometheus.Labels{
		"service":   service,
		"operation": operation,
		"status":    status,
	}

	mp.formsProcessed.With(labels).Inc()
}

// RecordResponseProcessed records response processing metrics
func (mp *MetricsProvider) RecordResponseProcessed(service, formID, status string) {
	labels := prometheus.Labels{
		"service": service,
		"form_id": formID,
		"status":  status,
	}

	mp.responsesProcessed.With(labels).Inc()
}

// RecordAnalyticsEvent records analytics event metrics
func (mp *MetricsProvider) RecordAnalyticsEvent(eventType, source string) {
	labels := prometheus.Labels{
		"event_type": eventType,
		"source":     source,
	}

	mp.analyticsEvents.With(labels).Inc()
}

// RecordError records error metrics by type
func (mp *MetricsProvider) RecordError(errorType, component, severity string) {
	labels := prometheus.Labels{
		"error_type": errorType,
		"component":  component,
		"severity":   severity,
	}

	mp.errorsByType.With(labels).Inc()
}

// HTTPMiddleware provides HTTP metrics middleware
func (mp *MetricsProvider) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Increment active connections
		mp.activeConnections.Inc()
		defer mp.activeConnections.Dec()

		// Create response writer wrapper to capture status code
		wrappedWriter := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Execute request
		next.ServeHTTP(wrappedWriter, r)

		// Record metrics
		duration := time.Since(start)
		mp.RecordHTTPRequest(r.Method, r.URL.Path, wrappedWriter.statusCode, duration)
	})
}

// StartSystemMetricsCollection starts periodic system metrics collection
func (mp *MetricsProvider) StartSystemMetricsCollection(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				mp.UpdateSystemMetrics()
			}
		}
	}()
}
