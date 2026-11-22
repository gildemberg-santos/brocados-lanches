import { beforeEach, describe, expect, it, vi } from 'vitest'

import { getData } from './http'
import { fetchDeliveryZones } from './deliveryZones'

vi.mock('./http', () => ({
  getData: vi.fn(),
}))

const mockedGetData = vi.mocked(getData)

describe('delivery zones service', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads active delivery zones from the API', async () => {
    const zones = [
      { id: 'zone-1', name: 'Centro', feeCents: 800, sortOrder: 1 },
      { id: 'zone-2', name: 'Prainha', feeCents: 1200, sortOrder: 2 },
    ]
    mockedGetData.mockResolvedValue(zones)

    await expect(fetchDeliveryZones()).resolves.toEqual(zones)
    expect(mockedGetData).toHaveBeenCalledWith('/api/v1/delivery-zones')
  })
})
