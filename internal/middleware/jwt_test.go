package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/markHiarley/payments/internal/middleware"
	"github.com/markHiarley/payments/internal/platform/apperr"
	"github.com/markHiarley/payments/internal/platform/httpapi"
)

func TestJWTAuthAnswersWithProblemEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		header string
	}{
		{name: "sem token"},
		{name: "token inválido", header: "Bearer not-a-token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/private", middleware.JWTAuth(), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			request := httptest.NewRequest(http.MethodGet, "/private", nil)
			if tt.header != "" {
				request.Header.Set("Authorization", tt.header)
			}
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
			}

			var problem httpapi.Problem
			if err := json.Unmarshal(recorder.Body.Bytes(), &problem); err != nil {
				t.Fatalf("decode problem: %v (body %q)", err, recorder.Body.String())
			}
			if problem.Code != apperr.CodeUnauthorized {
				t.Errorf("problem.code = %q, want %q", problem.Code, apperr.CodeUnauthorized)
			}
			if problem.TraceID == "" {
				t.Error("problem.trace_id is empty")
			}
			if problem.Status != http.StatusUnauthorized {
				t.Errorf("problem.status = %d, want %d", problem.Status, http.StatusUnauthorized)
			}
		})
	}
}
