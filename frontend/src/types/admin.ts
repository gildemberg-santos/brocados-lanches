export type AdminRole = 'owner' | 'operator'

export interface AdminUser {
  id: string
  name: string
  email: string
  role: AdminRole
}

export interface AdminPageMeta {
  page: number
  pageSize: number
  total: number
  totalPages: number
}

export interface AdminPage<T> {
  data: T[]
  meta: AdminPageMeta
}
