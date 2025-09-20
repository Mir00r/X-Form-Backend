const express = require('express');
const http = require('http');
const socketIo = require('socket.io');
const cors = require('cors');
const helmet = require('helmet');
const jwt = require('jsonwebtoken');
require('dotenv').config();

const { swaggerSpec, serve, setup } = require('./swagger');
const { createRealtimeRoutes } = require('./routes/realtime');
const { createHealthRoutes } = require('./routes/health');
const { createWebSocketRoutes } = require('./routes/websocket');

const app = express();
const server = http.createServer(app);

// Socket.IO server with CORS configuration
const io = socketIo(server, {
  cors: {
    origin: process.env.ALLOWED_ORIGINS?.split(',') || ["http://localhost:3000"],
    methods: ["GET", "POST"],
    credentials: true
  },
  transports: ['websocket', 'polling']
});

// Global middleware
app.use(helmet({
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      styleSrc: ["'self'", "'unsafe-inline'"],
      scriptSrc: ["'self'"],
      imgSrc: ["'self'", "data:", "https:"],
    },
  },
}));

app.use(cors({
  origin: process.env.ALLOWED_ORIGINS?.split(',') || ["http://localhost:3000"],
  credentials: true
}));

app.use(express.json({ limit: '10mb' }));
app.use(express.urlencoded({ extended: true, limit: '10mb' }));

// Request logging middleware
app.use((req, res, next) => {
  console.log(`${new Date().toISOString()} - ${req.method} ${req.path}`);
  next();
});

// Global variables for tracking
let connectionStats = {
  totalConnections: 0,
  activeConnections: 0,
  roomSubscriptions: {},
  eventsPerSecond: 0,
  startTime: Date.now()
};

// Update connection stats
const updateConnectionStats = () => {
  connectionStats.activeConnections = io.engine.clientsCount || 0;
  
  // Calculate rooms and subscriptions
  const rooms = io.sockets.adapter.rooms;
  connectionStats.roomSubscriptions = {};
  
  for (const [roomName, room] of rooms) {
    if (roomName.startsWith('form:')) {
      connectionStats.roomSubscriptions[roomName] = room.size;
    }
  }
};

// Swagger documentation
app.use('/api-docs', serve, setup);

/**
 * @swagger
 * /:
 *   get:
 *     summary: Service Information
 *     description: Get comprehensive information about the Realtime Service
 *     tags: [Health]
 *     responses:
 *       200:
 *         description: Service information retrieved successfully
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 service:
 *                   type: string
 *                   example: "Realtime Service API"
 *                 version:
 *                   type: string
 *                   example: "1.0.0"
 *                 description:
 *                   type: string
 *                   example: "Real-time WebSocket service for X-Form Backend"
 *                 endpoints:
 *                   type: object
 *                   properties:
 *                     health:
 *                       type: string
 *                       example: "/health"
 *                     swagger:
 *                       type: string
 *                       example: "/api-docs"
 *                     websocket:
 *                       type: string
 *                       example: "/ws"
 *                     stats:
 *                       type: string
 *                       example: "/stats"
 *                 features:
 *                   type: array
 *                   items:
 *                     type: string
 *                   example: ["Real-time messaging", "WebSocket connections", "Event broadcasting"]
 *                 timestamp:
 *                   type: string
 *                   format: date-time
 *                   example: "2024-01-15T10:30:00.000Z"
 */
app.get('/', (req, res) => {
  updateConnectionStats();
  
  res.json({
    service: 'Realtime Service API',
    version: process.env.npm_package_version || '1.0.0',
    description: 'Real-time WebSocket service for X-Form Backend providing live communication capabilities',
    endpoints: {
      health: '/health',
      swagger: '/api-docs',
      websocket: '/ws',
      stats: '/stats',
      events: '/events'
    },
    features: [
      '✅ Real-time WebSocket messaging',
      '✅ Form subscription management',
      '✅ Event broadcasting',
      '✅ Connection monitoring',
      '✅ Swagger API documentation',
      '✅ Health checks & metrics',
      '✅ CORS & security headers',
      '✅ JWT authentication support',
      '✅ Room-based communication',
      '✅ Horizontal scaling ready'
    ],
    websocket: {
      endpoint: `ws://localhost:${process.env.REALTIME_SERVICE_PORT || 8002}`,
      protocol: 'socket.io',
      version: '4.7.4',
      transports: ['websocket', 'polling']
    },
    connections: {
      active: connectionStats.activeConnections,
      total: connectionStats.totalConnections,
      rooms: Object.keys(connectionStats.roomSubscriptions).length
    },
    timestamp: new Date().toISOString()
  });
});

// Mount route modules
app.use('/health', createHealthRoutes(io, connectionStats));
app.use('/ws', createWebSocketRoutes(io, connectionStats));
app.use('/events', createRealtimeRoutes(io, connectionStats));

/**
 * @swagger
 * /stats:
 *   get:
 *     summary: Get Connection Statistics
 *     description: Get real-time statistics about WebSocket connections and activities
 *     tags: [Monitoring]
 *     responses:
 *       200:
 *         description: Connection statistics retrieved successfully
 *         content:
 *           application/json:
 *             schema:
 *               $ref: '#/components/schemas/ConnectionStats'
 *       500:
 *         description: Internal server error
 *         content:
 *           application/json:
 *             schema:
 *               $ref: '#/components/schemas/ErrorResponse'
 */
app.get('/stats', (req, res) => {
  try {
    updateConnectionStats();
    
    const uptime = Math.floor((Date.now() - connectionStats.startTime) / 1000);
    
    res.json({
      totalConnections: connectionStats.totalConnections,
      activeConnections: connectionStats.activeConnections,
      roomSubscriptions: connectionStats.roomSubscriptions,
      eventsPerSecond: connectionStats.eventsPerSecond,
      uptime: uptime,
      server: {
        memory: process.memoryUsage(),
        nodeVersion: process.version,
        platform: process.platform,
        pid: process.pid
      },
      timestamp: new Date().toISOString()
    });
  } catch (error) {
    res.status(500).json({
      error: 'Failed to retrieve statistics',
      message: error.message,
      timestamp: new Date().toISOString(),
      code: 'STATS_ERROR'
    });
  }
});

// JWT authentication middleware (optional)
const authenticateSocket = (socket, next) => {
  const token = socket.handshake.auth.token;
  
  if (!token && process.env.REQUIRE_AUTH === 'true') {
    return next(new Error('Authentication required'));
  }
  
  if (token) {
    try {
      const decoded = jwt.verify(token, process.env.JWT_SECRET || 'fallback-secret');
      socket.userId = decoded.userId;
      socket.user = decoded;
    } catch (err) {
      if (process.env.REQUIRE_AUTH === 'true') {
        return next(new Error('Invalid token'));
      }
      console.warn('Invalid token provided, but authentication not required');
    }
  }
  
  next();
};

// Socket.IO middleware
io.use(authenticateSocket);

// Socket.IO connection handling with comprehensive event management
io.on('connection', (socket) => {
  console.log(`✅ New client connected: ${socket.id} ${socket.userId ? `(User: ${socket.userId})` : '(Anonymous)'}`);
  
  connectionStats.totalConnections++;
  updateConnectionStats();

  /**
   * Form subscription events
   */
  
  // Subscribe to form updates
  socket.on('form:subscribe', (formId) => {
    if (!formId) {
      socket.emit('error', { message: 'Form ID is required' });
      return;
    }
    
    const roomName = `form:${formId}`;
    socket.join(roomName);
    
    console.log(`📝 Client ${socket.id} subscribed to form ${formId}`);
    
    // Notify about successful subscription
    socket.emit('form:subscribed', {
      formId,
      roomName,
      timestamp: new Date().toISOString()
    });
    
    // Notify others in the room about new subscriber
    socket.to(roomName).emit('form:new_subscriber', {
      socketId: socket.id,
      userId: socket.userId,
      formId,
      timestamp: new Date().toISOString()
    });
    
    updateConnectionStats();
  });

  // Unsubscribe from form updates
  socket.on('form:unsubscribe', (formId) => {
    if (!formId) {
      socket.emit('error', { message: 'Form ID is required' });
      return;
    }
    
    const roomName = `form:${formId}`;
    socket.leave(roomName);
    
    console.log(`📝 Client ${socket.id} unsubscribed from form ${formId}`);
    
    // Notify about successful unsubscription
    socket.emit('form:unsubscribed', {
      formId,
      roomName,
      timestamp: new Date().toISOString()
    });
    
    // Notify others in the room about subscriber leaving
    socket.to(roomName).emit('form:subscriber_left', {
      socketId: socket.id,
      userId: socket.userId,
      formId,
      timestamp: new Date().toISOString()
    });
    
    updateConnectionStats();
  });

  /**
   * Response events
   */
  
  // Handle new response submission
  socket.on('response:new', (data) => {
    if (!data.formId) {
      socket.emit('error', { message: 'Form ID is required in response data' });
      return;
    }
    
    const responseData = {
      ...data,
      socketId: socket.id,
      userId: socket.userId,
      timestamp: new Date().toISOString()
    };
    
    // Broadcast to all subscribers of this form
    io.to(`form:${data.formId}`).emit('response:update', responseData);
    
    console.log(`📊 New response for form ${data.formId} from ${socket.id}`);
    
    // Track events per second
    connectionStats.eventsPerSecond = (connectionStats.eventsPerSecond + 1) / 2;
  });

  // Handle response updates
  socket.on('response:update', (data) => {
    if (!data.formId || !data.responseId) {
      socket.emit('error', { message: 'Form ID and Response ID are required' });
      return;
    }
    
    const updateData = {
      ...data,
      socketId: socket.id,
      userId: socket.userId,
      timestamp: new Date().toISOString()
    };
    
    // Broadcast update to form subscribers
    io.to(`form:${data.formId}`).emit('response:updated', updateData);
    
    console.log(`📊 Response ${data.responseId} updated for form ${data.formId}`);
  });

  /**
   * Form lifecycle events
   */
  
  // Form published event
  socket.on('form:published', (data) => {
    if (!data.formId) {
      socket.emit('error', { message: 'Form ID is required' });
      return;
    }
    
    const publishData = {
      ...data,
      publishedBy: socket.userId,
      timestamp: new Date().toISOString()
    };
    
    // Broadcast to form subscribers
    io.to(`form:${data.formId}`).emit('form:published', publishData);
    
    console.log(`📢 Form ${data.formId} published`);
  });

  // Form closed event
  socket.on('form:closed', (data) => {
    if (!data.formId) {
      socket.emit('error', { message: 'Form ID is required' });
      return;
    }
    
    const closeData = {
      ...data,
      closedBy: socket.userId,
      timestamp: new Date().toISOString()
    };
    
    // Broadcast to form subscribers
    io.to(`form:${data.formId}`).emit('form:closed', closeData);
    
    console.log(`🔒 Form ${data.formId} closed`);
  });

  /**
   * Real-time collaboration events
   */
  
  // User typing indicator
  socket.on('user:typing', (data) => {
    if (!data.formId) return;
    
    socket.to(`form:${data.formId}`).emit('user:typing', {
      ...data,
      socketId: socket.id,
      userId: socket.userId,
      timestamp: new Date().toISOString()
    });
  });

  // User stopped typing
  socket.on('user:stopped_typing', (data) => {
    if (!data.formId) return;
    
    socket.to(`form:${data.formId}`).emit('user:stopped_typing', {
      ...data,
      socketId: socket.id,
      userId: socket.userId,
      timestamp: new Date().toISOString()
    });
  });

  /**
   * Connection management
   */
  
  // Handle ping for connection health
  socket.on('ping', () => {
    socket.emit('pong', {
      timestamp: new Date().toISOString(),
      socketId: socket.id
    });
  });

  // Get current room subscriptions
  socket.on('rooms:list', () => {
    const rooms = Array.from(socket.rooms).filter(room => room !== socket.id);
    socket.emit('rooms:list', {
      rooms,
      socketId: socket.id,
      timestamp: new Date().toISOString()
    });
  });

  // Handle disconnection
  socket.on('disconnect', (reason) => {
    console.log(`❌ Client disconnected: ${socket.id} (Reason: ${reason})`);
    
    updateConnectionStats();
    
    // Notify all rooms about disconnection
    const rooms = Array.from(socket.rooms);
    rooms.forEach(room => {
      if (room !== socket.id) {
        socket.to(room).emit('user:disconnected', {
          socketId: socket.id,
          userId: socket.userId,
          reason,
          timestamp: new Date().toISOString()
        });
      }
    });
  });

  // Handle connection errors
  socket.on('error', (error) => {
    console.error(`Socket error for ${socket.id}:`, error);
    socket.emit('error', {
      message: 'A connection error occurred',
      timestamp: new Date().toISOString()
    });
  });
});

// Error handling middleware
app.use((err, req, res, next) => {
  console.error('Error:', err);
  res.status(500).json({
    error: 'Internal server error',
    message: process.env.NODE_ENV === 'development' ? err.message : 'Something went wrong',
    timestamp: new Date().toISOString(),
    code: 'INTERNAL_SERVER_ERROR'
  });
});

// 404 handler
app.use('*', (req, res) => {
  res.status(404).json({
    error: 'Not found',
    message: `Route ${req.originalUrl} not found`,
    availableEndpoints: [
      '/',
      '/health',
      '/api-docs',
      '/ws',
      '/stats',
      '/events'
    ],
    timestamp: new Date().toISOString(),
    code: 'ROUTE_NOT_FOUND'
  });
});

const PORT = process.env.REALTIME_SERVICE_PORT || 8002;

// Graceful shutdown
const gracefulShutdown = () => {
  console.log('🛑 Shutting down gracefully...');
  
  // Close all socket connections
  io.close(() => {
    console.log('✅ All socket connections closed');
  });
  
  // Close HTTP server
  server.close(() => {
    console.log('✅ HTTP server closed');
    process.exit(0);
  });
  
  // Force close after 30 seconds
  setTimeout(() => {
    console.log('❌ Force closing server');
    process.exit(1);
  }, 30000);
};

// Handle shutdown signals
process.on('SIGTERM', gracefulShutdown);
process.on('SIGINT', gracefulShutdown);

// Start server
server.listen(PORT, () => {
  console.log('🚀 Realtime Service starting...');
  console.log(`📊 Server running on port ${PORT}`);
  console.log(`📖 Swagger documentation available at: http://localhost:${PORT}/api-docs`);
  console.log(`🔍 Health check available at: http://localhost:${PORT}/health`);
  console.log(`🌐 WebSocket endpoint: ws://localhost:${PORT}`);
  console.log(`📋 Service info available at: http://localhost:${PORT}/`);
  console.log('✅ Ready to accept connections');
});

module.exports = { app, server, io };
