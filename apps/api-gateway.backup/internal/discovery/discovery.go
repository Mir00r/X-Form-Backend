package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/config"
	"github.com/gin-gonic/gin"
)

// ServiceDiscovery handles service discovery and health monitoring
type ServiceDiscovery struct {
	config   *config.Config
	services map[string]*ServiceInfo
	mutex    sync.RWMutex
	client   *http.Client
	stopChan chan struct{}
}

// ServiceInfo represents information about a discovered service
type ServiceInfo struct {
	Name            string            `json:"name"`
	URL             string            `json:"url"`
	HealthEndpoint  string            `json:"health_endpoint"`
	Status          ServiceStatus     `json:"status"`
	LastHealthCheck time.Time         `json:"last_health_check"`
	ResponseTime    time.Duration     `json:"response_time"`
	ErrorCount      int               `json:"error_count"`
	Metadata        map[string]string `json:"metadata"`
	Version         string            `json:"version"`
	Tags            []string          `json:"tags"`
	Weight          int               `json:"weight"`
	Circuit         *CircuitBreaker   `json:"circuit"`
}

// ServiceStatus represents the status of a service
type ServiceStatus string

const (
	StatusHealthy     ServiceStatus = "healthy"
	StatusUnhealthy   ServiceStatus = "unhealthy"
	StatusUnknown     ServiceStatus = "unknown"
	StatusMaintenance ServiceStatus = "maintenance"
)

// CircuitBreaker represents circuit breaker state for a service
type CircuitBreaker struct {
	State        CircuitState  `json:"state"`
	FailureCount int           `json:"failure_count"`
	LastFailure  time.Time     `json:"last_failure"`
	NextAttempt  time.Time     `json:"next_attempt"`
	Threshold    int           `json:"threshold"`
	Timeout      time.Duration `json:"timeout"`
}

// CircuitState represents circuit breaker states
type CircuitState string

const (
	CircuitClosed   CircuitState = "closed"
	CircuitOpen     CircuitState = "open"
	CircuitHalfOpen CircuitState = "half_open"
)

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Version   string                 `json:"version,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Duration  time.Duration          `json:"duration"`
}

// LoadBalancer handles load balancing between service instances
type LoadBalancer struct {
	strategy LoadBalancingStrategy
	services []*ServiceInfo
	mutex    sync.RWMutex
	index    int
}

// LoadBalancingStrategy defines load balancing strategies
type LoadBalancingStrategy string

const (
	RoundRobin LoadBalancingStrategy = "round_robin"
	Random     LoadBalancingStrategy = "random"
	LeastConn  LoadBalancingStrategy = "least_conn"
	Weighted   LoadBalancingStrategy = "weighted"
)

// NewServiceDiscovery creates a new service discovery instance
func NewServiceDiscovery(cfg *config.Config) *ServiceDiscovery {
	sd := &ServiceDiscovery{
		config:   cfg,
		services: make(map[string]*ServiceInfo),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		stopChan: make(chan struct{}),
	}

	// Initialize services from configuration
	sd.initializeServices()

	// Start health monitoring
	go sd.startHealthMonitoring()

	return sd
}

// initializeServices initializes services from configuration
func (sd *ServiceDiscovery) initializeServices() {
	sd.mutex.Lock()
	defer sd.mutex.Unlock()

	// Add configured services
	services := map[string]config.ServiceConfig{
		"auth-service":          sd.config.Services.AuthService,
		"form-service":          sd.config.Services.FormService,
		"response-service":      sd.config.Services.ResponseService,
		"analytics-service":     sd.config.Services.AnalyticsService,
		"collaboration-service": sd.config.Services.CollaborationService,
		"realtime-service":      sd.config.Services.RealtimeService,
	}

	for name, serviceConfig := range services {
		serviceInfo := &ServiceInfo{
			Name:           name,
			URL:            serviceConfig.URL,
			HealthEndpoint: serviceConfig.HealthEndpoint,
			Status:         StatusUnknown,
			ErrorCount:     0,
			Weight:         1,
			Tags:           []string{"microservice", "x-form"},
			Circuit: &CircuitBreaker{
				State:     CircuitClosed,
				Threshold: 5,
				Timeout:   30 * time.Second,
			},
		}

		sd.services[name] = serviceInfo
		log.Printf("Registered service: %s at %s", name, serviceConfig.URL)
	}
}

// RegisterService registers a new service
func (sd *ServiceDiscovery) RegisterService(ctx context.Context, service *ServiceInfo) error {
	sd.mutex.Lock()
	defer sd.mutex.Unlock()

	if service.Circuit == nil {
		service.Circuit = &CircuitBreaker{
			State:     CircuitClosed,
			Threshold: 5,
			Timeout:   30 * time.Second,
		}
	}

	sd.services[service.Name] = service
	log.Printf("Registered service: %s at %s", service.Name, service.URL)

	return nil
}

// DeregisterService removes a service from discovery
func (sd *ServiceDiscovery) DeregisterService(ctx context.Context, serviceName string) error {
	sd.mutex.Lock()
	defer sd.mutex.Unlock()

	delete(sd.services, serviceName)
	log.Printf("Deregistered service: %s", serviceName)

	return nil
}

// GetService retrieves service information
func (sd *ServiceDiscovery) GetService(serviceName string) (*ServiceInfo, error) {
	sd.mutex.RLock()
	defer sd.mutex.RUnlock()

	service, exists := sd.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	return service, nil
}

// GetHealthyService returns a healthy service instance using load balancing
func (sd *ServiceDiscovery) GetHealthyService(serviceName string) (*ServiceInfo, error) {
	sd.mutex.RLock()
	defer sd.mutex.RUnlock()

	// For now, return the single service instance if healthy
	service, exists := sd.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	// Check circuit breaker
	if service.Circuit.State == CircuitOpen {
		if time.Now().Before(service.Circuit.NextAttempt) {
			return nil, fmt.Errorf("service %s circuit breaker is open", serviceName)
		}
		// Transition to half-open
		service.Circuit.State = CircuitHalfOpen
	}

	if service.Status == StatusHealthy || service.Status == StatusUnknown {
		return service, nil
	}

	return nil, fmt.Errorf("service %s is not healthy (status: %s)", serviceName, service.Status)
}

// GetAllServices returns all registered services
func (sd *ServiceDiscovery) GetAllServices() map[string]*ServiceInfo {
	sd.mutex.RLock()
	defer sd.mutex.RUnlock()

	// Create a copy to avoid data races
	result := make(map[string]*ServiceInfo)
	for name, service := range sd.services {
		serviceCopy := *service
		result[name] = &serviceCopy
	}

	return result
}

// HealthCheck performs a health check on a specific service
func (sd *ServiceDiscovery) HealthCheck(ctx context.Context, service *ServiceInfo) (*HealthResponse, error) {
	start := time.Now()

	healthURL, err := url.JoinPath(service.URL, service.HealthEndpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid health endpoint URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := sd.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	var healthResp HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		// If we can't decode JSON, assume service is healthy if status is 200
		healthResp = HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now(),
		}
	}

	healthResp.Duration = duration
	return &healthResp, nil
}

// startHealthMonitoring starts the background health monitoring routine
func (sd *ServiceDiscovery) startHealthMonitoring() {
	ticker := time.NewTicker(sd.config.Observability.HealthCheck.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sd.performHealthChecks()
		case <-sd.stopChan:
			return
		}
	}
}

// performHealthChecks performs health checks on all services
func (sd *ServiceDiscovery) performHealthChecks() {
	sd.mutex.RLock()
	services := make([]*ServiceInfo, 0, len(sd.services))
	for _, service := range sd.services {
		services = append(services, service)
	}
	sd.mutex.RUnlock()

	for _, service := range services {
		go sd.checkServiceHealth(service)
	}
}

// checkServiceHealth performs health check on a single service
func (sd *ServiceDiscovery) checkServiceHealth(service *ServiceInfo) {
	ctx, cancel := context.WithTimeout(context.Background(), sd.config.Observability.HealthCheck.Timeout)
	defer cancel()

	healthResp, err := sd.HealthCheck(ctx, service)

	sd.mutex.Lock()
	defer sd.mutex.Unlock()

	service.LastHealthCheck = time.Now()

	if err != nil {
		service.ErrorCount++
		service.Status = StatusUnhealthy

		// Handle circuit breaker
		service.Circuit.FailureCount++
		service.Circuit.LastFailure = time.Now()

		if service.Circuit.FailureCount >= service.Circuit.Threshold {
			service.Circuit.State = CircuitOpen
			service.Circuit.NextAttempt = time.Now().Add(service.Circuit.Timeout)
		}

		log.Printf("Health check failed for service %s: %v", service.Name, err)
	} else {
		service.Status = StatusHealthy
		service.ResponseTime = healthResp.Duration
		service.ErrorCount = 0

		// Reset circuit breaker on success
		if service.Circuit.State == CircuitHalfOpen || service.Circuit.State == CircuitClosed {
			service.Circuit.State = CircuitClosed
			service.Circuit.FailureCount = 0
		}

		// Update version if available
		if healthResp.Version != "" {
			service.Version = healthResp.Version
		}
	}
}

// ProxyRequest proxies a request to a healthy service instance
func (sd *ServiceDiscovery) ProxyRequest(serviceName string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		service, err := sd.GetHealthyService(serviceName)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": fmt.Sprintf("Service %s unavailable: %v", serviceName, err),
				"code":  "SERVICE_UNAVAILABLE",
			})
			return
		}

		// Create target URL
		targetURL, err := url.Parse(service.URL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid service URL",
				"code":  "INVALID_SERVICE_URL",
			})
			return
		}

		// Update target URL with request path
		targetURL.Path = c.Request.URL.Path
		targetURL.RawQuery = c.Request.URL.RawQuery

		// Create new request
		req, err := http.NewRequest(c.Request.Method, targetURL.String(), c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create proxy request",
				"code":  "PROXY_REQUEST_FAILED",
			})
			return
		}

		// Copy headers
		for name, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(name, value)
			}
		}

		// Add tracing headers
		req.Header.Set("X-Request-ID", c.GetString("request_id"))
		req.Header.Set("X-Gateway-Service", serviceName)

		// Execute request
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			// Update service error count
			sd.mutex.Lock()
			service.ErrorCount++
			sd.mutex.Unlock()

			c.JSON(http.StatusBadGateway, gin.H{
				"error": "Service request failed",
				"code":  "SERVICE_REQUEST_FAILED",
			})
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for name, values := range resp.Header {
			for _, value := range values {
				c.Header(name, value)
			}
		}

		// Copy response status and body
		c.Status(resp.StatusCode)

		// Stream response body
		buffer := make([]byte, 32*1024) // 32KB buffer
		for {
			n, err := resp.Body.Read(buffer)
			if n > 0 {
				c.Writer.Write(buffer[:n])
			}
			if err != nil {
				break
			}
		}
	})
}

// GetServicesEndpoint provides service discovery information
func (sd *ServiceDiscovery) GetServicesEndpoint() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		services := sd.GetAllServices()

		response := gin.H{
			"services":  services,
			"count":     len(services),
			"timestamp": time.Now(),
		}

		c.JSON(http.StatusOK, response)
	})
}

// GetServiceHealthEndpoint provides health information for a specific service
func (sd *ServiceDiscovery) GetServiceHealthEndpoint() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		serviceName := c.Param("service")

		service, err := sd.GetService(serviceName)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
				"code":  "SERVICE_NOT_FOUND",
			})
			return
		}

		response := gin.H{
			"service":           service.Name,
			"status":            service.Status,
			"url":               service.URL,
			"last_health_check": service.LastHealthCheck,
			"response_time":     service.ResponseTime.String(),
			"error_count":       service.ErrorCount,
			"circuit_breaker":   service.Circuit,
			"version":           service.Version,
			"tags":              service.Tags,
		}

		c.JSON(http.StatusOK, response)
	})
}

// GetServiceMetricsEndpoint provides metrics for service discovery
func (sd *ServiceDiscovery) GetServiceMetricsEndpoint() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		services := sd.GetAllServices()

		totalServices := len(services)
		healthyServices := 0
		unhealthyServices := 0
		unknownServices := 0

		for _, service := range services {
			switch service.Status {
			case StatusHealthy:
				healthyServices++
			case StatusUnhealthy:
				unhealthyServices++
			case StatusUnknown:
				unknownServices++
			}
		}

		response := gin.H{
			"total_services":     totalServices,
			"healthy_services":   healthyServices,
			"unhealthy_services": unhealthyServices,
			"unknown_services":   unknownServices,
			"health_percentage":  float64(healthyServices) / float64(totalServices) * 100,
			"timestamp":          time.Now(),
		}

		c.JSON(http.StatusOK, response)
	})
}

// Stop stops the service discovery monitoring
func (sd *ServiceDiscovery) Stop() {
	close(sd.stopChan)
}
