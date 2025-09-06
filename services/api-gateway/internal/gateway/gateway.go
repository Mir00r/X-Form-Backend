package gateway

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/config"
	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/proxy"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Gateway handles routing requests to microservices
type Gateway struct {
	config       *config.Config
	proxyManager *proxy.ProxyManager
	httpClient   *http.Client
}

// New creates a new gateway instance
func New(cfg *config.Config, proxyManager *proxy.ProxyManager) *Gateway {
	return &Gateway{
		config:       cfg,
		proxyManager: proxyManager,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProxyToAuth proxies requests to the auth service
func (g *Gateway) ProxyToAuth(c *gin.Context) {
	path := strings.TrimPrefix(c.Request.URL.Path, "/api/v1")
	g.proxyRequest(c, g.config.AuthServiceURL, path)
}

// ProxyToForm proxies requests to the form service
func (g *Gateway) ProxyToForm(c *gin.Context) {
	path := c.Request.URL.Path
	if strings.HasPrefix(path, "/api/v1/forms") {
		path = strings.TrimPrefix(path, "/api/v1")
	}
	g.proxyRequest(c, g.config.FormServiceURL, path)
}

// ProxyToResponse proxies requests to the response service
func (g *Gateway) ProxyToResponse(c *gin.Context) {
	path := c.Request.URL.Path
	if strings.HasPrefix(path, "/api/v1/responses") {
		path = strings.TrimPrefix(path, "/api/v1")
	}
	g.proxyRequest(c, g.config.ResponseServiceURL, path)
}

// ProxyToAnalytics proxies requests to the analytics service
func (g *Gateway) ProxyToAnalytics(c *gin.Context) {
	path := strings.TrimPrefix(c.Request.URL.Path, "/api/v1")
	g.proxyRequest(c, g.config.AnalyticsServiceURL, path)
}

// ProxyToFile proxies requests to the file service
func (g *Gateway) ProxyToFile(c *gin.Context) {
	path := strings.TrimPrefix(c.Request.URL.Path, "/api/v1")
	g.proxyRequest(c, g.config.FileServiceURL, path)
}

// proxyRequest handles the actual proxying of requests
func (g *Gateway) proxyRequest(c *gin.Context, targetBaseURL, targetPath string) {
	// Build target URL
	targetURL := targetBaseURL + targetPath
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	// Create request body reader
	var bodyReader io.Reader
	if c.Request.Body != nil {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logrus.WithError(err).Error("Failed to read request body")
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_request",
				"message": "Failed to read request body",
			})
			return
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create new request
	req, err := http.NewRequestWithContext(
		context.Background(),
		c.Request.Method,
		targetURL,
		bodyReader,
	)
	if err != nil {
		logrus.WithError(err).Error("Failed to create request")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "proxy_error",
			"message": "Failed to create request",
		})
		return
	}

	// Copy headers (excluding hop-by-hop headers)
	for key, values := range c.Request.Header {
		if !isHopByHopHeader(key) {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	// Add gateway headers
	req.Header.Set("X-Forwarded-For", c.ClientIP())
	req.Header.Set("X-Forwarded-Host", c.Request.Host)
	req.Header.Set("X-Forwarded-Proto", getProtocol(c.Request))

	// Add request ID if available
	if requestID := c.GetString("RequestID"); requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}

	// Add user context headers if authenticated
	if userID := c.GetString("UserID"); userID != "" {
		req.Header.Set("X-User-ID", userID)
	}
	if email := c.GetString("Email"); email != "" {
		req.Header.Set("X-User-Email", email)
	}

	// Execute request
	start := time.Now()
	resp, err := g.httpClient.Do(req)
	duration := time.Since(start)

	// Log the request
	logrus.WithFields(logrus.Fields{
		"method":      c.Request.Method,
		"target_url":  targetURL,
		"duration_ms": duration.Milliseconds(),
		"request_id":  c.GetString("RequestID"),
	}).Info("Proxied request")

	if err != nil {
		logrus.WithError(err).Error("Request failed")
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "proxy_error",
			"message": "Failed to reach the service",
		})
		return
	}
	defer resp.Body.Close()

	// Copy response headers (excluding hop-by-hop headers)
	for key, values := range resp.Header {
		if !isHopByHopHeader(key) {
			for _, value := range values {
				c.Header(key, value)
			}
		}
	}

	// Copy response body
	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}

// isHopByHopHeader checks if a header is hop-by-hop
func isHopByHopHeader(header string) bool {
	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}

	headerLower := strings.ToLower(header)
	for _, hopHeader := range hopByHopHeaders {
		if strings.ToLower(hopHeader) == headerLower {
			return true
		}
	}
	return false
}

// getProtocol determines the protocol used in the request
func getProtocol(req *http.Request) string {
	if req.TLS != nil {
		return "https"
	}
	if proto := req.Header.Get("X-Forwarded-Proto"); proto != "" {
		return proto
	}
	return "http"
}
