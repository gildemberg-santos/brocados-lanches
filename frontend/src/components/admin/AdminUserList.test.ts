import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import AdminUserList from './AdminUserList.vue'
import type { AdminUser } from '../../types/adminUsers'

const user: AdminUser = {
  id: 'u1',
  name: 'Ana',
  email: 'ana@example.com',
  role: 'operator',
  active: true,
  lastLoginAt: null,
  createdAt: '',
  updatedAt: '',
}

const props = {
  items: [user],
  state: 'success' as const,
  error: '',
  isEmpty: false,
  meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
  actorId: 'actor',
}

describe('AdminUserList', () => {
  it('mostra usuário e ações', () => {
    const wrapper = mount(AdminUserList, { props })

    expect(wrapper.text()).toContain('Ana')
    expect(wrapper.text()).toContain('ana@example.com')
    expect(wrapper.text()).toContain('Desativar')
  })

  it('não oferece desativação para o próprio usuário', () => {
    const wrapper = mount(AdminUserList, { props: { ...props, actorId: 'u1' } })

    expect(wrapper.text()).not.toContain('Desativar')
  })

  it('emite edição e alternância de ativo', async () => {
    const wrapper = mount(AdminUserList, { props })

    await wrapper.findAll('.user-row__button')[0]?.trigger('click')
    await wrapper.findAll('.user-row__button')[1]?.trigger('click')

    expect(wrapper.emitted('edit')?.[0]).toEqual([user])
    expect(wrapper.emitted('toggle')?.[0]).toEqual([user])
  })

  it('mostra estado vazio', () => {
    const wrapper = mount(AdminUserList, { props: { ...props, isEmpty: true, items: [] } })

    expect(wrapper.text()).toContain('Nenhum usuário encontrado')
  })
})
