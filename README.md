# fullstack-golang-next-template


production-ready full-stack template. go backend + next.js frontend. batteries included.

---

## what's included

**backend**

- `chi` — lightweight http router with middleware support
- `gorm` — sql orm for go
- `postgres` — persistent storage
- `redis` — cache and ephemeral state
- jwt authentication
- single openapi spec at `backend/internal/swagger/openapi.yaml`

**frontend**

- next.js 15 with typescript
- bun as package manager and runtime
- msw (mock service worker) for request-level api mocking in tests — no brittle function stubs

**infrastructure**

- docker compose for local test services
- ansible roles for provisioning and deployment
- github actions ci on push and pr

**quality**

- test-first workflow
- contract tests that catch drift between openapi spec and implementation
- coverage reporting
- security-focused middleware defaults (hardened response headers)

---

## prerequisites

- **go 1.22+** — https://go.dev/dl/
- **bun** — https://bun.sh
- **docker desktop** — https://www.docker.com/products/docker-desktop/
- **make** — preinstalled on macos/linux; windows: `choco install make` or use wsl
- **deployment only**: a linux server with ssh access

---

## quick start

**1. clone**

```bash
git clone <your-repo-url>
cd go-project
```

**2. bootstrap**

```bash
make bootstrap
```

first run: prompts for your go module name and renames it throughout the codebase.  
subsequent runs: skips setup, goes straight to dependency install and test verification.  
completion writes `.bootstrap-done` locally (git-ignored) so the prompt never repeats.

**3. configure env**

```bash
cp .env.example .env
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env
```

required variables:

| variable | description |
|---|---|
| `DATABASE_URL` | postgres connection string |
| `REDIS_URL` | redis connection string |
| `JWT_SECRET` | token signing secret — generate with `openssl rand -hex 32` |
| `APP_ENV` | set to `development` locally |

**4. start dev**

```bash
make dev
```

backend on `:5000`, frontend on `:3000`.

**5. open**

- app: http://localhost:3000
- api docs (swagger ui): http://localhost:5000/swagger/ui

---

## tests

```bash
make test                          # unit tests
make validate-contracts            # openapi spec vs implementation
bash scripts/test-integration.sh   # integration tests (requires docker)
make coverage                      # coverage report
make check                         # full pre-deploy gate — run this before every deploy
```

---

## project structure

```text
go-project/
├── backend/                    # go api server
│   ├── internal/               # private packages (router, auth, db, cache, config, contract tests)
│   ├── middleware/             # shared http middleware (auth, rate limit, headers, validation)
│   ├── migrations/             # sql migration files
│   └── main.go                 # backend entrypoint
├── frontend/                   # next.js application
│   └── src/
│       ├── app/                # app router pages, layout, sitemap, robots
│       ├── lib/                # api client and shared frontend utilities
│       └── mocks/              # msw api mocks used in tests
├── roles/                      # ansible deployment roles
├── docs/                       # architecture docs, adrs, and security notes
├── scripts/                    # automation scripts (bootstrap, checks, integration helpers)
├── docker-compose.test.yml     # postgres/redis test infrastructure
├── Makefile                    # common project commands
└── AGENTS.md                   # contributor and ai agent workflow guide
```

---

## deployment

### what you need

- linux server (ubuntu 22.04 recommended) with ssh access
- server ip address or hostname
- user with sudo privileges on the server

### steps

**1. update hosts**

edit `hosts.ini` and replace the placeholder with your server ip:

```ini
[servers]
your.server.ip.here
```

**2. review playbook**

`playbook.yml` runs roles that provision: docker, nginx, postgres, redis, the backend service, and the frontend deployment artifacts. review if needed.

**3. set production env**

```bash
APP_ENV=production
```

**4. run the pre-deploy gate**

```bash
make check
```

**5. deploy**

```bash
ansible-playbook -i hosts.ini playbook.yml
```

**6. after deploy**

- nginx serves the frontend on ports 80/443
- backend api runs on `:5000` behind nginx
- postgres and redis run as docker containers on the server

### first deploy checklist

- [ ] `hosts.ini` updated with real server ip
- [ ] `.env` has production values (strong `JWT_SECRET`, real `DATABASE_URL`)
- [ ] `make check` passes locally
- [ ] ssh key configured for your server user
- [ ] ports 80 and 443 open in server firewall

---

## making your first change

1. open `backend/internal/swagger/openapi.yaml` and add your new route
2. run `make validate-contracts` — it will fail; this is expected
3. write a failing test in `backend/internal/router/`
4. write the handler
5. run `make test` — confirm it passes
6. run `make validate-contracts` — confirm it passes
7. add the msw mock in `frontend/src/mocks/handlers.ts`
8. commit

---

## authentication

the backend accepts two auth mechanisms simultaneously:

- bearer token via `Authorization: Bearer <token>` header
- httponly cookie named `auth_token`

**priority order**

1. bearer token present → use it
2. no bearer token → fall back to `auth_token` cookie
3. neither present → `401`

**browser clients**

login sets `auth_token` automatically. requests send it automatically with `credentials: "include"` in the api client. no extra cookie handling needed.

**cli clients**

```bash
curl -H "Authorization: Bearer <your-token>" https://yourapi.com/protected
```

mcp and programmatic clients use the same bearer header pattern.

**getting a token**

`POST /auth/login` with `{ email, password }` — returns token data as json and sets the `auth_token` cookie.

**logout**

`POST /auth/logout` clears the `auth_token` cookie. bearer-only clients should discard the stored token client-side.

**xss tradeoff**

- `localStorage` tokens are readable by javascript
- httponly cookies are not
- this template supports both: cookie flow for browsers, bearer flow for cli/mcp/api clients

---

## customization

### analytics — umami

privacy-friendly, self-hostable, gdpr-compliant.

file: `frontend/src/app/layout.tsx` → inside the `<head>` of root layout

```html
<script
  defer
  src="https://your-umami-instance.com/script.js"
  data-website-id="your-website-id"
/>
```

replace `src` with your umami instance url and `data-website-id` with your dashboard id.  
self-host docs: https://umami.is/docs/install

---

### error tracking — sentry

**backend** (already wired)

file: `backend/internal/telemetry/sentry.go` — no code changes needed. set in `.env`:

```bash
SENTRY_DSN=your-dsn-from-sentry-dashboard
SENTRY_ENVIRONMENT=production
SENTRY_RELEASE=v1.0.0
```

dsn location: sentry dashboard → project → settings → client keys

**frontend** (not wired yet)

```bash
cd frontend && bun add @sentry/nextjs
bunx @sentry/wizard@latest -i nextjs
```

set in `.env`: `NEXT_PUBLIC_SENTRY_DSN=your-dsn`  
docs: https://docs.sentry.io/platforms/javascript/guides/nextjs/

---

### email — resend

no email code exists yet. add it with:

1. create `backend/internal/mailer/mailer.go`
2. install sdk: `go get github.com/resendlabs/resend-go`
3. set in `.env`: `RESEND_API_KEY=your-key`

docs: https://resend.com/docs/send-with-go

---

### payments — stripe

no payment code exists yet. add it with:

- backend webhook handler: `backend/internal/router/webhook.go`
- frontend checkout page: `frontend/src/app/checkout/page.tsx`
- set in `.env`:

```bash
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...
```

docs: https://stripe.com/docs/webhooks

---

### file storage — s3-compatible

works with supabase storage, cloudflare r2, aws s3. no storage code exists yet. add it with:

1. create `backend/internal/storage/storage.go`
2. install sdk: `go get github.com/aws/aws-sdk-go-v2`
3. set in `.env`:

```bash
STORAGE_ENDPOINT=https://your-endpoint
STORAGE_BUCKET=your-bucket
STORAGE_KEY=your-access-key
STORAGE_SECRET=your-secret-key
```

---

### oauth / social login

no oauth code exists yet. add it with:

- extend: `backend/internal/router/auth.go`
- recommended library: `golang.org/x/oauth2`
- add callback routes in `backend/internal/swagger/openapi.yaml` first

---

### database gui

inspect postgres locally with [tableplus](https://tableplus.com) or pgadmin. use `DATABASE_URL` from `.env` as the connection string.

---

## environment variables

| variable | required | description |
|---|---|---|
| `APP_ENV` | required | `development` or `production`. controls cookie secure flag and graceful shutdown behavior. |
| `DATABASE_URL` | required | full postgres connection string including host, port, user, password, and database name. |
| `REDIS_URL` | required | redis connection string. |
| `JWT_SECRET` | required | secret for signing auth tokens. generate with `openssl rand -hex 32`. |
| `SERVER_HOST` | optional | backend bind host. default: `0.0.0.0`. |
| `SERVER_PORT` | optional | backend api port. default: `5000`. |
| `LOGGER_LEVEL` | optional | log verbosity: `debug`, `info`, `warn`, `error`. |
| `LOGGER_PRETTY` | optional | pretty console log output (`true`/`false`). |
| `SENTRY_DSN` | optional | sentry project dsn. leave empty to disable. |
| `SENTRY_ENVIRONMENT` | optional | sentry environment tag, usually aligned with `APP_ENV`. |
| `SENTRY_RELEASE` | optional | release version tag sent to sentry. |
| `SENTRY_TRACES_SAMPLE_RATE` | optional | trace sample rate for sentry performance events. |
| `TEST_DATABASE_URL` | dev only | postgres url used by tests. local default points to port `5433`. |
| `TEST_REDIS_URL` | dev only | redis url used by tests. local default points to port `6380`. |
| `STAGE_STATUS` | optional | backend runtime mode (`dev` or `prod`). |
| `SERVER_READ_TIMEOUT` | optional | http read timeout in seconds. |
| `SERVER_WRITE_TIMEOUT` | optional | http write timeout in seconds. |
| `SERVER_IDLE_TIMEOUT` | optional | http idle timeout in seconds. |
| `DB_ENABLE` | optional | enables postgres integration when `true`. |
| `DB_HOST` | optional | postgres host when `DB_ENABLE=true`. |
| `DB_PORT` | optional | postgres port when `DB_ENABLE=true`. |
| `DB_USER` | optional | postgres user when `DB_ENABLE=true`. |
| `DB_PASSWORD` | optional | postgres password when `DB_ENABLE=true`. |
| `DB_NAME` | optional | postgres database name when `DB_ENABLE=true`. |
| `DB_SSL_MODE` | optional | postgres ssl mode (`disable`, `require`, etc). |
| `DB_TIMEZONE` | optional | postgres timezone setting. |
| `REDIS_ENABLE` | optional | enables redis integration when `true`. |
| `REDIS_HOST` | optional | redis host when `REDIS_ENABLE=true`. |
| `REDIS_PORT` | optional | redis port when `REDIS_ENABLE=true`. |
| `REDIS_PASSWORD` | optional | redis password when required. |
| `REDIS_DB` | optional | redis db index. |
| `NEXT_PUBLIC_UMAMI_WEBSITE_ID` | optional | umami site identifier used by the frontend script. |
| `NEXT_PUBLIC_UMAMI_SCRIPT_URL` | optional | umami script url loaded by the frontend. |

copy `.env.example` to `.env` and fill in all required variables before running `make dev` or `make bootstrap`. never commit `.env` — blocked by the pre-commit hook.

---

## getting help

- `AGENTS.md` — contributor conventions and ai agent workflow guide
- `docs/ARCHITECTURE.md` — system design overview
- `docs/adr/` — decision records explaining why key choices were made
- github issues — bugs and questions
