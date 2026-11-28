import { getData, patchData } from './http'
import type { AdminPixInput, AdminPixSettings } from '../types/adminPix'

const BASE_PATH = '/api/v1/admin/integrations/pix'

export function fetchAdminPix(): Promise<AdminPixSettings> {
  return getData<AdminPixSettings>(BASE_PATH)
}

export function updateAdminPix(input: AdminPixInput): Promise<AdminPixSettings> {
  return patchData<AdminPixSettings>(BASE_PATH, input)
}
