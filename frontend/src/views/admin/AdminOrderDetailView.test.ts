import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { playNotificationSound } from '../../features/orders/notificationSound'
import { fetchAdminOrder, updateAdminOrderStatus } from '../../services/adminOrders'
import { subscribeOrderEvents, type OrderEvent } from '../../services/realtime'
import AdminOrderDetailView from './AdminOrderDetailView.vue'

vi.mock('../../features/orders/notificationSound', () => ({
  playNotificationSound: vi.fn(() => true),
}))
vi.mock('../../services/adminOrders', () => ({
  fetchAdminOrder: vi.fn(),
  fetchAdminOrders: vi.fn(),
  updateAdminOrderStatus: vi.fn(),
}))
vi.mock('../../services/realtime', () => ({ subscribeOrderEvents: vi.fn(() => vi.fn()) }))

const mockedDetail = vi.mocked(fetchAdminOrder)
const mockedPlayNotificationSound = vi.mocked(playNotificationSound)
const mockedSubscribeOrderEvents = vi.mocked(subscribeOrderEvents)
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

function emitOrderEvent(event: OrderEvent): void {
  const handler = mockedSubscribeOrderEvents.mock.calls.at(-1)?.[1]
  if (!handler) throw new Error('canal de eventos não foi assinado')
  handler(event)
}

describe('AdminOrderDetailView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockedSubscribeOrderEvents.mockReturnValue(vi.fn())
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

  it('recarrega o detalhe apenas quando o evento é deste pedido', async () => {
    mockedDetail
      .mockResolvedValueOnce(detail)
      .mockResolvedValueOnce({ ...detail, status: 'preparing' })
    const wrapper = await mountView()

    expect(mockedSubscribeOrderEvents).toHaveBeenCalledWith(
      '/api/v1/admin/orders/events',
      expect.any(Function),
      expect.any(Object),
    )
    emitOrderEvent({ type: 'order.updated', orderId: 'outro-pedido', status: 'ready' })
    await flushPromises()
    expect(mockedDetail).toHaveBeenCalledTimes(1)

    emitOrderEvent({ type: 'order.updated', orderId: 'order-1', status: 'preparing' })
    await flushPromises()
    expect(mockedDetail).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('Em preparo')
  })

  it('avisa com toast e som somente quando o status deste pedido muda', async () => {
    mockedDetail.mockResolvedValue(detail)
    const wrapper = await mountView()

    emitOrderEvent({ type: 'connected' })
    emitOrderEvent({ type: 'order.updated', orderId: 'outro-pedido', status: 'ready' })
    expect(wrapper.find('.status-toast').exists()).toBe(false)
    expect(mockedPlayNotificationSound).not.toHaveBeenCalled()

    mockedDetail.mockResolvedValue({ ...detail, status: 'preparing' })
    emitOrderEvent({
      type: 'order.updated',
      orderId: 'order-1',
      orderNumber: 1042,
      status: 'preparing',
      updatedAt: '2026-11-27T12:00:00Z',
    })
    await flushPromises()

    expect(wrapper.get('.status-toast').text()).toContain(`Pedido #${1042}: Em preparo`)
    expect(mockedPlayNotificationSound).toHaveBeenCalledTimes(1)
  })

  it('mostra o resumo do perfil de conveniência quando o cliente é recorrente', async () => {
    mockedDetail.mockResolvedValue({
      ...detail,
      customerProfile: {
        name: 'Maria Silva',
        phone: '+5585999998708',
        orderCount: 3,
        lastOrderAt: '2026-11-20T12:00:00Z',
        fulfillment: 'delivery',
        zoneName: 'Centro',
        paymentMethod: 'pix',
        optedOut: false,
        retainedUntil: '2028-11-20T12:00:00Z',
      },
    })

    const wrapper = await mountView()

    expect(wrapper.get('#order-profile-title').text()).toBe('Cliente recorrente')
    expect(wrapper.text()).toContain('Pedidos com este telefone')
    expect(wrapper.text()).toContain('3')
    expect(wrapper.text()).toContain('Última zona')
    expect(wrapper.text()).toContain('Centro')
    expect(wrapper.text()).toContain('Perfil mantido até')
  })

  it('informa quando o perfil foi desativado', async () => {
    mockedDetail.mockResolvedValue({
      ...detail,
      customerProfile: {
        name: 'Maria Silva',
        phone: '+5585999998708',
        orderCount: 0,
        optedOut: true,
      },
    })

    const wrapper = await mountView()

    expect(wrapper.text()).toContain('Perfil desativado por solicitação do cliente')
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
    expect(wrapper.get('.admin-page__success').text()).toContain('Status do pedido atualizado')
    expect(wrapper.text()).toContain('Confirmado')
    expect(wrapper.get('.status-toast').text()).toContain('Confirmado')
    expect(mockedPlayNotificationSound).toHaveBeenCalledTimes(1)

    mockedDetail.mockResolvedValue({ ...detail, status: 'confirmed' })
    emitOrderEvent({
      type: 'order.updated',
      orderId: 'order-1',
      status: 'confirmed',
      updatedAt: '2026-11-27T12:10:00Z',
    })
    await flushPromises()
    expect(mockedPlayNotificationSound).toHaveBeenCalledTimes(1)
  })

  it('mostra erro amigável quando o detalhe não carrega', async () => {
    mockedDetail.mockRejectedValue(new Error('boom'))

    const wrapper = await mountView()

    expect(wrapper.text()).toContain('Não foi possível carregar os pedidos')
  })
})
