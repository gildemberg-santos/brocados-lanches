import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchMenu } from '../services/menu'
import { fetchStoreStatus } from '../services/store'
import { makeCategory, makeMenu, makeProduct } from '../test/fixtures/catalog'
import type { PublicStoreStatus } from '../types/store'
import HomeView from './HomeView.vue'

vi.mock('../services/menu', () => ({
  fetchMenu: vi.fn(),
}))
vi.mock('../services/store', () => ({
  fetchStoreStatus: vi.fn(),
}))

const mockedFetchMenu = vi.mocked(fetchMenu)
const mockedFetchStoreStatus = vi.mocked(fetchStoreStatus)

const StubView = defineComponent({
  name: 'StubView',
  render: () => h('div'),
})

async function mountHome() {
  const pinia = createPinia()
  setActivePinia(pinia)

  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'home', component: StubView },
      { path: '/cardapio', name: 'menu', component: StubView },
      { path: '/produto/:slug', name: 'product', component: StubView },
    ],
  })

  await router.push('/')
  await router.isReady()

  const wrapper = mount(HomeView, {
    global: {
      plugins: [pinia, router],
    },
  })

  return { wrapper, router }
}

const openStatus: PublicStoreStatus = {
  name: 'Brocados Lanches',
  status: 'open',
  isOpenManual: true,
  timezone: 'America/Fortaleza',
  deliveryTime: '30-90 min',
  whatsapp: '+5585999999999',
  instagram: '@brocados',
  address: 'Rua de Teste, 10',
  openingHours: [],
  nextOpeningAt: null,
  updatedAt: '2026-11-22T12:00:00Z',
}

const successMenu = makeMenu({
  categories: [
    makeCategory({
      id: 'category-1',
      name: 'Hambúrgueres',
      slug: 'hamburgueres',
      products: [
        makeProduct({
          id: 'product-1',
          name: 'Brocado Cheddar',
          slug: 'brocado-cheddar',
          priceCents: 2300,
          featured: true,
        }),
      ],
    }),
    makeCategory({
      id: 'category-2',
      name: 'Bebidas',
      slug: 'bebidas',
      products: [
        makeProduct({
          id: 'product-2',
          name: 'Coca-Cola Lata',
          slug: 'coca-cola-lata',
          priceCents: 600,
          featured: false,
        }),
      ],
    }),
  ],
})

describe('HomeView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockedFetchMenu.mockResolvedValue(successMenu)
    mockedFetchStoreStatus.mockResolvedValue(openStatus)
  })

  it('shows loading before rendering categories and featured products', async () => {
    const { wrapper } = await mountHome()

    expect(wrapper.text()).toContain('Carregando cardápio')
    expect(wrapper.text()).toContain('Consultando status da loja...')

    await flushPromises()

    expect(wrapper.text()).toContain('Seu lanche, do seu jeito.')
    expect(wrapper.text()).toContain('Hambúrgueres')
    expect(wrapper.text()).toContain('Mais pedidos')
    expect(wrapper.text()).toContain('Brocado Cheddar')
    expect(wrapper.find('a[href="/produto/brocado-cheddar"]').exists()).toBe(true)
  })

  it('shows the effective store status, delivery time and contacts from the API', async () => {
    const { wrapper } = await mountHome()
    await flushPromises()

    const banner = wrapper.get('[data-testid="store-status-banner"]')

    expect(banner.text()).toContain('Loja aberta')
    expect(banner.text()).toContain('Entrega: 30-90 min')
    expect(banner.get('a').attributes('href')).toBe('https://wa.me/5585999999999')
    expect(banner.text()).toContain('+5585999999999')
    expect(banner.text()).toContain('@brocados')
    expect(banner.text()).toContain('Rua de Teste, 10')
    expect(banner.text()).not.toContain('Próxima abertura')
  })

  it('formats the next opening in America/Fortaleza when the store is closed', async () => {
    mockedFetchStoreStatus.mockResolvedValue({
      ...openStatus,
      status: 'closed',
      nextOpeningAt: '2026-11-23T02:00:00Z',
    })

    const { wrapper } = await mountHome()
    await flushPromises()

    const banner = wrapper.get('[data-testid="store-status-banner"]')
    const formatted = new Intl.DateTimeFormat('pt-BR', {
      dateStyle: 'short',
      timeStyle: 'short',
      timeZone: 'America/Fortaleza',
    }).format(new Date('2026-11-23T02:00:00Z'))

    expect(banner.text()).toContain('Loja fechada')
    expect(banner.text()).toContain(`Próxima abertura: ${formatted}`)
    expect(banner.text()).not.toContain('2026-11-23T02:00:00Z')
  })

  it('shows a neutral message with retry when the status fails without breaking the menu', async () => {
    mockedFetchStoreStatus.mockRejectedValueOnce(new Error('mongo down'))

    const { wrapper } = await mountHome()
    await flushPromises()

    let banner = wrapper.get('[data-testid="store-status-banner"]')
    expect(banner.text()).toContain('Status da loja indisponível.')
    expect(banner.text()).not.toContain('mongo')
    expect(wrapper.text()).toContain('Hambúrgueres')

    await banner.get('button').trigger('click')
    await flushPromises()

    expect(mockedFetchStoreStatus).toHaveBeenCalledTimes(2)
    banner = wrapper.get('[data-testid="store-status-banner"]')
    expect(banner.text()).toContain('Loja aberta')
  })

  it('shows a neutral state when the API returns no status', async () => {
    mockedFetchStoreStatus.mockResolvedValue(null as unknown as PublicStoreStatus)

    const { wrapper } = await mountHome()
    await flushPromises()

    const banner = wrapper.get('[data-testid="store-status-banner"]')
    expect(banner.text()).toContain('Status da loja indisponível.')
    expect(banner.text()).toContain('Tentar novamente')
    expect(banner.text()).not.toContain('Loja aberta')
  })

  it('omits empty contacts instead of inventing values', async () => {
    mockedFetchStoreStatus.mockResolvedValue({
      ...openStatus,
      deliveryTime: '',
      whatsapp: '',
      instagram: '',
      address: '',
      nextOpeningAt: null,
    })

    const { wrapper } = await mountHome()
    await flushPromises()

    const banner = wrapper.get('[data-testid="store-status-banner"]')
    expect(banner.text()).toContain('Loja aberta')
    expect(banner.text()).not.toContain('Entrega:')
    expect(banner.text()).not.toContain('WhatsApp')
    expect(banner.text()).not.toContain('Instagram')
    expect(banner.text()).not.toContain('Endereço')
    expect(banner.text()).not.toContain('Próxima abertura')
  })

  it('navigates to the menu with the selected category', async () => {
    const { wrapper, router } = await mountHome()
    await flushPromises()

    const categoryButton = wrapper
      .findAll('button')
      .find((button) => button.text() === 'Hambúrgueres')

    expect(categoryButton).toBeDefined()
    await categoryButton?.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.fullPath).toBe('/cardapio?categoria=hamburgueres')
  })

  it('shows the error state and retries', async () => {
    mockedFetchMenu.mockRejectedValue(new Error('network error'))

    const { wrapper } = await mountHome()
    await flushPromises()

    expect(wrapper.text()).toContain('Não foi possível carregar o cardápio.')

    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(mockedFetchMenu).toHaveBeenCalledTimes(2)
  })

  it('shows an empty catalog state', async () => {
    mockedFetchMenu.mockResolvedValue(makeMenu({ categories: [] }))

    const { wrapper } = await mountHome()
    await flushPromises()

    expect(wrapper.text()).toContain('Cardápio em preparação')
    expect(wrapper.text()).toContain('Tentar novamente')
  })
})
