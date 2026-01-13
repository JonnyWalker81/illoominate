package auth

import (
	"context"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/fulldisclosure/api/internal/domain"
)

// MembershipLoader loads membership for a user and project
type MembershipLoader interface {
	GetByProjectAndUser(ctx context.Context, projectID, userID string) (*domain.Membership, error)
}

// SupabaseAuthMiddleware creates middleware that validates Supabase JWT tokens
func SupabaseAuthMiddleware(validator *SupabaseValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := ExtractTokenFromHeader(r)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userID, claims, err := validator.ValidateToken(r.Context(), token)
			if err != nil {
				log.Warn().Err(err).Msg("JWT validation failed")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add user info to context
			ctx := ContextWithUserID(r.Context(), userID)
			ctx = ContextWithUserEmail(ctx, claims.Email)
			ctx = ContextWithAuthMethod(ctx, AuthMethodSupabase)

			// Log successful auth
			log.Debug().
				Str("user_id", userID.String()).
				Str("email", claims.Email).
				Msg("User authenticated via Supabase JWT")

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalSupabaseAuthMiddleware validates Supabase JWT if present but doesn't require it
func OptionalSupabaseAuthMiddleware(validator *SupabaseValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := ExtractTokenFromHeader(r)
			if err != nil {
				// No token, continue without auth
				next.ServeHTTP(w, r)
				return
			}

			userID, claims, err := validator.ValidateToken(r.Context(), token)
			if err != nil {
				// Invalid token, continue without auth (optional auth)
				next.ServeHTTP(w, r)
				return
			}

			// Add user info to context
			ctx := ContextWithUserID(r.Context(), userID)
			ctx = ContextWithUserEmail(ctx, claims.Email)
			ctx = ContextWithAuthMethod(ctx, AuthMethodSupabase)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SDKAuthMiddleware creates middleware that validates SDK tokens
func SDKAuthMiddleware(validator *SDKTokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := ExtractSDKTokenFromHeader(r)
			if err != nil {
				http.Error(w, "Unauthorized: missing SDK token", http.StatusUnauthorized)
				return
			}

			origin := r.Header.Get("Origin")
			projectID, err := validator.ValidateToken(r.Context(), token, origin)
			if err != nil {
				log.Warn().Err(err).Msg("SDK token validation failed")
				http.Error(w, "Unauthorized: invalid SDK token", http.StatusUnauthorized)
				return
			}

			// Add project info to context
			ctx := ContextWithSDKProject(r.Context(), projectID)
			ctx = ContextWithAuthMethod(ctx, AuthMethodSDK)

			log.Debug().
				Str("project_id", projectID.String()).
				Msg("Request authenticated via SDK token")

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireMembershipMiddleware ensures the user has a membership in the project
// The project ID must be extracted from the URL path parameter
func RequireMembershipMiddleware(loader MembershipLoader, projectIDExtractor func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			projectID := projectIDExtractor(r)
			if projectID == "" {
				http.Error(w, "Bad Request: missing project ID", http.StatusBadRequest)
				return
			}

			membership, err := loader.GetByProjectAndUser(r.Context(), projectID, userID.String())
			if err != nil {
				log.Warn().
					Err(err).
					Str("user_id", userID.String()).
					Str("project_id", projectID).
					Msg("Membership lookup failed")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if membership == nil {
				http.Error(w, "Forbidden: not a project member", http.StatusForbidden)
				return
			}

			// Add membership to context
			ctx := ContextWithMembership(r.Context(), membership)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireTeamRoleMiddleware ensures the user has a team role (not community)
func RequireTeamRoleMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			membership, ok := MembershipFromContext(r.Context())
			if !ok {
				http.Error(w, "Forbidden: no membership", http.StatusForbidden)
				return
			}

			if !membership.Role.IsTeamRole() {
				http.Error(w, "Forbidden: team role required", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRoleMiddleware ensures the user has at least the specified role
func RequireRoleMiddleware(minRole domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			membership, ok := MembershipFromContext(r.Context())
			if !ok {
				http.Error(w, "Forbidden: no membership", http.StatusForbidden)
				return
			}

			if !membership.HasRoleAtLeast(minRole) {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdminMiddleware ensures the user is an admin or owner
func RequireAdminMiddleware() func(http.Handler) http.Handler {
	return RequireRoleMiddleware(domain.RoleAdmin)
}

// RequireOwnerMiddleware ensures the user is the owner
func RequireOwnerMiddleware() func(http.Handler) http.Handler {
	return RequireRoleMiddleware(domain.RoleOwner)
}
