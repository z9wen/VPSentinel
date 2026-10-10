# VPSentinel Server

The Gin API authenticates agents and administrators, stores metric batches in
PostgreSQL, receives Grafana incidents, and performs retention cleanup.

## Authentication

- Agent credentials are unique random tokens stored as SHA-256 hashes.
- Admin login uses bcrypt password hashes and issues a 15-minute Access JWT.
- The seven-day Refresh JWT is stored in an HttpOnly cookie and its hash is
  rotated transactionally in PostgreSQL on every refresh.
- Grafana webhooks use a separate constant-time-compared header secret.

See [the full endpoint and token flow](../docs/api.md).

## Native configuration

Required variables:

- `VPSENTINEL_DATABASE_URL`
- `VPSENTINEL_AGENT_TOKEN` and `VPSENTINEL_NODE_ID` for the bootstrap node
- `VPSENTINEL_ADMIN_EMAIL` and `VPSENTINEL_ADMIN_PASSWORD`
- `VPSENTINEL_JWT_SECRET` (at least 32 characters)
- `VPSENTINEL_GRAFANA_WEBHOOK_SECRET` (at least 16 characters)

Optional variables include `VPSENTINEL_ACCESS_TOKEN_TTL` (default `15m`),
`VPSENTINEL_REFRESH_TOKEN_TTL` (default `168h`), `VPSENTINEL_RETENTION_DAYS`
(default `30`), `VPSENTINEL_MAX_BODY_BYTES`, `VPSENTINEL_WEB_ORIGIN`,
`VPSENTINEL_COOKIE_SECURE`, `VPSENTINEL_AUTO_MIGRATE` (default `true`), and
`VPSENTINEL_TRUSTED_PROXIES`: comma-separated IPs or CIDRs of reverse proxies
whose `X-Real-IP` header names the client. It defaults to none, so a directly
exposed server ignores the header.

## Migrations

`vpsentinel-server migrate` applies pending migrations and exits; it needs only
`VPSENTINEL_DATABASE_URL`. Compose runs it as a one-shot job before the API and
sets `VPSENTINEL_AUTO_MIGRATE=false`, so the API never migrates on startup.

```bash
go test ./...
go run ./cmd/vpsentinel-server migrate
go run ./cmd/vpsentinel-server
```

Production deployment uses the root Compose stack rather than a native server.
