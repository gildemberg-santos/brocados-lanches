import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it } from 'vitest'

import { useCartStore } from '../stores/cart'
import { makeCatalogProduct } from '../test/fixtures/catalog'
import MobileTabBar from './MobileTabBar.vue'

const StubView = defineComponent({ name: 'StubView', render: () => h('div') })

async function mountTabBar(path = '/') {
  const pinia = createPinia()
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
  await router.push(path)
  await router.isReady()

  return mount(MobileTabBar, { global: { plugins: [pinia, router] } })
}

describe('MobileTabBar', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('oferece as quatro áreas principais com rótulo acessível', async () => {
    const wrapper = await mountTabBar()
    const nav = wrapper.get('nav[aria-label="Navegação inferior"]')
    const links = nav.findAll('a')

    expect(links.map((link) => link.text())).toEqual(['Início', 'Cardápio', 'Pedidos', 'Carrinho'])
    expect(links.map((link) => link.attributes('href'))).toEqual([
      '/',
      '/cardapio',
      '/meus-pedidos',
      '/carrinho',
    ])
  })

  it('marca a rota atual como ativa', async () => {
    const wrapper = await mountTabBar('/cardapio')
    const links = wrapper.findAll('a')

    expect(links[1].classes()).toContain('router-link-active')
    expect(links[0].classes()).not.toContain('router-link-active')
  })

  it('mostra a contagem do carrinho no rótulo acessível e no badge', async () => {
    const wrapper = await mountTabBar()
    const cart = useCartStore()
    cart.add(makeCatalogProduct({ id: 'product-1' }), {}, [], '', 2)

    await wrapper.vm.$nextTick()

    const cartLink = wrapper.get('.mobile-tab-bar__link--cart')
    expect(cartLink.attributes('aria-label')).toBe('Carrinho, 2 itens')
    expect(cartLink.get('.mobile-tab-bar__badge').text()).toBe('2')
  })

  it('não mostra badge quando o carrinho está vazio', async () => {
    const wrapper = await mountTabBar()

    expect(wrapper.find('.mobile-tab-bar__badge').exists()).toBe(false)
    expect(wrapper.get('.mobile-tab-bar__link--cart').attributes('aria-label')).toBe('Carrinho')
  })
})
