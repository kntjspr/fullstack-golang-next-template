# ADR 0002: Use GORM with Postgres

- Status: Accepted

## Context
The project needs a relational database and a Go data access approach that balances development speed with production-grade control.

## Decision
Use GORM as the ORM and Postgres as the primary datastore.

## Consequences
- Rapid model-driven setup and test bootstrapping through GORM APIs.
- `AutoMigrate` is used in test setup for fast schema alignment.
- Production schema evolution is controlled through raw SQL migrations in `backend/migrations/`.
- Developers must keep model structs and SQL migrations aligned.
