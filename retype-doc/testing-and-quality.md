# Testing And Quality

This repository enforces quality with layered tests and contract checks.

## Test Matrix

- Unit tests:
  - Backend: `go test ./backend/...`
  - Frontend: `bun test`
- Contract validation:
  - OpenAPI lint (`backend/scripts/validate-openapi.sh`)
  - Contract tests (`go test ./backend/internal/contract -count=1`)
- Integration tests:
  - `go test -tags integration ./backend/internal/integration/...`

## Recommended Sequence Before PR

```bash
docker compose -f docker-compose.test.yml up -d
make test
make validate-contracts
bash scripts/test-integration.sh
make lint
```

For full CI-equivalent validation:

```bash
make ci-local
```

## OpenAPI Rules

- Canonical file is `backend/internal/swagger/openapi.yaml`.
- `backend/internal/swagger/spec.go` embeds this path at compile time.
- Moving the file breaks build and contract pipeline.

## Red-Green Protocol

1. Add/modify test and confirm assertion failure.
2. Implement minimal code to pass.
3. Re-run tests.
4. Update contract/mocks if behavior changed.

Compile errors are not accepted as red phase; tests must execute and fail assertions.

## CI Gates Summary

- `unit-tests`: every push/PR
- `contract-tests`: every push/PR
- `security`: every push/PR
- `integration-tests`: push to `main` and PRs targeting `main`

Do not merge while required jobs are red.
