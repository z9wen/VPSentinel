# VPSentinel

[![CI](https://github.com/z9wen/VPSentinel/actions/workflows/ci.yml/badge.svg)](https://github.com/z9wen/VPSentinel/actions/workflows/ci.yml)

VPSentinel is a lightweight, self-hosted monitoring and incident platform for
small Linux VPS fleets. A low-overhead Go agent pushes versioned host metrics
over HTTPS to a Gin API, PostgreSQL retains operational history, Grafana
evaluates alerts, and a React control plane manages nodes and incidents.

## What this project demonstrates

- Push-based Go agent with partial collection, bounded queueing, idempotent
  batches, exponential backoff, full jitter, and graceful shutdown.
- Per-node hashed tokens with generation, listing, rotation, and revocation.
- Admin dual-token authentication: short-lived Access JWT plus rotating Refresh
  JWT in an HttpOnly cookie, with server-side hash storage and logout revocation.
- Versioned PostgreSQL migrations, indexed time-series queries, advisory-lock
  migration safety, and automatic 30-day retention cleanup.
- Provisioned Grafana fleet/node dashboards, resource rules, offline detection,
  and firing/resolved webhooks stored as incident history.
- Tailwind/DaisyUI React operations UI with fleet state, latest metrics, token
  operations, incident history, responsive states, and Grafana deep links.
- Reproducible Docker Compose deployment, stable/LTS base images, Nginx TLS
  example, hardened systemd agent, CI checks, multi-architecture binaries, and
  provenance-attested container images published to GHCR.

## Architecture

```text
Linux VPS agents -- HTTPS + unique token --> Nginx --> Go API --> PostgreSQL
                                                        ^           ^
                                                        |           |
React control plane -- Access JWT ----------------------+       Grafana
                                                                    |
                                      firing/resolved webhook -------+
```

The public host exposes only Nginx. Agents initiate outbound connections, while
PostgreSQL, Grafana, and the API remain on an internal Compose network. See the
[architecture and trust boundaries](docs/architecture.md).

## Quick start

Prerequisites: Docker Engine, Docker Compose v2, and OpenSSL.

```bash
./scripts/install-server.sh
```

The installer generates `.env` with mode `0600`, builds the stack, applies
migrations, and waits for health checks. Open `http://127.0.0.1:8080` locally.
Before public use, configure the supplied [Nginx TLS proxy](deploy/nginx/vpsentinel.conf),
set the public web origin, and enable secure cookies.

Useful documentation:

- [Server deployment](docs/deployment.md)
- [Agent installation and token rotation](docs/agent-installation.md)
- [API and dual-token model](docs/api.md)
- [Node-offline runbook](docs/runbooks/node-offline.md)
- [Failure-drill postmortem template](docs/postmortems/template.md)

## Development checks

```bash
make test
make lint
make build
make compose-config
```

CI also starts PostgreSQL for migration and token-lifecycle integration tests,
validates provisioning assets, builds both application images, and publishes
Linux amd64/arm64 binaries with checksums after successful main-branch builds.
Each GitHub release also publishes multi-platform images with release, commit,
and `latest` tags:

```bash
docker pull ghcr.io/z9wen/vpsentinel-server:latest
docker pull ghcr.io/z9wen/vpsentinel-web:latest
```

## Operational evidence still required

The software MVP is implemented, but repository code cannot prove a real
deployment. Before using the project as a headline resume example, operate at
least two real nodes for 24 hours, execute the failure drills in
`PROJECT_PLAN.md`, replace the postmortem template with measured timestamps,
and add screenshots with credentials and host identifiers removed.

## Deliberate v0.1 limits

This release is single-admin and monitoring-only. It does not provide remote
shell execution, multi-tenancy, maintenance windows, disk-backed agent spooling,
or infrastructure provisioning. Those remain post-MVP work so the core
monitoring and recovery path stays understandable and auditable.
