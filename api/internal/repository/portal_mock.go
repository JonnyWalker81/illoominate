package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/fulldisclosure/api/internal/domain"
)

// MockPortalRepository is a mock implementation of PortalRepository for testing
type MockPortalRepository struct {
	mock.Mock
}

// NewMockPortalRepository creates a new mock portal repository
func NewMockPortalRepository() *MockPortalRepository {
	return &MockPortalRepository{}
}

func (m *MockPortalRepository) CreateProfile(ctx context.Context, userID, projectID uuid.UUID) error {
	args := m.Called(ctx, userID, projectID)
	return args.Error(0)
}

func (m *MockPortalRepository) GetProfile(ctx context.Context, userID, projectID uuid.UUID) (*domain.PortalUserProfile, error) {
	args := m.Called(ctx, userID, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PortalUserProfile), args.Error(1)
}

func (m *MockPortalRepository) UpdateNotificationPrefs(ctx context.Context, userID, projectID uuid.UUID, prefs domain.PortalNotificationPreferences) error {
	args := m.Called(ctx, userID, projectID, prefs)
	return args.Error(0)
}

func (m *MockPortalRepository) LinkSDKUsersByEmail(ctx context.Context, userID, projectID uuid.UUID, email string) (int64, error) {
	args := m.Called(ctx, userID, projectID, email)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPortalRepository) GetLinkedFeedback(ctx context.Context, userID, projectID uuid.UUID) ([]domain.PortalFeedbackSummary, error) {
	args := m.Called(ctx, userID, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PortalFeedbackSummary), args.Error(1)
}

func (m *MockPortalRepository) ListPublicFeatures(ctx context.Context, projectID uuid.UUID, userID *uuid.UUID, limit, offset int) ([]domain.PortalFeedbackSummary, int, error) {
	args := m.Called(ctx, projectID, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]domain.PortalFeedbackSummary), args.Int(1), args.Error(2)
}

func (m *MockPortalRepository) CreateVote(ctx context.Context, feedbackID, userID, projectID uuid.UUID) error {
	args := m.Called(ctx, feedbackID, userID, projectID)
	return args.Error(0)
}

func (m *MockPortalRepository) DeleteVote(ctx context.Context, feedbackID, userID uuid.UUID) error {
	args := m.Called(ctx, feedbackID, userID)
	return args.Error(0)
}

func (m *MockPortalRepository) HasVoted(ctx context.Context, feedbackID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, feedbackID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPortalRepository) GetVotedFeedbackIDs(ctx context.Context, userID uuid.UUID, feedbackIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	args := m.Called(ctx, userID, feedbackIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]bool), args.Error(1)
}

// Ensure MockPortalRepository implements PortalRepository
var _ PortalRepository = (*MockPortalRepository)(nil)
