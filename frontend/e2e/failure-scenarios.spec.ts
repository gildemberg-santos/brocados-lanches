import { randomUUID } from 'node:crypto'

import { expect, test } from '@playwright/test'
import type { APIRequestContext, APIResponse } from '@playwright/test'

import {
  adminOrderDetail,
  adminOrderIdByNumber,
  buildOrderRequest,
  createOrder,
  ensureAdminApi,
  expectApiError,
  fetchCatalog,
  quoteOrder,
  SESSION_OPERATOR,
  type CatalogFixture,
  type MenuPayload,
  type QuotePayload,
} from './helpers/api'

interface AdminStorePayload {
  data: { isOpen: boolean }
}

interface AdminCatalogListPayload<T> {
  data: T[]
}

interface AdminProduct {
  id: string
  available: boolean
}

interface AdminZone {
  id: string
  active: boolean
}

interface StorePayload {
  data: { status: string }
}

interface TrackingPayload {
  data: { orderNumber: number; publicId: string }
}

test.describe.serial('cenários críticos de falha', () => {
  let adminApi: APIRequestContext
  let fixture: CatalogFixture
  let original: { isOpen: boolean; available: boolean; active: boolean }

  test.beforeAll(async ({ playwright }) => {
    adminApi = await ensureAdminApi(playwright)

    const storeResponse = await adminApi.get('/api/v1/admin/store')
    expect(storeResponse.ok()).toBeTruthy()
    const store = ((await storeResponse.json()) as AdminStorePayload).data

    const productsResponse = await adminApi.get('/api/v1/admin/products?page=1&pageSize=100')
    expect(productsResponse.ok()).toBeTruthy()
    const products = ((await productsResponse.json()) as AdminCatalogListPayload<AdminProduct>).data

    const zonesResponse = await adminApi.get('/api/v1/admin/delivery-zones?page=1&pageSize=100')
    expect(zonesResponse.ok()).toBeTruthy()
    const zones = ((await zonesResponse.json()) as AdminCatalogListPayload<AdminZone>).data

    fixture = await fetchCatalog(adminApi)
    const product = products.find((item) => item.id === fixture.product.id)
    const zone = zones.find((item) => item.id === fixture.zoneId)
    expect(product).toBeDefined()
    expect(zone).toBeDefined()
    original = {
      isOpen: store.isOpen,
      available: (product as AdminProduct).available,
      active: (zone as AdminZone).active,
    }
  })

  test.afterAll(async () => {
    if (!original || !fixture) {
      await adminApi?.dispose()
      return
    }
    await adminApi.patch('/api/v1/admin/store', { data: { isOpen: original.isOpen } })
    await adminApi.patch(`/api/v1/admin/products/${fixture.product.id}`, {
      data: { available: original.available },
    })
    await adminApi.patch(`/api/v1/admin/delivery-zones/${fixture.zoneId}`, {
      data: { active: original.active },
    })
    await adminApi.dispose()
  })

  test('loja fechada pelo kill switch bloqueia cotação e checkout', async ({ page, request }) => {
    const response = await adminApi.patch('/api/v1/admin/store', { data: { isOpen: false } })
    expect(response.ok()).toBeTruthy()

    try {
      const storeResponse = await request.get('/api/v1/store')
      expect(storeResponse.ok()).toBeTruthy()
      expect(((await storeResponse.json()) as StorePayload).data.status).toBe('closed')

      await expectApiError(await quoteOrder(request, buildOrderRequest(fixture)), 'STORE_CLOSED')

      await page.goto(`/produto/${fixture.product.slug}`)
      for (const group of fixture.product.optionGroups.filter((item) => item.required)) {
        const option = group.options.find((item) => item.available)
        if (option) {
          await page.locator(`input[name="group-${group.id}"][value="${option.id}"]`).check()
        }
      }
      await page.getByRole('button', { name: 'Adicionar ao carrinho' }).click()
      await page.goto('/checkout')
      // Retirada não exige zona: a única validação pendente é o estado da loja.
      await page.getByLabel('Retirada na loja').check()
      await expect(page.getByRole('alert')).toContainText('A loja está fechada')
    } finally {
      const restored = await adminApi.patch('/api/v1/admin/store', {
        data: { isOpen: original.isOpen },
      })
      expect(restored.ok()).toBeTruthy()
    }
  })

  test('produto esgotado rejeita a cotação e aparece no catálogo', async ({ request }) => {
    const response = await adminApi.patch(`/api/v1/admin/products/${fixture.product.id}`, {
      data: { available: false },
    })
    expect(response.ok()).toBeTruthy()

    try {
      await expectApiError(
        await quoteOrder(request, buildOrderRequest(fixture)),
        'PRODUCT_UNAVAILABLE',
      )

      const menuResponse = await request.get('/api/v1/menu')
      expect(menuResponse.ok()).toBeTruthy()
      const menu = (await menuResponse.json()) as MenuPayload
      const product = menu.data.categories
        .flatMap((category) => category.products)
        .find((item) => item.id === fixture.product.id)
      expect(product?.available).toBe(false)
    } finally {
      const restored = await adminApi.patch(`/api/v1/admin/products/${fixture.product.id}`, {
        data: { available: original.available },
      })
      expect(restored.ok()).toBeTruthy()
    }
  })

  test('zona inativa rejeita a cotação e sai da lista pública', async ({ request }) => {
    const response = await adminApi.patch(`/api/v1/admin/delivery-zones/${fixture.zoneId}`, {
      data: { active: false },
    })
    expect(response.ok()).toBeTruthy()

    try {
      await expectApiError(
        await quoteOrder(request, buildOrderRequest(fixture)),
        'ZONE_UNAVAILABLE',
      )

      const zonesResponse = await request.get('/api/v1/delivery-zones')
      expect(zonesResponse.ok()).toBeTruthy()
      const zones = (await zonesResponse.json()) as { data: Array<{ id: string }> }
      expect(zones.data.some((zone) => zone.id === fixture.zoneId)).toBe(false)
    } finally {
      const restored = await adminApi.patch(`/api/v1/admin/delivery-zones/${fixture.zoneId}`, {
        data: { active: original.active },
      })
      expect(restored.ok()).toBeTruthy()
    }
  })

  // Os projetos desktop e mobile compartilham a mesma API e o limite de 10
  // pedidos/min por IP. Os cenários que gravam pedido rodam só no desktop.
  const skipOrderCreationOnMobile = (isMobile: boolean): void => {
    test.skip(isMobile, 'criação de pedido é validada apenas no projeto desktop')
  }

  test('retirada zera a taxa, troco é congelado e transição inválida é rejeitada', async ({
    request,
    isMobile,
  }) => {
    skipOrderCreationOnMobile(isMobile)
    const payload = buildOrderRequest(fixture, {
      fulfillment: 'pickup',
      payment: { method: 'cash', changeForCents: 10_000 },
    })
    const quoteResponse = await quoteOrder(request, payload)
    expect(quoteResponse.ok()).toBeTruthy()
    const quote = ((await quoteResponse.json()) as QuotePayload).data
    expect(quote.deliveryFeeCents).toBe(0)
    expect(quote.totalCents).toBe(quote.subtotalCents)

    const createdResponse = await createOrder(request, payload)
    expect(createdResponse.status()).toBe(201)
    const created = ((await createdResponse.json()) as TrackingPayload).data
    const id = await adminOrderIdByNumber(adminApi, created.orderNumber)
    const detail = await adminOrderDetail(adminApi, id)
    expect(detail.fulfillment).toBe('pickup')
    expect(detail.deliveryFeeCents).toBe(0)
    expect(detail.totalCents).toBe(detail.deliveryFeeCents + quote.subtotalCents)
    expect(detail.payment.method).toBe('cash')
    expect(detail.payment.changeForCents).toBe(10_000)

    const invalid: APIResponse = await adminApi.patch(`/api/v1/admin/orders/${id}/status`, {
      data: { status: 'delivered' },
    })
    await expectApiError(invalid, 'CONFLICT')

    const valid = await adminApi.patch(`/api/v1/admin/orders/${id}/status`, {
      data: { status: 'confirmed' },
    })
    expect(valid.ok()).toBeTruthy()

    const updated = await adminOrderDetail(adminApi, id)
    expect(updated.status).toBe('confirmed')
    expect(updated.statusHistory.at(-1)?.status).toBe('confirmed')
  })

  test('mesma chave idempotente devolve o pedido; corpo diferente conflita', async ({
    request,
    isMobile,
  }) => {
    skipOrderCreationOnMobile(isMobile)
    const key = randomUUID()
    const payload = buildOrderRequest(fixture)

    const firstResponse = await createOrder(request, payload, key)
    expect(firstResponse.status()).toBe(201)
    const first = ((await firstResponse.json()) as TrackingPayload).data

    const replayResponse = await createOrder(request, payload, key)
    expect(replayResponse.status()).toBe(200)
    const replay = ((await replayResponse.json()) as TrackingPayload).data
    expect(replay.publicId).toBe(first.publicId)
    expect(replay.orderNumber).toBe(first.orderNumber)

    await expectApiError(
      await createOrder(request, buildOrderRequest(fixture, { quantity: 2 }), key),
      'CONFLICT',
    )
  })

  test('sessão encerrada deixa de autorizar rotas do painel', async ({ playwright }) => {
    const context = await ensureAdminApi(playwright, SESSION_OPERATOR)
    try {
      expect((await context.get('/api/v1/admin/auth/me')).ok()).toBeTruthy()

      const logout = await context.post('/api/v1/admin/auth/logout')
      expect(logout.status()).toBe(204)

      await expectApiError(await context.get('/api/v1/admin/auth/me'), 'UNAUTHORIZED')
    } finally {
      await context.dispose()
    }
  })
})
