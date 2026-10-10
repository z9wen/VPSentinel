import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { useSession } from './useSession'

function json(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
}

describe('useSession', () => {
  afterEach(() => vi.unstubAllGlobals())

  it('shares one refresh request between concurrent callers', async () => {
    // The server rotates the refresh token on every use, so a second concurrent
    // refresh would present an already-rotated token and sign the operator out.
    let refreshes = 0
    let validToken: string | null = null
    vi.stubGlobal('fetch', vi.fn(async (path: string, init: RequestInit) => {
      if (path === '/api/v1/auth/refresh') {
        refreshes++
        validToken = `token-${refreshes}`
        return json({ access_token: validToken, expires_at: '2026-07-25T12:15:00Z' })
      }
      const authorization = new Headers(init.headers).get('Authorization')
      return authorization === `Bearer ${validToken}` ? json({ ok: true }) : json({ error: { message: 'expired' } }, 401)
    }))

    const { result } = renderHook(() => useSession(), { reactStrictMode: true })
    await waitFor(() => expect(result.current.state).toBe('authenticated'))
    expect(refreshes).toBe(1)

    validToken = null
    await act(async () => {
      await Promise.all([result.current.authedRequest('/api/v1/nodes'), result.current.authedRequest('/api/v1/alerts')])
    })
    expect(refreshes).toBe(2)
    expect(result.current.state).toBe('authenticated')
  })
})
