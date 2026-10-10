import { Plus, Trash2 } from 'lucide-react'
import { useState } from 'react'
import type { AuthedRequest, Node } from '../../types'
import { relativeTime } from '../../lib/format'
import { Button } from '../../components/ui/Button'
import { ConfirmDialog } from '../../components/ui/ConfirmDialog'
import { EmptyState } from '../../components/ui/EmptyState'
import { Panel } from '../../components/ui/Panel'
import { StatusBadge } from '../../components/ui/StatusBadge'

type Props = {
  nodes: Node[]
  authedRequest: AuthedRequest
  onRefresh: () => Promise<void>
  onSelect: (node: Node) => void
  onAdd: () => void
}

export function NodesView({ nodes, authedRequest, onRefresh, onSelect, onAdd }: Props) {
  const [deleting, setDeleting] = useState<Node | null>(null)

  async function remove(node: Node) {
    await authedRequest(`/api/v1/nodes/${node.id}`, { method: 'DELETE' })
    await onRefresh()
  }

  if (nodes.length === 0) {
    return (
      <section className="rounded-box border border-base-300 bg-base-100">
        <EmptyState
          title="No nodes yet"
          text="Register the first VPS, then install the agent with the token it receives."
          action={
            <Button size="sm" variant="primary" onClick={onAdd}>
              <Plus size={14} />
              Add node
            </Button>
          }
        />
      </section>
    )
  }

  return (
    <>
      <Panel title={`${nodes.length} ${nodes.length === 1 ? 'node' : 'nodes'}`} description="Status bands: online ≤ 2 min, degraded ≤ 5 min, offline beyond that.">
        <div className="overflow-x-auto">
          <table className="table">
            <thead>
              <tr className="text-xs text-base-content/55">
                <th className="font-medium">Node</th>
                <th className="font-medium">Status</th>
                <th className="font-medium">Last seen</th>
                <th className="hidden font-medium md:table-cell">Environment</th>
                <th className="hidden font-medium lg:table-cell">Agent</th>
                <th><span className="sr-only">Actions</span></th>
              </tr>
            </thead>
            <tbody>
              {nodes.map((node) => (
                <tr key={node.id} className="border-base-300">
                  <td>
                    <button type="button" className="text-left text-sm font-medium hover:underline" onClick={() => onSelect(node)}>
                      {node.name}
                    </button>
                    <div className="font-mono text-xs text-base-content/50">{node.hostname || node.id}</div>
                  </td>
                  <td><StatusBadge status={node.status} /></td>
                  <td className="whitespace-nowrap text-sm tabular-nums text-base-content/70">{relativeTime(node.last_seen_at)}</td>
                  <td className="hidden text-sm text-base-content/70 md:table-cell">{node.environment}</td>
                  <td className="hidden lg:table-cell">
                    <div className="font-mono text-xs">{node.agent_version}</div>
                    <div className="text-xs text-base-content/50">{[node.os_name, node.kernel_version].filter(Boolean).join(' · ') || '—'}</div>
                  </td>
                  <td>
                    <div className="flex justify-end gap-1">
                      <Button size="sm" variant="ghost" onClick={() => onSelect(node)}>Inspect</Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        className="btn-square text-base-content/60 hover:text-error"
                        onClick={() => setDeleting(node)}
                        aria-label={`Delete ${node.name}`}
                        title="Delete node"
                      >
                        <Trash2 size={15} />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Panel>

      <ConfirmDialog
        open={deleting !== null}
        title={`Delete ${deleting?.name ?? 'node'}?`}
        description="This removes the node, its agent tokens, and all retained metrics. Any agent still running on the host will be rejected."
        confirmLabel="Delete node"
        onConfirm={() => (deleting ? remove(deleting) : Promise.resolve())}
        onClose={() => setDeleting(null)}
      />
    </>
  )
}
