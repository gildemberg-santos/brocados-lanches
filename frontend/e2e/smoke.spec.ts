import { expect, test } from '@playwright/test'

test.describe('smoke da base', () => {
  test('home renderiza e mostra a API conectada', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByRole('heading', { level: 1, name: 'Estrutura inicial' })).toBeVisible()
    await expect(page.getByText('API: conectada')).toBeVisible()
    await expect(page.getByText('MongoDB: ok')).toBeVisible()
  })

  test('health e readiness respondem pelo proxy', async ({ request }) => {
    const health = await request.get('/healthz')
    expect(health.ok()).toBe(true)
    expect(await health.json()).toEqual({ status: 'ok' })

    const ready = await request.get('/readyz')
    expect(ready.ok()).toBe(true)
    expect(await ready.json()).toMatchObject({ status: 'ok', checks: { mongo: 'ok' } })
  })

  test('botão principal tem alvo de toque e não cria rolagem horizontal', async ({ page }) => {
    await page.goto('/')

    const button = page.getByRole('button', { name: 'Verificar novamente' })
    await expect(button).toBeVisible()

    const box = await button.boundingBox()
    expect(box?.height ?? 0).toBeGreaterThanOrEqual(44)

    const viewportWidth = page.viewportSize()?.width ?? 0
    const scrollWidth = await page.evaluate(() => document.documentElement.scrollWidth)
    expect(scrollWidth).toBeLessThanOrEqual(viewportWidth)
  })
})
