/**
 * Event System for Microservices Communication
 * Handles event publishing and subscription for decoupled service communication
 */

const EventEmitter = require('events');
const logger = require('../utils/logger');

class EventSystem extends EventEmitter {
  constructor() {
    super();
    this.setMaxListeners(50); // Increase default limit
    this.eventStore = new Map(); // Simple in-memory event store
    this.subscribers = new Map(); // Track active subscribers
    this.retryAttempts = 3;
    this.retryDelay = 1000;
  }

  /**
   * Publish an event to the event system
   * @param {string} eventType - Type of event
   * @param {Object} eventData - Event payload
   * @param {string} correlationId - Request correlation ID
   * @param {Object} metadata - Additional metadata
   */
  async publishEvent(eventType, eventData, correlationId, metadata = {}) {
    const event = {
      id: this.generateEventId(),
      type: eventType,
      data: eventData,
      correlationId,
      metadata: {
        ...metadata,
        publishedAt: new Date().toISOString(),
        publishedBy: 'response-service',
        version: '1.0'
      }
    };

    try {
      // Store event for potential replay
      this.eventStore.set(event.id, event);

      logger.info('Publishing event', {
        eventId: event.id,
        eventType,
        correlationId,
        dataKeys: Object.keys(eventData)
      });

      // Emit the event
      this.emit(eventType, event);

      // Log successful publication
      logger.logBusiness('EVENT_PUBLISHED', eventType, event.id, {
        correlationId,
        eventData: JSON.stringify(eventData)
      }, { correlationId });

      return event.id;

    } catch (error) {
      logger.error('Failed to publish event', {
        eventType,
        correlationId,
        error: error.message
      });
      throw error;
    }
  }

  /**
   * Subscribe to events with error handling and retry logic
   * @param {string} eventType - Type of event to subscribe to
   * @param {Function} handler - Event handler function
   * @param {Object} options - Subscription options
   */
  subscribeToEvent(eventType, handler, options = {}) {
    const {
      retryOnError = true,
      maxRetries = this.retryAttempts,
      retryDelay = this.retryDelay,
      subscriberName = 'anonymous'
    } = options;

    const wrappedHandler = async (event) => {
      const startTime = Date.now();
      let attempts = 0;

      while (attempts <= maxRetries) {
        try {
          logger.info('Processing event', {
            eventId: event.id,
            eventType,
            subscriber: subscriberName,
            attempt: attempts + 1,
            correlationId: event.correlationId
          });

          await handler(event);

          const duration = Date.now() - startTime;
          logger.info('Event processed successfully', {
            eventId: event.id,
            eventType,
            subscriber: subscriberName,
            duration,
            correlationId: event.correlationId
          });

          break; // Success, exit retry loop

        } catch (error) {
          attempts++;
          
          logger.error('Event processing failed', {
            eventId: event.id,
            eventType,
            subscriber: subscriberName,
            attempt: attempts,
            maxRetries: maxRetries + 1,
            error: error.message,
            correlationId: event.correlationId
          });

          if (attempts > maxRetries || !retryOnError) {
            logger.error('Event processing failed permanently', {
              eventId: event.id,
              eventType,
              subscriber: subscriberName,
              totalAttempts: attempts,
              correlationId: event.correlationId
            });

            // Publish dead letter event
            await this.publishDeadLetterEvent(event, error, subscriberName);
            break;
          }

          // Wait before retry with exponential backoff
          const delay = retryDelay * Math.pow(2, attempts - 1);
          await this.sleep(delay);
        }
      }
    };

    // Register the wrapped handler
    this.on(eventType, wrappedHandler);

    // Track subscription
    if (!this.subscribers.has(eventType)) {
      this.subscribers.set(eventType, []);
    }
    this.subscribers.get(eventType).push({
      name: subscriberName,
      handler: wrappedHandler,
      registeredAt: new Date().toISOString()
    });

    logger.info('Event subscription registered', {
      eventType,
      subscriber: subscriberName
    });

    return wrappedHandler;
  }

  /**
   * Unsubscribe from events
   * @param {string} eventType - Type of event
   * @param {Function} handler - Handler to remove
   */
  unsubscribeFromEvent(eventType, handler) {
    this.removeListener(eventType, handler);
    
    // Remove from subscribers tracking
    const subscribers = this.subscribers.get(eventType);
    if (subscribers) {
      const index = subscribers.findIndex(sub => sub.handler === handler);
      if (index !== -1) {
        subscribers.splice(index, 1);
      }
    }

    logger.info('Event subscription removed', { eventType });
  }

  /**
   * Publish a dead letter event when processing fails
   * @param {Object} originalEvent - The original event that failed
   * @param {Error} error - The error that occurred
   * @param {string} subscriberName - Name of the subscriber that failed
   */
  async publishDeadLetterEvent(originalEvent, error, subscriberName) {
    try {
      await this.publishEvent('event.processing.failed', {
        originalEvent,
        error: {
          message: error.message,
          stack: error.stack,
          name: error.name
        },
        failedSubscriber: subscriberName,
        failedAt: new Date().toISOString()
      }, originalEvent.correlationId);

    } catch (deadLetterError) {
      logger.error('Failed to publish dead letter event', {
        originalEventId: originalEvent.id,
        error: deadLetterError.message
      });
    }
  }

  /**
   * Get event by ID (for replay or debugging)
   * @param {string} eventId - Event ID
   * @returns {Object|null} Event data
   */
  getEvent(eventId) {
    return this.eventStore.get(eventId) || null;
  }

  /**
   * Get events by type (for debugging or replay)
   * @param {string} eventType - Event type
   * @param {Object} options - Query options
   * @returns {Array} Events
   */
  getEventsByType(eventType, options = {}) {
    const { limit = 100, since } = options;
    const events = Array.from(this.eventStore.values())
      .filter(event => {
        if (event.type !== eventType) return false;
        if (since && new Date(event.metadata.publishedAt) < new Date(since)) return false;
        return true;
      })
      .sort((a, b) => new Date(b.metadata.publishedAt) - new Date(a.metadata.publishedAt))
      .slice(0, limit);

    return events;
  }

  /**
   * Get subscription statistics
   * @returns {Object} Subscription stats
   */
  getSubscriptionStats() {
    const stats = {};
    
    for (const [eventType, subscribers] of this.subscribers.entries()) {
      stats[eventType] = {
        subscriberCount: subscribers.length,
        subscribers: subscribers.map(sub => ({
          name: sub.name,
          registeredAt: sub.registeredAt
        }))
      };
    }

    return stats;
  }

  /**
   * Clear old events from memory (simple cleanup)
   * @param {number} maxAge - Maximum age in milliseconds
   */
  cleanupOldEvents(maxAge = 24 * 60 * 60 * 1000) { // Default: 24 hours
    const cutoff = new Date(Date.now() - maxAge);
    let cleaned = 0;

    for (const [eventId, event] of this.eventStore.entries()) {
      if (new Date(event.metadata.publishedAt) < cutoff) {
        this.eventStore.delete(eventId);
        cleaned++;
      }
    }

    logger.info('Event cleanup completed', {
      eventsRemoved: cleaned,
      remainingEvents: this.eventStore.size
    });

    return cleaned;
  }

  /**
   * Generate unique event ID
   * @returns {string} Event ID
   */
  generateEventId() {
    return `evt_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  /**
   * Sleep utility
   * @param {number} ms - Milliseconds to sleep
   * @returns {Promise}
   */
  sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

// Response Service Event Types
const ResponseEvents = {
  // Response lifecycle events
  RESPONSE_CREATED: 'response.created',
  RESPONSE_UPDATED: 'response.updated',
  RESPONSE_SUBMITTED: 'response.submitted',
  RESPONSE_DELETED: 'response.deleted',
  
  // Validation events
  RESPONSE_VALIDATED: 'response.validated',
  RESPONSE_VALIDATION_FAILED: 'response.validation.failed',
  
  // Analytics events
  ANALYTICS_GENERATED: 'analytics.generated',
  ANALYTICS_REQUESTED: 'analytics.requested',
  
  // System events
  SERVICE_STARTED: 'service.started',
  SERVICE_STOPPED: 'service.stopped',
  HEALTH_CHECK_FAILED: 'health.check.failed',
  
  // Integration events
  FORM_SERVICE_CALLED: 'integration.form_service.called',
  FORM_SERVICE_FAILED: 'integration.form_service.failed',
  
  // Error events
  EVENT_PROCESSING_FAILED: 'event.processing.failed'
};

// Create singleton instance
const eventSystem = new EventSystem();

// Set up automatic cleanup (runs every hour)
setInterval(() => {
  eventSystem.cleanupOldEvents();
}, 60 * 60 * 1000);

module.exports = {
  eventSystem,
  ResponseEvents
};
