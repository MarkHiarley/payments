package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/markHiarley/payments/internal/platform/httpapi"
)

// ErrorHandler renders the first error attached to the request context as an
// RFC 9457 problem document. Controllers only report the error; they never pick
// a status code.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		httpapi.TraceID(c)

		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		httpapi.WriteProblem(c, c.Errors[0].Err)
	}
}
