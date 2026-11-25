import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { ApiError } from '../services/http'
import { fetchTrackedOrder } from '../services/orders'
import { fetchStoreStatus } from '../services/store'
import type { PublicStoreStatus } from '../types/store'
import type { TrackedOrder } from '../types/order'
import OrderView from './OrderView.vue'

vi.mock('../services/orders', () => ({ fetchTrackedOrder: vi.fn() }))
vi.mock('../services/store', () => ({ fetchStoreStatus: vi.fn() }))

const mockedFetchTrackedOrder = vi.mocked(fetchTrackedOrder)
const StubView = defineComponent({ name: 'StubView', render: () => h('div') })
const tracked: TrackedOrder = {
  orderNumber: 1042,
  publicId: 'public-123',
  status: 'confirmed',
  fulfillment: 'delivery',
  items: [{ name: 'Brocado Cheddar', quantity: 2 }],
  subtotalCents: 5400,
  deliveryFeeCents: 800,
  totalCents: 6200,
  createdAt: '2026-11-22T02:30:00Z',
  updatedAt: '2026-11-22T03:10:00Z',
  estimatedTime: '35-50 min',
}
const storeStatus: PublicStoreStatus = {
  name: 'Brocados Lanches',
  status: 'open',
  isOpenManual: true,
  timezone: 'America/Fortaleza',
  deliveryTime: '35-50 min',
  whatsapp: '+5585999999999',
  instagram: 'https://instagram.com/brocados',
  address: 'Rua das Flores, 100',
  openingHours: [],
  nextOpeningAt: null,
  updatedAt: '2026-11-22T22:00:00Z',
}

function createTestRouter() {
  return createRouter({
    history: createWebHistory(),
    routes: [
      { path: '/cardapio', name: 'menu', component: StubView },
      { path: '/pedido/:publicId', name: 'order', component: StubView },
    ],
  })
}

async function mountOrder(
  state: Record<string, unknown> = {},
  orderFixture: TrackedOrder = tracked,
) {
  window.history.replaceState(state, '', '/pedido/public-123')
  mockedFetchTrackedOrder.mockResolvedValue(orderFixture)

  const pinia = createPinia()
  setActivePinia(pinia)
  const router = createTestRouter()
  await router.replace({ name: 'order', params: { publicId: 'public-123' } })
  await router.isReady()

  const wrapper = mount(OrderView, { global: { plugins: [pinia, router] } })
  await flushPromises()
  return wrapper
}

describe('OrderView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    sessionStorage.clear()
    vi.mocked(fetchStoreStatus).mockResolvedValue(storeStatus)
  })

  it('mostra o pedido, a timeline, o registro e o handoff do WhatsApp', async () => {
    const wrapper = await mountOrder({ whatsappUrl: 'https://wa.me/5585999999999?text=pedido' })

    expect(wrapper.get('h1').text()).toBe('Pedido registrado')
    expect(wrapper.text()).toContain('Pedido #')
    expect(wrapper.text()).toContain('1042')
    expect(wrapper.text()).toContain('Confirmado')
    expect(wrapper.text()).toContain('35-50 min')
    expect(wrapper.text()).toContain('2× Brocado Cheddar')
    expect(wrapper.text()).toContain('62,00')
    expect(wrapper.text()).toContain('Pedido registrado no servidor em')
    expect(wrapper.text()).toContain('Andamento do pedido')
    expect(wrapper.get('[aria-current="step"]').text()).toContain('Confirmado')
    expect(wrapper.get('.handoff-link').text()).toContain('Abrir WhatsApp')
  })

  it('copia o link de handoff', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText },
    })
    const wrapper = await mountOrder({ whatsappUrl: 'https://wa.me/5585999999999?text=pedido' })
    const copyButton = wrapper.findAll('button').find((button) => button.text() === 'Copiar link')

    await copyButton!.trigger('click')
    await flushPromises()

    expect(writeText).toHaveBeenCalledWith('https://wa.me/5585999999999?text=pedido')
    expect(wrapper.text()).toContain('Link copiado')
  })

  it('usa o link guardado após recarregar a página', async () => {
    sessionStorage.setItem('brocados:handoff:public-123', 'https://wa.me/5585999999999?text=salvo')
    const wrapper = await mountOrder()

    expect(wrapper.get('.handoff-link').attributes('href')).toContain('text=salvo')
  })

  it('informa o registro e exibe o contato da loja sem link de handoff', async () => {
    const wrapper = await mountOrder()

    expect(wrapper.get('h1').text()).toBe('Acompanhe seu pedido')
    expect(wrapper.find('.handoff-link').exists()).toBe(false)
    expect(wrapper.text()).toContain('Pedido registrado no servidor em')
    expect(wrapper.text()).toContain('Fale com a loja')
    expect(wrapper.text()).toContain('Abrir WhatsApp da loja')
    expect(wrapper.text()).toContain('Rua das Flores, 100')
  })

  it('usa título de pedido finalizado quando o status é entregue', async () => {
    const wrapper = await mountOrder({}, { ...tracked, status: 'delivered' })

    expect(wrapper.get('h1').text()).toBe('Pedido finalizado')
    expect(wrapper.get('[aria-current="step"]').text()).toContain('Entregue')
  })

  it('mostra carregamento antes da resposta', async () => {
    mockedFetchTrackedOrder.mockImplementation(() => new Promise(() => {}))
    const pinia = createPinia()
    setActivePinia(pinia)
    const router = createTestRouter()
    await router.replace({ name: 'order', params: { publicId: 'public-123' } })
    await router.isReady()

    const wrapper = mount(OrderView, { global: { plugins: [pinia, router] } })
    await flushPromises()

    expect(wrapper.get('h1').text()).toBe('Carregando pedido')
  })

  it('mostra erro genérico e permite tentar novamente', async () => {
    mockedFetchTrackedOrder.mockRejectedValue(new Error('falha'))
    const pinia = createPinia()
    setActivePinia(pinia)
    const router = createTestRouter()
    await router.replace({ name: 'order', params: { publicId: 'public-123' } })
    await router.isReady()
    const wrapper = mount(OrderView, { global: { plugins: [pinia, router] } })
    await flushPromises()

    expect(wrapper.get('h1').text()).toBe('Acompanhamento indisponível')
    expect(wrapper.text()).toContain('Não foi possível carregar')
    expect(wrapper.text()).toContain('Tentar novamente')

    mockedFetchTrackedOrder.mockResolvedValue(tracked)
    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(wrapper.get('h1').text()).toBe('Acompanhe seu pedido')
    expect(wrapper.text()).toContain('Pedido #')
    expect(wrapper.text()).toContain('1042')
  })

  it('não revela dados do pedido em caso de 404', async () => {
    mockedFetchTrackedOrder.mockRejectedValue(
      new ApiError(404, 'NOT_FOUND', 'pedido de terceiros com telefone 85999999999'),
    )
    const pinia = createPinia()
    setActivePinia(pinia)
    const router = createTestRouter()
    await router.replace({ name: 'order', params: { publicId: 'public-123' } })
    await router.isReady()
    const wrapper = mount(OrderView, { global: { plugins: [pinia, router] } })
    await flushPromises()

    expect(wrapper.get('h1').text()).toBe('Acompanhamento indisponível')
    expect(wrapper.text()).toContain('Pedido não encontrado ou código de acompanhamento inválido.')
    expect(wrapper.text()).not.toContain('85999999999')
    expect(wrapper.text()).not.toContain('1042')
  })
})
