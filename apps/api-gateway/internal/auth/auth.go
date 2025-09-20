// Package auth provides authentication and authorization services
// Implements JWT token validation and user management following security best practices
package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Mir00r/X-Form-Backend/enhanced-architecture/api-gateway/internal/config"
)

// Claims represents JWT token claims
type Claims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	SessionID   string   `json:"session_id,omitempty"`
	DeviceID    string   `json:"device_id,omitempty"`

	// Standard JWT claims
	jwt.RegisteredClaims
}

// User represents user information
type User struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	Username    string     `json:"username"`
	Role        string     `json:"role"`
	Roles       []string   `json:"roles,omitempty"`
	Permissions []string   `json:"permissions,omitempty"`
	IsActive    bool       `json:"is_active"`
	LastLogin   *time.Time `json:"last_login,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// AuthService handles authentication and authorization
type AuthService struct {
	config     config.JWTConfig
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

// AuthError represents an authentication error
type AuthError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *AuthError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Token validation errors
var (
	ErrTokenMissing     = &AuthError{Code: "TOKEN_MISSING", Message: "Authentication token is missing"}
	ErrTokenInvalid     = &AuthError{Code: "TOKEN_INVALID", Message: "Authentication token is invalid"}
	ErrTokenExpired     = &AuthError{Code: "TOKEN_EXPIRED", Message: "Authentication token has expired"}
	ErrTokenMalformed   = &AuthError{Code: "TOKEN_MALFORMED", Message: "Authentication token is malformed"}
	ErrTokenSignature   = &AuthError{Code: "TOKEN_SIGNATURE", Message: "Authentication token signature is invalid"}
	ErrInsufficientRole = &AuthError{Code: "INSUFFICIENT_ROLE", Message: "Insufficient role for this operation"}
	ErrUserNotFound     = &AuthError{Code: "USER_NOT_FOUND", Message: "User not found"}
	ErrUserInactive     = &AuthError{Code: "USER_INACTIVE", Message: "User account is inactive"}
)

// NewAuthService creates a new authentication service
func NewAuthService(cfg config.JWTConfig) (*AuthService, error) {
	service := &AuthService{
		config: cfg,
	}

	// Load public key for token validation
	if cfg.PublicKey != "" {
		publicKey, err := parsePublicKey(cfg.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
		service.publicKey = publicKey
	}

	// Load private key for token signing (if needed)
	if cfg.PrivateKey != "" {
		privateKey, err := parsePrivateKey(cfg.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		service.privateKey = privateKey
	}

	return service, nil
}

// ValidateToken validates and parses a JWT token
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrTokenMissing
	}

	// Parse token with claims
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, s.getSigningKey)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, ErrTokenSignature
		}
		return nil, &AuthError{
			Code:    "TOKEN_PARSE_ERROR",
			Message: "Failed to parse token",
			Details: err.Error(),
		}
	}

	// Check if token is valid
	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	// Extract claims
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, &AuthError{
			Code:    "TOKEN_CLAIMS_INVALID",
			Message: "Invalid token claims",
		}
	}

	// Additional validation
	if err := s.validateClaims(claims); err != nil {
		return nil, err
	}

	return claims, nil
}

// GenerateToken generates a new JWT token for a user
func (s *AuthService) GenerateToken(user *User) (string, error) {
	if s.privateKey == nil {
		return "", &AuthError{
			Code:    "SIGNING_KEY_MISSING",
			Message: "Token signing key is not configured",
		}
	}

	now := time.Now()

	// Create claims
	claims := &Claims{
		UserID:      user.ID,
		Email:       user.Email,
		Username:    user.Username,
		Role:        user.Role,
		Roles:       user.Roles,
		Permissions: user.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			Subject:   user.ID,
			Audience:  jwt.ClaimStrings{s.config.Audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.ExpirationTime)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        generateTokenID(),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Sign token
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", &AuthError{
			Code:    "TOKEN_SIGN_ERROR",
			Message: "Failed to sign token",
			Details: err.Error(),
		}
	}

	return tokenString, nil
}

// ValidateRole checks if the user has the required role
func (s *AuthService) ValidateRole(claims *Claims, requiredRole string) error {
	if claims == nil {
		return ErrTokenInvalid
	}

	// Check primary role
	if claims.Role == requiredRole {
		return nil
	}

	// Check additional roles
	for _, role := range claims.Roles {
		if role == requiredRole {
			return nil
		}
	}

	return &AuthError{
		Code:    "INSUFFICIENT_ROLE",
		Message: fmt.Sprintf("Required role '%s' not found", requiredRole),
		Details: fmt.Sprintf("User has role '%s'", claims.Role),
	}
}

// ValidatePermission checks if the user has the required permission
func (s *AuthService) ValidatePermission(claims *Claims, requiredPermission string) error {
	if claims == nil {
		return ErrTokenInvalid
	}

	// Check permissions
	for _, permission := range claims.Permissions {
		if permission == requiredPermission {
			return nil
		}
	}

	return &AuthError{
		Code:    "INSUFFICIENT_PERMISSION",
		Message: fmt.Sprintf("Required permission '%s' not found", requiredPermission),
	}
}

// ValidateRoleHierarchy checks if the user has sufficient role level
func (s *AuthService) ValidateRoleHierarchy(claims *Claims, requiredLevel RoleLevel) error {
	if claims == nil {
		return ErrTokenInvalid
	}

	userLevel := GetRoleLevel(claims.Role)
	if userLevel < requiredLevel {
		return &AuthError{
			Code:    "INSUFFICIENT_ROLE_LEVEL",
			Message: fmt.Sprintf("Required role level %d, user has level %d", requiredLevel, userLevel),
		}
	}

	return nil
}

// RefreshToken generates a new token if the current one is close to expiry
func (s *AuthService) RefreshToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		// Allow refresh for expired tokens within grace period
		if !errors.Is(err, ErrTokenExpired) {
			return "", err
		}

		// Parse expired token to get claims
		token, parseErr := jwt.ParseWithClaims(tokenString, &Claims{}, s.getSigningKey)
		if parseErr != nil {
			return "", err // Return original error
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			return "", err
		}

		// Check if token is within refresh grace period (using RefreshTime)
		if time.Since(claims.ExpiresAt.Time) > s.config.RefreshTime {
			return "", &AuthError{
				Code:    "TOKEN_REFRESH_EXPIRED",
				Message: "Token is beyond refresh grace period",
			}
		}
	}

	// Generate new token with same claims but updated timestamps
	user := &User{
		ID:          claims.UserID,
		Email:       claims.Email,
		Username:    claims.Username,
		Role:        claims.Role,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
		IsActive:    true, // Assume active for refresh
	}

	return s.GenerateToken(user)
}

// RevokeToken adds a token to the revocation list
func (s *AuthService) RevokeToken(tokenString string) error {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return err
	}

	// In production, store revoked tokens in Redis or database
	// For now, just validate the token exists
	_ = claims

	// TODO: Implement token revocation storage
	// - Add token ID to revocation list
	// - Set expiration time for cleanup

	return nil
}

// Helper methods

// getSigningKey returns the key for token verification
func (s *AuthService) getSigningKey(token *jwt.Token) (interface{}, error) {
	// Verify signing method
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, &AuthError{
			Code:    "INVALID_SIGNING_METHOD",
			Message: fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"]),
		}
	}

	if s.publicKey == nil {
		return nil, &AuthError{
			Code:    "PUBLIC_KEY_MISSING",
			Message: "Public key not configured",
		}
	}

	return s.publicKey, nil
}

// validateClaims performs additional validation on token claims
func (s *AuthService) validateClaims(claims *Claims) error {
	now := time.Now()

	// Check expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(now) {
		return ErrTokenExpired
	}

	// Check not before
	if claims.NotBefore != nil && claims.NotBefore.After(now) {
		return &AuthError{
			Code:    "TOKEN_NOT_YET_VALID",
			Message: "Token is not yet valid",
		}
	}

	// Check issuer
	if s.config.Issuer != "" && claims.Issuer != s.config.Issuer {
		return &AuthError{
			Code:    "INVALID_ISSUER",
			Message: "Token issuer is invalid",
		}
	}

	// Check audience
	if s.config.Audience != "" {
		validAudience := false
		for _, aud := range claims.Audience {
			if aud == s.config.Audience {
				validAudience = true
				break
			}
		}
		if !validAudience {
			return &AuthError{
				Code:    "INVALID_AUDIENCE",
				Message: "Token audience is invalid",
			}
		}
	}

	// Check required fields
	if claims.UserID == "" {
		return &AuthError{
			Code:    "MISSING_USER_ID",
			Message: "Token is missing user ID",
		}
	}

	return nil
}

// Role hierarchy and permissions

// RoleLevel represents the hierarchy level of a role
type RoleLevel int

const (
	RoleLevelGuest RoleLevel = iota
	RoleLevelUser
	RoleLevelModerator
	RoleLevelAdmin
	RoleLevelSuperAdmin
)

// Role definitions
const (
	RoleGuest      = "guest"
	RoleUser       = "user"
	RoleModerator  = "moderator"
	RoleAdmin      = "admin"
	RoleSuperAdmin = "super_admin"
)

// GetRoleLevel returns the hierarchy level for a role
func GetRoleLevel(role string) RoleLevel {
	switch role {
	case RoleGuest:
		return RoleLevelGuest
	case RoleUser:
		return RoleLevelUser
	case RoleModerator:
		return RoleLevelModerator
	case RoleAdmin:
		return RoleLevelAdmin
	case RoleSuperAdmin:
		return RoleLevelSuperAdmin
	default:
		return RoleLevelGuest
	}
}

// Utility functions

// parsePublicKey parses RSA public key from PEM format
func parsePublicKey(keyData string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(keyData))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("key is not RSA public key")
	}

	return rsaPub, nil
}

// parsePrivateKey parses RSA private key from PEM format
func parsePrivateKey(keyData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(keyData))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// generateTokenID generates a unique token ID
func generateTokenID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// DefaultJWTConfig returns a default JWT configuration
func DefaultJWTConfig() config.JWTConfig {
	return config.JWTConfig{
		Secret:         "your-secret-key", // Should be set via environment
		ExpirationTime: time.Hour * 24,    // 24 hours
		RefreshTime:    time.Hour * 2,     // 2 hours refresh window
		Issuer:         "x-form-api-gateway",
		Audience:       "x-form-services",
		Algorithm:      "RS256",
	}
}
