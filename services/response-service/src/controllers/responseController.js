/**
 * Response Controller for Response Service
 * Handles all response-related business logic with comprehensive validation and error handling
 */

const { createSuccessResponse, createErrorResponse } = require('../dto/response-dtos');
const { NotFoundError, ValidationError, ForbiddenError, ConflictError } = require('../middleware/errorHandler');
const logger = require('../utils/logger');

// Mock database operations (replace with actual database integration)
const mockDatabase = {
  responses: new Map(),
  forms: new Map()
};

// Initialize with some mock data
mockDatabase.forms.set('f123e4567-e89b-12d3-a456-426614174000', {
  id: 'f123e4567-e89b-12d3-a456-426614174000',
  title: 'Customer Feedback Survey',
  status: 'active',
  organizationId: 'org123'
});

/**
 * Create a new form response
 */
const createResponse = async (req, res) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  const startTime = Date.now();
  
  try {
    const { formId, respondentId, respondentEmail, respondentName, responses, metadata, isDraft, isPartial } = req.body;
    
    logger.info('Creating new response', {
      correlationId,
      formId,
      respondentId,
      responseCount: responses?.length,
      isDraft,
      isPartial
    });

    // Validate form exists and is active
    const form = mockDatabase.forms.get(formId);
    if (!form) {
      throw new NotFoundError('Form not found');
    }

    if (form.status !== 'active') {
      throw new ValidationError('Form is not accepting responses', {
        formStatus: form.status
      });
    }

    // Generate response ID
    const responseId = `r${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    // Create response object
    const newResponse = {
      id: responseId,
      formId,
      formTitle: form.title,
      respondentId: respondentId || null,
      respondentEmail: respondentEmail || null,
      respondentName: respondentName || null,
      responses: responses.map(response => ({
        ...response,
        id: `resp_${Date.now()}_${Math.random().toString(36).substr(2, 6)}`
      })),
      status: isDraft ? 'draft' : (isPartial ? 'partial' : 'completed'),
      isDraft: isDraft || false,
      isPartial: isPartial || false,
      metadata: {
        ...metadata,
        submissionSource: 'api',
        ipAddress: req.ip,
        userAgent: req.get('User-Agent')
      },
      submittedAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      version: 1
    };

    // Store in mock database
    mockDatabase.responses.set(responseId, newResponse);

    const duration = Date.now() - startTime;
    
    logger.logBusiness('RESPONSE_CREATED', 'response', responseId, {
      formId,
      respondentId,
      status: newResponse.status,
      responseCount: responses.length
    }, { correlationId });

    logger.info('Response created successfully', {
      correlationId,
      responseId,
      formId,
      duration
    });

    res.status(201).json(
      createSuccessResponse(
        newResponse,
        'Response submitted successfully',
        correlationId
      )
    );

  } catch (error) {
    const duration = Date.now() - startTime;
    
    logger.error('Failed to create response', {
      correlationId,
      error: error.message,
      duration,
      formId: req.body?.formId
    });

    throw error;
  }
};

/**
 * Get responses with filtering and pagination
 */
const getResponses = async (req, res) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  const startTime = Date.now();
  
  try {
    const {
      formId,
      respondentEmail,
      status,
      startDate,
      endDate,
      page = 1,
      limit = 20,
      sortBy = 'submittedAt',
      sortOrder = 'desc',
      search
    } = req.query;

    logger.debug('Fetching responses', {
      correlationId,
      filters: { formId, respondentEmail, status, startDate, endDate },
      pagination: { page, limit },
      sorting: { sortBy, sortOrder },
      search
    });

    // Get all responses and apply filters
    let filteredResponses = Array.from(mockDatabase.responses.values());

    // Apply filters
    if (formId) {
      filteredResponses = filteredResponses.filter(response => response.formId === formId);
    }

    if (respondentEmail) {
      filteredResponses = filteredResponses.filter(response => 
        response.respondentEmail?.toLowerCase().includes(respondentEmail.toLowerCase())
      );
    }

    if (status) {
      filteredResponses = filteredResponses.filter(response => response.status === status);
    }

    if (startDate) {
      const start = new Date(startDate);
      filteredResponses = filteredResponses.filter(response => 
        new Date(response.submittedAt) >= start
      );
    }

    if (endDate) {
      const end = new Date(endDate);
      filteredResponses = filteredResponses.filter(response => 
        new Date(response.submittedAt) <= end
      );
    }

    if (search) {
      filteredResponses = filteredResponses.filter(response =>
        response.responses.some(resp => 
          resp.textValue?.toLowerCase().includes(search.toLowerCase())
        )
      );
    }

    // Apply sorting
    filteredResponses.sort((a, b) => {
      let aValue = a[sortBy];
      let bValue = b[sortBy];
      
      if (sortBy === 'submittedAt' || sortBy === 'updatedAt') {
        aValue = new Date(aValue);
        bValue = new Date(bValue);
      }
      
      if (sortOrder === 'desc') {
        return bValue > aValue ? 1 : -1;
      } else {
        return aValue > bValue ? 1 : -1;
      }
    });

    // Apply pagination
    const totalItems = filteredResponses.length;
    const totalPages = Math.ceil(totalItems / limit);
    const startIndex = (page - 1) * limit;
    const endIndex = startIndex + limit;
    
    const paginatedResponses = filteredResponses.slice(startIndex, endIndex);

    // Create response summaries
    const responseSummaries = paginatedResponses.map(response => ({
      id: response.id,
      formId: response.formId,
      formTitle: response.formTitle,
      respondentEmail: response.respondentEmail,
      respondentName: response.respondentName,
      status: response.status,
      submittedAt: response.submittedAt,
      updatedAt: response.updatedAt,
      responseCount: response.responses.length
    }));

    const result = {
      responses: responseSummaries,
      pagination: {
        currentPage: parseInt(page),
        totalPages,
        totalItems,
        itemsPerPage: parseInt(limit),
        hasNext: page < totalPages,
        hasPrevious: page > 1
      }
    };

    const duration = Date.now() - startTime;

    logger.info('Responses fetched successfully', {
      correlationId,
      totalItems,
      returnedItems: responseSummaries.length,
      page,
      duration
    });

    res.json(
      createSuccessResponse(
        result,
        'Responses retrieved successfully',
        correlationId
      )
    );

  } catch (error) {
    const duration = Date.now() - startTime;
    
    logger.error('Failed to fetch responses', {
      correlationId,
      error: error.message,
      duration
    });

    throw error;
  }
};

/**
 * Get a specific response by ID
 */
const getResponse = async (req, res) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  const { id } = req.params;
  const startTime = Date.now();
  
  try {
    logger.debug('Fetching response', {
      correlationId,
      responseId: id
    });

    const response = mockDatabase.responses.get(id);
    if (!response) {
      throw new NotFoundError('Response not found');
    }

    // Check authorization (users can only view their own responses unless admin)
    const userRole = req.user?.role;
    const userId = req.user?.id;
    
    if (userRole !== 'admin' && userRole !== 'form_manager' && userRole !== 'analyst') {
      if (response.respondentId && response.respondentId !== userId) {
        throw new ForbiddenError('Access denied to this response');
      }
    }

    const duration = Date.now() - startTime;

    logger.info('Response fetched successfully', {
      correlationId,
      responseId: id,
      formId: response.formId,
      duration
    });

    res.json(
      createSuccessResponse(
        response,
        'Response retrieved successfully',
        correlationId
      )
    );

  } catch (error) {
    const duration = Date.now() - startTime;
    
    logger.error('Failed to fetch response', {
      correlationId,
      responseId: id,
      error: error.message,
      duration
    });

    throw error;
  }
};

/**
 * Update an existing response
 */
const updateResponse = async (req, res) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  const { id } = req.params;
  const startTime = Date.now();
  
  try {
    logger.info('Updating response', {
      correlationId,
      responseId: id
    });

    const existingResponse = mockDatabase.responses.get(id);
    if (!existingResponse) {
      throw new NotFoundError('Response not found');
    }

    // Check if response can be updated
    if (existingResponse.status === 'completed' && req.user?.role !== 'admin') {
      throw new ForbiddenError('Cannot update completed responses');
    }

    // Check authorization
    const userRole = req.user?.role;
    const userId = req.user?.id;
    
    if (userRole !== 'admin' && userRole !== 'form_manager') {
      if (existingResponse.respondentId && existingResponse.respondentId !== userId) {
        throw new ForbiddenError('Access denied to update this response');
      }
    }

    const { responses, respondentEmail, respondentName, metadata, isDraft, isPartial } = req.body;

    // Update response
    const updatedResponse = {
      ...existingResponse,
      ...(responses && { responses: responses.map(response => ({
        ...response,
        id: response.id || `resp_${Date.now()}_${Math.random().toString(36).substr(2, 6)}`
      })) }),
      ...(respondentEmail !== undefined && { respondentEmail }),
      ...(respondentName !== undefined && { respondentName }),
      ...(metadata && { metadata: { ...existingResponse.metadata, ...metadata } }),
      ...(isDraft !== undefined && { isDraft }),
      ...(isPartial !== undefined && { isPartial }),
      status: isDraft ? 'draft' : (isPartial ? 'partial' : 'completed'),
      updatedAt: new Date().toISOString(),
      version: existingResponse.version + 1
    };

    // Store updated response
    mockDatabase.responses.set(id, updatedResponse);

    const duration = Date.now() - startTime;

    logger.logBusiness('RESPONSE_UPDATED', 'response', id, {
      formId: updatedResponse.formId,
      status: updatedResponse.status,
      version: updatedResponse.version
    }, { correlationId });

    logger.info('Response updated successfully', {
      correlationId,
      responseId: id,
      formId: updatedResponse.formId,
      duration
    });

    res.json(
      createSuccessResponse(
        updatedResponse,
        'Response updated successfully',
        correlationId
      )
    );

  } catch (error) {
    const duration = Date.now() - startTime;
    
    logger.error('Failed to update response', {
      correlationId,
      responseId: id,
      error: error.message,
      duration
    });

    throw error;
  }
};

/**
 * Delete a response
 */
const deleteResponse = async (req, res) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  const { id } = req.params;
  const startTime = Date.now();
  
  try {
    logger.info('Deleting response', {
      correlationId,
      responseId: id
    });

    const response = mockDatabase.responses.get(id);
    if (!response) {
      throw new NotFoundError('Response not found');
    }

    // Check authorization
    const userRole = req.user?.role;
    const userId = req.user?.id;
    
    if (userRole !== 'admin' && userRole !== 'form_manager') {
      if (response.respondentId && response.respondentId !== userId) {
        throw new ForbiddenError('Access denied to delete this response');
      }
      
      // Regular users can only delete draft responses
      if (response.status !== 'draft') {
        throw new ForbiddenError('Can only delete draft responses');
      }
    }

    // Delete the response
    mockDatabase.responses.delete(id);

    const duration = Date.now() - startTime;

    logger.logBusiness('RESPONSE_DELETED', 'response', id, {
      formId: response.formId,
      status: response.status
    }, { correlationId });

    logger.info('Response deleted successfully', {
      correlationId,
      responseId: id,
      formId: response.formId,
      duration
    });

    res.status(204).send();

  } catch (error) {
    const duration = Date.now() - startTime;
    
    logger.error('Failed to delete response', {
      correlationId,
      responseId: id,
      error: error.message,
      duration
    });

    throw error;
  }
};

/**
 * Get responses for a specific form
 */
const getFormResponses = async (req, res) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  const { formId } = req.params;
  
  try {
    // Delegate to getResponses with formId filter
    req.query.formId = formId;
    await getResponses(req, res);
  } catch (error) {
    logger.error('Failed to fetch form responses', {
      correlationId,
      formId,
      error: error.message
    });

    throw error;
  }
};

/**
 * Export responses in various formats
 */
const exportResponses = async (req, res) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  const { formId } = req.params;
  const { format = 'csv', includeMetadata = false } = req.query;
  const startTime = Date.now();
  
  try {
    logger.info('Exporting responses', {
      correlationId,
      formId,
      format,
      includeMetadata
    });

    // Get form responses
    const responses = Array.from(mockDatabase.responses.values())
      .filter(response => response.formId === formId);

    if (responses.length === 0) {
      throw new NotFoundError('No responses found for this form');
    }

    const form = mockDatabase.forms.get(formId);
    const filename = `${form?.title || 'form'}_responses_${new Date().toISOString().split('T')[0]}`;

    let contentType;
    let fileExtension;
    let exportData;

    switch (format.toLowerCase()) {
      case 'csv':
        contentType = 'text/csv';
        fileExtension = 'csv';
        exportData = convertToCSV(responses, includeMetadata);
        break;
      
      case 'excel':
        contentType = 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet';
        fileExtension = 'xlsx';
        exportData = convertToExcel(responses, includeMetadata);
        break;
      
      case 'json':
        contentType = 'application/json';
        fileExtension = 'json';
        exportData = JSON.stringify(responses, null, 2);
        break;
      
      default:
        throw new ValidationError('Unsupported export format', {
          supportedFormats: ['csv', 'excel', 'json']
        });
    }

    const duration = Date.now() - startTime;

    logger.logBusiness('RESPONSES_EXPORTED', 'form', formId, {
      format,
      responseCount: responses.length,
      includeMetadata
    }, { correlationId });

    logger.info('Responses exported successfully', {
      correlationId,
      formId,
      format,
      responseCount: responses.length,
      duration
    });

    res.setHeader('Content-Type', contentType);
    res.setHeader('Content-Disposition', `attachment; filename="${filename}.${fileExtension}"`);
    res.send(exportData);

  } catch (error) {
    const duration = Date.now() - startTime;
    
    logger.error('Failed to export responses', {
      correlationId,
      formId,
      error: error.message,
      duration
    });

    throw error;
  }
};

/**
 * Helper function to convert responses to CSV format
 */
function convertToCSV(responses, includeMetadata) {
  if (responses.length === 0) return '';

  // Get all unique question IDs
  const questionIds = new Set();
  responses.forEach(response => {
    response.responses.forEach(resp => {
      questionIds.add(resp.questionId);
    });
  });

  // Create CSV header
  const headers = [
    'Response ID',
    'Form ID',
    'Respondent Email',
    'Respondent Name',
    'Status',
    'Submitted At',
    'Updated At',
    ...Array.from(questionIds).map(id => `Question_${id}`)
  ];

  if (includeMetadata) {
    headers.push('IP Address', 'User Agent', 'Time Spent');
  }

  // Create CSV rows
  const rows = [headers];
  
  responses.forEach(response => {
    const row = [
      response.id,
      response.formId,
      response.respondentEmail || '',
      response.respondentName || '',
      response.status,
      response.submittedAt,
      response.updatedAt
    ];

    // Add question responses
    questionIds.forEach(questionId => {
      const questionResponse = response.responses.find(resp => resp.questionId === questionId);
      row.push(questionResponse ? questionResponse.textValue || JSON.stringify(questionResponse.value) : '');
    });

    if (includeMetadata) {
      row.push(
        response.metadata?.ipAddress || '',
        response.metadata?.userAgent || '',
        response.metadata?.timeSpent || ''
      );
    }

    rows.push(row);
  });

  // Convert to CSV string
  return rows.map(row => 
    row.map(cell => `"${String(cell).replace(/"/g, '""')}"`).join(',')
  ).join('\n');
}

/**
 * Helper function to convert responses to Excel format (simplified)
 */
function convertToExcel(responses, includeMetadata) {
  // For simplicity, return CSV format with Excel content type
  // In a real implementation, you would use a library like 'exceljs'
  return convertToCSV(responses, includeMetadata);
}

module.exports = {
  createResponse,
  getResponses,
  getResponse,
  updateResponse,
  deleteResponse,
  getFormResponses,
  exportResponses
};
