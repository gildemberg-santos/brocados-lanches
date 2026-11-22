import { mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import SiteHeader from './SiteHeader.vue'

const StubView = defineComponent({
  name: 'StubView',
  render: () => h('div'),
})

async function mountHeader() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'home', component: StubView },
      { path: '/cardapio', name: 'menu', component: StubView },
    ],
  })

  await router.push('/')
  await router.isReady()

  return mount(SiteHeader, {
    global: {
      plugins: [router],
    },
  })
}

describe('SiteHeader', () => {
  it('links the brand to the home page', async () => {
    const wrapper = await mountHeader()

    expect(wrapper.get('a[aria-label="Brocados Lanches, início"]').attributes('href')).toBe('/')
    expect(wrapper.text()).toContain('BROCADOS')
  })

  it('offers navigation to home and menu', async () => {
    const wrapper = await mountHeader()
    const links = wrapper.findAll('.site-header__link')

    expect(links.map((link) => link.text())).toEqual(['Início', 'Cardápio'])
    expect(links[1].attributes('href')).toBe('/cardapio')
  })
})
