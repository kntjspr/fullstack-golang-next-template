package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"github.com/tidwall/gjson"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/create-go-app/chi-go-template/internal/models"
)

// SetupTestDB creates a PostgreSQL test database connection and migrates all models.
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("TEST_DATABASE_URL is required")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test postgres database: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("automigrate test models: %v", err)
	}

	return db
}

// SetupTestRedis creates a redis client for integration tests.
func SetupTestRedis(t *testing.T) *redis.Client {
	t.Helper()

	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		t.Fatal("TEST_REDIS_URL is required")
	}

	options, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("parse TEST_REDIS_URL: %v", err)
	}

	client := redis.NewClient(options)
	if err := client.Ping(t.Context()).Err(); err != nil {
		t.Fatalf("ping test redis: %v", err)
	}

	return client
}

// NewTestServer wraps the provided router in a real HTTP test server.
func NewTestServer(r chi.Router) *httptest.Server {
	return httptest.NewServer(r)
}

// MustJSON marshals v to JSON or fails the test.
func MustJSON(t *testing.T, v any) []byte {
	t.Helper()

	payload, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}

	return payload
}

// AssertStatus checks that the response status code matches expected.
func AssertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()

	if resp.StatusCode != expected {
		t.Fatalf("unexpected status: got %d, want %d", resp.StatusCode, expected)
	}
}

// AssertJSONField checks a JSONPath-like gjson path value.
func AssertJSONField(t *testing.T, body []byte, path string, expected any) {
	t.Helper()

	result := gjson.GetBytes(body, path)
	if !result.Exists() {
		t.Fatalf("JSON path %q not found", path)
	}

	actualJSON, err := json.Marshal(result.Value())
	if err != nil {
		t.Fatalf("marshal actual JSON field %q: %v", path, err)
	}

	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("marshal expected JSON field %q: %v", path, err)
	}

	if string(actualJSON) != string(expectedJSON) {
		t.Fatalf("JSON path %q mismatch: got %s, want %s", path, actualJSON, expectedJSON)
	}
}

// TeardownTestDB closes database resources.
func TeardownTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql.DB from gorm DB: %v", err)
	}

	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close test DB: %v", err)
	}
}
