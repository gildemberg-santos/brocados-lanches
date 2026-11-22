import { expect, test } from '@playwright/test'

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
})
