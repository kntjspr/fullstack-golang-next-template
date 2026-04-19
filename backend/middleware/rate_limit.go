package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type rateBucket struct {
	count       int
	windowStart time.Time
	lastSeen    time.Time
}

// RateLimiter tracks request counts by client and time window.
type RateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	nowFunc func() time.Time
	buckets map[string]rateBucket

	lastCleanup time.Time
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
	clientKey := clientIP(req)
	now := r.nowFunc().UTC()

	r.mu.Lock()
	defer r.mu.Unlock()

	r.cleanupStaleBuckets(now)

	bucket := r.buckets[clientKey]
	if bucket.windowStart.IsZero() || now.Sub(bucket.windowStart) >= r.window {
		bucket = rateBucket{
			count:       1,
			windowStart: now,
			lastSeen:    now,
		}
		r.buckets[clientKey] = bucket
		return true
	}

	if bucket.count >= r.limit {
		bucket.lastSeen = now
		r.buckets[clientKey] = bucket
		return false
	}

	bucket.count++
	bucket.lastSeen = now
	r.buckets[clientKey] = bucket
	return true
}

func (r *RateLimiter) cleanupStaleBuckets(now time.Time) {
	if !r.lastCleanup.IsZero() && now.Sub(r.lastCleanup) < r.window {
		return
	}

	for key, bucket := range r.buckets {
		if now.Sub(bucket.lastSeen) >= r.window {
			delete(r.buckets, key)
		}
	}
	r.lastCleanup = now
}

func clientIP(req *http.Request) string {
	if ip := firstForwardedIP(req.Header.Get("CF-Connecting-IP")); ip != "" {
		return ip
	}
	if ip := firstForwardedIP(req.Header.Get("X-Forwarded-For")); ip != "" {
		return ip
	}
	if ip := firstForwardedIP(req.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}

	remoteAddr := strings.TrimSpace(req.RemoteAddr)
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

func firstForwardedIP(value string) string {
	for _, part := range strings.Split(value, ",") {
		candidate := strings.TrimSpace(part)
		if candidate == "" {
			continue
		}

		host, _, err := net.SplitHostPort(candidate)
		if err == nil {
			candidate = host
		}

		if parsed := net.ParseIP(candidate); parsed != nil {
			return parsed.String()
		}
	}

	return ""
}
