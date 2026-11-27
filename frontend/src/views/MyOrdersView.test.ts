import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { saveOrderAccessToken } from '../features/orders/orderAccessToken'
import { fetchOrderHistory } from '../services/orders'
import MyOrdersView from './MyOrdersView.vue'

vi.mock('../services/orders', () => ({ fetchOrderHistory: vi.fn() }))

const mockedFetch = vi.mocked(fetchOrderHistory)
const accessToken = 'a'.repeat(43)

const StubView = defineComponent({ name: 'StubView', render: () => h('div') })

async function mountView() {
  const pinia = createPinia()
  setActivePinia(pinia)
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/meus-pedidos', name: 'my-orders', component: MyOrdersView },
      { path: '/pedido/:publicId', name: 'order', component: StubView },
      { path: '/cardapio', name: 'menu', component: StubView },
      { path: '/privacidade', name: 'privacy', component: StubView },
    ],
  })
  await router.push('/meus-pedidos')
  await router.isReady()

  const wrapper = mount(MyOrdersView, {
    global: { plugins: [pinia, router] },
  })
  await flushPromises()
  return wrapper
}

describe('MyOrdersView', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('explica que não há histórico quando o navegador não tem acesso', async () => {
    const wrapper = await mountView()

    expect(wrapper.text()).toContain('Nenhum histórico neste navegador')
    expect(mockedFetch).not.toHaveBeenCalled()
  })

  it('lista os pedidos com status, total e link de acompanhamento', async () => {
    saveOrderAccessToken(accessToken)
    mockedFetch.mockResolvedValue({
      data: [
        {
          orderNumber: 1042,
          publicId: 'public-123',
          status: 'pending',
          fulfillment: 'delivery',
          items: [{ name: 'Bacon Cheddar', quantity: 2 }],
          subtotalCents: 4400,
          deliveryFeeCents: 800,
          totalCents: 5200,
          createdAt: '2026-11-22T02:30:00Z',
          updatedAt: '2026-11-22T02:30:00Z',
        },
      ],
      meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
    })

    const wrapper = await mountView()

    expect(wrapper.text()).toContain('1042')
    expect(wrapper.text()).toContain('2× Bacon Cheddar')
    expect(wrapper.text()).toContain('R$ 52,00')
    expect(wrapper.get('.history-card').attributes('href')).toBe('/pedido/public-123')
  })

  it('oferece nova tentativa quando a consulta falha', async () => {
    saveOrderAccessToken(accessToken)
    mockedFetch.mockRejectedValueOnce(new Error('falha de rede')).mockResolvedValueOnce({
      data: [],
      meta: { page: 1, pageSize: 20, total: 0, totalPages: 0 },
    })

    const wrapper = await mountView()

    expect(wrapper.text()).toContain('Não foi possível carregar')
    await wrapper.get('.error-state button').trigger('click')
    await flushPromises()

    expect(mockedFetch).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('Você ainda não fez pedidos')
  })

  it('navega entre páginas quando há mais de uma página', async () => {
    saveOrderAccessToken(accessToken)
    mockedFetch.mockImplementation(async (_token, query) => ({
      data: [
        {
          orderNumber: 1042,
          publicId: 'public-123',
          status: 'pending',
          fulfillment: 'delivery',
          items: [{ name: 'Bacon Cheddar', quantity: 2 }],
          subtotalCents: 4400,
          deliveryFeeCents: 800,
          totalCents: 5200,
          createdAt: '2026-11-22T02:30:00Z',
          updatedAt: '2026-11-22T02:30:00Z',
        },
      ],
      meta: {
        page: query?.page ?? 1,
        pageSize: 20,
        total: 21,
        totalPages: 2,
      },
    }))

    const wrapper = await mountView()
    await wrapper.get('.history-pagination__button:last-child').trigger('click')
    await flushPromises()

    expect(mockedFetch).toHaveBeenLastCalledWith(accessToken, { page: 2, pageSize: 20 })
    expect(wrapper.text()).toContain('Página 2 de 2')
  })
})
