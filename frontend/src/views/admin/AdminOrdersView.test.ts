import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { deleteAdminOrder, fetchAdminOrders } from '../../services/adminOrders'
import { ApiError } from '../../services/http'
import { subscribeOrderEvents, type OrderEvent } from '../../services/realtime'
import { useAdminAuthStore } from '../../stores/adminAuth'
import AdminOrdersView from './AdminOrdersView.vue'

vi.mock('../../services/adminOrders', () => ({
  deleteAdminOrder: vi.fn(),
  fetchAdminOrders: vi.fn(),
}))
vi.mock('../../services/realtime', () => ({ subscribeOrderEvents: vi.fn(() => vi.fn()) }))

const mockedList = vi.mocked(fetchAdminOrders)
const mockedDelete = vi.mocked(deleteAdminOrder)
const mockedSubscribeOrderEvents = vi.mocked(subscribeOrderEvents)
const StubView = defineComponent({ render: () => h('div') })

function setRole(role: 'owner' | 'operator'): void {
  const auth = useAdminAuthStore()
  auth.user = { id: 'u1', name: 'Dono', email: 'dono@example.com', role }
}

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

function emitOrderEvent(event: OrderEvent): void {
  const handler = mockedSubscribeOrderEvents.mock.calls.at(-1)?.[1]
  if (!handler) throw new Error('canal de eventos não foi assinado')
  handler(event)
}

describe('AdminOrdersView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockedSubscribeOrderEvents.mockReturnValue(vi.fn())
    mockedDelete.mockResolvedValue(undefined)
    setActivePinia(createPinia())
    setRole('owner')
  })

  async function mountView() {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/admin/pedidos', name: 'admin-orders', component: StubView },
        { path: '/admin/pedidos/:id', name: 'admin-order-detail', component: StubView },
      ],
    })
    router.push('/admin/pedidos')
    await router.isReady()

    const wrapper = mount(AdminOrdersView, {
      global: { plugins: [router] },
    })
    await flushPromises()
    return wrapper
  }

  it('mostra a fila mascarada com link para o detalhe', async () => {
    mockedList.mockResolvedValue({
      data: [listItem],
      meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
    })

    const wrapper = await mountView()

    expect(wrapper.text()).toContain('1042')
    expect(wrapper.text()).toContain('M*** S*** · ****8708')
    expect(wrapper.text()).toContain('Pendente')
    expect(wrapper.text()).toContain('R$ 62,00')
    expect(
      wrapper
        .findAll('a[href="/admin/pedidos/order-1"]')
        .some((link) => link.text() === 'Ver detalhes'),
    ).toBe(true)
  })

  it('assina o canal autenticado e recarrega a fila quando um pedido muda', async () => {
    mockedList.mockResolvedValue({
      data: [listItem],
      meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
    })
    await mountView()

    expect(mockedSubscribeOrderEvents).toHaveBeenCalledWith(
      '/api/v1/admin/orders/events',
      expect.any(Function),
      expect.any(Object),
    )
    emitOrderEvent({ type: 'connected' })
    await flushPromises()
    expect(mockedList).toHaveBeenCalledTimes(1)

    emitOrderEvent({ type: 'order.updated', orderId: 'order-1', status: 'ready' })
    await flushPromises()
    expect(mockedList).toHaveBeenCalledTimes(2)
  })

  it('exibe estado vazio quando não há pedidos', async () => {
    mockedList.mockResolvedValue({
      data: [],
      meta: { page: 1, pageSize: 20, total: 0, totalPages: 0 },
    })

    const wrapper = await mountView()

    expect(wrapper.text()).toContain('Nenhum pedido encontrado')
  })

  it('aplica filtros e permite tentar novamente após erro', async () => {
    mockedList.mockRejectedValueOnce(new Error('boom')).mockResolvedValueOnce({
      data: [],
      meta: { page: 1, pageSize: 20, total: 0, totalPages: 0 },
    })

    const wrapper = await mountView()

    expect(wrapper.text()).toContain('Não foi possível carregar os pedidos')
    await wrapper.get('[role="alert"] button').trigger('click')
    await flushPromises()
    expect(mockedList).toHaveBeenCalledTimes(2)

    await wrapper.find('#order-status').setValue('cancelled')
    await wrapper.find('form').trigger('submit')
    await flushPromises()
    expect(mockedList).toHaveBeenLastCalledWith(
      expect.objectContaining({ status: 'cancelled', page: 1 }),
    )
  })

  it('só oferece exclusão para owner e pede confirmação antes de excluir', async () => {
    mockedList.mockResolvedValue({
      data: [listItem],
      meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
    })
    const wrapper = await mountView()

    await wrapper.get('[aria-label="Excluir pedido 1042"]').trigger('click')
    expect(wrapper.text()).toContain('Excluir pedido 1042? Esta ação é irreversível.')
    expect(mockedDelete).not.toHaveBeenCalled()

    setRole('operator')
    await flushPromises()
    expect(wrapper.find('[aria-label="Excluir pedido 1042"]').exists()).toBe(false)
  })

  it('exclui o pedido após confirmação e remove da fila', async () => {
    mockedList.mockResolvedValue({
      data: [listItem],
      meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
    })
    const wrapper = await mountView()

    await wrapper.get('[aria-label="Excluir pedido 1042"]').trigger('click')
    await wrapper.get('[aria-label="Confirmar exclusão do pedido 1042"] button').trigger('click')
    await flushPromises()

    expect(mockedDelete).toHaveBeenCalledWith('order-1')
    expect(wrapper.text()).not.toContain('1042')
    expect(wrapper.text()).toContain('Pedido excluído.')
  })

  it('mantém a confirmação aberta quando a exclusão falha', async () => {
    mockedList.mockResolvedValue({
      data: [listItem],
      meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
    })
    mockedDelete.mockRejectedValueOnce(new ApiError(403, 'FORBIDDEN', 'raw'))
    const wrapper = await mountView()

    await wrapper.get('[aria-label="Excluir pedido 1042"]').trigger('click')
    await wrapper.get('[aria-label="Confirmar exclusão do pedido 1042"] button').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Você não tem permissão para esta ação.')
    expect(wrapper.find('[aria-label="Confirmar exclusão do pedido 1042"]').exists()).toBe(true)
  })

  it('recarrega a fila quando outro operador exclui um pedido', async () => {
    mockedList.mockResolvedValue({
      data: [listItem],
      meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
    })
    await mountView()

    emitOrderEvent({ type: 'order.deleted', orderId: 'order-1', orderNumber: 1042 })
    await flushPromises()

    expect(mockedList).toHaveBeenCalledTimes(2)
  })
})
