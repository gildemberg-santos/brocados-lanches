import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import PriceText from './PriceText.vue'

describe('PriceText', () => {
  it('formats integer cents as Brazilian currency', () => {
    const wrapper = mount(PriceText, { props: { valueCents: 2300 } })
    const text = wrapper.text()

    expect(text).toContain('R$')
    expect(text).toContain('23,00')
  })

  it('formats decimal cents without turning them into reais', () => {
    const wrapper = mount(PriceText, { props: { valueCents: 5490 } })

    expect(wrapper.text()).toContain('54,90')
  })
})
