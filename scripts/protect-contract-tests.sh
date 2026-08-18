#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BASE_REF="${1:-origin/main}"
CONTRACT_FILE="backend/service_test.go"

cd "$ROOT_DIR"

if ! git rev-parse --verify --quiet "$BASE_REF" >/dev/null; then
  echo "contract protection requires an available base ref: $BASE_REF" >&2
  exit 1
fi

BASE_COMMIT="$(git merge-base "$BASE_REF" HEAD)"
if git diff --no-ext-diff "$BASE_COMMIT" -- "$CONTRACT_FILE" | grep -E '^-[^-]'; then
  echo "CRITICAL ERROR: existing contract test content was modified or deleted" >&2
  exit 1
fi
