package middleware

import (
	"github.com/gin-gonic/gin"
)

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

func AuthMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// TODO: Implement JWT auth middleware
		c.Next()
	})
}

func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// TODO: Implement JWT auth validation
		c.Next()
	})
}

func OptionalAuth(jwtSecret string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// TODO: Implement optional JWT auth validation
		c.Next()
	})
}
