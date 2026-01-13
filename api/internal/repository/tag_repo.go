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

type tagRepository struct {
	db DBTX
}

// NewTagRepository creates a new tag repository
func NewTagRepository(db *pgxpool.Pool) TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) Create(ctx context.Context, t *domain.Tag) error {
	query := `
		INSERT INTO tags (id, project_id, name, slug, color, description)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(ctx, query,
		t.ID,
		t.ProjectID,
		t.Name,
		t.Slug,
		t.Color,
		t.Description,
	)

	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	return nil
}

func (r *tagRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tag, error) {
	query := `
		SELECT id, project_id, name, slug, color, description
		FROM tags
		WHERE id = $1
	`

	var t domain.Tag
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID,
		&t.ProjectID,
		&t.Name,
		&t.Slug,
		&t.Color,
		&t.Description,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return &t, nil
}

func (r *tagRepository) GetBySlug(ctx context.Context, projectID uuid.UUID, slug string) (*domain.Tag, error) {
	query := `
		SELECT id, project_id, name, slug, color, description
		FROM tags
		WHERE project_id = $1 AND slug = $2
	`

	var t domain.Tag
	err := r.db.QueryRow(ctx, query, projectID, slug).Scan(
		&t.ID,
		&t.ProjectID,
		&t.Name,
		&t.Slug,
		&t.Color,
		&t.Description,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get tag by slug: %w", err)
	}

	return &t, nil
}

func (r *tagRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Tag, error) {
	query := `
		SELECT t.id, t.project_id, t.name, t.slug, t.color, t.description,
			   (SELECT COUNT(*) FROM feedback_tags ft WHERE ft.tag_id = t.id) as feedback_count
		FROM tags t
		WHERE t.project_id = $1
		ORDER BY t.name ASC
	`

	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	defer rows.Close()

	var tags []domain.Tag
	for rows.Next() {
		var t domain.Tag
		var feedbackCount int
		if err := rows.Scan(
			&t.ID,
			&t.ProjectID,
			&t.Name,
			&t.Slug,
			&t.Color,
			&t.Description,
			&feedbackCount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		t.FeedbackCount = feedbackCount
		tags = append(tags, t)
	}

	return tags, nil
}

func (r *tagRepository) ListByFeedback(ctx context.Context, feedbackID uuid.UUID) ([]domain.Tag, error) {
	query := `
		SELECT t.id, t.project_id, t.name, t.slug, t.color, t.description
		FROM tags t
		JOIN feedback_tags ft ON ft.tag_id = t.id
		WHERE ft.feedback_id = $1
		ORDER BY t.name ASC
	`

	rows, err := r.db.Query(ctx, query, feedbackID)
	if err != nil {
		return nil, fmt.Errorf("failed to list feedback tags: %w", err)
	}
	defer rows.Close()

	var tags []domain.Tag
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(
			&t.ID,
			&t.ProjectID,
			&t.Name,
			&t.Slug,
			&t.Color,
			&t.Description,
		); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, t)
	}

	return tags, nil
}

func (r *tagRepository) Update(ctx context.Context, t *domain.Tag) error {
	query := `
		UPDATE tags
		SET name = $2, slug = $3, color = $4, description = $5
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		t.ID,
		t.Name,
		t.Slug,
		t.Color,
		t.Description,
	)

	if err != nil {
		return fmt.Errorf("failed to update tag: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *tagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tags WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
