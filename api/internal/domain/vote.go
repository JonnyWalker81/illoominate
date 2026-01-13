package domain

import (
	"time"

	"github.com/google/uuid"
)

// Vote represents a user's vote on a feedback item
type Vote struct {
	ID         uuid.UUID `json:"id"`
	FeedbackID uuid.UUID `json:"feedback_id"`
	UserID     uuid.UUID `json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// VoteResult represents the result of a vote operation
type VoteResult struct {
	FeedbackID uuid.UUID `json:"feedback_id"`
	VoteCount  int       `json:"vote_count"`
	HasVoted   bool      `json:"has_voted"`
}
