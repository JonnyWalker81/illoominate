package domain

import "fmt"

// Role represents a user's role within a project
type Role string

const (
	RoleCommunity Role = "community"
	RoleViewer    Role = "viewer"
	RoleMember    Role = "member"
	RoleAdmin     Role = "admin"
	RoleOwner     Role = "owner"
)

// RoleHierarchy defines the permission level for each role
var RoleHierarchy = map[Role]int{
	RoleCommunity: 0,
	RoleViewer:    1,
	RoleMember:    2,
	RoleAdmin:     3,
	RoleOwner:     4,
}

// IsValid checks if the role is a valid role
func (r Role) IsValid() bool {
	_, ok := RoleHierarchy[r]
	return ok
}

// Level returns the permission level of the role
func (r Role) Level() int {
	return RoleHierarchy[r]
}

// IsTeamRole returns true if the role is a team role (not community)
func (r Role) IsTeamRole() bool {
	return r == RoleViewer || r == RoleMember || r == RoleAdmin || r == RoleOwner
}

// CanModify returns true if the role can modify content
func (r Role) CanModify() bool {
	return r == RoleMember || r == RoleAdmin || r == RoleOwner
}

// IsAdminOrOwner returns true if the role is admin or owner
func (r Role) IsAdminOrOwner() bool {
	return r == RoleAdmin || r == RoleOwner
}

// IsOwner returns true if the role is owner
func (r Role) IsOwner() bool {
	return r == RoleOwner
}

// HasAtLeast checks if the role has at least the given permission level
func (r Role) HasAtLeast(required Role) bool {
	return r.Level() >= required.Level()
}

// String returns the string representation of the role
func (r Role) String() string {
	return string(r)
}

// ParseRole parses a string into a Role
func ParseRole(s string) (Role, error) {
	r := Role(s)
	if !r.IsValid() {
		return "", fmt.Errorf("invalid role: %s", s)
	}
	return r, nil
}

// AllRoles returns all valid roles
func AllRoles() []Role {
	return []Role{RoleCommunity, RoleViewer, RoleMember, RoleAdmin, RoleOwner}
}

// TeamRoles returns all team roles
func TeamRoles() []Role {
	return []Role{RoleViewer, RoleMember, RoleAdmin, RoleOwner}
}
