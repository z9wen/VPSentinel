import { useState } from 'react'
import { Outlet } from '@tanstack/react-router'
import { ConsoleShell } from './components/ConsoleShell'
import { LoadingScreen } from './components/LoadingScreen'
import { Login } from './components/Login'
import { Button } from './components/ui/Button'
import { AddNodeDialog } from './features/nodes/AddNodeDialog'
import { NodeDetailSheet } from './features/nodes/NodeDetailSheet'
import { useFleet } from './hooks/useFleet'
import { useSession } from './hooks/useSession'
import { ConsoleContext, type ConsoleContextValue } from './hooks/useConsole'

function App() {
  const session = useSession()
  const fleet = useFleet(session.authedRequest, session.state === 'authenticated')
  const [showAddNode, setShowAddNode] = useState(false)
  // Track the selection by ID so the open sheet follows background refreshes.
  const [selectedNodeID, setSelectedNodeID] = useState<string | null>(null)

  if (session.state === 'loading') return <LoadingScreen />
  if (session.state === 'anonymous') return <Login onAuthenticated={session.authenticate} />

  const selectedNode = fleet.nodes.find((node) => node.id === selectedNodeID) ?? null
  const context: ConsoleContextValue = {
    nodes: fleet.nodes,
    alerts: fleet.alerts,
    activeAlerts: fleet.activeAlerts,
    nodeName: fleet.nodeName,
    authedRequest: session.authedRequest,
    refresh: fleet.load,
    selectNode: (node) => setSelectedNodeID(node.id),
    addNode: () => setShowAddNode(true),
  }

  return (
    <ConsoleContext.Provider value={context}>
      <ConsoleShell
        nodeCount={fleet.nodes.length}
        incidentCount={fleet.activeAlerts.length}
        updatedAt={fleet.updatedAt}
        fetching={fleet.fetching}
        onRefresh={() => void fleet.load()}
        onAddNode={() => setShowAddNode(true)}
        onLogout={() => void session.logout()}
      >
        {fleet.error && (
          <div className="mb-6 flex items-center justify-between gap-3 rounded-box border border-error/25 bg-error/10 px-4 py-3 text-sm" role="alert">
            <span>{fleet.error}</span>
            <Button size="sm" variant="ghost" onClick={() => void fleet.load()}>Retry</Button>
          </div>
        )}
        {fleet.loading ? (
          <div className="space-y-6">
            <div className="skeleton h-28 rounded-box" />
            <div className="skeleton h-72 rounded-box" />
          </div>
        ) : (
          <Outlet />
        )}
        <AddNodeDialog open={showAddNode} authedRequest={session.authedRequest} onClose={() => setShowAddNode(false)} onCreated={fleet.load} />
        <NodeDetailSheet node={selectedNode} authedRequest={session.authedRequest} onClose={() => setSelectedNodeID(null)} />
      </ConsoleShell>
    </ConsoleContext.Provider>
  )
}

export default App
