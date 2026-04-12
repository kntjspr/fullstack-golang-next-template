#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if ! command -v gosec >/dev/null 2>&1; then
  echo "gosec not found. Installing gosec..."
  go install github.com/securego/gosec/v2/cmd/gosec@latest
  export PATH="$(go env GOPATH)/bin:$PATH"
fi

echo "Running gosec on backend..."
gosec -exclude-dir=backend/.gopath -exclude-dir=backend/.gomodcache -exclude-dir=backend/.gocache ./backend/...

echo "Running bun audit in frontend..."
(
  cd frontend
  bun audit
)
