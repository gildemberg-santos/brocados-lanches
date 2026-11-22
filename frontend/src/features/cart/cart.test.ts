import { describe, expect, it } from 'vitest'

import {
  makeAddon,
  makeCatalogProduct,
  makeOption,
  makeOptionGroup,
} from '../../test/fixtures/catalog'
import {
  buildCartLines,
  calculateSubtotalCents,
  createCartItem,
  createCartKey,
  normalizeQuantity,
} from './cart'

const group = makeOptionGroup({
  id: 'doneness',
  name: 'Ponto da carne',
  required: true,
  minSelections: 1,
  maxSelections: 1,
  options: [
    makeOption({ id: 'rare', name: 'Ao ponto', priceCents: 0 }),
    makeOption({ id: 'well-done', name: 'Bem passado', priceCents: 200 }),
  ],
})
const addon = makeAddon({ id: 'bacon', name: 'Bacon crocante', priceCents: 500 })
const product = makeCatalogProduct({
  id: 'product-1',
  name: 'Brocado Cheddar',
  priceCents: 2300,
  optionGroups: [group],
  addons: [addon],
})

describe('regras do carrinho', () => {
  it('normaliza quantidade para inteiro dentro dos limites', () => {
    expect(normalizeQuantity(0)).toBe(1)
    expect(normalizeQuantity(3.8)).toBe(3)
    expect(normalizeQuantity(500)).toBe(99)
    expect(normalizeQuantity(Number.NaN)).toBe(1)
  })

  it('gera a mesma chave para escolhas equivalentes', () => {
    const first = createCartKey({
      productId: 'product-1',
      optionIds: ['rare', 'extra'],
      addonIds: ['bacon', 'cheese'],
      note: ' Sem cebola ',
    })
    const second = createCartKey({
      productId: 'product-1',
      optionIds: ['extra', 'rare'],
      addonIds: ['cheese', 'bacon'],
      note: 'sem cebola',
    })

    expect(first).toBe(second)
  })

  it('cria item apenas com IDs, quantidade e escolhas do catálogo', () => {
    const item = createCartItem(
      product,
      { doneness: ['well-done'] },
      ['bacon'],
      '  sem cebola  ',
      2,
    )

    expect(item).toEqual({
      key: createCartKey({
        productId: 'product-1',
        optionIds: ['well-done'],
        addonIds: ['bacon'],
        note: 'sem cebola',
      }),
      productId: 'product-1',
      optionIds: ['well-done'],
      addonIds: ['bacon'],
      quantity: 2,
      note: 'sem cebola',
    })
    expect(JSON.stringify(item)).not.toContain('priceCents')
  })

  it('rejeita produto esgotado, opção obrigatória ausente e adicional inválido', () => {
    expect(
      createCartItem({ ...product, available: false }, { doneness: ['rare'] }, [], '', 1),
    ).toBeNull()
    expect(createCartItem(product, { doneness: [] }, [], '', 1)).toBeNull()
    expect(createCartItem(product, { doneness: ['unknown'] }, [], '', 1)).toBeNull()
    expect(createCartItem(product, { doneness: ['rare'] }, ['unknown'], '', 1)).toBeNull()
  })

  it('recalcula preço e nomes a partir do catálogo e sinaliza itens indisponíveis', () => {
    const validItem = createCartItem(
      product,
      { doneness: ['well-done'] },
      ['bacon'],
      'sem cebola',
      2,
    )
    expect(validItem).not.toBeNull()

    const staleItem = {
      ...validItem!,
      key: 'stale-key',
      productId: 'removed-product',
    }
    const invalidItem = {
      ...validItem!,
      key: 'invalid-key',
      optionIds: ['missing-option'],
    }

    const result = buildCartLines(
      [validItem!, staleItem, invalidItem],
      [product, makeCatalogProduct({ id: 'removed-product', available: false })],
    )

    expect(result.lines).toHaveLength(1)
    expect(result.lines[0]).toMatchObject({
      optionNames: ['Ponto da carne: Bem passado'],
      addonNames: ['Bacon crocante'],
      note: 'sem cebola',
      quantity: 2,
      unitPriceCents: 3000,
      lineTotalCents: 6000,
    })
    expect(result.unavailableKeys).toEqual(['stale-key', 'invalid-key'])
    expect(calculateSubtotalCents(result.lines)).toBe(6000)
  })
})
