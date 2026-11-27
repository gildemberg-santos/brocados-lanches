import { afterEach, describe, expect, it, vi } from 'vitest'

import { ApiError, get, getData, patch, patchData, post, postData, postEmpty } from './http'

interface MockResponse {
  ok: boolean
  status: number
  json: () => Promise<unknown>
}

function jsonResponse(body: unknown, status = 200): MockResponse {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  }
}

function invalidJsonResponse(status = 500): MockResponse {
  return {
    ok: false,
    status,
    json: async () => {
      throw new Error('invalid json')
    },
  }
}

describe('http', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    vi.useRealTimers()
  })

  it('returns JSON on success', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ status: 'ok' }))
    vi.stubGlobal('fetch', fetchMock)

    await expect(get('/readyz')).resolves.toEqual({ status: 'ok' })
    expect(fetchMock).toHaveBeenCalledWith('/readyz', {
      headers: {
        Accept: 'application/json',
      },
      credentials: 'include',
      signal: expect.any(AbortSignal),
    })
  })

  it('sends additional headers on GET without changing defaults', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ data: [] }))
    vi.stubGlobal('fetch', fetchMock)

    await get('/api/v1/orders/history', {
      headers: { Authorization: 'Bearer token-de-teste' },
    })

    expect(fetchMock).toHaveBeenCalledWith('/api/v1/orders/history', {
      headers: {
        Accept: 'application/json',
        Authorization: 'Bearer token-de-teste',
      },
      credentials: 'include',
      signal: expect.any(AbortSignal),
    })
  })

  it('unwraps the data envelope', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse({ data: { id: 'abc' } })))

    await expect(getData('/api/v1/example')).resolves.toEqual({ id: 'abc' })
  })

  it('throws ApiError with the API error payload', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        jsonResponse(
          {
            error: {
              code: 'STORE_CLOSED',
              message: 'Loja fechada.',
            },
          },
          503,
        ),
      ),
    )

    await expect(get('/api/v1/orders')).rejects.toMatchObject({
      name: 'ApiError',
      status: 503,
      code: 'STORE_CLOSED',
      message: 'Loja fechada.',
    })
  })

  it('preserves field details from the API error payload', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        jsonResponse(
          {
            error: {
              code: 'VALIDATION_ERROR',
              message: 'Dados inválidos.',
              details: [{ field: 'priceCents', code: 'OUT_OF_RANGE', message: 'Preço inválido.' }],
            },
          },
          422,
        ),
      ),
    )

    const error = await get('/api/v1/admin/products').catch((caught: unknown) => caught)

    expect(error).toMatchObject({
      code: 'VALIDATION_ERROR',
      details: [{ field: 'priceCents', code: 'OUT_OF_RANGE', message: 'Preço inválido.' }],
    })
  })

  it('throws a generic ApiError when the error body is not JSON', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(invalidJsonResponse()))

    const error = await get('/readyz').catch((caught: unknown) => caught)

    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({
      status: 500,
      code: 'HTTP_ERROR',
    })
  })

  it('posts JSON with the idempotency header and unwraps data', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ data: { id: 'abc' } }))
    vi.stubGlobal('fetch', fetchMock)

    await expect(
      postData('/api/v1/orders', { items: [] }, { idempotencyKey: 'chave-idempotente' }),
    ).resolves.toEqual({ id: 'abc' })
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/orders', {
      method: 'POST',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
        'Idempotency-Key': 'chave-idempotente',
      },
      credentials: 'include',
      body: JSON.stringify({ items: [] }),
      signal: expect.any(AbortSignal),
    })
  })

  it('maps POST errors to ApiError', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValue(
          jsonResponse({ error: { code: 'CONFLICT', message: 'Pedido duplicado.' } }, 409),
        ),
    )

    await expect(post('/api/v1/orders', {})).rejects.toMatchObject({
      name: 'ApiError',
      status: 409,
      code: 'CONFLICT',
    })
  })

  it('posts without a body and accepts 204', async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 204 })
    vi.stubGlobal('fetch', fetchMock)

    await expect(postEmpty('/api/v1/admin/auth/logout')).resolves.toBeUndefined()
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/admin/auth/logout', {
      method: 'POST',
      headers: {
        Accept: 'application/json',
      },
      credentials: 'include',
      signal: expect.any(AbortSignal),
    })
  })

  it('maps empty POST errors to ApiError', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValue(jsonResponse({ error: { code: 'FORBIDDEN', message: 'Negado.' } }, 403)),
    )

    await expect(postEmpty('/api/v1/admin/auth/logout')).rejects.toMatchObject({
      status: 403,
      code: 'FORBIDDEN',
    })
  })

  it('patches JSON with credentials and unwraps data', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ data: { id: 'abc' } }))
    vi.stubGlobal('fetch', fetchMock)

    await expect(
      patchData('/api/v1/admin/orders/abc/status', { status: 'confirmed' }),
    ).resolves.toEqual({ id: 'abc' })
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/admin/orders/abc/status', {
      method: 'PATCH',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({ status: 'confirmed' }),
      signal: expect.any(AbortSignal),
    })
  })

  it('maps PATCH errors to ApiError', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValue(
          jsonResponse({ error: { code: 'CONFLICT', message: 'Transição inválida.' } }, 409),
        ),
    )

    await expect(patch('/api/v1/admin/orders/abc/status', {})).rejects.toMatchObject({
      name: 'ApiError',
      status: 409,
      code: 'CONFLICT',
    })
  })

  it('aborta a requisição por timeout e devolve uma mensagem amigável', async () => {
    vi.useFakeTimers()
    const fetchMock = vi.fn(
      (_url: string, init: RequestInit) =>
        new Promise<Response>((_resolve, reject) => {
          init.signal?.addEventListener('abort', () => {
            reject(new DOMException('Aborted', 'AbortError'))
          })
        }),
    )
    vi.stubGlobal('fetch', fetchMock)

    const pending = get('/readyz')
    vi.advanceTimersByTime(10_000)

    await expect(pending).rejects.toMatchObject({
      name: 'ApiError',
      status: 0,
      code: 'TIMEOUT',
      message: 'A conexão demorou muito. Verifique sua conexão e tente novamente.',
    })
    expect(vi.getTimerCount()).toBe(0)
  })

  it('limpa o timer quando a requisição termina com sucesso', async () => {
    vi.useFakeTimers()
    const clearTimeoutSpy = vi.spyOn(globalThis, 'clearTimeout')
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse({ status: 'ok' })))

    await expect(get('/readyz')).resolves.toEqual({ status: 'ok' })

    expect(clearTimeoutSpy).toHaveBeenCalledTimes(1)
    expect(vi.getTimerCount()).toBe(0)
  })
})
