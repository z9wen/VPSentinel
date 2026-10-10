# Server deployment

## Prerequisites

- A Linux host (amd64 or arm64) with Docker Engine and Docker Compose v2.
- A DNS record such as `monitor.example.com` pointing to the host.
- Nginx stable and a trusted TLS certificate on the host.

## Install

Production hosts run the tested release images instead of building from source.
Pick a tag from the GitHub releases page:

```bash
git clone https://github.com/z9wen/VPSentinel.git
cd VPSentinel
git checkout vYYYY.MM.DD.TIMESTAMP
VPSENTINEL_VERSION=vYYYY.MM.DD.TIMESTAMP \
  VPSENTINEL_WEB_ORIGIN=https://monitor.example.com \
  VPSENTINEL_COOKIE_SECURE=true \
  ./scripts/install-server.sh
```

Without `VPSENTINEL_VERSION` the installer builds images from the checkout,
which suits development. The installer writes generated credentials to `.env`
with mode `0600`; back it up like any other production secret. It also records
the release and Compose files there, so plain `docker compose ps` and
`docker compose logs` target the same deployment afterwards.

Startup order is enforced by Compose: PostgreSQL becomes healthy, the one-shot
`migrate` job applies schema migrations under an advisory lock and exits, and
only then does the API start. A failed migration stops the rollout instead of
crash-looping the API.

Copy `deploy/nginx/vpsentinel.conf` to the host Nginx configuration, replace the
example domain and certificate paths, then validate and reload:

```bash
sudo nginx -t
sudo systemctl reload nginx
```

The stack publishes only `127.0.0.1:8080`. Confirm that ports 5432, 3000, and
the Go API are not externally reachable, and restrict inbound firewall rules to
SSH, HTTP, and HTTPS.

### Client addresses

Rate limits and request logs use the real client address. The host Nginx sets
`X-Real-IP`, the bundled web proxy trusts it only from loopback and private
peers, and the API reads it only from the private Compose network
(`VPSENTINEL_TRUSTED_PROXIES`). Keep the web port bound to loopback or a private
address; exposing it publicly would let clients forge `X-Real-IP`. Logins are
limited to 10 attempts per minute per client address.

## Reference deployment: two VPS

A small real fleet is enough to produce operational evidence:

| Host | Role | Inbound ports |
| --- | --- | --- |
| Oracle Cloud VM (arm64) | Control plane + agent | 22, 80, 443 |
| Second VPS | Agent only | 22 (none needed by VPSentinel) |

1. Install the control plane on the Oracle VM as above, then the host Nginx and
   a certificate for the monitor domain.
2. Oracle Cloud filters traffic twice: the VCN security list (or network
   security group) and a host firewall in the image. Open 80 and 443 in both.
   On Ubuntu images, insert ACCEPT rules above the existing REJECT rule shown by
   `sudo iptables -L INPUT --line-numbers` and persist them with
   `sudo netfilter-persistent save`; on Oracle Linux use
   `sudo firewall-cmd --permanent --add-service=http --add-service=https` and
   `sudo firewall-cmd --reload`.
3. In the console, add one node per host and install each agent with the
   release installer ([agent installation](agent-installation.md)). The agent
   only makes outbound HTTPS requests, so the second VPS needs no new inbound
   rule.
4. Add an external uptime check against `https://monitor.example.com/readyz`.
   Grafana runs on the monitored control plane, so it cannot report its own
   outage; the external check covers that blind spot.

## Upgrade and rollback

```bash
git fetch --tags
git checkout vNEW
VPSENTINEL_VERSION=vNEW ./scripts/install-server.sh
docker compose ps
```

The installer pulls the new images, runs the migrate job, and recreates only
changed containers. Migrations are forward-only: before an upgrade, take a
PostgreSQL backup and note the running tag. To roll back, rerun the installer
with the previous tag; restore the backup only when a migration is incompatible
with the old binary.

## Published container images

Successful main-branch releases publish Linux amd64/arm64 images to GitHub
Container Registry:

```bash
docker pull ghcr.io/z9wen/vpsentinel-server:vYYYY.MM.DD.TIMESTAMP
docker pull ghcr.io/z9wen/vpsentinel-web:vYYYY.MM.DD.TIMESTAMP
```

Pin the dated release tag rather than `latest`. Every image also receives an
immutable `sha-<commit>` tag, OCI source/revision labels, an SBOM, and build
provenance.

## Backup

```bash
docker compose exec -T postgres pg_dump -U vpsentinel -Fc vpsentinel > vpsentinel.dump
```

Store dumps encrypted outside the monitoring host and periodically test a
restore into a disposable PostgreSQL instance.
