package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/kntjspr/fullstack-golang-next-template/internal/testutil"
)

func TestRecoverMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		panicValue any
		logCapture func(string, ...any)
		assertFunc func(t *testing.T, status int, body string, logged bool)
	}{
		{
			name:       "TestRecover_PanicReturns500",
			panicValue: "boom",
			assertFunc: func(t *testing.T, status int, body string, _ bool) {
				t.Helper()
				if status != http.StatusInternalServerError {
					t.Fatalf("unexpected status: got %d want %d", status, http.StatusInternalServerError)
				}

				var payload map[string]string
				if err := json.Unmarshal([]byte(body), &payload); err != nil {
					t.Fatalf("decode JSON body: %v", err)
				}
				if payload["error"] == "" {
					t.Fatalf("expected JSON error body, got %s", body)
				}
			},
		},
		{
			name: "TestRecover_NoPanicPassesThrough",
			assertFunc: func(t *testing.T, status int, body string, _ bool) {
				t.Helper()
				if status != http.StatusCreated {
					t.Fatalf("unexpected status: got %d want %d", status, http.StatusCreated)
				}
				if body != "created" {
					t.Fatalf("unexpected body: got %q want %q", body, "created")
				}
			},
		},
		{
			name:       "TestRecover_LogsPanic",
			panicValue: "logged-panic",
			assertFunc: func(t *testing.T, status int, _ string, logged bool) {
				t.Helper()
				if status != http.StatusInternalServerError {
					t.Fatalf("unexpected status: got %d want %d", status, http.StatusInternalServerError)
				}
				if !logged {
					t.Fatal("expected logger to be called for panic")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logged := false
			logFn := func(format string, args ...any) {
				if strings.TrimSpace(format) != "" {
					logged = true
				}
			}

			r := chi.NewRouter()
			r.Use(Recover(logFn))
			r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
				if tc.panicValue != nil {
					panic(tc.panicValue)
				}
				w.WriteHeader(http.StatusCreated)
				_, _ = io.WriteString(w, "created")
			})

			server := testutil.NewTestServer(r)
			defer server.Close()

			resp, err := http.Get(server.URL + "/") //nolint:noctx
			if err != nil {
				t.Fatalf("execute request: %v", err)
			}
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read response body: %v", err)
			}

			tc.assertFunc(t, resp.StatusCode, string(bodyBytes), logged)
		})
	}
}
