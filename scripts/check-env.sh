#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if [ -f .env ]; then
  set -a
  # shellcheck disable=SC1091
  . ./.env
  set +a
fi

required_vars=(DATABASE_URL REDIS_URL JWT_SECRET STAGE_STATUS)
missing_vars=()

for var_name in "${required_vars[@]}"; do
  if [ -z "${!var_name:-}" ]; then
    missing_vars+=("$var_name")
  fi
done

if [ "${#missing_vars[@]}" -gt 0 ]; then
  echo "Missing required environment variables:" >&2
  for var_name in "${missing_vars[@]}"; do
    echo "- $var_name" >&2
  done
  exit 1
fi

if [ "${STAGE_STATUS}" != "dev" ] && [ "${STAGE_STATUS}" != "prod" ]; then
  echo "STAGE_STATUS must be one of: dev, prod" >&2
  exit 1
fi

echo "Environment check passed."
