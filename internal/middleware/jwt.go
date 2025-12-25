package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markHiarley/payments/internal/auth"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token de autorização não fornecido",
			})
			c.Abort()
			return
		}

		tokenString, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}

		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token inválido ou expirado",
			})
			c.Abort()
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
			c.Next()
			return
		}

		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		c.Set("email", claims.Email)
		c.Set("claims", claims)
		c.Next()
	}
}
