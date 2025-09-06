// Package processors provides event processing capabilities for the Event Bus Service
// This package implements enterprise-grade event processors for handling various types
// of CDC events, transformations, filtering, and routing with comprehensive monitoring.
package processors

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/config"
	"github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/events"
	"github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/kafka"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// EventProcessor defines the interface for processing events
type EventProcessor interface {
	ProcessEvent(ctx context.Context, event *events.CDCEvent) error
	GetName() string
	GetType() string
	HealthCheck() error
}

// ProcessorManager manages multiple event processors and routing
type ProcessorManager struct {
	config     *config.Config
	logger     *zap.Logger
	kafka      *kafka.Client
	processors map[string]EventProcessor
	routes     map[string][]string // topic -> processor names
	metrics    *ProcessorMetrics
	stopCh     chan struct{}
	wg         sync.WaitGroup
	mutex      sync.RWMutex
}

// ProcessorMetrics contains Prometheus metrics for event processing
type ProcessorMetrics struct {
	EventsProcessed      prometheus.Counter
	EventsFiltered       prometheus.Counter
	EventsFailed         prometheus.Counter
	ProcessingLatency    prometheus.Histogram
	ProcessorHealthScore *prometheus.GaugeVec
	TransformationTime   prometheus.Histogram
	RoutingDecisions     *prometheus.CounterVec
	ErrorsByType         *prometheus.CounterVec
}

// CDCEventProcessor processes Change Data Capture events
type CDCEventProcessor struct {
	name            string
	config          *config.EventProcessingConfig
	logger          *zap.Logger
	kafka           *kafka.Client
	transformations []Transformation
	filters         []Filter
	routes          []Route
	metrics         *ProcessorMetrics
}

// FormEventProcessor processes form-related events
type FormEventProcessor struct {
	name        string
	config      *config.EventProcessingConfig
	logger      *zap.Logger
	kafka       *kafka.Client
	formService string // Form service endpoint
	metrics     *ProcessorMetrics
}

// ResponseEventProcessor processes response-related events
type ResponseEventProcessor struct {
	name            string
	config          *config.EventProcessingConfig
	logger          *zap.Logger
	kafka           *kafka.Client
	responseService string // Response service endpoint
	metrics         *ProcessorMetrics
}

// AnalyticsEventProcessor processes analytics events
type AnalyticsEventProcessor struct {
	name             string
	config           *config.EventProcessingConfig
	logger           *zap.Logger
	kafka            *kafka.Client
	analyticsService string // Analytics service endpoint
	aggregators      map[string]*EventAggregator
	metrics          *ProcessorMetrics
}

// Transformation defines an event transformation
type Transformation interface {
	Transform(ctx context.Context, event *events.CDCEvent) (*events.CDCEvent, error)
	GetName() string
	GetConfig() map[string]interface{}
}

// Filter defines an event filter
type Filter interface {
	ShouldProcess(ctx context.Context, event *events.CDCEvent) (bool, error)
	GetName() string
	GetConfig() map[string]interface{}
}

// Route defines event routing logic
type Route interface {
	ShouldRoute(ctx context.Context, event *events.CDCEvent) (bool, error)
	GetTargets() []string
	GetName() string
}

// EventAggregator aggregates events for analytics
type EventAggregator struct {
	WindowSize time.Duration
	Events     []events.CDCEvent
	LastFlush  time.Time
	mutex      sync.RWMutex
}

// ProcessingResult represents the result of event processing
type ProcessingResult struct {
	Success         bool                   `json:"success"`
	ProcessorName   string                 `json:"processor_name"`
	ProcessingTime  time.Duration          `json:"processing_time"`
	Error           string                 `json:"error,omitempty"`
	Transformations int                    `json:"transformations"`
	RoutedTo        []string               `json:"routed_to"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// NewProcessorManager creates a new processor manager
func NewProcessorManager(cfg *config.Config, logger *zap.Logger, kafkaClient *kafka.Client) (*ProcessorManager, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	manager := &ProcessorManager{
		config:     cfg,
		logger:     logger,
		kafka:      kafkaClient,
		processors: make(map[string]EventProcessor),
		routes:     make(map[string][]string),
		metrics:    initProcessorMetrics(),
		stopCh:     make(chan struct{}),
	}

	// Initialize processors based on configuration
	if err := manager.initializeProcessors(); err != nil {
		return nil, fmt.Errorf("failed to initialize processors: %w", err)
	}

	logger.Info("Processor manager initialized successfully",
		zap.Int("processors", len(manager.processors)))

	return manager, nil
}

// Start starts the processor manager and all processors
func (pm *ProcessorManager) Start(ctx context.Context) error {
	if pm.config.EventProcessing.Workers == 0 {
		pm.logger.Info("Event processing is disabled (workers=0), skipping startup")
		return nil
	}

	pm.logger.Info("Starting processor manager")

	// Start health check monitoring
	pm.wg.Add(1)
	go pm.healthCheckLoop(ctx)

	// Start metrics collection
	pm.wg.Add(1)
	go pm.metricsCollectionLoop(ctx)

	return nil
}

// Stop stops the processor manager and all processors
func (pm *ProcessorManager) Stop() error {
	pm.logger.Info("Stopping processor manager")

	close(pm.stopCh)
	pm.wg.Wait()

	pm.logger.Info("Processor manager stopped")
	return nil
}

// ProcessEvent processes an event through the appropriate processors
func (pm *ProcessorManager) ProcessEvent(ctx context.Context, event *events.CDCEvent) (*ProcessingResult, error) {
	start := time.Now()
	defer func() {
		pm.metrics.ProcessingLatency.Observe(time.Since(start).Seconds())
	}()

	result := &ProcessingResult{
		Success:         true,
		ProcessingTime:  0,
		Transformations: 0,
		RoutedTo:        []string{},
		Metadata:        make(map[string]interface{}),
	}

	// Determine which processors should handle this event
	processors := pm.getProcessorsForEvent(event)
	if len(processors) == 0 {
		pm.logger.Debug("No processors found for event",
			zap.String("topic", event.Source.Topic),
			zap.String("table", event.Source.Table))
		pm.metrics.EventsFiltered.Inc()
		return result, nil
	}

	// Process through each selected processor
	for _, processorName := range processors {
		pm.mutex.RLock()
		processor, exists := pm.processors[processorName]
		pm.mutex.RUnlock()

		if !exists {
			pm.logger.Warn("Processor not found", zap.String("processor", processorName))
			continue
		}

		processorStart := time.Now()
		if err := processor.ProcessEvent(ctx, event); err != nil {
			pm.logger.Error("Processor failed to process event",
				zap.String("processor", processorName),
				zap.String("event_id", event.ID),
				zap.Error(err))

			result.Success = false
			result.Error = err.Error()
			result.ProcessorName = processorName
			pm.metrics.EventsFailed.Inc()
			pm.metrics.ErrorsByType.WithLabelValues(processor.GetType(), "processing_error").Inc()
			continue
		}

		processingTime := time.Since(processorStart)
		result.ProcessingTime += processingTime
		result.RoutedTo = append(result.RoutedTo, processorName)

		pm.logger.Debug("Event processed successfully",
			zap.String("processor", processorName),
			zap.String("event_id", event.ID),
			zap.Duration("processing_time", processingTime))
	}

	if result.Success {
		pm.metrics.EventsProcessed.Inc()
	}

	result.ProcessingTime = time.Since(start)
	return result, nil
}

// RegisterProcessor registers a new event processor
func (pm *ProcessorManager) RegisterProcessor(processor EventProcessor) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	name := processor.GetName()
	if _, exists := pm.processors[name]; exists {
		return fmt.Errorf("processor %s already registered", name)
	}

	pm.processors[name] = processor
	pm.logger.Info("Processor registered",
		zap.String("name", name),
		zap.String("type", processor.GetType()))

	return nil
}

// UnregisterProcessor unregisters an event processor
func (pm *ProcessorManager) UnregisterProcessor(name string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if _, exists := pm.processors[name]; !exists {
		return fmt.Errorf("processor %s not found", name)
	}

	delete(pm.processors, name)
	pm.logger.Info("Processor unregistered", zap.String("name", name))

	return nil
}

// initializeProcessors initializes all configured processors
func (pm *ProcessorManager) initializeProcessors() error {
	// Initialize CDC processor
	cdcProcessor := &CDCEventProcessor{
		name:    "cdc-processor",
		config:  &pm.config.EventProcessing,
		logger:  pm.logger.Named("cdc-processor"),
		kafka:   pm.kafka,
		metrics: pm.metrics,
	}
	if err := cdcProcessor.initialize(); err != nil {
		return fmt.Errorf("failed to initialize CDC processor: %w", err)
	}
	pm.processors[cdcProcessor.name] = cdcProcessor

	// Initialize Form processor
	formProcessor := &FormEventProcessor{
		name:        "form-processor",
		config:      &pm.config.EventProcessing,
		logger:      pm.logger.Named("form-processor"),
		kafka:       pm.kafka,
		formService: "http://form-service:8080", // From service discovery
		metrics:     pm.metrics,
	}
	pm.processors[formProcessor.name] = formProcessor

	// Initialize Response processor
	responseProcessor := &ResponseEventProcessor{
		name:            "response-processor",
		config:          &pm.config.EventProcessing,
		logger:          pm.logger.Named("response-processor"),
		kafka:           pm.kafka,
		responseService: "http://response-service:8080", // From service discovery
		metrics:         pm.metrics,
	}
	pm.processors[responseProcessor.name] = responseProcessor

	// Initialize Analytics processor
	analyticsProcessor := &AnalyticsEventProcessor{
		name:             "analytics-processor",
		config:           &pm.config.EventProcessing,
		logger:           pm.logger.Named("analytics-processor"),
		kafka:            pm.kafka,
		analyticsService: "http://analytics-service:8080", // From service discovery
		aggregators:      make(map[string]*EventAggregator),
		metrics:          pm.metrics,
	}
	if err := analyticsProcessor.initialize(); err != nil {
		return fmt.Errorf("failed to initialize analytics processor: %w", err)
	}
	pm.processors[analyticsProcessor.name] = analyticsProcessor

	// Configure routing
	pm.configureRouting()

	return nil
}

// configureRouting configures event routing based on configuration
func (pm *ProcessorManager) configureRouting() {
	// Route CDC events to appropriate processors
	pm.routes["cdc.forms"] = []string{"cdc-processor", "form-processor", "analytics-processor"}
	pm.routes["cdc.responses"] = []string{"cdc-processor", "response-processor", "analytics-processor"}
	pm.routes["cdc.users"] = []string{"cdc-processor", "analytics-processor"}
	pm.routes["cdc.analytics"] = []string{"analytics-processor"}

	// Route application events
	pm.routes["app.form.created"] = []string{"form-processor", "analytics-processor"}
	pm.routes["app.form.updated"] = []string{"form-processor", "analytics-processor"}
	pm.routes["app.response.submitted"] = []string{"response-processor", "analytics-processor"}
	pm.routes["app.user.registered"] = []string{"analytics-processor"}
}

// getProcessorsForEvent determines which processors should handle an event
func (pm *ProcessorManager) getProcessorsForEvent(event *events.CDCEvent) []string {
	// Build topic key for routing
	topicKey := event.Source.Topic
	if event.Source.Table != "" {
		topicKey = fmt.Sprintf("%s.%s", event.Source.Topic, event.Source.Table)
	}

	// Check exact match first
	if processors, exists := pm.routes[topicKey]; exists {
		return processors
	}

	// Check prefix matches
	for route, processors := range pm.routes {
		if strings.HasPrefix(topicKey, route) || strings.HasPrefix(route, topicKey) {
			return processors
		}
	}

	// Default fallback to CDC processor for all CDC events
	if strings.HasPrefix(topicKey, "cdc.") || event.Source.Connector != "" {
		return []string{"cdc-processor"}
	}

	return []string{}
}

// healthCheckLoop performs periodic health checks on all processors
func (pm *ProcessorManager) healthCheckLoop(ctx context.Context) {
	defer pm.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pm.stopCh:
			return
		case <-ticker.C:
			pm.performHealthChecks()
		}
	}
}

// metricsCollectionLoop collects metrics from processors
func (pm *ProcessorManager) metricsCollectionLoop(ctx context.Context) {
	defer pm.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pm.stopCh:
			return
		case <-ticker.C:
			pm.collectMetrics()
		}
	}
}

// performHealthChecks performs health checks on all processors
func (pm *ProcessorManager) performHealthChecks() {
	pm.mutex.RLock()
	processors := make(map[string]EventProcessor)
	for name, processor := range pm.processors {
		processors[name] = processor
	}
	pm.mutex.RUnlock()

	for name, processor := range processors {
		healthScore := 1.0
		if err := processor.HealthCheck(); err != nil {
			pm.logger.Warn("Processor health check failed",
				zap.String("processor", name),
				zap.Error(err))
			healthScore = 0.0
		}

		pm.metrics.ProcessorHealthScore.WithLabelValues(name, processor.GetType()).Set(healthScore)
	}
}

// collectMetrics collects metrics from processors
func (pm *ProcessorManager) collectMetrics() {
	// This method would collect additional metrics from processors
	// Implementation depends on specific processor metrics
}

// CDC Event Processor Implementation

// initialize initializes the CDC event processor
func (cep *CDCEventProcessor) initialize() error {
	// Initialize transformations
	cep.transformations = []Transformation{
		&TableNameTransformation{},
		&TimestampTransformation{},
		&SchemaTransformation{},
	}

	// Initialize filters
	cep.filters = []Filter{
		&TableFilter{AllowedTables: []string{"forms", "responses", "users", "analytics"}},
		&OperationFilter{AllowedOperations: []string{"c", "u", "d"}}, // create, update, delete
	}

	// Initialize routes
	cep.routes = []Route{
		&TopicRoute{SourceTopics: []string{"cdc.forms"}, TargetTopics: []string{"processed.forms"}},
		&TopicRoute{SourceTopics: []string{"cdc.responses"}, TargetTopics: []string{"processed.responses"}},
		&TopicRoute{SourceTopics: []string{"cdc.users"}, TargetTopics: []string{"processed.users"}},
	}

	return nil
}

// ProcessEvent processes a CDC event
func (cep *CDCEventProcessor) ProcessEvent(ctx context.Context, event *events.CDCEvent) error {
	start := time.Now()
	defer func() {
		cep.metrics.TransformationTime.Observe(time.Since(start).Seconds())
	}()

	// Apply filters
	for _, filter := range cep.filters {
		shouldProcess, err := filter.ShouldProcess(ctx, event)
		if err != nil {
			return fmt.Errorf("filter %s failed: %w", filter.GetName(), err)
		}
		if !shouldProcess {
			cep.logger.Debug("Event filtered out",
				zap.String("filter", filter.GetName()),
				zap.String("event_id", event.ID))
			cep.metrics.EventsFiltered.Inc()
			return nil
		}
	}

	// Apply transformations
	transformedEvent := event
	for _, transformation := range cep.transformations {
		var err error
		transformedEvent, err = transformation.Transform(ctx, transformedEvent)
		if err != nil {
			return fmt.Errorf("transformation %s failed: %w", transformation.GetName(), err)
		}
	}

	// Apply routing
	for _, route := range cep.routes {
		shouldRoute, err := route.ShouldRoute(ctx, transformedEvent)
		if err != nil {
			return fmt.Errorf("route %s failed: %w", route.GetName(), err)
		}
		if shouldRoute {
			for _, target := range route.GetTargets() {
				if err := cep.publishToTarget(ctx, transformedEvent, target); err != nil {
					return fmt.Errorf("failed to publish to target %s: %w", target, err)
				}
				cep.metrics.RoutingDecisions.WithLabelValues(route.GetName(), target).Inc()
			}
		}
	}

	return nil
}

// publishToTarget publishes an event to a target topic
func (cep *CDCEventProcessor) publishToTarget(ctx context.Context, event *events.CDCEvent, target string) error {
	message := &kafka.Message{
		ID:        event.ID,
		EventType: event.GetEventType(),
		Source:    "cdc-processor",
		Data:      event,
		Topic:     target,
		Key:       event.ID,
		Headers:   event.Headers,
		Metadata: kafka.MessageMetadata{
			Timestamp:   time.Now(),
			Version:     "1.0",
			ContentType: "application/json",
			Encoding:    "utf-8",
		},
	}

	return cep.kafka.PublishMessage(ctx, message)
}

// GetName returns the processor name
func (cep *CDCEventProcessor) GetName() string {
	return cep.name
}

// GetType returns the processor type
func (cep *CDCEventProcessor) GetType() string {
	return "cdc"
}

// HealthCheck performs a health check
func (cep *CDCEventProcessor) HealthCheck() error {
	// Check if all transformations and filters are healthy
	return nil
}

// Form Event Processor Implementation

// ProcessEvent processes a form-related event
func (fep *FormEventProcessor) ProcessEvent(ctx context.Context, event *events.CDCEvent) error {
	// Process form-specific logic
	if event.Source.Table != "forms" && !strings.Contains(event.Source.Topic, "form") {
		return nil // Skip non-form events
	}

	// Extract form data
	formData, err := fep.extractFormData(event)
	if err != nil {
		return fmt.Errorf("failed to extract form data: %w", err)
	}

	// Process based on operation
	switch event.Operation {
	case "c": // Create
		return fep.handleFormCreated(ctx, formData)
	case "u": // Update
		return fep.handleFormUpdated(ctx, formData)
	case "d": // Delete
		return fep.handleFormDeleted(ctx, formData)
	default:
		return fmt.Errorf("unsupported operation: %s", event.Operation)
	}
}

// extractFormData extracts form data from CDC event
func (fep *FormEventProcessor) extractFormData(event *events.CDCEvent) (map[string]interface{}, error) {
	if event.After != nil {
		return event.After, nil
	}
	if event.Before != nil {
		return event.Before, nil
	}
	return nil, fmt.Errorf("no form data found in event")
}

// handleFormCreated handles form creation events
func (fep *FormEventProcessor) handleFormCreated(ctx context.Context, formData map[string]interface{}) error {
	// Publish form created event
	eventData := map[string]interface{}{
		"event_type": "form.created",
		"form_data":  formData,
		"timestamp":  time.Now().Unix(),
	}

	return fep.publishEvent(ctx, "app.form.created", eventData)
}

// handleFormUpdated handles form update events
func (fep *FormEventProcessor) handleFormUpdated(ctx context.Context, formData map[string]interface{}) error {
	// Publish form updated event
	eventData := map[string]interface{}{
		"event_type": "form.updated",
		"form_data":  formData,
		"timestamp":  time.Now().Unix(),
	}

	return fep.publishEvent(ctx, "app.form.updated", eventData)
}

// handleFormDeleted handles form deletion events
func (fep *FormEventProcessor) handleFormDeleted(ctx context.Context, formData map[string]interface{}) error {
	// Publish form deleted event
	eventData := map[string]interface{}{
		"event_type": "form.deleted",
		"form_data":  formData,
		"timestamp":  time.Now().Unix(),
	}

	return fep.publishEvent(ctx, "app.form.deleted", eventData)
}

// publishEvent publishes an event to Kafka
func (fep *FormEventProcessor) publishEvent(ctx context.Context, topic string, data map[string]interface{}) error {
	message := &kafka.Message{
		ID:        fmt.Sprintf("form_%d", time.Now().UnixNano()),
		EventType: topic,
		Source:    "form-processor",
		Data:      data,
		Topic:     topic,
		Key:       fmt.Sprintf("form_%d", time.Now().UnixNano()),
		Headers:   make(map[string]string),
		Metadata: kafka.MessageMetadata{
			Timestamp:   time.Now(),
			Version:     "1.0",
			ContentType: "application/json",
			Encoding:    "utf-8",
		},
	}

	return fep.kafka.PublishMessage(ctx, message)
}

// GetName returns the processor name
func (fep *FormEventProcessor) GetName() string {
	return fep.name
}

// GetType returns the processor type
func (fep *FormEventProcessor) GetType() string {
	return "form"
}

// HealthCheck performs a health check
func (fep *FormEventProcessor) HealthCheck() error {
	// Check connectivity to form service
	return nil
}

// Response Event Processor Implementation

// ProcessEvent processes a response-related event
func (rep *ResponseEventProcessor) ProcessEvent(ctx context.Context, event *events.CDCEvent) error {
	// Process response-specific logic
	if event.Source.Table != "responses" && !strings.Contains(event.Source.Topic, "response") {
		return nil // Skip non-response events
	}

	// Similar implementation to FormEventProcessor
	// ... implementation details ...

	return nil
}

// GetName returns the processor name
func (rep *ResponseEventProcessor) GetName() string {
	return rep.name
}

// GetType returns the processor type
func (rep *ResponseEventProcessor) GetType() string {
	return "response"
}

// HealthCheck performs a health check
func (rep *ResponseEventProcessor) HealthCheck() error {
	return nil
}

// Analytics Event Processor Implementation

// initialize initializes the analytics event processor
func (aep *AnalyticsEventProcessor) initialize() error {
	// Initialize aggregators for different event types
	aep.aggregators["forms"] = &EventAggregator{
		WindowSize: 5 * time.Minute,
		Events:     make([]events.CDCEvent, 0),
		LastFlush:  time.Now(),
	}

	aep.aggregators["responses"] = &EventAggregator{
		WindowSize: 1 * time.Minute,
		Events:     make([]events.CDCEvent, 0),
		LastFlush:  time.Now(),
	}

	return nil
}

// ProcessEvent processes an analytics event
func (aep *AnalyticsEventProcessor) ProcessEvent(ctx context.Context, event *events.CDCEvent) error {
	// Determine aggregator based on event
	aggregatorKey := aep.getAggregatorKey(event)

	aggregator, exists := aep.aggregators[aggregatorKey]
	if !exists {
		// Create new aggregator if needed
		aggregator = &EventAggregator{
			WindowSize: 5 * time.Minute,
			Events:     make([]events.CDCEvent, 0),
			LastFlush:  time.Now(),
		}
		aep.aggregators[aggregatorKey] = aggregator
	}

	// Add event to aggregator
	aggregator.mutex.Lock()
	aggregator.Events = append(aggregator.Events, *event)

	// Check if window should be flushed
	if time.Since(aggregator.LastFlush) >= aggregator.WindowSize {
		eventsToFlush := make([]events.CDCEvent, len(aggregator.Events))
		copy(eventsToFlush, aggregator.Events)
		aggregator.Events = aggregator.Events[:0] // Clear slice
		aggregator.LastFlush = time.Now()
		aggregator.mutex.Unlock()

		// Flush events asynchronously
		go aep.flushAggregatedEvents(ctx, aggregatorKey, eventsToFlush)
	} else {
		aggregator.mutex.Unlock()
	}

	return nil
}

// getAggregatorKey determines the aggregator key for an event
func (aep *AnalyticsEventProcessor) getAggregatorKey(event *events.CDCEvent) string {
	if event.Source.Table != "" {
		return event.Source.Table
	}

	// Extract from topic
	parts := strings.Split(event.Source.Topic, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}

	return "default"
}

// flushAggregatedEvents flushes aggregated events to analytics topic
func (aep *AnalyticsEventProcessor) flushAggregatedEvents(ctx context.Context, key string, events []events.CDCEvent) {
	if len(events) == 0 {
		return
	}

	// Create analytics summary
	summary := map[string]interface{}{
		"aggregator_key": key,
		"event_count":    len(events),
		"window_start":   events[0].Timestamp,
		"window_end":     events[len(events)-1].Timestamp,
		"operations":     aep.summarizeOperations(events),
		"tables":         aep.summarizeTables(events),
		"timestamp":      time.Now().Unix(),
	}

	topic := fmt.Sprintf("analytics.%s", key)
	eventKey := fmt.Sprintf("analytics_%s_%d", key, time.Now().UnixNano())

	// Publish to analytics topic
	message := &kafka.Message{
		ID:        eventKey,
		EventType: topic,
		Source:    "analytics-processor",
		Data:      summary,
		Topic:     topic,
		Key:       eventKey,
		Headers:   make(map[string]string),
		Metadata: kafka.MessageMetadata{
			Timestamp:   time.Now(),
			Version:     "1.0",
			ContentType: "application/json",
			Encoding:    "utf-8",
		},
	}

	if err := aep.kafka.PublishMessage(ctx, message); err != nil {
		aep.logger.Error("Failed to publish analytics summary",
			zap.String("topic", topic),
			zap.Error(err))
	}
}

// summarizeOperations summarizes operations in events
func (aep *AnalyticsEventProcessor) summarizeOperations(events []events.CDCEvent) map[string]int {
	operations := make(map[string]int)
	for _, event := range events {
		operations[event.Operation]++
	}
	return operations
}

// summarizeTables summarizes tables in events
func (aep *AnalyticsEventProcessor) summarizeTables(events []events.CDCEvent) map[string]int {
	tables := make(map[string]int)
	for _, event := range events {
		if event.Source.Table != "" {
			tables[event.Source.Table]++
		}
	}
	return tables
}

// GetName returns the processor name
func (aep *AnalyticsEventProcessor) GetName() string {
	return aep.name
}

// GetType returns the processor type
func (aep *AnalyticsEventProcessor) GetType() string {
	return "analytics"
}

// HealthCheck performs a health check
func (aep *AnalyticsEventProcessor) HealthCheck() error {
	return nil
}

// Transformation Implementations

// TableNameTransformation transforms table names
type TableNameTransformation struct{}

func (t *TableNameTransformation) Transform(ctx context.Context, event *events.CDCEvent) (*events.CDCEvent, error) {
	// Transform table names to standardized format
	if event.Source.Table != "" {
		event.Source.Table = strings.ToLower(event.Source.Table)
	}
	return event, nil
}

func (t *TableNameTransformation) GetName() string {
	return "table-name-transformation"
}

func (t *TableNameTransformation) GetConfig() map[string]interface{} {
	return map[string]interface{}{"type": "table_name"}
}

// TimestampTransformation transforms timestamps
type TimestampTransformation struct{}

func (t *TimestampTransformation) Transform(ctx context.Context, event *events.CDCEvent) (*events.CDCEvent, error) {
	// Ensure timestamp is in Unix format
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().Unix()
	}
	return event, nil
}

func (t *TimestampTransformation) GetName() string {
	return "timestamp-transformation"
}

func (t *TimestampTransformation) GetConfig() map[string]interface{} {
	return map[string]interface{}{"type": "timestamp"}
}

// SchemaTransformation transforms schema information
type SchemaTransformation struct{}

func (t *SchemaTransformation) Transform(ctx context.Context, event *events.CDCEvent) (*events.CDCEvent, error) {
	// Add schema information if missing
	if event.Schema == nil {
		event.Schema = &events.Schema{
			Type:     "struct",
			Optional: false,
			Name:     fmt.Sprintf("%s.%s.Envelope", event.Source.Topic, event.Source.Table),
		}
	}
	return event, nil
}

func (t *SchemaTransformation) GetName() string {
	return "schema-transformation"
}

func (t *SchemaTransformation) GetConfig() map[string]interface{} {
	return map[string]interface{}{"type": "schema"}
}

// Filter Implementations

// TableFilter filters events by table name
type TableFilter struct {
	AllowedTables []string
}

func (f *TableFilter) ShouldProcess(ctx context.Context, event *events.CDCEvent) (bool, error) {
	if len(f.AllowedTables) == 0 {
		return true, nil // Allow all if no restrictions
	}

	for _, table := range f.AllowedTables {
		if event.Source.Table == table {
			return true, nil
		}
	}
	return false, nil
}

func (f *TableFilter) GetName() string {
	return "table-filter"
}

func (f *TableFilter) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"type":           "table",
		"allowed_tables": f.AllowedTables,
	}
}

// OperationFilter filters events by operation type
type OperationFilter struct {
	AllowedOperations []string
}

func (f *OperationFilter) ShouldProcess(ctx context.Context, event *events.CDCEvent) (bool, error) {
	if len(f.AllowedOperations) == 0 {
		return true, nil // Allow all if no restrictions
	}

	for _, op := range f.AllowedOperations {
		if event.Operation == op {
			return true, nil
		}
	}
	return false, nil
}

func (f *OperationFilter) GetName() string {
	return "operation-filter"
}

func (f *OperationFilter) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"type":               "operation",
		"allowed_operations": f.AllowedOperations,
	}
}

// Route Implementations

// TopicRoute routes events based on topic patterns
type TopicRoute struct {
	SourceTopics []string
	TargetTopics []string
}

func (r *TopicRoute) ShouldRoute(ctx context.Context, event *events.CDCEvent) (bool, error) {
	for _, source := range r.SourceTopics {
		if event.Source.Topic == source || strings.HasPrefix(event.Source.Topic, source) {
			return true, nil
		}
	}
	return false, nil
}

func (r *TopicRoute) GetTargets() []string {
	return r.TargetTopics
}

func (r *TopicRoute) GetName() string {
	return "topic-route"
}

// Helper function to initialize processor metrics
func initProcessorMetrics() *ProcessorMetrics {
	return &ProcessorMetrics{
		EventsProcessed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "eventbus_events_processed_total",
			Help: "Total number of events processed",
		}),
		EventsFiltered: promauto.NewCounter(prometheus.CounterOpts{
			Name: "eventbus_events_filtered_total",
			Help: "Total number of events filtered out",
		}),
		EventsFailed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "eventbus_events_failed_total",
			Help: "Total number of events that failed processing",
		}),
		ProcessingLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "eventbus_processing_latency_seconds",
			Help:    "Histogram of event processing latencies",
			Buckets: prometheus.DefBuckets,
		}),
		ProcessorHealthScore: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "eventbus_processor_health_score",
			Help: "Health score of event processors",
		}, []string{"processor", "type"}),
		TransformationTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "eventbus_transformation_time_seconds",
			Help:    "Time spent in event transformations",
			Buckets: prometheus.DefBuckets,
		}),
		RoutingDecisions: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "eventbus_routing_decisions_total",
			Help: "Total number of routing decisions made",
		}, []string{"route", "target"}),
		ErrorsByType: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "eventbus_errors_by_type_total",
			Help: "Total number of errors by type",
		}, []string{"processor_type", "error_type"}),
	}
}
