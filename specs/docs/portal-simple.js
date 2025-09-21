const express = require('express');
const swaggerUi = require('swagger-ui-express');
const YAML = require('yaml');
const fs = require('fs');
const path = require('path');
const cors = require('cors');

const app = express();
const port = process.env.PORT || 3000;

// Middleware
app.use(cors());
app.use(express.json());

// Load OpenAPI specs
function loadSpec(specPath) {
  try {
    const content = fs.readFileSync(specPath, 'utf8');
    return YAML.parse(content);
  } catch (error) {
    console.error(`Error loading spec ${specPath}:`, error.message);
    return null;
  }
}

// Routes
app.get('/', (req, res) => {
  res.send(`
    <html>
      <head>
        <title>X-Form Backend API Documentation</title>
        <style>
          body { font-family: Arial, sans-serif; margin: 40px; }
          .header { background: #f8f9fa; padding: 20px; border-radius: 8px; margin-bottom: 30px; }
          .service { background: white; border: 1px solid #ddd; padding: 20px; margin: 10px 0; border-radius: 8px; }
          .service h3 { margin-top: 0; color: #0366d6; }
          a { color: #0366d6; text-decoration: none; }
          a:hover { text-decoration: underline; }
          .status { color: #28a745; font-weight: bold; }
        </style>
      </head>
      <body>
        <div class="header">
          <h1>🚀 X-Form Backend API Documentation</h1>
          <p>Centralized API documentation for all microservices</p>
          <p class="status">✅ GitHub Spec Kit Active</p>
        </div>

        <div class="service">
          <h3>📋 Combined API Documentation</h3>
          <p>Complete API documentation for all services</p>
          <a href="/docs">→ View Interactive Documentation</a>
        </div>

        <div class="service">
          <h3>🔐 Auth Service</h3>
          <p>User authentication and management</p>
          <a href="/docs/auth">→ View Auth API</a>
        </div>

        <div class="service">
          <h3>📝 Form Service</h3>
          <p>Form creation and management</p>
          <a href="/docs/form">→ View Form API</a>
        </div>

        <div class="service">
          <h3>📊 Response Service</h3>
          <p>Form response collection and management</p>
          <a href="/docs/response">→ View Response API</a>
        </div>

        <div class="service">
          <h3>⚡ Realtime Service</h3>
          <p>WebSocket communication and live features</p>
          <a href="/docs/realtime">→ View Realtime API</a>
        </div>

        <div class="service">
          <h3>📈 Analytics Service</h3>
          <p>Data analytics and reporting</p>
          <a href="/docs/analytics">→ View Analytics API</a>
        </div>

        <hr style="margin: 30px 0;">
        
        <h3>🔧 Developer Resources</h3>
        <ul>
          <li><a href="/openapi.yaml">Download OpenAPI YAML</a></li>
          <li><a href="/openapi.json">Download OpenAPI JSON</a></li>
          <li><a href="/health">Health Check</a></li>
        </ul>
      </body>
    </html>
  `);
});

// Combined API docs
app.get('/docs', (req, res, next) => {
  const mainSpec = loadSpec(path.join(__dirname, '../openapi.yaml'));
  if (!mainSpec) {
    return res.status(404).json({ error: 'Main specification not found' });
  }
  
  const options = {
    customCss: '.swagger-ui .topbar { display: none }',
    customSiteTitle: 'X-Form Backend API Documentation'
  };
  
  swaggerUi.setup(mainSpec, options)(req, res, next);
});

app.use('/docs', swaggerUi.serve);

// Service-specific docs
app.get('/docs/:service', (req, res, next) => {
  const service = req.params.service;
  const specPath = path.join(__dirname, `../services/${service}-service.yaml`);
  const spec = loadSpec(specPath);
  
  if (!spec) {
    return res.status(404).json({ error: `Specification for ${service} service not found` });
  }
  
  const options = {
    customCss: '.swagger-ui .topbar { display: none }',
    customSiteTitle: `${service.charAt(0).toUpperCase() + service.slice(1)} Service API`
  };
  
  swaggerUi.setup(spec, options)(req, res, next);
});

// API endpoints
app.get('/openapi.yaml', (req, res) => {
  const specPath = path.join(__dirname, '../openapi.yaml');
  if (fs.existsSync(specPath)) {
    res.setHeader('Content-Type', 'application/x-yaml');
    res.sendFile(specPath);
  } else {
    res.status(404).json({ error: 'OpenAPI specification not found' });
  }
});

app.get('/openapi.json', (req, res) => {
  const spec = loadSpec(path.join(__dirname, '../openapi.yaml'));
  if (spec) {
    res.json(spec);
  } else {
    res.status(404).json({ error: 'OpenAPI specification not found' });
  }
});

app.get('/health', (req, res) => {
  res.json({
    status: 'healthy',
    timestamp: new Date().toISOString(),
    version: '1.0.0',
    services: {
      'auth-service': fs.existsSync(path.join(__dirname, '../services/auth-service.yaml')),
      'form-service': fs.existsSync(path.join(__dirname, '../services/form-service.yaml')),
      'response-service': fs.existsSync(path.join(__dirname, '../services/response-service.yaml')),
      'realtime-service': fs.existsSync(path.join(__dirname, '../services/realtime-service.yaml')),
      'analytics-service': fs.existsSync(path.join(__dirname, '../services/analytics-service.yaml'))
    }
  });
});

// Start server
app.listen(port, () => {
  console.log(`🚀 GitHub Spec Kit Documentation Portal running at http://localhost:${port}`);
  console.log(`📚 Main documentation: http://localhost:${port}/docs`);
  console.log(`❤️  Health check: http://localhost:${port}/health`);
  console.log(`📋 OpenAPI spec: http://localhost:${port}/openapi.yaml`);
});

// Error handling
app.use((err, req, res, next) => {
  console.error('Portal error:', err);
  res.status(500).json({ error: 'Internal server error' });
});

module.exports = app;
