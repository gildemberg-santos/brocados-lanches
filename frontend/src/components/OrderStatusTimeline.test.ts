import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import OrderStatusTimeline from './OrderStatusTimeline.vue'

const base = {
  createdAt: '2026-11-22T02:30:00Z',
  updatedAt: '2026-11-22T03:10:00Z',
}

describe('OrderStatusTimeline', () => {
  it('mostra a sequência de entrega até o status atual', () => {
    const wrapper = mount(OrderStatusTimeline, {
      props: { ...base, status: 'preparing', fulfillment: 'delivery' },
    })
    const items = wrapper.findAll('li')
    const labels = items.map((item) => item.find('.timeline__label').text())

    expect(labels).toEqual([
      'Pedido registrado',
      'Confirmado',
      'Em preparo',
      'Pronto',
      'Saiu para entrega',
      'Entregue',
    ])
    expect(items[0].find('time').attributes('datetime')).toBe(base.createdAt)
    expect(items[2].attributes('aria-current')).toBe('step')
    expect(items[2].find('time').attributes('datetime')).toBe(base.updatedAt)
    expect(items[3].text()).toContain('Aguardando')
    expect(wrapper.text()).not.toContain('out_for_delivery')
  })

  it('mostra a sequência de retirada sem etapas de entrega', () => {
    const wrapper = mount(OrderStatusTimeline, {
      props: { ...base, status: 'ready', fulfillment: 'pickup' },
    })

    expect(wrapper.findAll('li')).toHaveLength(5)
    expect(wrapper.text()).toContain('Retirado')
    expect(wrapper.text()).not.toContain('Saiu para entrega')
    expect(wrapper.text()).not.toContain('Entregue')
  })

  it('não inventa etapas para pedido cancelado', () => {
    const wrapper = mount(OrderStatusTimeline, {
      props: { ...base, status: 'cancelled', fulfillment: 'delivery' },
    })
    const items = wrapper.findAll('li')

    expect(items).toHaveLength(1)
    expect(items[0].text()).toContain('Cancelado')
    expect(items[0].attributes('aria-current')).toBe('step')
  })

  it('trata status fora da sequência sem quebrar a timeline', () => {
    const wrapper = mount(OrderStatusTimeline, {
      props: { ...base, status: 'desconhecido', fulfillment: 'pickup', updatedAt: 'invalida' },
    })

    expect(wrapper.findAll('li')).toHaveLength(1)
    expect(wrapper.text()).toContain('desconhecido')
    expect(wrapper.text()).toContain('—')
  })
})
