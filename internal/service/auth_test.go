package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danilov-go/gophprofile/internal/auth"
	"github.com/danilov-go/gophprofile/internal/service"
)

const (
	secretKey = "test-secret-key"
)

func TestService_GetUserID(t *testing.T) {
	svc := &service.Service{}
	generateToken := func(id, secret string) string {
		t.Helper()
		token, err := auth.GenerateToken(id, secret)
		require.NoError(t, err)
		return token
	}
	tests := []struct {
		name     string
		expToken string
		expID    string
		expErr   string
	}{
		{
			name:     "положительный тест",
			expToken: generateToken(userID, secretKey),
			expID:    userID,
		},
		{
			name:     "пустой UserID",
			expToken: generateToken("", secretKey),
			expErr:   "в токене отсутствует user_id",
		},
		{
			name:     "неверный секретный ключ",
			expToken: generateToken(userID, "unknow"),
			expErr:   "signature is invalid",
		},
		{
			name:     "невалидный токен",
			expToken: "invalid",
			expErr:   "token is malformed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := svc.GetUserID(tt.expToken, secretKey)
			if tt.expErr != "" {
				require.ErrorContains(t, err, tt.expErr)
				assert.Empty(t, id)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expID, id)
		})
	}
}
