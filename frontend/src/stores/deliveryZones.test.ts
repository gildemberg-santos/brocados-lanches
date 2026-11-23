import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchDeliveryZones } from '../services/deliveryZones'
import { useDeliveryZonesStore } from './deliveryZones'

vi.mock('../services/deliveryZones', () => ({
  fetchDeliveryZones: vi.fn(),
}))

const mockedFetchDeliveryZones = vi.mocked(fetchDeliveryZones)

describe('delivery zones store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('loads zones once and exposes the cheapest active fee', async () => {
    mockedFetchDeliveryZones.mockResolvedValue([
      { id: 'zone-1', name: 'Centro', feeCents: 1000, sortOrder: 1 },
      { id: 'zone-2', name: 'Prainha', feeCents: 800, sortOrder: 2 },
    ])

    const store = useDeliveryZonesStore()
    await store.ensureLoaded()
    await store.ensureLoaded()

    expect(mockedFetchDeliveryZones).toHaveBeenCalledTimes(1)
    expect(store.state).toBe('success')
    expect(store.cheapestFeeCents).toBe(800)
  })

  it('loads nothing when there are no active zones', async () => {
    mockedFetchDeliveryZones.mockResolvedValue([])

    const store = useDeliveryZonesStore()
    await store.ensureLoaded()

    expect(store.zones).toEqual([])
    expect(store.cheapestFeeCents).toBeNull()
  })

  it('keeps the fee unknown when the request fails and retries on demand', async () => {
    mockedFetchDeliveryZones
      .mockRejectedValueOnce(new Error('network error'))
      .mockResolvedValueOnce([{ id: 'zone-1', name: 'Centro', feeCents: 800, sortOrder: 1 }])

    const store = useDeliveryZonesStore()
    await store.ensureLoaded()

    expect(store.state).toBe('error')
    expect(store.error).toBe('Não foi possível carregar as taxas de entrega. Tente novamente.')
    expect(store.cheapestFeeCents).toBeNull()

    await store.loadZones(true)
    expect(store.cheapestFeeCents).toBe(800)
  })
})
