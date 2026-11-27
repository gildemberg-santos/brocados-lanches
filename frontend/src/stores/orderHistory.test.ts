import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { saveOrderAccessToken } from '../features/orders/orderAccessToken'
import { fetchOrderHistory } from '../services/orders'
import { ApiError } from '../services/http'
import { useOrderHistoryStore } from './orderHistory'

vi.mock('../services/orders', () => ({
  fetchOrderHistory: vi.fn(),
}))

const mockedFetch = vi.mocked(fetchOrderHistory)
const accessToken = 'a'.repeat(43)

const order = {
  orderNumber: 1042,
  publicId: 'public-123',
  status: 'pending',
  fulfillment: 'delivery' as const,
  items: [{ name: 'Bacon Cheddar', quantity: 2 }],
  subtotalCents: 4400,
  deliveryFeeCents: 800,
  totalCents: 5200,
  createdAt: '2026-11-22T02:30:00Z',
  updatedAt: '2026-11-22T02:30:00Z',
}

describe('orderHistory store', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
    setActivePinia(createPinia())
  })

  it('marca sucesso vazio quando não há token salvo', async () => {
    const store = useOrderHistoryStore()

    await store.load()

    expect(store.hasToken).toBe(false)
    expect(store.state).toBe('success')
    expect(store.isEmpty).toBe(true)
    expect(mockedFetch).not.toHaveBeenCalled()
  })

  it('carrega o histórico com o token salvo', async () => {
    saveOrderAccessToken(accessToken)
    mockedFetch.mockResolvedValue({
      data: [order],
      meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
    })
    const store = useOrderHistoryStore()

    await store.load()

    expect(store.hasToken).toBe(true)
    expect(store.state).toBe('success')
    expect(store.orders).toEqual([order])
    expect(store.isEmpty).toBe(false)
    expect(mockedFetch).toHaveBeenCalledWith(accessToken, { page: 1, pageSize: 20 })
  })

  it('navega apenas para páginas válidas', async () => {
    saveOrderAccessToken(accessToken)
    mockedFetch.mockResolvedValue({
      data: [],
      meta: { page: 2, pageSize: 20, total: 25, totalPages: 2 },
    })
    const store = useOrderHistoryStore()
    await store.load()

    await store.goToPage(2)

    expect(mockedFetch).toHaveBeenLastCalledWith(accessToken, { page: 2, pageSize: 20 })

    await store.goToPage(9)

    expect(mockedFetch).toHaveBeenCalledTimes(2)
    expect(store.meta.page).toBe(2)
  })

  it('mostra mensagem amigável em caso de erro', async () => {
    saveOrderAccessToken(accessToken)
    mockedFetch.mockRejectedValue(new ApiError(500, 'INTERNAL_ERROR', 'mongo: boom'))
    const store = useOrderHistoryStore()

    await store.load()

    expect(store.state).toBe('error')
    expect(store.error).toContain('Não foi possível carregar')
    expect(store.error).not.toContain('mongo')
    expect(store.orders).toEqual([])
  })
})
