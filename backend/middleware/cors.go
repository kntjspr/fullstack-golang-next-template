package middleware

import (
	"net/http"
	"strings"
)

// CORS applies CORS headers for explicitly allowed origins.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := map[string]struct{}{}
	for _, origin := range allowedOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			allowed[trimmed] = struct{}{}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			_, isAllowed := allowed[origin]

			if isAllowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == http.MethodOptions && isAllowed {
				requestMethod := r.Header.Get("Access-Control-Request-Method")
				if requestMethod == "" {
					requestMethod = "GET,POST,PUT,PATCH,DELETE,OPTIONS"
				} else {
					requestMethod += ",OPTIONS"
				}
				w.Header().Set("Access-Control-Allow-Methods", requestMethod)

				requestHeaders := r.Header.Get("Access-Control-Request-Headers")
				if requestHeaders == "" {
					requestHeaders = "Authorization,Content-Type"
				}
				w.Header().Set("Access-Control-Allow-Headers", requestHeaders)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
