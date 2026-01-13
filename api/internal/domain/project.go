package domain

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// Project represents a project/tenant in the system
type Project struct {
	ID         uuid.UUID       `json:"id"`
	Name       string          `json:"name"`
	Slug       string          `json:"slug"`
	ProjectKey string          `json:"project_key"`
	Settings   ProjectSettings `json:"settings"`
	LogoURL    *string         `json:"logo_url,omitempty"`
	PrimaryColor string        `json:"primary_color"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	ArchivedAt *time.Time      `json:"archived_at,omitempty"`
}

// ProjectSettings holds configurable settings for a project
type ProjectSettings struct {
	DefaultVisibility       DefaultVisibilitySettings `json:"default_visibility"`
	AllowAnonymousFeedback  bool                      `json:"allow_anonymous_feedback"`
	RequireEmailForAnonymous bool                     `json:"require_email_for_anonymous"`
	VotingEnabled           bool                      `json:"voting_enabled"`
	CommunityCommentsEnabled bool                     `json:"community_comments_enabled"`
	AutoCloseDuplicates     bool                      `json:"auto_close_duplicates"`
	NotificationPreferences NotificationPreferences   `json:"notification_preferences"`
}

// DefaultVisibilitySettings defines default visibility per feedback type
type DefaultVisibilitySettings struct {
	Bug     Visibility `json:"bug"`
	Feature Visibility `json:"feature"`
	General Visibility `json:"general"`
}

// NotificationPreferences defines notification settings
type NotificationPreferences struct {
	NewFeedback   bool `json:"new_feedback"`
	StatusChanges bool `json:"status_changes"`
	NewComments   bool `json:"new_comments"`
}

// DefaultProjectSettings returns the default settings for a new project
func DefaultProjectSettings() ProjectSettings {
	return ProjectSettings{
		DefaultVisibility: DefaultVisibilitySettings{
			Bug:     VisibilityTeamOnly,
			Feature: VisibilityCommunity,
			General: VisibilityTeamOnly,
		},
		AllowAnonymousFeedback:   true,
		RequireEmailForAnonymous: false,
		VotingEnabled:            true,
		CommunityCommentsEnabled: true,
		AutoCloseDuplicates:      true,
		NotificationPreferences: NotificationPreferences{
			NewFeedback:   true,
			StatusChanges: true,
			NewComments:   true,
		},
	}
}

// GetDefaultVisibility returns the default visibility for a feedback type
func (s *ProjectSettings) GetDefaultVisibility(feedbackType FeedbackType) Visibility {
	switch feedbackType {
	case FeedbackTypeBug:
		return s.DefaultVisibility.Bug
	case FeedbackTypeFeature:
		return s.DefaultVisibility.Feature
	case FeedbackTypeGeneral:
		return s.DefaultVisibility.General
	default:
		return VisibilityTeamOnly
	}
}

// Scan implements sql.Scanner for ProjectSettings
func (s *ProjectSettings) Scan(src interface{}) error {
	if src == nil {
		*s = DefaultProjectSettings()
		return nil
	}
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return fmt.Errorf("cannot scan %T into ProjectSettings", src)
	}
}

// IsArchived returns true if the project is archived
func (p *Project) IsArchived() bool {
	return p.ArchivedAt != nil
}

// Validate validates the project data
func (p *Project) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(p.Name) > 100 {
		return fmt.Errorf("name must be 100 characters or less")
	}
	if p.Slug == "" {
		return fmt.Errorf("slug is required")
	}
	if !isValidSlug(p.Slug) {
		return fmt.Errorf("invalid slug format: must be lowercase alphanumeric with hyphens")
	}
	if len(p.Slug) > 50 {
		return fmt.Errorf("slug must be 50 characters or less")
	}
	return nil
}

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func isValidSlug(s string) bool {
	return slugRegex.MatchString(s)
}
