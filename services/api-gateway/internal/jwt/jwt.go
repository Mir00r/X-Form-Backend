package jwt

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTService handles JWT validation and JWKS integration
type JWTService struct {
	config   *config.Config
	keyCache *KeyCache
	client   *http.Client
}

// KeyCache stores JWKS keys with cache management
type KeyCache struct {
	keys      map[string]*rsa.PublicKey
	mutex     sync.RWMutex
	lastFetch time.Time
	jwksURL   string
}

// JWKS represents JSON Web Key Set structure
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kty string   `json:"kty"`           // Key Type
	Use string   `json:"use"`           // Public Key Use
	Kid string   `json:"kid"`           // Key ID
	Alg string   `json:"alg"`           // Algorithm
	N   string   `json:"n"`             // RSA Modulus
	E   string   `json:"e"`             // RSA Exponent
	X5c []string `json:"x5c,omitempty"` // X.509 Certificate Chain
	X5t string   `json:"x5t,omitempty"` // X.509 Certificate SHA-1 Thumbprint
}

// CustomClaims extends jwt.RegisteredClaims with application-specific claims
type CustomClaims struct {
	UserID      string                 `json:"user_id,omitempty"`
	Username    string                 `json:"username,omitempty"`
	Email       string                 `json:"email,omitempty"`
	Roles       []string               `json:"roles,omitempty"`
	Permissions []string               `json:"permissions,omitempty"`
	Tenant      string                 `json:"tenant,omitempty"`
	Scope       string                 `json:"scope,omitempty"`
	ClientID    string                 `json:"client_id,omitempty"`
	TokenType   string                 `json:"token_type,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	jwt.RegisteredClaims
}

// ValidationResult represents JWT validation result
type ValidationResult struct {
	Valid     bool          `json:"valid"`
	Claims    *CustomClaims `json:"claims,omitempty"`
	Error     string        `json:"error,omitempty"`
	KeyID     string        `json:"key_id,omitempty"`
	Algorithm string        `json:"algorithm,omitempty"`
}

// NewJWTService creates a new JWT service instance
func NewJWTService(cfg *config.Config) *JWTService {
	service := &JWTService{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		keyCache: &KeyCache{
			keys:    make(map[string]*rsa.PublicKey),
			jwksURL: cfg.Security.JWKS.Endpoint,
		},
	}

	// Start background key refresh if JWKS is enabled
	if cfg.Security.JWKS.Endpoint != "" {
		go service.startKeyRefresh()
	}

	return service
}

// ValidateToken validates a JWT token using either JWKS or shared secret
func (js *JWTService) ValidateToken(tokenString string) *ValidationResult {
	// Remove Bearer prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Parse token to get header information
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return js.getValidationKey(token)
	})

	result := &ValidationResult{}

	if err != nil {
		result.Error = fmt.Sprintf("token validation failed: %v", err)
		return result
	}

	if !token.Valid {
		result.Error = "token is invalid"
		return result
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		result.Error = "invalid token claims"
		return result
	}

	// Additional validation
	if err := js.validateClaims(claims); err != nil {
		result.Error = fmt.Sprintf("claims validation failed: %v", err)
		return result
	}

	result.Valid = true
	result.Claims = claims

	// Extract key ID and algorithm from token header
	if header, ok := token.Header["kid"]; ok {
		result.KeyID = header.(string)
	}
	if alg, ok := token.Header["alg"]; ok {
		result.Algorithm = alg.(string)
	}

	return result
}

// getValidationKey returns the key for token validation
func (js *JWTService) getValidationKey(token *jwt.Token) (interface{}, error) {
	// Check algorithm
	if _, ok := token.Method.(*jwt.SigningMethodRSA); ok {
		// RSA key - use JWKS
		return js.getRSAKey(token)
	} else if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
		// HMAC key - use shared secret
		return []byte(js.config.Security.JWT.Secret), nil
	}

	return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
}

// getRSAKey retrieves RSA public key from JWKS
func (js *JWTService) getRSAKey(token *jwt.Token) (*rsa.PublicKey, error) {
	if js.config.Security.JWKS.Endpoint == "" {
		return nil, fmt.Errorf("JWKS endpoint not configured")
	}

	// Get key ID from token header
	kidInterface, ok := token.Header["kid"]
	if !ok {
		return nil, fmt.Errorf("token header missing kid")
	}

	kid, ok := kidInterface.(string)
	if !ok {
		return nil, fmt.Errorf("kid header is not a string")
	}

	// Get key from cache
	key, err := js.keyCache.getKey(kid)
	if err != nil {
		// Refresh keys and try again
		if refreshErr := js.refreshKeys(); refreshErr != nil {
			return nil, fmt.Errorf("failed to refresh keys: %v", refreshErr)
		}
		key, err = js.keyCache.getKey(kid)
		if err != nil {
			return nil, fmt.Errorf("key not found after refresh: %v", err)
		}
	}

	return key, nil
}

// validateClaims performs additional claims validation
func (js *JWTService) validateClaims(claims *CustomClaims) error {
	// Validate issuer
	if js.config.Security.JWT.Issuer != "" && claims.Issuer != js.config.Security.JWT.Issuer {
		return fmt.Errorf("invalid issuer: expected %s, got %s", js.config.Security.JWT.Issuer, claims.Issuer)
	}

	// Validate audience
	if js.config.Security.JWT.Audience != "" {
		validAudience := false
		for _, aud := range claims.Audience {
			if aud == js.config.Security.JWT.Audience {
				validAudience = true
				break
			}
		}
		if !validAudience {
			return fmt.Errorf("invalid audience: expected %s", js.config.Security.JWT.Audience)
		}
	}

	// Validate expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("token has expired")
	}

	// Validate not before
	if claims.NotBefore != nil && claims.NotBefore.After(time.Now()) {
		return fmt.Errorf("token not yet valid")
	}

	// Validate issued at (with clock skew tolerance)
	if claims.IssuedAt != nil && claims.IssuedAt.After(time.Now().Add(5*time.Minute)) {
		return fmt.Errorf("token issued in the future")
	}

	return nil
}

// refreshKeys fetches fresh keys from JWKS endpoint
func (js *JWTService) refreshKeys() error {
	if js.config.Security.JWKS.Endpoint == "" {
		return fmt.Errorf("JWKS endpoint not configured")
	}

	resp, err := js.client.Get(js.config.Security.JWKS.Endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status: %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %v", err)
	}

	return js.keyCache.updateKeys(jwks.Keys)
}

// startKeyRefresh starts background key refresh routine
func (js *JWTService) startKeyRefresh() {
	// Initial fetch
	if err := js.refreshKeys(); err != nil {
		log.Printf("Initial JWKS fetch failed: %v", err)
	}

	// Periodic refresh
	ticker := time.NewTicker(js.config.Security.JWKS.RefreshInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := js.refreshKeys(); err != nil {
			log.Printf("JWKS refresh failed: %v", err)
		} else {
			log.Printf("JWKS keys refreshed successfully")
		}
	}
}

// getKey retrieves a key from cache
func (kc *KeyCache) getKey(kid string) (*rsa.PublicKey, error) {
	kc.mutex.RLock()
	defer kc.mutex.RUnlock()

	key, exists := kc.keys[kid]
	if !exists {
		return nil, fmt.Errorf("key with ID %s not found", kid)
	}

	return key, nil
}

// updateKeys updates the key cache with new keys
func (kc *KeyCache) updateKeys(jwkKeys []JWK) error {
	kc.mutex.Lock()
	defer kc.mutex.Unlock()

	newKeys := make(map[string]*rsa.PublicKey)

	for _, jwk := range jwkKeys {
		if jwk.Kty != "RSA" {
			log.Printf("Skipping non-RSA key: %s", jwk.Kid)
			continue
		}

		pubKey, err := jwkToRSAPublicKey(jwk)
		if err != nil {
			log.Printf("Failed to convert JWK to RSA public key for kid %s: %v", jwk.Kid, err)
			continue
		}

		newKeys[jwk.Kid] = pubKey
	}

	kc.keys = newKeys
	kc.lastFetch = time.Now()

	log.Printf("Updated key cache with %d keys", len(newKeys))
	return nil
}

// jwkToRSAPublicKey converts a JWK to RSA public key
func jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// This is a simplified implementation
	// In production, you would use a proper JWK library like github.com/lestrrat-go/jwx
	return nil, fmt.Errorf("JWK to RSA conversion not implemented - use proper JWK library")
}

// JWTMiddleware creates Gin middleware for JWT validation
func (js *JWTService) JWTMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Skip authentication for health check and public endpoints
		if js.isPublicEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "MISSING_TOKEN",
			})
			c.Abort()
			return
		}

		// Validate token
		result := js.ValidateToken(authHeader)
		if !result.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": result.Error,
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Add claims to context
		c.Set("jwt_claims", result.Claims)
		c.Set("user_id", result.Claims.UserID)
		c.Set("username", result.Claims.Username)
		c.Set("email", result.Claims.Email)
		c.Set("roles", result.Claims.Roles)
		c.Set("permissions", result.Claims.Permissions)
		c.Set("tenant", result.Claims.Tenant)

		// Add validation metadata
		c.Set("jwt_key_id", result.KeyID)
		c.Set("jwt_algorithm", result.Algorithm)

		c.Next()
	})
}

// isPublicEndpoint checks if an endpoint should skip authentication
func (js *JWTService) isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/health",
		"/metrics",
		"/swagger",
		"/docs",
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/refresh",
		"/api/v1/public",
	}

	for _, publicPath := range publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
	}

	return false
}

// ExtractClaims extracts JWT claims from Gin context
func ExtractClaims(c *gin.Context) (*CustomClaims, error) {
	claims, exists := c.Get("jwt_claims")
	if !exists {
		return nil, fmt.Errorf("JWT claims not found in context")
	}

	customClaims, ok := claims.(*CustomClaims)
	if !ok {
		return nil, fmt.Errorf("invalid JWT claims type")
	}

	return customClaims, nil
}

// HasRole checks if the user has a specific role
func HasRole(c *gin.Context, role string) bool {
	claims, err := ExtractClaims(c)
	if err != nil {
		return false
	}

	for _, userRole := range claims.Roles {
		if userRole == role {
			return true
		}
	}

	return false
}

// HasPermission checks if the user has a specific permission
func HasPermission(c *gin.Context, permission string) bool {
	claims, err := ExtractClaims(c)
	if err != nil {
		return false
	}

	for _, userPermission := range claims.Permissions {
		if userPermission == permission {
			return true
		}
	}

	return false
}

// RequireRole creates middleware that requires a specific role
func RequireRole(role string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if !HasRole(c, role) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("Role '%s' required", role),
				"code":  "INSUFFICIENT_ROLE",
			})
			c.Abort()
			return
		}
		c.Next()
	})
}

// RequirePermission creates middleware that requires a specific permission
func RequirePermission(permission string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if !HasPermission(c, permission) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("Permission '%s' required", permission),
				"code":  "INSUFFICIENT_PERMISSION",
			})
			c.Abort()
			return
		}
		c.Next()
	})
}

// RequireAnyRole creates middleware that requires any of the specified roles
func RequireAnyRole(roles ...string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		for _, role := range roles {
			if HasRole(c, role) {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": fmt.Sprintf("One of roles %v required", roles),
			"code":  "INSUFFICIENT_ROLE",
		})
		c.Abort()
	})
}

// RequireAnyPermission creates middleware that requires any of the specified permissions
func RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		for _, permission := range permissions {
			if HasPermission(c, permission) {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": fmt.Sprintf("One of permissions %v required", permissions),
			"code":  "INSUFFICIENT_PERMISSION",
		})
		c.Abort()
	})
}

// ValidateJWTEndpoint provides an endpoint for JWT validation
func (js *JWTService) ValidateJWTEndpoint() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		var request struct {
			Token string `json:"token" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"details": err.Error(),
			})
			return
		}

		result := js.ValidateToken(request.Token)

		if result.Valid {
			c.JSON(http.StatusOK, gin.H{
				"valid":     true,
				"claims":    result.Claims,
				"key_id":    result.KeyID,
				"algorithm": result.Algorithm,
			})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"valid": false,
				"error": result.Error,
			})
		}
	})
}

// GetJWKSEndpoint provides JWKS information endpoint
func (js *JWTService) GetJWKSEndpoint() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		info := gin.H{
			"jwks_enabled":     js.config.Security.JWKS.Endpoint != "",
			"jwks_endpoint":    js.config.Security.JWKS.Endpoint,
			"cache_timeout":    js.config.Security.JWKS.CacheTimeout.String(),
			"refresh_interval": js.config.Security.JWKS.RefreshInterval.String(),
			"last_refresh":     js.keyCache.lastFetch,
		}

		if js.config.Security.JWKS.Endpoint != "" {
			js.keyCache.mutex.RLock()
			keyCount := len(js.keyCache.keys)
			js.keyCache.mutex.RUnlock()

			info["cached_keys_count"] = keyCount
		}

		c.JSON(http.StatusOK, info)
	})
}
