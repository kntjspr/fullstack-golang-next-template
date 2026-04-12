package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/kntjspr/fullstack-golang-next-template/internal/testutil"
)

type validateTestRequest struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required,min=2"`
	Age   int    `json:"age" validate:"required,gte=0"`
}

type validateErrorResponse struct {
	Error  string `json:"error"`
	Fields []struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	} `json:"fields"`
}

func TestValidateBody_Valid(t *testing.T) {
	handler := newValidationTestHandler(t)
	server := testutil.NewTestServer(handler)
	defer server.Close()

	resp, body := doJSONRequest(t, server.URL+"/", map[string]any{
		"email": "valid@example.com",
		"name":  "Alice",
		"age":   30,
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", resp.StatusCode, http.StatusOK)
	}

	var payload validateTestRequest
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Email != "valid@example.com" || payload.Name != "Alice" || payload.Age != 30 {
		t.Fatalf("unexpected validated payload: %+v", payload)
	}
}

func TestValidateBody_MissingRequired(t *testing.T) {
	handler := newValidationTestHandler(t)
	server := testutil.NewTestServer(handler)
	defer server.Close()

	resp, body := doJSONRequest(t, server.URL+"/", map[string]any{
		"email": "valid@example.com",
		"age":   30,
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status: got %d want %d", resp.StatusCode, http.StatusUnprocessableEntity)
	}

	assertValidationErrorContainsField(t, body, "name")
}

func TestValidateBody_TypeMismatch(t *testing.T) {
	handler := newValidationTestHandler(t)
	server := testutil.NewTestServer(handler)
	defer server.Close()

	resp, body := doRawJSONRequest(t, server.URL+"/", `{"email":"valid@example.com","name":"Alice","age":"thirty"}`)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, http.StatusUnprocessableEntity, body)
	}

	assertValidationErrorContainsField(t, body, "body")
}

func TestValidateBody_SQLInjectionString(t *testing.T) {
	handler := newValidationTestHandler(t)
	server := testutil.NewTestServer(handler)
	defer server.Close()

	resp, body := doJSONRequest(t, server.URL+"/", map[string]any{
		"email": "valid@example.com",
		"name":  "'; DROP TABLE users; --",
		"age":   30,
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, http.StatusOK, body)
	}

	var payload validateTestRequest
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Name != "'; DROP TABLE users; --" {
		t.Fatalf("expected SQL-like string to pass through unchanged, got %q", payload.Name)
	}
}

func TestValidateBody_XSSPayload(t *testing.T) {
	handler := newValidationTestHandler(t)
	server := testutil.NewTestServer(handler)
	defer server.Close()

	resp, body := doJSONRequest(t, server.URL+"/", map[string]any{
		"email": "valid@example.com",
		"name":  "<script>alert(1)</script>",
		"age":   30,
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, http.StatusOK, body)
	}

	var payload validateTestRequest
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Name != "<script>alert(1)</script>" {
		t.Fatalf("expected XSS-like payload to pass through unchanged, got %q", payload.Name)
	}
}

func TestValidateBody_EmptyBody(t *testing.T) {
	handler := newValidationTestHandler(t)
	server := testutil.NewTestServer(handler)
	defer server.Close()

	resp, body := doRawRequest(t, server.URL+"/", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, http.StatusUnprocessableEntity, body)
	}

	assertValidationErrorContainsField(t, body, "body")
}

func TestValidateBody_OversizedBody(t *testing.T) {
	handler := newValidationTestHandler(t)
	server := testutil.NewTestServer(handler)
	defer server.Close()

	oversized := strings.Repeat("a", maxValidateBodyBytes+1)
	resp, body := doRawJSONRequest(
		t,
		server.URL+"/",
		`{"email":"valid@example.com","name":"Alice","age":30,"blob":"`+oversized+`"}`,
	)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, http.StatusRequestEntityTooLarge, body)
	}
}

func newValidationTestHandler(t *testing.T) chi.Router {
	t.Helper()

	r := chi.NewRouter()
	r.Use(ValidateBody(validateTestRequest{}))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		payload := GetValidatedBody[validateTestRequest](req.Context())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	})
	return r
}

func doJSONRequest(t *testing.T, url string, body map[string]any) (*http.Response, []byte) {
	t.Helper()

	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}
	return doRawRequest(t, url, encoded)
}

func doRawJSONRequest(t *testing.T, url string, raw string) (*http.Response, []byte) {
	t.Helper()
	return doRawRequest(t, url, []byte(raw))
}

func doRawRequest(t *testing.T, url string, body []byte) (*http.Response, []byte) {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		t.Fatalf("read response body: %v", err)
	}

	return resp, responseBody
}

func assertValidationErrorContainsField(t *testing.T, body []byte, field string) {
	t.Helper()

	var payload validateErrorResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode validation error response: %v body=%s", err, body)
	}

	if payload.Error != "validation failed" {
		t.Fatalf("unexpected error: got %q want %q", payload.Error, "validation failed")
	}

	for _, item := range payload.Fields {
		if item.Field == field {
			return
		}
	}

	t.Fatalf("expected field %q in validation errors, got %+v", field, payload.Fields)
}
