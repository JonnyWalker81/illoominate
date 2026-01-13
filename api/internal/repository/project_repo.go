package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fulldisclosure/api/internal/domain"
)

type projectRepository struct {
	db DBTX
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *pgxpool.Pool) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(ctx context.Context, p *domain.Project) error {
	query := `
		INSERT INTO projects (id, name, slug, project_key, settings, logo_url, primary_color, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		p.ID,
		p.Name,
		p.Slug,
		p.ProjectKey,
		p.Settings,
		p.LogoURL,
		p.PrimaryColor,
	).Scan(&p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

func (r *projectRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	query := `
		SELECT id, name, slug, project_key, settings, logo_url, primary_color, created_at, updated_at
		FROM projects
		WHERE id = $1
	`

	var p domain.Project
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.Name,
		&p.Slug,
		&p.ProjectKey,
		&p.Settings,
		&p.LogoURL,
		&p.PrimaryColor,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &p, nil
}

func (r *projectRepository) GetBySlug(ctx context.Context, slug string) (*domain.Project, error) {
	query := `
		SELECT id, name, slug, project_key, settings, logo_url, primary_color, created_at, updated_at
		FROM projects
		WHERE slug = $1
	`

	var p domain.Project
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&p.ID,
		&p.Name,
		&p.Slug,
		&p.ProjectKey,
		&p.Settings,
		&p.LogoURL,
		&p.PrimaryColor,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get project by slug: %w", err)
	}

	return &p, nil
}

func (r *projectRepository) GetByProjectKey(ctx context.Context, key string) (*domain.Project, error) {
	query := `
		SELECT id, name, slug, project_key, settings, logo_url, primary_color, created_at, updated_at
		FROM projects
		WHERE project_key = $1
	`

	var p domain.Project
	err := r.db.QueryRow(ctx, query, key).Scan(
		&p.ID,
		&p.Name,
		&p.Slug,
		&p.ProjectKey,
		&p.Settings,
		&p.LogoURL,
		&p.PrimaryColor,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get project by key: %w", err)
	}

	return &p, nil
}

func (r *projectRepository) Update(ctx context.Context, p *domain.Project) error {
	query := `
		UPDATE projects
		SET name = $2, slug = $3, settings = $4, logo_url = $5, primary_color = $6, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRow(ctx, query,
		p.ID,
		p.Name,
		p.Slug,
		p.Settings,
		p.LogoURL,
		p.PrimaryColor,
	).Scan(&p.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

func (r *projectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM projects WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
