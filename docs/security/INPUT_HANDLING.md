# Input Handling

## Request validation middleware (`middleware.ValidateBody`)

The request validation middleware performs two checks for JSON request bodies:

1. Type validation via `encoding/json` decoding into a typed Go struct.
2. Constraint validation via `go-playground/validator` tags (for example `required`, `email`, `min`, `gte`).

Behavior summary:

- Valid payloads are stored in request context and retrieved with `middleware.GetValidatedBody[T](ctx)`.
- Missing, malformed, or constraint-invalid payloads return `422 Unprocessable Entity` with:
  - `error: "validation failed"`
  - `fields: [{"field": "...", "message": "..."}]`
- Oversized payloads (> 1MB) return `413 Request Entity Too Large`.

This middleware validates structure and constraints only.

## Sanitization helpers (`internal/sanitize`)

Sanitization is opt-in and explicit at call sites.

- `sanitize.StripHTML(s string)` removes script blocks and HTML tags from text.
- `sanitize.TruncateString(s string, max int)` truncates strings by rune count (Unicode-safe).

These helpers are utilities, not globally enforced middleware.

## What this layer does not do

### SQL injection prevention is not handled by sanitizing strings

The API layer does not sanitize input specifically for SQL injection.

Instead, SQL injection prevention is achieved by using parameterized queries through GORM (for example `Where("email = ?", email)`), where values are bound as parameters rather than string-concatenated SQL.

This is the correct defense at the data access layer and avoids brittle, lossy input rewriting.

### Output encoding is a frontend responsibility

This backend accepts and validates data as typed values. It does not perform output escaping for every rendering context.

When displaying user-generated content in the UI, the frontend must apply context-appropriate output encoding/escaping (HTML, attribute, URL, JavaScript context, etc.) to prevent XSS.
