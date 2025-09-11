// Package middleware implements the 7-step API Gateway process from the architecture diagram
// Following industry best practices for middleware design patterns
package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Mir00r/X-Form-Backend/enhanced-architecture/api-gateway/internal/config"
	"github.com/Mir00r/X-Form-Backend/enhanced-architecture/api-gateway/pkg/logger"
	"github.com/Mir00r/X-Form-Backend/enhanced-architecture/api-gateway/pkg/metrics"
)

// MiddlewareError represents a middleware-specific error
type MiddlewareError struct {
	Code    int
	Message string
	Details string
}

func (e *MiddlewareError) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.Details)
}

// HandlerFunc represents an HTTP handler function
type HandlerFunc func(http.ResponseWriter, *http.Request)

// Middleware represents a middleware function
type Middleware func(HandlerFunc) HandlerFunc

// Chain represents a middleware chain
type Chain struct {
	middlewares []Middleware
}

// NewChain creates a new middleware chain
func NewChain(middlewares ...Middleware) *Chain {
	return &Chain{middlewares: middlewares}
}

// Then applies the middleware chain to a handler
func (c *Chain) Then(handler HandlerFunc) HandlerFunc {
	if len(c.middlewares) == 0 {
		return handler
	}

	h := handler
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}
	return h
}

// Append adds middleware to the chain
func (c *Chain) Append(middlewares ...Middleware) *Chain {
	newChain := &Chain{
		middlewares: make([]Middleware, len(c.middlewares)+len(middlewares)),
	}
	copy(newChain.middlewares, c.middlewares)
	copy(newChain.middlewares[len(c.middlewares):], middlewares)
	return newChain
}

// RequestID middleware adds a unique request ID to each request
func RequestID() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Check if request ID already exists
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				// Generate new UUID for request tracking
				requestID = generateUUID()
			}

			// Add request ID to context and response header
			ctx := context.WithValue(r.Context(), "request_id", requestID)
			w.Header().Set("X-Request-ID", requestID)

			next(w, r.WithContext(ctx))
		}
	}
}

// StructuredLogger middleware provides structured logging with request context
func StructuredLogger(logger logger.Logger) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get request information
			requestID := getRequestID(r.Context())
			path := r.URL.Path
			method := r.Method
			clientIP := getClientIPSimple(r)
			userAgent := r.UserAgent()

			// Add request context to logger
			loggerWithContext := logger.WithFields(map[string]interface{}{
				"request_id": requestID,
				"method":     method,
				"path":       path,
				"client_ip":  clientIP,
				"user_agent": userAgent,
			})

			// Create wrapped response writer
			ww := &wrappedResponseWriter{ResponseWriter: w, statusCode: 200}

			// Process request
			next(ww, r)

			// Log request completion
			duration := time.Since(start)
			statusCode := ww.statusCode

			logLevel := "info"
			if statusCode >= 400 {
				logLevel = "error"
			} else if statusCode >= 300 {
				logLevel = "warn"
			}

			loggerWithContext.WithFields(map[string]interface{}{
				"status_code": statusCode,
				"latency_ms":  duration.Milliseconds(),
				"body_size":   ww.bytesWritten,
			}).Log(logLevel, "Request completed")
		}
	}
}

// Metrics middleware collects request metrics for monitoring
func Metrics(collector *metrics.Collector) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			collector.IncrementRequestsInFlight()
			defer collector.DecrementRequestsInFlight()

			// Create wrapped response writer
			ww := &wrappedResponseWriter{ResponseWriter: w, statusCode: 200}

			// Process request
			next(ww, r)

			// Record metrics
			duration := time.Since(start)
			collector.RecordHTTPRequest(r.Method, r.URL.Path, ww.statusCode, duration, ww.bytesWritten)
		}
	}
}

// Recovery middleware handles panics gracefully
func Recovery(logger logger.Logger) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					requestID := getRequestID(r.Context())
					logger.WithFields(map[string]interface{}{
						"request_id": requestID,
						"method":     r.Method,
						"path":       r.URL.Path,
						"panic":      err,
					}).Error("Panic recovered")

					// Return internal server error
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next(w, r)
		}
	}
}

// CORS middleware handles Cross-Origin Resource Sharing
func CORS(corsConfig config.CORSConfig) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if CORS is enabled
			if !corsConfig.Enabled {
				next(w, r)
				return
			}

			// Check if origin is allowed
			if !isOriginAllowed(origin, corsConfig.AllowedOrigins) {
				http.Error(w, "Origin not allowed", http.StatusForbidden)
				return
			}

			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(corsConfig.AllowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(corsConfig.AllowedHeaders, ", "))

			if len(corsConfig.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(corsConfig.ExposedHeaders, ", "))
			}

			if corsConfig.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if corsConfig.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(corsConfig.MaxAge))
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next(w, r)
		}
	}
}

// SecurityHeaders middleware adds security headers
func SecurityHeaders() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Content Security Policy
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

			// Frame Options
			w.Header().Set("X-Frame-Options", "DENY")

			// Content Type Options
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// XSS Protection
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Referrer Policy
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// HSTS (only for HTTPS)
			if r.TLS != nil {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			}

			// Remove server information
			w.Header().Set("Server", "")

			next(w, r)
		}
	}
}

// Step 1: Parameter Validation Middleware
func ParameterValidation(validationConfig config.ValidationConfig) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Skip validation for certain methods and paths
			if shouldSkipValidation(r.Method, r.URL.Path) {
				next(w, r)
				return
			}

			// Basic validation implementation
			if err := validateRequest(r, validationConfig); err != nil {
				http.Error(w, fmt.Sprintf("Validation failed: %s", err.Error()), http.StatusBadRequest)
				return
			}

			next(w, r)
		}
	}
}

// Step 2: Whitelist Validation Middleware
func WhitelistValidation(whitelistConfig config.WhitelistConfig) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Skip if whitelist is disabled
			if !whitelistConfig.Enabled {
				next(w, r)
				return
			}

			clientIP := getClientIPFromConfig(r, whitelistConfig)

			// Check if IP is blocked
			if isIPBlocked(clientIP, whitelistConfig.BlockedIPs) {
				http.Error(w, "IP address is blocked", http.StatusForbidden)
				return
			}

			// Check if IP is in whitelist (if whitelist is configured)
			if len(whitelistConfig.AllowedIPs) > 0 && !isIPAllowed(clientIP, whitelistConfig.AllowedIPs) {
				http.Error(w, "IP address not in whitelist", http.StatusForbidden)
				return
			}

			// Add client IP to context
			ctx := context.WithValue(r.Context(), "client_ip", clientIP)
			next(w, r.WithContext(ctx))
		}
	}
}

// Step 3: Authentication Middleware
func Authentication(authConfig config.JWTConfig) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for public endpoints
			if isPublicEndpoint(r.URL.Path) {
				next(w, r)
				return
			}

			// Extract token from request
			token := extractToken(r)
			if token == "" {
				http.Error(w, "Authentication token is required", http.StatusUnauthorized)
				return
			}

			// Basic token validation (simplified)
			if !validateToken(token, authConfig) {
				http.Error(w, "Invalid authentication token", http.StatusUnauthorized)
				return
			}

			// Add user information to context (simplified)
			ctx := context.WithValue(r.Context(), "user_authenticated", true)
			next(w, r.WithContext(ctx))
		}
	}
}

// Step 4: Rate Limiting Middleware
func RateLimit(rateLimitConfig config.RateLimitConfig) Middleware {
	// Simple in-memory rate limiter (in production, use Redis)
	rateLimiters := make(map[string]*simpleRateLimiter)

	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Skip if rate limiting is disabled
			if !rateLimitConfig.Enabled {
				next(w, r)
				return
			}

			// Get client identifier
			clientID := getClientIdentifier(r)

			// Get or create rate limiter for client
			limiter, exists := rateLimiters[clientID]
			if !exists {
				limiter = newSimpleRateLimiter(rateLimitConfig.RPS, rateLimitConfig.Burst)
				rateLimiters[clientID] = limiter
			}

			// Check if request is allowed
			if !limiter.Allow() {
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rateLimitConfig.RPS))
				w.Header().Set("X-RateLimit-Remaining", "0")
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rateLimitConfig.RPS))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(rateLimitConfig.Burst-1))

			next(w, r)
		}
	}
}

// Timeout middleware implements request timeout
func Timeout(timeout time.Duration) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Create a channel to signal completion
			done := make(chan struct{})

			// Process request in goroutine
			go func() {
				defer close(done)
				next(w, r.WithContext(ctx))
			}()

			// Wait for completion or timeout
			select {
			case <-done:
				// Request completed normally
				return
			case <-ctx.Done():
				// Request timed out
				http.Error(w, "Request timeout", http.StatusRequestTimeout)
				return
			}
		}
	}
}

// Helper functions

// wrappedResponseWriter wraps http.ResponseWriter to capture response information
type wrappedResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (w *wrappedResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *wrappedResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += int64(n)
	return n, err
}

// Simple rate limiter implementation
type simpleRateLimiter struct {
	maxRequests int
	windowSize  time.Duration
	requests    []time.Time
}

func newSimpleRateLimiter(maxRequests, burst int) *simpleRateLimiter {
	return &simpleRateLimiter{
		maxRequests: maxRequests,
		windowSize:  time.Minute,
		requests:    make([]time.Time, 0),
	}
}

func (rl *simpleRateLimiter) Allow() bool {
	now := time.Now()

	// Remove old requests outside the window
	cutoff := now.Add(-rl.windowSize)
	var validRequests []time.Time
	for _, req := range rl.requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}
	rl.requests = validRequests

	// Check if we can allow this request
	if len(rl.requests) >= rl.maxRequests {
		return false
	}

	// Add this request
	rl.requests = append(rl.requests, now)
	return true
}

// Utility functions

func generateUUID() string {
	// Simple UUID generation (in production, use proper UUID library)
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func getRequestID(ctx context.Context) string {
	if id, ok := ctx.Value("request_id").(string); ok {
		return id
	}
	return ""
}

func getClientIPSimple(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Use RemoteAddr
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}

	return r.RemoteAddr
}

func getClientIPFromConfig(r *http.Request, config config.WhitelistConfig) string {
	if config.TrustProxy && config.ProxyHeader != "" {
		if ip := r.Header.Get(config.ProxyHeader); ip != "" {
			return strings.Split(ip, ",")[0]
		}
	}
	return getClientIPSimple(r)
}

func shouldSkipValidation(method, path string) bool {
	skipPaths := []string{
		"/health",
		"/metrics",
		"/swagger",
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	return method == http.MethodGet || method == http.MethodOptions
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

func isIPBlocked(ip string, blockedIPs []string) bool {
	for _, blocked := range blockedIPs {
		if matchIP(ip, blocked) {
			return true
		}
	}
	return false
}

func isIPAllowed(ip string, allowedIPs []string) bool {
	for _, allowed := range allowedIPs {
		if matchIP(ip, allowed) {
			return true
		}
	}
	return false
}

func matchIP(ip, pattern string) bool {
	if strings.Contains(pattern, "/") {
		// CIDR notation
		_, network, err := net.ParseCIDR(pattern)
		if err != nil {
			return false
		}
		ipAddr := net.ParseIP(ip)
		return network.Contains(ipAddr)
	}
	return ip == pattern
}

func isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/health",
		"/metrics",
		"/api/v1/auth/login",
		"/api/v1/auth/signup",
		"/public/",
		"/swagger/",
	}

	for _, publicPath := range publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
	}
	return false
}

func extractToken(r *http.Request) string {
	// Try Authorization header first
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// Try query parameter
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}

	// Try cookie
	if cookie, err := r.Cookie("token"); err == nil {
		return cookie.Value
	}

	return ""
}

func getClientIdentifier(r *http.Request) string {
	// Simple client identification
	if userID := r.Context().Value("user_id"); userID != nil {
		return fmt.Sprintf("user:%v", userID)
	}
	return fmt.Sprintf("ip:%s", getClientIPSimple(r))
}

// Helper functions for simplified middleware implementation

// validateRequest performs basic request validation
func validateRequest(r *http.Request, config config.ValidationConfig) error {
	// Skip validation if disabled
	if !config.Enabled {
		return nil
	}

	// Basic request size validation (using a reasonable default since MaxBodySize isn't in config)
	maxSize := int64(10 << 20) // 10MB default
	if r.ContentLength > maxSize {
		return fmt.Errorf("request size exceeds maximum allowed: %d bytes", maxSize)
	}

	return nil
}

// validateToken performs basic token validation
func validateToken(token string, config config.JWTConfig) bool {
	// Basic token validation (simplified for this implementation)
	if len(token) < 10 {
		return false
	}

	// In a real implementation, you would parse and validate the JWT
	// For now, just check if it's not empty and has reasonable length
	return len(token) > 0 && len(token) < 2048
}
