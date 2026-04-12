package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestSecurityHeaders_AllPresent(t *testing.T) {
	h := testSecurityHeadersHandler()
	server := httptest.NewTLSServer(h)
	defer server.Close()

	resp, err := server.Client().Get(server.URL) //nolint:noctx
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()

	headers := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Content-Security-Policy":   "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'",
		"Permissions-Policy":        "camera=(), microphone=(), geolocation=()",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
	}

	for key, expected := range headers {
		if got := resp.Header.Get(key); got != expected {
			t.Fatalf("header %s mismatch: got %q want %q", key, got, expected)
		}
	}
}

func TestSecurityHeaders_CSP_NoUnsafeInlineScript(t *testing.T) {
	h := testSecurityHeadersHandler()
	server := httptest.NewServer(h)
	defer server.Close()

	resp, err := http.Get(server.URL) //nolint:noctx
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()

	csp := resp.Header.Get("Content-Security-Policy")
	if csp == "" {
		t.Fatal("missing Content-Security-Policy header")
	}
	if strings.Contains(csp, "script-src 'self' 'unsafe-inline'") {
		t.Fatalf("script-src unexpectedly contains unsafe-inline: %q", csp)
	}
}

func TestSecurityHeaders_HSTS_OnlyOnHTTPS(t *testing.T) {
	h := testSecurityHeadersHandler()

	httpServer := httptest.NewServer(h)
	defer httpServer.Close()

	httpsServer := httptest.NewTLSServer(h)
	defer httpsServer.Close()

	httpResp, err := http.Get(httpServer.URL) //nolint:noctx
	if err != nil {
		t.Fatalf("execute http request: %v", err)
	}
	defer httpResp.Body.Close()

	httpsResp, err := httpsServer.Client().Get(httpsServer.URL) //nolint:noctx
	if err != nil {
		t.Fatalf("execute https request: %v", err)
	}
	defer httpsResp.Body.Close()

	if got := httpResp.Header.Get("Strict-Transport-Security"); got != "" {
		t.Fatalf("unexpected HSTS header on HTTP: %q", got)
	}

	hsts := httpsResp.Header.Get("Strict-Transport-Security")
	if hsts != "max-age=31536000; includeSubDomains" {
		t.Fatalf("unexpected HSTS header on HTTPS: got %q", hsts)
	}
}

func TestCORS_AllowedOrigin(t *testing.T) {
	r := chi.NewRouter()
	r.Use(CORS([]string{"https://allowed.example.com"}))
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	})

	server := httptest.NewServer(r)
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Origin", "https://allowed.example.com")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "https://allowed.example.com" {
		t.Fatalf("unexpected Access-Control-Allow-Origin: got %q want %q", got, "https://allowed.example.com")
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	r := chi.NewRouter()
	r.Use(CORS([]string{"https://allowed.example.com"}))
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	})

	server := httptest.NewServer(r)
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Origin", "https://blocked.example.com")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unexpected Access-Control-Allow-Origin for blocked origin: %q", got)
	}
}

func testSecurityHeadersHandler() http.Handler {
	r := chi.NewRouter()
	r.Use(SecurityHeaders)
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	})
	return r
}
