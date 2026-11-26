import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchDeliveryZones } from '../services/deliveryZones'
import { fetchMenu } from '../services/menu'
import { useCartStore } from '../stores/cart'
import type { CartItem } from '../types/cart'
import type { CatalogProduct, Menu } from '../types/catalog'
import {
  makeCategory,
  makeMenu,
  makeOption,
  makeOptionGroup,
  makeProduct,
} from '../test/fixtures/catalog'
import CartView from './CartView.vue'

vi.mock('../services/menu', () => ({ fetchMenu: vi.fn() }))
vi.mock('../services/deliveryZones', () => ({ fetchDeliveryZones: vi.fn() }))

const mockedFetchMenu = vi.mocked(fetchMenu)
const mockedFetchDeliveryZones = vi.mocked(fetchDeliveryZones)

const StubView = defineComponent({ name: 'StubView', render: () => h('div') })

const product = makeProduct({
  id: 'product-1',
  name: 'Brocado Cheddar',
  slug: 'brocado-cheddar',
  priceCents: 2300,
  optionGroups: [
    makeOptionGroup({
      id: 'doneness',
      name: 'Ponto da carne',
      options: [
        makeOption({ id: 'rare', name: 'Ao ponto', priceCents: 0 }),
        makeOption({ id: 'well-done', name: 'Bem passado', priceCents: 200 }),
      ],
    }),
  ],
})
const catalogProduct: CatalogProduct = {
  ...product,
  categoryId: 'category-1',
  categoryName: 'Hambúrgueres',
  categorySlug: 'hamburgueres',
}
const fullMenu = makeMenu({ categories: [makeCategory({ products: [product] })] })

function addItem(cart: ReturnType<typeof useCartStore>, optionId = 'rare', quantity = 1): CartItem {
  cart.add(
    catalogProduct,
    { doneness: [optionId] },
    [],
    optionId === 'rare' ? '' : 'sem cebola',
    quantity,
  )

  return cart.items[cart.items.length - 1]
}

interface MountOptions {
  seed?: boolean
  menu?: Menu
  menuError?: boolean
}

async function mountCart(options: MountOptions = {}) {
  const pinia = createPinia()
  setActivePinia(pinia)
  const cart = useCartStore()

  if (options.seed) {
    addItem(cart)
  }

  if (options.menuError) {
    mockedFetchMenu.mockRejectedValue(new Error('network error'))
  } else {
    mockedFetchMenu.mockResolvedValue(options.menu ?? fullMenu)
  }

  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'home', component: StubView },
      { path: '/cardapio', name: 'menu', component: StubView },
      { path: '/produto/:slug', name: 'product', component: StubView },
      { path: '/carrinho', name: 'cart', component: StubView },
      { path: '/checkout', name: 'checkout', component: StubView },
    ],
  })
  await router.push('/carrinho')
  await router.isReady()

  const wrapper = mount(CartView, { global: { plugins: [pinia, router] } })
  await flushPromises()

  return { wrapper, cart, router }
}

describe('CartView', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
  })

  it('shows the empty state with a link to the menu', async () => {
    const { wrapper } = await mountCart({ menu: makeMenu() })

    expect(wrapper.text()).toContain('Seu carrinho está vazio')
    expect(wrapper.get('a[href="/cardapio"]').attributes('href')).toBe('/cardapio')
  })

  it('lists items with catalog prices and the subtotal, without a delivery fee', async () => {
    const { wrapper } = await mountCart({ seed: true })

    expect(wrapper.text()).toContain('Brocado Cheddar')
    expect(wrapper.text()).toContain('Subtotal (1)')
    expect(wrapper.text()).toContain('23,00')
    expect(wrapper.text()).not.toContain('31,00')
    expect(wrapper.text()).toContain('A taxa de entrega aparece no checkout')
    expect(wrapper.text()).toContain('O total é calculado no servidor no próximo passo.')
  })

  it('does not load delivery zones before checkout', async () => {
    const { wrapper } = await mountCart({ seed: true })

    expect(mockedFetchDeliveryZones).not.toHaveBeenCalled()
    expect(wrapper.text()).not.toContain('estimada · menor taxa ativa')
    expect(wrapper.text()).not.toContain('Total parcial')
  })

  it('shows the selected options and note', async () => {
    const { wrapper } = await mountCart({ seed: true })

    expect(wrapper.text()).toContain('Ponto da carne: Ao ponto')
  })

  it('changes quantity, removes and clears items', async () => {
    const { wrapper, cart } = await mountCart({ seed: true })

    await wrapper.get('[aria-label="Aumentar quantidade de Brocado Cheddar"]').trigger('click')
    expect(wrapper.text()).toContain('46,00')

    await wrapper.get('[aria-label="Remover Brocado Cheddar do carrinho"]').trigger('click')
    expect(wrapper.text()).toContain('Seu carrinho está vazio')

    addItem(cart)
    await flushPromises()
    const clearButton = wrapper
      .findAll('button')
      .find((button) => button.text() === 'Limpar carrinho')
    expect(clearButton).toBeDefined()
    await clearButton!.trigger('click')

    expect(cart.isEmpty).toBe(true)
  })

  it('drops items that are no longer in the catalog', async () => {
    const { wrapper, cart } = await mountCart({
      seed: true,
      menu: makeMenu({ categories: [] }),
    })

    expect(cart.isEmpty).toBe(true)
    expect(wrapper.text()).toContain('Seu carrinho está vazio')
  })

  it('shows a retry state when the menu fails', async () => {
    const { wrapper } = await mountCart({ seed: true, menuError: true })

    expect(wrapper.text()).toContain('Não foi possível carregar o cardápio.')
    expect(mockedFetchMenu).toHaveBeenCalledTimes(1)

    mockedFetchMenu.mockResolvedValue(fullMenu)
    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Brocado Cheddar')
  })

  it('navigates to checkout from the cart CTA', async () => {
    const { wrapper, router } = await mountCart({ seed: true })
    const continueButton = wrapper.findAll('button').find((button) => button.text() === 'Continuar')

    expect(continueButton?.attributes('disabled')).toBeUndefined()
    await continueButton!.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.name).toBe('checkout')
  })
})
