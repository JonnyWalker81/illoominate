package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/fulldisclosure/api/internal/domain"
)

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error  string            `json:"error"`
	Code   string            `json:"code"`
	Fields map[string]string `json:"fields,omitempty"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data interface{}    `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// JSON writes a JSON response
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Error().Err(err).Msg("Failed to encode JSON response")
		}
	}
}

// Error writes an error response
func Error(w http.ResponseWriter, status int, code string, message string) {
	JSON(w, status, ErrorResponse{
		Error: message,
		Code:  code,
	})
}

// ValidationError writes a validation error response
func ValidationError(w http.ResponseWriter, fields map[string]string) {
	JSON(w, http.StatusBadRequest, ErrorResponse{
		Error:  "Validation failed",
		Code:   "VALIDATION_ERROR",
		Fields: fields,
	})
}

// HandleError writes an appropriate error response based on the error type
func HandleError(w http.ResponseWriter, err error) {
	var domainErr *domain.DomainError
	if errors.As(err, &domainErr) {
		JSON(w, domainErr.HTTPStatus(), ErrorResponse{
			Error: domainErr.Message,
			Code:  domainErr.Code,
		})
		return
	}

	if errors.Is(err, domain.ErrNotFound) {
		Error(w, http.StatusNotFound, "NOT_FOUND", "Resource not found")
		return
	}

	if errors.Is(err, domain.ErrForbidden) {
		Error(w, http.StatusForbidden, "FORBIDDEN", "Access denied")
		return
	}

	if errors.Is(err, domain.ErrUnauthorized) {
		Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	// Log unexpected errors
	log.Error().Err(err).Msg("Unexpected error")
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
}

// Paginated writes a paginated response
func Paginated(w http.ResponseWriter, data interface{}, total, page, perPage, totalPages int) {
	JSON(w, http.StatusOK, PaginatedResponse{
		Data: data,
		Meta: PaginationMeta{
			Total:      total,
			Page:       page,
			PerPage:    perPage,
			TotalPages: totalPages,
		},
	})
}

// Created writes a 201 Created response
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// NoContent writes a 204 No Content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// DecodeJSON decodes JSON from request body
func DecodeJSON(r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return domain.NewDomainError("INVALID_JSON", "Invalid JSON in request body", 400)
	}
	return nil
}
