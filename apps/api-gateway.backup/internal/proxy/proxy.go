package proxy

import (
	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/config"
)

// ProxyManager manages service proxying
type ProxyManager struct {
	config *config.Config
}

// NewManager creates a new proxy manager
func NewManager(cfg *config.Config) *ProxyManager {
	return &ProxyManager{
		config: cfg,
	}
}
