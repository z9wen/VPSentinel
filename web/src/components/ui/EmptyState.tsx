import { CircleCheck, CircleDashed } from 'lucide-react'
import type { ReactNode } from 'react'

type Props = {
  title: string
  text: ReactNode
  good?: boolean
  action?: ReactNode
}

export function EmptyState({ title, text, good = false, action }: Props) {
  const Icon = good ? CircleCheck : CircleDashed
  return (
    <div className="flex min-h-40 flex-col items-center justify-center px-6 py-10 text-center">
      <Icon size={22} className={good ? 'text-success' : 'text-base-content/35'} />
      <p className="mt-3 text-sm font-medium">{title}</p>
      <p className="mt-1 max-w-sm text-sm text-base-content/55">{text}</p>
      {action && <div className="mt-4">{action}</div>}
    </div>
  )
}
