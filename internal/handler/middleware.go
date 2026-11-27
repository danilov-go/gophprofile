package handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/danilov-go/gophprofile/internal/models"
)

type UserHandlerFunc func(w http.ResponseWriter, r *http.Request, userID string)

func (h *Handler) ValidateUserID(key string, next UserHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(authHeader, "Bearer ")
		if !ok || token == "" {
			h.writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{
				Error:   "Unauthorized",
				Details: "Authorization Bearer token is required",
			})
			return
		}
		userID, err := h.service.GetUserID(token, key)
		if err != nil || userID == "" {
			h.writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{
				Error:   "Unauthorized",
				Details: "Invalid or expired token",
			})
			return
		}
		if urlUserID := chi.URLParam(r, "user_id"); urlUserID != "" && urlUserID != userID {
			h.writeJSON(w, http.StatusForbidden, models.ErrorResponse{
				Error:   "Forbidden",
				Details: "user_id does not match authenticated user",
			})
			return
		}
		next(w, r, userID)
	}
}
