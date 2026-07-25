#!/bin/sh
set -eu

if [ "$(id -u)" -ne 0 ]; then
    echo "Run this installer as root." >&2
    exit 1
fi
: "${VPSENTINEL_VERSION:?Set VPSENTINEL_VERSION to a release tag such as v2026.07.25.123456}"
: "${VPSENTINEL_SERVER_URL:?Set VPSENTINEL_SERVER_URL to the HTTPS metrics endpoint}"
: "${VPSENTINEL_TOKEN:?Set VPSENTINEL_TOKEN to the node token shown by the API}"

case "$(uname -m)" in
    x86_64) architecture=amd64 ;;
    aarch64|arm64) architecture=arm64 ;;
    *) echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
temporary_dir=$(mktemp -d)
trap 'rm -rf "$temporary_dir"' EXIT HUP INT TERM
asset="vpsentinel-agent-linux-${architecture}"
release_url="https://github.com/z9wen/VPSentinel/releases/download/${VPSENTINEL_VERSION}"

curl --fail --location --silent --show-error "$release_url/$asset" --output "$temporary_dir/$asset"
curl --fail --location --silent --show-error "$release_url/checksums.txt" --output "$temporary_dir/checksums.txt"
(cd "$temporary_dir" && grep "  ${asset}$" checksums.txt | sha256sum --check -)

if ! id vpsentinel >/dev/null 2>&1; then
    useradd --system --home-dir /nonexistent --shell /usr/sbin/nologin vpsentinel
fi
install -o root -g root -m 0755 "$temporary_dir/$asset" /usr/local/bin/vpsentinel-agent
install -d -o root -g vpsentinel -m 0750 /etc/vpsentinel
umask 027
{
    printf 'VPSENTINEL_SERVER_URL=%s\n' "$VPSENTINEL_SERVER_URL"
    printf 'VPSENTINEL_TOKEN=%s\n' "$VPSENTINEL_TOKEN"
    printf 'VPSENTINEL_INTERVAL=%s\n' "${VPSENTINEL_INTERVAL:-15s}"
    printf 'VPSENTINEL_DISK_PATH=%s\n' "${VPSENTINEL_DISK_PATH:-/}"
} >/etc/vpsentinel/agent.env
chown root:vpsentinel /etc/vpsentinel/agent.env
chmod 0640 /etc/vpsentinel/agent.env
install -o root -g root -m 0644 "$script_dir/../deploy/systemd/vpsentinel-agent.service" /etc/systemd/system/vpsentinel-agent.service

systemctl daemon-reload
systemctl enable --now vpsentinel-agent
systemctl --no-pager --full status vpsentinel-agent
