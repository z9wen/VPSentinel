import { CheckCircle2, CircleDashed } from 'lucide-react'

export function EmptyState({ title, text, good = false }: { title: string; text: string; good?: boolean }) {
  const Icon = good ? CheckCircle2 : CircleDashed
  return <div className="flex min-h-44 flex-col items-center justify-center p-6 text-center"><Icon className={good ? 'mb-3 text-success' : 'mb-3 text-base-content/40'} /><strong className="text-sm">{title}</strong><p className="mt-1 max-w-lg text-xs text-base-content/55">{text}</p></div>
}
