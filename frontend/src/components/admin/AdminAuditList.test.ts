import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import AdminAuditList from './AdminAuditList.vue'
import type { AdminAuditLog } from '../../types/adminUsers'

const log: AdminAuditLog = {
  id: 'a1',
  actorId: 'actor',
  action: 'update',
  entity: 'products',
  entityId: 'p1',
  before: {},
  after: {},
  createdAt: '2026-11-22T12:00:00Z',
}

const props = {
  items: [log],
  state: 'success' as const,
  error: '',
  isEmpty: false,
  meta: { page: 1, pageSize: 20, total: 1, totalPages: 2 },
}

describe('AdminAuditList', () => {
  it('mostra entidade e ação sem PII', () => {
    const wrapper = mount(AdminAuditList, { props })

    expect(wrapper.text()).toContain('Produtos')
    expect(wrapper.text()).toContain('Atualização')
    expect(wrapper.text()).not.toContain('phone')
  })

  it('emite troca de página', async () => {
    const wrapper = mount(AdminAuditList, { props })
    const buttons = wrapper.findAll('.audit-list__button')

    await buttons[1]?.trigger('click')

    expect(wrapper.emitted('page')?.[0]).toEqual([2])
  })

  it('mostra estado vazio', () => {
    const wrapper = mount(AdminAuditList, { props: { ...props, isEmpty: true, items: [] } })

    expect(wrapper.text()).toContain('Nenhuma alteração registrada')
  })
})
