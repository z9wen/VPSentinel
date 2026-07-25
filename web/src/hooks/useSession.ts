import { useCallback, useEffect, useState } from 'react'
import { APIError, request } from '../lib/api'
import type { AccessResponse, AuthedRequest, SessionState } from '../types'

export function useSession() {
  const [state, setState] = useState<SessionState>('loading')
  const [accessToken, setAccessToken] = useState('')

  const refresh = useCallback(async () => {
    try {
      const result = await request<AccessResponse>('/api/v1/auth/refresh', { method: 'POST' })
      setAccessToken(result.access_token)
      setState('authenticated')
      return result.access_token
    } catch {
      setAccessToken('')
      setState('anonymous')
      return ''
    }
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
