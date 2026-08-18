#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TOOL_BIN_DIR="$(go env GOPATH)/bin"
MUTESTING_BIN="$TOOL_BIN_DIR/go-mutesting"

if [ ! -x "$MUTESTING_BIN" ]; then
  echo "go-mutesting not found. Installing go-mutesting..."
  go install github.com/avito-tech/go-mutesting/cmd/go-mutesting@latest
fi

WORKTREE_DIR="$(mktemp -d)"

cleanup() {
  git -C "$ROOT_DIR" worktree remove --force "$WORKTREE_DIR" >/dev/null 2>&1 || true
  rmdir "$WORKTREE_DIR" >/dev/null 2>&1 || true
}
trap cleanup EXIT

# go-mutesting edits source files while evaluating mutants. Run it in an
# isolated worktree and apply the caller's staged and unstaged changes so a
# failed mutation run cannot leave the real checkout altered.
git -C "$ROOT_DIR" worktree add --detach "$WORKTREE_DIR" HEAD >/dev/null
git -C "$ROOT_DIR" diff --binary HEAD | git -C "$WORKTREE_DIR" apply --whitespace=nowarn

cd "$WORKTREE_DIR/backend"

echo "Mutating stable core logic..."
"$MUTESTING_BIN" ./core/...
