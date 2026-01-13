package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/fulldisclosure/api/internal/domain"
	"github.com/fulldisclosure/api/internal/repository"
)

type feedbackService struct {
	feedbackRepo repository.FeedbackRepository
	tagRepo      repository.TagRepository
	projectRepo  repository.ProjectRepository
}

// NewFeedbackService creates a new feedback service
func NewFeedbackService(
	feedbackRepo repository.FeedbackRepository,
	tagRepo repository.TagRepository,
	projectRepo repository.ProjectRepository,
) FeedbackService {
	return &feedbackService{
		feedbackRepo: feedbackRepo,
		tagRepo:      tagRepo,
		projectRepo:  projectRepo,
	}
}

func (s *feedbackService) Create(ctx context.Context, req CreateFeedbackRequest) (*domain.Feedback, error) {
	// Get project to determine default visibility
	project, err := s.projectRepo.GetByID(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Determine visibility based on project settings if not explicitly set
	visibility := req.Visibility
	if visibility == "" {
		visibility = project.Settings.GetDefaultVisibility(req.Type)
	}

	// Convert source metadata to JSON
	var sourceMetadata json.RawMessage
	if req.SourceMetadata != nil {
		var err error
		sourceMetadata, err = json.Marshal(req.SourceMetadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal source metadata: %w", err)
		}
	}

	feedback := &domain.Feedback{
		ID:             uuid.New(),
		ProjectID:      req.ProjectID,
		AuthorID:       req.AuthorID,
		Title:          req.Title,
		Description:    req.Description,
		Type:           req.Type,
		Status:         domain.FeedbackStatusNew,
		Visibility:     visibility,
		SubmitterEmail: req.SubmitterEmail,
		SubmitterName:  req.SubmitterName,
		Source:         req.Source,
		SourceMetadata: sourceMetadata,
	}

	if err := s.feedbackRepo.Create(ctx, feedback); err != nil {
		return nil, fmt.Errorf("failed to create feedback: %w", err)
	}

	return feedback, nil
}

func (s *feedbackService) GetByID(ctx context.Context, projectID, feedbackID uuid.UUID, userRole domain.Role) (*domain.Feedback, error) {
	feedback, err := s.feedbackRepo.GetByID(ctx, feedbackID)
	if err != nil {
		return nil, err
	}

	// Check if feedback belongs to the project
	if feedback.ProjectID != projectID {
		return nil, domain.ErrNotFound
	}

	// Check visibility permissions
	if !feedback.CanView(userRole) {
		return nil, domain.ErrForbidden
	}

	// If feedback is merged, redirect to canonical
	if feedback.CanonicalID != nil {
		return s.GetByID(ctx, projectID, *feedback.CanonicalID, userRole)
	}

	// Load tags
	tags, err := s.tagRepo.ListByFeedback(ctx, feedbackID)
	if err != nil {
		return nil, fmt.Errorf("failed to load tags: %w", err)
	}
	feedback.Tags = tags

	return feedback, nil
}

func (s *feedbackService) List(ctx context.Context, projectID uuid.UUID, filter FeedbackFilter, userRole domain.Role) (*FeedbackListResult, error) {
	// Apply visibility filter based on role
	var visibility *domain.Visibility
	if !userRole.IsTeamRole() {
		// Community users can only see COMMUNITY visibility
		v := domain.VisibilityCommunity
		visibility = &v
	} else if filter.Visibility != nil {
		visibility = filter.Visibility
	}

	// Calculate pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 || perPage > 100 {
		perPage = 50
	}
	offset := (page - 1) * perPage

	repoFilter := repository.FeedbackFilter{
		Type:       filter.Type,
		Status:     filter.Status,
		Visibility: visibility,
		TagIDs:     filter.TagIDs,
		AssignedTo: filter.AssignedTo,
		Search:     filter.Search,
		SortBy:     filter.SortBy,
		SortOrder:  filter.SortOrder,
		Limit:      perPage,
		Offset:     offset,
	}

	feedbacks, total, err := s.feedbackRepo.List(ctx, projectID, repoFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list feedback: %w", err)
	}

	totalPages := (total + perPage - 1) / perPage

	return &FeedbackListResult{
		Feedbacks:  feedbacks,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (s *feedbackService) Update(ctx context.Context, feedbackID uuid.UUID, req UpdateFeedbackRequest, actorID uuid.UUID) (*domain.Feedback, error) {
	feedback, err := s.feedbackRepo.GetByID(ctx, feedbackID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.Title != nil {
		feedback.Title = *req.Title
	}
	if req.Description != nil {
		feedback.Description = *req.Description
	}
	if req.Status != nil {
		oldStatus := feedback.Status
		feedback.Status = *req.Status

		// Set resolved_at when completing
		if *req.Status == domain.FeedbackStatusCompleted && oldStatus != domain.FeedbackStatusCompleted {
			now := time.Now()
			feedback.ResolvedAt = &now
		} else if *req.Status != domain.FeedbackStatusCompleted {
			feedback.ResolvedAt = nil
		}
	}
	if req.Severity != nil {
		feedback.Severity = req.Severity
	}
	if req.Visibility != nil {
		feedback.Visibility = *req.Visibility
	}
	if req.AssignedTo != nil {
		feedback.AssignedTo = req.AssignedTo
	}

	if err := s.feedbackRepo.Update(ctx, feedback); err != nil {
		return nil, fmt.Errorf("failed to update feedback: %w", err)
	}

	// Update tags if provided
	if len(req.TagIDs) > 0 {
		// Get current tags
		currentTags, err := s.tagRepo.ListByFeedback(ctx, feedbackID)
		if err != nil {
			return nil, fmt.Errorf("failed to get current tags: %w", err)
		}

		currentTagIDs := make(map[uuid.UUID]bool)
		for _, t := range currentTags {
			currentTagIDs[t.ID] = true
		}

		newTagIDs := make(map[uuid.UUID]bool)
		for _, id := range req.TagIDs {
			newTagIDs[id] = true
		}

		// Add new tags
		for id := range newTagIDs {
			if !currentTagIDs[id] {
				if err := s.feedbackRepo.AddTag(ctx, feedbackID, id); err != nil {
					return nil, fmt.Errorf("failed to add tag: %w", err)
				}
			}
		}

		// Remove old tags
		for id := range currentTagIDs {
			if !newTagIDs[id] {
				if err := s.feedbackRepo.RemoveTag(ctx, feedbackID, id); err != nil {
					return nil, fmt.Errorf("failed to remove tag: %w", err)
				}
			}
		}
	}

	// Reload with tags
	return s.feedbackRepo.GetByID(ctx, feedbackID)
}

func (s *feedbackService) Merge(ctx context.Context, sourceID, canonicalID uuid.UUID, actorID uuid.UUID) (*domain.Feedback, error) {
	// Verify both exist
	source, err := s.feedbackRepo.GetByID(ctx, sourceID)
	if err != nil {
		return nil, fmt.Errorf("source feedback not found: %w", err)
	}

	canonical, err := s.feedbackRepo.GetByID(ctx, canonicalID)
	if err != nil {
		return nil, fmt.Errorf("canonical feedback not found: %w", err)
	}

	// Verify they belong to the same project
	if source.ProjectID != canonical.ProjectID {
		return nil, domain.NewDomainError("CROSS_PROJECT_MERGE", "Cannot merge feedback from different projects", 400)
	}

	// Cannot merge into a merged item
	if canonical.CanonicalID != nil {
		return nil, domain.NewDomainError("MERGED_CANONICAL", "Cannot merge into already merged feedback", 400)
	}

	// Cannot merge an already merged item
	if source.CanonicalID != nil {
		return nil, domain.NewDomainError("ALREADY_MERGED", "Source feedback is already merged", 400)
	}

	// Perform the merge
	if err := s.feedbackRepo.Merge(ctx, sourceID, canonicalID); err != nil {
		return nil, fmt.Errorf("failed to merge feedback: %w", err)
	}

	// Return updated canonical
	return s.feedbackRepo.GetByID(ctx, canonicalID)
}
