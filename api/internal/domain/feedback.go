package domain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FeedbackType represents the type of feedback
type FeedbackType string

const (
	FeedbackTypeBug     FeedbackType = "bug"
	FeedbackTypeFeature FeedbackType = "feature"
	FeedbackTypeGeneral FeedbackType = "general"
)

// IsValid checks if the feedback type is valid
func (t FeedbackType) IsValid() bool {
	return t == FeedbackTypeBug || t == FeedbackTypeFeature || t == FeedbackTypeGeneral
}

// Visibility represents the visibility level of feedback/comments
type Visibility string

const (
	VisibilityTeamOnly  Visibility = "TEAM_ONLY"
	VisibilityCommunity Visibility = "COMMUNITY"
)

// IsValid checks if the visibility is valid
func (v Visibility) IsValid() bool {
	return v == VisibilityTeamOnly || v == VisibilityCommunity
}

// IsVisibleTo checks if content with this visibility is visible to the given role
func (v Visibility) IsVisibleTo(role Role) bool {
	if v == VisibilityCommunity {
		return true // Everyone can see community content
	}
	// TEAM_ONLY requires team role
	return role.IsTeamRole()
}

// FeedbackStatus represents the status of feedback
type FeedbackStatus string

const (
	StatusNew         FeedbackStatus = "new"
	StatusUnderReview FeedbackStatus = "under_review"
	StatusPlanned     FeedbackStatus = "planned"
	StatusInProgress  FeedbackStatus = "in_progress"
	StatusCompleted   FeedbackStatus = "completed"
	StatusDeclined    FeedbackStatus = "declined"
	StatusDuplicate   FeedbackStatus = "duplicate"

	// Aliases for compatibility
	FeedbackStatusNew       = StatusNew
	FeedbackStatusCompleted = StatusCompleted
)

// IsValid checks if the status is valid
func (s FeedbackStatus) IsValid() bool {
	switch s {
	case StatusNew, StatusUnderReview, StatusPlanned, StatusInProgress,
		StatusCompleted, StatusDeclined, StatusDuplicate:
		return true
	}
	return false
}

// IsResolved returns true if the status is a resolved status
func (s FeedbackStatus) IsResolved() bool {
	return s == StatusCompleted || s == StatusDeclined || s == StatusDuplicate
}

// Severity represents the severity level (primarily for bugs)
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// IsValid checks if the severity is valid
func (s Severity) IsValid() bool {
	return s == SeverityLow || s == SeverityMedium || s == SeverityHigh || s == SeverityCritical
}

// Feedback represents a feedback item (bug, feature request, or general feedback)
type Feedback struct {
	ID          uuid.UUID      `json:"id"`
	ProjectID   uuid.UUID      `json:"project_id"`
	AuthorID    *uuid.UUID     `json:"author_id,omitempty"`
	AssignedTo  *uuid.UUID     `json:"assigned_to,omitempty"`
	CanonicalID *uuid.UUID     `json:"canonical_id,omitempty"`

	Title       string         `json:"title"`
	Description string         `json:"description"`
	Type        FeedbackType   `json:"type"`
	Status      FeedbackStatus `json:"status"`
	Severity    *Severity      `json:"severity,omitempty"`
	Visibility  Visibility     `json:"visibility"`

	VoteCount    int `json:"vote_count"`
	CommentCount int `json:"comment_count"`

	// Anonymous submitter info
	SubmitterEmail      *string `json:"submitter_email,omitempty"`
	SubmitterName       *string `json:"submitter_name,omitempty"`
	SubmitterIdentifier *string `json:"submitter_identifier,omitempty"`

	// Source tracking
	Source         string          `json:"source"`
	SourceURL      *string         `json:"source_url,omitempty"`
	SourceMetadata json.RawMessage `json:"source_metadata,omitempty"`

	// Timestamps
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`

	// Relationships (populated by service layer)
	Tags     []Tag    `json:"tags,omitempty"`
	Author   *User    `json:"author,omitempty"`
	Assignee *User    `json:"assignee,omitempty"`

	// Computed fields (populated by service layer)
	HasVoted bool `json:"has_voted,omitempty"`
}

// IsMerged returns true if this feedback has been merged into another
func (f *Feedback) IsMerged() bool {
	return f.CanonicalID != nil
}

// IsAnonymous returns true if the feedback was submitted anonymously
func (f *Feedback) IsAnonymous() bool {
	return f.AuthorID == nil
}

// CanBeViewedBy checks if the feedback can be viewed by a user with the given role
func (f *Feedback) CanBeViewedBy(role Role) bool {
	return f.Visibility.IsVisibleTo(role)
}

// CanView checks if the feedback can be viewed by a user with the given role (alias)
func (f *Feedback) CanView(role Role) bool {
	return f.CanBeViewedBy(role)
}

// Validate validates the feedback data
func (f *Feedback) Validate() error {
	if f.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(f.Title) > 200 {
		return fmt.Errorf("title must be 200 characters or less")
	}
	if f.Description == "" {
		return fmt.Errorf("description is required")
	}
	if !f.Type.IsValid() {
		return fmt.Errorf("invalid feedback type: %s", f.Type)
	}
	if !f.Status.IsValid() {
		return fmt.Errorf("invalid feedback status: %s", f.Status)
	}
	if !f.Visibility.IsValid() {
		return fmt.Errorf("invalid visibility: %s", f.Visibility)
	}
	if f.Severity != nil && !f.Severity.IsValid() {
		return fmt.Errorf("invalid severity: %s", *f.Severity)
	}
	return nil
}

// User represents a minimal user for embedding in other types
type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email,omitempty"`
	Name      string    `json:"name,omitempty"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
}
