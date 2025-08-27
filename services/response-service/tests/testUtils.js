/**
 * Test Configuration and Setup
 * Comprehensive testing utilities and configurations for the Response Service
 */

const config = require('../config/enhanced');

// Test Database Configuration
const testConfig = {
  database: {
    type: 'memory', // Use in-memory database for tests
    name: 'response_service_test',
    host: 'localhost',
    port: 27017,
    username: '',
    password: '',
    options: {
      useNewUrlParser: true,
      useUnifiedTopology: true
    }
  },
  
  jwt: {
    secret: 'test-jwt-secret-key',
    expiresIn: '1h'
  },
  
  server: {
    port: 0, // Use random available port for tests
    environment: 'test'
  },
  
  logging: {
    level: 'error', // Reduce noise in tests
    format: 'simple'
  },
  
  rateLimit: {
    windowMs: 1000,
    max: 1000 // Allow many requests in tests
  },
  
  cache: {
    enabled: false // Disable caching in tests
  },
  
  events: {
    enabled: true,
    maxListeners: 100
  },
  
  features: {
    analytics: true,
    fileUploads: true,
    realTimeUpdates: false,
    advancedValidation: true,
    responseExport: true,
    multiLanguage: false
  }
};

// Mock Data for Testing
const mockData = {
  forms: [
    {
      id: 'form_test_1',
      title: 'Test Survey Form',
      description: 'A test form for unit testing',
      status: 'active',
      schema: {
        fields: [
          {
            id: 'q1',
            type: 'text',
            label: 'What is your name?',
            required: true,
            validation: {
              minLength: 2,
              maxLength: 50
            }
          },
          {
            id: 'q2',
            type: 'email',
            label: 'What is your email?',
            required: true,
            validation: {
              format: 'email'
            }
          },
          {
            id: 'q3',
            type: 'radio',
            label: 'How satisfied are you?',
            required: false,
            options: [
              { value: 'very_satisfied', label: 'Very Satisfied' },
              { value: 'satisfied', label: 'Satisfied' },
              { value: 'neutral', label: 'Neutral' },
              { value: 'dissatisfied', label: 'Dissatisfied' }
            ]
          },
          {
            id: 'q4',
            type: 'checkbox',
            label: 'Which features do you use?',
            required: false,
            options: [
              { value: 'feature_a', label: 'Feature A' },
              { value: 'feature_b', label: 'Feature B' },
              { value: 'feature_c', label: 'Feature C' }
            ]
          },
          {
            id: 'q5',
            type: 'textarea',
            label: 'Additional comments',
            required: false,
            validation: {
              maxLength: 500
            }
          }
        ]
      },
      settings: {
        allowMultipleSubmissions: false,
        requireAuth: false,
        collectMetadata: true
      }
    }
  ],
  
  responses: [
    {
      id: 'resp_test_1',
      formId: 'form_test_1',
      formTitle: 'Test Survey Form',
      respondentEmail: 'test@example.com',
      status: 'completed',
      responses: [
        {
          questionId: 'q1',
          questionType: 'text',
          value: 'John Doe'
        },
        {
          questionId: 'q2',
          questionType: 'email',
          value: 'john@example.com'
        },
        {
          questionId: 'q3',
          questionType: 'radio',
          value: 'satisfied'
        },
        {
          questionId: 'q4',
          questionType: 'checkbox',
          value: ['feature_a', 'feature_c']
        },
        {
          questionId: 'q5',
          questionType: 'textarea',
          value: 'Great service, keep it up!'
        }
      ],
      metadata: {
        userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
        ipAddress: '192.168.1.100',
        referrer: 'https://example.com/survey',
        timeSpent: 120,
        deviceType: 'desktop',
        location: {
          country: 'United States',
          city: 'New York'
        }
      },
      submittedAt: '2023-12-07T10:30:00Z',
      updatedAt: '2023-12-07T10:30:00Z'
    },
    {
      id: 'resp_test_2',
      formId: 'form_test_1',
      formTitle: 'Test Survey Form',
      respondentEmail: 'jane@example.com',
      status: 'partial',
      responses: [
        {
          questionId: 'q1',
          questionType: 'text',
          value: 'Jane Smith'
        },
        {
          questionId: 'q2',
          questionType: 'email',
          value: 'jane@example.com'
        }
      ],
      metadata: {
        userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)',
        ipAddress: '192.168.1.101',
        timeSpent: 45,
        deviceType: 'mobile'
      },
      submittedAt: '2023-12-07T11:00:00Z',
      updatedAt: '2023-12-07T11:15:00Z'
    }
  ],
  
  users: [
    {
      id: 'user_test_1',
      email: 'admin@example.com',
      role: 'admin',
      name: 'Test Admin',
      permissions: ['read', 'write', 'delete', 'analytics']
    },
    {
      id: 'user_test_2',
      email: 'user@example.com',
      role: 'user',
      name: 'Test User',
      permissions: ['read', 'write']
    }
  ]
};

// Test Utilities
class TestUtils {
  /**
   * Generate a valid JWT token for testing
   */
  static generateTestToken(user = mockData.users[0]) {
    const jwt = require('jsonwebtoken');
    return jwt.sign(
      {
        userId: user.id,
        email: user.email,
        role: user.role,
        permissions: user.permissions
      },
      testConfig.jwt.secret,
      { expiresIn: testConfig.jwt.expiresIn }
    );
  }

  /**
   * Generate correlation ID for testing
   */
  static generateCorrelationId() {
    return `test_cor_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  /**
   * Create test request headers
   */
  static createTestHeaders(includeAuth = true, user = mockData.users[0]) {
    const headers = {
      'Content-Type': 'application/json',
      'X-Correlation-ID': this.generateCorrelationId()
    };

    if (includeAuth) {
      headers['Authorization'] = `Bearer ${this.generateTestToken(user)}`;
    }

    return headers;
  }

  /**
   * Create a test response object
   */
  static createTestResponse(overrides = {}) {
    return {
      id: `resp_test_${Date.now()}`,
      formId: 'form_test_1',
      formTitle: 'Test Survey Form',
      respondentEmail: 'test@example.com',
      status: 'draft',
      responses: [
        {
          questionId: 'q1',
          questionType: 'text',
          value: 'Test Response'
        }
      ],
      metadata: {
        userAgent: 'Test User Agent',
        ipAddress: '127.0.0.1',
        timeSpent: 60,
        deviceType: 'desktop'
      },
      submittedAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      ...overrides
    };
  }

  /**
   * Create a test file upload object
   */
  static createTestFile(overrides = {}) {
    return {
      fieldname: 'file',
      originalname: 'test-file.pdf',
      encoding: '7bit',
      mimetype: 'application/pdf',
      buffer: Buffer.from('test file content'),
      size: 1024,
      ...overrides
    };
  }

  /**
   * Wait for a specified amount of time
   */
  static async sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  /**
   * Create a mock Express request object
   */
  static createMockRequest(overrides = {}) {
    return {
      method: 'GET',
      url: '/api/v1/test',
      headers: this.createTestHeaders(),
      params: {},
      query: {},
      body: {},
      user: mockData.users[0],
      correlationId: this.generateCorrelationId(),
      ...overrides
    };
  }

  /**
   * Create a mock Express response object
   */
  static createMockResponse() {
    const res = {
      status: jest.fn().mockReturnThis(),
      json: jest.fn().mockReturnThis(),
      send: jest.fn().mockReturnThis(),
      set: jest.fn().mockReturnThis(),
      cookie: jest.fn().mockReturnThis(),
      redirect: jest.fn().mockReturnThis()
    };
    return res;
  }

  /**
   * Create a mock next function
   */
  static createMockNext() {
    return jest.fn();
  }

  /**
   * Validate response structure
   */
  static validateApiResponse(response) {
    expect(response).toHaveProperty('success');
    expect(response).toHaveProperty('data');
    expect(response).toHaveProperty('message');
    expect(response).toHaveProperty('correlationId');
    expect(response).toHaveProperty('timestamp');
    expect(typeof response.success).toBe('boolean');
    expect(typeof response.message).toBe('string');
    expect(typeof response.correlationId).toBe('string');
    expect(typeof response.timestamp).toBe('string');
  }

  /**
   * Validate error response structure
   */
  static validateErrorResponse(response) {
    expect(response).toHaveProperty('success', false);
    expect(response).toHaveProperty('error');
    expect(response.error).toHaveProperty('code');
    expect(response.error).toHaveProperty('message');
    expect(response).toHaveProperty('correlationId');
    expect(response).toHaveProperty('timestamp');
  }

  /**
   * Setup test database with mock data
   */
  static async setupTestDatabase() {
    // This would be implemented based on the chosen database
    // For now, we'll use in-memory storage
    const mockDatabase = {
      responses: new Map(),
      forms: new Map(),
      users: new Map()
    };

    // Load mock data
    mockData.responses.forEach(response => {
      mockDatabase.responses.set(response.id, response);
    });

    mockData.forms.forEach(form => {
      mockDatabase.forms.set(form.id, form);
    });

    mockData.users.forEach(user => {
      mockDatabase.users.set(user.id, user);
    });

    return mockDatabase;
  }

  /**
   * Cleanup test database
   */
  static async cleanupTestDatabase(database) {
    if (database) {
      database.responses.clear();
      database.forms.clear();
      database.users.clear();
    }
  }

  /**
   * Setup test server
   */
  static async setupTestServer() {
    const express = require('express');
    const app = express();
    
    // Basic middleware
    app.use(express.json());
    app.use(express.urlencoded({ extended: true }));
    
    // Test routes would be added here
    
    const server = app.listen(0); // Use random port
    const port = server.address().port;
    
    return { app, server, port };
  }

  /**
   * Cleanup test server
   */
  static async cleanupTestServer(server) {
    if (server) {
      return new Promise((resolve) => {
        server.close(resolve);
      });
    }
  }
}

// Jest Configuration
const jestConfig = {
  testEnvironment: 'node',
  testMatch: [
    '**/__tests__/**/*.js',
    '**/?(*.)+(spec|test).js'
  ],
  collectCoverageFrom: [
    'src/**/*.js',
    '!src/**/*.test.js',
    '!src/**/*.spec.js',
    '!src/config/**',
    '!src/scripts/**'
  ],
  coverageDirectory: 'coverage',
  coverageReporters: ['text', 'lcov', 'html'],
  setupFilesAfterEnv: ['<rootDir>/tests/setup.js'],
  testTimeout: 10000,
  verbose: true,
  detectOpenHandles: true,
  forceExit: true
};

module.exports = {
  testConfig,
  mockData,
  TestUtils,
  jestConfig
};
