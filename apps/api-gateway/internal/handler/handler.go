// Package handler provides HTTP handlers for the API Gateway
// Implements the Step 6 (Request Transformation) and Step 7 (Reverse Proxy) from the architecture diagram
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/Mir00r/X-Form-Backend/enhanced-architecture/api-gateway/internal/config"
	"github.com/Mir00r/X-Form-Backend/enhanced-architecture/api-gateway/pkg/logger"
	"github.com/Mir00r/X-Form-Backend/enhanced-architecture/api-gateway/pkg/metrics"
)

// Handler provides HTTP request handling functionality
type Handler struct {
	config   *config.Config
	logger   logger.Logger
	metrics  *metrics.Collector
	services map[string]*Service
	proxies  map[string]*httputil.ReverseProxy
}

// Service represents an upstream service configuration
type Service struct {
	Name            string            `json:"name"`
	BaseURL         string            `json:"base_url"`
	HealthCheckPath string            `json:"health_check_path"`
	Timeout         time.Duration     `json:"timeout"`
	Headers         map[string]string `json:"headers"`
	CircuitBreaker  *CircuitBreaker   `json:"circuit_breaker"`
	LoadBalancer    *LoadBalancer     `json:"load_balancer"`
}

// CircuitBreaker represents circuit breaker configuration
type CircuitBreaker struct {
	Enabled           bool          `json:"enabled"`
	FailureThreshold  int           `json:"failure_threshold"`
	RecoveryTimeout   time.Duration `json:"recovery_timeout"`
	TestRequestVolume int           `json:"test_request_volume"`
	state             CircuitState
	failures          int
	lastFailureTime   time.Time
	testRequests      int
}

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// LoadBalancer represents load balancer configuration
type LoadBalancer struct {
	Strategy  string   `json:"strategy"` // round_robin, weighted, least_connections
	Instances []string `json:"instances"`
	current   int
}

// NewHandler creates a new handler instance
func NewHandler(cfg *config.Config, logger logger.Logger, metrics *metrics.Collector) *Handler {
	h := &Handler{
		config:   cfg,
		logger:   logger,
		metrics:  metrics,
		services: make(map[string]*Service),
		proxies:  make(map[string]*httputil.ReverseProxy),
	}

	// Initialize services from configuration
	h.initializeServices()

	return h
}

// initializeServices initializes service configurations and reverse proxies
func (h *Handler) initializeServices() {
	// Example service configurations (in production, load from config)
	services := map[string]*Service{
		"auth-service": {
			Name:            "auth-service",
			BaseURL:         "http://localhost:8001",
			HealthCheckPath: "/health",
			Timeout:         time.Second * 30,
			Headers:         map[string]string{"Service": "auth-service"},
			CircuitBreaker: &CircuitBreaker{
				Enabled:           true,
				FailureThreshold:  5,
				RecoveryTimeout:   time.Second * 30,
				TestRequestVolume: 3,
				state:             CircuitClosed,
			},
			LoadBalancer: &LoadBalancer{
				Strategy:  "round_robin",
				Instances: []string{"http://localhost:8001"},
			},
		},
		"form-service": {
			Name:            "form-service",
			BaseURL:         "http://localhost:8002",
			HealthCheckPath: "/health",
			Timeout:         time.Second * 30,
			Headers:         map[string]string{"Service": "form-service"},
			CircuitBreaker: &CircuitBreaker{
				Enabled:           true,
				FailureThreshold:  5,
				RecoveryTimeout:   time.Second * 30,
				TestRequestVolume: 3,
				state:             CircuitClosed,
			},
			LoadBalancer: &LoadBalancer{
				Strategy:  "round_robin",
				Instances: []string{"http://localhost:8002"},
			},
		},
		"response-service": {
			Name:            "response-service",
			BaseURL:         "http://localhost:8003",
			HealthCheckPath: "/health",
			Timeout:         time.Second * 30,
			Headers:         map[string]string{"Service": "response-service"},
			CircuitBreaker: &CircuitBreaker{
				Enabled:           true,
				FailureThreshold:  5,
				RecoveryTimeout:   time.Second * 30,
				TestRequestVolume: 3,
				state:             CircuitClosed,
			},
			LoadBalancer: &LoadBalancer{
				Strategy:  "round_robin",
				Instances: []string{"http://localhost:8003"},
			},
		},
		"collaboration-service": {
			Name:            "collaboration-service",
			BaseURL:         "http://localhost:8004",
			HealthCheckPath: "/health",
			Timeout:         time.Second * 30,
			Headers:         map[string]string{"Service": "collaboration-service"},
			CircuitBreaker: &CircuitBreaker{
				Enabled:           true,
				FailureThreshold:  5,
				RecoveryTimeout:   time.Second * 30,
				TestRequestVolume: 3,
				state:             CircuitClosed,
			},
			LoadBalancer: &LoadBalancer{
				Strategy:  "round_robin",
				Instances: []string{"http://localhost:8004"},
			},
		},
		"realtime-service": {
			Name:            "realtime-service",
			BaseURL:         "http://localhost:8005",
			HealthCheckPath: "/health",
			Timeout:         time.Second * 30,
			Headers:         map[string]string{"Service": "realtime-service"},
			CircuitBreaker: &CircuitBreaker{
				Enabled:           true,
				FailureThreshold:  5,
				RecoveryTimeout:   time.Second * 30,
				TestRequestVolume: 3,
				state:             CircuitClosed,
			},
			LoadBalancer: &LoadBalancer{
				Strategy:  "round_robin",
				Instances: []string{"http://localhost:8005"},
			},
		},
		"analytics-service": {
			Name:            "analytics-service",
			BaseURL:         "http://localhost:8006",
			HealthCheckPath: "/health",
			Timeout:         time.Second * 30,
			Headers:         map[string]string{"Service": "analytics-service"},
			CircuitBreaker: &CircuitBreaker{
				Enabled:           true,
				FailureThreshold:  5,
				RecoveryTimeout:   time.Second * 30,
				TestRequestVolume: 3,
				state:             CircuitClosed,
			},
			LoadBalancer: &LoadBalancer{
				Strategy:  "round_robin",
				Instances: []string{"http://localhost:8006"},
			},
		},
	}

	// Initialize services and proxies
	for name, service := range services {
		h.services[name] = service
		h.proxies[name] = h.createReverseProxy(service)
	}
}

// createReverseProxy creates a reverse proxy for a service
func (h *Handler) createReverseProxy(service *Service) *httputil.ReverseProxy {
	target, _ := url.Parse(service.BaseURL)

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Customize the director to handle request transformation
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Step 6: Request Transformation
		h.transformRequest(req, service)
	}

	// Customize error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		h.handleProxyError(w, r, service, err)
	}

	// Customize response modifier
	proxy.ModifyResponse = func(resp *http.Response) error {
		return h.transformResponse(resp, service)
	}

	return proxy
}

// Step 6: Request Transformation
func (h *Handler) transformRequest(req *http.Request, service *Service) {
	// Add service-specific headers
	for key, value := range service.Headers {
		req.Header.Set(key, value)
	}

	// Add correlation ID for tracing
	if requestID := req.Header.Get("X-Request-ID"); requestID != "" {
		req.Header.Set("X-Correlation-ID", requestID)
	}

	// Add authentication headers if present in context
	if userID := req.Context().Value("user_id"); userID != nil {
		req.Header.Set("X-User-ID", fmt.Sprintf("%v", userID))
	}

	if userRole := req.Context().Value("user_role"); userRole != nil {
		req.Header.Set("X-User-Role", fmt.Sprintf("%v", userRole))
	}

	// Add timestamp
	req.Header.Set("X-Gateway-Timestamp", time.Now().UTC().Format(time.RFC3339))

	// Add client IP
	if clientIP := req.Context().Value("client_ip"); clientIP != nil {
		req.Header.Set("X-Client-IP", fmt.Sprintf("%v", clientIP))
	}

	// Service-specific transformations
	switch service.Name {
	case "auth-service":
		h.transformAuthRequest(req)
	case "form-service":
		h.transformFormRequest(req)
	}
}

// transformAuthRequest applies auth-service specific transformations
func (h *Handler) transformAuthRequest(req *http.Request) {
	// Add specific headers for auth service
	req.Header.Set("X-Service-Version", "v1")
}

// transformFormRequest applies form-service specific transformations
func (h *Handler) transformFormRequest(req *http.Request) {
	// Add specific headers for form service
	req.Header.Set("X-Service-Version", "v1")
	req.Header.Set("X-Content-Type", "application/json")
}

// transformResponse modifies the response from upstream services
func (h *Handler) transformResponse(resp *http.Response, service *Service) error {
	// Add service identification headers
	resp.Header.Set("X-Served-By", service.Name)
	resp.Header.Set("X-Gateway", "x-form-api-gateway")

	// Record metrics
	h.metrics.RecordUpstreamRequest(
		service.Name,
		resp.Request.Method,
		resp.StatusCode,
		time.Since(time.Now()), // This should be calculated from request start
	)

	return nil
}

// handleProxyError handles errors from upstream services
func (h *Handler) handleProxyError(w http.ResponseWriter, r *http.Request, service *Service, err error) {
	// Record error metrics
	h.metrics.RecordUpstreamError(service.Name, "proxy_error")

	// Update circuit breaker
	if service.CircuitBreaker != nil && service.CircuitBreaker.Enabled {
		h.recordFailure(service.CircuitBreaker)
	}

	// Log error
	h.logger.WithFields(map[string]interface{}{
		"service":    service.Name,
		"error":      err.Error(),
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("X-Request-ID"),
	}).Error("Upstream service error")

	// Return appropriate error response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadGateway)

	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":       "SERVICE_UNAVAILABLE",
			"message":    "Upstream service is currently unavailable",
			"service":    service.Name,
			"request_id": r.Header.Get("X-Request-ID"),
		},
	}

	json.NewEncoder(w).Encode(response)
}

// Step 7: Reverse Proxy Handler
func (h *Handler) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	// Determine target service based on request path
	serviceName := h.determineService(r.URL.Path)
	if serviceName == "" {
		h.handleNotFound(w, r)
		return
	}

	service, exists := h.services[serviceName]
	if !exists {
		h.handleServiceNotFound(w, r, serviceName)
		return
	}

	// Check circuit breaker
	if service.CircuitBreaker != nil && service.CircuitBreaker.Enabled {
		if !h.checkCircuitBreaker(service.CircuitBreaker) {
			h.handleCircuitOpen(w, r, serviceName)
			return
		}
	}

	// Get the appropriate proxy
	proxy, exists := h.proxies[serviceName]
	if !exists {
		h.handleServiceNotFound(w, r, serviceName)
		return
	}

	// Record start time for metrics
	start := time.Now()

	// Add timeout to request context
	ctx, cancel := context.WithTimeout(r.Context(), service.Timeout)
	defer cancel()
	r = r.WithContext(ctx)

	// Forward the request
	proxy.ServeHTTP(w, r)

	// Record success
	if service.CircuitBreaker != nil && service.CircuitBreaker.Enabled {
		h.recordSuccess(service.CircuitBreaker)
	}

	// Record metrics (duration calculation is simplified here)
	h.metrics.RecordUpstreamRequest(serviceName, r.Method, 200, time.Since(start))
}

// determineService determines which service to route to based on the request path
func (h *Handler) determineService(path string) string {
	// Service routing rules
	routes := map[string]string{
		"/api/v1/auth/":          "auth-service",
		"/api/v1/forms/":         "form-service",
		"/api/v1/responses/":     "response-service",
		"/api/v1/collaboration/": "collaboration-service",
		"/api/v1/realtime/":      "realtime-service",
		"/api/v1/analytics/":     "analytics-service",
		"/ws/":                   "realtime-service",
	}

	for prefix, service := range routes {
		if strings.HasPrefix(path, prefix) {
			return service
		}
	}

	return ""
}

// Circuit Breaker Implementation

// checkCircuitBreaker checks if requests can pass through the circuit breaker
func (h *Handler) checkCircuitBreaker(cb *CircuitBreaker) bool {
	now := time.Now()

	switch cb.state {
	case CircuitClosed:
		return true

	case CircuitOpen:
		if now.Sub(cb.lastFailureTime) > cb.RecoveryTimeout {
			cb.state = CircuitHalfOpen
			cb.testRequests = 0
			h.metrics.SetCircuitBreakerState("service", metrics.CircuitBreakerHalfOpen)
			return true
		}
		return false

	case CircuitHalfOpen:
		if cb.testRequests < cb.TestRequestVolume {
			cb.testRequests++
			return true
		}
		return false

	default:
		return false
	}
}

// recordFailure records a failure in the circuit breaker
func (h *Handler) recordFailure(cb *CircuitBreaker) {
	cb.failures++
	cb.lastFailureTime = time.Now()

	if cb.state == CircuitClosed && cb.failures >= cb.FailureThreshold {
		cb.state = CircuitOpen
		h.metrics.RecordCircuitBreakerTrip("service")
		h.metrics.SetCircuitBreakerState("service", metrics.CircuitBreakerOpen)
	} else if cb.state == CircuitHalfOpen {
		cb.state = CircuitOpen
		cb.testRequests = 0
		h.metrics.SetCircuitBreakerState("service", metrics.CircuitBreakerOpen)
	}
}

// recordSuccess records a success in the circuit breaker
func (h *Handler) recordSuccess(cb *CircuitBreaker) {
	if cb.state == CircuitHalfOpen {
		cb.failures = 0
		cb.state = CircuitClosed
		cb.testRequests = 0
		h.metrics.SetCircuitBreakerState("service", metrics.CircuitBreakerClosed)
	} else if cb.state == CircuitClosed {
		cb.failures = 0
	}
}

// Health Check Handler
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := h.checkServiceHealth()

	w.Header().Set("Content-Type", "application/json")
	if status.Overall == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(status)
}

// HealthStatus represents the health status of the gateway and services
type HealthStatus struct {
	Overall   string                   `json:"overall"`
	Gateway   string                   `json:"gateway"`
	Services  map[string]ServiceHealth `json:"services"`
	Timestamp time.Time                `json:"timestamp"`
}

// ServiceHealth represents the health status of a single service
type ServiceHealth struct {
	Status       string        `json:"status"`
	ResponseTime time.Duration `json:"response_time"`
	LastCheck    time.Time     `json:"last_check"`
	Error        string        `json:"error,omitempty"`
}

// checkServiceHealth checks the health of all registered services
func (h *Handler) checkServiceHealth() HealthStatus {
	status := HealthStatus{
		Overall:   "healthy",
		Gateway:   "healthy",
		Services:  make(map[string]ServiceHealth),
		Timestamp: time.Now(),
	}

	unhealthyCount := 0

	for name, service := range h.services {
		health := h.checkSingleServiceHealth(service)
		status.Services[name] = health

		if health.Status != "healthy" {
			unhealthyCount++
		}
	}

	// Determine overall health
	if unhealthyCount > 0 {
		if unhealthyCount == len(h.services) {
			status.Overall = "unhealthy"
		} else {
			status.Overall = "degraded"
		}
	}

	return status
}

// checkSingleServiceHealth checks the health of a single service
func (h *Handler) checkSingleServiceHealth(service *Service) ServiceHealth {
	start := time.Now()

	// Create health check request
	healthURL := service.BaseURL + service.HealthCheckPath
	req, err := http.NewRequest(http.MethodGet, healthURL, nil)
	if err != nil {
		return ServiceHealth{
			Status:       "unhealthy",
			ResponseTime: 0,
			LastCheck:    start,
			Error:        fmt.Sprintf("Failed to create request: %v", err),
		}
	}

	// Set timeout
	client := &http.Client{
		Timeout: time.Second * 5,
	}

	// Make request
	resp, err := client.Do(req)
	responseTime := time.Since(start)

	if err != nil {
		return ServiceHealth{
			Status:       "unhealthy",
			ResponseTime: responseTime,
			LastCheck:    start,
			Error:        fmt.Sprintf("Request failed: %v", err),
		}
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return ServiceHealth{
			Status:       "healthy",
			ResponseTime: responseTime,
			LastCheck:    start,
		}
	}

	return ServiceHealth{
		Status:       "unhealthy",
		ResponseTime: responseTime,
		LastCheck:    start,
		Error:        fmt.Sprintf("Health check returned status %d", resp.StatusCode),
	}
}

// Metrics Handler
func (h *Handler) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.metrics.Handler().ServeHTTP(w, r)
}

// WebSocket Handler for realtime service
func (h *Handler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Check if this is a WebSocket upgrade request
	if r.Header.Get("Upgrade") != "websocket" {
		http.Error(w, "This endpoint requires WebSocket upgrade", http.StatusBadRequest)
		return
	}

	// Forward to realtime service
	service := h.services["realtime-service"]
	if service == nil {
		http.Error(w, "Realtime service not available", http.StatusServiceUnavailable)
		return
	}

	proxy := h.proxies["realtime-service"]
	if proxy == nil {
		http.Error(w, "Realtime service proxy not available", http.StatusServiceUnavailable)
		return
	}

	proxy.ServeHTTP(w, r)
}

// Error handlers

func (h *Handler) handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)

	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":       "ROUTE_NOT_FOUND",
			"message":    "The requested route was not found",
			"path":       r.URL.Path,
			"request_id": r.Header.Get("X-Request-ID"),
		},
	}

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) handleServiceNotFound(w http.ResponseWriter, r *http.Request, serviceName string) {
	h.logger.WithFields(map[string]interface{}{
		"service":    serviceName,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("X-Request-ID"),
	}).Error("Service not found")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)

	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":       "SERVICE_NOT_FOUND",
			"message":    "The requested service is not available",
			"service":    serviceName,
			"request_id": r.Header.Get("X-Request-ID"),
		},
	}

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) handleCircuitOpen(w http.ResponseWriter, r *http.Request, serviceName string) {
	h.logger.WithFields(map[string]interface{}{
		"service":    serviceName,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("X-Request-ID"),
	}).Warn("Circuit breaker is open")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)

	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":       "CIRCUIT_BREAKER_OPEN",
			"message":    "Service is temporarily unavailable due to high failure rate",
			"service":    serviceName,
			"request_id": r.Header.Get("X-Request-ID"),
		},
	}

	json.NewEncoder(w).Encode(response)
}

// GatewayHandlers represents the collection of HTTP handlers for the gateway
type GatewayHandlers struct {
	handler *Handler
}

// NewGatewayHandlers creates a new instance of GatewayHandlers
func NewGatewayHandlers(cfg *config.Config, logger logger.Logger, metrics *metrics.Collector) *GatewayHandlers {
	handler := NewHandler(cfg, logger, metrics)
	return &GatewayHandlers{
		handler: handler,
	}
}

// ProxyRequest handles proxying requests to backend services
func (gh *GatewayHandlers) ProxyRequest(w http.ResponseWriter, r *http.Request) {
	gh.handler.ProxyHandler(w, r)
}

// ProxyToService proxies a request to a specific service by name
func (gh *GatewayHandlers) ProxyToService(w http.ResponseWriter, r *http.Request, serviceName string) {
	gh.handler.ProxyToService(w, r, serviceName)
}

// HealthCheck handles health check requests
func (gh *GatewayHandlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	gh.handler.HealthHandler(w, r)
}

// Metrics handles metrics requests
func (gh *GatewayHandlers) Metrics(w http.ResponseWriter, r *http.Request) {
	gh.handler.MetricsHandler(w, r)
}

// WebSocket handles WebSocket requests
func (gh *GatewayHandlers) WebSocket(w http.ResponseWriter, r *http.Request) {
	gh.handler.WebSocketHandler(w, r)
}

// ProxyToService proxies a request to a specific service by name
func (h *Handler) ProxyToService(w http.ResponseWriter, r *http.Request, serviceName string) {
	service, exists := h.services[serviceName]
	if !exists {
		h.handleServiceNotFound(w, r, serviceName)
		return
	}

	// Check circuit breaker
	if service.CircuitBreaker != nil && service.CircuitBreaker.Enabled {
		if !h.checkCircuitBreaker(service.CircuitBreaker) {
			h.handleCircuitOpen(w, r, serviceName)
			return
		}
	}

	// Get the appropriate proxy
	proxy, exists := h.proxies[serviceName]
	if !exists {
		h.handleServiceNotFound(w, r, serviceName)
		return
	}

	// Record start time for metrics
	start := time.Now()

	// Add timeout to request context
	ctx, cancel := context.WithTimeout(r.Context(), service.Timeout)
	defer cancel()
	r = r.WithContext(ctx)

	// Add service-specific headers
	h.transformRequest(r, service)

	// Forward the request
	proxy.ServeHTTP(w, r)

	// Record success if the request was successful
	if service.CircuitBreaker != nil && service.CircuitBreaker.Enabled {
		h.recordSuccess(service.CircuitBreaker)
	}

	// Record metrics (duration calculation is simplified here)
	h.metrics.RecordUpstreamRequest(serviceName, r.Method, 200, time.Since(start))
}

// SetupRoutes sets up the HTTP routes for the gateway
func (h *Handler) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/health", h.HealthHandler)

	// Metrics endpoint
	mux.HandleFunc("/metrics", h.MetricsHandler)

	// WebSocket endpoint
	mux.HandleFunc("/ws/", h.WebSocketHandler)

	// Main proxy handler (catch-all)
	mux.HandleFunc("/", h.ProxyHandler)

	return mux
}
