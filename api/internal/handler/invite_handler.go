package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/fulldisclosure/api/internal/auth"
	"github.com/fulldisclosure/api/internal/service"
)

// InviteHandler handles invite endpoints
type InviteHandler struct {
	inviteSvc service.InviteService
}

// NewInviteHandler creates a new invite handler
func NewInviteHandler(inviteSvc service.InviteService) *InviteHandler {
	return &InviteHandler{
		inviteSvc: inviteSvc,
	}
}

// AcceptInvite handles POST /invites/:token/accept
func (h *InviteHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		Error(w, http.StatusBadRequest, "INVALID_TOKEN", "Token is required")
		return
	}

	userID := auth.MustUserIDFromContext(r.Context())

	membership, err := h.inviteSvc.Accept(r.Context(), token, userID)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, membership)
}
