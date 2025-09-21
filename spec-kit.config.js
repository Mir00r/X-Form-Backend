/**
 * GitHub Spec Kit Configuration for X-Form Backend
 * Centralized API specification management for microservices architecture
 */

module.exports = {
  // Project Information
  project: {
    name: 'X-Form Backend API',
    version: '2.0.0',
    description: 'Comprehensive API specifications for X-Form microservices platform',
    organization: 'X-Form Team',
    repository: 'https://github.com/Mir00r/X-Form-Backend',
    license: 'MIT'
  },

  // Specification Sources
  specs: {
    // Main API specification entry point
    main: './specs/openapi.yaml',
    
    // Service-specific specifications
    services: [
      {
        name: 'auth-service',
        path: './specs/services/auth-service.yaml',
        baseUrl: '/auth',
        version: '1.0.0',
        port: 3001,
        technology: 'Node.js + TypeScript + Express'
      },
      {
        name: 'form-service',
        path: './specs/services/form-service.yaml',
        baseUrl: '/forms',
        version: '1.0.0',
        port: 8001,
        technology: 'Go + Gin + GORM'
      },
      {
        name: 'response-service',
        path: './specs/services/response-service.yaml',
        baseUrl: '/responses',
        version: '1.0.0',
        port: 3002,
        technology: 'Node.js + TypeScript + Express'
      },
      {
        name: 'realtime-service',
        path: './specs/services/realtime-service.yaml',
        baseUrl: '/ws',
        version: '1.0.0',
        port: 8002,
        technology: 'Go + WebSockets + Redis'
      },
      {
        name: 'analytics-service',
        path: './specs/services/analytics-service.yaml',
        baseUrl: '/analytics',
        version: '1.0.0',
        port: 5001,
        technology: 'Python + FastAPI'
      }
    ],

    // Component specifications
    components: {
      schemas: './specs/components/schemas',
      parameters: './specs/components/parameters',
      responses: './specs/components/responses',
      examples: './specs/components/examples',
      headers: './specs/components/headers',
      securitySchemes: './specs/components/security'
    }
  },

  // Output Configuration
  output: {
    // Generated documentation directory
    docs: './specs/docs',
    
    // Bundled specifications
    bundle: './specs/dist',
    
    // Format preferences
    formats: ['yaml', 'json'],
    
    // Documentation themes
    theme: {
      name: 'redoc',
      options: {
        theme: {
          colors: {
            primary: {
              main: '#2563eb'
            }
          },
          typography: {
            fontSize: '14px',
            fontFamily: '"Inter", "Helvetica Neue", Arial, sans-serif'
          }
        },
        expandResponses: '200,201',
        hideDownloadButton: false,
        disableSearch: false,
        pathInMiddlePanel: true,
        menuToggle: true,
        scrollYOffset: 60
      }
    }
  },

  // Validation Rules
  validation: {
    // OpenAPI specification validation
    openapi: {
      version: '3.0.3',
      strictMode: true,
      validateExamples: true,
      validateSchemas: true
    },

    // Custom validation rules
    rules: {
      // Enforce consistent naming conventions
      naming: {
        operationIds: 'camelCase',
        parameters: 'camelCase',
        schemas: 'PascalCase',
        properties: 'camelCase'
      },

      // Require specific fields
      required: {
        operationId: true,
        summary: true,
        description: true,
        tags: true,
        responses: ['200', '400', '401', '500']
      },

      // Response standards
      responses: {
        successFormat: 'standardized',
        errorFormat: 'standardized',
        includeExamples: true
      },

      // Security requirements
      security: {
        enforceAuthentication: true,
        allowedSchemes: ['BearerAuth', 'ApiKeyAuth']
      }
    }
  },

  // Linting Configuration
  linting: {
    // Use Spectral for OpenAPI linting
    spectral: {
      extends: [
        '@stoplight/spectral-oai',
        '@stoplight/spectral-formats'
      ],
      rules: {
        // Custom rules for X-Form Backend
        'x-form-operation-summary': {
          description: 'Operation summary should be descriptive',
          given: '$.paths.*[get,post,put,patch,delete]',
          then: {
            field: 'summary',
            function: 'length',
            functionOptions: {
              min: 10,
              max: 80
            }
          }
        },
        'x-form-response-examples': {
          description: 'Responses should include examples',
          given: '$.paths.*[get,post,put,patch,delete].responses.*',
          then: {
            field: 'content.application/json.example',
            function: 'truthy'
          }
        },
        'x-form-tags-required': {
          description: 'Operations must have tags',
          given: '$.paths.*[get,post,put,patch,delete]',
          then: {
            field: 'tags',
            function: 'length',
            functionOptions: {
              min: 1
            }
          }
        }
      }
    }
  },

  // Development Server Configuration
  server: {
    port: 3000,
    host: 'localhost',
    cors: true,
    
    // API Gateway simulation
    gateway: {
      enabled: true,
      baseUrl: '/api/v1',
      services: {
        'auth-service': 'http://localhost:3001',
        'form-service': 'http://localhost:8001',
        'response-service': 'http://localhost:3002',
        'realtime-service': 'http://localhost:8002',
        'analytics-service': 'http://localhost:5001'
      }
    },

    // Mock server configuration
    mock: {
      enabled: true,
      dynamic: true,
      delay: {
        min: 100,
        max: 1000
      }
    }
  },

  // Testing Configuration
  testing: {
    // Contract testing
    contracts: {
      enabled: true,
      provider: 'pact',
      outputDir: './specs/tests/contracts'
    },

    // API testing
    api: {
      enabled: true,
      framework: 'newman',
      collections: './specs/tests/postman',
      environments: {
        development: './specs/tests/environments/dev.json',
        staging: './specs/tests/environments/staging.json',
        production: './specs/tests/environments/prod.json'
      }
    },

    // Performance testing
    performance: {
      enabled: true,
      tool: 'k6',
      scripts: './specs/tests/performance',
      thresholds: {
        'http_req_duration': ['p(95)<1000'],
        'http_req_failed': ['rate<0.01']
      }
    }
  },

  // CI/CD Integration
  cicd: {
    // GitHub Actions integration
    github: {
      workflows: {
        validate: '.github/workflows/spec-validation.yml',
        docs: '.github/workflows/docs-generation.yml',
        contracts: '.github/workflows/contract-testing.yml'
      }
    },

    // Quality gates
    quality: {
      coverage: {
        minimum: 80,
        include: ['schemas', 'paths', 'responses']
      },
      breaking: {
        detect: true,
        allowBreaking: false,
        exceptions: ['development']
      }
    }
  },

  // Plugin Configuration
  plugins: {
    // Code generation
    codegen: {
      enabled: true,
      generators: [
        {
          name: 'typescript-client',
          output: './generated/typescript-client',
          config: {
            npmName: '@x-form/api-client',
            withInterfaces: true,
            withSeparateModelsAndApi: true
          }
        },
        {
          name: 'go-client',
          output: './generated/go-client',
          config: {
            packageName: 'xformapi',
            withGoCodegen: true
          }
        },
        {
          name: 'python-client',
          output: './generated/python-client',
          config: {
            packageName: 'xform_api',
            projectName: 'x-form-api-client'
          }
        }
      ]
    },

    // Documentation enhancements
    docs: {
      enabled: true,
      generators: [
        {
          name: 'redoc',
          output: './specs/docs/redoc.html',
          theme: 'custom'
        },
        {
          name: 'swagger-ui',
          output: './specs/docs/swagger-ui',
          theme: 'custom'
        },
        {
          name: 'asyncapi',
          input: './specs/services/realtime-service.yaml',
          output: './specs/docs/asyncapi.html'
        }
      ]
    }
  },

  // Environment Configuration
  environments: {
    development: {
      baseUrl: 'http://localhost:8080/api/v1',
      services: {
        'auth-service': 'http://localhost:3001',
        'form-service': 'http://localhost:8001',
        'response-service': 'http://localhost:3002',
        'realtime-service': 'http://localhost:8002',
        'analytics-service': 'http://localhost:5001'
      }
    },
    staging: {
      baseUrl: 'https://api-staging.x-form.com/api/v1',
      services: {
        'auth-service': 'https://auth-staging.x-form.com',
        'form-service': 'https://forms-staging.x-form.com',
        'response-service': 'https://responses-staging.x-form.com',
        'realtime-service': 'https://realtime-staging.x-form.com',
        'analytics-service': 'https://analytics-staging.x-form.com'
      }
    },
    production: {
      baseUrl: 'https://api.x-form.com/api/v1',
      services: {
        'auth-service': 'https://auth.x-form.com',
        'form-service': 'https://forms.x-form.com',
        'response-service': 'https://responses.x-form.com',
        'realtime-service': 'https://realtime.x-form.com',
        'analytics-service': 'https://analytics.x-form.com'
      }
    }
  }
};
