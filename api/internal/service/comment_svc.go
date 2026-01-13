package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/fulldisclosure/api/internal/domain"
	"github.com/fulldisclosure/api/internal/repository"
)

type commentService struct {
	commentRepo  repository.CommentRepository
	feedbackRepo repository.FeedbackRepository
}

// NewCommentService creates a new comment service
func NewCommentService(
	commentRepo repository.CommentRepository,
	feedbackRepo repository.FeedbackRepository,
) CommentService {
	return &commentService{
		commentRepo:  commentRepo,
		feedbackRepo: feedbackRepo,
	}
}

func (s *commentService) Create(ctx context.Context, req CreateCommentRequest) (*domain.Comment, error) {
	// Verify feedback exists
	feedback, err := s.feedbackRepo.GetByID(ctx, req.FeedbackID)
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback: %w", err)
	}

	// Validate parent if provided
	if req.ParentID != nil {
		parent, err := s.commentRepo.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent comment not found: %w", err)
		}

		// Parent must belong to same feedback
		if parent.FeedbackID != req.FeedbackID {
			return nil, domain.NewDomainError("INVALID_PARENT", "Parent comment belongs to different feedback", 400)
		}
	}

	comment := &domain.Comment{
		ID:         uuid.New(),
		FeedbackID: feedback.ID,
		AuthorID:   req.AuthorID,
		ParentID:   req.ParentID,
		Body:       req.Body,
		Visibility: req.Visibility,
		IsEdited:   false,
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return comment, nil
}

func (s *commentService) ListByFeedback(ctx context.Context, feedbackID uuid.UUID, includeTeamOnly bool) ([]domain.Comment, error) {
	comments, err := s.commentRepo.ListByFeedback(ctx, feedbackID, includeTeamOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}

	return comments, nil
}

func (s *commentService) Update(ctx context.Context, commentID uuid.UUID, body string, actorID uuid.UUID) (*domain.Comment, error) {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	// Only author can edit
	if comment.AuthorID != actorID {
		return nil, domain.ErrForbidden
	}

	comment.Body = body

	if err := s.commentRepo.Update(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	return comment, nil
}

func (s *commentService) Delete(ctx context.Context, commentID uuid.UUID, actorID uuid.UUID) error {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return err
	}

	// Only author can delete (admins handled at handler level)
	if comment.AuthorID != actorID {
		return domain.ErrForbidden
	}

	return s.commentRepo.Delete(ctx, commentID)
}
