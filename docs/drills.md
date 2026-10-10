# Failure drills on a disposable AWS fleet

`deploy/terraform/drill` creates a short-lived environment for the failure
drills in `PROJECT_PLAN.md`: one control plane running the released stack and
up to five monitored nodes. Create it, run the drills, record the timings, and
destroy it the same day.

## What it creates

- A VPC with one public subnet and no NAT gateway. Instances have outbound
  HTTPS only; nothing accepts inbound traffic from the internet.
- A `t4g.small` control plane (Amazon Linux 2023, arm64) that installs the
  release with `scripts/install-server.sh`, registers `drill-1..N`, and
  publishes each agent token as an SSM SecureString. It also monitors itself as
  `primary-monitor`.
- `t4g.nano` drill nodes. Each has its own IAM role that can read only its own
  token, waits for it, and installs the agent with the release installer.
- IMDSv2 with a hop limit of 1, so containers cannot reach instance-role
  credentials. Operators connect through SSM Session Manager; there are no SSH
  keys or open SSH ports.
- Optionally, an account-wide AWS Budgets alert.

Drill nodes reach the control plane over plain HTTP on its private address, so
traffic and tokens stay inside the VPC. The production path terminates TLS at
the host Nginx instead.

## Prerequisites

- Terraform 1.6 or newer, AWS CLI v2, and the Session Manager plugin.
- Short-lived AWS credentials (for example `aws sso login`) allowed to manage
  EC2, VPC, IAM roles, SSM parameters, and Budgets.
- A VPSentinel release that includes the `migrate` job and the
  `install-agent.sh` release asset (any release built after they were added).

## Bring it up

```bash
cp deploy/terraform/drill/terraform.tfvars.example deploy/terraform/drill/terraform.tfvars
# set release_tag and budget_alert_email
make drill-up
```

Bootstrap takes about five minutes after `apply` returns. Follow it with the
`bootstrap_log_command` output, then open the console:

```bash
make drill-password   # prints the admin password (user admin@example.com)
make drill-console    # keep running; browse http://127.0.0.1:18080
```

## Helpers

```bash
tf() { terraform -chdir=deploy/terraform/drill "$@"; }
node_id() { tf output -json agent_instance_ids | jq -r --arg name "$1" '.[$name]'; }
run_on() {
  aws ssm send-command --instance-ids "$1" --document-name AWS-RunShellScript \
    --parameters "$(jq -n --arg command "$2" '{commands: [$command]}')" \
    --query Command.CommandId --output text
}
now() { date -u +%FT%TZ; }
```

Record every timestamp in UTC. The incident list in the console shows when
each Grafana notification arrived.

## Drill 1: agent stopped

```bash
now; run_on "$(node_id drill-1)" 'systemctl stop vpsentinel-agent'
# wait for "Node offline" on drill-1 (rule: no report for 2 minutes)
now; run_on "$(node_id drill-1)" 'systemctl start vpsentinel-agent'
# wait for the matching "Resolved" event
```

Measure the time from the stop to the firing event and from the start to the
resolved event. Target: alert within 3 minutes, recovery within 2 minutes.

## Drill 2: disk threshold

Fill the root filesystem to 85% with a removable file, then remove it:

```bash
now; run_on "$(node_id drill-2)" 'size=$(df --output=size -B1 / | tail -1); used=$(df --output=used -B1 / | tail -1); fallocate -l $((size * 85 / 100 - used)) /var/tmp/vpsentinel-drill.fill'
# "Disk usage warning" fires after 5 minutes above 80%
now; run_on "$(node_id drill-2)" 'rm -f /var/tmp/vpsentinel-drill.fill'
```

Check that the summary names the node, the current value, and the threshold.

## Drill 3: control plane unavailable

```bash
now; run_on "$(tf output -raw control_plane_instance_id)" 'cd /opt/vpsentinel && docker compose stop server'
# wait 3 minutes, then restore
now; run_on "$(tf output -raw control_plane_instance_id)" 'cd /opt/vpsentinel && docker compose start server'
```

Agents keep collecting and retry with jittered backoff; after recovery their
journals show batches carrying the queued samples, and the Grafana panels have
no gap. Expect "Node offline" for every node during the outage: the monitor
cannot tell its own failure from a fleet-wide one. Record this; it is the
reason for the external uptime check in the deployment guide.

## Drill 4: revoked token

1. In the console, revoke drill-2's token. Its journal shows `401` rejections
   that never contain the token, and "Node offline" fires.
2. Issue a new token, store it where the node can read it, and reload the
   agent. The token never appears in SSM command history:

```bash
aws ssm put-parameter --overwrite --type SecureString \
  --name /vpsentinel-drill/agent-tokens/drill-2 --value 'NEW_TOKEN'
now; run_on "$(node_id drill-2)" 'token=$(aws ssm get-parameter --name /vpsentinel-drill/agent-tokens/drill-2 --with-decryption --query Parameter.Value --output text) && sed -i "s|^VPSENTINEL_TOKEN=.*|VPSENTINEL_TOKEN=$token|" /etc/vpsentinel/agent.env && systemctl restart vpsentinel-agent'
```

Confirm reporting resumes and the incident resolves.

## Record the evidence

| Drill | Action (UTC) | Alert received | Recovered | Notes |
| --- | --- | --- | --- | --- |
| 1 Agent stopped | | | | |
| 2 Disk threshold | | | | |
| 3 Control plane down | | | | |
| 4 Revoked token | | | | |

Turn the most instructive drill into a postmortem with
`docs/postmortems/template.md`, and keep sanitised screenshots of the console
and Grafana.

## Tear down

```bash
make drill-down
aws ec2 describe-instances --region ap-southeast-1 \
  --filters Name=tag:Project,Values=VPSentinel Name=instance-state-name,Values=pending,running \
  --query 'Reservations[].Instances[].InstanceId'
```

`destroy` removes the instances, network, roles, and token parameters. The local
`terraform.tfstate` holds this environment's secrets; delete it after
destroying.

## Cost

Three instances for 48 hours cost a few US dollars: Graviton instance hours,
one public IPv4 address per instance, and small gp3 volumes. Check current
regional pricing before you start, and set `budget_alert_email` so AWS Budgets
warns you if anything is left running.
