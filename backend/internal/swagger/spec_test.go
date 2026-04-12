package swagger_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/kntjspr/fullstack-golang-next-template/internal/router"
	"github.com/kntjspr/fullstack-golang-next-template/internal/testutil"
)

func newSwaggerTestServer(t *testing.T) string {
	t.Helper()

	r := chi.NewRouter()
	router.GetRoutes(r, nil, nil, nil)

	server := testutil.NewTestServer(r)
	t.Cleanup(server.Close)

	return server.URL
}

func TestUIHandler_Returns200(t *testing.T) {
	baseURL := newSwaggerTestServer(t)

	resp, err := http.Get(baseURL + "/swagger/ui")
	if err != nil {
		t.Fatalf("request /swagger/ui: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestUIHandler_ReturnsHTML(t *testing.T) {
	baseURL := newSwaggerTestServer(t)

	resp, err := http.Get(baseURL + "/swagger/ui")
	if err != nil {
		t.Fatalf("request /swagger/ui: %v", err)
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "text/html") {
		t.Fatalf("unexpected content-type: %q", contentType)
	}
}

func TestUIHandler_ContainsRedocScript(t *testing.T) {
	baseURL := newSwaggerTestServer(t)

	resp, err := http.Get(baseURL + "/swagger/ui")
	if err != nil {
		t.Fatalf("request /swagger/ui: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	if !strings.Contains(strings.ToLower(string(body)), "redoc") {
		t.Fatalf("expected response body to contain redoc, body=%s", body)
	}
}
