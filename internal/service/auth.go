package service

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"

	"github.com/danilov-go/gophprofile/internal/models"
)

// GetUserID проверяет валидность токена и возвращает UserID.
func (s *Service) GetUserID(tokenString, key string) (string, error) {
	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(key), nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", errors.New("невалидный токен")
	}
	userID := claims.UserID
	if userID == "" {
		return "", errors.New("в токене отсутствует user_id")
	}
	return userID, nil
}
