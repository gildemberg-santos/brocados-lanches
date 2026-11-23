import { mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import { makeProduct } from '../test/fixtures/catalog'
import ProductCard from './ProductCard.vue'

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

  return mount(ProductCard, {
    props: { product },
    global: {
      plugins: [router],
    },
  })
}

describe('ProductCard', () => {
  it('links to the product detail with name, description and price', async () => {
    const wrapper = await mountCard()

    const link = wrapper.get('a')
    expect(link.attributes('href')).toBe('/produto/brocado-cheddar')
    expect(link.attributes('aria-label')).toBe('Ver detalhes de Brocado Cheddar')
    expect(wrapper.text()).toContain('Brocado Cheddar')
    expect(wrapper.text()).toContain('Blend artesanal com cheddar.')
    expect(wrapper.text()).toContain('23,00')
    expect(wrapper.text()).toContain('Ver item')
  })

  it('marks a sold-out product without an enabled add action', async () => {
    const wrapper = await mountCard(makeProduct({ available: false }))

    expect(wrapper.text()).toContain('Esgotado')
    expect(wrapper.text()).toContain('Indisponível')
    expect(wrapper.get('a').attributes('aria-label')).toBe(
      'Ver detalhes de Brocado Cheddar (esgotado)',
    )
  })

  it('uses the product name as image alt when no alt was provided', async () => {
    const wrapper = await mountCard()

    expect(wrapper.get('[role="img"]').attributes('aria-label')).toBe('Brocado Cheddar')
  })
})
