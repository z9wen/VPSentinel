import * as Dialog from '@radix-ui/react-dialog'
import { X } from 'lucide-react'
import type { ReactNode } from 'react'

type Props = {
  open: boolean
  title: ReactNode
  description?: ReactNode
  onClose: () => void
  children: ReactNode
  footer?: ReactNode
  // A sheet slides in from the right and suits detail views; a dialog suits short forms.
  variant?: 'dialog' | 'sheet'
}

const layouts = {
  dialog: 'left-1/2 top-1/2 max-h-[90vh] w-[calc(100%-2rem)] max-w-md -translate-x-1/2 -translate-y-1/2 rounded-box border data-[state=open]:animate-dialog-in',
  sheet: 'inset-y-0 right-0 w-full max-w-xl border-l data-[state=open]:animate-sheet-in',
}

export function Modal({ open, title, description, onClose, children, footer, variant = 'dialog' }: Props) {
  return (
    <Dialog.Root open={open} onOpenChange={(next) => !next && onClose()}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-50 bg-black/50 data-[state=open]:animate-overlay-in motion-reduce:animate-none" />
        <Dialog.Content
          className={`fixed z-50 flex flex-col border-base-300 bg-base-100 shadow-2xl outline-none motion-reduce:animate-none ${layouts[variant]}`}
          // Radix links a rendered Description automatically; opt out explicitly otherwise.
          {...(description ? {} : { 'aria-describedby': undefined })}
        >
          <header className="flex items-start justify-between gap-4 border-b border-base-300 px-5 py-4">
            <div className="min-w-0">
              <Dialog.Title className="truncate text-base font-semibold">{title}</Dialog.Title>
              {description && <Dialog.Description asChild><div className="mt-1 text-sm text-base-content/60">{description}</div></Dialog.Description>}
            </div>
            <Dialog.Close className="btn btn-square btn-ghost btn-sm -mr-2 -mt-1" aria-label="Close">
              <X size={18} />
            </Dialog.Close>
          </header>
          <div className="min-h-0 flex-1 overflow-y-auto px-5 py-5">{children}</div>
          {footer && <footer className="flex justify-end gap-2 border-t border-base-300 px-5 py-3">{footer}</footer>}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
