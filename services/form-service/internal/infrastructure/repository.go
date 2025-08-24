// Package infrastructure contains implementations of domain repositories and external integrations
// This layer handles persistence, external APIs, and other infrastructure concerns
package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/Mir00r/X-Form-Backend/services/form-service/internal/domain"
)

// PostgreSQLFormRepository implements FormRepository using PostgreSQL
// Follows the Repository pattern to abstract data access
type PostgreSQLFormRepository struct {
	db *sql.DB
}

// NewPostgreSQLFormRepository creates a new PostgreSQL form repository
func NewPostgreSQLFormRepository(db *sql.DB) *PostgreSQLFormRepository {
	return &PostgreSQLFormRepository{
		db: db,
	}
}

// Create implements domain.FormRepository.Create
func (r *PostgreSQLFormRepository) Create(ctx context.Context, form *domain.Form) error {
	query := `
		INSERT INTO forms (
			id, user_id, title, description, questions, settings, 
			status, created_at, updated_at, published_at, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	// Serialize complex fields to JSON
	questionsJSON, err := json.Marshal(form.Questions)
	if err != nil {
		return fmt.Errorf("failed to marshal questions: %w", err)
	}

	settingsJSON, err := json.Marshal(form.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	metadataJSON, err := json.Marshal(form.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		form.ID,
		form.UserID,
		form.Title,
		form.Description,
		questionsJSON,
		settingsJSON,
		form.Status,
		form.CreatedAt,
		form.UpdatedAt,
		form.PublishedAt,
		metadataJSON,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return domain.NewConflictError("form", "form with this ID already exists")
		}
		return fmt.Errorf("failed to insert form: %w", err)
	}

	return nil
}

// GetByID implements domain.FormRepository.GetByID
func (r *PostgreSQLFormRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Form, error) {
	query := `
		SELECT id, user_id, title, description, questions, settings,
			   status, created_at, updated_at, published_at, metadata
		FROM forms 
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	form, err := r.scanForm(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Form not found
		}
		return nil, fmt.Errorf("failed to scan form: %w", err)
	}

	return form, nil
}

// GetByUserID implements domain.FormRepository.GetByUserID
func (r *PostgreSQLFormRepository) GetByUserID(
	ctx context.Context,
	userID uuid.UUID,
	offset, limit int,
) ([]*domain.Form, int64, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM forms WHERE user_id = $1`
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count user forms: %w", err)
	}

	// Get forms with pagination
	query := `
		SELECT id, user_id, title, description, questions, settings,
			   status, created_at, updated_at, published_at, metadata
		FROM forms 
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query user forms: %w", err)
	}
	defer rows.Close()

	var forms []*domain.Form
	for rows.Next() {
		form, err := r.scanForm(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan form: %w", err)
		}
		forms = append(forms, form)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating forms: %w", err)
	}

	return forms, total, nil
}

// Update implements domain.FormRepository.Update
func (r *PostgreSQLFormRepository) Update(ctx context.Context, form *domain.Form) error {
	query := `
		UPDATE forms 
		SET title = $2, description = $3, questions = $4, settings = $5,
			status = $6, updated_at = $7, published_at = $8, metadata = $9
		WHERE id = $1
	`

	// Serialize complex fields to JSON
	questionsJSON, err := json.Marshal(form.Questions)
	if err != nil {
		return fmt.Errorf("failed to marshal questions: %w", err)
	}

	settingsJSON, err := json.Marshal(form.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	metadataJSON, err := json.Marshal(form.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query,
		form.ID,
		form.Title,
		form.Description,
		questionsJSON,
		settingsJSON,
		form.Status,
		form.UpdatedAt,
		form.PublishedAt,
		metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to update form: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.NewNotFoundError("form", form.ID.String())
	}

	return nil
}

// Delete implements domain.FormRepository.Delete
func (r *PostgreSQLFormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM forms WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete form: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.NewNotFoundError("form", id.String())
	}

	return nil
}

// GetPublishedForms implements domain.FormRepository.GetPublishedForms
func (r *PostgreSQLFormRepository) GetPublishedForms(
	ctx context.Context,
	filters domain.FormFilters,
) ([]*domain.Form, error) {
	query := `
		SELECT id, user_id, title, description, questions, settings,
			   status, created_at, updated_at, published_at, metadata
		FROM forms 
		WHERE status = $1
	`
	args := []interface{}{domain.FormStatusPublished}
	argIndex := 2

	// Apply filters
	if filters.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *filters.UserID)
		argIndex++
	}

	if filters.Search != nil {
		query += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex+1)
		searchPattern := "%" + *filters.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	if filters.CreatedAfter != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filters.CreatedAfter)
		argIndex++
	}

	if filters.CreatedBefore != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filters.CreatedBefore)
		argIndex++
	}

	// Add ordering and pagination
	query += " ORDER BY created_at DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++

		if filters.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, filters.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query published forms: %w", err)
	}
	defer rows.Close()

	var forms []*domain.Form
	for rows.Next() {
		form, err := r.scanForm(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan form: %w", err)
		}
		forms = append(forms, form)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating forms: %w", err)
	}

	return forms, nil
}

// UpdateStatus implements domain.FormRepository.UpdateStatus
func (r *PostgreSQLFormRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status domain.FormStatus,
) error {
	var query string
	var args []interface{}

	if status == domain.FormStatusPublished {
		query = `UPDATE forms SET status = $2, published_at = $3, updated_at = $4 WHERE id = $1`
		now := time.Now()
		args = []interface{}{id, status, now, now}
	} else {
		query = `UPDATE forms SET status = $2, updated_at = $3 WHERE id = $1`
		args = []interface{}{id, status, time.Now()}
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update form status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.NewNotFoundError("form", id.String())
	}

	return nil
}

// Helper methods

// Scanner interface for scanning rows (implements Interface Segregation Principle)
type Scanner interface {
	Scan(dest ...interface{}) error
}

// scanForm scans a database row into a Form entity
func (r *PostgreSQLFormRepository) scanForm(scanner Scanner) (*domain.Form, error) {
	var form domain.Form
	var questionsJSON, settingsJSON, metadataJSON []byte
	var publishedAt sql.NullTime

	err := scanner.Scan(
		&form.ID,
		&form.UserID,
		&form.Title,
		&form.Description,
		&questionsJSON,
		&settingsJSON,
		&form.Status,
		&form.CreatedAt,
		&form.UpdatedAt,
		&publishedAt,
		&metadataJSON,
	)

	if err != nil {
		return nil, err
	}

	// Deserialize JSON fields
	if err := json.Unmarshal(questionsJSON, &form.Questions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal questions: %w", err)
	}

	if err := json.Unmarshal(settingsJSON, &form.Settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &form.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	if publishedAt.Valid {
		form.PublishedAt = &publishedAt.Time
	}

	return &form, nil
}

// isUniqueViolation checks if error is a unique constraint violation
func isUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505" // unique_violation
	}
	return false
}

// DatabaseMigrator handles database schema migrations
type DatabaseMigrator struct {
	db *sql.DB
}

// NewDatabaseMigrator creates a new database migrator
func NewDatabaseMigrator(db *sql.DB) *DatabaseMigrator {
	return &DatabaseMigrator{db: db}
}

// Migrate runs database migrations
func (m *DatabaseMigrator) Migrate(ctx context.Context) error {
	migrations := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
		`
		CREATE TABLE IF NOT EXISTS forms (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL,
			title VARCHAR(200) NOT NULL,
			description TEXT,
			questions JSONB NOT NULL DEFAULT '[]',
			settings JSONB NOT NULL DEFAULT '{}',
			status VARCHAR(20) NOT NULL DEFAULT 'draft',
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			published_at TIMESTAMP WITH TIME ZONE,
			metadata JSONB DEFAULT '{}'
		)
		`,
		`CREATE INDEX IF NOT EXISTS idx_forms_user_id ON forms(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_forms_status ON forms(status)`,
		`CREATE INDEX IF NOT EXISTS idx_forms_created_at ON forms(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_forms_published_at ON forms(published_at)`,
		`CREATE INDEX IF NOT EXISTS idx_forms_title_search ON forms USING gin(to_tsvector('english', title))`,
	}

	for _, migration := range migrations {
		if _, err := m.db.ExecContext(ctx, migration); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	return nil
}
