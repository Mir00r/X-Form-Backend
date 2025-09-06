package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Mir00r/X-Form-Backend/services/api-gateway/docs" // Import for swagger
	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/handlers"
	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/middleware"
)

// @title           X-Form API Gateway
// @version         2.0.0
// @description     Comprehensive API Gateway for X-Form microservices architecture with advanced features including authentication, form management, response collection, real-time analytics, and comprehensive monitoring.
// @description
// @description     ## Features
// @description     - **Authentication & Authorization**: JWT-based authentication with role-based access control
// @description     - **Form Management**: Create, update, delete, and publish forms with advanced field types
// @description     - **Response Collection**: Collect and validate form responses with file uploads
// @description     - **Real-time Analytics**: Comprehensive analytics with insights and dashboard capabilities
// @description     - **Data Export**: Export responses in multiple formats (CSV, Excel, JSON, PDF)
// @description     - **Rate Limiting**: Advanced rate limiting with per-user and per-endpoint controls
// @description     - **Monitoring**: Health checks, metrics, and performance monitoring
// @description     - **Security**: CORS, security headers, request validation, and data protection
// @description
// @description     ## Authentication
// @description     This API uses JWT (JSON Web Tokens) for authentication. Include the token in the Authorization header:
// @description     ```
// @description     Authorization: Bearer <your-jwt-token>
// @description     ```
// @description
// @description     ## Rate Limiting
// @description     API requests are rate limited to ensure fair usage:
// @description     - Anonymous users: 100 requests per hour
// @description     - Authenticated users: 1000 requests per hour
// @description     - Premium users: 10000 requests per hour
// @description
// @description     ## Error Handling
// @description     The API returns consistent error responses with detailed information:
// @description     - 4xx errors: Client-side issues (validation, authentication, etc.)
// @description     - 5xx errors: Server-side issues
// @description
// @description     ## Pagination
// @description     List endpoints support pagination with the following parameters:
// @description     - `page`: Page number (default: 1)
// @description     - `limit`: Items per page (default: 10, max: 100)
// @description     - `sort`: Sort field
// @description     - `order`: Sort order (asc/desc)

// @termsOfService  https://x-form.com/terms
// @contact.name    X-Form API Support Team
// @contact.url     https://x-form.com/support
// @contact.email   api-support@x-form.com
// @license.name    MIT License
// @license.url     https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @schemes   http https
// @produce   application/json
// @consumes  application/json

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and your JWT token. Example: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API key for service-to-service communication

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-this-in-production"
		log.Println("Warning: Using default JWT secret. Set JWT_SECRET environment variable for production.")
	}

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	r := gin.Default()

	// Add enhanced middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.Recovery())

	// Metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Enhanced health check endpoints
	r.GET("/health", handlers.EnhancedHealthCheck)
	r.GET("/ready", handlers.EnhancedReady)
	r.GET("/live", handlers.EnhancedLive)

	// Swagger documentation with enhanced configuration
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Redirect root to documentation
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})

	// API v1 routes with enhanced handlers
	v1 := r.Group("/api/v1")
	{
		// Enhanced Authentication service routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handlers.EnhancedRegister)
			auth.POST("/login", handlers.EnhancedLogin)
			auth.POST("/logout", middleware.AuthRequired(jwtSecret), handlers.EnhancedLogout)
			auth.POST("/refresh", handlers.EnhancedRefreshToken)
			auth.GET("/profile", middleware.AuthRequired(jwtSecret), handlers.EnhancedGetProfile)
			auth.PUT("/profile", middleware.AuthRequired(jwtSecret), handlers.EnhancedUpdateProfile)
			auth.DELETE("/profile", middleware.AuthRequired(jwtSecret), handlers.EnhancedDeleteProfile)
		}

		// Enhanced Form service routes
		forms := v1.Group("/forms")
		{
			forms.GET("", middleware.OptionalAuth(jwtSecret), handlers.EnhancedListForms)
			forms.POST("", middleware.AuthRequired(jwtSecret), handlers.EnhancedCreateForm)
			forms.GET("/:id", middleware.OptionalAuth(jwtSecret), handlers.EnhancedGetForm)
			forms.PUT("/:id", middleware.AuthRequired(jwtSecret), handlers.UpdateForm)
			forms.DELETE("/:id", middleware.AuthRequired(jwtSecret), handlers.DeleteForm)
			forms.POST("/:id/publish", middleware.AuthRequired(jwtSecret), handlers.PublishForm)
			forms.POST("/:id/unpublish", middleware.AuthRequired(jwtSecret), handlers.UnpublishForm)
		}

		// Enhanced Response service routes
		responses := v1.Group("/responses")
		{
			responses.GET("", middleware.AuthRequired(jwtSecret), handlers.EnhancedListResponses)
			responses.POST("/:formId/submit", handlers.EnhancedSubmitResponse)
			responses.GET("/:id", middleware.AuthRequired(jwtSecret), handlers.EnhancedGetResponse)
			responses.PUT("/:id", middleware.AuthRequired(jwtSecret), handlers.UpdateResponse)
			responses.DELETE("/:id", middleware.AuthRequired(jwtSecret), handlers.DeleteResponse)
		}

		// Enhanced Analytics service routes (protected)
		analytics := v1.Group("/analytics", middleware.AuthRequired(jwtSecret))
		{
			analytics.GET("/forms/:formId", handlers.EnhancedGetFormAnalytics)
			analytics.GET("/responses/:responseId", handlers.GetResponseAnalytics)
			analytics.GET("/dashboard", handlers.EnhancedGetDashboard)
			analytics.POST("/export", handlers.EnhancedExportData)
			analytics.GET("/export/:job_id/status", handlers.EnhancedGetExportStatus)
		}
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 X-Form API Gateway v2.0.0 starting...")
	log.Printf("🌐 Server starting on port %s", port)
	log.Printf("📚 Enhanced Swagger documentation: http://localhost:%s/swagger/index.html", port)
	log.Printf("💚 Health check: http://localhost:%s/health", port)
	log.Printf("📊 Metrics: http://localhost:%s/metrics", port)
	log.Printf("⚡ Features: Enhanced auth, comprehensive analytics, advanced forms")

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
