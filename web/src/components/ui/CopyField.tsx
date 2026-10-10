import { Check, Copy } from 'lucide-react'
import { useState } from 'react'

export function CopyField({ label, value, multiline = false }: { label: string; value: string; multiline?: boolean }) {
  const [status, setStatus] = useState<'idle' | 'copied' | 'failed'>('idle')

  async function copy() {
    try {
      // The Clipboard API exists only in secure contexts (HTTPS or localhost).
      await navigator.clipboard.writeText(value)
      setStatus('copied')
    } catch {
      setStatus('failed')
    }
  }

  return (
    <div>
      <div className="mb-1.5 flex items-center justify-between">
        <span className="text-xs font-medium text-base-content/70">{label}</span>
        <button type="button" className="btn btn-ghost btn-xs gap-1.5" onClick={() => void copy()}>
          {status === 'copied' ? <Check size={13} className="text-success" /> : <Copy size={13} />}
          {status === 'copied' ? 'Copied' : status === 'failed' ? 'Select to copy' : 'Copy'}
        </button>
      </div>
      <pre className={`overflow-x-auto rounded-field border border-base-300 bg-base-200 px-3 py-2.5 font-mono text-xs leading-5 select-all ${multiline ? '' : 'whitespace-pre-wrap break-all'}`}>
        {value}
      </pre>
    </div>
  )
}
