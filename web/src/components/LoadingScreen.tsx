import { Brand } from './Brand'

export function LoadingScreen() {
  return <main className="grid min-h-screen place-content-center justify-items-center gap-5"><Brand compact /><p className="font-mono text-[10px] tracking-[.2em] text-base-content/45">RESTORING SESSION</p><span className="loading loading-dots loading-md text-primary" /></main>
}
