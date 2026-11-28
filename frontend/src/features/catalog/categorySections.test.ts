import { describe, expect, it } from 'vitest'

import {
  CATEGORY_EDGE_OFFSET,
  categorySectionId,
  resolveActiveCategory,
  type CategorySectionMetric,
} from './categorySections'

const sections: CategorySectionMetric[] = [
  { slug: 'combos', top: 300, bottom: 900 },
  { slug: 'hamburgueres', top: 950, bottom: 1600 },
  { slug: 'bebidas', top: 1650, bottom: 2100 },
]

describe('categorySections', () => {
  it('gera um ID estável a partir do slug', () => {
    expect(categorySectionId('hamburgueres')).toBe('categoria-hamburgueres')
  })

  it('sempre ativa a primeira categoria no topo da página', () => {
    expect(resolveActiveCategory(sections, 0, 'down', 2000)).toBe('combos')
    expect(resolveActiveCategory(sections, 20, 'up', 2000)).toBe('combos')
  })

  it('sempre ativa a última categoria no fim da página', () => {
    expect(resolveActiveCategory(sections, 2000, 'up', 2000)).toBe('bebidas')
  })

  it('ao descer, ativa a última seção cujo topo passou do limite', () => {
    expect(resolveActiveCategory(sections, 500, 'down', 2000)).toBe('combos')
    expect(resolveActiveCategory(sections, 930, 'down', 2000)).toBe('combos')
    expect(resolveActiveCategory(sections, 950, 'down', 2000)).toBe('hamburgueres')
    expect(resolveActiveCategory(sections, 1661, 'down', 2000)).toBe('bebidas')
  })

  it('ao subir, ativa a primeira seção cuja base ainda passa do limite', () => {
    expect(resolveActiveCategory(sections, 1800, 'up', 2000)).toBe('bebidas')
    expect(resolveActiveCategory(sections, 1490, 'up', 2000)).toBe('hamburgueres')
    expect(resolveActiveCategory(sections, 850, 'up', 2000)).toBe('combos')
  })

  it('usa um limite negativo para a última seção ao subir', () => {
    const withShortLast = [
      { slug: 'primeira', top: 100, bottom: 200 },
      { slug: 'ultima', top: 250, bottom: 300 },
    ]

    expect(resolveActiveCategory(withShortLast, 301, 'up', 500)).toBe('ultima')
    expect(resolveActiveCategory(withShortLast, 389, 'up', 500)).toBe('ultima')
    expect(resolveActiveCategory(withShortLast, 391, 'up', 500)).toBe('primeira')
  })

  it('respeita o limite de 10 px e não ativa seção sem medição', () => {
    expect(CATEGORY_EDGE_OFFSET).toBe(10)
    expect(resolveActiveCategory([], 500, 'down', 2000)).toBe('')
  })
})
