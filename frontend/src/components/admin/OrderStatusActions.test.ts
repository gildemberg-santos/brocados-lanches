import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { describe, expect, it } from 'vitest'

import OrderStatusActions from './OrderStatusActions.vue'

function buttonByText(wrapper: ReturnType<typeof mount>, text: string) {
  const button = wrapper.findAll('button').find((candidate) => candidate.text() === text)
  if (!button) throw new Error(`Botão não encontrado: ${text}`)
  return button
}

describe('OrderStatusActions', () => {
  it('oferece apenas transições válidas para entrega', () => {
    const wrapper = mount(OrderStatusActions, {
      props: { status: 'pending', fulfillment: 'delivery' },
    })

    expect(wrapper.text()).toContain('Confirmar pedido')
    expect(wrapper.text()).toContain('Cancelar pedido')
    expect(wrapper.text()).not.toContain('Marcar como retirado')
  })

  it('emite a transição escolhida sem dados pessoais', async () => {
    const wrapper = mount(OrderStatusActions, {
      props: { status: 'confirmed', fulfillment: 'delivery' },
    })

    await buttonByText(wrapper, 'Iniciar preparo').trigger('click')

    expect(wrapper.emitted('change')).toEqual([[{ status: 'preparing' }]])
  })

  it('exige motivo para cancelar e o envia sem dados do cliente', async () => {
    const wrapper = mount(OrderStatusActions, {
      props: { status: 'pending', fulfillment: 'pickup' },
    })

    await buttonByText(wrapper, 'Cancelar pedido').trigger('click')
    await wrapper.find('form').trigger('submit')
    expect(wrapper.find('[role="alert"]').text()).toContain('Informe o motivo')
    expect(wrapper.emitted('change')).toBeUndefined()

    await wrapper.find('textarea').setValue('Sobres de cozinha')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.emitted('change')).toEqual([
      [{ status: 'cancelled', reason: 'Sobres de cozinha' }],
    ])
  })

  it('em retirada pronta não oferece cancelamento', () => {
    const wrapper = mount(OrderStatusActions, {
      props: { status: 'ready', fulfillment: 'pickup' },
    })

    expect(wrapper.text()).toContain('Marcar como retirado')
    expect(wrapper.text()).not.toContain('Cancelar pedido')
  })

  it('limpa o formulário quando o status muda', async () => {
    const wrapper = mount(OrderStatusActions, {
      props: { status: 'pending', fulfillment: 'delivery' },
    })
    await buttonByText(wrapper, 'Cancelar pedido').trigger('click')
    await wrapper.find('textarea').setValue('Motivo')

    await wrapper.setProps({ status: 'confirmed' })
    await nextTick()

    expect(wrapper.find('textarea').exists()).toBe(false)
  })
})
