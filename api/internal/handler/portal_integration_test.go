package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/fulldisclosure/api/internal/auth"
	"github.com/fulldisclosure/api/internal/domain"
	"github.com/fulldisclosure/api/internal/repository"
)

// Integration tests for the portal email linking flow
// These tests simulate the full request/response flow through middleware and handlers

func TestPortalEmailLinkingFlow(t *testing.T) {
	// This test simulates: SDK identify → portal login → verify linked feedback appears

	t.Run("complete flow: SDK user gets linked when portal user accesses endpoint", func(t *testing.T) {
		// Setup
		mockRepo := repository.NewMockPortalRepository()
		logger := zerolog.Nop()
		handlers := NewPortalHandlers(mockRepo, logger)

		// Test data
		portalUserID := uuid.New()
		projectID := uuid.New()
		userEmail := "test@example.com"

		// Simulate existing SDK feedback (submitted before portal signup)
		sdkFeedback := []domain.PortalFeedbackSummary{
			{
				ID:           uuid.New(),
				Title:        "Feature Request from SDK",
				Description:  "Submitted via mobile app",
				Type:         "feature",
				Status:       "new",
				VoteCount:    3,
				CommentCount: 1,
				HasVoted:     false,
			},
			{
				ID:           uuid.New(),
				Title:        "Bug Report from SDK",
				Description:  "Found a crash",
				Type:         "bug",
				Status:       "in_progress",
				VoteCount:    5,
				CommentCount: 2,
				HasVoted:     false,
			},
		}

		// Step 1: Mock - Profile creation when middleware runs
		mockRepo.On("CreateProfile", mock.Anything, portalUserID, projectID).Return(nil).Once()

		// Step 2: Mock - SDK users linking by email
		mockRepo.On("LinkSDKUsersByEmail", mock.Anything, portalUserID, projectID, userEmail).Return(int64(2), nil).Once()

		// Step 3: Mock - Getting linked feedback
		mockRepo.On("GetLinkedFeedback", mock.Anything, portalUserID, projectID).Return(sdkFeedback, nil).Once()

		// Create router with middleware
		r := chi.NewRouter()
		r.Route("/portal/{projectId}", func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				// Simulate authenticated user context
				return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					ctx := auth.ContextWithUserID(req.Context(), portalUserID)
					ctx = auth.ContextWithUserEmail(ctx, userEmail)
					next.ServeHTTP(w, req.WithContext(ctx))
				})
			})
			r.Use(PortalAccessMiddleware(mockRepo, logger))
			r.Get("/my-feedback", handlers.ListMyFeedback)
		})

		// Make request to /my-feedback (this triggers middleware + handler)
		req := httptest.NewRequest("GET", "/portal/"+projectID.String()+"/my-feedback", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		// Assertions
		assert.Equal(t, http.StatusOK, rec.Code)

		var response []domain.PortalFeedbackSummary
		err := json.NewDecoder(rec.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Len(t, response, 2)
		assert.Equal(t, "Feature Request from SDK", response[0].Title)
		assert.Equal(t, "Bug Report from SDK", response[1].Title)

		mockRepo.AssertExpectations(t)
	})

	t.Run("middleware creates profile on first access even if no SDK users to link", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		logger := zerolog.Nop()
		handlers := NewPortalHandlers(mockRepo, logger)

		portalUserID := uuid.New()
		projectID := uuid.New()
		userEmail := "newuser@example.com"

		// Mock - Profile creation succeeds
		mockRepo.On("CreateProfile", mock.Anything, portalUserID, projectID).Return(nil).Once()

		// Mock - No SDK users found with this email
		mockRepo.On("LinkSDKUsersByEmail", mock.Anything, portalUserID, projectID, userEmail).Return(int64(0), nil).Once()

		// Mock - No linked feedback
		mockRepo.On("GetLinkedFeedback", mock.Anything, portalUserID, projectID).Return([]domain.PortalFeedbackSummary{}, nil).Once()

		r := chi.NewRouter()
		r.Route("/portal/{projectId}", func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					ctx := auth.ContextWithUserID(req.Context(), portalUserID)
					ctx = auth.ContextWithUserEmail(ctx, userEmail)
					next.ServeHTTP(w, req.WithContext(ctx))
				})
			})
			r.Use(PortalAccessMiddleware(mockRepo, logger))
			r.Get("/my-feedback", handlers.ListMyFeedback)
		})

		req := httptest.NewRequest("GET", "/portal/"+projectID.String()+"/my-feedback", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response []domain.PortalFeedbackSummary
		err := json.NewDecoder(rec.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Empty(t, response)

		mockRepo.AssertExpectations(t)
	})
}

func TestPortalVotingFlow(t *testing.T) {
	// Tests the voting flow: vote → refresh → vote still present

	t.Run("vote persists after page refresh", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		logger := zerolog.Nop()
		handlers := NewPortalHandlers(mockRepo, logger)

		portalUserID := uuid.New()
		projectID := uuid.New()
		feedbackID := uuid.New()
		userEmail := "voter@example.com"

		// Create router
		r := chi.NewRouter()
		r.Route("/portal/{projectId}", func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					ctx := auth.ContextWithUserID(req.Context(), portalUserID)
					ctx = auth.ContextWithUserEmail(ctx, userEmail)
					next.ServeHTTP(w, req.WithContext(ctx))
				})
			})
			r.Use(PortalAccessMiddleware(mockRepo, logger))
			r.Post("/feature-requests/{feedbackId}/vote", handlers.Vote)
			r.Get("/feature-requests", handlers.ListFeatures)
		})

		// Step 1: User votes on a feature
		mockRepo.On("CreateProfile", mock.Anything, portalUserID, projectID).Return(nil).Once()
		mockRepo.On("LinkSDKUsersByEmail", mock.Anything, portalUserID, projectID, userEmail).Return(int64(0), nil).Once()
		mockRepo.On("CreateVote", mock.Anything, feedbackID, portalUserID, projectID).Return(nil).Once()

		voteReq := httptest.NewRequest("POST", "/portal/"+projectID.String()+"/feature-requests/"+feedbackID.String()+"/vote", nil)
		voteRec := httptest.NewRecorder()
		r.ServeHTTP(voteRec, voteReq)

		assert.Equal(t, http.StatusNoContent, voteRec.Code)

		// Step 2: User refreshes page (simulated by listing features)
		featuresAfterVote := []domain.PortalFeedbackSummary{
			{
				ID:        feedbackID,
				Title:     "Voted Feature",
				Type:      "feature",
				Status:    "planned",
				VoteCount: 6, // Vote count increased
				HasVoted:  true, // Shows user has voted
			},
		}

		mockRepo.On("CreateProfile", mock.Anything, portalUserID, projectID).Return(nil).Once()
		mockRepo.On("LinkSDKUsersByEmail", mock.Anything, portalUserID, projectID, userEmail).Return(int64(0), nil).Once()
		mockRepo.On("ListPublicFeatures", mock.Anything, projectID, &portalUserID, 20, 0).Return(featuresAfterVote, 1, nil).Once()

		listReq := httptest.NewRequest("GET", "/portal/"+projectID.String()+"/feature-requests", nil)
		listRec := httptest.NewRecorder()
		r.ServeHTTP(listRec, listReq)

		assert.Equal(t, http.StatusOK, listRec.Code)

		var response struct {
			Data  []domain.PortalFeedbackSummary `json:"data"`
			Total int                            `json:"total"`
		}
		err := json.NewDecoder(listRec.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Len(t, response.Data, 1)
		assert.True(t, response.Data[0].HasVoted, "Vote should persist after refresh")
		assert.Equal(t, 6, response.Data[0].VoteCount)

		mockRepo.AssertExpectations(t)
	})

	t.Run("unvote removes vote", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		logger := zerolog.Nop()
		handlers := NewPortalHandlers(mockRepo, logger)

		portalUserID := uuid.New()
		projectID := uuid.New()
		feedbackID := uuid.New()
		userEmail := "voter@example.com"

		r := chi.NewRouter()
		r.Route("/portal/{projectId}", func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					ctx := auth.ContextWithUserID(req.Context(), portalUserID)
					ctx = auth.ContextWithUserEmail(ctx, userEmail)
					next.ServeHTTP(w, req.WithContext(ctx))
				})
			})
			r.Use(PortalAccessMiddleware(mockRepo, logger))
			r.Delete("/feature-requests/{feedbackId}/vote", handlers.Unvote)
			r.Get("/feature-requests", handlers.ListFeatures)
		})

		// Step 1: User removes their vote
		mockRepo.On("CreateProfile", mock.Anything, portalUserID, projectID).Return(nil).Once()
		mockRepo.On("LinkSDKUsersByEmail", mock.Anything, portalUserID, projectID, userEmail).Return(int64(0), nil).Once()
		mockRepo.On("DeleteVote", mock.Anything, feedbackID, portalUserID).Return(nil).Once()

		unvoteReq := httptest.NewRequest("DELETE", "/portal/"+projectID.String()+"/feature-requests/"+feedbackID.String()+"/vote", nil)
		unvoteRec := httptest.NewRecorder()
		r.ServeHTTP(unvoteRec, unvoteReq)

		assert.Equal(t, http.StatusNoContent, unvoteRec.Code)

		// Step 2: Verify vote is gone on refresh
		featuresAfterUnvote := []domain.PortalFeedbackSummary{
			{
				ID:        feedbackID,
				Title:     "Unvoted Feature",
				Type:      "feature",
				Status:    "planned",
				VoteCount: 5, // Vote count decreased
				HasVoted:  false, // No longer voted
			},
		}

		mockRepo.On("CreateProfile", mock.Anything, portalUserID, projectID).Return(nil).Once()
		mockRepo.On("LinkSDKUsersByEmail", mock.Anything, portalUserID, projectID, userEmail).Return(int64(0), nil).Once()
		mockRepo.On("ListPublicFeatures", mock.Anything, projectID, &portalUserID, 20, 0).Return(featuresAfterUnvote, 1, nil).Once()

		listReq := httptest.NewRequest("GET", "/portal/"+projectID.String()+"/feature-requests", nil)
		listRec := httptest.NewRecorder()
		r.ServeHTTP(listRec, listReq)

		assert.Equal(t, http.StatusOK, listRec.Code)

		var response struct {
			Data  []domain.PortalFeedbackSummary `json:"data"`
			Total int                            `json:"total"`
		}
		err := json.NewDecoder(listRec.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Data[0].HasVoted, "Vote should be removed after unvote")
		assert.Equal(t, 5, response.Data[0].VoteCount)

		mockRepo.AssertExpectations(t)
	})
}

func TestPortalNotificationSettingsFlow(t *testing.T) {
	t.Run("user updates notification preferences", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		logger := zerolog.Nop()
		handlers := NewPortalHandlers(mockRepo, logger)

		portalUserID := uuid.New()
		projectID := uuid.New()
		userEmail := "user@example.com"

		r := chi.NewRouter()
		r.Route("/portal/{projectId}", func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					ctx := auth.ContextWithUserID(req.Context(), portalUserID)
					ctx = auth.ContextWithUserEmail(ctx, userEmail)
					next.ServeHTTP(w, req.WithContext(ctx))
				})
			})
			r.Use(PortalAccessMiddleware(mockRepo, logger))
			r.Get("/me", handlers.GetProfile)
			r.Patch("/me/notifications", handlers.UpdateNotificationPreferences)
		})

		// Step 1: Get initial profile with default preferences
		initialProfile := &domain.PortalUserProfile{
			ID:        uuid.New(),
			UserID:    portalUserID,
			ProjectID: projectID,
			NotificationPreferences: domain.PortalNotificationPreferences{
				StatusChanges:           true, // Default: on
				NewCommentsOnMyFeedback: true, // Default: on
			},
		}

		mockRepo.On("CreateProfile", mock.Anything, portalUserID, projectID).Return(nil).Once()
		mockRepo.On("LinkSDKUsersByEmail", mock.Anything, portalUserID, projectID, userEmail).Return(int64(0), nil).Once()
		mockRepo.On("GetProfile", mock.Anything, portalUserID, projectID).Return(initialProfile, nil).Once()

		getReq := httptest.NewRequest("GET", "/portal/"+projectID.String()+"/me", nil)
		getRec := httptest.NewRecorder()
		r.ServeHTTP(getRec, getReq)

		assert.Equal(t, http.StatusOK, getRec.Code)

		var profile domain.PortalUserProfile
		json.NewDecoder(getRec.Body).Decode(&profile)
		assert.True(t, profile.NotificationPreferences.StatusChanges)
		assert.True(t, profile.NotificationPreferences.NewCommentsOnMyFeedback)

		// Step 2: User disables status change notifications
		newPrefs := domain.PortalNotificationPreferences{
			StatusChanges:           false, // Changed
			NewCommentsOnMyFeedback: true,
		}

		updatedProfile := &domain.PortalUserProfile{
			ID:                      initialProfile.ID,
			UserID:                  portalUserID,
			ProjectID:               projectID,
			NotificationPreferences: newPrefs,
		}

		mockRepo.On("CreateProfile", mock.Anything, portalUserID, projectID).Return(nil).Once()
		mockRepo.On("LinkSDKUsersByEmail", mock.Anything, portalUserID, projectID, userEmail).Return(int64(0), nil).Once()
		mockRepo.On("UpdateNotificationPrefs", mock.Anything, portalUserID, projectID, newPrefs).Return(nil).Once()
		mockRepo.On("GetProfile", mock.Anything, portalUserID, projectID).Return(updatedProfile, nil).Once()

		updateBody, _ := json.Marshal(newPrefs)
		updateReq := httptest.NewRequest("PATCH", "/portal/"+projectID.String()+"/me/notifications",
			bytes.NewReader(updateBody))
		updateReq.Header.Set("Content-Type", "application/json")
		updateRec := httptest.NewRecorder()
		r.ServeHTTP(updateRec, updateReq)

		assert.Equal(t, http.StatusOK, updateRec.Code)

		// Verify the response contains updated preferences
		var responseProfile domain.PortalUserProfile
		json.NewDecoder(updateRec.Body).Decode(&responseProfile)
		assert.False(t, responseProfile.NotificationPreferences.StatusChanges)

		mockRepo.AssertExpectations(t)
	})
}
