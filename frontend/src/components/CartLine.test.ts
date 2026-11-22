import { mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import { makeCatalogProduct } from '../test/fixtures/catalog'
import type { CartLine as CartLineModel } from '../types/cart'
import CartLine from './CartLine.vue'

const StubView = defineComponent({ name: 'StubView', render: () => h('div') })

async function mountLine(line: CartLineModel) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: StubView },
      { path: '/produto/:slug', name: 'product', component: StubView },
    ],
  })
  await router.push('/')
  await router.isReady()

  return mount(CartLine, { props: { line }, global: { plugins: [router] } })
}

const line: CartLineModel = {
  key: 'product-1|rare|bacon|sem cebola',
  product: makeCatalogProduct({
    id: 'product-1',
    name: 'Brocado Cheddar',
    slug: 'brocado-cheddar',
  }),
  optionNames: ['Ponto da carne: Ao ponto'],
  addonNames: ['Bacon crocante'],
  note: 'sem cebola',
  quantity: 2,
  unitPriceCents: 2800,
  lineTotalCents: 5600,
}

describe('CartLine', () => {
  it('shows name, choices, note and prices from the catalog', async () => {
    const wrapper = await mountLine(line)

    expect(wrapper.text()).toContain('Brocado Cheddar')
    expect(wrapper.text()).toContain('Ponto da carne: Ao ponto')
    expect(wrapper.text()).toContain('Adicionais: Bacon crocante')
    expect(wrapper.text()).toContain('Obs.: sem cebola')
    expect(wrapper.text()).toContain('28,00')
    expect(wrapper.text()).toContain('56,00')
  })

  it('emits quantity changes and removal', async () => {
    const wrapper = await mountLine(line)

    await wrapper.get('[aria-label="Aumentar quantidade de Brocado Cheddar"]').trigger('click')
    await wrapper.get('[aria-label="Remover Brocado Cheddar do carrinho"]').trigger('click')

    expect(wrapper.emitted('update:quantity')).toEqual([[3]])
    expect(wrapper.emitted('remove')).toHaveLength(1)
  })
})
