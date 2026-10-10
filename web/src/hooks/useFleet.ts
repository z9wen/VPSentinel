import { useCallback, useMemo } from 'react'
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

  // Alerts arrive newest first, so the first event per node and rule is its current state.
  const activeAlerts = useMemo(() => {
    const latest = new Map<string, Alert>()
    alerts.forEach((alert) => {
      const key = `${alert.node_id ?? 'global'}:${alert.rule_name}`
      if (!latest.has(key)) latest.set(key, alert)
    })
    return [...latest.values()].filter((alert) => alert.state === 'firing')
  }, [alerts])

  const names = useMemo(() => new Map(nodes.map((node) => [node.id, node.name])), [nodes])
  const nodeName = useCallback((id: string | null) => (id ? names.get(id) ?? id : 'Control plane'), [names])

  const { refetch } = query
  const load = useCallback(async () => {
    await refetch()
  }, [refetch])

  return {
    nodes,
    alerts,
    activeAlerts,
    nodeName,
    loading: query.isLoading,
    fetching: query.isFetching,
    updatedAt: query.dataUpdatedAt,
    error: query.error instanceof Error ? query.error.message : '',
    load,
  }
}
