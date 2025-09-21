/**
 * X-Form Backend - Unified API Documentation Portal
 * Aggregates all service specifications using GitHub Spec Kit
 */

const express = require('express');
const cors = require('cors');
const helmet = require('helmet');
const path = require('path');
const fs = require('fs');
const yaml = require('yaml');
const swaggerUi = require('swagger-ui-express');
const { createProxyMiddleware } = require('http-proxy-middleware');

class UnifiedDocsPortal {
  constructor() {
    this.app = express();
    this.port = process.env.DOCS_PORT || 3000;
    this.setupMiddleware();
    this.setupRoutes();
    this.setupErrorHandling();
  }

  setupMiddleware() {
    // Security middleware
    this.app.use(helmet({
      contentSecurityPolicy: {
        directives: {
          defaultSrc: ["'self'"],
          styleSrc: ["'self'", "'unsafe-inline'", "https://fonts.googleapis.com"],
          fontSrc: ["'self'", "https://fonts.gstatic.com"],
          scriptSrc: ["'self'", "'unsafe-inline'"],
          imgSrc: ["'self'", "data:", "https:"],
        },
      },
    }));

    // CORS
    this.app.use(cors({
      origin: process.env.ALLOWED_ORIGINS?.split(',') || ['http://localhost:3000'],
      credentials: true
    }));

    // JSON parsing
    this.app.use(express.json());
    this.app.use(express.urlencoded({ extended: true }));

    // Static files
    this.app.use('/static', express.static(path.join(__dirname, '../docs')));
  }

  setupRoutes() {
    // Main documentation portal
    this.app.get('/', this.renderPortalIndex.bind(this));

    // Service-specific documentation
    this.app.get('/docs/:service', this.renderServiceDocs.bind(this));

    // API specifications
    this.app.get('/specs/:service', this.getServiceSpec.bind(this));
    this.app.get('/specs', this.getAllSpecs.bind(this));

    // Combined specification
    this.app.get('/openapi.yaml', this.getCombinedSpec.bind(this));
    this.app.get('/openapi.json', this.getCombinedSpecJSON.bind(this));

    // Interactive documentation
    this.setupSwaggerUI();

    // Health check
    this.app.get('/health', (req, res) => {
      res.json({
        success: true,
        data: {
          status: 'healthy',
          timestamp: new Date().toISOString(),
          version: '1.0.0',
          services: this.getAvailableServices()
        }
      });
    });

    // API testing interface
    this.app.get('/test', this.renderAPITester.bind(this));

    // Service proxy for testing
    this.setupServiceProxies();
  }

  setupSwaggerUI() {
    // Combined API documentation
    this.app.use('/docs', swaggerUi.serve);
    this.app.get('/docs', swaggerUi.setup(null, {
      explorer: true,
      swaggerOptions: {
        urls: [
          {
            url: '/openapi.json',
            name: 'X-Form Backend API (Complete)'
          },
          {
            url: '/specs/auth-service',
            name: 'Auth Service'
          },
          {
            url: '/specs/form-service',
            name: 'Form Service'
          },
          {
            url: '/specs/response-service',
            name: 'Response Service'
          },
          {
            url: '/specs/analytics-service',
            name: 'Analytics Service'
          },
          {
            url: '/specs/realtime-service',
            name: 'Realtime Service'
          }
        ],
        defaultModelsExpandDepth: 2,
        defaultModelExpandDepth: 2,
        displayRequestDuration: true,
        filter: true,
        showExtensions: true,
        showCommonExtensions: true,
        tryItOutEnabled: true
      },
      customCss: `
        .swagger-ui .topbar { display: none; }
        .swagger-ui .info { margin: 40px 0; }
        .swagger-ui .info .title { 
          color: #2563eb; 
          font-size: 2.5rem; 
          font-weight: bold;
          margin-bottom: 1rem;
        }
        .swagger-ui .info .description { 
          font-size: 1.1rem; 
          line-height: 1.6;
          color: #374151;
        }
        .swagger-ui .scheme-container { 
          background: #f8fafc; 
          padding: 20px;
          border-radius: 8px;
          margin: 20px 0;
        }
      `,
      customSiteTitle: 'X-Form Backend API Documentation'
    }));
  }

  setupServiceProxies() {
    const services = {
      'auth-service': process.env.AUTH_SERVICE_URL || 'http://localhost:3001',
      'form-service': process.env.FORM_SERVICE_URL || 'http://localhost:8001',
      'response-service': process.env.RESPONSE_SERVICE_URL || 'http://localhost:3002',
      'realtime-service': process.env.REALTIME_SERVICE_URL || 'http://localhost:8002',
      'analytics-service': process.env.ANALYTICS_SERVICE_URL || 'http://localhost:5001'
    };

    Object.entries(services).forEach(([serviceName, serviceUrl]) => {
      this.app.use(`/proxy/${serviceName}`, createProxyMiddleware({
        target: serviceUrl,
        changeOrigin: true,
        pathRewrite: {
          [`^/proxy/${serviceName}`]: ''
        },
        onError: (err, req, res) => {
          res.status(502).json({
            success: false,
            error: {
              code: 'SERVICE_UNAVAILABLE',
              message: `${serviceName} is not available`,
              details: { service: serviceName, target: serviceUrl }
            }
          });
        }
      }));
    });
  }

  renderPortalIndex(req, res) {
    const services = this.getAvailableServices();
    
    const html = `
    <!DOCTYPE html>
    <html lang="en">
    <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <title>X-Form Backend API Documentation Portal</title>
      <link href="https://cdn.tailwindcss.com/2.2.19/tailwind.min.css" rel="stylesheet">
      <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" rel="stylesheet">
    </head>
    <body class="bg-gray-50">
      <div class="min-h-screen">
        <!-- Header -->
        <div class="bg-white shadow-sm border-b">
          <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div class="flex justify-between items-center py-6">
              <div class="flex items-center">
                <h1 class="text-3xl font-bold text-gray-900">X-Form Backend</h1>
                <span class="ml-3 px-3 py-1 bg-blue-100 text-blue-800 text-sm font-medium rounded-full">API Documentation</span>
              </div>
              <div class="flex items-center space-x-4">
                <a href="/health" class="text-gray-500 hover:text-gray-700">
                  <i class="fas fa-heartbeat"></i> Health
                </a>
                <a href="https://github.com/Mir00r/X-Form-Backend" class="text-gray-500 hover:text-gray-700">
                  <i class="fab fa-github"></i> GitHub
                </a>
              </div>
            </div>
          </div>
        </div>

        <!-- Main Content -->
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
          <!-- Overview -->
          <div class="mb-12">
            <div class="text-center">
              <h2 class="text-4xl font-bold text-gray-900 mb-4">
                Comprehensive API Documentation
              </h2>
              <p class="text-xl text-gray-600 max-w-3xl mx-auto">
                Modern, microservices-based form management platform built with Clean Architecture principles.
                Explore our APIs, test endpoints, and integrate with confidence.
              </p>
            </div>
          </div>

          <!-- Quick Access -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-8 mb-12">
            <div class="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow">
              <div class="flex items-center mb-4">
                <i class="fas fa-book text-blue-500 text-2xl mr-3"></i>
                <h3 class="text-xl font-semibold text-gray-900">Complete API Docs</h3>
              </div>
              <p class="text-gray-600 mb-4">Interactive documentation for all services with examples and testing capabilities.</p>
              <a href="/docs" class="inline-flex items-center text-blue-600 hover:text-blue-800 font-medium">
                View Documentation <i class="fas fa-arrow-right ml-2"></i>
              </a>
            </div>

            <div class="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow">
              <div class="flex items-center mb-4">
                <i class="fas fa-flask text-green-500 text-2xl mr-3"></i>
                <h3 class="text-xl font-semibold text-gray-900">API Testing</h3>
              </div>
              <p class="text-gray-600 mb-4">Test API endpoints directly from the browser with our integrated testing interface.</p>
              <a href="/test" class="inline-flex items-center text-green-600 hover:text-green-800 font-medium">
                Start Testing <i class="fas fa-arrow-right ml-2"></i>
              </a>
            </div>

            <div class="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow">
              <div class="flex items-center mb-4">
                <i class="fas fa-download text-purple-500 text-2xl mr-3"></i>
                <h3 class="text-xl font-semibold text-gray-900">OpenAPI Specs</h3>
              </div>
              <p class="text-gray-600 mb-4">Download complete OpenAPI specifications for code generation and integration.</p>
              <a href="/openapi.yaml" class="inline-flex items-center text-purple-600 hover:text-purple-800 font-medium">
                Download Specs <i class="fas fa-arrow-right ml-2"></i>
              </a>
            </div>
          </div>

          <!-- Services -->
          <div class="mb-12">
            <h3 class="text-2xl font-bold text-gray-900 mb-8 text-center">Available Services</h3>
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              ${services.map(service => `
                <div class="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow">
                  <div class="flex items-center justify-between mb-4">
                    <h4 class="text-lg font-semibold text-gray-900">${service.name}</h4>
                    <span class="px-2 py-1 bg-${service.status === 'healthy' ? 'green' : 'red'}-100 text-${service.status === 'healthy' ? 'green' : 'red'}-800 text-xs font-medium rounded">
                      ${service.status}
                    </span>
                  </div>
                  <p class="text-gray-600 text-sm mb-4">${service.description}</p>
                  <div class="flex items-center justify-between">
                    <span class="text-sm text-gray-500">${service.technology}</span>
                    <a href="/docs/${service.id}" class="text-blue-600 hover:text-blue-800 text-sm font-medium">
                      View Docs <i class="fas fa-external-link-alt ml-1"></i>
                    </a>
                  </div>
                </div>
              `).join('')}
            </div>
          </div>

          <!-- Architecture Overview -->
          <div class="bg-white rounded-lg shadow-md p-8">
            <h3 class="text-2xl font-bold text-gray-900 mb-6">Architecture Overview</h3>
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
              <div>
                <h4 class="text-lg font-semibold text-gray-900 mb-4">🏗️ Design Principles</h4>
                <ul class="space-y-2 text-gray-600">
                  <li><strong>Clean Architecture:</strong> Domain-driven design with SOLID principles</li>
                  <li><strong>Microservices:</strong> Independent, scalable service architecture</li>
                  <li><strong>Event-Driven:</strong> Asynchronous communication patterns</li>
                  <li><strong>API-First:</strong> Contract-driven development approach</li>
                </ul>
              </div>
              <div>
                <h4 class="text-lg font-semibold text-gray-900 mb-4">🔧 Technology Stack</h4>
                <ul class="space-y-2 text-gray-600">
                  <li><strong>Gateway:</strong> Traefik (All-in-One)</li>
                  <li><strong>Services:</strong> Node.js, Go, Python</li>
                  <li><strong>Databases:</strong> PostgreSQL, Redis, Firestore</li>
                  <li><strong>Documentation:</strong> OpenAPI 3.0 + GitHub Spec Kit</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </div>
    </body>
    </html>
    `;

    res.send(html);
  }

  renderAPITester(req, res) {
    const html = `
    <!DOCTYPE html>
    <html lang="en">
    <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <title>X-Form API Tester</title>
      <link href="https://cdn.tailwindcss.com/2.2.19/tailwind.min.css" rel="stylesheet">
      <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" rel="stylesheet">
    </head>
    <body class="bg-gray-50">
      <div class="min-h-screen p-8">
        <div class="max-w-6xl mx-auto">
          <div class="bg-white rounded-lg shadow-lg p-8">
            <h1 class="text-3xl font-bold text-gray-900 mb-8">API Testing Interface</h1>
            
            <!-- API Tester will be implemented here -->
            <div class="text-center py-12">
              <i class="fas fa-flask text-6xl text-blue-500 mb-4"></i>
              <h2 class="text-2xl font-semibold text-gray-700 mb-4">Interactive API Testing</h2>
              <p class="text-gray-600 mb-8">Test API endpoints directly from your browser</p>
              <a href="/docs" class="inline-flex items-center px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
                <i class="fas fa-book mr-2"></i>
                Use Swagger UI for Testing
              </a>
            </div>
          </div>
        </div>
      </div>
    </body>
    </html>
    `;

    res.send(html);
  }

  async getServiceSpec(req, res) {
    const { service } = req.params;
    const specPath = path.join(__dirname, `../specs/services/${service}.yaml`);
    
    try {
      if (!fs.existsSync(specPath)) {
        return res.status(404).json({
          success: false,
          error: {
            code: 'SPEC_NOT_FOUND',
            message: `Specification for service '${service}' not found`
          }
        });
      }

      const specContent = fs.readFileSync(specPath, 'utf8');
      const spec = yaml.parse(specContent);
      
      res.json(spec);
    } catch (error) {
      res.status(500).json({
        success: false,
        error: {
          code: 'SPEC_PARSE_ERROR',
          message: 'Failed to parse service specification',
          details: error.message
        }
      });
    }
  }

  async getAllSpecs(req, res) {
    const specsDir = path.join(__dirname, '../specs/services');
    
    try {
      const files = fs.readdirSync(specsDir).filter(file => file.endsWith('.yaml'));
      const specs = {};

      for (const file of files) {
        const serviceName = file.replace('.yaml', '');
        const specPath = path.join(specsDir, file);
        const specContent = fs.readFileSync(specPath, 'utf8');
        specs[serviceName] = yaml.parse(specContent);
      }

      res.json({
        success: true,
        data: specs,
        meta: {
          services: Object.keys(specs),
          count: Object.keys(specs).length
        }
      });
    } catch (error) {
      res.status(500).json({
        success: false,
        error: {
          code: 'SPECS_LOAD_ERROR',
          message: 'Failed to load service specifications',
          details: error.message
        }
      });
    }
  }

  async getCombinedSpec(req, res) {
    const mainSpecPath = path.join(__dirname, '../specs/openapi.yaml');
    
    try {
      const specContent = fs.readFileSync(mainSpecPath, 'utf8');
      res.type('yaml').send(specContent);
    } catch (error) {
      res.status(500).json({
        success: false,
        error: {
          code: 'MAIN_SPEC_ERROR',
          message: 'Failed to load main specification',
          details: error.message
        }
      });
    }
  }

  async getCombinedSpecJSON(req, res) {
    const mainSpecPath = path.join(__dirname, '../specs/openapi.yaml');
    
    try {
      const specContent = fs.readFileSync(mainSpecPath, 'utf8');
      const spec = yaml.parse(specContent);
      res.json(spec);
    } catch (error) {
      res.status(500).json({
        success: false,
        error: {
          code: 'MAIN_SPEC_ERROR',
          message: 'Failed to load main specification',
          details: error.message
        }
      });
    }
  }

  getAvailableServices() {
    return [
      {
        id: 'auth-service',
        name: 'Auth Service',
        description: 'User authentication and authorization',
        technology: 'Node.js + TypeScript + Express',
        port: 3001,
        status: 'healthy'
      },
      {
        id: 'form-service',
        name: 'Form Service',
        description: 'Form management and configuration',
        technology: 'Go + Gin + GORM',
        port: 8001,
        status: 'healthy'
      },
      {
        id: 'response-service',
        name: 'Response Service',
        description: 'Form response collection and management',
        technology: 'Node.js + TypeScript + Express',
        port: 3002,
        status: 'healthy'
      },
      {
        id: 'realtime-service',
        name: 'Realtime Service',
        description: 'Real-time communication and WebSockets',
        technology: 'Go + WebSockets + Redis',
        port: 8002,
        status: 'healthy'
      },
      {
        id: 'analytics-service',
        name: 'Analytics Service',
        description: 'Analytics and reporting',
        technology: 'Python + FastAPI',
        port: 5001,
        status: 'healthy'
      }
    ];
  }

  setupErrorHandling() {
    // 404 handler
    this.app.use((req, res) => {
      res.status(404).json({
        success: false,
        error: {
          code: 'NOT_FOUND',
          message: 'The requested resource was not found',
          path: req.path
        }
      });
    });

    // Global error handler
    this.app.use((err, req, res, next) => {
      console.error('Unhandled error:', err);
      
      res.status(500).json({
        success: false,
        error: {
          code: 'INTERNAL_ERROR',
          message: 'An unexpected error occurred',
          ...(process.env.NODE_ENV === 'development' && { details: err.message })
        }
      });
    });
  }

  start() {
    this.app.listen(this.port, () => {
      console.log(`
🚀 X-Form API Documentation Portal is running!

📚 Documentation: http://localhost:${this.port}
🔍 Interactive Docs: http://localhost:${this.port}/docs  
🧪 API Tester: http://localhost:${this.port}/test
📋 OpenAPI Spec: http://localhost:${this.port}/openapi.yaml
❤️  Health Check: http://localhost:${this.port}/health

Environment: ${process.env.NODE_ENV || 'development'}
      `);
    });
  }
}

// Start the portal if this file is run directly
if (require.main === module) {
  const portal = new UnifiedDocsPortal();
  portal.start();
}

module.exports = UnifiedDocsPortal;
