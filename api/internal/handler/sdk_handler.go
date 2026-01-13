package handler

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/fulldisclosure/api/internal/auth"
	"github.com/fulldisclosure/api/internal/domain"
	"github.com/fulldisclosure/api/internal/service"
)

// SDKHandler handles SDK endpoints
type SDKHandler struct {
	feedbackSvc   service.FeedbackService
	attachmentSvc service.AttachmentService
}

// NewSDKHandler creates a new SDK handler
func NewSDKHandler(
	feedbackSvc service.FeedbackService,
	attachmentSvc service.AttachmentService,
) *SDKHandler {
	return &SDKHandler{
		feedbackSvc:   feedbackSvc,
		attachmentSvc: attachmentSvc,
	}
}

// SubmitFeedback handles POST /sdk/feedback
func (h *SDKHandler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	// Get project ID from SDK auth context
	projectID, ok := auth.SDKProjectFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "SDK authentication required")
		return
	}

	// Optionally get user ID if authenticated via Supabase JWT
	var authorID *uuid.UUID
	if userID, ok := auth.UserIDFromContext(r.Context()); ok {
		authorID = &userID
	}

	var req struct {
		Title          string                 `json:"title"`
		Description    string                 `json:"description"`
		Type           domain.FeedbackType    `json:"type"`
		SubmitterEmail *string                `json:"submitter_email"`
		SubmitterName  *string                `json:"submitter_name"`
		SourceMetadata map[string]interface{} `json:"source_metadata"`
	}

	if err := DecodeJSON(r, &req); err != nil {
		HandleError(w, err)
		return
	}

	// Validation
	errors := make(map[string]string)
	if req.Title == "" {
		errors["title"] = "Title is required"
	}
	if req.Description == "" {
		errors["description"] = "Description is required"
	}
	if req.Type == "" {
		req.Type = domain.FeedbackTypeGeneral
	}
	if len(errors) > 0 {
		ValidationError(w, errors)
		return
	}

	// Determine source from header
	source := "sdk"
	if sdkSource := r.Header.Get("X-SDK-Source"); sdkSource != "" {
		source = "sdk-" + sdkSource // e.g., "sdk-ios", "sdk-android", "sdk-web"
	}

	feedback, err := h.feedbackSvc.Create(r.Context(), service.CreateFeedbackRequest{
		ProjectID:      projectID,
		AuthorID:       authorID,
		Title:          req.Title,
		Description:    req.Description,
		Type:           req.Type,
		SubmitterEmail: req.SubmitterEmail,
		SubmitterName:  req.SubmitterName,
		Source:         source,
		SourceMetadata: req.SourceMetadata,
	})
	if err != nil {
		HandleError(w, err)
		return
	}

	Created(w, map[string]interface{}{
		"id":         feedback.ID,
		"created_at": feedback.CreatedAt,
	})
}

// InitiateUpload handles POST /sdk/attachments/init
func (h *SDKHandler) InitiateUpload(w http.ResponseWriter, r *http.Request) {
	// Get uploader ID if authenticated
	var uploaderID *uuid.UUID
	if userID, ok := auth.UserIDFromContext(r.Context()); ok {
		uploaderID = &userID
	}

	var req struct {
		FeedbackID  uuid.UUID `json:"feedback_id"`
		Filename    string    `json:"filename"`
		ContentType string    `json:"content_type"`
		SizeBytes   int64     `json:"size_bytes"`
	}

	if err := DecodeJSON(r, &req); err != nil {
		HandleError(w, err)
		return
	}

	// Validation
	errors := make(map[string]string)
	if req.FeedbackID == uuid.Nil {
		errors["feedback_id"] = "Feedback ID is required"
	}
	if req.Filename == "" {
		errors["filename"] = "Filename is required"
	}
	if req.ContentType == "" {
		errors["content_type"] = "Content type is required"
	}
	if req.SizeBytes <= 0 {
		errors["size_bytes"] = "Size must be greater than 0"
	}
	if len(errors) > 0 {
		ValidationError(w, errors)
		return
	}

	uploadInfo, err := h.attachmentSvc.InitiateUpload(
		r.Context(),
		req.FeedbackID,
		uploaderID,
		req.Filename,
		req.ContentType,
		req.SizeBytes,
	)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, uploadInfo)
}

// CompleteUpload handles POST /sdk/attachments/complete
func (h *SDKHandler) CompleteUpload(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AttachmentID uuid.UUID `json:"attachment_id"`
	}

	if err := DecodeJSON(r, &req); err != nil {
		HandleError(w, err)
		return
	}

	if req.AttachmentID == uuid.Nil {
		ValidationError(w, map[string]string{"attachment_id": "Attachment ID is required"})
		return
	}

	attachment, err := h.attachmentSvc.CompleteUpload(r.Context(), req.AttachmentID)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, attachment)
}
