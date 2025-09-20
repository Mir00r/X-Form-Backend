// Package middleware implements the 7-step API Gateway process from the architecture diagram
// Following industry best practices for middleware design patterns
package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Mir00r/X-Form-Backend/enhanced-architecture/api-gateway/internal/config"
	"github.com/Mir00r/X-Form-Backend/enhanced-architecture/api-gateway/pkg/logger"
	"github.com/Mir00r/X-Form-Backend/enhanced-architecture/api-gateway/pkg/metrics"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// Context key types to avoid collisions
type contextKey string

const (
	DiscoveredServiceKey contextKey = "discovered_service"
	ServiceNameKey       contextKey = "service_name"
	ServiceHostKey       contextKey = "service_host"
	ServicePortKey       contextKey = "service_port"
	ServiceInstanceIDKey contextKey = "service_instance_id"
	RequestIDKey         contextKey = "request_id"
	ClientIPKey          contextKey = "client_ip"
	UserAuthenticatedKey contextKey = "user_authenticated"
	UserIDKey            contextKey = "user_id"
	UserRoleKey          contextKey = "user_role"
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

// ServiceRegistry represents an enhanced service registry for service discovery
type ServiceRegistry struct {
	services      map[string][]*ServiceInstance
	serviceHealth map[string]*ServiceHealth
	config        *ServiceRegistryConfig
	logger        logger.Logger
	metrics       *metrics.Collector
	loadBalancer  *LoadBalancer
	healthChecker *ServiceHealthChecker
	mutex         sync.RWMutex
}

// ServiceInstance represents a service instance
type ServiceInstance struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Host      string            `json:"host"`
	Port      int               `json:"port"`
	Health    string            `json:"health"`
	LastCheck time.Time         `json:"last_check"`
	Metadata  map[string]string `json:"metadata"`
	Weight    int               `json:"weight"`
	Tags      []string          `json:"tags"`
}

// ServiceHealth tracks health information for a service
type ServiceHealth struct {
	HealthyInstances   int
	UnhealthyInstances int
	LastHealthCheck    time.Time
	ResponseTime       time.Duration
	FailureCount       int
	SuccessCount       int
}

// ServiceRegistryConfig holds configuration for service registry
type ServiceRegistryConfig struct {
	RealTimeHealthCheck bool
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
	MaxRetries          int
}

// LoadBalancer provides load balancing functionality
type LoadBalancer struct {
	strategy string
	counters map[string]int
	mutex    sync.Mutex
}

// ServiceHealthChecker performs health checks
type ServiceHealthChecker struct {
	client *http.Client
	config *ServiceRegistryConfig
}

// NewServiceRegistry creates a new enhanced service registry
func NewServiceRegistry(logger logger.Logger, metrics *metrics.Collector) *ServiceRegistry {
	config := &ServiceRegistryConfig{
		RealTimeHealthCheck: true,
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		MaxRetries:          3,
	}

	healthChecker := &ServiceHealthChecker{
		client: &http.Client{
			Timeout: config.HealthCheckTimeout,
		},
		config: config,
	}

	loadBalancer := &LoadBalancer{
		strategy: "round_robin",
		counters: make(map[string]int),
	}

	registry := &ServiceRegistry{
		services:      make(map[string][]*ServiceInstance),
		serviceHealth: make(map[string]*ServiceHealth),
		config:        config,
		logger:        logger,
		metrics:       metrics,
		loadBalancer:  loadBalancer,
		healthChecker: healthChecker,
	}

	// Initialize with default services
	registry.initializeServices()

	// Start background health checker
	go registry.startHealthChecker()

	return registry
}

// initializeServices initializes the service registry with default services
func (sr *ServiceRegistry) initializeServices() {
	defaultServices := map[string]*ServiceInstance{
		"auth-service": {
			ID:        "auth-service-1",
			Name:      "auth-service",
			Host:      "auth-service",
			Port:      3001,
			Health:    "healthy",
			LastCheck: time.Now(),
			Metadata:  map[string]string{"version": "v1", "region": "local"},
			Weight:    1,
			Tags:      []string{"auth", "security"},
		},
		"form-service": {
			ID:        "form-service-1",
			Name:      "form-service",
			Host:      "form-service",
			Port:      8001,
			Health:    "healthy",
			LastCheck: time.Now(),
			Metadata:  map[string]string{"version": "v1", "region": "local"},
			Weight:    1,
			Tags:      []string{"forms", "storage"},
		},
		"response-service": {
			ID:        "response-service-1",
			Name:      "response-service",
			Host:      "response-service",
			Port:      3002,
			Health:    "healthy",
			LastCheck: time.Now(),
			Metadata:  map[string]string{"version": "v1", "region": "local"},
			Weight:    1,
			Tags:      []string{"responses", "data"},
		},
		"analytics-service": {
			ID:        "analytics-service-1",
			Name:      "analytics-service",
			Host:      "analytics-service",
			Port:      5001,
			Health:    "healthy",
			LastCheck: time.Now(),
			Metadata:  map[string]string{"version": "v1", "region": "local"},
			Weight:    1,
			Tags:      []string{"analytics", "reporting"},
		},
		"realtime-service": {
			ID:        "realtime-service-1",
			Name:      "realtime-service",
			Host:      "realtime-service",
			Port:      8002,
			Health:    "healthy",
			LastCheck: time.Now(),
			Metadata:  map[string]string{"version": "v1", "region": "local"},
			Weight:    1,
			Tags:      []string{"realtime", "websockets"},
		},
	}

	// Initialize services slice for each service name
	for name, service := range defaultServices {
		sr.services[name] = []*ServiceInstance{service}
		sr.serviceHealth[name] = &ServiceHealth{
			HealthyInstances:   1,
			UnhealthyInstances: 0,
			LastHealthCheck:    time.Now(),
			ResponseTime:       50 * time.Millisecond,
			FailureCount:       0,
			SuccessCount:       1,
		}
	}
}

// GetHealthyService returns a healthy service instance using load balancing
func (sr *ServiceRegistry) GetHealthyService(serviceName string) (*ServiceInstance, error) {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	instances, exists := sr.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", serviceName)
	}

	// Filter healthy instances
	var healthyInstances []*ServiceInstance
	for _, instance := range instances {
		if instance.Health == "healthy" {
			healthyInstances = append(healthyInstances, instance)
		}
	}

	if len(healthyInstances) == 0 {
		return nil, fmt.Errorf("no healthy instances available for service: %s", serviceName)
	}

	// Apply load balancing
	return sr.loadBalancer.SelectInstance(serviceName, healthyInstances), nil
}

// PerformHealthCheck performs a health check on a service instance
func (sr *ServiceRegistry) PerformHealthCheck(instance *ServiceInstance) (bool, time.Duration) {
	return sr.healthChecker.CheckHealth(instance)
}

// MarkUnhealthy marks a service instance as unhealthy
func (sr *ServiceRegistry) MarkUnhealthy(instanceID string) {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	for _, instances := range sr.services {
		for _, instance := range instances {
			if instance.ID == instanceID {
				instance.Health = "unhealthy"
				instance.LastCheck = time.Now()
				sr.logger.Warn(fmt.Sprintf("Marked instance %s as unhealthy", instanceID))
				return
			}
		}
	}
}

// RecordAccess records access to a service instance for load balancing
func (sr *ServiceRegistry) RecordAccess(instanceID string) {
	// This could be used for weighted load balancing in the future
	sr.logger.Debug(fmt.Sprintf("Recorded access to instance: %s", instanceID))
}

// SelectInstance selects an instance using load balancing
func (lb *LoadBalancer) SelectInstance(serviceName string, instances []*ServiceInstance) *ServiceInstance {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	if len(instances) == 0 {
		return nil
	}

	if len(instances) == 1 {
		return instances[0]
	}

	// Simple round-robin load balancing
	counter := lb.counters[serviceName]
	selectedInstance := instances[counter%len(instances)]
	lb.counters[serviceName] = counter + 1

	return selectedInstance
}

// CheckHealth performs a health check on a service instance
func (hc *ServiceHealthChecker) CheckHealth(instance *ServiceInstance) (bool, time.Duration) {
	start := time.Now()
	healthURL := fmt.Sprintf("http://%s:%d/health", instance.Host, instance.Port)

	resp, err := hc.client.Get(healthURL)
	duration := time.Since(start)

	if err != nil {
		return false, duration
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300, duration
}

// startHealthChecker starts the background health checking process
func (sr *ServiceRegistry) startHealthChecker() {
	ticker := time.NewTicker(sr.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		sr.performBulkHealthCheck()
	}
}

// performBulkHealthCheck performs health checks on all registered services
func (sr *ServiceRegistry) performBulkHealthCheck() {
	sr.mutex.RLock()
	allInstances := make([]*ServiceInstance, 0)
	for _, instances := range sr.services {
		allInstances = append(allInstances, instances...)
	}
	sr.mutex.RUnlock()

	for _, instance := range allInstances {
		go func(inst *ServiceInstance) {
			healthy, responseTime := sr.healthChecker.CheckHealth(inst)

			sr.mutex.Lock()
			inst.LastCheck = time.Now()
			if healthy {
				inst.Health = "healthy"
			} else {
				inst.Health = "unhealthy"
			}
			sr.mutex.Unlock()

			// Update health metrics
			if health, exists := sr.serviceHealth[inst.Name]; exists {
				if healthy {
					health.SuccessCount++
				} else {
					health.FailureCount++
				}
				health.LastHealthCheck = time.Now()
				health.ResponseTime = responseTime
			}
		}(instance)
	}
}

// Step 5: Enhanced Service Discovery Middleware
func ServiceDiscoveryMiddleware(registry *ServiceRegistry, logger logger.Logger, metrics *metrics.Collector) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Extract service name from the request path
			serviceName := extractServiceName(r.URL.Path)
			if serviceName == "" {
				// If no specific service, proceed without service discovery
				next(w, r)
				return
			}

			// Get healthy service instance with load balancing
			serviceInstance, err := registry.GetHealthyService(serviceName)
			if err != nil {
				logger.Error(fmt.Sprintf("Service discovery failed: %s (path: %s) - %v", serviceName, r.URL.Path, err))

				// Record metrics
				if metrics != nil {
					metrics.RecordError("service_not_found", "service_discovery")
				}

				http.Error(w, fmt.Sprintf("Service '%s' not available", serviceName), http.StatusServiceUnavailable)
				return
			}

			// Perform real-time health check
			if registry.config.RealTimeHealthCheck {
				healthy, responseTime := registry.PerformHealthCheck(serviceInstance)
				if !healthy {
					logger.Warn(fmt.Sprintf("Service failed health check: %s at %s:%d",
						serviceName, serviceInstance.Host, serviceInstance.Port))

					// Mark service as unhealthy and try to get another instance
					registry.MarkUnhealthy(serviceInstance.ID)

					// Try to get another healthy instance
					altInstance, altErr := registry.GetHealthyService(serviceName)
					if altErr != nil {
						logger.Error(fmt.Sprintf("No healthy instances available for service: %s", serviceName))
						if metrics != nil {
							metrics.RecordError("service_unhealthy", "service_discovery")
						}
						http.Error(w, fmt.Sprintf("Service '%s' is unhealthy", serviceName), http.StatusServiceUnavailable)
						return
					}
					serviceInstance = altInstance
				} else {
					// Update response time metrics - simplified logging for now
					logger.Debug(fmt.Sprintf("Health check response time: %v for service %s", responseTime, serviceName))
				}
			}

			// Add service information to request context
			ctx := r.Context()
			ctx = context.WithValue(ctx, DiscoveredServiceKey, serviceInstance)
			ctx = context.WithValue(ctx, ServiceNameKey, serviceName)
			ctx = context.WithValue(ctx, ServiceHostKey, serviceInstance.Host)
			ctx = context.WithValue(ctx, ServicePortKey, serviceInstance.Port)
			ctx = context.WithValue(ctx, ServiceInstanceIDKey, serviceInstance.ID)

			// Update service metadata in headers for downstream services
			r.Header.Set("X-Service-Name", serviceInstance.Name)
			r.Header.Set("X-Service-Host", serviceInstance.Host)
			r.Header.Set("X-Service-Port", strconv.Itoa(serviceInstance.Port))
			r.Header.Set("X-Service-Instance-ID", serviceInstance.ID)
			r.Header.Set("X-Service-Version", serviceInstance.Metadata["version"])
			r.Header.Set("X-Service-Region", serviceInstance.Metadata["region"])
			r.Header.Set("X-Load-Balancer-Backend", fmt.Sprintf("%s:%d", serviceInstance.Host, serviceInstance.Port))

			// Record successful service discovery - simplified for now
			discoveryDuration := time.Since(start)
			logger.Debug(fmt.Sprintf("Service discovery successful for %s in %v", serviceName, discoveryDuration))

			logger.Debug(fmt.Sprintf("Service discovered successfully: %s at %s:%d (instance: %s, duration: %v)",
				serviceName, serviceInstance.Host, serviceInstance.Port, serviceInstance.ID, discoveryDuration))

			// Update service access tracking
			registry.RecordAccess(serviceInstance.ID)

			// Continue to next middleware with enriched context
			next(w, r.WithContext(ctx))
		}
	}
}

// extractServiceName extracts the service name from the request path
func extractServiceName(path string) string {
	// Remove leading slash and split by slash
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		return ""
	}

	// Map path prefixes to service names
	pathToService := map[string]string{
		"auth":          "auth-service",
		"forms":         "form-service",
		"responses":     "response-service",
		"analytics":     "analytics-service",
		"collaboration": "collaboration-service",
		"realtime":      "realtime-service",
		"events":        "event-bus-service",
		"files":         "file-upload-service",
		"upload":        "file-upload-service",
		"api/v1/auth":   "auth-service",
		"api/v1/forms":  "form-service",
		"api/auth":      "auth-service",
		"api/forms":     "form-service",
	}

	// Try exact match first
	if serviceName, exists := pathToService[parts[0]]; exists {
		return serviceName
	}

	// Try with api prefix
	if len(parts) >= 2 && parts[0] == "api" {
		if serviceName, exists := pathToService[parts[1]]; exists {
			return serviceName
		}

		// Try api/v1 prefix
		if len(parts) >= 3 && parts[1] == "v1" {
			if serviceName, exists := pathToService[parts[2]]; exists {
				return serviceName
			}
		}
	}

	return ""
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
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
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

			// Create wrapped response writer
			ww := &wrappedResponseWriter{ResponseWriter: w, statusCode: 200}

			// Process request
			next(ww, r)

			// Log request completion
			duration := time.Since(start)
			statusCode := ww.statusCode

			logMessage := fmt.Sprintf("Request completed: method=%s path=%s status=%d client_ip=%s user_agent=%s latency_ms=%d body_size=%d request_id=%s",
				method, path, statusCode, clientIP, userAgent, duration.Milliseconds(), ww.bytesWritten, requestID)

			if statusCode >= 400 {
				logger.Error(logMessage)
			} else if statusCode >= 300 {
				logger.Warn(logMessage)
			} else {
				logger.Info(logMessage)
			}
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
					logger.Errorf("Panic recovered: request_id=%s method=%s path=%s panic=%v",
						requestID, r.Method, r.URL.Path, err)

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
	// Pre-parse CIDR blocks for better performance
	var parsedAllowedNetworks []*net.IPNet
	var parsedBlockedNetworks []*net.IPNet
	var plainAllowedIPs []string
	var plainBlockedIPs []string

	// Parse allowed IPs
	for _, ipOrCIDR := range whitelistConfig.AllowedIPs {
		if strings.Contains(ipOrCIDR, "/") {
			_, network, err := net.ParseCIDR(ipOrCIDR)
			if err == nil {
				parsedAllowedNetworks = append(parsedAllowedNetworks, network)
			}
		} else {
			plainAllowedIPs = append(plainAllowedIPs, ipOrCIDR)
		}
	}

	// Parse blocked IPs
	for _, ipOrCIDR := range whitelistConfig.BlockedIPs {
		if strings.Contains(ipOrCIDR, "/") {
			_, network, err := net.ParseCIDR(ipOrCIDR)
			if err == nil {
				parsedBlockedNetworks = append(parsedBlockedNetworks, network)
			}
		} else {
			plainBlockedIPs = append(plainBlockedIPs, ipOrCIDR)
		}
	}

	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Skip if whitelist is disabled
			if !whitelistConfig.Enabled {
				next(w, r)
				return
			}

			// Skip for certain paths (health checks, metrics, etc.)
			if isPublicEndpoint(r.URL.Path) {
				next(w, r)
				return
			}

			clientIP := getClientIPFromConfig(r, whitelistConfig)
			ipAddr := net.ParseIP(clientIP)
			if ipAddr == nil {
				// Invalid IP address
				http.Error(w, "Invalid IP address", http.StatusForbidden)
				return
			}

			// Check if IP is blocked (explicit IPs)
			for _, blockedIP := range plainBlockedIPs {
				if clientIP == blockedIP {
					http.Error(w, "IP address is blocked", http.StatusForbidden)
					return
				}
			}

			// Check if IP is in blocked networks
			for _, network := range parsedBlockedNetworks {
				if network.Contains(ipAddr) {
					http.Error(w, "IP address is blocked", http.StatusForbidden)
					return
				}
			}

			// If whitelist is configured, check if IP is allowed
			if len(whitelistConfig.AllowedIPs) > 0 {
				allowed := false

				// Check explicit IPs
				for _, allowedIP := range plainAllowedIPs {
					if clientIP == allowedIP {
						allowed = true
						break
					}
				}

				// Check networks
				if !allowed {
					for _, network := range parsedAllowedNetworks {
						if network.Contains(ipAddr) {
							allowed = true
							break
						}
					}
				}

				if !allowed {
					http.Error(w, "IP address not in whitelist", http.StatusForbidden)
					return
				}
			}

			// Add client IP to context
			ctx := context.WithValue(r.Context(), ClientIPKey, clientIP)
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
			ctx := context.WithValue(r.Context(), UserAuthenticatedKey, true)
			next(w, r.WithContext(ctx))
		}
	}
}

// Step 4: Rate Limiting Middleware
func RateLimit(rateLimitConfig config.RateLimitConfig) Middleware {
	// Redis-based distributed rate limiter with simple fallback
	globalLimiters := make(map[string]*HybridRateLimiter)
	endpointLimiters := make(map[string]map[string]*HybridRateLimiter)

	// Get Redis URL from config (fallback to env var or local)
	redisURL := rateLimitConfig.RedisURL
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Skip if rate limiting is disabled
			if !rateLimitConfig.Enabled {
				next(w, r)
				return
			}

			// Get client identifier
			clientID := getClientIdentifier(r)
			path := r.URL.Path

			// Check for endpoint-specific rate limits
			var rps int
			var window time.Duration
			var endpointLimit bool

			for pattern, limit := range rateLimitConfig.Endpoints {
				if matchPath(path, pattern) {
					rps = limit.RPS
					window = limit.Window
					endpointLimit = true
					break
				}
			}

			// Use global limits if no endpoint-specific limits found
			if !endpointLimit {
				rps = rateLimitConfig.RPS
				window = rateLimitConfig.Window
			}

			// Get or create appropriate rate limiter
			var limiter *HybridRateLimiter
			var rateLimitKey string

			if endpointLimit {
				// Use endpoint-specific limiter
				if _, exists := endpointLimiters[path]; !exists {
					endpointLimiters[path] = make(map[string]*HybridRateLimiter)
				}

				if _, exists := endpointLimiters[path][clientID]; !exists {
					endpointLimiters[path][clientID] = NewHybridRateLimiter(redisURL, window, rps)
				}

				limiter = endpointLimiters[path][clientID]
				rateLimitKey = fmt.Sprintf("rate_limit:%s:%s", path, clientID)
			} else {
				// Use global limiter
				if _, exists := globalLimiters[clientID]; !exists {
					globalLimiters[clientID] = NewHybridRateLimiter(redisURL, window, rps)
				}

				limiter = globalLimiters[clientID]
				rateLimitKey = fmt.Sprintf("rate_limit:global:%s", clientID)
			}

			// Check if request is allowed
			if !limiter.Allow(rateLimitKey) {
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rps))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(window).Unix(), 10))
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			// Add rate limit headers (simplified for Redis-based limiting)
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rps))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(rps/2)) // Approximate remaining
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(window).Unix(), 10))

			next(w, r)
		}
	}
}

// Step 6: Circuit Breaker Middleware
func CircuitBreaker(config config.CircuitBreakerConfig) Middleware {
	// Advanced circuit breaker with multiple failure modes
	breakers := make(map[string]*AdvancedCircuitBreaker)

	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !config.Enabled {
				next(w, r)
				return
			}

			// Determine circuit breaker key (service or endpoint based)
			key := getCircuitBreakerKey(r, config.Mode)

			// Get or create circuit breaker for this key
			if _, exists := breakers[key]; !exists {
				breakers[key] = NewAdvancedCircuitBreaker(config)
			}

			breaker := breakers[key]

			// Check if circuit is open
			if !breaker.AllowRequest() {
				w.Header().Set("X-Circuit-Breaker-State", breaker.GetState())
				http.Error(w, "Service temporarily unavailable - circuit breaker open", http.StatusServiceUnavailable)
				return
			}

			// Create response recorder to capture status
			recorder := &StatusRecorder{ResponseWriter: w}
			start := time.Now()

			// Execute request
			next(recorder, r)

			// Record result in circuit breaker
			duration := time.Since(start)
			success := recorder.Status < 500 && recorder.Status != 0
			breaker.RecordResult(success, duration)
		}
	}
}

// AdvancedCircuitBreaker implements sophisticated circuit breaking
type AdvancedCircuitBreaker struct {
	config          config.CircuitBreakerConfig
	state           CircuitBreakerState
	failureCount    int
	successCount    int
	requestCount    int
	lastFailureTime time.Time
	lastSuccessTime time.Time
	nextAttemptTime time.Time
	responseTimeSum time.Duration
	recentResults   []CircuitBreakerResult
	healthChecker   *HealthChecker
	mutex           sync.RWMutex
}

// CircuitBreakerState represents the circuit breaker state
type CircuitBreakerState int

const (
	Closed CircuitBreakerState = iota
	Open
	HalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case Closed:
		return "closed"
	case Open:
		return "open"
	case HalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerResult represents the result of a request
type CircuitBreakerResult struct {
	Success   bool
	Duration  time.Duration
	Timestamp time.Time
}

// HealthChecker provides health check functionality
type HealthChecker struct {
	endpoint string
	timeout  time.Duration
	client   *http.Client
}

// NewAdvancedCircuitBreaker creates a new advanced circuit breaker
func NewAdvancedCircuitBreaker(config config.CircuitBreakerConfig) *AdvancedCircuitBreaker {
	healthChecker := &HealthChecker{
		endpoint: config.HealthCheckURL,
		timeout:  config.HealthCheckTimeout,
		client: &http.Client{
			Timeout: config.HealthCheckTimeout,
		},
	}

	return &AdvancedCircuitBreaker{
		config:        config,
		state:         Closed,
		recentResults: make([]CircuitBreakerResult, 0, config.WindowSize),
		healthChecker: healthChecker,
	}
}

// AllowRequest checks if a request should be allowed
func (cb *AdvancedCircuitBreaker) AllowRequest() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()

	switch cb.state {
	case Closed:
		return true
	case Open:
		// Check if it's time to try a health check
		if now.After(cb.nextAttemptTime) {
			if cb.config.HealthCheckURL != "" && cb.isHealthy() {
				cb.state = HalfOpen
				cb.successCount = 0
				return true
			}
			// Update next attempt time
			cb.nextAttemptTime = now.Add(cb.config.RetryInterval)
		}
		return false
	case HalfOpen:
		// Allow limited requests to test recovery
		return cb.successCount < cb.config.HalfOpenMaxRequests
	default:
		return false
	}
}

// RecordResult records the result of a request
func (cb *AdvancedCircuitBreaker) RecordResult(success bool, duration time.Duration) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	result := CircuitBreakerResult{
		Success:   success,
		Duration:  duration,
		Timestamp: now,
	}

	// Add to recent results (sliding window)
	cb.recentResults = append(cb.recentResults, result)
	if len(cb.recentResults) > cb.config.WindowSize {
		cb.recentResults = cb.recentResults[1:]
	}

	cb.requestCount++
	cb.responseTimeSum += duration

	if success {
		cb.successCount++
		cb.lastSuccessTime = now

		// If in half-open state and enough successes, close circuit
		if cb.state == HalfOpen && cb.successCount >= cb.config.SuccessThreshold {
			cb.state = Closed
			cb.failureCount = 0
		}
	} else {
		cb.failureCount++
		cb.lastFailureTime = now

		// Check if should open circuit
		if cb.shouldOpen() {
			cb.state = Open
			cb.nextAttemptTime = now.Add(cb.config.RetryInterval)
		}
	}
}

// shouldOpen determines if the circuit should be opened
func (cb *AdvancedCircuitBreaker) shouldOpen() bool {
	// Need minimum requests before making decisions
	if cb.requestCount < cb.config.MinRequests {
		return false
	}

	// Check failure rate
	failureRate := cb.calculateFailureRate()
	if failureRate >= cb.config.FailureThreshold {
		return true
	}

	// Check consecutive failures
	if cb.failureCount >= cb.config.MaxFailures {
		return true
	}

	// Check response time threshold
	avgResponseTime := cb.getAverageResponseTime()
	if avgResponseTime > cb.config.ResponseTimeThreshold {
		return true
	}

	return false
}

// calculateFailureRate calculates the current failure rate
func (cb *AdvancedCircuitBreaker) calculateFailureRate() float64 {
	if len(cb.recentResults) == 0 {
		return 0.0
	}

	failures := 0
	for _, result := range cb.recentResults {
		if !result.Success {
			failures++
		}
	}

	return float64(failures) / float64(len(cb.recentResults))
}

// getAverageResponseTime calculates average response time
func (cb *AdvancedCircuitBreaker) getAverageResponseTime() time.Duration {
	if cb.requestCount == 0 {
		return 0
	}
	return cb.responseTimeSum / time.Duration(cb.requestCount)
}

// isHealthy performs a health check
func (cb *AdvancedCircuitBreaker) isHealthy() bool {
	if cb.healthChecker.endpoint == "" {
		return false
	}

	resp, err := cb.healthChecker.client.Get(cb.healthChecker.endpoint)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// GetState returns the current circuit breaker state
func (cb *AdvancedCircuitBreaker) GetState() string {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state.String()
}

// GetStats returns circuit breaker statistics
func (cb *AdvancedCircuitBreaker) GetStats() map[string]interface{} {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	return map[string]interface{}{
		"state":             cb.state.String(),
		"failure_count":     cb.failureCount,
		"success_count":     cb.successCount,
		"request_count":     cb.requestCount,
		"failure_rate":      cb.calculateFailureRate(),
		"avg_response_time": cb.getAverageResponseTime().String(),
		"last_failure":      cb.lastFailureTime,
		"last_success":      cb.lastSuccessTime,
	}
}

// StatusRecorder captures the HTTP status code
type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *StatusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

// getCircuitBreakerKey generates a key for the circuit breaker
func getCircuitBreakerKey(r *http.Request, mode string) string {
	switch mode {
	case "service":
		return extractServiceName(r.URL.Path)
	case "endpoint":
		return r.URL.Path
	case "global":
		return "global"
	default:
		return extractServiceName(r.URL.Path)
	}
}

// Step 7: Advanced Metrics Collection Middleware
func AdvancedMetricsMiddleware(metrics *metrics.Collector, logger logger.Logger) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Generate trace ID for distributed tracing
			traceID := generateTraceID()

			// Add trace ID to context and headers
			ctx := context.WithValue(r.Context(), "trace_id", traceID)
			r = r.WithContext(ctx)
			r.Header.Set("X-Trace-ID", traceID)

			// Create enhanced response recorder
			recorder := &EnhancedResponseRecorder{
				ResponseWriter: w,
				statusCode:     200,
				responseSize:   0,
				startTime:      start,
			}

			// Increment requests in flight
			if metrics != nil {
				metrics.IncrementRequestsInFlight()
				defer metrics.DecrementRequestsInFlight()
			}

			// Extract request metadata
			method := r.Method
			path := r.URL.Path
			userAgent := r.Header.Get("User-Agent")
			clientIP := getClientIdentifier(r) // Use existing function
			serviceName := extractServiceName(path)

			// Set response headers with monitoring information
			recorder.Header().Set("X-Trace-ID", traceID)
			recorder.Header().Set("X-Request-Start", strconv.FormatInt(start.UnixNano(), 10))

			// Execute the request
			next(recorder, r)

			// Calculate metrics
			duration := time.Since(start)
			statusCode := recorder.statusCode
			responseSize := recorder.responseSize

			// Record comprehensive metrics
			if metrics != nil {
				// Core HTTP metrics
				metrics.RecordRequestWithTrace(method, path, statusCode, duration, responseSize, traceID)

				// Service-specific metrics
				if serviceName != "" {
					metrics.RecordServiceLatency(serviceName, method, duration)

					// Record upstream metrics
					metrics.RecordUpstreamRequest(serviceName, method, statusCode, duration)

					// Record errors if applicable
					if statusCode >= 400 {
						errorType := categorizeError(statusCode)
						metrics.RecordUpstreamError(serviceName, errorType)
					}
				}

				// Business metrics based on path patterns
				recordBusinessMetrics(metrics, path, method, statusCode)

				// Performance categorization
				recordPerformanceMetrics(metrics, duration, path, statusCode)
			}

			// Enhanced structured logging
			logFields := map[string]interface{}{
				"trace_id":      traceID,
				"method":        method,
				"path":          path,
				"status_code":   statusCode,
				"duration_ms":   duration.Milliseconds(),
				"response_size": responseSize,
				"client_ip":     clientIP,
				"user_agent":    userAgent,
				"service":       serviceName,
				"timestamp":     start.Format(time.RFC3339Nano),
			}

			// Add user context if available
			if userID, ok := r.Context().Value(UserIDKey).(string); ok && userID != "" {
				logFields["user_id"] = userID
			}

			if userRole, ok := r.Context().Value(UserRoleKey).(string); ok && userRole != "" {
				logFields["user_role"] = userRole
			}

			// Enhanced structured logging with production-ready context
			logMessage := fmt.Sprintf("%s %s completed", method, path)
			logLvl := determineLogLevel(statusCode, duration)

			// Create structured fields
			fields := map[string]interface{}{
				"trace_id": traceID, "method": method, "path": path, "status_code": statusCode,
				"duration_ms": duration.Milliseconds(), "client_ip": clientIP, "response_size": responseSize,
				"service": serviceName, "timestamp": time.Now().Format(time.RFC3339Nano),
			}

			// Add user context if available
			if userID, ok := r.Context().Value(UserIDKey).(string); ok && userID != "" {
				fields["user_id"] = userID
			}

			// Use structured logging based on level
			switch logLvl {
			case "error":
				logger.WithFields(fields).Error(fmt.Sprintf("Request failed: %s", logMessage))
			case "warn":
				logger.WithFields(fields).Warn(fmt.Sprintf("Slow request: %s", logMessage))
			case "info":
				logger.WithFields(fields).Info(fmt.Sprintf("Request completed: %s", logMessage))
			default:
				logger.WithFields(fields).Debug(fmt.Sprintf("Request: %s", logMessage))
			}
		}
	}
}

// EnhancedResponseRecorder captures detailed response information
type EnhancedResponseRecorder struct {
	http.ResponseWriter
	statusCode   int
	responseSize int64
	startTime    time.Time
	headers      http.Header
}

func (r *EnhancedResponseRecorder) WriteHeader(status int) {
	r.statusCode = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *EnhancedResponseRecorder) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseSize += int64(size)
	return size, err
}

// Helper functions for advanced metrics

func generateTraceID() string {
	// Simple trace ID generation (in production, use proper distributed tracing)
	return fmt.Sprintf("trace-%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
}

func categorizeError(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "server_error"
	case statusCode >= 400:
		return "client_error"
	case statusCode >= 300:
		return "redirect"
	default:
		return "success"
	}
}

func recordBusinessMetrics(metrics *metrics.Collector, path, method string, statusCode int) {
	// Record business-specific metrics based on API endpoints
	if strings.Contains(path, "/forms") && method == "POST" && statusCode < 400 {
		metrics.RecordBusinessMetric("form_submission", "api", "success")
	} else if strings.Contains(path, "/forms") && method == "POST" && statusCode >= 400 {
		metrics.RecordBusinessMetric("form_submission", "api", "failure")
	}

	if strings.Contains(path, "/auth/register") && method == "POST" && statusCode < 400 {
		metrics.RecordBusinessMetric("user_registration", "api", "success")
	}

	if strings.Contains(path, "/responses") && method == "POST" && statusCode < 400 {
		metrics.RecordBusinessMetric("form_submission", "response", "success")
	}
}

func recordPerformanceMetrics(metrics *metrics.Collector, duration time.Duration, path string, statusCode int) {
	// Categorize performance
	if duration > 5*time.Second {
		metrics.RecordError("very_slow_request", "performance")
	} else if duration > 1*time.Second {
		metrics.RecordError("slow_request", "performance")
	}

	// Record by endpoint type
	if strings.Contains(path, "/api/") {
		if duration > 500*time.Millisecond {
			metrics.RecordError("slow_api_request", "performance")
		}
	}
}

func determineLogLevel(statusCode int, duration time.Duration) string {
	if statusCode >= 500 {
		return "error"
	}
	if statusCode >= 400 {
		return "warn"
	}
	if duration > 1*time.Second {
		return "warn"
	}
	if statusCode >= 300 {
		return "info"
	}
	return "debug"
}

// matchPath checks if a path matches a pattern
func matchPath(path, pattern string) bool {
	// Exact match
	if path == pattern {
		return true
	}

	// Prefix match with wildcard
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}

	// Regex pattern matching for advanced routing
	if strings.HasPrefix(pattern, "^") || strings.Contains(pattern, ".*") ||
		strings.Contains(pattern, "+") || strings.Contains(pattern, "[") {
		// Compile and match regex pattern
		if matched, err := regexp.MatchString(pattern, path); err == nil && matched {
			return true
		}
	}

	// Path parameter matching (e.g., /users/{id} matches /users/123)
	if strings.Contains(pattern, "{") && strings.Contains(pattern, "}") {
		patternParts := strings.Split(pattern, "/")
		pathParts := strings.Split(path, "/")

		if len(patternParts) != len(pathParts) {
			return false
		}

		for i, part := range patternParts {
			if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
				// This is a parameter, any value matches
				continue
			}
			if part != pathParts[i] {
				return false
			}
		}
		return true
	}

	return false
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

// RedisRateLimiter provides distributed rate limiting using Redis
type RedisRateLimiter struct {
	client      *redis.Client
	windowSize  time.Duration
	maxRequests int
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(redisURL string, windowSize time.Duration, maxRequests int) (*RedisRateLimiter, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisRateLimiter{
		client:      client,
		windowSize:  windowSize,
		maxRequests: maxRequests,
	}, nil
}

// Allow implements sliding window rate limiting using Redis
func (r *RedisRateLimiter) Allow(key string) (bool, error) {
	ctx := context.Background()
	now := time.Now()
	windowStart := now.Add(-r.windowSize)

	// Redis pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Remove expired entries
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart.UnixNano(), 10))

	// Count current requests in window
	countCmd := pipe.ZCard(ctx, key)

	// Add current request timestamp
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: now.UnixNano(),
	})

	// Set expiration for the key
	pipe.Expire(ctx, key, r.windowSize+time.Minute)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("Redis pipeline execution failed: %w", err)
	}

	// Check if limit exceeded
	currentCount := countCmd.Val()
	return currentCount < int64(r.maxRequests), nil
}

// Close closes the Redis connection
func (r *RedisRateLimiter) Close() error {
	return r.client.Close()
}

// Simple rate limiter implementation (fallback)
type simpleRateLimiter struct {
	maxRequests int
	windowSize  time.Duration
	requests    []time.Time
}

// HybridRateLimiter provides rate limiting with Redis fallback
type HybridRateLimiter struct {
	redis    *RedisRateLimiter
	simple   *simpleRateLimiter
	useRedis bool
}

// NewHybridRateLimiter creates a new hybrid rate limiter
func NewHybridRateLimiter(redisURL string, windowSize time.Duration, maxRequests int) *HybridRateLimiter {
	hybrid := &HybridRateLimiter{
		simple: newSimpleRateLimiter(maxRequests, maxRequests), // burst = maxRequests
	}

	// Try to initialize Redis
	if redisURL != "" {
		redisLimiter, err := NewRedisRateLimiter(redisURL, windowSize, maxRequests)
		if err == nil {
			hybrid.redis = redisLimiter
			hybrid.useRedis = true
		}
	}

	return hybrid
}

// Allow checks if the request should be allowed
func (h *HybridRateLimiter) Allow(key string) bool {
	if h.useRedis && h.redis != nil {
		allowed, err := h.redis.Allow(key)
		if err == nil {
			return allowed
		}
		// Fallback to simple limiter if Redis fails
		h.useRedis = false
	}
	return h.simple.Allow()
}

// Close closes the Redis connection if available
func (h *HybridRateLimiter) Close() error {
	if h.redis != nil {
		return h.redis.Close()
	}
	return nil
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

// validateToken performs JWT token validation with proper parsing
func validateToken(token string, config config.JWTConfig) bool {
	// Parse and validate JWT token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Validate the algorithm
		switch config.Algorithm {
		case "HS256":
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.Secret), nil
		case "HS384", "HS512":
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.Secret), nil
		case "RS256", "RS384", "RS512":
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// Parse public key for RSA validation
			return parseRSAPublicKey(config.PublicKey)
		default:
			return nil, fmt.Errorf("unsupported algorithm: %s", config.Algorithm)
		}
	})

	if err != nil {
		return false
	}

	// Validate token and claims
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		// Check expiration
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return false
			}
		}

		// Check not before
		if nbf, ok := claims["nbf"].(float64); ok {
			if time.Now().Unix() < int64(nbf) {
				return false
			}
		}

		// Check issued at
		if iat, ok := claims["iat"].(float64); ok {
			if time.Now().Unix() < int64(iat) {
				return false
			}
		}

		return true
	}

	return false
}

// parseRSAPublicKey parses RSA public key from PEM format
func parseRSAPublicKey(publicKeyPEM string) (interface{}, error) {
	// For now, return the key as-is (in a real implementation, parse PEM format)
	// This would use crypto/x509 and crypto/rsa packages
	if len(publicKeyPEM) == 0 {
		return nil, fmt.Errorf("empty public key")
	}

	// Simplified validation - in production, parse actual PEM key
	return []byte(publicKeyPEM), nil
}
