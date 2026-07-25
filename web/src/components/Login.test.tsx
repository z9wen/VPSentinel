import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { Login } from './Login'

describe('Login', () => {
  afterEach(() => vi.unstubAllGlobals())

  it('submits credentials and returns the access token', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ access_token: 'access-token', expires_at: '2026-07-25T12:15:00Z' }), { status: 200, headers: { 'Content-Type': 'application/json' } }))
    vi.stubGlobal('fetch', fetchMock)
    const authenticated = vi.fn()
    render(<Login onAuthenticated={authenticated} />)

    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'admin@example.com' } })
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'correct-password' } })
    fireEvent.click(screen.getByRole('button', { name: 'Open control plane' }))

    await waitFor(() => expect(authenticated).toHaveBeenCalledWith('access-token'))
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/auth/login', expect.objectContaining({ method: 'POST', credentials: 'include' }))
  })
})
