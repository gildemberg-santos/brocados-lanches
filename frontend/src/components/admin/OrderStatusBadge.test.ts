import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import OrderStatusBadge from './OrderStatusBadge.vue'

describe('OrderStatusBadge', () => {
  it('mostra o rótulo em português', () => {
    const wrapper = mount(OrderStatusBadge, { props: { status: 'out_for_delivery' } })
    expect(wrapper.text()).toBe('Saiu para entrega')
  })

  it('mantém valor desconhecido legível sem quebrar', () => {
    const wrapper = mount(OrderStatusBadge, { props: { status: 'desconhecido' } })
    expect(wrapper.text()).toBe('desconhecido')
  })
})
