import { beforeEach, describe, expect, it, vi } from 'vitest'

import { getData } from './http'
import { fetchStoreStatus } from './store'
import type { PublicStoreStatus } from '../types/store'

vi.mock('./http', () => ({ getData: vi.fn() }))

const status: PublicStoreStatus = {
  name: 'Brocados Lanches',
  status: 'closed',
  isOpenManual: false,
  timezone: 'America/Fortaleza',
  deliveryTime: '30-90 min',
  whatsapp: '+5585999999999',
  instagram: '',
  address: 'Rua de Teste, 10',
  openingHours: [],
  nextOpeningAt: null,
  updatedAt: '2026-11-22T12:00:00Z',
}

beforeEach(() => vi.clearAllMocks())

describe('fetchStoreStatus', () => {
  it('busca o status efetivo no endpoint público', async () => {
    vi.mocked(getData).mockResolvedValue(status)

    await expect(fetchStoreStatus()).resolves.toEqual(status)
    expect(getData).toHaveBeenCalledWith('/api/v1/store')
  })
})
