# Troubleshooting

Common issues and direct fixes.

## Bootstrap Fails On Go Version

Symptom:

- `error: Go 1.22+ is required`

Fix:

- Upgrade Go and confirm with `go version`.

## Contract Validation Fails With OpenAPI Warnings

Symptom:

- `make validate-contracts` fails even when lint output says warning.

Cause:

- `backend/scripts/validate-openapi.sh` treats warnings as failures.

Fix:

- Resolve all OpenAPI warnings, not only errors.

## Frontend Cannot Reach API In Production

Symptom:

- Network errors or requests still target old API host.

Cause:

- `NEXT_PUBLIC_API_URL` is build-time for static export.

Fix:

1. Update `frontend/.env` with correct API origin.
2. Rebuild frontend (`bun run build`).
3. Redeploy frontend artifacts.

## Cross-Origin Auth Fails After Login

Symptom:

- Login response succeeds but subsequent authenticated requests fail.

Likely causes:

- CORS middleware not mounted.
- Cookie flags incompatible with cross-origin use.
- Backend origin not allowlisted.

Fix:

1. Mount/configure CORS middleware in backend.
2. Verify cookie `Secure` and `SameSite` policy for cross-origin mode.
3. Confirm frontend requests use `credentials: "include"` when cookie auth is intended.

## Secure Cookie Not Set In Production

Symptom:

- `auth_token` cookie appears without `Secure`.

Cause:

- Cookie secure flag checks `STAGE_STATUS=prod` in auth handler.

Fix:

- Set `STAGE_STATUS=prod` in backend runtime.

## Nginx On Frontend Host Cannot Proxy Backend

Symptom:

- 502 from frontend host when proxying `/healthz` or swagger paths.

Cause:

- Template proxy target is `cgapp-backend` container name, unreachable across hosts.

Fix:

1. Remove API proxy locations on frontend-only host, or
2. Point proxy to reachable backend origin, or
3. Keep backend local to that host/network.

## Tests Fail Due To Missing Postgres/Redis

Symptom:

- backend tests fail on connection errors.

Fix:

```bash
docker compose -f docker-compose.test.yml up -d
export TEST_DATABASE_URL=postgres://postgres:test@localhost:5433/testdb?sslmode=disable
export TEST_REDIS_URL=redis://localhost:6380
make test
```
