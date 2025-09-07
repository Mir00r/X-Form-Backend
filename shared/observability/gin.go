package observability

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// GinMiddleware returns Gin middleware for observability
func (p *Provider) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Increment active connections
		p.metrics.IncrementActiveConnections()
		defer p.metrics.DecrementActiveConnections()

		// Start tracing span
		ctx, span := p.tracing.StartSpan(c.Request.Context(), fmt.Sprintf("HTTP %s %s", c.Request.Method, c.FullPath()))
		defer span.End()

		// Update request context with span
		c.Request = c.Request.WithContext(ctx)

		// Add request attributes to span
		p.tracing.AddSpanAttributes(span,
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.route", c.FullPath()),
			attribute.String("http.scheme", c.Request.URL.Scheme),
			attribute.String("http.host", c.Request.Host),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("component", "http"),
		)

		if c.Request.ContentLength > 0 {
			p.tracing.AddSpanAttributes(span,
				attribute.Int64("http.request.size", c.Request.ContentLength),
			)
		}

		// Extract user information
		userID := p.extractUserIDFromGin(c)
		if userID != "" {
			p.tracing.AddSpanAttributes(span, attribute.String("user.id", userID))
			p.errors.SetUserContext(userID, "", "")
		}

		// Add breadcrumb
		p.errors.AddBreadcrumb(
			fmt.Sprintf("HTTP %s %s", c.Request.Method, c.FullPath()),
			"http.request",
			map[string]interface{}{
				"method": c.Request.Method,
				"url":    c.Request.URL.String(),
				"route":  c.FullPath(),
			},
		)

		// Set trace context in Gin context for downstream use
		c.Set("trace_id", p.tracing.ExtractTraceID(ctx))
		c.Set("span_id", p.tracing.ExtractSpanID(ctx))

		// Continue with request
		c.Next()

		// Record metrics after request completion
		duration := time.Since(start)
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}
		statusCode := c.Writer.Status()
		statusCodeStr := strconv.Itoa(statusCode)

		p.metrics.RecordHTTPRequest(
			c.Request.Method,
			endpoint,
			statusCodeStr,
			userID,
			duration,
			c.Request.ContentLength,
			int64(c.Writer.Size()),
		)

		// Add response attributes to span
		p.tracing.AddSpanAttributes(span,
			attribute.Int("http.status_code", statusCode),
			attribute.Int("http.response.size", c.Writer.Size()),
			attribute.Float64("http.duration_ms", float64(duration.Nanoseconds())/1e6),
		)

		// Set span status based on HTTP status code
		if statusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", statusCode))
		} else {
			span.SetStatus(codes.Ok, "")
		}

		// Capture errors for 4xx/5xx responses
		if statusCode >= 400 {
			requestID := p.extractRequestIDFromGin(c)

			errorContext := ErrorContext{
				UserID:     userID,
				RequestID:  requestID,
				TraceID:    p.tracing.ExtractTraceID(ctx),
				SpanID:     p.tracing.ExtractSpanID(ctx),
				Endpoint:   endpoint,
				Method:     c.Request.Method,
				StatusCode: statusCode,
				Component:  "http",
				Operation:  "request",
				Extra: map[string]interface{}{
					"route":       c.FullPath(),
					"remote_addr": c.ClientIP(),
				},
			}

			if statusCode >= 500 {
				p.errors.CaptureError(
					fmt.Errorf("HTTP %d error: %s %s", statusCode, c.Request.Method, c.FullPath()),
					errorContext,
				)
				p.metrics.RecordServiceError("http_error", "http", "error")
			} else {
				p.metrics.RecordServiceError("http_client_error", "http", "warning")
			}
		}

		// Record successful operation
		if statusCode < 400 {
			p.metrics.RecordServiceOperation("http_request", "success", "http")
		}
	}
}

// extractUserIDFromGin extracts user ID from Gin context
func (p *Provider) extractUserIDFromGin(c *gin.Context) string {
	// Try different methods to extract user ID
	if userID := c.GetHeader("X-User-ID"); userID != "" {
		return userID
	}

	// Try from Gin context (set by auth middleware)
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			return uid
		}
	}

	// Try from JWT claims
	if claims, exists := c.Get("claims"); exists {
		if claimsMap, ok := claims.(map[string]interface{}); ok {
			if userID, ok := claimsMap["user_id"].(string); ok {
				return userID
			}
			if sub, ok := claimsMap["sub"].(string); ok {
				return sub
			}
		}
	}

	return ""
}

// extractRequestIDFromGin extracts request ID from Gin context
func (p *Provider) extractRequestIDFromGin(c *gin.Context) string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	if requestID := c.GetHeader("Request-ID"); requestID != "" {
		return requestID
	}
	if requestID, exists := c.Get("request_id"); exists {
		if rid, ok := requestID.(string); ok {
			return rid
		}
	}
	return ""
}

// ProxyObservability provides observability for proxy requests
func (p *Provider) ProxyObservability(c *gin.Context, targetService, targetURL string) func(int, error) {
	start := time.Now()

	// Start span for proxy operation
	ctx, span := p.tracing.StartSpan(c.Request.Context(), fmt.Sprintf("Proxy %s %s", c.Request.Method, targetService))
	defer span.End()

	// Add span attributes
	p.tracing.AddSpanAttributes(span,
		attribute.String("proxy.target_service", targetService),
		attribute.String("proxy.target_url", targetURL),
		attribute.String("proxy.source_method", c.Request.Method),
		attribute.String("proxy.source_path", c.Request.URL.Path),
		attribute.String("component", "proxy"),
	)

	// Add breadcrumb
	p.errors.AddBreadcrumb(
		fmt.Sprintf("Proxy to %s", targetService),
		"proxy.request",
		map[string]interface{}{
			"target_service": targetService,
			"target_url":     targetURL,
			"source_method":  c.Request.Method,
			"source_path":    c.Request.URL.Path,
		},
	)

	// Return cleanup function
	return func(statusCode int, err error) {
		duration := time.Since(start)
		statusCodeStr := strconv.Itoa(statusCode)

		// Add response attributes
		p.tracing.AddSpanAttributes(span,
			attribute.Int("proxy.status_code", statusCode),
			attribute.Float64("proxy.duration_ms", float64(duration.Nanoseconds())/1e6),
		)

		if err != nil {
			p.tracing.RecordError(span, err)

			// Capture error
			errorContext := ErrorContext{
				UserID:     p.extractUserIDFromGin(c),
				RequestID:  p.extractRequestIDFromGin(c),
				TraceID:    p.tracing.ExtractTraceID(ctx),
				SpanID:     p.tracing.ExtractSpanID(ctx),
				Component:  "proxy",
				Operation:  "forward",
				StatusCode: statusCode,
				Extra: map[string]interface{}{
					"target_service": targetService,
					"target_url":     targetURL,
					"source_method":  c.Request.Method,
					"source_path":    c.Request.URL.Path,
				},
			}
			p.errors.CaptureError(err, errorContext)
			p.metrics.RecordServiceError("proxy_error", "proxy", "error")
		}

		// Record external service call metrics
		p.metrics.RecordExternalServiceCall(targetService, c.Request.Method, targetURL, statusCodeStr, duration)

		// Record proxy operation metric
		status := "success"
		if err != nil || statusCode >= 500 {
			status = "error"
		} else if statusCode >= 400 {
			status = "client_error"
		}
		p.metrics.RecordServiceOperation("proxy_request", status, "proxy")
	}
}
