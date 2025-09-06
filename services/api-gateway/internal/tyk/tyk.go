package tyk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/config"
	"github.com/gin-gonic/gin"
)

// TykService handles Tyk API Management integration
type TykService struct {
	config       *config.Config
	client       *http.Client
	dashboardURL string
	gatewayURL   string
	apiKey       string
	orgID        string
}

// NewTykService creates a new Tyk service instance
func NewTykService(cfg *config.Config) *TykService {
	return &TykService{
		config:       cfg,
		dashboardURL: cfg.Tyk.DashboardURL,
		gatewayURL:   cfg.Tyk.GatewayURL,
		apiKey:       cfg.Tyk.APIKey,
		orgID:        cfg.Tyk.OrgID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TykAPIDefinition represents a Tyk API definition
type TykAPIDefinition struct {
	ID                        string                 `json:"id,omitempty"`
	Name                      string                 `json:"name"`
	Slug                      string                 `json:"slug"`
	APIID                     string                 `json:"api_id"`
	ORGID                     string                 `json:"org_id"`
	UseKeylessAccess          bool                   `json:"use_keyless_access"`
	UseOauth2                 bool                   `json:"use_oauth2"`
	UseOpenID                 bool                   `json:"use_openid"`
	OpenIDOptions             *OpenIDOptions         `json:"openid_options,omitempty"`
	Oauth2Meta                *Oauth2Meta            `json:"oauth2_meta,omitempty"`
	Auth                      *AuthConfig            `json:"auth,omitempty"`
	UseBasicAuth              bool                   `json:"use_basic_auth"`
	BasicAuth                 *BasicAuthConfig       `json:"basic_auth,omitempty"`
	UseMutualTLS              bool                   `json:"use_mutual_tls"`
	ClientCertificates        []string               `json:"client_certificates,omitempty"`
	UpstreamCertificates      map[string]string      `json:"upstream_certificates,omitempty"`
	PinnedPublicKeys          map[string]string      `json:"pinned_public_keys,omitempty"`
	Proxy                     ProxyConfig            `json:"proxy"`
	DisableRateLimit          bool                   `json:"disable_rate_limit"`
	DisableQuota              bool                   `json:"disable_quota"`
	CustomMiddleware          *CustomMiddleware      `json:"custom_middleware,omitempty"`
	CustomMiddlewareBundle    string                 `json:"custom_middleware_bundle,omitempty"`
	CacheOptions              *CacheOptions          `json:"cache_options,omitempty"`
	SessionLifetime           int64                  `json:"session_lifetime"`
	Active                    bool                   `json:"active"`
	AuthProvider              *AuthProvider          `json:"auth_provider,omitempty"`
	SessionProvider           *SessionProvider       `json:"session_provider,omitempty"`
	EventHandlers             *EventHandlers         `json:"event_handlers,omitempty"`
	EnableBatchRequestSupport bool                   `json:"enable_batch_request_support"`
	EnableIPWhiteListing      bool                   `json:"enable_ip_whitelisting"`
	AllowedIPs                []string               `json:"allowed_ips,omitempty"`
	EnableIPBlacklisting      bool                   `json:"enable_ip_blacklisting"`
	BlacklistedIPs            []string               `json:"blacklisted_ips,omitempty"`
	DontSetQuotasOnCreate     bool                   `json:"dont_set_quotas_on_create"`
	ExpireAnalyticsAfter      int64                  `json:"expire_analytics_after"`
	ResponseProcessors        []ResponseProcessor    `json:"response_processors,omitempty"`
	CORS                      *CORSConfig            `json:"CORS,omitempty"`
	Domain                    string                 `json:"domain,omitempty"`
	DoNotTrack                bool                   `json:"do_not_track"`
	Tags                      []string               `json:"tags,omitempty"`
	EnableContextVars         bool                   `json:"enable_context_vars"`
	ConfigData                map[string]interface{} `json:"config_data,omitempty"`
	TagHeaders                []string               `json:"tag_headers,omitempty"`
	GlobalRateLimit           *GlobalRateLimit       `json:"global_rate_limit,omitempty"`
	StripListenPath           bool                   `json:"strip_listen_path"`
	VersionData               VersionData            `json:"version_data"`
}

// ProxyConfig defines the proxy configuration for Tyk
type ProxyConfig struct {
	PreserveHostHeader          bool              `json:"preserve_host_header"`
	ListenPath                  string            `json:"listen_path"`
	TargetURL                   string            `json:"target_url"`
	StripListenPath             bool              `json:"strip_listen_path"`
	EnableLoadBalancing         bool              `json:"enable_load_balancing"`
	Targets                     []string          `json:"targets,omitempty"`
	CheckHostAgainstUptimeTests bool              `json:"check_host_against_uptime_tests"`
	ServiceDiscovery            *ServiceDiscovery `json:"service_discovery,omitempty"`
	Transport                   *TransportConfig  `json:"transport,omitempty"`
}

// OpenIDOptions for OpenID Connect
type OpenIDOptions struct {
	Providers         []OIDProvider `json:"providers"`
	SegregateByClient bool          `json:"segregate_by_client"`
}

// OIDProvider represents an OpenID provider
type OIDProvider struct {
	Issuer    string            `json:"issuer"`
	ClientIDs map[string]string `json:"client_ids"`
}

// Oauth2Meta for OAuth2 configuration
type Oauth2Meta struct {
	AllowedAccessTypes     []string `json:"allowed_access_types"`
	AllowedAuthorizeTypes  []string `json:"allowed_authorize_types"`
	AuthorizeLoginRedirect string   `json:"auth_login_redirect"`
}

// AuthConfig for authentication configuration
type AuthConfig struct {
	UseParam          bool             `json:"use_param"`
	ParamName         string           `json:"param_name,omitempty"`
	UseCookie         bool             `json:"use_cookie"`
	CookieName        string           `json:"cookie_name,omitempty"`
	AuthHeaderName    string           `json:"auth_header_name,omitempty"`
	UseCertificate    bool             `json:"use_certificate"`
	ValidateSignature bool             `json:"validate_signature"`
	Signature         *SignatureConfig `json:"signature,omitempty"`
}

// SignatureConfig for signature validation
type SignatureConfig struct {
	Algorithm        string `json:"algorithm"`
	Header           string `json:"header"`
	Secret           string `json:"secret"`
	AllowedClockSkew int64  `json:"allowed_clock_skew"`
}

// BasicAuthConfig for basic authentication
type BasicAuthConfig struct {
	DisableCaching             bool                `json:"disable_caching"`
	CacheTTL                   int                 `json:"cache_ttl"`
	ExtractCredentialsFromBody *ExtractCredentials `json:"extract_from_body,omitempty"`
}

// ExtractCredentials for credential extraction
type ExtractCredentials struct {
	FormUserNameParameter string `json:"form_username_parameter"`
	FormPasswordParameter string `json:"form_password_parameter"`
}

// ServiceDiscovery configuration
type ServiceDiscovery struct {
	UseDiscoveryService bool   `json:"use_discovery_service"`
	QueryEndpoint       string `json:"query_endpoint"`
	UseNestedQuery      bool   `json:"use_nested_query"`
	ParentDataPath      string `json:"parent_data_path"`
	DataPath            string `json:"data_path"`
	PortDataPath        string `json:"port_data_path"`
	TargetPath          string `json:"target_path"`
	UseTargetList       bool   `json:"use_target_list"`
	CacheTimeout        int64  `json:"cache_timeout"`
	EndpointReturnsList bool   `json:"endpoint_returns_list"`
}

// TransportConfig for transport layer configuration
type TransportConfig struct {
	SSLInsecureSkipVerify   bool     `json:"ssl_insecure_skip_verify"`
	SSLCipherSuites         []string `json:"ssl_ciphers,omitempty"`
	SSLMinVersion           int      `json:"ssl_min_version"`
	SSLForceCommonNameCheck bool     `json:"ssl_force_common_name_check"`
	ProxyURL                string   `json:"proxy_url,omitempty"`
}

// CustomMiddleware configuration
type CustomMiddleware struct {
	Pre         []MiddlewareDefinition `json:"pre,omitempty"`
	Post        []MiddlewareDefinition `json:"post,omitempty"`
	PostKeyAuth []MiddlewareDefinition `json:"post_key_auth,omitempty"`
	AuthCheck   *MiddlewareDefinition  `json:"auth_check,omitempty"`
	Response    []MiddlewareDefinition `json:"response,omitempty"`
	Driver      string                 `json:"driver"`
	IdExtractor *IdExtractor           `json:"id_extractor,omitempty"`
}

// MiddlewareDefinition represents a middleware definition
type MiddlewareDefinition struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	RequireSession bool   `json:"require_session"`
	RawBodyOnly    bool   `json:"raw_body_only"`
}

// IdExtractor for ID extraction from requests
type IdExtractor struct {
	ExtractFrom     string            `json:"extract_from"`
	ExtractWith     string            `json:"extract_with"`
	ExtractorConfig map[string]string `json:"extractor_config,omitempty"`
}

// CacheOptions for response caching
type CacheOptions struct {
	CacheTimeout               int64    `json:"cache_timeout"`
	EnableCache                bool     `json:"enable_cache"`
	CacheAllSafeRequests       bool     `json:"cache_all_safe_requests"`
	CacheResponseCodes         []int    `json:"cache_response_codes,omitempty"`
	EnableUpstreamCacheControl bool     `json:"enable_upstream_cache_control"`
	CacheControlTTLHeader      string   `json:"cache_control_ttl_header,omitempty"`
	CacheByHeaders             []string `json:"cache_by_headers,omitempty"`
}

// AuthProvider configuration
type AuthProvider struct {
	Name          string                 `json:"name"`
	StorageEngine string                 `json:"storage_engine"`
	Meta          map[string]interface{} `json:"meta,omitempty"`
}

// SessionProvider configuration
type SessionProvider struct {
	Name          string                 `json:"name"`
	StorageEngine string                 `json:"storage_engine"`
	Meta          map[string]interface{} `json:"meta,omitempty"`
}

// EventHandlers for API events
type EventHandlers struct {
	Events map[string][]EventHandler `json:"events,omitempty"`
}

// EventHandler represents an event handler
type EventHandler struct {
	HandlerName string                 `json:"handler_name"`
	HandlerMeta map[string]interface{} `json:"handler_meta,omitempty"`
}

// ResponseProcessor for response processing
type ResponseProcessor struct {
	Name    string                 `json:"name"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// CORSConfig for CORS configuration
type CORSConfig struct {
	Enable             bool     `json:"enable"`
	AllowedOrigins     []string `json:"allowed_origins,omitempty"`
	AllowedMethods     []string `json:"allowed_methods,omitempty"`
	AllowedHeaders     []string `json:"allowed_headers,omitempty"`
	ExposedHeaders     []string `json:"exposed_headers,omitempty"`
	AllowCredentials   bool     `json:"allow_credentials"`
	MaxAge             int      `json:"max_age"`
	OptionsPassthrough bool     `json:"options_passthrough"`
	Debug              bool     `json:"debug"`
}

// GlobalRateLimit for global rate limiting
type GlobalRateLimit struct {
	Rate int64 `json:"rate"`
	Per  int64 `json:"per"`
}

// VersionData for API versioning
type VersionData struct {
	NotVersioned   bool                         `json:"not_versioned"`
	DefaultVersion string                       `json:"default_version,omitempty"`
	Versions       map[string]VersionDefinition `json:"versions,omitempty"`
}

// VersionDefinition represents a version definition
type VersionDefinition struct {
	Name                string            `json:"name"`
	Expires             string            `json:"expires,omitempty"`
	Paths               *VersionPaths     `json:"paths,omitempty"`
	UseExtendedPaths    bool              `json:"use_extended_paths"`
	ExtendedPaths       *ExtendedPathsSet `json:"extended_paths,omitempty"`
	GlobalHeaders       map[string]string `json:"global_headers,omitempty"`
	GlobalHeadersRemove []string          `json:"global_headers_remove,omitempty"`
	GlobalSizeLimit     int64             `json:"global_size_limit"`
	OverrideTarget      string            `json:"override_target,omitempty"`
}

// VersionPaths for version path configuration
type VersionPaths struct {
	Ignored   []string `json:"ignored,omitempty"`
	WhiteList []string `json:"white_list,omitempty"`
	BlackList []string `json:"black_list,omitempty"`
}

// ExtendedPathsSet for extended path configurations
type ExtendedPathsSet struct {
	Ignored             []ExtendedPath        `json:"ignored,omitempty"`
	WhiteList           []ExtendedPath        `json:"white_list,omitempty"`
	BlackList           []ExtendedPath        `json:"black_list,omitempty"`
	Cached              []string              `json:"cached,omitempty"`
	Transform           []TemplateMeta        `json:"transform,omitempty"`
	TransformResponse   []TemplateMeta        `json:"transform_response,omitempty"`
	TransformJQ         []TransformJQMeta     `json:"transform_jq,omitempty"`
	TransformJQResponse []TransformJQMeta     `json:"transform_jq_response,omitempty"`
	HardTimeouts        []HardTimeoutMeta     `json:"hard_timeouts,omitempty"`
	CircuitBreakers     []CircuitBreakerMeta  `json:"circuit_breakers,omitempty"`
	URLRewrites         []URLRewriteMeta      `json:"url_rewrites,omitempty"`
	VirtualEndpoints    []VirtualMeta         `json:"virtual,omitempty"`
	SizeLimit           []RequestSizeMeta     `json:"size_limits,omitempty"`
	MethodTransforms    []MethodTransformMeta `json:"method_transforms,omitempty"`
	TrackEndpoints      []TrackEndpointMeta   `json:"track_endpoints,omitempty"`
	DoNotTrackEndpoints []TrackEndpointMeta   `json:"do_not_track_endpoints,omitempty"`
	ValidateJSON        []ValidatePathMeta    `json:"validate_json,omitempty"`
	InternalEndpoints   []InternalMeta        `json:"internal,omitempty"`
}

// ExtendedPath represents an extended path configuration
type ExtendedPath struct {
	Path         string `json:"path"`
	Method       string `json:"method,omitempty"`
	IgnoreCase   bool   `json:"ignore_case,omitempty"`
	MatchPattern string `json:"match_pattern,omitempty"`
}

// TemplateMeta for request/response transformation
type TemplateMeta struct {
	Path         string       `json:"path"`
	Method       string       `json:"method"`
	TemplateData TemplateData `json:"template_data"`
}

// TemplateData for transformation templates
type TemplateData struct {
	TemplateMode   string `json:"template_mode"`
	Template       string `json:"template"`
	EnableSession  bool   `json:"enable_session"`
	InputType      string `json:"input_type"`
	TemplateSource string `json:"template_source"`
}

// TransformJQMeta for JQ transformations
type TransformJQMeta struct {
	Path   string `json:"path"`
	Method string `json:"method"`
	Filter string `json:"filter"`
}

// HardTimeoutMeta for hard timeouts
type HardTimeoutMeta struct {
	Path    string `json:"path"`
	Method  string `json:"method"`
	Timeout int    `json:"timeout"`
}

// CircuitBreakerMeta for circuit breaker configuration
type CircuitBreakerMeta struct {
	Path                 string  `json:"path"`
	Method               string  `json:"method"`
	ThresholdPercent     float64 `json:"threshold_percent"`
	Samples              int64   `json:"samples"`
	ReturnToServiceAfter int     `json:"return_to_service_after"`
}

// URLRewriteMeta for URL rewriting
type URLRewriteMeta struct {
	Path         string           `json:"path"`
	Method       string           `json:"method"`
	MatchPattern string           `json:"match_pattern"`
	RewriteTo    string           `json:"rewrite_to"`
	Triggers     []RoutingTrigger `json:"triggers,omitempty"`
	RewriteRaw   bool             `json:"rewrite_raw"`
}

// RoutingTrigger for routing triggers
type RoutingTrigger struct {
	On        string                `json:"on"`
	Options   RoutingTriggerOptions `json:"options"`
	RewriteTo string                `json:"rewrite_to"`
}

// RoutingTriggerOptions for routing trigger options
type RoutingTriggerOptions struct {
	HeaderMatches         map[string]RoutingTriggerHeaderOptions  `json:"header_matches,omitempty"`
	QueryValMatches       map[string]RoutingTriggerQueryOptions   `json:"query_val_matches,omitempty"`
	PathPartMatches       map[string]RoutingTriggerPathOptions    `json:"path_part_matches,omitempty"`
	SessionMetaMatches    map[string]RoutingTriggerSessionOptions `json:"session_meta_matches,omitempty"`
	RequestContextMatches map[string]RoutingTriggerContextOptions `json:"request_context_matches,omitempty"`
}

// RoutingTriggerHeaderOptions for header-based routing
type RoutingTriggerHeaderOptions struct {
	MatchRx string `json:"match_rx"`
	Reverse bool   `json:"reverse"`
}

// RoutingTriggerQueryOptions for query-based routing
type RoutingTriggerQueryOptions struct {
	MatchRx string `json:"match_rx"`
	Reverse bool   `json:"reverse"`
}

// RoutingTriggerPathOptions for path-based routing
type RoutingTriggerPathOptions struct {
	MatchRx string `json:"match_rx"`
	Reverse bool   `json:"reverse"`
}

// RoutingTriggerSessionOptions for session-based routing
type RoutingTriggerSessionOptions struct {
	MatchRx string `json:"match_rx"`
	Reverse bool   `json:"reverse"`
}

// RoutingTriggerContextOptions for context-based routing
type RoutingTriggerContextOptions struct {
	MatchRx string `json:"match_rx"`
	Reverse bool   `json:"reverse"`
}

// VirtualMeta for virtual endpoints
type VirtualMeta struct {
	ResponseFunctionName string `json:"response_function_name"`
	FunctionSourceType   string `json:"function_source_type"`
	FunctionSourceURI    string `json:"function_source_uri"`
	Path                 string `json:"path"`
	Method               string `json:"method"`
	UseSession           bool   `json:"use_session"`
	ProxyOnError         bool   `json:"proxy_on_error"`
}

// RequestSizeMeta for request size limits
type RequestSizeMeta struct {
	Path      string `json:"path"`
	Method    string `json:"method"`
	SizeLimit int64  `json:"size_limit"`
}

// MethodTransformMeta for HTTP method transformation
type MethodTransformMeta struct {
	Path     string `json:"path"`
	Method   string `json:"method"`
	ToMethod string `json:"to_method"`
}

// TrackEndpointMeta for endpoint tracking
type TrackEndpointMeta struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

// ValidatePathMeta for JSON validation
type ValidatePathMeta struct {
	Path              string                 `json:"path"`
	Method            string                 `json:"method"`
	Schema            map[string]interface{} `json:"schema"`
	ErrorResponseCode int                    `json:"error_response_code"`
}

// InternalMeta for internal endpoints
type InternalMeta struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

// TykKeyDefinition represents a Tyk API key definition
type TykKeyDefinition struct {
	AllowedIPs            []string               `json:"allowed_ips,omitempty"`
	APIModel              map[string]string      `json:"api_model,omitempty"`
	APIAccess             []AccessDefinition     `json:"access_rights"`
	BasicAuthData         *BasicAuthData         `json:"basic_auth_data,omitempty"`
	HMACEnabled           bool                   `json:"hmac_enabled"`
	HmacSecret            string                 `json:"hmac_secret,omitempty"`
	IsInactive            bool                   `json:"is_inactive"`
	ApplyPolicyID         string                 `json:"apply_policy_id,omitempty"`
	ApplyPolicies         []string               `json:"apply_policies,omitempty"`
	DataExpires           int64                  `json:"data_expires"`
	MonitorQuotaReaches   bool                   `json:"monitor"`
	EnableDetailRecording bool                   `json:"enable_detail_recording"`
	MetaData              map[string]interface{} `json:"meta_data,omitempty"`
	OrgID                 string                 `json:"org_id"`
	OauthClientID         string                 `json:"oauth_client_id,omitempty"`
	OAuthKeys             map[string]string      `json:"oauth_keys,omitempty"`
	PartitionByID         string                 `json:"partition_by_id,omitempty"`
	QuotaMax              int64                  `json:"quota_max"`
	QuotaRenewalRate      int64                  `json:"quota_renewal_rate"`
	QuotaRemaining        int64                  `json:"quota_remaining"`
	QuotaRenews           int64                  `json:"quota_renews"`
	Rate                  float64                `json:"rate"`
	Per                   int64                  `json:"per"`
	ThrottleInterval      float64                `json:"throttle_interval"`
	ThrottleRetryLimit    int                    `json:"throttle_retry_limit"`
	MaxQueryDepth         int                    `json:"max_query_depth"`
	DateCreated           time.Time              `json:"date_created,omitempty"`
	Tags                  []string               `json:"tags,omitempty"`
	Alias                 string                 `json:"alias,omitempty"`
	LastUpdated           string                 `json:"last_updated,omitempty"`
	IDExtractorDeadLine   int64                  `json:"id_extractor_deadline"`
	SessionState          *SessionState          `json:"session_state,omitempty"`
}

// AccessDefinition defines API access rights
type AccessDefinition struct {
	APIID          string       `json:"api_id"`
	APIName        string       `json:"api_name"`
	Versions       []string     `json:"versions,omitempty"`
	AllowedURLs    []AccessSpec `json:"allowed_urls,omitempty"`
	Limit          *APILimit    `json:"limit,omitempty"`
	AllowanceScope string       `json:"allowance_scope,omitempty"`
}

// AccessSpec defines URL access specifications
type AccessSpec struct {
	URL     string   `json:"url"`
	Methods []string `json:"methods"`
}

// APILimit defines API usage limits
type APILimit struct {
	Rate               float64 `json:"rate"`
	Per                int64   `json:"per"`
	ThrottleInterval   float64 `json:"throttle_interval"`
	ThrottleRetryLimit int     `json:"throttle_retry_limit"`
	MaxQueryDepth      int     `json:"max_query_depth"`
	QuotaMax           int64   `json:"quota_max"`
	QuotaRenewalRate   int64   `json:"quota_renewal_rate"`
	QuotaRenews        int64   `json:"quota_renews"`
	QuotaRemaining     int64   `json:"quota_remaining"`
}

// BasicAuthData for basic authentication
type BasicAuthData struct {
	Password string `json:"password"`
	Hash     string `json:"hash,omitempty"`
}

// SessionState represents session state
type SessionState struct {
	Rate                  float64                     `json:"rate"`
	Per                   int64                       `json:"per"`
	ThrottleInterval      float64                     `json:"throttle_interval"`
	ThrottleRetryLimit    int                         `json:"throttle_retry_limit"`
	MaxQueryDepth         int                         `json:"max_query_depth"`
	QuotaMax              int64                       `json:"quota_max"`
	QuotaRenewalRate      int64                       `json:"quota_renewal_rate"`
	QuotaRenews           int64                       `json:"quota_renews"`
	QuotaRemaining        int64                       `json:"quota_remaining"`
	AccessRights          map[string]AccessDefinition `json:"access_rights"`
	OrgID                 string                      `json:"org_id"`
	OauthClientID         string                      `json:"oauth_client_id,omitempty"`
	OauthKeys             map[string]string           `json:"oauth_keys,omitempty"`
	BasicAuthData         *BasicAuthData              `json:"basic_auth_data,omitempty"`
	JWTData               *JWTData                    `json:"jwt_data,omitempty"`
	HMACEnabled           bool                        `json:"hmac_enabled"`
	HmacSecret            string                      `json:"hmac_secret,omitempty"`
	IsInactive            bool                        `json:"is_inactive"`
	ApplyPolicyID         string                      `json:"apply_policy_id,omitempty"`
	DataExpires           int64                       `json:"data_expires"`
	Monitor               bool                        `json:"monitor"`
	EnableDetailRecording bool                        `json:"enable_detail_recording"`
	MetaData              map[string]interface{}      `json:"meta_data,omitempty"`
	Tags                  []string                    `json:"tags,omitempty"`
	Alias                 string                      `json:"alias,omitempty"`
	LastUpdated           string                      `json:"last_updated,omitempty"`
	IDExtractorDeadLine   int64                       `json:"id_extractor_deadline"`
}

// JWTData for JWT configuration
type JWTData struct {
	Secret string `json:"secret"`
}

// TykPolicy represents a Tyk security policy
type TykPolicy struct {
	ID                   string                      `json:"_id,omitempty"`
	Name                 string                      `json:"name"`
	Rate                 float64                     `json:"rate"`
	Per                  int64                       `json:"per"`
	QuotaMax             int64                       `json:"quota_max"`
	QuotaRenewalRate     int64                       `json:"quota_renewal_rate"`
	AccessRights         map[string]AccessDefinition `json:"access_rights"`
	HMACEnabled          bool                        `json:"hmac_enabled"`
	Active               bool                        `json:"active"`
	IsInactive           bool                        `json:"is_inactive"`
	Tags                 []string                    `json:"tags,omitempty"`
	KeyExpiresIn         int64                       `json:"key_expires_in"`
	Partitions           Partitions                  `json:"partitions"`
	ThrottleInterval     float64                     `json:"throttle_interval"`
	ThrottleRetryLimit   int                         `json:"throttle_retry_limit"`
	MaxQueryDepth        int                         `json:"max_query_depth"`
	GraphQLIntrospection *GraphQLIntrospection       `json:"graphql_introspection,omitempty"`
	OrgID                string                      `json:"org_id"`
	MetaData             map[string]interface{}      `json:"meta_data,omitempty"`
	LastUpdated          string                      `json:"last_updated,omitempty"`
}

// Partitions for policy partitioning
type Partitions struct {
	Quota      bool `json:"quota"`
	RateLimit  bool `json:"rate_limit"`
	Complexity bool `json:"complexity"`
	Acl        bool `json:"acl"`
}

// GraphQLIntrospection for GraphQL APIs
type GraphQLIntrospection struct {
	Disabled bool `json:"disabled"`
}

// CreateAPIDefinition creates a new API definition in Tyk
func (ts *TykService) CreateAPIDefinition(ctx context.Context, apiDef *TykAPIDefinition) error {
	if !ts.config.Tyk.Enabled {
		log.Printf("Tyk is disabled, skipping API definition creation for %s", apiDef.Name)
		return nil
	}

	// Set default values
	apiDef.ORGID = ts.orgID
	apiDef.Active = true

	jsonData, err := json.Marshal(apiDef)
	if err != nil {
		return fmt.Errorf("failed to marshal API definition: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ts.dashboardURL+"/api/apis", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", ts.apiKey)

	resp, err := ts.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API definition creation failed: %d - %s", resp.StatusCode, string(body))
	}

	log.Printf("Successfully created API definition: %s", apiDef.Name)
	return nil
}

// CreateAPIKey creates a new API key in Tyk
func (ts *TykService) CreateAPIKey(ctx context.Context, keyDef *TykKeyDefinition) (string, error) {
	if !ts.config.Tyk.Enabled {
		log.Printf("Tyk is disabled, skipping API key creation")
		return "disabled", nil
	}

	// Set default values
	keyDef.OrgID = ts.orgID

	jsonData, err := json.Marshal(keyDef)
	if err != nil {
		return "", fmt.Errorf("failed to marshal key definition: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ts.dashboardURL+"/api/keys", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", ts.apiKey)

	resp, err := ts.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API key creation failed: %d - %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	keyID, ok := result["key_id"].(string)
	if !ok {
		return "", fmt.Errorf("key_id not found in response")
	}

	log.Printf("Successfully created API key: %s", keyID)
	return keyID, nil
}

// CreatePolicy creates a new security policy in Tyk
func (ts *TykService) CreatePolicy(ctx context.Context, policy *TykPolicy) (string, error) {
	if !ts.config.Tyk.Enabled {
		log.Printf("Tyk is disabled, skipping policy creation for %s", policy.Name)
		return "disabled", nil
	}

	// Set default values
	policy.OrgID = ts.orgID
	policy.Active = true

	jsonData, err := json.Marshal(policy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal policy: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ts.dashboardURL+"/api/portal/policies", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", ts.apiKey)

	resp, err := ts.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("policy creation failed: %d - %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	policyID, ok := result["Message"].(string)
	if !ok {
		return "", fmt.Errorf("policy ID not found in response")
	}

	log.Printf("Successfully created policy: %s", policy.Name)
	return policyID, nil
}

// GetAPIDefinitionForService creates a Tyk API definition for a service
func (ts *TykService) GetAPIDefinitionForService(serviceName, serviceURL, listenPath string) *TykAPIDefinition {
	return &TykAPIDefinition{
		Name:             fmt.Sprintf("X-Form %s Service", serviceName),
		Slug:             fmt.Sprintf("x-form-%s", serviceName),
		APIID:            fmt.Sprintf("x-form-%s-api", serviceName),
		UseKeylessAccess: false,
		UseOauth2:        false,
		UseOpenID:        true,
		OpenIDOptions: &OpenIDOptions{
			Providers: []OIDProvider{
				{
					Issuer: ts.config.Security.JWT.Issuer,
					ClientIDs: map[string]string{
						"x-form-client": ts.config.Security.JWT.Audience,
					},
				},
			},
		},
		Auth: &AuthConfig{
			UseParam:       false,
			UseCookie:      false,
			AuthHeaderName: "Authorization",
			UseCertificate: ts.config.Security.MTLS.Enabled,
		},
		Proxy: ProxyConfig{
			PreserveHostHeader: true,
			ListenPath:         listenPath,
			TargetURL:          serviceURL,
			StripListenPath:    true,
			Transport: &TransportConfig{
				SSLInsecureSkipVerify: false,
				SSLMinVersion:         771, // TLS 1.2
			},
		},
		DisableRateLimit: false,
		DisableQuota:     false,
		SessionLifetime:  int64(ts.config.Security.JWT.ExpirationTime.Seconds()),
		CORS: &CORSConfig{
			Enable:           ts.config.Security.CORS.Enabled,
			AllowedOrigins:   ts.config.Security.CORS.AllowedOrigins,
			AllowedMethods:   ts.config.Security.CORS.AllowedMethods,
			AllowedHeaders:   ts.config.Security.CORS.AllowedHeaders,
			AllowCredentials: ts.config.Security.CORS.AllowCredentials,
			MaxAge:           ts.config.Security.CORS.MaxAge,
		},
		GlobalRateLimit: &GlobalRateLimit{
			Rate: int64(ts.config.Security.RateLimit.GlobalLimit),
			Per:  3600, // 1 hour
		},
		StripListenPath: true,
		VersionData: VersionData{
			NotVersioned: true,
		},
		Tags: []string{"x-form", serviceName, "microservice"},
	}
}

// GetDefaultPolicy creates a default security policy for X-Form services
func (ts *TykService) GetDefaultPolicy(name string, accessRights map[string]AccessDefinition) *TykPolicy {
	return &TykPolicy{
		Name:             name,
		Rate:             float64(ts.config.Security.RateLimit.PerUserLimit),
		Per:              3600,                                                  // 1 hour
		QuotaMax:         int64(ts.config.Security.RateLimit.PerUserLimit * 24), // Daily quota
		QuotaRenewalRate: 86400,                                                 // 24 hours
		AccessRights:     accessRights,
		Active:           true,
		KeyExpiresIn:     int64(ts.config.Security.JWT.ExpirationTime.Seconds()),
		Tags:             []string{"x-form", "default"},
		Partitions: Partitions{
			Quota:      true,
			RateLimit:  true,
			Complexity: false,
			Acl:        true,
		},
	}
}

// TykMiddleware creates Gin middleware that integrates with Tyk
func (ts *TykService) TykMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Add Tyk-specific headers
		c.Header("X-Tyk-Gateway", "enabled")

		// Extract Tyk headers if present
		if apiKey := c.GetHeader("X-Tyk-Api-Key"); apiKey != "" {
			c.Set("tyk_api_key", apiKey)
		}

		if policyID := c.GetHeader("X-Tyk-Policy-Id"); policyID != "" {
			c.Set("tyk_policy_id", policyID)
		}

		// Add metadata for Tyk integration
		c.Set("tyk_enabled", ts.config.Tyk.Enabled)
		c.Set("tyk_gateway_url", ts.gatewayURL)

		c.Next()
	})
}

// ReloadAPIDefinitions reloads API definitions in Tyk Gateway
func (ts *TykService) ReloadAPIDefinitions(ctx context.Context) error {
	if !ts.config.Tyk.Enabled {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", ts.gatewayURL+"/tyk/reload", nil)
	if err != nil {
		return fmt.Errorf("failed to create reload request: %w", err)
	}

	req.Header.Set("X-Tyk-Authorization", ts.apiKey)

	resp, err := ts.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reload APIs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API reload failed: %d - %s", resp.StatusCode, string(body))
	}

	log.Printf("Successfully reloaded Tyk API definitions")
	return nil
}
