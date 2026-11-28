import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchAdminPix, updateAdminPix } from '../services/adminPix'
import { useAdminPixStore } from './adminPix'

vi.mock('../services/adminPix', () => ({ fetchAdminPix: vi.fn(), updateAdminPix: vi.fn() }))

const settings = {
  active: false,
  key: '',
  merchantName: '',
  merchantCity: '',
  createdAt: '2026-11-27T12:00:00Z',
  updatedAt: '2026-11-27T12:00:00Z',
}

describe('store de configuração do Pix', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.mocked(fetchAdminPix).mockReset()
    vi.mocked(updateAdminPix).mockReset()
  })

  it('carrega a configuração existente', async () => {
    vi.mocked(fetchAdminPix).mockResolvedValue(settings)
    const store = useAdminPixStore()

    await store.load()

    expect(store.state).toBe('success')
    expect(store.settings).toEqual(settings)
  })

  it('mostra erro de carregamento', async () => {
    vi.mocked(fetchAdminPix).mockRejectedValue(new Error('indisponível'))
    const store = useAdminPixStore()

    await store.load()

    expect(store.state).toBe('error')
    expect(store.error).not.toBe('')
  })

  it('salva e confirma a configuração', async () => {
    const input = {
      active: true,
      key: '123e4567-e89b-12d3-a456-426614174000',
      merchantName: 'Brocados Lanches',
      merchantCity: 'Aquiraz',
    }
    vi.mocked(updateAdminPix).mockResolvedValue({ ...settings, ...input })
    const store = useAdminPixStore()

    await expect(store.save(input)).resolves.toBe(true)

    expect(store.saveSuccess).not.toBe('')
    expect(store.settings?.key).toBe(input.key)
  })

  it('mostra erro de salvamento', async () => {
    vi.mocked(updateAdminPix).mockRejectedValue(new Error('indisponível'))
    const store = useAdminPixStore()

    await expect(store.save(settings)).resolves.toBe(false)

    expect(store.saveError).not.toBe('')
    expect(store.saving).toBe(false)
  })
})
