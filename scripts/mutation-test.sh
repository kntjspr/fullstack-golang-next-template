#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR/backend"

TOOL_BIN_DIR="$(go env GOPATH)/bin"
MUTESTING_BIN="$TOOL_BIN_DIR/go-mutesting"

if [ ! -x "$MUTESTING_BIN" ]; then
  echo "go-mutesting not found. Installing go-mutesting..."
  go install github.com/avito-tech/go-mutesting/cmd/go-mutesting@latest
fi

echo "Mutating stable core logic..."
"$MUTESTING_BIN" ./core/...
