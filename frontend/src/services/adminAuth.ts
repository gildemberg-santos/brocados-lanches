import type { AdminUser } from '../types/admin'

import { getData, postData, postEmpty } from './http'

interface AdminAuthResponse {
  user: AdminUser
}

/** A sessão é lida somente no servidor; o guard confia no 401 da API. */
export function fetchAdminUser(): Promise<AdminUser> {
  return getData<AdminAuthResponse>('/api/v1/admin/auth/me').then((body) => body.user)
}

export function loginAdmin(email: string, password: string): Promise<AdminUser> {
  return postData<AdminAuthResponse>('/api/v1/admin/auth/login', { email, password }).then(
    (body) => body.user,
  )
}

export function logoutAdmin(): Promise<void> {
  return postEmpty('/api/v1/admin/auth/logout')
}
