package handler

import (
	"net/http"
	"strings"

	"github.com/danilov-go/gophprofile/internal/models"
	"github.com/go-chi/chi/v5"
)

type UserHandlerFunc func(w http.ResponseWriter, r *http.Request, userID string)

func (h *Handler) ValidateUserID(next UserHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headerUserID := strings.TrimSpace(r.Header.Get("X-User-ID"))
		if headerUserID == "" || len(headerUserID) > 255 {
			h.writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
				Error:   "Invalid userID",
				Details: "X-User-ID header is required and must be <= 255 chars",
			})
			return
		}
		if userID := chi.URLParam(r, "user_id"); userID != "" && userID != headerUserID {
			h.writeJSON(w, http.StatusForbidden, models.ErrorResponse{
				Error:   "Forbidden",
				Details: "user_id does not match X-User-ID header",
			})
			return
		}
		next(w, r, headerUserID)
	}
}
