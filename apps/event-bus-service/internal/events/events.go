// Package events defines event structures and types for the Event Bus Service
// This package provides comprehensive event modeling for Change Data Capture,
// application events, and event metadata with enterprise-grade features.
package events

import (
	"encoding/json"
	"fmt"
	"time"
)

// CDCEvent represents a Change Data Capture event from Debezium
type CDCEvent struct {
	ID        string                 `json:"id"`
	Schema    *Schema                `json:"schema,omitempty"`
	Payload   *Payload               `json:"payload"`
	Source    *Source                `json:"source"`
	Operation string                 `json:"op"`    // c, u, d, r (create, update, delete, read)
	Timestamp int64                  `json:"ts_ms"` // Timestamp in milliseconds
	Before    map[string]interface{} `json:"before,omitempty"`
	After     map[string]interface{} `json:"after,omitempty"`
	Headers   map[string]string      `json:"headers,omitempty"`
	Metadata  *EventMetadata         `json:"metadata,omitempty"`
}

// Schema represents the schema information for an event
type Schema struct {
	Type       string                 `json:"type"`
	Fields     []SchemaField          `json:"fields,omitempty"`
	Optional   bool                   `json:"optional"`
	Name       string                 `json:"name,omitempty"`
	Version    int                    `json:"version,omitempty"`
	Doc        string                 `json:"doc,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// SchemaField represents a field in the schema
type SchemaField struct {
	Type       string                 `json:"type"`
	Optional   bool                   `json:"optional"`
	Field      string                 `json:"field"`
	Default    interface{}            `json:"default,omitempty"`
	Doc        string                 `json:"doc,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// Payload represents the event payload
type Payload struct {
	Before      map[string]interface{} `json:"before,omitempty"`
	After       map[string]interface{} `json:"after,omitempty"`
	Source      *Source                `json:"source"`
	Operation   string                 `json:"op"`
	Timestamp   int64                  `json:"ts_ms"`
	Transaction *Transaction           `json:"transaction,omitempty"`
}

// Source represents the source information of an event
type Source struct {
	Version   string `json:"version"`
	Connector string `json:"connector"`
	Name      string `json:"name"`
	Timestamp int64  `json:"ts_ms"`
	Snapshot  string `json:"snapshot,omitempty"`
	Database  string `json:"db"`
	Sequence  string `json:"sequence,omitempty"`
	Schema    string `json:"schema"`
	Table     string `json:"table"`
	TxID      int64  `json:"txId,omitempty"`
	LSN       int64  `json:"lsn,omitempty"`
	XMIN      int64  `json:"xmin,omitempty"`
	Topic     string `json:"topic"`
}

// Transaction represents transaction information
type Transaction struct {
	ID                  string `json:"id"`
	TotalOrder          int64  `json:"total_order"`
	DataCollectionOrder int64  `json:"data_collection_order"`
}

// EventMetadata represents additional metadata for events
type EventMetadata struct {
	ProcessingTime time.Time              `json:"processing_time"`
	ProcessorID    string                 `json:"processor_id"`
	Version        string                 `json:"version"`
	Correlation    *CorrelationData       `json:"correlation,omitempty"`
	Security       *SecurityInfo          `json:"security,omitempty"`
	Quality        *QualityMetrics        `json:"quality,omitempty"`
	Routing        *RoutingInfo           `json:"routing,omitempty"`
	Custom         map[string]interface{} `json:"custom,omitempty"`
}

// CorrelationData represents correlation information for event tracing
type CorrelationData struct {
	TraceID      string            `json:"trace_id"`
	SpanID       string            `json:"span_id"`
	ParentSpanID string            `json:"parent_span_id,omitempty"`
	SessionID    string            `json:"session_id,omitempty"`
	RequestID    string            `json:"request_id,omitempty"`
	UserID       string            `json:"user_id,omitempty"`
	TenantID     string            `json:"tenant_id,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
}

// SecurityInfo represents security information for events
type SecurityInfo struct {
	Signature     string                 `json:"signature,omitempty"`
	SignatureAlgo string                 `json:"signature_algo,omitempty"`
	Checksum      string                 `json:"checksum,omitempty"`
	Encrypted     bool                   `json:"encrypted"`
	AccessLevel   string                 `json:"access_level,omitempty"`
	Permissions   []string               `json:"permissions,omitempty"`
	Claims        map[string]interface{} `json:"claims,omitempty"`
}

// QualityMetrics represents data quality metrics for events
type QualityMetrics struct {
	Completeness float64         `json:"completeness"`
	Validity     float64         `json:"validity"`
	Consistency  float64         `json:"consistency"`
	Accuracy     float64         `json:"accuracy"`
	Issues       []QualityIssue  `json:"issues,omitempty"`
	Score        float64         `json:"score"`
	Checks       map[string]bool `json:"checks,omitempty"`
}

// QualityIssue represents a data quality issue
type QualityIssue struct {
	Type        string                 `json:"type"`
	Field       string                 `json:"field,omitempty"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// RoutingInfo represents routing information for events
type RoutingInfo struct {
	SourceTopic     string   `json:"source_topic"`
	TargetTopics    []string `json:"target_topics"`
	RoutingRules    []string `json:"routing_rules,omitempty"`
	ProcessorChain  []string `json:"processor_chain,omitempty"`
	RetryCount      int      `json:"retry_count"`
	MaxRetries      int      `json:"max_retries"`
	DeadLetterTopic string   `json:"dead_letter_topic,omitempty"`
}

// ApplicationEvent represents application-level events
type ApplicationEvent struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"`
	Source          string                 `json:"source"`
	Subject         string                 `json:"subject"`
	Time            time.Time              `json:"time"`
	Data            map[string]interface{} `json:"data"`
	DataContentType string                 `json:"datacontenttype,omitempty"`
	SpecVersion     string                 `json:"specversion"`
	Extensions      map[string]interface{} `json:"extensions,omitempty"`
	Metadata        *EventMetadata         `json:"metadata,omitempty"`
}

// FormEvent represents form-specific events
type FormEvent struct {
	*ApplicationEvent
	FormID    string                 `json:"form_id"`
	FormTitle string                 `json:"form_title"`
	FormType  string                 `json:"form_type"`
	OwnerID   string                 `json:"owner_id"`
	Status    string                 `json:"status"`
	Fields    []FormField            `json:"fields,omitempty"`
	Settings  map[string]interface{} `json:"settings,omitempty"`
	Version   int                    `json:"version"`
}

// FormField represents a form field
type FormField struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Label      string                 `json:"label"`
	Required   bool                   `json:"required"`
	Options    []string               `json:"options,omitempty"`
	Validation map[string]interface{} `json:"validation,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// ResponseEvent represents response-specific events
type ResponseEvent struct {
	*ApplicationEvent
	ResponseID  string                 `json:"response_id"`
	FormID      string                 `json:"form_id"`
	UserID      string                 `json:"user_id"`
	Answers     map[string]interface{} `json:"answers"`
	Status      string                 `json:"status"`
	SubmittedAt time.Time              `json:"submitted_at"`
	IP          string                 `json:"ip,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Score       float64                `json:"score,omitempty"`
}

// UserEvent represents user-specific events
type UserEvent struct {
	*ApplicationEvent
	UserID      string                 `json:"user_id"`
	Email       string                 `json:"email"`
	Username    string                 `json:"username,omitempty"`
	Role        string                 `json:"role"`
	Status      string                 `json:"status"`
	Profile     map[string]interface{} `json:"profile,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
	LastLoginAt *time.Time             `json:"last_login_at,omitempty"`
}

// AnalyticsEvent represents analytics-specific events
type AnalyticsEvent struct {
	*ApplicationEvent
	SessionID  string                 `json:"session_id"`
	UserID     string                 `json:"user_id,omitempty"`
	EventName  string                 `json:"event_name"`
	Properties map[string]interface{} `json:"properties"`
	Metrics    map[string]float64     `json:"metrics,omitempty"`
	Dimensions map[string]string      `json:"dimensions,omitempty"`
	Page       string                 `json:"page,omitempty"`
	Referrer   string                 `json:"referrer,omitempty"`
	Device     *DeviceInfo            `json:"device,omitempty"`
	Location   *LocationInfo          `json:"location,omitempty"`
}

// DeviceInfo represents device information
type DeviceInfo struct {
	Type         string `json:"type"`    // desktop, mobile, tablet
	OS           string `json:"os"`      // windows, macos, ios, android, linux
	Browser      string `json:"browser"` // chrome, firefox, safari, edge
	Version      string `json:"version"` // browser version
	Viewport     string `json:"viewport,omitempty"`
	ScreenSize   string `json:"screen_size,omitempty"`
	TouchEnabled bool   `json:"touch_enabled"`
}

// LocationInfo represents location information
type LocationInfo struct {
	Country   string  `json:"country,omitempty"`
	Region    string  `json:"region,omitempty"`
	City      string  `json:"city,omitempty"`
	Timezone  string  `json:"timezone,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Accuracy  float64 `json:"accuracy,omitempty"`
}

// EventBatch represents a batch of events
type EventBatch struct {
	ID        string                 `json:"id"`
	Events    []CDCEvent             `json:"events"`
	BatchSize int                    `json:"batch_size"`
	CreatedAt time.Time              `json:"created_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Checksum  string                 `json:"checksum,omitempty"`
}

// EventFilter represents criteria for filtering events
type EventFilter struct {
	EventTypes    []string          `json:"event_types,omitempty"`
	Sources       []string          `json:"sources,omitempty"`
	Tables        []string          `json:"tables,omitempty"`
	Operations    []string          `json:"operations,omitempty"`
	TimeRange     *TimeRange        `json:"time_range,omitempty"`
	Conditions    []FilterCondition `json:"conditions,omitempty"`
	IncludeFields []string          `json:"include_fields,omitempty"`
	ExcludeFields []string          `json:"exclude_fields,omitempty"`
}

// TimeRange represents a time range filter
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// FilterCondition represents a filter condition
type FilterCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, gte, lte, in, nin, regex
	Value    interface{} `json:"value"`
	Type     string      `json:"type,omitempty"` // string, number, boolean, date
}

// EventTransformation represents an event transformation
type EventTransformation struct {
	Name    string                 `json:"name"`
	Type    string                 `json:"type"`
	Config  map[string]interface{} `json:"config"`
	Input   *EventFilter           `json:"input,omitempty"`
	Output  *EventSchema           `json:"output,omitempty"`
	Enabled bool                   `json:"enabled"`
}

// EventSchema represents an event schema definition
type EventSchema struct {
	Name       string                 `json:"name"`
	Version    string                 `json:"version"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required,omitempty"`
	Examples   []interface{}          `json:"examples,omitempty"`
}

// EventRoute represents an event routing configuration
type EventRoute struct {
	Name            string         `json:"name"`
	Description     string         `json:"description,omitempty"`
	Enabled         bool           `json:"enabled"`
	Priority        int            `json:"priority"`
	Filter          *EventFilter   `json:"filter"`
	Targets         []RouteTarget  `json:"targets"`
	Transformations []string       `json:"transformations,omitempty"`
	ErrorHandling   *ErrorHandling `json:"error_handling,omitempty"`
}

// RouteTarget represents a routing target
type RouteTarget struct {
	Type     string                 `json:"type"` // kafka, webhook, service, storage
	Endpoint string                 `json:"endpoint"`
	Config   map[string]interface{} `json:"config,omitempty"`
	Enabled  bool                   `json:"enabled"`
	Retry    *RetryConfig           `json:"retry,omitempty"`
}

// RetryConfig represents retry configuration
type RetryConfig struct {
	MaxAttempts   int           `json:"max_attempts"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	Multiplier    float64       `json:"multiplier"`
	JitterEnabled bool          `json:"jitter_enabled"`
}

// ErrorHandling represents error handling configuration
type ErrorHandling struct {
	Strategy        string        `json:"strategy"` // retry, skip, deadletter
	MaxRetries      int           `json:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay"`
	DeadLetterTopic string        `json:"dead_letter_topic,omitempty"`
	AlertOnError    bool          `json:"alert_on_error"`
}

// EventStats represents statistics for events
type EventStats struct {
	TotalEvents       int64              `json:"total_events"`
	EventsByType      map[string]int64   `json:"events_by_type"`
	EventsBySource    map[string]int64   `json:"events_by_source"`
	EventsByOperation map[string]int64   `json:"events_by_operation"`
	ProcessingLatency map[string]float64 `json:"processing_latency"`
	ErrorRate         float64            `json:"error_rate"`
	Throughput        float64            `json:"throughput"`
	TimeRange         *TimeRange         `json:"time_range"`
	LastUpdated       time.Time          `json:"last_updated"`
}

// Helper methods

// ToJSON converts an event to JSON
func (e *CDCEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON creates an event from JSON
func (e *CDCEvent) FromJSON(data []byte) error {
	return json.Unmarshal(data, e)
}

// GetEventType returns the event type based on source and operation
func (e *CDCEvent) GetEventType() string {
	return fmt.Sprintf("%s.%s.%s", e.Source.Database, e.Source.Table, e.Operation)
}

// IsCreate returns true if this is a create operation
func (e *CDCEvent) IsCreate() bool {
	return e.Operation == "c"
}

// IsUpdate returns true if this is an update operation
func (e *CDCEvent) IsUpdate() bool {
	return e.Operation == "u"
}

// IsDelete returns true if this is a delete operation
func (e *CDCEvent) IsDelete() bool {
	return e.Operation == "d"
}

// IsRead returns true if this is a read operation (initial snapshot)
func (e *CDCEvent) IsRead() bool {
	return e.Operation == "r"
}

// GetPrimaryKey extracts the primary key from the event data
func (e *CDCEvent) GetPrimaryKey() interface{} {
	// Try to get from after data first (for create/update)
	if e.After != nil {
		if id, exists := e.After["id"]; exists {
			return id
		}
	}

	// Fallback to before data (for delete)
	if e.Before != nil {
		if id, exists := e.Before["id"]; exists {
			return id
		}
	}

	return nil
}

// HasChangedField checks if a specific field has changed in an update event
func (e *CDCEvent) HasChangedField(fieldName string) bool {
	if !e.IsUpdate() {
		return false
	}

	if e.Before == nil || e.After == nil {
		return false
	}

	beforeValue, beforeExists := e.Before[fieldName]
	afterValue, afterExists := e.After[fieldName]

	// If field existence changed
	if beforeExists != afterExists {
		return true
	}

	// If both exist, compare values
	if beforeExists && afterExists {
		return beforeValue != afterValue
	}

	return false
}

// GetChangedFields returns a list of fields that changed in an update event
func (e *CDCEvent) GetChangedFields() []string {
	if !e.IsUpdate() {
		return nil
	}

	if e.Before == nil || e.After == nil {
		return nil
	}

	var changedFields []string

	// Check all fields in after
	for field := range e.After {
		if e.HasChangedField(field) {
			changedFields = append(changedFields, field)
		}
	}

	// Check for removed fields (exist in before but not in after)
	for field := range e.Before {
		if _, exists := e.After[field]; !exists {
			changedFields = append(changedFields, field)
		}
	}

	return changedFields
}

// AddMetadata adds metadata to the event
func (e *CDCEvent) AddMetadata(key string, value interface{}) {
	if e.Metadata == nil {
		e.Metadata = &EventMetadata{
			Custom: make(map[string]interface{}),
		}
	}
	if e.Metadata.Custom == nil {
		e.Metadata.Custom = make(map[string]interface{})
	}
	e.Metadata.Custom[key] = value
}

// GetMetadata retrieves metadata from the event
func (e *CDCEvent) GetMetadata(key string) (interface{}, bool) {
	if e.Metadata == nil || e.Metadata.Custom == nil {
		return nil, false
	}
	value, exists := e.Metadata.Custom[key]
	return value, exists
}

// SetCorrelationID sets the correlation ID for the event
func (e *CDCEvent) SetCorrelationID(traceID, spanID string) {
	if e.Metadata == nil {
		e.Metadata = &EventMetadata{}
	}
	if e.Metadata.Correlation == nil {
		e.Metadata.Correlation = &CorrelationData{}
	}
	e.Metadata.Correlation.TraceID = traceID
	e.Metadata.Correlation.SpanID = spanID
}

// ToApplicationEvent converts a CDC event to an application event
func (e *CDCEvent) ToApplicationEvent() *ApplicationEvent {
	return &ApplicationEvent{
		ID:          e.ID,
		Type:        e.GetEventType(),
		Source:      fmt.Sprintf("%s.%s", e.Source.Connector, e.Source.Name),
		Subject:     e.Source.Table,
		Time:        time.Unix(0, e.Timestamp*int64(time.Millisecond)),
		Data:        e.After,
		SpecVersion: "1.0",
		Metadata:    e.Metadata,
	}
}

// Validate validates the event structure
func (e *CDCEvent) Validate() error {
	if e.ID == "" {
		return fmt.Errorf("event ID is required")
	}

	if e.Source == nil {
		return fmt.Errorf("event source is required")
	}

	if e.Operation == "" {
		return fmt.Errorf("event operation is required")
	}

	if e.Timestamp == 0 {
		return fmt.Errorf("event timestamp is required")
	}

	// Validate operation-specific requirements
	switch e.Operation {
	case "c": // Create
		if e.After == nil {
			return fmt.Errorf("create events must have 'after' data")
		}
	case "u": // Update
		if e.Before == nil || e.After == nil {
			return fmt.Errorf("update events must have both 'before' and 'after' data")
		}
	case "d": // Delete
		if e.Before == nil {
			return fmt.Errorf("delete events must have 'before' data")
		}
	}

	return nil
}

// Clone creates a deep copy of the event
func (e *CDCEvent) Clone() *CDCEvent {
	data, _ := json.Marshal(e)
	var clone CDCEvent
	json.Unmarshal(data, &clone)
	return &clone
}

// Package level helper functions

// NewCDCEvent creates a new CDC event with default values
func NewCDCEvent(id, operation string, source *Source) *CDCEvent {
	return &CDCEvent{
		ID:        id,
		Operation: operation,
		Source:    source,
		Timestamp: time.Now().UnixMilli(),
		Headers:   make(map[string]string),
	}
}

// NewApplicationEvent creates a new application event
func NewApplicationEvent(eventType, source, subject string, data map[string]interface{}) *ApplicationEvent {
	return &ApplicationEvent{
		ID:              fmt.Sprintf("%s_%d", eventType, time.Now().UnixNano()),
		Type:            eventType,
		Source:          source,
		Subject:         subject,
		Time:            time.Now(),
		Data:            data,
		SpecVersion:     "1.0",
		DataContentType: "application/json",
	}
}

// NewFormEvent creates a new form event
func NewFormEvent(eventType, formID, formTitle string, data map[string]interface{}) *FormEvent {
	appEvent := NewApplicationEvent(eventType, "form-service", formID, data)
	return &FormEvent{
		ApplicationEvent: appEvent,
		FormID:           formID,
		FormTitle:        formTitle,
		Status:           "active",
		Fields:           []FormField{},
		Settings:         make(map[string]interface{}),
		Version:          1,
	}
}

// NewResponseEvent creates a new response event
func NewResponseEvent(eventType, responseID, formID, userID string, answers map[string]interface{}) *ResponseEvent {
	appEvent := NewApplicationEvent(eventType, "response-service", responseID, answers)
	return &ResponseEvent{
		ApplicationEvent: appEvent,
		ResponseID:       responseID,
		FormID:           formID,
		UserID:           userID,
		Answers:          answers,
		Status:           "submitted",
		SubmittedAt:      time.Now(),
	}
}

// NewAnalyticsEvent creates a new analytics event
func NewAnalyticsEvent(eventName, sessionID string, properties map[string]interface{}) *AnalyticsEvent {
	appEvent := NewApplicationEvent("analytics."+eventName, "analytics-service", sessionID, properties)
	return &AnalyticsEvent{
		ApplicationEvent: appEvent,
		SessionID:        sessionID,
		EventName:        eventName,
		Properties:       properties,
		Metrics:          make(map[string]float64),
		Dimensions:       make(map[string]string),
	}
}

// MergeBatches merges multiple event batches into one
func MergeBatches(batches ...*EventBatch) *EventBatch {
	if len(batches) == 0 {
		return nil
	}

	merged := &EventBatch{
		ID:        fmt.Sprintf("merged_%d", time.Now().UnixNano()),
		Events:    []CDCEvent{},
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	for _, batch := range batches {
		if batch != nil {
			merged.Events = append(merged.Events, batch.Events...)
		}
	}

	merged.BatchSize = len(merged.Events)
	return merged
}

// FilterEvents filters events based on the provided filter criteria
func FilterEvents(events []CDCEvent, filter *EventFilter) []CDCEvent {
	if filter == nil {
		return events
	}

	var filtered []CDCEvent

	for _, event := range events {
		if matchesFilter(&event, filter) {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// matchesFilter checks if an event matches the filter criteria
func matchesFilter(event *CDCEvent, filter *EventFilter) bool {
	// Check event types
	if len(filter.EventTypes) > 0 {
		eventType := event.GetEventType()
		found := false
		for _, t := range filter.EventTypes {
			if t == eventType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check sources
	if len(filter.Sources) > 0 {
		found := false
		for _, s := range filter.Sources {
			if s == event.Source.Name || s == event.Source.Database {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check tables
	if len(filter.Tables) > 0 {
		found := false
		for _, t := range filter.Tables {
			if t == event.Source.Table {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check operations
	if len(filter.Operations) > 0 {
		found := false
		for _, op := range filter.Operations {
			if op == event.Operation {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check time range
	if filter.TimeRange != nil {
		eventTime := time.Unix(0, event.Timestamp*int64(time.Millisecond))
		if eventTime.Before(filter.TimeRange.From) || eventTime.After(filter.TimeRange.To) {
			return false
		}
	}

	// Check conditions
	for _, condition := range filter.Conditions {
		if !matchesCondition(event, &condition) {
			return false
		}
	}

	return true
}

// matchesCondition checks if an event matches a specific condition
func matchesCondition(event *CDCEvent, condition *FilterCondition) bool {
	// This is a simplified implementation
	// In production, you would implement proper field access and comparison logic

	var fieldValue interface{}

	// Get field value from event data
	if event.After != nil {
		if val, exists := event.After[condition.Field]; exists {
			fieldValue = val
		}
	}
	if fieldValue == nil && event.Before != nil {
		if val, exists := event.Before[condition.Field]; exists {
			fieldValue = val
		}
	}

	// Compare based on operator
	switch condition.Operator {
	case "eq":
		return fieldValue == condition.Value
	case "ne":
		return fieldValue != condition.Value
	// Add more operators as needed
	default:
		return false
	}
}
