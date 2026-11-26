import type { CustomerProfileLookupResponse } from '../types/order'

import { postData } from './http'

/** Consulta explícita de conveniência; o telefone vai no corpo, nunca na URL. */
export function fetchCustomerProfile(phone: string): Promise<CustomerProfileLookupResponse> {
  return postData<CustomerProfileLookupResponse>('/api/v1/customers/profile', { phone })
}
