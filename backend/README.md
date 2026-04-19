# go-chi backend template for [Create Go App CLI](https://github.com/create-go-app/cli)

<img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go" alt="go version" />&nbsp;<a href="https://goreportcard.com/report/github.com/create-go-app/fiber-go-template" target="_blank"><img src="https://img.shields.io/badge/Go_report-A+-success?style=for-the-badge&logo=none" alt="go report" /></a>&nbsp;<img src="https://img.shields.io/badge/license-Apache_2.0-red?style=for-the-badge&logo=none" alt="license" />

[Chi](https://go-chi.io/#/) is a lightweight, idiomatic and composable router for building Go HTTP services.

## ⚡️ Quick start

1. Create a new project with Fiber:

```bash
cgapp create

# Choose a backend framework:
#   net/http
#   fiber
# > chi
```

2. Rename `.env.example` to `.env` and fill it with your environment values.
3. Install [Docker](https://www.docker.com/get-started) and the following useful Go tools to your system:

- [golang-migrate/migrate](https://github.com/golang-migrate/migrate#cli-usage) for apply migrations
- [github.com/securego/gosec](https://github.com/securego/gosec) for checking Go security issues
- [github.com/go-critic/go-critic](https://github.com/go-critic/go-critic) for checking Go the best practice issues
- [github.com/golangci/golangci-lint](https://github.com/golangci/golangci-lint) for checking Go linter issues

4. Run project by this command:

```bash
make docker.run
```

5. Go to [127.0.0.1:5000/hc/status](http://127.0.0.1:5000/hc/status) and see `HTTP 200 OK`.

## 📦 Used packages

| Name                                                                  | Version  | Type       |
|-----------------------------------------------------------------------| -------- | ---------- |
| [go-chi/chi](https://github.com/go-chi/chi)                           | `v5.0.7` | core       |
| [joho/godotenv](https://github.com/joho/godotenv)                     | `v1.4.0` | config     |
| [go-chi/render](https://github.com/go-chi/render)                     | `v1.x`   | rendering  |
| [rs/zerolog](https://github.com/rs/zerolog)                           | `v1.x`   | logger     |
| [gorm.io/gorm](https://github.com/go-gorm/gorm)                       | `v1.x`   | ORM        |
| [gorm.io/driver/postgres](https://github.com/go-gorm/postgres)        | `v1.x`   | db driver  |
| [go-redis/redis](https://github.com/redis/go-redis)                   | `v9.x`   | cache      |
| [getsentry/sentry-go](https://github.com/getsentry/sentry-go)         | `v0.x`   | telemetry  |

## ⚙️ Configuration

```ini
# .env

# Stage status to start server:
#   - "dev", for start server without graceful shutdown
#   - "prod", for start server with graceful shutdown
STAGE_STATUS="dev"

# Server settings:
SERVER_HOST="0.0.0.0"
SERVER_PORT=5000
SERVER_READ_TIMEOUT=5
SERVER_WRITE_TIMEOUT=10
SERVER_IDLE_TIMEOUT=120

# Logger settings:
LOGGER_LEVEL="info"
LOGGER_PRETTY=false

# Database settings:
DB_ENABLE=false
DB_HOST="127.0.0.1"
DB_PORT=5432
DB_USER="postgres"
DB_PASSWORD="password"
DB_NAME="postgres"
DB_SSL_MODE="disable"
DB_TIMEZONE="UTC"

# Redis settings:
REDIS_ENABLE=false
REDIS_HOST="127.0.0.1"
REDIS_PORT=6379
REDIS_PASSWORD=""
REDIS_DB=0

# Sentry settings:
SENTRY_DSN=""
SENTRY_ENVIRONMENT="development"
SENTRY_RELEASE=""
SENTRY_TRACES_SAMPLE_RATE=0
```

This project is based on create-go-app/chi-go-template
by Vic Shóstak, licensed under Apache 2.0.
