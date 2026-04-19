package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

var (
	ErrMissingAuthorization       = errors.New("missing authorization")
	ErrInvalidAuthorizationHeader = errors.New("invalid authorization header")
)

// ExtractAuthToken gets auth token from bearer header first, then auth_token cookie.
func ExtractAuthToken(r *http.Request) (string, error) {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader != "" {
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return "", ErrInvalidAuthorizationHeader
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if token == "" {
			return "", ErrInvalidAuthorizationHeader
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

	return "", ErrMissingAuthorization
}

// WriteJSONError writes a standard JSON error payload.
func WriteJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
