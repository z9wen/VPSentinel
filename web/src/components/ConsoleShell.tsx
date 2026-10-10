import { Activity, ExternalLink, LayoutDashboard, LogOut, Plus, RefreshCw, Server } from 'lucide-react'
import type { ReactNode } from 'react'
import { Link, useRouterState } from '@tanstack/react-router'
import { useServerVersion } from '../hooks/useServerVersion'
import { useTheme } from '../hooks/useTheme'
import { formatClock } from '../lib/format'
import { Brand } from './Brand'
import { ThemeToggle } from './ThemeToggle'
import { Button } from './ui/Button'

const pages = {
  '/overview': { label: 'Overview', Icon: LayoutDashboard, title: 'Overview', description: 'Fleet health and active incidents across every monitored host.' },
  '/nodes': { label: 'Nodes', Icon: Server, title: 'Nodes', description: 'Register hosts, issue agent tokens, and inspect the latest sample.' },
  '/alerts': { label: 'Incidents', Icon: Activity, title: 'Incidents', description: 'Firing and resolved events delivered by Grafana alerting.' },
} as const

type Path = keyof typeof pages

type Props = {
  nodeCount: number
  incidentCount: number
  updatedAt: number
  fetching: boolean
  children: ReactNode
  onRefresh: () => void
  onAddNode: () => void
  onLogout: () => void
}

export function ConsoleShell({ nodeCount, incidentCount, updatedAt, fetching, children, onRefresh, onAddNode, onLogout }: Props) {
  const pathname = useRouterState({ select: (state) => state.location.pathname })
  const current: Path = pathname.startsWith('/nodes') ? '/nodes' : pathname.startsWith('/alerts') ? '/alerts' : '/overview'
  const page = pages[current]
  const { theme, toggle } = useTheme()
  const { version } = useServerVersion()
  const counts: Record<Path, number> = { '/overview': 0, '/nodes': nodeCount, '/alerts': incidentCount }

  return (
    <div className="min-h-screen bg-base-200 lg:grid lg:grid-cols-[15rem_minmax(0,1fr)]">
      <aside className="sticky top-0 z-30 flex flex-col border-b border-base-300 bg-base-100 lg:h-screen lg:border-b-0 lg:border-r">
        <div className="flex h-14 items-center justify-between px-4 lg:h-16 lg:px-5">
          <Brand />
          <div className="flex items-center gap-1 lg:hidden">
            <ThemeToggle theme={theme} onToggle={toggle} />
            <button type="button" className="btn btn-square btn-ghost btn-sm text-base-content/70" onClick={onLogout} aria-label="Sign out" title="Sign out">
              <LogOut size={16} />
            </button>
          </div>
        </div>

        <nav className="flex gap-1 overflow-x-auto px-3 pb-2 lg:flex-1 lg:flex-col lg:pb-0 lg:pt-3" aria-label="Main">
          {(Object.keys(pages) as Path[]).map((path) => {
            const { label, Icon } = pages[path]
            const count = counts[path]
            return (
              <Link
                key={path}
                to={path}
                className="flex shrink-0 items-center gap-2.5 rounded-field px-3 py-2 text-sm font-medium transition-colors"
                activeProps={{ className: 'bg-base-200 text-base-content' }}
                inactiveProps={{ className: 'text-base-content/65 hover:bg-base-200/70 hover:text-base-content' }}
              >
                <Icon size={16} />
                {label}
                {count > 0 && (
                  <span className={`ml-auto min-w-5 rounded-full px-1.5 text-center text-xs tabular-nums ${path === '/alerts' ? 'bg-error text-error-content' : 'bg-base-300/70 text-base-content/70'}`}>
                    {count}
                  </span>
                )}
              </Link>
            )
          })}
          <a
            href="/grafana/"
            target="_blank"
            rel="noreferrer"
            className="flex shrink-0 items-center gap-2.5 rounded-field px-3 py-2 text-sm font-medium text-base-content/65 transition-colors hover:bg-base-200/70 hover:text-base-content lg:mt-auto"
          >
            <ExternalLink size={16} />
            Grafana
          </a>
        </nav>

        <footer className="hidden items-center gap-1 border-t border-base-300 px-3 py-3 lg:flex">
          <span className="truncate px-2 font-mono text-xs text-base-content/45" title="Server version">
            {version ? `server ${version}` : ''}
          </span>
          <span className="ml-auto" />
          <ThemeToggle theme={theme} onToggle={toggle} />
          <button type="button" className="btn btn-square btn-ghost btn-sm text-base-content/70" onClick={onLogout} aria-label="Sign out" title="Sign out">
            <LogOut size={16} />
          </button>
        </footer>
      </aside>

      <main className="min-w-0">
        <div className="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-10 lg:py-10">
          <header className="mb-6 flex flex-wrap items-end justify-between gap-4">
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">{page.title}</h1>
              <p className="mt-1 text-sm text-base-content/60">{page.description}</p>
            </div>
            <div className="flex items-center gap-2">
              {updatedAt > 0 && (
                <span className="hidden text-xs tabular-nums text-base-content/50 sm:inline">Updated {formatClock(updatedAt)}</span>
              )}
              <Button size="sm" onClick={onRefresh} disabled={fetching}>
                <RefreshCw size={14} className={fetching ? 'animate-spin' : ''} />
                Refresh
              </Button>
              {current === '/nodes' && (
                <Button size="sm" variant="primary" onClick={onAddNode}>
                  <Plus size={14} />
                  Add node
                </Button>
              )}
            </div>
          </header>
          {children}
        </div>
      </main>
    </div>
  )
}
