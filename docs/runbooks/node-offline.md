# Runbook: node offline

## Trigger

Grafana reports no accepted heartbeat or metric batch for more than two minutes.

## Triage

1. Check the node's last-seen timestamp and recent resource graphs.
2. Test whether the hosted service is reachable from an independent network.
3. Check the provider status page and the control-plane `/readyz` endpoint.
4. If SSH is available, inspect `systemctl status vpsentinel-agent` and
   `journalctl -u vpsentinel-agent --since '15 minutes ago'`.
5. Classify the failure as node, network, agent, credential, or control plane.

## Recovery

- Restart only the failed agent after capturing logs.
- For `401`, rotate the node token and verify reporting before revoking any
  remaining working token.
- For control-plane failure, follow the deployment rollback procedure and
  confirm queued samples drain without a request spike.

## Verification

Confirm a resolved notification arrives, the node becomes online, new samples
have current timestamps, and hosted services are healthy. Record detection,
acknowledgement, recovery, cause, and follow-up actions in a postmortem.
