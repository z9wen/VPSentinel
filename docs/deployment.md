# Server deployment

## Prerequisites

- A Linux host with Docker Engine and Docker Compose v2.
- A DNS record such as `monitor.example.com` pointing to the host.
- Nginx stable and a trusted TLS certificate on the host.

## Install

```bash
git clone https://github.com/z9wen/VPSentinel.git
cd VPSentinel
./scripts/install-server.sh
```

The installer writes generated credentials to `.env` with mode `0600`. Back up
this file using the same controls as other production secrets. Change
`VPSENTINEL_ADMIN_EMAIL`, set the public `VPSENTINEL_WEB_ORIGIN`, and set
`VPSENTINEL_COOKIE_SECURE=true` before public use.

Copy `deploy/nginx/vpsentinel.conf` to the host Nginx configuration, replace the
example domain and certificate paths, then validate and reload:

```bash
sudo nginx -t
sudo systemctl reload nginx
```

The Compose network publishes only `127.0.0.1:8080`. Confirm that ports 5432,
3000, and the Go API are not externally reachable. Restrict inbound firewall
rules to SSH, HTTP, and HTTPS.

## Upgrade and rollback

```bash
git pull --ff-only
docker compose build
docker compose up -d --wait
docker compose ps
```

Database migrations are forward-only. Before an upgrade, take a PostgreSQL
backup and record the running Git commit. To roll back application containers,
checkout the recorded commit and rebuild; restore the database backup only when
a migration is incompatible with the old binary.

## Published container images

Successful main-branch releases publish Linux amd64/arm64 images to GitHub
Container Registry:

```bash
docker pull ghcr.io/z9wen/vpsentinel-server:latest
docker pull ghcr.io/z9wen/vpsentinel-web:latest
```

Production deployments should pin the dated release tag shown on the GitHub
release instead of relying on `latest`. Every image also receives an immutable
`sha-<commit>` tag, OCI source/revision labels, an SBOM, and build provenance.

## Backup

```bash
docker compose exec -T postgres pg_dump -U vpsentinel -Fc vpsentinel > vpsentinel.dump
```

Store dumps encrypted outside the monitoring host and periodically test a
restore into a disposable PostgreSQL instance.
