import * as Dialog from '@radix-ui/react-dialog'
import { X } from 'lucide-react'
import type { ReactNode } from 'react'

type Props = {
  open: boolean
  title: string
  onClose: () => void
  children: ReactNode
  wide?: boolean
}

export function Modal({ open, title, onClose, children, wide = false }: Props) {
  return (
    <Dialog.Root open={open} onOpenChange={(next) => !next && onClose()}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-black/75 backdrop-blur-sm" />
        <Dialog.Content className={`fixed left-1/2 top-1/2 z-50 max-h-[90vh] w-[calc(100%-2rem)] -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-box border border-base-300 bg-base-100 p-6 shadow-2xl ${wide ? 'max-w-3xl' : 'max-w-lg'}`}>
          <header className="mb-6 flex items-start justify-between gap-4">
            <div>
              <p className="mb-1 font-mono text-[10px] tracking-[.2em] text-primary">VPSENTINEL</p>
              <Dialog.Title className="text-2xl font-semibold tracking-tight">{title}</Dialog.Title>
            </div>
            <Dialog.Close className="btn btn-square btn-ghost btn-sm" aria-label="Close dialog"><X size={18} /></Dialog.Close>
          </header>
          {children}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
