package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/fulldisclosure/api/internal/domain"
	"github.com/fulldisclosure/api/internal/repository"
)

type tagService struct {
	tagRepo repository.TagRepository
}

// NewTagService creates a new tag service
func NewTagService(tagRepo repository.TagRepository) TagService {
	return &tagService{
		tagRepo: tagRepo,
	}
}

func (s *tagService) Create(ctx context.Context, projectID uuid.UUID, name string, color *string) (*domain.Tag, error) {
	tag := domain.NewTag(projectID, name, color)

	// Check if slug already exists
	existing, err := s.tagRepo.GetBySlug(ctx, projectID, tag.Slug)
	if err != nil && err != domain.ErrNotFound {
		return nil, fmt.Errorf("failed to check existing tag: %w", err)
	}

	if existing != nil {
		return nil, domain.NewDomainError("TAG_EXISTS", "A tag with this name already exists", 400)
	}

	if err := s.tagRepo.Create(ctx, tag); err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return tag, nil
}

func (s *tagService) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Tag, error) {
	tags, err := s.tagRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	return tags, nil
}

func (s *tagService) Update(ctx context.Context, tagID uuid.UUID, name *string, color *string) (*domain.Tag, error) {
	tag, err := s.tagRepo.GetByID(ctx, tagID)
	if err != nil {
		return nil, err
	}

	if name != nil {
		tag.Name = *name
		tag.Slug = domain.GenerateSlug(*name)

		// Check if new slug conflicts with existing tag
		existing, err := s.tagRepo.GetBySlug(ctx, tag.ProjectID, tag.Slug)
		if err != nil && err != domain.ErrNotFound {
			return nil, fmt.Errorf("failed to check existing tag: %w", err)
		}

		if existing != nil && existing.ID != tagID {
			return nil, domain.NewDomainError("TAG_EXISTS", "A tag with this name already exists", 400)
		}
	}

	if color != nil {
		tag.Color = *color
	}

	if err := s.tagRepo.Update(ctx, tag); err != nil {
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}

	return tag, nil
}

func (s *tagService) Delete(ctx context.Context, tagID uuid.UUID) error {
	return s.tagRepo.Delete(ctx, tagID)
}
