import { createPinia, setActivePinia } from 'pinia'
import type {
  NavigationGuardNext,
  RouteLocationNormalized,
  RouteLocationNormalizedLoaded,
} from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { useAdminAuthStore } from '../stores/adminAuth'
import type { AdminRole, AdminUser } from '../types/admin'
import { redirectAuthenticatedFromLogin, requireAdminAuth } from './adminGuard'

function makeRoute(fullPath: string, roles?: AdminRole[]): RouteLocationNormalized {
  return {
    fullPath,
    meta: roles ? { roles } : {},
  } as unknown as RouteLocationNormalized
}

const from = {} as RouteLocationNormalizedLoaded

function callGuard(guard: typeof requireAdminAuth, to: RouteLocationNormalized): Promise<unknown> {
  const next = vi.fn() as NavigationGuardNext
  return guard.call(undefined, to, from, next) as Promise<unknown>
}

function user(role: AdminRole): AdminUser {
  return { id: 'user-1', name: 'Dono', email: 'dono@example.com', role }
}

describe('guard do painel administrativo', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('redireciona visitante sem sessão para login com o destino desejado', async () => {
    const auth = useAdminAuthStore()
    vi.spyOn(auth, 'bootstrap').mockResolvedValue()

    const result = await callGuard(requireAdminAuth, makeRoute('/admin/produtos'))

    expect(result).toEqual({
      name: 'admin-login',
      query: { redirect: '/admin/produtos' },
    })
  })

  it('permite a navegação de usuário autenticado', async () => {
    const auth = useAdminAuthStore()
    vi.spyOn(auth, 'bootstrap').mockResolvedValue()
    auth.user = user('operator')

    const result = await callGuard(requireAdminAuth, makeRoute('/admin/pedidos'))

    expect(result).toBe(true)
  })

  it('bloqueia operador em rota de owner', async () => {
    const auth = useAdminAuthStore()
    vi.spyOn(auth, 'bootstrap').mockResolvedValue()
    auth.user = user('operator')

    const result = await callGuard(requireAdminAuth, makeRoute('/admin/usuarios', ['owner']))

    expect(result).toEqual({ name: 'admin-forbidden' })
  })

  it('permite owner em rota de owner', async () => {
    const auth = useAdminAuthStore()
    vi.spyOn(auth, 'bootstrap').mockResolvedValue()
    auth.user = user('owner')

    const result = await callGuard(requireAdminAuth, makeRoute('/admin/usuarios', ['owner']))

    expect(result).toBe(true)
  })

  it('devolve visitante autenticado ao painel ao abrir o login', async () => {
    const auth = useAdminAuthStore()
    vi.spyOn(auth, 'bootstrap').mockResolvedValue()
    auth.user = user('operator')

    const result = await callGuard(redirectAuthenticatedFromLogin, makeRoute('/admin/login'))

    expect(result).toEqual({ name: 'admin-dashboard' })
  })

  it('mantém o login aberto para visitante sem sessão', async () => {
    const auth = useAdminAuthStore()
    vi.spyOn(auth, 'bootstrap').mockResolvedValue()

    const result = await callGuard(redirectAuthenticatedFromLogin, makeRoute('/admin/login'))

    expect(result).toBe(true)
  })
})
