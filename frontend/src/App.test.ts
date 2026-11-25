import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import App from './App.vue'

const StubView = defineComponent({
  name: 'StubView',
  render: () => h('main', { id: 'conteudo', tabindex: '-1' }, 'stub'),
})

async function mountApp() {
  const pinia = createPinia()
  setActivePinia(pinia)

  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: StubView },
      { path: '/cardapio', component: StubView },
      { path: '/carrinho', component: StubView },
      { path: '/admin', component: StubView, meta: { admin: true } },
    ],
  })

  await router.push('/')
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
    expect(wrapper.get('.skip-link').attributes('href')).toBe('#conteudo')
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
  })
})
