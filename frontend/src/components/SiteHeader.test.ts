import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchHealth } from '../services/health'
import { useCartStore } from '../stores/cart'
import { makeCatalogProduct } from '../test/fixtures/catalog'
import SiteHeader from './SiteHeader.vue'

vi.mock('../services/health', () => ({ fetchHealth: vi.fn() }))

const mockedFetchHealth = vi.mocked(fetchHealth)

const StubView = defineComponent({
  name: 'StubView',
  render: () => h('div'),
})

async function mountHeader(pinia = createPinia()) {
  setActivePinia(pinia)

  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'home', component: StubView },
      { path: '/cardapio', name: 'menu', component: StubView },
      { path: '/meus-pedidos', name: 'my-orders', component: StubView },
      { path: '/carrinho', name: 'cart', component: StubView },
    ],
  })

  await router.push('/')
  await router.isReady()

  return mount(SiteHeader, {
    global: {
      plugins: [pinia, router],
    },
  })
}

describe('SiteHeader', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
    mockedFetchHealth.mockResolvedValue({ status: 'ok', checks: { mongo: 'ok' } })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('links the brand to the home page', async () => {
    const wrapper = await mountHeader()

    expect(wrapper.get('a[aria-label="Brocados Lanches, início"]').attributes('href')).toBe('/')
    expect(wrapper.text()).toContain('BROCADOS')
  })

  it('offers navigation to home, menu, order history and cart', async () => {
    const wrapper = await mountHeader()
    const links = wrapper.findAll('.site-header__link:not(.site-header__link--cart)')

    expect(links.map((link) => link.text())).toEqual(['Início', 'Cardápio', 'Pedidos'])
    expect(links[1].attributes('href')).toBe('/cardapio')
    expect(links[2].attributes('href')).toBe('/meus-pedidos')
    expect(wrapper.get('.site-header__link--cart').attributes('href')).toBe('/carrinho')
  })

  it('shows the cart item count when there are items', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    useCartStore().add(makeCatalogProduct({ id: 'product-1' }), {}, [], '', 2)

    const wrapper = await mountHeader(pinia)

    expect(wrapper.get('.site-header__link--cart').text()).toContain('2')
    expect(wrapper.get('.site-header__link--cart').attributes('aria-label')).toBe(
      'Carrinho, 2 itens',
    )
  })

  it('não exibe aviso quando a API responde', async () => {
    const wrapper = await mountHeader()
    await flushPromises()

    expect(wrapper.find('[role="status"]').exists()).toBe(false)
  })

  it('mostra aviso acessível sem bloquear o carrinho quando a conexão cai', async () => {
    mockedFetchHealth.mockRejectedValue(new Error('falha de rede'))
    const pinia = createPinia()
    setActivePinia(pinia)
    useCartStore().add(makeCatalogProduct({ id: 'product-1' }), {}, [], '', 2)

    const wrapper = await mountHeader(pinia)
    await flushPromises()

    expect(wrapper.get('[role="status"]').text()).toBe(
      'Conexão perdida. Verifique sua conexão e tente novamente em instantes.',
    )
    const cartLink = wrapper.get('.site-header__link--cart')
    expect(cartLink.attributes('href')).toBe('/carrinho')
    expect(cartLink.attributes('aria-label')).toBe('Carrinho, 2 itens')
    expect(cartLink.attributes('aria-disabled')).toBeUndefined()
  })

  it('monitora a saúde da API e encerra o monitoramento ao desmontar', async () => {
    vi.useFakeTimers()

    const pinia = createPinia()
    setActivePinia(pinia)
    const wrapper = await mountHeader(pinia)

    expect(mockedFetchHealth).toHaveBeenCalledTimes(1)
    vi.advanceTimersByTime(30_000)
    expect(mockedFetchHealth).toHaveBeenCalledTimes(2)

    wrapper.unmount()
    vi.advanceTimersByTime(30_000)
    expect(mockedFetchHealth).toHaveBeenCalledTimes(2)
  })
})
