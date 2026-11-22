import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchAdminUser, loginAdmin, logoutAdmin } from '../services/adminAuth'
import { ApiError } from '../services/http'
import type { AdminUser } from '../types/admin'
import { useAdminAuthStore } from './adminAuth'

vi.mock('../services/adminAuth', () => ({
  fetchAdminUser: vi.fn(),
  loginAdmin: vi.fn(),
  logoutAdmin: vi.fn(),
}))

const owner: AdminUser = {
  id: 'user-1',
  name: 'Dono',
  email: 'dono@example.com',
  role: 'owner',
}

const operator: AdminUser = {
  id: 'user-2',
  name: 'Operador',
  email: 'operador@example.com',
  role: 'operator',
}

describe('store de autenticação administrativa', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('carrega a sessão uma vez e expõe papel e nome', async () => {
    vi.mocked(fetchAdminUser).mockResolvedValue(owner)
    const auth = useAdminAuthStore()

    await auth.bootstrap()
    await auth.bootstrap()

    expect(fetchAdminUser).toHaveBeenCalledTimes(1)
    expect(auth.isAuthenticated).toBe(true)
    expect(auth.isOwner).toBe(true)
    expect(auth.role).toBe('owner')
    expect(auth.userName).toBe('Dono')
    expect(auth.loading).toBe(false)
  })

  it('trata 401 como visitante sem mensagem de erro', async () => {
    vi.mocked(fetchAdminUser).mockRejectedValue(new ApiError(401, 'UNAUTHORIZED', 'Sem sessão'))
    const auth = useAdminAuthStore()

    await auth.bootstrap()

    expect(auth.isAuthenticated).toBe(false)
    expect(auth.error).toBe('')
    expect(auth.initialized).toBe(true)
  })

  it('mostra mensagem em erro inesperado e permite recarregar', async () => {
    vi.mocked(fetchAdminUser)
      .mockRejectedValueOnce(new Error('offline'))
      .mockResolvedValueOnce(operator)
    const auth = useAdminAuthStore()

    await auth.bootstrap()

    expect(auth.error).toContain('Não foi possível verificar a sessão')

    await auth.bootstrap(true)

    expect(auth.user).toEqual(operator)
    expect(auth.isOwner).toBe(false)
    expect(fetchAdminUser).toHaveBeenCalledTimes(2)
  })

  it('faz login e guarda o usuário devolvido pela API', async () => {
    vi.mocked(loginAdmin).mockResolvedValue(operator)
    const auth = useAdminAuthStore()

    await expect(auth.login('operador@example.com', 'senha-forte-123')).resolves.toBe(true)

    expect(auth.user).toEqual(operator)
    expect(auth.error).toBe('')
    expect(auth.isAuthenticated).toBe(true)
  })

  it('devolve mensagem genérica quando o login falha', async () => {
    vi.mocked(loginAdmin).mockRejectedValue(new ApiError(401, 'UNAUTHORIZED', 'credenciais'))
    const auth = useAdminAuthStore()

    await expect(auth.login('dono@example.com', 'errada')).resolves.toBe(false)

    expect(auth.user).toBeNull()
    expect(auth.error).toBe('E-mail ou senha incorretos.')
  })

  it('limpa a sessão local no logout mesmo se a API falhar', async () => {
    vi.mocked(loginAdmin).mockResolvedValue(owner)
    vi.mocked(logoutAdmin).mockRejectedValue(new Error('offline'))
    const auth = useAdminAuthStore()
    await auth.login('dono@example.com', 'senha-forte-123')

    await auth.logout()

    expect(auth.user).toBeNull()
    expect(auth.isAuthenticated).toBe(false)
    expect(auth.loading).toBe(false)
  })

  it('limpa estado com clear', async () => {
    vi.mocked(fetchAdminUser).mockResolvedValue(owner)
    const auth = useAdminAuthStore()
    await auth.bootstrap()

    auth.clear()

    expect(auth.user).toBeNull()
    expect(auth.initialized).toBe(true)
  })
})
