// Mirrors deploy/grafana/provisioning/alerting/rules.yml so a meter changes
// colour at the same value where Grafana starts evaluating an alert.
export const thresholds = {
  cpu: { warning: 85 },
  memory: { warning: 90 },
  disk: { warning: 80, critical: 90 },
}
