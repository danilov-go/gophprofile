package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap/zaptest"

	"github.com/danilov-go/gophprofile/internal/handler/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/danilov-go/gophprofile/internal/handler"
	"github.com/danilov-go/gophprofile/internal/models"
)

const (
	avatarID = "1f0f950c-fabd-47cc-b72e-4a268ef169e1"
	userID   = "userID"
	unknowID = "unknowID"
	fileName = "avatar.jpg"
)

var (
	imageData = []byte("\x89PNG\r\n\x1a\nimageData")
	errTest   = errors.New("test error")
)

func TestHandler_GetForm(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	testDir := t.TempDir()
	expHTML := "<html><body>Upload Form</body></html>"
	require.NoError(t, os.WriteFile(filepath.Join(testDir, "index.html"), []byte(expHTML), 0o600))
	svc := mocks.NewMockService(t)
	h := handler.NewHandler(svc, logger, testDir)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.GetForm(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, expHTML, rec.Body.String())
}

func TestHandler_GetGallery(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	testDir := t.TempDir()
	expHTML := "<html><body>Gallery</body></html>"
	require.NoError(t, os.WriteFile(filepath.Join(testDir, "gallery.html"), []byte(expHTML), 0o600))
	svc := mocks.NewMockService(t)
	h := handler.NewHandler(svc, logger, testDir)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/gallery", nil)
	rec := httptest.NewRecorder()
	h.GetGallery(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, expHTML, rec.Body.String())
}

func TestHandler_Health(t *testing.T) {
	tests := []struct {
		name           string
		healthResp     models.HealthResponse
		expectedStatus int
	}{
		{
			name: "положительный тест",
			healthResp: models.HealthResponse{
				DB:     "ok",
				S3:     "ok",
				Broker: "ok",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "ошибка БД",
			healthResp: models.HealthResponse{
				DB:     errTest.Error(),
				S3:     "ok",
				Broker: "ok",
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name: "ошибка S3",
			healthResp: models.HealthResponse{
				DB:     "ok",
				S3:     errTest.Error(),
				Broker: "ok",
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name: "ошибка брокера",
			healthResp: models.HealthResponse{
				DB:     "ok",
				S3:     "ok",
				Broker: errTest.Error(),
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()
			svc := mocks.NewMockService(t)
			svc.On("Health", mock.Anything).Return(tt.healthResp).Once()
			h := handler.NewHandler(svc, logger, "")
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()
			h.Health(rec, req)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
			var resp models.HealthResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, tt.healthResp, resp)
		})
	}
}
