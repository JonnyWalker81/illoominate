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

type commentRepository struct {
	db DBTX
}

// NewCommentRepository creates a new comment repository
func NewCommentRepository(db *pgxpool.Pool) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, c *domain.Comment) error {
	query := `
		INSERT INTO comments (id, feedback_id, author_id, parent_id, body, visibility, is_edited, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, false, NOW(), NOW())
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		c.ID,
		c.FeedbackID,
		c.AuthorID,
		c.ParentID,
		c.Body,
		c.Visibility,
	).Scan(&c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	return nil
}

func (r *commentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	query := `
		SELECT id, feedback_id, author_id, parent_id, body, visibility, is_edited, created_at, updated_at
		FROM comments
		WHERE id = $1
	`

	var c domain.Comment
	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID,
		&c.FeedbackID,
		&c.AuthorID,
		&c.ParentID,
		&c.Body,
		&c.Visibility,
		&c.IsEdited,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	return &c, nil
}

func (r *commentRepository) ListByFeedback(ctx context.Context, feedbackID uuid.UUID, includeTeamOnly bool) ([]domain.Comment, error) {
	var query string
	var args []interface{}

	if includeTeamOnly {
		query = `
			SELECT id, feedback_id, author_id, parent_id, body, visibility, is_edited, created_at, updated_at
			FROM comments
			WHERE feedback_id = $1
			ORDER BY created_at ASC
		`
		args = []interface{}{feedbackID}
	} else {
		query = `
			SELECT id, feedback_id, author_id, parent_id, body, visibility, is_edited, created_at, updated_at
			FROM comments
			WHERE feedback_id = $1 AND visibility = 'COMMUNITY'
			ORDER BY created_at ASC
		`
		args = []interface{}{feedbackID}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}
	defer rows.Close()

	var comments []domain.Comment
	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(
			&c.ID,
			&c.FeedbackID,
			&c.AuthorID,
			&c.ParentID,
			&c.Body,
			&c.Visibility,
			&c.IsEdited,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, c)
	}

	return comments, nil
}

func (r *commentRepository) Update(ctx context.Context, c *domain.Comment) error {
	query := `
		UPDATE comments
		SET body = $2, is_edited = true, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRow(ctx, query, c.ID, c.Body).Scan(&c.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("failed to update comment: %w", err)
	}

	c.IsEdited = true
	return nil
}

func (r *commentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM comments WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
