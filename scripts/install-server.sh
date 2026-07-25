#!/bin/sh
set -eu

project_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
environment_file="$project_dir/.env"

for command_name in docker openssl; do
    if ! command -v "$command_name" >/dev/null 2>&1; then
        echo "$command_name is required." >&2
        exit 1
    fi
done
if ! docker compose version >/dev/null 2>&1; then
    echo "Docker Compose v2 is required." >&2
    exit 1
fi

ensure_value() {
    key=$1
    value=$2
    if ! grep -q "^${key}=" "$environment_file" 2>/dev/null; then
        printf '%s=%s\n' "$key" "$value" >>"$environment_file"
    fi
}

umask 077
touch "$environment_file"
ensure_value POSTGRES_PASSWORD "$(openssl rand -hex 24)"
ensure_value VPSENTINEL_AGENT_TOKEN "$(openssl rand -hex 32)"
ensure_value VPSENTINEL_NODE_ID primary-monitor
ensure_value VPSENTINEL_ADMIN_EMAIL admin@example.com
ensure_value VPSENTINEL_ADMIN_PASSWORD "$(openssl rand -base64 24 | tr -d '\n')"
ensure_value VPSENTINEL_JWT_SECRET "$(openssl rand -hex 32)"
ensure_value VPSENTINEL_GRAFANA_WEBHOOK_SECRET "$(openssl rand -hex 24)"
ensure_value GRAFANA_ADMIN_USER admin
ensure_value GRAFANA_ADMIN_PASSWORD "$(openssl rand -base64 24 | tr -d '\n')"
ensure_value VPSENTINEL_HTTP_PORT 8080
ensure_value VPSENTINEL_COOKIE_SECURE false
ensure_value VPSENTINEL_RETENTION_DAYS 30
chmod 600 "$environment_file"

cd "$project_dir"
docker compose up --build --detach --wait --wait-timeout 180
docker compose ps

echo "VPSentinel is listening on http://127.0.0.1:8080."
echo "Admin credentials and bootstrap agent token are stored in $environment_file (mode 0600)."
echo "Configure the host Nginx TLS proxy before connecting internet-facing agents."
