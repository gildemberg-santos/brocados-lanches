import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AdminUsersView from './AdminUsersView.vue'
import { fetchAdminUsers } from '../../services/adminUsers'

vi.mock('../../services/adminUsers', () => ({
  fetchAdminUsers: vi.fn(),
  createAdminUser: vi.fn(),
  updateAdminUser: vi.fn(),
  fetchAdminAuditLogs: vi.fn(),
}))

beforeEach(() => {
  setActivePinia(createPinia())
  vi.clearAllMocks()
  vi.mocked(fetchAdminUsers).mockResolvedValue({
    data: [
      {
        id: 'u1',
        name: 'Ana',
        email: 'ana@example.com',
        role: 'operator',
        active: true,
        lastLoginAt: null,
        createdAt: '',
        updatedAt: '',
      },
    ],
    meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
  })
})

describe('AdminUsersView', () => {
  it('carrega e exibe os usuários', async () => {
    const wrapper = mount(AdminUsersView)
    await flushPromises()

    expect(wrapper.get('h1').text()).toBe('Usuários')
    expect(wrapper.text()).toContain('ana@example.com')
  })

  it('abre o formulário de novo usuário', async () => {
    const wrapper = mount(AdminUsersView)
    await flushPromises()

    await wrapper
      .findAll('button')
      .find((button) => button.text() === 'Novo usuário')
      ?.trigger('click')

    expect(wrapper.get('[role="dialog"]').attributes('aria-modal')).toBe('true')
    expect(wrapper.get('[data-testid="admin-modal"] h2').text()).toBe('Novo usuário')
    expect(wrapper.find('#admin-user-password').exists()).toBe(true)
  })

  it('aplica busca ao enviar o filtro', async () => {
    const wrapper = mount(AdminUsersView)
    await flushPromises()

    await wrapper.find('#admin-users-search').setValue('ana')
    await wrapper.find('.admin-page__search').trigger('submit')
    await flushPromises()

    expect(fetchAdminUsers).toHaveBeenLastCalledWith(expect.objectContaining({ q: 'ana' }))
  })
})
