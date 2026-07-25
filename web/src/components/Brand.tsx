import { ShieldCheck } from 'lucide-react'

export function Brand({ compact = false }: { compact?: boolean }) {
  return <div className="flex items-center gap-3 font-semibold tracking-tight"><span className="grid size-10 place-items-center rounded-xl border border-primary/30 bg-primary/10 text-primary"><ShieldCheck size={20} /></span>{!compact && <span>VPSentinel<small className="block font-mono text-[9px] font-normal uppercase tracking-widest text-base-content/45">Control plane</small></span>}</div>
}
