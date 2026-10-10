import { useState } from 'react'
import type { FormEvent } from 'react'
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

  return (
    <main className="grid min-h-screen place-items-center bg-base-200 px-4 py-12">
      <div className="w-full max-w-sm">
        <div className="mb-8 flex justify-center">
          <Brand />
        </div>

        <form className="rounded-box border border-base-300 bg-base-100 p-6 shadow-sm sm:p-8" onSubmit={submit}>
          <h1 className="text-lg font-semibold">Sign in</h1>
          <p className="mt-1 text-sm text-base-content/60">Use the administrator account configured on the server.</p>

          <div className="mt-6 space-y-4">
            <label className="block">
              <span className="mb-1.5 block text-sm font-medium">Email</span>
              <input
                className="input w-full"
                type="email"
                autoComplete="username"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                required
                autoFocus
              />
            </label>
            <label className="block">
              <span className="mb-1.5 block text-sm font-medium">Password</span>
              <input
                className="input w-full"
                type="password"
                autoComplete="current-password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                required
              />
            </label>
          </div>

          {error && (
            <p className="mt-4 rounded-field border border-error/25 bg-error/10 px-3 py-2 text-sm" role="alert">
              {error}
            </p>
          )}

          <Button variant="primary" className="mt-6 w-full" disabled={submitting}>
            {submitting && <span className="loading loading-spinner loading-xs" />}
            {submitting ? 'Signing in…' : 'Sign in'}
          </Button>
        </form>

        <p className="mt-6 text-center text-xs text-base-content/50">Self-hosted monitoring for Linux VPS fleets</p>
      </div>
    </main>
  )
}
