#!/bin/sh
# Installs or upgrades the VPSentinel control plane with Docker Compose.
#
#   ./scripts/install-server.sh                         build images from this checkout
#   VPSENTINEL_VERSION=vX ./scripts/install-server.sh   run published release images
#
# Secrets are generated once into .env (mode 0600) and kept across upgrades.
# VPSENTINEL_BIND_ADDRESS, VPSENTINEL_HTTP_PORT, VPSENTINEL_WEB_ORIGIN, and
# VPSENTINEL_COOKIE_SECURE given in the environment are written to .env too.
set -eu

project_dir=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
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

set_value() {
    key=$1
    value=$2
    if grep -q "^${key}=" "$environment_file" 2>/dev/null; then
        temporary_file=$(mktemp "$environment_file.XXXXXX")
        grep -v "^${key}=" "$environment_file" >"$temporary_file" || true
        mv "$temporary_file" "$environment_file"
    fi
    printf '%s=%s\n' "$key" "$value" >>"$environment_file"
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

for key in VPSENTINEL_BIND_ADDRESS VPSENTINEL_HTTP_PORT VPSENTINEL_WEB_ORIGIN VPSENTINEL_COOKIE_SECURE; do
    eval "value=\${$key:-}"
    if [ -n "$value" ]; then
        set_value "$key" "$value"
    fi
done

if [ -n "${VPSENTINEL_VERSION:-}" ]; then
    # Recording the release and compose files in .env makes later plain
    # `docker compose ps|logs|up` commands target the same deployment.
    set_value VPSENTINEL_VERSION "$VPSENTINEL_VERSION"
    set_value COMPOSE_FILE docker-compose.yml:deploy/compose/release.yml
fi
chmod 600 "$environment_file"

cd "$project_dir"
if grep -q '^COMPOSE_FILE=.*release.yml' "$environment_file"; then
    docker compose pull --quiet migrate server web
    docker compose up --detach --wait --wait-timeout 300 --no-build --remove-orphans
else
    docker compose up --detach --wait --wait-timeout 300 --build --remove-orphans
fi
docker compose ps

http_port=$(grep '^VPSENTINEL_HTTP_PORT=' "$environment_file" | cut -d= -f2)
bind_address=$(grep '^VPSENTINEL_BIND_ADDRESS=' "$environment_file" | cut -d= -f2)
echo "VPSentinel is listening on http://${bind_address:-127.0.0.1}:${http_port}."
echo "Admin credentials and the bootstrap agent token are stored in $environment_file (mode 0600)."
echo "Put a TLS reverse proxy in front before connecting internet-facing agents (deploy/nginx/vpsentinel.conf)."
