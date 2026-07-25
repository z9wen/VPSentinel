# Agent installation and rotation

Create a node in the Web UI, generate an agent token, and copy it immediately.
The plaintext token is never available from the server again.

On the monitored Linux host, clone the repository or copy the installer and
systemd unit, then run:

```bash
sudo VPSENTINEL_VERSION=vYYYY.MM.DD.TIMESTAMP \
  VPSENTINEL_SERVER_URL=https://monitor.example.com/api/v1/agent/metrics \
  VPSENTINEL_TOKEN=the-one-time-token \
  ./scripts/install-agent.sh
```

The installer verifies the release checksum, creates a non-login service user,
writes `/etc/vpsentinel/agent.env`, and enables a hardened systemd unit.

Verify operation with:

```bash
systemctl status vpsentinel-agent
journalctl -u vpsentinel-agent --since '10 minutes ago'
```

To rotate a credential, generate a second token, update the environment file,
restart the agent, verify a new metric arrives, and only then revoke the old
token. This overlap prevents a monitoring gap.
