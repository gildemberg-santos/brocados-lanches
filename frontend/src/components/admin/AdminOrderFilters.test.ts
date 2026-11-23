import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import AdminOrderFilters from './AdminOrderFilters.vue'

describe('AdminOrderFilters', () => {
  it('envia somente os filtros preenchidos', async () => {
    const wrapper = mount(AdminOrderFilters, { props: { filters: {} } })

    await wrapper.find('#order-search').setValue('  8708  ')
    await wrapper.find('#order-status').setValue('pending')
    await wrapper.find('#order-fulfillment').setValue('delivery')
    await wrapper.find('#order-from').setValue('2026-11-01')
    await wrapper.find('#order-to').setValue('2026-11-30')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.emitted('apply')).toEqual([
      [
        {
          status: 'pending',
          fulfillment: 'delivery',
          q: '8708',
          from: '2026-11-01T00:00:00Z',
          to: '2026-11-30T23:59:59Z',
        },
      ],
    ])
  })

  it('limpa os filtros e o formulário', async () => {
    const wrapper = mount(AdminOrderFilters, {
      props: { filters: { status: 'pending', q: '8708' } },
    })

    const limpar = wrapper.findAll('button').find((button) => button.text() === 'Limpar')
    if (!limpar) throw new Error('Botão Limpar não encontrado')
    await limpar.trigger('click')

    expect(wrapper.emitted('apply')).toEqual([[{}]])
    expect((wrapper.find('#order-search').element as HTMLInputElement).value).toBe('')
    expect((wrapper.find('#order-status').element as HTMLSelectElement).value).toBe('')
  })
})
