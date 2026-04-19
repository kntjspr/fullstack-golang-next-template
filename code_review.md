# Code Review — fullstack-golang-next-template

**Reviewer mood:** ☕ ran out. Let's get into it.

---

## CRITICAL — Ship-blockers

### 1. Plaintext password comparison — this is inexcusable

[auth.go:L76](file:///home/xo/temp/go-project/backend/internal/router/auth.go#L76)

```go
if user.PasswordHash != payload.Password {
```

The field is literally called `PasswordHash` and you're doing a **raw string equality check** against the plaintext password from the request body. There is no bcrypt, no argon2, no scrypt — nothing. The password is stored and compared **in cleartext**. This means:

- Every password in the database is readable to anyone with DB access.
- A SQL dump leak = every user account is compromised instantly.
- The field name `PasswordHash` is a lie. It's `PasswordPlaintext`.

The test fixtures also reinforce this anti-pattern — `password_hash: "correct-password"` is stored verbatim. This isn't a shortcut, it's a liability.

> [!CAUTION]
> This is a textbook CWE-256 (Plaintext Storage of a Password). If this ships to production, you deserve everything that happens next.

---

### 2. Hardcoded JWT secret fallback in production code

[token.go:L12](file:///home/xo/temp/go-project/backend/internal/auth/token.go#L12)

```go
const defaultJWTSecret = "test-secret"
```

[token.go:L107-L114](file:///home/xo/temp/go-project/backend/internal/auth/token.go#L107-L114)

```go
func signingSecret() string {
    secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
    if secret == "" {
        return defaultJWTSecret
    }
    return secret
}
```

If `JWT_SECRET` is unset or empty, **every token in the system is signed with `"test-secret"`**. There is no startup validation, no panic, no warning. The app just silently runs with a known secret. Any attacker who reads your source code (it's on GitHub, by the way) can forge arbitrary JWTs.

This should `log.Fatal()` if `JWT_SECRET` is empty in production. Falling back to a hardcoded secret is the kind of thing that makes security researchers write blog posts about you.

---

### 3. Double expiry check in token validation — one of them is redundant (and confusing)

[token.go:L66-L97](file:///home/xo/temp/go-project/backend/internal/auth/token.go#L66-L97)

The `jwt.Parse` call already checks `exp` and returns `jwt.ErrTokenExpired`. Then you manually check it again:

```go
if time.Now().UTC().After(expiresAt) {
    return nil, ErrTokenExpired
}
```

This means you have two competing expiry checks with potentially different clock reads. If the token is right on the boundary, the library might say "valid" and your manual check might say "expired" (or vice versa, depending on clock skew within the same goroutine). Pick one. The library already does this correctly.

---

### 4. `ValidateBody` middleware creates a **new validator instance per request wrapper**, not per request

[validate.go:L33-L34](file:///home/xo/temp/go-project/backend/middleware/validate.go#L33-L34)

```go
func ValidateBody[T any](schema T) func(http.Handler) http.Handler {
    validate := validator.New()
```

Wait — actually, this creates one validator per middleware instance (per route registration), which is fine. But the **schema is passed by value** on line 40:

```go
payload := schema
```

For struct types this copies the zero-value schema each time, which works. But if `T` is a pointer type, every request mutates the **same object**. The generic constraint doesn't prevent `T` from being `*SomeStruct`. This is a race condition waiting to happen under concurrent load.

---

### 5. The `.env` file is committed to the repository

[.env](file:///home/xo/temp/go-project/backend/.env)

The actual `.env` file is tracked in git with default credentials:

```
DB_PASSWORD="password"
```

Yes, `.env.example` also exists with the same content. But `.env` should be in `.gitignore`. The [.gitignore](file:///home/xo/temp/go-project/backend/.gitignore) likely has it, but the file is still tracked. If someone puts real credentials in there and commits, the secret is in git history forever.

Also the root `.env` at [.env](file:///home/xo/temp/go-project/.env) — same problem.

---

## HIGH — Needs fixing before merge

### 6. Package-level mutable singleton in healthcheck

[handlers.go:L18](file:///home/xo/temp/go-project/backend/internal/router/healthcheck/handlers.go#L18)

```go
var checks = &service{}
```

[handlers.go:L31-L34](file:///home/xo/temp/go-project/backend/internal/router/healthcheck/handlers.go#L31-L34)

```go
func setDependencies(sqlDB *sql.DB, redisClient *redis.Client) {
    checks.sqlDB = sqlDB
    checks.redisClient = redisClient
}
```

This is a **package-level mutable global** being written to at startup without any synchronization. If two goroutines ever call `setDependencies` concurrently (e.g., in tests), you have a data race. The entire healthcheck package relies on mutation of a shared pointer. Just... pass the dependencies through the handler function like every other handler in this codebase does. Your auth and users handlers already do this correctly with closures. Why is healthcheck special?

---

### 7. `RateLimiter` never evicts old entries — unbounded memory growth

[rate_limit.go:L22-L23](file:///home/xo/temp/go-project/backend/middleware/rate_limit.go#L22-L23)

```go
type RateLimiter struct {
    // ...
    buckets map[string]rateBucket
}
```

New entries are added for every unique client IP. Old entries are only overwritten when the same IP returns after the window expires. If you're behind a CDN or receive traffic from many unique IPs, this map grows without bound. There's no TTL sweep, no LRU eviction, nothing. In a long-running process, this is a slow memory leak.

---

### 8. Rate limiter ignores `X-Forwarded-For` / `X-Real-IP`

[rate_limit.go:L50](file:///home/xo/temp/go-project/backend/middleware/rate_limit.go#L50)

```go
clientKey := clientIP(req.RemoteAddr)
```

If you're behind a reverse proxy (nginx, cloudflare, etc.), `RemoteAddr` is the proxy's IP, not the client's. Every client gets the same rate limit bucket. The AGENTS.md mentions CloudFlare. This rate limiter is useless behind it.

---

### 9. CORS middleware reflects the request method directly

[cors.go:L30-L36](file:///home/xo/temp/go-project/backend/middleware/cors.go#L30-L36)

```go
requestMethod := r.Header.Get("Access-Control-Request-Method")
if requestMethod == "" {
    requestMethod = "GET,POST,PUT,PATCH,DELETE,OPTIONS"
} else {
    requestMethod += ",OPTIONS"
}
w.Header().Set("Access-Control-Allow-Methods", requestMethod)
```

You're reflecting the `Access-Control-Request-Method` header back to the client and just appending `,OPTIONS`. So if a client sends `Access-Control-Request-Method: PROPFIND`, you'll respond with `Allow-Methods: PROPFIND,OPTIONS`. This is a permissive reflection that effectively allows any method. Either validate against a known list or just use the hardcoded fallback every time.

---

### 10. `Recover` middleware doesn't capture stack traces

[recover.go:L15](file:///home/xo/temp/go-project/backend/middleware/recover.go#L15)

```go
logFn("panic recovered: %v", recovered)
```

You log the panic value but not the stack trace. When something panics in production, you get `panic recovered: runtime error: index out of range [3] with length 2` with zero context about where it happened. Use `debug.Stack()` or at minimum `runtime.Stack()`. Without a stack trace, this log line is almost useless.

---

### 11. No graceful shutdown

[run.go](file:///home/xo/temp/go-project/backend/cmd/run.go)

```go
func Run(c *config.Config, r *chi.Mux) error {
    server := &http.Server{...}
    return server.ListenAndServe()
}
```

The `.env` has `STAGE_STATUS` for dev vs prod, and the AGENTS.md mentions graceful shutdown in prod mode. But the `Run` function just calls `ListenAndServe()` and returns. No signal handling, no `Shutdown()` context, no drain period. When you deploy and SIGTERM hits, in-flight requests get axed mid-response.

---

### 12. `config.NewConfig()` uses `sync.Once` — un-testable singleton

[config.go:L77-L80](file:///home/xo/temp/go-project/backend/internal/config/config.go#L77-L80)

```go
var (
    once     sync.Once
    instance *Config
)
```

Once `NewConfig()` is called, the config is frozen forever in that process. Tests can't override it. Multiple test cases that need different configs? Too bad. The `sync.Once` pattern for config is an anti-pattern in Go test suites. Just return a new `*Config` every time — the caller can cache it if they want.

---

### 13. Dockerfile uses Go 1.22 but `go.mod` says 1.25

[Dockerfile:L1](file:///home/xo/temp/go-project/backend/Dockerfile#L1): `FROM golang:1.22-alpine`

[go.mod:L3](file:///home/xo/temp/go-project/backend/go.mod#L3): `go 1.25.0`

These don't match. The Dockerfile builds with 1.22, the module requires 1.25. This either fails at build time (if 1.25 features are used) or silently downgrades the Go version. Either way, nobody has built this Docker image recently.

---

## MEDIUM — Code quality & design

### 14. Duplicated `tokenFromRequest` / `extractToken` logic

[auth.go:L141-L164](file:///home/xo/temp/go-project/backend/internal/router/auth.go#L141-L164) and [middleware/auth.go:L63-L86](file:///home/xo/temp/go-project/backend/middleware/auth.go#L63-L86)

These are the **exact same function** copy-pasted into two different packages. `tokenFromRequest` in the router and `extractToken` in the middleware do identical things — check `Authorization: Bearer` header, fall back to `auth_token` cookie. This is a maintenance nightmare. When you fix a bug in one, you'll forget the other.

---

### 15. Duplicated `writeJSONError` / `writeAuthError`

[middleware/auth.go:L102-L106](file:///home/xo/temp/go-project/backend/middleware/auth.go#L102-L106) and [router/auth.go:L214-L218](file:///home/xo/temp/go-project/backend/internal/router/auth.go#L214-L218)

Same song, different package. Two identical JSON error writers that could be one shared utility.

---

### 16. `extractOpenAPITitle` is a hand-rolled YAML parser

[spec.go:L60-L88](file:///home/xo/temp/go-project/backend/internal/swagger/spec.go#L60-L88)

You wrote a custom YAML parser to extract the title from the OpenAPI spec. You iterate lines, track whether you're inside `info:`, check indentation manually. Meanwhile, the codebase already depends on `gopkg.in/yaml.v3` as a transitive dependency. Just unmarshal the spec into a struct and read `.Info.Title`. This hand-rolled parser will break on:
- Titles with colons in them (`title: "My API: v2"`)
- Specs that use `\r\n` line endings
- Different YAML indentation styles

---

### 17. `generateUUIDv4` silently returns empty string on failure

[request_id.go:L40-L42](file:///home/xo/temp/go-project/backend/middleware/request_id.go#L40-L42)

```go
if _, err := rand.Read(b[:]); err != nil {
    return ""
}
```

If `crypto/rand` fails (which is extremely rare but possible on depleted entropy), the request gets an empty request ID. This means log correlation breaks silently. Either use `google/uuid` (which panics on entropy failure, which is arguably correct), or at minimum log the error.

---

### 18. `Secure` cookie flag checks wrong env var

[auth.go:L177](file:///home/xo/temp/go-project/backend/internal/router/auth.go#L177)

```go
Secure: strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production"),
```

The config system uses `STAGE_STATUS` for dev/prod mode (see `.env`). But the cookie security flag checks `APP_ENV`. These are different variables. If someone sets `STAGE_STATUS=prod` but not `APP_ENV`, cookies won't be secure. Classic env var naming inconsistency.

---

### 19. `config.NewConfig()` panics on bad input instead of returning errors

[config.go:L91](file:///home/xo/temp/go-project/backend/internal/config/config.go#L91)

```go
panic("wrong server port (check your .env)")
```

There are 8 `panic()` calls in the config loading function. Panics are not errors — they're unrecoverable crashes with terrible UX. Return `(*Config, error)` like an adult. The caller (`main.go`) already has error handling for everything else.

---

### 20. `password.TrimSpace` on login

[auth.go:L56](file:///home/xo/temp/go-project/backend/internal/router/auth.go#L56)

```go
payload.Password = strings.TrimSpace(payload.Password)
```

You're trimming whitespace from the password. If a user's password is `"  hunter2  "`, it becomes `"hunter2"`. This silently changes their credential. Passwords should be accepted as-is, byte-for-byte. Trimming email is fine; trimming passwords is a bug.

---

### 21. MSW mock handlers don't validate request bodies

[handlers.ts:L138-L144](file:///home/xo/temp/go-project/frontend/src/mocks/handlers.ts#L138-L144)

```typescript
http.post("/auth/login", async () => {
    return HttpResponse.json(loginResponseExample, { ... });
}),
```

The login mock always returns success regardless of what credentials are sent. The other handlers have `validateOptionalBody` but login, refresh, and user profile skip validation entirely. Frontend tests against these mocks will never catch malformed request payloads.

---

### 22. Frontend `parseErrorMessage` checks wrong field

[api-client.ts:L55-L58](file:///home/xo/temp/go-project/frontend/src/lib/api-client.ts#L55-L58)

```typescript
const body = (await response.json()) as { message?: unknown };
if (typeof body.message === "string" && body.message.length > 0) {
    return body.message;
}
```

The backend returns errors as `{"error": "..."}` (see `writeAuthError`), but the frontend checks for `body.message`. These don't match. The frontend will never extract the actual error message from auth errors and will always fall back to the generic `"Request failed with status XXX"`.

---

### 23. OpenAPI spec has auth on healthcheck and swagger endpoints

[openapi.yaml:L20-L22](file:///home/xo/temp/go-project/backend/internal/swagger/openapi.yaml#L20-L22)

```yaml
/hc/status:
    get:
      security:
        - BearerAuth: []
        - CookieAuth: []
```

The spec claims healthcheck, swagger, and openapi.yaml endpoints require authentication. The actual router doesn't enforce auth on any of these routes. The contract test passes because it sends auth headers on every request, masking the discrepancy. This is spec drift — the documented API doesn't match reality.

---

### 24. `auth/refresh` doesn't exist in the MSW handlers

The backend has `POST /auth/refresh` but the [MSW handlers](file:///home/xo/temp/go-project/frontend/src/mocks/handlers.ts) don't mock it. Any frontend code that tries to refresh tokens in tests will get an unhandled request.

---

### 25. No `auth/refresh` test for cookie-based flow

The auth tests only test token refresh via `Authorization: Bearer`. The AGENTS.md says "Both work simultaneously" for Bearer and cookie auth. But there's no test that sends a refresh request using only the `auth_token` cookie. You don't know if cookie-based refresh actually works.

---

## LOW — Nitpicks & cleanup

### 26. `TestLogoutReturns200WithoutAuth` reads body twice

[auth_test.go:L439-L451](file:///home/xo/temp/go-project/backend/internal/router/auth_test.go#L439-L451)

```go
defer resp.Body.Close()
// ...
if resp.StatusCode != http.StatusOK {
    rawBody, _ := io.ReadAll(resp.Body)
    // ...
}
rawBody, err := io.ReadAll(resp.Body) // second read
```

If the status check branch is taken, the body is already consumed. The second `ReadAll` gets nothing. Not a bug in the happy path, but the error path logs an empty body.

---

### 27. Coverage files committed to the repo

[backend/backend-coverage.out](file:///home/xo/temp/go-project/backend/backend-coverage.out) and [backend/coverage.out](file:///home/xo/temp/go-project/backend/coverage.out) — 36KB and 29KB of coverage data tracked in git. These are build artifacts. Add them to `.gitignore`.

---

### 28. `backend/api` directory exists but purpose is unclear

There's a `backend/api` directory alongside `backend/api_test.http` (41 bytes). What's in it? If it's empty scaffold, delete it. If it matters, it's undocumented.

---

### 29. `.npm-cache` directory in frontend

[frontend/.npm-cache](file:///home/xo/temp/go-project/frontend/.npm-cache) — the project mandates bun, never npm. Why is there an npm cache directory? This is either a leftover from someone running npm anyway, or a CI artifact leak.

---

### 30. `CLAUDE.md` is just `@AGENTS.md`

[CLAUDE.md](file:///home/xo/temp/go-project/frontend/CLAUDE.md) contains literally one line: `@AGENTS.md`. If this is a symlink mechanism for Claude, fine. But it's a bit silly to track a 11-byte file whose entire purpose is "go read the other file."

---

## Summary

| Severity | Count | Themes |
|----------|-------|--------|
| **Critical** | 5 | Plaintext passwords, hardcoded JWT secret, committed `.env`, race condition in validation middleware, redundant expiry check |
| **High** | 8 | Mutable globals, memory leaks, missing graceful shutdown, Go version mismatch, un-testable singleton config |
| **Medium** | 12 | Duplicated code, incorrect env var, spec drift, frontend-backend error format mismatch, wrong password trimming |
| **Low** | 5 | Build artifacts in git, empty directories, cosmetic issues |

The bones are there — contract testing, middleware layering, dual auth strategy, test infrastructure with real Postgres/Redis. But the security fundamentals are missing. No password hashing means this codebase cannot touch production in its current state.

Fix the criticals. Then we'll talk.
