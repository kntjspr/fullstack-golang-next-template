# Fullstack Golang Next Template Documentation

This docs site is the operational guide for developing, testing, and deploying this repository.

Use this as the source of truth for onboarding and deployment decisions. The root `README.md` is still useful, but this docs set is intentionally more explicit about assumptions, warnings, and production nuances.

## Read This First

- Frontend package manager is **Bun only**. Do not use npm/yarn.
- OpenAPI contract source is **only** `backend/internal/swagger/openapi.yaml`.
- Tests must use real Postgres/Redis from `docker-compose.test.yml` on ports `5433` and `6380`.
- Auth strategy is dual-mode: `Authorization: Bearer <token>` has priority; `auth_token` cookie is fallback.
- Default Ansible automation is single-host oriented.

## Critical Deployment Warnings

1. Current playbook/inventory (`playbook.yml`, `hosts.ini`) applies all roles to one host group (`cgapp_project`).
2. Nginx templates proxy to `cgapp-backend` container by name; this only works on the same Docker network/host.
3. Frontend API endpoint is build-time (`NEXT_PUBLIC_API_URL`), so API host changes require frontend rebuild/redeploy.
4. CORS middleware exists but is not mounted in `backend/main.go`.
5. Cookie `Secure` flag is controlled by `STAGE_STATUS`: `prod` enables it, `dev` disables it.

If you deploy frontend and backend on different hosts, read [Deployment Split Hosts](deployment-split-host.md) before touching Ansible.

## Documentation Map

- [Onboarding](onboarding.md): first-day and first-week process.
- [Local Development](local-development.md): day-to-day developer workflow.
- [Testing And Quality](testing-and-quality.md): test matrix, gates, and contract discipline.
- [Deployment Overview](deployment-overview.md): topology choices and tradeoffs.
- [Deployment Single Host](deployment-single-host.md): current default automation path.
- [Deployment Split Hosts](deployment-split-host.md): required changes for frontend/backend separation.
- [Environment Variables Reference](reference-environment.md): backend/frontend variable catalog.
- [Repository Map](reference-repository-map.md): key files and responsibilities.
- [Troubleshooting](troubleshooting.md): common failure patterns and fixes.

## Fast Path

If you just cloned the repo:

1. Read [Onboarding](onboarding.md).
2. Run `make bootstrap`.
3. Start stack with `make dev`.
4. Run `make test` and `make validate-contracts`.
5. Before deployment, pick topology from [Deployment Overview](deployment-overview.md).
