package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/fulldisclosure/api/internal/domain"
)

// CreateFeedbackRequest contains the data needed to create feedback
type CreateFeedbackRequest struct {
	ProjectID      uuid.UUID
	AuthorID       *uuid.UUID
	Title          string
	Description    string
	Type           domain.FeedbackType
	Visibility     domain.Visibility
	SubmitterEmail *string
	SubmitterName  *string
	Source         string
	SourceMetadata map[string]interface{}
}

// UpdateFeedbackRequest contains the data that can be updated
type UpdateFeedbackRequest struct {
	Title       *string
	Description *string
	Status      *domain.FeedbackStatus
	Severity    *domain.Severity
	Visibility  *domain.Visibility
	AssignedTo  *uuid.UUID
	TagIDs      []uuid.UUID
}

// FeedbackFilter defines filter options for listing feedback
type FeedbackFilter struct {
	Type       *domain.FeedbackType
	Status     *domain.FeedbackStatus
	Visibility *domain.Visibility
	TagIDs     []uuid.UUID
	AssignedTo *uuid.UUID
	Search     *string
	SortBy     string
	SortOrder  string
	Page       int
	PerPage    int
}

// FeedbackListResult contains paginated feedback results
type FeedbackListResult struct {
	Feedbacks  []domain.Feedback
	Total      int
	Page       int
	PerPage    int
	TotalPages int
}

// VoteResult contains the result of a vote operation
type VoteResult struct {
	FeedbackID uuid.UUID
	VoteCount  int
	HasVoted   bool
}

// InviteMemberRequest contains the data needed to invite a member
type InviteMemberRequest struct {
	ProjectID uuid.UUID
	InviterID uuid.UUID
	Email     string
	Role      domain.Role
}

// CreateCommentRequest contains the data needed to create a comment
type CreateCommentRequest struct {
	FeedbackID uuid.UUID
	AuthorID   uuid.UUID
	ParentID   *uuid.UUID
	Body       string
	Visibility domain.Visibility
}

// CreateProjectRequest contains the data needed to create a project
type CreateProjectRequest struct {
	Name         string
	OwnerID      uuid.UUID
	LogoURL      *string
	PrimaryColor string
}

// FeedbackService defines the business logic interface for feedback
type FeedbackService interface {
	Create(ctx context.Context, req CreateFeedbackRequest) (*domain.Feedback, error)
	GetByID(ctx context.Context, projectID, feedbackID uuid.UUID, userRole domain.Role) (*domain.Feedback, error)
	List(ctx context.Context, projectID uuid.UUID, filter FeedbackFilter, userRole domain.Role) (*FeedbackListResult, error)
	Update(ctx context.Context, feedbackID uuid.UUID, req UpdateFeedbackRequest, actorID uuid.UUID) (*domain.Feedback, error)
	Merge(ctx context.Context, sourceID, canonicalID uuid.UUID, actorID uuid.UUID) (*domain.Feedback, error)
}

// VoteService defines the business logic interface for votes
type VoteService interface {
	Vote(ctx context.Context, projectID, feedbackID, userID uuid.UUID) (*VoteResult, error)
	Unvote(ctx context.Context, projectID, feedbackID, userID uuid.UUID) (*VoteResult, error)
	HasVoted(ctx context.Context, feedbackID, userID uuid.UUID) (bool, error)
	EnrichWithVoteStatus(ctx context.Context, feedbacks []domain.Feedback, userID uuid.UUID) error
}

// CommentService defines the business logic interface for comments
type CommentService interface {
	Create(ctx context.Context, req CreateCommentRequest) (*domain.Comment, error)
	ListByFeedback(ctx context.Context, feedbackID uuid.UUID, includeTeamOnly bool) ([]domain.Comment, error)
	Update(ctx context.Context, commentID uuid.UUID, body string, actorID uuid.UUID) (*domain.Comment, error)
	Delete(ctx context.Context, commentID uuid.UUID, actorID uuid.UUID) error
}

// ProjectService defines the business logic interface for projects
type ProjectService interface {
	Create(ctx context.Context, req CreateProjectRequest) (*domain.Project, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Project, error)
	Update(ctx context.Context, id uuid.UUID, name *string, settings *domain.ProjectSettings, actorID uuid.UUID) (*domain.Project, error)
}

// MembershipService defines the business logic interface for memberships
type MembershipService interface {
	CheckAccess(ctx context.Context, projectID, userID uuid.UUID, requiredRole domain.Role) error
	GetUserRole(ctx context.Context, projectID, userID uuid.UUID) (domain.Role, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Membership, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Membership, error)
	UpdateRole(ctx context.Context, membershipID uuid.UUID, newRole domain.Role, actorID uuid.UUID) (*domain.Membership, error)
	Remove(ctx context.Context, membershipID uuid.UUID, actorID uuid.UUID) error
}

// InviteService defines the business logic interface for invites
type InviteService interface {
	Create(ctx context.Context, req InviteMemberRequest) (*domain.Invite, error)
	Accept(ctx context.Context, token string, userID uuid.UUID) (*domain.Membership, error)
	Revoke(ctx context.Context, inviteID uuid.UUID, actorID uuid.UUID) error
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Invite, error)
}

// TagService defines the business logic interface for tags
type TagService interface {
	Create(ctx context.Context, projectID uuid.UUID, name string, color *string) (*domain.Tag, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Tag, error)
	Update(ctx context.Context, tagID uuid.UUID, name *string, color *string) (*domain.Tag, error)
	Delete(ctx context.Context, tagID uuid.UUID) error
}

// AttachmentService defines the business logic interface for attachments
type AttachmentService interface {
	InitiateUpload(ctx context.Context, feedbackID uuid.UUID, uploaderID *uuid.UUID, filename string, contentType string, sizeBytes int64) (*UploadInfo, error)
	CompleteUpload(ctx context.Context, attachmentID uuid.UUID) (*domain.Attachment, error)
	GetDownloadURL(ctx context.Context, attachmentID uuid.UUID) (string, error)
	Delete(ctx context.Context, attachmentID uuid.UUID, actorID uuid.UUID) error
}

// UploadInfo contains signed URL info for uploading
type UploadInfo struct {
	AttachmentID uuid.UUID
	UploadURL    string
	ExpiresAt    string
}
