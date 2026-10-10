import type { ReactNode } from 'react'

type Props = {
  title: string
  description?: ReactNode
  action?: ReactNode
  children: ReactNode
  className?: string
}

export function Panel({ title, description, action, children, className = '' }: Props) {
  return (
    <section className={`overflow-hidden rounded-box border border-base-300 bg-base-100 ${className}`}>
      <header className="flex flex-wrap items-center justify-between gap-3 border-b border-base-300 px-4 py-3 sm:px-5">
        <div className="min-w-0">
          <h2 className="text-sm font-semibold">{title}</h2>
          {description && <div className="mt-0.5 text-xs text-base-content/55">{description}</div>}
        </div>
        {action}
      </header>
      {children}
    </section>
  )
}
