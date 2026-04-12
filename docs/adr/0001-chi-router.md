# ADR 0001: Use chi Router

- Status: Accepted

## Context
The backend needs an HTTP router that is lightweight, composable, and compatible with idiomatic net/http middleware patterns.

## Decision
Adopt `chi` as the backend HTTP router instead of gorilla/mux, gin, or echo.

## Consequences
- Lightweight dependency with strong stdlib compatibility.
- Easy middleware composition for security, auth, validation, and observability.
- Lower abstraction overhead than full framework routers.
- Team keeps explicit control over request/response handling.
