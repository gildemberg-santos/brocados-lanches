import { expect, test } from '@playwright/test'

test.describe('smoke da base', () => {
  test('home renderiza a marca e o CTA do cardápio', async ({ page }) => {
    await page.goto('/')

    await expect(
      page.getByRole('heading', { level: 1, name: 'Seu lanche, do seu jeito.' }),
    ).toBeVisible()
    await expect(page.getByRole('link', { name: 'Abrir cardápio' })).toBeVisible()
  })

  test('navega da home para o cardápio', async ({ page }) => {
    await page.goto('/')
    await page.getByRole('link', { name: 'Abrir cardápio' }).click()

    await expect(page).toHaveURL(/\/cardapio$/)
    await expect(
      page.getByRole('heading', { level: 1, name: 'Escolha seu próximo lanche' }),
    ).toBeVisible()
  })

  test('health e readiness respondem pelo proxy', async ({ request }) => {
    const health = await request.get('/healthz')
    expect(health.ok()).toBe(true)
    expect(await health.json()).toEqual({ status: 'ok' })

    const ready = await request.get('/readyz')
    expect(ready.ok()).toBe(true)
    expect(await ready.json()).toMatchObject({ status: 'ok', checks: { mongo: 'ok' } })
  })

  test('CTA principal tem alvo de toque e não cria rolagem horizontal', async ({ page }) => {
    await page.goto('/')

    const cta = page.getByRole('link', { name: 'Abrir cardápio' })
    await expect(cta).toBeVisible()

    const box = await cta.boundingBox()
    expect(box?.height ?? 0).toBeGreaterThanOrEqual(44)

    const viewportWidth = page.viewportSize()?.width ?? 0
    const scrollWidth = await page.evaluate(() => document.documentElement.scrollWidth)
    expect(scrollWidth).toBeLessThanOrEqual(viewportWidth)
  })
})
