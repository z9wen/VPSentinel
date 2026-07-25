import { Activity, LayoutDashboard, LogOut, Plus, RefreshCw, Server } from 'lucide-react'
import type { ReactNode } from 'react'
import { Link, useRouterState } from '@tanstack/react-router'
import { Brand } from './Brand'
import { Button } from './ui/Button'

const navigation = [
  { path: '/overview', label: 'Overview', Icon: LayoutDashboard },
  { path: '/nodes', label: 'Nodes', Icon: Server },
  { path: '/alerts', label: 'Incidents', Icon: Activity },
]

type Props = {
  nodeCount: number
  incidentCount: number
  children: ReactNode
  onRefresh: () => void
  onAddNode: () => void
  onLogout: () => void
}

export function ConsoleShell({ nodeCount, incidentCount, children, onRefresh, onAddNode, onLogout }: Props) {
	const pathname = useRouterState({ select: (state) => state.location.pathname })
	const view = pathname.startsWith('/nodes') ? 'nodes' : pathname.startsWith('/alerts') ? 'alerts' : 'overview'
  const title = view === 'overview' ? 'Fleet overview' : view === 'nodes' ? 'Monitored nodes' : 'Incident history'
  return <div className="min-h-screen bg-base-200 lg:grid lg:grid-cols-[250px_1fr]">
    <aside className="border-b border-base-300 bg-base-100 p-4 lg:sticky lg:top-0 lg:h-screen lg:border-b-0 lg:border-r lg:p-6"><Brand /><nav className="mt-5 grid grid-cols-3 gap-1 lg:mt-12 lg:grid-cols-1">{navigation.map(({ path, label, Icon }) => { const count = path === '/nodes' ? nodeCount : path === '/alerts' ? incidentCount : 0; return <Link key={path} to={path} className="btn justify-start border-0" activeProps={{ className: 'btn-active bg-base-300 text-base-content' }} inactiveProps={{ className: 'btn-ghost text-base-content/55' }}><Icon size={17} /><span className="hidden sm:inline">{label}</span>{count > 0 && <span className={`badge badge-sm ml-auto ${path === '/alerts' ? 'badge-error' : 'badge-ghost'}`}>{count}</span>}</Link> })}</nav><div className="absolute bottom-6 hidden w-[202px] items-center gap-2 border-t border-base-300 pt-5 text-xs text-base-content/45 lg:flex"><span className="size-2 rounded-full bg-success shadow-[0_0_8px] shadow-success" />Control plane ready<button className="btn btn-ghost btn-xs ml-auto" onClick={onLogout}><LogOut size={14} /></button></div></aside>
    <main className="min-w-0 p-4 sm:p-6 lg:p-10 xl:p-14"><header className="mb-8 flex flex-wrap items-end justify-between gap-5"><div><p className="font-mono text-[10px] tracking-[.2em] text-base-content/40">OPERATIONS / {view.toUpperCase()}</p><h1 className="mt-2 text-3xl font-semibold tracking-[-.04em] sm:text-4xl">{title}</h1></div><div className="flex gap-2"><Button size="sm" onClick={onRefresh}><RefreshCw size={14} />Refresh</Button>{view === 'nodes' && <Button variant="primary" size="sm" onClick={onAddNode}><Plus size={14} />Add node</Button>}</div></header>{children}</main>
  </div>
}
