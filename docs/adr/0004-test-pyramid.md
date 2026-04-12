# ADR 0004: Adopt a 70/20/10 Test Pyramid

- Status: Accepted

## Context
The project needs a test strategy that preserves fast feedback while still catching integration and contract regressions.

## Decision
Adopt a test pyramid of approximately 70% unit tests, 20% contract tests, and 10% integration tests.

## Consequences
- Most local runs stay fast due to unit-test dominance.
- Contract tests catch API drift between OpenAPI, backend handlers, and frontend mocks.
- Integration tests provide production-like confidence and run on main-branch critical paths.
- CI complexity increases slightly, but confidence at merge time is substantially better.
