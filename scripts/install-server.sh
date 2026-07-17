#!/bin/sh
set -eu

project_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
environment_file="$project_dir/.env"

if ! command -v docker >/dev/null 2>&1; then
    echo "Docker is required. Install Docker Engine and rerun this script." >&2
    exit 1
fi
if ! docker compose version >/dev/null 2>&1; then
    echo "Docker Compose v2 is required." >&2
    exit 1
fi
if ! command -v openssl >/dev/null 2>&1; then
    echo "OpenSSL is required to generate installation secrets." >&2
    exit 1
fi

if [ ! -f "$environment_file" ]; then
    postgres_password=$(openssl rand -hex 24)
    agent_token=$(openssl rand -hex 32)
    umask 077
    {
        echo "POSTGRES_PASSWORD=$postgres_password"
        echo "VPSENTINEL_AGENT_TOKEN=$agent_token"
        echo "VPSENTINEL_NODE_ID=primary-monitor"
    } >"$environment_file"
    echo "Generated $environment_file with mode 0600."
fi
chmod 600 "$environment_file"

cd "$project_dir"
docker compose up --build --detach --wait --wait-timeout 120
docker compose ps

echo "VPSentinel server is listening on http://127.0.0.1:8080."
echo "Put an HTTPS reverse proxy in front of this address before connecting public VPS agents."
