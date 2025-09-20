const { firestore, FieldValue } = require('../config/firebase');
const { Response, ResponseAnalytics, FormIntegration } = require('../models');
const { PaginationHelper, DateHelper, ErrorHelper } = require('../utils/helpers');
const { customValidators } = require('../validators');
const logger = require('../utils/logger');

/**
 * Response Service - Core business logic for form responses
 */
class ResponseService {
  constructor() {
    this.collection = firestore.collection('responses');
    this.analyticsCollection = firestore.collection('response_analytics');
    this.integrationsCollection = firestore.collection('form_integrations');
  }

  /**
   * Create a new response
   * @param {Object} responseData - Response data
   * @param {Object} formSchema - Form schema for validation
   * @param {Object} metadata - Additional metadata
   */
  async createResponse(responseData, formSchema = null, metadata = {}) {
    try {
      // Create response model
      const response = new Response({
        ...responseData,
        ipAddress: metadata.ipAddress,
        userAgent: metadata.userAgent,
        sessionId: metadata.sessionId,
      });

      // Validate basic response structure
      const validation = response.validate();
      if (!validation.isValid) {
        throw ErrorHelper.createError('Response validation failed', 400, 'VALIDATION_ERROR');
      }

      // Validate response answers against form schema
      if (formSchema && response.responses) {
        const answerValidation = await customValidators.validateResponseAnswers(
          response.responses,
          formSchema
        );
        
        if (!answerValidation.isValid) {
          response.validationResults = {
            isValid: false,
            errors: answerValidation.errors,
            validatedAt: FieldValue.serverTimestamp(),
          };
          
          // Still save the response but mark validation issues
          logger.warn('Response created with validation errors', {
            formId: response.formId,
            errors: answerValidation.errors,
          });
        } else {
          response.validationResults = {
            isValid: true,
            errors: [],
            validatedAt: FieldValue.serverTimestamp(),
          };
        }
      }

      // Calculate duration if times are provided
      response.calculateDuration();

      // Add to audit history
      response.addToHistory('created', response.submitterId, {
        source: metadata.source || 'web',
        formVersion: formSchema?.version,
      });

      // Save to Firestore
      const docRef = await this.collection.add(response.toFirestore());
      response.id = docRef.id;

      logger.info('Response created', {
        responseId: response.id,
        formId: response.formId,
        submitterId: response.submitterId,
        isComplete: response.isComplete,
        duration: response.duration,
      });

      // Update analytics asynchronously
      this._updateAnalytics(response).catch(error => {
        logger.error('Failed to update analytics for new response:', error);
      });

      // Trigger integrations asynchronously
      this._triggerIntegrations(response).catch(error => {
        logger.error('Failed to trigger integrations for new response:', error);
      });

      return response;
    } catch (error) {
      logger.error('Failed to create response:', error);
      throw ErrorHelper.handleDbError(error);
    }
  }

  /**
   * Update an existing response
   * @param {string} responseId - Response ID
   * @param {Object} updates - Update data
   * @param {string} userId - User making the update
   */
  async updateResponse(responseId, updates, userId = null) {
    try {
      const docRef = this.collection.doc(responseId);
      const doc = await docRef.get();

      if (!doc.exists) {
        throw ErrorHelper.createError('Response not found', 404, 'RESPONSE_NOT_FOUND');
      }

      const response = Response.fromFirestore(doc);
      
      // Track original values for audit
      const originalValues = { ...response };

      // Apply updates
      Object.keys(updates).forEach(key => {
        if (updates[key] !== undefined) {
          response[key] = updates[key];
        }
      });

      // Update timestamp
      response.touch();

      // Add to audit history
      response.addToHistory('updated', userId, {
        updatedFields: Object.keys(updates),
        originalValues: Object.keys(updates).reduce((acc, key) => {
          acc[key] = originalValues[key];
          return acc;
        }, {}),
      });

      // Validate updated response
      const validation = response.validate();
      if (!validation.isValid) {
        throw ErrorHelper.createError('Updated response validation failed', 400, 'VALIDATION_ERROR');
      }

      // Save changes
      await docRef.update(response.toFirestore());

      logger.info('Response updated', {
        responseId,
        updatedFields: Object.keys(updates),
        userId,
      });

      // Trigger integrations for updates if needed
      if (updates.responses || updates.status) {
        this._triggerIntegrations(response, 'updated').catch(error => {
          logger.error('Failed to trigger integrations for updated response:', error);
        });
      }

      return response;
    } catch (error) {
      logger.error('Failed to update response:', error);
      throw ErrorHelper.handleDbError(error);
    }
  }

  /**
   * Get response by ID
   * @param {string} responseId - Response ID
   */
  async getResponse(responseId) {
    try {
      const doc = await this.collection.doc(responseId).get();
      
      if (!doc.exists) {
        throw ErrorHelper.createError('Response not found', 404, 'RESPONSE_NOT_FOUND');
      }

      return Response.fromFirestore(doc);
    } catch (error) {
      logger.error('Failed to get response:', error);
      throw ErrorHelper.handleDbError(error);
    }
  }

  /**
   * Query responses with filters and pagination
   * @param {Object} filters - Query filters
   * @param {Object} pagination - Pagination options
   */
  async queryResponses(filters = {}, pagination = {}) {
    try {
      const { page, limit } = PaginationHelper.validateParams(
        pagination.page,
        pagination.limit
      );

      let query = this.collection;

      // Apply filters
      if (filters.formId) {
        query = query.where('formId', '==', filters.formId);
      }

      if (filters.submitterId) {
        query = query.where('submitterId', '==', filters.submitterId);
      }

      if (filters.submitterEmail) {
        query = query.where('submitterEmail', '==', filters.submitterEmail);
      }

      if (filters.status && filters.status !== 'all') {
        query = query.where('status', '==', filters.status);
      }

      if (filters.isComplete !== undefined) {
        query = query.where('isComplete', '==', filters.isComplete);
      }

      if (filters.isDraft !== undefined) {
        query = query.where('isDraft', '==', filters.isDraft);
      }

      // Date range filtering
      if (filters.startDate) {
        query = query.where('submittedAt', '>=', new Date(filters.startDate));
      }

      if (filters.endDate) {
        query = query.where('submittedAt', '<=', new Date(filters.endDate));
      }

      // Score filtering
      if (filters.minScore !== undefined) {
        query = query.where('score', '>=', filters.minScore);
      }

      if (filters.maxScore !== undefined) {
        query = query.where('score', '<=', filters.maxScore);
      }

      // Tags filtering (array-contains for single tag)
      if (filters.tags && filters.tags.length > 0) {
        query = query.where('tags', 'array-contains-any', filters.tags);
      }

      // Sorting
      const sortBy = filters.sortBy || 'submittedAt';
      const sortOrder = filters.sortOrder || 'desc';
      query = query.orderBy(sortBy, sortOrder);

      // Count total for pagination
      const countSnapshot = await query.get();
      const total = countSnapshot.size;

      // Apply pagination
      const offset = PaginationHelper.getOffset(page, limit);
      if (offset > 0) {
        query = query.offset(offset);
      }
      query = query.limit(limit);

      // Execute query
      const snapshot = await query.get();
      const responses = snapshot.docs.map(doc => Response.fromFirestore(doc));

      // Search filtering (post-query for complex text search)
      let filteredResponses = responses;
      if (filters.search) {
        const searchTerm = filters.search.toLowerCase();
        filteredResponses = responses.filter(response => {
          // Search in response answers
          const responseText = JSON.stringify(response.responses).toLowerCase();
          return responseText.includes(searchTerm) ||
                 (response.submitterEmail && response.submitterEmail.toLowerCase().includes(searchTerm)) ||
                 (response.submitterName && response.submitterName.toLowerCase().includes(searchTerm));
        });
      }

      const paginationInfo = PaginationHelper.createPagination(page, limit, total);

      logger.info('Responses queried', {
        filters,
        resultCount: filteredResponses.length,
        total,
        page,
        limit,
      });

      return {
        items: filteredResponses,
        pagination: paginationInfo,
      };
    } catch (error) {
      logger.error('Failed to query responses:', error);
      throw ErrorHelper.handleDbError(error);
    }
  }

  /**
   * Delete response (soft delete by default)
   * @param {string} responseId - Response ID
   * @param {boolean} permanent - Whether to permanently delete
   * @param {string} userId - User performing deletion
   */
  async deleteResponse(responseId, permanent = false, userId = null) {
    try {
      const docRef = this.collection.doc(responseId);
      const doc = await docRef.get();

      if (!doc.exists) {
        throw ErrorHelper.createError('Response not found', 404, 'RESPONSE_NOT_FOUND');
      }

      if (permanent) {
        // Permanent deletion
        await docRef.delete();
        
        logger.info('Response permanently deleted', {
          responseId,
          userId,
        });
      } else {
        // Soft delete
        const response = Response.fromFirestore(doc);
        response.status = 'deleted';
        response.addToHistory('deleted', userId);
        response.touch();

        await docRef.update(response.toFirestore());

        logger.info('Response soft deleted', {
          responseId,
          userId,
        });
      }

      // Trigger integrations for deletion
      this._triggerIntegrations({ id: responseId }, 'deleted').catch(error => {
        logger.error('Failed to trigger integrations for deleted response:', error);
      });

      return true;
    } catch (error) {
      logger.error('Failed to delete response:', error);
      throw ErrorHelper.handleDbError(error);
    }
  }

  /**
   * Bulk update responses
   * @param {Array} responseIds - Array of response IDs
   * @param {Object} updates - Updates to apply
   * @param {string} userId - User performing updates
   */
  async bulkUpdateResponses(responseIds, updates, userId = null) {
    try {
      const batch = firestore.batch();
      const updatedResponses = [];

      for (const responseId of responseIds) {
        const docRef = this.collection.doc(responseId);
        const doc = await docRef.get();

        if (doc.exists) {
          const response = Response.fromFirestore(doc);
          
          // Apply updates
          Object.keys(updates).forEach(key => {
            if (updates[key] !== undefined) {
              response[key] = updates[key];
            }
          });

          response.touch();
          response.addToHistory('bulk_updated', userId, {
            updatedFields: Object.keys(updates),
          });

          batch.update(docRef, response.toFirestore());
          updatedResponses.push(response);
        }
      }

      await batch.commit();

      logger.info('Bulk update completed', {
        responseCount: updatedResponses.length,
        updates,
        userId,
      });

      return {
        updated: updatedResponses.length,
        responses: updatedResponses,
      };
    } catch (error) {
      logger.error('Failed to bulk update responses:', error);
      throw ErrorHelper.handleDbError(error);
    }
  }

  /**
   * Bulk delete responses
   * @param {Array} responseIds - Array of response IDs
   * @param {boolean} permanent - Whether to permanently delete
   * @param {string} userId - User performing deletion
   */
  async bulkDeleteResponses(responseIds, permanent = false, userId = null) {
    try {
      const batch = firestore.batch();
      let deletedCount = 0;

      for (const responseId of responseIds) {
        const docRef = this.collection.doc(responseId);
        const doc = await docRef.get();

        if (doc.exists) {
          if (permanent) {
            batch.delete(docRef);
          } else {
            const response = Response.fromFirestore(doc);
            response.status = 'deleted';
            response.addToHistory('bulk_deleted', userId);
            response.touch();
            batch.update(docRef, response.toFirestore());
          }
          deletedCount++;
        }
      }

      await batch.commit();

      logger.info('Bulk delete completed', {
        deletedCount,
        permanent,
        userId,
      });

      return { deleted: deletedCount };
    } catch (error) {
      logger.error('Failed to bulk delete responses:', error);
      throw ErrorHelper.handleDbError(error);
    }
  }

  /**
   * Get response statistics for a form
   * @param {string} formId - Form ID
   * @param {Object} dateRange - Date range for stats
   */
  async getResponseStats(formId, dateRange = {}) {
    try {
      let query = this.collection.where('formId', '==', formId);

      // Apply date range
      if (dateRange.startDate) {
        query = query.where('submittedAt', '>=', new Date(dateRange.startDate));
      }
      if (dateRange.endDate) {
        query = query.where('submittedAt', '<=', new Date(dateRange.endDate));
      }

      const snapshot = await query.get();
      const responses = snapshot.docs.map(doc => Response.fromFirestore(doc));

      // Calculate statistics
      const stats = {
        total: responses.length,
        completed: responses.filter(r => r.isComplete).length,
        drafts: responses.filter(r => r.isDraft).length,
        deleted: responses.filter(r => r.status === 'deleted').length,
        flagged: responses.filter(r => r.status === 'flagged').length,
        averageCompletionTime: 0,
        completionRate: 0,
        lastResponse: null,
      };

      // Calculate completion rate
      stats.completionRate = stats.total > 0 ? (stats.completed / stats.total) * 100 : 0;

      // Calculate average completion time
      const completedWithDuration = responses.filter(r => r.isComplete && r.duration);
      if (completedWithDuration.length > 0) {
        const totalDuration = completedWithDuration.reduce((sum, r) => sum + r.duration, 0);
        stats.averageCompletionTime = totalDuration / completedWithDuration.length;
      }

      // Get last response
      if (responses.length > 0) {
        const sortedResponses = responses.sort((a, b) => 
          new Date(b.submittedAt.toDate()) - new Date(a.submittedAt.toDate())
        );
        stats.lastResponse = sortedResponses[0].getSummary();
      }

      return stats;
    } catch (error) {
      logger.error('Failed to get response stats:', error);
      throw ErrorHelper.handleDbError(error);
    }
  }

  /**
   * Update analytics for a response
   * @private
   */
  async _updateAnalytics(response) {
    try {
      const date = DateHelper.formatForFilename(new Date(response.submittedAt.toDate()));
      const analyticsId = `${response.formId}_${date}`;
      
      const analyticsRef = this.analyticsCollection.doc(analyticsId);
      const analyticsDoc = await analyticsRef.get();

      if (analyticsDoc.exists) {
        const analytics = ResponseAnalytics.fromFirestore(analyticsDoc);
        analytics.updateWithResponse(response);
        await analyticsRef.update(analytics.toFirestore());
      } else {
        const analytics = new ResponseAnalytics({
          formId: response.formId,
          date,
          period: 'daily',
        });
        analytics.updateWithResponse(response);
        await analyticsRef.set(analytics.toFirestore());
      }
    } catch (error) {
      logger.error('Failed to update analytics:', error);
    }
  }

  /**
   * Trigger integrations for a response
   * @private
   */
  async _triggerIntegrations(response, action = 'created') {
    try {
      // Get active integrations for the form
      const integrationsSnapshot = await this.integrationsCollection
        .where('formId', '==', response.formId)
        .where('isActive', '==', true)
        .get();

      const integrations = integrationsSnapshot.docs.map(doc => 
        FormIntegration.fromFirestore(doc)
      );

      // Process each integration
      for (const integration of integrations) {
        try {
          // This would typically be handled by a separate integration service
          // For now, just log the integration trigger
          logger.info('Integration triggered', {
            integrationId: integration.id,
            type: integration.type,
            action,
            responseId: response.id,
          });

          // Update integration sync status
          integration.updateSyncStatus('pending');
          await this.integrationsCollection.doc(integration.id).update(integration.toFirestore());
        } catch (integrationError) {
          logger.error('Integration trigger failed:', integrationError);
        }
      }
    } catch (error) {
      logger.error('Failed to trigger integrations:', error);
    }
  }
}

module.exports = ResponseService;
