import { mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import { makeProduct } from '../test/fixtures/catalog'
import ProductCardCompact from './ProductCardCompact.vue'

const StubView = defineComponent({
  name: 'StubView',
  render: () => h('div'),
})

async function mountCard(product = makeProduct()) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'home', component: StubView },
      { path: '/produto/:slug', name: 'product', component: StubView },
    ],
  })

  await router.push('/')
  await router.isReady()

  return mount(ProductCardCompact, {
    props: { product },
    global: {
      plugins: [router],
    },
  })
}

describe('ProductCardCompact', () => {
  it('shows the compact product summary and links to the detail', async () => {
    const wrapper = await mountCard()

    expect(wrapper.get('a').attributes('href')).toBe('/produto/brocado-cheddar')
    expect(wrapper.text()).toContain('Brocado Cheddar')
    expect(wrapper.text()).toContain('23,00')
  })

  it('marks a sold-out product', async () => {
    const wrapper = await mountCard(makeProduct({ available: false }))

    expect(wrapper.text()).toContain('Esgotado')
    expect(wrapper.classes()).toContain('product-card-compact--sold-out')
  })
})
