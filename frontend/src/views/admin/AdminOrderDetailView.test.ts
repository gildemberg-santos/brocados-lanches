import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchAdminOrder, updateAdminOrderStatus } from '../../services/adminOrders'
import AdminOrderDetailView from './AdminOrderDetailView.vue'

vi.mock('../../services/adminOrders', () => ({
  fetchAdminOrder: vi.fn(),
  fetchAdminOrders: vi.fn(),
  updateAdminOrderStatus: vi.fn(),
}))

const mockedDetail = vi.mocked(fetchAdminOrder)
const mockedUpdate = vi.mocked(updateAdminOrderStatus)
const StubView = defineComponent({ render: () => h('div') })

const detail = {
  id: 'order-1',
  orderNumber: 1042,
  publicId: 'abc123',
  status: 'pending',
  fulfillment: 'delivery' as const,
  customer: { name: 'Maria Silva', phone: '+5585999998708' },
  address: {
    street: 'Rua A',
    number: '12',
    complement: '',
    district: 'Centro',
    city: 'Aquiraz',
    state: 'CE',
    zipCode: '61700000',
  },
  zone: { id: 'zone-1', name: 'Centro', feeCents: 800 },
  items: [
    {
      productId: 'prod-1',
      name: 'Burguer',
      unitPriceCents: 2300,
      quantity: 2,
      options: [{ optionId: 'opt-1', name: 'Cheddar', kind: 'addon', priceCents: 400 }],
      note: '',
      lineTotalCents: 5400,
    },
  ],
  subtotalCents: 5400,
  deliveryFeeCents: 800,
  totalCents: 6200,
  payment: { method: 'cash' as const },
  note: '',
  statusHistory: [{ status: 'pending', at: '2026-11-22T02:30:00Z' }],
  confirmedAt: '',
  cancelledAt: '',
  cancelReason: '',
  createdAt: '2026-11-22T02:30:00Z',
  updatedAt: '2026-11-22T02:30:00Z',
}

describe('AdminOrderDetailView', () => {
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
        { path: '/pedido/:publicId', name: 'order', component: StubView },
      ],
    })
    await router.push('/admin/pedidos/order-1')
    await router.isReady()

    const wrapper = mount(AdminOrderDetailView, { global: { plugins: [router] } })
    await flushPromises()
    return wrapper
  }

  it('mostra snapshot, cliente, itens e histórico', async () => {
    mockedDetail.mockResolvedValue(detail)

    const wrapper = await mountView()

    expect(wrapper.text()).toContain('Pedido ')
    expect(wrapper.text()).toContain('1042')
    expect(wrapper.text()).toContain('Maria Silva')
    expect(wrapper.text()).toContain('+5585999998708')
    expect(wrapper.text()).toContain('2× Burguer')
    expect(wrapper.text()).toContain('Cheddar')
    expect(wrapper.text()).toContain('62,00')
    expect(wrapper.text()).toContain('Pendente')
  })

  it('envia a transição e confirma o resultado', async () => {
    mockedDetail.mockResolvedValue(detail)
    mockedUpdate.mockResolvedValue({ ...detail, status: 'confirmed' })

    const wrapper = await mountView()
    const button = wrapper.findAll('button').find((item) => item.text() === 'Confirmar pedido')
    if (!button) throw new Error('Botão de confirmação não encontrado')
    await button.trigger('click')
    await flushPromises()

    expect(mockedUpdate).toHaveBeenCalledWith('order-1', { status: 'confirmed' })
    expect(wrapper.get('[role="status"]').text()).toContain('Status do pedido atualizado')
    expect(wrapper.text()).toContain('Confirmado')
  })

  it('mostra erro amigável quando o detalhe não carrega', async () => {
    mockedDetail.mockRejectedValue(new Error('boom'))

    const wrapper = await mountView()

    expect(wrapper.text()).toContain('Não foi possível carregar os pedidos')
  })
})
