package auth

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/markHiarley/payments/internal/models"
)

var JWT_SECRET_TOKEN []byte

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Could not find or load .env file. Using system environment variables.")
	}
	JWT_SECRET_TOKEN = []byte(os.Getenv("JWT_SECRET_TOKEN"))
}

func GenerateAccessToken(email string) (string, error) {
	claims := models.JWTClaims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "web-socket",
			Subject:   email,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenToString, err := accessToken.SignedString(JWT_SECRET_TOKEN)

	if err != nil {
		return "", fmt.Errorf("erro ao assinar token: %v", err)
	}

	return tokenToString, nil

}

func GenerateRefreshToken(email string) (string, error) {
	claims := models.JWTClaims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "web-socket",
			Subject:   email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenToString, err := token.SignedString([]byte(JWT_SECRET_TOKEN))
	if err != nil {
		return "", fmt.Errorf("erro ao assinar token: %v", err)
	}

	return tokenToString, nil
}

func ValidateToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inválido: %v", token.Header["alg"])
		}
		return []byte(JWT_SECRET_TOKEN), nil
	})

	if err != nil {
		return nil, fmt.Errorf("erro ao validar token: %v", err)
	}

	if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("token inválido")
}
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("token de autorização não fornecido")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) > len(bearerPrefix) && authHeader[:len(bearerPrefix)] == bearerPrefix {
		return authHeader[len(bearerPrefix):], nil
	}

	return "", fmt.Errorf("formato de autorização inválido")
}
