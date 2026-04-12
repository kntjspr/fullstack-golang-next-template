package router

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/kntjspr/fullstack-golang-next-template/internal/auth"
	"github.com/kntjspr/fullstack-golang-next-template/internal/testutil"
)

type getMeResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

func setupUsersRouteTestServer(t *testing.T) (string, string) {
	t.Helper()

	t.Setenv("JWT_SECRET", "users-route-secret")
	t.Setenv("TEST_DATABASE_URL", "postgres://postgres:test@localhost:5433/testdb?sslmode=disable")
	t.Setenv("TEST_REDIS_URL", "redis://localhost:6380")

	db := testutil.SetupTestDB(t)
	t.Cleanup(func() {
		testutil.TeardownTestDB(t, db)
	})

	user := testutil.CreateTestUser(t, db, map[string]any{
		"role": "user",
	})

	r := chi.NewRouter()
	GetRoutes(r, nil, nil, db)

	server := testutil.NewTestServer(r)
	t.Cleanup(server.Close)

	return server.URL, user.ID
}

func TestGetMe_Authenticated(t *testing.T) {
	baseURL, userID := setupUsersRouteTestServer(t)

	token, err := auth.GenerateToken(userID, "user")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, baseURL+"/users/me", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		rawBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, http.StatusOK, rawBody)
	}

	var payload getMeResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.ID == "" || payload.Email == "" || payload.Role == "" || payload.CreatedAt == "" {
		t.Fatalf("expected non-empty profile fields, got %+v", payload)
	}
}

func TestGetMe_Unauthenticated(t *testing.T) {
	baseURL, _ := setupUsersRouteTestServer(t)

	req, err := http.NewRequest(http.MethodGet, baseURL+"/users/me", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		rawBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, http.StatusUnauthorized, rawBody)
	}
}

func TestGetMe_InvalidToken(t *testing.T) {
	baseURL, _ := setupUsersRouteTestServer(t)

	req, err := http.NewRequest(http.MethodGet, baseURL+"/users/me", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer malformed-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		rawBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, http.StatusUnauthorized, rawBody)
	}
}
