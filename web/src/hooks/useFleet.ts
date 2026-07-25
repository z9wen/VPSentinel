import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import type { Alert, AuthedRequest, Node } from '../types'

const emptyNodes: Node[] = []
const emptyAlerts: Alert[] = []

export function useFleet(authedRequest: AuthedRequest, enabled: boolean) {
  const query = useQuery({
    queryKey: ['fleet'],
    enabled,
    refetchInterval: 30_000,
    retry: 1,
    queryFn: async () => {
      const [nodeResult, alertResult] = await Promise.all([
        authedRequest<{ nodes: Node[] }>('/api/v1/nodes'),
        authedRequest<{ alerts: Alert[] }>('/api/v1/alerts?limit=100'),
      ])
      return { nodes: nodeResult.nodes, alerts: alertResult.alerts }
    },
  })

  const nodes = query.data?.nodes ?? emptyNodes
  const alerts = query.data?.alerts ?? emptyAlerts
  const activeAlerts = useMemo(() => {
    const latest = new Map<string, Alert>()
    alerts.forEach((alert) => {
      const key = `${alert.node_id ?? 'global'}:${alert.rule_name}`
      if (!latest.has(key)) latest.set(key, alert)
    })
    return [...latest.values()].filter((alert) => alert.state === 'firing')
  }, [alerts])

  return {
    nodes,
    alerts,
    activeAlerts,
    loading: query.isLoading,
    error: query.error instanceof Error ? query.error.message : '',
    load: async () => { await query.refetch() },
  }
}
