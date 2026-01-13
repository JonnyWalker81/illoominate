package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/fulldisclosure/api/internal/auth"
	"github.com/fulldisclosure/api/internal/domain"
	"github.com/fulldisclosure/api/internal/service"
)

// CommunityHandler handles community portal endpoints
type CommunityHandler struct {
	feedbackSvc service.FeedbackService
	voteSvc     service.VoteService
	commentSvc  service.CommentService
}

// NewCommunityHandler creates a new community handler
func NewCommunityHandler(
	feedbackSvc service.FeedbackService,
	voteSvc service.VoteService,
	commentSvc service.CommentService,
) *CommunityHandler {
	return &CommunityHandler{
		feedbackSvc: feedbackSvc,
		voteSvc:     voteSvc,
		commentSvc:  commentSvc,
	}
}

// ListFeatures handles GET /community/projects/:projectId/feature-requests
func (h *CommunityHandler) ListFeatures(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "Invalid project ID")
		return
	}

	membership := auth.MustMembershipFromContext(r.Context())

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "vote_count"
	}

	status := r.URL.Query().Get("status")

	filter := service.FeedbackFilter{
		Page:      page,
		PerPage:   20,
		SortBy:    sortBy,
		SortOrder: "desc",
	}

	// Only show feature requests
	featureType := domain.FeedbackTypeFeature
	filter.Type = &featureType

	// Parse status filter
	if status != "" {
		s := domain.FeedbackStatus(status)
		filter.Status = &s
	}

	result, err := h.feedbackSvc.List(r.Context(), projectID, filter, membership.Role)
	if err != nil {
		HandleError(w, err)
		return
	}

	// Enrich with vote status if authenticated
	userID, ok := auth.UserIDFromContext(r.Context())
	if ok {
		if err := h.voteSvc.EnrichWithVoteStatus(r.Context(), result.Feedbacks, userID); err != nil {
			HandleError(w, err)
			return
		}
	}

	Paginated(w, result.Feedbacks, result.Total, result.Page, result.PerPage, result.TotalPages)
}

// GetFeature handles GET /community/projects/:projectId/feature-requests/:feedbackId
func (h *CommunityHandler) GetFeature(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "Invalid project ID")
		return
	}

	feedbackID, err := uuid.Parse(chi.URLParam(r, "feedbackId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_FEEDBACK_ID", "Invalid feedback ID")
		return
	}

	membership := auth.MustMembershipFromContext(r.Context())

	feedback, err := h.feedbackSvc.GetByID(r.Context(), projectID, feedbackID, membership.Role)
	if err != nil {
		HandleError(w, err)
		return
	}

	// Only return feature requests via community endpoint
	if feedback.Type != domain.FeedbackTypeFeature {
		Error(w, http.StatusNotFound, "NOT_FOUND", "Feature request not found")
		return
	}

	// Enrich with vote status
	userID, ok := auth.UserIDFromContext(r.Context())
	if ok {
		hasVoted, err := h.voteSvc.HasVoted(r.Context(), feedbackID, userID)
		if err == nil {
			feedback.HasVoted = hasVoted
		}
	}

	JSON(w, http.StatusOK, feedback)
}

// Vote handles POST /community/projects/:projectId/feature-requests/:feedbackId/vote
func (h *CommunityHandler) Vote(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "Invalid project ID")
		return
	}

	feedbackID, err := uuid.Parse(chi.URLParam(r, "feedbackId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_FEEDBACK_ID", "Invalid feedback ID")
		return
	}

	userID := auth.MustUserIDFromContext(r.Context())

	result, err := h.voteSvc.Vote(r.Context(), projectID, feedbackID, userID)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, result)
}

// Unvote handles DELETE /community/projects/:projectId/feature-requests/:feedbackId/vote
func (h *CommunityHandler) Unvote(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "Invalid project ID")
		return
	}

	feedbackID, err := uuid.Parse(chi.URLParam(r, "feedbackId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_FEEDBACK_ID", "Invalid feedback ID")
		return
	}

	userID := auth.MustUserIDFromContext(r.Context())

	result, err := h.voteSvc.Unvote(r.Context(), projectID, feedbackID, userID)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, result)
}

// ListComments handles GET /community/projects/:projectId/feature-requests/:feedbackId/comments
func (h *CommunityHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	feedbackID, err := uuid.Parse(chi.URLParam(r, "feedbackId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_FEEDBACK_ID", "Invalid feedback ID")
		return
	}

	// Community users can only see COMMUNITY visibility comments
	comments, err := h.commentSvc.ListByFeedback(r.Context(), feedbackID, false)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, comments)
}

// CreateComment handles POST /community/projects/:projectId/feature-requests/:feedbackId/comments
func (h *CommunityHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	feedbackID, err := uuid.Parse(chi.URLParam(r, "feedbackId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_FEEDBACK_ID", "Invalid feedback ID")
		return
	}

	userID := auth.MustUserIDFromContext(r.Context())

	var req struct {
		Body string `json:"body"`
	}

	if err := DecodeJSON(r, &req); err != nil {
		HandleError(w, err)
		return
	}

	if req.Body == "" {
		ValidationError(w, map[string]string{"body": "Body is required"})
		return
	}

	comment, err := h.commentSvc.Create(r.Context(), service.CreateCommentRequest{
		FeedbackID: feedbackID,
		AuthorID:   userID,
		Body:       req.Body,
		Visibility: domain.VisibilityCommunity, // Community comments are always public
	})
	if err != nil {
		HandleError(w, err)
		return
	}

	Created(w, comment)
}
