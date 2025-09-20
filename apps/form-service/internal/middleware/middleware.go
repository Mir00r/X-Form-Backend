package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// =============================================================================
// Simple Rate Limiting (without external dependencies)
// =============================================================================

type simpleRateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

var globalRateLimiter = &simpleRateLimiter{
	requests: make(map[string][]time.Time),
	limit:    100,         // 100 requests
	window:   time.Minute, // per minute
}

// RateLimiting provides simple rate limiting functionality
func RateLimiting() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := c.ClientIP()

		globalRateLimiter.mutex.Lock()
		defer globalRateLimiter.mutex.Unlock()

		now := time.Now()

		// Clean old requests
		if requests, exists := globalRateLimiter.requests[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < globalRateLimiter.window {
					validRequests = append(validRequests, reqTime)
				}
			}
			globalRateLimiter.requests[clientIP] = validRequests
		}

		// Check if limit exceeded
		if len(globalRateLimiter.requests[clientIP]) >= globalRateLimiter.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "Too many requests. Please try again later.",
				},
				"timestamp": time.Now(),
			})
			c.Abort()
			return
		}

		// Add current request
		globalRateLimiter.requests[clientIP] = append(globalRateLimiter.requests[clientIP], now)

		c.Next()
	})
}

// =============================================================================
// CORS and Security
// =============================================================================

func CorsMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

func CORS() gin.HandlerFunc {
	return CorsMiddleware()
}

func Security() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Next()
	})
}

// =============================================================================
// Authentication
// =============================================================================

func AuthMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// TODO: Implement JWT auth middleware
		c.Set("userID", "demo-user-123")
		c.Set("authenticated", true)
		c.Next()
	})
}

func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Simplified auth for demo - sets a demo user
		c.Set("userID", "demo-user-123")
		c.Set("authenticated", true)
		c.Next()
	})
}

func OptionalAuth(jwtSecret string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Optional auth - may or may not set user
		c.Set("userID", "demo-user-123")
		c.Set("authenticated", true)
		c.Next()
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

// GetUserID retrieves user ID from request context
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("userID"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// IsAuthenticated checks if request is authenticated
func IsAuthenticated(c *gin.Context) bool {
	if auth, exists := c.Get("authenticated"); exists {
		if authenticated, ok := auth.(bool); ok {
			return authenticated
		}
	}
	return false
}
