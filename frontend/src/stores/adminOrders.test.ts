import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import {
  deleteAdminOrder,
  fetchAdminOrder,
  fetchAdminOrders,
  updateAdminOrderStatus,
} from '../services/adminOrders'
import { ApiError } from '../services/http'
import { useAdminOrdersStore } from './adminOrders'

vi.mock('../services/adminOrders', () => ({
  deleteAdminOrder: vi.fn(),
  fetchAdminOrder: vi.fn(),
  fetchAdminOrders: vi.fn(),
  updateAdminOrderStatus: vi.fn(),
}))

const mockedList = vi.mocked(fetchAdminOrders)
const mockedDetail = vi.mocked(fetchAdminOrder)
const mockedUpdate = vi.mocked(updateAdminOrderStatus)
const mockedDelete = vi.mocked(deleteAdminOrder)

const listItem = {
  id: 'order-1',
  orderNumber: 1042,
  publicId: 'abc123',
  status: 'pending',
  fulfillment: 'delivery' as const,
  customerName: 'M*** S***',
  phone: '****8708',
  itemCount: 2,
  subtotalCents: 5400,
  deliveryFeeCents: 800,
  totalCents: 6200,
  createdAt: '2026-11-22T02:30:00Z',
  updatedAt: '2026-11-22T02:30:00Z',
}

const detail = {
  ...listItem,
  customer: { name: 'Maria Silva', phone: '+5585999998708' },
  items: [],
  payment: { method: 'cash' as const },
  statusHistory: [{ status: 'pending', at: listItem.createdAt }],
}

describe('adminOrders store', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
  })

  it('carrega a fila e os metadados', async () => {
    mockedList.mockResolvedValue({
      data: [listItem],
      meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
    })
    const store = useAdminOrdersStore()

    await store.load()

    expect(store.state).toBe('success')
    expect(store.orders).toEqual([listItem])
    expect(store.isEmpty).toBe(false)
    expect(store.totalPages).toBe(1)
  })

  it('marca fila vazia e erro com mensagem amigável', async () => {
    const store = useAdminOrdersStore()
    mockedList.mockResolvedValueOnce({
      data: [],
      meta: { page: 1, pageSize: 20, total: 0, totalPages: 0 },
    })

    await store.load()
    expect(store.isEmpty).toBe(true)

    mockedList.mockRejectedValueOnce(new ApiError(500, 'INTERNAL_ERROR', 'mongo: boom'))
    await store.load()
    expect(store.state).toBe('error')
    expect(store.error).toContain('Não foi possível carregar')
    expect(store.error).not.toContain('mongo')
  })

  it('normaliza filtros e reinicia a página', async () => {
    mockedList.mockResolvedValue({
      data: [],
      meta: { page: 1, pageSize: 20, total: 0, totalPages: 0 },
    })
    const store = useAdminOrdersStore()

    await store.applyFilters({
      page: 3,
      status: 'pending',
      fulfillment: undefined,
      q: '   ',
      from: '',
    })

    expect(store.filters).toEqual({
      page: 1,
      pageSize: 20,
      sort: 'createdAt',
      order: 'desc',
      status: 'pending',
    })
    expect(mockedList).toHaveBeenCalledWith(store.filters)
  })

  it('não navega para página fora do intervalo', async () => {
    mockedList.mockResolvedValue({
      data: [],
      meta: { page: 1, pageSize: 20, total: 0, totalPages: 1 },
    })
    const store = useAdminOrdersStore()
    await store.load()

    await store.goToPage(9)

    expect(mockedList).toHaveBeenCalledTimes(1)
  })

  it('carrega o detalhe uma vez e permite forçar recarga', async () => {
    mockedDetail.mockResolvedValue(detail)
    const store = useAdminOrdersStore()

    await store.loadDetail('order-1')
    await store.loadDetail('order-1')
    expect(mockedDetail).toHaveBeenCalledTimes(1)

    await store.loadDetail('order-1', true)
    expect(mockedDetail).toHaveBeenCalledTimes(2)
    expect(store.detail).toEqual(detail)
  })

  it('atualiza o detalhe após transição e comunica conflito', async () => {
    mockedUpdate.mockResolvedValueOnce({ ...detail, status: 'confirmed' })
    const store = useAdminOrdersStore()
    store.detail = detail

    store.orders = [listItem]
    mockedList.mockResolvedValue({
      data: [listItem],
      meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
    })
    await store.load()

    await expect(store.changeStatus('order-1', { status: 'confirmed' })).resolves.toBe(true)
    expect(store.detail?.status).toBe('confirmed')
    expect(store.orders[0]?.status).toBe('confirmed')
    expect(store.actionSuccess).toContain('atualizado')

    mockedUpdate.mockRejectedValueOnce(new ApiError(409, 'CONFLICT', 'raw'))
    await expect(store.changeStatus('order-1', { status: 'preparing' })).resolves.toBe(false)
    expect(store.actionError).toContain('mudou de status')
    expect(store.actionLoading).toBe(false)
  })

  it('recarrega a fila após evento sem alterar o estado de carregamento', async () => {
    mockedList
      .mockResolvedValueOnce({
        data: [listItem],
        meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
      })
      .mockResolvedValueOnce({
        data: [{ ...listItem, status: 'ready', updatedAt: '2026-11-26T20:00:00Z' }],
        meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
      })
    const store = useAdminOrdersStore()
    await store.load()

    await store.refreshFromEvent()

    expect(mockedList).toHaveBeenCalledTimes(2)
    expect(store.state).toBe('success')
    expect(store.orders[0]?.status).toBe('ready')

    mockedList.mockRejectedValueOnce(new Error('rede indisponível'))
    await store.refreshFromEvent()
    expect(store.orders[0]?.status).toBe('ready')
  })

  it('atualiza o detalhe aberto após evento de outro operador', async () => {
    mockedDetail
      .mockResolvedValueOnce(detail)
      .mockResolvedValueOnce({ ...detail, status: 'preparing' })
    const store = useAdminOrdersStore()
    await store.loadDetail('order-1')

    await store.refreshDetailFromEvent('order-1')
    expect(store.detail?.status).toBe('preparing')

    await store.refreshDetailFromEvent('outro-pedido')
    expect(mockedDetail).toHaveBeenCalledTimes(2)
  })

  it('exclui o pedido somente após confirmação do servidor e ajusta a paginação', async () => {
    mockedList.mockResolvedValue({
      data: [listItem, { ...listItem, id: 'order-2', orderNumber: 1043 }],
      meta: { page: 1, pageSize: 20, total: 2, totalPages: 1 },
    })
    mockedDelete.mockResolvedValue(undefined)
    const store = useAdminOrdersStore()
    await store.load()

    await expect(store.deleteOrder('order-1')).resolves.toBe(true)

    expect(mockedDelete).toHaveBeenCalledWith('order-1')
    expect(store.orders.map((item) => item.id)).toEqual(['order-2'])
    expect(store.meta.total).toBe(1)
    expect(store.deleteSuccess).toContain('excluído')
    expect(store.deleteLoading).toBe(false)
  })

  it('mantém o pedido na fila quando a exclusão falha', async () => {
    mockedList.mockResolvedValue({
      data: [listItem],
      meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
    })
    mockedDelete.mockRejectedValueOnce(new ApiError(403, 'FORBIDDEN', 'raw'))
    const store = useAdminOrdersStore()
    await store.load()

    await expect(store.deleteOrder('order-1')).resolves.toBe(false)

    expect(store.orders).toEqual([listItem])
    expect(store.deleteError).toContain('permissão')
    expect(store.deleteLoading).toBe(false)
  })
})
