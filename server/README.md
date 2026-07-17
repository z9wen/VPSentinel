# VPSentinel Server

The Gin server accepts authenticated metric batches and stores them in PostgreSQL.

## One-command installation

From the repository root:

```bash
./scripts/install-server.sh
```

The installer generates a mode `0600` `.env` file, starts PostgreSQL, builds the
server image, and starts the stack. On startup, the server waits for PostgreSQL,
acquires a PostgreSQL advisory lock, verifies applied migration checksums, and
applies pending embedded migrations before listening for HTTP requests.

The service binds to `127.0.0.1:8080`. Put Nginx or another HTTPS reverse proxy
in front of it before connecting agents over the internet.

## Endpoints

- `POST /api/v1/agent/metrics`
- `GET /healthz`
- `GET /readyz`
- `GET /version`

## Run the server natively

Start PostgreSQL separately, then configure:

```bash
export VPSENTINEL_AGENT_TOKEN=replace-with-at-least-16-characters
export VPSENTINEL_DATABASE_URL='postgres://vpsentinel:password@127.0.0.1:5432/vpsentinel?sslmode=disable'
go run ./cmd/vpsentinel-server
```

Configuration:

- `VPSENTINEL_LISTEN_ADDRESS` (default `:8080`)
- `VPSENTINEL_AGENT_TOKEN` (required; temporary single-node authenticator)
- `VPSENTINEL_NODE_ID` (default `development-node`)
- `VPSENTINEL_DATABASE_URL` (required)
- `VPSENTINEL_AUTO_MIGRATE` (default `true`)
- `VPSENTINEL_DATABASE_START_TIMEOUT` (default `30s`)
- `VPSENTINEL_MAX_BODY_BYTES` (default `1048576`)
- `VPSENTINEL_SHUTDOWN_TIMEOUT` (default `10s`)
- `GIN_MODE` (default `release`)

The static agent token is an initial development step. A PostgreSQL-backed
per-node token authenticator and token rotation API will replace it.
