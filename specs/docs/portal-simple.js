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

// Auto-discover service specifications
function discoverServices() {
  const servicesDir = path.join(__dirname, '../services');
  const services = [];
  
  try {
    const files = fs.readdirSync(servicesDir);
    files.forEach(file => {
      if (file.endsWith('-service.yaml')) {
        const serviceName = file.replace('-service.yaml', '');
        const specPath = path.join(servicesDir, file);
        const spec = loadSpec(specPath);
        
        if (spec && spec.info) {
          services.push({
            name: serviceName,
            title: spec.info.title || serviceName,
            description: spec.info.description ? spec.info.description.split('\n')[0] : `${serviceName} service`,
            version: spec.info.version || '1.0.0',
            icon: getServiceIcon(serviceName),
            path: `/docs/${serviceName}`
          });
        }
      }
    });
  } catch (error) {
    console.error('Error discovering services:', error.message);
  }
  
  return services.sort((a, b) => a.name.localeCompare(b.name));
}

// Get service icon based on service name
function getServiceIcon(serviceName) {
  const icons = {
    'auth': '🔐',
    'form': '📝',
    'response': '📊',
    'realtime': '⚡',
    'analytics': '📈',
    'collaboration': '👥',
    'event-bus': '🔄',
    'api-gateway': '🌐',
    'file': '📁'
  };
  return icons[serviceName] || '🔧';
}

// Routes
app.get('/', (req, res) => {
  const services = discoverServices();
  
  let servicesHtml = '';
  services.forEach(service => {
    servicesHtml += `
      <div class="service">
        <h3>${service.icon} ${service.title}</h3>
        <p>${service.description}</p>
        <p><small>Version: ${service.version}</small></p>
        <a href="${service.path}">→ View ${service.name} API</a>
      </div>
    `;
  });
  
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
          .stats { margin: 20px 0; padding: 15px; background: #e8f5e8; border-radius: 5px; }
        </style>
      </head>
      <body>
        <div class="header">
          <h1>🚀 X-Form Backend API Documentation</h1>
          <p>Centralized API documentation for all microservices</p>
          <p class="status">✅ GitHub Spec Kit Active</p>
          <div class="stats">
            <strong>📊 Services Detected: ${services.length}</strong>
          </div>
        </div>

        <div class="service">
          <h3>� Combined API Documentation</h3>
          <p>Complete API documentation for all services</p>
          <a href="/docs">→ View Interactive Documentation</a>
        </div>

        ${servicesHtml}

        <hr style="margin: 30px 0;">
        
        <h3>🔧 Developer Resources</h3>
        <ul>
          <li><a href="/openapi.yaml">Download OpenAPI YAML</a></li>
          <li><a href="/openapi.json">Download OpenAPI JSON</a></li>
          <li><a href="/health">Health Check</a></li>
          <li><a href="/services">Services API (JSON)</a></li>
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
  const services = discoverServices();
  const serviceHealth = {};
  
  services.forEach(service => {
    serviceHealth[`${service.name}-service`] = true;
  });
  
  res.json({
    status: 'healthy',
    timestamp: new Date().toISOString(),
    version: '1.0.0',
    servicesDetected: services.length,
    services: serviceHealth
  });
});

// New services API endpoint
app.get('/services', (req, res) => {
  const services = discoverServices();
  res.json({
    count: services.length,
    services: services
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
