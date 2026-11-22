import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import QuantityStepper from './QuantityStepper.vue'

describe('QuantityStepper', () => {
  it('emits the next value and keeps the current value announced', async () => {
    const wrapper = mount(QuantityStepper, {
      props: { modelValue: 2, label: 'quantidade de Brocado Cheddar' },
    })

    await wrapper.get('[aria-label="Aumentar quantidade de Brocado Cheddar"]').trigger('click')
    await wrapper.get('[aria-label="Diminuir quantidade de Brocado Cheddar"]').trigger('click')

    expect(wrapper.emitted('update:modelValue')).toEqual([[3], [1]])
    expect(wrapper.get('output').text()).toBe('2')
  })

  it('disables the buttons at the limits', async () => {
    const wrapper = mount(QuantityStepper, { props: { modelValue: 1, min: 1, max: 2 } })

    expect(wrapper.get('[aria-label="Diminuir quantidade"]').attributes('disabled')).toBeDefined()

    await wrapper.get('[aria-label="Aumentar quantidade"]').trigger('click')
    await wrapper.setProps({ modelValue: 2 })

    expect(wrapper.get('[aria-label="Aumentar quantidade"]').attributes('disabled')).toBeDefined()
  })
})
