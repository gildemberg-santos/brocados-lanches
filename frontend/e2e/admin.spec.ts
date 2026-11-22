import { randomUUID } from 'node:crypto'

import { expect, test } from '@playwright/test'
import type { APIRequestContext } from '@playwright/test'

const API_URL = 'http://127.0.0.1:8080'
const EMAIL = 'dono@teste.local'
const PASSWORD = 'senha-de-teste-12345'
const CUSTOMER_PHONE = '+5585999990001'

interface MenuPayload {
  data: {
    categories: Array<{
      products: Array<{
        id: string
        available: boolean
        optionGroups: Array<{
          required: boolean
          options: Array<{ id: string; available: boolean }>
        }>
      }>
    }>
  }
}

interface ZonesPayload {
  data: Array<{ id: string }>
}

interface OrderPayload {
  data: { orderNumber: number; publicId: string }
}

async function createTestOrder(request: APIRequestContext): Promise<OrderPayload['data']> {
  const menuResponse = await request.get(`${API_URL}/api/v1/menu`)
  expect(menuResponse.ok()).toBeTruthy()
  const menu = (await menuResponse.json()) as MenuPayload
  const product = menu.data.categories
    .flatMap((category) => category.products)
    .find((item) => item.available)
  expect(product).toBeDefined()

  const zonesResponse = await request.get(`${API_URL}/api/v1/delivery-zones`)
  expect(zonesResponse.ok()).toBeTruthy()
  const zones = (await zonesResponse.json()) as ZonesPayload
  expect(zones.data.length).toBeGreaterThan(0)

  const optionIds = (product?.optionGroups ?? [])
    .filter((group) => group.required)
    .map((group) => group.options.find((option) => option.available)?.id)
    .filter((id): id is string => Boolean(id))

  const response = await request.post(`${API_URL}/api/v1/orders`, {
    headers: { 'Idempotency-Key': randomUUID() },
    data: {
      items: [
        {
          productId: product?.id,
          quantity: 1,
          optionIds,
          addonIds: [],
          note: '',
        },
      ],
      fulfillment: 'delivery',
      zoneId: zones.data[0]?.id,
      customer: { name: 'Cliente Teste E2E', phone: CUSTOMER_PHONE },
      address: {
        street: 'Rua de Teste',
        number: '10',
        complement: '',
        district: 'Centro',
        city: 'Aquiraz',
        state: 'CE',
        zipCode: '61700000',
      },
      payment: { method: 'cash' },
      note: '',
    },
  })

  expect(response.ok()).toBeTruthy()
  return ((await response.json()) as OrderPayload).data
}

async function login(page: import('@playwright/test').Page): Promise<void> {
  await page.goto('/admin/login')
  await page.getByLabel('E-mail').fill(EMAIL)
  await page.getByLabel('Senha').fill(PASSWORD)
  await page.getByRole('button', { name: 'Entrar' }).click()
  await expect(page).toHaveURL(/\/admin$/)
}

test.describe('acesso ao painel administrativo', () => {
  test('rota protegida redireciona visitante para login com o destino', async ({ page }) => {
    await page.goto('/admin/pedidos')

    await expect(page).toHaveURL(/\/admin\/login\?redirect=/)
    expect(new URL(page.url()).searchParams.get('redirect')).toBe('/admin/pedidos')
    await expect(page.getByRole('heading', { level: 1, name: 'Acessar painel' })).toBeVisible()
  })

  test('credenciais inválidas mostram erro genérico', async ({ page }) => {
    await page.goto('/admin/login')
    await page.getByLabel('E-mail').fill('ninguem@example.com')
    await page.getByLabel('Senha').fill('senha-errada-123')
    await page.getByRole('button', { name: 'Entrar' }).click()

    await expect(page.getByRole('alert')).toHaveText('E-mail ou senha incorretos.')
    await expect(page).toHaveURL(/\/admin\/login$/)
  })

  test('pedido criado aparece na fila e muda de status pelo painel', async ({ page, request }) => {
    const created = await createTestOrder(request)

    await login(page)
    await page.goto('/admin/pedidos')
    await expect(page.getByRole('heading', { level: 1, name: 'Pedidos' })).toBeVisible()
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth),
    ).toBeLessThanOrEqual(0)

    await page.getByLabel('Buscar').fill('0001')
    await page.getByRole('button', { name: 'Filtrar' }).click()

    const linha = page.locator('.order-row').filter({ hasText: `#${created.orderNumber}` })
    await expect(linha).toBeVisible()
    await expect(linha).toContainText('****0001')

    await linha.getByRole('link', { name: 'Ver detalhes' }).click()
    await expect(
      page.getByRole('heading', { level: 1, name: `Pedido #${created.orderNumber}` }),
    ).toBeVisible()
    await expect(page.getByText('Cliente Teste E2E')).toBeVisible()
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth),
    ).toBeLessThanOrEqual(0)

    await page.getByRole('button', { name: 'Confirmar pedido' }).click()
    await expect(page.getByRole('status')).toContainText('Status do pedido atualizado')
    await expect(page.getByText('Confirmado', { exact: true }).first()).toBeVisible()
  })
})
