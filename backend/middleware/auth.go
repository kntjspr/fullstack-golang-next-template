package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/create-go-app/chi-go-template/internal/auth"
)

type authContextKey string

const (
	authUserIDKey authContextKey = "auth_user_id"
	authRoleKey   authContextKey = "auth_role"
)

// WithAuthContext stores authenticated user fields in request context.
func WithAuthContext(ctx context.Context, userID, role string) context.Context {
	ctx = context.WithValue(ctx, authUserIDKey, userID)
	ctx = context.WithValue(ctx, authRoleKey, role)
	return ctx
}

// UserIDFromContext returns authenticated user id from context.
func UserIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(authUserIDKey).(string)
	return value
}

// RoleFromContext returns authenticated role from context.
func RoleFromContext(ctx context.Context) string {
	value, _ := ctx.Value(authRoleKey).(string)
	return value
}

// Authenticator validates auth tokens (Bearer header first, then auth cookie) and enriches request context with claims.
func Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := extractToken(r)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, err.Error())
			return
		}

		claims, err := auth.ValidateToken(token)
		if err != nil {
			if errors.Is(err, auth.ErrTokenExpired) {
				writeJSONError(w, http.StatusUnauthorized, "token expired")
				return
			}
			writeJSONError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		ctx := WithAuthContext(r.Context(), claims.UserID, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractToken(r *http.Request) (string, error) {
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
		cookieToken := strings.TrimSpace(cookie.Value)
		if cookieToken != "" {
			return cookieToken, nil
		}
	}

	return "", errors.New("missing authorization")
}

// RoleRequired enforces a required role in the current request context.
func RoleRequired(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := RoleFromContext(r.Context())
			if role != requiredRole {
				writeJSONError(w, http.StatusForbidden, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
