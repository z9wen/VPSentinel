import type { ReactNode } from 'react'

export function Panel({ eyebrow, title, action, children }: { eyebrow: string; title: string; action?: ReactNode; children: ReactNode }) {
  return <section className="overflow-hidden rounded-box border border-base-300 bg-base-100"><header className="flex min-h-20 items-center justify-between gap-4 border-b border-base-300 px-5 py-4"><div><p className="mb-1 font-mono text-[10px] tracking-[.18em] text-base-content/45">{eyebrow}</p><h2 className="font-semibold">{title}</h2></div>{action}</header>{children}</section>
}
