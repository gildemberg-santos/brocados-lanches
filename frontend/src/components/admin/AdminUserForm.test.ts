import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import AdminUserForm from './AdminUserForm.vue'

describe('AdminUserForm', () => {
  it('emite os dados para criar usuário', async () => {
    const wrapper = mount(AdminUserForm, {
      props: { item: null, saving: false, error: '', success: '' },
    })

    await wrapper.find('#admin-user-name').setValue('Ana')
    await wrapper.find('#admin-user-email').setValue('ana@example.com')
    await wrapper.find('#admin-user-password').setValue('senha-forte-123')
    await wrapper.find('#admin-user-role').setValue('owner')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.emitted('submit')?.[0]).toEqual([
      {
        name: 'Ana',
        email: 'ana@example.com',
        password: 'senha-forte-123',
        role: 'owner',
      },
    ])
  })

  it('bloqueia senha temporária curta', async () => {
    const wrapper = mount(AdminUserForm, {
      props: { item: null, saving: false, error: '', success: '' },
    })

    await wrapper.find('#admin-user-name').setValue('Ana')
    await wrapper.find('#admin-user-email').setValue('ana@example.com')
    await wrapper.find('#admin-user-password').setValue('curta')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.emitted('submit')).toBeUndefined()
    expect(wrapper.text()).toContain('12 caracteres')
    expect(wrapper.get('#admin-user-password').attributes('aria-invalid')).toBe('true')
    expect(wrapper.get('#admin-user-password').attributes('aria-describedby')).toBe(
      'admin-user-password-error',
    )
    expect(wrapper.get('#admin-user-password-error').text()).toContain('12 caracteres')
  })

  it('associa os erros de nome e e-mail aos campos correspondentes', async () => {
    const wrapper = mount(AdminUserForm, {
      props: { item: null, saving: false, error: '', success: '' },
    })

    await wrapper.find('form').trigger('submit')

    expect(wrapper.get('#admin-user-name').attributes('aria-invalid')).toBe('true')
    expect(wrapper.get('#admin-user-name').attributes('aria-describedby')).toBe(
      'admin-user-name-error',
    )
    expect(wrapper.get('#admin-user-email').attributes('aria-invalid')).toBe('true')
    expect(wrapper.get('#admin-user-email').attributes('aria-describedby')).toBe(
      'admin-user-email-error',
    )
  })

  it('não pede senha ao editar', () => {
    const wrapper = mount(AdminUserForm, {
      props: {
        item: {
          id: 'u1',
          name: 'Ana',
          email: 'ana@example.com',
          role: 'operator',
          active: true,
          lastLoginAt: null,
          createdAt: '',
          updatedAt: '',
        },
        saving: false,
        error: '',
        success: '',
      },
    })

    expect(wrapper.find('#admin-user-password').exists()).toBe(false)
    expect((wrapper.find('#admin-user-name').element as HTMLInputElement).value).toBe('Ana')
  })
})
