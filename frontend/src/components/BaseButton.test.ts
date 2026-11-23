import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import BaseButton from './BaseButton.vue'

describe('BaseButton', () => {
  it('renders the default primary button and forwards attributes', () => {
    const wrapper = mount(BaseButton, {
      attrs: { 'data-test': 'action' },
      slots: { default: 'Salvar' },
    })

    expect(wrapper.get('button').attributes('type')).toBe('button')
    expect(wrapper.get('button').attributes('data-test')).toBe('action')
    expect(wrapper.classes()).toContain('base-button--primary')
  })

  it('emits click when enabled', async () => {
    const onClick = vi.fn()
    const wrapper = mount(BaseButton, {
      attrs: { onClick },
      slots: { default: 'Salvar' },
    })

    await wrapper.get('button').trigger('click')

    expect(onClick).toHaveBeenCalledTimes(1)
  })

  it('does not emit click when disabled', async () => {
    const onClick = vi.fn()
    const wrapper = mount(BaseButton, {
      props: { disabled: true },
      attrs: { onClick },
      slots: { default: 'Salvar' },
    })

    await wrapper.get('button').trigger('click')

    expect(wrapper.get('button').attributes('disabled')).toBeDefined()
    expect(onClick).not.toHaveBeenCalled()
  })

  it('shows a loading state and blocks interaction', () => {
    const wrapper = mount(BaseButton, {
      props: { loading: true },
      slots: { default: 'Salvando' },
    })

    expect(wrapper.get('button').attributes('aria-busy')).toBe('true')
    expect(wrapper.get('button').attributes('disabled')).toBeDefined()
    expect(wrapper.find('.base-button__spinner').exists()).toBe(true)
  })
})
