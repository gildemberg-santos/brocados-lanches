import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchHealth } from '../services/health'
import { useHealthStore } from './health'

vi.mock('../services/health', () => ({
  fetchHealth: vi.fn(),
}))

const mockedFetchHealth = vi.mocked(fetchHealth)

describe('useHealthStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('marks the service online when the API and MongoDB are ready', async () => {
    mockedFetchHealth.mockResolvedValue({
      status: 'ok',
      checks: { mongo: 'ok' },
    })

    const store = useHealthStore()
    await store.check()

    expect(store.state).toBe('online')
    expect(store.mongo).toBe('ok')
    expect(store.error).toBe('')
    expect(store.isReady).toBe(true)
  })

  it('marks the service offline when readiness is degraded', async () => {
    mockedFetchHealth.mockResolvedValue({
      status: 'degraded',
      checks: { mongo: 'error' },
    })

    const store = useHealthStore()
    await store.check()

    expect(store.state).toBe('offline')
    expect(store.mongo).toBe('error')
    expect(store.error).toBe('Serviço indisponível.')
    expect(store.isReady).toBe(false)
  })

  it('marks the service offline when the request fails', async () => {
    mockedFetchHealth.mockRejectedValue(new Error('network error'))

    const store = useHealthStore()
    await store.check()

    expect(store.state).toBe('offline')
    expect(store.mongo).toBe('error')
    expect(store.error).toBe('Não foi possível falar com a API.')
    expect(store.isReady).toBe(false)
  })
})
