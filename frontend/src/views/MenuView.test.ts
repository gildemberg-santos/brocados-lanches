import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchMenu } from '../services/menu'
import { makeCategory, makeMenu, makeProduct } from '../test/fixtures/catalog'
import MenuView from './MenuView.vue'

vi.mock('../services/menu', () => ({
  fetchMenu: vi.fn(),
}))

const mockedFetchMenu = vi.mocked(fetchMenu)
const scrollIntoView = vi.fn()
const scrollToTop = vi.fn()

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

  const host = document.createElement('div')
  document.body.appendChild(host)

  const wrapper = mount(MenuView, {
    attachTo: host,
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
    Object.defineProperty(Element.prototype, 'scrollIntoView', {
      configurable: true,
      value: scrollIntoView,
    })
    vi.spyOn(window, 'scrollTo').mockImplementation(scrollToTop)
    Object.defineProperty(window, 'requestAnimationFrame', {
      configurable: true,
      writable: true,
      value: vi.fn(() => 1),
    })
    Object.defineProperty(window, 'cancelAnimationFrame', {
      configurable: true,
      writable: true,
      value: vi.fn(),
    })
  })

  afterEach(() => {
    vi.restoreAllMocks()
    document.body.innerHTML = ''
  })

  it('mostra carregando, seções por categoria, esgotado e contagem', async () => {
    mockedFetchMenu.mockResolvedValue(menu)

    const { wrapper } = await mountMenu()

    expect(wrapper.text()).toContain('Carregando cardápio')

    await flushPromises()

    expect(wrapper.text()).toContain('3 produtos encontrados')
    expect(wrapper.text()).toContain('Hambúrgueres')
    expect(wrapper.text()).toContain('Bebidas')
    expect(wrapper.text()).toContain('Brocadão Cheddar')
    expect(wrapper.text()).toContain('Coca-Cola Lata')
    expect(wrapper.text()).toContain('Esgotado')
    expect(wrapper.find('#categoria-hamburgueres').exists()).toBe(true)
    expect(wrapper.find('#categoria-bebidas').exists()).toBe(true)
    expect(wrapper.find('a[href="/produto/brocadao-cheddar"]').exists()).toBe(true)
  })

  it('destaca a primeira categoria no carregamento, antes de qualquer scroll', async () => {
    mockedFetchMenu.mockResolvedValue(menu)

    const { wrapper } = await mountMenu()
    await flushPromises()

    const hamburgueresButton = wrapper
      .findAll('button')
      .find((button) => button.text() === 'Hambúrgueres')

    expect(hamburgueresButton?.attributes('aria-pressed')).toBe('true')
    expect(scrollToTop).not.toHaveBeenCalled()
  })

  it('filtra por busca ignorando acentos e atualiza para o singular', async () => {
    mockedFetchMenu.mockResolvedValue(menu)

    const { wrapper } = await mountMenu()
    await flushPromises()

    await wrapper.get('input[type="search"]').setValue('brocadao')
    await flushPromises()

    expect(wrapper.text()).toContain('1 produto encontrado')
    expect(wrapper.text()).toContain('Brocadão Cheddar')
    expect(wrapper.text()).not.toContain('Coca-Cola Lata')
    expect(wrapper.find('nav[aria-label="Categorias do cardápio"]').exists()).toBe(false)
  })

  it('abre a categoria da query como âncora sem filtrar as demais', async () => {
    mockedFetchMenu.mockResolvedValue(menu)

    const { wrapper } = await mountMenu('/cardapio?categoria=bebidas')
    await flushPromises()

    const bebidasButton = wrapper.findAll('button').find((button) => button.text() === 'Bebidas')

    expect(bebidasButton?.attributes('aria-pressed')).toBe('true')
    expect(scrollIntoView).toHaveBeenCalledWith({ behavior: 'smooth', block: 'start' })
    expect(wrapper.text()).toContain('Brocadão Cheddar')
    expect(wrapper.text()).toContain('Coca-Cola Lata')
    expect(wrapper.text()).toContain('3 produtos encontrados')
  })

  it('ao clicar em uma categoria, reposiciona a página e atualiza a query', async () => {
    mockedFetchMenu.mockResolvedValue(menu)

    const { wrapper, router } = await mountMenu()
    await flushPromises()

    const hamburgueresButton = wrapper
      .findAll('button')
      .find((button) => button.text() === 'Hambúrgueres')

    expect(hamburgueresButton).toBeDefined()
    await hamburgueresButton?.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.query.categoria).toBe('hamburgueres')
    expect(hamburgueresButton?.attributes('aria-pressed')).toBe('true')
    expect(scrollIntoView).toHaveBeenCalledWith({ behavior: 'smooth', block: 'start' })
  })

  it('mostra estado vazio da busca e limpa a busca', async () => {
    mockedFetchMenu.mockResolvedValue(menu)

    const { wrapper } = await mountMenu()
    await flushPromises()

    await wrapper.get('input[type="search"]').setValue('sem resultado')
    await flushPromises()

    expect(wrapper.text()).toContain('0 produtos encontrados')
    expect(wrapper.text()).toContain('Nenhum produto encontrado')

    const clearButton = wrapper.findAll('button').find((button) => button.text() === 'Limpar busca')

    expect(clearButton).toBeDefined()
    await clearButton?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('3 produtos encontrados')
    expect(scrollToTop).toHaveBeenCalled()
  })

  it('mostra erro e tenta novamente', async () => {
    mockedFetchMenu.mockRejectedValue(new Error('network error'))

    const { wrapper } = await mountMenu()
    await flushPromises()

    expect(wrapper.text()).toContain('Não foi possível carregar o cardápio.')

    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(mockedFetchMenu).toHaveBeenCalledTimes(2)
  })

  it('mostra cardápio vazio', async () => {
    mockedFetchMenu.mockResolvedValue(makeMenu({ categories: [] }))

    const { wrapper } = await mountMenu()
    await flushPromises()

    expect(wrapper.text()).toContain('Cardápio em preparação')
    expect(wrapper.text()).not.toContain('produtos encontrados')
  })
})
