import { useCallback, useEffect, useRef, useState } from 'react'
import { APIError, request } from '../lib/api'
import type { AccessResponse, AuthedRequest, SessionState } from '../types'

export function useSession() {
  const [state, setState] = useState<SessionState>('loading')
  const [accessToken, setAccessToken] = useState('')
  const pendingRefresh = useRef<Promise<string> | null>(null)

  // Refresh tokens rotate on every use and the server rejects a reused one, so
  // concurrent callers (parallel 401s, StrictMode effects) must share one request.
  const refresh = useCallback(() => {
    pendingRefresh.current ??= (async () => {
      try {
        const result = await request<AccessResponse>('/api/v1/auth/refresh', { method: 'POST' })
        setAccessToken(result.access_token)
        setState('authenticated')
        return result.access_token
      } catch {
        setAccessToken('')
        setState('anonymous')
        return ''
      } finally {
        pendingRefresh.current = null
      }
    })()
    return pendingRefresh.current
  }, [])

  useEffect(() => { void refresh() }, [refresh])

  const authedRequest: AuthedRequest = useCallback(async <T,>(path: string, init: RequestInit = {}) => {
    try {
      return await request<T>(path, init, accessToken)
    } catch (caught) {
      if (caught instanceof APIError && caught.status === 401) {
        const nextToken = await refresh()
        if (nextToken) return request<T>(path, init, nextToken)
      }
      throw caught
    }
  }, [accessToken, refresh])

  async function logout() {
    await request('/api/v1/auth/logout', { method: 'POST' }).catch(() => undefined)
    setAccessToken('')
    setState('anonymous')
  }

  return {
    state,
    authedRequest,
    authenticate: (token: string) => { setAccessToken(token); setState('authenticated') },
    logout,
  }
}
