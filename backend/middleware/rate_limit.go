package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"
)

type rateBucket struct {
	count       int
	windowStart time.Time
}

// RateLimiter tracks request counts by client and time window.
type RateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	nowFunc func() time.Time
	buckets map[string]rateBucket
}

// NewRateLimiter creates a request limiter.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:   limit,
		window:  window,
		nowFunc: time.Now,
		buckets: map[string]rateBucket{},
	}
}

// Middleware enforces request limits with a fixed time window.
func (r *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if r.allow(req) {
			next.ServeHTTP(w, req)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "rate limit exceeded"})
	})
}

func (r *RateLimiter) allow(req *http.Request) bool {
	clientKey := clientIP(req.RemoteAddr)
	now := r.nowFunc().UTC()

	r.mu.Lock()
	defer r.mu.Unlock()

	bucket := r.buckets[clientKey]
	if bucket.windowStart.IsZero() || now.Sub(bucket.windowStart) >= r.window {
		bucket = rateBucket{
			count:       1,
			windowStart: now,
		}
		r.buckets[clientKey] = bucket
		return true
	}

	if bucket.count >= r.limit {
		return false
	}

	bucket.count++
	r.buckets[clientKey] = bucket
	return true
}

func clientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}
