import { useState } from 'react'
import type { ReactNode } from 'react'
import { Button } from './Button'
import { Modal } from './Modal'

type Props = {
  open: boolean
  title: string
  description: ReactNode
  confirmLabel: string
  onConfirm: () => Promise<void>
  onClose: () => void
}

export function ConfirmDialog({ open, title, description, confirmLabel, onConfirm, onClose }: Props) {
  const [pending, setPending] = useState(false)
  const [error, setError] = useState('')

  function close() {
    setError('')
    onClose()
  }

  async function confirm() {
    setPending(true)
    setError('')
    try {
      await onConfirm()
      close()
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : 'The request failed')
    } finally {
      setPending(false)
    }
  }

  return (
    <Modal
      open={open}
      title={title}
      onClose={close}
      footer={
        <>
          <Button size="sm" onClick={close} disabled={pending}>Cancel</Button>
          <Button size="sm" variant="danger" onClick={() => void confirm()} disabled={pending}>
            {pending && <span className="loading loading-spinner loading-xs" />}
            {confirmLabel}
          </Button>
        </>
      }
    >
      <div className="text-sm leading-6 text-base-content/75">{description}</div>
      {error && <p className="mt-3 text-sm text-error" role="alert">{error}</p>}
    </Modal>
  )
}
