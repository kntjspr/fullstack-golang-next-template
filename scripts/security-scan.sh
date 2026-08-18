#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

TOOL_BIN_DIR="$(go env GOPATH)/bin"
GOSEC_BIN="$TOOL_BIN_DIR/gosec"
GOVULNCHECK_BIN="$TOOL_BIN_DIR/govulncheck"

if [ ! -x "$GOSEC_BIN" ]; then
  echo "gosec not found. Installing gosec..."
  go install github.com/securego/gosec/v2/cmd/gosec@latest
fi

echo "Running gosec on backend..."
"$GOSEC_BIN" -exclude-dir=backend/.gopath -exclude-dir=backend/.gomodcache -exclude-dir=backend/.gocache ./backend/...

if [ ! -x "$GOVULNCHECK_BIN" ]; then
  echo "govulncheck not found. Installing govulncheck..."
  go install golang.org/x/vuln/cmd/govulncheck@latest
fi

echo "Running govulncheck on backend..."
"$GOVULNCHECK_BIN" -C backend ./...
