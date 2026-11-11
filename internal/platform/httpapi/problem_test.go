package httpapi_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/markHiarley/payments/internal/platform/apperr"
	"github.com/markHiarley/payments/internal/platform/httpapi"
)

func newContext(t *testing.T, path string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, path, nil)

	return ctx, recorder
}

func decodeProblem(t *testing.T, recorder *httptest.ResponseRecorder) httpapi.Problem {
	t.Helper()

	var problem httpapi.Problem
	if err := json.Unmarshal(recorder.Body.Bytes(), &problem); err != nil {
		t.Fatalf("decode problem: %v (body %q)", err, recorder.Body.String())
	}
	return problem
}

func TestWriteProblemStatusByKind(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"validation", apperr.Validation(apperr.CodeValidation, "campo inválido"), http.StatusBadRequest, apperr.CodeValidation},
		{"malformed json", apperr.Validation(apperr.CodeMalformedJSON, "json inválido"), http.StatusBadRequest, apperr.CodeMalformedJSON},
		{"unauthorized", apperr.Unauthorized(apperr.CodeUnauthorized, "credenciais inválidas"), http.StatusUnauthorized, apperr.CodeUnauthorized},
		{"forbidden", apperr.Forbidden(apperr.CodeForbidden, "sem permissão"), http.StatusForbidden, apperr.CodeForbidden},
		{"not found", apperr.NotFound(apperr.CodeAccountNotFound, "conta não encontrada"), http.StatusNotFound, apperr.CodeAccountNotFound},
		{"conflict", apperr.Conflict(apperr.CodeDuplicateExternalID, "duplicada"), http.StatusConflict, apperr.CodeDuplicateExternalID},
		{"precondition", apperr.Precondition(apperr.CodeInsufficientFunds, "saldo insuficiente"), http.StatusUnprocessableEntity, apperr.CodeInsufficientFunds},
		{"unavailable", apperr.Unavailable(apperr.CodeServiceUnavailable, "banco fora"), http.StatusServiceUnavailable, apperr.CodeServiceUnavailable},
		{"internal", apperr.Internal(errors.New("boom")), http.StatusInternalServerError, apperr.CodeInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, recorder := newContext(t, "/api/v1/transactions")

			httpapi.WriteProblem(ctx, tt.err)

			if recorder.Code != tt.status {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.status)
			}

			problem := decodeProblem(t, recorder)
			if problem.Status != tt.status {
				t.Errorf("problem.status = %d, want %d", problem.Status, tt.status)
			}
			if problem.Code != tt.code {
				t.Errorf("problem.code = %q, want %q", problem.Code, tt.code)
			}
			if problem.Type == "" || !strings.HasPrefix(problem.Type, "https://payments.local/errors/") {
				t.Errorf("problem.type = %q, want a payments.local error type", problem.Type)
			}
			if problem.Title == "" {
				t.Error("problem.title is empty")
			}
			if problem.Instance != "/api/v1/transactions" {
				t.Errorf("problem.instance = %q, want the request path", problem.Instance)
			}
			if problem.TraceID == "" {
				t.Error("problem.trace_id is empty")
			}
			if got := recorder.Header().Get("Content-Type"); got != "application/problem+json" {
				t.Errorf("Content-Type = %q, want application/problem+json", got)
			}
			if recorder.Header().Get(httpapi.TraceIDHeader) != problem.TraceID {
				t.Errorf("response %s = %q, want %q", httpapi.TraceIDHeader, recorder.Header().Get(httpapi.TraceIDHeader), problem.TraceID)
			}
		})
	}
}

func TestWriteProblemInternalDoesNotLeakCause(t *testing.T) {
	cause := errors.New(`failed to connect: host=db user=payments password=s3cr3t dbname=payments`)
	ctx, recorder := newContext(t, "/api/v1/accounts")

	httpapi.WriteProblem(ctx, cause)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}

	body := recorder.Body.String()
	if strings.Contains(body, "s3cr3t") || strings.Contains(body, "failed to connect") {
		t.Fatalf("response leaked the internal cause: %s", body)
	}

	problem := decodeProblem(t, recorder)
	if problem.Detail == "" {
		t.Error("problem.detail is empty")
	}
	if problem.TraceID == "" {
		t.Error("problem.trace_id is empty")
	}
}

func TestWriteProblemUnavailableHidesDetail(t *testing.T) {
	ctx, recorder := newContext(t, "/api/v1/accounts")
	err := apperr.Unavailable(apperr.CodeServiceUnavailable, "host=db user=payments password=s3cr3t")

	httpapi.WriteProblem(ctx, err)

	body := recorder.Body.String()
	if strings.Contains(body, "s3cr3t") {
		t.Fatalf("response leaked the unavailable detail: %s", body)
	}

	problem := decodeProblem(t, recorder)
	if problem.Status != http.StatusServiceUnavailable {
		t.Errorf("problem.status = %d, want %d", problem.Status, http.StatusServiceUnavailable)
	}
	if problem.Code != apperr.CodeServiceUnavailable {
		t.Errorf("problem.code = %q, want %q", problem.Code, apperr.CodeServiceUnavailable)
	}
	if problem.Detail != "serviço temporariamente indisponível" {
		t.Errorf("problem.detail = %q, want the generic message", problem.Detail)
	}
}

func TestWriteProblemPropagatesRequestID(t *testing.T) {
	ctx, recorder := newContext(t, "/api/v1/accounts")
	ctx.Request.Header.Set(httpapi.TraceIDHeader, "req-123")

	httpapi.WriteProblem(ctx, apperr.NotFound(apperr.CodeAccountNotFound, "conta não encontrada"))

	problem := decodeProblem(t, recorder)
	if problem.TraceID != "req-123" {
		t.Errorf("problem.trace_id = %q, want %q", problem.TraceID, "req-123")
	}
	if got := recorder.Header().Get(httpapi.TraceIDHeader); got != "req-123" {
		t.Errorf("response %s = %q, want %q", httpapi.TraceIDHeader, got, "req-123")
	}
}

func TestWriteProblemIncludesMeta(t *testing.T) {
	ctx, recorder := newContext(t, "/api/v1/accounts")
	conflict := apperr.Conflict(apperr.CodeAlreadyExists, "registro já cadastrado")
	conflict.Meta = map[string]string{"field": "email"}

	httpapi.WriteProblem(ctx, conflict)

	problem := decodeProblem(t, recorder)
	if got := problem.Meta["field"]; got != "email" {
		t.Errorf("problem.meta.field = %q, want %q", got, "email")
	}
}

func TestBindError(t *testing.T) {
	var typeErr *json.UnmarshalTypeError
	if err := json.Unmarshal([]byte(`{"amount":"dez"}`), &struct {
		Amount int64 `json:"amount"`
	}{}); !errors.As(err, &typeErr) {
		t.Fatalf("expected an UnmarshalTypeError, got %T", err)
	}

	tests := []struct {
		name string
		in   error
		code string
	}{
		{name: "empty body", in: io.EOF, code: apperr.CodeMalformedJSON},
		{name: "invalid json", in: &json.SyntaxError{Offset: 4}, code: apperr.CodeMalformedJSON},
		{name: "wrong type", in: typeErr, code: apperr.CodeValidation},
		{name: "validator failure", in: fmt.Errorf("Key: 'Account.Password' Error:Field validation for 'Password' failed"), code: apperr.CodeValidation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := httpapi.BindError(tt.in)

			if appErr.Kind != apperr.KindValidation {
				t.Errorf("Kind = %q, want %q", appErr.Kind, apperr.KindValidation)
			}
			if appErr.Code != tt.code {
				t.Errorf("Code = %q, want %q", appErr.Code, tt.code)
			}
			if strings.Contains(appErr.Detail, "Password") {
				t.Errorf("Detail = %q, want a message without binding internals", appErr.Detail)
			}
		})
	}
}
