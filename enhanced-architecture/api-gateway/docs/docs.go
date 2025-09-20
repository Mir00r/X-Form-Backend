package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "API Support",
            "email": "support@x-form.com"
        },
        "license": {
            "name": "MIT",
            "url": "https://opensource.org/licenses/MIT"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/health": {
            "get": {
                "description": "Check the health status of the API Gateway",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Health"
                ],
                "summary": "Health Check",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/HealthResponse"
                        }
                    }
                }
            }
        },
        "/metrics": {
            "get": {
                "description": "Get Prometheus metrics",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "Metrics"
                ],
                "summary": "Get Metrics",
                "responses": {
                    "200": {
                        "description": "Metrics data"
                    }
                }
            }
        }
    },
    "definitions": {
        "HealthResponse": {
            "type": "object",
            "properties": {
                "status": {
                    "type": "string",
                    "example": "healthy"
                },
                "timestamp": {
                    "type": "string",
                    "example": "2023-01-01T00:00:00Z"
                },
                "services": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                }
            }
        }
    }
}`

var doc = &swag.Spec{
	Version:          "1.0.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{"http", "https"},
	Title:            "X-Form Enhanced API Gateway",
	Description:      "Enhanced API Gateway for X-Form Backend with authentication, rate limiting, and service discovery",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(doc.InfoInstanceName, doc)
}
