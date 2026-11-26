import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import CartSummary from './CartSummary.vue'

describe('CartSummary', () => {
  it('mostra apenas o subtotal e explica quando a taxa aparece', () => {
    const wrapper = mount(CartSummary, {
      props: { subtotalCents: 5600, itemCount: 2 },
    })

    expect(wrapper.text()).toContain('Subtotal (2)')
    expect(wrapper.text()).toContain('56,00')
    expect(wrapper.text()).not.toContain('Taxa de entrega estimada')
    expect(wrapper.text()).not.toContain('8,00')
    expect(wrapper.text()).not.toContain('64,00')
    expect(wrapper.text()).toContain('A taxa de entrega aparece no checkout')
    expect(wrapper.text()).toContain('O total é calculado no servidor')
  })

  it('atualiza o subtotal conforme a quantidade de itens', () => {
    const wrapper = mount(CartSummary, {
      props: { subtotalCents: 2300, itemCount: 1 },
    })

    expect(wrapper.text()).toContain('Subtotal (1)')
    expect(wrapper.text()).toContain('23,00')
    expect(wrapper.text()).not.toContain('Total parcial')
  })
})
