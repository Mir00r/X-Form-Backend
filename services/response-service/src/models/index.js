const { FieldValue } = require('firebase-admin/firestore');
const { ValidationHelper } = require('../utils/helpers');

/**
 * Base model class with common functionality
 */
class BaseModel {
  constructor(data = {}) {
    this.id = data.id || null;
    this.createdAt = data.createdAt || FieldValue.serverTimestamp();
    this.updatedAt = data.updatedAt || FieldValue.serverTimestamp();
  }

  /**
   * Convert model to Firestore document
   */
  toFirestore() {
    const data = { ...this };
    delete data.id; // ID is handled separately in Firestore
    return data;
  }

  /**
   * Create model from Firestore document
   */
  static fromFirestore(doc) {
    if (!doc.exists) return null;
    return new this({ id: doc.id, ...doc.data() });
  }

  /**
   * Update timestamps
   */
  touch() {
    this.updatedAt = FieldValue.serverTimestamp();
  }

  /**
   * Validate model data
   */
  validate() {
    return { isValid: true, errors: [] };
  }
}

/**
 * Response model representing a form submission
 */
class Response extends BaseModel {
  constructor(data = {}) {
    super(data);
    
    // Required fields
    this.formId = data.formId || null;
    this.responses = data.responses || {};
    
    // Submitter information
    this.submitterId = data.submitterId || null;
    this.submitterEmail = data.submitterEmail || null;
    this.submitterName = data.submitterName || null;
    this.isAnonymous = data.isAnonymous || false;
    
    // Submission metadata
    this.submittedAt = data.submittedAt || FieldValue.serverTimestamp();
    this.ipAddress = data.ipAddress || null;
    this.userAgent = data.userAgent || null;
    this.sessionId = data.sessionId || null;
    
    // Form state
    this.status = data.status || 'submitted'; // submitted, draft, flagged, deleted
    this.isDraft = data.isDraft || false;
    this.isComplete = data.isComplete !== undefined ? data.isComplete : true;
    
    // Processing metadata
    this.processingStatus = data.processingStatus || 'pending'; // pending, processed, failed
    this.lastProcessedAt = data.lastProcessedAt || null;
    this.processingErrors = data.processingErrors || [];
    
    // Analytics data
    this.startTime = data.startTime || null;
    this.endTime = data.endTime || null;
    this.duration = data.duration || null; // in seconds
    this.pageViews = data.pageViews || [];
    this.interactions = data.interactions || [];
    
    // File attachments
    this.attachments = data.attachments || [];
    
    // Validation and scoring
    this.validationResults = data.validationResults || {};
    this.score = data.score || null;
    this.tags = data.tags || [];
    
    // Integration data
    this.integrationData = data.integrationData || {};
    this.syncStatus = data.syncStatus || {}; // Google Sheets, webhooks, etc.
    
    // Audit trail
    this.version = data.version || 1;
    this.history = data.history || [];
  }

  /**
   * Validate response data
   */
  validate() {
    const errors = [];

    // Required field validation
    if (!this.formId) {
      errors.push({ field: 'formId', message: 'Form ID is required' });
    }

    if (!this.responses || typeof this.responses !== 'object') {
      errors.push({ field: 'responses', message: 'Responses must be an object' });
    }

    // Email validation if provided
    if (this.submitterEmail && !ValidationHelper.isValidEmail(this.submitterEmail)) {
      errors.push({ field: 'submitterEmail', message: 'Invalid email format' });
    }

    // Status validation
    const validStatuses = ['submitted', 'draft', 'flagged', 'deleted'];
    if (!validStatuses.includes(this.status)) {
      errors.push({ field: 'status', message: 'Invalid status value' });
    }

    // Processing status validation
    const validProcessingStatuses = ['pending', 'processed', 'failed'];
    if (!validProcessingStatuses.includes(this.processingStatus)) {
      errors.push({ field: 'processingStatus', message: 'Invalid processing status' });
    }

    return {
      isValid: errors.length === 0,
      errors
    };
  }

  /**
   * Add response to a question
   */
  addResponse(questionId, answer, metadata = {}) {
    this.responses[questionId] = {
      value: answer,
      answeredAt: FieldValue.serverTimestamp(),
      ...metadata
    };
    this.touch();
  }

  /**
   * Remove response to a question
   */
  removeResponse(questionId) {
    delete this.responses[questionId];
    this.touch();
  }

  /**
   * Add attachment
   */
  addAttachment(attachment) {
    this.attachments.push({
      id: attachment.id || Date.now().toString(),
      filename: attachment.filename,
      originalName: attachment.originalName,
      mimeType: attachment.mimeType,
      size: attachment.size,
      url: attachment.url,
      uploadedAt: FieldValue.serverTimestamp(),
      questionId: attachment.questionId || null,
    });
    this.touch();
  }

  /**
   * Calculate duration if start and end times are set
   */
  calculateDuration() {
    if (this.startTime && this.endTime) {
      const start = this.startTime.toDate ? this.startTime.toDate() : new Date(this.startTime);
      const end = this.endTime.toDate ? this.endTime.toDate() : new Date(this.endTime);
      this.duration = Math.floor((end - start) / 1000); // seconds
    }
  }

  /**
   * Add interaction tracking
   */
  addInteraction(type, data = {}) {
    this.interactions.push({
      type, // click, focus, blur, change, etc.
      timestamp: FieldValue.serverTimestamp(),
      ...data
    });
  }

  /**
   * Add to audit history
   */
  addToHistory(action, userId = null, details = {}) {
    this.history.push({
      action,
      userId,
      timestamp: FieldValue.serverTimestamp(),
      version: this.version,
      details
    });
    this.version += 1;
  }

  /**
   * Mark as processed
   */
  markAsProcessed(results = {}) {
    this.processingStatus = 'processed';
    this.lastProcessedAt = FieldValue.serverTimestamp();
    this.processingErrors = [];
    this.integrationData = { ...this.integrationData, ...results };
    this.touch();
  }

  /**
   * Mark as failed with errors
   */
  markAsFailed(errors = []) {
    this.processingStatus = 'failed';
    this.lastProcessedAt = FieldValue.serverTimestamp();
    this.processingErrors = errors;
    this.touch();
  }

  /**
   * Get response summary
   */
  getSummary() {
    return {
      id: this.id,
      formId: this.formId,
      submitterEmail: this.submitterEmail,
      submitterName: this.submitterName,
      submittedAt: this.submittedAt,
      status: this.status,
      isComplete: this.isComplete,
      responseCount: Object.keys(this.responses).length,
      duration: this.duration,
      score: this.score,
      tags: this.tags,
    };
  }
}

/**
 * Response analytics model for aggregated data
 */
class ResponseAnalytics extends BaseModel {
  constructor(data = {}) {
    super(data);
    
    this.formId = data.formId || null;
    this.date = data.date || null; // Date string (YYYY-MM-DD)
    this.period = data.period || 'daily'; // daily, weekly, monthly
    
    // Response metrics
    this.totalResponses = data.totalResponses || 0;
    this.completedResponses = data.completedResponses || 0;
    this.draftResponses = data.draftResponses || 0;
    this.averageCompletionTime = data.averageCompletionTime || 0;
    this.completionRate = data.completionRate || 0;
    
    // Question analytics
    this.questionAnalytics = data.questionAnalytics || {};
    
    // Traffic analytics
    this.uniqueVisitors = data.uniqueVisitors || 0;
    this.totalViews = data.totalViews || 0;
    this.bounceRate = data.bounceRate || 0;
    
    // Device/browser analytics
    this.deviceBreakdown = data.deviceBreakdown || {};
    this.browserBreakdown = data.browserBreakdown || {};
    this.locationBreakdown = data.locationBreakdown || {};
    
    // Conversion funnel
    this.funnelData = data.funnelData || {};
  }

  /**
   * Update analytics with new response
   */
  updateWithResponse(response) {
    this.totalResponses += 1;
    
    if (response.isComplete) {
      this.completedResponses += 1;
    } else if (response.isDraft) {
      this.draftResponses += 1;
    }
    
    // Update completion rate
    this.completionRate = this.totalResponses > 0 
      ? (this.completedResponses / this.totalResponses) * 100 
      : 0;
    
    // Update average completion time
    if (response.duration && response.isComplete) {
      const totalTime = (this.averageCompletionTime * (this.completedResponses - 1)) + response.duration;
      this.averageCompletionTime = totalTime / this.completedResponses;
    }
    
    this.touch();
  }
}

/**
 * Form integration settings model
 */
class FormIntegration extends BaseModel {
  constructor(data = {}) {
    super(data);
    
    this.formId = data.formId || null;
    this.type = data.type || null; // google_sheets, webhook, email, etc.
    this.isActive = data.isActive !== undefined ? data.isActive : true;
    
    // Integration-specific configuration
    this.config = data.config || {};
    
    // Google Sheets integration
    if (this.type === 'google_sheets') {
      this.config = {
        spreadsheetId: data.config?.spreadsheetId || null,
        worksheetName: data.config?.worksheetName || 'Responses',
        includeTimestamp: data.config?.includeTimestamp !== false,
        fieldMapping: data.config?.fieldMapping || {},
        ...data.config
      };
    }
    
    // Webhook integration
    if (this.type === 'webhook') {
      this.config = {
        url: data.config?.url || null,
        method: data.config?.method || 'POST',
        headers: data.config?.headers || {},
        includeMetadata: data.config?.includeMetadata !== false,
        retryPolicy: data.config?.retryPolicy || { maxRetries: 3, backoff: 'exponential' },
        ...data.config
      };
    }
    
    // Email integration
    if (this.type === 'email') {
      this.config = {
        recipients: data.config?.recipients || [],
        subject: data.config?.subject || 'New form response',
        template: data.config?.template || 'default',
        includeAttachments: data.config?.includeAttachments !== false,
        ...data.config
      };
    }
    
    // Sync status
    this.lastSyncAt = data.lastSyncAt || null;
    this.syncStatus = data.syncStatus || 'pending'; // pending, success, failed
    this.syncErrors = data.syncErrors || [];
    this.totalSynced = data.totalSynced || 0;
  }

  /**
   * Validate integration configuration
   */
  validate() {
    const errors = [];

    if (!this.formId) {
      errors.push({ field: 'formId', message: 'Form ID is required' });
    }

    if (!this.type) {
      errors.push({ field: 'type', message: 'Integration type is required' });
    }

    // Type-specific validation
    if (this.type === 'google_sheets' && !this.config.spreadsheetId) {
      errors.push({ field: 'config.spreadsheetId', message: 'Spreadsheet ID is required for Google Sheets integration' });
    }

    if (this.type === 'webhook') {
      if (!this.config.url) {
        errors.push({ field: 'config.url', message: 'Webhook URL is required' });
      } else if (!ValidationHelper.isValidUrl(this.config.url)) {
        errors.push({ field: 'config.url', message: 'Invalid webhook URL format' });
      }
    }

    if (this.type === 'email' && (!this.config.recipients || this.config.recipients.length === 0)) {
      errors.push({ field: 'config.recipients', message: 'Email recipients are required' });
    }

    return {
      isValid: errors.length === 0,
      errors
    };
  }

  /**
   * Update sync status
   */
  updateSyncStatus(status, error = null) {
    this.syncStatus = status;
    this.lastSyncAt = FieldValue.serverTimestamp();
    
    if (status === 'success') {
      this.syncErrors = [];
      this.totalSynced += 1;
    } else if (status === 'failed' && error) {
      this.syncErrors.push({
        error: error.message,
        timestamp: FieldValue.serverTimestamp(),
      });
    }
    
    this.touch();
  }
}

module.exports = {
  BaseModel,
  Response,
  ResponseAnalytics,
  FormIntegration,
};
