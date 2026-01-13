package domain

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// InviteStatus represents the status of an invite
type InviteStatus string

const (
	InviteStatusPending  InviteStatus = "pending"
	InviteStatusAccepted InviteStatus = "accepted"
	InviteStatusExpired  InviteStatus = "expired"
	InviteStatusRevoked  InviteStatus = "revoked"
)

// IsValid checks if the invite status is valid
func (s InviteStatus) IsValid() bool {
	return s == InviteStatusPending || s == InviteStatusAccepted ||
		s == InviteStatusExpired || s == InviteStatusRevoked
}

// Invite represents an invitation to join a project
type Invite struct {
	ID         uuid.UUID    `json:"id"`
	ProjectID  uuid.UUID    `json:"project_id"`
	InvitedBy  uuid.UUID    `json:"invited_by"`
	Email      string       `json:"email"`
	Role       Role         `json:"role"`
	Token      string       `json:"-"` // Never expose token in JSON
	Status     InviteStatus `json:"status"`
	ExpiresAt  time.Time    `json:"expires_at"`
	CreatedAt  time.Time    `json:"created_at"`
	AcceptedAt *time.Time   `json:"accepted_at,omitempty"`

	// Relationships (populated by service layer)
	Project   *Project `json:"project,omitempty"`
	InvitedByUser *User `json:"invited_by_user,omitempty"`
}

// IsExpired returns true if the invite has expired
func (i *Invite) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// IsPending returns true if the invite is still pending
func (i *Invite) IsPending() bool {
	return i.Status == InviteStatusPending && !i.IsExpired()
}

// CanBeAccepted returns true if the invite can be accepted
func (i *Invite) CanBeAccepted() bool {
	return i.Status == InviteStatusPending && !i.IsExpired()
}

// CanAccept returns true if the invite can be accepted (alias for CanBeAccepted)
func (i *Invite) CanAccept() bool {
	return i.CanBeAccepted()
}

// Validate validates the invite data
func (i *Invite) Validate() error {
	if i.Email == "" {
		return fmt.Errorf("email is required")
	}
	if !isValidEmail(i.Email) {
		return fmt.Errorf("invalid email format")
	}
	if !i.Role.IsValid() {
		return fmt.Errorf("invalid role: %s", i.Role)
	}
	// Cannot invite as owner
	if i.Role == RoleOwner {
		return fmt.Errorf("cannot invite as owner")
	}
	return nil
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// InviteURL generates the full invite acceptance URL
func (i *Invite) InviteURL(baseURL string) string {
	return fmt.Sprintf("%s/auth/accept-invite/%s", baseURL, i.Token)
}
