package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/kntjspr/fullstack-golang-next-template/internal/testutil"
)

func TestRateLimitMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() (*RateLimiter, int, time.Duration)
		assertFunc func(t *testing.T, serverURL string, limiter *RateLimiter, limit int, window time.Duration)
	}{
		{
			name: "TestRateLimit_AllowsUnderLimit",
			setup: func() (*RateLimiter, int, time.Duration) {
				limit := 100
				window := 2 * time.Second
				return NewRateLimiter(limit, window), limit, window
			},
			assertFunc: func(t *testing.T, serverURL string, _ *RateLimiter, _ int, _ time.Duration) {
				t.Helper()
				for i := 0; i < 99; i++ {
					status, body := requestStatusAndBody(t, serverURL)
					if status != http.StatusOK {
						t.Fatalf("request #%d returned %d: %s", i+1, status, body)
					}
				}
			},
		},
		{
			name: "TestRateLimit_BlocksAtLimit",
			setup: func() (*RateLimiter, int, time.Duration) {
				limit := 100
				window := 2 * time.Second
				return NewRateLimiter(limit, window), limit, window
			},
			assertFunc: func(t *testing.T, serverURL string, _ *RateLimiter, _ int, _ time.Duration) {
				t.Helper()
				for i := 0; i < 100; i++ {
					status, body := requestStatusAndBody(t, serverURL)
					if status != http.StatusOK {
						t.Fatalf("request #%d returned %d: %s", i+1, status, body)
					}
				}

				status, body := requestStatusAndBody(t, serverURL)
				if status != http.StatusTooManyRequests {
					t.Fatalf("request #101 returned %d: %s", status, body)
				}

				var payload map[string]string
				if err := json.Unmarshal([]byte(body), &payload); err != nil {
					t.Fatalf("decode rate-limit JSON body: %v", err)
				}
				if payload["error"] == "" {
					t.Fatalf("expected JSON error body, got %s", body)
				}
			},
		},
		{
			name: "TestRateLimit_ResetsAfterWindow",
			setup: func() (*RateLimiter, int, time.Duration) {
				limit := 100
				window := 200 * time.Millisecond
				return NewRateLimiter(limit, window), limit, window
			},
			assertFunc: func(t *testing.T, serverURL string, _ *RateLimiter, _ int, window time.Duration) {
				t.Helper()
				for i := 0; i < 100; i++ {
					status, body := requestStatusAndBody(t, serverURL)
					if status != http.StatusOK {
						t.Fatalf("request #%d returned %d: %s", i+1, status, body)
					}
				}

				time.Sleep(window + 25*time.Millisecond)

				status, body := requestStatusAndBody(t, serverURL)
				if status != http.StatusOK {
					t.Fatalf("request after reset returned %d: %s", status, body)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			limiter, limit, window := tc.setup()

			r := chi.NewRouter()
			r.Use(limiter.Middleware)
			r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"ok":true}`))
			})

			server := testutil.NewTestServer(r)
			defer server.Close()

			tc.assertFunc(t, server.URL+"/", limiter, limit, window)
		})
	}
}

func requestStatusAndBody(t *testing.T, url string) (int, string) {
	t.Helper()

	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		t.Fatalf("make request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	return resp.StatusCode, string(bodyBytes)
}

func TestRateLimit_UsesForwardedClientIP(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)

	reqOne := httptest.NewRequest(http.MethodGet, "/", nil)
	reqOne.RemoteAddr = "10.0.0.1:1234"
	reqOne.Header.Set("X-Forwarded-For", "198.51.100.10")
	if !limiter.allow(reqOne) {
		t.Fatal("first forwarded request should be allowed")
	}

	reqTwo := httptest.NewRequest(http.MethodGet, "/", nil)
	reqTwo.RemoteAddr = "10.0.0.1:1235"
	reqTwo.Header.Set("X-Forwarded-For", "198.51.100.10")
	if limiter.allow(reqTwo) {
		t.Fatal("second request from same forwarded IP should be limited")
	}
}

func TestRateLimit_EvictsStaleBuckets(t *testing.T) {
	limiter := NewRateLimiter(2, 100*time.Millisecond)
	start := time.Now().UTC()
	limiter.nowFunc = func() time.Time { return start }

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.1:4567"
	if !limiter.allow(req) {
		t.Fatal("initial request should be allowed")
	}
	if len(limiter.buckets) != 1 {
		t.Fatalf("expected one bucket, got %d", len(limiter.buckets))
	}

	limiter.nowFunc = func() time.Time { return start.Add(250 * time.Millisecond) }
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "203.0.113.2:4568"
	if !limiter.allow(req2) {
		t.Fatal("request after window should be allowed")
	}

	if len(limiter.buckets) != 1 {
		t.Fatalf("expected stale bucket eviction, got %d buckets", len(limiter.buckets))
	}
}
