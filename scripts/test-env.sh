#!/usr/bin/env bash

# Load test-service ports from the local .env file without evaluating its contents.
read_env_value() {
  local key="$1"
  local env_file="$2"

  [ -f "$env_file" ] || return 0
  awk -F= -v key="$key" '$1 == key { sub(/^[^=]*=/, ""); value = $0 } END { print value }' "$env_file"
}

load_test_environment() {
  local env_file="$1"

  TEST_POSTGRES_PORT="${TEST_POSTGRES_PORT:-$(read_env_value TEST_POSTGRES_PORT "$env_file")}"
  TEST_REDIS_PORT="${TEST_REDIS_PORT:-$(read_env_value TEST_REDIS_PORT "$env_file")}"
  TEST_POSTGRES_PORT="${TEST_POSTGRES_PORT:-5433}"
  TEST_REDIS_PORT="${TEST_REDIS_PORT:-6380}"

  TEST_DATABASE_URL="${TEST_DATABASE_URL:-postgres://postgres:test@localhost:${TEST_POSTGRES_PORT}/testdb?sslmode=disable}"
  TEST_REDIS_URL="${TEST_REDIS_URL:-redis://localhost:${TEST_REDIS_PORT}}"

  export TEST_POSTGRES_PORT TEST_REDIS_PORT TEST_DATABASE_URL TEST_REDIS_URL
}
