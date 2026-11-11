package apperr

import (
	"errors"
	"fmt"
)

// Error is a domain error. Kind drives the HTTP status, Code is the stable
// machine-readable identifier, Detail is safe to show to the caller and cause
// keeps the original error for server-side logs only.
type Error struct {
	Code   string
	Kind   Kind
	Title  string
	Detail string
	Meta   map[string]string

	cause error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s (%s): %s: %v", e.Code, e.Kind, e.Detail, e.cause)
	}
	return fmt.Sprintf("%s (%s): %s", e.Code, e.Kind, e.Detail)
}

func (e *Error) Unwrap() error {
	return e.cause
}

func newError(kind Kind, code, detail string) *Error {
	return &Error{
		Code:   code,
		Kind:   kind,
		Title:  title(kind),
		Detail: detail,
	}
}

func Validation(code, detail string) *Error {
	return newError(KindValidation, code, detail)
}

func Unauthorized(code, detail string) *Error {
	return newError(KindUnauthorized, code, detail)
}

func Forbidden(code, detail string) *Error {
	return newError(KindForbidden, code, detail)
}

func NotFound(code, detail string) *Error {
	return newError(KindNotFound, code, detail)
}

func Conflict(code, detail string) *Error {
	return newError(KindConflict, code, detail)
}

func Precondition(code, detail string) *Error {
	return newError(KindPrecondition, code, detail)
}

func Unavailable(code, detail string) *Error {
	return newError(KindUnavailable, code, detail)
}

// Internal wraps an unexpected error. Its detail is never sent to the client.
func Internal(cause error) *Error {
	return &Error{
		Code:   CodeInternal,
		Kind:   KindInternal,
		Title:  title(KindInternal),
		Detail: "erro interno inesperado",
		cause:  cause,
	}
}

// From returns err as *Error. Errors that were not produced by this package are
// reported as internal so no raw driver or SQL text can leak.
func From(err error) *Error {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return Internal(err)
}
