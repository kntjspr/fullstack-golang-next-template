# ADR 0008: Enforce AI Governance with Independent Checks

- Status: Accepted

## Context

Instructions alone cannot stop an implementation from weakening tests or introducing unsafe dependencies to satisfy an immediate task.

## Decision

Use local and CI enforcement for AI-authored changes:

- Pre-commit runs `go vet ./backend/...` and checks `gofmt -l .`.
- `make test` runs the race detector and coverage from day one.
- `make security` runs gosec and govulncheck.
- `make mutation-test` uses the maintained `avito-tech/go-mutesting` fork to mutate `backend/core/`; surviving mutations fail the command. The original `zimmski/go-mutesting` release is incompatible with this Go 1.25 project.
- Pull requests reject deletion or modification of existing lines in `backend/service_test.go`, while allowing additions.
- CI runs these checks from a clean checkout and applies the existing 70% coverage floor.

## Consequences

- Agent-reported local success is not merge authority.
- The contract suite remains extensible but cannot be silently weakened in a pull request.
- Mutation testing adds execution time, but confirms that core branches are asserted rather than merely executed.
