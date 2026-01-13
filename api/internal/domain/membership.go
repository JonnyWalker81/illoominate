package domain

import (
	"time"

	"github.com/google/uuid"
)

// Membership represents a user's membership in a project
type Membership struct {
	ID          uuid.UUID  `json:"id"`
	ProjectID   uuid.UUID  `json:"project_id"`
	UserID      uuid.UUID  `json:"user_id"`
	Role        Role       `json:"role"`
	DisplayName *string    `json:"display_name,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relationships (populated by service layer)
	Project *Project `json:"project,omitempty"`
	User    *User    `json:"user,omitempty"`
}

// CanViewTeamOnlyContent returns true if the member can view team-only content
func (m *Membership) CanViewTeamOnlyContent() bool {
	return m.Role.IsTeamRole()
}

// CanModifyFeedback returns true if the member can modify feedback
func (m *Membership) CanModifyFeedback() bool {
	return m.Role.CanModify()
}

// CanManageMembers returns true if the member can manage other members
func (m *Membership) CanManageMembers() bool {
	return m.Role.IsAdminOrOwner()
}

// CanManageSettings returns true if the member can manage project settings
func (m *Membership) CanManageSettings() bool {
	return m.Role.IsAdminOrOwner()
}

// CanDeleteProject returns true if the member can delete the project
func (m *Membership) CanDeleteProject() bool {
	return m.Role.IsOwner()
}

// HasRoleAtLeast checks if the member has at least the given role level
func (m *Membership) HasRoleAtLeast(required Role) bool {
	return m.Role.HasAtLeast(required)
}

// CanManageRole checks if the member can manage another member's role
func (m *Membership) CanManageRole(targetRole Role) bool {
	// Must be admin or owner to manage roles
	if !m.Role.IsAdminOrOwner() {
		return false
	}
	// Can only manage roles lower than your own
	return m.Role.Level() > targetRole.Level()
}

// MembershipWithUser includes user details for display
type MembershipWithUser struct {
	Membership
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}
