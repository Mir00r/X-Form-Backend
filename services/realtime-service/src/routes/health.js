const express = require('express');

/**
 * Create health check routes
 * @param {Object} io - Socket.IO instance
 * @param {Object} connectionStats - Connection statistics object
 * @returns {Object} Express router
 */
function createHealthRoutes(io, connectionStats) {
  const router = express.Router();

  /**
   * @swagger
   * /health:
   *   get:
   *     summary: Basic Health Check
   *     description: Check if the Realtime Service is running and responsive
   *     tags: [Health]
   *     responses:
   *       200:
   *         description: Service is healthy
   *         content:
   *           application/json:
   *             schema:
   *               $ref: '#/components/schemas/HealthCheck'
   *       503:
   *         description: Service is unhealthy
   *         content:
   *           application/json:
   *             schema:
   *               $ref: '#/components/schemas/ErrorResponse'
   */
  router.get('/', (req, res) => {
    try {
      const uptime = process.uptime();
      const memoryUsage = process.memoryUsage();
      
      // Update connection stats
      const activeConnections = io.engine ? io.engine.clientsCount : 0;
      
      res.json({
        status: 'healthy',
        timestamp: new Date().toISOString(),
        uptime: Math.floor(uptime),
        service: 'realtime-service',
        version: process.env.npm_package_version || '1.0.0',
        connections: {
          active: activeConnections,
          total: connectionStats.totalConnections
        },
        memory: {
          used: Math.round(memoryUsage.heapUsed / 1024 / 1024),
          total: Math.round(memoryUsage.heapTotal / 1024 / 1024),
          external: Math.round(memoryUsage.external / 1024 / 1024)
        },
        environment: process.env.NODE_ENV || 'development'
      });
    } catch (error) {
      res.status(503).json({
        status: 'unhealthy',
        error: error.message,
        timestamp: new Date().toISOString(),
        code: 'HEALTH_CHECK_FAILED'
      });
    }
  });

  /**
   * @swagger
   * /health/detailed:
   *   get:
   *     summary: Detailed Health Check
   *     description: Get comprehensive health information including dependencies and metrics
   *     tags: [Health]
   *     responses:
   *       200:
   *         description: Detailed health information
   *         content:
   *           application/json:
   *             schema:
   *               $ref: '#/components/schemas/DetailedHealthCheck'
   *       503:
   *         description: Service is unhealthy
   *         content:
   *           application/json:
   *             schema:
   *               $ref: '#/components/schemas/ErrorResponse'
   */
  router.get('/detailed', (req, res) => {
    try {
      const uptime = process.uptime();
      const memoryUsage = process.memoryUsage();
      
      // Update connection stats
      const activeConnections = io.engine ? io.engine.clientsCount : 0;
      
      // Get Socket.IO adapter information
      const rooms = io.sockets.adapter.rooms;
      const roomInfo = {};
      let totalRoomConnections = 0;
      
      for (const [roomName, room] of rooms) {
        if (roomName.startsWith('form:')) {
          roomInfo[roomName] = room.size;
          totalRoomConnections += room.size;
        }
      }

      res.json({
        status: 'healthy',
        timestamp: new Date().toISOString(),
        uptime: Math.floor(uptime),
        service: {
          name: 'realtime-service',
          version: process.env.npm_package_version || '1.0.0',
          environment: process.env.NODE_ENV || 'development',
          nodeVersion: process.version,
          platform: process.platform,
          pid: process.pid
        },
        connections: {
          active: activeConnections,
          total: connectionStats.totalConnections,
          roomSubscriptions: totalRoomConnections,
          rooms: Object.keys(roomInfo).length
        },
        rooms: roomInfo,
        memory: {
          used: Math.round(memoryUsage.heapUsed / 1024 / 1024),
          total: Math.round(memoryUsage.heapTotal / 1024 / 1024),
          external: Math.round(memoryUsage.external / 1024 / 1024),
          rss: Math.round(memoryUsage.rss / 1024 / 1024)
        },
        performance: {
          eventsPerSecond: connectionStats.eventsPerSecond,
          uptimeSeconds: uptime,
          cpuUsage: process.cpuUsage()
        },
        socketio: {
          version: '4.7.4',
          engine: io.engine ? 'running' : 'not initialized',
          transports: ['websocket', 'polling']
        }
      });
    } catch (error) {
      res.status(503).json({
        status: 'unhealthy',
        error: error.message,
        timestamp: new Date().toISOString(),
        code: 'DETAILED_HEALTH_CHECK_FAILED'
      });
    }
  });

  /**
   * @swagger
   * /health/live:
   *   get:
   *     summary: Liveness Probe
   *     description: Kubernetes liveness probe endpoint - checks if the application is alive
   *     tags: [Health]
   *     responses:
   *       200:
   *         description: Service is alive
   *         content:
   *           application/json:
   *             schema:
   *               type: object
   *               properties:
   *                 status:
   *                   type: string
   *                   example: "alive"
   *                 timestamp:
   *                   type: string
   *                   format: date-time
   *       503:
   *         description: Service is not alive
   */
  router.get('/live', (req, res) => {
    res.json({
      status: 'alive',
      timestamp: new Date().toISOString()
    });
  });

  /**
   * @swagger
   * /health/ready:
   *   get:
   *     summary: Readiness Probe
   *     description: Kubernetes readiness probe endpoint - checks if the service is ready to accept traffic
   *     tags: [Health]
   *     responses:
   *       200:
   *         description: Service is ready
   *         content:
   *           application/json:
   *             schema:
   *               type: object
   *               properties:
   *                 status:
   *                   type: string
   *                   example: "ready"
   *                 timestamp:
   *                   type: string
   *                   format: date-time
   *                 checks:
   *                   type: object
   *                   properties:
   *                     socketio:
   *                       type: string
   *                       example: "ready"
   *       503:
   *         description: Service is not ready
   */
  router.get('/ready', (req, res) => {
    try {
      // Check if Socket.IO is ready
      const socketioReady = io && io.engine ? 'ready' : 'not ready';
      
      if (socketioReady === 'ready') {
        res.json({
          status: 'ready',
          timestamp: new Date().toISOString(),
          checks: {
            socketio: socketioReady
          }
        });
      } else {
        res.status(503).json({
          status: 'not ready',
          timestamp: new Date().toISOString(),
          checks: {
            socketio: socketioReady
          }
        });
      }
    } catch (error) {
      res.status(503).json({
        status: 'not ready',
        error: error.message,
        timestamp: new Date().toISOString(),
        code: 'READINESS_CHECK_FAILED'
      });
    }
  });

  return router;
}

module.exports = { createHealthRoutes };
