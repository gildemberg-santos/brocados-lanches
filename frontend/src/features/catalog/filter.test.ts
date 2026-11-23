import { describe, expect, it } from 'vitest'

import { makeCatalogProduct } from '../../test/fixtures/catalog'
import { filterProducts, matchesProductQuery } from './filter'

const cheddar = makeCatalogProduct({
  id: 'product-1',
  name: 'Brocadão Cheddar',
  slug: 'brocadao-cheddar',
  description: 'Dois blends com cheddar cremoso.',
  categorySlug: 'hamburgueres',
})

const coke = makeCatalogProduct({
  id: 'product-2',
  name: 'Coca-Cola Lata',
  slug: 'coca-cola-lata',
  description: 'Lata gelada 350ml.',
  categorySlug: 'bebidas',
})

describe('catalog filter', () => {
  it('matches products by name ignoring accents and case', () => {
    expect(matchesProductQuery(cheddar, 'BROCADAO')).toBe(true)
  })

  it('matches products by description and requires every term', () => {
    expect(matchesProductQuery(cheddar, 'cheddar cremoso')).toBe(true)
    expect(matchesProductQuery(cheddar, 'cheddar gelada')).toBe(false)
  })

  it('matches every product when the query is blank', () => {
    expect(matchesProductQuery(coke, '   ')).toBe(true)
  })

  it('filters by category and query together', () => {
    expect(filterProducts([cheddar, coke], 'bebidas', '')).toEqual([coke])
    expect(filterProducts([cheddar, coke], '', 'lata')).toEqual([coke])
    expect(filterProducts([cheddar, coke], 'hamburgueres', 'lata')).toEqual([])
  })
})
