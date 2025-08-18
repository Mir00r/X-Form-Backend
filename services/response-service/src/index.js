const express = require('express');
const cors = require('cors');
const helmet = require('helmet');
require('dotenv').config();

const app = express();

// Middleware
app.use(helmet());
app.use(cors());
app.use(express.json());

// Health check endpoint
app.get('/health', (req, res) => {
  res.status(200).json({ 
    status: 'healthy',
    service: 'response-service',
    timestamp: new Date().toISOString()
  });
});

// API Routes
app.get('/api/responses', (req, res) => {
  res.json({ message: 'Response Service is running' });
});

// Form response endpoints
app.post('/api/responses', (req, res) => {
  // TODO: Implement response submission
  res.status(201).json({ 
    message: 'Response submitted successfully',
    responseId: `resp_${Date.now()}`
  });
});

app.get('/api/responses/:id', (req, res) => {
  // TODO: Implement response retrieval
  res.json({ 
    responseId: req.params.id,
    message: 'Response retrieved successfully'
  });
});

const PORT = process.env.RESPONSE_SERVICE_PORT || 3002;

app.listen(PORT, () => {
  console.log(`Response Service running on port ${PORT}`);
});

module.exports = app;
