package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/danilov-go/gophprofile/internal/handler"
	"github.com/danilov-go/gophprofile/internal/handler/mocks"
	"github.com/danilov-go/gophprofile/internal/models"
)

func TestHandler_DeleteAvatar(t *testing.T) {
	tests := []struct {
		name       string
		avatarID   string
		setupMocks func(svc *mocks.MockService)
		expCode    int
		expErr     models.ErrorResponse
	}{
		{
			name:     "положительный тест",
			avatarID: avatarID,
			setupMocks: func(svc *mocks.MockService) {
				svc.On("Delete", mock.Anything, avatarID, userID).Return(nil).Once()
			},
			expCode: http.StatusNoContent,
		},
		{
			name:     "пустой параметр avatar_id",
			avatarID: "",
			expCode:  http.StatusBadRequest,
			expErr: models.ErrorResponse{
				Error:   "Empty avatarID",
				Details: "avatar_id path parameter is required",
			},
		},
		{
			name:     "аватарка не найдена",
			avatarID: avatarID,
			setupMocks: func(svc *mocks.MockService) {
				svc.On("Delete", mock.Anything, avatarID, userID).Return(models.ErrAvatarNotFound).Once()
			},
			expCode: http.StatusNotFound,
		},
		{
			name:     "чужая аватарка",
			avatarID: avatarID,
			setupMocks: func(svc *mocks.MockService) {
				svc.On("Delete", mock.Anything, avatarID, userID).Return(models.ErrForbidden).Once()
			},
			expCode: http.StatusForbidden,
		},
		{
			name:     "ошибка сервиса",
			avatarID: avatarID,
			setupMocks: func(svc *mocks.MockService) {
				svc.On("Delete", mock.Anything, avatarID, userID).Return(errTest).Once()
			},
			expCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()
			svc := mocks.NewMockService(t)
			if tt.setupMocks != nil {
				tt.setupMocks(svc)
			}
			req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/api/v1/avatars/"+tt.avatarID, nil)
			if tt.avatarID != "" {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("avatar_id", tt.avatarID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			}
			rec := httptest.NewRecorder()
			handler.NewHandler(svc, logger, "").DeleteAvatar(rec, req, userID)
			assert.Equal(t, tt.expCode, rec.Code)
			if tt.expCode == http.StatusNoContent {
				assert.Empty(t, rec.Body.Bytes())
			}
			if tt.expErr.Error != "" {
				var errResp models.ErrorResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errResp))
				assert.Equal(t, tt.expErr, errResp)
			}
		})
	}
}

func TestHandler_DeleteUserAvatar(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(svc *mocks.MockService)
		expCode    int
	}{
		{
			name: "положительный тест",
			setupMocks: func(svc *mocks.MockService) {
				svc.On("DeleteByUserID", mock.Anything, userID).Return(nil).Once()
			},
			expCode: http.StatusNoContent,
		},
		{
			name: "аватарка не найдена",
			setupMocks: func(svc *mocks.MockService) {
				svc.On("DeleteByUserID", mock.Anything, userID).Return(models.ErrAvatarNotFound).Once()
			},
			expCode: http.StatusNotFound,
		},
		{
			name: "ошибка сервиса",
			setupMocks: func(svc *mocks.MockService) {
				svc.On("DeleteByUserID", mock.Anything, userID).Return(errTest).Once()
			},
			expCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()
			svc := mocks.NewMockService(t)
			tt.setupMocks(svc)
			req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/api/v1/users/"+userID+"/avatar", nil)
			rec := httptest.NewRecorder()
			handler.NewHandler(svc, logger, "").DeleteUserAvatar(rec, req, userID)
			assert.Equal(t, tt.expCode, rec.Code)
			if tt.expCode == http.StatusNoContent {
				assert.Empty(t, rec.Body.Bytes())
			}
		})
	}
}
