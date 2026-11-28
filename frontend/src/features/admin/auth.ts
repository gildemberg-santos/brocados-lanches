import { ApiError } from '../../services/http'
import type { AdminRole } from '../../types/admin'

export interface AdminNavItem {
  label: string
  name: string
  roles?: AdminRole[]
}

/** Itens do painel na ordem exibida no menu; papéis ausentes são de todos. */
export const adminNavItems: readonly AdminNavItem[] = [
  { label: 'Painel', name: 'admin-dashboard' },
  { label: 'Pedidos', name: 'admin-orders' },
  { label: 'Produtos', name: 'admin-products' },
  { label: 'Categorias', name: 'admin-categories' },
  { label: 'Adicionais', name: 'admin-addons' },
  { label: 'Locais e taxas', name: 'admin-zones' },
  { label: 'Loja', name: 'admin-store' },
  { label: 'Webhook', name: 'admin-webhook', roles: ['owner'] },
  { label: 'Pix', name: 'admin-pix', roles: ['owner'] },
  { label: 'Conta', name: 'admin-account' },
  { label: 'Usuários', name: 'admin-users', roles: ['owner'] },
  { label: 'Auditoria', name: 'admin-audit', roles: ['owner'] },
]

const primaryNavNames = new Set([
  'admin-dashboard',
  'admin-orders',
  'admin-products',
  'admin-account',
])

export function canAccessAdminNav(item: AdminNavItem, role: AdminRole | null): boolean {
  if (!item.roles || item.roles.length === 0) {
    return true
  }
  return role !== null && item.roles.includes(role)
}

export function visibleAdminNav(role: AdminRole | null): AdminNavItem[] {
  return adminNavItems.filter((item) => canAccessAdminNav(item, role))
}

export function primaryAdminNav(role: AdminRole | null): AdminNavItem[] {
  return visibleAdminNav(role).filter((item) => primaryNavNames.has(item.name))
}

export function adminRoleLabel(role: AdminRole | null): string {
  if (role === 'owner') {
    return 'Administrador'
  }
  if (role === 'operator') {
    return 'Operador'
  }
  return 'Sem sessão'
}

export function adminAuthErrorMessage(error: unknown): string {
  if (!(error instanceof ApiError)) {
    return 'Não foi possível entrar. Verifique sua conexão e tente novamente.'
  }

  if (error.status === 401) {
    return 'E-mail ou senha incorretos.'
  }
  if (error.status === 403) {
    return 'Você não tem permissão para acessar esta área.'
  }
  if (error.status === 429) {
    return 'Muitas tentativas. Aguarde um minuto e tente novamente.'
  }
  if (error.code === 'VALIDATION_ERROR') {
    return 'Revise os dados informados e tente novamente.'
  }
  return error.message
}
