# Deployment Split Hosts

Use this when frontend and backend are hosted on different servers and/or different origins.

This is not fully automated by current defaults. You must adapt Ansible and runtime config deliberately.

## Architecture Target

Example:

- Frontend: `https://app.example.com`
- Backend API: `https://api.example.com`

Frontend calls backend using absolute API origin from `NEXT_PUBLIC_API_URL`.

## Step 1: Decide Auth Model For Browser Traffic

Choose one and document it:

1. Bearer header first (simpler cross-origin behavior)
2. Cookie fallback across origins (requires strict cookie+CORS settings)

Current backend supports both simultaneously, with bearer header priority.

## Step 2: Configure DNS And TLS First

- Provision both origins before deployment.
- Enforce HTTPS on both origins.
- Validate certs for frontend and backend hosts independently.

## Step 3: Configure Frontend Build Env

`frontend/.env` (production values before build/deploy):

```env
NEXT_PUBLIC_API_URL="https://api.example.com"
NEXT_PUBLIC_SITE_URL="https://app.example.com"
```

Because frontend is static export (`output: "export"`), changing API origin later requires rebuild/redeploy.

## Step 4: Configure Backend For Cross-Origin Requests

Required outcomes:

- Backend must emit CORS headers for frontend origin.
- Credentialed requests must be supported if cookie mode is used.

Current caveat:

- CORS middleware exists (`backend/middleware/cors.go`) but is not mounted in `backend/main.go`.
- You must mount/configure CORS in code before relying on browser cross-origin requests.

## Step 5: Handle Cookie Security Nuances

Cross-origin cookie mode requires:

- `Secure=true`
- Usually `SameSite=None`
- Frontend requests with `credentials: "include"` (already present)

Current caveat:

- Cookie `Secure` flag logic checks `STAGE_STATUS=prod` in auth handlers.
- Main config and auth cookie security now follow the same `STAGE_STATUS` convention.

For production deployments that rely on secure cookies, set `STAGE_STATUS=prod`.

## Step 6: Split Ansible Execution

Current default playbook applies all roles to one group. For split hosts, use either:

1. Separate playbooks for frontend/backend host groups.
2. Separate inventories plus role tags.

Tag-based example:

```bash
# frontend host
ansible-playbook -i hosts.frontend.ini playbook.yml --tags docker,frontend,nginx

# backend host
ansible-playbook -i hosts.backend.ini playbook.yml --tags docker,backend,postgres,redis
```

Warning:

- Current Nginx templates proxy to `cgapp-backend` container name for selected routes.
- On frontend-only host this target is unreachable unless backend is local to same Docker network.
- Remove or re-point these proxy locations for split-host topology.

## Step 7: Validate End-To-End Cross-Origin Behavior

From browser and CLI:

1. Frontend loads from frontend origin.
2. API calls go to backend origin.
3. Preflight and actual CORS responses include expected headers.
4. Login and authenticated endpoints work with chosen auth mode.
5. Logout clears auth cookie when cookie mode is enabled.

## Suggested Hardening Checklist

- [ ] Explicit CORS allowlist config in backend.
- [ ] Nginx templates split by topology.
- [ ] Auth cookie policy audited for cross-origin production.
- [ ] Frontend build pipeline enforces correct `NEXT_PUBLIC_API_URL` per environment.
