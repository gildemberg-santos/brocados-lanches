import { describe, expect, it } from 'vitest'

import router from './index'

describe('router', () => {
  it('registers the home route', () => {
    const route = router.getRoutes().find((candidate) => candidate.name === 'home')

    expect(route?.path).toBe('/')
  })

  it('registers the menu route', () => {
    const route = router.getRoutes().find((candidate) => candidate.name === 'menu')

    expect(route?.path).toBe('/cardapio')
  })

  it('registers the product detail route by slug', () => {
    const route = router.getRoutes().find((candidate) => candidate.name === 'product')

    expect(route?.path).toBe('/produto/:slug')
  })

  it('registers the cart route', () => {
    const route = router.getRoutes().find((candidate) => candidate.name === 'cart')

    expect(route?.path).toBe('/carrinho')
  })

  it('registers the lazy order history route', () => {
    const route = router.getRoutes().find((candidate) => candidate.name === 'my-orders')

    expect(route?.path).toBe('/meus-pedidos')
    expect(typeof route?.components?.default).toBe('function')
    expect(route?.meta.title).toBe('Meus pedidos')
    expect(route?.meta.noindex).toBe(true)
  })

  it('registers the lazy admin login route', () => {
    const route = router.getRoutes().find((candidate) => candidate.name === 'admin-login')

    expect(route?.path).toBe('/admin/login')
    expect(typeof route?.components?.default).toBe('function')
  })

  it('marks owner-only admin routes with role metadata', () => {
    const usersRoute = router.getRoutes().find((candidate) => candidate.name === 'admin-users')

    expect(usersRoute?.path).toBe('/admin/usuarios')
    expect(usersRoute?.meta.roles).toEqual(['owner'])

    const webhookRoute = router.getRoutes().find((candidate) => candidate.name === 'admin-webhook')

    expect(webhookRoute?.path).toBe('/admin/integracoes/webhook')
    expect(webhookRoute?.meta.roles).toEqual(['owner'])
    expect(typeof webhookRoute?.components?.default).toBe('function')

    const pixRoute = router.getRoutes().find((candidate) => candidate.name === 'admin-pix')

    expect(pixRoute?.path).toBe('/admin/integracoes/pix')
    expect(pixRoute?.meta.roles).toEqual(['owner'])
    expect(typeof pixRoute?.components?.default).toBe('function')
  })

  it('registers the 403 admin route', () => {
    const route = router.getRoutes().find((candidate) => candidate.name === 'admin-forbidden')

    expect(route?.path).toBe('/admin/sem-acesso')
  })

  it.each([
    ['contact', '/contato', 'Contato'],
    ['privacy', '/privacidade', 'Privacidade'],
    ['terms', '/termos', 'Termos de uso'],
    ['credits', '/creditos', 'Créditos'],
  ])('registers the lazy %s route with its own metadata', (name, path, title) => {
    const route = router.getRoutes().find((candidate) => candidate.name === name)

    expect(route?.path).toBe(path)
    expect(typeof route?.components?.default).toBe('function')
    expect(route?.meta.title).toBe(title)
    expect(typeof route?.meta.description).toBe('string')
    expect(route?.meta.description).not.toHaveLength(0)
  })

  it('keeps published institutional pages indexable', () => {
    for (const name of ['privacy', 'terms', 'credits']) {
      const route = router.getRoutes().find((candidate) => candidate.name === name)

      expect(route?.meta.noindex).not.toBe(true)
    }
  })

  it('resolves unknown URLs to the not-found view', () => {
    const resolved = router.resolve('/endereco-inexistente')

    expect(resolved.name).toBe('not-found')
    expect(resolved.meta.title).toBe('Página não encontrada')
    expect(resolved.meta.noindex).toBe(true)
    expect(typeof resolved.matched[0]?.components?.default).toBe('function')
  })
})
