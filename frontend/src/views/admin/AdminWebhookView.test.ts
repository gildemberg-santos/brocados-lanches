import { createPinia, setActivePinia } from 'pinia'
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import * as api from '../../services/adminWebhook'
import AdminWebhookView from './AdminWebhookView.vue'

describe('AdminWebhookView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.restoreAllMocks()
  })

  it('renderiza o formulário e salva a configuração', async () => {
    vi.spyOn(api, 'fetchAdminWebhook').mockResolvedValue({
      url: 'https://hooks.exemplo.invalid/path',
      active: true,
      createdAt: '2026-11-26T12:00:00Z',
      updatedAt: '2026-11-26T12:00:00Z',
    })
    const updateSpy = vi.spyOn(api, 'updateAdminWebhook').mockResolvedValue({
      url: 'http://127.0.0.1:5678/hook',
      active: true,
      createdAt: '2026-11-26T12:00:00Z',
      updatedAt: '2026-11-26T12:00:00Z',
    })

    const wrapper = mount(AdminWebhookView, {
      global: {
        stubs: {
          BaseButton: { template: '<button type="submit"><slot /></button>' },
        },
      },
    })
    await Promise.resolve()
    await wrapper.vm.$nextTick()

    const input = wrapper.get('#webhook-url')
    expect((input.element as HTMLInputElement).value).toBe('https://hooks.exemplo.invalid/path')
    await input.setValue('http://127.0.0.1:5678/hook')
    await wrapper.get('form').trigger('submit')
    await Promise.resolve()

    expect(updateSpy).toHaveBeenCalledWith({ url: 'http://127.0.0.1:5678/hook', active: true })
  })

  it('bloqueia envio sem URL quando ativo', async () => {
    vi.spyOn(api, 'fetchAdminWebhook').mockResolvedValue({
      url: '',
      active: true,
      createdAt: '2026-11-26T12:00:00Z',
      updatedAt: '2026-11-26T12:00:00Z',
    })
    const updateSpy = vi.spyOn(api, 'updateAdminWebhook')

    const wrapper = mount(AdminWebhookView, {
      global: {
        stubs: {
          BaseButton: { template: '<button type="submit"><slot /></button>' },
        },
      },
    })
    await Promise.resolve()
    await wrapper.vm.$nextTick()

    await wrapper.get('form').trigger('submit')

    expect(updateSpy).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Informe a URL do webhook')
  })

  it('mostra estado de erro sem expor detalhes internos', async () => {
    vi.spyOn(api, 'fetchAdminWebhook').mockRejectedValue(new Error('internal url'))
    const wrapper = mount(AdminWebhookView)
    await Promise.resolve()
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('Não foi possível carregar')
    expect(wrapper.text()).not.toContain('internal url')
  })
})
