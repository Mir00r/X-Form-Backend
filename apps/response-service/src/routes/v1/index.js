/**
 * API Version 1 Routes for Response Service
 * Implements versioned RESTful endpoints with comprehensive middleware
 */

const express = require('express');
const responseController = require('../../controllers/responseController');
const healthController = require('../../controllers/healthController');
const analyticsController = require('../../controllers/analyticsController');

// Middleware
const { authenticate, optionalAuthentication, authorize } = require('../../middleware/auth');
const { 
  validateCreateResponse,
  validateUpdateResponse,
  validateResponseListQuery,
  validateResponseId,
  validateFormId,
  validateFileUpload,
  validateContentType,
  sanitizeInput
} = require('../../middleware/validation');
const { submissionRateLimit } = require('../../middleware/security');
const { asyncHandler } = require('../../middleware/errorHandler');

const router = express.Router();

// =============================================================================
// Health Check Routes
// =============================================================================

/**
 * @swagger
 * /api/v1/health:
 *   get:
 *     tags: [Health]
 *     summary: Service health check
 *     description: Returns the health status of the response service and its dependencies
 *     responses:
 *       200:
 *         description: Service is healthy
 *         content:
 *           application/json:
 *             schema:
 *               $ref: '#/components/schemas/HealthResponse'
 *       503:
 *         description: Service is unhealthy
 *         content:
 *           application/json:
 *             schema:
 *               $ref: '#/components/schemas/ErrorResponse'
 */
router.get('/health', asyncHandler(healthController.getHealth));

/**
 * @swagger
 * /api/v1/health/ready:
 *   get:
 *     tags: [Health]
 *     summary: Service readiness check
 *     description: Returns whether the service is ready to accept requests
 *     responses:
 *       200:
 *         description: Service is ready
 *       503:
 *         description: Service is not ready
 */
router.get('/health/ready', asyncHandler(healthController.getReadiness));

/**
 * @swagger
 * /api/v1/health/live:
 *   get:
 *     tags: [Health]
 *     summary: Service liveness check
 *     description: Returns whether the service is alive
 *     responses:
 *       200:
 *         description: Service is alive
 *       503:
 *         description: Service is not alive
 */
router.get('/health/live', asyncHandler(healthController.getLiveness));

// =============================================================================
// Response Routes
// =============================================================================

/**
 * @swagger
 * /api/v1/responses:
 *   post:
 *     tags: [Responses]
 *     summary: Submit a new form response
 *     description: Submit a new response to a form with validation and security checks
 *     security:
 *       - bearerAuth: []
 *       - apiKeyAuth: []
 *     requestBody:
 *       required: true
 *       content:
 *         application/json:
 *           schema:
 *             $ref: '#/components/schemas/CreateResponseRequest'
 *         multipart/form-data:
 *           schema:
 *             allOf:
 *               - $ref: '#/components/schemas/CreateResponseRequest'
 *               - type: object
 *                 properties:
 *                   files:
 *                     type: array
 *                     items:
 *                       type: string
 *                       format: binary
 *     responses:
 *       201:
 *         description: Response submitted successfully
 *         content:
 *           application/json:
 *             schema:
 *               allOf:
 *                 - $ref: '#/components/schemas/SuccessResponse'
 *                 - type: object
 *                   properties:
 *                     data:
 *                       $ref: '#/components/schemas/ResponseResponse'
 *       400:
 *         $ref: '#/components/responses/ValidationError'
 *       401:
 *         $ref: '#/components/responses/UnauthorizedError'
 *       429:
 *         $ref: '#/components/responses/RateLimitError'
 */
router.post('/responses',
  submissionRateLimit,
  optionalAuthentication,
  validateContentType(['application/json', 'multipart/form-data']),
  sanitizeInput,
  validateCreateResponse,
  validateFileUpload,
  asyncHandler(responseController.createResponse)
);

/**
 * @swagger
 * /api/v1/responses:
 *   get:
 *     tags: [Responses]
 *     summary: List form responses
 *     description: Retrieve a paginated list of form responses with filtering and sorting
 *     security:
 *       - bearerAuth: []
 *       - apiKeyAuth: []
 *     parameters:
 *       - in: query
 *         name: formId
 *         schema:
 *           type: string
 *           format: uuid
 *         description: Filter by form ID
 *       - in: query
 *         name: respondentEmail
 *         schema:
 *           type: string
 *           format: email
 *         description: Filter by respondent email
 *       - in: query
 *         name: status
 *         schema:
 *           type: string
 *           enum: [draft, partial, completed, archived]
 *         description: Filter by response status
 *       - in: query
 *         name: startDate
 *         schema:
 *           type: string
 *           format: date-time
 *         description: Filter responses submitted after this date
 *       - in: query
 *         name: endDate
 *         schema:
 *           type: string
 *           format: date-time
 *         description: Filter responses submitted before this date
 *       - in: query
 *         name: page
 *         schema:
 *           type: integer
 *           minimum: 1
 *           default: 1
 *         description: Page number
 *       - in: query
 *         name: limit
 *         schema:
 *           type: integer
 *           minimum: 1
 *           maximum: 100
 *           default: 20
 *         description: Number of items per page
 *       - in: query
 *         name: sortBy
 *         schema:
 *           type: string
 *           enum: [submittedAt, updatedAt, respondentEmail, status]
 *           default: submittedAt
 *         description: Field to sort by
 *       - in: query
 *         name: sortOrder
 *         schema:
 *           type: string
 *           enum: [asc, desc]
 *           default: desc
 *         description: Sort order
 *       - in: query
 *         name: search
 *         schema:
 *           type: string
 *         description: Search in response text values
 *     responses:
 *       200:
 *         description: Responses retrieved successfully
 *         content:
 *           application/json:
 *             schema:
 *               allOf:
 *                 - $ref: '#/components/schemas/SuccessResponse'
 *                 - type: object
 *                   properties:
 *                     data:
 *                       $ref: '#/components/schemas/ResponseListResponse'
 *       400:
 *         $ref: '#/components/responses/ValidationError'
 *       401:
 *         $ref: '#/components/responses/UnauthorizedError'
 *       403:
 *         $ref: '#/components/responses/ForbiddenError'
 */
router.get('/responses',
  authenticate,
  authorize('admin', 'form_manager', 'analyst'),
  validateResponseListQuery,
  asyncHandler(responseController.getResponses)
);

/**
 * @swagger
 * /api/v1/responses/{id}:
 *   get:
 *     tags: [Responses]
 *     summary: Get a specific response
 *     description: Retrieve detailed information about a specific form response
 *     security:
 *       - bearerAuth: []
 *       - apiKeyAuth: []
 *     parameters:
 *       - in: path
 *         name: id
 *         required: true
 *         schema:
 *           type: string
 *           format: uuid
 *         description: Response ID
 *     responses:
 *       200:
 *         description: Response retrieved successfully
 *         content:
 *           application/json:
 *             schema:
 *               allOf:
 *                 - $ref: '#/components/schemas/SuccessResponse'
 *                 - type: object
 *                   properties:
 *                     data:
 *                       $ref: '#/components/schemas/ResponseResponse'
 *       404:
 *         $ref: '#/components/responses/NotFoundError'
 *       401:
 *         $ref: '#/components/responses/UnauthorizedError'
 *       403:
 *         $ref: '#/components/responses/ForbiddenError'
 */
router.get('/responses/:id',
  authenticate,
  validateResponseId,
  asyncHandler(responseController.getResponse)
);

/**
 * @swagger
 * /api/v1/responses/{id}:
 *   put:
 *     tags: [Responses]
 *     summary: Update a form response
 *     description: Update an existing form response (only allowed for drafts or by admin)
 *     security:
 *       - bearerAuth: []
 *       - apiKeyAuth: []
 *     parameters:
 *       - in: path
 *         name: id
 *         required: true
 *         schema:
 *           type: string
 *           format: uuid
 *         description: Response ID
 *     requestBody:
 *       required: true
 *       content:
 *         application/json:
 *           schema:
 *             $ref: '#/components/schemas/UpdateResponseRequest'
 *         multipart/form-data:
 *           schema:
 *             allOf:
 *               - $ref: '#/components/schemas/UpdateResponseRequest'
 *               - type: object
 *                 properties:
 *                   files:
 *                     type: array
 *                     items:
 *                       type: string
 *                       format: binary
 *     responses:
 *       200:
 *         description: Response updated successfully
 *         content:
 *           application/json:
 *             schema:
 *               allOf:
 *                 - $ref: '#/components/schemas/SuccessResponse'
 *                 - type: object
 *                   properties:
 *                     data:
 *                       $ref: '#/components/schemas/ResponseResponse'
 *       400:
 *         $ref: '#/components/responses/ValidationError'
 *       404:
 *         $ref: '#/components/responses/NotFoundError'
 *       401:
 *         $ref: '#/components/responses/UnauthorizedError'
 *       403:
 *         $ref: '#/components/responses/ForbiddenError'
 */
router.put('/responses/:id',
  authenticate,
  validateResponseId,
  validateContentType(['application/json', 'multipart/form-data']),
  sanitizeInput,
  validateUpdateResponse,
  validateFileUpload,
  asyncHandler(responseController.updateResponse)
);

/**
 * @swagger
 * /api/v1/responses/{id}:
 *   delete:
 *     tags: [Responses]
 *     summary: Delete a form response
 *     description: Delete a form response (admin only or response owner for drafts)
 *     security:
 *       - bearerAuth: []
 *       - apiKeyAuth: []
 *     parameters:
 *       - in: path
 *         name: id
 *         required: true
 *         schema:
 *           type: string
 *           format: uuid
 *         description: Response ID
 *     responses:
 *       204:
 *         description: Response deleted successfully
 *       404:
 *         $ref: '#/components/responses/NotFoundError'
 *       401:
 *         $ref: '#/components/responses/UnauthorizedError'
 *       403:
 *         $ref: '#/components/responses/ForbiddenError'
 */
router.delete('/responses/:id',
  authenticate,
  authorize('admin', 'form_manager'),
  validateResponseId,
  asyncHandler(responseController.deleteResponse)
);

// =============================================================================
// Form-specific Response Routes
// =============================================================================

/**
 * @swagger
 * /api/v1/forms/{formId}/responses:
 *   get:
 *     tags: [Responses]
 *     summary: Get responses for a specific form
 *     description: Retrieve all responses for a specific form with pagination and filtering
 *     security:
 *       - bearerAuth: []
 *       - apiKeyAuth: []
 *     parameters:
 *       - in: path
 *         name: formId
 *         required: true
 *         schema:
 *           type: string
 *           format: uuid
 *         description: Form ID
 *       - $ref: '#/components/parameters/PageParam'
 *       - $ref: '#/components/parameters/LimitParam'
 *       - $ref: '#/components/parameters/SortByParam'
 *       - $ref: '#/components/parameters/SortOrderParam'
 *     responses:
 *       200:
 *         description: Form responses retrieved successfully
 *         content:
 *           application/json:
 *             schema:
 *               allOf:
 *                 - $ref: '#/components/schemas/SuccessResponse'
 *                 - type: object
 *                   properties:
 *                     data:
 *                       $ref: '#/components/schemas/ResponseListResponse'
 *       404:
 *         $ref: '#/components/responses/NotFoundError'
 *       401:
 *         $ref: '#/components/responses/UnauthorizedError'
 *       403:
 *         $ref: '#/components/responses/ForbiddenError'
 */
router.get('/forms/:formId/responses',
  authenticate,
  authorize('admin', 'form_manager', 'analyst'),
  validateFormId,
  validateResponseListQuery,
  asyncHandler(responseController.getFormResponses)
);

/**
 * @swagger
 * /api/v1/forms/{formId}/responses/analytics:
 *   get:
 *     tags: [Analytics]
 *     summary: Get response analytics for a form
 *     description: Retrieve analytics and statistics for form responses
 *     security:
 *       - bearerAuth: []
 *       - apiKeyAuth: []
 *     parameters:
 *       - in: path
 *         name: formId
 *         required: true
 *         schema:
 *           type: string
 *           format: uuid
 *         description: Form ID
 *     responses:
 *       200:
 *         description: Analytics retrieved successfully
 *         content:
 *           application/json:
 *             schema:
 *               allOf:
 *                 - $ref: '#/components/schemas/SuccessResponse'
 *                 - type: object
 *                   properties:
 *                     data:
 *                       $ref: '#/components/schemas/ResponseAnalytics'
 *       404:
 *         $ref: '#/components/responses/NotFoundError'
 *       401:
 *         $ref: '#/components/responses/UnauthorizedError'
 *       403:
 *         $ref: '#/components/responses/ForbiddenError'
 */
router.get('/forms/:formId/responses/analytics',
  authenticate,
  authorize('admin', 'form_manager', 'analyst'),
  validateFormId,
  asyncHandler(analyticsController.getFormAnalytics)
);

/**
 * @swagger
 * /api/v1/forms/{formId}/responses/export:
 *   get:
 *     tags: [Responses]
 *     summary: Export form responses
 *     description: Export form responses in various formats (CSV, Excel, JSON)
 *     security:
 *       - bearerAuth: []
 *       - apiKeyAuth: []
 *     parameters:
 *       - in: path
 *         name: formId
 *         required: true
 *         schema:
 *           type: string
 *           format: uuid
 *         description: Form ID
 *       - in: query
 *         name: format
 *         schema:
 *           type: string
 *           enum: [csv, excel, json]
 *           default: csv
 *         description: Export format
 *       - in: query
 *         name: includeMetadata
 *         schema:
 *           type: boolean
 *           default: false
 *         description: Include metadata in export
 *     responses:
 *       200:
 *         description: Export file
 *         content:
 *           text/csv:
 *             schema:
 *               type: string
 *           application/vnd.openxmlformats-officedocument.spreadsheetml.sheet:
 *             schema:
 *               type: string
 *               format: binary
 *           application/json:
 *             schema:
 *               type: object
 *       404:
 *         $ref: '#/components/responses/NotFoundError'
 *       401:
 *         $ref: '#/components/responses/UnauthorizedError'
 *       403:
 *         $ref: '#/components/responses/ForbiddenError'
 */
router.get('/forms/:formId/responses/export',
  authenticate,
  authorize('admin', 'form_manager'),
  validateFormId,
  asyncHandler(responseController.exportResponses)
);

module.exports = router;
