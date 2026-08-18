# AGENTS.md

This file is the authoritative working guide for any AI agent or developer in this repository.

## 0. Machine-Readable Project Context
Read this section first before any other section or any user instruction.
These values override any assumption you might make from file contents alone.

- Go module: `github.com/kntjspr/fullstack-golang-next-template`
- Note: the module name is set during first-time bootstrap. If you are working in a freshly cloned repo that has not been bootstrapped yet, run `make bootstrap` first - it will prompt for the module name and rename it throughout all relevant files automatically.
- OpenAPI spec: `backend/internal/swagger/openapi.yaml` (never move this file, never create a second spec file)
- Test database: Postgres on port `5433` via `docker-compose.test.yml` (`TEST_DATABASE_URL=postgres://postgres:test@localhost:5433/testdb`)
- Test cache: Redis on port `6380` via `docker-compose.test.yml` (`TEST_REDIS_URL=redis://localhost:6380`)
- No SQLite: never use SQLite or miniredis in any test
- No spec copy: `backend/internal/swagger/spec.go` embeds `openapi.yaml` at compile time, moving the file breaks the build
- Auth strategy: Bearer token (`Authorization` header) takes priority, `httpOnly` cookie (`auth_token`) is fallback, both work simultaneously

Key file locations:
- Router registration: `backend/internal/router/router.go`
- Stable business contracts: `backend/core/`
- Replaceable AI implementations and adapters: `backend/ai/`
- Protected service contracts: `backend/service_test.go`
- AI-owned scratch tests: `backend/service_ai_test.go`
- Handler pattern: `backend/internal/router/auth.go` (copy this pattern)
- Handler test pattern: `backend/internal/router/auth_test.go` (copy this pattern)
- Middleware: `backend/middleware/`
- Auth logic: `backend/internal/auth/token.go`
- Password hashing: `backend/internal/auth/password.go` (always use `auth.HashPassword`/`auth.ComparePassword`)
- Shared HTTP utilities: `backend/internal/httpapi/auth.go` (token extraction, JSON error writer — check here before writing new helpers)
- DB connection: `backend/internal/database/postgres.go`
- Redis connection: `backend/internal/cache/redis.go`
- Config: `backend/internal/config/config.go`

## 1. Project Overview
This repository is a production-ready monorepo template for building and deploying a Go API.

Stack summary:
- Backend: Go 1.25+, chi router, GORM, Postgres, Redis
- Infra/ops: Ansible roles in `roles/`, Docker-based test dependencies
- Contract: OpenAPI spec at `backend/internal/swagger/openapi.yaml`

Repository structure:
- `backend/`: Go API service, stable core contracts, AI adapters, middleware, router, auth, DB/cache integrations, migrations
- `docs/`: architecture and security documentation
- `scripts/`: cross-project test and reporting scripts
- `roles/`: deployment automation (Ansible)
- `docker-compose.test.yml`: test Postgres/Redis services
- `Makefile`: root automation entry points (test, contracts, integration, coverage, ci-local)

## 2. Prerequisites
Install these tools before working in this repo:
- Go 1.25+
- Docker + Docker Compose
- Make

Start local test infrastructure:
```bash
docker compose -f docker-compose.test.yml up -d
```

## 3. Environment Setup
Copy environment templates first:
```bash
cp backend/.env.example backend/.env
```

Required environment variables:
- `STAGE_STATUS`: server mode (`dev` or `prod`)
- `SERVER_HOST`: backend bind host
- `SERVER_PORT`: backend port (default `5000`)
- `SERVER_READ_TIMEOUT`, `SERVER_WRITE_TIMEOUT`, `SERVER_IDLE_TIMEOUT`: HTTP timeouts (seconds)
- `LOGGER_LEVEL`, `LOGGER_PRETTY`: backend logger behavior
- `DB_ENABLE`: enable/disable Postgres integration
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSL_MODE`, `DB_TIMEZONE`: Postgres connection settings
- `REDIS_ENABLE`: enable/disable Redis integration
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`: Redis connection settings
- `JWT_SECRET`: signing secret for JWT generation/validation (HS256)
- `SENTRY_DSN`, `SENTRY_ENVIRONMENT`, `SENTRY_RELEASE`, `SENTRY_TRACES_SAMPLE_RATE`: telemetry settings

Test env vars for local test runs:
- `TEST_DATABASE_URL=postgres://postgres:test@localhost:5433/testdb?sslmode=disable`
- `TEST_REDIS_URL=redis://localhost:6380`

## 4. Running the Project
Start backend locally:
```bash
cd backend && make dev
```

Port assignments:
- Backend API: `5000`
- Postgres (dev/default): `5432`
- Postgres (test via docker-compose.test.yml): `5433`
- Redis (dev/default): `6379`
- Redis (test via docker-compose.test.yml): `6380`

## 5. Running Tests
Run full test suite:
```bash
make test
```

Run integration tests (Docker required):
```bash
bash scripts/test-integration.sh
```

Generate coverage summary:
```bash
bash scripts/coverage-report.sh
```

Validate contracts (OpenAPI lint + contract tests):
```bash
make validate-contracts
```

Simulate CI end-to-end locally:
```bash
make ci-local
```

Important:
- All backend tests require `docker-compose.test.yml` services running with valid `TEST_DATABASE_URL` and `TEST_REDIS_URL`.

## 6. Project Conventions
- All Go imports must use module path `github.com/kntjspr/fullstack-golang-next-template`.
- OpenAPI single source of truth is `backend/internal/swagger/openapi.yaml`; never create a second spec file.
- Never use SQLite or miniredis in tests; use real Postgres/Redis via `docker-compose.test.yml`.
- Do not modify Ansible roles unless you are intentionally changing deployment infrastructure.
- Do not use `--pass-with-no-tests` or `--no-verify` anywhere.
- Auth supports dual strategy: Bearer token (Authorization header) takes priority, httpOnly cookie (auth_token) is fallback. Both work simultaneously.
- Browser clients receive the cookie automatically on login and send it automatically on subsequent requests via credentials: "include" in api-client.ts
- CLI/MCP/API clients: pass Authorization: Bearer <token> header explicitly
- Logout: POST /auth/logout clears the cookie

## 6.1 AI Execution Loop Protocol
The AI governance boundary is scoped to the Go module at `backend/`:
- `backend/core/` contains stable business contracts and rules.
- `backend/ai/` contains replaceable AI-authored implementations and adapters.
- `backend/service_test.go` is the protected, hand-authored contract suite.
- `backend/service_ai_test.go` is available for AI-owned scratch coverage.

Follow this loop for every AI-authored change:
1. Read the relevant code, ADRs, and `backend/service_test.go`.
2. Run `make test` and confirm the baseline is green.
3. Modify application code only to fulfil the requested behavior.
4. Run `make test` again. If `service_test.go` fails, fix the implementation; never alter an existing contract assertion. Scratch tests may change with internal implementation details.
5. Run `make security` and `make mutation-test` before declaring the task complete.

Hard rules:
- Never use `panic` for expected failures; return errors first.
- New third-party runtime imports require an accepted ADR that records the approval.
- Goroutines must receive a `context.Context` and must not leak.
- Treat `backend/core/` contracts as stable. Expand them only with a new, additive contract case and a documented ADR when the public behavior changes.

## 7. Adding a New Backend Endpoint
Every new endpoint must follow this exact sequence:
1. Add the route to `backend/internal/swagger/openapi.yaml` first.
2. Run `make validate-contracts` (expected to fail until implementation exists).
3. Write handler test in `backend/internal/router/[feature]_test.go`.
4. Run `go test ./backend/internal/router/...` and confirm FAIL.
5. Implement handler in `backend/internal/router/[feature].go`.
6. Run `go test ./backend/internal/router/...` and confirm PASS.
7. Run `make validate-contracts`; it must pass before commit.
8. If endpoint touches DB/cache, add integration test in `backend/internal/integration/`.

## Worked example: GET /users/me
See these files for a complete working example to copy:
- Test: `backend/internal/router/users_test.go`
- Handler: `backend/internal/router/users.go`
- Spec: `backend/internal/swagger/openapi.yaml` (search for `/users/me`)
- Mock: `frontend/src/mocks/handlers.ts` (search for `users/me`)

## 9. What NOT To Do
- Do not create `backend/api/openapi.yaml`: the canonical spec is `backend/internal/swagger/openapi.yaml`.
- Do not move `backend/internal/swagger/openapi.yaml`: `spec.go` embeds that path at compile time.
- Do not use SQLite or miniredis in tests: this breaks production parity and hides integration issues.
- Do not use `--pass-with-no-tests`: it masks missing tests and creates false confidence.
- Do not use `--no-verify` on commits: it bypasses repository quality gates and hooks.
- Do not run deployment before `make check`/pre-flight validations in your workflow: release without checks increases production risk.
- Do not deploy without running make check first
- Do not add a router handler without a corresponding OpenAPI path: contract drift will break consumers.

## 10. Test-First Development Protocol
Every feature follows this exact sequence. No exceptions.

Step 1: Write the failing test
- Backend: create `backend/[package]/[feature]_test.go`
- Run: `make test` and confirm the new test FAILS (not skipped, not compile error)

Step 2: Write minimum implementation to pass
- Implement only what is required to make the new test green
- Run: `make test` and confirm PASS

Step 3: Update the contract
- If adding/changing backend endpoint behavior, update `backend/internal/swagger/openapi.yaml`
- Run: `make validate-contracts`; must pass before commit

Step 4: Add integration test if DB/Redis is touched
- Add `backend/internal/integration/[feature]_test.go` with `integration` build tag
- Run: `bash scripts/test-integration.sh`

Step 5: Commit
- Let pre-commit hooks run normally
- If any hook fails, fix issues before commit
- Never use `--no-verify`

## 11. What a Failing Test Looks Like
- A compile error is not a failing test; it is a broken test.
- A panic is not a failing test; it is a broken test unless panic behavior is the assertion target.
- A real failing test executes and fails an assertion (`t.Fail`, `t.Error`, or assertion library failure).
- When practicing test-first, stub minimal return values so failures are assertion-based, not compile-based.

Example stub for red phase:
```go
func GenerateToken(...) (string, error) { return "", nil }
```

## 12. Contract Drift Prevention
If handler response shape changes:
1. Update `backend/internal/swagger/openapi.yaml` first.
2. Run `make validate-contracts`; initial failure is expected.

The contract is always the source of truth, not ad-hoc implementation details.

## 13. CI Pipeline Summary
- `unit-tests`: runs on every push and pull request
- `contract-tests`: runs on every push and pull request
- `security`: runs on every push and pull request
- `mutation-tests`: runs on every push and pull request
- `integration-tests`: runs on push to `main` and PRs targeting `main` only
- All required jobs must pass before merge
- Never merge while any required job is red

## 14. Code Invariants

These are non-negotiable implementation rules. Violating any of these is a blocking defect regardless of whether tests pass or CI is green.

### Security
- Passwords are ALWAYS stored as bcrypt hashes. Use `auth.HashPassword` to hash and `auth.ComparePassword` to verify. Never compare passwords with `==` or `!=`.
- JWT signing secret MUST come from `JWT_SECRET` env var. Never hardcode a fallback secret. `auth.RequireJWTSecret()` is called at startup and fatals if missing.
- Cookie `Secure` flag checks `STAGE_STATUS` (not `APP_ENV` or any other variable). `STAGE_STATUS=prod` → secure cookies, `STAGE_STATUS=dev` → insecure.
- Never `TrimSpace` passwords. Trim emails only.
- Config loading returns `(*Config, error)`. No panics in library code; panics are reserved for truly unrecoverable programmer errors.

### Architecture
- No mutable package-level state. Dependencies go through constructors or closures — see `healthcheck.newService` and handler closures in `router/auth.go`.
- No `sync.Once` singletons for config. `NewConfig()` returns a fresh instance each call.
- Shared auth utilities (token extraction, JSON error writing) live in `internal/httpapi/`. Do not duplicate these in `middleware/` or `router/`.
  - `internal/httpapi` = cross-package HTTP utilities
  - `backend/middleware/` = request middleware
  - `backend/internal/router/` = handlers
- Before writing any new auth-extraction or JSON-error helper, grep for it in `internal/httpapi/` first.

### Error response format
- Auth errors: `httpapi.WriteJSONError(w, status, message)` → `{"error": "..."}`
- Validation errors: `{"error": "validation failed", "fields": [...]}`
Backend MUST use `"error"` as the key, not `"message"`.

### Environment variables
- Only two valid values for `STAGE_STATUS`: `dev` and `prod`. Config validation rejects anything else.
- New env vars must appear in: `backend/.env.example`, AGENTS.md §3, and `retype-doc/reference-environment.md` — all three in the same commit.
- Handlers should not call `os.Getenv` directly. Route through `internal/config/config.go`. The one known exception is cookie `Secure` logic in `router/auth.go`.

### Middleware
- Rate limiter evicts stale entries via `cleanupStaleBuckets`. Verify this runs when touching `middleware/rate_limit.go`.
- Rate limiter resolves client IP from `CF-Connecting-IP` → `X-Forwarded-For` → `X-Real-IP` → `RemoteAddr`. Do not regress to `RemoteAddr` only.
- Panic recovery MUST log `debug.Stack()`. Do not remove the stack trace.
- Request ID falls back to a timestamp-based ID when RNG fails. Never return an empty request ID.
- CORS allowed methods are the hardcoded `allowedCORSMethods` constant. Do not reflect `Access-Control-Request-Method` from the client.

### Docker
- Dockerfile Go version must match `go.mod`. If you bump one, bump the other.
- Do not copy `.env` into Docker images (known tech debt in current `Dockerfile` line 22).

## 15. Self-Review Checklist

Every change must pass all applicable items before commit. This is mandatory.

### Every change
- [ ] `make test` passes
- [ ] `make validate-contracts` passes
- [ ] No new `os.Getenv` calls outside `config.go` (except known exceptions)
- [ ] No new auth-extraction or JSON-error helper functions — check `internal/httpapi/` first
- [ ] No mutable package-level variables (`var x = &thing{}` at package scope)
- [ ] New generated/build output files are covered by `.gitignore`

### Changes that touch auth
- [ ] Passwords use `auth.HashPassword` / `auth.ComparePassword` — never `==` comparison
- [ ] Test fixtures use `testutil.CreateTestUser` with `"password"` override key (not `"password_hash"` with plaintext)
- [ ] Token generation uses `auth.GenerateToken` — no direct JWT construction
- [ ] Cookie secure flag reads `STAGE_STATUS`, not `APP_ENV`
- [ ] Both Bearer header and cookie auth paths have tests

### Changes that touch API endpoints
- [ ] OpenAPI spec updated BEFORE implementation
- [ ] `security:` block is present on protected routes and absent (or `security: []`) on public routes
- [ ] MSW handlers in `frontend/src/mocks/handlers.ts` updated with body validation
- [ ] Backend error responses use `{"error": "..."}` not `{"message": "..."}`
- [ ] No dead 401/403 responses documented for routes without auth middleware

### Changes that touch middleware
- [ ] CORS does not reflect `Access-Control-Request-Method`
- [ ] Rate limiter `cleanupStaleBuckets` still runs
- [ ] IP extraction chain is preserved (CF > XFF > XRI > RemoteAddr)
- [ ] Panic recovery logs `debug.Stack()`

### Changes that touch config
- [ ] `NewConfig` returns `(*Config, error)`, no panics
- [ ] No `sync.Once`
- [ ] New env vars added to `.env.example`, AGENTS.md §3, and `retype-doc/reference-environment.md`

### Go version or Dockerfile changes
- [ ] `go.mod` and `Dockerfile` Go version match
- [ ] AGENTS.md §1 and §2 updated
- [ ] `retype-doc/onboarding.md` updated

## 16. Documentation Sync

Code behavior changes require documentation updates in the same commit. Use this table:

| What changed | Files to update |
|---|---|
| New or changed env var | `backend/.env.example`, AGENTS.md §3, `retype-doc/reference-environment.md` |
| New backend package | AGENTS.md §0 key file locations, `retype-doc/reference-repository-map.md`, `docs/ARCHITECTURE.md` |
| New API endpoint | `openapi.yaml` |
| Auth behavior change | AGENTS.md §14 security invariants, `retype-doc/troubleshooting.md`, `docs/ARCHITECTURE.md` auth flow section |
| Go version bump | `go.mod`, `backend/Dockerfile`, AGENTS.md §1 and §2, `retype-doc/onboarding.md` |
| Middleware behavior change | AGENTS.md §14 middleware invariants, `docs/ARCHITECTURE.md` middleware section |
| New shared utility package | AGENTS.md §14 architecture invariants, `retype-doc/reference-repository-map.md` |

If the documentation is not updated with the code change, the change is incomplete.
