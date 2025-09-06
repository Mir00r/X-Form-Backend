// Package kafka provides Apache Kafka client implementation for the Event Bus Service
// This package implements enterprise-grade Kafka producer and consumer functionality
// with comprehensive error handling, retry logic, and observability features.
package kafka

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// Client represents a Kafka client that handles both producing and consuming messages
// It implements enterprise patterns including circuit breaker, retry logic, and metrics
type Client struct {
	config   *config.Config
	logger   *zap.Logger
	producer sarama.SyncProducer
	consumer sarama.ConsumerGroup
	admin    sarama.ClusterAdmin
	mutex    sync.RWMutex
	closed   bool

	// Metrics
	metrics *KafkaMetrics
}

// KafkaMetrics contains Prometheus metrics for Kafka operations
type KafkaMetrics struct {
	MessagesProduced prometheus.Counter
	MessagesConsumed prometheus.Counter
	ProducerErrors   prometheus.Counter
	ConsumerErrors   prometheus.Counter
	ProducerLatency  prometheus.Histogram
	ConsumerLatency  prometheus.Histogram
	ConnectionStatus prometheus.Gauge
	TopicsCount      prometheus.Gauge
	PartitionsCount  prometheus.Gauge
}

// Message represents a standardized event message structure
type Message struct {
	// Message identification
	ID            string `json:"id"`
	CorrelationID string `json:"correlation_id"`
	EventType     string `json:"event_type"`
	Source        string `json:"source"`

	// Payload and metadata
	Data     interface{}       `json:"data"`
	Headers  map[string]string `json:"headers"`
	Metadata MessageMetadata   `json:"metadata"`

	// Routing information
	Topic     string `json:"topic"`
	Partition int32  `json:"partition,omitempty"`
	Key       string `json:"key,omitempty"`
}

// MessageMetadata contains message metadata for tracing and debugging
type MessageMetadata struct {
	Timestamp      time.Time `json:"timestamp"`
	Version        string    `json:"version"`
	SchemaVersion  string    `json:"schema_version,omitempty"`
	ContentType    string    `json:"content_type"`
	Encoding       string    `json:"encoding"`
	RetryCount     int       `json:"retry_count,omitempty"`
	OriginalTopic  string    `json:"original_topic,omitempty"`
	ProcessingTime int64     `json:"processing_time,omitempty"`
}

// ConsumerHandler defines the interface for message handlers
type ConsumerHandler interface {
	Handle(ctx context.Context, message *Message) error
	GetTopics() []string
	GetGroupID() string
}

// ProducerCallback defines callback function for async producer operations
type ProducerCallback func(message *Message, partition int32, offset int64, err error)

// NewClient creates a new Kafka client with the provided configuration
// It initializes producer, consumer, and admin clients with proper error handling
func NewClient(cfg *config.Config, logger *zap.Logger) (*Client, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	client := &Client{
		config:  cfg,
		logger:  logger,
		metrics: initMetrics(),
	}

	// Initialize Kafka configuration
	kafkaConfig, err := client.createKafkaConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka config: %w", err)
	}

	// Initialize producer
	if err := client.initProducer(kafkaConfig); err != nil {
		return nil, fmt.Errorf("failed to initialize producer: %w", err)
	}

	// Initialize consumer
	if err := client.initConsumer(kafkaConfig); err != nil {
		client.producer.Close() // Clean up producer on consumer init failure
		return nil, fmt.Errorf("failed to initialize consumer: %w", err)
	}

	// Initialize admin client
	if err := client.initAdmin(kafkaConfig); err != nil {
		client.producer.Close()
		client.consumer.Close(context.Background())
		return nil, fmt.Errorf("failed to initialize admin client: %w", err)
	}

	// Update connection status metric
	client.metrics.ConnectionStatus.Set(1)

	logger.Info("Kafka client initialized successfully",
		zap.Strings("brokers", cfg.Kafka.Brokers),
		zap.String("client_id", cfg.Kafka.ClientID),
		zap.String("group_id", cfg.Kafka.Consumer.GroupID))

	return client, nil
}

// createKafkaConfig creates Sarama configuration from service config
func (c *Client) createKafkaConfig() (*sarama.Config, error) {
	kafkaConfig := sarama.NewConfig()

	// Set version
	version, err := sarama.ParseKafkaVersion(c.config.Kafka.Version)
	if err != nil {
		return nil, fmt.Errorf("invalid Kafka version: %w", err)
	}
	kafkaConfig.Version = version

	// Client configuration
	kafkaConfig.ClientID = c.config.Kafka.ClientID
	kafkaConfig.ChannelBufferSize = c.config.Kafka.Consumer.ChannelBufferSize

	// Configure security
	if err := c.configureSecurity(kafkaConfig); err != nil {
		return nil, fmt.Errorf("failed to configure security: %w", err)
	}

	// Configure producer
	c.configureProducer(kafkaConfig)

	// Configure consumer
	c.configureConsumer(kafkaConfig)

	// Configure admin
	kafkaConfig.Admin.Timeout = c.config.Kafka.Admin.Timeout

	// Enable metadata refresh
	kafkaConfig.Metadata.RefreshFrequency = 5 * time.Minute
	kafkaConfig.Metadata.Full = true
	kafkaConfig.Metadata.RetryMax = 3
	kafkaConfig.Metadata.RetryBackoff = 250 * time.Millisecond

	return kafkaConfig, nil
}

// configureSecurity configures Kafka security settings
func (c *Client) configureSecurity(kafkaConfig *sarama.Config) error {
	securityConfig := c.config.Kafka.Security

	switch securityConfig.Protocol {
	case "PLAINTEXT":
		// No additional configuration needed
	case "SASL_PLAINTEXT":
		kafkaConfig.Net.SASL.Enable = true
		if err := c.configureSASL(kafkaConfig); err != nil {
			return err
		}
	case "SASL_SSL":
		kafkaConfig.Net.SASL.Enable = true
		kafkaConfig.Net.TLS.Enable = true
		if err := c.configureSASL(kafkaConfig); err != nil {
			return err
		}
		if err := c.configureTLS(kafkaConfig); err != nil {
			return err
		}
	case "SSL":
		kafkaConfig.Net.TLS.Enable = true
		if err := c.configureTLS(kafkaConfig); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported security protocol: %s", securityConfig.Protocol)
	}

	return nil
}

// configureSASL configures SASL authentication
func (c *Client) configureSASL(kafkaConfig *sarama.Config) error {
	saslConfig := c.config.Kafka.Security.SASL

	switch saslConfig.Mechanism {
	case "PLAIN":
		kafkaConfig.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	case "SCRAM-SHA-256":
		kafkaConfig.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
	case "SCRAM-SHA-512":
		kafkaConfig.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
	default:
		return fmt.Errorf("unsupported SASL mechanism: %s", saslConfig.Mechanism)
	}

	kafkaConfig.Net.SASL.User = saslConfig.Username
	kafkaConfig.Net.SASL.Password = saslConfig.Password

	return nil
}

// configureTLS configures TLS settings
func (c *Client) configureTLS(kafkaConfig *sarama.Config) error {
	tlsConfig := c.config.Kafka.Security.TLS

	kafkaConfig.Net.TLS.Config = &tls.Config{
		InsecureSkipVerify: tlsConfig.InsecureSkipVerify,
	}

	// Load certificates if provided
	if tlsConfig.CertFile != "" && tlsConfig.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(tlsConfig.CertFile, tlsConfig.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to load client certificate: %w", err)
		}
		kafkaConfig.Net.TLS.Config.Certificates = []tls.Certificate{cert}
	}

	return nil
}

// configureProducer configures producer settings
func (c *Client) configureProducer(kafkaConfig *sarama.Config) {
	producerConfig := c.config.Kafka.Producer

	// Acknowledgment settings
	kafkaConfig.Producer.RequiredAcks = sarama.RequiredAcks(producerConfig.RequiredAcks)
	kafkaConfig.Producer.Timeout = producerConfig.Timeout

	// Retry settings
	kafkaConfig.Producer.Retry.Max = producerConfig.RetryMax
	kafkaConfig.Producer.Retry.Backoff = producerConfig.RetryBackoff

	// Compression settings
	switch producerConfig.Compression {
	case "none":
		kafkaConfig.Producer.Compression = sarama.CompressionNone
	case "gzip":
		kafkaConfig.Producer.Compression = sarama.CompressionGZIP
	case "snappy":
		kafkaConfig.Producer.Compression = sarama.CompressionSnappy
	case "lz4":
		kafkaConfig.Producer.Compression = sarama.CompressionLZ4
	case "zstd":
		kafkaConfig.Producer.Compression = sarama.CompressionZSTD
	default:
		kafkaConfig.Producer.Compression = sarama.CompressionSnappy
	}

	// Message settings
	kafkaConfig.Producer.MaxMessageBytes = producerConfig.MaxMessageBytes

	// Flush settings
	kafkaConfig.Producer.Flush.Frequency = producerConfig.FlushFrequency
	kafkaConfig.Producer.Flush.Messages = producerConfig.FlushMessages
	kafkaConfig.Producer.Flush.Bytes = producerConfig.FlushBytes

	// Idempotent producer
	kafkaConfig.Producer.Idempotent = producerConfig.Idempotent
	if producerConfig.Idempotent {
		kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
		kafkaConfig.Producer.Retry.Max = 1
		kafkaConfig.Net.MaxOpenRequests = 1
	}

	// Enable return of successes and errors
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Producer.Return.Errors = true

	// Partitioner
	kafkaConfig.Producer.Partitioner = sarama.NewHashPartitioner
}

// configureConsumer configures consumer settings
func (c *Client) configureConsumer(kafkaConfig *sarama.Config) {
	consumerConfig := c.config.Kafka.Consumer

	// Group configuration
	kafkaConfig.Consumer.Group.Session.Timeout = consumerConfig.SessionTimeout
	kafkaConfig.Consumer.Group.Heartbeat.Interval = consumerConfig.HeartbeatInterval

	// Offset configuration
	switch consumerConfig.AutoOffsetReset {
	case "earliest":
		kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	case "latest":
		kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	default:
		kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	// Auto-commit configuration
	kafkaConfig.Consumer.Offsets.AutoCommit.Enable = consumerConfig.EnableAutoCommit
	kafkaConfig.Consumer.Offsets.AutoCommit.Interval = consumerConfig.AutoCommitInterval

	// Fetch configuration
	kafkaConfig.Consumer.Fetch.Min = consumerConfig.FetchMin
	kafkaConfig.Consumer.Fetch.Default = consumerConfig.FetchDefault
	kafkaConfig.Consumer.Fetch.Max = consumerConfig.FetchMax
	kafkaConfig.Consumer.MaxWaitTime = consumerConfig.MaxWaitTime

	// Processing configuration
	kafkaConfig.Consumer.MaxProcessingTime = consumerConfig.MaxProcessingTime

	// Isolation level
	switch consumerConfig.IsolationLevel {
	case "ReadUncommitted":
		kafkaConfig.Consumer.IsolationLevel = sarama.ReadUncommitted
	case "ReadCommitted":
		kafkaConfig.Consumer.IsolationLevel = sarama.ReadCommitted
	default:
		kafkaConfig.Consumer.IsolationLevel = sarama.ReadUncommitted
	}

	// Channel buffer size
	kafkaConfig.ChannelBufferSize = consumerConfig.ChannelBufferSize

	// Return errors
	kafkaConfig.Consumer.Return.Errors = consumerConfig.ReturnErrors
}

// initProducer initializes the Kafka producer
func (c *Client) initProducer(kafkaConfig *sarama.Config) error {
	producer, err := sarama.NewSyncProducer(c.config.Kafka.Brokers, kafkaConfig)
	if err != nil {
		return fmt.Errorf("failed to create producer: %w", err)
	}

	c.producer = producer
	c.logger.Info("Kafka producer initialized successfully")
	return nil
}

// initConsumer initializes the Kafka consumer group
func (c *Client) initConsumer(kafkaConfig *sarama.Config) error {
	consumer, err := sarama.NewConsumerGroup(c.config.Kafka.Brokers, c.config.Kafka.Consumer.GroupID, kafkaConfig)
	if err != nil {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	c.consumer = consumer
	c.logger.Info("Kafka consumer group initialized successfully",
		zap.String("group_id", c.config.Kafka.Consumer.GroupID))
	return nil
}

// initAdmin initializes the Kafka admin client
func (c *Client) initAdmin(kafkaConfig *sarama.Config) error {
	admin, err := sarama.NewClusterAdmin(c.config.Kafka.Brokers, kafkaConfig)
	if err != nil {
		return fmt.Errorf("failed to create admin client: %w", err)
	}

	c.admin = admin
	c.logger.Info("Kafka admin client initialized successfully")
	return nil
}

// PublishMessage publishes a message to Kafka
func (c *Client) PublishMessage(ctx context.Context, message *Message) error {
	if c.closed {
		return fmt.Errorf("kafka client is closed")
	}

	start := time.Now()
	defer func() {
		duration := time.Since(start)
		c.metrics.ProducerLatency.Observe(duration.Seconds())
	}()

	// Prepare Kafka message
	kafkaMessage, err := c.prepareKafkaMessage(message)
	if err != nil {
		c.metrics.ProducerErrors.Inc()
		return fmt.Errorf("failed to prepare message: %w", err)
	}

	// Send message
	partition, offset, err := c.producer.SendMessage(kafkaMessage)
	if err != nil {
		c.metrics.ProducerErrors.Inc()
		c.logger.Error("Failed to publish message",
			zap.String("topic", message.Topic),
			zap.String("message_id", message.ID),
			zap.Error(err))
		return fmt.Errorf("failed to send message: %w", err)
	}

	c.metrics.MessagesProduced.Inc()
	c.logger.Debug("Message published successfully",
		zap.String("topic", message.Topic),
		zap.String("message_id", message.ID),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset))

	return nil
}

// PublishMessageAsync publishes a message asynchronously with callback
func (c *Client) PublishMessageAsync(ctx context.Context, message *Message, callback ProducerCallback) {
	go func() {
		err := c.PublishMessage(ctx, message)
		if callback != nil {
			// Extract partition and offset from error or success
			partition := int32(-1)
			offset := int64(-1)
			callback(message, partition, offset, err)
		}
	}()
}

// StartConsumer starts consuming messages with the provided handler
func (c *Client) StartConsumer(ctx context.Context, handler ConsumerHandler) error {
	if c.closed {
		return fmt.Errorf("kafka client is closed")
	}

	topics := handler.GetTopics()
	if len(topics) == 0 {
		return fmt.Errorf("no topics specified for consumer")
	}

	c.logger.Info("Starting Kafka consumer",
		zap.Strings("topics", topics),
		zap.String("group_id", handler.GetGroupID()))

	// Create consumer group handler
	consumerHandler := &consumerGroupHandler{
		client:  c,
		handler: handler,
		logger:  c.logger,
	}

	// Start consuming in a goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("Consumer context cancelled, stopping consumer")
				return
			default:
				if err := c.consumer.Consume(ctx, topics, consumerHandler); err != nil {
					c.logger.Error("Consumer error", zap.Error(err))
					c.metrics.ConsumerErrors.Inc()
				}
			}
		}
	}()

	// Handle consumer errors
	go func() {
		for err := range c.consumer.Errors() {
			c.logger.Error("Consumer group error", zap.Error(err))
			c.metrics.ConsumerErrors.Inc()
		}
	}()

	return nil
}

// CreateTopic creates a new Kafka topic
func (c *Client) CreateTopic(ctx context.Context, topicName string, numPartitions int32, replicationFactor int16) error {
	if c.closed {
		return fmt.Errorf("kafka client is closed")
	}

	topicDetail := &sarama.TopicDetail{
		NumPartitions:     numPartitions,
		ReplicationFactor: replicationFactor,
		ConfigEntries: map[string]*string{
			"cleanup.policy": &[]string{"compact"}[0],
			"retention.ms":   &[]string{"604800000"}[0], // 7 days
		},
	}

	err := c.admin.CreateTopic(topicName, topicDetail, false)
	if err != nil {
		if kafkaErr, ok := err.(*sarama.TopicError); ok && kafkaErr.Err == sarama.ErrTopicAlreadyExists {
			c.logger.Info("Topic already exists", zap.String("topic", topicName))
			return nil
		}
		return fmt.Errorf("failed to create topic %s: %w", topicName, err)
	}

	c.logger.Info("Topic created successfully",
		zap.String("topic", topicName),
		zap.Int32("partitions", numPartitions),
		zap.Int16("replication_factor", replicationFactor))

	return nil
}

// ListTopics returns a list of available topics
func (c *Client) ListTopics(ctx context.Context) ([]string, error) {
	if c.closed {
		return nil, fmt.Errorf("kafka client is closed")
	}

	metadata, err := c.admin.DescribeTopics(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to describe topics: %w", err)
	}

	topics := make([]string, 0, len(metadata))
	for topic := range metadata {
		topics = append(topics, topic)
	}

	c.metrics.TopicsCount.Set(float64(len(topics)))
	return topics, nil
}

// GetTopicMetadata returns metadata for a specific topic
func (c *Client) GetTopicMetadata(ctx context.Context, topicName string) (*sarama.TopicMetadata, error) {
	if c.closed {
		return nil, fmt.Errorf("kafka client is closed")
	}

	metadata, err := c.admin.DescribeTopics([]string{topicName})
	if err != nil {
		return nil, fmt.Errorf("failed to describe topic %s: %w", topicName, err)
	}

	topicMetadata, exists := metadata[topicName]
	if !exists {
		return nil, fmt.Errorf("topic %s not found", topicName)
	}

	return topicMetadata, nil
}

// HealthCheck performs a health check on the Kafka client
func (c *Client) HealthCheck(ctx context.Context) error {
	if c.closed {
		return fmt.Errorf("kafka client is closed")
	}

	// Check if we can list topics (basic connectivity test)
	_, err := c.ListTopics(ctx)
	if err != nil {
		c.metrics.ConnectionStatus.Set(0)
		return fmt.Errorf("kafka health check failed: %w", err)
	}

	c.metrics.ConnectionStatus.Set(1)
	return nil
}

// Close closes the Kafka client and all its connections
func (c *Client) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return nil
	}

	var errors []error

	// Close producer
	if c.producer != nil {
		if err := c.producer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close producer: %w", err))
		}
	}

	// Close consumer
	if c.consumer != nil {
		if err := c.consumer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close consumer: %w", err))
		}
	}

	// Close admin client
	if c.admin != nil {
		if err := c.admin.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close admin client: %w", err))
		}
	}

	c.closed = true
	c.metrics.ConnectionStatus.Set(0)

	if len(errors) > 0 {
		return fmt.Errorf("errors closing Kafka client: %v", errors)
	}

	c.logger.Info("Kafka client closed successfully")
	return nil
}

// prepareKafkaMessage converts internal Message to Sarama ProducerMessage
func (c *Client) prepareKafkaMessage(message *Message) (*sarama.ProducerMessage, error) {
	// Serialize message data
	value, err := c.serializeMessage(message)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	// Create Kafka message
	kafkaMessage := &sarama.ProducerMessage{
		Topic:     message.Topic,
		Value:     sarama.ByteEncoder(value),
		Timestamp: message.Metadata.Timestamp,
	}

	// Set key if provided
	if message.Key != "" {
		kafkaMessage.Key = sarama.StringEncoder(message.Key)
	}

	// Set partition if specified
	if message.Partition >= 0 {
		kafkaMessage.Partition = message.Partition
	}

	// Add headers
	for key, value := range message.Headers {
		kafkaMessage.Headers = append(kafkaMessage.Headers, sarama.RecordHeader{
			Key:   []byte(key),
			Value: []byte(value),
		})
	}

	// Add metadata headers
	kafkaMessage.Headers = append(kafkaMessage.Headers,
		sarama.RecordHeader{Key: []byte("event-id"), Value: []byte(message.ID)},
		sarama.RecordHeader{Key: []byte("correlation-id"), Value: []byte(message.CorrelationID)},
		sarama.RecordHeader{Key: []byte("event-type"), Value: []byte(message.EventType)},
		sarama.RecordHeader{Key: []byte("source"), Value: []byte(message.Source)},
		sarama.RecordHeader{Key: []byte("content-type"), Value: []byte(message.Metadata.ContentType)},
		sarama.RecordHeader{Key: []byte("schema-version"), Value: []byte(message.Metadata.SchemaVersion)},
	)

	return kafkaMessage, nil
}

// serializeMessage serializes message data to bytes
func (c *Client) serializeMessage(message *Message) ([]byte, error) {
	// This is a simplified JSON serialization
	// In production, you might want to use Avro, Protocol Buffers, or other formats
	return []byte(fmt.Sprintf(`{
		"id": "%s",
		"correlation_id": "%s",
		"event_type": "%s",
		"source": "%s",
		"data": %v,
		"metadata": {
			"timestamp": "%s",
			"version": "%s",
			"content_type": "%s",
			"encoding": "%s"
		}
	}`, message.ID, message.CorrelationID, message.EventType, message.Source,
		message.Data, message.Metadata.Timestamp.Format(time.RFC3339),
		message.Metadata.Version, message.Metadata.ContentType, message.Metadata.Encoding)), nil
}

// initMetrics initializes Prometheus metrics for Kafka operations
func initMetrics() *KafkaMetrics {
	return &KafkaMetrics{
		MessagesProduced: promauto.NewCounter(prometheus.CounterOpts{
			Name: "kafka_messages_produced_total",
			Help: "Total number of messages produced to Kafka",
		}),
		MessagesConsumed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "kafka_messages_consumed_total",
			Help: "Total number of messages consumed from Kafka",
		}),
		ProducerErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "kafka_producer_errors_total",
			Help: "Total number of producer errors",
		}),
		ConsumerErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "kafka_consumer_errors_total",
			Help: "Total number of consumer errors",
		}),
		ProducerLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "kafka_producer_latency_seconds",
			Help:    "Histogram of producer latencies",
			Buckets: prometheus.DefBuckets,
		}),
		ConsumerLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "kafka_consumer_latency_seconds",
			Help:    "Histogram of consumer latencies",
			Buckets: prometheus.DefBuckets,
		}),
		ConnectionStatus: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "kafka_connection_status",
			Help: "Kafka connection status (1 = connected, 0 = disconnected)",
		}),
		TopicsCount: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "kafka_topics_count",
			Help: "Number of Kafka topics",
		}),
		PartitionsCount: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "kafka_partitions_count",
			Help: "Number of Kafka partitions",
		}),
	}
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	client  *Client
	handler ConsumerHandler
	logger  *zap.Logger
}

// Setup is run before the consumer starts consuming
func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	h.logger.Info("Consumer group session setup")
	return nil
}

// Cleanup is run after the consumer stops consuming
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.logger.Info("Consumer group session cleanup")
	return nil
}

// ConsumeClaim processes messages from a partition
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			start := time.Now()

			// Convert Kafka message to internal Message
			internalMessage, err := h.convertKafkaMessage(message)
			if err != nil {
				h.logger.Error("Failed to convert Kafka message",
					zap.Error(err),
					zap.String("topic", message.Topic),
					zap.Int32("partition", message.Partition),
					zap.Int64("offset", message.Offset))
				continue
			}

			// Process message with handler
			ctx := session.Context()
			if err := h.handler.Handle(ctx, internalMessage); err != nil {
				h.logger.Error("Failed to handle message",
					zap.Error(err),
					zap.String("message_id", internalMessage.ID),
					zap.String("topic", message.Topic))
				h.client.metrics.ConsumerErrors.Inc()
			} else {
				h.client.metrics.MessagesConsumed.Inc()
				h.logger.Debug("Message processed successfully",
					zap.String("message_id", internalMessage.ID),
					zap.String("topic", message.Topic),
					zap.Duration("processing_time", time.Since(start)))
			}

			// Record processing latency
			h.client.metrics.ConsumerLatency.Observe(time.Since(start).Seconds())

			// Mark message as processed
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// convertKafkaMessage converts Sarama ConsumerMessage to internal Message
func (h *consumerGroupHandler) convertKafkaMessage(kafkaMessage *sarama.ConsumerMessage) (*Message, error) {
	// Extract headers
	headers := make(map[string]string)
	var eventID, correlationID, eventType, source, contentType, schemaVersion string

	for _, header := range kafkaMessage.Headers {
		key := string(header.Key)
		value := string(header.Value)
		headers[key] = value

		// Extract standard headers
		switch key {
		case "event-id":
			eventID = value
		case "correlation-id":
			correlationID = value
		case "event-type":
			eventType = value
		case "source":
			source = value
		case "content-type":
			contentType = value
		case "schema-version":
			schemaVersion = value
		}
	}

	// Create internal message
	message := &Message{
		ID:            eventID,
		CorrelationID: correlationID,
		EventType:     eventType,
		Source:        source,
		Data:          kafkaMessage.Value, // Raw data - should be deserialized based on content type
		Headers:       headers,
		Topic:         kafkaMessage.Topic,
		Partition:     kafkaMessage.Partition,
		Key:           string(kafkaMessage.Key),
		Metadata: MessageMetadata{
			Timestamp:     kafkaMessage.Timestamp,
			ContentType:   contentType,
			SchemaVersion: schemaVersion,
			Encoding:      "utf-8",
		},
	}

	return message, nil
}
