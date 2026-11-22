import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import { makeOption, makeOptionGroup } from '../test/fixtures/catalog'
import OptionGroup from './OptionGroup.vue'

const twoOptions = [
  makeOption({ id: 'rare', name: 'Ao ponto', priceCents: 0 }),
  makeOption({ id: 'well', name: 'Bem passado', priceCents: 200 }),
]

describe('OptionGroup', () => {
  it('uses radios and emits a single selection', async () => {
    const wrapper = mount(OptionGroup, {
      props: {
        group: makeOptionGroup({ options: twoOptions }),
        selectedIds: [],
      },
    })

    expect(wrapper.get('fieldset').attributes('aria-required')).toBe('true')
    expect(wrapper.findAll('input[type="radio"]')).toHaveLength(2)
    expect(wrapper.text()).toContain('Obrigatório · escolha 1')
    expect(wrapper.text()).toContain('2,00')

    await wrapper.findAll('input[type="radio"]')[1].setValue()

    expect(wrapper.emitted('update:selectedIds')?.[0]).toEqual([['well']])
  })

  it('uses checkboxes and blocks new selections at the maximum', () => {
    const wrapper = mount(OptionGroup, {
      props: {
        group: makeOptionGroup({
          required: false,
          minSelections: 0,
          maxSelections: 2,
          options: [makeOption({ id: 'a' }), makeOption({ id: 'b' }), makeOption({ id: 'c' })],
        }),
        selectedIds: ['a', 'b'],
      },
    })

    const checkboxes = wrapper.findAll('input[type="checkbox"]')
    expect(checkboxes).toHaveLength(3)
    expect(checkboxes[2].attributes('disabled')).toBeDefined()
    expect(wrapper.text()).toContain('Opcional · escolha até 2')
  })

  it('emits the remaining ids when a checkbox is unchecked', async () => {
    const wrapper = mount(OptionGroup, {
      props: {
        group: makeOptionGroup({
          required: false,
          minSelections: 0,
          maxSelections: 2,
          options: [makeOption({ id: 'a' }), makeOption({ id: 'b' }), makeOption({ id: 'c' })],
        }),
        selectedIds: ['a', 'b'],
      },
    })

    await wrapper.findAll('input[type="checkbox"]')[0].setValue(false)

    expect(wrapper.emitted('update:selectedIds')?.[0]).toEqual([['b']])
  })

  it('shows validation errors and an empty group message', () => {
    const wrapper = mount(OptionGroup, {
      props: {
        group: makeOptionGroup({ options: [] }),
        selectedIds: [],
        error: 'Este grupo está sem opções disponíveis.',
      },
    })

    expect(wrapper.get('[role="alert"]').text()).toBe('Este grupo está sem opções disponíveis.')
    expect(wrapper.text()).toContain('Nenhuma opção disponível.')
  })

  it('does not emit when an unavailable option is selected', async () => {
    const wrapper = mount(OptionGroup, {
      props: {
        group: makeOptionGroup({
          required: false,
          options: [makeOption({ id: 'off', available: false })],
        }),
        selectedIds: [],
      },
    })

    const input = wrapper.get('input[type="radio"]')
    expect(input.attributes('disabled')).toBeDefined()

    await input.setValue()

    expect(wrapper.emitted('update:selectedIds')).toBeUndefined()
  })
})
