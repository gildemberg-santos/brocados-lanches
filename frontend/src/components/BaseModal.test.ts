import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { afterEach, describe, expect, it } from 'vitest'

import BaseModal from './BaseModal.vue'

afterEach(() => {
  document.body.style.overflow = ''
  document.body.innerHTML = ''
})

describe('BaseModal', () => {
  it('não renderiza enquanto estiver fechado', () => {
    const wrapper = mount(BaseModal, {
      props: { open: false, title: 'Novo produto' },
    })

    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
  })

  it('abre centralizado, nomeia o diálogo, foca o primeiro campo e trava a rolagem', async () => {
    const wrapper = mount(BaseModal, {
      attachTo: document.body,
      props: { open: true, title: 'Novo produto' },
      slots: { default: '<input id="primeiro" /><button type="button">Salvar</button>' },
    })
    await nextTick()

    const dialog = wrapper.get('[role="dialog"]')
    expect(dialog.attributes('aria-modal')).toBe('true')
    expect(dialog.attributes('aria-labelledby')).toBeTruthy()
    expect(wrapper.get('h2').text()).toBe('Novo produto')
    expect(wrapper.get('[data-testid="admin-modal"]').classes()).toContain('base-modal')
    expect((document.activeElement as HTMLElement).id).toBe('primeiro')
    expect(document.body.style.overflow).toBe('hidden')

    wrapper.unmount()
  })

  it('fecha com ESC e retorna o foco para o elemento anterior', async () => {
    const trigger = document.createElement('button')
    document.body.appendChild(trigger)
    trigger.focus()

    const wrapper = mount(BaseModal, {
      attachTo: document.body,
      props: { open: true, title: 'Novo produto' },
      slots: { default: '<input id="primeiro" />' },
    })
    await nextTick()
    expect(document.activeElement?.id).toBe('primeiro')

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    expect(wrapper.emitted('close')).toHaveLength(1)

    await wrapper.setProps({ open: false })
    expect(document.activeElement).toBe(trigger)
    expect(document.body.style.overflow).toBe('')
  })

  it('prende o foco dentro do diálogo ao tabular', async () => {
    const wrapper = mount(BaseModal, {
      attachTo: document.body,
      props: { open: true, title: 'Novo produto' },
      slots: {
        default: '<input id="primeiro" /><button id="ultimo" type="button">Salvar</button>',
      },
    })
    await nextTick()

    const first = wrapper.get('[aria-label="Fechar"]').element as HTMLElement
    const field = wrapper.get('#primeiro').element as HTMLElement
    const last = wrapper.get('#ultimo').element as HTMLElement

    expect(document.activeElement).toBe(field)

    last.focus()
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', bubbles: true }))
    expect(document.activeElement).toBe(first)

    document.dispatchEvent(
      new KeyboardEvent('keydown', { key: 'Tab', shiftKey: true, bubbles: true }),
    )
    expect(document.activeElement).toBe(last)

    wrapper.unmount()
  })

  it('emite close ao clicar no backdrop e no botão de fechar', async () => {
    const wrapper = mount(BaseModal, {
      attachTo: document.body,
      props: { open: true, title: 'Novo produto' },
      slots: { default: '<input id="primeiro" />' },
    })
    await nextTick()

    await wrapper.get('.base-modal__backdrop').trigger('click')
    await wrapper.get('[aria-label="Fechar"]').trigger('click')

    expect(wrapper.emitted('close')).toHaveLength(2)
    wrapper.unmount()
  })
})
