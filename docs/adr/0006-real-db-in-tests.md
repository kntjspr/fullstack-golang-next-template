# ADR 0006: Use Real Postgres and Redis in Tests

- Status: Accepted

## Context
In-memory substitutes (for example SQLite/miniredis) can hide behavior differences and reduce confidence in test outcomes.

## Decision
Use real Postgres and Redis via `docker-compose.test.yml` for backend unit, contract, and integration workflows; do not use SQLite or miniredis.

## Consequences
- Test behavior has high parity with production data stores and protocols.
- Startup and execution are slightly slower than fully in-memory tests.
- Compatibility gaps and false positives are reduced significantly.
- `docker-compose.test.yml` must be available and running for backend test execution.
