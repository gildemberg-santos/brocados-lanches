import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AdminAuditView from './AdminAuditView.vue'
import { fetchAdminAuditLogs } from '../../services/adminUsers'

vi.mock('../../services/adminUsers', () => ({
  fetchAdminAuditLogs: vi.fn(),
  fetchAdminUsers: vi.fn(),
  createAdminUser: vi.fn(),
  updateAdminUser: vi.fn(),
}))

beforeEach(() => {
  setActivePinia(createPinia())
  vi.clearAllMocks()
  vi.mocked(fetchAdminAuditLogs).mockResolvedValue({
    data: [
      {
        id: 'a1',
        actorId: 'actor',
        action: 'update',
        entity: 'products',
        entityId: 'p1',
        before: {},
        after: {},
        createdAt: '2026-11-22T12:00:00Z',
      },
    ],
    meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
  })
})

describe('AdminAuditView', () => {
  it('carrega e exibe a auditoria', async () => {
    const wrapper = mount(AdminAuditView)
    await flushPromises()

    expect(wrapper.get('h1').text()).toBe('Auditoria')
    expect(wrapper.text()).toContain('Produtos')
  })

  it('filtra por entidade e ação', async () => {
    const wrapper = mount(AdminAuditView)
    await flushPromises()

    await wrapper.find('#audit-entity').setValue('orders')
    await wrapper.find('#audit-action').setValue('create')
    await wrapper.find('.audit-filters').trigger('submit')
    await flushPromises()

    expect(fetchAdminAuditLogs).toHaveBeenLastCalledWith(
      expect.objectContaining({
        entity: 'orders',
        action: 'create',
      }),
    )
  })
})
