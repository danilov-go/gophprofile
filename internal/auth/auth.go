// Package auth реализует генерацию JWT-токена.
package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/danilov-go/gophprofile/internal/models"
)

// GenerateToken создает JWT-токен для указанного userID.
func GenerateToken(userID, secret string) (string, error) {
	now := time.Now()
	claims := models.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(models.TokenExp)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID: userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
