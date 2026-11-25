import { randomUUID } from 'node:crypto'

import { expect } from '@playwright/test'
import type { APIRequestContext, APIResponse, Playwright } from '@playwright/test'

export const API_URL = 'http://127.0.0.1:8080'

const OWNER_EMAIL = 'dono@teste.local'
const OWNER_PASSWORD = 'senha-de-teste-12345'

// Login por e-mail tem limite de 5/min por IP. Um operador exclusivo dos
// cenários de falha evita competir com os logins das demais specs.
const OPERATOR = {
  name: 'Operador Cenários E2E',
  email: 'operador@teste.local',
  password: 'senha-de-teste-12345',
  role: 'operator',
}

export interface ApiErrorEnvelope {
  error: { code: string; message: string }
}

export interface MenuProduct {
  id: string
  slug: string
  name: string
  available: boolean
  optionGroups: Array<{
    id: string
    required: boolean
    options: Array<{ id: string; available: boolean }>
  }>
}

export interface MenuPayload {
  data: {
    categories: Array<{ products: MenuProduct[] }>
  }
}

export interface ZonesPayload {
  data: Array<{ id: string; feeCents: number }>
}

export interface QuotePayload {
  data: { subtotalCents: number; deliveryFeeCents: number; totalCents: number }
}

export interface CreatedOrderPayload {
  data: { orderNumber: number; publicId: string; status: string; totalCents: number }
}

export interface AdminOrderListPayload {
  data: Array<{ id: string; orderNumber: number; publicId: string }>
}

export interface AdminOrderDetailPayload {
  data: {
    id: string
    orderNumber: number
    status: string
    fulfillment: string
    deliveryFeeCents: number
    totalCents: number
    payment: { method: string; changeForCents: number }
    statusHistory: Array<{ status: string }>
  }
}

export interface OrderRequest {
  items: Array<{
    productId: string
    quantity: number
    optionIds: string[]
    addonIds: string[]
    note: string
  }>
  fulfillment: 'delivery' | 'pickup'
  zoneId?: string
  customer: { name: string; phone: string }
  address: {
    street: string
    number: string
    complement: string
    district: string
    city: string
    state: string
    zipCode: string
  }
  payment: { method: string; changeForCents?: number }
  note: string
}

export interface CatalogFixture {
  product: MenuProduct
  zoneId: string
}

export function uniquePhone(): string {
  const suffix = String(Math.floor(Math.random() * 100_000)).padStart(5, '0')
  return `+55859999${suffix}`
}

export async function expectApiError(response: APIResponse, code: string): Promise<void> {
  expect(response.ok()).toBeFalsy()
  const body = (await response.json()) as ApiErrorEnvelope
  expect(body.error.code).toBe(code)
}

export async function loginAdmin(context: APIRequestContext): Promise<void> {
  const response = await context.post('/api/v1/admin/auth/login', {
    data: { email: OWNER_EMAIL, password: OWNER_PASSWORD },
  })
  expect(response.ok()).toBeTruthy()
}

export async function loginTestOperator(context: APIRequestContext): Promise<void> {
  const response = await context.post('/api/v1/admin/auth/login', {
    data: { email: OPERATOR.email, password: OPERATOR.password },
  })
  expect(response.ok()).toBeTruthy()
}

export async function ensureAdminApi(playwright: Playwright): Promise<APIRequestContext> {
  const context = await playwright.request.newContext({ baseURL: API_URL })
  const loginOperator = () =>
    context.post('/api/v1/admin/auth/login', {
      data: { email: OPERATOR.email, password: OPERATOR.password },
    })

  let response = await loginOperator()
  if (!response.ok()) {
    const owner = await playwright.request.newContext({ baseURL: API_URL })
    try {
      await loginAdmin(owner)
      const created = await owner.post('/api/v1/admin/users', { data: OPERATOR })
      expect([201, 409]).toContain(created.status())
    } finally {
      await owner.dispose()
    }
    response = await loginOperator()
  }
  expect(response.ok()).toBeTruthy()
  return context
}

export async function fetchCatalog(context: APIRequestContext): Promise<CatalogFixture> {
  const menuResponse = await context.get('/api/v1/menu')
  expect(menuResponse.ok()).toBeTruthy()
  const menu = (await menuResponse.json()) as MenuPayload
  const product = menu.data.categories
    .flatMap((category) => category.products)
    .find((item) => item.available)
  expect(product).toBeDefined()

  const zonesResponse = await context.get('/api/v1/delivery-zones')
  expect(zonesResponse.ok()).toBeTruthy()
  const zones = (await zonesResponse.json()) as ZonesPayload
  expect(zones.data.length).toBeGreaterThan(0)

  return { product: product as MenuProduct, zoneId: zones.data[0]?.id ?? '' }
}

export function buildOrderRequest(
  fixture: CatalogFixture,
  options: Partial<Pick<OrderRequest, 'fulfillment' | 'payment'>> & { quantity?: number } = {},
): OrderRequest {
  const fulfillment = options.fulfillment ?? 'delivery'
  const optionIds = (fixture.product.optionGroups ?? [])
    .filter((group) => group.required)
    .map((group) => group.options.find((option) => option.available)?.id)
    .filter((id): id is string => Boolean(id))

  return {
    items: [
      {
        productId: fixture.product.id,
        quantity: options.quantity ?? 1,
        optionIds,
        addonIds: [],
        note: '',
      },
    ],
    fulfillment,
    ...(fulfillment === 'delivery' ? { zoneId: fixture.zoneId } : {}),
    customer: { name: 'Cliente Cenários E2E', phone: uniquePhone() },
    address: {
      street: 'Rua de Teste',
      number: '10',
      complement: '',
      district: 'Centro',
      city: 'Aquiraz',
      state: 'CE',
      zipCode: '61700000',
    },
    payment: options.payment ?? { method: 'pix' },
    note: '',
  }
}

export async function quoteOrder(
  context: APIRequestContext,
  request: OrderRequest,
): Promise<APIResponse> {
  // A cotação não aceita dados pessoais, pagamento nem observação do pedido.
  return context.post('/api/v1/orders/quote', {
    data: {
      items: request.items,
      fulfillment: request.fulfillment,
      ...(request.zoneId ? { zoneId: request.zoneId } : {}),
    },
  })
}

export async function createOrder(
  context: APIRequestContext,
  request: OrderRequest,
  key = randomUUID(),
): Promise<APIResponse> {
  return context.post('/api/v1/orders', {
    headers: { 'Idempotency-Key': key },
    data: request,
  })
}

export async function adminOrderIdByNumber(
  context: APIRequestContext,
  orderNumber: number,
): Promise<string> {
  const response = await context.get(`/api/v1/admin/orders?q=${orderNumber}&page=1&pageSize=100`)
  expect(response.ok()).toBeTruthy()
  const list = (await response.json()) as AdminOrderListPayload
  const order = list.data.find((item) => item.orderNumber === orderNumber)
  expect(order).toBeDefined()
  return (order as AdminOrderListPayload['data'][number]).id
}

export async function adminOrderDetail(
  context: APIRequestContext,
  id: string,
): Promise<AdminOrderDetailPayload['data']> {
  const response = await context.get(`/api/v1/admin/orders/${id}`)
  expect(response.ok()).toBeTruthy()
  return ((await response.json()) as AdminOrderDetailPayload).data
}
