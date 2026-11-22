export type AdminRole = 'owner' | 'operator'

export interface AdminUser {
  id: string
  name: string
  email: string
  role: AdminRole
}
