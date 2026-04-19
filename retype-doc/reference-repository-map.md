# Repository Map

Key paths and responsibilities.

## Root

- `AGENTS.md`: authoritative process and constraints for contributors/agents.
- `README.md`: quick-start and high-level project overview.
- `Makefile`: standard automation entry points.
- `docker-compose.test.yml`: test Postgres/Redis services.
- `playbook.yml`, `hosts.ini`: Ansible deployment entrypoints.

## Backend

- `backend/main.go`: process wiring and middleware registration.
- `backend/internal/router/router.go`: route registration root.
- `backend/internal/router/auth.go`: auth handlers and cookie behavior.
- `backend/internal/auth/token.go`: token generation/validation logic.
- `backend/internal/auth/password.go`: bcrypt hash and verify helpers — always use these, never compare passwords directly.
- `backend/internal/httpapi/auth.go`: shared HTTP utilities (token extraction, JSON error writer) used by both middleware and handlers.
- `backend/internal/config/config.go`: environment loading and config model.
- `backend/internal/database/postgres.go`: Postgres connection lifecycle.
- `backend/internal/cache/redis.go`: Redis connection lifecycle.
- `backend/internal/swagger/openapi.yaml`: canonical API contract.
- `backend/internal/contract/`: API contract tests.
- `backend/middleware/`: reusable middleware modules.

## Frontend

- `frontend/src/app/`: Next.js App Router pages/layouts.
- `frontend/src/lib/api-client.ts`: fetch wrapper, auth header/cookie behavior.
- `frontend/src/mocks/handlers.ts`: MSW handlers mirroring backend contract.
- `frontend/src/lib/__tests__/`: frontend unit/integration-level tests.
- `frontend/next.config.ts`: static export mode.

## Infrastructure

- `roles/docker/`: Docker setup role.
- `roles/frontend/`: frontend deploy/build role.
- `roles/backend/`: backend image/container role.
- `roles/postgres/`: postgres provisioning role.
- `roles/redis/`: redis provisioning role.
- `roles/nginx/`: nginx templating and container role.

## Additional Docs

- `docs/ARCHITECTURE.md`: architecture deep dive.
- `docs/DEPLOYMENT_TOPOLOGIES.md`: topology differences and constraints.
- `docs/security/INPUT_HANDLING.md`: input validation/sanitization policy.
- `docs/adr/`: architecture decision records.
