package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/kntjspr/fullstack-golang-next-template/internal/auth"
	"github.com/kntjspr/fullstack-golang-next-template/internal/testutil"
)

func setupAuthTestServer(t *testing.T, appEnv string) (baseURL, email, password string) {
	t.Helper()

	t.Setenv("JWT_SECRET", "router-auth-secret")
	t.Setenv("TEST_DATABASE_URL", "postgres://postgres:test@localhost:5433/testdb?sslmode=disable")
	t.Setenv("TEST_REDIS_URL", "redis://localhost:6380")
	if appEnv != "" {
		t.Setenv("STAGE_STATUS", appEnv)
	}

	db := testutil.SetupTestDB(t)
	t.Cleanup(func() {
		testutil.TeardownTestDB(t, db)
	})

	now := time.Now().UTC().UnixNano()
	pass := "correct-password"
	user := testutil.CreateTestUser(t, db, map[string]any{
		"id":       fmt.Sprintf("router-user-%d", now),
		"email":    fmt.Sprintf("router-user-%d@example.com", now),
		"password": pass,
		"role":     "user",
	})

	r := chi.NewRouter()
	RegisterAuthRoutes(r, db)
	server := testutil.NewTestServer(r)
	t.Cleanup(server.Close)

	return server.URL, user.Email, pass
}

func TestAuthRoutes(t *testing.T) {
	t.Setenv("JWT_SECRET", "router-auth-secret")
	t.Setenv("TEST_DATABASE_URL", "postgres://postgres:test@localhost:5433/testdb?sslmode=disable")
	t.Setenv("TEST_REDIS_URL", "redis://localhost:6380")

	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	now := time.Now().UTC().UnixNano()
	testUser := testutil.CreateTestUser(t, db, map[string]any{
		"id":       fmt.Sprintf("router-user-%d", now),
		"email":    fmt.Sprintf("router-user-%d@example.com", now),
		"password": "correct-password",
		"role":     "user",
	})

	tests := []struct {
		name         string
		method       string
		path         string
		body         any
		headers      map[string]string
		wantStatus   int
		wantError    string
		assertFunc   func(t *testing.T, responseBody []byte)
		sameAsStatus string
	}{
		{
			name:       "TestLoginSuccess",
			method:     http.MethodPost,
			path:       "/auth/login",
			wantStatus: http.StatusOK,
			body: map[string]string{
				"email":    testUser.Email,
				"password": "correct-password",
			},
			assertFunc: func(t *testing.T, responseBody []byte) {
				t.Helper()
				var payload map[string]string
				if err := json.Unmarshal(responseBody, &payload); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if payload["token"] == "" {
					t.Fatal("token is missing in login response")
				}
				if payload["expires_at"] == "" {
					t.Fatal("expires_at is missing in login response")
				}
			},
		},
		{
			name:       "TestLoginInvalidPassword",
			method:     http.MethodPost,
			path:       "/auth/login",
			wantStatus: http.StatusUnauthorized,
			wantError:  "invalid credentials",
			body: map[string]string{
				"email":    testUser.Email,
				"password": "wrong-password",
			},
		},
		{
			name:       "TestLoginUserNotFound",
			method:     http.MethodPost,
			path:       "/auth/login",
			wantStatus: http.StatusUnauthorized,
			wantError:  "invalid credentials",
			body: map[string]string{
				"email":    "missing@example.com",
				"password": "whatever",
			},
		},
		{
			name:       "TestLoginMissingFields",
			method:     http.MethodPost,
			path:       "/auth/login",
			wantStatus: http.StatusUnprocessableEntity,
			body:       map[string]string{},
			assertFunc: func(t *testing.T, responseBody []byte) {
				t.Helper()
				var payload map[string][]string
				if err := json.Unmarshal(responseBody, &payload); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if len(payload["errors"]) == 0 {
					t.Fatalf("expected validation errors, got %s", responseBody)
				}
			},
		},
		{
			name:       "TestRefreshValid",
			method:     http.MethodPost,
			path:       "/auth/refresh",
			wantStatus: http.StatusOK,
			assertFunc: func(t *testing.T, responseBody []byte) {
				t.Helper()

				var payload map[string]string
				if err := json.Unmarshal(responseBody, &payload); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if payload["token"] == "" {
					t.Fatal("refresh token is missing")
				}
				if payload["expires_at"] == "" {
					t.Fatal("refresh expires_at is missing")
				}
			},
		},
		{
			name:       "TestRefreshExpired",
			method:     http.MethodPost,
			path:       "/auth/refresh",
			wantStatus: http.StatusUnauthorized,
			wantError:  "token expired",
		},
	}

	loginToken, err := auth.GenerateToken(testUser.ID, testUser.Role, time.Hour)
	if err != nil {
		t.Fatalf("generate login token: %v", err)
	}
	loginClaims, err := auth.ValidateToken(loginToken)
	if err != nil {
		t.Fatalf("validate login token: %v", err)
	}

	expiredToken, err := auth.GenerateToken(testUser.ID, testUser.Role, -1*time.Minute)
	if err != nil {
		t.Fatalf("generate expired token: %v", err)
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := chi.NewRouter()
			RegisterAuthRoutes(r, db)
			server := testutil.NewTestServer(r)
			defer server.Close()

			var requestBody io.Reader
			if tc.body != nil {
				encoded, err := json.Marshal(tc.body)
				if err != nil {
					t.Fatalf("marshal body: %v", err)
				}
				requestBody = bytes.NewReader(encoded)
			}

			req, err := http.NewRequest(tc.method, server.URL+tc.path, requestBody)
			if err != nil {
				t.Fatalf("create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			switch tc.name {
			case "TestRefreshValid":
				req.Header.Set("Authorization", "Bearer "+loginToken)
			case "TestRefreshExpired":
				req.Header.Set("Authorization", "Bearer "+expiredToken)
			}

			for key, value := range tc.headers {
				req.Header.Set(key, value)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("execute request: %v", err)
			}
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read response: %v", err)
			}

			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, tc.wantStatus, bodyBytes)
			}

			if tc.wantError != "" {
				var payload map[string]string
				if err := json.Unmarshal(bodyBytes, &payload); err != nil {
					t.Fatalf("decode error response: %v", err)
				}
				if payload["error"] != tc.wantError {
					t.Fatalf("unexpected error: got %q want %q", payload["error"], tc.wantError)
				}
				return
			}

			if tc.name == "TestRefreshValid" {
				var payload map[string]string
				if err := json.Unmarshal(bodyBytes, &payload); err != nil {
					t.Fatalf("decode refresh response: %v", err)
				}
				refreshedClaims, err := auth.ValidateToken(payload["token"])
				if err != nil {
					t.Fatalf("validate refreshed token: %v", err)
				}
				if !refreshedClaims.ExpiresAt.After(loginClaims.ExpiresAt) {
					t.Fatalf("refreshed token expiry should be later than login token expiry")
				}
			}

			if tc.assertFunc != nil {
				tc.assertFunc(t, bodyBytes)
			}
		})
	}
}

func TestLoginSetsHTTPOnlyCookie(t *testing.T) {
	baseURL, email, password := setupAuthTestServer(t, "development")

	body, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/auth/login", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		rawBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, http.StatusOK, rawBody)
	}

	var authCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "auth_token" {
			authCookie = c
			break
		}
	}

	if authCookie == nil {
		t.Fatal("expected auth_token cookie to be set")
	}
	if !authCookie.HttpOnly {
		t.Fatal("expected auth_token cookie to be HttpOnly")
	}
}

func TestLoginCookieSecureFlagInProd(t *testing.T) {
	baseURL, email, password := setupAuthTestServer(t, "prod")

	body, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/auth/login", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		rawBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, http.StatusOK, rawBody)
	}

	var authCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "auth_token" {
			authCookie = c
			break
		}
	}
	if authCookie == nil {
		t.Fatal("expected auth_token cookie to be set")
	}
	if !authCookie.Secure {
		t.Fatal("expected auth_token cookie to be Secure in production")
	}
}

func TestLoginCookieSecureFlagAbsentInDev(t *testing.T) {
	baseURL, email, password := setupAuthTestServer(t, "development")

	body, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/auth/login", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		rawBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, http.StatusOK, rawBody)
	}

	var authCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "auth_token" {
			authCookie = c
			break
		}
	}
	if authCookie == nil {
		t.Fatal("expected auth_token cookie to be set")
	}
	if authCookie.Secure {
		t.Fatal("expected auth_token cookie to be non-Secure in development")
	}
}

func TestLogoutClearsCookie(t *testing.T) {
	baseURL, _, _ := setupAuthTestServer(t, "development")

	req, err := http.NewRequest(http.MethodPost, baseURL+"/auth/logout", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		rawBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, http.StatusOK, rawBody)
	}

	var authCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "auth_token" {
			authCookie = c
			break
		}
	}
	if authCookie == nil {
		t.Fatal("expected auth_token cookie to be cleared")
	}
	if authCookie.MaxAge != -1 {
		t.Fatalf("expected cleared cookie MaxAge -1, got %d", authCookie.MaxAge)
	}
}

func TestLogoutReturns200WithoutAuth(t *testing.T) {
	baseURL, _, _ := setupAuthTestServer(t, "development")

	req, err := http.NewRequest(http.MethodPost, baseURL+"/auth/logout", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, http.StatusOK, rawBody)
	}

	if !strings.Contains(string(rawBody), "logged out") {
		t.Fatalf("unexpected response body: %s", rawBody)
	}
}

func TestRefreshTokenFromCookie(t *testing.T) {
	t.Setenv("JWT_SECRET", "router-auth-secret")
	t.Setenv("TEST_DATABASE_URL", "postgres://postgres:test@localhost:5433/testdb?sslmode=disable")
	t.Setenv("TEST_REDIS_URL", "redis://localhost:6380")
	t.Setenv("STAGE_STATUS", "dev")

	db := testutil.SetupTestDB(t)
	t.Cleanup(func() {
		testutil.TeardownTestDB(t, db)
	})

	user := testutil.CreateTestUser(t, db, map[string]any{
		"password": "correct-password",
		"role":     "user",
	})

	loginToken, err := auth.GenerateToken(user.ID, user.Role, time.Hour)
	if err != nil {
		t.Fatalf("generate login token: %v", err)
	}

	r := chi.NewRouter()
	RegisterAuthRoutes(r, db)
	server := testutil.NewTestServer(r)
	t.Cleanup(server.Close)

	req, err := http.NewRequest(http.MethodPost, server.URL+"/auth/refresh", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: loginToken})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, http.StatusOK, rawBody)
	}

	var payload map[string]string
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if payload["token"] == "" {
		t.Fatal("expected refreshed token in response")
	}
}
