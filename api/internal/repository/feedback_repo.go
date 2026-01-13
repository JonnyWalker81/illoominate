package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fulldisclosure/api/internal/domain"
)

type feedbackRepository struct {
	db DBTX
}

// NewFeedbackRepository creates a new feedback repository
func NewFeedbackRepository(db *pgxpool.Pool) FeedbackRepository {
	return &feedbackRepository{db: db}
}

func (r *feedbackRepository) Create(ctx context.Context, f *domain.Feedback) error {
	query := `
		INSERT INTO feedback (
			id, project_id, author_id, assigned_to, title, description,
			type, status, severity, visibility, submitter_email, submitter_name,
			source, source_metadata, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW())
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		f.ID,
		f.ProjectID,
		f.AuthorID,
		f.AssignedTo,
		f.Title,
		f.Description,
		f.Type,
		f.Status,
		f.Severity,
		f.Visibility,
		f.SubmitterEmail,
		f.SubmitterName,
		f.Source,
		f.SourceMetadata,
	).Scan(&f.CreatedAt, &f.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create feedback: %w", err)
	}

	return nil
}

func (r *feedbackRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Feedback, error) {
	query := `
		SELECT
			id, project_id, author_id, assigned_to, canonical_id, title, description,
			type, status, severity, visibility, vote_count, comment_count,
			submitter_email, submitter_name, source, source_metadata,
			created_at, updated_at, resolved_at
		FROM feedback
		WHERE id = $1
	`

	var f domain.Feedback
	err := r.db.QueryRow(ctx, query, id).Scan(
		&f.ID,
		&f.ProjectID,
		&f.AuthorID,
		&f.AssignedTo,
		&f.CanonicalID,
		&f.Title,
		&f.Description,
		&f.Type,
		&f.Status,
		&f.Severity,
		&f.Visibility,
		&f.VoteCount,
		&f.CommentCount,
		&f.SubmitterEmail,
		&f.SubmitterName,
		&f.Source,
		&f.SourceMetadata,
		&f.CreatedAt,
		&f.UpdatedAt,
		&f.ResolvedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get feedback: %w", err)
	}

	return &f, nil
}

func (r *feedbackRepository) List(ctx context.Context, projectID uuid.UUID, filter FeedbackFilter) ([]domain.Feedback, int, error) {
	// Build WHERE clause
	conditions := []string{"project_id = $1", "canonical_id IS NULL"} // Exclude merged items
	args := []interface{}{projectID}
	argIndex := 2

	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, *filter.Type)
		argIndex++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.Visibility != nil {
		conditions = append(conditions, fmt.Sprintf("visibility = $%d", argIndex))
		args = append(args, *filter.Visibility)
		argIndex++
	}

	if filter.AssignedTo != nil {
		conditions = append(conditions, fmt.Sprintf("assigned_to = $%d", argIndex))
		args = append(args, *filter.AssignedTo)
		argIndex++
	}

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(title ILIKE $%d OR description ILIKE $%d)",
			argIndex, argIndex,
		))
		args = append(args, "%"+*filter.Search+"%")
		argIndex++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM feedback WHERE %s", whereClause)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count feedback: %w", err)
	}

	// Build ORDER BY clause
	sortBy := "created_at"
	if filter.SortBy != "" {
		switch filter.SortBy {
		case "updated_at", "vote_count", "created_at":
			sortBy = filter.SortBy
		}
	}

	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	// List query
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	listQuery := fmt.Sprintf(`
		SELECT
			id, project_id, author_id, assigned_to, canonical_id, title, description,
			type, status, severity, visibility, vote_count, comment_count,
			submitter_email, submitter_name, source, source_metadata,
			created_at, updated_at, resolved_at
		FROM feedback
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, sortOrder, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list feedback: %w", err)
	}
	defer rows.Close()

	var feedbacks []domain.Feedback
	for rows.Next() {
		var f domain.Feedback
		if err := rows.Scan(
			&f.ID,
			&f.ProjectID,
			&f.AuthorID,
			&f.AssignedTo,
			&f.CanonicalID,
			&f.Title,
			&f.Description,
			&f.Type,
			&f.Status,
			&f.Severity,
			&f.Visibility,
			&f.VoteCount,
			&f.CommentCount,
			&f.SubmitterEmail,
			&f.SubmitterName,
			&f.Source,
			&f.SourceMetadata,
			&f.CreatedAt,
			&f.UpdatedAt,
			&f.ResolvedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan feedback: %w", err)
		}
		feedbacks = append(feedbacks, f)
	}

	return feedbacks, total, nil
}

func (r *feedbackRepository) Update(ctx context.Context, f *domain.Feedback) error {
	query := `
		UPDATE feedback
		SET
			assigned_to = $2,
			title = $3,
			description = $4,
			status = $5,
			severity = $6,
			visibility = $7,
			resolved_at = $8,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRow(ctx, query,
		f.ID,
		f.AssignedTo,
		f.Title,
		f.Description,
		f.Status,
		f.Severity,
		f.Visibility,
		f.ResolvedAt,
	).Scan(&f.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("failed to update feedback: %w", err)
	}

	return nil
}

func (r *feedbackRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM feedback WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete feedback: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *feedbackRepository) Merge(ctx context.Context, sourceID, canonicalID uuid.UUID) error {
	// Set canonical_id on source feedback and transfer votes
	query := `
		WITH updated AS (
			UPDATE feedback
			SET canonical_id = $2, updated_at = NOW()
			WHERE id = $1
			RETURNING id
		),
		vote_transfer AS (
			UPDATE votes
			SET feedback_id = $2
			WHERE feedback_id = $1 AND NOT EXISTS (
				SELECT 1 FROM votes WHERE feedback_id = $2 AND user_id = votes.user_id
			)
		)
		SELECT id FROM updated
	`

	var updatedID uuid.UUID
	err := r.db.QueryRow(ctx, query, sourceID, canonicalID).Scan(&updatedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("failed to merge feedback: %w", err)
	}

	return nil
}

func (r *feedbackRepository) AddTag(ctx context.Context, feedbackID, tagID uuid.UUID) error {
	query := `
		INSERT INTO feedback_tags (feedback_id, tag_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`

	_, err := r.db.Exec(ctx, query, feedbackID, tagID)
	if err != nil {
		return fmt.Errorf("failed to add tag: %w", err)
	}

	return nil
}

func (r *feedbackRepository) RemoveTag(ctx context.Context, feedbackID, tagID uuid.UUID) error {
	query := `DELETE FROM feedback_tags WHERE feedback_id = $1 AND tag_id = $2`

	_, err := r.db.Exec(ctx, query, feedbackID, tagID)
	if err != nil {
		return fmt.Errorf("failed to remove tag: %w", err)
	}

	return nil
}
