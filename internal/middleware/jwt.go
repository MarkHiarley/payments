package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/markHiarley/payments/internal/auth"
	"github.com/markHiarley/payments/internal/platform/apperr"
	"github.com/markHiarley/payments/internal/platform/httpapi"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			httpapi.WriteProblem(c, apperr.Unauthorized(apperr.CodeUnauthorized, "token de autorização não fornecido"))
			return
		}

		tokenString, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			httpapi.WriteProblem(c, apperr.Unauthorized(apperr.CodeUnauthorized, "cabeçalho de autorização inválido"))
			return
		}

		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			httpapi.WriteProblem(c, apperr.Unauthorized(apperr.CodeUnauthorized, "token inválido ou expirado"))
			return
		}

		c.Set("email", claims.Email)
		c.Set("claims", claims)

		c.Next()
	}
}

func OptionalJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenString, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			httpapi.WriteProblem(c, apperr.Unauthorized(apperr.CodeUnauthorized, "cabeçalho de autorização inválido"))
			return
		}

		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			httpapi.WriteProblem(c, apperr.Unauthorized(apperr.CodeUnauthorized, "token inválido ou expirado"))
			return
		}

		c.Set("email", claims.Email)
		c.Set("claims", claims)
		c.Next()
	}
}
