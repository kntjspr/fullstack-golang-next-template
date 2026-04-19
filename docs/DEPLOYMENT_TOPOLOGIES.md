# Deployment Topologies And Onboarding

This document explains what changes when frontend and backend are hosted on different servers.

## Current Reality In This Repository

The default deployment path is **single-host** and **single inventory group**:

- `playbook.yml` applies `docker`, `frontend`, `backend`, `postgres`, `redis`, and `nginx` roles to the same `cgapp_project` host group.
- Nginx templates assume local Docker-network access to container `cgapp-backend`.
- Frontend is built as static export (`next.config.ts` has `output: "export"`).
- Frontend API base URL is compile-time from `NEXT_PUBLIC_API_URL`.

If you deploy frontend and backend to separate hosts, treat it as a different topology with extra setup.

## Topology Comparison

### A) Same host (supported by current automation)

- `https://app.example.com` serves static frontend from Nginx container.
- Nginx can proxy backend routes to `cgapp-backend` container on the same Docker network.
- One Ansible run can provision all components.

### B) Split hosts (requires manual adaptation)

- Frontend host serves only static assets.
- Backend host serves API separately (directly or through its own reverse proxy).
- Frontend calls backend via absolute URL from `NEXT_PUBLIC_API_URL` (for example `https://api.example.com`).
- CORS and auth cookie policy become mandatory design decisions, not defaults.

## What Changes For Docker

- Single-host mode uses one Docker network where Nginx and backend are mutually reachable by container name.
- Split-host mode has separate Docker networks per host; container-name proxying across hosts does not work.
- Any Nginx `proxy_pass http://cgapp-backend:...` on a frontend-only host will fail unless backend is local.

## What Changes For Ansible

Current playbook is not split-aware by default. You need one of these patterns:

1. Create separate playbooks per role group.
2. Keep one playbook but run targeted tags against separate inventories.

Example tag-based split execution:

```bash
# frontend host
ansible-playbook -i hosts.frontend.ini playbook.yml --tags docker,frontend,nginx

# backend host
ansible-playbook -i hosts.backend.ini playbook.yml --tags docker,backend,postgres,redis
```

Important:

- Nginx templates in `roles/nginx/templates/` currently include backend proxy rules for selected routes.
- For frontend-only host, either remove API proxy locations or point them to a reachable backend origin.
- If backend is behind its own Nginx, manage that in backend-host config separately.

## Onboarding Delta Vs README Quick Start

The local development onboarding in `README.md` stays the same.

The deployment onboarding changes as follows:

1. Decide topology before provisioning.
2. Create DNS for both origins (for example `app.example.com` and `api.example.com`).
3. Configure `frontend/.env`:
   - `NEXT_PUBLIC_API_URL=https://api.example.com`
   - `NEXT_PUBLIC_SITE_URL=https://app.example.com`
4. Build frontend with those production env values.
5. Configure backend public URL/proxy and TLS on backend host.
6. Add/enable backend CORS allowlist for `https://app.example.com`.
7. Choose auth mode for browser traffic:
   - bearer-token header mode (simpler across origins)
   - cookie mode (requires correct cookie flags and CORS credentials support)
8. Run Ansible in split mode (separate inventories/tags or separate playbooks).
9. Validate from browser with real cross-origin requests.

## Auth And Cookie Nuances For Split-Host

When frontend and backend are on different origins:

- Cookies are cross-site from the browser perspective.
- Cross-site cookies require `Secure` and generally `SameSite=None`.
- `credentials: "include"` must be used by frontend (already present in `api-client.ts`).
- Backend must return matching CORS headers for credentialed requests.

If you do not configure cross-site cookie policy correctly, login may appear to succeed but authenticated requests will fail because cookie is not sent.

## Current Known Assumptions To Document And Fix

These are important constraints in the current codebase:

- Backend CORS middleware exists (`backend/middleware/cors.go`) but is not mounted in `backend/main.go`.
- Cookie `Secure` flag in auth handlers depends on `STAGE_STATUS=prod`.
- Nginx templates proxy only a subset of backend routes (`/hc`, `/healthz`, `/swagger`, `/openapi.yaml`), not full API surface.
- Frontend API base URL is baked at build time; changing API host requires rebuild/redeploy of frontend static bundle.

Document these in PRs/releases until code-level fixes are merged.
