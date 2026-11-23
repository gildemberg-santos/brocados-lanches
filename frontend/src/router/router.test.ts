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

  it('registers the lazy admin login route', () => {
    const route = router.getRoutes().find((candidate) => candidate.name === 'admin-login')

    expect(route?.path).toBe('/admin/login')
    expect(typeof route?.components?.default).toBe('function')
  })

  it('marks owner-only admin routes with role metadata', () => {
    const route = router.getRoutes().find((candidate) => candidate.name === 'admin-users')

    expect(route?.path).toBe('/admin/usuarios')
    expect(route?.meta.roles).toEqual(['owner'])
  })

  it('registers the 403 admin route', () => {
    const route = router.getRoutes().find((candidate) => candidate.name === 'admin-forbidden')

    expect(route?.path).toBe('/admin/sem-acesso')
  })
})
