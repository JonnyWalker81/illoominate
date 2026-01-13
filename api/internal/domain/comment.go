package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Comment represents a comment on a feedback item
type Comment struct {
	ID         uuid.UUID  `json:"id"`
	FeedbackID uuid.UUID  `json:"feedback_id"`
	AuthorID   uuid.UUID  `json:"author_id"`
	ParentID   *uuid.UUID `json:"parent_id,omitempty"`
	Body       string     `json:"body"`
	Visibility Visibility `json:"visibility"`
	IsEdited   bool       `json:"is_edited"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`

	// Relationships (populated by service layer)
	Author   *User      `json:"author,omitempty"`
	Replies  []Comment  `json:"replies,omitempty"`
}

// IsDeleted returns true if the comment has been soft-deleted
func (c *Comment) IsDeleted() bool {
	return c.DeletedAt != nil
}

// IsReply returns true if this is a reply to another comment
func (c *Comment) IsReply() bool {
	return c.ParentID != nil
}

// CanBeViewedBy checks if the comment can be viewed by a user with the given role
func (c *Comment) CanBeViewedBy(role Role) bool {
	if c.IsDeleted() {
		return false
	}
	return c.Visibility.IsVisibleTo(role)
}

// Validate validates the comment data
func (c *Comment) Validate() error {
	if c.Body == "" {
		return fmt.Errorf("body is required")
	}
	if len(c.Body) > 10000 {
		return fmt.Errorf("body must be 10000 characters or less")
	}
	if !c.Visibility.IsValid() {
		return fmt.Errorf("invalid visibility: %s", c.Visibility)
	}
	return nil
}
