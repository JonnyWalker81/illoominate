package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/fulldisclosure/api/internal/auth"
	"github.com/fulldisclosure/api/internal/domain"
	"github.com/fulldisclosure/api/internal/repository"
)

// PortalHandlers contains all portal-related HTTP handlers
type PortalHandlers struct {
	repo   repository.PortalRepository
	logger zerolog.Logger
}

// NewPortalHandlers creates a new PortalHandlers instance
func NewPortalHandlers(repo repository.PortalRepository, logger zerolog.Logger) *PortalHandlers {
	return &PortalHandlers{
		repo:   repo,
		logger: logger,
	}
}

// GetProfile returns the user's portal profile for a project
func (h *PortalHandlers) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "Invalid project ID")
		return
	}

	profile, err := h.repo.GetProfile(r.Context(), userID, projectID)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, profile)
}

// UpdateNotificationPreferences updates the user's notification preferences
func (h *PortalHandlers) UpdateNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "Invalid project ID")
		return
	}

	var prefs domain.PortalNotificationPreferences
	if err := DecodeJSON(r, &prefs); err != nil {
		HandleError(w, err)
		return
	}

	if err := h.repo.UpdateNotificationPrefs(r.Context(), userID, projectID, prefs); err != nil {
		HandleError(w, err)
		return
	}

	// Return the updated profile
	profile, err := h.repo.GetProfile(r.Context(), userID, projectID)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, profile)
}

// ListMyFeedback returns all feedback linked to the current user via SDK
func (h *PortalHandlers) ListMyFeedback(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "Invalid project ID")
		return
	}

	feedback, err := h.repo.GetLinkedFeedback(r.Context(), userID, projectID)
	if err != nil {
		HandleError(w, err)
		return
	}

	if feedback == nil {
		feedback = []domain.PortalFeedbackSummary{}
	}

	JSON(w, http.StatusOK, feedback)
}

// ListFeatures returns public feature requests for the project
func (h *PortalHandlers) ListFeatures(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "Invalid project ID")
		return
	}

	// Get optional user ID for has_voted tracking
	var userID *uuid.UUID
	if uid, ok := auth.UserIDFromContext(r.Context()); ok {
		userID = &uid
	}

	// Parse pagination
	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	feedback, total, err := h.repo.ListPublicFeatures(r.Context(), projectID, userID, limit, offset)
	if err != nil {
		HandleError(w, err)
		return
	}

	if feedback == nil {
		feedback = []domain.PortalFeedbackSummary{}
	}

	page := (offset / limit) + 1
	totalPages := (total + limit - 1) / limit

	Paginated(w, feedback, total, page, limit, totalPages)
}

// Vote adds a vote to a feature request
func (h *PortalHandlers) Vote(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

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

	if err := h.repo.CreateVote(r.Context(), feedbackID, userID, projectID); err != nil {
		HandleError(w, err)
		return
	}

	NoContent(w)
}

// Unvote removes a vote from a feature request
func (h *PortalHandlers) Unvote(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	feedbackID, err := uuid.Parse(chi.URLParam(r, "feedbackId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_FEEDBACK_ID", "Invalid feedback ID")
		return
	}

	if err := h.repo.DeleteVote(r.Context(), feedbackID, userID); err != nil {
		HandleError(w, err)
		return
	}

	NoContent(w)
}

// PortalAccessMiddleware creates profiles and links SDK users on first visit
func PortalAccessMiddleware(repo repository.PortalRepository, logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := auth.UserIDFromContext(r.Context())
			if !ok {
				Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}

			projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
			if err != nil {
				Error(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "Invalid project ID")
				return
			}

			email, _ := auth.UserEmailFromContext(r.Context())

			// Create profile if not exists (idempotent)
			if err := repo.CreateProfile(r.Context(), userID, projectID); err != nil {
				logger.Error().Err(err).
					Str("user_id", userID.String()).
					Str("project_id", projectID.String()).
					Msg("Failed to create portal profile")
				// Continue anyway - non-critical error
			}

			// Auto-link SDK users with matching email
			if email != "" {
				linkedCount, err := repo.LinkSDKUsersByEmail(r.Context(), userID, projectID, email)
				if err != nil {
					logger.Error().Err(err).
						Str("user_id", userID.String()).
						Str("project_id", projectID.String()).
						Str("email", email).
						Msg("Failed to link SDK users")
					// Continue anyway - non-critical error
				} else if linkedCount > 0 {
					logger.Info().
						Str("user_id", userID.String()).
						Str("project_id", projectID.String()).
						Str("email", email).
						Int64("linked_count", linkedCount).
						Msg("Linked SDK users to portal user")
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
