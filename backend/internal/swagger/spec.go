package swagger

import (
	"bytes"
	_ "embed"
	"fmt"
	"html"
	"net/http"
	"strings"
)

var (
	//go:embed openapi.yaml
	openAPISpec  []byte
	openAPITitle = extractOpenAPITitle(openAPISpec)
)

// SpecHandler serves OpenAPI specification for Swagger UI.
func SpecHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openAPISpec)
}

// UIHandler serves a lightweight Redoc page for visual API docs.
func UIHandler(w http.ResponseWriter, _ *http.Request) {
	// relax script-src for this route only — redoc bundle loads from cdn.jsdelivr.net
	w.Header().Set("Content-Security-Policy",
		"default-src 'self'; script-src 'self' https://cdn.jsdelivr.net; "+
			"style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' https://fonts.gstatic.com; "+
			"worker-src blob:")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(redocHTML(openAPITitle)))
}

func redocHTML(title string) string {
	escapedTitle := html.EscapeString(strings.TrimSpace(title))
	if escapedTitle == "" {
		escapedTitle = "API Docs"
	}

	return fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>%s</title>
  <style>
    body { margin: 0; padding: 0; }
  </style>
</head>
<body>
  <redoc spec-url="/swagger/spec"></redoc>
  <script src="https://cdn.jsdelivr.net/npm/redoc/bundles/redoc.standalone.js"></script>
</body>
</html>`, escapedTitle)
}

func extractOpenAPITitle(spec []byte) string {
	inInfo := false

	for _, rawLine := range bytes.Split(spec, []byte("\n")) {
		line := string(rawLine)
		trimmed := strings.TrimSpace(line)

		if trimmed == "info:" {
			inInfo = true
			continue
		}

		if inInfo && line != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			inInfo = false
		}

		if !inInfo || !strings.HasPrefix(trimmed, "title:") {
			continue
		}

		title := strings.TrimSpace(strings.TrimPrefix(trimmed, "title:"))
		title = strings.Trim(title, `"'`)
		if title != "" {
			return title
		}
	}

	return "API Docs"
}
