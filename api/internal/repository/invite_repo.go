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

type inviteRepository struct {
	db DBTX
}

// NewInviteRepository creates a new invite repository
func NewInviteRepository(db *pgxpool.Pool) InviteRepository {
	return &inviteRepository{db: db}
}

func (r *inviteRepository) Create(ctx context.Context, i *domain.Invite) error {
	query := `
		INSERT INTO invites (id, project_id, invited_by, email, role, token, status, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		RETURNING created_at
	`

	err := r.db.QueryRow(ctx, query,
		i.ID,
		i.ProjectID,
		i.InvitedBy,
		i.Email,
		i.Role,
		i.Token,
		i.Status,
		i.ExpiresAt,
	).Scan(&i.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create invite: %w", err)
	}

	return nil
}

func (r *inviteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Invite, error) {
	query := `
		SELECT id, project_id, invited_by, email, role, token, status, expires_at, created_at, accepted_at
		FROM invites
		WHERE id = $1
	`

	var i domain.Invite
	err := r.db.QueryRow(ctx, query, id).Scan(
		&i.ID,
		&i.ProjectID,
		&i.InvitedBy,
		&i.Email,
		&i.Role,
		&i.Token,
		&i.Status,
		&i.ExpiresAt,
		&i.CreatedAt,
		&i.AcceptedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get invite: %w", err)
	}

	return &i, nil
}

func (r *inviteRepository) GetByToken(ctx context.Context, token string) (*domain.Invite, error) {
	query := `
		SELECT id, project_id, invited_by, email, role, token, status, expires_at, created_at, accepted_at
		FROM invites
		WHERE token = $1
	`

	var i domain.Invite
	err := r.db.QueryRow(ctx, query, token).Scan(
		&i.ID,
		&i.ProjectID,
		&i.InvitedBy,
		&i.Email,
		&i.Role,
		&i.Token,
		&i.Status,
		&i.ExpiresAt,
		&i.CreatedAt,
		&i.AcceptedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get invite by token: %w", err)
	}

	return &i, nil
}

func (r *inviteRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Invite, error) {
	query := `
		SELECT id, project_id, invited_by, email, role, token, status, expires_at, created_at, accepted_at
		FROM invites
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list invites: %w", err)
	}
	defer rows.Close()

	var invites []domain.Invite
	for rows.Next() {
		var i domain.Invite
		if err := rows.Scan(
			&i.ID,
			&i.ProjectID,
			&i.InvitedBy,
			&i.Email,
			&i.Role,
			&i.Token,
			&i.Status,
			&i.ExpiresAt,
			&i.CreatedAt,
			&i.AcceptedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan invite: %w", err)
		}
		invites = append(invites, i)
	}

	return invites, nil
}

func (r *inviteRepository) Update(ctx context.Context, i *domain.Invite) error {
	query := `
		UPDATE invites
		SET status = $2, accepted_at = $3
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		i.ID,
		i.Status,
		i.AcceptedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update invite: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *inviteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM invites WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete invite: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *inviteRepository) DeleteExpired(ctx context.Context) (int, error) {
	query := `
		DELETE FROM invites
		WHERE status = 'pending' AND expires_at < NOW()
	`

	result, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired invites: %w", err)
	}

	return int(result.RowsAffected()), nil
}
