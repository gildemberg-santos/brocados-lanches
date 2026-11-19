import { afterEach, describe, expect, it, vi } from 'vitest'

import { ApiError, get, getData } from './http'

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
  })

  it('returns JSON on success', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ status: 'ok' }))
    vi.stubGlobal('fetch', fetchMock)

    await expect(get('/readyz')).resolves.toEqual({ status: 'ok' })
    expect(fetchMock).toHaveBeenCalledWith('/readyz', {
      headers: {
        Accept: 'application/json',
      },
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

  it('throws a generic ApiError when the error body is not JSON', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(invalidJsonResponse()))

    const error = await get('/readyz').catch((caught: unknown) => caught)

    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({
      status: 500,
      code: 'HTTP_ERROR',
    })
  })
})
