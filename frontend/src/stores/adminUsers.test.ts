import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import {
  createAdminUser,
  fetchAdminAuditLogs,
  fetchAdminUsers,
  updateAdminUser,
} from '../services/adminUsers'
import { useAdminUsersStore } from './adminUsers'
import type { AdminUser } from '../types/adminUsers'

vi.mock('../services/adminUsers', () => ({
  createAdminUser: vi.fn(),
  fetchAdminAuditLogs: vi.fn(),
  fetchAdminUsers: vi.fn(),
  updateAdminUser: vi.fn(),
}))

const user: AdminUser = {
  id: 'u1',
  name: 'Ana',
  email: 'ana@example.com',
  role: 'operator',
  active: true,
  lastLoginAt: null,
  createdAt: '2026-11-20T12:00:00Z',
  updatedAt: '2026-11-20T12:00:00Z',
}
const page = { data: [user], meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 } }

beforeEach(() => {
  setActivePinia(createPinia())
  vi.clearAllMocks()
  vi.mocked(fetchAdminUsers).mockResolvedValue(page)
  vi.mocked(fetchAdminAuditLogs).mockResolvedValue({ data: [], meta: page.meta })
})

describe('useAdminUsersStore', () => {
  it('carrega usuários e metadados', async () => {
    const store = useAdminUsersStore()
    await store.load()

    expect(store.state).toBe('success')
    expect(store.items).toEqual([user])
    expect(store.meta.total).toBe(1)
  })

  it('aplica filtros e paginação', async () => {
    const store = useAdminUsersStore()
    await store.load()
    await store.applyFilters({ q: 'ana' })
    await store.goToPage(2)

    expect(fetchAdminUsers).toHaveBeenLastCalledWith(expect.objectContaining({ page: 2 }))
    expect(store.filters.q).toBe('ana')
  })

  it('cria usuário e recarrega a lista', async () => {
    vi.mocked(createAdminUser).mockResolvedValue(user)
    const store = useAdminUsersStore()
    await store.load()

    const saved = await store.save({
      name: 'Ana',
      email: 'ana@example.com',
      password: 'senha-forte-123',
      role: 'operator',
    })

    expect(saved).toBe(true)
    expect(createAdminUser).toHaveBeenCalledOnce()
    expect(fetchAdminUsers).toHaveBeenCalledTimes(2)
    expect(store.saveSuccess).toContain('senha temporária')
  })

  it('atualiza usuário existente', async () => {
    vi.mocked(updateAdminUser).mockResolvedValue(user)
    const store = useAdminUsersStore()
    await store.load()

    await store.save(
      { name: 'Ana', email: 'ana@example.com', password: '', role: 'operator' },
      'u1',
    )

    expect(updateAdminUser).toHaveBeenCalledWith('u1', expect.any(Object))
  })

  it('alterna ativo e recarrega', async () => {
    vi.mocked(updateAdminUser).mockResolvedValue(user)
    const store = useAdminUsersStore()
    await store.load()

    await store.toggle('u1', false)

    expect(updateAdminUser).toHaveBeenCalledWith('u1', { active: false })
  })

  it('reporta erro de validação sem expor detalhes técnicos', async () => {
    vi.mocked(createAdminUser).mockRejectedValue(new Error('mongo down'))
    const store = useAdminUsersStore()

    const saved = await store.save({
      name: 'Ana',
      email: 'ana@example.com',
      password: 'senha-forte-123',
      role: 'operator',
    })

    expect(saved).toBe(false)
    expect(store.saveError).toContain('Não foi possível salvar')
    expect(store.saveError).not.toContain('mongo')
  })

  it('carrega auditoria com filtros', async () => {
    const store = useAdminUsersStore()
    await store.loadAudit()
    await store.applyAuditFilters({ entity: 'products' })

    expect(fetchAdminAuditLogs).toHaveBeenLastCalledWith(
      expect.objectContaining({ entity: 'products' }),
    )
    expect(store.auditState).toBe('success')
  })
})
