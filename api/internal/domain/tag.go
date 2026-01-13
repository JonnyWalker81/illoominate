package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Tag represents a label that can be applied to feedback
type Tag struct {
	ID          uuid.UUID `json:"id"`
	ProjectID   uuid.UUID `json:"project_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Color       string    `json:"color"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`

	// Computed fields (populated by service layer)
	FeedbackCount int `json:"feedback_count,omitempty"`
}

// NewTag creates a new tag with the given parameters
func NewTag(projectID uuid.UUID, name string, color *string) *Tag {
	tag := &Tag{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      name,
		Slug:      GenerateSlug(name),
		Color:     "#6B7280", // Default gray color
	}
	if color != nil {
		tag.Color = *color
	}
	return tag
}

// Validate validates the tag data
func (t *Tag) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(t.Name) > 50 {
		return fmt.Errorf("name must be 50 characters or less")
	}
	if t.Slug == "" {
		t.Slug = GenerateSlug(t.Name)
	}
	if !isValidTagSlug(t.Slug) {
		return fmt.Errorf("invalid slug format")
	}
	if t.Color != "" && !isValidHexColor(t.Color) {
		return fmt.Errorf("invalid color format: must be a hex color (e.g., #6B7280)")
	}
	return nil
}

var tagSlugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var hexColorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func isValidTagSlug(s string) bool {
	return tagSlugRegex.MatchString(s)
}

func isValidHexColor(s string) bool {
	return hexColorRegex.MatchString(s)
}

// GenerateSlug generates a URL-safe slug from a string
func GenerateSlug(s string) string {
	// Convert to lowercase
	slug := strings.ToLower(s)
	// Replace spaces and underscores with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	// Remove any character that isn't alphanumeric or hyphen
	slug = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(slug, "")
	// Replace multiple hyphens with single hyphen
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	// Trim hyphens from ends
	slug = strings.Trim(slug, "-")
	return slug
}
