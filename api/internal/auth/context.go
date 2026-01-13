package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/fulldisclosure/api/internal/domain"
)

type contextKey string

const (
	userIDKey      contextKey = "user_id"
	userEmailKey   contextKey = "user_email"
	membershipKey  contextKey = "membership"
	sdkProjectKey  contextKey = "sdk_project_id"
	authMethodKey  contextKey = "auth_method"
)

// AuthMethod indicates how the request was authenticated
type AuthMethod string

const (
	AuthMethodSupabase AuthMethod = "supabase"
	AuthMethodSDK      AuthMethod = "sdk"
	AuthMethodNone     AuthMethod = "none"
)

// ContextWithUserID adds the user ID to the context
func ContextWithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext retrieves the user ID from context
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}

// MustUserIDFromContext retrieves the user ID from context or panics
func MustUserIDFromContext(ctx context.Context) uuid.UUID {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		panic("user_id not found in context")
	}
	return userID
}

// ContextWithUserEmail adds the user email to the context
func ContextWithUserEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, userEmailKey, email)
}

// UserEmailFromContext retrieves the user email from context
func UserEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(userEmailKey).(string)
	return email, ok
}

// MustUserEmailFromContext retrieves the user email from context or panics
func MustUserEmailFromContext(ctx context.Context) string {
	email, ok := UserEmailFromContext(ctx)
	if !ok {
		panic("user_email not found in context")
	}
	return email
}

// ContextWithMembership adds the membership to the context
func ContextWithMembership(ctx context.Context, membership *domain.Membership) context.Context {
	return context.WithValue(ctx, membershipKey, membership)
}

// MembershipFromContext retrieves the membership from context
func MembershipFromContext(ctx context.Context) (*domain.Membership, bool) {
	membership, ok := ctx.Value(membershipKey).(*domain.Membership)
	return membership, ok
}

// MustMembershipFromContext retrieves the membership from context or panics
func MustMembershipFromContext(ctx context.Context) *domain.Membership {
	membership, ok := MembershipFromContext(ctx)
	if !ok {
		panic("membership not found in context")
	}
	return membership
}

// ContextWithSDKProject adds the SDK project ID to the context
func ContextWithSDKProject(ctx context.Context, projectID uuid.UUID) context.Context {
	return context.WithValue(ctx, sdkProjectKey, projectID)
}

// SDKProjectFromContext retrieves the SDK project ID from context
func SDKProjectFromContext(ctx context.Context) (uuid.UUID, bool) {
	projectID, ok := ctx.Value(sdkProjectKey).(uuid.UUID)
	return projectID, ok
}

// ContextWithAuthMethod adds the auth method to the context
func ContextWithAuthMethod(ctx context.Context, method AuthMethod) context.Context {
	return context.WithValue(ctx, authMethodKey, method)
}

// AuthMethodFromContext retrieves the auth method from context
func AuthMethodFromContext(ctx context.Context) AuthMethod {
	method, ok := ctx.Value(authMethodKey).(AuthMethod)
	if !ok {
		return AuthMethodNone
	}
	return method
}

// IsAuthenticated checks if the request has valid authentication
func IsAuthenticated(ctx context.Context) bool {
	return AuthMethodFromContext(ctx) != AuthMethodNone
}
