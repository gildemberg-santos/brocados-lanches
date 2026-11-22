import { beforeEach, describe, expect, it, vi } from 'vitest'

import { CART_SCHEMA_VERSION, CART_STORAGE_KEY, loadCart, saveCart } from './storage'

describe('persistência do carrinho', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.restoreAllMocks()
  })

  it('começa vazio quando não há carrinho salvo', () => {
    expect(loadCart()).toEqual({
      snapshot: { version: CART_SCHEMA_VERSION, items: [] },
      notice: '',
    })
  })

  it('persiste e recupera apenas IDs, escolhas e quantidades', () => {
    const snapshot = {
      version: CART_SCHEMA_VERSION,
      items: [
        {
          key: 'product-1|rare||sem cebola',
          productId: 'product-1',
          optionIds: ['rare'],
          addonIds: [],
          quantity: 2,
          note: 'sem cebola',
        },
      ],
    }

    saveCart(snapshot)

    expect(loadCart()).toEqual({ snapshot, notice: '' })
    expect(localStorage.getItem(CART_STORAGE_KEY)).toContain('productId')
    expect(localStorage.getItem(CART_STORAGE_KEY)).not.toContain('priceCents')
  })

  it('descarta versão antiga com aviso', () => {
    localStorage.setItem(
      CART_STORAGE_KEY,
      JSON.stringify({ version: 0, items: [{ productId: 'x' }] }),
    )

    const loaded = loadCart()

    expect(loaded.snapshot.items).toEqual([])
    expect(loaded.notice).toBe('O formato do carrinho mudou. Os itens anteriores foram removidos.')
    expect(localStorage.getItem(CART_STORAGE_KEY)).toBeNull()
  })

  it('reinicia um carrinho corrompido com aviso', () => {
    localStorage.setItem(CART_STORAGE_KEY, '{quebrado')

    const loaded = loadCart()

    expect(loaded.snapshot.items).toEqual([])
    expect(loaded.notice).toBe('Não foi possível ler o carrinho salvo. Ele foi reiniciado.')
  })

  it('remove entradas inválidas e deduplica escolhas sem perder itens válidos', () => {
    localStorage.setItem(
      CART_STORAGE_KEY,
      JSON.stringify({
        version: CART_SCHEMA_VERSION,
        items: [
          {
            productId: 'product-1',
            optionIds: ['rare', 'rare'],
            addonIds: [],
            quantity: 2,
            note: '  sem cebola  ',
          },
          { productId: '', optionIds: [], addonIds: [], quantity: 1, note: '' },
          { productId: 'product-2', optionIds: [], addonIds: [], quantity: 0, note: '' },
        ],
      }),
    )

    const loaded = loadCart()

    expect(loaded.snapshot.items).toHaveLength(1)
    expect(loaded.snapshot.items[0]).toMatchObject({
      productId: 'product-1',
      optionIds: ['rare'],
      quantity: 2,
      note: 'sem cebola',
    })
    expect(loaded.notice).toBe('Itens inválidos foram removidos do carrinho.')
    expect(JSON.parse(localStorage.getItem(CART_STORAGE_KEY) ?? '{}').items).toHaveLength(1)
  })

  it('mantém o carrinho em memória se o armazenamento falhar', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('storage blocked')
    })
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('storage blocked')
    })

    expect(loadCart().snapshot.items).toEqual([])
    expect(() => saveCart({ version: CART_SCHEMA_VERSION, items: [] })).not.toThrow()
  })
})
