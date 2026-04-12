# ADR 0003: JWT-Based Authentication

- Status: Accepted

## Context
The API requires stateless authentication suitable for horizontally scaled services.

## Decision
Use JWT access tokens with either RS256 or HS256 signing, short token expiry, and a refresh-token style renewal pattern.

## Consequences
- Stateless auth scales cleanly across multiple API instances.
- Access tokens can remain short-lived to reduce risk exposure.
- Refresh token rotation policy is required to manage replay and token theft risks.
- Signing key/secret management becomes a critical operational control.
