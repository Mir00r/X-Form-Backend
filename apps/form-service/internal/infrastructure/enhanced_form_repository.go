// Enhanced Form Repository with Microservices Best Practices
// Implements comprehensive data access layer with proper error handling

package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/domain"
)

// FormRepository implements domain.FormRepository interface
type FormRepository struct {
	db *gorm.DB
}

// NewFormRepository creates a new form repository instance
func NewFormRepository(db *gorm.DB) domain.FormRepository {
	return &FormRepository{
		db: db,
	}
}

// =============================================================================
// Form CRUD Operations
// =============================================================================

// Create creates a new form
func (r *FormRepository) Create(ctx context.Context, form *domain.Form) error {
	if form == nil {
		return domain.NewValidationError("form", "form cannot be nil")
	}

	// Set ID if not already set
	if form.ID == uuid.Nil {
		form.ID = uuid.New()
	}

	// Set creation timestamp
	now := time.Now()
	if form.CreatedAt.IsZero() {
		form.CreatedAt = now
	}
	form.UpdatedAt = now

	if err := r.db.WithContext(ctx).Create(form).Error; err != nil {
		return fmt.Errorf("failed to create form: %w", err)
	}

	return nil
}

// GetByID retrieves a form by its ID
func (r *FormRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Form, error) {
	if id == uuid.Nil {
		return nil, domain.NewValidationError("id", "form ID cannot be nil")
	}

	var form domain.Form
	err := r.db.WithContext(ctx).
		Preload("Questions").
		Preload("Questions.Options").
		First(&form, "id = ?", id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.NewNotFoundError("form", id.String())
		}
		return nil, fmt.Errorf("failed to get form: %w", err)
	}

	return &form, nil
}

// GetByUserID retrieves forms by user ID with pagination
func (r *FormRepository) GetByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*domain.Form, int64, error) {
	if userID == uuid.Nil {
		return nil, 0, domain.NewValidationError("userID", "user ID cannot be nil")
	}

	var forms []*domain.Form
	var total int64

	// Count total records first
	err := r.db.WithContext(ctx).
		Model(&domain.Form{}).
		Where("user_id = ?", userID).
		Count(&total).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to count forms: %w", err)
	}

	// Get paginated results
	err = r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&forms).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get forms by user ID: %w", err)
	}

	return forms, total, nil
}

// Update updates an existing form
func (r *FormRepository) Update(ctx context.Context, form *domain.Form) error {
	if form == nil {
		return domain.NewValidationError("form", "form cannot be nil")
	}

	if form.ID == uuid.Nil {
		return domain.NewValidationError("id", "form ID cannot be nil")
	}

	// Update timestamp
	form.UpdatedAt = time.Now()

	// Use a transaction to update form and questions
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update form
	if err := tx.Save(form).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update form: %w", err)
	}

	// Delete existing questions
	if err := tx.Where("form_id = ?", form.ID).Delete(&domain.Question{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete existing questions: %w", err)
	}

	// Create new questions
	for i := range form.Questions {
		form.Questions[i].ID = uuid.New()
		// Set form_id if needed by your GORM model
	}

	if len(form.Questions) > 0 {
		if err := tx.Create(&form.Questions).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create questions: %w", err)
		}
	}

	return tx.Commit().Error
}

// Delete deletes a form by ID
func (r *FormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return domain.NewValidationError("id", "form ID cannot be nil")
	}

	// Use transaction for cascading delete
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete questions first (if not using foreign key constraints)
	if err := tx.Where("form_id = ?", id).Delete(&domain.Question{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete questions: %w", err)
	}

	// Delete form
	result := tx.Delete(&domain.Form{}, "id = ?", id)
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete form: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return domain.NewNotFoundError("form", id.String())
	}

	return tx.Commit().Error
}

// =============================================================================
// Additional Required Methods
// =============================================================================

// GetPublishedForms retrieves published forms with filters
func (r *FormRepository) GetPublishedForms(ctx context.Context, filters domain.FormFilters) ([]*domain.Form, error) {
	var forms []*domain.Form

	query := r.db.WithContext(ctx).
		Where("status = ?", domain.FormStatusPublished)

	// Apply filters
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}
	if filters.Search != nil {
		query = query.Where("title ILIKE ?", fmt.Sprintf("%%%s%%", *filters.Search))
	}
	if filters.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *filters.CreatedAfter)
	}
	if filters.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *filters.CreatedBefore)
	}

	err := query.
		Limit(filters.Limit).
		Offset(filters.Offset).
		Order("created_at DESC").
		Find(&forms).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get published forms: %w", err)
	}

	return forms, nil
}

// UpdateStatus changes form status
func (r *FormRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.FormStatus) error {
	if id == uuid.Nil {
		return domain.NewValidationError("id", "form ID cannot be nil")
	}

	result := r.db.WithContext(ctx).
		Model(&domain.Form{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("failed to update form status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.NewNotFoundError("form", id.String())
	}

	return nil
}

// =============================================================================
// Health Check and Utilities
// =============================================================================

// HealthCheck verifies database connectivity
func (r *FormRepository) HealthCheck(ctx context.Context) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}
