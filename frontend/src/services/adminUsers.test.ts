import { beforeEach, describe, expect, it, vi } from 'vitest'

import { get, patchData, postData } from './http'
import {
  createAdminUser,
  fetchAdminAuditLogs,
  fetchAdminUsers,
  updateAdminUser,
} from './adminUsers'

vi.mock('./http', () => ({ get: vi.fn(), patchData: vi.fn(), postData: vi.fn() }))

beforeEach(() => vi.clearAllMocks())

describe('adminUsers service', () => {
  it('lista usuários com paginação e filtros', async () => {
    vi.mocked(get).mockResolvedValue({ data: [], meta: {} })

    await fetchAdminUsers({ page: 2, pageSize: 10, q: 'ana', active: true })

    expect(get).toHaveBeenCalledWith('/api/v1/admin/users?page=2&pageSize=10&q=ana&active=true')
  })

  it('cria usuário com senha temporária', async () => {
    vi.mocked(postData).mockResolvedValue({ id: 'u1' })

    await createAdminUser({
      name: 'Ana',
      email: 'ana@example.com',
      password: 'senha-forte-123',
      role: 'operator',
    })

    expect(postData).toHaveBeenCalledWith(
      '/api/v1/admin/users',
      expect.objectContaining({
        email: 'ana@example.com',
        role: 'operator',
      }),
    )
  })

  it('atualiza usuário pelo id codificado', async () => {
    vi.mocked(patchData).mockResolvedValue({ id: 'u1' })

    await updateAdminUser('abc/123', { active: false })

    expect(patchData).toHaveBeenCalledWith('/api/v1/admin/users/abc%2F123', { active: false })
  })

  it('lista auditoria com filtros', async () => {
    vi.mocked(get).mockResolvedValue({ data: [], meta: {} })

    await fetchAdminAuditLogs({ entity: 'products', from: '2026-11-01T00:00:00Z' })

    expect(get).toHaveBeenCalledWith(
      '/api/v1/admin/audit-logs?entity=products&from=2026-11-01T00%3A00%3A00Z',
    )
  })
})
