package middleware

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"time"
)

const RequestIDHeader = "X-Request-ID"

type requestIDContextKey string

const requestIDKey requestIDContextKey = "request_id"

var randRead = rand.Read

// RequestID injects request identifiers into request context and response headers.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = generateUUIDv4()
			if requestID == "" {
				requestID = fallbackRequestID()
			}
		}

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		w.Header().Set(RequestIDHeader, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext returns the current request identifier from context.
func RequestIDFromContext(ctx context.Context) string {
	if value, ok := ctx.Value(requestIDKey).(string); ok {
		return value
	}
	return ""
}

func generateUUIDv4() string {
	var b [16]byte
	if _, err := randRead(b[:]); err != nil {
		return ""
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		b[0:4],
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:16],
	)
}

func fallbackRequestID() string {
	return fmt.Sprintf("req-%d", time.Now().UTC().UnixNano())
}
