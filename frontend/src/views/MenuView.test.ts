import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchMenu } from '../services/menu'
import { makeCategory, makeMenu, makeProduct } from '../test/fixtures/catalog'
import MenuView from './MenuView.vue'

vi.mock('../services/menu', () => ({
  fetchMenu: vi.fn(),
}))

const mockedFetchMenu = vi.mocked(fetchMenu)

const StubView = defineComponent({
  name: 'StubView',
  render: () => h('div'),
})

async function mountMenu(path = '/cardapio') {
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

  await router.push(path)
  await router.isReady()

  const wrapper = mount(MenuView, {
    global: {
      plugins: [pinia, router],
    },
  })

  return { wrapper, router }
}

const menu = makeMenu({
  categories: [
    makeCategory({
      id: 'category-1',
      name: 'Hambúrgueres',
      slug: 'hamburgueres',
      products: [
        makeProduct({
          id: 'product-1',
          name: 'Brocadão Cheddar',
          slug: 'brocadao-cheddar',
          description: 'Cheddar cremoso.',
          priceCents: 3000,
        }),
        makeProduct({
          id: 'product-sold-out',
          name: 'Brocado Esgotado',
          slug: 'brocado-esgotado',
          available: false,
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
          description: 'Lata gelada 350ml.',
          priceCents: 600,
          featured: false,
        }),
      ],
    }),
  ],
})

describe('MenuView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows loading, products, sold-out state and result count', async () => {
    mockedFetchMenu.mockResolvedValue(menu)

    const { wrapper } = await mountMenu()

    expect(wrapper.text()).toContain('Carregando cardápio')

    await flushPromises()

    expect(wrapper.text()).toContain('3 produtos encontrados')
    expect(wrapper.text()).toContain('Brocadão Cheddar')
    expect(wrapper.text()).toContain('Coca-Cola Lata')
    expect(wrapper.text()).toContain('Esgotado')
    expect(wrapper.find('a[href="/produto/brocadao-cheddar"]').exists()).toBe(true)
  })

  it('filters by search ignoring accents and updates to a singular count', async () => {
    mockedFetchMenu.mockResolvedValue(menu)

    const { wrapper } = await mountMenu()
    await flushPromises()

    await wrapper.get('input[type="search"]').setValue('brocadao')
    await flushPromises()

    expect(wrapper.text()).toContain('1 produto encontrado')
    expect(wrapper.text()).toContain('Brocadão Cheddar')
    expect(wrapper.text()).not.toContain('Coca-Cola Lata')
  })

  it('filters by category from the route query and clears the filter', async () => {
    mockedFetchMenu.mockResolvedValue(menu)

    const { wrapper, router } = await mountMenu('/cardapio?categoria=bebidas')
    await flushPromises()

    expect(wrapper.text()).toContain('Coca-Cola Lata')
    expect(wrapper.text()).not.toContain('Brocadão Cheddar')
    expect(wrapper.text()).toContain('1 produto encontrado')

    const allButton = wrapper.findAll('button').find((button) => button.text() === 'Todos')

    expect(allButton).toBeDefined()
    await allButton?.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.query.categoria).toBeUndefined()
    expect(wrapper.text()).toContain('3 produtos encontrados')
  })

  it('shows a search empty state and clears all filters', async () => {
    mockedFetchMenu.mockResolvedValue(menu)

    const { wrapper } = await mountMenu()
    await flushPromises()

    await wrapper.get('input[type="search"]').setValue('sem resultado')
    await flushPromises()

    expect(wrapper.text()).toContain('0 produtos encontrados')
    expect(wrapper.text()).toContain('Nenhum produto encontrado')

    const clearButton = wrapper
      .findAll('button')
      .find((button) => button.text() === 'Limpar filtros')

    expect(clearButton).toBeDefined()
    await clearButton?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('3 produtos encontrados')
  })

  it('shows an error state and retries', async () => {
    mockedFetchMenu.mockRejectedValue(new Error('network error'))

    const { wrapper } = await mountMenu()
    await flushPromises()

    expect(wrapper.text()).toContain('Não foi possível carregar o cardápio.')

    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(mockedFetchMenu).toHaveBeenCalledTimes(2)
  })

  it('shows an empty catalog state', async () => {
    mockedFetchMenu.mockResolvedValue(makeMenu({ categories: [] }))

    const { wrapper } = await mountMenu()
    await flushPromises()

    expect(wrapper.text()).toContain('Cardápio em preparação')
    expect(wrapper.text()).not.toContain('produtos encontrados')
  })
})
