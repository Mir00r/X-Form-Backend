const express = require('express');
const http = require('http');
const socketIo = require('socket.io');
const cors = require('cors');
const helmet = require('helmet');
require('dotenv').config();

const app = express();
const server = http.createServer(app);
const io = socketIo(server, {
  cors: {
    origin: process.env.ALLOWED_ORIGINS?.split(',') || ["http://localhost:3000"],
    methods: ["GET", "POST"]
  }
});

// Middleware
app.use(helmet());
app.use(cors());
app.use(express.json());

// Health check endpoint
app.get('/health', (req, res) => {
  res.status(200).json({ 
    status: 'healthy',
    service: 'realtime-service',
    timestamp: new Date().toISOString()
  });
});

// Socket.IO connection handling
io.on('connection', (socket) => {
  console.log('New client connected:', socket.id);

  // Handle form updates
  socket.on('form:subscribe', (formId) => {
    socket.join(`form:${formId}`);
    console.log(`Client ${socket.id} subscribed to form ${formId}`);
  });

  socket.on('form:unsubscribe', (formId) => {
    socket.leave(`form:${formId}`);
    console.log(`Client ${socket.id} unsubscribed from form ${formId}`);
  });

  // Handle response updates
  socket.on('response:new', (data) => {
    io.to(`form:${data.formId}`).emit('response:update', data);
  });

  socket.on('disconnect', () => {
    console.log('Client disconnected:', socket.id);
  });
});

const PORT = process.env.REALTIME_SERVICE_PORT || 8002;

server.listen(PORT, () => {
  console.log(`Realtime Service running on port ${PORT}`);
});

module.exports = { app, server };
