import { beforeEach, describe, expect, it, vi } from 'vitest'

import { getData, patchData } from './http'
import { fetchAdminStore, updateAdminStore } from './adminStore'

vi.mock('./http', () => ({
  getData: vi.fn(),
  patchData: vi.fn(),
}))

const mockedGetData = vi.mocked(getData)
const mockedPatchData = vi.mocked(patchData)

describe('adminStore service', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('carrega as configurações da loja', async () => {
    const settings = { name: 'Brocados', isOpen: true }
    mockedGetData.mockResolvedValue(settings)

    await expect(fetchAdminStore()).resolves.toBe(settings)
    expect(mockedGetData).toHaveBeenCalledWith('/api/v1/admin/store')
  })

  it('envia o PATCH completo com centavos e horários inalterados', async () => {
    mockedPatchData.mockResolvedValue({ name: 'Brocados', isOpen: false })
    const input = {
      name: 'Brocados',
      isOpen: false,
      openingHours: [],
      deliveryTime: '35-50 min',
      whatsapp: '+5585999999999',
      instagram: '@brocados',
      address: 'Rua de Teste, 10',
    }

    await expect(updateAdminStore(input)).resolves.toEqual({ name: 'Brocados', isOpen: false })
    expect(mockedPatchData).toHaveBeenCalledWith('/api/v1/admin/store', input)
  })
})
