import { beforeEach, describe, expect, it, vi } from 'vitest'

import { getData, postData } from './http'
import { createOrder, fetchOrderQuote, fetchTrackedOrder } from './orders'

vi.mock('./http', () => ({
  getData: vi.fn(),
  postData: vi.fn(),
}))

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

  it('loads tracking data by public ID', async () => {
    mockedGetData.mockResolvedValue({ publicId: 'public-123' })

    await fetchTrackedOrder('public/123')

    expect(mockedGetData).toHaveBeenCalledWith('/api/v1/orders/public%2F123')
  })
})
