import { beforeEach, describe, expect, it, vi } from 'vitest'

import { get, getData, patchData } from './http'
import { fetchAdminOrder, fetchAdminOrders, updateAdminOrderStatus } from './adminOrders'

vi.mock('./http', () => ({
  get: vi.fn(),
  getData: vi.fn(),
  patchData: vi.fn(),
}))

const mockedGet = vi.mocked(get)
const mockedGetData = vi.mocked(getData)
const mockedPatchData = vi.mocked(patchData)

describe('adminOrders service', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('carrega a fila sem query quando nenhum filtro é informado', async () => {
    const page = {
      data: [],
      meta: { page: 1, pageSize: 20, total: 0, totalPages: 0 },
    }
    mockedGet.mockResolvedValue(page)

    await expect(fetchAdminOrders()).resolves.toEqual(page)
    expect(mockedGet).toHaveBeenCalledWith('/api/v1/admin/orders')
  })

  it('serializa apenas os filtros preenchidos', async () => {
    mockedGet.mockResolvedValue({ data: [], meta: {} })

    await fetchAdminOrders({
      page: 2,
      pageSize: 10,
      status: 'pending',
      fulfillment: 'delivery',
      q: '8708',
      sort: 'number',
      order: 'asc',
      from: '2026-11-22T00:00:00Z',
    })

    expect(mockedGet).toHaveBeenCalledWith(
      '/api/v1/admin/orders?page=2&pageSize=10&status=pending&fulfillment=delivery&from=2026-11-22T00%3A00%3A00Z&q=8708&sort=number&order=asc',
    )
  })

  it('carrega o detalhe codificando o identificador', async () => {
    mockedGetData.mockResolvedValue({ id: 'abc/123' })

    await fetchAdminOrder('abc/123')

    expect(mockedGetData).toHaveBeenCalledWith('/api/v1/admin/orders/abc%2F123')
  })

  it('envia a transição de status sem dados do cliente', async () => {
    mockedPatchData.mockResolvedValue({ id: 'abc', status: 'cancelled' })
    const request = { status: 'cancelled', reason: 'Cliente desistiu' }

    await expect(updateAdminOrderStatus('abc', request)).resolves.toEqual({
      id: 'abc',
      status: 'cancelled',
    })
    expect(mockedPatchData).toHaveBeenCalledWith('/api/v1/admin/orders/abc/status', request)
  })
})
