import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import CategoryPill from './CategoryPill.vue'

describe('CategoryPill', () => {
  it('communicates the active state and emits select', async () => {
    const wrapper = mount(CategoryPill, {
      props: { label: 'Hambúrgueres', active: true },
    })

    expect(wrapper.get('button').attributes('aria-pressed')).toBe('true')
    expect(wrapper.text()).toBe('Hambúrgueres')

    await wrapper.get('button').trigger('click')

    expect(wrapper.emitted('select')).toHaveLength(1)
  })

  it('does not mark an inactive filter as pressed', () => {
    const wrapper = mount(CategoryPill, {
      props: { label: 'Bebidas', active: false },
    })

    expect(wrapper.get('button').attributes('aria-pressed')).toBe('false')
    expect(wrapper.classes()).not.toContain('category-pill--active')
  })
})
