const express = require('express');

/**
 * Create WebSocket management routes
 * @param {Object} io - Socket.IO instance
 * @param {Object} connectionStats - Connection statistics object
 * @returns {Object} Express router
 */
function createWebSocketRoutes(io, connectionStats) {
  const router = express.Router();

  /**
   * @swagger
   * /ws/info:
   *   get:
   *     summary: WebSocket Connection Info
   *     description: Get information about WebSocket endpoint and connection details
   *     tags: [WebSocket]
   *     responses:
   *       200:
   *         description: WebSocket information
   *         content:
   *           application/json:
   *             schema:
   *               $ref: '#/components/schemas/WebSocketInfo'
   */
  router.get('/info', (req, res) => {
    const port = process.env.REALTIME_SERVICE_PORT || 8002;
    
    res.json({
      endpoint: `ws://localhost:${port}`,
      protocol: 'socket.io',
      version: '4.7.4',
      transports: ['websocket', 'polling'],
      cors: {
        enabled: true,
        origins: process.env.ALLOWED_ORIGINS?.split(',') || ["http://localhost:3000"]
      },
      authentication: {
        required: process.env.REQUIRE_AUTH === 'true',
        method: 'JWT',
        header: 'auth.token'
      },
      events: {
        supported: [
          'form:subscribe',
          'form:unsubscribe',
          'response:new',
          'response:update',
          'form:published',
          'form:closed',
          'user:typing',
          'user:stopped_typing',
          'ping'
        ],
        responses: [
          'form:subscribed',
          'form:unsubscribed',
          'response:update',
          'response:updated',
          'form:published',
          'form:closed',
          'user:typing',
          'user:stopped_typing',
          'pong',
          'error'
        ]
      },
      connection: {
        timeout: 20000,
        pingInterval: 25000,
        pingTimeout: 5000
      },
      timestamp: new Date().toISOString()
    });
  });

  /**
   * @swagger
   * /ws/connections:
   *   get:
   *     summary: Active WebSocket Connections
   *     description: Get list of active WebSocket connections and their details
   *     tags: [WebSocket]
   *     responses:
   *       200:
   *         description: Active connections list
   *         content:
   *           application/json:
   *             schema:
   *               type: object
   *               properties:
   *                 totalConnections:
   *                   type: integer
   *                   example: 150
   *                 activeConnections:
   *                   type: integer
   *                   example: 42
   *                 connections:
   *                   type: array
   *                   items:
   *                     $ref: '#/components/schemas/SocketConnection'
   *                 timestamp:
   *                   type: string
   *                   format: date-time
   */
  router.get('/connections', (req, res) => {
    try {
      const sockets = io.sockets.sockets;
      const connections = [];
      
      for (const [socketId, socket] of sockets) {
        const rooms = Array.from(socket.rooms).filter(room => room !== socketId);
        
        connections.push({
          socketId,
          userId: socket.userId || null,
          connected: socket.connected,
          rooms: rooms,
          handshake: {
            time: socket.handshake.time,
            address: socket.handshake.address,
            userAgent: socket.handshake.headers['user-agent']
          }
        });
      }
      
      res.json({
        totalConnections: connectionStats.totalConnections,
        activeConnections: connections.length,
        connections: connections,
        timestamp: new Date().toISOString()
      });
    } catch (error) {
      res.status(500).json({
        error: 'Failed to get connections',
        message: error.message,
        timestamp: new Date().toISOString(),
        code: 'CONNECTIONS_ERROR'
      });
    }
  });

  /**
   * @swagger
   * /ws/rooms:
   *   get:
   *     summary: Active WebSocket Rooms
   *     description: Get information about active WebSocket rooms and their subscribers
   *     tags: [WebSocket]
   *     responses:
   *       200:
   *         description: Active rooms information
   *         content:
   *           application/json:
   *             schema:
   *               type: object
   *               properties:
   *                 totalRooms:
   *                   type: integer
   *                   example: 15
   *                 rooms:
   *                   type: object
   *                   additionalProperties:
   *                     type: object
   *                     properties:
   *                       name:
   *                         type: string
   *                         example: "form:123"
   *                       subscribers:
   *                         type: integer
   *                         example: 5
   *                       sockets:
   *                         type: array
   *                         items:
   *                           type: string
   *                         example: ["socket1", "socket2"]
   *                 timestamp:
   *                   type: string
   *                   format: date-time
   */
  router.get('/rooms', (req, res) => {
    try {
      const rooms = io.sockets.adapter.rooms;
      const roomInfo = {};
      let totalFormRooms = 0;
      
      for (const [roomName, room] of rooms) {
        if (roomName.startsWith('form:')) {
          totalFormRooms++;
          roomInfo[roomName] = {
            name: roomName,
            subscribers: room.size,
            sockets: Array.from(room)
          };
        }
      }
      
      res.json({
        totalRooms: totalFormRooms,
        rooms: roomInfo,
        timestamp: new Date().toISOString()
      });
    } catch (error) {
      res.status(500).json({
        error: 'Failed to get rooms',
        message: error.message,
        timestamp: new Date().toISOString(),
        code: 'ROOMS_ERROR'
      });
    }
  });

  /**
   * @swagger
   * /ws/broadcast:
   *   post:
   *     summary: Broadcast Message to Room
   *     description: Send a message to all clients in a specific room (Admin endpoint)
   *     tags: [WebSocket]
   *     security:
   *       - BearerAuth: []
   *     requestBody:
   *       required: true
   *       content:
   *         application/json:
   *           schema:
   *             type: object
   *             required:
   *               - room
   *               - event
   *               - data
   *             properties:
   *               room:
   *                 type: string
   *                 description: Room name to broadcast to
   *                 example: "form:12345"
   *               event:
   *                 type: string
   *                 description: Event name to emit
   *                 example: "admin:announcement"
   *               data:
   *                 type: object
   *                 description: Data to send with the event
   *                 example: {"message": "System maintenance in 10 minutes"}
   *     responses:
   *       200:
   *         description: Message broadcasted successfully
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
   *                   example: "Message broadcasted to room form:12345"
   *                 recipients:
   *                   type: integer
   *                   example: 5
   *                 timestamp:
   *                   type: string
   *                   format: date-time
   *       400:
   *         description: Invalid request
   *       401:
   *         description: Unauthorized
   *       404:
   *         description: Room not found
   */
  router.post('/broadcast', (req, res) => {
    try {
      const { room, event, data } = req.body;
      
      if (!room || !event || !data) {
        return res.status(400).json({
          error: 'Missing required fields',
          required: ['room', 'event', 'data'],
          timestamp: new Date().toISOString(),
          code: 'MISSING_FIELDS'
        });
      }
      
      // Check if room exists
      const roomExists = io.sockets.adapter.rooms.has(room);
      
      if (!roomExists) {
        return res.status(404).json({
          error: 'Room not found',
          room: room,
          timestamp: new Date().toISOString(),
          code: 'ROOM_NOT_FOUND'
        });
      }
      
      // Get room size
      const roomSize = io.sockets.adapter.rooms.get(room)?.size || 0;
      
      // Broadcast message
      const broadcastData = {
        ...data,
        broadcastBy: 'admin',
        timestamp: new Date().toISOString()
      };
      
      io.to(room).emit(event, broadcastData);
      
      res.json({
        success: true,
        message: `Message broadcasted to room ${room}`,
        recipients: roomSize,
        event: event,
        timestamp: new Date().toISOString()
      });
      
    } catch (error) {
      res.status(500).json({
        error: 'Failed to broadcast message',
        message: error.message,
        timestamp: new Date().toISOString(),
        code: 'BROADCAST_ERROR'
      });
    }
  });

  /**
   * @swagger
   * /ws/disconnect/{socketId}:
   *   post:
   *     summary: Disconnect Socket
   *     description: Force disconnect a specific socket connection (Admin endpoint)
   *     tags: [WebSocket]
   *     security:
   *       - BearerAuth: []
   *     parameters:
   *       - in: path
   *         name: socketId
   *         required: true
   *         schema:
   *           type: string
   *         description: Socket ID to disconnect
   *     responses:
   *       200:
   *         description: Socket disconnected successfully
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
   *                   example: "Socket disconnected successfully"
   *                 socketId:
   *                   type: string
   *                   example: "abcd1234"
   *                 timestamp:
   *                   type: string
   *                   format: date-time
   *       404:
   *         description: Socket not found
   *       500:
   *         description: Internal server error
   */
  router.post('/disconnect/:socketId', (req, res) => {
    try {
      const { socketId } = req.params;
      
      const socket = io.sockets.sockets.get(socketId);
      
      if (!socket) {
        return res.status(404).json({
          error: 'Socket not found',
          socketId: socketId,
          timestamp: new Date().toISOString(),
          code: 'SOCKET_NOT_FOUND'
        });
      }
      
      // Disconnect the socket
      socket.disconnect(true);
      
      res.json({
        success: true,
        message: 'Socket disconnected successfully',
        socketId: socketId,
        timestamp: new Date().toISOString()
      });
      
    } catch (error) {
      res.status(500).json({
        error: 'Failed to disconnect socket',
        message: error.message,
        timestamp: new Date().toISOString(),
        code: 'DISCONNECT_ERROR'
      });
    }
  });

  return router;
}

module.exports = { createWebSocketRoutes };
