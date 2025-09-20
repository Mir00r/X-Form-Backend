package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kamkaiz/x-form-backend/collaboration-service/internal/models"
)

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token expired")
	ErrInsufficientPerms = errors.New("insufficient permissions")
	ErrUserNotFound      = errors.New("user not found")
)

// Claims represents JWT claims
type Claims struct {
	UserID      string   `json:"sub"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	SessionID   string   `json:"session_id"`
	jwt.RegisteredClaims
}

// Service handles authentication and authorization
type Service struct {
	jwtSecret     []byte
	serviceSecret []byte
	expiration    time.Duration
}

// NewService creates a new auth service
func NewService(jwtSecret, serviceSecret string, expiration time.Duration) *Service {
	return &Service{
		jwtSecret:     []byte(jwtSecret),
		serviceSecret: []byte(serviceSecret),
		expiration:    expiration,
	}
}

// ValidateToken validates a JWT token and returns claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	// Remove "Bearer " prefix if present
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateServiceToken validates a service-to-service token
func (s *Service) ValidateServiceToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.serviceSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// CreateUser creates a user model from claims
func (s *Service) CreateUser(claims *Claims) *models.User {
	return &models.User{
		ID:          claims.UserID,
		Email:       claims.Email,
		Name:        claims.Name,
		Role:        claims.Role,
		Permissions: claims.Permissions,
		ConnectedAt: time.Now(),
		LastSeen:    time.Now(),
		IsOnline:    true,
		SessionID:   claims.SessionID,
	}
}

// CheckFormPermission checks if a user has permission to access a form
func (s *Service) CheckFormPermission(ctx context.Context, userID, formID, permission string) (*models.FormPermission, error) {
	// This would typically query a database or cache
	// For now, we'll implement basic logic

	// TODO: Implement actual permission checking logic
	// This could involve:
	// 1. Checking if user is form owner
	// 2. Checking collaboration permissions
	// 3. Checking organization/team permissions

	// Mock implementation for demonstration
	formPerm := &models.FormPermission{
		FormID:      formID,
		UserID:      userID,
		Role:        "editor", // This would be fetched from database
		Permissions: []string{"view", "edit", "collaborate"},
		GrantedAt:   time.Now(),
		GrantedBy:   "system",
	}

	if !formPerm.HasPermission(permission) {
		return nil, ErrInsufficientPerms
	}

	return formPerm, nil
}

// PermissionChecker interface for checking permissions
type PermissionChecker interface {
	CheckFormPermission(ctx context.Context, userID, formID, permission string) (*models.FormPermission, error)
	HasPermission(ctx context.Context, userID string, permissions []string, requiredPermission string) bool
}

// HasPermission checks if a user has a specific permission
func (s *Service) HasPermission(ctx context.Context, userID string, permissions []string, requiredPermission string) bool {
	for _, perm := range permissions {
		if perm == requiredPermission || perm == "admin" || perm == "super_admin" {
			return true
		}
	}
	return false
}

// ExtractTokenFromQuery extracts token from WebSocket query parameters
func ExtractTokenFromQuery(query map[string][]string) string {
	if tokens, exists := query["token"]; exists && len(tokens) > 0 {
		return tokens[0]
	}

	if auth, exists := query["authorization"]; exists && len(auth) > 0 {
		return auth[0]
	}

	return ""
}

// ExtractTokenFromHeader extracts token from HTTP headers
func ExtractTokenFromHeader(header string) string {
	if header == "" {
		return ""
	}

	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}

	return header
}

// ValidatePermissions validates required permissions for an action
type ActionPermissions struct {
	View        []string
	Edit        []string
	Collaborate []string
	Admin       []string
}

// DefaultFormPermissions returns default permission requirements for form actions
func DefaultFormPermissions() ActionPermissions {
	return ActionPermissions{
		View:        []string{"view", "edit", "collaborate", "admin"},
		Edit:        []string{"edit", "admin"},
		Collaborate: []string{"collaborate", "edit", "admin"},
		Admin:       []string{"admin"},
	}
}

// CheckActionPermission checks if a user can perform a specific action
func (s *Service) CheckActionPermission(ctx context.Context, userID, formID, action string) error {
	permissions := DefaultFormPermissions()

	var requiredPerms []string
	switch action {
	case "view":
		requiredPerms = permissions.View
	case "edit":
		requiredPerms = permissions.Edit
	case "collaborate":
		requiredPerms = permissions.Collaborate
	case "admin":
		requiredPerms = permissions.Admin
	default:
		return ErrInsufficientPerms
	}

	// Check if user has any of the required permissions
	for _, perm := range requiredPerms {
		if formPerm, err := s.CheckFormPermission(ctx, userID, formID, perm); err == nil && formPerm != nil {
			return nil
		}
	}

	return ErrInsufficientPerms
}

// TokenInfo represents information about a token
type TokenInfo struct {
	Valid       bool      `json:"valid"`
	UserID      string    `json:"userId"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
	ExpiresAt   time.Time `json:"expiresAt"`
	IssuedAt    time.Time `json:"issuedAt"`
}

// GetTokenInfo returns information about a token without validating it fully
func (s *Service) GetTokenInfo(tokenString string) (*TokenInfo, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return &TokenInfo{Valid: false}, err
	}

	return &TokenInfo{
		Valid:       true,
		UserID:      claims.UserID,
		Email:       claims.Email,
		Role:        claims.Role,
		Permissions: claims.Permissions,
		ExpiresAt:   claims.ExpiresAt.Time,
		IssuedAt:    claims.IssuedAt.Time,
	}, nil
}

// AuthMiddleware represents authentication middleware context
type AuthMiddleware struct {
	authService *Service
	required    bool
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authService *Service, required bool) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		required:    required,
	}
}

// AuthContext represents authentication context
type AuthContext struct {
	User        *models.User `json:"user"`
	Claims      *Claims      `json:"claims"`
	Permissions []string     `json:"permissions"`
	IsValid     bool         `json:"isValid"`
}

// CreateAuthContext creates authentication context from token
func (s *Service) CreateAuthContext(tokenString string) (*AuthContext, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return &AuthContext{IsValid: false}, err
	}

	user := s.CreateUser(claims)

	return &AuthContext{
		User:        user,
		Claims:      claims,
		Permissions: claims.Permissions,
		IsValid:     true,
	}, nil
}

// IsFormOwner checks if a user is the owner of a form
func (s *Service) IsFormOwner(ctx context.Context, userID, formID string) (bool, error) {
	// This would typically query the form service or database
	// For now, we'll implement a mock check

	// TODO: Implement actual ownership checking
	// This could involve calling the form service API

	return true, nil // Mock implementation
}

// GetUserPermissions retrieves user permissions for a form
func (s *Service) GetUserPermissions(ctx context.Context, userID, formID string) ([]string, error) {
	// This would typically query a permissions database or cache
	// For now, we'll return default permissions

	// TODO: Implement actual permission retrieval
	// This could involve:
	// 1. Checking user role in the organization
	// 2. Checking direct form permissions
	// 3. Checking team/group permissions

	return []string{"view", "edit", "collaborate"}, nil // Mock implementation
}

// RevokeToken revokes a token (would typically blacklist it)
func (s *Service) RevokeToken(ctx context.Context, tokenString string) error {
	// This would typically add the token to a blacklist
	// Implementation would depend on your token revocation strategy

	// TODO: Implement token revocation
	// This could involve:
	// 1. Adding token to Redis blacklist
	// 2. Updating token revocation list
	// 3. Broadcasting revocation to other services

	return nil // Mock implementation
}

// RefreshToken creates a new token from an existing valid token
func (s *Service) RefreshToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Create new claims with extended expiration
	newClaims := &Claims{
		UserID:      claims.UserID,
		Email:       claims.Email,
		Name:        claims.Name,
		Role:        claims.Role,
		Permissions: claims.Permissions,
		SessionID:   claims.SessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	return token.SignedString(s.jwtSecret)
}

// CanAccessForm checks if a user can access a specific form
func (s *Service) CanAccessForm(user *models.User, formID string) bool {
	// TODO: Implement actual form access check
	// This would typically involve checking the user's permissions for the form

	if user.Role == "admin" || user.Role == "super_admin" {
		return true
	}

	// Check if user has view permissions
	for _, permission := range user.Permissions {
		if permission == "forms:view" || permission == "forms:edit" || permission == fmt.Sprintf("form:%s:view", formID) {
			return true
		}
	}

	return false
}

// CanEditForm checks if a user can edit a specific form
func (s *Service) CanEditForm(user *models.User, formID string) bool {
	// TODO: Implement actual form edit permission check
	// This would typically involve checking if the user is the owner or has edit permissions

	if user.Role == "admin" || user.Role == "super_admin" {
		return true
	}

	// Check if user has edit permissions
	for _, permission := range user.Permissions {
		if permission == "forms:edit" || permission == fmt.Sprintf("form:%s:edit", formID) {
			return true
		}
	}

	return false
}
