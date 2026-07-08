package http_errors

import (
	"errors"
	"log/slog"
	"net/http"

	domain_errors "github.com/te0tl/go-clean-arch-base/core/pkg/domain/errors"

	"github.com/gin-gonic/gin"
)

// ProjectStatusMapper lets a project map its own error sentinels (e.g.
// ErrRateLimited → 429) to an HTTP status for JSON APIs. It runs AFTER the
// canonical mapping fails. Return 0 to fall through to the default 500.
// Set it once at boot if needed; leave nil otherwise.
var ProjectStatusMapper func(err error) int

// StatusForError maps an error to an HTTP status using the canonical sentinels
// from core/pkg/domain/errors. This is the JSON-API counterpart of the htmx
// package's StatusForError, kept in sync with the same category → status table.
func StatusForError(err error) int {
	switch {
	case errors.Is(err, domain_errors.ErrInvalidInput),
		errors.Is(err, domain_errors.ErrValidation):
		return http.StatusUnprocessableEntity

	case errors.Is(err, domain_errors.ErrNotFound):
		return http.StatusNotFound

	case errors.Is(err, domain_errors.ErrUnauthorized):
		return http.StatusUnauthorized

	case errors.Is(err, domain_errors.ErrForbidden):
		return http.StatusForbidden

	case errors.Is(err, domain_errors.ErrAlreadyExists),
		errors.Is(err, domain_errors.ErrConflict):
		return http.StatusConflict

	case errors.Is(err, domain_errors.ErrExternalServiceUnavailable):
		return http.StatusBadGateway
	}

	if ProjectStatusMapper != nil {
		if status := ProjectStatusMapper(err); status != 0 {
			return status
		}
	}

	return http.StatusInternalServerError
}

func codeForStatus(status int) string {
	switch {
	case status == http.StatusUnauthorized:
		return ERROR_CODE_UNAUTHORIZED
	case status >= 400 && status < 500:
		return ERROR_CODE_INVALID_INPUT
	default:
		return ERROR_CODE_INTERNAL_SERVER_ERROR
	}
}

// Respond writes err as a JSON ErrorResponse with the mapped HTTP status and
// stashes it in the gin context so the logger middleware records it. Use in
// error paths of JSON handlers:
//
//	if err := ctrl.uc.Execute(...); err != nil {
//	    http_errors.Respond(c, err)
//	    return
//	}
//
// gin binding failures (validator.ValidationErrors) are reported as 400 with
// per-field details; everything else goes through the canonical mapping.
func Respond(c *gin.Context, err error) {
	if err != nil {
		c.Set(ERROR_MESSAGE_CONTEXT_KEY, err)
	}

	if fieldErrors := ExtractValidationErrors(err); len(fieldErrors) > 0 {
		status, body := NewValidationErrorResponse(fieldErrors)
		c.JSON(status, body)
		return
	}

	status := StatusForError(err)
	if status >= http.StatusInternalServerError {
		slog.Error("unmapped or internal error in Respond", "error", err)
	}

	c.JSON(status, ErrorResponse{
		Message: err.Error(),
		Code:    codeForStatus(status),
	})
}
