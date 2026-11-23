import type {
  CreateOrderRequest,
  CreatedOrder,
  OrderQuote,
  OrderQuoteRequest,
  TrackedOrder,
} from '../types/order'

import { getData, postData } from './http'

export function fetchOrderQuote(request: OrderQuoteRequest): Promise<OrderQuote> {
  return postData<OrderQuote>('/api/v1/orders/quote', request)
}

export function createOrder(
  request: CreateOrderRequest,
  idempotencyKey: string,
): Promise<CreatedOrder> {
  return postData<CreatedOrder>('/api/v1/orders', request, { idempotencyKey })
}

export function fetchTrackedOrder(publicId: string): Promise<TrackedOrder> {
  return getData<TrackedOrder>(`/api/v1/orders/${encodeURIComponent(publicId)}`)
}
