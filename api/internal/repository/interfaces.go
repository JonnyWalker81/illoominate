package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/fulldisclosure/api/internal/domain"
)

// FeedbackFilter defines filter options for listing feedback
type FeedbackFilter struct {
	Type       *domain.FeedbackType
	Status     *domain.FeedbackStatus
	Visibility *domain.Visibility
	TagIDs     []uuid.UUID
	AssignedTo *uuid.UUID
	Search     *string
	SortBy     string // "created_at", "updated_at", "vote_count"
	SortOrder  string // "asc", "desc"
	Limit      int
	Offset     int
}

// FeedbackRepository defines the data access interface for feedback
type FeedbackRepository interface {
	Create(ctx context.Context, f *domain.Feedback) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Feedback, error)
	List(ctx context.Context, projectID uuid.UUID, filter FeedbackFilter) ([]domain.Feedback, int, error)
	Update(ctx context.Context, f *domain.Feedback) error
	Delete(ctx context.Context, id uuid.UUID) error
	Merge(ctx context.Context, sourceID, canonicalID uuid.UUID) error
	AddTag(ctx context.Context, feedbackID, tagID uuid.UUID) error
	RemoveTag(ctx context.Context, feedbackID, tagID uuid.UUID) error
}

// VoteRepository defines the data access interface for votes
type VoteRepository interface {
	Create(ctx context.Context, v *domain.Vote) error
	Delete(ctx context.Context, feedbackID, userID uuid.UUID) error
	Exists(ctx context.Context, feedbackID, userID uuid.UUID) (bool, error)
	CountByFeedback(ctx context.Context, feedbackID uuid.UUID) (int, error)
	ListByUser(ctx context.Context, userID uuid.UUID, feedbackIDs []uuid.UUID) (map[uuid.UUID]bool, error)
}

// CommentRepository defines the data access interface for comments
type CommentRepository interface {
	Create(ctx context.Context, c *domain.Comment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error)
	ListByFeedback(ctx context.Context, feedbackID uuid.UUID, includeTeamOnly bool) ([]domain.Comment, error)
	Update(ctx context.Context, c *domain.Comment) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ProjectRepository defines the data access interface for projects
type ProjectRepository interface {
	Create(ctx context.Context, p *domain.Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Project, error)
	GetByProjectKey(ctx context.Context, key string) (*domain.Project, error)
	Update(ctx context.Context, p *domain.Project) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// MembershipRepository defines the data access interface for memberships
type MembershipRepository interface {
	Create(ctx context.Context, m *domain.Membership) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Membership, error)
	GetByProjectAndUser(ctx context.Context, projectID, userID string) (*domain.Membership, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Membership, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Membership, error)
	Update(ctx context.Context, m *domain.Membership) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// InviteRepository defines the data access interface for invites
type InviteRepository interface {
	Create(ctx context.Context, i *domain.Invite) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Invite, error)
	GetByToken(ctx context.Context, token string) (*domain.Invite, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Invite, error)
	Update(ctx context.Context, i *domain.Invite) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) (int, error)
}

// TagRepository defines the data access interface for tags
type TagRepository interface {
	Create(ctx context.Context, t *domain.Tag) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Tag, error)
	GetBySlug(ctx context.Context, projectID uuid.UUID, slug string) (*domain.Tag, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Tag, error)
	ListByFeedback(ctx context.Context, feedbackID uuid.UUID) ([]domain.Tag, error)
	Update(ctx context.Context, t *domain.Tag) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// AttachmentRepository defines the data access interface for attachments
type AttachmentRepository interface {
	Create(ctx context.Context, a *domain.Attachment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Attachment, error)
	ListByFeedback(ctx context.Context, feedbackID uuid.UUID) ([]domain.Attachment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.AttachmentStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ActivityRepository defines the data access interface for activity logs
type ActivityRepository interface {
	Create(ctx context.Context, projectID, feedbackID uuid.UUID, actorID *uuid.UUID, action string, changes map[string]interface{}) error
	ListByFeedback(ctx context.Context, feedbackID uuid.UUID, limit, offset int) ([]ActivityEntry, int, error)
}

// ActivityEntry represents an activity log entry
type ActivityEntry struct {
	ID         uuid.UUID
	ProjectID  uuid.UUID
	FeedbackID *uuid.UUID
	ActorID    *uuid.UUID
	Action     string
	Changes    map[string]interface{}
	CreatedAt  string
}

// PortalRepository defines the data access interface for portal operations
type PortalRepository interface {
	// Profile operations
	CreateProfile(ctx context.Context, userID, projectID uuid.UUID) error
	GetProfile(ctx context.Context, userID, projectID uuid.UUID) (*domain.PortalUserProfile, error)
	UpdateNotificationPrefs(ctx context.Context, userID, projectID uuid.UUID, prefs domain.PortalNotificationPreferences) error

	// SDK user linking
	LinkSDKUsersByEmail(ctx context.Context, userID, projectID uuid.UUID, email string) (int64, error)

	// Feedback operations
	GetLinkedFeedback(ctx context.Context, userID, projectID uuid.UUID) ([]domain.PortalFeedbackSummary, error)
	ListPublicFeatures(ctx context.Context, projectID uuid.UUID, userID *uuid.UUID, limit, offset int) ([]domain.PortalFeedbackSummary, int, error)

	// Voting operations
	CreateVote(ctx context.Context, feedbackID, userID, projectID uuid.UUID) error
	DeleteVote(ctx context.Context, feedbackID, userID uuid.UUID) error
	HasVoted(ctx context.Context, feedbackID, userID uuid.UUID) (bool, error)
	GetVotedFeedbackIDs(ctx context.Context, userID uuid.UUID, feedbackIDs []uuid.UUID) (map[uuid.UUID]bool, error)
}
