package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/fulldisclosure/api/internal/domain"
	"github.com/fulldisclosure/api/internal/repository"
)

type membershipService struct {
	membershipRepo repository.MembershipRepository
}

// NewMembershipService creates a new membership service
func NewMembershipService(membershipRepo repository.MembershipRepository) MembershipService {
	return &membershipService{
		membershipRepo: membershipRepo,
	}
}

func (s *membershipService) CheckAccess(ctx context.Context, projectID, userID uuid.UUID, requiredRole domain.Role) error {
	membership, err := s.membershipRepo.GetByProjectAndUser(ctx, projectID.String(), userID.String())
	if err != nil {
		return fmt.Errorf("failed to get membership: %w", err)
	}

	if membership == nil {
		return domain.ErrForbidden
	}

	if !membership.HasRoleAtLeast(requiredRole) {
		return domain.ErrForbidden
	}

	return nil
}

func (s *membershipService) GetUserRole(ctx context.Context, projectID, userID uuid.UUID) (domain.Role, error) {
	membership, err := s.membershipRepo.GetByProjectAndUser(ctx, projectID.String(), userID.String())
	if err != nil {
		return "", fmt.Errorf("failed to get membership: %w", err)
	}

	if membership == nil {
		return "", domain.ErrNotFound
	}

	return membership.Role, nil
}

func (s *membershipService) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Membership, error) {
	memberships, err := s.membershipRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list memberships: %w", err)
	}

	return memberships, nil
}

func (s *membershipService) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Membership, error) {
	memberships, err := s.membershipRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user memberships: %w", err)
	}

	return memberships, nil
}

func (s *membershipService) UpdateRole(ctx context.Context, membershipID uuid.UUID, newRole domain.Role, actorID uuid.UUID) (*domain.Membership, error) {
	membership, err := s.membershipRepo.GetByID(ctx, membershipID)
	if err != nil {
		return nil, err
	}

	// Cannot change owner role
	if membership.Role == domain.RoleOwner {
		return nil, domain.NewDomainError("CANNOT_CHANGE_OWNER", "Cannot change the owner's role", 400)
	}

	// Cannot promote to owner
	if newRole == domain.RoleOwner {
		return nil, domain.NewDomainError("CANNOT_PROMOTE_OWNER", "Cannot promote to owner role", 400)
	}

	// Get actor's membership to check permissions
	actorMembership, err := s.membershipRepo.GetByProjectAndUser(ctx, membership.ProjectID.String(), actorID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get actor membership: %w", err)
	}

	if actorMembership == nil {
		return nil, domain.ErrForbidden
	}

	// Only admins and owners can change roles
	if !actorMembership.CanManageRole(newRole) {
		return nil, domain.ErrForbidden
	}

	membership.Role = newRole

	if err := s.membershipRepo.Update(ctx, membership); err != nil {
		return nil, fmt.Errorf("failed to update membership: %w", err)
	}

	return membership, nil
}

func (s *membershipService) Remove(ctx context.Context, membershipID uuid.UUID, actorID uuid.UUID) error {
	membership, err := s.membershipRepo.GetByID(ctx, membershipID)
	if err != nil {
		return err
	}

	// Cannot remove owner
	if membership.Role == domain.RoleOwner {
		return domain.NewDomainError("CANNOT_REMOVE_OWNER", "Cannot remove the owner from the project", 400)
	}

	// Get actor's membership to check permissions
	actorMembership, err := s.membershipRepo.GetByProjectAndUser(ctx, membership.ProjectID.String(), actorID.String())
	if err != nil {
		return fmt.Errorf("failed to get actor membership: %w", err)
	}

	if actorMembership == nil {
		return domain.ErrForbidden
	}

	// Users can remove themselves
	if membership.UserID == actorID {
		return s.membershipRepo.Delete(ctx, membershipID)
	}

	// Only admins and owners can remove others
	if !actorMembership.Role.IsAdminOrOwner() {
		return domain.ErrForbidden
	}

	// Admins cannot remove other admins or owners
	if actorMembership.Role == domain.RoleAdmin && membership.Role.IsAdminOrOwner() {
		return domain.ErrForbidden
	}

	return s.membershipRepo.Delete(ctx, membershipID)
}
