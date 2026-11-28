import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchStoreStatus } from '../services/store'
import { useStoreStatusStore } from './storeStatus'
import type { PublicStoreStatus } from '../types/store'

vi.mock('../services/store', () => ({ fetchStoreStatus: vi.fn() }))

const status: PublicStoreStatus = {
  name: 'Brocados Lanches',
  status: 'open',
  isOpenManual: true,
  timezone: 'America/Fortaleza',
  deliveryTime: '30-90 min',
  whatsapp: '+5585999999999',
  instagram: '@brocados',
  address: 'Rua de Teste, 10',
  openingHours: [],
  pix: { active: false, key: '', merchantName: '', merchantCity: '' },
  nextOpeningAt: null,
  updatedAt: '2026-11-22T12:00:00Z',
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.clearAllMocks()
  vi.mocked(fetchStoreStatus).mockResolvedValue(status)
})

describe('useStoreStatusStore', () => {
  it('carrega o status efetivo do servidor', async () => {
    const store = useStoreStatusStore()

    await store.load()

    expect(store.state).toBe('success')
    expect(store.status?.status).toBe('open')
    expect(fetchStoreStatus).toHaveBeenCalledOnce()
  })

  it('não recarrega quando já carregou', async () => {
    const store = useStoreStatusStore()

    await store.load()
    await store.load()

    expect(fetchStoreStatus).toHaveBeenCalledOnce()
  })

  it('expõe erro em português sem dados brutos', async () => {
    vi.mocked(fetchStoreStatus).mockRejectedValue(new Error('mongo down'))
    const store = useStoreStatusStore()

    await store.load()

    expect(store.state).toBe('error')
    expect(store.error).toContain('Não foi possível')
    expect(store.error).not.toContain('mongo')
  })

  it('força recarga após salvar configurações', async () => {
    const store = useStoreStatusStore()
    await store.load()

    await store.refresh()

    expect(fetchStoreStatus).toHaveBeenCalledTimes(2)
  })
})
