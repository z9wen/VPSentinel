import { Check, Clipboard } from 'lucide-react'
import { useState } from 'react'
import { Button } from './Button'

export function OneTimeToken({ token, onClose }: { token: string; onClose: () => void }) {
  const [copied, setCopied] = useState(false)
  async function copy() {
    await navigator.clipboard.writeText(token)
    setCopied(true)
  }
  return <div className="rounded-box border border-primary/30 bg-primary/5 p-5"><p className="font-mono text-[10px] tracking-[.18em] text-primary">SHOWN ONCE</p><h3 className="mt-2 font-semibold">Store this agent token securely</h3><p className="mt-1 text-xs leading-5 text-base-content/60">The server retained only its hash. Losing it requires a new token.</p><code className="my-4 block break-all rounded-lg border border-base-300 bg-base-300/30 p-3 text-xs text-primary">{token}</code><div className="flex gap-2"><Button variant="primary" size="sm" onClick={() => void copy()}>{copied ? <Check size={15} /> : <Clipboard size={15} />}{copied ? 'Copied' : 'Copy token'}</Button><Button size="sm" onClick={onClose}>Done</Button></div></div>
}
