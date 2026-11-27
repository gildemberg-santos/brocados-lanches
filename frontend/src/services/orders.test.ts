import { beforeEach, describe, expect, it, vi } from 'vitest'

import { get, getData, postData } from './http'
import { createOrder, fetchOrderHistory, fetchOrderQuote, fetchTrackedOrder } from './orders'

vi.mock('./http', () => ({
  get: vi.fn(),
  getData: vi.fn(),
  postData: vi.fn(),
}))

const mockedGet = vi.mocked(get)
const mockedGetData = vi.mocked(getData)
const mockedPostData = vi.mocked(postData)

describe('orders service', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('quotes an order on the server', async () => {
    const quote = {
      subtotalCents: 5400,
      deliveryFeeCents: 800,
      totalCents: 6200,
      currency: 'BRL',
      items: [],
      warnings: [],
    }
    mockedPostData.mockResolvedValue(quote)
    const request = {
      items: [],
      fulfillment: 'delivery' as const,
      zoneId: 'zona-1',
    }

    await expect(fetchOrderQuote(request)).resolves.toEqual(quote)
    expect(mockedPostData).toHaveBeenCalledWith('/api/v1/orders/quote', request)
  })

  it('sends the idempotency key when creating an order', async () => {
    const created = {
      orderNumber: 1042,
      publicId: 'public-123',
      status: 'pending',
      fulfillment: 'delivery' as const,
      subtotalCents: 5400,
      deliveryFeeCents: 800,
      totalCents: 6200,
      createdAt: '2026-11-22T22:00:00Z',
      whatsappUrl: 'https://wa.me/5585999999999?text=pedido',
      accessToken: 'a'.repeat(43),
    }
    mockedPostData.mockResolvedValue(created)
    const request = {
      items: [],
      fulfillment: 'delivery' as const,
      zoneId: 'zona-1',
      customer: { name: 'Maria', phone: '+5585999999999' },
      payment: { method: 'cash' as const },
      note: '',
    }

    await expect(createOrder(request, 'chave-idempotente')).resolves.toEqual(created)
    expect(mockedPostData).toHaveBeenCalledWith('/api/v1/orders', request, {
      idempotencyKey: 'chave-idempotente',
    })
  })

  it('reenvia o token de histórico ao criar um pedido', async () => {
    const created = {
      orderNumber: 1042,
      publicId: 'public-123',
      status: 'pending',
      fulfillment: 'delivery' as const,
      subtotalCents: 5400,
      deliveryFeeCents: 800,
      totalCents: 6200,
      createdAt: '2026-11-22T22:00:00Z',
      whatsappUrl: 'https://wa.me/5585999999999?text=pedido',
      accessToken: 'b'.repeat(43),
    }
    mockedPostData.mockResolvedValue(created)
    const request = {
      items: [],
      fulfillment: 'delivery' as const,
      zoneId: 'zona-1',
      customer: { name: 'Maria', phone: '+5585999999999' },
      payment: { method: 'cash' as const },
      note: '',
    }
    const accessToken = 'a'.repeat(43)

    await expect(createOrder(request, 'chave-idempotente', accessToken)).resolves.toEqual(created)
    expect(mockedPostData).toHaveBeenCalledWith('/api/v1/orders', request, {
      idempotencyKey: 'chave-idempotente',
      headers: { Authorization: `Bearer ${accessToken}` },
    })
  })

  it('busca o histórico paginado com o token no cabeçalho', async () => {
    const page = {
      data: [],
      meta: { page: 2, pageSize: 10, total: 0, totalPages: 0 },
    }
    mockedGet.mockResolvedValue(page)
    const accessToken = 'a'.repeat(43)

    await expect(fetchOrderHistory(accessToken, { page: 2, pageSize: 10 })).resolves.toEqual(page)
    expect(mockedGet).toHaveBeenCalledWith('/api/v1/orders/history?page=2&pageSize=10', {
      headers: { Authorization: `Bearer ${accessToken}` },
    })
  })

  it('loads tracking data by public ID', async () => {
    mockedGetData.mockResolvedValue({ publicId: 'public-123' })

    await fetchTrackedOrder('public/123')

    expect(mockedGetData).toHaveBeenCalledWith('/api/v1/orders/public%2F123')
  })
})
