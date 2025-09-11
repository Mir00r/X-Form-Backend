package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	// Setup test environment
	_ = os.Setenv("ENV", "test")
	_ = os.Setenv("LOG_LEVEL", "error")

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Run tests
	code := m.Run()

	// Cleanup
	os.Exit(code)
}

func setupTestRouter() *gin.Engine {
	router := gin.New()
	setupGinRoutes(router)
	return router
}

func TestHealthEndpoint(t *testing.T) {
	router := setupTestRouter()

	// Create a request to the health endpoint
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check content type
	expectedContentType := "application/json; charset=utf-8"
	if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, expectedContentType)
	}

	// Check response body structure
	var healthResponse HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &healthResponse); err != nil {
		t.Errorf("Could not parse JSON response: %v", err)
	}

	if healthResponse.Status != "healthy" {
		t.Errorf("handler returned wrong status: got %v want %v", healthResponse.Status, "healthy")
	}
}

func TestMetricsEndpoint(t *testing.T) {
	router := setupTestRouter()

	// Create a request to the metrics endpoint
	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check content type
	expectedContentType := "text/plain; charset=utf-8"
	if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, expectedContentType)
	}

	// Check response body contains expected metrics
	body := rr.Body.String()
	if !contains(body, "api_gateway_requests_total") {
		t.Errorf("handler response does not contain expected metrics: got %v", body)
	}
}

func TestSwaggerEndpoint(t *testing.T) {
	router := setupTestRouter()

	// Create a request to the swagger endpoint
	req, err := http.NewRequest("GET", "/swagger/index.html", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(rr, req)

	// Check status code (should be 200 for HTML or 301/302 for redirect)
	if status := rr.Code; status != http.StatusOK && status != http.StatusMovedPermanently && status != http.StatusFound {
		t.Errorf("swagger endpoint returned unexpected status code: got %v", status)
	}
}

func TestGatewayInfoEndpoint(t *testing.T) {
	router := setupTestRouter()

	// Create a request to the root endpoint
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response body structure
	var infoResponse GatewayInfoResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &infoResponse); err != nil {
		t.Errorf("Could not parse JSON response: %v", err)
	}

	if infoResponse.Message != "Enhanced X-Form API Gateway" {
		t.Errorf("handler returned wrong message: got %v want %v", infoResponse.Message, "Enhanced X-Form API Gateway")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
