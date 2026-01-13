package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fulldisclosure/api/internal/domain"
)

type portalRepository struct {
	db DBTX
}

// NewPortalRepository creates a new portal repository
func NewPortalRepository(db *pgxpool.Pool) PortalRepository {
	return &portalRepository{db: db}
}

// NewPortalRepositoryWithTx creates a portal repository with a transaction
func NewPortalRepositoryWithTx(tx DBTX) PortalRepository {
	return &portalRepository{db: tx}
}

func (r *portalRepository) CreateProfile(ctx context.Context, userID, projectID uuid.UUID) error {
	query := `
		INSERT INTO portal_user_profiles (user_id, project_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, project_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, query, userID, projectID)
	if err != nil {
		return fmt.Errorf("failed to create portal profile: %w", err)
	}

	return nil
}

func (r *portalRepository) GetProfile(ctx context.Context, userID, projectID uuid.UUID) (*domain.PortalUserProfile, error) {
	query := `
		SELECT id, user_id, project_id, notification_preferences, created_at, updated_at
		FROM portal_user_profiles
		WHERE user_id = $1 AND project_id = $2
	`

	var profile domain.PortalUserProfile
	var prefsJSON []byte

	err := r.db.QueryRow(ctx, query, userID, projectID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.ProjectID,
		&prefsJSON,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get portal profile: %w", err)
	}

	if err := json.Unmarshal(prefsJSON, &profile.NotificationPreferences); err != nil {
		// Use defaults if JSON is malformed
		profile.NotificationPreferences = domain.DefaultPortalNotificationPreferences()
	}

	return &profile, nil
}

func (r *portalRepository) UpdateNotificationPrefs(ctx context.Context, userID, projectID uuid.UUID, prefs domain.PortalNotificationPreferences) error {
	prefsJSON, err := json.Marshal(prefs)
	if err != nil {
		return fmt.Errorf("failed to marshal notification preferences: %w", err)
	}

	query := `
		UPDATE portal_user_profiles
		SET notification_preferences = $1, updated_at = NOW()
		WHERE user_id = $2 AND project_id = $3
	`

	result, err := r.db.Exec(ctx, query, prefsJSON, userID, projectID)
	if err != nil {
		return fmt.Errorf("failed to update notification preferences: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *portalRepository) LinkSDKUsersByEmail(ctx context.Context, userID, projectID uuid.UUID, email string) (int64, error) {
	query := `
		UPDATE sdk_users
		SET linked_user_id = $1, updated_at = NOW()
		WHERE project_id = $2
		AND LOWER(email) = LOWER($3)
		AND linked_user_id IS NULL
	`

	result, err := r.db.Exec(ctx, query, userID, projectID, email)
	if err != nil {
		return 0, fmt.Errorf("failed to link SDK users: %w", err)
	}

	return result.RowsAffected(), nil
}

func (r *portalRepository) GetLinkedFeedback(ctx context.Context, userID, projectID uuid.UUID) ([]domain.PortalFeedbackSummary, error) {
	query := `
		SELECT
			f.id, f.title, f.description, f.type, f.status,
			f.vote_count, f.comment_count, f.created_at, f.updated_at,
			EXISTS(SELECT 1 FROM portal_votes pv WHERE pv.feedback_id = f.id AND pv.user_id = $1) as has_voted
		FROM feedback f
		WHERE f.project_id = $2
		AND f.canonical_id IS NULL
		AND f.sdk_user_id IN (
			SELECT id FROM sdk_users WHERE linked_user_id = $1
		)
		ORDER BY f.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get linked feedback: %w", err)
	}
	defer rows.Close()

	var feedback []domain.PortalFeedbackSummary
	for rows.Next() {
		var f domain.PortalFeedbackSummary
		if err := rows.Scan(
			&f.ID, &f.Title, &f.Description, &f.Type, &f.Status,
			&f.VoteCount, &f.CommentCount, &f.CreatedAt, &f.UpdatedAt,
			&f.HasVoted,
		); err != nil {
			return nil, fmt.Errorf("failed to scan feedback: %w", err)
		}
		feedback = append(feedback, f)
	}

	return feedback, nil
}

func (r *portalRepository) ListPublicFeatures(ctx context.Context, projectID uuid.UUID, userID *uuid.UUID, limit, offset int) ([]domain.PortalFeedbackSummary, int, error) {
	// Build has_voted subquery based on whether user is provided
	var hasVotedExpr string
	var args []interface{}
	argIndex := 1

	args = append(args, projectID)
	argIndex++

	if userID != nil {
		hasVotedExpr = fmt.Sprintf("EXISTS(SELECT 1 FROM portal_votes pv WHERE pv.feedback_id = f.id AND pv.user_id = $%d)", argIndex)
		args = append(args, *userID)
		argIndex++
	} else {
		hasVotedExpr = "false"
	}

	// Count query
	countQuery := `
		SELECT COUNT(*)
		FROM feedback f
		WHERE f.project_id = $1
		AND f.canonical_id IS NULL
		AND f.type = 'feature'
		AND f.visibility = 'public'
	`

	var total int
	if err := r.db.QueryRow(ctx, countQuery, projectID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count features: %w", err)
	}

	// Main query
	query := fmt.Sprintf(`
		SELECT
			f.id, f.title, f.description, f.type, f.status,
			f.vote_count, f.comment_count, f.created_at, f.updated_at,
			%s as has_voted
		FROM feedback f
		WHERE f.project_id = $1
		AND f.canonical_id IS NULL
		AND f.type = 'feature'
		AND f.visibility = 'public'
		ORDER BY f.vote_count DESC, f.created_at DESC
		LIMIT $%d OFFSET $%d
	`, hasVotedExpr, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list features: %w", err)
	}
	defer rows.Close()

	var feedback []domain.PortalFeedbackSummary
	for rows.Next() {
		var f domain.PortalFeedbackSummary
		if err := rows.Scan(
			&f.ID, &f.Title, &f.Description, &f.Type, &f.Status,
			&f.VoteCount, &f.CommentCount, &f.CreatedAt, &f.UpdatedAt,
			&f.HasVoted,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan feature: %w", err)
		}
		feedback = append(feedback, f)
	}

	return feedback, total, nil
}

func (r *portalRepository) CreateVote(ctx context.Context, feedbackID, userID, projectID uuid.UUID) error {
	query := `
		INSERT INTO portal_votes (id, feedback_id, user_id, project_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (feedback_id, user_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, query, uuid.New(), feedbackID, userID, projectID)
	if err != nil {
		if strings.Contains(err.Error(), "fk_portal_votes_feedback") {
			return domain.ErrNotFound
		}
		return fmt.Errorf("failed to create vote: %w", err)
	}

	// Update vote count on feedback
	updateQuery := `
		UPDATE feedback
		SET vote_count = (
			SELECT COUNT(*) FROM portal_votes WHERE feedback_id = $1
		) + (
			SELECT COUNT(*) FROM votes WHERE feedback_id = $1
		)
		WHERE id = $1
	`

	_, err = r.db.Exec(ctx, updateQuery, feedbackID)
	if err != nil {
		return fmt.Errorf("failed to update vote count: %w", err)
	}

	return nil
}

func (r *portalRepository) DeleteVote(ctx context.Context, feedbackID, userID uuid.UUID) error {
	query := `DELETE FROM portal_votes WHERE feedback_id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, feedbackID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete vote: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	// Update vote count on feedback
	updateQuery := `
		UPDATE feedback
		SET vote_count = (
			SELECT COUNT(*) FROM portal_votes WHERE feedback_id = $1
		) + (
			SELECT COUNT(*) FROM votes WHERE feedback_id = $1
		)
		WHERE id = $1
	`

	_, err = r.db.Exec(ctx, updateQuery, feedbackID)
	if err != nil {
		return fmt.Errorf("failed to update vote count: %w", err)
	}

	return nil
}

func (r *portalRepository) HasVoted(ctx context.Context, feedbackID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM portal_votes WHERE feedback_id = $1 AND user_id = $2)`

	var exists bool
	err := r.db.QueryRow(ctx, query, feedbackID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check vote: %w", err)
	}

	return exists, nil
}

func (r *portalRepository) GetVotedFeedbackIDs(ctx context.Context, userID uuid.UUID, feedbackIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	if len(feedbackIDs) == 0 {
		return make(map[uuid.UUID]bool), nil
	}

	query := `SELECT feedback_id FROM portal_votes WHERE user_id = $1 AND feedback_id = ANY($2)`

	rows, err := r.db.Query(ctx, query, userID, feedbackIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get voted feedback IDs: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]bool)
	for rows.Next() {
		var feedbackID uuid.UUID
		if err := rows.Scan(&feedbackID); err != nil {
			return nil, fmt.Errorf("failed to scan feedback ID: %w", err)
		}
		result[feedbackID] = true
	}

	return result, nil
}
