import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'

import type { OptionSelections } from '../features/catalog/customization'
import { createCartItem, MAX_CART_QUANTITY, normalizeQuantity } from '../features/cart/cart'
import { CART_SCHEMA_VERSION, loadCart, saveCart } from '../features/cart/storage'
import type { CartItem } from '../types/cart'
import type { CatalogProduct } from '../types/catalog'

const UNAVAILABLE_NOTICE = 'Itens que não estão mais disponíveis foram removidos do carrinho.'

export const useCartStore = defineStore('cart', () => {
  const loaded = loadCart()
  const items = ref<CartItem[]>(loaded.snapshot.items)
  const notice = ref(loaded.notice)
  const version = CART_SCHEMA_VERSION

  watch(
    items,
    (value) => {
      saveCart({ version, items: value })
    },
    { deep: true, flush: 'sync' },
  )

  const count = computed(() => items.value.reduce((total, item) => total + item.quantity, 0))
  const isEmpty = computed(() => items.value.length === 0)

  function add(
    product: CatalogProduct,
    selections: OptionSelections,
    addonIds: string[],
    note: string,
    quantity = 1,
  ): boolean {
    const item = createCartItem(product, selections, addonIds, note, quantity)

    if (item === null) {
      return false
    }

    const existing = items.value.find((candidate) => candidate.key === item.key)

    if (existing) {
      existing.quantity = normalizeQuantity(existing.quantity + quantity)
    } else {
      items.value.push(item)
    }

    return true
  }

  function setQuantity(key: string, quantity: number): void {
    if (quantity < 1) {
      remove(key)
      return
    }

    const item = items.value.find((candidate) => candidate.key === key)

    if (item) {
      item.quantity = normalizeQuantity(quantity)
    }
  }

  function increase(key: string): void {
    const item = items.value.find((candidate) => candidate.key === key)

    if (item) {
      setQuantity(key, item.quantity + 1)
    }
  }

  function decrease(key: string): void {
    const item = items.value.find((candidate) => candidate.key === key)

    if (item && item.quantity <= 1) {
      remove(key)
      return
    }

    if (item) {
      setQuantity(key, item.quantity - 1)
    }
  }

  function remove(key: string): void {
    items.value = items.value.filter((item) => item.key !== key)
  }

  function clear(): void {
    items.value = []
  }

  function removeUnavailable(keys: string[]): void {
    if (keys.length === 0) {
      return
    }

    const removed = new Set(keys)
    items.value = items.value.filter((item) => !removed.has(item.key))
    notice.value = UNAVAILABLE_NOTICE
  }

  function dismissNotice(): void {
    notice.value = ''
  }

  return {
    version,
    items,
    notice,
    count,
    isEmpty,
    add,
    setQuantity,
    increase,
    decrease,
    remove,
    clear,
    removeUnavailable,
    dismissNotice,
    maxQuantity: MAX_CART_QUANTITY,
  }
})
