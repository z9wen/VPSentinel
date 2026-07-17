# VPSentinel

[![CI](https://github.com/z9wen/VPSentinel/actions/workflows/ci.yml/badge.svg)](https://github.com/z9wen/VPSentinel/actions/workflows/ci.yml)

VPSentinel is a lightweight, self-hosted monitoring system for a small fleet of
Linux VPS instances. The Go agent pushes versioned host metrics to a Gin server,
which stores them in PostgreSQL for later dashboarding and alerting.

## Start the server stack

```bash
./scripts/install-server.sh
```

The installation starts PostgreSQL and automatically applies versioned database
migrations before the API begins accepting traffic. See [server/README.md](server/README.md)
for configuration and security notes.
