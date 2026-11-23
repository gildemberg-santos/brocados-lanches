import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import CartSummary from './CartSummary.vue'

describe('CartSummary', () => {
  it('shows subtotal, estimated fee and estimated total', () => {
    const wrapper = mount(CartSummary, {
      props: { subtotalCents: 5600, deliveryFeeCents: 800, itemCount: 2 },
    })

    expect(wrapper.text()).toContain('Itens (2)')
    expect(wrapper.text()).toContain('56,00')
    expect(wrapper.text()).toContain('8,00')
    expect(wrapper.text()).toContain('64,00')
    expect(wrapper.text()).toContain('estimada · menor taxa ativa')
    expect(wrapper.text()).toContain('Retirada não paga taxa de entrega.')
  })

  it('keeps the total partial when the fee is unknown', () => {
    const wrapper = mount(CartSummary, {
      props: { subtotalCents: 2300, deliveryFeeCents: null, itemCount: 1 },
    })

    expect(wrapper.text()).toContain('A calcular no checkout')
    expect(wrapper.text()).toContain('Total parcial')
    expect(wrapper.text()).not.toContain('Total estimado')
  })
})
