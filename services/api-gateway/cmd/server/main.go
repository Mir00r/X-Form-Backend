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

// @title           API Gateway
// @version         1.0
// @description     A comprehensive API Gateway for X-Form microservices
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

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

	// Add middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.Recovery())

	// Metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check endpoint
	r.GET("/health", handlers.HealthCheck)

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Auth service routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.POST("/logout", middleware.AuthRequired(jwtSecret), handlers.Logout)
			auth.POST("/refresh", handlers.RefreshToken)
			auth.GET("/profile", middleware.AuthRequired(jwtSecret), handlers.GetProfile)
			auth.PUT("/profile", middleware.AuthRequired(jwtSecret), handlers.UpdateProfile)
			auth.DELETE("/profile", middleware.AuthRequired(jwtSecret), handlers.DeleteProfile)
		}

		// Form service routes
		forms := v1.Group("/forms")
		{
			forms.GET("", middleware.OptionalAuth(jwtSecret), handlers.ListForms)
			forms.POST("", middleware.AuthRequired(jwtSecret), handlers.CreateForm)
			forms.GET("/:id", middleware.OptionalAuth(jwtSecret), handlers.GetForm)
			forms.PUT("/:id", middleware.AuthRequired(jwtSecret), handlers.UpdateForm)
			forms.DELETE("/:id", middleware.AuthRequired(jwtSecret), handlers.DeleteForm)
			forms.POST("/:id/publish", middleware.AuthRequired(jwtSecret), handlers.PublishForm)
			forms.POST("/:id/unpublish", middleware.AuthRequired(jwtSecret), handlers.UnpublishForm)
		}

		// Response service routes
		responses := v1.Group("/responses")
		{
			responses.GET("", middleware.AuthRequired(jwtSecret), handlers.ListResponses)
			responses.POST("/:formId/submit", handlers.SubmitResponse)
			responses.GET("/:id", middleware.AuthRequired(jwtSecret), handlers.GetResponse)
			responses.PUT("/:id", middleware.AuthRequired(jwtSecret), handlers.UpdateResponse)
			responses.DELETE("/:id", middleware.AuthRequired(jwtSecret), handlers.DeleteResponse)
		}

		// Analytics service routes (protected)
		analytics := v1.Group("/analytics", middleware.AuthRequired(jwtSecret))
		{
			analytics.GET("/forms/:formId", handlers.GetFormAnalytics)
			analytics.GET("/responses/:responseId", handlers.GetResponseAnalytics)
			analytics.GET("/dashboard", handlers.GetDashboard)
		}
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API Gateway starting on port %s", port)
	log.Printf("Swagger documentation available at: http://localhost:%s/swagger/index.html", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
