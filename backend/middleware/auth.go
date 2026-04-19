package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/kntjspr/fullstack-golang-next-template/internal/auth"
	"github.com/kntjspr/fullstack-golang-next-template/internal/httpapi"
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
		token, err := httpapi.ExtractAuthToken(r)
		if err != nil {
			httpapi.WriteJSONError(w, http.StatusUnauthorized, err.Error())
			return
		}

		claims, err := auth.ValidateToken(token)
		if err != nil {
			if errors.Is(err, auth.ErrTokenExpired) {
				httpapi.WriteJSONError(w, http.StatusUnauthorized, "token expired")
				return
			}
			httpapi.WriteJSONError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		ctx := WithAuthContext(r.Context(), claims.UserID, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RoleRequired enforces a required role in the current request context.
func RoleRequired(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := RoleFromContext(r.Context())
			if role != requiredRole {
				httpapi.WriteJSONError(w, http.StatusForbidden, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
