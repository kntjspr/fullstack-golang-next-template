# Environment Variables Reference

This page consolidates backend/frontend environment variables used by runtime and build.

## Backend Core Runtime

- `STAGE_STATUS`: `dev` or `prod` server mode.
- `SERVER_HOST`: bind host.
- `SERVER_PORT`: bind port.
- `SERVER_READ_TIMEOUT`, `SERVER_WRITE_TIMEOUT`, `SERVER_IDLE_TIMEOUT`: HTTP timeouts in seconds.
- `LOGGER_LEVEL`, `LOGGER_PRETTY`: logging behavior.
- `JWT_SECRET`: JWT signing secret.

## Backend Database

- `DB_ENABLE`
- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `DB_SSL_MODE`
- `DB_TIMEZONE`

## Backend Redis

- `REDIS_ENABLE`
- `REDIS_HOST`
- `REDIS_PORT`
- `REDIS_PASSWORD`
- `REDIS_DB`

## Backend Telemetry

- `SENTRY_DSN`
- `SENTRY_ENVIRONMENT`
- `SENTRY_RELEASE`
- `SENTRY_TRACES_SAMPLE_RATE`

## Frontend Build-Time Public Variables

- `NEXT_PUBLIC_API_URL`: API base URL used by `frontend/src/lib/api-client.ts`.
- `NEXT_PUBLIC_SITE_URL`: canonical site URL for sitemap/robots.
- `NEXT_PUBLIC_UMAMI_WEBSITE_ID`: optional analytics website id.
- `NEXT_PUBLIC_UMAMI_SCRIPT_URL`: optional analytics script URL.

## Test Runtime Variables

- `TEST_POSTGRES_PORT`: host port for test Postgres. Defaults to `5433`; bootstrap selects the next free port when occupied and stores it in the root `.env`.
- `TEST_REDIS_PORT`: host port for test Redis. Defaults to `6380`; bootstrap selects the next free port when occupied and stores it in the root `.env`.
- `TEST_DATABASE_URL`: test database connection URL. The default uses `TEST_POSTGRES_PORT`.
- `TEST_REDIS_URL`: test Redis connection URL. The default uses `TEST_REDIS_PORT`.

These are used by test scripts and Make targets.

## Auth Cookie Security

Auth cookie security checks `STAGE_STATUS` in `backend/internal/router/auth.go`.

- `STAGE_STATUS=prod` -> secure cookie true
- `STAGE_STATUS=dev` -> secure cookie false
