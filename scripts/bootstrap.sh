#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="$ROOT_DIR/docker-compose.test.yml"
BOOTSTRAP_DONE_FILE="$ROOT_DIR/.bootstrap-done"

STEP=""
trap 'echo "error: failed during step: $STEP" >&2' ERR

TEST_DATABASE_URL="postgres://postgres:test@localhost:5433/testdb?sslmode=disable"
TEST_REDIS_URL="redis://localhost:6380"
FIRST_TIME_RAN=false

if [ ! -f "$BOOTSTRAP_DONE_FILE" ]; then
  FIRST_TIME_RAN=true

  echo ""
  echo "=== First-time setup ==="
  echo "This appears to be a fresh clone. Let's configure your project."
  echo ""

  # Prompt for new module name
  echo "Current Go module name: https://github.com/kntjspr/fullstack-golang-next-template"
  echo "Project Author: Kent Jasper Sisi <kntjspr@pm.me>"
  echo "Enter your new Go module name (e.g. github.com/yourname/myproject)"
  echo "Press Enter to keep the current name:"
  read -r NEW_MODULE_NAME

  if [ -n "$NEW_MODULE_NAME" ]; then
    echo "==> Renaming module to: $NEW_MODULE_NAME"

    OLD_MODULE="github.com/kntjspr/fullstack-golang-next-template"

    # Replace in all .go files
    find "$ROOT_DIR/backend" -type f -name "*.go" \
      -not -path "$ROOT_DIR/backend/.gocache/*" \
      -not -path "$ROOT_DIR/backend/.gomodcache/*" \
      -not -path "$ROOT_DIR/backend/.gopath/*" | \
      xargs sed -i "s|$OLD_MODULE|$NEW_MODULE_NAME|g"

    # Replace in go.mod
    sed -i "s|$OLD_MODULE|$NEW_MODULE_NAME|g" "$ROOT_DIR/backend/go.mod"

    # Replace in go.work
    if [ -f "$ROOT_DIR/go.work" ]; then
      sed -i "s|$OLD_MODULE|$NEW_MODULE_NAME|g" "$ROOT_DIR/go.work"
    fi

    # Replace in AGENTS.md section 0
    sed -i "s|$OLD_MODULE|$NEW_MODULE_NAME|g" "$ROOT_DIR/AGENTS.md"

    # Replace in README.md
    sed -i "s|$OLD_MODULE|$NEW_MODULE_NAME|g" "$ROOT_DIR/README.md"

    # Replace in any frontend files that reference the module
    FRONTEND_MATCHES="$(
      find "$ROOT_DIR/frontend" -type f \( -name "*.ts" -o -name "*.tsx" -o -name "*.json" \) \
        -not -path "*/node_modules/*" | \
        xargs grep -l "$OLD_MODULE" 2>/dev/null || true
    )"
    if [ -n "$FRONTEND_MATCHES" ]; then
      printf "%s\n" "$FRONTEND_MATCHES" | xargs sed -i "s|$OLD_MODULE|$NEW_MODULE_NAME|g"
    fi

    # Update go.sum by running go mod tidy
    echo "==> Running go mod tidy after module rename..."
    cd "$ROOT_DIR/backend" && go mod tidy && cd "$ROOT_DIR"

    echo "✓ Module renamed to $NEW_MODULE_NAME"
  else
    echo "==> Keeping existing module name."
  fi

  # Create the indicator file
  cat > "$BOOTSTRAP_DONE_FILE" << 'EOF'
# This file marks that first-time bootstrap setup has been completed.
# It was created by scripts/bootstrap.sh on first run.
# Do not delete this file unless you want bootstrap to re-run first-time setup.
# This file is intentionally ignored by git (see .gitignore).
EOF

  echo "✓ First-time setup complete."
  echo ""
else
  echo "==> First-time setup already done (found .bootstrap-done). Skipping."
  echo ""
fi

check_command() {
  local cmd="$1"
  local install_hint="$2"

  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "error: required command '$cmd' not found. $install_hint" >&2
    exit 1
  fi
}

check_go_version() {
  local version raw major minor

  raw="$(go version | awk '{print $3}')"
  version="${raw#go}"
  major="${version%%.*}"
  minor="${version#*.}"
  minor="${minor%%.*}"

  if [ -z "$major" ] || [ -z "$minor" ]; then
    echo "error: unable to parse Go version from: $raw" >&2
    exit 1
  fi

  if [ "$major" -lt 1 ] || { [ "$major" -eq 1 ] && [ "$minor" -lt 22 ]; }; then
    echo "error: Go 1.22+ is required. Found: $raw" >&2
    exit 1
  fi
}

wait_for_postgres() {
  local container_id attempt

  container_id="$(docker compose -f "$COMPOSE_FILE" ps -q postgres)"
  if [ -z "$container_id" ]; then
    echo "error: postgres container not found" >&2
    exit 1
  fi

  for attempt in $(seq 1 30); do
    if docker exec "$container_id" pg_isready -U postgres -d testdb >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done

  echo "error: postgres did not become healthy within 30 seconds" >&2
  exit 1
}

wait_for_redis() {
  local container_id attempt

  container_id="$(docker compose -f "$COMPOSE_FILE" ps -q redis)"
  if [ -z "$container_id" ]; then
    echo "error: redis container not found" >&2
    exit 1
  fi

  for attempt in $(seq 1 30); do
    if [ "$(docker exec "$container_id" redis-cli ping 2>/dev/null)" = "PONG" ]; then
      return 0
    fi
    sleep 1
  done

  echo "error: redis did not become healthy within 30 seconds" >&2
  exit 1
}

run_backend_connectivity_check() {
  local backend_pid=0 started=0 ready=0 attempt
  local log_file

  log_file="$(mktemp)"

  cleanup() {
    if [ "$backend_pid" -ne 0 ] && kill -0 "$backend_pid" >/dev/null 2>&1; then
      kill "$backend_pid" >/dev/null 2>&1 || true
      wait "$backend_pid" >/dev/null 2>&1 || true
    fi
    rm -f "$log_file"
  }

  trap cleanup RETURN

  (
    cd "$ROOT_DIR"
    env \
      SERVER_HOST=127.0.0.1 \
      SERVER_PORT=5001 \
      SERVER_READ_TIMEOUT=5 \
      SERVER_WRITE_TIMEOUT=10 \
      SERVER_IDLE_TIMEOUT=120 \
      LOGGER_LEVEL=info \
      LOGGER_PRETTY=false \
      DB_ENABLE=true \
      DB_HOST=127.0.0.1 \
      DB_PORT=5433 \
      DB_USER=postgres \
      DB_PASSWORD=test \
      DB_NAME=testdb \
      DB_SSL_MODE=disable \
      DB_TIMEZONE=UTC \
      REDIS_ENABLE=true \
      REDIS_HOST=127.0.0.1 \
      REDIS_PORT=6380 \
      REDIS_PASSWORD= \
      REDIS_DB=0 \
      SENTRY_DSN= \
      SENTRY_ENVIRONMENT=development \
      SENTRY_RELEASE= \
      SENTRY_TRACES_SAMPLE_RATE=0 \
      JWT_SECRET=bootstrap-secret \
      go run ./backend >"$log_file" 2>&1
  ) &
  backend_pid=$!
  started=1

  for attempt in $(seq 1 30); do
    if curl -fsS http://127.0.0.1:5001/healthz >/dev/null 2>&1; then
      ready=1
      break
    fi

    if [ "$started" -eq 1 ] && ! kill -0 "$backend_pid" >/dev/null 2>&1; then
      break
    fi

    sleep 1
  done

  if [ "$ready" -ne 1 ]; then
    echo "error: backend connectivity check failed" >&2
    echo "backend log:" >&2
    sed -n '1,160p' "$log_file" >&2 || true
    exit 1
  fi
}

cd "$ROOT_DIR"

echo "[1/11] Checking prerequisites"
STEP="Checking prerequisites"
check_command go "Install Go 1.22 or newer."
check_go_version
check_command bun "Install Bun from https://bun.sh/."
bun --version >/dev/null
check_command docker "Install Docker Engine/Desktop."
docker info >/dev/null
check_command make "Install GNU Make."
make --version >/dev/null
docker compose version >/dev/null

echo "[2/11] Configuring environment"
STEP="Configuring environment"
if [ -f "$ROOT_DIR/.env" ]; then
  echo "skipping .env (already exists)"
else
  if [ ! -f "$ROOT_DIR/.env.example" ]; then
    echo "error: .env.example not found at repo root" >&2
    exit 1
  fi
  cp "$ROOT_DIR/.env.example" "$ROOT_DIR/.env"
fi

echo "[3/11] Installing git hooks"
STEP="Installing git hooks"
echo "==> Installing git hooks..."
bash scripts/install-hooks.sh

echo "[4/11] Installing frontend dependencies"
STEP="Installing frontend dependencies"
(cd "$ROOT_DIR/frontend" && bun install)

echo "[5/11] Starting test infrastructure"
STEP="Starting test infrastructure"
docker compose -f "$COMPOSE_FILE" up -d

echo "[6/11] Waiting for Postgres health check"
STEP="Waiting for Postgres health check"
wait_for_postgres

echo "[7/11] Waiting for Redis health check"
STEP="Waiting for Redis health check"
wait_for_redis

echo "[8/11] Running backend DB connectivity check"
STEP="Running backend DB connectivity check"
run_backend_connectivity_check

echo "[9/11] Validating OpenAPI contracts"
STEP="Validating OpenAPI contracts"
make validate-contracts

echo "[10/11] Running backend tests"
STEP="Running backend tests"
TEST_DATABASE_URL="$TEST_DATABASE_URL" TEST_REDIS_URL="$TEST_REDIS_URL" go test ./backend/...

echo "[11/11] Running frontend tests"
STEP="Running frontend tests"
(cd "$ROOT_DIR/frontend" && bun test)

echo ""
echo "✓ Prerequisites checked"
echo "✓ Environment configured"
echo "✓ Git hooks installed"
echo "✓ Dependencies installed"
echo "✓ Test infrastructure running"
echo "✓ Backend tests passed"
echo "✓ Frontend tests passed"
echo "✓ Contracts validated"
if [ "$FIRST_TIME_RAN" = true ]; then
  echo "✓ First-time setup complete (module name configured)"
fi
echo "Ready. Run 'make dev' to start the development server."
