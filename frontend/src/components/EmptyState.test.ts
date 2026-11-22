import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import EmptyState from './EmptyState.vue'

describe('EmptyState', () => {
  it('presents the title, description and an action slot', () => {
    const wrapper = mount(EmptyState, {
      props: {
        title: 'Nenhum produto encontrado',
        description: 'Tente outra busca ou limpe os filtros.',
      },
      slots: {
        actions: '<button type="button">Limpar filtros</button>',
      },
    })

    expect(wrapper.get('h2').text()).toBe('Nenhum produto encontrado')
    expect(wrapper.text()).toContain('Tente outra busca')
    expect(wrapper.get('button').text()).toBe('Limpar filtros')
  })
})
