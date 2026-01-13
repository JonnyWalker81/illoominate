package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/fulldisclosure/api/internal/domain"
)

func TestMockPortalRepository_CreateProfile(t *testing.T) {
	mockRepo := NewMockPortalRepository()
	userID := uuid.New()
	projectID := uuid.New()

	mockRepo.On("CreateProfile", mock.Anything, userID, projectID).Return(nil)

	err := mockRepo.CreateProfile(context.Background(), userID, projectID)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMockPortalRepository_GetProfile(t *testing.T) {
	mockRepo := NewMockPortalRepository()
	userID := uuid.New()
	projectID := uuid.New()

	expectedProfile := &domain.PortalUserProfile{
		ID:        uuid.New(),
		UserID:    userID,
		ProjectID: projectID,
		NotificationPreferences: domain.PortalNotificationPreferences{
			StatusChanges:           true,
			NewCommentsOnMyFeedback: true,
		},
	}

	mockRepo.On("GetProfile", mock.Anything, userID, projectID).Return(expectedProfile, nil)

	profile, err := mockRepo.GetProfile(context.Background(), userID, projectID)
	assert.NoError(t, err)
	assert.Equal(t, expectedProfile.ID, profile.ID)
	assert.Equal(t, expectedProfile.UserID, profile.UserID)
	mockRepo.AssertExpectations(t)
}

func TestMockPortalRepository_UpdateNotificationPrefs(t *testing.T) {
	mockRepo := NewMockPortalRepository()
	userID := uuid.New()
	projectID := uuid.New()

	prefs := domain.PortalNotificationPreferences{
		StatusChanges:           false,
		NewCommentsOnMyFeedback: true,
	}

	mockRepo.On("UpdateNotificationPrefs", mock.Anything, userID, projectID, prefs).Return(nil)

	err := mockRepo.UpdateNotificationPrefs(context.Background(), userID, projectID, prefs)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMockPortalRepository_LinkSDKUsersByEmail(t *testing.T) {
	mockRepo := NewMockPortalRepository()
	userID := uuid.New()
	projectID := uuid.New()
	email := "test@example.com"

	// Simulates linking 2 SDK users with matching email
	mockRepo.On("LinkSDKUsersByEmail", mock.Anything, userID, projectID, email).Return(int64(2), nil)

	count, err := mockRepo.LinkSDKUsersByEmail(context.Background(), userID, projectID, email)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
	mockRepo.AssertExpectations(t)
}

func TestMockPortalRepository_GetLinkedFeedback(t *testing.T) {
	mockRepo := NewMockPortalRepository()
	userID := uuid.New()
	projectID := uuid.New()

	expectedFeedback := []domain.PortalFeedbackSummary{
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

	mockRepo.On("GetLinkedFeedback", mock.Anything, userID, projectID).Return(expectedFeedback, nil)

	feedback, err := mockRepo.GetLinkedFeedback(context.Background(), userID, projectID)
	assert.NoError(t, err)
	assert.Len(t, feedback, 1)
	assert.Equal(t, expectedFeedback[0].Title, feedback[0].Title)
	mockRepo.AssertExpectations(t)
}

func TestMockPortalRepository_ListPublicFeatures(t *testing.T) {
	mockRepo := NewMockPortalRepository()
	projectID := uuid.New()
	userID := uuid.New()

	expectedFeatures := []domain.PortalFeedbackSummary{
		{
			ID:       uuid.New(),
			Title:    "Feature 1",
			Type:     "feature",
			Status:   "planned",
			HasVoted: true,
		},
	}

	mockRepo.On("ListPublicFeatures", mock.Anything, projectID, &userID, 20, 0).Return(expectedFeatures, 1, nil)

	features, total, err := mockRepo.ListPublicFeatures(context.Background(), projectID, &userID, 20, 0)
	assert.NoError(t, err)
	assert.Len(t, features, 1)
	assert.Equal(t, 1, total)
	assert.True(t, features[0].HasVoted)
	mockRepo.AssertExpectations(t)
}

func TestMockPortalRepository_CreateVote(t *testing.T) {
	mockRepo := NewMockPortalRepository()
	feedbackID := uuid.New()
	userID := uuid.New()
	projectID := uuid.New()

	mockRepo.On("CreateVote", mock.Anything, feedbackID, userID, projectID).Return(nil)

	err := mockRepo.CreateVote(context.Background(), feedbackID, userID, projectID)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMockPortalRepository_DeleteVote(t *testing.T) {
	mockRepo := NewMockPortalRepository()
	feedbackID := uuid.New()
	userID := uuid.New()

	mockRepo.On("DeleteVote", mock.Anything, feedbackID, userID).Return(nil)

	err := mockRepo.DeleteVote(context.Background(), feedbackID, userID)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMockPortalRepository_HasVoted(t *testing.T) {
	mockRepo := NewMockPortalRepository()
	feedbackID := uuid.New()
	userID := uuid.New()

	mockRepo.On("HasVoted", mock.Anything, feedbackID, userID).Return(true, nil)

	hasVoted, err := mockRepo.HasVoted(context.Background(), feedbackID, userID)
	assert.NoError(t, err)
	assert.True(t, hasVoted)
	mockRepo.AssertExpectations(t)
}

func TestMockPortalRepository_GetVotedFeedbackIDs(t *testing.T) {
	mockRepo := NewMockPortalRepository()
	userID := uuid.New()
	feedbackIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	expectedResult := map[uuid.UUID]bool{
		feedbackIDs[0]: true,
		feedbackIDs[2]: true,
	}

	mockRepo.On("GetVotedFeedbackIDs", mock.Anything, userID, feedbackIDs).Return(expectedResult, nil)

	result, err := mockRepo.GetVotedFeedbackIDs(context.Background(), userID, feedbackIDs)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.True(t, result[feedbackIDs[0]])
	assert.False(t, result[feedbackIDs[1]])
	assert.True(t, result[feedbackIDs[2]])
	mockRepo.AssertExpectations(t)
}
