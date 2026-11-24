package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/danilov-go/gophprofile/internal/handler"
	"github.com/danilov-go/gophprofile/internal/handler/mocks"
	"github.com/danilov-go/gophprofile/internal/models"
)

func TestHandler_ValidateUserID(t *testing.T) {
	tests := []struct {
		name      string
		headerID  string
		urlID     string
		expCode   int
		expUserID string
		expErr    models.ErrorResponse
	}{
		{
			name:      "положительный тест",
			headerID:  userID,
			urlID:     userID,
			expCode:   http.StatusOK,
			expUserID: userID,
		},
		{
			name:    "заголовок X-User-ID отсутствует",
			expCode: http.StatusBadRequest,
			expErr: models.ErrorResponse{
				Error:   "Invalid userID",
				Details: "X-User-ID header is required and must be <= 255 chars",
			},
		},
		{
			name:     "длина X-User-ID превышает 255 символов",
			headerID: strings.Repeat("a", 256),
			expCode:  http.StatusBadRequest,
		},
		{
			name:     "user_id не совпадает с X-User-ID",
			headerID: userID,
			urlID:    unknowID,
			expCode:  http.StatusForbidden,
			expErr: models.ErrorResponse{
				Error:   "Forbidden",
				Details: "user_id does not match X-User-ID header",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()
			svc := mocks.NewMockService(t)
			var capturedUserID string
			next := func(w http.ResponseWriter, r *http.Request, id string) {
				capturedUserID = id
				w.WriteHeader(http.StatusOK)
			}
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/avatars", nil)
			if tt.headerID != "" {
				req.Header.Set("X-User-Id", tt.headerID)
			}
			if tt.urlID != "" {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("user_id", tt.urlID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			}
			rec := httptest.NewRecorder()
			handler.NewHandler(svc, logger, "").ValidateUserID(next).ServeHTTP(rec, req)
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
