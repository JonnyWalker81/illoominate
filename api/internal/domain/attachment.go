package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AttachmentStatus represents the upload status of an attachment
type AttachmentStatus string

const (
	AttachmentStatusPending  AttachmentStatus = "pending"
	AttachmentStatusUploaded AttachmentStatus = "uploaded"
	AttachmentStatusFailed   AttachmentStatus = "failed"
	AttachmentStatusDeleted  AttachmentStatus = "deleted"
)

// IsValid checks if the attachment status is valid
func (s AttachmentStatus) IsValid() bool {
	return s == AttachmentStatusPending || s == AttachmentStatusUploaded ||
		s == AttachmentStatusFailed || s == AttachmentStatusDeleted
}

// Attachment represents a file attached to feedback or a comment
type Attachment struct {
	ID              uuid.UUID        `json:"id"`
	FeedbackID      *uuid.UUID       `json:"feedback_id,omitempty"`
	CommentID       *uuid.UUID       `json:"comment_id,omitempty"`
	UploadedBy      *uuid.UUID       `json:"uploaded_by,omitempty"`
	Filename        string           `json:"filename"`
	ContentType     string           `json:"content_type"`
	SizeBytes       int64            `json:"size_bytes"`
	StoragePath     string           `json:"-"` // Path in cloud storage
	GCSBucket       string           `json:"-"`
	GCSPath         string           `json:"-"`
	Status          AttachmentStatus `json:"status"`
	UploadExpiresAt *time.Time       `json:"upload_expires_at,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UploadedAt      *time.Time       `json:"uploaded_at,omitempty"`

	// URLs (populated by service layer, not stored)
	DownloadURL string `json:"download_url,omitempty"`
}

// AllowedContentTypes defines the allowlist of content types for uploads
var AllowedContentTypes = map[string]bool{
	// Images
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
	"image/svg+xml": true,

	// Videos
	"video/mp4":  true,
	"video/webm": true,
	"video/quicktime": true,

	// Documents
	"application/pdf": true,
	"text/plain":      true,
}

// MaxAttachmentSize is the maximum allowed attachment size (25MB)
const MaxAttachmentSize = 25 * 1024 * 1024

// IsContentTypeAllowed checks if the content type is in the allowlist
func IsContentTypeAllowed(contentType string) bool {
	// Check exact match
	if AllowedContentTypes[contentType] {
		return true
	}
	// Check wildcard patterns (e.g., image/*)
	parts := strings.Split(contentType, "/")
	if len(parts) == 2 {
		return AllowedContentTypes[parts[0]+"/*"]
	}
	return false
}

// Validate validates the attachment data
func (a *Attachment) Validate() error {
	if a.Filename == "" {
		return fmt.Errorf("filename is required")
	}
	if len(a.Filename) > 255 {
		return fmt.Errorf("filename must be 255 characters or less")
	}
	if a.ContentType == "" {
		return fmt.Errorf("content_type is required")
	}
	if !IsContentTypeAllowed(a.ContentType) {
		return fmt.Errorf("content type not allowed: %s", a.ContentType)
	}
	if a.SizeBytes <= 0 {
		return fmt.Errorf("size_bytes must be positive")
	}
	if a.SizeBytes > MaxAttachmentSize {
		return fmt.Errorf("file size exceeds maximum allowed (%d bytes)", MaxAttachmentSize)
	}
	return nil
}

// IsUploaded returns true if the attachment has been uploaded
func (a *Attachment) IsUploaded() bool {
	return a.Status == AttachmentStatusUploaded
}

// IsPending returns true if the attachment is pending upload
func (a *Attachment) IsPending() bool {
	return a.Status == AttachmentStatusPending
}

// AttachmentUploadResult contains the result of initiating an upload
type AttachmentUploadResult struct {
	AttachmentID uuid.UUID `json:"attachment_id"`
	UploadURL    string    `json:"upload_url"`
	ExpiresAt    time.Time `json:"expires_at"`
}
