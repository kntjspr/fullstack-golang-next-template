#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

TEST_DATABASE_URL="${TEST_DATABASE_URL:-postgres://postgres:test@localhost:5433/testdb?sslmode=disable}"
TEST_REDIS_URL="${TEST_REDIS_URL:-redis://localhost:6380}"

echo "=== [1/7] Environment check ==="
bash scripts/check-env.sh

echo "=== [2/7] OpenAPI contract validation ==="
bash scripts/validate-openapi.sh

echo "=== [3/7] Backend unit tests ==="
TEST_DATABASE_URL="$TEST_DATABASE_URL" TEST_REDIS_URL="$TEST_REDIS_URL" go test ./backend/...

echo "=== [4/7] Frontend unit tests ==="
(
  cd frontend
  bun test
)

echo "=== [5/7] Contract tests ==="
TEST_DATABASE_URL="$TEST_DATABASE_URL" TEST_REDIS_URL="$TEST_REDIS_URL" go test ./backend/internal/contract/...

echo "=== [6/7] Security scan ==="
bash scripts/security-scan.sh

echo "=== [7/7] Build verification ==="
go build ./backend/...
(
  cd frontend
  bun run build
)

echo ""
echo "All checks passed. Safe to deploy."
