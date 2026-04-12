package middleware

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/create-go-app/chi-go-template/internal/testutil"
)

func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		origin     string
		headers    map[string]string
		assertFunc func(t *testing.T, resp *http.Response, body string)
	}{
		{
			name:   "TestCORS_AllowedOrigin",
			method: http.MethodGet,
			origin: "https://allowed.example.com",
			assertFunc: func(t *testing.T, resp *http.Response, _ string) {
				t.Helper()
				if resp.Header.Get("Access-Control-Allow-Origin") != "https://allowed.example.com" {
					t.Fatalf("missing allow-origin header for allowed origin")
				}
			},
		},
		{
			name:   "TestCORS_DisallowedOrigin",
			method: http.MethodGet,
			origin: "https://blocked.example.com",
			assertFunc: func(t *testing.T, resp *http.Response, _ string) {
				t.Helper()
				if value := resp.Header.Get("Access-Control-Allow-Origin"); value != "" {
					t.Fatalf("unexpected allow-origin header %q for blocked origin", value)
				}
			},
		},
		{
			name:   "TestCORS_Preflight",
			method: http.MethodOptions,
			origin: "https://allowed.example.com",
			headers: map[string]string{
				"Access-Control-Request-Method":  "POST",
				"Access-Control-Request-Headers": "Authorization, Content-Type",
			},
			assertFunc: func(t *testing.T, resp *http.Response, _ string) {
				t.Helper()
				if resp.StatusCode != http.StatusNoContent {
					t.Fatalf("unexpected preflight status: got %d want %d", resp.StatusCode, http.StatusNoContent)
				}
				if resp.Header.Get("Access-Control-Allow-Origin") != "https://allowed.example.com" {
					t.Fatalf("preflight allow-origin mismatch")
				}
				if methods := resp.Header.Get("Access-Control-Allow-Methods"); !strings.Contains(methods, "POST") {
					t.Fatalf("expected preflight methods to include POST, got %q", methods)
				}
				if headers := resp.Header.Get("Access-Control-Allow-Headers"); !strings.Contains(headers, "Authorization") {
					t.Fatalf("expected preflight headers to include Authorization, got %q", headers)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Use(CORS([]string{"https://allowed.example.com"}))
			r.MethodFunc(http.MethodGet, "/", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`ok`))
			})
			r.MethodFunc(http.MethodOptions, "/", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})

			server := testutil.NewTestServer(r)
			defer server.Close()

			req, err := http.NewRequest(tc.method, server.URL+"/", nil)
			if err != nil {
				t.Fatalf("create request: %v", err)
			}
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("execute request: %v", err)
			}
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}

			tc.assertFunc(t, resp, string(bodyBytes))
		})
	}
}
