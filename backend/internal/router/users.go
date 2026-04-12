package router

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/kntjspr/fullstack-golang-next-template/internal/models"
	"github.com/kntjspr/fullstack-golang-next-template/middleware"
)

type getMePayload struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// UsersRoutes mounts user-related handlers into the passed router.
func UsersRoutes(r chi.Router, db *gorm.DB) {
	r.Route("/users", func(ur chi.Router) {
		ur.Use(middleware.Authenticator)
		ur.Get("/me", GetMe(db))
	})
}

// GetMe returns the authenticated user's profile.
func GetMe(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			writeAuthError(w, http.StatusInternalServerError, "database unavailable")
			return
		}

		userID := strings.TrimSpace(middleware.UserIDFromContext(r.Context()))
		if userID == "" {
			writeAuthError(w, http.StatusUnauthorized, "missing authorization")
			return
		}

		var user models.User
		if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				writeAuthError(w, http.StatusNotFound, "user not found")
				return
			}

			writeAuthError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(getMePayload{
			ID:        user.ID,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt.UTC(),
		})
	}
}
