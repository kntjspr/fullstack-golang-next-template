#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SPEC_PATH="$ROOT_DIR/backend/internal/swagger/openapi.yaml"
TMP_OUTPUT="$(mktemp)"
trap 'rm -f "$TMP_OUTPUT"' EXIT

bunx @redocly/cli lint "$SPEC_PATH" 2>&1 | tee "$TMP_OUTPUT"

if rg -i "\bwarning\b" "$TMP_OUTPUT" >/dev/null; then
  echo "OpenAPI lint reported warnings; failing contract validation."
  exit 1
fi
