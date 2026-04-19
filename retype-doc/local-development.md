# Local Development

This page covers day-to-day development commands and guardrails.

## Startup Patterns

Run full stack:

```bash
make dev
```

Run services separately:

```bash
# backend
cd backend && make dev

# frontend
cd frontend && bun run dev
```

## Test Infrastructure

Start test Postgres/Redis:

```bash
docker compose -f docker-compose.test.yml up -d
```

Stop test services:

```bash
make down
```

Test service defaults:

- Postgres: `localhost:5433`
- Redis: `localhost:6380`
- `TEST_DATABASE_URL=postgres://postgres:test@localhost:5433/testdb?sslmode=disable`
- `TEST_REDIS_URL=redis://localhost:6380`

## Daily Command Reference

```bash
make test                # backend + frontend tests
make lint                # go + frontend lint
make build               # backend + frontend build
make validate-contracts  # OpenAPI lint + contract tests
make check               # full pre-deploy checks
make ci-local            # full CI simulation
```

## Local API/Auth Notes

- Browser clients use cookie fallback automatically (`credentials: "include"` in API client).
- If a bearer token exists in local storage, it is sent and takes priority over cookie auth.
- `POST /auth/logout` clears `auth_token` cookie.

## Contract And Mock Discipline

When API shapes change:

1. Update OpenAPI spec first.
2. Update backend handlers/tests.
3. Update `frontend/src/mocks/handlers.ts`.
4. Re-run contract validation.

Skipping MSW updates creates false positives in frontend tests.

## Do Not Do These

- Do not use npm or yarn; Bun is mandatory.
- Do not add another OpenAPI file.
- Do not use SQLite/miniredis in tests.
- Do not bypass checks with `--pass-with-no-tests` or `--no-verify`.
