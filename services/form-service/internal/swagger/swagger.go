// Package swagger contains OpenAPI documentation configuration for the Form Service
// Following microservices best practices for API documentation
package swagger

import (
	"github.com/swaggo/swag"
)

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0.0",
	Host:             "",
	BasePath:         "/api/v1",
	Schemes:          []string{"http", "https"},
	Title:            "Form Service API",
	Description:      "Comprehensive form management service following microservices best practices",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  swaggerTemplate,
}

// swaggerTemplate contains the complete OpenAPI specification
const swaggerTemplate = `{
  "openapi": "3.0.0",
  "info": {
    "title": "Form Service API",
    "description": "Comprehensive form management service built with Clean Architecture and following microservices best practices.\n\n**Features:**\n- Create, update, and manage forms\n- Dynamic question types and validation\n- Form publishing and response collection\n- Advanced filtering and search\n- Comprehensive monitoring and health checks\n\n**Architecture:**\n- Clean Architecture with SOLID principles\n- Microservices best practices\n- API versioning\n- Comprehensive DTOs\n- Input validation\n- Rate limiting\n- Circuit breakers\n- Structured logging",
    "version": "1.0.0",
    "contact": {
      "name": "Form Service Team",
      "email": "form-service@example.com",
      "url": "https://api.example.com/support"
    },
    "license": {
      "name": "MIT",
      "url": "https://opensource.org/licenses/MIT"
    }
  },
  "servers": [
    {
      "url": "https://api.example.com/v1",
      "description": "Production server"
    },
    {
      "url": "https://staging-api.example.com/v1",
      "description": "Staging server"
    },
    {
      "url": "http://localhost:8080/api/v1",
      "description": "Development server"
    }
  ],
  "security": [
    {
      "BearerAuth": []
    }
  ],
  "components": {
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT",
        "description": "JWT token for authentication. Format: Bearer {token}"
      }
    },
    "schemas": {
      "BaseResponse": {
        "type": "object",
        "properties": {
          "success": {
            "type": "boolean",
            "description": "Indicates if the request was successful"
          },
          "message": {
            "type": "string",
            "description": "Human-readable message"
          },
          "correlationId": {
            "type": "string",
            "format": "uuid",
            "description": "Unique correlation ID for request tracing"
          },
          "timestamp": {
            "type": "string",
            "format": "date-time",
            "description": "Response timestamp in ISO 8601 format"
          },
          "version": {
            "type": "string",
            "description": "API version"
          }
        },
        "required": ["success", "timestamp", "version"]
      },
      "SuccessResponse": {
        "allOf": [
          {"$ref": "#/components/schemas/BaseResponse"},
          {
            "type": "object",
            "properties": {
              "data": {
                "description": "Response data"
              }
            }
          }
        ]
      },
      "ErrorResponse": {
        "allOf": [
          {"$ref": "#/components/schemas/BaseResponse"},
          {
            "type": "object",
            "properties": {
              "error": {
                "$ref": "#/components/schemas/ErrorDetail"
              }
            }
          }
        ]
      },
      "ErrorDetail": {
        "type": "object",
        "properties": {
          "code": {
            "type": "string",
            "description": "Error code for programmatic handling"
          },
          "message": {
            "type": "string",
            "description": "Human-readable error message"
          },
          "details": {
            "description": "Additional error details"
          },
          "fields": {
            "type": "object",
            "additionalProperties": {
              "type": "string"
            },
            "description": "Field-specific error messages"
          },
          "metadata": {
            "type": "object",
            "description": "Additional metadata"
          },
          "requestId": {
            "type": "string",
            "description": "Request ID for debugging"
          },
          "path": {
            "type": "string",
            "description": "API path that caused the error"
          },
          "timestamp": {
            "type": "string",
            "format": "date-time",
            "description": "Error timestamp"
          }
        },
        "required": ["code", "message", "timestamp"]
      },
      "PaginatedResponse": {
        "allOf": [
          {"$ref": "#/components/schemas/BaseResponse"},
          {
            "type": "object",
            "properties": {
              "data": {
                "description": "Paginated data"
              },
              "pagination": {
                "$ref": "#/components/schemas/Pagination"
              }
            }
          }
        ]
      },
      "Pagination": {
        "type": "object",
        "properties": {
          "page": {
            "type": "integer",
            "minimum": 1,
            "description": "Current page number"
          },
          "pageSize": {
            "type": "integer",
            "minimum": 1,
            "maximum": 100,
            "description": "Number of items per page"
          },
          "total": {
            "type": "integer",
            "minimum": 0,
            "description": "Total number of items"
          },
          "totalPages": {
            "type": "integer",
            "minimum": 0,
            "description": "Total number of pages"
          },
          "hasNext": {
            "type": "boolean",
            "description": "Whether there are more pages"
          },
          "hasPrev": {
            "type": "boolean",
            "description": "Whether there are previous pages"
          }
        }
      },
      "CreateFormRequest": {
        "type": "object",
        "properties": {
          "title": {
            "type": "string",
            "minLength": 1,
            "maxLength": 255,
            "description": "Form title",
            "example": "Customer Feedback Form"
          },
          "description": {
            "type": "string",
            "maxLength": 1000,
            "description": "Form description",
            "example": "Please provide your feedback about our service"
          },
          "isAnonymous": {
            "type": "boolean",
            "description": "Whether responses are anonymous",
            "example": false
          },
          "isPublic": {
            "type": "boolean",
            "description": "Whether form is publicly accessible",
            "example": true
          },
          "allowMultiple": {
            "type": "boolean",
            "description": "Whether multiple responses are allowed",
            "example": false
          },
          "expiresAt": {
            "type": "string",
            "format": "date-time",
            "description": "Form expiration date",
            "example": "2024-12-31T23:59:59Z"
          },
          "settings": {
            "$ref": "#/components/schemas/FormSettings"
          },
          "questions": {
            "type": "array",
            "items": {
              "$ref": "#/components/schemas/CreateQuestionRequest"
            },
            "minItems": 1,
            "description": "Form questions"
          },
          "tags": {
            "type": "array",
            "items": {
              "type": "string",
              "maxLength": 50
            },
            "maxItems": 10,
            "description": "Form tags"
          },
          "category": {
            "type": "string",
            "maxLength": 100,
            "description": "Form category",
            "example": "feedback"
          }
        },
        "required": ["title", "questions"]
      },
      "FormSettings": {
        "type": "object",
        "properties": {
          "requireLogin": {
            "type": "boolean",
            "description": "Whether login is required"
          },
          "collectEmail": {
            "type": "boolean",
            "description": "Whether to collect respondent email"
          },
          "showProgressBar": {
            "type": "boolean",
            "description": "Whether to show progress bar"
          },
          "allowDrafts": {
            "type": "boolean",
            "description": "Whether to allow draft responses"
          },
          "notifyOnSubmission": {
            "type": "boolean",
            "description": "Whether to notify on form submission"
          },
          "customCss": {
            "type": "string",
            "maxLength": 5000,
            "description": "Custom CSS styles"
          },
          "redirectUrl": {
            "type": "string",
            "format": "uri",
            "maxLength": 500,
            "description": "Redirect URL after submission"
          },
          "thankYouMessage": {
            "type": "string",
            "maxLength": 1000,
            "description": "Thank you message"
          },
          "metadata": {
            "type": "object",
            "description": "Additional metadata"
          }
        }
      },
      "CreateQuestionRequest": {
        "type": "object",
        "properties": {
          "type": {
            "type": "string",
            "enum": ["text", "textarea", "number", "email", "date", "checkbox", "radio", "select", "file"],
            "description": "Question type"
          },
          "label": {
            "type": "string",
            "minLength": 1,
            "maxLength": 500,
            "description": "Question label"
          },
          "description": {
            "type": "string",
            "maxLength": 1000,
            "description": "Question description"
          },
          "required": {
            "type": "boolean",
            "description": "Whether question is required"
          },
          "order": {
            "type": "integer",
            "minimum": 0,
            "description": "Question order"
          },
          "options": {
            "type": "array",
            "items": {
              "$ref": "#/components/schemas/QuestionOption"
            },
            "description": "Question options for select/radio/checkbox"
          },
          "validation": {
            "$ref": "#/components/schemas/QuestionValidation"
          },
          "conditional": {
            "$ref": "#/components/schemas/ConditionalLogic"
          },
          "metadata": {
            "type": "object",
            "description": "Additional metadata"
          }
        },
        "required": ["type", "label", "order"]
      },
      "QuestionOption": {
        "type": "object",
        "properties": {
          "value": {
            "type": "string",
            "maxLength": 255,
            "description": "Option value"
          },
          "label": {
            "type": "string",
            "maxLength": 255,
            "description": "Option label"
          },
          "order": {
            "type": "integer",
            "minimum": 0,
            "description": "Option order"
          }
        },
        "required": ["value", "label", "order"]
      },
      "QuestionValidation": {
        "type": "object",
        "properties": {
          "minLength": {
            "type": "integer",
            "minimum": 0,
            "description": "Minimum length for text inputs"
          },
          "maxLength": {
            "type": "integer",
            "minimum": 1,
            "description": "Maximum length for text inputs"
          },
          "pattern": {
            "type": "string",
            "description": "Regex pattern for validation"
          },
          "minValue": {
            "type": "number",
            "description": "Minimum value for numeric inputs"
          },
          "maxValue": {
            "type": "number",
            "description": "Maximum value for numeric inputs"
          },
          "allowedTypes": {
            "type": "array",
            "items": {
              "type": "string"
            },
            "description": "Allowed file types"
          },
          "maxFileSize": {
            "type": "integer",
            "minimum": 1,
            "description": "Maximum file size in bytes"
          }
        }
      },
      "ConditionalLogic": {
        "type": "object",
        "properties": {
          "showIf": {
            "type": "array",
            "items": {
              "$ref": "#/components/schemas/Condition"
            },
            "description": "Conditions to show question"
          },
          "hideIf": {
            "type": "array",
            "items": {
              "$ref": "#/components/schemas/Condition"
            },
            "description": "Conditions to hide question"
          },
          "logic": {
            "type": "string",
            "enum": ["AND", "OR"],
            "description": "Logic operator for multiple conditions"
          }
        }
      },
      "Condition": {
        "type": "object",
        "properties": {
          "questionId": {
            "type": "string",
            "format": "uuid",
            "description": "Referenced question ID"
          },
          "operator": {
            "type": "string",
            "enum": ["equals", "not_equals", "contains", "not_contains", "greater_than", "less_than"],
            "description": "Comparison operator"
          },
          "value": {
            "description": "Comparison value"
          }
        },
        "required": ["questionId", "operator", "value"]
      },
      "FormResponse": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string",
            "format": "uuid",
            "description": "Form ID"
          },
          "title": {
            "type": "string",
            "description": "Form title"
          },
          "description": {
            "type": "string",
            "description": "Form description"
          },
          "status": {
            "type": "string",
            "enum": ["draft", "published", "closed", "archived"],
            "description": "Form status"
          },
          "isAnonymous": {
            "type": "boolean",
            "description": "Whether responses are anonymous"
          },
          "isPublic": {
            "type": "boolean",
            "description": "Whether form is publicly accessible"
          },
          "allowMultiple": {
            "type": "boolean",
            "description": "Whether multiple responses are allowed"
          },
          "createdBy": {
            "$ref": "#/components/schemas/UserInfo"
          },
          "settings": {
            "$ref": "#/components/schemas/FormSettings"
          },
          "questions": {
            "type": "array",
            "items": {
              "$ref": "#/components/schemas/QuestionResponse"
            }
          },
          "tags": {
            "type": "array",
            "items": {
              "type": "string"
            }
          },
          "category": {
            "type": "string",
            "description": "Form category"
          },
          "statistics": {
            "$ref": "#/components/schemas/FormStatistics"
          },
          "createdAt": {
            "type": "string",
            "format": "date-time"
          },
          "updatedAt": {
            "type": "string",
            "format": "date-time"
          },
          "publishedAt": {
            "type": "string",
            "format": "date-time"
          },
          "expiresAt": {
            "type": "string",
            "format": "date-time"
          }
        }
      },
      "UserInfo": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string",
            "format": "uuid"
          },
          "username": {
            "type": "string"
          },
          "email": {
            "type": "string",
            "format": "email"
          },
          "name": {
            "type": "string"
          },
          "avatar": {
            "type": "string",
            "format": "uri"
          }
        }
      },
      "QuestionResponse": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string",
            "format": "uuid"
          },
          "type": {
            "type": "string"
          },
          "label": {
            "type": "string"
          },
          "description": {
            "type": "string"
          },
          "required": {
            "type": "boolean"
          },
          "order": {
            "type": "integer"
          },
          "options": {
            "type": "array",
            "items": {
              "$ref": "#/components/schemas/QuestionOption"
            }
          },
          "validation": {
            "$ref": "#/components/schemas/QuestionValidation"
          },
          "conditional": {
            "$ref": "#/components/schemas/ConditionalLogic"
          },
          "metadata": {
            "type": "object"
          },
          "createdAt": {
            "type": "string",
            "format": "date-time"
          },
          "updatedAt": {
            "type": "string",
            "format": "date-time"
          }
        }
      },
      "FormStatistics": {
        "type": "object",
        "properties": {
          "totalResponses": {
            "type": "integer",
            "description": "Total number of responses"
          },
          "uniqueResponders": {
            "type": "integer",
            "description": "Number of unique responders"
          },
          "completionRate": {
            "type": "number",
            "minimum": 0,
            "maximum": 1,
            "description": "Completion rate (0-1)"
          },
          "averageTimeSeconds": {
            "type": "integer",
            "description": "Average completion time in seconds"
          },
          "lastResponse": {
            "type": "string",
            "format": "date-time",
            "description": "Timestamp of last response"
          },
          "responseRate": {
            "type": "number",
            "minimum": 0,
            "maximum": 1,
            "description": "Response rate (0-1)"
          }
        }
      },
      "HealthCheck": {
        "type": "object",
        "properties": {
          "status": {
            "type": "string",
            "enum": ["healthy", "degraded", "unhealthy"],
            "description": "Overall service health status"
          },
          "service": {
            "type": "string",
            "description": "Service name"
          },
          "version": {
            "type": "string",
            "description": "Service version"
          },
          "environment": {
            "type": "string",
            "description": "Environment name"
          },
          "timestamp": {
            "type": "string",
            "format": "date-time",
            "description": "Health check timestamp"
          },
          "uptime": {
            "type": "string",
            "description": "Service uptime"
          },
          "dependencies": {
            "$ref": "#/components/schemas/HealthDependencies"
          },
          "metrics": {
            "$ref": "#/components/schemas/HealthMetrics"
          },
          "features": {
            "type": "array",
            "items": {
              "type": "string"
            },
            "description": "Enabled features"
          }
        }
      },
      "HealthDependencies": {
        "type": "object",
        "properties": {
          "database": {
            "$ref": "#/components/schemas/DependencyStatus"
          },
          "authService": {
            "$ref": "#/components/schemas/DependencyStatus"
          },
          "emailService": {
            "$ref": "#/components/schemas/DependencyStatus"
          },
          "storageService": {
            "$ref": "#/components/schemas/DependencyStatus"
          }
        }
      },
      "DependencyStatus": {
        "type": "object",
        "properties": {
          "status": {
            "type": "string",
            "enum": ["healthy", "degraded", "unhealthy"],
            "description": "Dependency health status"
          },
          "responseTimeMs": {
            "type": "integer",
            "description": "Response time in milliseconds"
          },
          "lastChecked": {
            "type": "string",
            "format": "date-time",
            "description": "Last check timestamp"
          },
          "error": {
            "type": "string",
            "description": "Error message if unhealthy"
          }
        }
      },
      "HealthMetrics": {
        "type": "object",
        "properties": {
          "requestsPerSecond": {
            "type": "number",
            "description": "Current requests per second"
          },
          "averageResponseTimeMs": {
            "type": "integer",
            "description": "Average response time in milliseconds"
          },
          "errorRate": {
            "type": "number",
            "description": "Current error rate (0-1)"
          },
          "activeConnections": {
            "type": "integer",
            "description": "Number of active connections"
          },
          "memoryUsagePercent": {
            "type": "number",
            "description": "Memory usage percentage"
          },
          "cpuUsagePercent": {
            "type": "number",
            "description": "CPU usage percentage"
          },
          "diskUsagePercent": {
            "type": "number",
            "description": "Disk usage percentage"
          }
        }
      }
    },
    "responses": {
      "200": {
        "description": "Success",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/SuccessResponse"
            }
          }
        }
      },
      "201": {
        "description": "Created",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/SuccessResponse"
            }
          }
        }
      },
      "400": {
        "description": "Bad Request",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          }
        }
      },
      "401": {
        "description": "Unauthorized",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          }
        }
      },
      "403": {
        "description": "Forbidden",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          }
        }
      },
      "404": {
        "description": "Not Found",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          }
        }
      },
      "409": {
        "description": "Conflict",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          }
        }
      },
      "422": {
        "description": "Unprocessable Entity",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          }
        }
      },
      "429": {
        "description": "Too Many Requests",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          }
        }
      },
      "500": {
        "description": "Internal Server Error",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          }
        }
      },
      "503": {
        "description": "Service Unavailable",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          }
        }
      }
    }
  },
  "paths": {}
}`
