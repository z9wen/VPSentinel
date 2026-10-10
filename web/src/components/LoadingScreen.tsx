import { BrandMark } from './Brand'

export function LoadingScreen() {
  return (
    <main className="grid min-h-screen place-content-center justify-items-center gap-4 bg-base-200">
      <BrandMark className="size-10" />
      <span className="loading loading-dots loading-sm text-base-content/40" aria-label="Restoring session" />
    </main>
  )
}
