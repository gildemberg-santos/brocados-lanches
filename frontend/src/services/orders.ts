import type {
  CreateOrderRequest,
  CreatedOrder,
  OrderHistoryPage,
  OrderHistoryQuery,
  OrderQuote,
  OrderQuoteRequest,
  TrackedOrder,
} from '../types/order'

import { get, getData, postData } from './http'

function bearerHeaders(accessToken: string): Record<string, string> {
  return { Authorization: `Bearer ${accessToken}` }
}

export function fetchOrderQuote(request: OrderQuoteRequest): Promise<OrderQuote> {
  return postData<OrderQuote>('/api/v1/orders/quote', request)
}

export function createOrder(
  request: CreateOrderRequest,
  idempotencyKey: string,
  accessToken?: string,
): Promise<CreatedOrder> {
  return postData<CreatedOrder>('/api/v1/orders', request, {
    idempotencyKey,
    ...(accessToken ? { headers: bearerHeaders(accessToken) } : {}),
  })
}

export function fetchOrderHistory(
  accessToken: string,
  query: OrderHistoryQuery = {},
): Promise<OrderHistoryPage> {
  const params = new URLSearchParams()
  if (query.page !== undefined) {
    params.set('page', String(query.page))
  }
  if (query.pageSize !== undefined) {
    params.set('pageSize', String(query.pageSize))
  }
  const search = params.toString()

  return get<OrderHistoryPage>(`/api/v1/orders/history${search ? `?${search}` : ''}`, {
    headers: bearerHeaders(accessToken),
  })
}

export function fetchTrackedOrder(publicId: string): Promise<TrackedOrder> {
  return getData<TrackedOrder>(`/api/v1/orders/${encodeURIComponent(publicId)}`)
}
