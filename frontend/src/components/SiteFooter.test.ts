import { mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import SiteFooter from './SiteFooter.vue'

const StubView = defineComponent({ name: 'StubView', render: () => h('div') })

async function mountFooter() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: StubView },
      { path: '/contato', component: StubView },
      { path: '/privacidade', component: StubView },
      { path: '/termos', component: StubView },
      { path: '/creditos', component: StubView },
    ],
  })
  await router.push('/')
  await router.isReady()

  return mount(SiteFooter, {
    global: { plugins: [router] },
  })
}

describe('SiteFooter', () => {
  it('exposes institutional navigation with an accessible label', async () => {
    const wrapper = await mountFooter()
    const nav = wrapper.find('nav[aria-label="Links institucionais"]')

    expect(nav.exists()).toBe(true)
  })

  it('links to contact, privacy, terms and credits without inventing contact data', async () => {
    const wrapper = await mountFooter()
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
    expect(wrapper.text()).not.toMatch(/\+55/)
  })

  it('termina após os links, sem reservar espaço das barras fixas', async () => {
    const wrapper = await mountFooter()

    expect(wrapper.classes()).toEqual(['site-footer'])
  })
})
