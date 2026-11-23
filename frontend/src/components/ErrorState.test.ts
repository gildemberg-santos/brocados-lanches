import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import ErrorState from './ErrorState.vue'

describe('ErrorState', () => {
  it('shows the friendly message and emits retry', async () => {
    const wrapper = mount(ErrorState, {
      props: { message: 'Não foi possível carregar o cardápio.' },
    })

    expect(wrapper.get('[role="alert"]').text()).toContain('Não foi possível carregar')
    expect(wrapper.text()).toContain('Não foi possível carregar o cardápio.')

    await wrapper.get('button').trigger('click')

    expect(wrapper.emitted('retry')).toHaveLength(1)
  })

  it('allows a custom retry label', () => {
    const wrapper = mount(ErrorState, {
      props: { message: 'Falhou.', retryLabel: 'Recarregar' },
    })

    expect(wrapper.get('button').text()).toBe('Recarregar')
  })
})
