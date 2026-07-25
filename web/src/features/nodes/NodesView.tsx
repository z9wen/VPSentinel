import { Eye, KeyRound, Server, Trash2 } from 'lucide-react'
import { useState } from 'react'
import type { AuthedRequest, Node } from '../../types'
import { relativeTime } from '../../lib/format'
import { Button } from '../../components/ui/Button'
import { EmptyState } from '../../components/ui/EmptyState'
import { OneTimeToken } from '../../components/ui/OneTimeToken'
import { StatusBadge } from '../../components/ui/StatusBadge'

type Props = { nodes: Node[]; authedRequest: AuthedRequest; onRefresh: () => Promise<void>; onSelect: (node: Node) => void }

export function NodesView({ nodes, authedRequest, onRefresh, onSelect }: Props) {
  const [token, setToken] = useState('')
  const [busy, setBusy] = useState('')

  async function generate(node: Node) {
    setBusy(node.id)
    try {
      const result = await authedRequest<{ token: string }>(`/api/v1/nodes/${node.id}/tokens`, { method: 'POST', body: JSON.stringify({ label: `web-${new Date().toISOString().slice(0, 10)}` }) })
      setToken(result.token)
    } finally {
      setBusy('')
    }
  }

  async function remove(node: Node) {
    if (!window.confirm(`Delete ${node.name} and all of its metrics?`)) return
    await authedRequest(`/api/v1/nodes/${node.id}`, { method: 'DELETE' })
    await onRefresh()
  }

  return <div className="space-y-5">{token && <OneTimeToken token={token} onClose={() => setToken('')} />}{nodes.length === 0 ? <section className="rounded-box border border-base-300 bg-base-100"><EmptyState title="No nodes yet" text="Register the first VPS and install its agent." /></section> : <section className="grid gap-4 md:grid-cols-2 2xl:grid-cols-3">{nodes.map((node) => <article className="card border border-base-300 bg-base-100" key={node.id}><div className="card-body p-5"><div className="flex justify-between"><StatusBadge status={node.status} /><span className="font-mono text-[10px] text-base-content/35">{node.id.slice(0, 8)}</span></div><div className="mt-4"><h2 className="card-title text-lg"><Server size={17} className="text-primary" />{node.name}</h2><p className="text-xs text-base-content/45">{node.hostname || 'Waiting for first heartbeat'}</p></div><dl className="my-3 space-y-2 text-xs">{[['Environment', node.environment], ['Last seen', relativeTime(node.last_seen_at)], ['Version', node.agent_version]].map(([label, value]) => <div className="flex justify-between border-b border-base-300 pb-2" key={label}><dt className="text-base-content/40">{label}</dt><dd className="font-mono">{value}</dd></div>)}</dl><div className="card-actions mt-2"><Button size="sm" onClick={() => onSelect(node)}><Eye size={14} />Inspect</Button><Button size="sm" disabled={busy === node.id} onClick={() => void generate(node)}>{busy === node.id ? <span className="loading loading-spinner loading-xs" /> : <KeyRound size={14} />}New token</Button><Button variant="danger" size="sm" className="ml-auto btn-square" title="Delete node" onClick={() => void remove(node)}><Trash2 size={15} /></Button></div></div></article>)}</section>}</div>
}
