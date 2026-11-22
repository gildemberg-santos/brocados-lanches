import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchMenu } from '../services/menu'
import { useCartStore } from '../stores/cart'
import {
  makeAddon,
  makeCategory,
  makeMenu,
  makeOption,
  makeOptionGroup,
  makeProduct,
} from '../test/fixtures/catalog'
import ProductView from './ProductView.vue'

vi.mock('../services/menu', () => ({
  fetchMenu: vi.fn(),
}))

const mockedFetchMenu = vi.mocked(fetchMenu)

const StubView = defineComponent({
  name: 'StubView',
  render: () => h('div'),
})

const product = makeProduct({
  id: 'product-1',
  name: 'Brocado Cheddar',
  slug: 'brocado-cheddar',
  description: 'Blend artesanal com cheddar derretido.',
  priceCents: 2300,
  allergens: ['leite', 'glúten'],
  optionGroups: [
    makeOptionGroup({
      id: 'doneness',
      name: 'Ponto da carne',
      required: true,
      minSelections: 1,
      maxSelections: 1,
      options: [
        makeOption({ id: 'rare', name: 'Ao ponto', priceCents: 0 }),
        makeOption({ id: 'well-done', name: 'Bem passado', priceCents: 200 }),
      ],
    }),
  ],
  addons: [makeAddon({ id: 'bacon', name: 'Bacon crocante', priceCents: 500 })],
})

const menu = makeMenu({
  categories: [makeCategory({ products: [product] })],
})

async function mountProduct(path = '/produto/brocado-cheddar') {
  const pinia = createPinia()
  setActivePinia(pinia)

  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'home', component: StubView },
      { path: '/cardapio', name: 'menu', component: StubView },
      { path: '/produto/:slug', name: 'product', component: StubView },
      { path: '/carrinho', name: 'cart', component: StubView },
    ],
  })

  await router.push(path)
  await router.isReady()

  const wrapper = mount(ProductView, {
    global: {
      plugins: [pinia, router],
    },
  })

  return { wrapper, router }
}

describe('ProductView', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
  })

  it('shows loading and then the product with options, addons and base price', async () => {
    mockedFetchMenu.mockResolvedValue(menu)

    const { wrapper } = await mountProduct()

    expect(wrapper.text()).toContain('Carregando produto')

    await flushPromises()

    expect(wrapper.text()).toContain('Brocado Cheddar')
    expect(wrapper.text()).toContain('Blend artesanal com cheddar derretido.')
    expect(wrapper.text()).toContain('Alergênicos: leite, glúten')
    expect(wrapper.text()).toContain('Ponto da carne')
    expect(wrapper.text()).toContain('Bacon crocante')
    expect(wrapper.text()).toContain('23,00')
    expect(wrapper.text()).toContain('0/140 caracteres')
  })

  it('blocks submit when a required option is missing', async () => {
    mockedFetchMenu.mockResolvedValue(menu)

    const { wrapper } = await mountProduct()
    await flushPromises()

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('Revise as opções obrigatórias antes de adicionar.')
    expect(wrapper.text()).toContain('Escolha uma opção.')
    expect(wrapper.text()).not.toContain('Adicionado ao carrinho.')
  })

  it('adds the configured item to the cart with IDs and choices only', async () => {
    mockedFetchMenu.mockResolvedValue(menu)

    const { wrapper } = await mountProduct()
    await flushPromises()

    await wrapper.findAll('input[type="radio"]')[1].setValue()
    await wrapper.findAll('input[type="checkbox"]')[0].setValue(true)
    await wrapper.get('#product-note').setValue('sem cebola')
    await flushPromises()

    expect(wrapper.text()).toContain('30,00')
    expect(wrapper.text()).toContain('10/140 caracteres')

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('Adicionado ao carrinho.')
    expect(wrapper.text()).not.toContain('Revise as opções obrigatórias')

    const cart = useCartStore()
    expect(cart.count).toBe(1)
    expect(cart.items[0]).toMatchObject({
      productId: 'product-1',
      optionIds: ['well-done'],
      addonIds: ['bacon'],
      note: 'sem cebola',
    })
    expect(localStorage.getItem('brocados.cart')).not.toContain('3000')
  })

  it('merges repeated configurations into the same cart line', async () => {
    mockedFetchMenu.mockResolvedValue(menu)

    const { wrapper } = await mountProduct()
    await flushPromises()

    await wrapper.findAll('input[type="radio"]')[0].setValue()
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    const cart = useCartStore()
    expect(cart.items).toHaveLength(1)
    expect(cart.count).toBe(2)
  })

  it('disables the add action when the product is sold out', async () => {
    mockedFetchMenu.mockResolvedValue(
      makeMenu({
        categories: [
          makeCategory({
            products: [makeProduct({ ...product, available: false })],
          }),
        ],
      }),
    )

    const { wrapper } = await mountProduct()
    await flushPromises()

    expect(wrapper.text()).toContain('Este item está esgotado.')
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeDefined()
  })

  it('shows a not-found state for an unknown slug', async () => {
    mockedFetchMenu.mockResolvedValue(makeMenu({ categories: [] }))

    const { wrapper } = await mountProduct('/produto/nao-existe')
    await flushPromises()

    expect(wrapper.text()).toContain('Produto não encontrado')
    expect(wrapper.find('a[href="/cardapio"]').exists()).toBe(true)
  })

  it('shows an error state and retries', async () => {
    mockedFetchMenu.mockRejectedValue(new Error('network error'))

    const { wrapper } = await mountProduct()
    await flushPromises()

    expect(wrapper.text()).toContain('Não foi possível carregar o cardápio.')

    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(mockedFetchMenu).toHaveBeenCalledTimes(2)
  })
})
