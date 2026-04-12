package router

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/create-go-app/chi-go-template/internal/auth"
	"github.com/create-go-app/chi-go-template/internal/models"
)

const (
	loginTokenTTL   = time.Hour
	refreshTokenTTL = 2 * time.Hour
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

// RegisterAuthRoutes mounts auth-related handlers into the passed router.
func RegisterAuthRoutes(r chi.Router, db *gorm.DB) {
	r.Route("/auth", func(ar chi.Router) {
		ar.Post("/login", loginHandler(db))
		ar.Post("/refresh", refreshHandler())
		ar.Post("/logout", logoutHandler())
	})
}

func loginHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			writeAuthError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}

		var payload loginRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeValidationErrors(w, []string{"invalid JSON body"})
			return
		}

		payload.Email = strings.TrimSpace(payload.Email)
		payload.Password = strings.TrimSpace(payload.Password)

		validationErrors := make([]string, 0, 2)
		if payload.Email == "" {
			validationErrors = append(validationErrors, "email is required")
		}
		if payload.Password == "" {
			validationErrors = append(validationErrors, "password is required")
		}
		if len(validationErrors) > 0 {
			writeValidationErrors(w, validationErrors)
			return
		}

		var user models.User
		if err := db.Where("email = ?", payload.Email).First(&user).Error; err != nil {
			writeAuthError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		if user.PasswordHash != payload.Password {
			writeAuthError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		token, err := auth.GenerateToken(user.ID, user.Role, loginTokenTTL)
		if err != nil {
			writeAuthError(w, http.StatusInternalServerError, "failed to issue token")
			return
		}
		claims, err := auth.ValidateToken(token)
		if err != nil {
			writeAuthError(w, http.StatusInternalServerError, "failed to issue token")
			return
		}

		setAuthCookie(w, token, claims.ExpiresAt)
		writeTokenResponse(w, http.StatusOK, token, claims.ExpiresAt)
	}
}

func refreshHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := tokenFromRequest(r)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, err.Error())
			return
		}

		claims, err := auth.ValidateToken(token)
		if err != nil {
			if errors.Is(err, auth.ErrTokenExpired) {
				writeAuthError(w, http.StatusUnauthorized, "token expired")
				return
			}
			writeAuthError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		newToken, err := auth.GenerateToken(claims.UserID, claims.Role, refreshTokenTTL)
		if err != nil {
			writeAuthError(w, http.StatusInternalServerError, "failed to issue token")
			return
		}
		newClaims, err := auth.ValidateToken(newToken)
		if err != nil {
			writeAuthError(w, http.StatusInternalServerError, "failed to issue token")
			return
		}

		setAuthCookie(w, newToken, newClaims.ExpiresAt)
		writeTokenResponse(w, http.StatusOK, newToken, newClaims.ExpiresAt)
	}
}

func logoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		clearAuthCookie(w)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "logged out"})
	}
}

func tokenFromRequest(r *http.Request) (string, error) {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader != "" {
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return "", errors.New("invalid authorization header")
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if token == "" {
			return "", errors.New("invalid authorization header")
		}
		return token, nil
	}

	cookie, err := r.Cookie("auth_token")
	if err == nil {
		token := strings.TrimSpace(cookie.Value)
		if token != "" {
			return token, nil
		}
	}

	return "", errors.New("missing authorization")
}

func setAuthCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production"),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
		Expires:  expiresAt.UTC(),
	})
}

func clearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production"),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0).UTC(),
	})
}

func writeTokenResponse(w http.ResponseWriter, status int, token string, expiresAt time.Time) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(tokenResponse{
		Token:     token,
		ExpiresAt: expiresAt.UTC().Format(time.RFC3339),
	})
}

func writeValidationErrors(w http.ResponseWriter, validationErrors []string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	_ = json.NewEncoder(w).Encode(map[string][]string{
		"errors": validationErrors,
	})
}

func writeAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
