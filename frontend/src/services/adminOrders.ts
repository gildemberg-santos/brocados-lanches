import { get, getData, patchData } from './http'
import type { AdminPage } from '../types/admin'
import type {
  AdminOrderDetail,
  AdminOrderListItem,
  AdminOrderListQuery,
  AdminOrderStatusRequest,
} from '../types/order'

const BASE_PATH = '/api/v1/admin/orders'

function buildQuery(query: AdminOrderListQuery): string {
  const params = new URLSearchParams()
  const values: Record<string, string | number | undefined> = {
    page: query.page,
    pageSize: query.pageSize,
    status: query.status,
    fulfillment: query.fulfillment,
    from: query.from,
    to: query.to,
    q: query.q,
    sort: query.sort,
    order: query.order,
  }

  for (const [key, value] of Object.entries(values)) {
    if (value !== undefined && value !== '') {
      params.set(key, String(value))
    }
  }

  const serialized = params.toString()
  return serialized ? `?${serialized}` : ''
}

export function fetchAdminOrders(
  query: AdminOrderListQuery = {},
): Promise<AdminPage<AdminOrderListItem>> {
  return get<AdminPage<AdminOrderListItem>>(`${BASE_PATH}${buildQuery(query)}`)
}

export function fetchAdminOrder(id: string): Promise<AdminOrderDetail> {
  return getData<AdminOrderDetail>(`${BASE_PATH}/${encodeURIComponent(id)}`)
}

export function updateAdminOrderStatus(
  id: string,
  request: AdminOrderStatusRequest,
): Promise<AdminOrderDetail> {
  return patchData<AdminOrderDetail>(`${BASE_PATH}/${encodeURIComponent(id)}/status`, request)
}
