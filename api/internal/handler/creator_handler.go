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

// CreatorHandler handles creator console endpoints
type CreatorHandler struct {
	feedbackSvc   service.FeedbackService
	voteSvc       service.VoteService
	commentSvc    service.CommentService
	tagSvc        service.TagService
	membershipSvc service.MembershipService
	inviteSvc     service.InviteService
	projectSvc    service.ProjectService
}

// NewCreatorHandler creates a new creator handler
func NewCreatorHandler(
	feedbackSvc service.FeedbackService,
	voteSvc service.VoteService,
	commentSvc service.CommentService,
	tagSvc service.TagService,
	membershipSvc service.MembershipService,
	inviteSvc service.InviteService,
	projectSvc service.ProjectService,
) *CreatorHandler {
	return &CreatorHandler{
		feedbackSvc:   feedbackSvc,
		voteSvc:       voteSvc,
		commentSvc:    commentSvc,
		tagSvc:        tagSvc,
		membershipSvc: membershipSvc,
		inviteSvc:     inviteSvc,
		projectSvc:    projectSvc,
	}
}

// ListFeedback handles GET /creator/projects/:projectId/feedback
func (h *CreatorHandler) ListFeedback(w http.ResponseWriter, r *http.Request) {
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

	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 50
	}

	filter := service.FeedbackFilter{
		Page:      page,
		PerPage:   perPage,
		SortBy:    r.URL.Query().Get("sort"),
		SortOrder: r.URL.Query().Get("order"),
		Search:    strPtr(r.URL.Query().Get("search")),
	}

	// Parse type filter
	if t := r.URL.Query().Get("type"); t != "" {
		ft := domain.FeedbackType(t)
		filter.Type = &ft
	}

	// Parse status filter
	if s := r.URL.Query().Get("status"); s != "" {
		fs := domain.FeedbackStatus(s)
		filter.Status = &fs
	}

	// Parse visibility filter
	if v := r.URL.Query().Get("visibility"); v != "" {
		fv := domain.Visibility(v)
		filter.Visibility = &fv
	}

	result, err := h.feedbackSvc.List(r.Context(), projectID, filter, membership.Role)
	if err != nil {
		HandleError(w, err)
		return
	}

	Paginated(w, result.Feedbacks, result.Total, result.Page, result.PerPage, result.TotalPages)
}

// GetFeedback handles GET /creator/projects/:projectId/feedback/:feedbackId
func (h *CreatorHandler) GetFeedback(w http.ResponseWriter, r *http.Request) {
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

	JSON(w, http.StatusOK, feedback)
}

// CreateFeedback handles POST /creator/projects/:projectId/feedback
func (h *CreatorHandler) CreateFeedback(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "Invalid project ID")
		return
	}

	userID := auth.MustUserIDFromContext(r.Context())

	var req struct {
		Title       string               `json:"title"`
		Description string               `json:"description"`
		Type        domain.FeedbackType  `json:"type"`
		Visibility  domain.Visibility    `json:"visibility"`
		Severity    *domain.Severity     `json:"severity"`
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
		errors["type"] = "Type is required"
	}
	if len(errors) > 0 {
		ValidationError(w, errors)
		return
	}

	feedback, err := h.feedbackSvc.Create(r.Context(), service.CreateFeedbackRequest{
		ProjectID:   projectID,
		AuthorID:    &userID,
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Visibility:  req.Visibility,
		Source:      "web",
	})
	if err != nil {
		HandleError(w, err)
		return
	}

	Created(w, feedback)
}

// UpdateFeedback handles PATCH /creator/projects/:projectId/feedback/:feedbackId
func (h *CreatorHandler) UpdateFeedback(w http.ResponseWriter, r *http.Request) {
	feedbackID, err := uuid.Parse(chi.URLParam(r, "feedbackId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_FEEDBACK_ID", "Invalid feedback ID")
		return
	}

	userID := auth.MustUserIDFromContext(r.Context())

	var req struct {
		Title       *string               `json:"title"`
		Description *string               `json:"description"`
		Status      *domain.FeedbackStatus `json:"status"`
		Severity    *domain.Severity      `json:"severity"`
		Visibility  *domain.Visibility    `json:"visibility"`
		AssignedTo  *uuid.UUID            `json:"assigned_to"`
		TagIDs      []uuid.UUID           `json:"tag_ids"`
	}

	if err := DecodeJSON(r, &req); err != nil {
		HandleError(w, err)
		return
	}

	feedback, err := h.feedbackSvc.Update(r.Context(), feedbackID, service.UpdateFeedbackRequest{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Severity:    req.Severity,
		Visibility:  req.Visibility,
		AssignedTo:  req.AssignedTo,
		TagIDs:      req.TagIDs,
	}, userID)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, feedback)
}

// MergeFeedback handles POST /creator/projects/:projectId/feedback/:feedbackId/merge
func (h *CreatorHandler) MergeFeedback(w http.ResponseWriter, r *http.Request) {
	feedbackID, err := uuid.Parse(chi.URLParam(r, "feedbackId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_FEEDBACK_ID", "Invalid feedback ID")
		return
	}

	userID := auth.MustUserIDFromContext(r.Context())

	var req struct {
		CanonicalID uuid.UUID `json:"canonical_id"`
	}

	if err := DecodeJSON(r, &req); err != nil {
		HandleError(w, err)
		return
	}

	if req.CanonicalID == uuid.Nil {
		ValidationError(w, map[string]string{"canonical_id": "Canonical ID is required"})
		return
	}

	feedback, err := h.feedbackSvc.Merge(r.Context(), feedbackID, req.CanonicalID, userID)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, feedback)
}

// ListTags handles GET /creator/projects/:projectId/tags
func (h *CreatorHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "Invalid project ID")
		return
	}

	tags, err := h.tagSvc.ListByProject(r.Context(), projectID)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, tags)
}

// CreateTag handles POST /creator/projects/:projectId/tags
func (h *CreatorHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "Invalid project ID")
		return
	}

	var req struct {
		Name  string  `json:"name"`
		Color *string `json:"color"`
	}

	if err := DecodeJSON(r, &req); err != nil {
		HandleError(w, err)
		return
	}

	if req.Name == "" {
		ValidationError(w, map[string]string{"name": "Name is required"})
		return
	}

	tag, err := h.tagSvc.Create(r.Context(), projectID, req.Name, req.Color)
	if err != nil {
		HandleError(w, err)
		return
	}

	Created(w, tag)
}

// UpdateTag handles PATCH /creator/projects/:projectId/tags/:tagId
func (h *CreatorHandler) UpdateTag(w http.ResponseWriter, r *http.Request) {
	tagID, err := uuid.Parse(chi.URLParam(r, "tagId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_TAG_ID", "Invalid tag ID")
		return
	}

	var req struct {
		Name  *string `json:"name"`
		Color *string `json:"color"`
	}

	if err := DecodeJSON(r, &req); err != nil {
		HandleError(w, err)
		return
	}

	tag, err := h.tagSvc.Update(r.Context(), tagID, req.Name, req.Color)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, tag)
}

// DeleteTag handles DELETE /creator/projects/:projectId/tags/:tagId
func (h *CreatorHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	tagID, err := uuid.Parse(chi.URLParam(r, "tagId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_TAG_ID", "Invalid tag ID")
		return
	}

	if err := h.tagSvc.Delete(r.Context(), tagID); err != nil {
		HandleError(w, err)
		return
	}

	NoContent(w)
}

// ListMembers handles GET /creator/projects/:projectId/members
func (h *CreatorHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "Invalid project ID")
		return
	}

	members, err := h.membershipSvc.ListByProject(r.Context(), projectID)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, members)
}

// UpdateMember handles PATCH /creator/projects/:projectId/members/:memberId
func (h *CreatorHandler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	memberID, err := uuid.Parse(chi.URLParam(r, "memberId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_MEMBER_ID", "Invalid member ID")
		return
	}

	userID := auth.MustUserIDFromContext(r.Context())

	var req struct {
		Role domain.Role `json:"role"`
	}

	if err := DecodeJSON(r, &req); err != nil {
		HandleError(w, err)
		return
	}

	if req.Role == "" {
		ValidationError(w, map[string]string{"role": "Role is required"})
		return
	}

	membership, err := h.membershipSvc.UpdateRole(r.Context(), memberID, req.Role, userID)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, membership)
}

// RemoveMember handles DELETE /creator/projects/:projectId/members/:memberId
func (h *CreatorHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	memberID, err := uuid.Parse(chi.URLParam(r, "memberId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_MEMBER_ID", "Invalid member ID")
		return
	}

	userID := auth.MustUserIDFromContext(r.Context())

	if err := h.membershipSvc.Remove(r.Context(), memberID, userID); err != nil {
		HandleError(w, err)
		return
	}

	NoContent(w)
}

// InviteMember handles POST /creator/projects/:projectId/members/invite
func (h *CreatorHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "Invalid project ID")
		return
	}

	userID := auth.MustUserIDFromContext(r.Context())

	var req struct {
		Email string      `json:"email"`
		Role  domain.Role `json:"role"`
	}

	if err := DecodeJSON(r, &req); err != nil {
		HandleError(w, err)
		return
	}

	errors := make(map[string]string)
	if req.Email == "" {
		errors["email"] = "Email is required"
	}
	if req.Role == "" {
		errors["role"] = "Role is required"
	}
	if len(errors) > 0 {
		ValidationError(w, errors)
		return
	}

	invite, err := h.inviteSvc.Create(r.Context(), service.InviteMemberRequest{
		ProjectID: projectID,
		InviterID: userID,
		Email:     req.Email,
		Role:      req.Role,
	})
	if err != nil {
		HandleError(w, err)
		return
	}

	Created(w, invite)
}

// GetSettings handles GET /creator/projects/:projectId/settings
func (h *CreatorHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "Invalid project ID")
		return
	}

	project, err := h.projectSvc.GetByID(r.Context(), projectID)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, project)
}

// UpdateSettings handles PATCH /creator/projects/:projectId/settings
func (h *CreatorHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "Invalid project ID")
		return
	}

	userID := auth.MustUserIDFromContext(r.Context())

	var req struct {
		Name     *string                 `json:"name"`
		Settings *domain.ProjectSettings `json:"settings"`
	}

	if err := DecodeJSON(r, &req); err != nil {
		HandleError(w, err)
		return
	}

	project, err := h.projectSvc.Update(r.Context(), projectID, req.Name, req.Settings, userID)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, project)
}

// Helper function
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
