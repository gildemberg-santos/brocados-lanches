import { getData, patchData } from './http'
import type { AdminWebhookInput, AdminWebhookSettings } from '../types/adminWebhook'

const BASE_PATH = '/api/v1/admin/integrations/webhook'

export function fetchAdminWebhook(): Promise<AdminWebhookSettings> {
  return getData<AdminWebhookSettings>(BASE_PATH)
}

export function updateAdminWebhook(input: AdminWebhookInput): Promise<AdminWebhookSettings> {
  return patchData<AdminWebhookSettings>(BASE_PATH, input)
}
