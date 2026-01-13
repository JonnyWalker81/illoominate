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

type membershipRepository struct {
	db DBTX
}

// NewMembershipRepository creates a new membership repository
func NewMembershipRepository(db *pgxpool.Pool) MembershipRepository {
	return &membershipRepository{db: db}
}

func (r *membershipRepository) Create(ctx context.Context, m *domain.Membership) error {
	query := `
		INSERT INTO memberships (id, project_id, user_id, role, display_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		m.ID,
		m.ProjectID,
		m.UserID,
		m.Role,
		m.DisplayName,
	).Scan(&m.CreatedAt, &m.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create membership: %w", err)
	}

	return nil
}

func (r *membershipRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Membership, error) {
	query := `
		SELECT id, project_id, user_id, role, display_name, created_at, updated_at
		FROM memberships
		WHERE id = $1
	`

	var m domain.Membership
	err := r.db.QueryRow(ctx, query, id).Scan(
		&m.ID,
		&m.ProjectID,
		&m.UserID,
		&m.Role,
		&m.DisplayName,
		&m.CreatedAt,
		&m.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get membership: %w", err)
	}

	return &m, nil
}

func (r *membershipRepository) GetByProjectAndUser(ctx context.Context, projectID, userID string) (*domain.Membership, error) {
	query := `
		SELECT id, project_id, user_id, role, display_name, created_at, updated_at
		FROM memberships
		WHERE project_id = $1 AND user_id = $2
	`

	var m domain.Membership
	err := r.db.QueryRow(ctx, query, projectID, userID).Scan(
		&m.ID,
		&m.ProjectID,
		&m.UserID,
		&m.Role,
		&m.DisplayName,
		&m.CreatedAt,
		&m.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Not found, but not an error - user just isn't a member
		}
		return nil, fmt.Errorf("failed to get membership: %w", err)
	}

	return &m, nil
}

func (r *membershipRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Membership, error) {
	query := `
		SELECT id, project_id, user_id, role, display_name, created_at, updated_at
		FROM memberships
		WHERE project_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list memberships: %w", err)
	}
	defer rows.Close()

	var memberships []domain.Membership
	for rows.Next() {
		var m domain.Membership
		if err := rows.Scan(
			&m.ID,
			&m.ProjectID,
			&m.UserID,
			&m.Role,
			&m.DisplayName,
			&m.CreatedAt,
			&m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan membership: %w", err)
		}
		memberships = append(memberships, m)
	}

	return memberships, nil
}

func (r *membershipRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Membership, error) {
	query := `
		SELECT m.id, m.project_id, m.user_id, m.role, m.display_name, m.created_at, m.updated_at
		FROM memberships m
		WHERE m.user_id = $1
		ORDER BY m.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user memberships: %w", err)
	}
	defer rows.Close()

	var memberships []domain.Membership
	for rows.Next() {
		var m domain.Membership
		if err := rows.Scan(
			&m.ID,
			&m.ProjectID,
			&m.UserID,
			&m.Role,
			&m.DisplayName,
			&m.CreatedAt,
			&m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan membership: %w", err)
		}
		memberships = append(memberships, m)
	}

	return memberships, nil
}

func (r *membershipRepository) Update(ctx context.Context, m *domain.Membership) error {
	query := `
		UPDATE memberships
		SET role = $2, display_name = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRow(ctx, query,
		m.ID,
		m.Role,
		m.DisplayName,
	).Scan(&m.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("failed to update membership: %w", err)
	}

	return nil
}

func (r *membershipRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM memberships WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete membership: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
