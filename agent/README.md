# VPSentinel Agent

The agent collects host metrics and sends versioned JSON batches to the VPSentinel server.

## Run a local dry-run

```bash
cd agent
go run ./cmd/vpsentinel-agent --dry-run
```

## Send metrics

```bash
export VPSENTINEL_SERVER_URL=https://monitor.example.com/api/v1/agent/metrics
export VPSENTINEL_TOKEN=replace-with-a-node-token
go run ./cmd/vpsentinel-agent
```

Supported environment variables:

- `VPSENTINEL_SERVER_URL`
- `VPSENTINEL_TOKEN`
- `VPSENTINEL_DISK_PATH` (default `/`)
- `VPSENTINEL_INTERVAL` (default `15s`)
- `VPSENTINEL_REQUEST_TIMEOUT` (default `10s`)
- `VPSENTINEL_QUEUE_CAPACITY` (default `120`)
- `VPSENTINEL_DRY_RUN` (default `false`)
- `VPSENTINEL_ALLOW_INSECURE_HTTP` (default `false`; local development only)

Command-line flags with the equivalent names override environment values.

Plain HTTP is rejected by default because it exposes the bearer token. For a
local development server only, set `VPSENTINEL_ALLOW_INSECURE_HTTP=true`.

The collector and sender run independently. Temporary server failures fill a
bounded in-memory queue while host collection continues; successful responses
acknowledge samples by their stable `sample_id`.
