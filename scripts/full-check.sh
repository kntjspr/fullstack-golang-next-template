#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

source "$ROOT_DIR/scripts/test-env.sh"
load_test_environment "$ROOT_DIR/.env"

echo "=== [1/5] Environment check ==="
bash scripts/check-env.sh

echo "=== [2/5] OpenAPI contract validation ==="
bash scripts/validate-openapi.sh

echo "=== [3/5] Backend unit tests ==="
TEST_DATABASE_URL="$TEST_DATABASE_URL" TEST_REDIS_URL="$TEST_REDIS_URL" go test ./backend/...

echo "=== [4/5] Contract tests ==="
TEST_DATABASE_URL="$TEST_DATABASE_URL" TEST_REDIS_URL="$TEST_REDIS_URL" go test ./backend/internal/contract/...

echo "=== [5/5] Security scan ==="
bash scripts/security-scan.sh

echo ""
echo "All checks passed. Safe to deploy."
