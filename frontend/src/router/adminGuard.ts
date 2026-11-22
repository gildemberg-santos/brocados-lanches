import type { NavigationGuardWithThis } from 'vue-router'

import { useAdminAuthStore } from '../stores/adminAuth'

function hasRequiredRole(roles: unknown, role: string | null): boolean {
  if (!Array.isArray(roles) || roles.length === 0) {
    return true
  }
  return role !== null && roles.includes(role)
}

/** Guard de conveniência; a autorização real está no servidor. */
export const requireAdminAuth: NavigationGuardWithThis<undefined> = async (to) => {
  const auth = useAdminAuthStore()
  await auth.bootstrap()

  if (!auth.isAuthenticated) {
    return { name: 'admin-login', query: { redirect: to.fullPath } }
  }
  if (!hasRequiredRole(to.meta.roles, auth.role)) {
    return { name: 'admin-forbidden' }
  }
  return true
}

export const redirectAuthenticatedFromLogin: NavigationGuardWithThis<undefined> = async () => {
  const auth = useAdminAuthStore()
  await auth.bootstrap()
  return auth.isAuthenticated ? { name: 'admin-dashboard' } : true
}
