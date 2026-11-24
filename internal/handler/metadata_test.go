package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/danilov-go/gophprofile/internal/handler"
	"github.com/danilov-go/gophprofile/internal/handler/mocks"
	"github.com/danilov-go/gophprofile/internal/models"
)

func TestHandler_GetAvatarMetadata(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	meta := models.MetaDataResponse{
		ID:         avatarID,
		UserID:     userID,
		FileName:   fileName,
		MimeType:   "image/jpeg",
		Size:       1024,
		Dimensions: models.Dimensions{Width: 700, Height: 700},
		Thumbnails: []models.Thumbnail{
			{Size: string(models.SizeSmall), URL: "/api/v1/avatars/" + avatarID + "?size=" + string(models.SizeSmall)},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
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
				svc.On("GetMetadata", mock.Anything, avatarID).Return(meta, nil).Once()
			},
			expCode: http.StatusOK,
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
				svc.On("GetMetadata", mock.Anything, avatarID).Return(models.MetaDataResponse{}, models.ErrAvatarNotFound).Once()
			},
			expCode: http.StatusNotFound,
		},
		{
			name:     "ошибка сервиса",
			avatarID: avatarID,
			setupMocks: func(svc *mocks.MockService) {
				svc.On("GetMetadata", mock.Anything, avatarID).Return(models.MetaDataResponse{}, errTest).Once()
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
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/avatars/"+tt.avatarID+"/metadata", nil)
			if tt.avatarID != "" {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("avatar_id", tt.avatarID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			}
			rec := httptest.NewRecorder()
			handler.NewHandler(svc, logger, "").GetAvatarMetadata(rec, req)
			assert.Equal(t, tt.expCode, rec.Code)
			assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
			if tt.expCode == http.StatusOK {
				var actual models.MetaDataResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &actual))
				assert.Equal(t, meta, actual)
			}
			if tt.expErr.Error != "" {
				var errResp models.ErrorResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errResp))
				assert.Equal(t, tt.expErr, errResp)
			}
		})
	}
}

func TestHandler_GetUserAvatars(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	avatars := []models.MetaDataResponse{
		{
			ID:         avatarID,
			UserID:     userID,
			FileName:   fileName,
			MimeType:   "image/jpeg",
			Size:       512,
			Dimensions: models.Dimensions{Width: 700, Height: 700},
			Thumbnails: []models.Thumbnail{},
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}
	tests := []struct {
		name       string
		userID     string
		setupMocks func(svc *mocks.MockService)
		expCode    int
		expErr     models.ErrorResponse
	}{
		{
			name:   "положительный тест",
			userID: userID,
			setupMocks: func(svc *mocks.MockService) {
				svc.On("GetAvatarsByUserID", mock.Anything, userID).Return(avatars, nil).Once()
			},
			expCode: http.StatusOK,
		},
		{
			name:    "пустой параметр user_id",
			userID:  "",
			expCode: http.StatusBadRequest,
			expErr: models.ErrorResponse{
				Error:   "Empty userID",
				Details: "user_id path parameter is required",
			},
		},
		{
			name:   "ошибка сервиса",
			userID: userID,
			setupMocks: func(svc *mocks.MockService) {
				svc.On("GetAvatarsByUserID", mock.Anything, userID).Return(nil, errTest).Once()
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
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/users/"+tt.userID+"/avatars", nil)
			if tt.userID != "" {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("user_id", tt.userID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			}
			rec := httptest.NewRecorder()
			handler.NewHandler(svc, logger, "").GetUserAvatars(rec, req)
			assert.Equal(t, tt.expCode, rec.Code)
			assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
			if tt.expCode == http.StatusOK {
				var actual []models.MetaDataResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &actual))
				assert.Equal(t, avatars, actual)
			}
			if tt.expErr.Error != "" {
				var errResp models.ErrorResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errResp))
				assert.Equal(t, tt.expErr, errResp)
			}
		})
	}
}
