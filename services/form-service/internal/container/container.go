// Package container provides dependency injection for the application
// Following the Dependency Inversion Principle from SOLID principles
package container

import (
	"fmt"

	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/config"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/database"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/domain"
	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/infrastructure"
	formhandlers "github.com/Mir00r/X-Form-Backend/services/form-service/internal/interface/http"
	"gorm.io/gorm"
)

// Container holds all application dependencies
// Implements Inversion of Control (IoC) pattern
type Container struct {
	Config      *config.Config
	DB          *gorm.DB
	FormRepo    domain.FormRepository
	FormHandler *formhandlers.FormHandler
}

// New creates and wires up all application dependencies
// This function serves as the composition root for the application
func New() (*Container, error) {
	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run database migrations
	if err := database.Migrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Get SQL database instance for repositories
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get SQL database instance: %w", err)
	}

	// Initialize repositories (Infrastructure Layer)
	formRepo := infrastructure.NewPostgreSQLFormRepository(sqlDB)

	// For now, create a minimal HTTP handler that works with the current setup
	// TODO: Add proper application service layer once dependencies are resolved
	formHandler := formhandlers.NewFormHandler(nil) // We'll pass nil for now

	return &Container{
		Config:      cfg,
		DB:          db,
		FormRepo:    formRepo,
		FormHandler: formHandler,
	}, nil
}

// Close gracefully shuts down the container and cleans up resources
func (c *Container) Close() error {
	if c.DB != nil {
		sqlDB, err := c.DB.DB()
		if err != nil {
			return fmt.Errorf("failed to get SQL database instance for closing: %w", err)
		}
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}
	}
	return nil
}
