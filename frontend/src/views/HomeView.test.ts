import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchMenu } from '../services/menu'
import { makeCategory, makeMenu, makeProduct } from '../test/fixtures/catalog'
import HomeView from './HomeView.vue'

vi.mock('../services/menu', () => ({
  fetchMenu: vi.fn(),
}))

const mockedFetchMenu = vi.mocked(fetchMenu)

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
  })

  it('shows loading before rendering categories and featured products', async () => {
    mockedFetchMenu.mockResolvedValue(successMenu)

    const { wrapper } = await mountHome()

    expect(wrapper.text()).toContain('Carregando cardápio')

    await flushPromises()

    expect(wrapper.text()).toContain('Seu lanche, do seu jeito.')
    expect(wrapper.text()).toContain('Hambúrgueres')
    expect(wrapper.text()).toContain('Mais pedidos')
    expect(wrapper.text()).toContain('Brocado Cheddar')
    expect(wrapper.find('a[href="/produto/brocado-cheddar"]').exists()).toBe(true)
  })

  it('navigates to the menu with the selected category', async () => {
    mockedFetchMenu.mockResolvedValue(successMenu)

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
