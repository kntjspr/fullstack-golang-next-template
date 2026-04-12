#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="$ROOT_DIR/docker-compose.test.yml"

cleanup() {
  docker compose -f "$COMPOSE_FILE" down
}
trap cleanup EXIT

docker compose -f "$COMPOSE_FILE" up -d

wait_for_healthy() {
  local service="$1"
  local max_attempts=60
  local attempt=1

  while [ "$attempt" -le "$max_attempts" ]; do
    local container_id
    container_id="$(docker compose -f "$COMPOSE_FILE" ps -q "$service")"
    if [ -n "$container_id" ]; then
      local status
      status="$(docker inspect --format='{{if .State.Health}}{{.State.Health.Status}}{{else}}unknown{{end}}' "$container_id")"
      if [ "$status" = "healthy" ]; then
        return 0
      fi
    fi

    sleep 1
    attempt=$((attempt + 1))
  done

  echo "service '$service' did not become healthy" >&2
  return 1
}

wait_for_healthy postgres
wait_for_healthy redis

export TEST_DATABASE_URL="postgres://postgres:test@localhost:5433/testdb?sslmode=disable"
export TEST_REDIS_URL="redis://localhost:6380"

cd "$ROOT_DIR/backend"
go test -tags integration ./internal/integration/...
