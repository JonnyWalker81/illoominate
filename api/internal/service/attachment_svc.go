package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/fulldisclosure/api/internal/domain"
	"github.com/fulldisclosure/api/internal/repository"
	"github.com/fulldisclosure/api/internal/storage"
)

type attachmentService struct {
	attachmentRepo repository.AttachmentRepository
	feedbackRepo   repository.FeedbackRepository
	storage        storage.ObjectStorage
	uploadExpiry   time.Duration
	downloadExpiry time.Duration
}

// NewAttachmentService creates a new attachment service
func NewAttachmentService(
	attachmentRepo repository.AttachmentRepository,
	feedbackRepo repository.FeedbackRepository,
	storage storage.ObjectStorage,
) AttachmentService {
	return &attachmentService{
		attachmentRepo: attachmentRepo,
		feedbackRepo:   feedbackRepo,
		storage:        storage,
		uploadExpiry:   15 * time.Minute,
		downloadExpiry: 1 * time.Hour,
	}
}

func (s *attachmentService) InitiateUpload(ctx context.Context, feedbackID uuid.UUID, uploaderID *uuid.UUID, filename string, contentType string, sizeBytes int64) (*UploadInfo, error) {
	// Validate content type
	if !domain.IsContentTypeAllowed(contentType) {
		return nil, domain.NewDomainError("INVALID_CONTENT_TYPE", "This file type is not allowed", 400)
	}

	// Validate file size
	if sizeBytes > domain.MaxAttachmentSize {
		return nil, domain.NewDomainError("FILE_TOO_LARGE", "File exceeds maximum size of 25MB", 400)
	}

	// Verify feedback exists
	feedback, err := s.feedbackRepo.GetByID(ctx, feedbackID)
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback: %w", err)
	}

	// Generate storage path
	attachmentID := uuid.New()
	storagePath := fmt.Sprintf("projects/%s/attachments/%s/%s", feedback.ProjectID, attachmentID, filename)

	// Create attachment record
	attachment := &domain.Attachment{
		ID:          attachmentID,
		FeedbackID:  &feedbackID,
		UploadedBy:  uploaderID,
		Filename:    filename,
		ContentType: contentType,
		SizeBytes:   sizeBytes,
		StoragePath: storagePath,
		Status:      domain.AttachmentStatusPending,
	}

	if err := s.attachmentRepo.Create(ctx, attachment); err != nil {
		return nil, fmt.Errorf("failed to create attachment record: %w", err)
	}

	// Generate signed upload URL
	uploadURL, err := s.storage.GenerateUploadURL(ctx, storagePath, contentType, s.uploadExpiry)
	if err != nil {
		// Clean up attachment record
		_ = s.attachmentRepo.Delete(ctx, attachmentID)
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return &UploadInfo{
		AttachmentID: attachmentID,
		UploadURL:    uploadURL,
		ExpiresAt:    time.Now().Add(s.uploadExpiry).Format(time.RFC3339),
	}, nil
}

func (s *attachmentService) CompleteUpload(ctx context.Context, attachmentID uuid.UUID) (*domain.Attachment, error) {
	attachment, err := s.attachmentRepo.GetByID(ctx, attachmentID)
	if err != nil {
		return nil, err
	}

	if attachment.Status != domain.AttachmentStatusPending {
		return nil, domain.NewDomainError("INVALID_STATUS", "Upload already completed or failed", 400)
	}

	// Verify file exists in storage
	exists, err := s.storage.Exists(ctx, attachment.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to verify upload: %w", err)
	}

	if !exists {
		// Mark as failed
		_ = s.attachmentRepo.UpdateStatus(ctx, attachmentID, domain.AttachmentStatusFailed)
		return nil, domain.NewDomainError("UPLOAD_NOT_FOUND", "Upload was not completed", 400)
	}

	// Mark as uploaded
	if err := s.attachmentRepo.UpdateStatus(ctx, attachmentID, domain.AttachmentStatusUploaded); err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	attachment.Status = domain.AttachmentStatusUploaded
	return attachment, nil
}

func (s *attachmentService) GetDownloadURL(ctx context.Context, attachmentID uuid.UUID) (string, error) {
	attachment, err := s.attachmentRepo.GetByID(ctx, attachmentID)
	if err != nil {
		return "", err
	}

	if attachment.Status != domain.AttachmentStatusUploaded {
		return "", domain.NewDomainError("ATTACHMENT_NOT_READY", "Attachment is not available for download", 400)
	}

	url, err := s.storage.GenerateDownloadURL(ctx, attachment.StoragePath, s.downloadExpiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	return url, nil
}

func (s *attachmentService) Delete(ctx context.Context, attachmentID uuid.UUID, actorID uuid.UUID) error {
	attachment, err := s.attachmentRepo.GetByID(ctx, attachmentID)
	if err != nil {
		return err
	}

	// Delete from storage
	if err := s.storage.Delete(ctx, attachment.StoragePath); err != nil {
		return fmt.Errorf("failed to delete from storage: %w", err)
	}

	// Mark as deleted in DB
	return s.attachmentRepo.UpdateStatus(ctx, attachmentID, domain.AttachmentStatusDeleted)
}
