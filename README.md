# Create Go App Full-Stack Template
A production-ready full-stack template: Go backend + Next.js frontend, batteries included.

## 2. What's included
Backend: You get a Go API service with the chi router (a lightweight HTTP router and middleware stack), GORM (an ORM for working with SQL data in Go), Postgres for persistent data, Redis for fast cache/state operations, JWT-based authentication, and a single OpenAPI specification at `backend/internal/swagger/openapi.yaml`.

Frontend: You get a Next.js 15 app with TypeScript, Bun as the package manager/runtime, and MSW (Mock Service Worker) so frontend tests can mock API behavior through real request interception instead of brittle fake function stubs.

Infrastructure: You get Docker Compose for local test services, Ansible roles for provisioning and deployment, and GitHub Actions CI to run automated checks on pushes and pull requests.

Quality: You get a test-first workflow, contract tests that catch API drift between spec and implementation, coverage reporting, and security-focused middleware defaults such as hardened response headers.

## 3. Prerequisites
Install these tools before you start:

1. Go 1.22+ — install from https://go.dev/dl/
2. Bun — install from https://bun.sh
3. Docker Desktop — install from https://www.docker.com/products/docker-desktop/
4. Make — preinstalled on most macOS/Linux systems; on Windows, install with Chocolatey (`choco install make`) or use WSL
5. For deployment: a Linux server with SSH access

## 4. Quick start (local development)
1. Clone this repository:
```bash
git clone <your-repo-url>
cd go-project
```

2. Run bootstrap:
```bash
make bootstrap
```
- On first run: asks for your Go module name and renames it throughout the codebase automatically.
- On subsequent runs: skips first-time setup and goes straight to dependency install and test verification.
- After bootstrap completes, a `.bootstrap-done` file is created locally (git-ignored) so the first-time prompt never shows again.

3. Copy `.env.example` to `.env` and fill in values:
```bash
cp .env.example .env
```
Required variables:
- `DATABASE_URL`: connection string for your Postgres database
- `REDIS_URL`: connection string for your Redis instance
- `JWT_SECRET`: a random secret string used to sign auth tokens (generate one with `openssl rand -hex 32`)
- `APP_ENV`: set to `development` locally

4. Start local development:
```bash
make dev
```
This starts the backend on `:8080` and the frontend on `:3000`.

5. Open the app:
- http://localhost:3000

## 5. Running tests
- `make test` — runs all unit tests
- `make validate-contracts` — checks API spec matches implementation
- `bash scripts/test-integration.sh` — runs integration tests (needs Docker)
- `make coverage` — prints coverage report
- `make check` — full pre-deploy gate (run this before every deploy)
- API docs: http://localhost:8080/swagger/ui (available when backend is running)

## 6. Project structure
```text
go-project/
├── backend/                    # Go API server
│   ├── internal/               # Private packages (router, auth, DB, cache, config, contract tests)
│   ├── middleware/             # Shared HTTP middleware (auth, rate limit, headers, validation)
│   ├── migrations/             # SQL migration files
│   └── main.go                 # Backend entrypoint
├── frontend/                   # Next.js application
│   └── src/
│       ├── app/                # App Router pages, layout, sitemap, robots
│       ├── lib/                # API client and shared frontend utilities
│       └── mocks/              # MSW API mocks used in tests
├── roles/                      # Ansible deployment roles
├── docs/                       # Architecture docs, ADRs, and security notes
├── scripts/                    # Automation scripts (bootstrap, checks, integration helpers)
├── docker-compose.test.yml     # Postgres/Redis test infrastructure
├── Makefile                    # Common project commands
└── AGENTS.md                   # Contributor and AI agent workflow guide
```

## 7. Deployment
### What you need
- A Linux server (Ubuntu 22.04 recommended) with SSH access
- Your server's IP address or hostname
- A user with sudo privileges on the server

### Steps
1. Edit `hosts.ini` and replace the placeholder IP with your server IP:
```ini
[servers]
your.server.ip.here
```

2. Review `playbook.yml` if needed. It runs roles that provision Docker, Nginx, Postgres, Redis, the backend service, and the frontend deployment artifacts.

3. Set `APP_ENV=production` in your `.env`.

4. Run the full gate locally:
```bash
make check
```

5. Run the Ansible playbook:
```bash
ansible-playbook -i hosts.ini playbook.yml
```

6. After deploy:
- Nginx serves the frontend on ports 80/443
- Backend API runs on `:8080` behind Nginx
- Postgres and Redis run as Docker containers on the server

### First deploy checklist
- [ ] `hosts.ini` updated with real server IP
- [ ] `.env` has production values (strong `JWT_SECRET`, real `DATABASE_URL`)
- [ ] `make check` passes locally
- [ ] SSH key is set up for your server user
- [ ] Server has ports 80 and 443 open in firewall

## 8. Making your first change
1. Open `backend/internal/swagger/openapi.yaml` and add your new route.
2. Run `make validate-contracts` (it will fail; this is expected).
3. Write a failing test in `backend/internal/router/`.
4. Write the handler.
5. Run `make test` and confirm it passes.
6. Run `make validate-contracts` and confirm it passes.
7. Add the MSW mock in `frontend/src/mocks/handlers.ts`.
8. Commit.

## 9. Getting help
- Read `AGENTS.md` for contributor conventions
- Read `docs/ARCHITECTURE.md` for system design
- Read `docs/adr/` for why key decisions were made
- Open an issue on GitHub for bugs or questions

## 10. Authentication
The backend accepts two auth mechanisms at the same time:
- Bearer token in `Authorization: Bearer <token>`
- httpOnly cookie named `auth_token`

Priority order:
1. If a Bearer token is present, it is used first.
2. If no Bearer token is present, backend falls back to `auth_token` cookie.
3. If neither is present, request is rejected with `401`.

Browser apps: login sets `auth_token` automatically and requests send it automatically (with `credentials: "include"` in the API client), so no extra browser-side cookie handling is needed.

CLI clients:
```bash
curl -H "Authorization: Bearer <your-token>" https://yourapi.com/protected
```

MCP/programmatic clients use the same Bearer header pattern.

Logout behavior:
- `POST /auth/logout` clears the `auth_token` cookie.
- Bearer-only clients should discard stored tokens on the client side.

To get a token:
- `POST /auth/login` with `{email, password}`.
- Response returns token data in JSON and also sets the `auth_token` cookie.

XSS tradeoff:
- `localStorage` tokens are readable by JavaScript.
- httpOnly cookies are not readable by JavaScript.
- This template supports both: cookie flow for browsers, Bearer flow for CLI/MCP/API clients.

## 12. Customizing this template

### Analytics - Umami
Privacy-friendly, self-hostable, GDPR-compliant analytics.
- File to edit: `frontend/src/app/layout.tsx`
- Where: inside the `<head>` section of the root layout
- Add:
```html
<script
  defer
  src="https://your-umami-instance.com/script.js"
  data-website-id="your-website-id"
/>
```
- Replace `src` with your Umami instance URL and `data-website-id` with your dashboard website ID.
- Self-host docs: https://umami.is/docs/install

### Error tracking - Sentry
Captures runtime errors and performance traces.

Backend (already wired):
- File: `backend/internal/telemetry/sentry.go` (no code changes required)
- Set in `.env`:
```bash
SENTRY_DSN=your-dsn-from-sentry-dashboard
SENTRY_ENVIRONMENT=production
SENTRY_RELEASE=v1.0.0
```
- DSN location: Sentry dashboard -> Project -> Settings -> Client Keys

Frontend (not wired yet):
1. `cd frontend && bun add @sentry/nextjs`
2. `bunx @sentry/wizard@latest -i nextjs`
3. Set in `.env`: `NEXT_PUBLIC_SENTRY_DSN=your-dsn`
4. Docs: https://docs.sentry.io/platforms/javascript/guides/nextjs/

### Email - Resend (recommended)
Developer-friendly transactional email API.

No email code exists yet. Add it with:
1. Create `backend/internal/mailer/mailer.go`
2. Install SDK: `go get github.com/resendlabs/resend-go`
3. Set in `.env`: `RESEND_API_KEY=your-key`
4. Docs: https://resend.com/docs/send-with-go

### Payments - Stripe
No payment code exists yet. Add it with:
- Backend webhook handler: create `backend/internal/router/webhook.go`
- Frontend checkout page: create `frontend/src/app/checkout/page.tsx`
- Set in `.env`:
```bash
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...
```
- Docs: https://stripe.com/docs/webhooks

### File storage - S3 compatible (Supabase Storage, Cloudflare R2, AWS S3)
No storage code exists yet. Add it with:
1. Create `backend/internal/storage/storage.go`
2. Install SDK: `go get github.com/aws/aws-sdk-go-v2`
3. Set in `.env`:
```bash
STORAGE_ENDPOINT=https://your-endpoint
STORAGE_BUCKET=your-bucket
STORAGE_KEY=your-access-key
STORAGE_SECRET=your-secret-key
```

### OAuth / Social login
No OAuth code exists yet. Add it with:
- File to extend: `backend/internal/router/auth.go`
- Recommended library: `golang.org/x/oauth2`
- Add callback routes in `backend/internal/swagger/openapi.yaml` first

### Database GUI
To inspect Postgres locally:
- Recommended: TablePlus (https://tableplus.com) or pgAdmin
- Connection string: use `DATABASE_URL` from `.env`

## 13. Environment variable reference

| Variable | Required | Description |
| --- | --- | --- |
| APP_ENV | Required | `development` or `production`. Controls cookie Secure flag and graceful shutdown behavior. |
| DATABASE_URL | Required | Full Postgres connection string including host, port, user, password, and database name. |
| REDIS_URL | Required | Redis connection string. |
| JWT_SECRET | Required | Secret for signing auth tokens. Generate with: `openssl rand -hex 32`. |
| SERVER_HOST | Optional | Backend bind host. Default: `0.0.0.0`. |
| SERVER_PORT | Optional | Backend API port. Default: `8080`. |
| LOGGER_LEVEL | Optional | Log verbosity: `debug`, `info`, `warn`, `error`. |
| SENTRY_DSN | Optional | Sentry project DSN. Leave empty to disable error tracking. |
| SENTRY_ENVIRONMENT | Optional | Sentry environment tag, usually aligned with `APP_ENV`. |
| SENTRY_RELEASE | Optional | Release version tag sent to Sentry. |
| TEST_DATABASE_URL | Dev only | Postgres URL used by tests (local default points to port `5433`). |
| TEST_REDIS_URL | Dev only | Redis URL used by tests (local default points to port `6380`). |
| STAGE_STATUS | Optional | Backend runtime mode (`dev` or `prod`). |
| SERVER_READ_TIMEOUT | Optional | HTTP read timeout in seconds. |
| SERVER_WRITE_TIMEOUT | Optional | HTTP write timeout in seconds. |
| SERVER_IDLE_TIMEOUT | Optional | HTTP idle timeout in seconds. |
| LOGGER_PRETTY | Optional | Pretty console log output (`true`/`false`). |
| DB_ENABLE | Optional | Enables Postgres integration when `true`. |
| DB_HOST | Optional | Postgres host when `DB_ENABLE=true`. |
| DB_PORT | Optional | Postgres port when `DB_ENABLE=true`. |
| DB_USER | Optional | Postgres user when `DB_ENABLE=true`. |
| DB_PASSWORD | Optional | Postgres password when `DB_ENABLE=true`. |
| DB_NAME | Optional | Postgres database name when `DB_ENABLE=true`. |
| DB_SSL_MODE | Optional | Postgres SSL mode (`disable`, `require`, etc). |
| DB_TIMEZONE | Optional | Postgres timezone setting. |
| REDIS_ENABLE | Optional | Enables Redis integration when `true`. |
| REDIS_HOST | Optional | Redis host when `REDIS_ENABLE=true`. |
| REDIS_PORT | Optional | Redis port when `REDIS_ENABLE=true`. |
| REDIS_PASSWORD | Optional | Redis password when required. |
| REDIS_DB | Optional | Redis DB index. |
| SENTRY_TRACES_SAMPLE_RATE | Optional | Trace sample rate for Sentry performance events. |
| NEXT_PUBLIC_UMAMI_WEBSITE_ID | Optional | Umami site identifier used by the frontend script. |
| NEXT_PUBLIC_UMAMI_SCRIPT_URL | Optional | Umami script URL loaded by the frontend. |

Copy `.env.example` to `.env` and fill in all Required variables before running `make dev` or `make bootstrap`. Never commit your `.env` file - it is blocked by the pre-commit hook.


Fully ready and compatible for AI assisted development and modern workflows :D
