const swaggerJsdoc = require('swagger-jsdoc');
const swaggerUi = require('swagger-ui-express');

const options = {
  definition: {
    openapi: '3.0.0',
    info: {
      title: 'Realtime Service API',
      version: '1.0.0',
      description: `
# Realtime Service API

Real-time WebSocket service for X-Form Backend providing live communication capabilities.

## Features

- **Real-time Communication**: WebSocket-based real-time messaging
- **Form Subscriptions**: Subscribe to live form updates and responses
- **Event Broadcasting**: Broadcast events to subscribed clients
- **Health Monitoring**: Comprehensive health checks and monitoring
- **Secure Connections**: JWT-based authentication and CORS protection

## Architecture

- **WebSocket Server**: Socket.IO for real-time bidirectional communication
- **Event-Driven**: Event-based architecture for scalable real-time features
- **Redis Integration**: Optional Redis adapter for horizontal scaling
- **RESTful APIs**: Traditional REST endpoints for service management
- **Microservices Ready**: Designed for microservices architecture

## WebSocket Events

### Client to Server Events
- \`form:subscribe\` - Subscribe to form updates
- \`form:unsubscribe\` - Unsubscribe from form updates
- \`response:new\` - Send new response data

### Server to Client Events
- \`response:update\` - Real-time response updates
- \`form:published\` - Form publication notifications
- \`form:closed\` - Form closure notifications

## Security Features

- CORS protection with configurable origins
- Helmet.js security headers
- JWT token validation (optional)
- Rate limiting and connection management
      `,
      contact: {
        name: 'Realtime Service Team',
        email: 'realtime-service@example.com',
        url: 'https://api.example.com/support'
      },
      license: {
        name: 'MIT',
        url: 'https://opensource.org/licenses/MIT'
      }
    },
    servers: [
      {
        url: 'https://api.example.com/realtime',
        description: 'Production server'
      },
      {
        url: 'https://staging-api.example.com/realtime',
        description: 'Staging server'
      },
      {
        url: 'http://localhost:8002',
        description: 'Development server'
      }
    ],
    tags: [
      {
        name: 'Health',
        description: 'Health check and monitoring endpoints'
      },
      {
        name: 'WebSocket',
        description: 'WebSocket connection information and documentation'
      },
      {
        name: 'Events',
        description: 'Real-time event management'
      },
      {
        name: 'Monitoring',
        description: 'Service monitoring and metrics'
      }
    ],
    components: {
      securitySchemes: {
        BearerAuth: {
          type: 'http',
          scheme: 'bearer',
          bearerFormat: 'JWT',
          description: 'JWT Bearer token for authentication'
        }
      },
      schemas: {
        HealthCheck: {
          type: 'object',
          required: ['status', 'timestamp', 'service'],
          properties: {
            status: {
              type: 'string',
              enum: ['healthy', 'unhealthy'],
              description: 'Current health status of the service',
              example: 'healthy'
            },
            service: {
              type: 'string',
              description: 'Service name',
              example: 'realtime-service'
            },
            version: {
              type: 'string',
              description: 'Service version',
              example: '1.0.0'
            },
            timestamp: {
              type: 'string',
              format: 'date-time',
              description: 'Timestamp when health check was performed',
              example: '2024-01-15T10:30:00.000Z'
            },
            uptime: {
              type: 'integer',
              description: 'Service uptime in seconds',
              minimum: 0,
              example: 3600
            },
            connections: {
              $ref: '#/components/schemas/ConnectionSummary'
            },
            memory: {
              $ref: '#/components/schemas/MemoryUsage'
            },
            environment: {
              type: 'string',
              description: 'Current environment',
              enum: ['development', 'staging', 'production'],
              example: 'development'
            }
          }
        },
        DetailedHealthCheck: {
          type: 'object',
          required: ['status', 'timestamp', 'service'],
          properties: {
            status: {
              type: 'string',
              enum: ['healthy', 'unhealthy'],
              description: 'Current health status of the service',
              example: 'healthy'
            },
            timestamp: {
              type: 'string',
              format: 'date-time',
              example: '2024-01-15T10:30:00.000Z'
            },
            uptime: {
              type: 'integer',
              description: 'Service uptime in seconds',
              example: 3600
            },
            service: {
              $ref: '#/components/schemas/ServiceInfo'
            },
            connections: {
              $ref: '#/components/schemas/DetailedConnectionInfo'
            },
            rooms: {
              type: 'object',
              additionalProperties: {
                type: 'integer'
              },
              description: 'Active rooms and their subscriber counts',
              example: {
                'form:123': 5,
                'form:456': 3
              }
            },
            memory: {
              $ref: '#/components/schemas/DetailedMemoryUsage'
            },
            performance: {
              $ref: '#/components/schemas/PerformanceMetrics'
            },
            socketio: {
              $ref: '#/components/schemas/SocketIOStatus'
            }
          }
        },
        ServiceInfo: {
          type: 'object',
          properties: {
            name: {
              type: 'string',
              example: 'realtime-service'
            },
            version: {
              type: 'string',
              example: '1.0.0'
            },
            environment: {
              type: 'string',
              enum: ['development', 'staging', 'production'],
              example: 'development'
            },
            nodeVersion: {
              type: 'string',
              example: 'v18.17.0'
            },
            platform: {
              type: 'string',
              example: 'darwin'
            },
            pid: {
              type: 'integer',
              example: 12345
            }
          }
        },
        ConnectionSummary: {
          type: 'object',
          properties: {
            active: {
              type: 'integer',
              description: 'Number of currently active connections',
              minimum: 0,
              example: 25
            },
            total: {
              type: 'integer',
              description: 'Total connections since service started',
              minimum: 0,
              example: 150
            }
          }
        },
        DetailedConnectionInfo: {
          type: 'object',
          properties: {
            active: {
              type: 'integer',
              description: 'Number of currently active connections',
              example: 25
            },
            total: {
              type: 'integer',
              description: 'Total connections since service started',
              example: 150
            },
            roomSubscriptions: {
              type: 'integer',
              description: 'Total number of room subscriptions',
              example: 45
            },
            rooms: {
              type: 'integer',
              description: 'Number of active rooms',
              example: 8
            }
          }
        },
        MemoryUsage: {
          type: 'object',
          properties: {
            used: {
              type: 'integer',
              description: 'Used memory in MB',
              example: 128
            },
            total: {
              type: 'integer',
              description: 'Total allocated memory in MB',
              example: 256
            },
            external: {
              type: 'integer',
              description: 'External memory usage in MB',
              example: 16
            }
          }
        },
        DetailedMemoryUsage: {
          type: 'object',
          properties: {
            used: {
              type: 'integer',
              description: 'Used heap memory in MB',
              example: 128
            },
            total: {
              type: 'integer',
              description: 'Total heap memory in MB',
              example: 256
            },
            external: {
              type: 'integer',
              description: 'External memory usage in MB',
              example: 16
            },
            rss: {
              type: 'integer',
              description: 'Resident Set Size in MB',
              example: 180
            }
          }
        },
        PerformanceMetrics: {
          type: 'object',
          properties: {
            eventsPerSecond: {
              type: 'number',
              description: 'Average events processed per second',
              example: 2.5
            },
            uptimeSeconds: {
              type: 'number',
              description: 'Service uptime in seconds',
              example: 3600
            },
            cpuUsage: {
              type: 'object',
              properties: {
                user: {
                  type: 'integer',
                  description: 'User CPU time in microseconds'
                },
                system: {
                  type: 'integer',
                  description: 'System CPU time in microseconds'
                }
              }
            }
          }
        },
        SocketIOStatus: {
          type: 'object',
          properties: {
            version: {
              type: 'string',
              example: '4.7.4'
            },
            engine: {
              type: 'string',
              enum: ['running', 'not initialized'],
              example: 'running'
            },
            transports: {
              type: 'array',
              items: {
                type: 'string',
                enum: ['websocket', 'polling']
              },
              example: ['websocket', 'polling']
            }
          }
        },
        WebSocketInfo: {
          type: 'object',
          required: ['endpoint', 'protocol', 'version'],
          properties: {
            endpoint: {
              type: 'string',
              format: 'uri',
              description: 'WebSocket connection endpoint',
              example: 'ws://localhost:8002'
            },
            protocol: {
              type: 'string',
              description: 'WebSocket protocol used',
              example: 'socket.io'
            },
            version: {
              type: 'string',
              description: 'Socket.IO version',
              example: '4.7.4'
            },
            transports: {
              type: 'array',
              items: {
                type: 'string',
                enum: ['websocket', 'polling']
              },
              description: 'Supported transport methods',
              example: ['websocket', 'polling']
            },
            cors: {
              $ref: '#/components/schemas/CorsConfig'
            },
            authentication: {
              $ref: '#/components/schemas/AuthConfig'
            },
            events: {
              $ref: '#/components/schemas/EventConfig'
            },
            connection: {
              $ref: '#/components/schemas/ConnectionConfig'
            },
            timestamp: {
              type: 'string',
              format: 'date-time',
              example: '2024-01-15T10:30:00.000Z'
            }
          }
        },
        CorsConfig: {
          type: 'object',
          properties: {
            enabled: {
              type: 'boolean',
              description: 'Whether CORS is enabled',
              example: true
            },
            origins: {
              type: 'array',
              items: {
                type: 'string'
              },
              description: 'Allowed origins for CORS',
              example: ['http://localhost:3000']
            }
          }
        },
        AuthConfig: {
          type: 'object',
          properties: {
            required: {
              type: 'boolean',
              description: 'Whether authentication is required',
              example: false
            },
            method: {
              type: 'string',
              description: 'Authentication method',
              example: 'JWT'
            },
            header: {
              type: 'string',
              description: 'Authentication header field',
              example: 'auth.token'
            }
          }
        },
        EventConfig: {
          type: 'object',
          properties: {
            supported: {
              type: 'array',
              items: {
                type: 'string'
              },
              description: 'Supported client-to-server events',
              example: ['form:subscribe', 'form:unsubscribe', 'response:new']
            },
            responses: {
              type: 'array',
              items: {
                type: 'string'
              },
              description: 'Server-to-client response events',
              example: ['form:subscribed', 'response:update', 'error']
            }
          }
        },
        ConnectionConfig: {
          type: 'object',
          properties: {
            timeout: {
              type: 'integer',
              description: 'Connection timeout in milliseconds',
              example: 20000
            },
            pingInterval: {
              type: 'integer',
              description: 'Ping interval in milliseconds',
              example: 25000
            },
            pingTimeout: {
              type: 'integer',
              description: 'Ping timeout in milliseconds',
              example: 5000
            }
          }
        },
        SocketConnection: {
          type: 'object',
          properties: {
            socketId: {
              type: 'string',
              description: 'Unique socket identifier',
              example: 'abc123def456'
            },
            userId: {
              type: 'string',
              nullable: true,
              description: 'Associated user ID (if authenticated)',
              example: 'user_123'
            },
            connected: {
              type: 'boolean',
              description: 'Whether socket is currently connected',
              example: true
            },
            rooms: {
              type: 'array',
              items: {
                type: 'string'
              },
              description: 'Rooms the socket has joined',
              example: ['form:123', 'form:456']
            },
            handshake: {
              $ref: '#/components/schemas/HandshakeInfo'
            }
          }
        },
        HandshakeInfo: {
          type: 'object',
          properties: {
            time: {
              type: 'string',
              format: 'date-time',
              description: 'Connection time',
              example: '2024-01-15T10:30:00.000Z'
            },
            address: {
              type: 'string',
              description: 'Client IP address',
              example: '127.0.0.1'
            },
            userAgent: {
              type: 'string',
              description: 'Client user agent',
              example: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)'
            }
          }
        },
        EventSubscription: {
          type: 'object',
          required: ['event', 'formId'],
          properties: {
            event: {
              type: 'string',
              description: 'Event name',
              enum: ['form:subscribe', 'form:unsubscribe'],
              example: 'form:subscribe'
            },
            formId: {
              type: 'string',
              format: 'uuid',
              description: 'Form identifier to subscribe to',
              example: '123e4567-e89b-12d3-a456-426614174000'
            },
            clientId: {
              type: 'string',
              description: 'Client socket ID',
              example: 'socket_abc123'
            },
            timestamp: {
              type: 'string',
              format: 'date-time',
              description: 'Subscription timestamp',
              example: '2024-01-15T10:30:00.000Z'
            }
          }
        },
        FormResponse: {
          type: 'object',
          required: ['responseId', 'data'],
          properties: {
            responseId: {
              type: 'string',
              format: 'uuid',
              description: 'Unique response identifier',
              example: '456e7890-e89b-12d3-a456-426614174001'
            },
            data: {
              type: 'object',
              additionalProperties: true,
              description: 'Form response data with question-answer pairs',
              example: {
                question1: 'John Doe',
                question2: 'john@example.com',
                question3: 'I love this form!'
              }
            },
            metadata: {
              $ref: '#/components/schemas/ResponseMetadata'
            }
          }
        },
        ResponseMetadata: {
          type: 'object',
          properties: {
            submittedAt: {
              type: 'string',
              format: 'date-time',
              description: 'Response submission timestamp',
              example: '2024-01-15T10:30:00.000Z'
            },
            submittedBy: {
              type: 'string',
              description: 'User ID who submitted the response',
              example: 'user_789'
            },
            ipAddress: {
              type: 'string',
              description: 'IP address of the submitter',
              example: '192.168.1.100'
            },
            userAgent: {
              type: 'string',
              description: 'User agent of the submitter',
              example: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)'
            },
            completionTime: {
              type: 'integer',
              description: 'Time taken to complete the form in seconds',
              example: 120
            }
          }
        },
        ResponseEvent: {
          type: 'object',
          required: ['formId', 'responseId', 'data'],
          properties: {
            formId: {
              type: 'string',
              format: 'uuid',
              description: 'Form identifier',
              example: '123e4567-e89b-12d3-a456-426614174000'
            },
            responseId: {
              type: 'string',
              format: 'uuid',
              description: 'Response identifier',
              example: '456e7890-e89b-12d3-a456-426614174001'
            },
            data: {
              type: 'object',
              additionalProperties: true,
              description: 'Response data',
              example: {
                question1: 'John Doe',
                question2: 'john@example.com'
              }
            },
            timestamp: {
              type: 'string',
              format: 'date-time',
              description: 'Event timestamp',
              example: '2024-01-15T10:30:00.000Z'
            },
            userId: {
              type: 'string',
              format: 'uuid',
              nullable: true,
              description: 'User who submitted the response',
              example: '789e0123-e89b-12d3-a456-426614174002'
            },
            socketId: {
              type: 'string',
              description: 'Socket ID that triggered the event',
              example: 'socket_abc123'
            },
            source: {
              type: 'string',
              enum: ['websocket', 'api'],
              description: 'Source of the event',
              example: 'api'
            }
          }
        },
        FormStatusUpdate: {
          type: 'object',
          required: ['status'],
          properties: {
            status: {
              type: 'string',
              enum: ['published', 'closed', 'archived', 'draft'],
              description: 'New form status',
              example: 'published'
            },
            reason: {
              type: 'string',
              description: 'Reason for status change',
              example: 'Form is now ready for responses'
            },
            updatedBy: {
              type: 'string',
              description: 'ID of user who updated the status',
              example: 'admin123'
            }
          }
        },
        ConnectionStats: {
          type: 'object',
          required: ['totalConnections', 'activeConnections'],
          properties: {
            totalConnections: {
              type: 'integer',
              minimum: 0,
              description: 'Total connections since service started',
              example: 150
            },
            activeConnections: {
              type: 'integer',
              minimum: 0,
              description: 'Currently active connections',
              example: 25
            },
            roomSubscriptions: {
              type: 'object',
              additionalProperties: {
                type: 'integer'
              },
              description: 'Active room subscriptions with subscriber counts',
              example: {
                'form:123e4567-e89b-12d3-a456-426614174000': 5,
                'form:456e7890-e89b-12d3-a456-426614174001': 3
              }
            },
            eventsPerSecond: {
              type: 'number',
              minimum: 0,
              description: 'Average events processed per second',
              example: 2.5
            },
            uptime: {
              type: 'integer',
              minimum: 0,
              description: 'Service uptime in seconds',
              example: 3600
            },
            server: {
              $ref: '#/components/schemas/ServerInfo'
            },
            timestamp: {
              type: 'string',
              format: 'date-time',
              description: 'Timestamp when stats were generated',
              example: '2024-01-15T10:30:00.000Z'
            }
          }
        },
        ServerInfo: {
          type: 'object',
          properties: {
            memory: {
              $ref: '#/components/schemas/DetailedMemoryUsage'
            },
            nodeVersion: {
              type: 'string',
              description: 'Node.js version',
              example: 'v18.17.0'
            },
            platform: {
              type: 'string',
              description: 'Operating system platform',
              example: 'darwin'
            },
            pid: {
              type: 'integer',
              description: 'Process ID',
              example: 12345
            }
          }
        },
        BroadcastRequest: {
          type: 'object',
          required: ['room', 'event', 'data'],
          properties: {
            room: {
              type: 'string',
              description: 'Room name to broadcast to',
              pattern: '^form:[a-zA-Z0-9_-]+$',
              example: 'form:12345'
            },
            event: {
              type: 'string',
              description: 'Event name to emit',
              example: 'admin:announcement'
            },
            data: {
              type: 'object',
              additionalProperties: true,
              description: 'Data to send with the event',
              example: {
                message: 'System maintenance in 10 minutes',
                priority: 'high'
              }
            }
          }
        },
        BroadcastResponse: {
          type: 'object',
          properties: {
            success: {
              type: 'boolean',
              description: 'Whether broadcast was successful',
              example: true
            },
            message: {
              type: 'string',
              description: 'Success message',
              example: 'Message broadcasted to room form:12345'
            },
            recipients: {
              type: 'integer',
              minimum: 0,
              description: 'Number of recipients',
              example: 5
            },
            event: {
              type: 'string',
              description: 'Event that was broadcasted',
              example: 'admin:announcement'
            },
            timestamp: {
              type: 'string',
              format: 'date-time',
              example: '2024-01-15T10:30:00.000Z'
            }
          }
        },
        NotificationRequest: {
          type: 'object',
          required: ['event', 'data'],
          properties: {
            event: {
              type: 'string',
              description: 'Event name to emit',
              example: 'form:updated'
            },
            data: {
              type: 'object',
              additionalProperties: true,
              description: 'Event data',
              example: {
                title: 'New Form Title',
                updatedBy: 'admin'
              }
            },
            urgent: {
              type: 'boolean',
              description: 'Whether this is an urgent notification',
              default: false,
              example: false
            }
          }
        },
        ErrorResponse: {
          type: 'object',
          required: ['error', 'timestamp'],
          properties: {
            error: {
              type: 'string',
              description: 'Error message',
              example: 'Service unavailable'
            },
            message: {
              type: 'string',
              description: 'Detailed error description',
              example: 'The service is temporarily unavailable'
            },
            timestamp: {
              type: 'string',
              format: 'date-time',
              description: 'Error timestamp',
              example: '2024-01-15T10:30:00.000Z'
            },
            code: {
              type: 'string',
              description: 'Error code for programmatic handling',
              example: 'REALTIME_SERVICE_ERROR'
            },
            details: {
              type: 'object',
              additionalProperties: true,
              description: 'Additional error details',
              example: {
                statusCode: 503,
                retryAfter: 30
              }
            }
          }
        }
      }
    },
    externalDocs: {
      description: 'Socket.IO Documentation',
      url: 'https://socket.io/docs/v4/'
    }
  },
  apis: ['./src/**/*.js'], // Path to the API files
};

const swaggerSpec = swaggerJsdoc(options);

module.exports = {
  swaggerSpec,
  serve: swaggerUi.serve,
  setup: swaggerUi.setup(swaggerSpec, {
    customCss: `
      .swagger-ui .topbar { 
        display: none; 
      }
      .swagger-ui .info .title { 
        color: #1e40af; 
        font-size: 2.5rem;
        font-weight: 700;
        margin-bottom: 1rem;
      }
      .swagger-ui .info .description {
        color: #374151;
        font-size: 1.1rem;
        line-height: 1.6;
      }
      .swagger-ui .scheme-container { 
        background: linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%); 
        padding: 24px; 
        margin: 24px 0; 
        border-radius: 12px; 
        border: 1px solid #cbd5e1;
        box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
      }
      .swagger-ui .opblock.opblock-get {
        border-color: #059669;
        background: rgba(5, 150, 105, 0.1);
      }
      .swagger-ui .opblock.opblock-post {
        border-color: #dc2626;
        background: rgba(220, 38, 38, 0.1);
      }
      .swagger-ui .opblock.opblock-put {
        border-color: #d97706;
        background: rgba(217, 119, 6, 0.1);
      }
      .swagger-ui .opblock.opblock-delete {
        border-color: #dc2626;
        background: rgba(220, 38, 38, 0.15);
      }
      .swagger-ui .opblock-summary {
        font-weight: 600;
        font-size: 1rem;
      }
      .swagger-ui .opblock-description-wrapper p {
        color: #4b5563;
        margin-bottom: 0.5rem;
      }
      .swagger-ui .btn.authorize {
        background: #3b82f6;
        border-color: #3b82f6;
        color: white;
        font-weight: 600;
        padding: 8px 16px;
        border-radius: 8px;
      }
      .swagger-ui .btn.authorize:hover {
        background: #2563eb;
        border-color: #2563eb;
      }
      .swagger-ui .model-box {
        background: #f9fafb;
        border: 1px solid #e5e7eb;
        border-radius: 8px;
        padding: 16px;
        margin: 8px 0;
      }
      .swagger-ui .model .property {
        padding: 8px 0;
        border-bottom: 1px solid #f3f4f6;
      }
      .swagger-ui .model .property:last-child {
        border-bottom: none;
      }
      .swagger-ui .parameters-col_description input[type=text] {
        border-radius: 6px;
        border: 1px solid #d1d5db;
        padding: 8px 12px;
      }
      .swagger-ui .parameters-col_description input[type=text]:focus {
        border-color: #3b82f6;
        box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
        outline: none;
      }
      .swagger-ui .response-col_status {
        font-weight: 600;
        font-size: 0.9rem;
      }
      .swagger-ui .response-col_description {
        color: #6b7280;
      }
      .swagger-ui .responses-inner h4 {
        color: #1f2937;
        font-weight: 600;
        margin-bottom: 12px;
      }
      .swagger-ui .operation-tag-content {
        max-width: none;
      }
      .swagger-ui .opblock-tag {
        border-bottom: 2px solid #e5e7eb;
        padding: 16px 0;
        margin-bottom: 16px;
      }
      .swagger-ui .opblock-tag a {
        font-size: 1.5rem;
        font-weight: 700;
        color: #1f2937;
        text-decoration: none;
      }
      .swagger-ui .opblock-tag small {
        color: #6b7280;
        font-size: 0.9rem;
        font-weight: 400;
        margin-left: 8px;
      }
      .swagger-ui .info .title small {
        background: #dbeafe;
        color: #1e40af;
        padding: 4px 8px;
        border-radius: 6px;
        font-size: 0.8rem;
        font-weight: 600;
        margin-left: 12px;
      }
      .swagger-ui .markdown p {
        margin-bottom: 0.8rem;
        line-height: 1.6;
      }
      .swagger-ui .markdown h1, .swagger-ui .markdown h2, .swagger-ui .markdown h3 {
        color: #1f2937;
        font-weight: 600;
        margin-top: 1.5rem;
        margin-bottom: 0.8rem;
      }
      .swagger-ui .markdown ul, .swagger-ui .markdown ol {
        margin-bottom: 1rem;
        padding-left: 1.5rem;
      }
      .swagger-ui .markdown li {
        margin-bottom: 0.4rem;
        line-height: 1.5;
      }
      .swagger-ui .markdown code {
        background: #f3f4f6;
        padding: 2px 6px;
        border-radius: 4px;
        font-size: 0.9rem;
        color: #dc2626;
        font-weight: 600;
      }
      .swagger-ui .highlight-code .microlight {
        background: #1f2937;
        border-radius: 8px;
        padding: 16px;
      }
      @media (max-width: 768px) {
        .swagger-ui .info .title {
          font-size: 2rem;
        }
        .swagger-ui .scheme-container {
          padding: 16px;
          margin: 16px 0;
        }
      }
    `,
    customSiteTitle: 'X-Form Realtime Service API Documentation',
    customfavIcon: 'https://swagger.io/favicon-32x32.png',
    swaggerOptions: {
      persistAuthorization: true,
      displayRequestDuration: true,
      filter: true,
      tryItOutEnabled: true,
      requestInterceptor: (req) => {
        req.headers['X-API-Client'] = 'Swagger-UI';
        return req;
      },
      responseInterceptor: (res) => {
        console.log('Response received:', res.status);
        return res;
      },
      docExpansion: 'list',
      defaultModelsExpandDepth: 2,
      defaultModelExpandDepth: 3,
      showExtensions: true,
      showCommonExtensions: true,
      tagsSorter: 'alpha',
      operationsSorter: 'alpha'
    },
    explorer: true
  })
};
