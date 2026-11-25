import { afterEach, describe, expect, it, vi } from 'vitest'

import { fetchAdminUser, loginAdmin, logoutAdmin } from './adminAuth'

function jsonResponse(body: unknown, status = 200): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  } as Response
}

const owner = {
  id: 'user-1',
  name: 'Dono',
  email: 'dono@example.com',
  role: 'owner',
}

describe('serviço de autenticação administrativa', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('lê /me com credenciais e devolve o usuário do envelope', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ data: { user: owner } }))
    vi.stubGlobal('fetch', fetchMock)

    await expect(fetchAdminUser()).resolves.toEqual(owner)
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/admin/auth/me', {
      headers: { Accept: 'application/json' },
      credentials: 'include',
      signal: expect.any(AbortSignal),
    })
  })

  it('envia apenas e-mail e senha no login', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ data: { user: owner } }))
    vi.stubGlobal('fetch', fetchMock)

    await expect(loginAdmin('dono@example.com', 'senha-forte-123')).resolves.toEqual(owner)
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/admin/auth/login', {
      method: 'POST',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({ email: 'dono@example.com', password: 'senha-forte-123' }),
      signal: expect.any(AbortSignal),
    })
  })

  it('faz logout sem corpo e aceita 204', async () => {
    const response = { ok: true, status: 204 } as Response
    const fetchMock = vi.fn().mockResolvedValue(response)
    vi.stubGlobal('fetch', fetchMock)

    await expect(logoutAdmin()).resolves.toBeUndefined()
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/admin/auth/logout', {
      method: 'POST',
      headers: { Accept: 'application/json' },
      credentials: 'include',
      signal: expect.any(AbortSignal),
    })
  })

  it('propaga erro da API no logout', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValue(
          jsonResponse({ error: { code: 'UNAUTHORIZED', message: 'Sem sessão.' } }, 401),
        ),
    )

    await expect(logoutAdmin()).rejects.toMatchObject({ status: 401, code: 'UNAUTHORIZED' })
  })
})
