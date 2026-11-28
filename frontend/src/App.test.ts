import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import App from './App.vue'
import { useCartStore } from './stores/cart'
import { makeCatalogProduct } from './test/fixtures/catalog'

const StubView = defineComponent({
  name: 'StubView',
  render: () => h('main', { id: 'conteudo', tabindex: '-1' }, 'stub'),
})

async function mountApp(path = '/', seedCart = false) {
  const pinia = createPinia()
  setActivePinia(pinia)
  if (seedCart) {
    useCartStore().add(makeCatalogProduct({ id: 'product-1' }), {}, [], '', 1)
  }

  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'home', component: StubView },
      { path: '/cardapio', component: StubView },
      { path: '/carrinho', component: StubView },
      { path: '/checkout', component: StubView, meta: { hideTabBar: true } },
      { path: '/admin', component: StubView, meta: { admin: true } },
    ],
  })

  await router.push(path)
  await router.isReady()

  return mount(App, {
    global: {
      plugins: [pinia, router],
    },
  })
}

describe('App', () => {
  it('renders the shared header, footer and the active route through RouterView', async () => {
    const wrapper = await mountApp()

    expect(wrapper.text()).toContain('BROCADOS')
    expect(wrapper.text()).toContain('stub')
    expect(wrapper.find('nav[aria-label="Links institucionais"]').exists()).toBe(true)
  })

  it('offers a skip link to the main content', async () => {
    const wrapper = await mountApp()

    expect(wrapper.get('.skip-link').attributes('href')).toBe('#conteudo')
    expect(wrapper.get('#conteudo').attributes('tabindex')).toBe('-1')
    expect(wrapper.findAll('main')).toHaveLength(1)
  })

  it('shows the institutional links on public routes', async () => {
    const wrapper = await mountApp()
    const links = wrapper.findAll('.site-footer__link')

    expect(links.map((link) => link.text())).toEqual([
      'Contato',
      'Privacidade',
      'Termos',
      'Créditos',
    ])
    expect(links.map((link) => link.attributes('href'))).toEqual([
      '/contato',
      '/privacidade',
      '/termos',
      '/creditos',
    ])
  })

  it('hides the public header on admin routes', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/', component: StubView },
        { path: '/admin', component: StubView, meta: { admin: true } },
      ],
    })
    await router.push('/admin')
    await router.isReady()

    const wrapper = mount(App, { global: { plugins: [pinia, router] } })

    expect(wrapper.find('.site-header').exists()).toBe(false)
    expect(wrapper.find('.site-footer').exists()).toBe(false)
    expect(wrapper.find('[data-testid="mobile-tab-bar"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="mobile-bottom-spacer"]').exists()).toBe(false)
    expect(wrapper.get('.skip-link').attributes('href')).toBe('#conteudo')
  })

  it('shows the mobile tab bar and ends the footer above it', async () => {
    const wrapper = await mountApp()

    expect(wrapper.find('[data-testid="mobile-tab-bar"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="mobile-bottom-spacer"]').classes()).toContain(
      'mobile-bottom-spacer--tab',
    )
  })

  it('hides the mobile tab bar and its spacer on focused routes', async () => {
    const wrapper = await mountApp('/checkout')

    expect(wrapper.find('[data-testid="mobile-tab-bar"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="mobile-bottom-spacer"]').exists()).toBe(false)
  })

  it('reserves space for the cart bar above the tab bar on the home route', async () => {
    const wrapper = await mountApp('/', true)

    expect(wrapper.get('[data-testid="mobile-bottom-spacer"]').classes()).toContain(
      'mobile-bottom-spacer--cart',
    )
  })

  it('hides the footer on the admin login route', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/', component: StubView },
        { path: '/admin/login', component: StubView },
      ],
    })
    await router.push('/admin/login')
    await router.isReady()

    const wrapper = mount(App, { global: { plugins: [pinia, router] } })

    expect(wrapper.find('.site-header').exists()).toBe(false)
    expect(wrapper.find('.site-footer').exists()).toBe(false)
    expect(wrapper.find('[data-testid="mobile-bottom-spacer"]').exists()).toBe(false)
  })
})
