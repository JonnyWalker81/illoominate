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

type voteRepository struct {
	db DBTX
}

// NewVoteRepository creates a new vote repository
func NewVoteRepository(db *pgxpool.Pool) VoteRepository {
	return &voteRepository{db: db}
}

func (r *voteRepository) Create(ctx context.Context, v *domain.Vote) error {
	query := `
		INSERT INTO votes (id, feedback_id, user_id, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (feedback_id, user_id) DO NOTHING
		RETURNING created_at
	`

	err := r.db.QueryRow(ctx, query,
		v.ID,
		v.FeedbackID,
		v.UserID,
	).Scan(&v.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Vote already exists - not an error, just idempotent
			return nil
		}
		return fmt.Errorf("failed to create vote: %w", err)
	}

	return nil
}

func (r *voteRepository) Delete(ctx context.Context, feedbackID, userID uuid.UUID) error {
	query := `DELETE FROM votes WHERE feedback_id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, feedbackID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete vote: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *voteRepository) Exists(ctx context.Context, feedbackID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM votes WHERE feedback_id = $1 AND user_id = $2)`

	var exists bool
	err := r.db.QueryRow(ctx, query, feedbackID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check vote: %w", err)
	}

	return exists, nil
}

func (r *voteRepository) CountByFeedback(ctx context.Context, feedbackID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM votes WHERE feedback_id = $1`

	var count int
	err := r.db.QueryRow(ctx, query, feedbackID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count votes: %w", err)
	}

	return count, nil
}

func (r *voteRepository) ListByUser(ctx context.Context, userID uuid.UUID, feedbackIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	if len(feedbackIDs) == 0 {
		return make(map[uuid.UUID]bool), nil
	}

	query := `SELECT feedback_id FROM votes WHERE user_id = $1 AND feedback_id = ANY($2)`

	rows, err := r.db.Query(ctx, query, userID, feedbackIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list votes: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]bool)
	for rows.Next() {
		var feedbackID uuid.UUID
		if err := rows.Scan(&feedbackID); err != nil {
			return nil, fmt.Errorf("failed to scan vote: %w", err)
		}
		result[feedbackID] = true
	}

	return result, nil
}
