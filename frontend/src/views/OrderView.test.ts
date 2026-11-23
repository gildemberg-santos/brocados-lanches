import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchTrackedOrder } from '../services/orders'
import OrderView from './OrderView.vue'

vi.mock('../services/orders', () => ({ fetchTrackedOrder: vi.fn() }))

const mockedFetchTrackedOrder = vi.mocked(fetchTrackedOrder)
const StubView = defineComponent({ name: 'StubView', render: () => h('div') })
const tracked = {
  orderNumber: 1042,
  publicId: 'public-123',
  status: 'confirmed',
  fulfillment: 'delivery' as const,
  items: [{ name: 'Brocado Cheddar', quantity: 2 }],
  subtotalCents: 5400,
  deliveryFeeCents: 800,
  totalCents: 6200,
  createdAt: '2026-11-22T22:00:00Z',
  updatedAt: '2026-11-22T22:00:00Z',
  estimatedTime: '35-50 min',
}

async function mountOrder(state: Record<string, unknown> = {}) {
  window.history.replaceState(state, '', '/pedido/public-123')
  mockedFetchTrackedOrder.mockResolvedValue(tracked)

  const pinia = createPinia()
  setActivePinia(pinia)
  const router = createRouter({
    history: createWebHistory(),
    routes: [
      { path: '/cardapio', name: 'menu', component: StubView },
      { path: '/pedido/:publicId', name: 'order', component: StubView },
    ],
  })
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
  })

  it('shows the tracked order and the WhatsApp handoff link', async () => {
    const wrapper = await mountOrder({ whatsappUrl: 'https://wa.me/5585999999999?text=pedido' })

    expect(wrapper.text()).toContain('Pedido #')
    expect(wrapper.text()).toContain('1042')
    expect(wrapper.text()).toContain('Confirmado')
    expect(wrapper.text()).toContain('35-50 min')
    expect(wrapper.text()).toContain('2× Brocado Cheddar')
    expect(wrapper.text()).toContain('62,00')
    expect(wrapper.get('a[href^="https://wa.me/"]').text()).toContain('Abrir WhatsApp')
  })

  it('copies the handoff link', async () => {
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

  it('uses the stored handoff link after a reload', async () => {
    sessionStorage.setItem('brocados:handoff:public-123', 'https://wa.me/5585999999999?text=salvo')
    const wrapper = await mountOrder()

    expect(wrapper.get('a[href^="https://wa.me/"]').attributes('href')).toContain('text=salvo')
  })

  it('does not expose a handoff link when none is available', async () => {
    const wrapper = await mountOrder()

    expect(wrapper.find('a[href^="https://wa.me/"]').exists()).toBe(false)
  })

  it('shows a retry state when tracking fails', async () => {
    mockedFetchTrackedOrder.mockRejectedValue(new Error('falha'))

    const pinia = createPinia()
    setActivePinia(pinia)
    const router = createRouter({
      history: createWebHistory(),
      routes: [
        { path: '/cardapio', name: 'menu', component: StubView },
        { path: '/pedido/:publicId', name: 'order', component: StubView },
      ],
    })
    await router.replace({ name: 'order', params: { publicId: 'public-123' } })
    await router.isReady()
    const wrapper = mount(OrderView, { global: { plugins: [pinia, router] } })
    await flushPromises()

    expect(wrapper.text()).toContain('Não foi possível carregar')
    expect(wrapper.text()).toContain('Tentar novamente')

    mockedFetchTrackedOrder.mockResolvedValue(tracked)
    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Pedido #')
    expect(wrapper.text()).toContain('1042')
  })
})
