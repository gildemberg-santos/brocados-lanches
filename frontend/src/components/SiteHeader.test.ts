import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it } from 'vitest'

import { useCartStore } from '../stores/cart'
import { makeCatalogProduct } from '../test/fixtures/catalog'
import SiteHeader from './SiteHeader.vue'

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
  })

  it('links the brand to the home page', async () => {
    const wrapper = await mountHeader()

    expect(wrapper.get('a[aria-label="Brocados Lanches, início"]').attributes('href')).toBe('/')
    expect(wrapper.text()).toContain('BROCADOS')
  })

  it('offers navigation to home, menu and cart', async () => {
    const wrapper = await mountHeader()
    const links = wrapper.findAll('.site-header__link:not(.site-header__link--cart)')

    expect(links.map((link) => link.text())).toEqual(['Início', 'Cardápio'])
    expect(links[1].attributes('href')).toBe('/cardapio')
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
})
