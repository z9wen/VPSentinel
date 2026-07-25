import { useState } from 'react'
import type { FormEvent } from 'react'
import type { AuthedRequest, Node } from '../../types'
import { Button } from '../../components/ui/Button'
import { Modal } from '../../components/ui/Modal'
import { OneTimeToken } from '../../components/ui/OneTimeToken'

type Props = { open: boolean; authedRequest: AuthedRequest; onClose: () => void; onCreated: () => Promise<void> }

export function AddNodeDialog({ open, authedRequest, onClose, onCreated }: Props) {
  const [name, setName] = useState('')
  const [hostname, setHostname] = useState('')
  const [environment, setEnvironment] = useState('production')
  const [error, setError] = useState('')
  const [token, setToken] = useState('')

  async function submit(event: FormEvent) {
    event.preventDefault()
    setError('')
    try {
      const node = await authedRequest<Node>('/api/v1/nodes', { method: 'POST', body: JSON.stringify({ name, hostname, environment }) })
      const result = await authedRequest<{ token: string }>(`/api/v1/nodes/${node.id}/tokens`, { method: 'POST', body: JSON.stringify({ label: 'initial-install' }) })
      setToken(result.token)
      await onCreated()
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : 'Could not register node')
    }
  }

  function close() {
    setToken('')
    onClose()
  }

  return <Modal open={open} title={token ? 'Save the agent token' : 'Register a node'} onClose={close}>{token ? <OneTimeToken token={token} onClose={close} /> : <form className="space-y-4" onSubmit={submit}><label className="fieldset"><span className="fieldset-legend">Display name</span><input className="input w-full" value={name} onChange={(event) => setName(event.target.value)} placeholder="Sydney API" required /></label><label className="fieldset"><span className="fieldset-legend">Expected hostname <span className="text-base-content/35">(optional)</span></span><input className="input w-full" value={hostname} onChange={(event) => setHostname(event.target.value)} placeholder="api-syd-01" /></label><label className="fieldset"><span className="fieldset-legend">Environment</span><select className="select w-full" value={environment} onChange={(event) => setEnvironment(event.target.value)}><option>production</option><option>staging</option><option>development</option></select></label>{error && <div className="alert alert-error text-sm">{error}</div>}<Button variant="primary" className="w-full">Register and create token</Button></form>}</Modal>
}
