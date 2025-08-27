/**
 * Form Service Integration Layer
 * Handles communication with the form service for validating forms and fetching form schemas
 */

const axios = require('axios');
const logger = require('../utils/logger');
const { ValidationError, ServiceUnavailableError } = require('../middleware/errorHandler');

class FormServiceIntegration {
  constructor() {
    this.baseURL = process.env.FORM_SERVICE_URL || 'http://localhost:8080';
    this.timeout = parseInt(process.env.FORM_SERVICE_TIMEOUT) || 10000;
    this.apiVersion = process.env.FORM_SERVICE_API_VERSION || 'v1';
    this.retryAttempts = parseInt(process.env.FORM_SERVICE_RETRY_ATTEMPTS) || 3;
    this.retryDelay = parseInt(process.env.FORM_SERVICE_RETRY_DELAY) || 1000;
    
    // Create axios instance with default configuration
    this.client = axios.create({
      baseURL: `${this.baseURL}/api/${this.apiVersion}`,
      timeout: this.timeout,
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        'User-Agent': 'Response-Service/1.0'
      }
    });

    // Add request interceptor for logging and authentication
    this.client.interceptors.request.use(
      (config) => {
        // Add API key if available
        if (process.env.FORM_SERVICE_API_KEY) {
          config.headers['X-API-Key'] = process.env.FORM_SERVICE_API_KEY;
        }

        // Add correlation ID for tracing
        if (config.correlationId) {
          config.headers['X-Correlation-ID'] = config.correlationId;
        }

        logger.info('Form service request initiated', {
          method: config.method.toUpperCase(),
          url: config.url,
          correlationId: config.correlationId
        });

        return config;
      },
      (error) => {
        logger.error('Form service request configuration error', {
          error: error.message
        });
        return Promise.reject(error);
      }
    );

    // Add response interceptor for logging and error handling
    this.client.interceptors.response.use(
      (response) => {
        logger.info('Form service response received', {
          status: response.status,
          url: response.config.url,
          correlationId: response.config.correlationId,
          responseTime: Date.now() - response.config.metadata?.startTime
        });
        return response;
      },
      (error) => {
        const status = error.response?.status;
        const message = error.response?.data?.message || error.message;
        
        logger.error('Form service request failed', {
          status,
          message,
          url: error.config?.url,
          correlationId: error.config?.correlationId
        });

        return Promise.reject(error);
      }
    );
  }

  /**
   * Validate if a form exists and is active
   * @param {string} formId - The form ID to validate
   * @param {string} correlationId - Request correlation ID
   * @returns {Promise<Object>} Form validation result
   */
  async validateForm(formId, correlationId) {
    try {
      logger.info('Validating form with form service', { formId, correlationId });

      const response = await this.retryRequest(async () => {
        return await this.client.get(`/forms/${formId}/validate`, {
          correlationId,
          metadata: { startTime: Date.now() }
        });
      });

      const formData = response.data.data;

      logger.info('Form validation successful', {
        formId,
        isActive: formData.isActive,
        hasSchema: !!formData.schema,
        correlationId
      });

      return {
        isValid: true,
        isActive: formData.isActive,
        schema: formData.schema,
        title: formData.title,
        description: formData.description,
        settings: formData.settings
      };

    } catch (error) {
      logger.error('Form validation failed', {
        formId,
        error: error.message,
        status: error.response?.status,
        correlationId
      });

      if (error.response?.status === 404) {
        return {
          isValid: false,
          error: 'Form not found'
        };
      }

      if (error.response?.status >= 500) {
        throw new ServiceUnavailableError('Form service is currently unavailable');
      }

      throw new ValidationError(`Failed to validate form: ${error.message}`);
    }
  }

  /**
   * Get form schema for validation
   * @param {string} formId - The form ID
   * @param {string} correlationId - Request correlation ID
   * @returns {Promise<Object>} Form schema
   */
  async getFormSchema(formId, correlationId) {
    try {
      logger.info('Fetching form schema', { formId, correlationId });

      const response = await this.retryRequest(async () => {
        return await this.client.get(`/forms/${formId}/schema`, {
          correlationId,
          metadata: { startTime: Date.now() }
        });
      });

      const schema = response.data.data;

      logger.info('Form schema retrieved successfully', {
        formId,
        fieldsCount: schema.fields?.length || 0,
        correlationId
      });

      return schema;

    } catch (error) {
      logger.error('Failed to fetch form schema', {
        formId,
        error: error.message,
        status: error.response?.status,
        correlationId
      });

      if (error.response?.status === 404) {
        throw new ValidationError('Form not found');
      }

      if (error.response?.status >= 500) {
        throw new ServiceUnavailableError('Form service is currently unavailable');
      }

      throw new ValidationError(`Failed to fetch form schema: ${error.message}`);
    }
  }

  /**
   * Get form statistics from form service
   * @param {string} formId - The form ID
   * @param {string} correlationId - Request correlation ID
   * @returns {Promise<Object>} Form statistics
   */
  async getFormStatistics(formId, correlationId) {
    try {
      logger.info('Fetching form statistics', { formId, correlationId });

      const response = await this.retryRequest(async () => {
        return await this.client.get(`/forms/${formId}/statistics`, {
          correlationId,
          metadata: { startTime: Date.now() }
        });
      });

      const statistics = response.data.data;

      logger.info('Form statistics retrieved successfully', {
        formId,
        views: statistics.views,
        submissions: statistics.submissions,
        correlationId
      });

      return statistics;

    } catch (error) {
      logger.warn('Failed to fetch form statistics', {
        formId,
        error: error.message,
        status: error.response?.status,
        correlationId
      });

      // Return empty statistics if service is unavailable
      return {
        views: 0,
        submissions: 0,
        lastSubmission: null
      };
    }
  }

  /**
   * Notify form service about new response submission
   * @param {string} formId - The form ID
   * @param {string} responseId - The response ID
   * @param {string} correlationId - Request correlation ID
   * @returns {Promise<boolean>} Success status
   */
  async notifyResponseSubmission(formId, responseId, correlationId) {
    try {
      logger.info('Notifying form service about response submission', {
        formId,
        responseId,
        correlationId
      });

      await this.retryRequest(async () => {
        return await this.client.post(`/forms/${formId}/submissions/notify`, {
          responseId,
          submittedAt: new Date().toISOString()
        }, {
          correlationId,
          metadata: { startTime: Date.now() }
        });
      });

      logger.info('Form service notified successfully', {
        formId,
        responseId,
        correlationId
      });

      return true;

    } catch (error) {
      logger.warn('Failed to notify form service', {
        formId,
        responseId,
        error: error.message,
        status: error.response?.status,
        correlationId
      });

      // Don't fail the response submission if notification fails
      return false;
    }
  }

  /**
   * Get list of forms for analytics
   * @param {string} correlationId - Request correlation ID
   * @returns {Promise<Array>} List of forms
   */
  async getFormsList(correlationId) {
    try {
      logger.info('Fetching forms list for analytics', { correlationId });

      const response = await this.retryRequest(async () => {
        return await this.client.get('/forms', {
          params: {
            status: 'active',
            limit: 1000
          },
          correlationId,
          metadata: { startTime: Date.now() }
        });
      });

      const forms = response.data.data;

      logger.info('Forms list retrieved successfully', {
        formsCount: forms.length,
        correlationId
      });

      return forms;

    } catch (error) {
      logger.warn('Failed to fetch forms list', {
        error: error.message,
        status: error.response?.status,
        correlationId
      });

      // Return empty array if service is unavailable
      return [];
    }
  }

  /**
   * Retry mechanism for HTTP requests
   * @param {Function} requestFunction - The request function to retry
   * @param {number} attempts - Number of retry attempts
   * @returns {Promise} Request result
   */
  async retryRequest(requestFunction, attempts = this.retryAttempts) {
    let lastError;

    for (let i = 0; i < attempts; i++) {
      try {
        return await requestFunction();
      } catch (error) {
        lastError = error;

        // Don't retry on client errors (4xx)
        if (error.response?.status && error.response.status < 500) {
          throw error;
        }

        // Wait before retrying (exponential backoff)
        if (i < attempts - 1) {
          const delay = this.retryDelay * Math.pow(2, i);
          logger.warn(`Request failed, retrying in ${delay}ms`, {
            attempt: i + 1,
            maxAttempts: attempts,
            error: error.message
          });
          await this.sleep(delay);
        }
      }
    }

    throw lastError;
  }

  /**
   * Sleep utility for retry delays
   * @param {number} ms - Milliseconds to sleep
   * @returns {Promise}
   */
  sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  /**
   * Check form service health
   * @param {string} correlationId - Request correlation ID
   * @returns {Promise<Object>} Health status
   */
  async checkHealth(correlationId) {
    try {
      const startTime = Date.now();
      
      const response = await this.client.get('/health', {
        timeout: 5000, // Shorter timeout for health checks
        correlationId,
        metadata: { startTime }
      });

      const responseTime = Date.now() - startTime;

      return {
        status: 'healthy',
        responseTime,
        version: response.data.version
      };

    } catch (error) {
      logger.error('Form service health check failed', {
        error: error.message,
        correlationId
      });

      return {
        status: 'unhealthy',
        error: error.message,
        responseTime: null
      };
    }
  }
}

// Export singleton instance
module.exports = new FormServiceIntegration();
