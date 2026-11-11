package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/markHiarley/payments/internal/platform/apperr"
)

const (
	// TraceIDKey is the gin context key holding the request trace id.
	TraceIDKey = "trace_id"
	// TraceIDHeader propagates the trace id in both directions.
	TraceIDHeader = "X-Request-Id"

	problemBaseURL = "https://payments.local/errors/"

	contentTypeProblemJSON = "application/problem+json"
)

// Problem is the RFC 9457 representation returned for every failed request.
type Problem struct {
	Type     string            `json:"type"`
	Title    string            `json:"title"`
	Status   int               `json:"status"`
	Detail   string            `json:"detail,omitempty"`
	Instance string            `json:"instance"`
	Code     string            `json:"code"`
	TraceID  string            `json:"trace_id"`
	Meta     map[string]string `json:"meta,omitempty"`
}

// TraceID returns the trace id of the request, taking it from the incoming
// X-Request-Id header or generating one, and echoes it back on the response.
func TraceID(c *gin.Context) string {
	if value, ok := c.Get(TraceIDKey); ok {
		if id, ok := value.(string); ok {
			return id
		}
	}

	id := c.GetHeader(TraceIDHeader)
	if id == "" {
		id = uuid.NewString()
	}

	c.Set(TraceIDKey, id)
	c.Writer.Header().Set(TraceIDHeader, id)
	return id
}

// WriteProblem writes err as an RFC 9457 problem document. Internal and
// unavailable failures keep a generic detail; their real cause only reaches the
// log.
func WriteProblem(c *gin.Context, err error) {
	appErr := apperr.From(err)
	status := statusFor(appErr.Kind)
	traceID := TraceID(c)

	problem := Problem{
		Type:     problemBaseURL + slugFor(appErr.Kind),
		Title:    appErr.Title,
		Status:   status,
		Detail:   appErr.Detail,
		Instance: c.Request.URL.Path,
		Code:     appErr.Code,
		TraceID:  traceID,
		Meta:     appErr.Meta,
	}

	if appErr.Kind == apperr.KindInternal || appErr.Kind == apperr.KindUnavailable {
		problem.Detail = genericDetail(appErr.Kind)
		logFailure(c, appErr, traceID)
	}

	c.Header("Content-Type", contentTypeProblemJSON)
	c.AbortWithStatusJSON(status, problem)
}

// BindError converts a request binding failure into a validation error without
// exposing the raw decoder or validator output. Only a body that is not valid
// JSON reports MALFORMED_JSON; a well-formed body with the wrong types or a
// failed field rule reports VALIDATION_ERROR.
func BindError(err error) *apperr.Error {
	var syntaxErr *json.SyntaxError

	if errors.Is(err, io.EOF) || errors.As(err, &syntaxErr) {
		return apperr.Validation(apperr.CodeMalformedJSON, "corpo da requisição não é um JSON válido")
	}

	return apperr.Validation(apperr.CodeValidation, "dados da requisição inválidos")
}

func statusFor(kind apperr.Kind) int {
	switch kind {
	case apperr.KindValidation:
		return http.StatusBadRequest
	case apperr.KindUnauthorized:
		return http.StatusUnauthorized
	case apperr.KindForbidden:
		return http.StatusForbidden
	case apperr.KindNotFound:
		return http.StatusNotFound
	case apperr.KindConflict:
		return http.StatusConflict
	case apperr.KindPrecondition:
		return http.StatusUnprocessableEntity
	case apperr.KindUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func slugFor(kind apperr.Kind) string {
	switch kind {
	case apperr.KindValidation:
		return "validation"
	case apperr.KindUnauthorized:
		return "unauthorized"
	case apperr.KindForbidden:
		return "forbidden"
	case apperr.KindNotFound:
		return "not-found"
	case apperr.KindConflict:
		return "conflict"
	case apperr.KindPrecondition:
		return "precondition-failed"
	case apperr.KindUnavailable:
		return "service-unavailable"
	default:
		return "internal"
	}
}

func genericDetail(kind apperr.Kind) string {
	if kind == apperr.KindUnavailable {
		return "serviço temporariamente indisponível"
	}
	return "erro interno inesperado"
}

func logFailure(c *gin.Context, appErr *apperr.Error, traceID string) {
	attrs := []any{
		"trace_id", traceID,
		"code", appErr.Code,
		"kind", string(appErr.Kind),
		"path", c.Request.URL.Path,
	}
	if cause := appErr.Unwrap(); cause != nil {
		attrs = append(attrs, "cause", cause)
	}

	slog.Error("request failed", attrs...)
}
