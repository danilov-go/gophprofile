package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/danilov-go/gophprofile/internal/auth"
	"github.com/danilov-go/gophprofile/internal/handler"
	"github.com/danilov-go/gophprofile/internal/handler/mocks"
	"github.com/danilov-go/gophprofile/internal/models"
)

const testSecret = "test-secret-key"

func TestHandler_ValidateUserID(t *testing.T) {
	token, err := auth.GenerateToken(userID, testSecret)
	require.NoError(t, err)
	tests := []struct {
		name       string
		authHeader string
		urlID      string
		setupMock  func(svc *mocks.MockService)
		expCode    int
		expUserID  string
		expErr     models.ErrorResponse
	}{
		{
			name:       "положительный тест",
			authHeader: "Bearer " + token,
			urlID:      userID,
			setupMock: func(svc *mocks.MockService) {
				svc.On("GetUserID", token, testSecret).Return(userID, nil).Once()
			},
			expCode:   http.StatusOK,
			expUserID: userID,
		},
		{
			name:    "заголовок Authorization отсутствует",
			expCode: http.StatusUnauthorized,
			expErr: models.ErrorResponse{
				Error:   "Unauthorized",
				Details: "Authorization Bearer token is required",
			},
		},
		{
			name:       "неверный формат заголовка Authorization",
			authHeader: "Basic " + token,
			expCode:    http.StatusUnauthorized,
			expErr: models.ErrorResponse{
				Error:   "Unauthorized",
				Details: "Authorization Bearer token is required",
			},
		},
		{
			name:       "невалидный или просроченный токен",
			authHeader: "Bearer invalid-token",
			setupMock: func(svc *mocks.MockService) {
				svc.On("GetUserID", "invalid-token", testSecret).Return("", errors.New("invalid token")).Once()
			},
			expCode: http.StatusUnauthorized,
			expErr: models.ErrorResponse{
				Error:   "Unauthorized",
				Details: "Invalid or expired token",
			},
		},
		{
			name:       "user_id в URL не совпадает с пользователем из токена",
			authHeader: "Bearer " + token,
			urlID:      unknowID,
			setupMock: func(svc *mocks.MockService) {
				svc.On("GetUserID", token, testSecret).Return(userID, nil).Once()
			},
			expCode: http.StatusForbidden,
			expErr: models.ErrorResponse{
				Error:   "Forbidden",
				Details: "user_id does not match authenticated user",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()
			svc := mocks.NewMockService(t)
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}
			var capturedUserID string
			next := func(w http.ResponseWriter, r *http.Request, id string) {
				capturedUserID = id
				w.WriteHeader(http.StatusOK)
			}
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/avatars", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			if tt.urlID != "" {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("user_id", tt.urlID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			}
			rec := httptest.NewRecorder()
			handler.NewHandler(svc, logger, "").ValidateUserID(testSecret, next).ServeHTTP(rec, req)
			assert.Equal(t, tt.expCode, rec.Code)
			assert.Equal(t, tt.expUserID, capturedUserID)
			if tt.expErr.Error != "" {
				var errResp models.ErrorResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errResp))
				assert.Equal(t, tt.expErr, errResp)
			}
		})
	}
}
