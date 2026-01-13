package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/fulldisclosure/api/internal/domain"
	"github.com/fulldisclosure/api/internal/repository"
)

type inviteService struct {
	inviteRepo     repository.InviteRepository
	membershipRepo repository.MembershipRepository
}

// NewInviteService creates a new invite service
func NewInviteService(
	inviteRepo repository.InviteRepository,
	membershipRepo repository.MembershipRepository,
) InviteService {
	return &inviteService{
		inviteRepo:     inviteRepo,
		membershipRepo: membershipRepo,
	}
}

func (s *inviteService) Create(ctx context.Context, req InviteMemberRequest) (*domain.Invite, error) {
	// Check if user is already a member
	existing, err := s.membershipRepo.GetByProjectAndUser(ctx, req.ProjectID.String(), req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing membership: %w", err)
	}

	if existing != nil {
		return nil, domain.NewDomainError("ALREADY_MEMBER", "User is already a member of this project", 400)
	}

	// Generate secure token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	invite := &domain.Invite{
		ID:        uuid.New(),
		ProjectID: req.ProjectID,
		InvitedBy: req.InviterID,
		Email:     req.Email,
		Role:      req.Role,
		Token:     token,
		Status:    domain.InviteStatusPending,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := s.inviteRepo.Create(ctx, invite); err != nil {
		return nil, fmt.Errorf("failed to create invite: %w", err)
	}

	return invite, nil
}

func (s *inviteService) Accept(ctx context.Context, token string, userID uuid.UUID) (*domain.Membership, error) {
	invite, err := s.inviteRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Check if invite is still valid
	if !invite.CanAccept() {
		return nil, domain.NewDomainError("INVITE_INVALID", "This invite is no longer valid", 400)
	}

	// Check if user is already a member
	existing, err := s.membershipRepo.GetByProjectAndUser(ctx, invite.ProjectID.String(), userID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to check existing membership: %w", err)
	}

	if existing != nil {
		// Already a member - mark invite as accepted and return existing membership
		invite.Status = domain.InviteStatusAccepted
		now := time.Now()
		invite.AcceptedAt = &now
		if err := s.inviteRepo.Update(ctx, invite); err != nil {
			return nil, fmt.Errorf("failed to update invite: %w", err)
		}
		return existing, nil
	}

	// Create membership
	membership := &domain.Membership{
		ID:        uuid.New(),
		ProjectID: invite.ProjectID,
		UserID:    userID,
		Role:      invite.Role,
	}

	if err := s.membershipRepo.Create(ctx, membership); err != nil {
		return nil, fmt.Errorf("failed to create membership: %w", err)
	}

	// Mark invite as accepted
	invite.Status = domain.InviteStatusAccepted
	now := time.Now()
	invite.AcceptedAt = &now
	if err := s.inviteRepo.Update(ctx, invite); err != nil {
		return nil, fmt.Errorf("failed to update invite: %w", err)
	}

	return membership, nil
}

func (s *inviteService) Revoke(ctx context.Context, inviteID uuid.UUID, actorID uuid.UUID) error {
	invite, err := s.inviteRepo.GetByID(ctx, inviteID)
	if err != nil {
		return err
	}

	if invite.Status != domain.InviteStatusPending {
		return domain.NewDomainError("INVITE_NOT_PENDING", "Can only revoke pending invites", 400)
	}

	invite.Status = domain.InviteStatusRevoked
	return s.inviteRepo.Update(ctx, invite)
}

func (s *inviteService) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Invite, error) {
	invites, err := s.inviteRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list invites: %w", err)
	}

	return invites, nil
}
