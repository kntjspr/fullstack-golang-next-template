#!/usr/bin/env bash
set -e

cp .githooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

echo "Git hooks installed. Secret scanning active on every commit."
