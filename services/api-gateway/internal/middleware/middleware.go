package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

// Prometheus metrics
var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "api_gateway_http_duration_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"method", "route", "status_code"})

	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_gateway_http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"method", "route", "status_code"})
)

// CORS middleware with comprehensive settings
func CORS() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// List of allowed origins - should be configurable
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"https://xform.dev",
			"https://app.xform.dev",
			"https://api.xform.dev",
		}

		// Check if origin is allowed
		isAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Request-ID")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
}

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("RequestID", requestID)
		c.Next()
	})
}

// RequestLogger middleware logs all requests with structured logging
func RequestLogger() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get request ID
		requestID, _ := c.Get("RequestID")

		// Build log entry
		entry := logrus.WithFields(logrus.Fields{
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       path,
			"query":      raw,
			"status":     c.Writer.Status(),
			"latency":    latency,
			"latency_ms": float64(latency.Nanoseconds()) / 1e6,
			"user_agent": c.Request.UserAgent(),
			"client_ip":  c.ClientIP(),
			"size":       c.Writer.Size(),
		})

		// Add user context if available
		if userID, exists := c.Get("UserID"); exists {
			entry = entry.WithField("user_id", userID)
		}

		// Log based on status code
		switch {
		case c.Writer.Status() >= 500:
			entry.Error("Server error")
		case c.Writer.Status() >= 400:
			entry.Warn("Client error")
		default:
			entry.Info("Request completed")
		}
	})
}

// Metrics middleware records Prometheus metrics
func Metrics() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		method := c.Request.Method
		route := c.FullPath()
		statusCode := string(rune(c.Writer.Status()))

		httpDuration.WithLabelValues(method, route, statusCode).Observe(duration)
		httpRequests.WithLabelValues(method, route, statusCode).Inc()
	})
}

// RateLimit middleware (placeholder - would integrate with Redis)
func RateLimit() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// TODO: Implement rate limiting with Redis
		// For now, just pass through
		c.Next()
	})
}

// AuthRequired middleware validates JWT tokens
func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authorization header format must be Bearer {token}",
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Set user context
			if userID, exists := claims["user_id"]; exists {
				c.Set("UserID", userID)
			}
			if sub, exists := claims["sub"]; exists {
				c.Set("Subject", sub)
			}
			if roles, exists := claims["roles"]; exists {
				c.Set("Roles", roles)
			}
			if email, exists := claims["email"]; exists {
				c.Set("Email", email)
			}

			// Set full claims for downstream services
			c.Set("Claims", claims)
		}

		c.Next()
	})
}

// OptionalAuth middleware validates JWT tokens but doesn't require them
func OptionalAuth(jwtSecret string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header, continue without authentication
			c.Next()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// Invalid format, continue without authentication
			c.Next()
			return
		}

		tokenString := tokenParts[1]

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err == nil && token.Valid {
			// Valid token - set user context
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if userID, exists := claims["user_id"]; exists {
					c.Set("UserID", userID)
				}
				if sub, exists := claims["sub"]; exists {
					c.Set("Subject", sub)
				}
				if roles, exists := claims["roles"]; exists {
					c.Set("Roles", roles)
				}
				c.Set("Claims", claims)
			}
		}

		c.Next()
	})
}

// SecurityHeaders middleware adds security headers
func SecurityHeaders() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		c.Next()
	})
}

// Recovery middleware with custom error handling
func Recovery() gin.HandlerFunc {
	return gin.RecoveryWithWriter(gin.DefaultErrorWriter, func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			logrus.WithFields(logrus.Fields{
				"error":      err,
				"request_id": c.GetString("RequestID"),
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
			}).Error("Panic recovered")
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_server_error",
			"message": "Internal server error occurred",
		})
	})
}
