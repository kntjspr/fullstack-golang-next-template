# Deployment Single Host

Use this for the default supported path in the current repository.

## Preconditions

- One Linux host (Ubuntu 22.04+ recommended)
- SSH access with sudo-capable user
- DNS for your domain
- Ports `80` and `443` open

## Step 1: Prepare Inventory

Edit `hosts.ini`:

```ini
[cgapp_project]
<your-server-ip-or-hostname>

[cgapp_project:vars]
project_domain=example.com
backend_port=5000
```

Also review:

- `server_dir`
- `server_user`
- `server_group`
- Postgres/Redis vars
- HTTPS redirect vars (`nginx_use_only_https`, `nginx_redirect_to_non_www`)

## Step 2: Configure Production Environments

Backend (`backend/.env`):

- `STAGE_STATUS=prod`
- DB/Redis values matching inventory/container setup
- Strong `JWT_SECRET`
- Sentry values if used

Frontend (`frontend/.env`):

- `NEXT_PUBLIC_API_URL=https://<your-domain>` or equivalent API origin
- `NEXT_PUBLIC_SITE_URL=https://<your-domain>`

Important: frontend is static export; `NEXT_PUBLIC_API_URL` is compiled into build output.

## Step 3: Run Local Pre-Deploy Gate

```bash
make check
```

Do not deploy with failing checks.

## Step 4: Deploy

```bash
ansible-playbook -i hosts.ini playbook.yml
```

## Step 5: Post-Deploy Validation

Validate:

1. `https://<domain>/` serves frontend.
2. `https://<domain>/healthz` returns backend health.
3. Auth login/refresh/logout works from browser.
4. DB/Redis containers are healthy.

## Known Limitations In Current Nginx Templates

Current Nginx proxy rules include only:

- `/hc`
- `/healthz`
- `/swagger`
- `/openapi.yaml`

If you expect all API paths under the same domain/proxy, extend templates accordingly.
