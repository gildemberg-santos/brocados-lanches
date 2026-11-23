import { get, patchData, postData } from './http'
import type { AdminPage } from '../types/admin'
import type {
  AdminAuditListQuery,
  AdminAuditLog,
  AdminUser,
  AdminUserCreateInput,
  AdminUserUpdateInput,
} from '../types/adminUsers'

const BASE_PATH = '/api/v1/admin'

function buildQuery(query: AdminAuditListQuery): string {
  const params = new URLSearchParams()
  const values: Record<string, string | number | boolean | undefined> = {
    page: query.page,
    pageSize: query.pageSize,
    q: query.q,
    active: query.active,
    sort: query.sort,
    order: query.order,
    actorId: query.actorId,
    action: query.action,
    entity: query.entity,
    from: query.from,
    to: query.to,
  }
  for (const [key, value] of Object.entries(values)) {
    if (value !== undefined && value !== '') {
      params.set(key, String(value))
    }
  }
  const serialized = params.toString()
  return serialized ? `?${serialized}` : ''
}

export function fetchAdminUsers(query: AdminAuditListQuery = {}): Promise<AdminPage<AdminUser>> {
  return get<AdminPage<AdminUser>>(`${BASE_PATH}/users${buildQuery(query)}`)
}

export function createAdminUser(input: AdminUserCreateInput): Promise<AdminUser> {
  return postData<AdminUser>(`${BASE_PATH}/users`, input)
}

export function updateAdminUser(id: string, input: AdminUserUpdateInput): Promise<AdminUser> {
  return patchData<AdminUser>(`${BASE_PATH}/users/${encodeURIComponent(id)}`, input)
}

export function fetchAdminAuditLogs(
  query: AdminAuditListQuery = {},
): Promise<AdminPage<AdminAuditLog>> {
  return get<AdminPage<AdminAuditLog>>(`${BASE_PATH}/audit-logs${buildQuery(query)}`)
}
