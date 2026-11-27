import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'

import StatusToast from './StatusToast.vue'

describe('StatusToast', () => {
  afterEach(() => {
    vi.useRealTimers()
  })

  it('anuncia a mensagem em região viva e permite fechar', async () => {
    const wrapper = mount(StatusToast, { props: { message: `Pedido #${1042}: Pronto` } })

    const status = wrapper.get('[role="status"]')
    expect(status.attributes('aria-live')).toBe('polite')
    expect(status.text()).toContain(`Pedido #${1042}: Pronto`)

    await wrapper.get('[aria-label="Fechar notificação"]').trigger('click')
    expect(wrapper.emitted('close')).toHaveLength(1)
  })

  it('fecha automaticamente após quatro segundos', async () => {
    vi.useFakeTimers()
    const wrapper = mount(StatusToast, { props: { message: `Pedido #${1042}: Pronto` } })

    vi.advanceTimersByTime(3999)
    expect(wrapper.emitted('close')).toBeUndefined()

    vi.advanceTimersByTime(1)
    expect(wrapper.emitted('close')).toHaveLength(1)
  })

  it('reinicia o tempo quando a mensagem muda', async () => {
    vi.useFakeTimers()
    const wrapper = mount(StatusToast, { props: { message: `Pedido #${1042}: Pronto` } })

    vi.advanceTimersByTime(3000)
    await wrapper.setProps({ message: `Pedido #${1042}: Saiu para entrega` })
    vi.advanceTimersByTime(2000)

    expect(wrapper.emitted('close')).toBeUndefined()
    vi.advanceTimersByTime(2000)
    expect(wrapper.emitted('close')).toHaveLength(1)
  })
})
