import { useState } from 'react'
import type { FormEvent } from 'react'
import { Activity, LockKeyhole, RadioTower } from 'lucide-react'
import { request } from '../lib/api'
import type { AccessResponse } from '../types'
import { Brand } from './Brand'
import { Button } from './ui/Button'

export function Login({ onAuthenticated }: { onAuthenticated: (token: string) => void }) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  async function submit(event: FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    setError('')
    try {
      const result = await request<AccessResponse>('/api/v1/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) })
      onAuthenticated(result.access_token)
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : 'Login failed')
    } finally {
      setSubmitting(false)
    }
  }

  return <main className="grid min-h-screen lg:grid-cols-[1.25fr_1fr]">
    <section className="hidden flex-col justify-between border-r border-base-300 bg-gradient-to-br from-base-200 via-base-100 to-base-300 p-12 lg:flex xl:p-20">
      <Brand />
      <div className="max-w-4xl"><p className="mb-4 font-mono text-xs tracking-[.2em] text-primary">SELF-HOSTED FLEET OBSERVABILITY</p><h1 className="text-6xl font-semibold leading-[.95] tracking-[-.06em] xl:text-8xl">Know when your infrastructure changes state.</h1><p className="mt-7 max-w-2xl text-lg leading-8 text-base-content/60">Live Linux telemetry, actionable incidents, and recovery signals—without opening an inbound agent port.</p></div>
      <div className="flex gap-7 text-xs text-base-content/50"><span className="flex items-center gap-2"><RadioTower size={14} className="text-primary" />HTTPS push agents</span><span className="flex items-center gap-2"><LockKeyhole size={14} className="text-primary" />Hashed node tokens</span><span className="flex items-center gap-2"><Activity size={14} className="text-primary" />30-day history</span></div>
    </section>
    <section className="grid place-items-center p-6"><form className="w-full max-w-sm space-y-5" onSubmit={submit}><div className="mb-10 lg:hidden"><Brand /></div><div><p className="font-mono text-[10px] tracking-[.2em] text-primary">OPERATOR ACCESS</p><h2 className="mt-2 text-3xl font-semibold tracking-tight">Sign in</h2><p className="mt-2 text-sm leading-6 text-base-content/55">Use the administrator credentials stored in the server environment.</p></div><label className="fieldset"><span className="fieldset-legend">Email</span><input className="input w-full" type="email" autoComplete="username" value={email} onChange={(event) => setEmail(event.target.value)} required /></label><label className="fieldset"><span className="fieldset-legend">Password</span><input className="input w-full" type="password" autoComplete="current-password" value={password} onChange={(event) => setPassword(event.target.value)} required /></label>{error && <div className="alert alert-error text-sm" role="alert">{error}</div>}<Button variant="primary" className="w-full" disabled={submitting}>{submitting ? <span className="loading loading-spinner loading-sm" /> : <LockKeyhole size={16} />}{submitting ? 'Authenticating…' : 'Open control plane'}</Button><p className="text-center font-mono text-[9px] text-base-content/40">Refresh sessions stay in an HttpOnly cookie.</p></form></section>
  </main>
}
