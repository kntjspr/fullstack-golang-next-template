# A default Makefile for fullstack-golang-next-template.
# Author: Kent Jasper Sisi <kntjspr@pm.me> (https://github.com/kntjspr)
# For more information, please visit https://github.com/kntjspr/fullstack-golang-next-template

.DEFAULT_GOAL := help

.PHONY: help bootstrap hooks dev test lint run build validate-contracts test-integration coverage ci-local check down

FRONTEND_PATH = $(PWD)/frontend
BACKEND_PATH = $(PWD)/backend
TEST_DATABASE_URL = postgres://postgres:test@localhost:5433/testdb?sslmode=disable
TEST_REDIS_URL = redis://localhost:6380

help:
	@echo "usage: make <target>"
	@echo "bootstrap     set up fresh dev environment"
	@echo "hooks         install git hooks"
	@echo "dev           start backend + frontend locally"
	@echo "test          run all tests"
	@echo "lint          run linters"
	@echo "build         build backend + frontend"
	@echo "coverage      run coverage report"
	@echo "validate-contracts  lint OpenAPI + run contract tests"
	@echo "ci-local      simulate full CI pipeline locally"
	@echo "check         run full pre-deploy regression gate"
	@echo "down          stop test infrastructure"

bootstrap:
	@bash scripts/bootstrap.sh

hooks:
	@bash scripts/install-hooks.sh

dev:
	@bash -ceu ' \
		set -euo pipefail; \
		(cd backend && go run .) & backend_pid=$$!; \
		(cd frontend && bun dev) & frontend_pid=$$!; \
		cleanup() { \
			kill "$$backend_pid" "$$frontend_pid" >/dev/null 2>&1 || true; \
		}; \
		trap cleanup INT TERM EXIT; \
		wait -n "$$backend_pid" "$$frontend_pid"; \
	'

test:
	@bash -ceu ' \
		set -euo pipefail; \
		export TEST_DATABASE_URL="$(TEST_DATABASE_URL)"; \
		export TEST_REDIS_URL="$(TEST_REDIS_URL)"; \
		go test ./backend/...; \
		cd frontend; \
		bun test; \
	'

lint:
	@bash -ceu ' \
		set -euo pipefail; \
		if ! command -v golangci-lint >/dev/null 2>&1; then \
			echo "golangci-lint is required. Install it and re-run make lint." >&2; \
			exit 1; \
		fi; \
		golangci-lint run ./backend/...; \
		cd frontend; \
		bun run lint; \
	'

run: test
	@if [ -d "$(FRONTEND_PATH)" ]; then cd $(FRONTEND_PATH) && bun run dev; fi
	@if [ -d "$(BACKEND_PATH)" ]; then cd $(BACKEND_PATH) && $(MAKE) run; fi

build:
	@bash -ceu ' \
		set -euo pipefail; \
		go build ./backend/...; \
		cd frontend; \
		bun run build; \
	'

validate-contracts:
	@cd $(BACKEND_PATH) && ./scripts/validate-openapi.sh
	@cd $(BACKEND_PATH) && go test ./internal/contract -count=1

test-integration:
	@bash scripts/test-integration.sh

coverage:
	@bash scripts/coverage-report.sh

ci-local:
	@bash -ceu ' \
		set -euo pipefail; \
		COMPOSE_FILE="docker-compose.test.yml"; \
		started_compose=0; \
		is_running() { \
			service="$$1"; \
			container_id="$$(docker compose -f "$$COMPOSE_FILE" ps -q "$$service")"; \
			if [ -z "$$container_id" ]; then return 1; fi; \
			state="$$(docker inspect --format="{{.State.Status}}" "$$container_id")"; \
			[ "$$state" = "running" ]; \
		}; \
		wait_for_healthy() { \
			service="$$1"; \
			for attempt in $$(seq 1 60); do \
				container_id="$$(docker compose -f "$$COMPOSE_FILE" ps -q "$$service")"; \
				if [ -n "$$container_id" ]; then \
					status="$$(docker inspect --format="{{if .State.Health}}{{.State.Health.Status}}{{else}}unknown{{end}}" "$$container_id")"; \
					if [ "$$status" = "healthy" ]; then return 0; fi; \
				fi; \
				sleep 1; \
			done; \
			echo "service $$service did not become healthy" >&2; \
			return 1; \
		}; \
		if ! is_running postgres || ! is_running redis; then \
			docker compose -f "$$COMPOSE_FILE" up -d; \
			started_compose=1; \
		fi; \
		trap '\''if [ "$$started_compose" -eq 1 ]; then docker compose -f "$$COMPOSE_FILE" down; fi'\'' EXIT; \
		wait_for_healthy postgres; \
		wait_for_healthy redis; \
		export TEST_DATABASE_URL="$(TEST_DATABASE_URL)"; \
		export TEST_REDIS_URL="$(TEST_REDIS_URL)"; \
		cd backend; \
		go test ./... -race -coverprofile=backend-coverage.out -covermode=atomic; \
		cd ../frontend; \
		bun install; \
		bun test --coverage; \
		cd ..; \
		cd backend; \
		bash scripts/validate-openapi.sh; \
		cd ..; \
		go test ./backend/internal/contract/...; \
		cd frontend; \
		bun install; \
		bun test src/lib/__tests__/api-client.test.ts; \
		cd ..; \
		go test -tags integration ./backend/internal/integration/...; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
		"$$(go env GOPATH)/bin/gosec" -exclude-dir=backend/.gopath -exclude-dir=backend/.gomodcache -exclude-dir=backend/.gocache ./backend/...; \
		cd frontend; \
		bun install; \
		bun audit; \
		cd ..; \
		cd backend; \
		go run . >/tmp/ci-local-backend.log 2>&1 & \
		backend_pid=$$!; \
		cd ..; \
		trap '\''kill "$$backend_pid" >/dev/null 2>&1 || true; if [ "$$started_compose" -eq 1 ]; then docker compose -f "$$COMPOSE_FILE" down; fi'\'' EXIT; \
		for _ in $$(seq 1 60); do \
			if curl -fsS http://127.0.0.1:5000/healthz >/dev/null; then break; fi; \
			sleep 1; \
		done; \
		headers="$$(curl -sS -D - -o /dev/null http://127.0.0.1:5000/healthz)"; \
		echo "$$headers" | grep -qi "^X-Content-Type-Options:"; \
		echo "$$headers" | grep -qi "^X-Frame-Options:"; \
		echo "$$headers" | grep -qi "^X-XSS-Protection:"; \
		echo "$$headers" | grep -qi "^Content-Security-Policy:"; \
	'

check:
	@bash scripts/full-check.sh

down:
	@docker compose -f docker-compose.test.yml down
