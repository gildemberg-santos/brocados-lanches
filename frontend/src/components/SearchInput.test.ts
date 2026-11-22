import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import SearchInput from './SearchInput.vue'

describe('SearchInput', () => {
  it('emits every typed value', async () => {
    const wrapper = mount(SearchInput, { props: { modelValue: '' } })

    await wrapper.get('input').setValue('cheddar')

    expect(wrapper.emitted('update:modelValue')).toEqual([['cheddar']])
  })

  it('renders a labelled search field and a clear action', async () => {
    const wrapper = mount(SearchInput, { props: { modelValue: 'cheddar' } })

    expect(wrapper.get('input').attributes('aria-label')).toBe('Buscar no cardápio')
    expect(wrapper.get('button').attributes('aria-label')).toBe('Limpar busca')

    await wrapper.get('button').trigger('click')

    expect(wrapper.emitted('update:modelValue')).toEqual([['']])
  })

  it('hides the clear action when the search is empty', () => {
    const wrapper = mount(SearchInput, { props: { modelValue: '' } })

    expect(wrapper.find('button').exists()).toBe(false)
  })
})
