import { describe, expect, it, vi } from 'vitest'

import { get } from './http'
import { fetchHealth } from './health'

vi.mock('./http', () => ({
  get: vi.fn(),
}))

describe('fetchHealth', () => {
  it('reads the readiness endpoint', async () => {
    vi.mocked(get).mockResolvedValue({
      status: 'ok',
      checks: { mongo: 'ok' },
    })

    await expect(fetchHealth()).resolves.toEqual({
      status: 'ok',
      checks: { mongo: 'ok' },
    })
    expect(get).toHaveBeenCalledWith('/readyz')
  })
})
