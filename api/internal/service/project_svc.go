package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/fulldisclosure/api/internal/domain"
	"github.com/fulldisclosure/api/internal/repository"
)

type projectService struct {
	projectRepo    repository.ProjectRepository
	membershipRepo repository.MembershipRepository
}

// NewProjectService creates a new project service
func NewProjectService(
	projectRepo repository.ProjectRepository,
	membershipRepo repository.MembershipRepository,
) ProjectService {
	return &projectService{
		projectRepo:    projectRepo,
		membershipRepo: membershipRepo,
	}
}

func (s *projectService) Create(ctx context.Context, req CreateProjectRequest) (*domain.Project, error) {
	// Generate unique slug
	slug := generateSlug(req.Name)

	// Check if slug exists, append random suffix if needed
	existing, err := s.projectRepo.GetBySlug(ctx, slug)
	if err != nil && err != domain.ErrNotFound {
		return nil, fmt.Errorf("failed to check slug: %w", err)
	}

	if existing != nil {
		// Append random suffix
		suffix := make([]byte, 4)
		rand.Read(suffix)
		slug = slug + "-" + hex.EncodeToString(suffix)
	}

	// Generate project key
	keyBytes := make([]byte, 16)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate project key: %w", err)
	}
	projectKey := "proj_" + hex.EncodeToString(keyBytes)

	project := &domain.Project{
		ID:           uuid.New(),
		Name:         req.Name,
		Slug:         slug,
		ProjectKey:   projectKey,
		Settings:     domain.DefaultProjectSettings(),
		LogoURL:      req.LogoURL,
		PrimaryColor: req.PrimaryColor,
	}

	if project.PrimaryColor == "" {
		project.PrimaryColor = "#6366f1" // Default indigo
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Create owner membership
	membership := &domain.Membership{
		ID:        uuid.New(),
		ProjectID: project.ID,
		UserID:    req.OwnerID,
		Role:      domain.RoleOwner,
	}

	if err := s.membershipRepo.Create(ctx, membership); err != nil {
		// Rollback project creation
		_ = s.projectRepo.Delete(ctx, project.ID)
		return nil, fmt.Errorf("failed to create owner membership: %w", err)
	}

	return project, nil
}

func (s *projectService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	return s.projectRepo.GetByID(ctx, id)
}

func (s *projectService) GetBySlug(ctx context.Context, slug string) (*domain.Project, error) {
	return s.projectRepo.GetBySlug(ctx, slug)
}

func (s *projectService) Update(ctx context.Context, id uuid.UUID, name *string, settings *domain.ProjectSettings, actorID uuid.UUID) (*domain.Project, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != nil {
		project.Name = *name
		// Update slug
		newSlug := generateSlug(*name)

		// Check if new slug conflicts
		existing, err := s.projectRepo.GetBySlug(ctx, newSlug)
		if err != nil && err != domain.ErrNotFound {
			return nil, fmt.Errorf("failed to check slug: %w", err)
		}

		if existing != nil && existing.ID != id {
			// Append random suffix
			suffix := make([]byte, 4)
			rand.Read(suffix)
			newSlug = newSlug + "-" + hex.EncodeToString(suffix)
		}

		project.Slug = newSlug
	}

	if settings != nil {
		project.Settings = *settings
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return project, nil
}

// generateSlug creates a URL-friendly slug from a name
func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}
