package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fulldisclosure/api/internal/auth"
)

// SDKTokenHandler handles SDK token management endpoints
type SDKTokenHandler struct {
	db *pgxpool.Pool
}

// NewSDKTokenHandler creates a new SDK token handler
func NewSDKTokenHandler(db *pgxpool.Pool) *SDKTokenHandler {
	return &SDKTokenHandler{db: db}
}

// SDKTokenResponse represents an SDK token in API responses
type SDKTokenResponse struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	Token          string     `json:"token,omitempty"` // Only returned on create
	TokenPrefix    string     `json:"token_prefix"`
	AllowedOrigins []string   `json:"allowed_origins"`
	RateLimit      int        `json:"rate_limit"`
	IsActive       bool       `json:"is_active"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// List handles GET /creator/projects/{projectId}/sdk-tokens
func (h *SDKTokenHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "Invalid project ID")
		return
	}

	query := `
		SELECT id, name, token_prefix, allowed_origins, rate_limit_per_minute,
		       is_active, last_used_at, created_at
		FROM sdk_tokens
		WHERE project_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := h.db.Query(r.Context(), query, projectID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch SDK tokens")
		return
	}
	defer rows.Close()

	var tokens []SDKTokenResponse
	for rows.Next() {
		var token SDKTokenResponse
		var allowedOrigins []string
		var lastUsedAt *time.Time

		err := rows.Scan(
			&token.ID,
			&token.Name,
			&token.TokenPrefix,
			&allowedOrigins,
			&token.RateLimit,
			&token.IsActive,
			&lastUsedAt,
			&token.CreatedAt,
		)
		if err != nil {
			Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to scan SDK token")
			return
		}

		token.AllowedOrigins = allowedOrigins
		if allowedOrigins == nil {
			token.AllowedOrigins = []string{}
		}
		token.LastUsedAt = lastUsedAt
		tokens = append(tokens, token)
	}

	if tokens == nil {
		tokens = []SDKTokenResponse{}
	}

	JSON(w, http.StatusOK, tokens)
}

// CreateSDKTokenRequest represents the request body for creating an SDK token
type CreateSDKTokenRequest struct {
	Name               string   `json:"name"`
	AllowedOrigins     []string `json:"allowed_origins"`
	RateLimitPerMinute int      `json:"rate_limit_per_minute"`
}

// Create handles POST /creator/projects/{projectId}/sdk-tokens
func (h *SDKTokenHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "Invalid project ID")
		return
	}

	var req CreateSDKTokenRequest
	if err := DecodeJSON(r, &req); err != nil {
		HandleError(w, err)
		return
	}

	// Validation
	errors := make(map[string]string)
	if req.Name == "" {
		errors["name"] = "Name is required"
	}
	if len(req.Name) > 100 {
		errors["name"] = "Name must be 100 characters or less"
	}
	if len(errors) > 0 {
		ValidationError(w, errors)
		return
	}

	// Set defaults
	if req.RateLimitPerMinute <= 0 {
		req.RateLimitPerMinute = 60
	}
	if req.AllowedOrigins == nil {
		req.AllowedOrigins = []string{}
	}

	// Generate token
	token, tokenHash, err := auth.GenerateToken()
	if err != nil {
		Error(w, http.StatusInternalServerError, "TOKEN_GENERATION_ERROR", "Failed to generate token")
		return
	}

	// Token prefix for display (first 12 chars)
	tokenPrefix := token[:12]

	// Insert into database
	var tokenID uuid.UUID
	var createdAt time.Time

	insertQuery := `
		INSERT INTO sdk_tokens (project_id, name, token_hash, token_prefix, allowed_origins, rate_limit_per_minute, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, true)
		RETURNING id, created_at
	`

	err = h.db.QueryRow(
		r.Context(),
		insertQuery,
		projectID,
		req.Name,
		tokenHash,
		tokenPrefix,
		req.AllowedOrigins,
		req.RateLimitPerMinute,
	).Scan(&tokenID, &createdAt)

	if err != nil {
		Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create SDK token")
		return
	}

	// Return response with the full token (only time it's shown!)
	response := SDKTokenResponse{
		ID:             tokenID,
		Name:           req.Name,
		Token:          token, // Only returned on create!
		TokenPrefix:    tokenPrefix,
		AllowedOrigins: req.AllowedOrigins,
		RateLimit:      req.RateLimitPerMinute,
		IsActive:       true,
		CreatedAt:      createdAt,
	}

	Created(w, response)
}

// Revoke handles DELETE /creator/projects/{projectId}/sdk-tokens/{tokenId}
func (h *SDKTokenHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "Invalid project ID")
		return
	}

	tokenID, err := uuid.Parse(chi.URLParam(r, "tokenId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_TOKEN_ID", "Invalid token ID")
		return
	}

	// Soft delete by setting revoked_at and is_active
	query := `
		UPDATE sdk_tokens
		SET revoked_at = NOW(), is_active = false
		WHERE id = $1 AND project_id = $2 AND revoked_at IS NULL
	`

	result, err := h.db.Exec(r.Context(), query, tokenID, projectID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to revoke SDK token")
		return
	}

	if result.RowsAffected() == 0 {
		Error(w, http.StatusNotFound, "NOT_FOUND", "SDK token not found")
		return
	}

	NoContent(w)
}
