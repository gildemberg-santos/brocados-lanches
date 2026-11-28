import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchAdminPix, updateAdminPix } from './adminPix'
import { getData, patchData } from './http'

vi.mock('./http', () => ({ getData: vi.fn(), patchData: vi.fn() }))

describe('serviço administrativo do Pix', () => {
  beforeEach(() => {
    vi.mocked(getData).mockReset()
    vi.mocked(patchData).mockReset()
  })

  it('lê a configuração da integração', async () => {
    vi.mocked(getData).mockResolvedValue({ active: false })

    await expect(fetchAdminPix()).resolves.toEqual({ active: false })
    expect(getData).toHaveBeenCalledWith('/api/v1/admin/integrations/pix')
  })

  it('salva a configuração da integração', async () => {
    const input = {
      active: true,
      key: '123e4567-e89b-12d3-a456-426614174000',
      merchantName: 'Brocados Lanches',
      merchantCity: 'Aquiraz',
    }
    vi.mocked(patchData).mockResolvedValue(input)

    await expect(updateAdminPix(input)).resolves.toEqual(input)
    expect(patchData).toHaveBeenCalledWith('/api/v1/admin/integrations/pix', input)
  })
})
