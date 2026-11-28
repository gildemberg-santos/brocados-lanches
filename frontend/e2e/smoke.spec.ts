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

  test('viewport cobre a safe-area do dispositivo', async ({ page }) => {
    await page.goto('/')

    const content = await page.locator('meta[name="viewport"]').getAttribute('content')
    expect(content).toContain('viewport-fit=cover')
  })

  test('navegação inferior aparece só no mobile e leva ao cardápio', async ({ page }, testInfo) => {
    await page.goto('/')
    const tabBar = page.getByTestId('mobile-tab-bar')

    if (testInfo.project.name === 'mobile') {
      await expect(tabBar).toBeVisible()
      await tabBar.getByRole('link', { name: 'Cardápio' }).click()
      await expect(page).toHaveURL(/\/cardapio$/)
    } else {
      await expect(tabBar).toBeHidden()
    }
  })

  test('rodapé termina acima da barra fixa no mobile', async ({ page }, testInfo) => {
    if (testInfo.project.name !== 'mobile') {
      test.skip()
    }

    await page.goto('/cardapio')
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight))

    const layout = await page.evaluate(() => {
      const footer = document.querySelector('.site-footer')
      const tab = document.querySelector('[data-testid="mobile-tab-bar"]')
      const spacer = document.querySelector('[data-testid="mobile-bottom-spacer"]')
      return {
        footerBottom: footer?.getBoundingClientRect().bottom ?? 0,
        tabTop: tab?.getBoundingClientRect().top ?? 0,
        spacerBg: spacer ? getComputedStyle(spacer).backgroundColor : '',
        bodyBg: getComputedStyle(document.body).backgroundColor,
      }
    })

    expect(layout.footerBottom).toBeLessThan(layout.tabTop)
    expect(layout.spacerBg).toBe(layout.bodyBg)
  })

  test('categorias grudam abaixo do header e o clique reposiciona a seção', async ({ page }) => {
    await page.goto('/cardapio')

    const nav = page.getByRole('navigation', { name: 'Categorias do cardápio' })
    await expect(nav).toBeVisible()

    const firstChip = nav.getByRole('button').first()
    const categoryName = (await firstChip.textContent())?.trim() ?? ''
    expect(categoryName.length).toBeGreaterThan(0)
    await expect(firstChip).toHaveAttribute('aria-pressed', 'true')

    // A faixa de chips pode rolar horizontalmente, mas a página não pode.
    const overflow = await page.evaluate(() => ({
      viewport: document.documentElement.clientWidth,
      content: document.documentElement.scrollWidth,
    }))
    expect(overflow.content).toBeLessThanOrEqual(overflow.viewport)

    await page.evaluate(() => window.scrollTo(0, 1200))

    const headerBottom = await page
      .locator('.site-header')
      .evaluate((element) => Math.round(element.getBoundingClientRect().bottom))
    await expect
      .poll(() => nav.evaluate((element) => Math.round(element.getBoundingClientRect().top)))
      .toBeGreaterThanOrEqual(headerBottom)

    await firstChip.click()

    await expect
      .poll(() =>
        page
          .getByRole('heading', { level: 2, name: categoryName })
          .evaluate((element) => Math.round(element.getBoundingClientRect().top)),
      )
      .toBeLessThan(180)
  })

  test('o último chip continua alcançável pela faixa rolável', async ({ page }) => {
    await page.goto('/cardapio')

    const nav = page.getByRole('navigation', { name: 'Categorias do cardápio' })
    const lastChip = nav.getByRole('button').last()
    const label = (await lastChip.textContent())?.trim() ?? ''

    await page.evaluate(() => {
      const inner = document.querySelector('[aria-label="Categorias do cardápio"] > div')
      if (inner) inner.scrollLeft = inner.scrollWidth
    })
    await lastChip.click()

    await expect(lastChip).toHaveAttribute('aria-pressed', 'true')
    await expect(page).toHaveURL(/categoria=/)
    await expect(page.getByRole('heading', { level: 2, name: label })).toBeInViewport()
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
