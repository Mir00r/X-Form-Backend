package traefik

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/config"
	"github.com/gin-gonic/gin"
)

// TraefikService handles Traefik ingress integration
type TraefikService struct {
	config *config.Config
	client *http.Client
}

// NewTraefikService creates a new Traefik service instance
func NewTraefikService(cfg *config.Config) *TraefikService {
	return &TraefikService{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TraefikRoute represents a Traefik dynamic route configuration
type TraefikRoute struct {
	Rule        string                 `json:"rule" yaml:"rule"`
	Service     string                 `json:"service" yaml:"service"`
	Priority    int                    `json:"priority,omitempty" yaml:"priority,omitempty"`
	Middlewares []string               `json:"middlewares,omitempty" yaml:"middlewares,omitempty"`
	TLS         *TraefikTLSConfig      `json:"tls,omitempty" yaml:"tls,omitempty"`
	Headers     map[string]string      `json:"headers,omitempty" yaml:"headers,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// TraefikService represents a Traefik service configuration
type TraefikServiceConfig struct {
	LoadBalancer *LoadBalancer          `json:"loadBalancer,omitempty" yaml:"loadBalancer,omitempty"`
	Weighted     *WeightedService       `json:"weighted,omitempty" yaml:"weighted,omitempty"`
	Mirroring    *MirroringService      `json:"mirroring,omitempty" yaml:"mirroring,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// LoadBalancer configuration for Traefik services
type LoadBalancer struct {
	Servers            []Server            `json:"servers" yaml:"servers"`
	Sticky             *Sticky             `json:"sticky,omitempty" yaml:"sticky,omitempty"`
	HealthCheck        *HealthCheck        `json:"healthCheck,omitempty" yaml:"healthCheck,omitempty"`
	PassHostHeader     *bool               `json:"passHostHeader,omitempty" yaml:"passHostHeader,omitempty"`
	ResponseForwarding *ResponseForwarding `json:"responseForwarding,omitempty" yaml:"responseForwarding,omitempty"`
}

// Server represents a backend server
type Server struct {
	URL    string `json:"url" yaml:"url"`
	Weight *int   `json:"weight,omitempty" yaml:"weight,omitempty"`
}

// Sticky session configuration
type Sticky struct {
	Cookie *StickyCookie `json:"cookie,omitempty" yaml:"cookie,omitempty"`
}

// StickyCookie configuration
type StickyCookie struct {
	Name     string `json:"name,omitempty" yaml:"name,omitempty"`
	Secure   *bool  `json:"secure,omitempty" yaml:"secure,omitempty"`
	HTTPOnly *bool  `json:"httpOnly,omitempty" yaml:"httpOnly,omitempty"`
	SameSite string `json:"sameSite,omitempty" yaml:"sameSite,omitempty"`
}

// HealthCheck configuration for load balancer
type HealthCheck struct {
	Path               string            `json:"path" yaml:"path"`
	Interval           string            `json:"interval,omitempty" yaml:"interval,omitempty"`
	Timeout            string            `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	HealthyStatusCodes []int             `json:"healthyStatusCodes,omitempty" yaml:"healthyStatusCodes,omitempty"`
	Headers            map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Hostname           string            `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	Port               int               `json:"port,omitempty" yaml:"port,omitempty"`
	Scheme             string            `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	FollowRedirects    *bool             `json:"followRedirects,omitempty" yaml:"followRedirects,omitempty"`
}

// ResponseForwarding configuration
type ResponseForwarding struct {
	FlushInterval string `json:"flushInterval,omitempty" yaml:"flushInterval,omitempty"`
}

// WeightedService for weighted load balancing
type WeightedService struct {
	Services []WeightedServiceEntry `json:"services" yaml:"services"`
	Sticky   *Sticky                `json:"sticky,omitempty" yaml:"sticky,omitempty"`
}

// WeightedServiceEntry represents a weighted service
type WeightedServiceEntry struct {
	Name   string `json:"name" yaml:"name"`
	Weight int    `json:"weight" yaml:"weight"`
}

// MirroringService for traffic mirroring
type MirroringService struct {
	Service string          `json:"service" yaml:"service"`
	Mirrors []MirrorService `json:"mirrors,omitempty" yaml:"mirrors,omitempty"`
}

// MirrorService represents a mirror service
type MirrorService struct {
	Name    string `json:"name" yaml:"name"`
	Percent int    `json:"percent" yaml:"percent"`
}

// TraefikTLSConfig for TLS configuration
type TraefikTLSConfig struct {
	CertResolver string            `json:"certResolver,omitempty" yaml:"certResolver,omitempty"`
	Domains      []TraefikDomain   `json:"domains,omitempty" yaml:"domains,omitempty"`
	Options      string            `json:"options,omitempty" yaml:"options,omitempty"`
	Store        string            `json:"store,omitempty" yaml:"store,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// TraefikDomain represents a TLS domain
type TraefikDomain struct {
	Main string   `json:"main" yaml:"main"`
	SANs []string `json:"sans,omitempty" yaml:"sans,omitempty"`
}

// TraefikMiddleware represents a Traefik middleware configuration
type TraefikMiddleware struct {
	AddPrefix         *AddPrefix         `json:"addPrefix,omitempty" yaml:"addPrefix,omitempty"`
	BasicAuth         *BasicAuth         `json:"basicAuth,omitempty" yaml:"basicAuth,omitempty"`
	Buffering         *Buffering         `json:"buffering,omitempty" yaml:"buffering,omitempty"`
	Chain             *Chain             `json:"chain,omitempty" yaml:"chain,omitempty"`
	CircuitBreaker    *CircuitBreaker    `json:"circuitBreaker,omitempty" yaml:"circuitBreaker,omitempty"`
	Compress          *Compress          `json:"compress,omitempty" yaml:"compress,omitempty"`
	ContentType       *ContentType       `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	DigestAuth        *DigestAuth        `json:"digestAuth,omitempty" yaml:"digestAuth,omitempty"`
	Errors            *ErrorPage         `json:"errors,omitempty" yaml:"errors,omitempty"`
	ForwardAuth       *ForwardAuth       `json:"forwardAuth,omitempty" yaml:"forwardAuth,omitempty"`
	Headers           *Headers           `json:"headers,omitempty" yaml:"headers,omitempty"`
	IPWhiteList       *IPWhiteList       `json:"ipWhiteList,omitempty" yaml:"ipWhiteList,omitempty"`
	InFlightReq       *InFlightReq       `json:"inFlightReq,omitempty" yaml:"inFlightReq,omitempty"`
	PassTLSClientCert *PassTLSClientCert `json:"passTLSClientCert,omitempty" yaml:"passTLSClientCert,omitempty"`
	RateLimit         *RateLimit         `json:"rateLimit,omitempty" yaml:"rateLimit,omitempty"`
	RedirectRegex     *RedirectRegex     `json:"redirectRegex,omitempty" yaml:"redirectRegex,omitempty"`
	RedirectScheme    *RedirectScheme    `json:"redirectScheme,omitempty" yaml:"redirectScheme,omitempty"`
	ReplacePath       *ReplacePath       `json:"replacePath,omitempty" yaml:"replacePath,omitempty"`
	ReplacePathRegex  *ReplacePathRegex  `json:"replacePathRegex,omitempty" yaml:"replacePathRegex,omitempty"`
	Retry             *Retry             `json:"retry,omitempty" yaml:"retry,omitempty"`
	StripPrefix       *StripPrefix       `json:"stripPrefix,omitempty" yaml:"stripPrefix,omitempty"`
	StripPrefixRegex  *StripPrefixRegex  `json:"stripPrefixRegex,omitempty" yaml:"stripPrefixRegex,omitempty"`
}

// Middleware configurations
type AddPrefix struct {
	Prefix string `json:"prefix" yaml:"prefix"`
}

type BasicAuth struct {
	Users        []string `json:"users,omitempty" yaml:"users,omitempty"`
	UsersFile    string   `json:"usersFile,omitempty" yaml:"usersFile,omitempty"`
	Realm        string   `json:"realm,omitempty" yaml:"realm,omitempty"`
	RemoveHeader bool     `json:"removeHeader,omitempty" yaml:"removeHeader,omitempty"`
	HeaderField  string   `json:"headerField,omitempty" yaml:"headerField,omitempty"`
}

type Buffering struct {
	MaxRequestBodyBytes  int64  `json:"maxRequestBodyBytes,omitempty" yaml:"maxRequestBodyBytes,omitempty"`
	MemRequestBodyBytes  int64  `json:"memRequestBodyBytes,omitempty" yaml:"memRequestBodyBytes,omitempty"`
	MaxResponseBodyBytes int64  `json:"maxResponseBodyBytes,omitempty" yaml:"maxResponseBodyBytes,omitempty"`
	MemResponseBodyBytes int64  `json:"memResponseBodyBytes,omitempty" yaml:"memResponseBodyBytes,omitempty"`
	RetryExpression      string `json:"retryExpression,omitempty" yaml:"retryExpression,omitempty"`
}

type Chain struct {
	Middlewares []string `json:"middlewares" yaml:"middlewares"`
}

type CircuitBreaker struct {
	Expression         string `json:"expression" yaml:"expression"`
	CheckPeriod        string `json:"checkPeriod,omitempty" yaml:"checkPeriod,omitempty"`
	FallbackDuration   string `json:"fallbackDuration,omitempty" yaml:"fallbackDuration,omitempty"`
	RecoveryDuration   string `json:"recoveryDuration,omitempty" yaml:"recoveryDuration,omitempty"`
	ResponseStatusCode int    `json:"responseStatusCode,omitempty" yaml:"responseStatusCode,omitempty"`
}

type Compress struct {
	ExcludedContentTypes []string `json:"excludedContentTypes,omitempty" yaml:"excludedContentTypes,omitempty"`
}

type ContentType struct {
	AutoDetect bool `json:"autoDetect,omitempty" yaml:"autoDetect,omitempty"`
}

type DigestAuth struct {
	Users        []string `json:"users,omitempty" yaml:"users,omitempty"`
	UsersFile    string   `json:"usersFile,omitempty" yaml:"usersFile,omitempty"`
	Realm        string   `json:"realm,omitempty" yaml:"realm,omitempty"`
	RemoveHeader bool     `json:"removeHeader,omitempty" yaml:"removeHeader,omitempty"`
	HeaderField  string   `json:"headerField,omitempty" yaml:"headerField,omitempty"`
}

type ErrorPage struct {
	Status  []string          `json:"status" yaml:"status"`
	Service string            `json:"service" yaml:"service"`
	Query   string            `json:"query,omitempty" yaml:"query,omitempty"`
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
}

type ForwardAuth struct {
	Address                  string            `json:"address" yaml:"address"`
	TLS                      *ForwardAuthTLS   `json:"tls,omitempty" yaml:"tls,omitempty"`
	TrustForwardHeader       bool              `json:"trustForwardHeader,omitempty" yaml:"trustForwardHeader,omitempty"`
	AuthResponseHeaders      []string          `json:"authResponseHeaders,omitempty" yaml:"authResponseHeaders,omitempty"`
	AuthResponseHeadersRegex string            `json:"authResponseHeadersRegex,omitempty" yaml:"authResponseHeadersRegex,omitempty"`
	AuthRequestHeaders       []string          `json:"authRequestHeaders,omitempty" yaml:"authRequestHeaders,omitempty"`
	AddAuthCookiesToResponse []string          `json:"addAuthCookiesToResponse,omitempty" yaml:"addAuthCookiesToResponse,omitempty"`
	Headers                  map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
}

type ForwardAuthTLS struct {
	Ca                 string `json:"ca,omitempty" yaml:"ca,omitempty"`
	CaOptional         bool   `json:"caOptional,omitempty" yaml:"caOptional,omitempty"`
	Cert               string `json:"cert,omitempty" yaml:"cert,omitempty"`
	Key                string `json:"key,omitempty" yaml:"key,omitempty"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify,omitempty" yaml:"insecureSkipVerify,omitempty"`
}

type Headers struct {
	CustomRequestHeaders          map[string]string `json:"customRequestHeaders,omitempty" yaml:"customRequestHeaders,omitempty"`
	CustomResponseHeaders         map[string]string `json:"customResponseHeaders,omitempty" yaml:"customResponseHeaders,omitempty"`
	AccessControlAllowCredentials bool              `json:"accessControlAllowCredentials,omitempty" yaml:"accessControlAllowCredentials,omitempty"`
	AccessControlAllowHeaders     []string          `json:"accessControlAllowHeaders,omitempty" yaml:"accessControlAllowHeaders,omitempty"`
	AccessControlAllowMethods     []string          `json:"accessControlAllowMethods,omitempty" yaml:"accessControlAllowMethods,omitempty"`
	AccessControlAllowOriginList  []string          `json:"accessControlAllowOriginList,omitempty" yaml:"accessControlAllowOriginList,omitempty"`
	AccessControlExposeHeaders    []string          `json:"accessControlExposeHeaders,omitempty" yaml:"accessControlExposeHeaders,omitempty"`
	AccessControlMaxAge           int64             `json:"accessControlMaxAge,omitempty" yaml:"accessControlMaxAge,omitempty"`
	AddVaryHeader                 bool              `json:"addVaryHeader,omitempty" yaml:"addVaryHeader,omitempty"`
	AllowedHosts                  []string          `json:"allowedHosts,omitempty" yaml:"allowedHosts,omitempty"`
	BrowserXssFilter              bool              `json:"browserXssFilter,omitempty" yaml:"browserXssFilter,omitempty"`
	ContentSecurityPolicy         string            `json:"contentSecurityPolicy,omitempty" yaml:"contentSecurityPolicy,omitempty"`
	ContentTypeNosniff            bool              `json:"contentTypeNosniff,omitempty" yaml:"contentTypeNosniff,omitempty"`
	CustomFrameOptionsValue       string            `json:"customFrameOptionsValue,omitempty" yaml:"customFrameOptionsValue,omitempty"`
	ForceSTSHeader                bool              `json:"forceSTSHeader,omitempty" yaml:"forceSTSHeader,omitempty"`
	FrameDeny                     bool              `json:"frameDeny,omitempty" yaml:"frameDeny,omitempty"`
	HostsProxyHeaders             []string          `json:"hostsProxyHeaders,omitempty" yaml:"hostsProxyHeaders,omitempty"`
	IsDevelopment                 bool              `json:"isDevelopment,omitempty" yaml:"isDevelopment,omitempty"`
	PublicKey                     string            `json:"publicKey,omitempty" yaml:"publicKey,omitempty"`
	ReferrerPolicy                string            `json:"referrerPolicy,omitempty" yaml:"referrerPolicy,omitempty"`
	SSLForceHost                  bool              `json:"sslForceHost,omitempty" yaml:"sslForceHost,omitempty"`
	SSLHost                       string            `json:"sslHost,omitempty" yaml:"sslHost,omitempty"`
	SSLProxyHeaders               map[string]string `json:"sslProxyHeaders,omitempty" yaml:"sslProxyHeaders,omitempty"`
	SSLRedirect                   bool              `json:"sslRedirect,omitempty" yaml:"sslRedirect,omitempty"`
	SSLTemporaryRedirect          bool              `json:"sslTemporaryRedirect,omitempty" yaml:"sslTemporaryRedirect,omitempty"`
	STSIncludeSubdomains          bool              `json:"stsIncludeSubdomains,omitempty" yaml:"stsIncludeSubdomains,omitempty"`
	STSPreload                    bool              `json:"stsPreload,omitempty" yaml:"stsPreload,omitempty"`
	STSSeconds                    int64             `json:"stsSeconds,omitempty" yaml:"stsSeconds,omitempty"`
}

type IPWhiteList struct {
	SourceRange []string          `json:"sourceRange" yaml:"sourceRange"`
	IPStrategy  *IPStrategy       `json:"ipStrategy,omitempty" yaml:"ipStrategy,omitempty"`
	Headers     map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
}

type IPStrategy struct {
	Depth       int      `json:"depth,omitempty" yaml:"depth,omitempty"`
	ExcludedIPs []string `json:"excludedIPs,omitempty" yaml:"excludedIPs,omitempty"`
}

type InFlightReq struct {
	Amount          int64            `json:"amount" yaml:"amount"`
	SourceCriterion *SourceCriterion `json:"sourceCriterion,omitempty" yaml:"sourceCriterion,omitempty"`
}

type SourceCriterion struct {
	IPStrategy        *IPStrategy `json:"ipStrategy,omitempty" yaml:"ipStrategy,omitempty"`
	RequestHeaderName string      `json:"requestHeaderName,omitempty" yaml:"requestHeaderName,omitempty"`
	RequestHost       bool        `json:"requestHost,omitempty" yaml:"requestHost,omitempty"`
}

type PassTLSClientCert struct {
	PEM  bool               `json:"pem,omitempty" yaml:"pem,omitempty"`
	Info *TLSClientCertInfo `json:"info,omitempty" yaml:"info,omitempty"`
}

type TLSClientCertInfo struct {
	NotAfter     bool                      `json:"notAfter,omitempty" yaml:"notAfter,omitempty"`
	NotBefore    bool                      `json:"notBefore,omitempty" yaml:"notBefore,omitempty"`
	Sans         bool                      `json:"sans,omitempty" yaml:"sans,omitempty"`
	Subject      *TLSCLientCertSubjectInfo `json:"subject,omitempty" yaml:"subject,omitempty"`
	Issuer       *TLSCLientCertIssuerInfo  `json:"issuer,omitempty" yaml:"issuer,omitempty"`
	SerialNumber bool                      `json:"serialNumber,omitempty" yaml:"serialNumber,omitempty"`
}

type TLSCLientCertSubjectInfo struct {
	Country            bool `json:"country,omitempty" yaml:"country,omitempty"`
	Province           bool `json:"province,omitempty" yaml:"province,omitempty"`
	Locality           bool `json:"locality,omitempty" yaml:"locality,omitempty"`
	Organization       bool `json:"organization,omitempty" yaml:"organization,omitempty"`
	OrganizationalUnit bool `json:"organizationalUnit,omitempty" yaml:"organizationalUnit,omitempty"`
	CommonName         bool `json:"commonName,omitempty" yaml:"commonName,omitempty"`
	SerialNumber       bool `json:"serialNumber,omitempty" yaml:"serialNumber,omitempty"`
	DomainComponent    bool `json:"domainComponent,omitempty" yaml:"domainComponent,omitempty"`
}

type TLSCLientCertIssuerInfo struct {
	Country            bool `json:"country,omitempty" yaml:"country,omitempty"`
	Province           bool `json:"province,omitempty" yaml:"province,omitempty"`
	Locality           bool `json:"locality,omitempty" yaml:"locality,omitempty"`
	Organization       bool `json:"organization,omitempty" yaml:"organization,omitempty"`
	OrganizationalUnit bool `json:"organizationalUnit,omitempty" yaml:"organizationalUnit,omitempty"`
	CommonName         bool `json:"commonName,omitempty" yaml:"commonName,omitempty"`
	SerialNumber       bool `json:"serialNumber,omitempty" yaml:"serialNumber,omitempty"`
	DomainComponent    bool `json:"domainComponent,omitempty" yaml:"domainComponent,omitempty"`
}

type RateLimit struct {
	Average         int64            `json:"average" yaml:"average"`
	Period          string           `json:"period,omitempty" yaml:"period,omitempty"`
	Burst           int64            `json:"burst,omitempty" yaml:"burst,omitempty"`
	SourceCriterion *SourceCriterion `json:"sourceCriterion,omitempty" yaml:"sourceCriterion,omitempty"`
}

type RedirectRegex struct {
	Regex       string `json:"regex" yaml:"regex"`
	Replacement string `json:"replacement" yaml:"replacement"`
	Permanent   bool   `json:"permanent,omitempty" yaml:"permanent,omitempty"`
}

type RedirectScheme struct {
	Scheme    string `json:"scheme" yaml:"scheme"`
	Port      string `json:"port,omitempty" yaml:"port,omitempty"`
	Permanent bool   `json:"permanent,omitempty" yaml:"permanent,omitempty"`
}

type ReplacePath struct {
	Path string `json:"path" yaml:"path"`
}

type ReplacePathRegex struct {
	Regex       string `json:"regex" yaml:"regex"`
	Replacement string `json:"replacement" yaml:"replacement"`
}

type Retry struct {
	Attempts        int    `json:"attempts" yaml:"attempts"`
	InitialInterval string `json:"initialInterval,omitempty" yaml:"initialInterval,omitempty"`
}

type StripPrefix struct {
	Prefixes   []string `json:"prefixes" yaml:"prefixes"`
	ForceSlash bool     `json:"forceSlash,omitempty" yaml:"forceSlash,omitempty"`
}

type StripPrefixRegex struct {
	Regex []string `json:"regex" yaml:"regex"`
}

// RegisterRoute registers a new route with Traefik middleware
func (ts *TraefikService) RegisterRoute(ctx context.Context, routeName string, route TraefikRoute) error {
	if !ts.config.Traefik.Enabled {
		log.Printf("Traefik is disabled, skipping route registration for %s", routeName)
		return nil
	}

	// Log route registration for debugging
	log.Printf("Registering Traefik route: %s with rule: %s", routeName, route.Rule)

	// In a real implementation, this would integrate with Traefik's API
	// For now, we'll return success to allow the system to work
	return nil
}

// CreateMiddlewareChain creates a chain of Traefik middlewares for the API Gateway
func (ts *TraefikService) CreateMiddlewareChain() []string {
	var middlewares []string

	// Add standard security middlewares
	if ts.config.Security.CORS.Enabled {
		middlewares = append(middlewares, "cors")
	}

	if ts.config.Security.RateLimit.Enabled {
		middlewares = append(middlewares, "rate-limit")
	}

	// Add authentication middleware
	if ts.config.Security.JWT.Secret != "" || ts.config.Security.JWKS.Endpoint != "" {
		middlewares = append(middlewares, "auth")
	}

	// Add circuit breaker
	middlewares = append(middlewares, "circuit-breaker")

	// Add request logging
	middlewares = append(middlewares, "request-logger")

	// Add compression
	middlewares = append(middlewares, "compress")

	return middlewares
}

// GetTraefikConfig generates Traefik configuration for the API Gateway
func (ts *TraefikService) GetTraefikConfig() map[string]interface{} {
	config := map[string]interface{}{
		"http": map[string]interface{}{
			"routers": map[string]interface{}{
				"api-gateway": map[string]interface{}{
					"rule":        fmt.Sprintf("Host(`%s`)", ts.getHostname()),
					"service":     "api-gateway",
					"entryPoints": []string{ts.config.Traefik.EntryPoint},
					"middlewares": ts.CreateMiddlewareChain(),
				},
			},
			"services": map[string]interface{}{
				"api-gateway": map[string]interface{}{
					"loadBalancer": map[string]interface{}{
						"servers": []map[string]interface{}{
							{
								"url": fmt.Sprintf("http://%s:%s", ts.config.Server.Host, ts.config.Server.Port),
							},
						},
						"healthCheck": map[string]interface{}{
							"path":               "/health",
							"interval":           "30s",
							"timeout":            "5s",
							"healthyStatusCodes": []int{200, 204},
							"followRedirects":    true,
						},
					},
				},
			},
			"middlewares": ts.getMiddlewareConfigs(),
		},
	}

	// Add TLS configuration if enabled
	if ts.config.Server.TLS.Enabled {
		config["http"].(map[string]interface{})["routers"].(map[string]interface{})["api-gateway"].(map[string]interface{})["tls"] = map[string]interface{}{
			"certResolver": "letsencrypt",
		}
	}

	return config
}

// getHostname returns the hostname for Traefik configuration
func (ts *TraefikService) getHostname() string {
	if host := ts.config.Server.Host; host != "" && host != "0.0.0.0" {
		return host
	}
	return "api-gateway.local"
}

// getMiddlewareConfigs returns middleware configurations for Traefik
func (ts *TraefikService) getMiddlewareConfigs() map[string]interface{} {
	middlewares := make(map[string]interface{})

	// CORS middleware
	if ts.config.Security.CORS.Enabled {
		middlewares["cors"] = map[string]interface{}{
			"headers": map[string]interface{}{
				"accessControlAllowMethods":     ts.config.Security.CORS.AllowedMethods,
				"accessControlAllowHeaders":     ts.config.Security.CORS.AllowedHeaders,
				"accessControlAllowOriginList":  ts.config.Security.CORS.AllowedOrigins,
				"accessControlAllowCredentials": ts.config.Security.CORS.AllowCredentials,
				"accessControlMaxAge":           ts.config.Security.CORS.MaxAge,
				"addVaryHeader":                 true,
			},
		}
	}

	// Rate limiting middleware
	if ts.config.Security.RateLimit.Enabled {
		middlewares["rate-limit"] = map[string]interface{}{
			"rateLimit": map[string]interface{}{
				"average": ts.config.Security.RateLimit.GlobalLimit,
				"burst":   ts.config.Security.RateLimit.GlobalLimit * 2,
				"period":  "1h",
			},
		}
	}

	// Security headers middleware
	middlewares["security-headers"] = map[string]interface{}{
		"headers": map[string]interface{}{
			"browserXssFilter":      true,
			"contentTypeNosniff":    true,
			"forceSTSHeader":        true,
			"frameDeny":             true,
			"stsIncludeSubdomains":  true,
			"stsPreload":            true,
			"stsSeconds":            31536000, // 1 year
			"contentSecurityPolicy": "default-src 'self'",
			"referrerPolicy":        "strict-origin-when-cross-origin",
			"customResponseHeaders": map[string]string{
				"X-API-Gateway": "X-Form-Backend",
				"X-Version":     ts.config.Version,
			},
		},
	}

	// Compression middleware
	middlewares["compress"] = map[string]interface{}{
		"compress": map[string]interface{}{
			"excludedContentTypes": []string{
				"text/event-stream",
				"application/grpc",
				"image/*",
				"video/*",
				"audio/*",
			},
		},
	}

	// Circuit breaker middleware
	middlewares["circuit-breaker"] = map[string]interface{}{
		"circuitBreaker": map[string]interface{}{
			"expression":         "NetworkErrorRatio() > 0.3 || ResponseCodeRatio(500, 600, 0, 600) > 0.3",
			"checkPeriod":        "10s",
			"fallbackDuration":   "30s",
			"recoveryDuration":   "30s",
			"responseStatusCode": 503,
		},
	}

	// Request ID middleware
	middlewares["request-id"] = map[string]interface{}{
		"headers": map[string]interface{}{
			"customRequestHeaders": map[string]string{
				"X-Request-ID": "{{ .UUID }}",
			},
		},
	}

	return middlewares
}

// TraefikMiddleware creates Gin middleware that integrates with Traefik
func (ts *TraefikService) TraefikMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Add Traefik-specific headers
		c.Header("X-Traefik-Gateway", "enabled")

		// Extract Traefik forwarded headers
		if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
			c.Set("original_ip", strings.Split(forwardedFor, ",")[0])
		}

		if forwardedProto := c.GetHeader("X-Forwarded-Proto"); forwardedProto != "" {
			c.Set("original_proto", forwardedProto)
		}

		if forwardedHost := c.GetHeader("X-Forwarded-Host"); forwardedHost != "" {
			c.Set("original_host", forwardedHost)
		}

		// Add request metadata for Traefik integration
		c.Set("traefik_enabled", ts.config.Traefik.Enabled)
		c.Set("entry_point", ts.config.Traefik.EntryPoint)

		c.Next()
	})
}

// HealthCheck provides health check endpoint for Traefik
func (ts *TraefikService) HealthCheck() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		status := map[string]interface{}{
			"status":  "healthy",
			"service": "x-form-api-gateway",
			"version": ts.config.Version,
			"time":    time.Now().UTC(),
			"traefik": map[string]interface{}{
				"enabled":     ts.config.Traefik.Enabled,
				"entry_point": ts.config.Traefik.EntryPoint,
			},
		}

		c.JSON(http.StatusOK, status)
	})
}
