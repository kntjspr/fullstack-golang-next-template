package middleware

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/kntjspr/fullstack-golang-next-template/internal/testutil"
)

func TestRequestIDMiddleware(t *testing.T) {
	tests := []struct {
		name                 string
		requestHeader        string
		expectedStatus       int
		expectSameHeader     bool
		expectGeneratedV4    bool
		expectContextRequest bool
	}{
		{
			name:                 "TestRequestID_GeneratesWhenMissing",
			expectedStatus:       http.StatusOK,
			expectGeneratedV4:    true,
			expectContextRequest: true,
		},
		{
			name:                 "TestRequestID_PreservesExisting",
			requestHeader:        "existing-request-id",
			expectedStatus:       http.StatusOK,
			expectSameHeader:     true,
			expectContextRequest: true,
		},
		{
			name:                 "TestRequestID_PropagatesInContext",
			expectedStatus:       http.StatusOK,
			expectGeneratedV4:    true,
			expectContextRequest: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Use(RequestID)
			r.Get("/", func(w http.ResponseWriter, req *http.Request) {
				requestID := RequestIDFromContext(req.Context())
				w.Header().Set("X-Request-ID-From-Context", requestID)
				w.WriteHeader(http.StatusOK)
				_, _ = io.WriteString(w, "ok")
			})

			server := testutil.NewTestServer(r)
			defer server.Close()

			req, err := http.NewRequest(http.MethodGet, server.URL+"/", nil)
			if err != nil {
				t.Fatalf("create request: %v", err)
			}
			if tc.requestHeader != "" {
				req.Header.Set(RequestIDHeader, tc.requestHeader)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("execute request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Fatalf("unexpected status code: got %d want %d", resp.StatusCode, tc.expectedStatus)
			}

			responseRequestID := resp.Header.Get(RequestIDHeader)
			contextRequestID := resp.Header.Get("X-Request-ID-From-Context")

			if tc.expectSameHeader && responseRequestID != tc.requestHeader {
				t.Fatalf("request id should be preserved: got %q want %q", responseRequestID, tc.requestHeader)
			}

			if tc.expectGeneratedV4 && !isUUIDv4(responseRequestID) {
				t.Fatalf("request id should be valid UUID v4, got %q", responseRequestID)
			}

			if tc.expectContextRequest && contextRequestID == "" {
				t.Fatal("context request id is empty")
			}

			if tc.expectContextRequest && contextRequestID != responseRequestID {
				t.Fatalf("context request id mismatch: got %q want %q", contextRequestID, responseRequestID)
			}
		})
	}
}

func TestRequestIDFallbackWhenRandomFails(t *testing.T) {
	originalRandRead := randRead
	randRead = func([]byte) (int, error) {
		return 0, errors.New("entropy unavailable")
	}
	t.Cleanup(func() {
		randRead = originalRandRead
	})

	r := chi.NewRouter()
	r.Use(RequestID)
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("X-Request-ID-From-Context", RequestIDFromContext(req.Context()))
		w.WriteHeader(http.StatusOK)
	})

	server := testutil.NewTestServer(r)
	defer server.Close()

	resp, err := http.Get(server.URL + "/") //nolint:noctx
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", resp.StatusCode, http.StatusOK)
	}

	requestID := resp.Header.Get(RequestIDHeader)
	if requestID == "" {
		t.Fatal("expected non-empty request id fallback")
	}
	if !strings.HasPrefix(requestID, "req-") {
		t.Fatalf("expected fallback request id prefix, got %q", requestID)
	}
}

func isUUIDv4(value string) bool {
	parts := strings.Split(value, "-")
	if len(parts) != 5 {
		return false
	}
	if len(parts[0]) != 8 || len(parts[1]) != 4 || len(parts[2]) != 4 || len(parts[3]) != 4 || len(parts[4]) != 12 {
		return false
	}
	return strings.HasPrefix(parts[2], "4")
}
