# Deployment Overview

The repository currently provides deployment automation that is biased toward one server hosting everything.

## Topology Options

## 1) Single Host (Current Default)

One host runs:

- Frontend static assets (Nginx-served)
- Backend API container
- Postgres container
- Redis container
- Nginx container

Characteristics:

- Works with current `playbook.yml` + `hosts.ini` defaults.
- Nginx can proxy to backend by Docker container name.
- Simpler operational setup.

## 2) Split Hosts (Frontend And Backend Separate)

Frontend host and backend host are different servers and usually different origins.

Characteristics:

- Requires Ansible inventory/playbook split or strict tag targeting.
- Requires explicit cross-origin auth/CORS strategy.
- Requires frontend rebuild when API origin changes.

## Current Automation Assumptions

- `playbook.yml` targets one inventory group: `cgapp_project`.
- Nginx templates use `proxy_pass http://cgapp-backend:<port>` for selected routes.
- This proxy pattern only works when backend container is reachable on the same Docker network.

## Decision Guide

Use Single Host if:

- You want fastest path with existing automation.
- Frontend and backend can share one public domain/server.

Use Split Hosts if:

- You need independent scaling/lifecycle for frontend and backend.
- You require separate network/security boundaries.
- You accept additional CORS/cookie and Ansible complexity.

After choosing, continue with either:

- [Deployment Single Host](deployment-single-host.md)
- [Deployment Split Hosts](deployment-split-host.md)
