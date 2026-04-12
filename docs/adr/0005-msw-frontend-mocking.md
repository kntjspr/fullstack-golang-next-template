# ADR 0005: Use MSW for Frontend API Mocking

- Status: Accepted

## Context
Frontend tests need realistic API behavior without brittle mocking tied to implementation internals.

## Decision
Use MSW request interception for frontend tests instead of `jest.mock()` API stubs.

## Consequences
- Frontend tests exercise the real `api-client` request path and response handling.
- Mock handlers in `frontend/src/mocks/handlers.ts` mirror OpenAPI-backed backend routes.
- Contract drift is surfaced earlier by tests that run through HTTP-level behavior.
- Handler maintenance discipline is required whenever API contracts change.
