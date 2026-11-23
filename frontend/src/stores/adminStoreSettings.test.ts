import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchAdminStore, updateAdminStore } from '../services/adminStore'
import { useAdminStoreSettingsStore } from './adminStoreSettings'
import { ApiError } from '../services/http'

vi.mock('../services/adminStore', () => ({
  fetchAdminStore: vi.fn(),
  updateAdminStore: vi.fn(),
}))

const mockedFetch = vi.mocked(fetchAdminStore)
const mockedUpdate = vi.mocked(updateAdminStore)

const settings = {
  name: 'Brocados Lanches',
  isOpen: true,
  openingHours: [],
  deliveryTime: '35-50 min',
  whatsapp: '+5585999999999',
  instagram: '@brocados',
  address: 'Rua de Teste, 10',
  createdAt: '2026-11-23T00:00:00Z',
  updatedAt: '2026-11-23T00:00:00Z',
}

describe('adminStoreSettings store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    mockedFetch.mockResolvedValue(settings)
  })

  it('carrega as configurações da loja', async () => {
    const store = useAdminStoreSettingsStore()

    await store.load()

    expect(store.state).toBe('success')
    expect(store.settings).toEqual(settings)
  })

  it('salva alterações e atualiza o estado local', async () => {
    mockedUpdate.mockResolvedValue({ ...settings, isOpen: false })
    const store = useAdminStoreSettingsStore()
    await store.load()

    await expect(store.save({ ...settings, isOpen: false })).resolves.toBe(true)

    expect(store.settings?.isOpen).toBe(false)
    expect(store.saveSuccess).toContain('salvas')
  })

  it('reporta erro sem expor detalhes internos', async () => {
    mockedUpdate.mockRejectedValue(new ApiError(403, 'FORBIDDEN', 'x'))
    const store = useAdminStoreSettingsStore()

    await expect(store.save({ ...settings })).resolves.toBe(false)
    expect(store.saveError).toContain('permissão')
  })
})
