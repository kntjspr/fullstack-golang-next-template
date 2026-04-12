package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/kntjspr/fullstack-golang-next-template/internal/auth"
	"github.com/kntjspr/fullstack-golang-next-template/internal/testutil"
)

func TestAuthenticator(t *testing.T) {
	t.Setenv("JWT_SECRET", "auth-middleware-secret")

	validToken, err := auth.GenerateToken("auth-user-1", "admin", time.Hour)
	if err != nil {
		t.Fatalf("generate valid token: %v", err)
	}
	expiredToken, err := auth.GenerateToken("auth-user-2", "user", -1*time.Minute)
	if err != nil {
		t.Fatalf("generate expired token: %v", err)
	}

	tests := []struct {
		name         string
		header       string
		wantStatus   int
		wantError    string
		wantUserID   string
		wantRole     string
		expectPassTh bool
	}{
		{
			name:         "TestAuthenticator_ValidToken",
			header:       "Bearer " + validToken,
			wantStatus:   http.StatusOK,
			wantUserID:   "auth-user-1",
			wantRole:     "admin",
			expectPassTh: true,
		},
		{
			name:       "TestAuthenticator_MissingHeader",
			wantStatus: http.StatusUnauthorized,
			wantError:  "missing authorization",
		},
		{
			name:       "TestAuthenticator_MalformedBearer",
			header:     "Bearer",
			wantStatus: http.StatusUnauthorized,
			wantError:  "invalid authorization header",
		},
		{
			name:       "TestAuthenticator_ExpiredToken",
			header:     "Bearer " + expiredToken,
			wantStatus: http.StatusUnauthorized,
			wantError:  "token expired",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Use(Authenticator)
			r.Get("/", func(w http.ResponseWriter, req *http.Request) {
				payload := map[string]string{
					"user_id": UserIDFromContext(req.Context()),
					"role":    RoleFromContext(req.Context()),
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(payload)
			})

			server := testutil.NewTestServer(r)
			defer server.Close()

			req, err := http.NewRequest(http.MethodGet, server.URL+"/", nil)
			if err != nil {
				t.Fatalf("create request: %v", err)
			}
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("execute request: %v", err)
			}
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read response body: %v", err)
			}

			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, tc.wantStatus, bodyBytes)
			}

			if tc.wantError != "" {
				var payload map[string]string
				if err := json.Unmarshal(bodyBytes, &payload); err != nil {
					t.Fatalf("decode JSON error: %v", err)
				}
				if payload["error"] != tc.wantError {
					t.Fatalf("unexpected error message: got %q want %q", payload["error"], tc.wantError)
				}
				return
			}

			if tc.expectPassTh {
				var payload map[string]string
				if err := json.Unmarshal(bodyBytes, &payload); err != nil {
					t.Fatalf("decode JSON body: %v", err)
				}
				if payload["user_id"] != tc.wantUserID {
					t.Fatalf("unexpected user id: got %q want %q", payload["user_id"], tc.wantUserID)
				}
				if payload["role"] != tc.wantRole {
					t.Fatalf("unexpected role: got %q want %q", payload["role"], tc.wantRole)
				}
			}
		})
	}
}

func TestAuthenticator_BearerTakesPriorityOverCookie(t *testing.T) {
	t.Setenv("JWT_SECRET", "auth-middleware-secret")

	bearerToken, err := auth.GenerateToken("bearer-user", "admin", time.Hour)
	if err != nil {
		t.Fatalf("generate bearer token: %v", err)
	}
	cookieToken, err := auth.GenerateToken("cookie-user", "user", time.Hour)
	if err != nil {
		t.Fatalf("generate cookie token: %v", err)
	}

	r := chi.NewRouter()
	r.Use(Authenticator)
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		payload := map[string]string{
			"user_id": UserIDFromContext(req.Context()),
			"role":    RoleFromContext(req.Context()),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(payload)
	})

	server := testutil.NewTestServer(r)
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: cookieToken})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, http.StatusOK, bodyBytes)
	}

	var payload map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["user_id"] != "bearer-user" {
		t.Fatalf("unexpected user id: got %q want %q", payload["user_id"], "bearer-user")
	}
	if payload["role"] != "admin" {
		t.Fatalf("unexpected role: got %q want %q", payload["role"], "admin")
	}
}

func TestAuthenticator_FallsBackToCookie(t *testing.T) {
	t.Setenv("JWT_SECRET", "auth-middleware-secret")

	cookieToken, err := auth.GenerateToken("cookie-user", "user", time.Hour)
	if err != nil {
		t.Fatalf("generate cookie token: %v", err)
	}

	r := chi.NewRouter()
	r.Use(Authenticator)
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		payload := map[string]string{
			"user_id": UserIDFromContext(req.Context()),
			"role":    RoleFromContext(req.Context()),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(payload)
	})

	server := testutil.NewTestServer(r)
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: cookieToken})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, http.StatusOK, bodyBytes)
	}

	var payload map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["user_id"] != "cookie-user" {
		t.Fatalf("unexpected user id: got %q want %q", payload["user_id"], "cookie-user")
	}
}

func TestAuthenticator_InvalidCookieWithNoHeader(t *testing.T) {
	t.Setenv("JWT_SECRET", "auth-middleware-secret")

	r := chi.NewRouter()
	r.Use(Authenticator)
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := testutil.NewTestServer(r)
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: "invalid-token"})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, http.StatusUnauthorized, bodyBytes)
	}

	var payload map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["error"] != "invalid token" {
		t.Fatalf("unexpected error: got %q want %q", payload["error"], "invalid token")
	}
}

func TestAuthenticator_BothMissingReturns401(t *testing.T) {
	t.Setenv("JWT_SECRET", "auth-middleware-secret")

	r := chi.NewRouter()
	r.Use(Authenticator)
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := testutil.NewTestServer(r)
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, http.StatusUnauthorized, bodyBytes)
	}

	var payload map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["error"] != "missing authorization" {
		t.Fatalf("unexpected error: got %q want %q", payload["error"], "missing authorization")
	}
}

func TestRoleRequired(t *testing.T) {
	tests := []struct {
		name       string
		role       string
		required   string
		wantStatus int
		wantError  string
	}{
		{
			name:       "TestRoleRequired_CorrectRole",
			role:       "admin",
			required:   "admin",
			wantStatus: http.StatusOK,
		},
		{
			name:       "TestRoleRequired_WrongRole",
			role:       "user",
			required:   "admin",
			wantStatus: http.StatusForbidden,
			wantError:  "insufficient permissions",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					ctx := WithAuthContext(req.Context(), "role-test-user", tc.role)
					next.ServeHTTP(w, req.WithContext(ctx))
				})
			})
			r.With(RoleRequired(tc.required)).Get("/", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
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

			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("unexpected status: got %d want %d", resp.StatusCode, tc.wantStatus)
			}

			if tc.wantError != "" {
				var payload map[string]string
				if err := json.Unmarshal(bodyBytes, &payload); err != nil {
					t.Fatalf("decode JSON error: %v", err)
				}
				if payload["error"] != tc.wantError {
					t.Fatalf("unexpected error: got %q want %q", payload["error"], tc.wantError)
				}
			}
		})
	}
}
