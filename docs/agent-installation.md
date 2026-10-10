# Agent installation and rotation

Create a node in the console and copy its agent token immediately. The
plaintext token is never available from the server again.

On the monitored Linux host (amd64 or arm64), download the installer from the
same release you deploy and verify it against the release checksums:

```bash
TAG=vYYYY.MM.DD.TIMESTAMP
curl -fsSLO "https://github.com/z9wen/VPSentinel/releases/download/$TAG/install-agent.sh"
curl -fsSLO "https://github.com/z9wen/VPSentinel/releases/download/$TAG/checksums.txt"
grep '  install-agent.sh$' checksums.txt | sha256sum --check

sudo VPSENTINEL_VERSION="$TAG" \
  VPSENTINEL_SERVER_URL=https://monitor.example.com/api/v1/agent/metrics \
  VPSENTINEL_TOKEN=the-one-time-token \
  sh install-agent.sh
```

The installer downloads the agent binary and systemd unit from that release,
verifies both against `checksums.txt`, creates a non-login service user, writes
`/etc/vpsentinel/agent.env` (mode `0640`), and enables the hardened unit.
Rerunning it with a newer tag upgrades the binary and restarts the service.

Verify operation with:

```bash
systemctl status vpsentinel-agent
journalctl -u vpsentinel-agent --since '10 minutes ago'
```

Plain HTTP is refused unless `VPSENTINEL_ALLOW_INSECURE_HTTP=true` is set, which
is meant only for traffic that stays inside a private network.

To rotate a credential, generate a second token, update the environment file,
restart the agent, verify a new metric arrives, and only then revoke the old
token. This overlap prevents a monitoring gap.
