package handler_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/danilov-go/gophprofile/internal/handler"
	"github.com/danilov-go/gophprofile/internal/handler/mocks"
	"github.com/danilov-go/gophprofile/internal/models"
)

func TestHandler_UploadAvatar(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	expResp := models.UploadResponse{
		ID:        avatarID,
		UserID:    userID,
		URL:       "/api/v1/avatars/" + avatarID,
		Status:    models.StatusProcessing,
		CreatedAt: now,
	}
	tests := []struct {
		name       string
		field      string
		fileName   string
		data       []byte
		rawBody    string
		setupMocks func(svc *mocks.MockService)
		expCode    int
		expBody    any
	}{
		{
			name:     "положительный тест",
			field:    "image",
			fileName: fileName,
			data:     imageData,
			setupMocks: func(svc *mocks.MockService) {
				svc.On("SaveAvatar", mock.Anything, imageData, userID, fileName, "image/png").
					Return(models.Avatar{
						ID:               avatarID,
						UserID:           userID,
						ProcessingStatus: models.StatusProcessing,
						CreatedAt:        now,
					}, nil).Once()
			},
			expCode: http.StatusCreated,
			expBody: expResp,
		},
		{
			name:    "невалидный multipart запрос",
			rawBody: "not-multipart",
			expCode: http.StatusBadRequest,
		},
		{
			name:     "отсутствует поле image",
			field:    "wrong_field",
			fileName: fileName,
			data:     imageData,
			expCode:  http.StatusBadRequest,
		},
		{
			name:     "превышен максимальный размер файла",
			field:    "image",
			fileName: fileName,
			data:     make([]byte, models.MaxSize+1024),
			expCode:  http.StatusRequestEntityTooLarge,
		},
		{
			name:     "пустой файл",
			field:    "image",
			fileName: fileName,
			data:     []byte{},
			expCode:  http.StatusBadRequest,
		},
		{
			name:     "неподдерживаемый формат файла",
			field:    "image",
			fileName: "doc.txt",
			data:     []byte("plain text"),
			setupMocks: func(svc *mocks.MockService) {
				svc.On("SaveAvatar", mock.Anything, []byte("plain text"), userID, "doc.txt", "text/plain; charset=utf-8").
					Return(models.Avatar{}, models.ErrInvalidFormat).Once()
			},
			expCode: http.StatusBadRequest,
		},
		{
			name:     "ошибка сервиса",
			field:    "image",
			fileName: fileName,
			data:     imageData,
			setupMocks: func(svc *mocks.MockService) {
				svc.On("SaveAvatar", mock.Anything, imageData, userID, fileName, "image/png").
					Return(models.Avatar{}, errTest).Once()
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
			var req *http.Request
			if tt.rawBody != "" {
				req = httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/avatars", bytes.NewBufferString(tt.rawBody))
				req.Header.Set("Content-Type", "application/json")
			} else {
				var body bytes.Buffer
				w := multipart.NewWriter(&body)
				part, err := w.CreateFormFile(tt.field, tt.fileName)
				require.NoError(t, err)
				_, err = part.Write(tt.data)
				require.NoError(t, err)
				require.NoError(t, w.Close())

				req = httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/avatars", &body)
				req.Header.Set("Content-Type", w.FormDataContentType())
			}
			rec := httptest.NewRecorder()
			handler.NewHandler(svc, logger, "").UploadAvatar(rec, req, userID)
			assert.Equal(t, tt.expCode, rec.Code)
			assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
			if tt.expBody != nil {
				var actual models.UploadResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &actual))
				assert.Equal(t, tt.expBody, actual)
			}
		})
	}
}
