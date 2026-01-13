package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/fulldisclosure/api/internal/domain"
	"github.com/fulldisclosure/api/internal/repository"
)

type voteService struct {
	voteRepo     repository.VoteRepository
	feedbackRepo repository.FeedbackRepository
}

// NewVoteService creates a new vote service
func NewVoteService(
	voteRepo repository.VoteRepository,
	feedbackRepo repository.FeedbackRepository,
) VoteService {
	return &voteService{
		voteRepo:     voteRepo,
		feedbackRepo: feedbackRepo,
	}
}

func (s *voteService) Vote(ctx context.Context, projectID, feedbackID, userID uuid.UUID) (*VoteResult, error) {
	// Verify feedback exists and belongs to project
	feedback, err := s.feedbackRepo.GetByID(ctx, feedbackID)
	if err != nil {
		return nil, err
	}

	if feedback.ProjectID != projectID {
		return nil, domain.ErrNotFound
	}

	// Check if already voted
	exists, err := s.voteRepo.Exists(ctx, feedbackID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check vote: %w", err)
	}

	if exists {
		// Already voted - return current state (idempotent)
		count, err := s.voteRepo.CountByFeedback(ctx, feedbackID)
		if err != nil {
			return nil, fmt.Errorf("failed to count votes: %w", err)
		}
		return &VoteResult{
			FeedbackID: feedbackID,
			VoteCount:  count,
			HasVoted:   true,
		}, nil
	}

	// Create vote
	vote := &domain.Vote{
		ID:         uuid.New(),
		FeedbackID: feedbackID,
		UserID:     userID,
	}

	if err := s.voteRepo.Create(ctx, vote); err != nil {
		return nil, fmt.Errorf("failed to create vote: %w", err)
	}

	// Get updated count (trigger updates denormalized count)
	count, err := s.voteRepo.CountByFeedback(ctx, feedbackID)
	if err != nil {
		return nil, fmt.Errorf("failed to count votes: %w", err)
	}

	return &VoteResult{
		FeedbackID: feedbackID,
		VoteCount:  count,
		HasVoted:   true,
	}, nil
}

func (s *voteService) Unvote(ctx context.Context, projectID, feedbackID, userID uuid.UUID) (*VoteResult, error) {
	// Verify feedback exists and belongs to project
	feedback, err := s.feedbackRepo.GetByID(ctx, feedbackID)
	if err != nil {
		return nil, err
	}

	if feedback.ProjectID != projectID {
		return nil, domain.ErrNotFound
	}

	// Delete vote (will return ErrNotFound if doesn't exist)
	if err := s.voteRepo.Delete(ctx, feedbackID, userID); err != nil {
		if err == domain.ErrNotFound {
			// Not voted - return current state (idempotent)
			count, countErr := s.voteRepo.CountByFeedback(ctx, feedbackID)
			if countErr != nil {
				return nil, fmt.Errorf("failed to count votes: %w", countErr)
			}
			return &VoteResult{
				FeedbackID: feedbackID,
				VoteCount:  count,
				HasVoted:   false,
			}, nil
		}
		return nil, fmt.Errorf("failed to delete vote: %w", err)
	}

	// Get updated count
	count, err := s.voteRepo.CountByFeedback(ctx, feedbackID)
	if err != nil {
		return nil, fmt.Errorf("failed to count votes: %w", err)
	}

	return &VoteResult{
		FeedbackID: feedbackID,
		VoteCount:  count,
		HasVoted:   false,
	}, nil
}

func (s *voteService) HasVoted(ctx context.Context, feedbackID, userID uuid.UUID) (bool, error) {
	return s.voteRepo.Exists(ctx, feedbackID, userID)
}

func (s *voteService) EnrichWithVoteStatus(ctx context.Context, feedbacks []domain.Feedback, userID uuid.UUID) error {
	if len(feedbacks) == 0 {
		return nil
	}

	// Collect feedback IDs
	feedbackIDs := make([]uuid.UUID, len(feedbacks))
	for i, f := range feedbacks {
		feedbackIDs[i] = f.ID
	}

	// Get votes for all feedbacks
	votes, err := s.voteRepo.ListByUser(ctx, userID, feedbackIDs)
	if err != nil {
		return fmt.Errorf("failed to list votes: %w", err)
	}

	// Enrich feedbacks
	for i := range feedbacks {
		feedbacks[i].HasVoted = votes[feedbacks[i].ID]
	}

	return nil
}
