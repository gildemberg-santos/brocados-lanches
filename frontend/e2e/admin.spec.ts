import { randomUUID } from 'node:crypto'

import { expect, test } from '@playwright/test'
import type { APIRequestContext, Cookie, Page } from '@playwright/test'

import { ensureAdminApi, uniquePhone } from './helpers/api'

const API_URL = 'http://127.0.0.1:8080'

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

type CreatedTestOrder = OrderPayload['data'] & { phone: string }

async function createTestOrder(request: APIRequestContext): Promise<CreatedTestOrder> {
  const phone = uniquePhone()
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
      customer: { name: 'Cliente Teste E2E', phone },
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
  return { ...((await response.json()) as OrderPayload).data, phone }
}

// O login é feito uma vez por API e reaproveitado nos testes que só precisam
// de sessão: isso evita estourar o limite de 5 logins/min por e-mail.
let adminCookies: Cookie[] = []

test.describe('acesso ao painel administrativo', () => {
  test.beforeAll(async ({ playwright }) => {
    const api = await ensureAdminApi(playwright)
    try {
      adminCookies = (await api.storageState()).cookies
    } finally {
      await api.dispose()
    }
  })

  async function useAdminSession(page: Page): Promise<void> {
    await page.context().addCookies(adminCookies)
  }

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

  test('cadastro abre em modal centralizado e ESC fecha sem sair da lista', async ({ page }) => {
    await useAdminSession(page)
    await page.goto('/admin/categorias')

    await expect(page.getByRole('heading', { level: 1, name: 'Categorias' })).toBeVisible()
    await page.getByRole('button', { name: 'Nova categoria' }).click()

    const dialog = page.getByRole('dialog', { name: 'Nova categoria' })
    await expect(dialog).toBeVisible()
    const box = await dialog.boundingBox()
    const viewport = page.viewportSize()
    if (!box || !viewport) {
      throw new Error('Modal sem dimensões para validar o centramento')
    }
    expect(Math.abs(box.x + box.width / 2 - viewport.width / 2)).toBeLessThan(2)

    await page.keyboard.press('Escape')
    await expect(dialog).toHaveCount(0)
    expect(new URL(page.url()).pathname).toBe('/admin/categorias')
  })

  test('navegação e modal cabem na viewport em mobile e desktop', async ({ page }) => {
    await useAdminSession(page)
    const viewport = page.viewportSize()
    if (!viewport) {
      throw new Error('Viewport indefinida para validar o painel')
    }

    await page.goto('/admin')
    if (viewport.width < 768) {
      const menu = page.getByTestId('admin-menu-button')
      await expect(menu).toBeVisible()
      await menu.click()

      const drawer = page.getByTestId('admin-drawer')
      await expect(drawer).toBeVisible()
      expect(await page.evaluate(() => document.body.style.overflow)).toBe('hidden')

      await drawer.getByRole('link', { name: 'Pedidos' }).click()
      await expect(page).toHaveURL(/\/admin\/pedidos$/)
      await expect(drawer).toHaveCount(0)
      expect(await page.evaluate(() => document.body.style.overflow)).toBe('')
    } else {
      await expect(page.getByRole('navigation', { name: 'Navegação do painel' })).toBeVisible()
      await expect(page.getByTestId('admin-menu-button')).toBeHidden()
    }

    await page.goto('/admin/categorias')
    await page.getByRole('button', { name: 'Nova categoria' }).click()
    const dialog = page.getByRole('dialog')
    await expect(dialog).toBeVisible()
    const box = await dialog.boundingBox()
    if (!box) {
      throw new Error('Modal sem dimensões para validar a viewport')
    }
    expect(box.y).toBeGreaterThanOrEqual(0)
    expect(box.y + box.height).toBeLessThanOrEqual(viewport.height)
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth),
    ).toBeLessThanOrEqual(0)
  })

  test('checkboxes de ativação seguem o padrão neon do design system', async ({ page }) => {
    await useAdminSession(page)

    const telas = [
      { url: '/admin/produtos', abrir: 'Novo produto' },
      { url: '/admin/categorias', abrir: 'Nova categoria' },
      { url: '/admin/loja' },
    ]

    for (const tela of telas) {
      await page.goto(tela.url)
      if (tela.abrir) {
        await page.getByRole('button', { name: tela.abrir }).click()
        await expect(page.getByRole('dialog')).toBeVisible()
      }

      await expect(page.locator('input[type="checkbox"]').first()).toBeVisible()
      const estilos = await page.locator('input[type="checkbox"]').evaluateAll((inputs) => {
        const referencia = document.createElement('span')
        referencia.style.color = 'var(--brand-primary)'
        document.body.appendChild(referencia)
        const corDaMarca = getComputedStyle(referencia).color
        referencia.remove()

        return inputs.map((input) => {
          const style = getComputedStyle(input)
          return {
            accentColor: style.accentColor,
            corDaMarca,
            width: style.width,
            height: style.height,
          }
        })
      })

      expect(estilos.length, `checkboxes em ${tela.url}`).toBeGreaterThan(0)
      for (const estilo of estilos) {
        expect(estilo.accentColor, `accent em ${tela.url}`).toBe(estilo.corDaMarca)
        expect(estilo.width, `largura em ${tela.url}`).toBe('18px')
        expect(estilo.height, `altura em ${tela.url}`).toBe('18px')
      }
    }
  })

  test('pedido criado aparece na fila e muda de status pelo painel', async ({ page, request }) => {
    const created = await createTestOrder(request)

    await useAdminSession(page)
    await page.goto('/admin/pedidos')
    await expect(page.getByRole('heading', { level: 1, name: 'Pedidos' })).toBeVisible()
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth),
    ).toBeLessThanOrEqual(0)

    await page.getByLabel('Buscar').fill(String(created.orderNumber).padStart(4, '0'))
    await page.getByRole('button', { name: 'Filtrar' }).click()

    const linha = page.locator('.order-row').filter({ hasText: `#${created.orderNumber}` })
    await expect(linha).toBeVisible()
    await expect(linha).toContainText(`****${created.phone.slice(-4)}`)

    await linha.getByRole('link', { name: 'Ver detalhes' }).click()
    await expect(
      page.getByRole('heading', { level: 1, name: `Pedido #${created.orderNumber}` }),
    ).toBeVisible()
    await expect(page.getByText('Cliente Teste E2E')).toBeVisible()
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth),
    ).toBeLessThanOrEqual(0)

    await page.getByRole('button', { name: 'Confirmar pedido' }).click()
    await expect(page.locator('.admin-page__success')).toContainText('Status do pedido atualizado')
    await expect(page.getByText('Confirmado', { exact: true }).first()).toBeVisible()
  })

  test('mudança de status chega ao acompanhamento sem recarregar a página', async ({
    page,
    request,
  }) => {
    const created = await createTestOrder(request)
    await page.goto(`/pedido/${created.publicId}`)
    await expect(page.getByText('Atualização automática ativa.')).toBeVisible()

    const adminPage = await page.context().newPage()
    await useAdminSession(adminPage)
    await adminPage.goto('/admin/pedidos')
    await adminPage.getByLabel('Buscar').fill(String(created.orderNumber))
    await adminPage.getByRole('button', { name: 'Filtrar' }).click()
    const linha = adminPage.locator('.order-row').filter({ hasText: `#${created.orderNumber}` })
    await linha.getByRole('link', { name: 'Ver detalhes' }).click()
    await adminPage.getByRole('button', { name: 'Confirmar pedido' }).click()
    await expect(adminPage.locator('.admin-page__success')).toBeVisible()

    await expect(page.locator('.order-hero h2')).toHaveText('Confirmado', { timeout: 15_000 })
    expect(new URL(page.url()).pathname).toBe(`/pedido/${created.publicId}`)
    await adminPage.close()
  })
})
