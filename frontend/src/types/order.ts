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

export interface CreateOrderRequest extends OrderQuoteRequest {
  customer: OrderCustomer
  address?: OrderAddress
  payment: {
    method: PaymentMethod
    changeForCents?: number
  }
  note: string
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
