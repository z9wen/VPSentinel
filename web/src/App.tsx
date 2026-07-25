import { useState } from 'react'
import { Outlet } from '@tanstack/react-router'
import { ConsoleShell } from './components/ConsoleShell'
import { LoadingScreen } from './components/LoadingScreen'
import { Login } from './components/Login'
import { AddNodeDialog } from './features/nodes/AddNodeDialog'
import { NodeDetailDialog } from './features/nodes/NodeDetailDialog'
import { useFleet } from './hooks/useFleet'
import { useSession } from './hooks/useSession'
import { ConsoleContext, type ConsoleContextValue } from './hooks/useConsole'
import type { Node } from './types'

function App() {
  const session = useSession()
  const fleet = useFleet(session.authedRequest, session.state === 'authenticated')
  const [showAddNode, setShowAddNode] = useState(false)
  const [selectedNode, setSelectedNode] = useState<Node | null>(null)

  if (session.state === 'loading') return <LoadingScreen />
  if (session.state === 'anonymous') return <Login onAuthenticated={session.authenticate} />

  const context: ConsoleContextValue = {
    nodes: fleet.nodes,
    alerts: fleet.alerts,
    activeAlerts: fleet.activeAlerts,
    authedRequest: session.authedRequest,
    refresh: fleet.load,
    selectNode: setSelectedNode,
  }

  return <ConsoleContext.Provider value={context}><ConsoleShell nodeCount={fleet.nodes.length} incidentCount={fleet.activeAlerts.length} onRefresh={() => void fleet.load()} onAddNode={() => setShowAddNode(true)} onLogout={() => void session.logout()}>
    {fleet.error && <div className="alert alert-error mb-5 text-sm"><span>{fleet.error}</span><button className="btn btn-ghost btn-xs" onClick={() => void fleet.load()}>Retry</button></div>}
    {fleet.loading && fleet.nodes.length === 0 ? <div className="grid grid-cols-2 gap-3 xl:grid-cols-4">{[1, 2, 3, 4].map((value) => <div className="skeleton h-36" key={value} />)}</div> : <Outlet />}
    <AddNodeDialog open={showAddNode} authedRequest={session.authedRequest} onClose={() => setShowAddNode(false)} onCreated={fleet.load} />
    <NodeDetailDialog node={selectedNode} authedRequest={session.authedRequest} onClose={() => setSelectedNode(null)} />
  </ConsoleShell></ConsoleContext.Provider>
}

export default App
