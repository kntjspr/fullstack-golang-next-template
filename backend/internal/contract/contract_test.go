package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"

	"github.com/kntjspr/fullstack-golang-next-template/internal/auth"
	"github.com/kntjspr/fullstack-golang-next-template/internal/router"
	"github.com/kntjspr/fullstack-golang-next-template/internal/testutil"
)

func TestOpenAPIContract(t *testing.T) {
	specPath := contractSpecPath(t)
	t.Setenv("JWT_SECRET", "contract-secret")
	t.Setenv("TEST_DATABASE_URL", "postgres://postgres:test@localhost:5433/testdb?sslmode=disable")

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		t.Fatalf("load OpenAPI spec: %v", err)
	}
	if err := doc.Validate(loader.Context); err != nil {
		t.Fatalf("validate OpenAPI spec: %v", err)
	}

	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	testUser := testutil.CreateTestUser(t, db, map[string]any{
		"password": "contract-password",
		"role":     "user",
	})
	authToken, err := auth.GenerateToken(testUser.ID, testUser.Role)
	if err != nil {
		t.Fatalf("generate auth token: %v", err)
	}

	r := chi.NewRouter()
	router.GetRoutes(r, nil, nil, db)

	routerOps, err := collectRouterOperations(r)
	if err != nil {
		t.Fatalf("collect router operations: %v", err)
	}

	specOps := collectSpecOperations(doc)
	assertRouterCoverage(t, routerOps, specOps)

	server := testutil.NewTestServer(r)
	defer server.Close()

	for _, op := range sortedSpecOperations(doc) {
		t.Run(op.key, func(t *testing.T) {
			expectedStatus, responseRef, err := expected2xxResponse(op.operation)
			if err != nil {
				t.Fatalf("resolve documented 2xx response: %v", err)
			}

			requestBody, err := requestExampleBody(op.operation)
			if err != nil {
				t.Fatalf("resolve request example body: %v", err)
			}
			if op.path == "/auth/login" {
				requestBody = testutil.MustJSON(t, map[string]string{
					"email":    testUser.Email,
					"password": "contract-password",
				})
			}

			requestPath := operationRequestPath(op.path, op.pathItem, op.operation)
			req, err := http.NewRequest(op.method, server.URL+requestPath, bytes.NewReader(requestBody))
			if err != nil {
				t.Fatalf("build request: %v", err)
			}
			if len(requestBody) > 0 {
				req.Header.Set("Content-Type", "application/json")
			}
			if operationRequiresAuth(op.operation) {
				req.Header.Set("Authorization", "Bearer "+authToken)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("perform request: %v", err)
			}
			defer resp.Body.Close()

			testutil.AssertStatus(t, resp, expectedStatus)

			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read response body: %v", err)
			}

			if responseRef == nil || responseRef.Value == nil {
				return
			}

			if err := validateResponseBody(responseRef.Value, responseBody); err != nil {
				t.Fatalf("response schema validation failed: %v", err)
			}
		})
	}
}

func TestAllRoutesDocumented(t *testing.T) {
	specPath := contractSpecPath(t)

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		t.Fatalf("load OpenAPI spec: %v", err)
	}
	if err := doc.Validate(loader.Context); err != nil {
		t.Fatalf("validate OpenAPI spec: %v", err)
	}

	r := chi.NewRouter()
	router.GetRoutes(r, nil, nil, nil)

	routerPaths, err := collectRouterPaths(r)
	if err != nil {
		t.Fatalf("collect router paths: %v", err)
	}

	specPaths := collectSpecPaths(doc)
	undocumented := missingKeys(routerPaths, specPaths)
	if len(undocumented) > 0 {
		t.Fatalf("undocumented routes: %s", strings.Join(undocumented, ", "))
	}
}

func TestAllDocumentedRoutesExist(t *testing.T) {
	specPath := contractSpecPath(t)

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		t.Fatalf("load OpenAPI spec: %v", err)
	}
	if err := doc.Validate(loader.Context); err != nil {
		t.Fatalf("validate OpenAPI spec: %v", err)
	}

	r := chi.NewRouter()
	router.GetRoutes(r, nil, nil, nil)

	routerPaths, err := collectRouterPaths(r)
	if err != nil {
		t.Fatalf("collect router paths: %v", err)
	}

	specPaths := collectSpecPaths(doc)
	deadSpecPaths := missingKeys(specPaths, routerPaths)
	if len(deadSpecPaths) > 0 {
		t.Fatalf("dead spec paths: %s", strings.Join(deadSpecPaths, ", "))
	}
}

func operationRequiresAuth(operation *openapi3.Operation) bool {
	if operation == nil || operation.Security == nil {
		return false
	}

	return len(*operation.Security) > 0
}

func contractSpecPath(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine contract test file path")
	}

	return filepath.Join(filepath.Dir(file), "..", "swagger", "openapi.yaml")
}

func collectRouterPaths(r chi.Router) (map[string]struct{}, error) {
	paths := make(map[string]struct{})

	err := chi.Walk(r, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		method = strings.ToUpper(method)
		if method == http.MethodHead || method == http.MethodOptions {
			return nil
		}

		normalizedRoute := normalizeRoute(route)
		if normalizedRoute == "" {
			return nil
		}

		paths[normalizedRoute] = struct{}{}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return paths, nil
}

func collectRouterOperations(r chi.Router) (map[string]struct{}, error) {
	operations := make(map[string]struct{})

	err := chi.Walk(r, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		method = strings.ToUpper(method)
		if method == http.MethodHead || method == http.MethodOptions {
			return nil
		}

		normalizedRoute := normalizeRoute(route)
		operations[fmt.Sprintf("%s %s", method, normalizedRoute)] = struct{}{}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return operations, nil
}

func normalizeRoute(route string) string {
	if strings.Contains(route, "/*") {
		return strings.Replace(route, "/*", "/{path}", 1)
	}

	return route
}

func collectSpecOperations(doc *openapi3.T) map[string]struct{} {
	operations := make(map[string]struct{})
	for path, pathItem := range doc.Paths.Map() {
		for method := range operationsForPathItem(pathItem) {
			operations[fmt.Sprintf("%s %s", method, path)] = struct{}{}
		}
	}

	return operations
}

func collectSpecPaths(doc *openapi3.T) map[string]struct{} {
	paths := make(map[string]struct{})
	for path := range doc.Paths.Map() {
		paths[path] = struct{}{}
	}

	return paths
}

func missingKeys(left, right map[string]struct{}) []string {
	missing := make([]string, 0)
	for key := range left {
		if _, ok := right[key]; !ok {
			missing = append(missing, key)
		}
	}

	sort.Strings(missing)
	return missing
}

func assertRouterCoverage(t *testing.T, routerOps, specOps map[string]struct{}) {
	t.Helper()

	missing := make([]string, 0)
	for op := range routerOps {
		if _, ok := specOps[op]; !ok {
			missing = append(missing, op)
		}
	}

	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("router operations missing from OpenAPI spec: %s", strings.Join(missing, ", "))
	}
}

type specOperation struct {
	key       string
	method    string
	path      string
	pathItem  *openapi3.PathItem
	operation *openapi3.Operation
}

func sortedSpecOperations(doc *openapi3.T) []specOperation {
	ops := make([]specOperation, 0)
	for path, pathItem := range doc.Paths.Map() {
		for method, operation := range operationsForPathItem(pathItem) {
			ops = append(ops, specOperation{
				key:       fmt.Sprintf("%s %s", method, path),
				method:    method,
				path:      path,
				pathItem:  pathItem,
				operation: operation,
			})
		}
	}

	sort.Slice(ops, func(i, j int) bool {
		return ops[i].key < ops[j].key
	})

	return ops
}

func operationsForPathItem(pathItem *openapi3.PathItem) map[string]*openapi3.Operation {
	operations := map[string]*openapi3.Operation{}
	if pathItem.Get != nil {
		operations[http.MethodGet] = pathItem.Get
	}
	if pathItem.Post != nil {
		operations[http.MethodPost] = pathItem.Post
	}
	if pathItem.Put != nil {
		operations[http.MethodPut] = pathItem.Put
	}
	if pathItem.Patch != nil {
		operations[http.MethodPatch] = pathItem.Patch
	}
	if pathItem.Delete != nil {
		operations[http.MethodDelete] = pathItem.Delete
	}

	return operations
}

func expected2xxResponse(operation *openapi3.Operation) (int, *openapi3.ResponseRef, error) {
	type candidate struct {
		code int
		resp *openapi3.ResponseRef
	}

	candidates := make([]candidate, 0)
	for statusCode, resp := range operation.Responses.Map() {
		code, err := strconv.Atoi(statusCode)
		if err != nil {
			continue
		}
		if code >= 200 && code < 300 {
			candidates = append(candidates, candidate{code: code, resp: resp})
		}
	}

	if len(candidates) == 0 {
		return 0, nil, fmt.Errorf("operation %q has no documented 2xx response", operation.OperationID)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].code < candidates[j].code
	})

	return candidates[0].code, candidates[0].resp, nil
}

func requestExampleBody(operation *openapi3.Operation) ([]byte, error) {
	if operation.RequestBody == nil || operation.RequestBody.Value == nil {
		return nil, nil
	}

	mediaType := operation.RequestBody.Value.Content.Get("application/json")
	if mediaType == nil {
		return nil, nil
	}

	if mediaType.Example != nil {
		return json.Marshal(mediaType.Example)
	}

	for _, ex := range mediaType.Examples {
		if ex != nil && ex.Value != nil && ex.Value.Value != nil {
			return json.Marshal(ex.Value.Value)
		}
	}

	if mediaType.Schema != nil && mediaType.Schema.Value != nil && mediaType.Schema.Value.Example != nil {
		return json.Marshal(mediaType.Schema.Value.Example)
	}

	return nil, nil
}

func operationRequestPath(rawPath string, pathItem *openapi3.PathItem, operation *openapi3.Operation) string {
	params := make([]*openapi3.ParameterRef, 0, len(pathItem.Parameters)+len(operation.Parameters))
	params = append(params, pathItem.Parameters...)
	params = append(params, operation.Parameters...)

	resolvedPath := rawPath
	for _, paramRef := range params {
		if paramRef == nil || paramRef.Value == nil || paramRef.Value.In != "path" {
			continue
		}

		value := "example"
		if paramRef.Value.Example != nil {
			value = fmt.Sprint(paramRef.Value.Example)
		} else {
			for _, ex := range paramRef.Value.Examples {
				if ex != nil && ex.Value != nil && ex.Value.Value != nil {
					value = fmt.Sprint(ex.Value.Value)
					break
				}
			}
		}

		resolvedPath = strings.ReplaceAll(resolvedPath, fmt.Sprintf("{%s}", paramRef.Value.Name), value)
	}

	return resolvedPath
}

func validateResponseBody(response *openapi3.Response, body []byte) error {
	if response.Content == nil || len(response.Content) == 0 {
		return nil
	}

	mediaTypeName, mediaType := preferredMediaType(response.Content)
	if mediaType == nil || mediaType.Schema == nil || mediaType.Schema.Value == nil {
		return nil
	}

	var value any
	if strings.Contains(mediaTypeName, "json") {
		if len(bytes.TrimSpace(body)) == 0 {
			value = map[string]any{}
		} else if err := json.Unmarshal(body, &value); err != nil {
			return fmt.Errorf("decode JSON response: %w", err)
		}
	} else {
		value = string(body)
	}

	if err := mediaType.Schema.Value.VisitJSON(value); err != nil {
		return fmt.Errorf("schema mismatch: %w", err)
	}

	return nil
}

func preferredMediaType(content openapi3.Content) (string, *openapi3.MediaType) {
	if mt := content.Get("application/json"); mt != nil {
		return "application/json", mt
	}
	if mt := content.Get("application/yaml"); mt != nil {
		return "application/yaml", mt
	}
	if mt := content.Get("text/html"); mt != nil {
		return "text/html", mt
	}

	names := make([]string, 0, len(content))
	for name := range content {
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) == 0 {
		return "", nil
	}

	return names[0], content[names[0]]
}
