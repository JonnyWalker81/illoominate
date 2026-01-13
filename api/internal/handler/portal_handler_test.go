package handler

import (
	"bytes"
	"context"
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

// setupTestContext creates a chi router context with URL parameters
func setupTestContext(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for key, value := range params {
		rctx.URLParams.Add(key, value)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// withAuthContext adds user ID and email to request context
func withAuthContext(r *http.Request, userID uuid.UUID, email string) *http.Request {
	ctx := auth.ContextWithUserID(r.Context(), userID)
	ctx = auth.ContextWithUserEmail(ctx, email)
	return r.WithContext(ctx)
}

func TestPortalHandlers_GetProfile(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("success - returns profile", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		h := NewPortalHandlers(mockRepo, logger)

		userID := uuid.New()
		projectID := uuid.New()
		profile := &domain.PortalUserProfile{
			ID:        uuid.New(),
			UserID:    userID,
			ProjectID: projectID,
			NotificationPreferences: domain.PortalNotificationPreferences{
				StatusChanges:           true,
				NewCommentsOnMyFeedback: true,
			},
		}

		mockRepo.On("GetProfile", mock.Anything, userID, projectID).Return(profile, nil)

		req := httptest.NewRequest("GET", "/portal/"+projectID.String()+"/me", nil)
		req = setupTestContext(req, map[string]string{"projectId": projectID.String()})
		req = withAuthContext(req, userID, "test@example.com")

		rr := httptest.NewRecorder()
		h.GetProfile(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockRepo.AssertExpectations(t)

		var response domain.PortalUserProfile
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, profile.ID, response.ID)
	})

	t.Run("unauthorized - no user ID in context", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		h := NewPortalHandlers(mockRepo, logger)

		projectID := uuid.New()
		req := httptest.NewRequest("GET", "/portal/"+projectID.String()+"/me", nil)
		req = setupTestContext(req, map[string]string{"projectId": projectID.String()})
		// No auth context

		rr := httptest.NewRecorder()
		h.GetProfile(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("bad request - invalid project ID", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		h := NewPortalHandlers(mockRepo, logger)

		userID := uuid.New()
		req := httptest.NewRequest("GET", "/portal/invalid/me", nil)
		req = setupTestContext(req, map[string]string{"projectId": "invalid"})
		req = withAuthContext(req, userID, "test@example.com")

		rr := httptest.NewRecorder()
		h.GetProfile(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("not found - profile doesn't exist", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		h := NewPortalHandlers(mockRepo, logger)

		userID := uuid.New()
		projectID := uuid.New()

		mockRepo.On("GetProfile", mock.Anything, userID, projectID).Return(nil, domain.ErrNotFound)

		req := httptest.NewRequest("GET", "/portal/"+projectID.String()+"/me", nil)
		req = setupTestContext(req, map[string]string{"projectId": projectID.String()})
		req = withAuthContext(req, userID, "test@example.com")

		rr := httptest.NewRecorder()
		h.GetProfile(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestPortalHandlers_UpdateNotificationPreferences(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("success - updates preferences", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		h := NewPortalHandlers(mockRepo, logger)

		userID := uuid.New()
		projectID := uuid.New()

		prefs := domain.PortalNotificationPreferences{
			StatusChanges:           false,
			NewCommentsOnMyFeedback: true,
		}

		updatedProfile := &domain.PortalUserProfile{
			ID:                      uuid.New(),
			UserID:                  userID,
			ProjectID:               projectID,
			NotificationPreferences: prefs,
		}

		mockRepo.On("UpdateNotificationPrefs", mock.Anything, userID, projectID, prefs).Return(nil)
		mockRepo.On("GetProfile", mock.Anything, userID, projectID).Return(updatedProfile, nil)

		body, _ := json.Marshal(prefs)
		req := httptest.NewRequest("PATCH", "/portal/"+projectID.String()+"/me/notifications", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = setupTestContext(req, map[string]string{"projectId": projectID.String()})
		req = withAuthContext(req, userID, "test@example.com")

		rr := httptest.NewRecorder()
		h.UpdateNotificationPreferences(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized - no user ID", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		h := NewPortalHandlers(mockRepo, logger)

		projectID := uuid.New()
		body, _ := json.Marshal(domain.PortalNotificationPreferences{})
		req := httptest.NewRequest("PATCH", "/portal/"+projectID.String()+"/me/notifications", bytes.NewReader(body))
		req = setupTestContext(req, map[string]string{"projectId": projectID.String()})

		rr := httptest.NewRecorder()
		h.UpdateNotificationPreferences(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestPortalHandlers_ListMyFeedback(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("success - returns linked feedback", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		h := NewPortalHandlers(mockRepo, logger)

		userID := uuid.New()
		projectID := uuid.New()

		feedback := []domain.PortalFeedbackSummary{
			{
				ID:           uuid.New(),
				Title:        "Test Feedback",
				Description:  "Test description",
				Type:         "feature",
				Status:       "new",
				VoteCount:    5,
				CommentCount: 2,
				HasVoted:     false,
			},
		}

		mockRepo.On("GetLinkedFeedback", mock.Anything, userID, projectID).Return(feedback, nil)

		req := httptest.NewRequest("GET", "/portal/"+projectID.String()+"/my-feedback", nil)
		req = setupTestContext(req, map[string]string{"projectId": projectID.String()})
		req = withAuthContext(req, userID, "test@example.com")

		rr := httptest.NewRecorder()
		h.ListMyFeedback(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockRepo.AssertExpectations(t)

		var response []domain.PortalFeedbackSummary
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 1)
		assert.Equal(t, feedback[0].Title, response[0].Title)
	})

	t.Run("success - returns empty array when no feedback", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		h := NewPortalHandlers(mockRepo, logger)

		userID := uuid.New()
		projectID := uuid.New()

		mockRepo.On("GetLinkedFeedback", mock.Anything, userID, projectID).Return(nil, nil)

		req := httptest.NewRequest("GET", "/portal/"+projectID.String()+"/my-feedback", nil)
		req = setupTestContext(req, map[string]string{"projectId": projectID.String()})
		req = withAuthContext(req, userID, "test@example.com")

		rr := httptest.NewRecorder()
		h.ListMyFeedback(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockRepo.AssertExpectations(t)

		var response []domain.PortalFeedbackSummary
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Empty(t, response)
	})
}

func TestPortalHandlers_Vote(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("success - creates vote", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		h := NewPortalHandlers(mockRepo, logger)

		userID := uuid.New()
		projectID := uuid.New()
		feedbackID := uuid.New()

		mockRepo.On("CreateVote", mock.Anything, feedbackID, userID, projectID).Return(nil)

		req := httptest.NewRequest("POST", "/portal/"+projectID.String()+"/feature-requests/"+feedbackID.String()+"/vote", nil)
		req = setupTestContext(req, map[string]string{
			"projectId":  projectID.String(),
			"feedbackId": feedbackID.String(),
		})
		req = withAuthContext(req, userID, "test@example.com")

		rr := httptest.NewRecorder()
		h.Vote(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized - no user ID", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		h := NewPortalHandlers(mockRepo, logger)

		projectID := uuid.New()
		feedbackID := uuid.New()

		req := httptest.NewRequest("POST", "/portal/"+projectID.String()+"/feature-requests/"+feedbackID.String()+"/vote", nil)
		req = setupTestContext(req, map[string]string{
			"projectId":  projectID.String(),
			"feedbackId": feedbackID.String(),
		})

		rr := httptest.NewRecorder()
		h.Vote(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("bad request - invalid feedback ID", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		h := NewPortalHandlers(mockRepo, logger)

		userID := uuid.New()
		projectID := uuid.New()

		req := httptest.NewRequest("POST", "/portal/"+projectID.String()+"/feature-requests/invalid/vote", nil)
		req = setupTestContext(req, map[string]string{
			"projectId":  projectID.String(),
			"feedbackId": "invalid",
		})
		req = withAuthContext(req, userID, "test@example.com")

		rr := httptest.NewRecorder()
		h.Vote(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("not found - feedback doesn't exist", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		h := NewPortalHandlers(mockRepo, logger)

		userID := uuid.New()
		projectID := uuid.New()
		feedbackID := uuid.New()

		mockRepo.On("CreateVote", mock.Anything, feedbackID, userID, projectID).Return(domain.ErrNotFound)

		req := httptest.NewRequest("POST", "/portal/"+projectID.String()+"/feature-requests/"+feedbackID.String()+"/vote", nil)
		req = setupTestContext(req, map[string]string{
			"projectId":  projectID.String(),
			"feedbackId": feedbackID.String(),
		})
		req = withAuthContext(req, userID, "test@example.com")

		rr := httptest.NewRecorder()
		h.Vote(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestPortalHandlers_Unvote(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("success - removes vote", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		h := NewPortalHandlers(mockRepo, logger)

		userID := uuid.New()
		feedbackID := uuid.New()

		mockRepo.On("DeleteVote", mock.Anything, feedbackID, userID).Return(nil)

		req := httptest.NewRequest("DELETE", "/portal/proj/feature-requests/"+feedbackID.String()+"/vote", nil)
		req = setupTestContext(req, map[string]string{
			"projectId":  uuid.New().String(),
			"feedbackId": feedbackID.String(),
		})
		req = withAuthContext(req, userID, "test@example.com")

		rr := httptest.NewRecorder()
		h.Unvote(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found - vote doesn't exist", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		h := NewPortalHandlers(mockRepo, logger)

		userID := uuid.New()
		feedbackID := uuid.New()

		mockRepo.On("DeleteVote", mock.Anything, feedbackID, userID).Return(domain.ErrNotFound)

		req := httptest.NewRequest("DELETE", "/portal/proj/feature-requests/"+feedbackID.String()+"/vote", nil)
		req = setupTestContext(req, map[string]string{
			"projectId":  uuid.New().String(),
			"feedbackId": feedbackID.String(),
		})
		req = withAuthContext(req, userID, "test@example.com")

		rr := httptest.NewRecorder()
		h.Unvote(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestPortalHandlers_ListFeatures(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("success - returns public features with pagination", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		h := NewPortalHandlers(mockRepo, logger)

		projectID := uuid.New()

		features := []domain.PortalFeedbackSummary{
			{
				ID:           uuid.New(),
				Title:        "Feature 1",
				Description:  "Description 1",
				Type:         "feature",
				Status:       "planned",
				VoteCount:    10,
				CommentCount: 5,
				HasVoted:     false,
			},
			{
				ID:           uuid.New(),
				Title:        "Feature 2",
				Description:  "Description 2",
				Type:         "feature",
				Status:       "new",
				VoteCount:    5,
				CommentCount: 2,
				HasVoted:     false,
			},
		}

		// Note: userID is nil for unauthenticated requests
		mockRepo.On("ListPublicFeatures", mock.Anything, projectID, (*uuid.UUID)(nil), 20, 0).Return(features, 2, nil)

		req := httptest.NewRequest("GET", "/portal/"+projectID.String()+"/feature-requests", nil)
		req = setupTestContext(req, map[string]string{"projectId": projectID.String()})
		// No auth context - public endpoint

		rr := httptest.NewRecorder()
		h.ListFeatures(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("success - returns features with has_voted for authenticated user", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		h := NewPortalHandlers(mockRepo, logger)

		userID := uuid.New()
		projectID := uuid.New()

		features := []domain.PortalFeedbackSummary{
			{
				ID:       uuid.New(),
				Title:    "Feature 1",
				HasVoted: true,
			},
		}

		mockRepo.On("ListPublicFeatures", mock.Anything, projectID, &userID, 20, 0).Return(features, 1, nil)

		req := httptest.NewRequest("GET", "/portal/"+projectID.String()+"/feature-requests", nil)
		req = setupTestContext(req, map[string]string{"projectId": projectID.String()})
		req = withAuthContext(req, userID, "test@example.com")

		rr := httptest.NewRecorder()
		h.ListFeatures(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("success - respects pagination params", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		h := NewPortalHandlers(mockRepo, logger)

		projectID := uuid.New()

		mockRepo.On("ListPublicFeatures", mock.Anything, projectID, (*uuid.UUID)(nil), 10, 20).Return([]domain.PortalFeedbackSummary{}, 0, nil)

		req := httptest.NewRequest("GET", "/portal/"+projectID.String()+"/feature-requests?limit=10&offset=20", nil)
		req = setupTestContext(req, map[string]string{"projectId": projectID.String()})

		rr := httptest.NewRecorder()
		h.ListFeatures(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestPortalAccessMiddleware(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("creates profile and links SDK users", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()

		userID := uuid.New()
		projectID := uuid.New()
		email := "test@example.com"

		mockRepo.On("CreateProfile", mock.Anything, userID, projectID).Return(nil)
		mockRepo.On("LinkSDKUsersByEmail", mock.Anything, userID, projectID, email).Return(int64(1), nil)

		middleware := PortalAccessMiddleware(mockRepo, logger)

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/portal/"+projectID.String()+"/me", nil)
		req = setupTestContext(req, map[string]string{"projectId": projectID.String()})
		req = withAuthContext(req, userID, email)

		rr := httptest.NewRecorder()
		middleware(next).ServeHTTP(rr, req)

		assert.True(t, nextCalled)
		assert.Equal(t, http.StatusOK, rr.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized - no user ID in context", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		middleware := PortalAccessMiddleware(mockRepo, logger)

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		projectID := uuid.New()
		req := httptest.NewRequest("GET", "/portal/"+projectID.String()+"/me", nil)
		req = setupTestContext(req, map[string]string{"projectId": projectID.String()})
		// No auth context

		rr := httptest.NewRecorder()
		middleware(next).ServeHTTP(rr, req)

		assert.False(t, nextCalled)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("bad request - invalid project ID", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()
		middleware := PortalAccessMiddleware(mockRepo, logger)

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		userID := uuid.New()
		req := httptest.NewRequest("GET", "/portal/invalid/me", nil)
		req = setupTestContext(req, map[string]string{"projectId": "invalid"})
		req = withAuthContext(req, userID, "test@example.com")

		rr := httptest.NewRecorder()
		middleware(next).ServeHTTP(rr, req)

		assert.False(t, nextCalled)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("continues even if profile creation fails", func(t *testing.T) {
		mockRepo := repository.NewMockPortalRepository()

		userID := uuid.New()
		projectID := uuid.New()
		email := "test@example.com"

		// Profile creation fails
		mockRepo.On("CreateProfile", mock.Anything, userID, projectID).Return(assert.AnError)
		mockRepo.On("LinkSDKUsersByEmail", mock.Anything, userID, projectID, email).Return(int64(0), nil)

		middleware := PortalAccessMiddleware(mockRepo, logger)

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/portal/"+projectID.String()+"/me", nil)
		req = setupTestContext(req, map[string]string{"projectId": projectID.String()})
		req = withAuthContext(req, userID, email)

		rr := httptest.NewRecorder()
		middleware(next).ServeHTTP(rr, req)

		assert.True(t, nextCalled)
		mockRepo.AssertExpectations(t)
	})
}
