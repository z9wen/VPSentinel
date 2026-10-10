#!/bin/sh
# Installs or upgrades the VPSentinel agent as a hardened systemd service.
# Every file comes from the same GitHub release and is checked against its
# checksums.txt, so no repository checkout is needed on the host:
#
#   curl -fsSLO https://github.com/z9wen/VPSentinel/releases/download/<tag>/install-agent.sh
#   sudo VPSENTINEL_VERSION=<tag> \
#     VPSENTINEL_SERVER_URL=https://monitor.example.com/api/v1/agent/metrics \
#     VPSENTINEL_TOKEN=<one-time token> sh install-agent.sh
#
# Optional: VPSENTINEL_INTERVAL (default 15s), VPSENTINEL_DISK_PATH (default /),
# VPSENTINEL_ALLOW_INSECURE_HTTP=true for plain HTTP inside a private network.
set -eu

if [ "$(id -u)" -ne 0 ]; then
    echo "Run this installer as root." >&2
    exit 1
fi
: "${VPSENTINEL_VERSION:?Set VPSENTINEL_VERSION to a release tag such as v2026.07.25.123456}"
: "${VPSENTINEL_SERVER_URL:?Set VPSENTINEL_SERVER_URL to the metrics endpoint}"
: "${VPSENTINEL_TOKEN:?Set VPSENTINEL_TOKEN to the node token shown by the console}"

case "$(uname -m)" in
    x86_64) architecture=amd64 ;;
    aarch64|arm64) architecture=arm64 ;;
    *) echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

temporary_dir=$(mktemp -d)
trap 'rm -rf "$temporary_dir"' EXIT HUP INT TERM
release_url="https://github.com/z9wen/VPSentinel/releases/download/${VPSENTINEL_VERSION}"
binary="vpsentinel-agent-linux-${architecture}"
unit="vpsentinel-agent.service"

download() {
    curl --fail --location --silent --show-error --retry 3 "$release_url/$1" --output "$temporary_dir/$1"
}

download checksums.txt
for file in "$binary" "$unit"; do
    download "$file"
    if ! grep "  ${file}\$" "$temporary_dir/checksums.txt" >"$temporary_dir/$file.sha256"; then
        echo "$file is not listed in checksums.txt for $VPSENTINEL_VERSION." >&2
        exit 1
    fi
    (cd "$temporary_dir" && sha256sum --check --quiet "$file.sha256")
done

if ! id vpsentinel >/dev/null 2>&1; then
    useradd --system --home-dir /nonexistent --shell /usr/sbin/nologin vpsentinel
fi
install -o root -g root -m 0755 "$temporary_dir/$binary" /usr/local/bin/vpsentinel-agent
install -o root -g root -m 0644 "$temporary_dir/$unit" /etc/systemd/system/vpsentinel-agent.service
install -d -o root -g vpsentinel -m 0750 /etc/vpsentinel
umask 027
{
    printf 'VPSENTINEL_SERVER_URL=%s\n' "$VPSENTINEL_SERVER_URL"
    printf 'VPSENTINEL_TOKEN=%s\n' "$VPSENTINEL_TOKEN"
    printf 'VPSENTINEL_INTERVAL=%s\n' "${VPSENTINEL_INTERVAL:-15s}"
    printf 'VPSENTINEL_DISK_PATH=%s\n' "${VPSENTINEL_DISK_PATH:-/}"
    printf 'VPSENTINEL_ALLOW_INSECURE_HTTP=%s\n' "${VPSENTINEL_ALLOW_INSECURE_HTTP:-false}"
} >/etc/vpsentinel/agent.env
chown root:vpsentinel /etc/vpsentinel/agent.env
chmod 0640 /etc/vpsentinel/agent.env

systemctl daemon-reload
systemctl enable vpsentinel-agent
# restart (not "enable --now") so an upgrade replaces a running old binary.
systemctl restart vpsentinel-agent
systemctl --no-pager --full status vpsentinel-agent
