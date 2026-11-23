import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchAdminOrders } from '../../services/adminOrders'
import AdminOrdersView from './AdminOrdersView.vue'

vi.mock('../../services/adminOrders', () => ({
  fetchAdminOrders: vi.fn(),
}))

const mockedList = vi.mocked(fetchAdminOrders)
const StubView = defineComponent({ render: () => h('div') })

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

describe('AdminOrdersView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
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
})
