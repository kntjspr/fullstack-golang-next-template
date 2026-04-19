# Onboarding

This process is designed for a new engineer onboarding to this repository with minimum ambiguity.

## Audience

- Backend/frontend engineers contributing features.
- DevOps engineers preparing deployment automation.
- Reviewers validating contract and test discipline.

## Stage 0: Workstation Prerequisites

Install these before running anything:

- Go `1.25+`
- Bun
- Docker + Docker Compose
- GNU Make

Verify:

```bash
go version
bun --version
docker compose version
make --version
```

## Stage 1: Clone And Bootstrap

```bash
git clone <your-repo-url>
cd go-project
make bootstrap
```

What `make bootstrap` does:

1. First-run module rename prompt (optional) and project-wide module replacement.
2. Copies `.env` templates if missing.
3. Installs git hooks.
4. Installs frontend dependencies using Bun.
5. Starts test infra (`docker-compose.test.yml`).
6. Validates backend DB/Redis connectivity.
7. Runs contract validation and backend/frontend tests.

Expected final result:

- `.bootstrap-done` exists.
- Postgres/Redis test containers are running.
- Bootstrap exits with success and suggests `make dev`.

## Stage 2: Environment Setup

These files should exist after bootstrap:

- `.env`
- `backend/.env`
- `frontend/.env`

Required backend baseline values:

- `STAGE_STATUS=dev`
- `SERVER_HOST=0.0.0.0`
- `SERVER_PORT=5000`
- `DB_ENABLE` and `REDIS_ENABLE` according to your local needs
- `JWT_SECRET` for auth token signing

Required frontend baseline values:

- `NEXT_PUBLIC_API_URL=http://localhost:5000`
- `NEXT_PUBLIC_SITE_URL=http://localhost:3000`

## Stage 3: Run Locally

Terminal 1:

```bash
cd backend && make dev
```

Terminal 2:

```bash
cd frontend && bun run dev
```

Or run both together:

```bash
make dev
```

Local URLs:

- Frontend: `http://localhost:3000`
- Backend health: `http://localhost:5000/healthz`
- Swagger UI: `http://localhost:5000/swagger/ui`

## Stage 4: Validate Baseline Quality

```bash
make test
make validate-contracts
bash scripts/test-integration.sh
```

Use `make ci-local` when you need a local CI-equivalent pass.

## Stage 5: First Contribution Workflow

For backend endpoint changes, use strict sequence:

1. Update `backend/internal/swagger/openapi.yaml`.
2. Run `make validate-contracts` (initial failure is expected).
3. Add failing handler tests in `backend/internal/router/*_test.go`.
4. Implement handler in `backend/internal/router/*.go`.
5. Run router tests and full tests.
6. Update frontend MSW handlers in `frontend/src/mocks/handlers.ts`.
7. Run `make validate-contracts` until green.

For frontend features:

1. Add failing test in `frontend/src/lib/__tests__/`.
2. Implement feature.
3. Run `bun test`.
4. If backend call shape changed, sync MSW handlers and re-run contract validation.

## Stage 6: Deployment Readiness Decision

Before touching Ansible, decide deployment topology:

- Single host: use [Deployment Single Host](deployment-single-host.md).
- Split hosts (frontend/backend separated): use [Deployment Split Hosts](deployment-split-host.md).

Do not run default playbook in split-host environments without adaptation.

## Onboarding Exit Criteria

A developer is fully onboarded when all are true:

- Can run local stack and hit health endpoints.
- Can run `make test` and `make validate-contracts` successfully.
- Understands OpenAPI-first and test-first rules.
- Knows whether target environment is single-host or split-host.
- Understands current deployment caveats documented in this site.
