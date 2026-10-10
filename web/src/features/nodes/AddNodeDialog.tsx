import { useState } from 'react'
import type { FormEvent } from 'react'
import type { AuthedRequest, Node } from '../../types'
import { Button } from '../../components/ui/Button'
import { Modal } from '../../components/ui/Modal'
import { TokenReveal } from './TokenReveal'

type Props = { open: boolean; authedRequest: AuthedRequest; onClose: () => void; onCreated: () => Promise<void> }

export function AddNodeDialog({ open, authedRequest, onClose, onCreated }: Props) {
  const [name, setName] = useState('')
  const [hostname, setHostname] = useState('')
  const [environment, setEnvironment] = useState('production')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [token, setToken] = useState('')

  async function submit(event: FormEvent) {
    event.preventDefault()
    setError('')
    setSubmitting(true)
    try {
      const node = await authedRequest<Node>('/api/v1/nodes', { method: 'POST', body: JSON.stringify({ name, hostname, environment }) })
      const result = await authedRequest<{ token: string }>(`/api/v1/nodes/${node.id}/tokens`, { method: 'POST', body: JSON.stringify({ label: 'initial-install' }) })
      setToken(result.token)
      await onCreated()
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : 'Could not register node')
    } finally {
      setSubmitting(false)
    }
  }

  function close() {
    setName('')
    setHostname('')
    setEnvironment('production')
    setError('')
    setToken('')
    onClose()
  }

  return (
    <Modal
      open={open}
      title={token ? 'Install the agent' : 'Add node'}
      description={token ? `${name} is registered and waiting for its first report.` : 'Register a host and issue its first agent token.'}
      onClose={close}
      footer={token ? <Button size="sm" variant="primary" onClick={close}>Done</Button> : undefined}
    >
      {token ? (
        <TokenReveal token={token} />
      ) : (
        <form className="space-y-4" onSubmit={submit}>
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium">Display name</span>
            <input className="input w-full" value={name} onChange={(event) => setName(event.target.value)} placeholder="hk-api-01" required autoFocus />
          </label>
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium">
              Expected hostname <span className="font-normal text-base-content/50">(optional)</span>
            </span>
            <input className="input w-full" value={hostname} onChange={(event) => setHostname(event.target.value)} placeholder="api-hk-01.internal" />
          </label>
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium">Environment</span>
            <select className="select w-full" value={environment} onChange={(event) => setEnvironment(event.target.value)}>
              <option>production</option>
              <option>staging</option>
              <option>development</option>
            </select>
          </label>
          {error && <p className="text-sm text-error" role="alert">{error}</p>}
          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" size="sm" onClick={close}>Cancel</Button>
            <Button size="sm" variant="primary" disabled={submitting}>
              {submitting && <span className="loading loading-spinner loading-xs" />}
              Register node
            </Button>
          </div>
        </form>
      )}
    </Modal>
  )
}
