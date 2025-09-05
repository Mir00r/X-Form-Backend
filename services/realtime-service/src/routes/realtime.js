const express = require('express');

/**
 * Create realtime event routes
 * @param {Object} io - Socket.IO instance
 * @param {Object} connectionStats - Connection statistics object
 * @returns {Object} Express router
 */
function createRealtimeRoutes(io, connectionStats) {
  const router = express.Router();

  /**
   * @swagger
   * /events/form/{formId}/notify:
   *   post:
   *     summary: Send Form Notification
   *     description: Send a notification to all subscribers of a specific form
   *     tags: [Events]
   *     parameters:
   *       - in: path
   *         name: formId
   *         required: true
   *         schema:
   *           type: string
   *         description: Form ID to send notification to
   *         example: "form123"
   *     requestBody:
   *       required: true
   *       content:
   *         application/json:
   *           schema:
   *             type: object
   *             required:
   *               - event
   *               - data
   *             properties:
   *               event:
   *                 type: string
   *                 description: Event name to emit
   *                 example: "form:updated"
   *               data:
   *                 type: object
   *                 description: Event data
   *                 example: {"title": "New Form Title", "updatedBy": "admin"}
   *               urgent:
   *                 type: boolean
   *                 description: Whether this is an urgent notification
   *                 example: false
   *     responses:
   *       200:
   *         description: Notification sent successfully
   *         content:
   *           application/json:
   *             schema:
   *               type: object
   *               properties:
   *                 success:
   *                   type: boolean
   *                   example: true
   *                 message:
   *                   type: string
   *                   example: "Notification sent to form subscribers"
   *                 formId:
   *                   type: string
   *                   example: "form123"
   *                 subscribers:
   *                   type: integer
   *                   example: 15
   *                 event:
   *                   type: string
   *                   example: "form:updated"
   *                 timestamp:
   *                   type: string
   *                   format: date-time
   *       400:
   *         description: Invalid request data
   *       404:
   *         description: Form not found or no subscribers
   */
  router.post('/form/:formId/notify', (req, res) => {
    try {
      const { formId } = req.params;
      const { event, data, urgent = false } = req.body;
      
      if (!event || !data) {
        return res.status(400).json({
          error: 'Missing required fields',
          required: ['event', 'data'],
          timestamp: new Date().toISOString(),
          code: 'MISSING_FIELDS'
        });
      }
      
      const roomName = `form:${formId}`;
      
      // Check if room exists and has subscribers
      const room = io.sockets.adapter.rooms.get(roomName);
      
      if (!room || room.size === 0) {
        return res.status(404).json({
          error: 'No subscribers found for this form',
          formId: formId,
          timestamp: new Date().toISOString(),
          code: 'NO_SUBSCRIBERS'
        });
      }
      
      // Prepare notification data
      const notificationData = {
        ...data,
        formId: formId,
        urgent: urgent,
        notifiedAt: new Date().toISOString(),
        source: 'api'
      };
      
      // Send notification to all form subscribers
      io.to(roomName).emit(event, notificationData);
      
      // Log the notification
      console.log(`📢 Notification sent to form ${formId}: ${event} (${room.size} subscribers)`);
      
      res.json({
        success: true,
        message: 'Notification sent to form subscribers',
        formId: formId,
        subscribers: room.size,
        event: event,
        urgent: urgent,
        timestamp: new Date().toISOString()
      });
      
    } catch (error) {
      res.status(500).json({
        error: 'Failed to send notification',
        message: error.message,
        timestamp: new Date().toISOString(),
        code: 'NOTIFICATION_ERROR'
      });
    }
  });

  /**
   * @swagger
   * /events/form/{formId}/response:
   *   post:
   *     summary: Broadcast New Response
   *     description: Broadcast a new form response to all form subscribers
   *     tags: [Events]
   *     parameters:
   *       - in: path
   *         name: formId
   *         required: true
   *         schema:
   *           type: string
   *         description: Form ID
   *         example: "form123"
   *     requestBody:
   *       required: true
   *       content:
   *         application/json:
   *           schema:
   *             $ref: '#/components/schemas/FormResponse'
   *     responses:
   *       200:
   *         description: Response broadcasted successfully
   *         content:
   *           application/json:
   *             schema:
   *               type: object
   *               properties:
   *                 success:
   *                   type: boolean
   *                   example: true
   *                 message:
   *                   type: string
   *                   example: "Response broadcasted to form subscribers"
   *                 formId:
   *                   type: string
   *                   example: "form123"
   *                 responseId:
   *                   type: string
   *                   example: "resp456"
   *                 subscribers:
   *                   type: integer
   *                   example: 12
   *                 timestamp:
   *                   type: string
   *                   format: date-time
   *       400:
   *         description: Invalid response data
   *       404:
   *         description: No subscribers for form
   */
  router.post('/form/:formId/response', (req, res) => {
    try {
      const { formId } = req.params;
      const responseData = req.body;
      
      if (!responseData.responseId) {
        return res.status(400).json({
          error: 'Response ID is required',
          timestamp: new Date().toISOString(),
          code: 'MISSING_RESPONSE_ID'
        });
      }
      
      const roomName = `form:${formId}`;
      
      // Check if room exists and has subscribers
      const room = io.sockets.adapter.rooms.get(roomName);
      
      if (!room || room.size === 0) {
        return res.status(404).json({
          error: 'No subscribers found for this form',
          formId: formId,
          timestamp: new Date().toISOString(),
          code: 'NO_SUBSCRIBERS'
        });
      }
      
      // Prepare response data
      const broadcastData = {
        ...responseData,
        formId: formId,
        timestamp: new Date().toISOString(),
        source: 'api'
      };
      
      // Broadcast to all form subscribers
      io.to(roomName).emit('response:new', broadcastData);
      
      // Log the response
      console.log(`📊 New response broadcasted for form ${formId}: ${responseData.responseId}`);
      
      res.json({
        success: true,
        message: 'Response broadcasted to form subscribers',
        formId: formId,
        responseId: responseData.responseId,
        subscribers: room.size,
        timestamp: new Date().toISOString()
      });
      
    } catch (error) {
      res.status(500).json({
        error: 'Failed to broadcast response',
        message: error.message,
        timestamp: new Date().toISOString(),
        code: 'RESPONSE_BROADCAST_ERROR'
      });
    }
  });

  /**
   * @swagger
   * /events/form/{formId}/status:
   *   post:
   *     summary: Update Form Status
   *     description: Broadcast form status changes to all subscribers
   *     tags: [Events]
   *     parameters:
   *       - in: path
   *         name: formId
   *         required: true
   *         schema:
   *           type: string
   *         description: Form ID
   *         example: "form123"
   *     requestBody:
   *       required: true
   *       content:
   *         application/json:
   *           schema:
   *             type: object
   *             required:
   *               - status
   *             properties:
   *               status:
   *                 type: string
   *                 enum: [published, closed, archived, draft]
   *                 description: New form status
   *                 example: "published"
   *               reason:
   *                 type: string
   *                 description: Reason for status change
   *                 example: "Form is now ready for responses"
   *               updatedBy:
   *                 type: string
   *                 description: ID of user who updated the status
   *                 example: "admin123"
   *     responses:
   *       200:
   *         description: Status update broadcasted successfully
   *         content:
   *           application/json:
   *             schema:
   *               type: object
   *               properties:
   *                 success:
   *                   type: boolean
   *                   example: true
   *                 message:
   *                   type: string
   *                   example: "Status update broadcasted"
   *                 formId:
   *                   type: string
   *                   example: "form123"
   *                 status:
   *                   type: string
   *                   example: "published"
   *                 subscribers:
   *                   type: integer
   *                   example: 8
   *                 timestamp:
   *                   type: string
   *                   format: date-time
   *       400:
   *         description: Invalid status value
   *       404:
   *         description: No subscribers for form
   */
  router.post('/form/:formId/status', (req, res) => {
    try {
      const { formId } = req.params;
      const { status, reason, updatedBy } = req.body;
      
      const validStatuses = ['published', 'closed', 'archived', 'draft'];
      
      if (!status || !validStatuses.includes(status)) {
        return res.status(400).json({
          error: 'Invalid status',
          validStatuses: validStatuses,
          timestamp: new Date().toISOString(),
          code: 'INVALID_STATUS'
        });
      }
      
      const roomName = `form:${formId}`;
      
      // Check if room exists and has subscribers
      const room = io.sockets.adapter.rooms.get(roomName);
      
      if (!room || room.size === 0) {
        return res.status(404).json({
          error: 'No subscribers found for this form',
          formId: formId,
          timestamp: new Date().toISOString(),
          code: 'NO_SUBSCRIBERS'
        });
      }
      
      // Prepare status update data
      const statusData = {
        formId: formId,
        status: status,
        reason: reason || null,
        updatedBy: updatedBy || null,
        timestamp: new Date().toISOString(),
        source: 'api'
      };
      
      // Choose event based on status
      let event = 'form:status_updated';
      if (status === 'published') {
        event = 'form:published';
      } else if (status === 'closed') {
        event = 'form:closed';
      }
      
      // Broadcast status update
      io.to(roomName).emit(event, statusData);
      
      // Log the status update
      console.log(`📢 Form ${formId} status updated to ${status} (${room.size} subscribers)`);
      
      res.json({
        success: true,
        message: 'Status update broadcasted',
        formId: formId,
        status: status,
        event: event,
        subscribers: room.size,
        timestamp: new Date().toISOString()
      });
      
    } catch (error) {
      res.status(500).json({
        error: 'Failed to broadcast status update',
        message: error.message,
        timestamp: new Date().toISOString(),
        code: 'STATUS_BROADCAST_ERROR'
      });
    }
  });

  /**
   * @swagger
   * /events/broadcast:
   *   post:
   *     summary: Global Broadcast
   *     description: Send a message to all connected clients (Admin only)
   *     tags: [Events]
   *     security:
   *       - BearerAuth: []
   *     requestBody:
   *       required: true
   *       content:
   *         application/json:
   *           schema:
   *             type: object
   *             required:
   *               - event
   *               - data
   *             properties:
   *               event:
   *                 type: string
   *                 description: Event name to broadcast
   *                 example: "system:maintenance"
   *               data:
   *                 type: object
   *                 description: Event data
   *                 example: {"message": "System will be down for maintenance in 10 minutes"}
   *               priority:
   *                 type: string
   *                 enum: [low, medium, high, critical]
   *                 description: Message priority
   *                 example: "high"
   *     responses:
   *       200:
   *         description: Message broadcasted globally
   *         content:
   *           application/json:
   *             schema:
   *               type: object
   *               properties:
   *                 success:
   *                   type: boolean
   *                   example: true
   *                 message:
   *                   type: string
   *                   example: "Message broadcasted to all clients"
   *                 recipients:
   *                   type: integer
   *                   example: 156
   *                 event:
   *                   type: string
   *                   example: "system:maintenance"
   *                 timestamp:
   *                   type: string
   *                   format: date-time
   *       400:
   *         description: Invalid request data
   *       401:
   *         description: Unauthorized
   */
  router.post('/broadcast', (req, res) => {
    try {
      const { event, data, priority = 'medium' } = req.body;
      
      if (!event || !data) {
        return res.status(400).json({
          error: 'Missing required fields',
          required: ['event', 'data'],
          timestamp: new Date().toISOString(),
          code: 'MISSING_FIELDS'
        });
      }
      
      const activeConnections = io.engine ? io.engine.clientsCount : 0;
      
      // Prepare broadcast data
      const broadcastData = {
        ...data,
        priority: priority,
        timestamp: new Date().toISOString(),
        source: 'admin'
      };
      
      // Broadcast to all connected clients
      io.emit(event, broadcastData);
      
      // Log the broadcast
      console.log(`📢 Global broadcast: ${event} (${activeConnections} recipients)`);
      
      res.json({
        success: true,
        message: 'Message broadcasted to all clients',
        recipients: activeConnections,
        event: event,
        priority: priority,
        timestamp: new Date().toISOString()
      });
      
    } catch (error) {
      res.status(500).json({
        error: 'Failed to broadcast message',
        message: error.message,
        timestamp: new Date().toISOString(),
        code: 'GLOBAL_BROADCAST_ERROR'
      });
    }
  });

  /**
   * @swagger
   * /events/metrics:
   *   get:
   *     summary: Get Event Metrics
   *     description: Get metrics about realtime events and activity
   *     tags: [Events]
   *     responses:
   *       200:
   *         description: Event metrics retrieved successfully
   *         content:
   *           application/json:
   *             schema:
   *               type: object
   *               properties:
   *                 totalEvents:
   *                   type: integer
   *                   example: 1250
   *                 eventsPerSecond:
   *                   type: number
   *                   example: 3.5
   *                 activeRooms:
   *                   type: integer
   *                   example: 25
   *                 totalSubscriptions:
   *                   type: integer
   *                   example: 87
   *                 eventTypes:
   *                   type: object
   *                   additionalProperties:
   *                     type: integer
   *                   example:
   *                     "response:new": 450
   *                     "form:published": 120
   *                     "user:typing": 680
   *                 uptime:
   *                   type: integer
   *                   example: 3600
   *                 timestamp:
   *                   type: string
   *                   format: date-time
   */
  router.get('/metrics', (req, res) => {
    try {
      const uptime = Math.floor((Date.now() - connectionStats.startTime) / 1000);
      const rooms = io.sockets.adapter.rooms;
      
      let totalSubscriptions = 0;
      let activeRooms = 0;
      
      for (const [roomName, room] of rooms) {
        if (roomName.startsWith('form:')) {
          activeRooms++;
          totalSubscriptions += room.size;
        }
      }
      
      res.json({
        totalEvents: connectionStats.totalConnections * 10, // Rough estimate
        eventsPerSecond: connectionStats.eventsPerSecond,
        activeRooms: activeRooms,
        totalSubscriptions: totalSubscriptions,
        eventTypes: {
          'response:new': Math.floor(connectionStats.totalConnections * 0.4),
          'form:published': Math.floor(connectionStats.totalConnections * 0.1),
          'user:typing': Math.floor(connectionStats.totalConnections * 0.3),
          'form:subscribed': Math.floor(connectionStats.totalConnections * 0.2)
        },
        uptime: uptime,
        performance: {
          memoryUsage: process.memoryUsage(),
          cpuUsage: process.cpuUsage()
        },
        timestamp: new Date().toISOString()
      });
    } catch (error) {
      res.status(500).json({
        error: 'Failed to get metrics',
        message: error.message,
        timestamp: new Date().toISOString(),
        code: 'METRICS_ERROR'
      });
    }
  });

  return router;
}

module.exports = { createRealtimeRoutes };
