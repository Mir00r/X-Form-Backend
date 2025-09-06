package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
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

// @x-extension-openapi {"info":{"x-logo":{"url":"https://x-form.com/logo.png","altText":"X-Form Logo"}}}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get configuration from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-this-in-production"
		log.Println("Warning: Using default JWT secret. Set JWT_SECRET environment variable for production.")
	}

	// Set Gin mode based on environment
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug"
	}
	gin.SetMode(ginMode)

	// Initialize router with recovery middleware
	r := gin.New()

	// Add built-in middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Add custom middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.Recovery())

	// CORS configuration
	corsConfig := cors.Config{
		AllowOrigins:     []string{"*"}, // Configure appropriately for production
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-API-Key", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID", "X-Rate-Limit-Remaining", "X-Rate-Limit-Reset"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(corsConfig))

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// System health endpoints
	r.GET("/health", handlers.EnhancedHealthCheck)
	r.GET("/ready", handlers.EnhancedReady)
	r.GET("/live", handlers.EnhancedLive)
	r.GET("/metrics-detailed", middleware.OptionalAuth(jwtSecret), handlers.EnhancedMetrics)

	// Swagger documentation with custom configuration
	swaggerURL := ginSwagger.URL("/swagger/doc.json")
	swaggerConfig := ginSwagger.Config{
		URL:          "/swagger/doc.json",
		DeepLinking:  true,
		DocExpansion: "list",
		DomID:        "#swagger-ui",
		InstanceName: "swagger",
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, swaggerConfig, swaggerURL))

	// Redirect root to Swagger documentation
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})

	// API Documentation endpoints
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Authentication routes
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

		// Form management routes
		forms := v1.Group("/forms")
		{
			forms.GET("", middleware.OptionalAuth(jwtSecret), handlers.EnhancedListForms)
			forms.POST("", middleware.AuthRequired(jwtSecret), handlers.EnhancedCreateForm)
			forms.GET("/:id", middleware.OptionalAuth(jwtSecret), handlers.EnhancedGetForm)
			forms.PUT("/:id", middleware.AuthRequired(jwtSecret), handlers.UpdateForm)
			forms.DELETE("/:id", middleware.AuthRequired(jwtSecret), handlers.DeleteForm)
			forms.POST("/:id/publish", middleware.AuthRequired(jwtSecret), handlers.PublishForm)
			forms.POST("/:id/unpublish", middleware.AuthRequired(jwtSecret), handlers.UnpublishForm)
			forms.POST("/:id/duplicate", middleware.AuthRequired(jwtSecret), handlers.DuplicateForm)
			forms.GET("/:id/analytics", middleware.AuthRequired(jwtSecret), handlers.GetFormAnalytics)
			forms.GET("/:id/export", middleware.AuthRequired(jwtSecret), handlers.ExportFormData)
		}

		// Response collection routes
		responses := v1.Group("/responses")
		{
			responses.GET("", middleware.AuthRequired(jwtSecret), handlers.EnhancedListResponses)
			responses.POST("/:formId/submit", handlers.EnhancedSubmitResponse)
			responses.GET("/:id", middleware.AuthRequired(jwtSecret), handlers.EnhancedGetResponse)
			responses.PUT("/:id", middleware.AuthRequired(jwtSecret), handlers.UpdateResponse)
			responses.DELETE("/:id", middleware.AuthRequired(jwtSecret), handlers.DeleteResponse)
			responses.POST("/:id/validate", middleware.AuthRequired(jwtSecret), handlers.ValidateResponse)
			responses.GET("/:id/files", middleware.AuthRequired(jwtSecret), handlers.GetResponseFiles)
		}

		// Analytics routes (all protected)
		analytics := v1.Group("/analytics", middleware.AuthRequired(jwtSecret))
		{
			analytics.GET("/forms/:formId", handlers.EnhancedGetFormAnalytics)
			analytics.GET("/responses/:responseId", handlers.GetResponseAnalytics)
			analytics.GET("/dashboard", handlers.EnhancedGetDashboard)
			analytics.POST("/dashboard", handlers.CreateDashboard)
			analytics.PUT("/dashboard/:id", handlers.UpdateDashboard)
			analytics.DELETE("/dashboard/:id", handlers.DeleteDashboard)
			analytics.GET("/insights", handlers.GetInsights)
			analytics.POST("/export", handlers.EnhancedExportData)
			analytics.GET("/export/:job_id/status", handlers.EnhancedGetExportStatus)
			analytics.GET("/export/:job_id/download", handlers.DownloadExport)
		}

		// Real-time collaboration routes
		collaboration := v1.Group("/collaboration", middleware.AuthRequired(jwtSecret))
		{
			collaboration.GET("/rooms", handlers.ListCollaborationRooms)
			collaboration.POST("/rooms", handlers.CreateCollaborationRoom)
			collaboration.GET("/rooms/:id", handlers.GetCollaborationRoom)
			collaboration.POST("/rooms/:id/join", handlers.JoinCollaborationRoom)
			collaboration.POST("/rooms/:id/leave", handlers.LeaveCollaborationRoom)
			collaboration.GET("/rooms/:id/users", handlers.GetRoomUsers)
			collaboration.POST("/rooms/:id/messages", handlers.SendMessage)
			collaboration.GET("/rooms/:id/messages", handlers.GetMessages)
		}

		// File management routes
		files := v1.Group("/files", middleware.AuthRequired(jwtSecret))
		{
			files.POST("/upload", handlers.UploadFile)
			files.GET("/:id", handlers.GetFile)
			files.DELETE("/:id", handlers.DeleteFile)
			files.GET("/:id/download", handlers.DownloadFile)
			files.POST("/:id/virus-scan", handlers.VirusScanFile)
		}

		// User management routes (admin only)
		users := v1.Group("/users", middleware.AuthRequired(jwtSecret), middleware.AdminRequired())
		{
			users.GET("", handlers.ListUsers)
			users.GET("/:id", handlers.GetUser)
			users.PUT("/:id", handlers.UpdateUser)
			users.DELETE("/:id", handlers.DeleteUser)
			users.POST("/:id/suspend", handlers.SuspendUser)
			users.POST("/:id/activate", handlers.ActivateUser)
			users.GET("/:id/activity", handlers.GetUserActivity)
		}

		// Admin routes
		admin := v1.Group("/admin", middleware.AuthRequired(jwtSecret), middleware.AdminRequired())
		{
			admin.GET("/stats", handlers.GetSystemStats)
			admin.GET("/logs", handlers.GetSystemLogs)
			admin.POST("/maintenance", handlers.SetMaintenanceMode)
			admin.GET("/config", handlers.GetSystemConfig)
			admin.PUT("/config", handlers.UpdateSystemConfig)
			admin.POST("/cache/clear", handlers.ClearCache)
			admin.GET("/performance", handlers.GetPerformanceMetrics)
		}
	}

	// API v2 routes (for future expansion)
	v2 := r.Group("/api/v2")
	{
		v2.GET("/health", handlers.EnhancedHealthCheck)
		// Future API versions can be added here
	}

	// Get server configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	// Start server with configuration
	serverAddr := host + ":" + port

	log.Printf("🚀 X-Form API Gateway starting...")
	log.Printf("📱 Environment: %s", ginMode)
	log.Printf("🌐 Server Address: http://%s", serverAddr)
	log.Printf("📚 API Documentation: http://%s/swagger/index.html", serverAddr)
	log.Printf("💚 Health Check: http://%s/health", serverAddr)
	log.Printf("📊 Metrics: http://%s/metrics", serverAddr)
	log.Printf("⚡ Version: 2.0.0")

	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
