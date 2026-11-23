import { getData, patchData } from './http'
import type { AdminStoreInput, AdminStoreSettings } from '../types/adminCatalog'

const BASE_PATH = '/api/v1/admin/store'

export function fetchAdminStore(): Promise<AdminStoreSettings> {
  return getData<AdminStoreSettings>(BASE_PATH)
}

export function updateAdminStore(input: AdminStoreInput): Promise<AdminStoreSettings> {
  return patchData<AdminStoreSettings>(BASE_PATH, input)
}
