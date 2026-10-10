import { useConsole } from '../../hooks/useConsole'
import { AlertsView } from '../alerts/AlertsView'
import { NodesView } from '../nodes/NodesView'
import { OverviewView } from '../overview/OverviewView'

export function OverviewRoute() {
  const console = useConsole()
  return (
    <OverviewView
      nodes={console.nodes}
      alerts={console.alerts}
      activeAlerts={console.activeAlerts}
      nodeName={console.nodeName}
      onNode={console.selectNode}
    />
  )
}

export function NodesRoute() {
  const console = useConsole()
  return (
    <NodesView
      nodes={console.nodes}
      authedRequest={console.authedRequest}
      onRefresh={console.refresh}
      onSelect={console.selectNode}
      onAdd={console.addNode}
    />
  )
}

export function AlertsRoute() {
  const console = useConsole()
  return <AlertsView alerts={console.alerts} activeAlerts={console.activeAlerts} nodeName={console.nodeName} />
}
