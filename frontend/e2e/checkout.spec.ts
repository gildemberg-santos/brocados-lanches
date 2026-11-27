import { expect, test } from '@playwright/test'
import type { APIRequestContext, Page } from '@playwright/test'

import { CHECKOUT_OPERATOR, ensureAdminApi } from './helpers/api'

const API_URL = 'http://127.0.0.1:8080'

interface TrackedOrder {
  data: {
    orderNumber: number
    status: string
    items: Array<{ name: string; quantity: number }>
    totalCents: number
  }
}

function uniquePhone(): string {
  // E.164 válido para o validador: +55 + DDD(2) + 9 dígitos = 13 dígitos.
  const suffix = String(Math.floor(Math.random() * 100_000)).padStart(5, '0')
  return `+55859999${suffix}`
}

async function expectNoHorizontalScroll(page: Page): Promise<void> {
  const overflow = await page.evaluate(
    () => document.documentElement.scrollWidth - window.innerWidth,
  )
  expect(overflow).toBeLessThanOrEqual(0)
}

async function assertOrderPersisted(
  request: APIRequestContext,
  publicId: string,
  expectedTotalCents: number,
): Promise<number> {
  const response = await request.get(`${API_URL}/api/v1/orders/${publicId}`)
  expect(response.ok()).toBeTruthy()
  const order = (await response.json()) as TrackedOrder
  expect(order.data.status).toBe('pending')
  expect(order.data.items.length).toBeGreaterThan(0)
  expect(order.data.totalCents).toBe(expectedTotalCents)

  const rawBody = await response.text()
  for (const privateField of ['customer', 'address', 'phone']) {
    expect(rawBody).not.toContain(privateField)
  }
  return order.data.orderNumber
}

test.describe('fluxo completo de pedido pela interface', () => {
  test.beforeAll(async ({ playwright }) => {
    const api = await ensureAdminApi(playwright, CHECKOUT_OPERATOR)
    await api.dispose()
  })

  test('cardápio → personalização → carrinho → checkout → pedido persistido no painel', async ({
    page,
    request,
  }) => {
    await page.goto('/cardapio')
    await expect(
      page.getByRole('heading', { level: 1, name: 'Escolha seu próximo lanche' }),
    ).toBeVisible()
    await expectNoHorizontalScroll(page)

    const productLink = page.locator('.product-card__link').first()
    const productName = await productLink.locator('.product-card__name').innerText()
    await productLink.click()

    await expect(page.getByRole('heading', { level: 1, name: productName })).toBeVisible()
    await page.getByLabel('Observação opcional').fill('Sem cebola, por favor')
    await page.getByRole('button', { name: 'Adicionar ao carrinho' }).click()
    await expect(page.getByRole('status')).toContainText('Adicionado ao carrinho')

    await page.getByRole('link', { name: 'Ver carrinho' }).click()
    await expect(page.getByRole('heading', { level: 1, name: 'Seu carrinho' })).toBeVisible()
    await expect(page.getByText('Obs.: Sem cebola, por favor')).toBeVisible()
    await expectNoHorizontalScroll(page)

    await page
      .getByRole('group', { name: `quantidade de ${productName}` })
      .getByRole('button', { name: `Aumentar quantidade de ${productName}` })
      .click()
    await expect(page.locator('.cart-line .quantity-stepper__value')).toHaveText('2')

    await page.getByRole('button', { name: 'Continuar' }).click()
    await expect(page.getByRole('heading', { level: 1, name: 'Finalizar pedido' })).toBeVisible()

    const zoneOptions = page.getByLabel('Zona de entrega').locator('option')
    await page.getByLabel('Zona de entrega').selectOption({ index: 1 })
    expect(await zoneOptions.count()).toBeGreaterThan(1)

    await page.getByLabel('Rua', { exact: true }).fill('Rua de Teste E2E')
    await page.getByLabel('Número', { exact: true }).fill('10')
    await page.getByLabel('Bairro', { exact: true }).fill('Centro')
    await page.getByLabel('CEP', { exact: true }).fill('61700000')
    await page.getByLabel('Nome completo').fill('Cliente Fluxo E2E')
    const phone = uniquePhone()
    await page.getByLabel('Telefone').fill(phone)
    await page.getByLabel('Pix').check()
    await page.getByLabel('Observação (opcional)').fill('Pedido de teste automatizado')

    await expect(page.locator('.checkout-summary .summary-totals__final dd')).not.toHaveText(
      'R$ 0,00',
    )
    const displayedTotal = await page
      .locator('.checkout-summary .summary-totals__final dd')
      .innerText()
    const expectedTotalCents = Number(displayedTotal.replace(/[^\d,]/g, '').replace(',', '.')) * 100

    await expectNoHorizontalScroll(page)
    await page.getByRole('button', { name: 'Enviar pedido' }).click()

    await expect(page.getByRole('heading', { level: 1, name: 'Pedido registrado' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Abrir WhatsApp', exact: true })).toHaveAttribute(
      'href',
      /^https:\/\/wa\.me\/\d+\?text=/,
    )
    const publicId = new URL(page.url()).pathname.split('/').pop() ?? ''
    expect(publicId).not.toBe('')
    await expectNoHorizontalScroll(page)

    // O handoff fica na sessão: se a aba recarregar, o link continua disponível.
    await page.reload()
    await expect(page.getByRole('link', { name: 'Abrir WhatsApp', exact: true })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Copiar link' })).toBeVisible()

    const orderNumber = await assertOrderPersisted(
      request,
      publicId,
      Math.round(expectedTotalCents),
    )

    await page.goto('/admin/login')
    await page.getByLabel('E-mail').fill(CHECKOUT_OPERATOR.email)
    await page.getByLabel('Senha').fill(CHECKOUT_OPERATOR.password)
    await page.getByRole('button', { name: 'Entrar' }).click()
    await expect(page).toHaveURL(/\/admin$/)

    await page.goto('/admin/pedidos')
    await page.getByLabel('Buscar').fill(String(orderNumber).padStart(4, '0'))
    await page.getByRole('button', { name: 'Filtrar' }).click()
    const linha = page.locator('.order-row').filter({ hasText: `#${orderNumber}` })
    await expect(linha).toBeVisible()
    expect(await linha.innerText()).not.toContain(phone)

    // O histórico local reencontra o pedido sem senha nem consulta por telefone.
    await page.goto('/meus-pedidos')
    await expect(page.getByRole('heading', { level: 1, name: 'Meus pedidos' })).toBeVisible()
    await expectNoHorizontalScroll(page)
    const card = page.locator('.history-card').filter({ hasText: `#${orderNumber}` })
    await expect(card).toBeVisible()
    await expect(card).toHaveAttribute('href', `/pedido/${publicId}`)
    const cardText = await card.innerText()
    expect(cardText).not.toContain(phone)
    expect(cardText).not.toContain('Cliente Fluxo E2E')

    const historyToken = await page.evaluate(() => {
      const stored = localStorage.getItem('brocados:order-access-token')
      if (!stored) {
        return null
      }
      const parsed: unknown = JSON.parse(stored)
      return typeof parsed === 'object' &&
        parsed !== null &&
        'token' in parsed &&
        typeof parsed.token === 'string'
        ? parsed.token
        : null
    })
    expect(historyToken).not.toBeNull()
    const historyResponse = await request.get(`${API_URL}/api/v1/orders/history`, {
      headers: { Authorization: `Bearer ${historyToken ?? ''}` },
    })
    expect(historyResponse.ok()).toBeTruthy()
    expect(historyResponse.headers()['cache-control']).toBe('no-store')
    const historyBody = await historyResponse.text()
    expect(historyBody).toContain(publicId)
    expect(historyBody).not.toContain(phone)
    expect(historyBody).not.toContain('Cliente Fluxo E2E')

    await page.goto('/privacidade')
    await page.getByRole('button', { name: 'Apagar acesso aos meus pedidos' }).click()
    await expect(page.getByRole('status')).toContainText('Acesso aos pedidos apagado')
    await page.goto('/meus-pedidos')
    await expect(
      page.getByRole('heading', { level: 2, name: 'Nenhum histórico neste navegador' }),
    ).toBeVisible()
  })
})
