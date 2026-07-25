export type SessionState = 'loading' | 'anonymous' | 'authenticated'
export type View = 'overview' | 'nodes' | 'alerts'
export type NodeStatus = 'online' | 'degraded' | 'offline' | 'pending'

export type Node = {
  id: string
  name: string
  hostname: string
  environment: string
  display_address: string | null
  agent_version: string
  os_name: string | null
  kernel_version: string | null
  last_seen_at: string | null
  created_at: string
  status: NodeStatus
  grafana_url: string
}

export type Metric = {
  node_id: string
  recorded_at: string
  cpu_percent: number | null
  memory_used_bytes: number | null
  memory_total_bytes: number | null
  disk_used_bytes: number | null
  disk_total_bytes: number | null
  load_1: number | null
  load_5: number | null
  load_15: number | null
  uptime_seconds: number | null
  network_rx_bytes: number | null
  network_tx_bytes: number | null
  failed_collectors: string[]
}

export type Alert = {
  id: number
  node_id: string | null
  rule_name: string
  severity: string
  state: 'firing' | 'resolved'
  summary: string
  starts_at: string
  ends_at: string | null
  received_at: string
}

export type AccessResponse = { access_token: string; expires_at: string }
export type AuthedRequest = <T>(path: string, init?: RequestInit) => Promise<T>

export type AgentToken = {
  id: number
  node_id: string
  label: string
  created_at: string
  last_used_at: string | null
  revoked_at: string | null
}
