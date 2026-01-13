package domain

import (
	"time"

	"github.com/google/uuid"
)

// PortalUserProfile represents a portal user's profile for a specific project
type PortalUserProfile struct {
	ID                      uuid.UUID                     `json:"id"`
	UserID                  uuid.UUID                     `json:"user_id"`
	ProjectID               uuid.UUID                     `json:"project_id"`
	NotificationPreferences PortalNotificationPreferences `json:"notification_preferences"`
	CreatedAt               time.Time                     `json:"created_at"`
	UpdatedAt               time.Time                     `json:"updated_at"`
}

// PortalNotificationPreferences defines portal user notification settings
type PortalNotificationPreferences struct {
	StatusChanges              bool `json:"status_changes"`
	NewCommentsOnMyFeedback    bool `json:"new_comments_on_my_feedback"`
	NewCommentsOnVotedFeedback bool `json:"new_comments_on_voted_feedback,omitempty"`
	WeeklyDigest               bool `json:"weekly_digest,omitempty"`
}

// DefaultPortalNotificationPreferences returns default portal notification preferences
func DefaultPortalNotificationPreferences() PortalNotificationPreferences {
	return PortalNotificationPreferences{
		StatusChanges:           true,
		NewCommentsOnMyFeedback: true,
	}
}

// PortalVote represents a vote from a portal user
type PortalVote struct {
	ID         uuid.UUID `json:"id"`
	FeedbackID uuid.UUID `json:"feedback_id"`
	UserID     uuid.UUID `json:"user_id"`
	ProjectID  uuid.UUID `json:"project_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// NotificationQueueEntry represents a pending notification
type NotificationQueueEntry struct {
	ID               uuid.UUID              `json:"id"`
	UserID           uuid.UUID              `json:"user_id"`
	ProjectID        uuid.UUID              `json:"project_id"`
	NotificationType string                 `json:"notification_type"`
	Payload          map[string]interface{} `json:"payload"`
	SentAt           *time.Time             `json:"sent_at,omitempty"`
	FailedAt         *time.Time             `json:"failed_at,omitempty"`
	FailureReason    *string                `json:"failure_reason,omitempty"`
	RetryCount       int                    `json:"retry_count"`
	CreatedAt        time.Time              `json:"created_at"`
}

// Notification types
const (
	NotificationTypeStatusChanged    = "status_changed"
	NotificationTypeNewComment       = "new_comment"
	NotificationTypeFeedbackResolved = "feedback_resolved"
	NotificationTypeWeeklyDigest     = "weekly_digest"
)

// PortalFeedbackSummary is a lightweight feedback representation for portal users
type PortalFeedbackSummary struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Type         string    `json:"type"`
	Status       string    `json:"status"`
	VoteCount    int       `json:"vote_count"`
	CommentCount int       `json:"comment_count"`
	HasVoted     bool      `json:"has_voted"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
