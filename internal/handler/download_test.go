package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"

	"github.com/danilov-go/gophprofile/internal/handler"
	"github.com/danilov-go/gophprofile/internal/handler/mocks"
	"github.com/danilov-go/gophprofile/internal/models"
)

func TestHandler_GetUserAvatar(t *testing.T) {
	meta := models.Avatar{
		ID:       avatarID,
		UserID:   userID,
		MimeType: "image/webp",
	}
	tests := []struct {
		name       string
		userID     string
		query      string
		setupMocks func(svc *mocks.MockService)
		expCode    int
		expBody    []byte
		expMime    string
	}{
		{
			name:   "положительный тест",
			userID: userID,
			query:  "?size=medium&format=webp",
			setupMocks: func(svc *mocks.MockService) {
				svc.On("GetAvatarByUserID", mock.Anything, userID, "medium", "webp").
					Return(meta, imageData, nil).Once()
			},
			expCode: http.StatusOK,
			expBody: imageData,
			expMime: "image/webp",
		},
		{
			name:    "пустой параметр user_id",
			userID:  "",
			expCode: http.StatusBadRequest,
		},
		{
			name:   "аватарка не найдена",
			userID: userID,
			setupMocks: func(svc *mocks.MockService) {
				svc.On("GetAvatarByUserID", mock.Anything, userID, "", "").
					Return(models.Avatar{}, nil, models.ErrAvatarNotFound).Once()
			},
			expCode: http.StatusNotFound,
		},
		{
			name:   "невалидный размер или формат",
			userID: userID,
			query:  "?size=invalid&format=gif",
			setupMocks: func(svc *mocks.MockService) {
				svc.On("GetAvatarByUserID", mock.Anything, userID, "invalid", "gif").
					Return(models.Avatar{}, nil, models.ErrInvalidSize).Once()
			},
			expCode: http.StatusBadRequest,
		},
		{
			name:   "ошибка сервиса",
			userID: userID,
			setupMocks: func(svc *mocks.MockService) {
				svc.On("GetAvatarByUserID", mock.Anything, userID, "", "").
					Return(models.Avatar{}, nil, errTest).Once()
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
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/users/"+tt.userID+"/avatar"+tt.query, nil)
			if tt.userID != "" {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("user_id", tt.userID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			}
			rec := httptest.NewRecorder()
			handler.NewHandler(svc, logger, "").GetUserAvatar(rec, req)
			assert.Equal(t, tt.expCode, rec.Code)
			if tt.expCode == http.StatusOK {
				assert.Equal(t, tt.expBody, rec.Body.Bytes())
				assert.Equal(t, tt.expMime, rec.Header().Get("Content-Type"))
				assert.Equal(t, "max-age=86400", rec.Header().Get("Cache-Control"))
			}
		})
	}
}

func TestHandler_GetAvatar(t *testing.T) {
	meta := models.Avatar{
		ID:       avatarID,
		UserID:   userID,
		MimeType: "image/jpeg",
	}
	tests := []struct {
		name       string
		avatarID   string
		query      string
		setupMocks func(svc *mocks.MockService)
		expCode    int
		expBody    []byte
		expMime    string
	}{
		{
			name:     "положительный тест",
			avatarID: avatarID,
			query:    "?size=small&format=jpeg",
			setupMocks: func(svc *mocks.MockService) {
				svc.On("GetAvatar", mock.Anything, avatarID, "small", "jpeg").
					Return(meta, imageData, nil).Once()
			},
			expCode: http.StatusOK,
			expBody: imageData,
			expMime: "image/jpeg",
		},
		{
			name:     "пустой параметр avatar_id",
			avatarID: "",
			expCode:  http.StatusBadRequest,
		},
		{
			name:     "аватарка не найдена",
			avatarID: avatarID,
			setupMocks: func(svc *mocks.MockService) {
				svc.On("GetAvatar", mock.Anything, avatarID, "", "").
					Return(models.Avatar{}, nil, models.ErrAvatarNotFound).Once()
			},
			expCode: http.StatusNotFound,
		},
		{
			name:     "невалидный формат",
			avatarID: avatarID,
			query:    "?format=bmp",
			setupMocks: func(svc *mocks.MockService) {
				svc.On("GetAvatar", mock.Anything, avatarID, "", "bmp").
					Return(models.Avatar{}, nil, models.ErrInvalidFormat).Once()
			},
			expCode: http.StatusBadRequest,
		},
		{
			name:     "ошибка сервиса",
			avatarID: avatarID,
			setupMocks: func(svc *mocks.MockService) {
				svc.On("GetAvatar", mock.Anything, avatarID, "", "").
					Return(models.Avatar{}, nil, errTest).Once()
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
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/avatars/"+tt.avatarID+tt.query, nil)
			if tt.avatarID != "" {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("avatar_id", tt.avatarID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			}
			rec := httptest.NewRecorder()
			handler.NewHandler(svc, logger, "").GetAvatar(rec, req)
			assert.Equal(t, tt.expCode, rec.Code)
			if tt.expCode == http.StatusOK {
				assert.Equal(t, tt.expBody, rec.Body.Bytes())
				assert.Equal(t, tt.expMime, rec.Header().Get("Content-Type"))
				assert.Equal(t, "max-age=86400", rec.Header().Get("Cache-Control"))
			}
		})
	}
}
