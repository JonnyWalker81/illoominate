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

type attachmentRepository struct {
	db DBTX
}

// NewAttachmentRepository creates a new attachment repository
func NewAttachmentRepository(db *pgxpool.Pool) AttachmentRepository {
	return &attachmentRepository{db: db}
}

func (r *attachmentRepository) Create(ctx context.Context, a *domain.Attachment) error {
	query := `
		INSERT INTO attachments (id, feedback_id, uploaded_by, filename, content_type, size_bytes, storage_path, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		RETURNING created_at
	`

	err := r.db.QueryRow(ctx, query,
		a.ID,
		a.FeedbackID,
		a.UploadedBy,
		a.Filename,
		a.ContentType,
		a.SizeBytes,
		a.StoragePath,
		a.Status,
	).Scan(&a.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create attachment: %w", err)
	}

	return nil
}

func (r *attachmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Attachment, error) {
	query := `
		SELECT id, feedback_id, uploaded_by, filename, content_type, size_bytes, storage_path, status, created_at
		FROM attachments
		WHERE id = $1
	`

	var a domain.Attachment
	err := r.db.QueryRow(ctx, query, id).Scan(
		&a.ID,
		&a.FeedbackID,
		&a.UploadedBy,
		&a.Filename,
		&a.ContentType,
		&a.SizeBytes,
		&a.StoragePath,
		&a.Status,
		&a.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get attachment: %w", err)
	}

	return &a, nil
}

func (r *attachmentRepository) ListByFeedback(ctx context.Context, feedbackID uuid.UUID) ([]domain.Attachment, error) {
	query := `
		SELECT id, feedback_id, uploaded_by, filename, content_type, size_bytes, storage_path, status, created_at
		FROM attachments
		WHERE feedback_id = $1 AND status = 'uploaded'
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, feedbackID)
	if err != nil {
		return nil, fmt.Errorf("failed to list attachments: %w", err)
	}
	defer rows.Close()

	var attachments []domain.Attachment
	for rows.Next() {
		var a domain.Attachment
		if err := rows.Scan(
			&a.ID,
			&a.FeedbackID,
			&a.UploadedBy,
			&a.Filename,
			&a.ContentType,
			&a.SizeBytes,
			&a.StoragePath,
			&a.Status,
			&a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan attachment: %w", err)
		}
		attachments = append(attachments, a)
	}

	return attachments, nil
}

func (r *attachmentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.AttachmentStatus) error {
	query := `
		UPDATE attachments
		SET status = $2
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update attachment status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *attachmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM attachments WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete attachment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
