import { mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import CartSummaryBar from './CartSummaryBar.vue'

const StubView = defineComponent({ render: () => h('div') })

describe('CartSummaryBar', () => {
  it('mostra a quantidade e o subtotal do carrinho e leva à rota do carrinho', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/', component: StubView },
        { path: '/carrinho', name: 'cart', component: StubView },
      ],
    })
    await router.push('/')
    await router.isReady()

    const wrapper = mount(CartSummaryBar, {
      props: { count: 3, subtotalCents: 8450 },
      global: { plugins: [router] },
    })

    expect(wrapper.text()).toContain('Ver carrinho (3)')
    expect(wrapper.text()).toContain('84,50')
    expect(wrapper.get('a').attributes('href')).toBe('/carrinho')
  })
})
