import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'

import { makeCatalogProduct, makeOption, makeOptionGroup } from '../test/fixtures/catalog'
import { CART_SCHEMA_VERSION, CART_STORAGE_KEY } from '../features/cart/storage'
import { useCartStore } from './cart'

const product = makeCatalogProduct({
  id: 'product-1',
  name: 'Brocado Cheddar',
  optionGroups: [
    makeOptionGroup({
      id: 'doneness',
      options: [makeOption({ id: 'rare', name: 'Ao ponto' })],
    }),
  ],
})

describe('cart store', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
  })

  it('começa vazio na versão atual do schema', () => {
    const cart = useCartStore()

    expect(cart.version).toBe(CART_SCHEMA_VERSION)
    expect(cart.isEmpty).toBe(true)
    expect(cart.count).toBe(0)
  })

  it('adiciona itens, soma configurações iguais e persist IDs', () => {
    const cart = useCartStore()

    expect(cart.add(product, { doneness: ['rare'] }, [], 'sem cebola', 1)).toBe(true)
    expect(cart.add(product, { doneness: ['rare'] }, [], 'SEM CEBOLA', 2)).toBe(true)

    expect(cart.items).toHaveLength(1)
    expect(cart.items[0]).toMatchObject({ optionIds: ['rare'], quantity: 3, note: 'sem cebola' })
    expect(cart.count).toBe(3)
    expect(localStorage.getItem(CART_STORAGE_KEY)).toContain('product-1')
    expect(localStorage.getItem(CART_STORAGE_KEY)).not.toContain('2300')
  })

  it('recusa produto indisponível ou configuração inválida', () => {
    const cart = useCartStore()

    expect(cart.add({ ...product, available: false }, { doneness: ['rare'] }, [], '', 1)).toBe(
      false,
    )
    expect(cart.add(product, { doneness: [] }, [], '', 1)).toBe(false)
    expect(cart.add(product, { doneness: ['rare'] }, ['unknown'], '', 1)).toBe(false)
    expect(cart.isEmpty).toBe(true)
  })

  it('altera, remove e limita quantidades', () => {
    const cart = useCartStore()
    cart.add(product, { doneness: ['rare'] }, [], '', 1)
    const [item] = cart.items

    cart.increase(item.key)
    expect(cart.count).toBe(2)

    cart.setQuantity(item.key, 500)
    expect(cart.items[0].quantity).toBe(99)

    cart.setQuantity(item.key, 0)
    expect(cart.isEmpty).toBe(true)

    cart.add(product, { doneness: ['rare'] }, [], '', 1)
    cart.decrease(cart.items[0].key)
    expect(cart.isEmpty).toBe(true)
  })

  it('limpa o carrinho e persist o estado vazio', () => {
    const cart = useCartStore()
    cart.add(product, { doneness: ['rare'] }, [], '', 2)

    cart.clear()

    expect(cart.isEmpty).toBe(true)
    expect(JSON.parse(localStorage.getItem(CART_STORAGE_KEY) ?? '{}').items).toEqual([])
  })

  it('recupera itens salvos em uma nova instância da store', () => {
    const first = useCartStore()
    first.add(product, { doneness: ['rare'] }, [], 'sem cebola', 2)

    setActivePinia(createPinia())
    const second = useCartStore()

    expect(second.count).toBe(2)
    expect(second.items[0].note).toBe('sem cebola')
  })

  it('remove itens indisponíveis com aviso que pode ser dispensado', () => {
    const cart = useCartStore()
    cart.add(product, { doneness: ['rare'] }, [], '', 1)
    const key = cart.items[0].key

    cart.removeUnavailable([key])

    expect(cart.isEmpty).toBe(true)
    expect(cart.notice).toBe('Itens que não estão mais disponíveis foram removidos do carrinho.')

    cart.dismissNotice()
    expect(cart.notice).toBe('')
  })

  it('exibe o aviso de migração de um carrinho salvo em versão antiga', () => {
    localStorage.setItem(CART_STORAGE_KEY, JSON.stringify({ version: 0, items: [] }))

    const cart = useCartStore()

    expect(cart.notice).toBe('O formato do carrinho mudou. Os itens anteriores foram removidos.')
    expect(cart.isEmpty).toBe(true)
  })
})
