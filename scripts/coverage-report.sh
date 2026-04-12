#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="$ROOT_DIR/docker-compose.test.yml"
REPORT_DIR="$ROOT_DIR/reports"
REPORT_FILE="$REPORT_DIR/coverage-summary.txt"

BACKEND_THRESHOLD=70
FRONTEND_THRESHOLD=70

started_compose=0

is_running() {
  local service="$1"
  local container_id
  container_id="$(docker compose -f "$COMPOSE_FILE" ps -q "$service")"

  if [ -z "$container_id" ]; then
    return 1
  fi

  local state
  state="$(docker inspect --format='{{.State.Status}}' "$container_id")"
  [ "$state" = "running" ]
}

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

cleanup() {
  if [ "$started_compose" -eq 1 ]; then
    docker compose -f "$COMPOSE_FILE" down >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

if ! is_running postgres || ! is_running redis; then
  docker compose -f "$COMPOSE_FILE" up -d >/dev/null
  started_compose=1
fi

wait_for_healthy postgres
wait_for_healthy redis

export TEST_DATABASE_URL="postgres://postgres:test@localhost:5433/testdb?sslmode=disable"
export TEST_REDIS_URL="redis://localhost:6380"

pushd "$ROOT_DIR/backend" >/dev/null
backend_output="$(mktemp)"
GOFLAGS="" go test ./... -coverprofile=coverage.out -covermode=atomic | tee "$backend_output"
backend_coverage="$(
  awk '
    /^ok[[:space:]]+github.com\/create-go-app\/chi-go-template/ && /coverage:/ {
      for (i = 1; i <= NF; i++) {
        if ($i == "coverage:") {
          value = $(i+1)
          gsub(/%/, "", value)
          if (value ~ /^[0-9.]+$/) {
            sum += value
            count++
          }
        }
      }
    }
    END {
      if (count == 0) {
        print "0.00"
      } else {
        printf "%.2f", sum / count
      }
    }
  ' "$backend_output"
)"
rm -f "$backend_output"
popd >/dev/null

pushd "$ROOT_DIR/frontend" >/dev/null
frontend_output="$(mktemp)"
bun test --coverage 2>&1 | tee "$frontend_output"
frontend_coverage="$(
  sed -E 's/\x1B\[[0-9;]*[mK]//g' "$frontend_output" | awk -F'|' '
    /All files/ {
      gsub(/ /, "", $3)
      print $3
      exit
    }
  '
)"
rm -f "$frontend_output"
popd >/dev/null

if [ -z "$backend_coverage" ]; then
  backend_coverage="0.00"
fi
if [ -z "$frontend_coverage" ]; then
  frontend_coverage="0.00"
fi

backend_status="PASS"
frontend_status="PASS"

if ! awk -v c="$backend_coverage" -v t="$BACKEND_THRESHOLD" 'BEGIN { exit !(c+0 >= t+0) }'; then
  backend_status="FAIL"
fi

if ! awk -v c="$frontend_coverage" -v t="$FRONTEND_THRESHOLD" 'BEGIN { exit !(c+0 >= t+0) }'; then
  frontend_status="FAIL"
fi

mkdir -p "$REPORT_DIR"

{
  printf "Component | Covered | Threshold | Status\n"
  printf "backend   | %s%% | %s%%       | %s\n" "$backend_coverage" "$BACKEND_THRESHOLD" "$backend_status"
  printf "frontend  | %s%% | %s%%       | %s\n" "$frontend_coverage" "$FRONTEND_THRESHOLD" "$frontend_status"
} | tee "$REPORT_FILE"

if [ "$backend_status" = "FAIL" ] || [ "$frontend_status" = "FAIL" ]; then
  exit 1
fi
