import type { AdminCatalogListQuery } from './adminCatalog'

export type AdminRole = 'owner' | 'operator'

export interface AdminUser {
  id: string
  name: string
  email: string
  role: AdminRole
  active: boolean
  lastLoginAt: string | null
  createdAt: string
  updatedAt: string
}

export interface AdminUserCreateInput {
  name: string
  email: string
  password: string
  role: AdminRole
}

export interface AdminUserUpdateInput {
  name?: string
  role?: AdminRole
  active?: boolean
}

export interface AdminAuditLog {
  id: string
  actorId: string
  action: 'create' | 'update'
  entity: string
  entityId: string
  before: Record<string, unknown> | null
  after: Record<string, unknown> | null
  createdAt: string
}

export interface AdminAuditListQuery extends AdminCatalogListQuery {
  actorId?: string
  action?: string
  entity?: string
  from?: string
  to?: string
}
