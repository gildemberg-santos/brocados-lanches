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
})
