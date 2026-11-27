import type { AdminPageMeta } from './admin'

export type Fulfillment = 'delivery' | 'pickup'

export interface OrderQuoteRequest {
  items: Array<{
    productId: string
    quantity: number
    optionIds: string[]
    addonIds: string[]
    note: string
  }>
  fulfillment: Fulfillment
  zoneId?: string
}

export interface OrderQuoteLine {
  productId: string
  name: string
  quantity: number
  unitPriceCents: number
  optionsPriceCents: number
  lineTotalCents: number
  options: Array<{ name: string; priceCents: number }>
}

export interface OrderQuote {
  subtotalCents: number
  deliveryFeeCents: number
  totalCents: number
  currency: string
  items: OrderQuoteLine[]
  warnings: string[]
}

export interface OrderCustomer {
  name: string
  phone: string
}

export interface OrderAddress {
  street: string
  number: string
  complement: string
  district: string
  city: string
  state: string
  zipCode: string
}

export type PaymentMethod = 'cash' | 'card' | 'pix'

/** Dados mínimos devolvidos pela consulta de conveniência por telefone. */
export interface CustomerProfileLookup {
  name?: string
  phone?: string
  fulfillment?: Fulfillment
  zone?: { id: string; name: string }
  address?: OrderAddress
  paymentMethod?: PaymentMethod
}

export interface CustomerProfileLookupResponse {
  found: boolean
  profile?: CustomerProfileLookup
}

export interface CreateOrderRequest extends OrderQuoteRequest {
  customer: OrderCustomer
  address?: OrderAddress
  payment: {
    method: PaymentMethod
    changeForCents?: number
  }
  note: string
  saveCustomerProfile?: boolean
}

export interface CreatedOrder {
  orderNumber: number
  publicId: string
  status: string
  fulfillment: Fulfillment
  subtotalCents: number
  deliveryFeeCents: number
  totalCents: number
  createdAt: string
  whatsappUrl: string
  accessToken: string
}

export interface TrackedOrder {
  orderNumber: number
  publicId: string
  status: string
  fulfillment: Fulfillment
  items: Array<{ name: string; quantity: number }>
  subtotalCents: number
  deliveryFeeCents: number
  totalCents: number
  createdAt: string
  updatedAt: string
  estimatedTime?: string
}

export interface OrderHistoryQuery {
  page?: number
  pageSize?: number
}

export interface OrderHistoryPage {
  data: TrackedOrder[]
  meta: AdminPageMeta
}

export interface AdminOrderListItem {
  id: string
  orderNumber: number
  publicId: string
  status: string
  fulfillment: Fulfillment
  customerName: string
  phone: string
  itemCount: number
  subtotalCents: number
  deliveryFeeCents: number
  totalCents: number
  createdAt: string
  updatedAt: string
}

export interface AdminOrderItemOption {
  groupId?: string
  groupName?: string
  optionId: string
  name: string
  kind: string
  priceCents: number
}

export interface AdminOrderItem {
  productId: string
  name: string
  unitPriceCents: number
  quantity: number
  options: AdminOrderItemOption[]
  note?: string
  lineTotalCents: number
}

export interface AdminOrderZone {
  id: string
  name: string
  feeCents: number
}

export interface AdminOrderAddress extends OrderAddress {
  reference?: string
}

export interface AdminOrderStatusChange {
  status: string
  at: string
  actorId?: string
}

export interface AdminCustomerProfile {
  name: string
  phone: string
  orderCount: number
  lastOrderAt?: string
  address?: AdminOrderAddress
  fulfillment?: Fulfillment
  zoneName?: string
  paymentMethod?: PaymentMethod
  optedOut: boolean
  updatedAt?: string
  retainedUntil?: string
}

export interface AdminOrderDetail {
  id: string
  orderNumber: number
  publicId: string
  status: string
  fulfillment: Fulfillment
  customer: OrderCustomer
  customerProfile?: AdminCustomerProfile
  address?: AdminOrderAddress
  zone?: AdminOrderZone
  items: AdminOrderItem[]
  subtotalCents: number
  deliveryFeeCents: number
  totalCents: number
  payment: {
    method: PaymentMethod
    changeForCents?: number
  }
  note?: string
  statusHistory: AdminOrderStatusChange[]
  confirmedAt?: string
  cancelledAt?: string
  cancelReason?: string
  createdAt: string
  updatedAt: string
}

export interface AdminOrderListQuery {
  page?: number
  pageSize?: number
  status?: string
  fulfillment?: Fulfillment
  from?: string
  to?: string
  q?: string
  sort?: 'createdAt' | 'number' | 'status' | 'totalCents'
  order?: 'asc' | 'desc'
}

export interface AdminOrderStatusRequest {
  status: string
  reason?: string
}
