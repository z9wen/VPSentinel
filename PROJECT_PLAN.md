# VPSentinel Development Plan

## Implementation status (2026-07-25)

The repository now implements the software portion of the v0.1 MVP: agent
collection and retry behavior, PostgreSQL ingestion and management schema,
per-node token lifecycle, dual-token admin authentication, node/metric/incident
APIs, React operations UI, Grafana dashboards and alert provisioning, retention
cleanup, hardened systemd installation, Docker Compose, Nginx TLS examples, CI,
multi-architecture releases, and operator documentation.

The remaining apply-ready items require a real environment and must not be
claimed from repository code alone: operate two or more VPS agents for 24
hours, connect an external operator notification destination, execute the four
failure drills, record measured alert/recovery timings in a real postmortem,
verify the public firewall/TLS deployment, and add sanitised screenshots.

The original milestones below remain as the scope and acceptance record.

## 1. Project Summary

VPSentinel is a lightweight, self-hosted monitoring and alerting platform for a small fleet of Linux VPS instances running websites, APIs, and self-hosted services.

The first production target is the author's existing fleet of 2-3 VPS instances. The project should demonstrate practical DevOps and SRE skills: Linux operations, secure agent communication, metrics collection, dashboards, alerting, recovery notifications, incident records, automated builds, and repeatable deployment.

### Primary objective

Deliver a reliable MVP that can:

- monitor 2-3 real Linux VPS instances;
- show current fleet health and recent resource usage;
- alert when a node is offline or a resource threshold is exceeded;
- send a recovery notification when the condition clears;
- retain enough history to investigate an incident;
- deploy and upgrade without manually rebuilding every component.

### Portfolio objective

The repository should make the following evidence easy to verify:

- a real agent/server architecture written in Go;
- secure communication between internet-facing hosts;
- PostgreSQL schema design and retention management;
- Grafana dashboards and alert rules;
- Docker Compose and systemd operations;
- GitHub Actions for tests, linting, builds, and releases;
- failure testing, runbooks, and an incident postmortem.

## 2. MVP Scope

### Included in v0.1

#### Go agent

- Collect CPU usage, memory usage, disk usage, load average, uptime, and basic network counters.
- Send a heartbeat and metric batch at a configurable interval.
- Authenticate with a unique per-node token.
- Retry failed requests with exponential backoff and jitter.
- Log structured operational events.
- Run as a systemd service.
- Expose its version in every payload.

#### Go/Gin server

- Register and manage monitored nodes.
- Generate and revoke node tokens.
- Accept authenticated metric batches.
- Store node metadata, metrics, and alert events.
- Expose health and readiness endpoints.
- Provide APIs for the React application.
- Accept Grafana alert webhooks so incidents can be displayed in the application.

#### PostgreSQL

- Store users, nodes, token metadata, metric samples, and alert events.
- Retain 30 days of metric data initially.
- Provide indexes for node-and-time queries.
- Run schema changes through versioned migrations.

#### Grafana

- Fleet overview dashboard.
- Per-node CPU, memory, disk, load, uptime, and last-seen panels.
- Offline, disk, CPU, and memory alerts.
- Firing and resolved notifications.
- Dashboard and alert provisioning stored in Git.

#### React + TypeScript + Vite

- Admin login.
- Fleet overview with online, degraded, and offline states.
- Node detail and latest metrics.
- Active alerts and alert history.
- Links from nodes to the relevant Grafana dashboard.

#### Delivery and operations

- Docker Compose for server, PostgreSQL, Grafana, and web application.
- Nginx reverse proxy and HTTPS for the public deployment.
- GitHub Actions for Go tests, linting, frontend checks, image builds, and agent releases.
- Linux amd64 and arm64 agent binaries.
- Installation and upgrade documentation.

### Explicitly out of scope for v0.1

- Kubernetes.
- Terraform or Ansible as a release blocker.
- Remote shell or arbitrary command execution.
- Full log aggregation.
- Distributed tracing.
- Multi-tenant billing.
- Mobile applications.
- Machine-learning anomaly detection.
- Replacing Grafana with a custom charting system.

These may be added after the monitoring and alerting path is stable.

## 3. Architecture

### High-level data flow

    Linux VPS
      └── vpsentinel-agent
            ├── collect host metrics
            ├── batch samples
            └── POST over HTTPS with node token
                         |
                         v
    Nginx -> Go/Gin API -> PostgreSQL
                 |             |
                 |             └── Grafana queries metrics
                 |
                 └── React application queries fleet and incident APIs

    Grafana Alerting
      ├── sends operator notification
      └── sends webhook to Go/Gin API -> alert_events

### Architecture decisions

1. **Push-based agent**
   The current VPS fleet is distributed across the internet. Agents initiate outbound HTTPS connections, so no new public metrics port is required on each host.

2. **PostgreSQL for the MVP**
   At a 15-second interval, three nodes create about 518,400 rows in 30 days when one row represents one complete host sample. This is manageable for the initial scale with appropriate indexes and retention cleanup.

3. **Grafana owns visualisation and alert evaluation**
   React should manage nodes, status, configuration, and incidents. It should not duplicate Grafana's time-series visualisation features.

4. **No remote command execution**
   The first version is monitoring-only. This limits the security impact of a compromised control plane or token.

5. **Terraform later**
   The first deployment may be manual and documented. Infrastructure as code can be added after the application and operational requirements are understood.

## 4. Proposed Repository Layout

    .
    ├── agent/
    │   ├── cmd/vpsentinel-agent/
    │   ├── internal/collector/
    │   ├── internal/config/
    │   ├── internal/sender/
    │   └── internal/version/
    ├── server/
    │   ├── cmd/vpsentinel-server/
    │   ├── internal/api/
    │   ├── internal/auth/
    │   ├── internal/config/
    │   ├── internal/database/
    │   ├── internal/ingest/
    │   └── internal/incidents/
    ├── web/
    │   ├── src/
    │   └── public/
    ├── migrations/
    ├── deploy/
    │   ├── compose/
    │   ├── grafana/
    │   │   ├── dashboards/
    │   │   ├── provisioning/
    │   │   └── alerts/
    │   ├── nginx/
    │   └── systemd/
    ├── docs/
    │   ├── architecture.md
    │   ├── agent-installation.md
    │   ├── deployment.md
    │   ├── runbooks/
    │   └── postmortems/
    ├── scripts/
    ├── .github/workflows/
    ├── docker-compose.yml
    ├── Makefile
    ├── PROJECT_PLAN.md
    └── README.md

## 5. Initial Technology Choices

### Agent

- Go
- gopsutil for portable host metrics
- Standard library HTTP client
- log/slog for structured logging
- Environment file plus command-line flags for configuration

### Server

- Go and Gin
- pgx for PostgreSQL access
- Versioned SQL migrations
- log/slog for structured logs
- JWT for the web admin session
- Hashed, unique bearer tokens for agents

### Frontend

- React
- TypeScript
- Vite
- React Router
- TanStack Query for API state

### Operations

- PostgreSQL
- Grafana Unified Alerting
- Docker Compose
- Nginx
- GitHub Actions
- systemd

## 6. Data Model

### users

- id
- email
- password_hash
- created_at
- updated_at

### nodes

- id
- name
- hostname
- environment
- public_ip or display_address
- agent_version
- status
- last_seen_at
- created_at
- updated_at

### agent_tokens

- id
- node_id
- token_hash
- created_at
- last_used_at
- revoked_at

Never store the plaintext token after it is shown to the operator.

### metric_samples

- id
- node_id
- recorded_at
- received_at
- cpu_percent
- memory_used_bytes
- memory_total_bytes
- disk_used_bytes
- disk_total_bytes
- load_1
- load_5
- load_15
- uptime_seconds
- network_rx_bytes
- network_tx_bytes

Initial indexes:

- node_id, recorded_at descending
- recorded_at for retention cleanup
- optional BRIN index on recorded_at after the table grows

### alert_events

- id
- fingerprint
- node_id
- rule_name
- severity
- state
- summary
- starts_at
- ends_at
- received_at
- raw_payload

Use the Grafana fingerprint plus alert state to make webhook processing idempotent.

## 7. API Draft

### Agent endpoints

- POST /api/v1/agent/metrics
- POST /api/v1/agent/heartbeat

The metric payload should support a small batch so temporary network failures do not require one HTTP request per sample.

### Admin endpoints

- POST /api/v1/auth/login
- GET /api/v1/nodes
- POST /api/v1/nodes
- GET /api/v1/nodes/:id
- DELETE /api/v1/nodes/:id
- POST /api/v1/nodes/:id/tokens
- DELETE /api/v1/nodes/:id/tokens/:tokenId
- GET /api/v1/nodes/:id/metrics/latest
- GET /api/v1/alerts
- GET /api/v1/alerts/:id

### Integration and health endpoints

- POST /api/v1/integrations/grafana/webhook
- GET /healthz
- GET /readyz
- GET /version

## 8. Alert Rules

The first rules should be understandable and testable.

### Node offline

- Condition: no accepted heartbeat for more than 2 minutes.
- Severity: critical.
- Recovery: heartbeat is accepted again.

### Disk space

- Warning: disk usage above 80 percent for 5 minutes.
- Critical: disk usage above 90 percent for 5 minutes.
- Recovery: disk usage below the configured threshold.

### CPU

- Warning: CPU usage above 85 percent for 10 minutes.
- Recovery: CPU usage below the threshold.

### Memory

- Warning: memory usage above 90 percent for 10 minutes.
- Recovery: memory usage below the threshold.

Thresholds must be configurable rather than compiled into the agent.

## 9. Security Requirements

- All production agent traffic uses HTTPS with certificate verification enabled.
- Every node receives a different random token.
- Tokens are stored hashed on the server.
- Agent tokens are stored in a root-readable configuration file with mode 0600.
- Requests have a strict body-size limit and schema validation.
- Server timestamps the receipt of every sample and rejects unreasonable timestamps.
- Rate limiting is applied per token and source address.
- CORS is restricted to the deployed web origin.
- PostgreSQL and Grafana are not directly exposed to the internet.
- Secrets and real server addresses are not committed to Git.
- Example configuration files contain placeholders only.
- The agent never accepts remote shell commands in v0.1.

## 10. Reliability Requirements

- Metric collection failure for one collector must not crash the agent.
- HTTP failures use bounded exponential backoff with jitter.
- The agent continues collecting during temporary server failure.
- v0.1 may use a bounded in-memory queue.
- A disk-backed spool is a v0.2 reliability improvement.
- Duplicate metric batches are safe to retry through a batch identifier or uniqueness constraint.
- Server shutdown is graceful and stops accepting new requests before closing database connections.
- Database migrations run explicitly, not automatically on every application start.
- Retention cleanup is observable and logged.

Initial targets:

- Agent memory usage below 40 MB under normal operation.
- Agent average CPU usage below 1 percent on a typical VPS.
- Offline alert delivered within 3 minutes.
- Recovery notification delivered within 2 minutes after service restoration.

These are engineering targets and must be measured before being claimed in the README or resume.

## 11. Development Milestones

### Milestone 0: Repository foundation

- Create the repository layout.
- Initialise Go modules for agent and server.
- Initialise the React TypeScript application.
- Add Makefile commands for format, lint, test, build, and local startup.
- Add editor configuration, ignore files, and example environment files.
- Add a local Docker Compose stack with PostgreSQL and Grafana.

Exit condition: a new contributor can clone the repository and run the empty development stack.

### Milestone 1: Local agent

- Implement the collector interfaces.
- Collect CPU, memory, disk, load, uptime, and network counters.
- Add configuration validation.
- Produce a versioned JSON batch payload.
- Add structured logs and unit tests.
- Add a dry-run mode that prints a sample without sending it.

Exit condition: the agent produces stable, tested samples on Linux.

### Milestone 2: Ingestion server

- Create database migrations.
- Implement node creation and token generation.
- Authenticate metric requests.
- Validate and store metric batches.
- Update node last-seen state.
- Implement health, readiness, and version endpoints.
- Add server unit and PostgreSQL integration tests.

Exit condition: a local agent can send metrics that can be queried from PostgreSQL.

### Milestone 3: Real VPS deployment

- Build Linux amd64 and arm64 binaries.
- Create the systemd unit and installation documentation.
- Deploy agents to the existing VPS fleet.
- Deploy the server stack behind Nginx and HTTPS.
- Confirm that no database or Grafana port is publicly reachable.

Exit condition: 2-3 real nodes report continuously for at least 24 hours.

### Milestone 4: Grafana and alerting

- Provision the PostgreSQL data source.
- Build fleet and per-node dashboards.
- Create offline, disk, CPU, and memory rules.
- Configure one real notification channel.
- Send alert webhooks to the server.
- Verify firing and resolved notifications.

Exit condition: a controlled agent shutdown creates an alert and a restart resolves it.

### Milestone 5: React operations UI

- Implement admin login.
- Display fleet status and last-seen time.
- Display node details and latest metrics.
- Display active alerts and alert history.
- Link to the corresponding Grafana dashboard.
- Add loading, empty, error, and unauthorised states.

Exit condition: routine fleet and incident inspection does not require direct database access.

### Milestone 6: Reliability and security hardening

- Add request limits and rate limiting.
- Add token revocation and rotation.
- Make metric retries idempotent.
- Add graceful shutdown.
- Add retention cleanup.
- Measure agent CPU and memory overhead.
- Review deployment files for secrets and unsafe defaults.

Exit condition: security and reliability requirements in this document are either verified or documented as limitations.

### Milestone 7: CI, releases, and portfolio documentation

- Run Go tests and linting in GitHub Actions.
- Run frontend lint, type-check, test, and build.
- Build container images.
- Publish versioned agent binaries with checksums.
- Add an architecture diagram and screenshots.
- Add deployment and agent installation guides.
- Add at least one runbook and one failure-test postmortem.
- Document current limitations and roadmap.

Exit condition: a recruiter can understand the problem, architecture, operational evidence, and the author's responsibilities from the repository homepage.

## 12. Failure Drills

Run each test against a non-critical node or controlled environment.

### Drill 1: Agent stopped

1. Stop vpsentinel-agent.
2. Record the last accepted heartbeat.
3. Measure time to alert delivery.
4. Restart the agent.
5. Measure time to recovery notification.
6. Save the incident timeline.

### Drill 2: Disk threshold

1. Use a temporary test filesystem or controlled file allocation.
2. Cross the warning threshold.
3. Confirm the alert contains node, value, threshold, and suggested action.
4. Remove the test data.
5. Confirm recovery.

### Drill 3: Control plane unavailable

1. Stop or block the ingestion server briefly.
2. Confirm agents continue running and retry with backoff.
3. Restore the server.
4. Confirm queued samples or the latest state are delivered as designed.
5. Verify that retry behaviour does not overload the server.

### Drill 4: Invalid or revoked token

1. Revoke one node token.
2. Confirm requests are rejected and logged without leaking the token.
3. Rotate the token.
4. Confirm reporting resumes.

## 13. Testing Strategy

### Agent

- Unit-test collectors behind interfaces.
- Test configuration validation.
- Test JSON payload compatibility.
- Test retry timing with an injectable clock or short test policy.
- Test cancellation and graceful shutdown.

### Server

- Test authentication and token revocation.
- Test payload validation and size limits.
- Test duplicate batch handling.
- Test node status transitions.
- Test Grafana webhook idempotency.
- Run integration tests against PostgreSQL in CI.

### Frontend

- Test authentication state.
- Test fleet status rendering.
- Test loading, empty, error, and alert states.
- Add one end-to-end smoke test after the main flow is stable.

### Deployment

- Validate Docker Compose configuration.
- Validate Nginx configuration.
- Verify systemd unit hardening options.
- Run a fresh-install test on a disposable Linux VM or VPS.

## 14. Apply-Ready Checkpoint

The project is ready to place prominently on a resume when all of the following are true:

- at least two real VPS nodes have reported for 24 hours;
- dashboards show meaningful historical data;
- offline and resource alerts have been tested;
- recovery notifications work;
- the server is deployed over HTTPS;
- agent installation is repeatable;
- CI is green;
- the public repository contains no credentials or sensitive host information;
- README includes architecture, screenshots, setup, limitations, and a failure drill.

Terraform, Kubernetes, log aggregation, and advanced anomaly detection are not required for this checkpoint.

## 15. Post-MVP Roadmap

### v0.2

- Disk-backed agent spool.
- Process and systemd service checks.
- HTTP endpoint checks for websites and APIs.
- Agent upgrade workflow.
- Per-node alert thresholds.
- Silence and maintenance-window support.
- Loki log integration.

### v0.3

- Terraform for a disposable monitoring server or test environment.
- Ansible role for agent installation and upgrades.
- Prometheus-compatible long-term metric storage if PostgreSQL becomes a limitation.
- mTLS between agents and server.
- SLI and SLO dashboards for the monitoring platform itself.

### Later

- Multi-user roles.
- Multi-tenant support.
- Additional operating systems.
- Kubernetes deployment only if it solves a real operational requirement.

## 16. Immediate Next Tasks

1. Create the proposed repository directories.
2. Decide the metric payload schema and commit it under docs.
3. Implement agent dry-run collection.
4. Add PostgreSQL and Grafana to Docker Compose.
5. Create the first database migration.
6. Implement authenticated metric ingestion.
7. Deploy one test agent before building the full React interface.

The first vertical slice should be:

    one real agent
      -> authenticated HTTPS request
      -> one PostgreSQL metric row
      -> one Grafana panel
      -> one tested offline alert

Complete this path before adding additional collectors or UI features.
