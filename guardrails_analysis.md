# Guardrails Analysis: Why AI Agents Keep Introducing Bugs

## Root Cause

The current AGENTS.md is good at **process** (spec-first, test-first, contract validation) but says almost nothing about **implementation invariants**. An AI agent can follow every step in Section 7 perfectly — update the spec, write a test, implement the handler — and still produce plaintext password storage, hardcoded secrets, and copy-pasted utility functions, because nowhere does it say "don't do that."

The bugs from the initial review fall into three categories that AGENTS.md doesn't address:

### 1. Security invariants are not documented
The review found plaintext password comparison, hardcoded JWT fallback, and wrong env var for cookie security. None of these are prevented by "write test first" or "update OpenAPI spec." An agent that follows the current instructions perfectly will still invent its own password handling if it doesn't know bcrypt is mandatory.

### 2. Code reuse patterns are not documented  
The review found duplicated token extraction, duplicated JSON error writers, and a hand-rolled YAML parser when a library was already available. An agent working in `router/auth.go` doesn't know that `middleware/auth.go` already has the same function because nothing says "shared auth helpers live in `internal/httpapi`."

### 3. Architecture constraints are implicit
No mutable package-level state, no panics in library code, no hardcoded fallback secrets — these are obvious to a human reviewer but invisible to an AI that's focused on making the tests pass. The agent optimizes for "green CI" not "correct architecture."

---

## Retype-Doc Drift from Current Codebase

| File | What's stale | Current truth |
|------|-------------|---------------|
| [index.md:L20](file:///home/xo/temp/go-project/retype-doc/index.md#L20) | "CORS middleware exists but is not mounted in `backend/main.go`" | CORS middleware status needs verification — is it mounted now? If so, remove the warning. |
| [index.md:L21](file:///home/xo/temp/go-project/retype-doc/index.md#L21) | "Cookie `Secure` depends on `APP_ENV=production`" | Now depends on `STAGE_STATUS=prod`. Already fixed in `reference-environment.md` but stale here. |
| [onboarding.md:L15](file:///home/xo/temp/go-project/retype-doc/onboarding.md#L15) | "Go `1.22+`" | `go.mod` says `1.25.0`, Dockerfile says `golang:1.25-alpine` |
| [reference-repository-map.md](file:///home/xo/temp/go-project/retype-doc/reference-repository-map.md) | Missing `internal/httpapi/` package | New shared package added during review fixes |
| [reference-repository-map.md](file:///home/xo/temp/go-project/retype-doc/reference-repository-map.md) | Missing `internal/auth/password.go` | New file added during review fixes |
| [deployment-split-host.md:L51](file:///home/xo/temp/go-project/retype-doc/deployment-split-host.md#L51) | "CORS middleware...is not mounted in `backend/main.go`" | Needs verification against current `main.go` |
| [deployment-split-host.md:L64-L65](file:///home/xo/temp/go-project/retype-doc/deployment-split-host.md#L64-L65) | Cookie env var description was updated but still references old pattern | Already corrected to `STAGE_STATUS` ✅ |
| AGENTS.md [L37](file:///home/xo/temp/go-project/AGENTS.md#L37) | "Go 1.22+" | Should be "Go 1.25+" |

---

## Proposed Changes

### A. Add "Code Invariants" section to AGENTS.md

This is the missing piece. The current doc tells agents *what process to follow*, but not *what code patterns are mandatory*. Adding a section that codifies the invariants caught in the review prevents agents from re-inventing them wrong.

Proposed new section (insert after Section 6, renumber subsequent sections):

```markdown
## 7. Code Invariants

These are non-negotiable implementation rules. Violating any of these is a blocking defect regardless of whether tests pass.

### Security
- Passwords are ALWAYS stored as bcrypt hashes. Use `auth.HashPassword` to hash and `auth.ComparePassword` to verify. Never compare passwords with `==` or `!=`.
- JWT signing secret MUST come from `JWT_SECRET` env var. Never hardcode a fallback secret. `auth.RequireJWTSecret()` is called at startup and fatals if unset.
- Cookie `Secure` flag checks `STAGE_STATUS` (not `APP_ENV` or any other variable). `STAGE_STATUS=prod` means secure cookies.
- Never `TrimSpace` passwords. Trim emails, not credentials.
- Config loading returns `(*Config, error)`. Never panic in library code. Panics are reserved for truly unrecoverable programmer errors, not user-facing config validation.

### Architecture
- No mutable package-level state. If a handler needs dependencies, pass them through constructors or closures (see `healthcheck.newService`, handler closures in `router/auth.go`).
- No `sync.Once` singletons for config. `NewConfig()` returns a fresh instance. The caller decides caching.
- Shared auth helpers (token extraction, JSON error writing) live in `internal/httpapi/`. Do not duplicate these in `middleware/` or `router/`.
- The `internal/httpapi` package is for cross-package HTTP utilities. `middleware/` is for request middleware. `router/` is for handlers. Do not blur these boundaries.

### Error responses
- Auth errors use `httpapi.WriteJSONError(w, status, message)` which produces `{"error": "..."}`.
- Validation errors use `{"error": "validation failed", "fields": [...]}`.
- The frontend `api-client.ts` parses `body.error` first, then `body.message`. Backend error responses must use `"error"` as the key.

### Environment variables
- Only two valid values for `STAGE_STATUS`: `dev` and `prod`. Config validation rejects anything else.
- All env-driven config goes through `internal/config/config.go`. Handlers should not read `os.Getenv` directly except for `STAGE_STATUS` in cookie logic (which is a known exception).

### Middleware
- Rate limiter evicts stale entries. If you touch the rate limiter, verify that `cleanupStaleBuckets` runs.
- Rate limiter resolves client IP from `CF-Connecting-IP` > `X-Forwarded-For` > `X-Real-IP` > `RemoteAddr`. Do not regress to `RemoteAddr` only.
- Panic recovery logs `debug.Stack()`. Do not remove stack traces.
- Request ID falls back to timestamp-based ID when RNG fails. Never return an empty request ID.
- CORS methods are hardcoded in `allowedCORSMethods`. Do not reflect `Access-Control-Request-Method` from the client.

### Docker
- Dockerfile Go version must match `go.mod`. If you bump one, bump the other.
- Do not copy `.env` into production Docker images. The current Dockerfile does this and it's a known tech debt item.
```

### B. Add "Self-Review Checklist" section to AGENTS.md

This is the "before you commit" gate that prevents the most common AI mistakes:

```markdown
## 14. Self-Review Checklist

Before committing any change, verify ALL of these. This is mandatory, not advisory.

### Every change
- [ ] `make test` passes
- [ ] `make validate-contracts` passes
- [ ] No new `os.Getenv` calls outside `config.go` (except known exceptions)
- [ ] No duplicated utility functions — check `internal/httpapi/` before writing JSON error/auth extraction helpers
- [ ] No mutable package-level variables (`var x = &thing{}` at package scope)
- [ ] `.gitignore` covers any new generated files (coverage, build output, etc)

### Changes that touch auth
- [ ] Passwords go through `auth.HashPassword`/`auth.ComparePassword` — never `==`
- [ ] Test fixtures use `testutil.CreateTestUser` with `"password"` override key (not `"password_hash"` with plaintext)
- [ ] Token generation uses `auth.GenerateToken` — no direct JWT construction
- [ ] Cookie secure flag reads `STAGE_STATUS`, not `APP_ENV`
- [ ] Both Bearer header and cookie auth paths are tested

### Changes that touch API endpoints
- [ ] OpenAPI spec updated BEFORE implementation
- [ ] `security:` block present on protected routes, absent on public routes
- [ ] MSW handlers in `frontend/src/mocks/handlers.ts` updated with body validation
- [ ] Frontend `api-client.ts` error parsing still works (backend returns `{"error": "..."}`)
- [ ] No stale 401/403 responses documented in spec for unauthenticated routes

### Changes that touch middleware
- [ ] No `Access-Control-Request-Method` reflection in CORS
- [ ] Rate limiter cleanup logic still runs
- [ ] IP extraction chain is preserved (CF > XFF > XRI > RemoteAddr)
- [ ] Recovery middleware includes `debug.Stack()`

### Changes that touch config
- [ ] Returns `(*Config, error)`, not panic
- [ ] No `sync.Once`
- [ ] New env vars added to: `.env.example`, AGENTS.md Section 3, `retype-doc/reference-environment.md`
```

### C. Add "Documentation Sync" rule to AGENTS.md

```markdown
## 15. Documentation Sync

When you change code behavior, update these documentation sources in the SAME commit:

| What changed | Update these |
|-------------|-------------|
| New/changed env var | `backend/.env.example`, AGENTS.md §3, `retype-doc/reference-environment.md` |
| New backend package | AGENTS.md §0 key file locations, `retype-doc/reference-repository-map.md`, `docs/ARCHITECTURE.md` |
| New API endpoint | `openapi.yaml`, `frontend/src/mocks/handlers.ts`, `retype-doc/reference-repository-map.md` if it's a new package |
| Auth behavior change | AGENTS.md §7 security invariants, `retype-doc/troubleshooting.md`, `docs/ARCHITECTURE.md` auth flow section |
| Go version bump | `go.mod`, `Dockerfile`, AGENTS.md §2, `retype-doc/onboarding.md` |
| Middleware behavior change | AGENTS.md §7 middleware invariants, `docs/ARCHITECTURE.md` middleware section |

If documentation is not updated with the code change, the change is incomplete.
```

---

## Immediate Retype-Doc Fixes Needed

These should be applied now to make docs consistent with the current codebase:

1. **`index.md` L20**: verify CORS mount status, update or remove warning
2. **`index.md` L21**: change `APP_ENV=production` → `STAGE_STATUS=prod`  
3. **`onboarding.md` L15**: change `Go 1.22+` → `Go 1.25+`
4. **`reference-repository-map.md`**: add `internal/httpapi/` and `internal/auth/password.go`
5. **`deployment-split-host.md` L51**: verify CORS mount status
6. **`AGENTS.md` L37**: change `Go 1.22+` → `Go 1.25+`
7. **`docs/ARCHITECTURE.md` L62**: verify auth flow description matches current `token.go` (startup fatal, no fallback)

---

## Summary

The root problem isn't that AI agents are bad at following instructions — they follow them literally. The problem is that your instructions describe **process** without **invariants**. Adding Sections 7, 14, and 15 to AGENTS.md closes the gap between "did the CI pass?" and "is this code actually correct?"

> [!IMPORTANT]
> Do you want me to apply all of these changes (AGENTS.md additions + retype-doc fixes) now, or do you want to review/modify the proposed sections first?
