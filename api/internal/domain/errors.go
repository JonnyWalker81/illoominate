package domain

import (
	"errors"
	"fmt"
	"net/http"
)

// Domain errors that can be mapped to HTTP status codes
var (
	// Authentication/Authorization errors
	ErrUnauthorized     = NewDomainError("unauthorized", "authentication required", http.StatusUnauthorized)
	ErrForbidden        = NewDomainError("forbidden", "access denied", http.StatusForbidden)
	ErrInvalidToken     = NewDomainError("invalid_token", "invalid or expired token", http.StatusUnauthorized)
	ErrTokenExpired     = NewDomainError("token_expired", "token has expired", http.StatusUnauthorized)

	// Not found errors
	ErrNotFound         = NewDomainError("not_found", "resource not found", http.StatusNotFound)
	ErrProjectNotFound  = NewDomainError("project_not_found", "project not found", http.StatusNotFound)
	ErrFeedbackNotFound = NewDomainError("feedback_not_found", "feedback not found", http.StatusNotFound)
	ErrCommentNotFound  = NewDomainError("comment_not_found", "comment not found", http.StatusNotFound)
	ErrTagNotFound      = NewDomainError("tag_not_found", "tag not found", http.StatusNotFound)
	ErrInviteNotFound   = NewDomainError("invite_not_found", "invite not found", http.StatusNotFound)
	ErrMemberNotFound   = NewDomainError("member_not_found", "member not found", http.StatusNotFound)
	ErrAttachmentNotFound = NewDomainError("attachment_not_found", "attachment not found", http.StatusNotFound)

	// Conflict errors
	ErrConflict            = NewDomainError("conflict", "resource already exists", http.StatusConflict)
	ErrAlreadyMember       = NewDomainError("already_member", "user is already a member of this project", http.StatusConflict)
	ErrAlreadyVoted        = NewDomainError("already_voted", "user has already voted", http.StatusConflict)
	ErrSlugTaken           = NewDomainError("slug_taken", "slug is already in use", http.StatusConflict)
	ErrPendingInviteExists = NewDomainError("pending_invite_exists", "a pending invite already exists for this email", http.StatusConflict)

	// Validation errors
	ErrValidation       = NewDomainError("validation_error", "validation failed", http.StatusBadRequest)
	ErrInvalidInput     = NewDomainError("invalid_input", "invalid input data", http.StatusBadRequest)
	ErrMissingField     = NewDomainError("missing_field", "required field is missing", http.StatusBadRequest)

	// Business logic errors
	ErrInviteExpired    = NewDomainError("invite_expired", "invite has expired", http.StatusGone)
	ErrInviteRevoked    = NewDomainError("invite_revoked", "invite has been revoked", http.StatusGone)
	ErrCannotMergeSelf  = NewDomainError("cannot_merge_self", "cannot merge feedback into itself", http.StatusBadRequest)
	ErrCannotRemoveOwner = NewDomainError("cannot_remove_owner", "cannot remove the project owner", http.StatusForbidden)
	ErrCannotChangeOwnerRole = NewDomainError("cannot_change_owner_role", "cannot change the owner's role", http.StatusForbidden)
	ErrNoMembership     = NewDomainError("no_membership", "user is not a member of this project", http.StatusForbidden)

	// Rate limiting
	ErrRateLimited      = NewDomainError("rate_limited", "too many requests", http.StatusTooManyRequests)

	// Server errors
	ErrInternal         = NewDomainError("internal_error", "an internal error occurred", http.StatusInternalServerError)
)

// DomainError represents an application-level error with a code and HTTP status
type DomainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
	Err     error  `json:"-"`
}

// NewDomainError creates a new domain error
func NewDomainError(code, message string, status int) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *DomainError) Unwrap() error {
	return e.Err
}

// WithError wraps an underlying error
func (e *DomainError) WithError(err error) *DomainError {
	return &DomainError{
		Code:    e.Code,
		Message: e.Message,
		Status:  e.Status,
		Err:     err,
	}
}

// WithMessage creates a copy with a custom message
func (e *DomainError) WithMessage(msg string) *DomainError {
	return &DomainError{
		Code:    e.Code,
		Message: msg,
		Status:  e.Status,
		Err:     e.Err,
	}
}

// WithMessagef creates a copy with a formatted custom message
func (e *DomainError) WithMessagef(format string, args ...interface{}) *DomainError {
	return &DomainError{
		Code:    e.Code,
		Message: fmt.Sprintf(format, args...),
		Status:  e.Status,
		Err:     e.Err,
	}
}

// Is checks if the target error matches this error's code
func (e *DomainError) Is(target error) bool {
	var domainErr *DomainError
	if errors.As(target, &domainErr) {
		return e.Code == domainErr.Code
	}
	return false
}

// HTTPStatus returns the HTTP status code for the error
func (e *DomainError) HTTPStatus() int {
	return e.Status
}

// ValidationError represents a validation error with field-specific details
type ValidationError struct {
	*DomainError
	Fields map[string]string `json:"fields,omitempty"`
}

// NewValidationError creates a new validation error with field details
func NewValidationError(fields map[string]string) *ValidationError {
	return &ValidationError{
		DomainError: ErrValidation,
		Fields:      fields,
	}
}

// AddField adds a field error
func (e *ValidationError) AddField(field, message string) {
	if e.Fields == nil {
		e.Fields = make(map[string]string)
	}
	e.Fields[field] = message
}

// HasErrors returns true if there are any field errors
func (e *ValidationError) HasErrors() bool {
	return len(e.Fields) > 0
}

// ErrorResponse is the standard error response format for the API
type ErrorResponse struct {
	Error  string            `json:"error"`
	Code   string            `json:"code"`
	Fields map[string]string `json:"fields,omitempty"`
}

// ToErrorResponse converts a domain error to an error response
func ToErrorResponse(err error) ErrorResponse {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		resp := ErrorResponse{
			Error: domainErr.Message,
			Code:  domainErr.Code,
		}
		var validationErr *ValidationError
		if errors.As(err, &validationErr) {
			resp.Fields = validationErr.Fields
		}
		return resp
	}
	return ErrorResponse{
		Error: "an unexpected error occurred",
		Code:  "internal_error",
	}
}

// GetHTTPStatus returns the HTTP status for an error
func GetHTTPStatus(err error) int {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Status
	}
	return http.StatusInternalServerError
}
